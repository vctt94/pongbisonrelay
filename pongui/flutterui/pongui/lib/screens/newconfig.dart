import 'package:flutter/material.dart';

import 'package:pongui/components/shared_layout.dart';
import 'package:pongui/models/newconfig.dart';

class NewConfigScreen extends StatefulWidget {
  const NewConfigScreen({
    super.key,
    required this.model,
    required this.onConfigSaved,
  });

  final NewConfigModel model;
  final Future<void> Function() onConfigSaved;

  @override
  State<NewConfigScreen> createState() => _NewConfigScreenState();
}

class _NewConfigScreenState extends State<NewConfigScreen> {
  final _formKey = GlobalKey<FormState>();

  // text controllers
  late final _serverAddr = TextEditingController(text: widget.model.serverAddr);
  late final _grpcCert   = TextEditingController(text: widget.model.grpcCertPath);
  // late final _rpcCert    = TextEditingController(text: widget.model.rpcCertPath);
  // late final _rpcCliCert = TextEditingController(text: widget.model.rpcClientCertPath);
  // late final _rpcCliKey  = TextEditingController(text: widget.model.rpcClientKeyPath);
  // late final _wsURL      = TextEditingController(text: widget.model.rpcWebsocketURL);
  late final _debugLvl   = TextEditingController(text: widget.model.debugLevel);
  // late final _user       = TextEditingController(text: widget.model.rpcUser);
  // late final _pass       = TextEditingController(text: widget.model.rpcPass);

  bool _showPerfOverlay = false;
  String _cfgPath = '', _dataDir = '';

  @override
  void initState() {
    super.initState();
    _showPerfOverlay = widget.model.showPerfOverlay;
    _initHeaderInfo();
    
    // Listen for model changes to update text fields when async initialization completes
    widget.model.addListener(_onModelChanged);
    // Populate fields from golib defaults
    // Note: listener will update controllers when model fields change.
    widget.model.loadFromGoDefaults();
  }

  void _onModelChanged() {
    if (mounted) {
      // Update text controllers when model values change (after async init)
      if (widget.model.serverAddr.isNotEmpty &&
          _serverAddr.text != widget.model.serverAddr) {
        _serverAddr.text = widget.model.serverAddr;
      }
      if (widget.model.grpcCertPath.isNotEmpty &&
          _grpcCert.text != widget.model.grpcCertPath) {
        _grpcCert.text = widget.model.grpcCertPath;
      }
      if (widget.model.debugLevel.isNotEmpty &&
          _debugLvl.text != widget.model.debugLevel) {
        _debugLvl.text = widget.model.debugLevel;
      }
      _showPerfOverlay = widget.model.showPerfOverlay;
    }
  }

  @override
  void dispose() {
    widget.model.removeListener(_onModelChanged);
    _serverAddr.dispose();
    _grpcCert.dispose();
    _debugLvl.dispose();
    super.dispose();
  }

  Future<void> _initHeaderInfo() async {
    _dataDir = await widget.model.appDatadir();
    _cfgPath = await widget.model.getConfigFilePath();
    if (mounted) setState(() {});
  }

