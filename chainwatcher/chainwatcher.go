package chainwatcher

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
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
	// known stores currently known unspent UTXOs per pkScriptHex, keyed by
	// "txid:vout". This allows persisting deposits discovered in prior ticks
	// so that subsequent ticks can still report funding even if no new blocks
	// include outputs for that pk.
	known map[string]map[string]*pong.EscrowUTXO
}

func NewChainWatcher(log slog.Logger, c *rpcclient.Client) *ChainWatcher {
	return &ChainWatcher{
		log:         log,
		dcrd:        c,
		subs:        make(map[string]map[chan DepositUpdate]struct{}),
		quit:        make(chan struct{}),
		lastScanned: -1,
		pkBytes:     make(map[string][]byte),
		known:       make(map[string]map[string]*pong.EscrowUTXO),
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
		// w.log.Debugf("watcher: poll tick; no subscribers (tip=%d)", w.currentTip())
		return
	}
	keys := make([]string, 0, len(w.subs))
	for k := range w.subs {
		keys = append(keys, k)
	}
	// Snapshot pkBytes and known sizes for decision making without holding locks during RPCs
	pkbByKey := make(map[string][]byte, len(keys))
	knownSize := make(map[string]int, len(keys))
	for _, k := range keys {
		pkbByKey[k] = w.pkBytes[k]
		if km := w.known[k]; km != nil {
			knownSize[k] = len(km)
		} else {
			knownSize[k] = 0
		}
	}
	w.mu.RUnlock()

	tip := w.currentTip()
	// w.log.Debugf("watcher: poll tick; tip=%d, subs=%d", tip, subsCount)

	// Prepare collectors for discoveries this tick (per-pk)
	discoveredByPk := make(map[string][]*pong.EscrowUTXO, len(keys))
	latestMatchByPk := make(map[string]int64, len(keys))

	// ----- New blocks (scan once per tick across all pkScripts) -----
	// Scan when tip changed (forward or backward). On unknown/reorg, just scan current tip.
	shouldScanBlocks := tip >= 0 && (w.lastScanned == -1 || tip != w.lastScanned)
	if shouldScanBlocks {
		start := w.lastScanned + 1
		if w.lastScanned == -1 || start < 0 || start > tip {
			// First run or reorg/unknown -> only scan the current tip
			start = tip
		}
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
					// Compare output script with each subscribed pk (byte-equal)
					for _, pkHex := range keys {
						pkb := pkbByKey[pkHex]
						if pkb != nil && bytes.Equal(o.PkScript, pkb) {
							discoveredByPk[pkHex] = append(discoveredByPk[pkHex], &pong.EscrowUTXO{
								Txid:        mtx.TxHash().String(),
								Vout:        uint32(voutIdx),
								Value:       uint64(o.Value),
								PkScriptHex: pkHex,
							})
							if bh > latestMatchByPk[pkHex] {
								latestMatchByPk[pkHex] = bh
							}
						}
					}
				}
			}
		}
		// Advance lastScanned only once per tick
		w.lastScanned = tip
	}

	// ----- Mempool (0-conf) scan once if needed -----
	needMempool := false
	for _, pkHex := range keys {
		if len(discoveredByPk[pkHex]) == 0 && knownSize[pkHex] == 0 {
			needMempool = true
			break
		}
	}
	if needMempool {
		if txids, err := w.dcrd.GetRawMempool(ctx, "all"); err == nil {
			for _, th := range txids {
				v, err := w.dcrd.GetRawTransactionVerbose(ctx, th)
				if err != nil || v == nil {
					continue
				}
				for voutIdx, vout := range v.Vout {
					spkBytes, err := hex.DecodeString(vout.ScriptPubKey.Hex)
					if err != nil {
						continue
					}
					for _, pkHex := range keys {
						pkb := pkbByKey[pkHex]
						if pkb != nil && bytes.Equal(spkBytes, pkb) {
							atoms := uint64(vout.Value * 1e8)
							discoveredByPk[pkHex] = append(discoveredByPk[pkHex], &pong.EscrowUTXO{
								Txid:        v.Txid,
								Vout:        uint32(voutIdx),
								Value:       atoms,
								PkScriptHex: pkHex,
							})
						}
					}
				}
			}
		} else {
			// w.log.Debugf("watcher: GetRawMempool failed; skipping mempool scan for tick")
		}
	}

	// ----- Persist discoveries and compute current state per pkScript -----
	for _, pkHex := range keys {
		if list := discoveredByPk[pkHex]; len(list) > 0 {
			// Log only when discovered in new blocks
			if h := latestMatchByPk[pkHex]; h > 0 {
				confs := uint32((tip - h) + 1)
				w.log.Debugf("watcher: pk=%s found in new blocks; utxos=%d confs=%d (scanned %d..%d)",
					pkHex, len(list), confs, func() int64 {
						if w.lastScanned == -1 {
							return tip
						} else {
							return w.lastScanned
						}
					}(), tip)
			}
			w.mu.Lock()
			km := w.known[pkHex]
			if km == nil {
				km = make(map[string]*pong.EscrowUTXO)
				w.known[pkHex] = km
			}
			for _, u := range list {
				id := u.Txid + ":" + fmt.Sprintf("%d", u.Vout)
				km[id] = u
			}
			w.mu.Unlock()
		}

		// Build a list from the currently known entries (if any) and check if they remain unspent.
		w.mu.RLock()
		km := w.known[pkHex]
		ids := make([]string, 0, len(km))
		for id := range km {
			ids = append(ids, id)
		}
		w.mu.RUnlock()

		current := make([]*pong.EscrowUTXO, 0, len(ids))
		minConfs := int64(^uint32(0))
		for _, id := range ids {
			u := km[id]
			if u == nil {
				continue
			}
			var h chainhash.Hash
			if err := chainhash.Decode(&h, u.Txid); err != nil {
				continue
			}
			res, err := w.dcrd.GetTxOut(ctx, &h, u.Vout, 0, true)
			if err != nil || res == nil {
				w.mu.Lock()
				if set := w.known[pkHex]; set != nil {
					delete(set, id)
					if len(set) == 0 {
						delete(w.known, pkHex)
					}
				}
				w.mu.Unlock()
				continue
			}
			current = append(current, u)
			if res.Confirmations < minConfs {
				minConfs = res.Confirmations
			}
		}

		var confs uint32
		ok := false
		if len(current) > 0 {
			ok = true
			switch {
			case minConfs < 0:
				confs = 0
			case minConfs > int64(^uint32(0)):
				confs = ^uint32(0)
			default:
				confs = uint32(minConfs)
			}
		} else {
			confs = 0
		}

		w.broadcastUpdate(pkHex, DepositUpdate{
			PkScriptHex: pkHex,
			Confs:       confs,
			UTXOCount:   len(current),
			OK:          ok,
			At:          time.Now(),
			UTXOs:       current,
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
				// If no more subscribers for this pk, clear known cache to free memory.
				delete(w.known, k)
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
	// w.log.Debugf("watcher: broadcast pk=%s to %d listeners; ok=%t utxos=%d confs=%d", pk, len(chs), u.OK, u.UTXOCount, u.Confs)

	for _, ch := range chs {
		select {
		case ch <- u:
		default:
			// Drop if receiver is slow.
		}
	}
}
