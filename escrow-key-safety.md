## Escrow Deposit Safety and Session Key Ownership

### What the deposit address is
- **Deposit address**: A P2SH address generated from an escrow redeem script.
- **Built from**: Your ephemeral settlement session public key (33‑byte compressed) and a CSV timeout value.
- **Not your wallet address**: It’s an escrow address; funds are controlled by the session key, not your wallet login key.

### Redeem script paths (high‑level)
- **Win path (immediate)**: Requires a signature with the session pubkey to settle the game and pay the winner.
- **Timeout path (refund)**: After `CSV` blocks, requires a signature with the same session pubkey to refund back to the depositor.

Implication: **Both paths require the session private key**. If you lose the session key, the funds are effectively unrecoverable.

### Where the session key lives
- Generated client‑side freshly per session.
- Cached in memory and (optionally) persisted to disk as JSON:
  - File: `settlement_session_key.json` under the client data directory.
  - The app can also archive the current key into `historic_sessions/` with a match‑scoped filename.

### What happens if the server disappears
- You can still use the **timeout path** after `CSV` blocks to refund, but you must have the session private key from `settlement_session_key.json`.

### Recommended operational practices
- Back up `settlement_session_key.json` immediately after it is created.
- Archive a copy per match (after opening escrow) so you have historical keys mapped to games.
- Do not reuse session keys across matches.

### Current manual backup
1) Locate `settlement_session_key.json` in your client data directory.
2) Copy it to an encrypted backup location (e.g., password‑protected vault).
3) After a match, archive a copy named with the match ID in a separate safe location.

### Planned improvements (see TODOs)
- Automatic backup of new session keys to a safe location.
- A UI/command to list active escrows along with their session pubkeys and the path to the associated key file.
- A refund UI action that constructs and broadcasts a timeout‑path refund after `CSV` is satisfied.

### Recovery/refund overview (high‑level)
1) Wait until your deposit UTXO reaches the required `CSV` relative height.
2) Build a transaction spending the P2SH UTXO via the timeout branch (includes sequence/CSV), signing with the session private key.
3) Broadcast the transaction to reclaim funds.

Note: The app will provide an integrated refund UI so you don’t have to craft this manually.


