package server

import (
	"bytes"
	"context"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"github.com/decred/dcrd/rpcclient/v8"
	"github.com/decred/slog"
	"github.com/vctt94/pong-bisonrelay/pongrpc/grpc/pong"
)

type DepositUpdate struct {
	PkScriptHex string
	Confs       uint32
	UTXOCount   int
	OK          bool
	At          time.Time
	UTXOs       []*pong.EscrowUTXO
}

// chainWatcher is a minimal pusher: it scans the chain/mempool for every
// pkScript that currently has at least one subscriber, and broadcasts a
// DepositUpdate each tick. No per-script state is retained.
type chainWatcher struct {
	log  slog.Logger
	dcrd *rpcclient.Client

	mu   sync.RWMutex
	tip  int64
	subs map[string]map[chan DepositUpdate]struct{} // pkScriptHex -> set(chan)

	quit chan struct{}
}

func newChainWatcher(log slog.Logger, c *rpcclient.Client) *chainWatcher {
	return &chainWatcher{
		log:  log,
		dcrd: c,
		subs: make(map[string]map[chan DepositUpdate]struct{}),
		quit: make(chan struct{}),
	}
}

func (w *chainWatcher) stop() { close(w.quit) }

func (w *chainWatcher) run(ctx context.Context) {
	w.log.Infof("watcher: started")
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
	defer w.log.Infof("watcher: stopped")
	for {
		select {
		case <-ctx.Done():
			return
		case <-w.quit:
			return
		case <-t.C:
			w.pollOnce(ctx)
		}
	}
}

func (w *chainWatcher) pollOnce(ctx context.Context) {
	// Update tip (best effort).
	if _, h, err := w.dcrd.GetBestBlock(ctx); err == nil {
		w.mu.Lock()
		w.tip = h
		w.mu.Unlock()
	} else {
		w.log.Debugf("watcher: GetBestBlock failed: %v", err)
	}

	// Snapshot subscribed pkScripts.
	w.mu.RLock()
	subsCount := len(w.subs)
	if subsCount == 0 {
		w.mu.RUnlock()
		w.log.Debugf("watcher: poll tick; no subscribers (tip=%d)", w.currentTip())
		return
	}
	keys := make([]string, 0, len(w.subs))
	for k := range w.subs {
		keys = append(keys, k)
	}
	w.mu.RUnlock()
	w.log.Debugf("watcher: poll tick; tip=%d, subs=%d", w.currentTip(), subsCount)

	// Scan each subscribed pkScript.
	for _, pkHex := range keys {
		w.log.Debugf("watcher: scanning pk=%s", pkHex)
		var utxos []*pong.EscrowUTXO
		var confs uint32
		found := false

		// Lookback over recent blocks (bounded).
		if h := w.currentTip(); h >= 0 {
			const lookback = int64(1024)
			latestMatch := int64(-1)
			best := h
			for i := int64(0); i < lookback && best-i >= 0; i++ {
				bh := best - i
				hash, err := w.dcrd.GetBlockHash(ctx, bh)
				if err != nil {
					continue
				}
				bv, err := w.dcrd.GetBlockVerbose(ctx, hash, true)
				if err != nil {
					continue
				}
				for _, tx := range bv.RawTx {
					for voutIdx, vout := range tx.Vout {
						match := false
						if spkBytes, err := hex.DecodeString(vout.ScriptPubKey.Hex); err == nil {
							if pkBytes, err := hex.DecodeString(pkHex); err == nil {
								match = bytes.Equal(spkBytes, pkBytes)
							}
						}
						if match {
							atoms := uint64(vout.Value * 1e8)
							utxos = append(utxos, &pong.EscrowUTXO{
								Txid:        tx.Txid,
								Vout:        uint32(voutIdx),
								Value:       atoms,
								PkScriptHex: pkHex,
							})
							if bh > latestMatch {
								latestMatch = bh
							}
						}
					}
				}
			}
			if len(utxos) > 0 && latestMatch >= 0 {
				confs = uint32((best - latestMatch) + 1)
				found = true
				w.log.Debugf("watcher: pk=%s found in blocks; utxos=%d confs=%d", pkHex, len(utxos), confs)
			}
		}

		// If not in recent blocks, check mempool (0-conf).
		if !found {
			if txids, err := w.dcrd.GetRawMempool(ctx, "all"); err == nil {
				for _, th := range txids {
					v, err := w.dcrd.GetRawTransactionVerbose(ctx, th)
					if err != nil {
						continue
					}
					for voutIdx, vout := range v.Vout {
						match := false
						if spkBytes, err := hex.DecodeString(vout.ScriptPubKey.Hex); err == nil {
							if pkBytes, err := hex.DecodeString(pkHex); err == nil {
								match = bytes.Equal(spkBytes, pkBytes)
							}
						}
						if match {
							atoms := uint64(vout.Value * 1e8)
							utxos = append(utxos, &pong.EscrowUTXO{
								Txid:        v.Txid,
								Vout:        uint32(voutIdx),
								Value:       atoms,
								PkScriptHex: pkHex,
							})
						}
					}
				}
				if len(utxos) > 0 {
					confs = 0
					found = true
					w.log.Debugf("watcher: pk=%s found in mempool; utxos=%d", pkHex, len(utxos))
				}
			} else {
				w.log.Debugf("watcher: GetRawMempool failed; skipping mempool scan for %s", pkHex)
			}
		}

		// Push an update every tick (simple, stateless).
		w.broadcastUpdate(pkHex, DepositUpdate{
			PkScriptHex: pkHex,
			Confs:       confs,
			UTXOCount:   len(utxos),
			OK:          found,
			At:          time.Now(),
			UTXOs:       utxos,
		})
	}
}

