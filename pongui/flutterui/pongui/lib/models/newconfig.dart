import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:path/path.dart' as p;
import 'package:pongui/config.dart';
import 'package:golib_plugin/golib_plugin.dart';

class NewConfigModel extends ChangeNotifier {
  // ─── Editable fields ────────────────────────────────────────────────────

  String serverAddr = '';
  String grpcCertPath = '';
  String network = '';
  String debugLevel = '';
  bool showPerfOverlay = false;

  final List<String> appArgs;

  // ─── Construction ───────────────────────────────────────────────────────
  NewConfigModel(this.appArgs);

  factory NewConfigModel.fromConfig(Config c) => NewConfigModel([])
    ..serverAddr = c.serverAddr
    ..grpcCertPath = c.grpcCertPath
    ..debugLevel = c.debugLevel
    ..showPerfOverlay = c.showPerfOverlay;

  // ─── Helpers ────────────────────────────────────────────────────────────

  // Load current defaults from golib and update fields.
  Future<void> loadFromGoDefaults() async {
    final dir = await appDatadir();
    final cc = await Golib.getClientConfig(dataDir: dir);
    serverAddr = cc.serverAddr;
    grpcCertPath = cc.grpcCertPath;
    network = cc.network;
    debugLevel = cc.debugLevel;
    showPerfOverlay = cc.showPerfOverlay;
    notifyListeners();
  }

  Future<String> appDatadir() async {
    // Always resolve the sandboxed Application Support path on this platform.
    // Do not rely on golib defaults which may point to non-sandboxed locations.
    return await defaultAppDataDir();
  }

  Future<String> getConfigFilePath() async {
    final dataDir = await appDatadir();
    return p.join(dataDir, '$APPNAME.conf');
  }

  // ─── Save to disk ───────────────────────────────────────────────────────
  Future<void> saveConfig() async {
    // Delegate saving to golib to ensure single source of truth and format consistency.
    final dataDir = await appDatadir();
    final cfg = ClientConfig(
      serverAddr: serverAddr,
      grpcCertPath: grpcCertPath,
      network: network,
      debugLevel: debugLevel,
      showPerfOverlay: showPerfOverlay,
      dataDir: dataDir,
    );
    await Golib.saveClientConfig(cfg);
  }

  // Atomically apply new values and persist them via golib. Also ensure logs dir exists.
  Future<void> applyAndSave({
    required String serverAddr,
    required String grpcCertPath,
    required String debugLevel,
    required bool showPerfOverlay,
  }) async {
    this.serverAddr = serverAddr.trim();
    this.grpcCertPath = grpcCertPath.trim();
    this.debugLevel = debugLevel.trim();
    this.showPerfOverlay = showPerfOverlay;

    final dir = await appDatadir();
    final logs = Directory(p.join(dir, 'logs'));
    if (!await logs.exists()) {
      await logs.create(recursive: true);
    }

    await saveConfig();
  }

  // expose the resolved data directory to the UI for display
  // For compatibility; prefer calling appDatadir() to ensure fresh value from golib.
  String get dataDir => '';
}
