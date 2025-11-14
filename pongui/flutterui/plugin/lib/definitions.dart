// ignore_for_file: constant_identifier_names

import 'dart:async';
import 'dart:convert';
import 'dart:isolate';

import 'package:json_annotation/json_annotation.dart';
import 'package:golib_plugin/grpc/generated/pong.pbgrpc.dart';

import 'package:flutter/foundation.dart'; // for debugPrint

part 'definitions.g.dart';

// CSV blocks for escrow timelock - change this for testing (2 for testing, 64 for production)
const int CSV_BLOCKS = 2;
// Default bet size in atoms (0.01 DCR)
const int DEFAULT_BET_ATOMS = 1000000;

// Isolate entry point: decode protobuf -> build a plain Map payload
// (numbers only) that can cross isolates efficiently.
void _frameDecodeIsolate(SendPort mainPort) {
  final inbox = ReceivePort();
  mainPort.send(inbox.sendPort);
  inbox.listen((msg) {
    try {
      if (msg is TransferableTypedData) {
        final raw = msg.materialize().asUint8List();
        final gu = GameUpdate.fromBuffer(raw);
        // 14 x float32 + 2 x int32 = 64 bytes
        final bd = ByteData(64);
        var o = 0;
        void wf(double v) {
          bd.setFloat32(o, v.toDouble(), Endian.little);
          o += 4;
        }

        void wi(int v) {
          bd.setInt32(o, v, Endian.little);
          o += 4;
        }

        wf(gu.gameWidth);
        wf(gu.gameHeight);
        wf(gu.p1X);
        wf(gu.p1Y);
        wf(gu.p1Width);
        wf(gu.p1Height);
        wf(gu.p2X);
        wf(gu.p2Y);
        wf(gu.p2Width);
        wf(gu.p2Height);
        wf(gu.ballX);
        wf(gu.ballY);
        wf(gu.ballWidth);
        wf(gu.ballHeight);
        wi(gu.p1Score);
        wi(gu.p2Score);

        final out = TransferableTypedData.fromList([bd.buffer.asUint8List()]);
        mainPort.send(out);
      }
    } catch (e, st) {
      mainPort.send('err: $e\n$st');
    }
  });
}

@JsonSerializable()
class InitClient {
  @JsonKey(name: 'client_id')
  final String clientId; // Wallet-authenticated clientID (required)
  @JsonKey(name: 'server_addr')
  final String serverAddr;
  @JsonKey(name: 'grpc_cert_path')
  final String grpcCertPath;
  @JsonKey(name: 'datadir')
  final String dataDir;
  @JsonKey(name: 'log_file')
  final String logFile;
  @JsonKey(name: "msgs_root")
  final String msgsRoot;
  @JsonKey(name: 'debug_level')
  final String debugLevel;

  // rpc fields
  @JsonKey(name: 'rpc_websocket_url')
  final String rpcWebsockeURL;
  @JsonKey(name: 'rpc_cert_path')
  final String rpcCertPath;
  @JsonKey(name: 'rpc_client_cert_path')
  final String rpcClientCertpath;
  @JsonKey(name: 'rpc_client_key_path')
  final String rpcClientKeypath;
  @JsonKey(name: 'rpc_user')
  final String rpcUser;
  @JsonKey(name: 'rpc_pass')
  final String rpcPass;

  InitClient(
    this.clientId,
    this.serverAddr,
    this.grpcCertPath,
    this.dataDir,
    this.logFile,
    this.msgsRoot,
    this.debugLevel,
    this.rpcWebsockeURL,
    this.rpcCertPath,
    this.rpcClientCertpath,
    this.rpcClientKeypath,
    this.rpcUser,
    this.rpcPass,
  );

  Map<String, dynamic> toJson() => _$InitClientToJson(this);
}

@JsonSerializable()
class IDInit {
  @JsonKey(name: 'id')
  final String uid;
  @JsonKey(name: 'nick')
  final String nick;
  IDInit(this.uid, this.nick);
  factory IDInit.fromJson(Map<String, dynamic> json) => _$IDInitFromJson(json);

  Map<String, dynamic> toJson() => _$IDInitToJson(this);
}

@JsonSerializable()
class GetUserNickArgs {
  @JsonKey(name: 'uid')
  final String uid;

  GetUserNickArgs(this.uid);
  Map<String, dynamic> toJson() => _$GetUserNickArgsToJson(this);
}