  Future<void> _save() async {
    if (!_formKey.currentState!.validate()) return;
    try {
      await widget.model.applyAndSave(
        serverAddr: _serverAddr.text,
        grpcCertPath: _grpcCert.text,
        debugLevel: _debugLvl.text,
        showPerfOverlay: _showPerfOverlay,
      );
      await widget.onConfigSaved();

      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(const SnackBar(content: Text('Config saved!')));
        await _initHeaderInfo();           // refresh header box
      }
    } catch (e, st) {
      debugPrint('Error saving config: $e\n$st');
      if (mounted) {
        ScaffoldMessenger.of(context)
            .showSnackBar(SnackBar(content: Text('Error: $e')));
      }
    }
  }

  @override
  Widget build(BuildContext context) {
    return SharedLayout(
      title: 'Settings',
      child: Padding(
        padding: const EdgeInsets.all(16),
        child: Form(
          key: _formKey,
          child: SingleChildScrollView(
            child: Column(
              children: [
                _HeaderBox(cfgPath: _cfgPath, dataDir: _dataDir),
                const SizedBox(height: 20),
                _field(_serverAddr, 'Server Address', required: true),
                _field(_grpcCert,   'gRPC Server Cert Path'),
                // _field(_rpcCert,    'RPC Cert Path'),
                // _field(_rpcCliCert, 'RPC Client Cert Path'),
                // _field(_rpcCliKey,  'RPC Client Key Path'),
                // _field(_wsURL, 'RPC WebSocket URL', required: true),
                _field(_debugLvl, 'Debug Level'),
                // _field(_user, 'RPC User', required: true),
                // _field(_pass, 'RPC Password', required: true, obscure: true),
                const SizedBox(height: 12),
                Row(
                  mainAxisAlignment: MainAxisAlignment.spaceBetween,
                  children: [
                    const Text('Show Performance Overlay', style: TextStyle(color: Colors.white)),
                    Switch(value: _showPerfOverlay,
                           onChanged: (v) => setState(() => _showPerfOverlay = v)),
                  ],
                ),
                const SizedBox(height: 20),
                ElevatedButton(onPressed: _save, child: const Text('Save Config')),
              ],
            ),
          ),
        ),
      ),
    );
  }

  // simple builder for text fields
  Widget _field(TextEditingController c, String label,
      {bool required = false, bool obscure = false}) {
    return Padding(
      padding: const EdgeInsets.symmetric(vertical: 6),
      child: TextFormField(
        controller: c,
        obscureText: obscure,
        style: const TextStyle(color: Colors.white),
        decoration: InputDecoration(
          labelText: label,
          labelStyle: const TextStyle(color: Colors.white70),
          enabledBorder: const UnderlineInputBorder(
            borderSide: BorderSide(color: Colors.white54),
          ),
          focusedBorder: const UnderlineInputBorder(
            borderSide: BorderSide(color: Colors.blueAccent),
          ),
        ),
        validator: required
            ? (v) => v == null || v.isEmpty ? 'Required' : null
            : null,
      ),
    );
  }
}

// ─── Small header widget just for display ──────────────────────────────────
class _HeaderBox extends StatelessWidget {
  const _HeaderBox({required this.cfgPath, required this.dataDir});
  final String cfgPath, dataDir;

  @override
  Widget build(BuildContext context) {
    if (cfgPath.isEmpty) {
      return const Text('Loading...', style: TextStyle(color: Colors.white70));
    }
    return Container(
      width: double.infinity,
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: const Color(0xFF1B1E2C),
        borderRadius: BorderRadius.circular(12),
        border: Border.all(color: Colors.blueAccent.withOpacity(.3)),
      ),
      child: Column(
        crossAxisAlignment: CrossAxisAlignment.start,
        children: [
          const Row(
            children: [
              Icon(Icons.settings_applications, color: Colors.blueAccent),
              SizedBox(width: 8),
              Text('Config & Data Directory',
                  style: TextStyle(color: Colors.white, fontSize: 18, fontWeight: FontWeight.bold)),
            ],
          ),
          const SizedBox(height: 12),
          const Text('Config file:', style: TextStyle(color: Colors.white70)),
          _Code(cfgPath),
          const SizedBox(height: 8),
          const Text('Data directory:', style: TextStyle(color: Colors.white70)),
          _Code(dataDir),
          const SizedBox(height: 8),
        ],
      ),
    );
  }
}

class _Code extends StatelessWidget {
  const _Code(this.text);
  final String text;
  @override
  Widget build(BuildContext context) => Container(
        width: double.infinity,
        padding: const EdgeInsets.all(8),
        margin: const EdgeInsets.only(top: 4),
        decoration: BoxDecoration(
          color: const Color(0xFF0F0F0F),
          borderRadius: BorderRadius.circular(4),
          border: Border.all(color: Colors.grey.shade700),
        ),
        child: SelectableText(text,
            style: const TextStyle(color: Colors.white, fontFamily: 'monospace', fontSize: 12)),
      );
}
