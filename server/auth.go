package server

import (
	"bytes"
	"context"
	"crypto/sha256"
	"encoding/base64"
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
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

// nonceTTL defines how long a nonce is valid for login.
const nonceTTL = 3 * time.Minute

type nonceInfo struct {
	expiresAt time.Time
	used      bool
}

type sessionInfo struct {
	address    string
	uid        zkidentity.ShortID
	created    time.Time
	compPubkey []byte
	p2pkAddr   string
}

// Auth state kept in-memory (short lived).
type authState struct {
	mu        sync.RWMutex
	nonces    map[string]nonceInfo
	sessions  map[string]sessionInfo // token -> session
	addrToUID map[string]zkidentity.ShortID
	// Fast lookup of authenticated wallet keys per uid.
	uidToPubkey map[zkidentity.ShortID][]byte // 33B compressed
	uidToP2PK   map[zkidentity.ShortID]string // P2PK address string
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
	if s.auth.uidToPubkey == nil {
		s.auth.uidToPubkey = make(map[zkidentity.ShortID][]byte)
	}
	if s.auth.uidToP2PK == nil {
		s.auth.uidToP2PK = make(map[zkidentity.ShortID]string)
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

// RequestNonce implements gRPC nonce issuance for wallet auth.
func (s *Server) RequestNonce(ctx context.Context, _ *pong.RequestNonceRequest) (*pong.RequestNonceResponse, error) {
	s.initAuth()
	// Generate one-time nonce with TTL.
	nonce := fmt.Sprintf("login:%d-%08x", time.Now().Unix(), rand.Uint32())
	s.auth.mu.Lock()
	s.auth.nonces[nonce] = nonceInfo{expiresAt: time.Now().Add(nonceTTL)}
	s.auth.mu.Unlock()
	return &pong.RequestNonceResponse{Nonce: nonce, TtlSec: int32(nonceTTL.Seconds()), AddressHint: ""}, nil
}

// VerifyLogin verifies the signed nonce and establishes a session, also
// returning the recovered 33-byte compressed pubkey and its P2PK address.
func (s *Server) VerifyLogin(ctx context.Context, req *pong.VerifyLoginRequest) (*pong.VerifyLoginResponse, error) {
	if req == nil || strings.TrimSpace(req.Address) == "" || strings.TrimSpace(req.Nonce) == "" || strings.TrimSpace(req.Signature) == "" {
		return nil, statusError(http.StatusBadRequest, "missing fields")
	}

	addrStr := strings.TrimSpace(req.Address)
	b64sig := strings.TrimSpace(req.Signature)
	msg := strings.TrimSpace(req.Nonce)

	// Validate nonce.
	s.auth.mu.Lock()
	ni, ok := s.auth.nonces[msg]
	if !ok || ni.used || time.Now().After(ni.expiresAt) {
		s.auth.mu.Unlock()
		return nil, statusError(http.StatusUnauthorized, "invalid or expired nonce")
	}
	// Mark used to prevent oracle scans, regardless of result.
	ni.used = true
	s.auth.nonces[msg] = ni
	s.auth.mu.Unlock()

	// Detect network from address string.
	var params *chaincfg.Params
	if _, err := stdaddr.DecodeAddress(addrStr, chaincfg.MainNetParams()); err == nil {
		params = chaincfg.MainNetParams()
	} else if _, err := stdaddr.DecodeAddress(addrStr, chaincfg.TestNet3Params()); err == nil {
		params = chaincfg.TestNet3Params()
	} else if _, err := stdaddr.DecodeAddress(addrStr, chaincfg.SimNetParams()); err == nil {
		params = chaincfg.SimNetParams()
	} else {
		return nil, statusError(http.StatusUnauthorized, "invalid or unsupported address")
	}

	addr, err := stdaddr.DecodeAddress(addrStr, params)
	if err != nil {
		return nil, statusError(http.StatusUnauthorized, fmt.Sprintf("decode address: %v", err))
	}

	// Build Decred signed message digest.
	var buf bytes.Buffer
	const header = "Decred Signed Message:\n"
	if err := wire.WriteVarString(&buf, 0, header); err != nil {
		return nil, statusError(http.StatusUnauthorized, fmt.Sprintf("header write failed: %v", err))
	}
	if err := wire.WriteVarString(&buf, 0, msg); err != nil {
		return nil, statusError(http.StatusUnauthorized, fmt.Sprintf("msg write failed: %v", err))
	}
	digest := chainhash.HashB(buf.Bytes())

	// Recover pubkey and verify it maps to the provided P2PKH.
	sig, err := base64.StdEncoding.DecodeString(b64sig)
	if err != nil {
		return nil, statusError(http.StatusUnauthorized, fmt.Sprintf("base64 decode failed: %v", err))
	}
	pub, _, err := ecdsa.RecoverCompact(sig, digest)
	if err != nil {
		return nil, statusError(http.StatusUnauthorized, fmt.Sprintf("recover compact failed: %v", err))
	}
	got, _ := stdaddr.NewAddressPubKeyHashEcdsaSecp256k1V0(stdaddr.Hash160(pub.SerializeCompressed()), params)
	if got.String() != addr.String() {
		return nil, statusError(http.StatusUnauthorized, fmt.Sprintf("address mismatch: got %s, want %s", got.String(), addr.String()))
	}

	// Also compute P2PK address for payout convenience.
	p2pkAddr, err := stdaddr.NewAddressPubKeyEcdsaSecp256k1V0(pub, params)
	if err != nil {
		return nil, statusError(http.StatusUnauthorized, fmt.Sprintf("p2pk encode failed: %v", err))
	}

	// Establish session and stable client id derived from address.
	s.auth.mu.Lock()
	uid, has := s.auth.addrToUID[addrStr]
	if !has {
		sum := sha256.Sum256([]byte(addrStr))
		var sid zkidentity.ShortID
		sid.FromBytes(sum[:])
		uid = sid
		s.auth.addrToUID[addrStr] = uid
	}
	// Create token and store session with recovered pubkey + p2pk address.
	tok := fmt.Sprintf("sess_%d_%08x", time.Now().Unix(), rand.Uint32())
	s.auth.sessions[tok] = sessionInfo{
		address:    addrStr,
		uid:        uid,
		created:    time.Now(),
		compPubkey: pub.SerializeCompressed(),
		p2pkAddr:   p2pkAddr.String(),
	}
	s.auth.mu.Unlock()

	// Persist uid -> wallet key mappings for payout defaults.
	// Prevent overwriting existing payout pubkey to avoid changing escrow payouts.
	s.auth.mu.Lock()
	if existing, exists := s.auth.uidToPubkey[uid]; exists {
		// Verify the pubkey matches - same uid should have same pubkey
		if !bytes.Equal(existing, pub.SerializeCompressed()) {
			uidStr := uid.String()
			s.auth.mu.Unlock()
			return nil, statusError(http.StatusConflict, "uid "+uidStr+" already authenticated with different pubkey")
		}
		// Same pubkey, keep existing entry (no overwrite needed)
	} else {
		// First time authentication for this uid - store it
		s.auth.uidToPubkey[uid] = append([]byte(nil), pub.SerializeCompressed()...)
		s.auth.uidToP2PK[uid] = p2pkAddr.String()
	}
	s.auth.mu.Unlock()

	return &pong.VerifyLoginResponse{
		Ok:         true,
		Token:      tok,
		ClientId:   uid.String(),
		CompPubkey: pub.SerializeCompressed(),
		P2PkAddr:   p2pkAddr.String(),
	}, nil
}

// statusError converts an auth error to a generic error; in a full gRPC setup
// you'd use status codes, but we avoid a hard dependency here.
func statusError(_ int, msg string) error { return errors.New(msg) }