@JsonSerializable()
class LocalPlayer {
  @JsonKey(name: 'uid')
  final String uid;
  @JsonKey(name: 'nick')
  final String? nick;
  @JsonKey(name: 'bet_amt')
  final int betAmount;
  @JsonKey(name: 'ready')
  bool ready;
  // Whether this player is currently connected to the server (as far as the
  // client knows). Used to highlight disconnected opponents in waiting rooms.
  @JsonKey(name: 'connected', defaultValue: true)
  bool connected;

  LocalPlayer(
    this.uid,
    this.nick,
    this.betAmount, {
    this.ready = false,
    this.connected = true,
  });

  factory LocalPlayer.fromJson(Map<String, dynamic> json) =>
      _$LocalPlayerFromJson(json);
  Map<String, dynamic> toJson() => _$LocalPlayerToJson(this);

  factory LocalPlayer.fromProto(Player player) {
    return LocalPlayer(
      player.uid,
      player.nick,
      player.betAmt.toInt(),
      ready: player.ready,
      connected: player.connected,
    );
  }
}

@JsonSerializable()
class LocalWaitingRoom {
  @JsonKey(name: 'id')
  final String id;
  @JsonKey(name: 'host_id')
  final String host;
  @JsonKey(name: 'bet_amt')
  final int betAmt;
  @JsonKey(name: 'players', defaultValue: [])
  final List<LocalPlayer> players;

  const LocalWaitingRoom(
    this.id,
    this.host,
    this.betAmt, {
    this.players = const [],
  });

  factory LocalWaitingRoom.fromJson(Map<String, dynamic> json) =>
      _$LocalWaitingRoomFromJson(json);
  Map<String, dynamic> toJson() => _$LocalWaitingRoomToJson(this);

  factory LocalWaitingRoom.fromProto(WaitingRoom wr) {
    return LocalWaitingRoom(
      wr.id,
      wr.hostId,
      wr.betAmt.toInt(),
      players: wr.players
          .map((player) => LocalPlayer(
                player.uid,
                player.nick,
                player.betAmt.toInt(),
                ready: player.ready,
              ))
          .toList(),
    );
  }
}

@JsonSerializable()
class LocalInfo {
  final String id;
  final String nick;
  @JsonKey(name: 'server_version', defaultValue: "")
  final String serverVersion;
  @JsonKey(name: 'server_is_f2p', defaultValue: false)
  final bool serverIsF2P;
  LocalInfo(this.id, this.nick, {String? serverVersion, bool? serverIsF2P})
      : serverVersion = serverVersion ?? "",
        serverIsF2P = serverIsF2P ?? false;
  factory LocalInfo.fromJson(Map<String, dynamic> json) =>
      _$LocalInfoFromJson(json);
}

@JsonSerializable()
class ServerCert {
  @JsonKey(name: "inner_fingerprint")
  final String innerFingerprint;
  @JsonKey(name: "outer_fingerprint")
  final String outerFingerprint;
  const ServerCert(this.innerFingerprint, this.outerFingerprint);

  factory ServerCert.fromJson(Map<String, dynamic> json) =>
      _$ServerCertFromJson(json);
}

const connStateOffline = 0;
const connStateCheckingWallet = 1;
const connStateOnline = 2;

@JsonSerializable()
class ServerInfo {
  final String innerFingerprint;
  final String outerFingerprint;
  final String serverAddr;
  const ServerInfo(
      {required this.innerFingerprint,
      required this.outerFingerprint,
      required this.serverAddr});
  const ServerInfo.empty()
      : this(innerFingerprint: "", outerFingerprint: "", serverAddr: "");

  factory ServerInfo.fromJson(Map<String, dynamic> json) =>
      _$ServerInfoFromJson(json);
}

@JsonSerializable()
class RemoteUser {
  final String uid;
  final String nick;

  const RemoteUser(this.uid, this.nick);

  factory RemoteUser.fromJson(Map<String, dynamic> json) =>
      _$RemoteUserFromJson(json);
}

@JsonSerializable()
class PublicIdentity {
  final String name;
  final String nick;
  final String identity;

  PublicIdentity(this.name, this.nick, this.identity);
  factory PublicIdentity.fromJson(Map<String, dynamic> json) =>
      _$PublicIdentityFromJson(json);
}

@JsonSerializable()
class Account {
  final String name;
  @JsonKey(name: "unconfirmed_balance")
  final int unconfirmedBalance;
  @JsonKey(name: "confirmed_balance")
  final int confirmedBalance;
  @JsonKey(name: "internal_key_count")
  final int internalKeyCount;
  @JsonKey(name: "external_key_count")
  final int externalKeyCount;

