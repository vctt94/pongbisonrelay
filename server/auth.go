package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"math/rand"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/decred/dcrd/chaincfg/chainhash"
	"github.com/decred/dcrd/chaincfg/v3"
	"github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	"github.com/decred/dcrd/txscript/v4/stdaddr"
	"github.com/decred/dcrd/wire"
)

// nonceTTL defines how long a nonce is valid for login.
const nonceTTL = 3 * time.Minute

type nonceInfo struct {
	expiresAt time.Time
	used      bool
}

type sessionInfo struct {
	address string
	uid     zkidentity.ShortID
	created time.Time
}

// Auth state kept in-memory (short lived).
type authState struct {
	mu        sync.RWMutex
	nonces    map[string]nonceInfo
	sessions  map[string]sessionInfo // token -> session
	addrToUID map[string]zkidentity.ShortID
}

// initAuth initializes the in-memory auth state maps.
func (s *Server) initAuth() {
	s.auth.mu.Lock()
	defer s.auth.mu.Unlock()
	if s.auth.nonces == nil {
		s.auth.nonces = make(map[string]nonceInfo)
	}
	if s.auth.sessions == nil {
		s.auth.sessions = make(map[string]sessionInfo)
	}
	if s.auth.addrToUID == nil {
		s.auth.addrToUID = make(map[string]zkidentity.ShortID)
	}
}

// VerifySignMessage verifies a base64 compact ECDSA signature of msg by addr.
// It recovers the pubkey and compares the derived P2PKH address against the
// provided address. It detects the network from the provided address.
func VerifySignMessage(addrStr, b64sig, msg string) (bool, error) {
	// Trim common clipboard artifacts.
	addrStr = strings.TrimSpace(addrStr)
	b64sig = strings.TrimSpace(b64sig)
	msg = strings.TrimSpace(msg)

	// Decode address in both networks to identify the correct params.
	var params *chaincfg.Params
	if _, err := stdaddr.DecodeAddress(addrStr, chaincfg.MainNetParams()); err == nil {
		params = chaincfg.MainNetParams()
	} else if _, err := stdaddr.DecodeAddress(addrStr, chaincfg.TestNet3Params()); err == nil {
		params = chaincfg.TestNet3Params()
	} else if _, err := stdaddr.DecodeAddress(addrStr, chaincfg.SimNetParams()); err == nil {
		params = chaincfg.SimNetParams()
	} else {
		return false, errors.New("invalid or unsupported address")
	}

	addr, err := stdaddr.DecodeAddress(addrStr, params)
	if err != nil {
		return false, err
	}

	// Build Decred signed message digest.
	var buf bytes.Buffer
	// Write VarString(header) + VarString(message) using wire helpers
	const header = "Decred Signed Message:\n"
	if err := wire.WriteVarString(&buf, 0, header); err != nil {
		return false, err
	}
	if err := wire.WriteVarString(&buf, 0, msg); err != nil {
		return false, err
	}

	// dcrwallet uses chainhash.HashB (double BLAKE-256)
	digest := chainhash.HashB(buf.Bytes())

	sig, err := base64.StdEncoding.DecodeString(b64sig)
	if err != nil {
		return false, fmt.Errorf("base64 decode failed: %w", err)
	}
	pub, _, err := ecdsa.RecoverCompact(sig, digest)
	if err != nil {
		return false, fmt.Errorf("recover compact failed: %w", err)
	}
	got, _ := stdaddr.NewAddressPubKeyHashEcdsaSecp256k1V0(stdaddr.Hash160(pub.SerializeCompressed()), params)
	match := got.String() == addr.String()
	if !match {
		return false, fmt.Errorf("address mismatch: got %s, want %s", got.String(), addr.String())
	}
	return true, nil
}

// StartAuthHTTP starts a small HTTP server that implements /auth/request and
// /auth/verify for Decred sign-message authentication.
func (s *Server) StartAuthHTTP(addr string) error {
	s.initAuth()

	mux := http.NewServeMux()

	mux.HandleFunc("/auth/request", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		// Generate one-time nonce.
		nonce := fmt.Sprintf("login:%d-%08x", time.Now().Unix(), rand.Uint32())
		s.auth.mu.Lock()
		s.auth.nonces[nonce] = nonceInfo{expiresAt: time.Now().Add(nonceTTL)}
		s.auth.mu.Unlock()

		resp := map[string]interface{}{
			"nonce":        nonce,
			"ttl_sec":      int(nonceTTL.Seconds()),
			"address_hint": "",
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/auth/verify", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		var req struct {
			Address   string `json:"address"`
			Nonce     string `json:"nonce"`
			Signature string `json:"signature"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}
		req.Address = strings.TrimSpace(req.Address)
		req.Nonce = strings.TrimSpace(req.Nonce)
		req.Signature = strings.TrimSpace(req.Signature)
		if req.Address == "" || req.Nonce == "" || req.Signature == "" {
			http.Error(w, "missing fields", http.StatusBadRequest)
			return
		}

		// Validate nonce.
		s.auth.mu.Lock()
		ni, ok := s.auth.nonces[req.Nonce]
		if !ok || ni.used || time.Now().After(ni.expiresAt) {
			s.auth.mu.Unlock()
			http.Error(w, "invalid or expired nonce", http.StatusUnauthorized)
			return
		}
		// Tentatively mark used; will stay used regardless of verify result to prevent oracle scans.
		ni.used = true
		s.auth.nonces[req.Nonce] = ni
		s.auth.mu.Unlock()

		// Verify signature.
		s.log.Infof("Verifying signature - Address: %s, Nonce: %s, Signature: %s", req.Address, req.Nonce, req.Signature)
		okSig, err := VerifySignMessage(req.Address, req.Signature, req.Nonce)
		if err != nil || !okSig {
			s.log.Errorf("Signature verification failed - okSig: %v, err: %v", okSig, err)
			http.Error(w, "invalid signature", http.StatusUnauthorized)
			return
		}
		s.log.Infof("Signature verified successfully for address: %s", req.Address)

		// Establish session and client id derived from recovered key/address.
		// Use stable mapping per address when possible.
		s.auth.mu.Lock()
		uid, has := s.auth.addrToUID[req.Address]
		if !has {
			// Derive uid deterministically from address string hash.
			sum := sha256.Sum256([]byte(req.Address))
			var sid zkidentity.ShortID
			sid.FromBytes(sum[:])
			uid = sid
			s.auth.addrToUID[req.Address] = uid
		}
		// Create a random token.
		tok := fmt.Sprintf("sess_%d_%08x", time.Now().Unix(), rand.Uint32())
		s.auth.sessions[tok] = sessionInfo{address: req.Address, uid: uid, created: time.Now()}
		s.auth.mu.Unlock()

		resp := map[string]interface{}{
			"ok":        true,
			"token":     tok,
			"client_id": uid.String(),
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Basic CORS for local testing; adjust as needed.
	handler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Access-Control-Allow-Origin", "*")
		w.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
		w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
		if r.Method == http.MethodOptions {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		mux.ServeHTTP(w, r)
	})

	srv := &http.Server{Addr: addr, Handler: handler}
	s.httpServer = srv

	go func() {
		s.log.Infof("Auth HTTP server listening on http://%s", addr)
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			s.log.Errorf("auth http server error: %v", err)
		}
	}()
	return nil
}

// StopAuthHTTP gracefully shuts down the auth HTTP server.
func (s *Server) StopAuthHTTP(ctx context.Context) error {
	if s.httpServer == nil {
		return nil
	}
	return s.httpServer.Shutdown(ctx)
}
