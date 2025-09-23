package chainwatcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"strings"
	"sync"
	"time"

	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/rpcclient/v8"
	"github.com/decred/slog"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
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
type ChainWatcher struct {
	log  slog.Logger
	dcrd *rpcclient.Client

	mu   sync.RWMutex
	tip  int64
	subs map[string]map[chan DepositUpdate]struct{} // pkScriptHex -> set(chan)

	quit        chan struct{}
	lastScanned int64

	pkBytes map[string][]byte
}

func NewChainWatcher(log slog.Logger, c *rpcclient.Client) *ChainWatcher {
	return &ChainWatcher{
		log:         log,
		dcrd:        c,
		subs:        make(map[string]map[chan DepositUpdate]struct{}),
		quit:        make(chan struct{}),
		lastScanned: -1,
		pkBytes:     make(map[string][]byte),
	}
}

func (w *ChainWatcher) Stop() { close(w.quit) }

func (w *ChainWatcher) Run(ctx context.Context) {
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

func (w *ChainWatcher) pollOnce(ctx context.Context) {
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
	tip := w.currentTip()
	w.log.Debugf("watcher: poll tick; tip=%d, subs=%d", tip, subsCount)

	// Scan each subscribed pkScript.
	for _, pkHex := range keys {
		w.log.Debugf("watcher: scanning pk=%s", pkHex)

		w.mu.RLock()
		pkb := w.pkBytes[pkHex]
		w.mu.RUnlock()

		var utxos []*pong.EscrowUTXO
		var confs uint32
		found := false

		// ----- New blocks only -----
		if tip >= 0 {
			start := w.lastScanned + 1
			// First run or reorg/unknown -> just scan the current tip only (no big backfill).
			if w.lastScanned == -1 || start < 0 || start > tip {
				start = tip
			}

			latestMatch := int64(-1)
			for bh := start; bh <= tip; bh++ {
				hash, err := w.dcrd.GetBlockHash(ctx, bh)
				if err != nil {
					continue
				}
				msg, err := w.dcrd.GetBlock(ctx, hash)
				if err != nil || msg == nil {
					continue
				}
				for _, mtx := range msg.Transactions {
					for voutIdx, o := range mtx.TxOut {
						if pkb != nil && bytes.Equal(o.PkScript, pkb) {
							utxos = append(utxos, &pong.EscrowUTXO{
								Txid:        mtx.TxHash().String(),
								Vout:        uint32(voutIdx),
								Value:       uint64(o.Value), // atoms already
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
				confs = uint32((tip - latestMatch) + 1)
				found = true
				w.log.Debugf("watcher: pk=%s found in new blocks; utxos=%d confs=%d (scanned %d..%d)",
					pkHex, len(utxos), confs, start, tip)
			}

			// Advance lastScanned to the tip processed this tick.
			w.lastScanned = tip
		}

		// ----- Mempool (0-conf) as fallback -----
		if !found {
			if txids, err := w.dcrd.GetRawMempool(ctx, "all"); err == nil {
				for _, th := range txids {
					v, err := w.dcrd.GetRawTransactionVerbose(ctx, th)
					if err != nil || v == nil {
						continue
					}
					for voutIdx, vout := range v.Vout {
						// Compare mempool vout script (hex) to cached pkb.
						if pkb == nil {
							continue
						}
						spkBytes, err := hex.DecodeString(vout.ScriptPubKey.Hex)
						if err != nil {
							continue
						}
						if bytes.Equal(spkBytes, pkb) {
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

		// ----- Filter to unspent & compute min confirmations -----
		if len(utxos) > 0 {
			filtered := make([]*pong.EscrowUTXO, 0, len(utxos))
			minConfs := int64(^uint32(0)) // large; GetTxOut.Confirmations is int64
			for _, u := range utxos {
				var h chainhash.Hash
				if err := chainhash.Decode(&h, u.Txid); err != nil {
					continue
				}
				res, err := w.dcrd.GetTxOut(ctx, &h, u.Vout, 0, true)
				if err != nil || res == nil {
					// Spent or not found.
					continue
				}
				filtered = append(filtered, u)
				if res.Confirmations < minConfs {
					minConfs = res.Confirmations
				}
			}
			utxos = filtered
			if len(utxos) > 0 {
				found = true
				switch {
				case minConfs < 0:
					confs = 0
				case minConfs > int64(^uint32(0)):
					confs = ^uint32(0)
				default:
					confs = uint32(minConfs)
				}
				w.log.Debugf("watcher: pk=%s unspent utxos=%d confs(min)=%d", pkHex, len(utxos), confs)
			} else {
				found = false
				confs = 0
			}
		}

		// ----- Broadcast snapshot -----
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

func (w *ChainWatcher) currentTip() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.tip
}

// Subscribe adds a listener for pkScriptHex and returns the channel + unsubscribe.
// No initial snapshot is sent; first data arrives on next tick.
func (w *ChainWatcher) Subscribe(pkScriptHex string) (<-chan DepositUpdate, func()) {
	k := strings.ToLower(pkScriptHex)
	// cache decoded bytes (ignore error if bad hex; just don’t store)
	if b, err := hex.DecodeString(k); err == nil {
		w.mu.Lock()
		w.pkBytes[k] = b
		w.mu.Unlock()
	}

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
func (w *ChainWatcher) broadcastUpdate(pk string, u DepositUpdate) {
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