  Account(this.name, this.unconfirmedBalance, this.confirmedBalance,
      this.internalKeyCount, this.externalKeyCount);

  factory Account.fromJson(Map<String, dynamic> json) =>
      _$AccountFromJson(json);
}

@JsonSerializable()
class LogEntry {
  final String from;
  final String message;
  final bool internal;
  final int timestamp;
  LogEntry(this.from, this.message, this.internal, this.timestamp);

  factory LogEntry.fromJson(Map<String, dynamic> json) =>
      _$LogEntryFromJson(json);
}

@JsonSerializable()
class SendOnChain {
  final String addr;
  final int amount;
  @JsonKey(name: "from_account")
  final String fromAccount;

  SendOnChain(this.addr, this.amount, this.fromAccount);
  Map<String, dynamic> toJson() => _$SendOnChainToJson(this);
}

@JsonSerializable()
class LoadUserHistory {
  final String uid;
  @JsonKey(name: "is_gc")
  final bool isGC;
  final int page;
  @JsonKey(name: "page_num")
  final int pageNum;

  LoadUserHistory(this.uid, this.isGC, this.page, this.pageNum);
  Map<String, dynamic> toJson() => _$LoadUserHistoryToJson(this);
}

@JsonSerializable()
class WriteInvite {
  @JsonKey(name: "fund_amount")
  final int fundAmount;
  @JsonKey(name: "fund_account")
  final String fundAccount;
  @JsonKey(name: "gc_id")
  final String? gcid;
  final bool prepaid;

  WriteInvite(this.fundAmount, this.fundAccount, this.gcid, this.prepaid);
  Map<String, dynamic> toJson() => _$WriteInviteToJson(this);
}

@JsonSerializable()
class RedeemedInviteFunds {
  final String txid;
  final int total;

  RedeemedInviteFunds(this.txid, this.total);
  factory RedeemedInviteFunds.fromJson(Map<String, dynamic> json) =>
      _$RedeemedInviteFundsFromJson(json);
}

@JsonSerializable()
class CreateWaitingRoomArgs {
  @JsonKey(name: 'client_id')
  final String clientId;
  @JsonKey(name: 'bet_amt')
  final int betAmt;
  @JsonKey(name: 'escrow_id')
  final String? escrowId;

  CreateWaitingRoomArgs(this.clientId, this.betAmt, {this.escrowId});

  Map<String, dynamic> toJson() => _$CreateWaitingRoomArgsToJson(this);

  factory CreateWaitingRoomArgs.fromJson(Map<String, dynamic> json) =>
      _$CreateWaitingRoomArgsFromJson(json);
}

@JsonSerializable()
class RunState {
  @JsonKey(name: "dcrlnd_running")
  final bool dcrlndRunning;
  @JsonKey(name: "client_running")
  final bool clientRunning;

  RunState({required this.dcrlndRunning, required this.clientRunning});
  factory RunState.fromJson(Map<String, dynamic> json) =>
      _$RunStateFromJson(json);
}

@JsonSerializable()
class ZipLogsArgs {
  @JsonKey(name: "include_golib")
  final bool includeGolib;
  @JsonKey(name: "include_ln")
  final bool includeLn;
  @JsonKey(name: "only_last_file")
  final bool onlyLastFile;
  @JsonKey(name: "dest_path")
  final String destPath;

  ZipLogsArgs(
      this.includeGolib, this.includeLn, this.onlyLastFile, this.destPath);
  Map<String, dynamic> toJson() => _$ZipLogsArgsToJson(this);
}

const UINtfnPM = "pm";
const UINtfnGCM = "gcm";
const UINtfnGCMMention = "gcmmention";
const UINtfnMultiple = "multiple";

@JsonSerializable()
class UINotification {
  final String type;
  final String text;
  @JsonKey(defaultValue: 0)
  final int count;
  @JsonKey(defaultValue: "")
  final String from;

  UINotification(this.type, this.text, this.count, this.from);
  factory UINotification.fromJson(Map<String, dynamic> json) =>
      _$UINotificationFromJson(json);
}

@JsonSerializable()
class UINotificationsConfig {
  final bool pms;
  final bool gcms;
  @JsonKey(name: "gcmentions")
  final bool gcMentions;

  UINotificationsConfig(this.pms, this.gcms, this.gcMentions);
  factory UINotificationsConfig.disabled() =>
      UINotificationsConfig(false, false, false);
  factory UINotificationsConfig.fromJson(Map<String, dynamic> json) =>
      _$UINotificationsConfigFromJson(json);
  Map<String, dynamic> toJson() => _$UINotificationsConfigToJson(this);
}

