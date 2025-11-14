import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/components/home/main_content.dart';
import 'package:pongui/components/shared_layout.dart';
import 'package:pongui/models/pong.dart';
import 'package:golib_plugin/definitions.dart';
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
    final gameInProgress =
        context.select<PongModel, bool>((m) => m.isGameStarted);

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
                          color: const Color(0xFF1B1E2C),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Padding(
                            padding: const EdgeInsets.all(16.0),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                Row(
                                  mainAxisAlignment:
                                      MainAxisAlignment.spaceBetween,
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    // Left side: Bet amount and address
                                    Flexible(
                                      child: Wrap(
                                        spacing: 8,
                                        runSpacing: 8,
                                        crossAxisAlignment:
                                            WrapCrossAlignment.center,
                                        children: [
                                          Row(
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              const Icon(Icons.attach_money,
                                                  color: Colors.amber),
                                              const SizedBox(width: 8),
                                              Text(
                                                "Bet: ${pongModel.betAmt / 1e8}",
                                                style: const TextStyle(
                                                  color: Colors.white,
                                                  fontSize: 16,
                                                  fontWeight: FontWeight.bold,
                                                ),
                                              ),
                                            ],
                                          ),
                                          if (pongModel
                                              .walletAddress.isNotEmpty)
                                            Tooltip(
                                              message: pongModel.walletAddress,
                                              child: Container(
                                                padding:
                                                    const EdgeInsets.symmetric(
                                                        horizontal: 10,
                                                        vertical: 6),
                                                decoration: BoxDecoration(
                                                  color: Colors.green
                                                      .withOpacity(0.15),
                                                  borderRadius:
                                                      BorderRadius.circular(8),
                                                  border: Border.all(
                                                      color: Colors.greenAccent,
                                                      width: 1),
                                                ),
                                                child: Row(
                                                  mainAxisSize:
                                                      MainAxisSize.min,
                                                  children: [
                                                    const Icon(
                                                        Icons.check_circle,
                                                        color:
                                                            Colors.greenAccent,
                                                        size: 16),
                                                    const SizedBox(width: 6),
                                                    Text(
                                                      '${pongModel.walletAddress.substring(0, 8)}...${pongModel.walletAddress.substring(pongModel.walletAddress.length - 6)}',
                                                      style: const TextStyle(
                                                          color: Colors
                                                              .greenAccent,
                                                          fontFamily:
                                                              'monospace'),
                                                    ),
                                                  ],
                                                ),
                                              ),
                                            ),
                                        ],
                                      ),
                                    ),
                                    // Right side: Buttons
                                    Flexible(
                                      child: Wrap(
                                        spacing: 8,
                                        runSpacing: 8,
                                        crossAxisAlignment:
                                            WrapCrossAlignment.center,
                                        alignment: WrapAlignment.end,
                                        runAlignment: WrapAlignment.end,
                                        children: [
                                          if (!pongModel.serverIsF2P) ...[
                                            if (pongModel.escrowFunded) ...[
                                              Tooltip(
                                                message: pongModel.fundingStatus
                                                        .isNotEmpty
                                                    ? pongModel.fundingStatus
                                                    : (pongModel.escrowConfirmed
                                                        ? 'Deposit confirmed (${pongModel.escrowConfs})'
                                                        : 'Deposit seen (mempool)'),
                                                child: Row(
                                                  mainAxisSize:
                                                      MainAxisSize.min,
                                                  children: const [
                                                    Icon(Icons.check_circle,
                                                        color:
                                                            Colors.greenAccent,
                                                        size: 16),
                                                    SizedBox(width: 6),
                                                    Text('Funding seen',
                                                        style: TextStyle(
                                                            color: Colors
                                                                .greenAccent)),
                                                  ],
                                                ),
                                              ),
                                            ],
                                            if (pongModel.escrowId.isNotEmpty)
                                              Tooltip(
                                                message: pongModel.escrowId,
                                                child: Container(
                                                  padding: const EdgeInsets
                                                      .symmetric(
                                                      horizontal: 10,
                                                      vertical: 6),
                                                  decoration: BoxDecoration(
                                                    color: Colors.green
                                                        .withOpacity(0.15),
                                                    borderRadius:
                                                        BorderRadius.circular(
                                                            8),
                                                    border: Border.all(
                                                        color:
                                                            Colors.greenAccent,
                                                        width: 1),
                                                  ),
                                                  child: Row(
                                                    mainAxisSize:
                                                        MainAxisSize.min,
                                                    children: [
                                                      const Icon(Icons.lock,
                                                          color: Colors
                                                              .greenAccent,
                                                          size: 16),
                                                      const SizedBox(width: 6),
                                                      Flexible(
                                                        child: Text(
                                                          pongModel.escrowId
                                                                      .length >
                                                                  12
                                                              ? 'Escrow: ${pongModel.escrowId.substring(0, 8)}...${pongModel.escrowId.substring(pongModel.escrowId.length - 4)}'
                                                              : 'Escrow: ${pongModel.escrowId}',
                                                          style: const TextStyle(
                                                              color: Colors
                                                                  .greenAccent),
                                                          overflow: TextOverflow
                                                              .ellipsis,
                                                        ),
                                                      ),
                                                    ],
                                                  ),
                                                ),
                                              )
                                            else
                                              ElevatedButton.icon(
                                                onPressed: () async {
                                                  try {
                                                    if (!pongModel
                                                        .isWalletAuthenticated) {
                                                      ScaffoldMessenger.of(
                                                              context)
                                                          .showSnackBar(
                                                        const SnackBar(
                                                            content: Text(
                                                                'Please login first')),
                                                      );
                                                      return;
                                                    }
                                                    await Golib
                                                        .generateSettlementSessionKey();
                                                    final payout = pongModel
                                                        .payoutAddressOrPubkey;
                                                    if (payout.trim().isEmpty) {
                                                      ScaffoldMessenger.of(
                                                              context)
                                                          .showSnackBar(
                                                        const SnackBar(
                                                            content: Text(
                                                                'Payout address not set. Please login again.')),
                                                      );
                                                      return;
                                                    }
                                                    final betAtoms =
                                                        pongModel.betAmt > 0
                                                            ? pongModel.betAmt
                                                            : DEFAULT_BET_ATOMS;
                                                    final res =
                                                        await Golib.openEscrow(
                                                      payout: payout,
                                                      betAtoms: betAtoms,
                                                      csvBlocks: CSV_BLOCKS,
                                                    );
                                                    final id = (res['escrow_id']
                                                            as String?) ??
                                                        '';
                                                    final dep =
                                                        (res['deposit_address']
                                                                as String?) ??
                                                            '';
                                                    final pk =
                                                        (res['pk_script_hex']
                                                                as String?) ??
                                                            '';
                                                    final redeem =
                                                        (res['redeem_script_hex']
                                                                as String?) ??
                                                            '';
                                                    final csvBlocks =
                                                        (res['csv_blocks']
                                                                as int?) ??
                                                            CSV_BLOCKS;
                                                    if (id.isEmpty ||
                                                        dep.isEmpty ||
                                                        redeem.isEmpty ||
                                                        pk.isEmpty) {
                                                      if (!context.mounted)
                                                        return;
                                                      ScaffoldMessenger.of(
                                                              context)
                                                          .showSnackBar(
                                                        const SnackBar(
                                                            content: Text(
                                                                'Escrow error: missing critical data. Try again.')),
                                                      );
                                                      return;
                                                    }
                                                    final persisted =
                                                        await pongModel
                                                            .persistInitialEscrowInfo(
                                                      escrowId: id,
                                                      betAtoms: betAtoms,
                                                      csvBlocks: csvBlocks,
                                                      pkScriptHex: pk,
                                                      redeemScriptHex: redeem,
                                                    );
                                                    if (!persisted) {
                                                      if (!context.mounted)
                                                        return;
                                                      ScaffoldMessenger.of(
                                                              context)
                                                          .showSnackBar(
                                                        const SnackBar(
                                                            content: Text(
                                                                'Failed to save escrow info. Deposit address not shown.')),
                                                      );
                                                      return;
                                                    }
                                                    pongModel.setEscrowDetails(
                                                        id, dep,
                                                        pkScriptHex: pk,
                                                        redeemScriptHex: redeem,
                                                        csvBlocks: csvBlocks);
                                                    pongModel.setEscrowBetAtoms(
                                                        betAtoms);
                                                    if (!context.mounted)
                                                      return;
                                                    ScaffoldMessenger.of(
                                                            context)
                                                        .showSnackBar(
                                                      SnackBar(
                                                          content: Text(
                                                              'Escrow opened. Deposit to ${res['deposit_address']}')),
                                                    );
                                                  } catch (e) {
                                                    if (!context.mounted)
                                                      return;
                                                    ScaffoldMessenger.of(
                                                            context)
                                                        .showSnackBar(
                                                      SnackBar(
                                                          content: Text(
                                                              'Escrow error: $e')),
                                                    );
                                                  }
                                                },
                                                icon:
                                                    const Icon(Icons.lock_open),
                                                label:
                                                    const Text('Open Escrow'),
                                                style: ElevatedButton.styleFrom(
                                                    backgroundColor:
                                                        Colors.blueAccent),
                                              ),
                                          ],
                                          if (pongModel.currentWR == null)
                                            Builder(builder: (ctx) {
                                              final reason =
                                                  _createRoomDisabledReason(
                                                      pongModel);
                                              final canCreate = reason == null;
                                              final btn = ElevatedButton.icon(
                                                onPressed: canCreate
                                                    ? pongModel
                                                        .createWaitingRoom
                                                    : null,
                                                icon: const Icon(
                                                    Icons.meeting_room),
                                                label:
                                                    const Text('Create Room'),
                                                style: ElevatedButton.styleFrom(
                                                  backgroundColor: canCreate
                                                      ? Colors.blueGrey
                                                      : Colors
                                                          .blueGrey.shade200,
                                                ),
                                              );
                                              if (canCreate) {
                                                return btn;
                                              }
                                              return Tooltip(
                                                  message: reason,
                                                  child: AbsorbPointer(
                                                      child: btn));
                                            }),
                                          Builder(builder: (ctx) {
                                            final canPresign =
                                                pongModel.escrowConfirmed &&
                                                    pongModel.currentWR != null;
                                            final onPressed = canPresign
                                                ? () async {
                                                    final wr =
                                                        pongModel.currentWR!;
                                                    final matchId =
                                                        '${wr.id}|${wr.host}';
                                                    pongModel.lastMatchId =
                                                        matchId;
                                                    try {
                                                      await Golib.startPreSign(
                                                          matchId);
                                                      if (!ctx.mounted) return;
                                                      ScaffoldMessenger.of(ctx)
                                                          .showSnackBar(
                                                        const SnackBar(
                                                            content: Text(
                                                                'Presign completed')),
                                                      );
                                                    } catch (e) {
                                                      if (!ctx.mounted) return;
                                                      ScaffoldMessenger.of(ctx)
                                                          .showSnackBar(
                                                        SnackBar(
                                                            content: Text(
                                                                'Presign error: $e')),
                                                      );
                                                    }
                                                  }
                                                : null;
                                            final button = ElevatedButton.icon(
                                              onPressed: onPressed,
                                              icon:
                                                  const Icon(Icons.fact_check),
                                              label: const Text('Presign'),
                                              style: ElevatedButton.styleFrom(
                                                backgroundColor: canPresign
                                                    ? Colors.teal
                                                    : Colors.grey,
                                              ),
                                            );
                                            if (canPresign) return button;
                                            final msg = pongModel.currentWR ==
                                                    null
                                                ? 'Join or create a room to presign'
                                                : (pongModel.escrowConfirmed
                                                    ? ''
                                                    : 'Wait for 1 confirmation before presign');
                                            if (msg.isEmpty) {
                                              return button;
                                            }
                                            return Tooltip(
                                                message: msg,
                                                child: AbsorbPointer(
                                                    child: button));
                                          }),
                                        ],
                                      ),
                                    ),
                                  ],
                                ),
                                if (!pongModel.serverIsF2P &&
                                    pongModel.escrowFunded &&
                                    !pongModel.escrowConfirmed) ...[
                                  const SizedBox(height: 8),
                                  const Row(
                                    children: [
                                      Icon(Icons.info_outline,
                                          color: Colors.amberAccent, size: 18),
                                      SizedBox(width: 6),
                                      Expanded(
                                        child: Text(
                                          'Presign becomes available after your deposit has at least 1 confirmation.',
                                          style: TextStyle(
                                            color: Colors.amberAccent,
                                            fontSize: 12,
                                          ),
                                        ),
                                      ),
                                    ],
                                  ),
                                ],
                                if (pongModel.serverIsF2P) ...[
                                  const SizedBox(height: 12),
                                  _buildServerModeBanner(pongModel),
                                ],
                              ],
                            ),
                          ),
                        ),
                      ),
                    ),

                    // 2) Current waiting room info
                    // 2) Current waiting room info
                    Center(
                      child: Container(
                        width: MediaQuery.of(context).size.width * 0.85,
                        margin: const EdgeInsets.only(top: 16.0),
                        child: Card(
                          color:
                              const Color(0xFF1B1E2C), // Dark card background
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(12),
                          ),
                          child: Padding(
                            padding: const EdgeInsets.all(16.0),
                            child: Column(
                              crossAxisAlignment: CrossAxisAlignment.start,
                              children: [
                                if (pongModel.escrowDepositAddress.isNotEmpty &&
                                    pongModel.escrowInfoPersisted &&
                                    pongModel.escrowRefundSessionValid) ...[
                                  Row(
                                    children: [
                                      const Icon(Icons.account_balance_wallet,
                                          color: Colors.amber),
                                      const SizedBox(width: 8),
                                      const Text(
                                        'Deposit Address:',
                                        style: TextStyle(color: Colors.white70),
                                      ),
                                      const SizedBox(width: 8),
                                      Expanded(
                                        child: SelectableText(
                                          pongModel.escrowDepositAddress,
                                          style: const TextStyle(
                                              color: Colors.white),
                                        ),
                                      ),
                                      IconButton(
                                        tooltip: 'Copy',
                                        onPressed: () async {
                                          await Clipboard.setData(ClipboardData(
                                              text: pongModel
                                                  .escrowDepositAddress));
                                          if (!context.mounted) return;
                                          ScaffoldMessenger.of(context)
                                              .showSnackBar(
                                            const SnackBar(
                                                content:
                                                    Text('Address copied')),
                                          );
                                        },
                                        icon: const Icon(Icons.copy,
                                            color: Colors.white70),
                                      ),
                                    ],
                                  ),
                                  const SizedBox(height: 12),
                                  Container(
                                    padding: const EdgeInsets.all(12),
                                    decoration: BoxDecoration(
                                      color: Colors.amber.withOpacity(0.15),
                                      border: Border.all(
                                          color: Colors.amberAccent, width: 1),
                                      borderRadius: BorderRadius.circular(8),
                                    ),
                                    child: Row(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        const Icon(Icons.warning_amber_rounded,
                                            color: Colors.amberAccent),
                                        const SizedBox(width: 8),
                                        Expanded(
                                          child: Text(
                                            'Warning: Deposit exactly ${(pongModel.betAmt / 1e8).toStringAsFixed(2)} DCR (default). Do NOT send a different amount.',
                                            style: const TextStyle(
                                              color: Colors.amberAccent,
                                              fontWeight: FontWeight.w600,
                                            ),
                                          ),
                                        ),
                                      ],
                                    ),
                                  ),
                                  const SizedBox(height: 12),
                                ],
                                if (pongModel.escrowInfoError.isNotEmpty)
                                  Padding(
                                    padding:
                                        const EdgeInsets.only(bottom: 12.0),
                                    child: Text(
                                      pongModel.escrowInfoError,
                                      style: const TextStyle(
                                        color: Colors.redAccent,
                                        fontSize: 12,
                                      ),
                                    ),
                                  ),
                                if (pongModel
                                    .escrowRefundSessionError.isNotEmpty)
                                  Padding(
                                    padding:
                                        const EdgeInsets.only(bottom: 12.0),
                                    child: Text(
                                      pongModel.escrowRefundSessionError,
                                      style: const TextStyle(
                                        color: Colors.redAccent,
                                        fontSize: 12,
                                      ),
                                    ),
                                  ),
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
                                  mainAxisAlignment:
                                      MainAxisAlignment.spaceBetween,
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
                                Builder(
                                  builder: (context) {
                                    final room = pongModel.currentWR;
                                    int totalPlayers = 0;
                                    int readyPlayers = 0;
                                    if (room != null) {
                                      totalPlayers = room.players.length;
                                      readyPlayers = room.players
                                          .where((p) => p.ready)
                                          .length;
                                    }
                                    return Column(
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        Text(
                                          "Players: $totalPlayers / 2",
                                          style: const TextStyle(
                                            color: Colors.white,
                                          ),
                                        ),
                                        const SizedBox(height: 4),
                                        Text(
                                          "Ready: $readyPlayers / 2",
                                          style: TextStyle(
                                            color: readyPlayers == 2
                                                ? Colors.greenAccent
                                                : Colors.white70,
                                          ),
                                        ),
                                      ],
                                    );
                                  },
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
                                        onPressed: () =>
                                            pongModel.leaveWaitingRoom(),
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
                                      style:
                                          const TextStyle(color: Colors.white),
                                    ),
                                  ),
                                  Material(
                                    color: Colors.transparent,
                                    child: InkWell(
                                      onTap: () async {
                                        await Clipboard.setData(ClipboardData(
                                            text: pongModel.errorMessage));
                                        if (!context.mounted) return;
                                        ScaffoldMessenger.of(context)
                                            .showSnackBar(
                                          const SnackBar(
                                              content: Text(
                                                  'Error copied to clipboard')),
                                        );
                                      },
                                      borderRadius: BorderRadius.circular(20),
                                      child: const Padding(
                                        padding: EdgeInsets.all(8.0),
                                        child: Icon(Icons.copy,
                                            color: Colors.white, size: 20),
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

  Widget _buildServerModeBanner(PongModel model) {
    return Row(
      children: [
        Chip(
          avatar: const Icon(Icons.videogame_asset,
              color: Colors.greenAccent, size: 16),
          label: const Text(
            'Free-to-Play enabled',
            style: TextStyle(
                color: Colors.greenAccent, fontWeight: FontWeight.w600),
          ),
          backgroundColor: Colors.greenAccent.withOpacity(0.15),
        ),
        const SizedBox(width: 12),
        const Expanded(
          child: Text(
            'No escrow needed. Create or join rooms instantly.',
            style: TextStyle(color: Colors.white70),
          ),
        ),
      ],
    );
  }

  String? _createRoomDisabledReason(PongModel model) {
    if (model.currentWR != null) {
      return 'Already in a waiting room';
    }
    if (model.serverIsF2P) {
      return null;
    }
    if (model.betAmt <= 0) {
      return 'Set a bet amount before creating a room';
    }
    if (model.escrowId.isEmpty) {
      return 'Open an escrow using the "Open Escrow" button above';
    }
    if (!model.escrowFunded) {
      return 'Wait for escrow funding before creating a room';
    }
    return null;
  }
}
