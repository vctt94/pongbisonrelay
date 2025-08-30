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

// chainWatcher maintains a lightweight view of the current tip and relevant mempool/blocks
// for a small set of watched pkScripts. It polls dcrd periodically (WebSocket connection)
// to update confirmations and matching UTXOs for those scripts.
type chainWatcher struct {
	log  slog.Logger
	dcrd *rpcclient.Client

	mu       sync.RWMutex
	tip      int64
	watchSet map[string]*watchedEntry // key: pkScript hex

	quit chan struct{}
}

type watchedEntry struct {
	pkScriptHex     string
	redeemScriptHex string
	tag             string
	utxos           []*pong.EscrowUTXO
	confs           uint32
	lastUpdated     time.Time
}

func newChainWatcher(log slog.Logger, c *rpcclient.Client) *chainWatcher {
	return &chainWatcher{
		log:      log,
		dcrd:     c,
		watchSet: make(map[string]*watchedEntry),
		quit:     make(chan struct{}),
	}
}

func (w *chainWatcher) stop() { close(w.quit) }

// registerDeposit registers a pkScript to be monitored.
func (w *chainWatcher) registerDeposit(pkScriptHex, redeemScriptHex, tag string) {
	w.mu.Lock()
	defer w.mu.Unlock()
	k := strings.ToLower(pkScriptHex)
	r := strings.ToLower(redeemScriptHex)
	if _, ok := w.watchSet[k]; !ok {
		w.watchSet[k] = &watchedEntry{pkScriptHex: k, redeemScriptHex: r, tag: tag}
		w.log.Infof("watcher: registered tag=%s pkScript=%s", tag, k)
	}
}

// queryDeposit returns the last known utxos and confirmations for a pkScript.
func (w *chainWatcher) queryDeposit(pkScriptHex string) (utxos []*pong.EscrowUTXO, confs uint32, ok bool) {
	w.mu.RLock()
	defer w.mu.RUnlock()
	k := strings.ToLower(pkScriptHex)
	if e, exists := w.watchSet[k]; exists {
		return e.utxos, e.confs, true
	}
	return nil, 0, false
}

func (w *chainWatcher) run(ctx context.Context) {
	t := time.NewTicker(5 * time.Second)
	defer t.Stop()
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
	// Update tip height
	if _, h, err := w.dcrd.GetBestBlock(ctx); err == nil {
		w.mu.Lock()
		w.tip = h
		w.mu.Unlock()
	} else {
		w.log.Debugf("watcher: GetBestBlock failed: %v", err)
	}

	// No watched entries; nothing to do
	w.mu.RLock()
	if len(w.watchSet) == 0 {
		w.mu.RUnlock()
		return
	}
	// Snapshot keys to avoid holding lock during RPC calls
	keys := make([]string, 0, len(w.watchSet))
	for k := range w.watchSet {
		keys = append(keys, k)
	}
	w.mu.RUnlock()

	// For each watched script, try to locate UTXOs in recent blocks; if not found, check mempool
	for _, pkHex := range keys {
		var utxos []*pong.EscrowUTXO
		var confs uint32
		found := false

		// Scan recent blocks (bounded)
		if h := w.currentTip(); h >= 0 {
			const lookback = int64(1024)
			latestMatch := int64(-1)
			var bestHeight int64 = h
			candidates := make([]*pong.EscrowUTXO, 0, 2)
			for i := int64(0); i < lookback && bestHeight-i >= 0; i++ {
				bh := bestHeight - i
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
						var match bool
						if spkBytes, err := hex.DecodeString(vout.ScriptPubKey.Hex); err == nil {
							if pkBytes, err := hex.DecodeString(pkHex); err == nil {
								match = bytes.Equal(spkBytes, pkBytes)
							}
						}
						if match {
							atoms := uint64(vout.Value * 1e8)
							candidates = append(candidates, &pong.EscrowUTXO{
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
			if len(candidates) > 0 && latestMatch >= 0 {
				utxos = candidates
				confs = uint32((bestHeight - latestMatch) + 1)
				found = true
			}
		}

		// If not in recent blocks, try mempool (0 conf)
		if !found {
			if txids, err := w.dcrd.GetRawMempool(ctx, "all"); err == nil {
				mCandidates := make([]*pong.EscrowUTXO, 0, 2)
				for _, th := range txids {
					v, err := w.dcrd.GetRawTransactionVerbose(ctx, th)
					if err != nil {
						continue
					}
					for voutIdx, vout := range v.Vout {
						var match bool
						if spkBytes, err := hex.DecodeString(vout.ScriptPubKey.Hex); err == nil {
							if pkBytes, err := hex.DecodeString(pkHex); err == nil {
								match = bytes.Equal(spkBytes, pkBytes)
							}
						}
						if match {
							atoms := uint64(vout.Value * 1e8)
							mCandidates = append(mCandidates, &pong.EscrowUTXO{
								Txid:        v.Txid,
								Vout:        uint32(voutIdx),
								Value:       atoms,
								PkScriptHex: pkHex,
							})
						}
					}
				}
				if len(mCandidates) > 0 {
					utxos = mCandidates
					confs = 0
					found = true
				}
			} else {
				w.log.Debugf("watcher: GetRawMempool failed; skipping mempool scan for %s", pkHex)
			}
		}

		// Store results
		w.mu.Lock()
		if e, exists := w.watchSet[pkHex]; exists {
			if found {
				// Preserve redeemScriptHex if set
				for _, u := range utxos {
					if u.RedeemScriptHex == "" {
						u.RedeemScriptHex = e.redeemScriptHex
					}
				}
				e.utxos = utxos
				e.confs = confs
				e.lastUpdated = time.Now()
				w.log.Debugf("watcher: updated tag=%s pkScript=%s confs=%d utxos=%d", e.tag, pkHex, confs, len(utxos))
			} else {
				// No info found this round; keep previous values
				if time.Since(e.lastUpdated) > 30*time.Second {
					w.log.Debugf("watcher: no matches yet tag=%s pkScript=%s (lastUpdated=%s)", e.tag, pkHex, e.lastUpdated.Format(time.RFC3339))
				}
			}
		}
		w.mu.Unlock()
	}
}

func (w *chainWatcher) currentTip() int64 {
	w.mu.RLock()
	defer w.mu.RUnlock()
	return w.tip
}