mixin NtfStreams {
  // --- broadcast streams (safe for multiple listeners) ---
  final _acceptedInvitesCtrl = StreamController<RemoteUser>.broadcast();
  Stream<RemoteUser> get acceptedInvites => _acceptedInvitesCtrl.stream;

  final _logLinesCtrl = StreamController<String>.broadcast();
  Stream<String> get logLines => _logLinesCtrl.stream;

  final _rescanProgressCtrl = StreamController<int>.broadcast();
  Stream<int> get rescanWalletProgress => _rescanProgressCtrl.stream;

  final _uiNotificationsCtrl = StreamController<UINotification>.broadcast();
  Stream<UINotification> get uiNotifications => _uiNotificationsCtrl.stream;

  // high-frequency game updates
  final _gameUpdatesCtrl = StreamController<GameUpdate>.broadcast();
  Stream<GameUpdate> get gameUpdates => _gameUpdatesCtrl.stream;

  // Perf stats stream for UI overlays / debugging.
  final _perfStatsCtrl = StreamController<PerfStats>.broadcast();
  Stream<PerfStats> get perfStats => _perfStatsCtrl.stream;

  // --- simplified decoder pipeline (no isolate) ---
  // Keep only the latest raw frame to bound backlog during bursts.
  final List<Uint8List> _frameQueue = <Uint8List>[]; // capacity effectively 1
  bool _decoding = false; // true while draining the queue
  bool _disposed = false;
  // Small jitter buffer: decode as fast as possible, present at steady cadence.

  // Track last emit time (for stall warnings only; no gating)
  int _lastEmitMicros = 0;

  // Perf logging (no behavior change)
  Timer? _perfTimer;
  // No render timer: UI drives rendering on its own (e.g. via vsync/RenderLoop).
  int _framesIn = 0; // frames received from FFI isolate
  int _framesDecoded = 0; // frames decoded and emitted
  // No drop counter now that gating is removed
  final Stopwatch _decodeSw = Stopwatch();
  int _lastDecodeMs = 0;
  int _maxDecodeMs = 0;
  int _qMax = 0; // peak queue size per second
  int _ffiFwd = 0; // last forwarded frames/sec from FFI isolate
  // jitter (ms) for incoming notifications and emitted frames
  int _inDtMin = 1 << 30, _inDtMax = 0, _inDtSum = 0, _inDtCount = 0;
  int _outDtMin = 1 << 30, _outDtMax = 0, _outDtSum = 0, _outDtCount = 0;

  // call this from your existing notification hook
  void handleNotifications(int cmd, bool isError, Object? payload) {
    // Start periodic perf log the first time we receive anything.
    _perfTimer ??= Timer.periodic(const Duration(seconds: 1), (_) {
      final nowUs = DateTime.now().microsecondsSinceEpoch;
      // Calculate time since last emit, or 0 if no frames emitted yet
      final sinceLastMs =
          _lastEmitMicros > 0 ? ((nowUs - _lastEmitMicros) ~/ 1000) : 0;
      // debugPrint(
      //     '[ui] frames in=$_framesIn out=$_framesDecoded decode=${_lastDecodeMs}ms max=${_maxDecodeMs}ms');
      final inAvg = _inDtCount > 0 ? (_inDtSum ~/ _inDtCount) : 0;
      final outAvg = _outDtCount > 0 ? (_outDtSum ~/ _outDtCount) : 0;
      // Push stats to UI subscribers.
      try {
        if (!_perfStatsCtrl.isClosed) {
          _perfStatsCtrl.add(PerfStats(
            pingCurMs: sinceLastMs,
            framesIn: _framesIn,
            framesOut: _framesDecoded,
            decodeLastMs: _lastDecodeMs,
            decodeMaxMs: _maxDecodeMs,
            queueLen: _frameQueue.length,
            queueMax: _qMax,
            sinceLastEmitMs: sinceLastMs,
            ffiFwd: _ffiFwd,
            inDtMin: (_inDtCount > 0 && _inDtMin < (1 << 30)) ? _inDtMin : 0,
            inDtAvg: inAvg,
            inDtMax: _inDtMax,
            outDtMin:
                (_outDtCount > 0 && _outDtMin < (1 << 30)) ? _outDtMin : 0,
            outDtAvg: outAvg,
            outDtMax: _outDtMax,
          ));
        }
      } catch (_) {}
      // reset for next interval
      _framesIn = 0;
      _framesDecoded = 0;
      _maxDecodeMs = 0;
      _qMax = 0;
      _inDtMin = 1 << 30;
      _inDtMax = 0;
      _inDtSum = 0;
      _inDtCount = 0;
      _outDtMin = 1 << 30;
      _outDtMax = 0;
      _outDtSum = 0;
      _outDtCount = 0;
    });
    switch (cmd) {
      case NTNOP:
        break;

      case NTUINotification:
        // Move JSON decode and stream add to microtask to avoid blocking the hot path
        if (payload is String && payload.isNotEmpty) {
          scheduleMicrotask(() {
            try {
              final decoded = jsonDecode(payload);
              final n = UINotification.fromJson(
                Map<String, dynamic>.from(decoded),
              );
              if (!_uiNotificationsCtrl.isClosed) {
                _uiNotificationsCtrl.add(n);
              }
            } catch (e, st) {
              debugPrint(
                  'Failed to decode NTUINotification: $e\n$st\nPayload: $payload');
            }
          });
        }
        break;

      case NTGameFrame:
        // payload is raw bytes of your packed GameUpdate.
        final raw =
            (payload as TransferableTypedData).materialize().asUint8List();
        _framesIn++;
        _frameQueue.add(raw); // enqueue; no drop/coalesce
        if (!_decoding) {
          // Start draining the queue soon, off the hot path
          scheduleMicrotask(_drainDecode);
        }
        break;

      case NTClientStopped:
        // optional: surface a UI event here
        break;

      case NTPerfFwd:
        // Per-second forwarded count reported by FFI isolate.
        if (payload is int) {
          _ffiFwd = payload;
        } else if (payload is String) {
          final v = int.tryParse(payload.trim());
          if (v != null) _ffiFwd = v;
        }
        break;

      default:
        debugPrint('Unknown notification 0x${cmd.toRadixString(16)}');
    }
  }

  void _drainDecode() {
    if (_disposed) return;
    if (_frameQueue.isEmpty) {
      _decoding = false;
      return;
    }
    _decoding = true;

    final raw = _frameQueue.removeAt(0);

    try {
      // Decode protobuf payload sent by Go server.
      _decodeSw
        ..reset()
        ..start();
      final gu = GameUpdate.fromBuffer(raw);
      _decodeSw.stop();
      _lastDecodeMs = _decodeSw.elapsedMilliseconds;
      if (_lastDecodeMs > _maxDecodeMs) _maxDecodeMs = _lastDecodeMs;

      if (!_gameUpdatesCtrl.isClosed) _gameUpdatesCtrl.add(gu);
      final emitMicros = DateTime.now().microsecondsSinceEpoch;
      _lastEmitMicros = emitMicros;
      _framesDecoded++;
    } catch (e, st) {
      debugPrint('Failed to decode game frame: $e\n$st');
    } finally {
      _decoding = false;
      // Continue draining if there are more frames queued.
      if (!_disposed && _frameQueue.isNotEmpty) {
        scheduleMicrotask(_drainDecode);
      } else {
        _decoding = false;
      }
    }
  }

  // call when your owning object is disposed
  void disposeNtfStreams() {
    _disposed = true;
    _perfTimer?.cancel();
    _perfStatsCtrl.close();
    _acceptedInvitesCtrl.close();
    _logLinesCtrl.close();
    _rescanProgressCtrl.close();
    _uiNotificationsCtrl.close();
    _gameUpdatesCtrl.close();
  }
}

