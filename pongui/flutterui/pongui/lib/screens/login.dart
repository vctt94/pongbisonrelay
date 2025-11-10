import 'package:flutter/material.dart';
import 'package:pongui/models/pong.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:provider/provider.dart';
import 'package:pongui/components/refund_dialog.dart';

class LoginScreen extends StatefulWidget {
  const LoginScreen({super.key});

  @override
  State<LoginScreen> createState() => _LoginScreenState();
}

class _LoginScreenState extends State<LoginScreen> {
  final TextEditingController _addrCtrl = TextEditingController();
  final TextEditingController _sigCtrl = TextEditingController();
  String _nonce = '';
  String _status = '';
  bool _loading = false;
  bool _preInitDone = false;

  @override
  void dispose() {
    _addrCtrl.dispose();
    _sigCtrl.dispose();
    super.dispose();
  }

  @override
  void initState() {
    super.initState();
    // Pre-login minimal init so CT* commands have a valid handle.
    WidgetsBinding.instance.addPostFrameCallback((_) async {
      if (_preInitDone) return;
      _preInitDone = true;
      try {
        final model = context.read<PongModel>();
        await model.ensurePreloginInitialized();
      } catch (_) {
        // Best-effort preinit; errors are non-fatal for the login UI.
      }
    });
  }

  Future<void> _requestNonce(PongModel m) async {
    setState(() {
      _loading = true;
      _status = '';
    });
    try {
      final n = await Golib.requestNonce(m.cfg.serverAddr, m.cfg.grpcCertPath);
      setState(() {
        _nonce = n;
      });
    } catch (e) {
      setState(() {
        _status = 'Error: $e';
      });
    } finally {
      setState(() {
        _loading = false;
      });
    }
  }

