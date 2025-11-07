import 'dart:io';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/config.dart';
import 'package:path/path.dart' as p;

class DevToolsScreen extends StatefulWidget {
  const DevToolsScreen({super.key});

  @override
  State<DevToolsScreen> createState() => _DevToolsScreenState();
}

class _DevToolsScreenState extends State<DevToolsScreen> {
  final _profAddrCtrl = TextEditingController(text: '127.0.0.1:8118');
  final _profilesDirCtrl = TextEditingController();
  final _zipDestCtrl = TextEditingController();

  bool _busy = false;

  @override
  void initState() {
    super.initState();
    _initDefaults();
  }

  Future<void> _initDefaults() async {
    try {
      final base = await defaultAppDataDir();
      final profDir = p.join(base, 'profiles');
      final zipPath = p.join(base, 'profiles.zip');
      if (mounted) {
        _profilesDirCtrl.text = profDir;
        _zipDestCtrl.text = zipPath;
        setState(() {});
      }
    } catch (_) {}
  }

  Future<void> _startHttpProfiler() async {
    final addr = _profAddrCtrl.text.trim().isEmpty
        ? '127.0.0.1:8118'
        : _profAddrCtrl.text.trim();
    setState(() => _busy = true);
    try {
      await Golib.asyncCall(0x86, addr); // CTEnableProfiler
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(
          content: Text('pprof listening on http://$addr/debug/pprof/'),
        ),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text('Failed to start profiler: $e')));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _startTimedProfiling() async {
    final dir = _profilesDirCtrl.text.trim();
    if (dir.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Choose a profiles directory first')),
      );
      return;
    }
    setState(() => _busy = true);
    try {
      // Ensure directory exists (desktop only convenience)
      try { await Directory(dir).create(recursive: true); } catch (_) {}
      await Golib.asyncCall(0x89, dir); // CTEnableTimedProfiling
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Timed CPU profiling enabled → $dir')),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text('Failed to enable timed profiling: $e')));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  Future<void> _zipProfiles() async {
    final dest = _zipDestCtrl.text.trim();
    if (dest.isEmpty) {
      ScaffoldMessenger.of(context).showSnackBar(
        const SnackBar(content: Text('Choose a destination zip path first')),
      );
      return;
    }
    setState(() => _busy = true);
    try {
      // Ensure parent dir exists
      try { await Directory(p.dirname(dest)).create(recursive: true); } catch (_) {}
      await Golib.asyncCall(0x87, dest); // CTZipTimedProfilingLogs
      if (!mounted) return;
      ScaffoldMessenger.of(context).showSnackBar(
        SnackBar(content: Text('Zipped profiles → $dest')),
      );
    } catch (e) {
      if (!mounted) return;
      ScaffoldMessenger.of(context)
          .showSnackBar(SnackBar(content: Text('Failed to zip profiles: $e')));
    } finally {
      if (mounted) setState(() => _busy = false);
    }
  }

  @override
  void dispose() {
    _profAddrCtrl.dispose();
    _profilesDirCtrl.dispose();
    _zipDestCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Developer Tools'),
      ),
      body: AbsorbPointer(
        absorbing: _busy,
        child: ListView(
          padding: const EdgeInsets.all(16),
          children: [
            Card(
              color: const Color(0xFF1B1E2C),
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('HTTP pprof (live)',
                        style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _profAddrCtrl,
                      decoration: const InputDecoration(
                        labelText: 'Listen address (host:port)',
                        hintText: '127.0.0.1:8118',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    Wrap(
                      spacing: 8,
                      runSpacing: 8,
                      children: [
                        ElevatedButton.icon(
                          onPressed: _startHttpProfiler,
                          icon: const Icon(Icons.play_arrow),
                          label: const Text('Start HTTP Profiler'),
                        ),
                        TextButton(
                          onPressed: () {
                            final addr = _profAddrCtrl.text.trim().isEmpty
                                ? '127.0.0.1:8118'
                                : _profAddrCtrl.text.trim();
                            final url = 'http://$addr/debug/pprof/';
                            ScaffoldMessenger.of(context).showSnackBar(
                              SnackBar(content: Text('pprof URL copied: $url')),
                            );
                            Clipboard.setData(ClipboardData(text: url));
                          },
                          child: const Text('Copy pprof URL'),
                        ),
                      ],
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Use: go tool pprof http://<addr>/debug/pprof/profile?seconds=30',
                      style: TextStyle(color: Colors.white.withOpacity(0.8)),
                    ),
                  ],
                ),
              ),
            ),

            const SizedBox(height: 16),

            Card(
              color: const Color(0xFF1B1E2C),
              child: Padding(
                padding: const EdgeInsets.all(16.0),
                child: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const Text('Timed CPU Profiling (to files)',
                        style: TextStyle(fontSize: 18, fontWeight: FontWeight.bold)),
                    const SizedBox(height: 12),
                    TextField(
                      controller: _profilesDirCtrl,
                      decoration: const InputDecoration(
                        labelText: 'Profiles directory',
                        hintText: '/path/to/profiles',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    ElevatedButton.icon(
                      onPressed: _startTimedProfiling,
                      icon: const Icon(Icons.timelapse),
                      label: const Text('Start Timed Profiling'),
                    ),
                    const SizedBox(height: 16),
                    TextField(
                      controller: _zipDestCtrl,
                      decoration: const InputDecoration(
                        labelText: 'Zip destination',
                        hintText: '/path/to/pongui-profiles.zip',
                        border: OutlineInputBorder(),
                      ),
                    ),
                    const SizedBox(height: 12),
                    ElevatedButton.icon(
                      onPressed: _zipProfiles,
                      icon: const Icon(Icons.archive),
                      label: const Text('Zip Profiles'),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Profiles rotate hourly as profile-YYYY-MM-DDTHH-MM-SS.pprof',
                      style: TextStyle(color: Colors.white.withOpacity(0.8)),
                    ),
                  ],
                ),
              ),
            ),
          ],
        ),
      ),
    );
  }
}