abstract class PluginPlatform {
  Future<String?> get platformVersion => throw "unimplemented";
  String get majorPlatform => "unknown-major-plat";
  String get minorPlatform => "unknown-minor-plat";
  Future<void> setTag(String tag) async => throw "unimplemented";
  Future<void> hello() async => throw "unimplemented";
  Future<String> getURL(String url) async => throw "unimplemented";
  Future<String> nextTime() async => throw "unimplemented";
  Future<void> writeStr(String s) async => throw "unimplemented";
  Stream<String> readStream() async* {
    throw "unimplemented";
  }

  // These are only implemented in android.
  Future<void> startForegroundSvc() => throw "unimplemented";
  Future<void> stopForegroundSvc() => throw "unimplemented";
  Future<void> setNtfnsEnabled(bool enabled) => throw "unimplemented";

  // Expose UI notifications stream to UI code. Implemented by mixin NtfStreams in concrete platforms.
  Stream<UINotification> get uiNotifications => throw "unimplemented";
  // Expose binary game updates decoded into GameUpdate objects.
  Stream<GameUpdate> get gameUpdates => throw "unimplemented";
  // Expose perf stats for debugging overlays.
  Stream<PerfStats> get perfStats => throw "unimplemented";
  // No separate structured stream.

  Future<dynamic> asyncCall(int cmd, dynamic payload) async =>
      throw "unimplemented";

