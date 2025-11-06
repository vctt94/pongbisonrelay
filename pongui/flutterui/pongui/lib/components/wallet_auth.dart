import 'package:flutter/material.dart';
import 'package:pongui/models/pong.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:provider/provider.dart';

class WalletAuthDialog extends StatefulWidget {
  const WalletAuthDialog({super.key});

  @override
  State<WalletAuthDialog> createState() => _WalletAuthDialogState();
}

class _WalletAuthDialogState extends State<WalletAuthDialog> {
  final TextEditingController _addrCtrl = TextEditingController();
  final TextEditingController _sigCtrl = TextEditingController();
  String _nonce = '';
  String _status = '';
  bool _loading = false;


  Future<void> _requestNonce(PongModel m) async {
    setState(() { _loading = true; _status = ''; });
    try {
      final n = await Golib.requestNonce(m.cfg.serverAddr, m.cfg.grpcCertPath);
      setState(() { _nonce = n; });
    } catch (e) {
      setState(() { _status = 'Error: $e'; });
    } finally {
      setState(() { _loading = false; });
    }
  }

  Future<void> _verify(PongModel m) async {
    final addr = _addrCtrl.text.trim();
    final sig = _sigCtrl.text.trim();
    final nonce = _nonce.trim();
    if (addr.isEmpty || sig.isEmpty || nonce.isEmpty) {
      setState(() { _status = 'Fill address, request code, and paste signature'; });
      return;
    }
    setState(() { _loading = true; _status = ''; });
    try {
      final res = await Golib.verifyLogin(
        m.cfg.serverAddr,
        m.cfg.grpcCertPath,
        addr,
        nonce,
        sig,
      );
      if (!(res['ok'] == true)) {
        throw Exception('Invalid response');
      }
      final token = (res['token'] ?? '').toString();
      final clientId = (res['client_id'] ?? '').toString();
      if (clientId.isEmpty) {
        throw Exception('Missing client_id');
      }
      // Apply new identity and store payout address from recovered P2PK.
      // Pass p2pkAddr directly to applyWalletAuth so it's set before _initPongClient runs
      m.applyWalletAuth(
        newClientId: clientId,
        token: token,
        address: addr,
        p2pkAddr: (res['p2pk_addr'] ?? '').toString(),
      );
      if (!mounted) return;
      Navigator.of(context).pop();
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Wallet linked. You can play now.')),
      );
    } catch (e) {
      setState(() { _status = 'Error: $e'; });
    } finally {
      setState(() { _loading = false; });
    }
  }

  @override
  Widget build(BuildContext context) {
    final model = context.watch<PongModel>();
    return AlertDialog(
      title: const Text('Login with Decred Wallet'),
      content: SingleChildScrollView(
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            const Text('1) Paste a P2PKH address (D... or T...)'),
            const SizedBox(height: 8),
            TextField(
              controller: _addrCtrl,
              decoration: const InputDecoration(
                labelText: 'Address',
              ),
            ),
            const SizedBox(height: 12),
            Row(
              children: [
                ElevatedButton(
                  onPressed: _loading ? null : () => _requestNonce(model),
                  child: const Text('Request Code'),
                ),
                const SizedBox(width: 8),
                if (_nonce.isNotEmpty) Expanded(
                  child: SelectableText(_nonce),
                ),
              ],
            ),
            if (_nonce.isNotEmpty) ...[
              const SizedBox(height: 12),
              const Text('2) Sign this code in your wallet and paste the signature'),
              TextField(
                controller: _sigCtrl,
                minLines: 1,
                maxLines: 4,
                decoration: const InputDecoration(
                  labelText: 'Base64 Signature',
                ),
              ),
            ],
            if (_status.isNotEmpty) ...[
              const SizedBox(height: 8),
              Text(_status, style: const TextStyle(color: Colors.redAccent)),
            ],
          ],
        ),
      ),
      actions: [
        TextButton(
          onPressed: _loading ? null : () => Navigator.of(context).pop(),
          child: const Text('Cancel'),
        ),
        ElevatedButton(
          onPressed: (!_loading && _nonce.isNotEmpty) ? () => _verify(model) : null,
          child: const Text('Verify & Login'),
        ),
      ],
    );
  }
}