  Future<void> _verify(PongModel m) async {
    final addr = _addrCtrl.text.trim();
    final sig = _sigCtrl.text.trim();
    final nonce = _nonce.trim();
    if (addr.isEmpty || sig.isEmpty || nonce.isEmpty) {
      setState(() {
        _status = 'Fill address, request code, and paste signature';
      });
      return;
    }
    setState(() {
      _loading = true;
      _status = '';
    });
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
      // Navigate to home screen after successful login
      Navigator.of(context).pushReplacementNamed('/home');
    } catch (e) {
      setState(() {
        _status = 'Login failed: $e';
      });
    } finally {
      setState(() {
        _loading = false;
      });
    }
  }

  void _openRefundScreen() {
    Navigator.of(context).push(
      MaterialPageRoute(
        builder: (_) => const RefundEscrowsScreen(),
      ),
    );
  }

  @override
  Widget build(BuildContext context) {
    final model = context.watch<PongModel>();

    return Scaffold(
      body: Container(
        decoration: const BoxDecoration(
          gradient: LinearGradient(
            begin: Alignment.topLeft,
            end: Alignment.bottomRight,
            colors: [
              Color.fromARGB(255, 25, 23, 44),
              Color.fromARGB(255, 35, 33, 54),
            ],
          ),
        ),
        child: Center(
          child: SingleChildScrollView(
            child: Container(
              constraints: const BoxConstraints(maxWidth: 500),
              margin: const EdgeInsets.all(24),
              child: Card(
                color: const Color(0xFF1B1E2C),
                elevation: 8,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(16),
                ),
                child: Padding(
                  padding: const EdgeInsets.all(32.0),
                  child: Column(
                    mainAxisSize: MainAxisSize.min,
                    crossAxisAlignment: CrossAxisAlignment.stretch,
                    children: [
                      // Logo/Title
                      const Icon(
                        Icons.sports_esports,
                        size: 64,
                        color: Colors.blueAccent,
                      ),
                      const SizedBox(height: 16),
                      const Text(
                        'Pong Game',
                        textAlign: TextAlign.center,
                        style: TextStyle(
                          fontSize: 32,
                          fontWeight: FontWeight.bold,
                          color: Colors.white,
                        ),
                      ),
                      const SizedBox(height: 8),
                      const Text(
                        'Login with Decred Wallet',
                        textAlign: TextAlign.center,
                        style: TextStyle(
                          fontSize: 16,
                          color: Colors.white70,
                        ),
                      ),
                      const SizedBox(height: 32),

                      // Instructions
                      const Text(
                        '1. Enter your P2PKH address (D... or T...)',
                        style: TextStyle(color: Colors.white70, fontSize: 14),
                      ),
                      const SizedBox(height: 12),

                      // Address input
                      TextField(
                        controller: _addrCtrl,
                        enabled: !_loading,
                        style: const TextStyle(color: Colors.white),
                        decoration: InputDecoration(
                          labelText: 'Wallet Address',
                          labelStyle: const TextStyle(color: Colors.white54),
                          prefixIcon: const Icon(Icons.account_balance_wallet,
                              color: Colors.blueAccent),
                          filled: true,
                          fillColor: Colors.black26,
                          border: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                            borderSide: BorderSide.none,
                          ),
                          enabledBorder: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                            borderSide: const BorderSide(color: Colors.white24),
                          ),
                          focusedBorder: OutlineInputBorder(
                            borderRadius: BorderRadius.circular(8),
                            borderSide: const BorderSide(
                                color: Colors.blueAccent, width: 2),
                          ),
                        ),
                      ),
                      const SizedBox(height: 16),

                      // Request nonce button
                      const Text(
                        '2. Request a code to sign',
                        style: TextStyle(color: Colors.white70, fontSize: 14),
                      ),
                      const SizedBox(height: 12),
                      ElevatedButton.icon(
                        onPressed:
                            !_loading ? () => _requestNonce(model) : null,
                        icon: const Icon(Icons.refresh),
                        label: const Text('Request Code'),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.blueAccent,
                          padding: const EdgeInsets.symmetric(vertical: 16),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(8),
                          ),
                        ),
                      ),

                      // Show nonce
                      if (_nonce.isNotEmpty) ...[
                        const SizedBox(height: 16),
                        Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: Colors.green.withOpacity(0.1),
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(color: Colors.greenAccent),
                          ),
                          child: Column(
                            crossAxisAlignment: CrossAxisAlignment.start,
                            children: [
                              const Text(
                                'Code to sign:',
                                style: TextStyle(
                                    color: Colors.greenAccent,
                                    fontWeight: FontWeight.bold),
                              ),
                              const SizedBox(height: 8),
                              SelectableText(
                                _nonce,
                                style: const TextStyle(
                                    color: Colors.white,
                                    fontFamily: 'monospace'),
                              ),
                            ],
                          ),
                        ),
                        const SizedBox(height: 16),
                        const Text(
                          '3. Sign this code in your wallet and paste the signature',
                          style: TextStyle(color: Colors.white70, fontSize: 14),
                        ),
                        const SizedBox(height: 12),
                        TextField(
                          controller: _sigCtrl,
                          enabled: !_loading,
                          minLines: 2,
                          maxLines: 4,
                          style: const TextStyle(
                              color: Colors.white, fontFamily: 'monospace'),
                          decoration: InputDecoration(
                            labelText: 'Base64 Signature',
                            labelStyle: const TextStyle(color: Colors.white54),
                            prefixIcon: const Padding(
                              padding: EdgeInsets.only(bottom: 50),
                              child: Icon(Icons.key, color: Colors.blueAccent),
                            ),
                            filled: true,
                            fillColor: Colors.black26,
                            border: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(8),
                              borderSide: BorderSide.none,
                            ),
                            enabledBorder: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(8),
                              borderSide:
                                  const BorderSide(color: Colors.white24),
                            ),
                            focusedBorder: OutlineInputBorder(
                              borderRadius: BorderRadius.circular(8),
                              borderSide: const BorderSide(
                                  color: Colors.blueAccent, width: 2),
                            ),
                          ),
                        ),
                        const SizedBox(height: 16),
                        ElevatedButton.icon(
                          onPressed: (!_loading && _nonce.isNotEmpty)
                              ? () => _verify(model)
                              : null,
                          icon: _loading
                              ? const SizedBox(
                                  width: 20,
                                  height: 20,
                                  child: CircularProgressIndicator(
                                    strokeWidth: 2,
                                    valueColor: AlwaysStoppedAnimation<Color>(
                                        Colors.white),
                                  ),
                                )
                              : const Icon(Icons.login),
                          label: Text(_loading ? 'Verifying...' : 'Login'),
                          style: ElevatedButton.styleFrom(
                            backgroundColor: Colors.green,
                            padding: const EdgeInsets.symmetric(vertical: 16),
                            shape: RoundedRectangleBorder(
                              borderRadius: BorderRadius.circular(8),
                            ),
                          ),
                        ),
                      ],

                      // Error message
                      if (_status.isNotEmpty) ...[
                        const SizedBox(height: 16),
                        Container(
                          padding: const EdgeInsets.all(12),
                          decoration: BoxDecoration(
                            color: Colors.red.withOpacity(0.2),
                            borderRadius: BorderRadius.circular(8),
                            border: Border.all(color: Colors.redAccent),
                          ),
                          child: Row(
                            children: [
                              const Icon(Icons.error, color: Colors.redAccent),
                              const SizedBox(width: 8),
                              Expanded(
                                child: Text(
                                  _status,
                                  style:
                                      const TextStyle(color: Colors.redAccent),
                                ),
                              ),
                            ],
                          ),
                        ),
                      ],

                      const SizedBox(height: 32),

                      // Refund button - accessible without login
                      const Divider(color: Colors.white24),
                      const SizedBox(height: 16),
                      const Text(
                        'Need to refund a CSV-locked escrow?',
                        textAlign: TextAlign.center,
                        style: TextStyle(
                          fontSize: 14,
                          color: Colors.white70,
                          fontWeight: FontWeight.bold,
                        ),
                      ),
                      const SizedBox(height: 8),
                      Text(
                        'No login required. Refund escrows directly using your wallet.',
                        textAlign: TextAlign.center,
                        style: TextStyle(
                          fontSize: 12,
                          color: Colors.grey.shade400,
                          fontStyle: FontStyle.italic,
                        ),
                      ),
                      const SizedBox(height: 12),
                      ElevatedButton.icon(
                        onPressed: _openRefundScreen,
                        icon: const Icon(Icons.refresh, size: 20),
                        label: const Text('Refund Escrow'),
                        style: ElevatedButton.styleFrom(
                          backgroundColor: Colors.orange,
                          padding: const EdgeInsets.symmetric(vertical: 14),
                          shape: RoundedRectangleBorder(
                            borderRadius: BorderRadius.circular(8),
                          ),
                        ),
                      ),
                    ],
                  ),
                ),
              ),
            ),
          ),
        ),
      ),
    );
  }
}