  // Wallet-auth helpers (use golib over gRPC)
  Future<String> requestNonce(String serverAddr, String grpcCertPath) async {
    final res = await asyncCall(CTRequestNonce, {
      'server_addr': serverAddr,
      'grpc_cert_path': grpcCertPath,
    });
    if (res is Map) {
      final m = Map<String, dynamic>.from(res);
      return (m['nonce'] ?? '').toString();
    }
    return res?.toString() ?? '';
  }

  Future<Map<String, dynamic>> verifyLogin(
    String serverAddr,
    String grpcCertPath,
    String address,
    String nonce,
    String signature,
  ) async {
    final res = await asyncCall(CTVerifyLogin, {
      'server_addr': serverAddr,
      'grpc_cert_path': grpcCertPath,
      'address': address,
      'nonce': nonce,
      'signature': signature,
    });
    return Map<String, dynamic>.from(res as Map);
  }

  Future<String> asyncHello(String name) async {
    var r = await asyncCall(CTHello, name);
    return r as String;
  }

  Future<LocalInfo> initClient(InitClient args) async {
    var res = await asyncCall(CTInitClient, args);
    return LocalInfo.fromJson(res as Map<String, dynamic>);
  }

  Future<void> createLockFile(String rootDir) async =>
      await asyncCall(CTCreateLockFile, rootDir);
  Future<void> closeLockFile(String rootDir) async =>
      await asyncCall(CTCloseLockFile, rootDir);
  Future<String> userNick(String pid) async {
    return await asyncCall(CTGetUserNick, pid);
  }

  Future<List<LocalPlayer>> getWRPlayers() async {
    var res = await asyncCall(CTGetWRPlayers, "");
    if (res == null) {
      return [];
    }
    return (res as List)
        .map<LocalPlayer>((v) => LocalPlayer.fromJson(v))
        .toList();
  }

  Future<List<LocalWaitingRoom>> getWaitingRooms() async {
    var res = await asyncCall(CTGetWaitingRooms, "");
    if (res == null) {
      return [];
    }
    return (res as List).map<LocalWaitingRoom>((v) {
      return LocalWaitingRoom.fromJson(v);
    }).toList();
  }

  /// Get the current waiting room ID (if any) for this client from the
  /// running golib PongClient instance. Empty string means no active room.
  Future<String> getCurrentWaitingRoomId() async {
    final res = await asyncCall(CTGetCurrentWaitingRoom, "");
    if (res == null) return "";
    final map = Map<String, dynamic>.from(res as Map);
    final id = map['room_id'];
    if (id is String) return id;
    if (id == null) return "";
    return id.toString();
  }

  Future<LocalWaitingRoom> JoinWaitingRoom(String id,
      {String? escrowId}) async {
    try {
      // Always send JSON object so golib handler can consistently parse
      final payload = {
        'room_id': id,
        'escrow_id': escrowId ?? '',
      };
      final response = await asyncCall(CTJoinWaitingRoom, payload);

      if (response is Map<String, dynamic>) {
        return LocalWaitingRoom.fromJson(response);
      } else {
        throw Exception("Invalid response format: $response");
      }
    } catch (err) {
      throw Exception("Failed to join waiting room: $err");
    }
  }

  Future<LocalWaitingRoom> CreateWaitingRoom(CreateWaitingRoomArgs args) async {
    try {
      // Always ensure escrow_id is present in the payload
      final payload = {
        'client_id': args.clientId,
        'bet_amt': args.betAmt,
        'escrow_id': args.escrowId ?? '',
      };
      final response = await asyncCall(CTCreateWaitingRoom, payload);

      if (response is Map<String, dynamic>) {
        return LocalWaitingRoom.fromJson(response);
      } else {
        throw Exception("Invalid response format: $response");
      }
    } catch (err) {
      throw Exception("Failed to join waiting room: $err");
    }
  }

  Future<void> LeaveWaitingRoom(String id) async {
    await asyncCall(CTLeaveWaitingRoom, id);
  }

  // Escrow/Settlement methods
  Future<Map<String, String>> generateSettlementSessionKey() async {
    final res = await asyncCall(CTGenerateSessionKey, "");
    return Map<String, String>.from(res as Map);
  }

  Future<Map<String, dynamic>> openEscrow(
      {required String payout,
      required int betAtoms,
      int csvBlocks = CSV_BLOCKS}) async {
    final payload = {
      'payout': payout,
      'bet_atoms': betAtoms,
      'csv_blocks': csvBlocks,
    };
    final res = await asyncCall(CTOpenEscrow, payload);
    return Map<String, dynamic>.from(res as Map);
  }

