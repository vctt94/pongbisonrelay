import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/models/pong.dart';
import 'package:provider/provider.dart';

class RefundEscrowsScreen extends StatefulWidget {
  const RefundEscrowsScreen({super.key});

  @override
  State<RefundEscrowsScreen> createState() => _RefundEscrowsScreenState();
}

class _RefundEscrowsScreenState extends State<RefundEscrowsScreen> {
  String? _deletingEscrowId;
  final TextEditingController _confirmationCtrl = TextEditingController();

  @override
  void initState() {
    super.initState();
    WidgetsBinding.instance.addPostFrameCallback((_) {
      if (!mounted) return;
      context.read<PongModel>().loadHistoricEscrows();
    });
  }

  @override
  void dispose() {
    _confirmationCtrl.dispose();
    super.dispose();
  }

  Future<void> _refresh() {
    return context.read<PongModel>().loadHistoricEscrows();
  }

  void _openEscrowDialog(Map<String, dynamic> escrow) {
    showDialog(
      context: context,
      barrierDismissible: false,
      builder: (_) => RefundEscrowDialog(
        escrow: Map<String, dynamic>.from(escrow),
      ),
    );
  }

  Future<void> _confirmDeleteEscrow(String escrowId) async {
    _confirmationCtrl.clear();
    final confirm = await showDialog<bool>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Delete escrow record?'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Are you absolutely sure? Deleting this historic escrow entry '
              'removes the only local record used for refunds. '
              'If the refund has not been recovered yet, the funds may be '
              'PERMANENTLY LOST.',
            ),
            const SizedBox(height: 16),
            TextField(
              controller: _confirmationCtrl,
              decoration: const InputDecoration(
                labelText: 'Type OK to confirm',
              ),
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(false),
            child: const Text('Cancel'),
          ),
          TextButton(
            onPressed: () {
              final ok = _confirmationCtrl.text.trim().toLowerCase() == 'ok';
              Navigator.of(dialogContext).pop(ok);
            },
            child: const Text('Delete'),
          ),
        ],
      ),
    );
    if (confirm != true) {
      return;
    }

    setState(() {
      _deletingEscrowId = escrowId;
    });

    final messenger = ScaffoldMessenger.of(context);
    try {
      await context.read<PongModel>().deleteHistoricEscrow(escrowId);
      if (!mounted) return;
      messenger.showSnackBar(
        const SnackBar(content: Text('Escrow entry deleted.')),
      );
    } catch (e) {
      if (!mounted) return;
      messenger.showSnackBar(
        SnackBar(content: Text('Failed to delete escrow: $e')),
      );
    } finally {
      if (!mounted) return;
      setState(() {
        _deletingEscrowId = null;
      });
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Historic Escrow Sessions'),
        actions: [
          IconButton(
            tooltip: 'Refresh',
            onPressed: () => _refresh(),
            icon: const Icon(Icons.refresh),
          ),
        ],
      ),
      body: Consumer<PongModel>(
        builder: (context, model, _) {
          if (model.isLoadingHistoricEscrows &&
              model.historicEscrows.isEmpty &&
              model.historicEscrowsError.isEmpty) {
            return const Center(child: CircularProgressIndicator());
          }

          if (model.historicEscrowsError.isNotEmpty) {
            return Center(
              child: Padding(
                padding: const EdgeInsets.all(24),
                child: Column(
                  mainAxisSize: MainAxisSize.min,
                  children: [
                    Icon(Icons.error_outline,
                        size: 48, color: Colors.redAccent.shade200),
                    const SizedBox(height: 16),
                    Text(
                      model.historicEscrowsError,
                      textAlign: TextAlign.center,
                      style: const TextStyle(
                          fontSize: 14, color: Colors.redAccent),
                    ),
                    const SizedBox(height: 16),
                    ElevatedButton.icon(
                      onPressed: _refresh,
                      icon: const Icon(Icons.refresh),
                      label: const Text('Try Again'),
                    ),
                  ],
                ),
              ),
            );
          }

          if (model.historicEscrows.isEmpty) {
            return RefreshIndicator(
              onRefresh: _refresh,
              child: ListView(
                padding:
                    const EdgeInsets.symmetric(horizontal: 24, vertical: 48),
                children: const [
                  Icon(Icons.lock_clock, size: 64, color: Colors.white70),
                  SizedBox(height: 16),
                  Text(
                    'No historic escrow sessions found.\n\n'
                    'Refund information will appear here after matches archive '
                    'their settlement session keys.',
                    textAlign: TextAlign.center,
                    style: TextStyle(fontSize: 14, color: Colors.white70),
                  ),
                ],
              ),
            );
          }

          return RefreshIndicator(
            onRefresh: _refresh,
            child: ListView.separated(
              padding: const EdgeInsets.all(16),
              itemCount: model.historicEscrows.length,
              separatorBuilder: (_, __) => const SizedBox(height: 12),
              itemBuilder: (context, index) {
                final escrow = model.historicEscrows[index];
                final escrowId = escrow['escrow_id']?.toString() ?? '';
                final fundingTx = escrow['funding_txid']?.toString() ?? '';
                final fundingVout = _toInt(escrow['funding_vout']);
                final amountAtoms = _toInt(escrow['funded_amount']);
                final amountDcr = amountAtoms / 100000000;
                final csvBlocks = _toInt(escrow['csv_blocks']);
                final archivedAt = _toInt(escrow['archived_at']);
                final isDeleting = _deletingEscrowId == escrowId;
                final status = (escrow['status']?.toString() ?? '').toLowerCase();
                String statusLabel = 'Open';
                Color statusColor = Colors.grey;
                switch (status) {
                  case 'paid':
                    statusLabel = 'Paid';
                    statusColor = Colors.greenAccent;
                    break;
                  case 'tx built':
                  case 'tx_built':
                    statusLabel = 'Tx Built';
                    statusColor = Colors.lightBlueAccent;
                    break;
                  default:
                    statusLabel = 'Open';
                    statusColor = Colors.grey;
                }
                final archivedText = archivedAt > 0
                    ? DateTime.fromMillisecondsSinceEpoch(archivedAt)
                        .toLocal()
                        .toString()
                        .split('.')
                        .first
                    : 'Unknown';
                final needsFunding = fundingTx.isEmpty;

                return Card(
                  color: const Color(0xFF1B1E2C),
                  elevation: 2,
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(14),
                  ),
                  child: InkWell(
                    borderRadius: BorderRadius.circular(14),
                    onTap: () => _openEscrowDialog(escrow),
                    child: Padding(
                      padding: const EdgeInsets.all(16),
                      child: Column(
                        crossAxisAlignment: CrossAxisAlignment.start,
                        children: [
                          Row(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              Expanded(
                                child: Text(
                                  'Escrow ${_shorten(escrowId)}',
                                  style: const TextStyle(
                                    fontSize: 16,
                                    fontWeight: FontWeight.bold,
                                  ),
                                ),
                              ),
                              Wrap(
                                spacing: 8,
                                children: [
                                  Chip(
                                    visualDensity: VisualDensity.compact,
                                    backgroundColor: needsFunding
                                        ? Colors.orange.withOpacity(0.15)
                                        : Colors.green.withOpacity(0.15),
                                    label: Text(
                                      needsFunding
                                          ? 'Needs funding details'
                                          : 'Funding recorded',
                                      style: TextStyle(
                                        fontSize: 12,
                                        color: needsFunding
                                            ? Colors.orangeAccent
                                            : Colors.greenAccent,
                                      ),
                                    ),
                                  ),
                                  Chip(
                                    visualDensity: VisualDensity.compact,
                                    backgroundColor:
                                        statusColor.withOpacity(0.15),
                                    label: Text(
                                      statusLabel,
                                      style: TextStyle(
                                        fontSize: 12,
                                        color: statusColor,
                                      ),
                                    ),
                                  ),
                                ],
                              ),
                            ],
                          ),
                          const SizedBox(height: 12),
                          Text(
                            'Amount: ${amountDcr.toStringAsFixed(8)} DCR',
                            style: const TextStyle(fontSize: 13),
                          ),
                          Text(
                            'CSV blocks: ${csvBlocks > 0 ? csvBlocks : 'unknown'}',
                            style: const TextStyle(fontSize: 13),
                          ),
                          Text(
                            'Archived: $archivedText',
                            style: TextStyle(
                              fontSize: 12,
                              color: Colors.grey.shade400,
                            ),
                          ),
                          const SizedBox(height: 8),
                          if (fundingTx.isNotEmpty)
                            Text(
                              'Funding: ${_shorten(fundingTx, head: 10, tail: 10)}:$fundingVout',
                              style: const TextStyle(fontSize: 12),
                            )
                          else
                            const Text(
                              'Funding transaction not recorded yet',
                              style: TextStyle(
                                fontSize: 12,
                                color: Colors.orangeAccent,
                              ),
                            ),
                          const SizedBox(height: 16),
                          Align(
                            alignment: Alignment.centerRight,
                            child: Row(
                              mainAxisSize: MainAxisSize.min,
                              children: [
                                TextButton.icon(
                                  onPressed: isDeleting
                                      ? null
                                      : () => _confirmDeleteEscrow(escrowId),
                                  style: TextButton.styleFrom(
                                    foregroundColor: Colors.redAccent,
                                  ),
                                  icon: isDeleting
                                      ? const SizedBox(
                                          width: 16,
                                          height: 16,
                                          child: CircularProgressIndicator(
                                            strokeWidth: 2,
                                          ),
                                        )
                                      : const Icon(Icons.delete_outline,
                                          size: 18),
                                  label: Text(
                                      isDeleting ? 'Deleting...' : 'Delete'),
                                ),
                                const SizedBox(width: 8),
                                TextButton.icon(
                                  onPressed: () => _openEscrowDialog(escrow),
                                  icon: const Icon(Icons.open_in_new, size: 18),
                                  label: const Text('Review & Refund'),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
                  ),
                );
              },
            ),
          );
        },
      ),
    );
  }

  static int _toInt(dynamic value) {
    if (value is int) return value;
    if (value is num) return value.toInt();
    if (value is String) return int.tryParse(value) ?? 0;
    return 0;
  }

  static String _shorten(String value, {int head = 6, int tail = 4}) {
    if (value.isEmpty || value.length <= head + tail) {
      return value;
    }
    return '${value.substring(0, head)}...${value.substring(value.length - tail)}';
  }
}

class RefundEscrowDialog extends StatefulWidget {
  const RefundEscrowDialog({super.key, required this.escrow});

  final Map<String, dynamic> escrow;

  @override
  State<RefundEscrowDialog> createState() => _RefundEscrowDialogState();
}

class _RefundEscrowDialogState extends State<RefundEscrowDialog> {
  late Map<String, dynamic> _escrow;
  late TextEditingController _destAddressCtrl;
  late TextEditingController _csvBlocksCtrl;
  late TextEditingController _utxoValueCtrl;

  bool _isBuilding = false;
  bool _isUpdatingFunding = false;
  String? _statusMessage;
  bool _statusIsError = false;
  String? _refundTxHex;
  Map<String, dynamic>? _refundResult;

  @override
  void initState() {
    super.initState();
    _escrow = Map<String, dynamic>.from(widget.escrow);
    final model = context.read<PongModel>();
    final defaultDest = model.walletAddress.isNotEmpty
        ? model.walletAddress
        : model.payoutAddressOrPubkey;
    _destAddressCtrl = TextEditingController(text: defaultDest);
    final csvBlocks = _toInt(_escrow['csv_blocks']);
    _csvBlocksCtrl = TextEditingController(
      text: csvBlocks > 0 ? csvBlocks.toString() : '',
    );
    final storedAmount = _toInt(_escrow['funded_amount']);
    _utxoValueCtrl = TextEditingController(
      text: storedAmount > 0 ? storedAmount.toString() : '',
    );
  }

  @override
  void dispose() {
    _destAddressCtrl.dispose();
    _csvBlocksCtrl.dispose();
    _utxoValueCtrl.dispose();
    super.dispose();
  }

  String get _escrowId => _escrow['escrow_id']?.toString() ?? '';

  Future<void> _handleFundingUpdate() async {
    final currentTxid = _escrow['funding_txid']?.toString() ?? '';
    final currentVout = _toInt(_escrow['funding_vout']);

    final txidController = TextEditingController(text: currentTxid);
    final voutController = TextEditingController(
      text: currentVout >= 0 ? currentVout.toString() : '',
    );

    final result = await showDialog<Map<String, String>>(
      context: context,
      builder: (dialogContext) => AlertDialog(
        title: const Text('Funding Transaction'),
        content: Column(
          mainAxisSize: MainAxisSize.min,
          crossAxisAlignment: CrossAxisAlignment.start,
          children: [
            const Text(
              'Enter the transaction that funded this escrow. '
              'This information is required to build the refund transaction.',
              style: TextStyle(fontSize: 13),
            ),
            const SizedBox(height: 16),
            TextField(
              controller: txidController,
              decoration: const InputDecoration(
                labelText: 'Funding transaction ID',
                hintText:
                    '0000000000000000000000000000000000000000000000000000000000000000',
              ),
              maxLines: 1,
            ),
            const SizedBox(height: 12),
            TextField(
              controller: voutController,
              decoration: const InputDecoration(
                labelText: 'Output index (vout)',
                hintText: '0',
              ),
              keyboardType: TextInputType.number,
            ),
          ],
        ),
        actions: [
          TextButton(
            onPressed: () => Navigator.of(dialogContext).pop(),
            child: const Text('Cancel'),
          ),
          ElevatedButton(
            onPressed: () => Navigator.of(dialogContext).pop({
              'txid': txidController.text.trim(),
              'vout': voutController.text.trim(),
            }),
            child: const Text('Save'),
          ),
        ],
      ),
    );

    if (result == null) {
      return;
    }
    final txid = result['txid']?.trim() ?? '';
    final vout = int.tryParse(result['vout'] ?? '') ?? 0;
    if (txid.isEmpty) {
      setState(() {
        _statusMessage = 'Funding transaction ID is required.';
        _statusIsError = true;
      });
      return;
    }

    setState(() {
      _isUpdatingFunding = true;
      _statusMessage = null;
    });

    try {
      final model = context.read<PongModel>();
      await model.updateEscrowFundingTx(_escrowId, txid, vout);
      await model.loadHistoricEscrows();
      if (!mounted) return;
      setState(() {
        _escrow['funding_txid'] = txid;
        _escrow['funding_vout'] = vout;
        _statusMessage = 'Funding transaction saved.';
        _statusIsError = false;
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _statusMessage = e.toString();
        _statusIsError = true;
      });
    } finally {
      if (!mounted) return;
      setState(() {
        _isUpdatingFunding = false;
      });
    }
  }

  Future<void> _handleBuildRefund() async {
    final dest = _destAddressCtrl.text.trim();
    if (dest.isEmpty) {
      setState(() {
        _statusMessage = 'Destination address or pubkey is required.';
        _statusIsError = true;
      });
      return;
    }

    final csvInput = _csvBlocksCtrl.text.trim();
    final csvBlocks = csvInput.isNotEmpty
        ? int.tryParse(csvInput) ?? _toInt(_escrow['csv_blocks'])
        : _toInt(_escrow['csv_blocks']);

    final utxoValueInput = _utxoValueCtrl.text.trim();
    final utxoValue = utxoValueInput.isNotEmpty
        ? int.tryParse(utxoValueInput)
        : null;

    setState(() {
      _isBuilding = true;
      _statusMessage = 'Building refund transaction...';
      _statusIsError = false;
      _refundTxHex = null;
      _refundResult = null;
    });

    try {
      final model = context.read<PongModel>();
      final result = await model.buildRefundTransaction(
        _escrowId,
        dest,
        csvBlocks: csvBlocks > 0 ? csvBlocks : null,
        utxoValue: utxoValue,
      );
      if (!mounted) return;
      setState(() {
        _refundResult = result;
        if (result['can_refund'] == true) {
          _refundTxHex = result['refund_tx_hex']?.toString();
          _statusMessage = 'Refund transaction built successfully.';
          _statusIsError = false;
          // Update escrow info with latest utxo hints if provided
          if (result['utxo_txid'] != null) {
            _escrow['funding_txid'] = result['utxo_txid'];
          }
          if (result['utxo_vout'] != null) {
            _escrow['funding_vout'] = result['utxo_vout'];
          }
          if (result['utxo_value'] != null) {
            _escrow['funded_amount'] = result['utxo_value'];
          }
          // Mark escrow as having a refund tx built (not broadcast).
          _escrow['status'] = 'tx built';
        } else {
          _refundTxHex = null;
          final reason = result['reason']?.toString();
          _statusMessage = reason?.isNotEmpty == true
              ? 'Cannot refund: $reason'
              : 'Cannot refund this escrow.';
          _statusIsError = true;
        }
      });
      // Persist status change when refund tx is built.
      if (_refundTxHex != null && _refundTxHex!.isNotEmpty) {
        try {
          await Golib.updateHistoricEscrow({
            'escrow_id': _escrowId,
            'status': 'tx built',
          });
          // Also refresh the model's list so the caller screen updates.
          if (!mounted) return;
          await context.read<PongModel>().loadHistoricEscrows();
        } catch (_) {
          // Non-fatal; UI already updated.
        }
      }
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _statusMessage = e.toString();
        _statusIsError = true;
        _refundTxHex = null;
      });
    } finally {
      if (!mounted) return;
      setState(() {
        _isBuilding = false;
      });
    }
  }

  Future<void> _copyRefundTx() async {
    if (_refundTxHex == null || _refundTxHex!.isEmpty) return;
    await Clipboard.setData(ClipboardData(text: _refundTxHex!));
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Refund transaction copied to clipboard')),
    );
  }

  Future<void> _copyFundingTx() async {
    final fundingTx = _escrow['funding_txid']?.toString() ?? '';
    if (fundingTx.isEmpty) return;
    await Clipboard.setData(ClipboardData(text: fundingTx));
    if (!mounted) return;
    ScaffoldMessenger.of(context).showSnackBar(
      const SnackBar(content: Text('Funding transaction ID copied to clipboard')),
    );
  }

  @override
  Widget build(BuildContext context) {
    final fundingTx = _escrow['funding_txid']?.toString() ?? '';
    final fundingVout = _toInt(_escrow['funding_vout']);
    final amountAtoms = _toInt(_escrow['funded_amount']);
    final amountDcr = amountAtoms / 100000000;
    final csvBlocks = _toInt(_escrow['csv_blocks']);
    final archivedAt = _toInt(_escrow['archived_at']);
    final archivedText = archivedAt > 0
        ? DateTime.fromMillisecondsSinceEpoch(archivedAt)
            .toLocal()
            .toString()
            .split('.')
            .first
        : 'Unknown';

    return AlertDialog(
      title: const Text('Refund Escrow'),
      content: SingleChildScrollView(
        child: SizedBox(
          width: MediaQuery.of(context).size.width * 0.6,
          child: Column(
            crossAxisAlignment: CrossAxisAlignment.start,
            mainAxisSize: MainAxisSize.min,
            children: [
              SelectableText(
                'Escrow ID: $_escrowId',
                style:
                    const TextStyle(fontWeight: FontWeight.bold, fontSize: 14),
              ),
              const SizedBox(height: 12),
              Padding(
                padding: const EdgeInsets.symmetric(vertical: 4),
                child: Row(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    SizedBox(
                      width: 120,
                      child: Text(
                        'Funding',
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.grey.shade400,
                        ),
                      ),
                    ),
                    Expanded(
                      child: Row(
                        children: [
                          Expanded(
                            child: Text(
                              fundingTx.isNotEmpty
                                  ? '${_shorten(fundingTx, head: 12, tail: 12)}:${fundingVout >= 0 ? fundingVout : 0}'
                                  : 'Not recorded',
                              style: TextStyle(
                                fontSize: 12,
                                color: fundingTx.isEmpty ? Colors.orangeAccent : null,
                                fontStyle:
                                    fundingTx.isEmpty ? FontStyle.italic : FontStyle.normal,
                              ),
                            ),
                          ),
                          if (fundingTx.isNotEmpty)
                            IconButton(
                              icon: const Icon(Icons.copy, size: 16),
                              onPressed: _copyFundingTx,
                              tooltip: 'Copy funding transaction ID',
                              padding: EdgeInsets.zero,
                              constraints: const BoxConstraints(
                                minWidth: 24,
                                minHeight: 24,
                              ),
                              color: Colors.grey.shade400,
                            ),
                        ],
                      ),
                    ),
                  ],
                ),
              ),
              _InfoRow(
                label: 'Amount',
                value: amountAtoms > 0
                    ? '${amountDcr.toStringAsFixed(8)} DCR'
                    : 'Unknown',
              ),
              _InfoRow(
                label: 'CSV blocks',
                value: csvBlocks > 0 ? csvBlocks.toString() : 'Unknown',
              ),
              _InfoRow(
                label: 'Archived',
                value: archivedText,
              ),
              const SizedBox(height: 16),
              TextField(
                controller: _destAddressCtrl,
                decoration: const InputDecoration(
                  labelText: 'Refund destination (address or pubkey)',
                  hintText: 'Destination to receive the refund',
                ),
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _csvBlocksCtrl,
                decoration: InputDecoration(
                  labelText: 'CSV blocks override',
                  hintText: csvBlocks > 0 ? csvBlocks.toString() : 'e.g. 2',
                  helperText:
                      'Optional. Leave empty to use the stored CSV timelock.',
                ),
                keyboardType: TextInputType.number,
              ),
              const SizedBox(height: 12),
              TextField(
                controller: _utxoValueCtrl,
                decoration: const InputDecoration(
                  labelText: 'UTXO value (atoms)',
                  helperText: 'optional in case of wrong input',
                ),
                keyboardType: TextInputType.number,
              ),
              const SizedBox(height: 20),
              Wrap(
                spacing: 12,
                runSpacing: 12,
                children: [
                  OutlinedButton.icon(
                    onPressed: _isBuilding ? null : _handleFundingUpdate,
                    icon: _isUpdatingFunding
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(strokeWidth: 2),
                          )
                        : const Icon(Icons.edit),
                    label: Text(
                      fundingTx.isEmpty
                          ? 'Record funding transaction'
                          : 'Edit funding transaction',
                    ),
                  ),
                  ElevatedButton.icon(
                    onPressed: (_isBuilding || _isUpdatingFunding)
                        ? null
                        : _handleBuildRefund,
                    icon: _isBuilding
                        ? const SizedBox(
                            width: 16,
                            height: 16,
                            child: CircularProgressIndicator(
                              strokeWidth: 2,
                              color: Colors.white,
                            ),
                          )
                        : const Icon(Icons.currency_exchange),
                    label: Text(
                      _isBuilding ? 'Building...' : 'Build refund transaction',
                    ),
                  ),
                ],
              ),
              const SizedBox(height: 16),
              if (_statusMessage != null)
                _StatusBanner(
                  message: _statusMessage!,
                  isError: _statusIsError,
                ),
              if (_refundResult != null && _refundResult!['utxo_txid'] != null)
                Padding(
                  padding: const EdgeInsets.only(top: 12),
                  child: _InfoRow(
                    label: 'Refund UTXO',
                    value:
                        '${_shorten(_refundResult!['utxo_txid'].toString(), head: 12, tail: 12)}:${_refundResult!['utxo_vout']}',
                  ),
                ),
              if (_refundTxHex != null && _refundTxHex!.isNotEmpty) ...[
                const SizedBox(height: 16),
                const Text(
                  'Refund transaction (hex)',
                  style: TextStyle(fontWeight: FontWeight.bold),
                ),
                const SizedBox(height: 8),
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.black.withOpacity(0.35),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.grey.shade700),
                  ),
                  child: SelectableText(
                    _refundTxHex!,
                    style: const TextStyle(
                      fontFamily: 'monospace',
                      fontSize: 12,
                    ),
                    maxLines: 6,
                  ),
                ),
                const SizedBox(height: 8),
                const SizedBox(height: 12),
                Container(
                  padding: const EdgeInsets.all(12),
                  decoration: BoxDecoration(
                    color: Colors.blue.withOpacity(0.1),
                    borderRadius: BorderRadius.circular(8),
                    border: Border.all(color: Colors.blue.withOpacity(0.3)),
                  ),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Text(
                        'To rebroadcast this transaction, visit dcrdata:',
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.blue.shade200,
                          fontWeight: FontWeight.w500,
                        ),
                      ),
                      const SizedBox(height: 8),
                      InkWell(
                        onTap: () async {
                          const url = 'https://dcrdata.org/decodetx';
                          try {
                            if (Platform.isWindows) {
                              await Process.run('start', [url], runInShell: true);
                            } else if (Platform.isMacOS) {
                              await Process.run('open', [url]);
                            } else if (Platform.isLinux) {
                              await Process.run('xdg-open', [url]);
                            } else {
                              // Fallback: copy to clipboard
                              await Clipboard.setData(ClipboardData(text: url));
                              if (!mounted) return;
                              ScaffoldMessenger.of(context).showSnackBar(
                                const SnackBar(
                                  content: Text('URL copied to clipboard'),
                                ),
                              );
                            }
                          } catch (e) {
                            // Fallback: copy to clipboard if opening fails
                            await Clipboard.setData(ClipboardData(text: url));
                            if (!mounted) return;
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(
                                content: Text('URL copied to clipboard: $url'),
                              ),
                            );
                          }
                        },
                        child: Text(
                          'https://dcrdata.org/decodetx',
                          style: TextStyle(
                            fontSize: 12,
                            color: Colors.blue.shade300,
                            decoration: TextDecoration.underline,
                          ),
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'Paste the transaction hex above into the "Broadcast Tx" field on dcrdata to rebroadcast it to the network.',
                        style: TextStyle(
                          fontSize: 11,
                          color: Colors.grey.shade400,
                          fontStyle: FontStyle.italic,
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
      actions: [
        if (_refundTxHex != null && _refundTxHex!.isNotEmpty)
          TextButton.icon(
            onPressed: _copyRefundTx,
            icon: const Icon(Icons.copy, size: 18),
            label: const Text('Copy transaction'),
          ),
        TextButton(
          onPressed: () => Navigator.of(context).pop(),
          child: const Text('Close'),
        ),
      ],
    );
  }

  static int _toInt(dynamic value) {
    if (value is int) return value;
    if (value is num) return value.toInt();
    if (value is String) return int.tryParse(value) ?? 0;
    return 0;
  }

  static String _shorten(String value, {int head = 6, int tail = 4}) {
    if (value.isEmpty || value.length <= head + tail) {
      return value;
    }
    return '${value.substring(0, head)}...${value.substring(value.length - tail)}';
  }
}

