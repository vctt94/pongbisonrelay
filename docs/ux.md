
## End‑to‑end UX (players ↔ server)

  The following walks through the flow in plain terms from both players’ and the server’s perspectives.

  ### 1) Alice creates a room and funds her deposit
  - Alice creates a waiting room.
  - The app shows a deposit address and its locking code (pkScript).
  - Alice’s wallet should verify the locking code matches exactly before funding.
  - She funds exactly the required amount and waits for confirmations.

  ### 2) Bob joins and funds his deposit
  - Bob joins the room.
  - The server tells Bob the required amount, the deposit pkScript, and how many confirmations are needed.
  - Bob’s app reconstructs the locking code locally and only funds if the bytes match exactly.
  - Bob funds the exact amount and waits for confirmations.

  ### 3) Server prepares two payout drafts and asks for pre‑approval
  - After both deposits confirm, the server builds two possible payout transactions: “Alice wins” and “Bob wins”.
  - Each draft spends one deposit from Alice and one from Bob, paying everything (minus fee) to the winner’s address that each player provided.
  - The server asks each player to pre‑approve both outcomes for their own deposit.

  ### 4) Players pre‑approve on their device
  - Your app re‑checks the drafts (inputs, amounts, and your locking code). If anything differs, it aborts.
  - If all checks pass, your app sends a pre‑approval for each outcome; your private key never leaves your device.
  - These pre‑approvals cannot be repurposed: they only work with the exact drafts that pay the winner.

  ### 5) Match result and reveal
  - When the game ends, the server determines the winner and reveals a small secret tied to the winning outcome.
  - The winner’s app combines that secret with the stored pre‑approvals (yours and your opponent’s, provided for the winning side) to finish the payout transaction.
  - Either the winner or the server can broadcast the finished transaction, but it only pays the declared winner’s address.

  ### 6) Winner finalizes and broadcasts
  - The app sets an appropriate fee and broadcasts the winning draft.
  - Everyone can verify the result on‑chain.

  ### 7) Safety and refunds
  - Deposits are controlled solely by each player’s key. There is no server key in the deposit.
  - If the match never completes or the server withholds the reveal, your funds remain under your control.
  - You can move your own deposit at any time; the timeout is a built‑in fallback path.
  - The server cannot redirect funds to itself.

  Deposit script (simplified view):
  ```
  IF            # immediate spend path
    owner_signature
  ELSE          # after a delay (CSV)
    csv_delay
    owner_signature
  ENDIF
  ```