  Future<Map<String, dynamic>> refundEscrow(
      {required String escrowId,
      required String destAddr,
      int feeAtoms = 20000,
      int csvBlocks = CSV_BLOCKS,
      int? utxoValue}) async {
    final payload = {
      'escrow_id': escrowId,
      'dest_addr': destAddr,
      'fee_atoms': feeAtoms,
      'csv_blocks': csvBlocks,
      if (utxoValue != null && utxoValue > 0) 'utxo_value': utxoValue,
    };
    final res = await asyncCall(CTRefundEscrow, payload);
    return Map<String, dynamic>.from(res as Map);
  }

  Future<Map<String, dynamic>> validateRefundSession(String escrowId) async {
    final res = await asyncCall(CTValidateRefundSession, {
      'escrow_id': escrowId,
    });
    return Map<String, dynamic>.from(res as Map);
  }

  Future<List<Map<String, dynamic>>> listHistoricEscrows() async {
    // The Go side uses the already-initialized client handle and its data dir.
    // No payload needed here.
    final res = await asyncCall(CTListHistoricEscrows, "");
    if (res == null) return [];
    final data = Map<String, dynamic>.from(res as Map);
    final raw = data['escrows'];
    final List<dynamic> escrows =
        raw is List ? raw : (raw == null ? const [] : [raw]);
    return escrows.map((e) => Map<String, dynamic>.from(e)).toList();
  }

  Future<void> startPreSign(String matchId) async {
    await asyncCall(CTStartPreSign, {'match_id': matchId});
  }

  Future<void> archiveSettlementSessionKey(String matchId) async {
    await asyncCall(CTArchiveSessionKey, {'match_id': matchId});
  }

  Future<void> archiveSettlementSessionKeyWithEscrow(
      String matchId, Map<String, dynamic> escrowInfo) async {
    await asyncCall(CTArchiveSessionKey, {
      'match_id': matchId,
      'escrow_info': escrowInfo,
    });
  }

  Future<void> cacheEscrowInfo(Map<String, dynamic> escrowInfo) async {
    await asyncCall(CTCacheEscrowInfo, escrowInfo);
  }

  /// Cache wallet authentication information (wallet address and payout address)
  /// to persist it across hot reloads.
  Future<void> cacheWalletAuthInfo({
    required String walletAddress,
    required String payoutAddressOrPubkey,
  }) async {
    await asyncCall(CTCacheWalletAuthInfo, {
      'wallet_address': walletAddress,
      'payout_address_or_pubkey': payoutAddressOrPubkey,
    });
  }

  /// Get cached wallet authentication information.
  /// Returns a map with 'wallet_address' and 'payout_address_or_pubkey' keys.
  Future<Map<String, String>> getWalletAuthInfo() async {
    final result = await asyncCall(CTGetWalletAuthInfo, {});
    return Map<String, String>.from(result);
  }

  /// Get active escrow information from the cached session key file.
  /// Returns a map with escrow metadata (escrow_id, funding_txid, etc.) or empty map if none.
  Future<Map<String, dynamic>> getActiveEscrowInfo() async {
    final result = await asyncCall(CTGetActiveEscrowInfo, {});
    return Map<String, dynamic>.from(result);
  }

  Future<void> updateHistoricEscrow(Map<String, dynamic> escrowInfo) async {
    await asyncCall(CTUpdateHistoricEscrow, escrowInfo);
  }

  Future<void> deleteHistoricEscrow(String escrowId) async {
    await asyncCall(CTDeleteHistoricEscrow, {'escrow_id': escrowId});
  }

  // --- Config management via golib ---
  Future<ClientConfig> getClientConfig({String? dataDir}) async {
    final res = await asyncCall(CTGetClientConfig, {'data_dir': dataDir});
    return ClientConfig.fromJson(Map<String, dynamic>.from(res as Map));
  }

  Future<void> saveClientConfig(ClientConfig cfg) async {
    final payload = {
      'server_addr': cfg.serverAddr,
      'grpc_cert_path': cfg.grpcCertPath,
      'network': cfg.network,
      'debug': cfg.debugLevel,
      'show_perfoverlay': cfg.showPerfOverlay,
      'data_dir': cfg.dataDir,
    };
    await asyncCall(CTSaveClientConfig, payload);
  }

  // Player action methods (migrated from Dart gRPC)
  Future<void> sendInput(String input) async {
    await asyncCall(CTSendInput, {'input': input});
  }

