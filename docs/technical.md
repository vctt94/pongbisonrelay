
# Pong Settlement — Decred Schnorr (DCRv0) Adaptor Flow

This document explains how the **self-custodial Pong** settlement works **exactly** as implemented in the current client code, using **Decred Schnorr (DCRv0)** and adaptor signatures. It maps concepts to concrete functions and spells out what each side must verify.

---

## 0) Notation & Primitives

- Curve: secp256k1 with generator **G**, order **n**.
- Player key: secret **x**, public **X = x·G**.
- Referee adaptor secret: **γ**, adaptor point **T = γ·G** (33-byte compressed).
- Nonce: random scalar **k** from `crypto/rand` (reject 0 / overflow).
- Points:
  - **R = k·G**
  - **R' = R + T = k·G + T** (**must compress with even-Y**, prefix `0x02`).
- Coordinate:
  - **r' = x(R')** = 32-byte X-coordinate of **R'**.
- Challenge: **e = BLAKE256(r' || m)**, where **m** is the 32-byte **Decred sighash** of the specific input and branch draft.
- DCRv0 signing relation: **s' = k − e·x (mod n)** (pre-signature), final **s = s' + γ (mod n)** after reveal.
- Verifier reconstruction for finalized signature: **R' := s·G + e·X**, then require `R'.y` even and `x(R') == r'`.

---

## 1) Funding Phase (Escrow UTXOs)

1. **Server** assigns roles (A/B) and sends each client:
   - `required_atoms` (stake amount),
   - `deposit_pkScript_hex` (and optional address),
   - `min_confs`.
2. **Client** rebuilds the redeem script **locally** and derives the **pkScript**. Only fund if the bytes **exactly match** `deposit_pkScript_hex`.
3. **Client** sends exactly `required_atoms` from its own wallet to that script.
4. **Server** watches the chain by `pkScript` and value; when both sides are confirmed, it proceeds.

**Why the server can’t steal:** the deposit script contains only player keys/timelock paths—**no server key**—and you fund only after **byte-for-byte** script verification.

---

## 2) Server-Driven Settlement Stream

After both escrows are funded:

- Server computes two **branch drafts** (A-wins, B-wins) and, per input, produces `(m, T)`.
- Server sends **NeedPreSigs** message: `{ branches_to_presign: [A, B], drafts, inputs: [ input_id, redeem_script_hex, m_awins, T_awins, m_bwins, T_bwins ] }`.
- Client responds with **PreSigBatch** per branch: a list of `(input_id, r' (32B), s' (32B))`.

> Note: The server receives only `r'` and `s'`. It never sees the nonce `k` nor the full point `R'`.

Your entry point:
- `startSettlementStream()` opens the stream, handles role/funding updates, and reacts to `NeedPreSigs` by calling the **builder** that recomputes `m` and produces pre-sigs.

---

## 3) Computing a Pre-Signature (client)

Function: **`computePreSig(xPrivHex, mHex, TCompHex)`**

**Inputs:**
- `x` — secret scalar (32B),
- `m` — 32B Decred sighash for **this** input/branch,
- `T` — adaptor point (33B compressed).

**Steps:**

1. **Parse inputs & scalars**: reject zero/overflow.
2. **Sample nonce** `k` from `crypto/rand` until valid (non-zero, < n).
3. Compute **R = k·G**; then **R' = R + T** via `addPoints`.
4. **Parity rule:** if compressed `R'` does **not** start with `0x02` (even-Y), **resample** a fresh `k` and repeat.
5. Set `r' = x(R')` (the 32-byte X-coordinate).
6. Compute **e = BLAKE256(r' || m)**.
7. Compute **s' = k − e·x (mod n)** using `NegateVal` + `Add` on `ModNScalar`.
8. Return `(r', s')`. (You also return `R'` compressed for logging/debug.)

**Even-Y requirement (DCRv0):** Nonces/signatures are normalized so the reconstructed point has **even Y**; adaptor mode must **resample** `k` rather than flip `k→n−k`.

---

## 4) Recomputing `m` and Matching Inputs

- Use `findInputIndex(tx, inputID)` to map `"<txid>:<vout>"` to the index within the provided draft tx.
- Recompute **`m = CalcSignatureHash(redeemScript, SigHashAll, draftTx, idx, nil)`**.
- Compare your local `m` (hex) with the server’s `m_awins/m_bwins`. If mismatch → **abort** (do not presign).

This binds your presig to:
- the exact input,
- the redeem script (and version),
- all outputs & fees of that branch draft.

---

## 5) Finalization (after γ reveal)

When the server reveals **γ** for the winning branch:

1. Winner computes **`s = s' + γ (mod n)`**.
2. Attach `(r', s)` to the input’s witness as the DCRv0 Schnorr signature.
3. Broadcast the finalized transaction.

XXX
Current POC status: the server logs received pre‑signatures and reveals `γ`; the client should perform the `s = s' + γ` finalization and broadcast upon `RevealGamma`, but that wiring is not yet implemented.

**Verifier check (conceptual):**
```
R' := s·G + e·X
require evenY(R')
require x(R') == r'
```
This holds because:
```
s·G + e·X = (s' + γ)·G + e·X
          = (k − e·x)·G + γ·G + e·X
          = k·G − e·X + T + e·X
          = k·G + T
          = R'
```

---

## 6) Security Properties

1. **Server cannot redirect funds**: any output change alters `m`; your `(r', s')` becomes useless.
2. **Server cannot spend deposits**: no server key in the script; you verified `pkScript` before funding.
3. **Transparency**: seeing both `s'` and `s` yields **γ = s − s' (mod n)**, proving what the server revealed.
4. **Refund**: if the server disappears or withholds `γ`, you recover via the **CSV** refund path after the timelock.

---

## 7) Code Map (what does what)

- `findInputIndex`: resolve `input_id -> txin index` by prevout (txid:vout).
- `addPoints`: computes `R' = R + T` using Jacobian add; rejects infinity; converts to affine.
- `computePreSig`: implements the DCRv0 adaptor pre-sig: resample `k` → even-Y `R'` → `e = BLAKE256(r'||m)` → `s' = k − e·x`.
- `startSettlementStream`:
  - obtains/ensures `A_c`,
  - allocates escrow if needed,
  - runs funding watcher,
  - handles `NeedPreSigs` → builds and sends `PreSigBatch` for each requested branch,
  - reacts to `RevealGamma` (winner can finalize).

- `server/handlers.go` → `SettlementStream`: stores only `(r', s')` per input/branch; server never receives `k` or full `R'`.
- `server/handlers.go` → `FinalizeWinner`: logs presigs and indicates `s_final = s' + γ`; finalization is intended to occur on the client.

---

## 8) Invariants & Edge Cases

- Reject `k = 0` or overflow; reject **infinity** from `R + T`.
- Require `len(compress(R')) == 33` and prefix `0x02` (even-Y).
- Always recompute `m` locally; never trust only the server’s `m`.
- Match inputs by **prevout**, never by list index.
- If drafts include a different script version, pass it explicitly to `CalcSignatureHash`.
- If there’s a reorg, the server’s funding watcher restates status; client logic is unchanged.

---

## 9) Testing Checklist

- [ ] Deposit pkScript verification before funding (byte-exact).
- [ ] Funding watcher updates when a single party funds; proceeds only when both funded with min confs.
- [ ] For each input/branch:
  - [ ] Local `m` equals server’s `m_awins/m_bwins`.
  - [ ] `R'` compresses with even-Y.
  - [ ] `(r', s')` verifies after finalization (`s = s' + γ`) using the DCRv0 relation.
- [ ] Server cannot broadcast a tx that pays a different address (presigs won’t verify).
- [ ] Refund path works after CSV (simulate withheld `γ`).

---

## 10) Appendix — Script Views (OP codes)

This section lists the exact scripts used in the current implementation.

### Deposit redeem script (P2SH redeem, script version 0)
Spends are by the depositor only, with an immediate path and a CSV-delayed path.

```
OP_IF
  <A_c (33B)> 2 OP_CHECKSIGALTVERIFY
  OP_TRUE
OP_ELSE
  <csv_blocks> OP_CHECKSEQUENCEVERIFY OP_DROP
  <A_c (33B)> 2 OP_CHECKSIGALTVERIFY
  OP_TRUE
OP_ENDIF
```

Where:
- `<A_c>`: depositor’s compressed secp256k1 pubkey (33 bytes).
- `2`: Schnorr-secp256k1 signature type for OP_CHECKSIGALT.
- `csv_blocks`: relative timelock enforced via OP_CHECKSEQUENCEVERIFY.

### Deposit P2SH pkScript (script version 0)
```
OP_HASH160 <Hash160(redeem)> OP_EQUAL
```

### Unlocking scripts for the deposit
- Immediate path (selects OP_IF branch):
```
<sig65> OP_1 <redeem>
```
- CSV refund path (selects OP_ELSE branch):
```
<sig65> OP_0 <redeem>
```

Notes:
- `sig65 = r || s || 0x01`, where `0x01` is SigHashAll.
- CSV branch requires the input’s sequence to satisfy `csv_blocks` per Decred relative timelock rules.

### Winner payout output script (script version 0)
Payout is P2PK-alt to the winner’s compressed pubkey.
```
<X_winner (33B)> 2 OP_CHECKSIGALT
```

---

## 11) Appendix — Equations (DCRv0)

- Challenge: `e = BLAKE256(r || m)`
- Pre-sig:   `s' = k − e·x (mod n)`
- Reveal:    `T = γ·G`
- Final:     `s = s' + γ (mod n)`
- Verify:    `R' = s·G + e·X`, `evenY(R')`, `x(R') == r'`
