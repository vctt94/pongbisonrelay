import 'dart:io';
import 'package:flutter/foundation.dart';
import 'package:path/path.dart' as p;
import 'package:pongui/config.dart';

class NewConfigModel extends ChangeNotifier {
  // ─── Editable fields ────────────────────────────────────────────────────
  // String rpcUser         = 'defaultuser';
  // String rpcPass         = 'defaultpass';
  String serverAddr      = '104.131.180.29:50051';
  String grpcCertPath    = '';
  // String rpcCertPath     = '';
  // String rpcClientCertPath = '';
  // String rpcClientKeyPath  = '';
  // String rpcWebsocketURL = 'wss://127.0.0.1:7676/ws';
  String debugLevel      = 'info';
  bool   wantsLogNtfns   = false;
  bool   showPerfOverlay = false;

  final List<String> appArgs;
  String _appDataDir = '';
  // String _brDataDir = '';

  // ─── Construction ───────────────────────────────────────────────────────
  NewConfigModel(this.appArgs) {
    _initialiseDefaults();
  }

  factory NewConfigModel.fromConfig(Config c) => NewConfigModel([])
    // ..rpcUser            = c.rpcUser
    // ..rpcPass            = c.rpcPass
    ..serverAddr         = c.serverAddr
    ..grpcCertPath       = c.grpcCertPath
    // ..rpcCertPath        = c.rpcCertPath
    // ..rpcClientCertPath  = c.rpcClientCertPath
    // ..rpcClientKeyPath   = c.rpcClientKeyPath
    // ..rpcWebsocketURL    = c.rpcWebsocketURL
    ..debugLevel         = c.debugLevel
    ..wantsLogNtfns      = c.wantsLogNtfns
    ..showPerfOverlay    = c.showPerfOverlay;

  // ─── Helpers ────────────────────────────────────────────────────────────
  Future<void> _initialiseDefaults() async {
    _appDataDir = await defaultAppDataDir();
    // _brDataDir = await defaultAppDataBRUIGDir();

    grpcCertPath = p.join(_appDataDir, 'server.cert');
    // rpcCertPath  = p.join(_brDataDir, 'rpc.cert');
    // rpcClientCertPath = p.join(_brDataDir, 'rpc-client.cert');
    // rpcClientKeyPath  = p.join(_brDataDir, 'rpc-client.key');

    notifyListeners();
  }

  Future<String> appDatadir() async {
    // Ensure initialization has completed
    if (_appDataDir.isEmpty) {
      _appDataDir = await defaultAppDataDir();
    }
    return _appDataDir;
  }

  Future<String> getConfigFilePath() async {
    final dataDir = await appDatadir();
    return p.join(dataDir, 'pongui.conf');
  }

  // ─── Save to disk ───────────────────────────────────────────────────────
  Future<void> saveConfig() async {
    final cfgPath = await getConfigFilePath();
    final file    = File(cfgPath);

    final content = (StringBuffer()
      ..writeln('server=$serverAddr')
      ..writeln('grpccertpath=$grpcCertPath')
      ..writeln()
      ..writeln('[clientrpc]')
      // ..writeln('rpcuser=$rpcUser')
      // ..writeln('rpcpass=$rpcPass')
      // ..writeln('rpcwebsocketurl=$rpcWebsocketURL')
      // ..writeln('rpccertpath=$rpcCertPath')
      // ..writeln('rpcclientcertpath=$rpcClientCertPath')
      // ..writeln('rpcclientkeypath=$rpcClientKeyPath')
      ..writeln('wantsLogNtfns=${wantsLogNtfns ? "1" : "0"}')
      ..writeln('showPerfOverlay=${showPerfOverlay ? "1" : "0"}')
      ..writeln()
      ..writeln('[log]')
      ..writeln('debuglevel=$debugLevel')
    ).toString();

    await file.parent.create(recursive: true);
    await file.writeAsString(content);
  }

  // expose the resolved data directory to the UI for display
  String get dataDir => _appDataDir;
}