  Future<bool> signalReadyToPlay(String gameId) async {
    final res = await asyncCall(CTSignalReadyToPlay, {'game_id': gameId});
    return (res as Map<String, dynamic>)['success'] as bool;
  }

  Future<void> unreadyGameStream() async {
    await asyncCall(CTUnreadyGameStream, "");
  }

  Future<void> startGameStream() async {
    await asyncCall(CTStartGameStream, "");
  }
}

const int CTUnknown = 0x00;
const int CTHello = 0x01;
const int CTInitClient = 0x02;
const int CTGetUserNick = 0x03;
const int CTStopClient = 0x04;
const int CTGetWRPlayers = 0x05;
const int CTGetWaitingRooms = 0x06;
const int CTJoinWaitingRoom = 0x07;
const int CTCreateWaitingRoom = 0x08;
const int CTLeaveWaitingRoom = 0x09;
const int CTGenerateSessionKey = 0x0a;
const int CTOpenEscrow = 0x0b;
const int CTStartPreSign = 0x0c;
const int CTArchiveSessionKey = 0x0e;
const int CTRequestNonce = 0x0f;
const int CTVerifyLogin = 0x10;
const int CTSendInput = 0x11;
const int CTSignalReadyToPlay = 0x12;
const int CTUnreadyGameStream = 0x13;
const int CTStartGameStream = 0x14;
const int CTRefundEscrow = 0x15;
const int CTListHistoricEscrows = 0x16;
const int CTCacheEscrowInfo = 0x17;
const int CTCacheWalletAuthInfo = 0x1d;
const int CTGetWalletAuthInfo = 0x1e;
const int CTGetActiveEscrowInfo = 0x1f;
const int CTGetCurrentWaitingRoom = 0x20;
const int CTUpdateHistoricEscrow = 0x18;
const int CTDeleteHistoricEscrow = 0x19;

// Config management
const int CTGetClientConfig = 0x1a;
const int CTSaveClientConfig = 0x1b;
const int CTValidateRefundSession = 0x1c;

// Client/runtime state helpers
const int CTGetRunState = 0x83;

const int CTCreateLockFile = 0x60;
const int CTCloseLockFile = 0x61;

const int notificationsStartID = 0x1000;
// Notification types (must match golib)
const int NTUINotification = 0x1001;
const int NTClientStopped = 0x1002;
const int NTLogLine = 0x1003;
const int NTNOP = 0x1004;
// Binary game frame (raw pong.GameUpdate bytes)
const int NTGameFrame = 0x1011;
// Per-second forwarded NTGameFrame count from FFI isolate
const int NTPerfFwd = 0x10f0;

// Lightweight perf stats the UI can subscribe to for debugging spikes.
class PerfStats {
  final int framesIn;
  final int framesOut;
  final int decodeLastMs;
  final int decodeMaxMs;
  final int queueLen;
  final int queueMax;
  final int sinceLastEmitMs;
  final int ffiFwd;
  final int inDtMin;
  final int inDtAvg;
  final int inDtMax;
  final int outDtMin;
  final int outDtAvg;
  final int outDtMax;
  final int pingCurMs;

  const PerfStats({
    required this.framesIn,
    required this.framesOut,
    required this.decodeLastMs,
    required this.decodeMaxMs,
    required this.queueLen,
    required this.queueMax,
    required this.sinceLastEmitMs,
    required this.ffiFwd,
    required this.inDtMin,
    required this.inDtAvg,
    required this.inDtMax,
    required this.outDtMin,
    required this.outDtAvg,
    required this.outDtMax,
    required this.pingCurMs,
  });
}

// Lightweight config DTO managed by golib
class ClientConfig {
  final String serverAddr;
  final String grpcCertPath;
  final String network;
  final String debugLevel;
  final bool showPerfOverlay;
  final String dataDir;

  ClientConfig({
    required this.serverAddr,
    required this.grpcCertPath,
    required this.network,
    required this.debugLevel,
    required this.showPerfOverlay,
    required this.dataDir,
  });

  factory ClientConfig.fromJson(Map<String, dynamic> json) {
    return ClientConfig(
      serverAddr: (json['server_addr'] ?? '').toString(),
      grpcCertPath: (json['grpc_cert_path'] ?? '').toString(),
      network: (json['network'] ?? '').toString(),
      debugLevel: (json['debug'] ?? '').toString(),
      showPerfOverlay: json['show_perfoverlay'] == true,
      dataDir: (json['data_dir'] ?? '').toString(),
    );
  }
}
