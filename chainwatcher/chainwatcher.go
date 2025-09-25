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

type ChainWatcher struct {
	log  slog.Logger
	dcrd *rpcclient.Client

	mu      sync.RWMutex
	subs    map[string]map[chan DepositUpdate]struct{} // pkHex -> set(chan)
	pkBytes map[string][]byte                          // pkHex -> script bytes
	known   map[string]map[string]*pong.EscrowUTXO     // pkHex -> (txid:vout -> utxo)
}

func NewChainWatcher(log slog.Logger, c *rpcclient.Client) *ChainWatcher {
	return &ChainWatcher{
		log:     log,
		dcrd:    c,
		subs:    make(map[string]map[chan DepositUpdate]struct{}),
		pkBytes: make(map[string][]byte),
		known:   make(map[string]map[string]*pong.EscrowUTXO),
	}
}

func (w *ChainWatcher) Stop() {} // nothing to stop

// Subscribe unchanged from your version (caches pkBytes, returns ch + unsub).
func (w *ChainWatcher) Subscribe(pkScriptHex string) (<-chan DepositUpdate, func()) {
	k := strings.ToLower(pkScriptHex)
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
				delete(w.known, k)
				delete(w.pkBytes, k)
			}
		}
		rem := 0
		if set, ok := w.subs[k]; ok {
			rem = len(set)
		}
		w.mu.Unlock()
		w.log.Infof("watcher: unsubscribed pk=%s (subs=%d)", k, rem)
	}
	return ch, unsub
}

func (w *ChainWatcher) broadcastUpdate(pk string, u DepositUpdate) {
	w.mu.RLock()
	set := w.subs[pk]
	chs := make([]chan DepositUpdate, 0, len(set))
	for ch := range set {
		chs = append(chs, ch)
	}
	w.mu.RUnlock()

	for _, ch := range chs {
		select {
		case ch <- u:
		default: /* drop if slow */
		}
	}
}

// ----- Event handlers you call from rpcclient notifications -------------------

// Use this for OnTxAccepted (hash-only).
func (w *ChainWatcher) ProcessTxAcceptedHash(ctx context.Context, hash *chainhash.Hash) {
	if hash == nil {
		return
	}

	v, err := w.dcrd.GetRawTransactionVerbose(ctx, hash)
	if err != nil || v == nil {
		return
	}

	// snapshot subs + pkBytes
	w.mu.RLock()
	if len(w.subs) == 0 {
		w.mu.RUnlock()
		return
	}
	keys := make([]string, 0, len(w.subs))
	for k := range w.subs {
		keys = append(keys, k)
	}
	pkb := make(map[string][]byte, len(keys))
	for _, k := range keys {
		pkb[k] = w.pkBytes[k]
	}
	w.mu.RUnlock()

	discoveredByPk := make(map[string][]*pong.EscrowUTXO)
	for i, out := range v.Vout {
		spk, err := hex.DecodeString(out.ScriptPubKey.Hex)
		if err != nil {
			continue
		}
		for _, k := range keys {
			if b := pkb[k]; b != nil && bytes.Equal(spk, b) {
				// value comes as float; convert to atoms conservatively
				atoms := uint64(out.Value*1e8 + 0.5)
				discoveredByPk[k] = append(discoveredByPk[k], &pong.EscrowUTXO{
					Txid: v.Txid, Vout: uint32(i), Value: atoms, PkScriptHex: k,
				})
			}
		}
	}
	if len(discoveredByPk) == 0 {
		return
	}

	now := time.Now()
	for k, list := range discoveredByPk {
		w.mu.Lock()
		m := w.known[k]
		if m == nil {
			m = make(map[string]*pong.EscrowUTXO)
			w.known[k] = m
		}
		for _, u := range list {
			m[fmt.Sprintf("%s:%d", u.Txid, u.Vout)] = u
		}
		cur := make([]*pong.EscrowUTXO, 0, len(m))
		for _, u := range m {
			cur = append(cur, u)
		}
		w.mu.Unlock()

		w.broadcastUpdate(k, DepositUpdate{
			PkScriptHex: k, Confs: 0, UTXOCount: len(cur),
			OK: len(cur) > 0, At: now, UTXOs: cur,
		})
	}
}

// Call this for every OnBlockConnected to bump confirmations & prune spents.
func (w *ChainWatcher) ProcessBlockConnected(ctx context.Context) {
	// copy keys to check
	w.mu.RLock()
	keys := make([]string, 0, len(w.known))
	for k := range w.known {
		keys = append(keys, k)
	}
	w.mu.RUnlock()

	now := time.Now()
	for _, k := range keys {
		// copy ids so we can RPC without holding the lock
		w.mu.RLock()
		km := w.known[k]
		ids := make([]string, 0, len(km))
		for id := range km {
			ids = append(ids, id)
		}
		w.mu.RUnlock()

		cur := make([]*pong.EscrowUTXO, 0, len(ids))
		minConfs := int64(^uint32(0))

		for _, id := range ids {
			w.mu.RLock()
			u := w.known[k][id]
			w.mu.RUnlock()
			if u == nil {
				continue
			}

			var h chainhash.Hash
			if err := chainhash.Decode(&h, u.Txid); err != nil {
				continue
			}

			res, err := w.dcrd.GetTxOut(ctx, &h, u.Vout, 0, true)
			if err != nil || res == nil {
				// spent/unknown → evict
				w.mu.Lock()
				if set := w.known[k]; set != nil {
					delete(set, id)
					if len(set) == 0 {
						delete(w.known, k)
					}
				}
				w.mu.Unlock()
				continue
			}

			cur = append(cur, u)
			if res.Confirmations < minConfs {
				minConfs = res.Confirmations
			}
		}

		var confs uint32
		if len(cur) == 0 || minConfs < 0 {
			confs = 0
		} else if minConfs > int64(^uint32(0)) {
			confs = ^uint32(0)
		} else {
			confs = uint32(minConfs)
		}

		w.broadcastUpdate(k, DepositUpdate{
			PkScriptHex: k, Confs: confs, UTXOCount: len(cur),
			OK: len(cur) > 0, At: now, UTXOs: cur,
		})
	}
}
