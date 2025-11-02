import 'dart:convert';
import 'dart:io';

import 'package:flutter/material.dart';
import 'package:pongui/models/pong.dart';
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

  String _httpBase(PongModel m) {
    final parts = m.cfg.serverAddr.split(":");
    final host = parts.isNotEmpty ? parts[0] : '127.0.0.1';
    int port = 8080;
    if (parts.length > 1) {
      final p = int.tryParse(parts[1]);
      if (p != null) port = p + 1; // http runs on grpc+1
    }
    return "http://$host:$port";
  }

  Future<void> _requestNonce(PongModel m) async {
    setState(() { _loading = true; _status = ''; });
    try {
      final client = HttpClient();
      final req = await client.postUrl(Uri.parse("${_httpBase(m)}/auth/request"));
      req.headers.set(HttpHeaders.contentTypeHeader, 'application/json');
      req.add(utf8.encode(jsonEncode({"user_id": ""})));
      final res = await req.close();
      final body = await res.transform(utf8.decoder).join();
      if (res.statusCode != 200) {
        throw Exception('Request failed: ${res.statusCode} ${res.reasonPhrase}');
      }
      final data = jsonDecode(body) as Map<String, dynamic>;
      setState(() { _nonce = (data['nonce'] as String?) ?? ''; });
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
      final client = HttpClient();
      final req = await client.postUrl(Uri.parse("${_httpBase(m)}/auth/verify"));
      req.headers.set(HttpHeaders.contentTypeHeader, 'application/json');
      req.add(utf8.encode(jsonEncode({
        "address": addr,
        "nonce": nonce,
        "signature": sig,
      })));
      final res = await req.close();
      final body = await res.transform(utf8.decoder).join();
      if (res.statusCode != 200) {
        throw Exception('Verify failed: ${res.statusCode} ${res.reasonPhrase} ${body}');
      }
      final data = jsonDecode(body) as Map<String, dynamic>;
      final ok = (data['ok'] == true);
      if (!ok) {
        throw Exception('Invalid response');
      }
      final token = (data['token'] as String?) ?? '';
      final clientId = (data['client_id'] as String?) ?? '';
      if (clientId.isEmpty) {
        throw Exception('Missing client_id');
      }
      m.applyWalletAuth(newClientId: clientId, token: token);
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