func (w *chainWatcher) currentTip() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.tip
}

// Subscribe adds a listener for pkScriptHex and returns the channel + unsubscribe.
// No initial snapshot is sent; first data arrives on next tick.
func (w *chainWatcher) Subscribe(pkScriptHex string) (<-chan DepositUpdate, func()) {
	k := strings.ToLower(pkScriptHex)

	ch := make(chan DepositUpdate, 8)

	w.mu.Lock()
	if _, ok := w.subs[k]; !ok {
		w.subs[k] = make(map[chan DepositUpdate]struct{})
	}
	w.subs[k][ch] = struct{}{}
	n := len(w.subs[k])
	w.mu.Unlock()
	w.log.Infof("watcher: subscribed pk=%s (subs=%d)", k, n)

	unsub := func() {
		w.mu.Lock()
		if set, ok := w.subs[k]; ok {
			delete(set, ch)
			if len(set) == 0 {
				delete(w.subs, k)
			}
		}
		remaining := 0
		if set, ok := w.subs[k]; ok {
			remaining = len(set)
		}
		w.mu.Unlock()
		w.log.Infof("watcher: unsubscribed pk=%s (subs=%d)", k, remaining)
		// Do not close(ch): the producer may still try to send; let receiver stop by context.
	}
	return ch, unsub
}

// broadcastUpdate snapshots subscribers for pk, then best-effort sends (non-blocking).
func (w *chainWatcher) broadcastUpdate(pk string, u DepositUpdate) {
	w.mu.RLock()
	set := w.subs[pk]
	chs := make([]chan DepositUpdate, 0, len(set))
	for ch := range set {
		chs = append(chs, ch)
	}
	w.mu.RUnlock()
	w.log.Debugf("watcher: broadcast pk=%s to %d listeners; ok=%t utxos=%d confs=%d", pk, len(chs), u.OK, u.UTXOCount, u.Confs)

	for _, ch := range chs {
		select {
		case ch <- u:
		default:
			// Drop if receiver is slow.
		}
	}
}
