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
                                          if (pongModel.escrowDepositAddress
                                                  .isNotEmpty &&
                                              pongModel.escrowInfoPersisted &&
                                              pongModel
                                                  .escrowRefundSessionValid)
                                            Tooltip(
                                              message: pongModel
                                                  .escrowDepositAddress,
                                              child: InkWell(
                                                onTap: () async {
                                                  await Clipboard.setData(
                                                    ClipboardData(
                                                      text: pongModel
                                                          .escrowDepositAddress,
                                                    ),
                                                  );
                                                  if (!context.mounted) return;
                                                  ScaffoldMessenger.of(context)
                                                      .showSnackBar(
                                                    const SnackBar(
                                                      content: Text(
                                                          'Deposit address copied'),
                                                    ),
                                                  );
                                                },
                                                borderRadius:
                                                    BorderRadius.circular(8),
                                                child: Container(
                                                  padding: const EdgeInsets
                                                      .symmetric(
                                                      horizontal: 10,
                                                      vertical: 6),
                                                  decoration: BoxDecoration(
                                                    color: Colors.amber
                                                        .withOpacity(0.1),
                                                    borderRadius:
                                                        BorderRadius.circular(
                                                            8),
                                                    border: Border.all(
                                                      color: Colors.amberAccent
                                                          .withOpacity(0.6),
                                                      width: 1,
                                                    ),
                                                  ),
                                                  child: Row(
                                                    mainAxisSize:
                                                        MainAxisSize.min,
                                                    children: [
                                                      const Icon(
                                                        Icons
                                                            .account_balance_wallet,
                                                        color:
                                                            Colors.amberAccent,
                                                        size: 16,
                                                      ),
                                                      const SizedBox(width: 6),
                                                      Text(
                                                        '${pongModel.escrowDepositAddress.substring(0, 8)}...${pongModel.escrowDepositAddress.substring(pongModel.escrowDepositAddress.length - 6)}',
                                                        style: const TextStyle(
                                                          color: Colors
                                                              .amberAccent,
                                                          fontFamily:
                                                              'monospace',
                                                        ),
                                                      ),
                                                    ],
                                                  ),
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
                                                      depositAddress: dep,
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
                                        ],
                                      ),
                                    ),
                                  ],
                                ),
                                if (!pongModel.serverIsF2P) ...[
                                  const SizedBox(height: 8),
                                  _buildSettlementStatusRow(pongModel),
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

                    // 2) Escrow information (shown when escrow is opened and
                    // not currently in a waiting room)
                    if (!pongModel.serverIsF2P &&
                        pongModel.escrowId.isNotEmpty &&
                        pongModel.currentWR == null)
                      Center(
                        child: Container(
                          width: MediaQuery.of(context).size.width * 0.85,
                          margin: const EdgeInsets.only(top: 16.0),
                          child: Card(
                            elevation: 2,
                            color: const Color(0xFF1B1E2C),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(16),
                              side: BorderSide(
                                color: Colors.amber.withOpacity(0.3),
                                width: 1.5,
                              ),
                            ),
                            child: Container(
                              decoration: BoxDecoration(
                                borderRadius: BorderRadius.circular(16),
                                gradient: LinearGradient(
                                  begin: Alignment.topLeft,
                                  end: Alignment.bottomRight,
                                  colors: [
                                    Colors.amber.withOpacity(0.05),
                                    Colors.transparent,
                                  ],
                                ),
                              ),
                              child: Padding(
                                padding: const EdgeInsets.all(20.0),
                                child: Column(
                                  crossAxisAlignment: CrossAxisAlignment.start,
                                  children: [
                                    // Escrow ID and Status badge in same row
                                    Row(
                                      mainAxisAlignment:
                                          MainAxisAlignment.spaceBetween,
                                      crossAxisAlignment:
                                          CrossAxisAlignment.start,
                                      children: [
                                        // Escrow ID section - simplified
                                        Expanded(
                                          child: Column(
                                            crossAxisAlignment:
                                                CrossAxisAlignment.start,
                                            children: [
                                              const Text(
                                                'Escrow ID',
                                                style: TextStyle(
                                                  color: Colors.white70,
                                                  fontSize: 13,
                                                  fontWeight: FontWeight.w500,
                                                ),
                                              ),
                                              const SizedBox(height: 4),
                                              SelectableText(
                                                pongModel.escrowId,
                                                style: const TextStyle(
                                                  color: Colors.white,
                                                  fontFamily: 'monospace',
                                                  fontSize: 13,
                                                ),
                                              ),
                                            ],
                                          ),
                                        ),
                                        // Status badge
                                        Container(
                                          padding: const EdgeInsets.symmetric(
                                              horizontal: 12, vertical: 6),
                                          decoration: BoxDecoration(
                                            color: pongModel.escrowFunded
                                                ? Colors.greenAccent
                                                    .withOpacity(0.2)
                                                : Colors.amberAccent
                                                    .withOpacity(0.2),
                                            borderRadius:
                                                BorderRadius.circular(20),
                                            border: Border.all(
                                              color: pongModel.escrowFunded
                                                  ? Colors.greenAccent
                                                  : Colors.amberAccent,
                                              width: 1,
                                            ),
                                          ),
                                          child: Row(
                                            mainAxisSize: MainAxisSize.min,
                                            children: [
                                              Icon(
                                                pongModel.escrowFunded
                                                    ? Icons.check_circle
                                                    : Icons.pending,
                                                size: 14,
                                                color: pongModel.escrowFunded
                                                    ? Colors.greenAccent
                                                    : Colors.amberAccent,
                                              ),
                                              const SizedBox(width: 4),
                                              Text(
                                                pongModel.escrowFunded
                                                    ? 'Funded'
                                                    : 'Pending',
                                                style: TextStyle(
                                                  color: pongModel.escrowFunded
                                                      ? Colors.greenAccent
                                                      : Colors.amberAccent,
                                                  fontSize: 12,
                                                  fontWeight: FontWeight.w600,
                                                ),
                                              ),
                                            ],
                                          ),
                                        ),
                                      ],
                                    ),
                                    const SizedBox(height: 16),
                                    // Deposit Address section
                                    if (pongModel
                                            .escrowDepositAddress.isNotEmpty &&
                                        pongModel.escrowInfoPersisted &&
                                        pongModel.escrowRefundSessionValid) ...[
                                      Container(
                                        padding: const EdgeInsets.all(16),
                                        decoration: BoxDecoration(
                                          gradient: LinearGradient(
                                            begin: Alignment.topLeft,
                                            end: Alignment.bottomRight,
                                            colors: [
                                              Colors.amber.withOpacity(0.15),
                                              Colors.amber.withOpacity(0.05),
                                            ],
                                          ),
                                          borderRadius:
                                              BorderRadius.circular(12),
                                          border: Border.all(
                                            color: Colors.amberAccent
                                                .withOpacity(0.4),
                                            width: 1.5,
                                          ),
                                        ),
                                        child: Column(
                                          crossAxisAlignment:
                                              CrossAxisAlignment.start,
                                          children: [
                                            Row(
                                              children: [
                                                const Icon(
                                                    Icons
                                                        .account_balance_wallet,
                                                    color: Colors.amber,
                                                    size: 20),
                                                const SizedBox(width: 8),
                                                const Text(
                                                  'Deposit Address',
                                                  style: TextStyle(
                                                    color: Colors.amberAccent,
                                                    fontSize: 14,
                                                    fontWeight: FontWeight.bold,
                                                    letterSpacing: 0.3,
                                                  ),
                                                ),
                                              ],
                                            ),
                                            const SizedBox(height: 12),
                                            Row(
                                              children: [
                                                Expanded(
                                                  child: SelectableText(
                                                    pongModel
                                                        .escrowDepositAddress,
                                                    style: const TextStyle(
                                                      color: Colors.white,
                                                      fontFamily: 'monospace',
                                                      fontSize: 13,
                                                      letterSpacing: 0.5,
                                                      fontWeight:
                                                          FontWeight.w500,
                                                    ),
                                                  ),
                                                ),
                                                Material(
                                                  color: Colors.transparent,
                                                  child: InkWell(
                                                    onTap: () async {
                                                      await Clipboard.setData(
                                                          ClipboardData(
                                                              text: pongModel
                                                                  .escrowDepositAddress));
                                                      if (!context.mounted)
                                                        return;
                                                      ScaffoldMessenger.of(
                                                              context)
                                                          .showSnackBar(
                                                        const SnackBar(
                                                            content: Text(
                                                                'Address copied'),
                                                            duration: Duration(
                                                                seconds: 2)),
                                                      );
                                                    },
                                                    borderRadius:
                                                        BorderRadius.circular(
                                                            8),
                                                    child: Container(
                                                      padding:
                                                          const EdgeInsets.all(
                                                              8),
                                                      decoration: BoxDecoration(
                                                        color: Colors.amber
                                                            .withOpacity(0.2),
                                                        borderRadius:
                                                            BorderRadius
                                                                .circular(8),
                                                      ),
                                                      child: const Icon(
                                                          Icons.copy,
                                                          color: Colors.amber,
                                                          size: 18),
                                                    ),
                                                  ),
                                                ),
                                              ],
                                            ),
                                            const SizedBox(height: 14),
                                            Container(
                                              padding: const EdgeInsets.all(12),
                                              decoration: BoxDecoration(
                                                color: Colors.amberAccent
                                                    .withOpacity(0.15),
                                                borderRadius:
                                                    BorderRadius.circular(8),
                                                border: Border.all(
                                                  color: Colors.amberAccent
                                                      .withOpacity(0.3),
                                                  width: 1,
                                                ),
                                              ),
                                              child: Row(
                                                crossAxisAlignment:
                                                    CrossAxisAlignment.start,
                                                children: [
                                                  const Icon(
                                                      Icons
                                                          .warning_amber_rounded,
                                                      color: Colors.amberAccent,
                                                      size: 18),
                                                  const SizedBox(width: 10),
                                                  Expanded(
                                                    child: Text(
                                                      'Deposit exactly ${(pongModel.betAmt / 1e8).toStringAsFixed(2)} DCR. Default bet amount.',
                                                      style: const TextStyle(
                                                        color:
                                                            Colors.amberAccent,
                                                        fontSize: 12,
                                                        fontWeight:
                                                            FontWeight.w600,
                                                        height: 1.4,
                                                      ),
                                                    ),
                                                  ),
                                                ],
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                      const SizedBox(height: 16),
                                    ],
                                    // Funding status
                                    Container(
                                      padding: const EdgeInsets.symmetric(
                                          horizontal: 14, vertical: 12),
                                      decoration: BoxDecoration(
                                        color: pongModel.escrowFunded
                                            ? Colors.greenAccent
                                                .withOpacity(0.1)
                                            : Colors.amberAccent
                                                .withOpacity(0.1),
                                        borderRadius: BorderRadius.circular(10),
                                        border: Border.all(
                                          color: pongModel.escrowFunded
                                              ? Colors.greenAccent
                                                  .withOpacity(0.3)
                                              : Colors.amberAccent
                                                  .withOpacity(0.3),
                                          width: 1,
                                        ),
                                      ),
                                      child: Row(
                                        children: [
                                          Icon(
                                            pongModel.escrowFunded
                                                ? Icons.check_circle
                                                : Icons.pending_outlined,
                                            color: pongModel.escrowFunded
                                                ? Colors.greenAccent
                                                : Colors.amberAccent,
                                            size: 20,
                                          ),
                                          const SizedBox(width: 10),
                                          Expanded(
                                            child: Text(
                                              pongModel.escrowFunded
                                                  ? (pongModel.fundingStatus
                                                          .isNotEmpty
                                                      ? pongModel.fundingStatus
                                                      : (pongModel
                                                              .escrowConfirmed
                                                          ? 'Deposit confirmed (${pongModel.escrowConfs} confirmations)'
                                                          : 'Deposit seen in mempool'))
                                                  : 'Waiting for deposit...',
                                              style: TextStyle(
                                                color: pongModel.escrowFunded
                                                    ? Colors.greenAccent
                                                    : Colors.amberAccent,
                                                fontSize: 13,
                                                fontWeight: FontWeight.w600,
                                              ),
                                            ),
                                          ),
                                        ],
                                      ),
                                    ),
                                    // Error messages
                                    if (pongModel
                                        .escrowInfoError.isNotEmpty) ...[
                                      const SizedBox(height: 12),
                                      Container(
                                        padding: const EdgeInsets.all(12),
                                        decoration: BoxDecoration(
                                          color: Colors.redAccent
                                              .withOpacity(0.15),
                                          borderRadius:
                                              BorderRadius.circular(10),
                                          border: Border.all(
                                            color: Colors.redAccent
                                                .withOpacity(0.4),
                                            width: 1.5,
                                          ),
                                        ),
                                        child: Row(
                                          crossAxisAlignment:
                                              CrossAxisAlignment.start,
                                          children: [
                                            const Icon(Icons.error_outline,
                                                color: Colors.redAccent,
                                                size: 20),
                                            const SizedBox(width: 10),
                                            Expanded(
                                              child: Text(
                                                pongModel.escrowInfoError,
                                                style: const TextStyle(
                                                  color: Colors.redAccent,
                                                  fontSize: 12,
                                                  fontWeight: FontWeight.w500,
                                                  height: 1.4,
                                                ),
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                    if (pongModel.escrowRefundSessionError
                                        .isNotEmpty) ...[
                                      const SizedBox(height: 12),
                                      Container(
                                        padding: const EdgeInsets.all(12),
                                        decoration: BoxDecoration(
                                          color: Colors.redAccent
                                              .withOpacity(0.15),
                                          borderRadius:
                                              BorderRadius.circular(10),
                                          border: Border.all(
                                            color: Colors.redAccent
                                                .withOpacity(0.4),
                                            width: 1.5,
                                          ),
                                        ),
                                        child: Row(
                                          crossAxisAlignment:
                                              CrossAxisAlignment.start,
                                          children: [
                                            const Icon(Icons.error_outline,
                                                color: Colors.redAccent,
                                                size: 20),
                                            const SizedBox(width: 10),
                                            Expanded(
                                              child: Text(
                                                pongModel
                                                    .escrowRefundSessionError,
                                                style: const TextStyle(
                                                  color: Colors.redAccent,
                                                  fontSize: 12,
                                                  fontWeight: FontWeight.w500,
                                                  height: 1.4,
                                                ),
                                              ),
                                            ),
                                          ],
                                        ),
                                      ),
                                    ],
                                  ],
                                ),
                              ),
                            ),
                          ),
                        ),
                      ),

                    // 3) Current waiting room info
                    if (pongModel.currentWR != null)
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
                                  const Text(
                                    "Current Waiting Room",
                                    style: TextStyle(
                                      color: Colors.white,
                                      fontSize: 18,
                                      fontWeight: FontWeight.bold,
                                    ),
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
                                          // Room ID and Not Ready side by side
                                          Row(
                                            mainAxisAlignment:
                                                MainAxisAlignment.spaceBetween,
                                            children: [
                                              Row(
                                                children: [
                                                  const Icon(Icons.tag,
                                                      size: 16,
                                                      color: Colors.white54),
                                                  const SizedBox(width: 4),
                                                  Text(
                                                    "Room ID: ${pongModel.currentWR?.id ?? ""}",
                                                    style: const TextStyle(
                                                      color: Colors.white,
                                                    ),
                                                  ),
                                                ],
                                              ),
                                              Row(
                                                children: [
                                                  Icon(
                                                    pongModel.isReady
                                                        ? Icons.check_circle
                                                        : Icons.cancel,
                                                    size: 16,
                                                    color: pongModel.isReady
                                                        ? Colors.green
                                                        : Colors.white70,
                                                  ),
                                                  const SizedBox(width: 4),
                                                  Text(
                                                    pongModel.isReady
                                                        ? "Ready"
                                                        : "Not Ready",
                                                    style: TextStyle(
                                                      color: pongModel.isReady
                                                          ? Colors.green
                                                          : Colors.white70,
                                                      fontWeight: pongModel
                                                              .isReady
                                                          ? FontWeight.bold
                                                          : FontWeight.normal,
                                                    ),
                                                  ),
                                                ],
                                              ),
                                            ],
                                          ),
                                          const SizedBox(height: 8),
                                          // Players and Ready count side by side
                                          Row(
                                            mainAxisAlignment:
                                                MainAxisAlignment.spaceBetween,
                                            children: [
                                              Row(
                                                children: [
                                                  const Icon(Icons.people,
                                                      size: 16,
                                                      color: Colors
                                                          .lightBlueAccent),
                                                  const SizedBox(width: 4),
                                                  Text(
                                                    "Players: $totalPlayers / 2",
                                                    style: const TextStyle(
                                                      color: Colors.white,
                                                    ),
                                                  ),
                                                ],
                                              ),
                                              Row(
                                                children: [
                                                  const Icon(Icons.check_circle,
                                                      size: 16,
                                                      color:
                                                          Colors.greenAccent),
                                                  const SizedBox(width: 4),
                                                  Text(
                                                    "Ready: $readyPlayers / 2",
                                                    style: TextStyle(
                                                      color: readyPlayers == 2
                                                          ? Colors.greenAccent
                                                          : Colors.white70,
                                                    ),
                                                  ),
                                                ],
                                              ),
                                            ],
                                          ),
                                        ],
                                      );
                                    },
                                  ),

                                  // Add ready/leave buttons if in a room
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
                              ),
                            ),
                          ),
                        ),
                      ),

                    // 4) Error message if exists
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

                    // 5) Main content
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

  Widget _buildSettlementStatusRow(PongModel model) {
    if (model.serverIsF2P) {
      return const SizedBox.shrink();
    }

    String label;
    IconData icon;
    Color color;

    if (model.escrowId.isEmpty) {
      label = 'No active escrow. Open escrow to create or join waiting rooms';
      icon = Icons.savings_outlined;
      color = Colors.white70;
    } else if (!model.escrowFunded) {
      label = 'Escrow created. Waiting for your deposit to be seen on-chain.';
      icon = Icons.savings_outlined;
      color = Colors.white70;
    } else if (!model.escrowConfirmed) {
      label =
          'Deposit seen in mempool. Waiting for 1 confirmation before preparing settlement (presign).';
      icon = Icons.schedule;
      color = Colors.amberAccent;
    } else {
      if (!model.escrowRefundSessionValid) {
        label =
            'Escrow backup validation failed. Do not deposit more funds until this is fixed.';
        icon = Icons.error_outline;
        color = Colors.redAccent;
      } else if (model.presignInProgress) {
        label = 'Settlement presign in progress…';
        icon = Icons.shield_outlined;
        color = Colors.lightBlueAccent;
      } else if (model.presignCompleted) {
        label = 'Settlement prepared. Your payout is locked in for this match.';
        icon = Icons.verified_user;
        color = Colors.greenAccent;
      } else if (model.presignError.isNotEmpty) {
        label =
            'Presign error detected. We will retry automatically when conditions are met.';
        icon = Icons.error_outline;
        color = Colors.redAccent;
      } else if (model.currentWR == null) {
        label =
            'Escrow confirmed. Join or create a waiting room so we can pre-sign settlement.';
        icon = Icons.meeting_room_outlined;
        color = Colors.white70;
      } else if (model.currentWR!.players.length < 2) {
        label =
            'Escrow confirmed. Waiting for an opponent before we pre-sign settlement.';
        icon = Icons.people_alt_outlined;
        color = Colors.white70;
      } else {
        label = 'Escrow confirmed. Preparing settlement after players are ready.';
        icon = Icons.shield_outlined;
        color = Colors.lightBlueAccent;
      }
    }

    return Row(
      crossAxisAlignment: CrossAxisAlignment.center,
      children: [
        Icon(icon, color: color, size: 18),
        const SizedBox(width: 8),
        Expanded(
          child: Text(
            label,
            style: TextStyle(
              color: color,
              fontSize: 12,
            ),
          ),
        ),
      ],
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
