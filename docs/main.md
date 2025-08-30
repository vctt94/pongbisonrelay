
# Pong Bets: Self-Custodial Schnorr Settlement (Decred Schnorr, adaptor-sig POC)

This document explains how the Pong betting system works in a self-custodial way using **Decred’s EC‑Schnorr‑DCRv0** signatures with **adaptor signatures** and **script paths**. Players hold their own keys and funds; the server coordinates the game, escrow address derivation, and settlement hints **without taking custody**.

---

## What “self‑custodial” means here

- **Players keep private keys**: Keys are generated and stored client‑side. The server never sees private keys and cannot spend player funds.
- **Deposits are on‑chain to a P2SH redeem script**: Each player funds an address that encodes a redeem script with three paths: A‑wins, B‑wins, or a CSV timeout fallback for the depositor.
- **Settlement uses adaptor signatures**: Before the game ends, players produce pre‑signatures bound to the eventual winner branch. After the game, the server reveals a small secret that allows only the winning branch to finalize valid Schnorr signatures and sweep both deposits.
- **Timeout recovery**: If settlement does not complete, the depositor can unilaterally recover their own escrow after a relative CSV delay.

---

## Actors and components

- **Player A / Player B**: End users who hold keys and funds, produce pre‑signatures, and (if winner) finalize and broadcast the settlement transaction.
- **Referee server**: Coordinates matches, derives deposit scripts/addresses, serves adaptors, collects pre‑sigs, reveals winning adaptors, and can assemble drafts. It cannot spend player funds.
- **dcrd**: Full node used by the server for chain data, UTXO discovery, and confirmations.
- **Chain watcher**: A lightweight poller that watches specific `pkScript` values to report observed UTXOs and confirmations for funding status.

---


### Messages
- `m` is **exactly** the 32‑byte Decred input sighash. **Do not prehash again.**
- Schnorr challenge: `e = BLAKE256(r || m)`, where `r = x(R_final)`.

### Adaptor signatures
- For each input, the server derives an adaptor secret `γ_i` and point `T_i = γ_i·G` deterministically from `(match_id, input_id, branch, sighash, server_priv)` with strong domain separation. Normalizing `T_i` to even‑Y is fine for determinism, **but the critical rule is that the final nonce point you sign against must have even‑Y** (see below).
- Clients produce pre‑signatures against the **final** nonce point:
  1. Sample/derive nonce `k` (RFC6979).
  2. Compute `R = k·G`, then **form the final nonce point** `R' = R + T_i`.
  3. If `R'.y` is odd, advance the RFC6979 iterator (or resample `k`) and recompute until `R'` has **even‑Y**.
  4. Let `r' = x(R')` and `e = BLAKE256(r' || m)`.
  5. **Pre‑signature:** `s'_i = k − e·x (mod n)` (note the **minus**, per DCRv0).
- On the winning branch, the server reveals `γ_i`. The winner finalizes with `s_i = s'_i + γ_i (mod n)`. The pair `(r', s_i)` is a valid DCRv0 Schnorr signature for that input.

> **Why even‑Y?** Decred’s Schnorr is x‑only and requires the nonce point used for verification to have an **even y‑coordinate**. With adaptors you sign against `R' = R + T`, so enforce even‑Y on **`R'`**, not just on `T`.

---

## End‑to‑end flow

1) **Allocate escrow / create match**  
   Client obtains or generates `A_c` (and optionally `B_c`).  
   Call `PongReferee.AllocateEscrow(player_id, A_c, bet_atoms, csv)` or `CreateMatch(A_c, B_c, csv)`.  
   Server returns `match_id`, coefficients (`a_*`), branch keys (`X_*`), and deposit info: `redeem_script_hex`, `pk_script_hex`, and `deposit_address`.

2) **Fund deposit**  
   Player sends `bet_atoms` to `deposit_address` (P2SH for their redeem script).  
   Track via `FundingStatus(match_id, player_id)` (watcher streams UTXOs/confirmations) or submit exact UTXOs via `SubmitFunding` (server validates against chain and expected `pkScript`).

3) **Draft transaction and adaptors**  
   After funding is observed, call `GetDraftAndAdaptors(match_id, branch)`.  
   Server builds a draft that consumes **both** escrows and pays to the winner’s output. For each input, server returns the `m_sighash` and adaptor point `T_i`, keeping `γ_i` secret.

4) **Client pre‑signs (per input)**  
   Compute pre‑sigs `(r', s'_i)` **against the even‑Y final nonce point `R'`** (see steps above) and submit via `SubmitPreSig`. Pre‑sigs are bound to the exact sighashes and branch.

5) **Reveal and finalize (winner only)**  
   After the game, the server reveals `γ_i` for the winning branch (`RevealAdaptors`).  
   Winner computes `s_i = s'_i + γ_i (mod n)`, attaches final Schnorr signatures to the draft, and broadcasts.

6) **Timeout recovery (safety fallback)**  
   If settlement stalls, the depositor recovers funds after `CSV` blocks using the timeout path `P_i` (input sequence must satisfy CSV).

---


## Client usage (TUI/CLI hints)

- **Generate a key**: Create a fresh secp256k1 key; keep the private key client‑side only.
- **Allocate escrow**: 
...

---

## Security considerations

- **Key custody**: Private keys never leave the client. The server cannot sign on behalf of players.
- **Non‑custodial deposits**: On‑chain UTXOs are to P2SH redeem scripts that the server cannot spend without your signatures.
- **Draft‑tx substitution**: Clients must locally recompute the sighash from the provided draft and verify outputs/fees before pre‑signing.
- **Even‑Y rule (critical)**: Decred rejects signatures whose nonce point has odd‑Y. Enforce **even‑Y on the final nonce point `R'`** used in signing (iterate RFC6979 if needed).
- **Deterministic adaptors**: `γ_i` is derived with domain separation and includes the specific sighash and branch, reducing misuse and enabling reproducibility across restarts.
- **Adaptor reuse/cross‑binding**: Never reuse `γ/T` across inputs/branches. Keep per‑input, per‑branch, per‑sighash domain tags.
- **Timeout safety**: CSV fallback ensures unilateral recovery if reveals or broadcasts do not happen.
- **Privacy**: Funding appears as standard P2SH. The winner spend reveals the used branch; until spent, third parties cannot distinguish the branch.
- **PoC limitations**: Current implementation logs final `s` and validates flows; fully automated tx assembly/broadcast for all paths may require additional integration.

---
