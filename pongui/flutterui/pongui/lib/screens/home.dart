import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/components/home/main_content.dart';
import 'package:pongui/components/shared_layout.dart';
import 'package:pongui/models/pong.dart';
import 'package:provider/provider.dart';

class HomeScreen extends StatefulWidget {
  const HomeScreen({super.key});

  @override
  State<HomeScreen> createState() => _HomeScreenState();
}

class _HomeScreenState extends State<HomeScreen> {
  @override
  Widget build(BuildContext context) {
    // Only rebuild this widget when the "in game" flag toggles; other
    // updates are handled by Consumers inside each branch.
    final gameInProgress = context.select<PongModel, bool>((m) => m.isGameStarted);

    return SharedLayout(
      title: "Pong Game - Home",
      child: gameInProgress
          ? Padding(
              padding: const EdgeInsets.only(top: 12.0),
              child: Consumer<PongModel>(
                builder: (_, model, __) => MainContent(pongModel: model),
              ),
            )
          : Consumer<PongModel>(builder: (context, pongModel, _) {
              return SingleChildScrollView(
                padding: const EdgeInsets.only(bottom: 24),
                child: Column(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                // 1) Top area: bet status
                Center(
                  child: Container(
                    width: MediaQuery.of(context).size.width * 0.85,
                    margin: const EdgeInsets.only(top: 16.0),
                    child: Card(
                      color: const Color(0xFF1B1E2C), // Dark card background
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Padding(
                        padding: const EdgeInsets.all(16.0),
                        child: Row(
                          children: [
                            const Icon(Icons.attach_money, color: Colors.amber),
                            const SizedBox(width: 8),
                            Text(
                              "Bet: ${pongModel.betAmt / 1e8}",
                              style: const TextStyle(
                                color: Colors.white,
                                fontSize: 16,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const Spacer(),
                            if (pongModel.escrowFunded) ...[
                              Tooltip(
                                message: pongModel.fundingStatus.isNotEmpty ? pongModel.fundingStatus : (pongModel.escrowConfirmed ? 'Deposit confirmed (${pongModel.escrowConfs})' : 'Deposit seen (mempool)'),
                                child: Row(children: const [
                                  Icon(Icons.check_circle, color: Colors.greenAccent, size: 16),
                                  SizedBox(width: 6),
                                  Text('Funding seen', style: TextStyle(color: Colors.greenAccent)),
                                ]),
                              ),
                              const SizedBox(width: 12),
                            ],
                            // Consider escrow active strictly when there is a current escrow id.
                            if (pongModel.escrowId.isNotEmpty)
                              Container(
                                padding: const EdgeInsets.symmetric(horizontal: 10, vertical: 6),
                                decoration: BoxDecoration(
                                  color: Colors.green.withOpacity(0.15),
                                  borderRadius: BorderRadius.circular(8),
                                  border: Border.all(color: Colors.greenAccent, width: 1),
                                ),
                                child: Row(children: [
                                  const Icon(Icons.lock, color: Colors.greenAccent, size: 16),
                                  const SizedBox(width: 6),
                                  Text(
                                    pongModel.escrowId.isNotEmpty ? 'Escrow: ${pongModel.escrowId}' : 'Escrow active',
                                    style: const TextStyle(color: Colors.greenAccent),
                                  ),
                                ]),
                              )
                            else
                              Row(children: [
                                ElevatedButton.icon(
                                  onPressed: () async {
                                    try {
                                      await Golib.generateSettlementSessionKey();
                                      final payout = pongModel.payoutAddressOrPubkey;
                                      if (payout.trim().isEmpty) {
                                        ScaffoldMessenger.of(context).showSnackBar(
                                          const SnackBar(content: Text('Set payout address in Settings first.')),
                                        );
                                        return;
                                      }
                                      final betAtoms = pongModel.betAmt > 0 ? pongModel.betAmt : 100000000;
                                      final res = await Golib.openEscrow(
                                        payout: payout,
                                        betAtoms: betAtoms,
                                        csvBlocks: 64,
                                      );
                                      final id = (res['escrow_id'] as String?) ?? '';
                                      final dep = (res['deposit_address'] as String?) ?? '';
                                      final pk  = (res['pk_script_hex'] as String?) ?? '';
                                      pongModel.setEscrowDetails(id, dep, pk);
                                      pongModel.setEscrowBetAtoms(betAtoms);
                                      if (!context.mounted) return;
                                      ScaffoldMessenger.of(context).showSnackBar(
                                        SnackBar(content: Text('Escrow opened. Deposit to ${res['deposit_address']}')),
                                      );
                                    } catch (e) {
                                      if (!context.mounted) return;
                                      ScaffoldMessenger.of(context).showSnackBar(
                                        SnackBar(content: Text('Escrow error: $e')),
                                      );
                                    }
                                  },
                                  icon: const Icon(Icons.lock_open),
                                  label: const Text('Open Escrow'),
                                  style: ElevatedButton.styleFrom(backgroundColor: Colors.blueAccent),
                                ),
                                const SizedBox(width: 8),
                                if (pongModel.betAmt > 0 && pongModel.currentWR == null && pongModel.escrowFunded)
                                  ElevatedButton(
                                    onPressed: pongModel.createWaitingRoom,
                                    style: ElevatedButton.styleFrom(
                                      backgroundColor: Colors.blueGrey,
                                    ),
                                    child: const Text("Create Waiting Room"),
                                  ),
                              ]),
                            const SizedBox(width: 8),
                            // Presign action when escrow confirmed and in a room
                            Builder(builder: (ctx) {
                              final canPresign = pongModel.escrowConfirmed && pongModel.currentWR != null;
                              final onPressed = canPresign
                                  ? () async {
                                      final wr = pongModel.currentWR!;
                                      final matchId = '${wr.id}|${wr.host}';
                                      pongModel.lastMatchId = matchId;
                                      try {
                                        await Golib.startPreSign(matchId);
                                        if (!ctx.mounted) return;
                                        ScaffoldMessenger.of(ctx).showSnackBar(
                                          const SnackBar(content: Text('Presign completed')),
                                        );
                                      } catch (e) {
                                        if (!ctx.mounted) return;
                                        ScaffoldMessenger.of(ctx).showSnackBar(
                                          SnackBar(content: Text('Presign error: $e')),
                                        );
                                      }
                                    }
                                  : null;
                              final button = ElevatedButton.icon(
                                onPressed: onPressed,
                                icon: const Icon(Icons.fact_check),
                                label: const Text('Presign'),
                                style: ElevatedButton.styleFrom(
                                  backgroundColor: canPresign ? Colors.teal : Colors.grey,
                                ),
                              );
                              if (canPresign) return button;
                              final msg = pongModel.currentWR == null
                                  ? 'Join or create a room to presign'
                                  : (pongModel.escrowConfirmed ? '' : 'Wait for deposit confirmation');
                              if (msg.isEmpty) {
                                return button; // no tooltip when message is empty to avoid zero-size hit test
                              }
                              return Tooltip(message: msg, child: AbsorbPointer(child: button));
                            }),
                            const SizedBox(width: 8),
                            if (pongModel.betAmt > 0 && pongModel.currentWR == null && pongModel.escrowId.isNotEmpty && pongModel.escrowFunded)
                              ElevatedButton.icon(
                                onPressed: pongModel.createWaitingRoom,
                                icon: const Icon(Icons.meeting_room),
                                label: const Text('Create Room'),
                                style: ElevatedButton.styleFrom(backgroundColor: Colors.blueGrey),
                              ),
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
                
                // 2) Current waiting room info
                Center(
                  child: Container(
                    width: MediaQuery.of(context).size.width * 0.85,
                    margin: const EdgeInsets.only(top: 16.0),
                    child: Card(
                      color: const Color(0xFF1B1E2C), // Dark card background
                      shape: RoundedRectangleBorder(
                        borderRadius: BorderRadius.circular(12),
                      ),
                      child: Padding(
                        padding: const EdgeInsets.all(16.0),
                        child: Column(
                          crossAxisAlignment: CrossAxisAlignment.start,
                          children: [
                            if (pongModel.escrowDepositAddress.isNotEmpty) ...[
                              Row(
                                children: [
                                  const Icon(Icons.account_balance_wallet, color: Colors.amber),
                                  const SizedBox(width: 8),
                                  const Text(
                                    'Deposit Address:',
                                    style: TextStyle(color: Colors.white70),
                                  ),
                                  const SizedBox(width: 8),
                                  Expanded(
                                    child: SelectableText(
                                      pongModel.escrowDepositAddress,
                                      style: const TextStyle(color: Colors.white),
                                    ),
                                  ),
                                  IconButton(
                                    tooltip: 'Copy',
                                    onPressed: () async {
                                      await Clipboard.setData(ClipboardData(text: pongModel.escrowDepositAddress));
                                      if (!context.mounted) return;
                                      ScaffoldMessenger.of(context).showSnackBar(
                                        const SnackBar(content: Text('Address copied')),
                                      );
                                    },
                                    icon: const Icon(Icons.copy, color: Colors.white70),
                                  ),
                                ],
                              ),
                              const SizedBox(height: 12),
                            ],
                            const Text(
                              "Current Waiting Room",
                              style: TextStyle(
                                color: Colors.white,
                                fontSize: 18,
                                fontWeight: FontWeight.bold,
                              ),
                            ),
                            const SizedBox(height: 8),
                            Row(
                              mainAxisAlignment: MainAxisAlignment.spaceBetween,
                              children: [
                                Text(
                                  "Room ID: ${pongModel.currentWR?.id ?? ""}",
                                  style: const TextStyle(
                                    color: Colors.white,
                                  ),
                                ),
                                Text(
                                  pongModel.isReady ? "Ready" : "Not Ready",
                                  style: TextStyle(
                                    color: pongModel.isReady
                                        ? Colors.green
                                        : Colors.white,
                                    fontWeight: pongModel.isReady
                                        ? FontWeight.bold
                                        : FontWeight.normal,
                                  ),
                                ),
                              ],
                            ),
                            const SizedBox(height: 8),
                            Text(
                              "Players: ${pongModel.currentWR?.players.length ?? 0} / 2",
                              style: const TextStyle(
                                color: Colors.white,
                              ),
                            ),
                            
                            // Add ready/leave buttons if in a room
                            if (pongModel.currentWR != null) ...[
                              const SizedBox(height: 16),
                              Row(
                                mainAxisAlignment: MainAxisAlignment.end,
                                children: [
                                  ElevatedButton(
                                    onPressed: pongModel.toggleReady,
                                    style: ElevatedButton.styleFrom(
                                      backgroundColor: pongModel.isReady
                                          ? Colors.orange
                                          : Colors.green,
                                    ),
                                    child: Text(pongModel.isReady
                                        ? "Cancel Ready"
                                        : "Ready"),
                                  ),
                                  const SizedBox(width: 8),
                                  ElevatedButton(
                                    onPressed: () => pongModel.leaveWaitingRoom(),
                                    style: ElevatedButton.styleFrom(
                                      backgroundColor: Colors.redAccent,
                                    ),
                                    child: const Text("Leave Room"),
                                  ),
                                ],
                              ),
                            ],
                          ],
                        ),
                      ),
                    ),
                  ),
                ),
                
                // 3) Error message if exists
                if (pongModel.errorMessage.isNotEmpty)
                  Center(
                    child: Container(
                      width: MediaQuery.of(context).size.width * 0.85,
                      margin: const EdgeInsets.only(top: 16.0),
                      child: Card(
                        color: Colors.red.shade800,
                        shape: RoundedRectangleBorder(
                          borderRadius: BorderRadius.circular(12),
                        ),
                        child: Padding(
                          padding: const EdgeInsets.all(12.0),
                          child: Row(
                            children: [
                              const Icon(Icons.error, color: Colors.white),
                              const SizedBox(width: 8),
                              Expanded(
                                child: SelectableText(
                                  pongModel.errorMessage,
                                  style: const TextStyle(color: Colors.white),
                                ),
                              ),
                              Material(
                                color: Colors.transparent,
                                child: InkWell(
                                  onTap: () async {
                                    await Clipboard.setData(ClipboardData(text: pongModel.errorMessage));
                                    if (!context.mounted) return;
                                    ScaffoldMessenger.of(context).showSnackBar(
                                      const SnackBar(content: Text('Error copied to clipboard')),
                                    );
                                  },
                                  borderRadius: BorderRadius.circular(20),
                                  child: const Padding(
                                    padding: EdgeInsets.all(8.0),
                                    child: Icon(Icons.copy, color: Colors.white, size: 20),
                                  ),
                                ),
                              ),
                              Material(
                                color: Colors.transparent,
                                child: InkWell(
                                  onTap: () {
                                    pongModel.clearErrorMessage();
                                  },
                                  borderRadius: BorderRadius.circular(20),
                                  child: const Padding(
                                    padding: EdgeInsets.all(8.0),
                                    child: Icon(Icons.close,
                                        color: Colors.white, size: 20),
                                  ),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ),
                    ),
                  ),
                
                // 4) Main content
                Padding(
                  padding: const EdgeInsets.only(top: 12.0),
                  child: MainContent(pongModel: pongModel),
                ),
              ],
              ),
            );
            }),
    );
  }
}
