package server

// PreSignCtx stores all artifacts needed to finalize using the exact same
// draft and message digest that were used during the presign phase.
//
// It binds the presign to:
// - the specific input (txid:vout)
// - the redeem script and its version (implicitly v0 here)
// - the exact serialized draft transaction
// - the sighash digest used to compute s'
// - the adaptor point used during presign
//
// The winner can then finalize with s = s' + gamma (mod n) using the same m.
type PreSignCtx struct {
	InputID          string // "txid:vout"
	RedeemScriptHex  string
	DraftHex         string // exact serialized tx used at presign
	MHex             string // 32-byte sighash for (DraftHex, RedeemScriptHex, idx, SIGHASH_ALL)
	RPrimeCompressed []byte // 33 bytes, even-Y (0x02)
	SPrime32         []byte // 32 bytes
	TCompressed      []byte // 33 bytes (if used in adaptor domain)
	WinnerUID        string // tie to player/session (owner uid)
}