class _InfoRow extends StatelessWidget {
  const _InfoRow({
    required this.label,
    required this.value,
    this.valueStyle,
  });

  final String label;
  final String value;
  final TextStyle? valueStyle;

  @override
  Widget build(BuildContext context) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 4),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          SizedBox(
            width: 120,
            child: Text(
              label,
              style: TextStyle(
                fontSize: 12,
                color: Colors.grey.shade400,
              ),
            ),
          ),
          Expanded(
            child: Text(
              value,
              style: valueStyle ??
                  const TextStyle(
                    fontSize: 12,
                  ),
            ),
          ),
        ],
      ),
    );
  }
}

class _StatusBanner extends StatelessWidget {
  const _StatusBanner({required this.message, required this.isError});

  final String message;
  final bool isError;

  @override
  Widget build(BuildContext context) {
    final color = isError ? Colors.redAccent : Colors.greenAccent;
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(12),
      decoration: BoxDecoration(
        color: color.withOpacity(0.15),
        borderRadius: BorderRadius.circular(8),
        border: Border.all(color: color),
      ),
      child: Row(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          Icon(
            isError ? Icons.error : Icons.check_circle,
            color: color,
            size: 20,
          ),
          const SizedBox(width: 8),
          Expanded(
            child: Text(
              message,
              style: TextStyle(color: color, fontSize: 12),
            ),
          ),
        ],
      ),
    );
  }
}
