import 'dart:async';
import 'dart:convert';
import 'dart:developer' as developer;

import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';
import 'package:golib_plugin/golib_plugin.dart';
import 'package:pongui/components/pong_game.dart';
import 'package:pongui/components/helper.dart';
import 'package:pongui/config.dart';
import 'package:golib_plugin/grpc/generated/pong.pb.dart';
import 'package:pongui/models/notifications.dart';
import 'package:path/path.dart' as path;

const String kPongUIVersion = "0.0.1";

// Define a clear enum for game states
enum GameState {
  idle, // Initial state, not in a game or waiting room
  inWaitingRoom, // In a waiting room, not ready
  waitingRoomReady, // In a waiting room and marked as ready
  gameInitialized, // Game has started but not ready to play
  readyToPlay, // Signaled ready to play but waiting for opponent
  countdown, // Countdown in progress
  playing, // Active gameplay
  gameEnded // Game has finished
}

class PongModel extends ChangeNotifier {
  final Config cfg;
  late PongGame pongGame;
  final NotificationModel notificationModel;

  String clientId = '';
  String nick = '';
  int betAmt = 0;
  String escrowId = '';
  String payoutAddressOrPubkey = '';
  String escrowDepositAddress = '';
  String escrowPkScriptHex = '';
  String escrowRedeemScriptHex = '';
  int escrowBetAtoms = 0;
  int escrowCsvBlocks = CSV_BLOCKS;
  String escrowFundingTxid = '';
  int escrowFundingVout = -1;
  int escrowFundingValueAtoms = 0;
  bool escrowInfoPersisted = false;
  String escrowInfoError = '';
  bool escrowRefundSessionValid = false;
  String escrowRefundSessionError = '';
  String fundingStatus = '';
  int escrowConfs = 0;
  // Escrow funding flags derived from notifications
  bool escrowFunded = false; // true when deposit seen (0-conf OK)
  bool escrowConfirmed = false; // true when at least 1 confirmation
  String errorMessage = '';
  List<LocalWaitingRoom> waitingRooms = [];
  LocalWaitingRoom? currentWR;
  GameUpdate? gameState;

  // Historic escrow state (used for refunds)
  List<Map<String, dynamic>> historicEscrows = [];
  bool isLoadingHistoricEscrows = false;
  String historicEscrowsError = '';

  final SnapshotInterpolator interpolator = SnapshotInterpolator();
  final RenderLoop renderLoop = RenderLoop();
  StreamSubscription<GameUpdate>? _gameStreamSub;
  StreamSubscription<UINotification>? _uiNtfnSub;

  // Connection state
  bool isConnected = false;

  // Game related state
  GameState _currentGameState = GameState.idle;
  String currentGameId = '';
  String countdownMessage = '';
  String gameEndingMessage = '';
  // Ready-timeout (pre-countdown) UI state
  int readyCancelRemaining = 0;
  Timer? _readyCancelTimer;

  // Track last settlement match id used for presign so we can archive the
  // session key safely after game completion.
  String lastMatchId = '';

  // Pre-login minimal client init state (used for refunds before auth)
  bool _preloginInitialized = false;

  void setEscrowId(String id) {
    escrowId = id;
    notifyListeners();
  }

  void setEscrowDetails(String id, String depositAddr,
      {String? pkScriptHex, String? redeemScriptHex, int? csvBlocks}) {
    escrowId = id;
    escrowDepositAddress = depositAddr;
    escrowPkScriptHex = pkScriptHex ?? escrowPkScriptHex;
    if (redeemScriptHex != null && redeemScriptHex.isNotEmpty) {
      escrowRedeemScriptHex = redeemScriptHex;
    }
    if (csvBlocks != null && csvBlocks > 0) {
      escrowCsvBlocks = csvBlocks;
    }
    notifyListeners();
  }

  void setEscrowBetAtoms(int atoms) {
    escrowBetAtoms = atoms;
    escrowFundingValueAtoms = atoms;
    // Reflect intended bet in UI immediately
    betAmt = atoms;
    notifyListeners();
  }

  Future<bool> persistEscrowInfo(Map<String, dynamic> info,
      {String failureContext = 'Persisting escrow info'}) async {
    try {
      await Golib.cacheEscrowInfo(info);
      escrowInfoPersisted = true;
      escrowInfoError = '';
      notifyListeners();
      return true;
    } catch (e) {
      escrowInfoPersisted = false;
      escrowInfoError = '$failureContext failed: $e';
      fundingStatus = 'CRITICAL: escrow state not saved. Do not deposit.';
      notificationModel.showNotification(escrowInfoError);
      notifyListeners();
      return false;
    }
  }

  Future<bool> persistInitialEscrowInfo({
    required String escrowId,
    required int betAtoms,
    required int csvBlocks,
    required String pkScriptHex,
    required String redeemScriptHex,
  }) async {
    final ok = await persistEscrowInfo({
      'escrow_id': escrowId,
      'funded_amount': betAtoms,
      'pk_script_hex': pkScriptHex,
      'redeem_script_hex': redeemScriptHex,
      'csv_blocks': csvBlocks,
      'archived_at': DateTime.now().millisecondsSinceEpoch,
    }, failureContext: 'Saving initial escrow metadata');
    if (!ok) {
      return false;
    }
    try {
      final res = await Golib.validateRefundSession(escrowId);
      final valid = res['ok'] == true;
      escrowRefundSessionValid = valid;
      escrowRefundSessionError = valid
          ? ''
          : (res['reason']?.toString() ?? 'unknown validation error');
      if (!valid) {
        // Strengthen user warning if validation failed.
        fundingStatus =
            'CRITICAL: escrow session backup invalid. Deposit address hidden.';
        notifyListeners();
        return false;
      }
      notifyListeners();
      return true;
    } catch (e) {
      escrowRefundSessionValid = false;
      escrowRefundSessionError = 'Validation error: $e';
      fundingStatus =
          'CRITICAL: escrow session backup validation failed. Deposit address hidden.';
      notifyListeners();
      return false;
    }
  }

  // Getters for the game state
  GameState get currentGameState => _currentGameState;
  bool get isInGame =>
      _currentGameState != GameState.idle &&
      _currentGameState != GameState.inWaitingRoom &&
      _currentGameState != GameState.waitingRoomReady;
  bool get isReady => _currentGameState == GameState.waitingRoomReady;
  bool get isGameStarted =>
      _currentGameState != GameState.idle &&
      _currentGameState != GameState.inWaitingRoom &&
      _currentGameState != GameState.waitingRoomReady;
  bool get isReadyToPlay =>
      _currentGameState == GameState.readyToPlay ||
      _currentGameState == GameState.countdown ||
      _currentGameState == GameState.playing;
  bool get countdownStarted => _currentGameState == GameState.countdown;

  String authToken = '';
  bool isWalletAuthenticated = false;
  String walletAddress = '';
  bool serverIsF2P = false;
  String serverVersion = "";

  bool get escrowReadyForMatches => escrowId.isNotEmpty && escrowFunded;
  bool get canJoinRooms => serverIsF2P || escrowReadyForMatches;
  bool get canCreateRoomNow =>
      currentWR == null &&
      (serverIsF2P || (betAmt > 0 && escrowReadyForMatches));

  PongModel(this.cfg, this.notificationModel);

  // Initialize golib PongClient after wallet authentication (requires clientId)
  // Public helper to ensure prelogin initialization from UI code.
  Future<void> ensurePreloginInitialized() async {
    await _initPongClient(cfg, prelogin: true);
  }

  // Initialize golib PongClient. When prelogin=true and clientId is empty,
  // it creates a minimal local-only client and returns early.
  Future<void> _initPongClient(Config cfg, {bool prelogin = false}) async {
    try {
      // Early exits
      if (isConnected && !prelogin) return;
      if (prelogin && (_preloginInitialized || isConnected)) return;

      final appDataDir = await defaultAppDataDir();
      final logFilePath = path.join(appDataDir, "logs", "$APPNAME.log");

      // Build args; if clientId is empty and prelogin=true, pass empty id to trigger minimal client.
      final initClientId = clientId.isNotEmpty ? clientId : "";
      InitClient initArgs = InitClient(
          initClientId,
          cfg.serverAddr,
          cfg.grpcCertPath,
          appDataDir,
          logFilePath,
          "",
          cfg.debugLevel,
          cfg.rpcWebsocketURL,
          cfg.rpcCertPath,
          cfg.rpcClientCertPath,
          cfg.rpcClientKeyPath,
          cfg.rpcUser,
          cfg.rpcPass);

      developer.log("InitClient args: $initArgs");

      var localInfo = await Golib.initClient(initArgs);
      // If this was a prelogin init (no clientId), record success and return early.
      if (clientId.isEmpty) {
        _preloginInitialized = true;
        return;
      }

      // Full init: use IDs returned by golib as authoritative.
      if ((localInfo.id).isNotEmpty) clientId = localInfo.id;
      nick = localInfo.nick;
      serverIsF2P = localInfo.serverIsF2P;
      serverVersion = localInfo.serverVersion;

      // Query initial waiting rooms via golib
      waitingRooms = await Golib.getWaitingRooms();

      // Initialize game client (notifications come via golib now)
      pongGame = PongGame(clientId);

      isConnected = true;
      // Subscribe to game frames decoded in Go (JSON -> GameUpdate)
      _gameStreamSub ??= Golib.gameUpdates.listen((gu) {
        try {
          _handleGameUpdateFrame(gu);
        } catch (_) {}
      });
      // Subscribe to UI notifications forwarded by golib (also carries structured events)
      _uiNtfnSub ??= Golib.uiNotifications.listen((n) {
        try {
          switch (n.type) {
            case 'ready_timeout':
              _handleNtfnReadyTimeout(n);
              break;
            case 'bet_update':
              _handleNtfnBetUpdate(n);
              break;
            case 'server_config':
              _handleNtfnServerConfig(n);
              break;
            case 'game_started':
            case 'gamestarted':
              _handleNtfnGameStarted(n);
              break;
            case 'game_ready_to_play':
              _handleNtfnGameReadyToPlay(n);
              break;
            case 'countdown_update':
              _handleNtfnCountdown(n);
              break;
            case 'game_end':
              _handleNtfnGameEnd(n);
              break;
            case 'player_ready':
              _handleNtfnPlayerReady(n);
              break;
            case 'wr_created':
              _handleNtfnWRCreated(n);
              break;
            case 'wr_removed':
              _handleNtfnWRRemoved(n);
              break;
            case 'player_joined_wr':
            case 'player_left_wr':
              _handleNtfnPlayerWRUpdate(n);
              break;
            case 'connection_state':
              _handleNtfnConnectionState(n);
              break;
            default:
              if (n.text.isNotEmpty) {
                notificationModel.showNotification(n.text);
              }
          }
        } catch (_) {}
      });
      // Notifications are now handled by golib and forwarded automatically
      notifyListeners();
    } catch (exception) {
      // Surface startup/config errors to the UI
      errorMessage = "${exception.toString()}";
      isConnected = false;
      notifyListeners();
      // For prelogin callers (refunds screen), propagate the error so the UI
      // can present the real cause instead of falling through to "unknown handle".
      if (prelogin) {
        rethrow;
      }
    }
  }

  // Apply wallet-based auth: set clientId and initialize client.
  // p2pkAddr: P2PK address from wallet authentication (if provided)
  Future<void> applyWalletAuth(
      {required String newClientId,
      String token = '',
      String address = '',
      String p2pkAddr = ''}) async {
    clientId = newClientId;
    authToken = token;
    isWalletAuthenticated = true;
    walletAddress = address;
    // Set P2PK address from wallet authentication BEFORE initializing client
    // This ensures the authenticated P2PK address is used instead of config xpub
    if (p2pkAddr.isNotEmpty) {
      payoutAddressOrPubkey = p2pkAddr;
    }
    // Initialize the client with wallet-authenticated clientId
    await _initPongClient(cfg);
    notifyListeners();
  }

  Future<void> logout() async {
    // Treat logout as a full client stop so the server
    // can cleanly remove this player from any waiting room
    // via the standard disconnect path.
    try {
      await Golib.asyncCall(CTStopClient, "");
    } catch (_) {
      // Ignore errors if client is already stopped.
    }

    // Reset connection/client state so a future login
    // (including prelogin init on the login screen) will
    // reinitialize golib cleanly.
    isConnected = false;
    _preloginInitialized = false;
    clientId = '';
    serverVersion = "";
    waitingRooms = [];

    // Stop UI/game subscriptions tied to the old client.
    await _gameStreamSub?.cancel();
    _gameStreamSub = null;
    await _uiNtfnSub?.cancel();
    _uiNtfnSub = null;

    isWalletAuthenticated = false;
    walletAddress = '';
    authToken = '';
    payoutAddressOrPubkey = '';
    escrowId = '';
    escrowDepositAddress = '';
    escrowPkScriptHex = '';
    escrowRedeemScriptHex = '';
    escrowBetAtoms = 0;
    escrowCsvBlocks = CSV_BLOCKS;
    escrowFundingTxid = '';
    escrowFundingVout = -1;
    escrowFundingValueAtoms = 0;
    escrowInfoPersisted = false;
    escrowInfoError = '';
    escrowFunded = false;
    escrowConfirmed = false;
    currentWR = null;
    _currentGameState = GameState.idle;
    serverIsF2P = false;
    notifyListeners();
  }

  Map<String, dynamic> _extrasFrom(UINotification n) {
    try {
      final extras = n.from.isNotEmpty ? jsonDecode(n.from) : null;
      if (extras is Map<String, dynamic>) return extras;
    } catch (_) {}
    return const {};
  }

  void _handleNtfnBetUpdate(UINotification n) {
    final extras = _extrasFrom(n);
    final pid = (extras['player_id'] ?? '').toString();
    if (pid != clientId) return;
    final b = extras['bet_amt'];
    final c = extras['confs'];
    if (b is int) {
      betAmt = b;
    } else if (b is num) {
      betAmt = b.toInt();
    }
    if (c is int) {
      escrowConfs = c;
    } else if (c is num) {
      escrowConfs = c.toInt();
    }
    escrowFunded = escrowConfs >= 0;
    escrowConfirmed = escrowConfs >= 1;
    fundingStatus = escrowConfirmed
        ? 'Deposit confirmed ($escrowConfs)'
        : 'Deposit seen (mempool)';

    bool metadataUpdated = false;
    final txid = (extras['funding_txid'] ?? '').toString();
    if (txid.isNotEmpty) {
      escrowFundingTxid = txid;
      metadataUpdated = true;
    }
    final vout = extras['funding_vout'];
    if (vout is num && vout.toInt() >= 0) {
      escrowFundingVout = vout.toInt();
      metadataUpdated = true;
    }
    final amt = extras['funded_amount'];
    if (amt is num && amt.toInt() > 0) {
      escrowFundingValueAtoms = amt.toInt();
      metadataUpdated = true;
    }
    final redeem = (extras['redeem_script_hex'] ?? '').toString();
    if (redeem.isNotEmpty) {
      escrowRedeemScriptHex = redeem;
      metadataUpdated = true;
    }
    final pk = (extras['pk_script_hex'] ?? '').toString();
    if (pk.isNotEmpty) {
      escrowPkScriptHex = pk;
    }
    final csv = extras['csv_blocks'];
    if (csv is num && csv.toInt() > 0) {
      escrowCsvBlocks = csv.toInt();
    }

    if (metadataUpdated && escrowId.isNotEmpty) {
      final info = <String, dynamic>{
        'escrow_id': escrowId,
        'funded_amount': escrowFundingValueAtoms,
        'pk_script_hex': escrowPkScriptHex,
        'csv_blocks': escrowCsvBlocks,
        'archived_at': DateTime.now().millisecondsSinceEpoch,
      };
      if (escrowFundingTxid.isNotEmpty) {
        info['funding_txid'] = escrowFundingTxid;
      }
      if (escrowFundingVout >= 0) {
        info['funding_vout'] = escrowFundingVout;
      }
      if (escrowRedeemScriptHex.isNotEmpty) {
        info['redeem_script_hex'] = escrowRedeemScriptHex;
      }
      persistEscrowInfo(info,
          failureContext: 'Updating escrow funding metadata');
    }

    notifyListeners();
  }

  void _handleNtfnServerConfig(UINotification n) {
    final extras = _extrasFrom(n);
    final flag = extras['is_f2p'];
    bool next = serverIsF2P;
    if (flag is bool) {
      next = flag;
    } else if (flag is num) {
      next = flag != 0;
    } else if (flag is String) {
      next = flag.toLowerCase() == 'true';
    }
    final changed = next != serverIsF2P;
    serverIsF2P = next;
    final srvVer = extras['server_version'];
    if (srvVer is String && srvVer.isNotEmpty) {
      serverVersion = srvVer;
    }
    if (n.text.isNotEmpty) {
      notificationModel.showNotification(n.text);
    }
    if (changed) {
      notifyListeners();
    }
  }

  void _handleNtfnGameStarted(UINotification n) {
    final extras = _extrasFrom(n);
    final gid = (extras['game_id'] ?? '').toString();
    if (gid.isNotEmpty) currentGameId = gid;
    if (_currentGameState == GameState.idle ||
        _currentGameState == GameState.inWaitingRoom ||
        _currentGameState == GameState.waitingRoomReady) {
      _currentGameState = GameState.gameInitialized;
    }
    currentWR = null;
    notifyListeners();
  }

  void _handleNtfnGameReadyToPlay(UINotification n) {
    final extras = _extrasFrom(n);
    final gid = (extras['game_id'] ?? '').toString();
    if (gid.isNotEmpty) currentGameId = gid;
    _currentGameState = GameState.gameInitialized;
    notificationModel.showNotification("Game is ready!");
    notifyListeners();
  }

  void _handleNtfnCountdown(UINotification n) {
    String msg = n.text.trim();
    if (msg.isEmpty) {
      final extras = _extrasFrom(n);
      msg = (extras['message'] ?? '').toString();
    }
    countdownMessage = msg;
    _currentGameState = GameState.countdown;
    // Hide pre-start cancel countdown once actual countdown begins
    _stopReadyCancelCountdown();
    if (msg.contains('0')) {
      _currentGameState = GameState.playing;
    }
    notifyListeners();
  }

  void _handleNtfnGameEnd(UINotification n) {
    _stopGameStreamAndRenderLoop();
    gameEndingMessage = n.text.isNotEmpty ? n.text : 'Game ended';
    // Clear ready-timeout state when game ends
    _stopReadyCancelCountdown();
    _currentGameState = GameState.gameEnded;
    notifyListeners();
  }

  void _handleNtfnPlayerReady(UINotification n) {
    final extras = _extrasFrom(n);
    final pid = (extras['player_id'] ?? '').toString();
    final r = extras['ready'] == true;
    final wr = extras['waiting_room'];

    // When a waiting room snapshot is provided, update local room state so
    // both players can see each other's ready status.
    if (wr is Map<String, dynamic>) {
      final room = LocalWaitingRoom.fromJson(Map<String, dynamic>.from(wr));
      final idx = waitingRooms.indexWhere((r) => r.id == room.id);
      if (idx == -1) {
        waitingRooms = [room, ...waitingRooms];
      } else {
        waitingRooms[idx] = room;
      }
      if (currentWR?.id == room.id) {
        currentWR = room;
      }

      // Keep local readiness state in sync for this client while in a room.
      if (pid == clientId) {
        if (r) {
          _currentGameState = GameState.waitingRoomReady;
        } else if (_currentGameState == GameState.waitingRoomReady) {
          _currentGameState = GameState.inWaitingRoom;
        }
      }
      notifyListeners();
      return;
    }

    // Fallback for legacy notifications that only include player_id/ready and
    // no waiting room context: only adjust state if we're still in the
    // waiting-room phase to avoid clobbering in-game states.
    if (pid == clientId && r && _currentGameState == GameState.inWaitingRoom) {
      _currentGameState = GameState.waitingRoomReady;
      notifyListeners();
    }
  }

  void _handleNtfnWRCreated(UINotification n) {
    final extras = _extrasFrom(n);
    final wr = extras['waiting_room'];
    if (wr is Map<String, dynamic>) {
      final room = LocalWaitingRoom.fromJson(Map<String, dynamic>.from(wr));
      final idx = waitingRooms.indexWhere((r) => r.id == room.id);
      if (idx == -1) {
        waitingRooms = [room, ...waitingRooms];
      } else {
        waitingRooms[idx] = room;
      }
      notifyListeners();
    }
  }

  void _handleNtfnWRRemoved(UINotification n) {
    final extras = _extrasFrom(n);
    final rid = (extras['room_id'] ?? '').toString();
    if (rid.isEmpty) return;
    waitingRooms =
        waitingRooms.where((r) => r.id != rid).toList(growable: false);
    if (currentWR?.id == rid) currentWR = null;
    notifyListeners();
  }

  void _handleNtfnPlayerWRUpdate(UINotification n) {
    final extras = _extrasFrom(n);
    final wr = extras['waiting_room'];
    if (wr is Map<String, dynamic>) {
      final room = LocalWaitingRoom.fromJson(Map<String, dynamic>.from(wr));
      final idx = waitingRooms.indexWhere((r) => r.id == room.id);
      if (idx == -1) {
        waitingRooms = [room, ...waitingRooms];
      } else {
        waitingRooms[idx] = room;
      }
      if (currentWR?.id == room.id) currentWR = room;
      notifyListeners();
    }
  }

  void _handleNtfnConnectionState(UINotification n) async {
    final extras = _extrasFrom(n);
    final connected = extras['connected'];
    if (connected is bool) {
      final wasConnected = isConnected;
      isConnected = connected;

      // If connection was restored, refresh waiting rooms
      if (connected && !wasConnected) {
        try {
          waitingRooms = await Golib.getWaitingRooms();
        } catch (e) {
          developer
              .log("Failed to refresh waiting rooms after reconnection: $e");
        }
      }

      notifyListeners();
    }
  }

  // Removed JSON fallback for game updates. Binary-only via gameUpdates().

  void _handleGameUpdateFrame(GameUpdate gu) {
    // Push into interpolator for smooth rendering; retain last state for fallback.
    interpolator.push(gu);
    gameState = gu;
    renderLoop.start();
  }

  void resetGameState() {
    _currentGameState = GameState.idle;
    currentWR = null;
    betAmt = DEFAULT_BET_ATOMS;
    currentGameId = '';
    countdownMessage = '';
    readyCancelRemaining = 0;
    _readyCancelTimer?.cancel();
    _readyCancelTimer = null;
    gameEndingMessage = '';
    clearEscrowState();
    _stopGameStreamAndRenderLoop();
    notifyListeners();
  }

  void _handleNtfnReadyTimeout(UINotification n) {
    // Use count as seconds if provided; else try extras['seconds'].
    var secs = n.count;
    final extras = _extrasFrom(n);
    if (secs <= 0) {
      final v = extras['seconds'];
      if (v is int) secs = v;
      if (v is num) secs = v.toInt();
    }
    if (secs <= 0) {
      return;
    }
    // Start/restart a local UI countdown even if state has changed
    // (notification might arrive late, but we still want to show the countdown if relevant)
    _startReadyCancelCountdown(secs);
    // Show a subtle notification message
    notificationModel.showNotification(
      'Waiting for both players to be ready…',
    );
  }

  void _startReadyCancelCountdown(int seconds) {
    _readyCancelTimer?.cancel();
    readyCancelRemaining = seconds;
    notifyListeners();
    _readyCancelTimer = Timer.periodic(const Duration(seconds: 1), (t) {
      if (readyCancelRemaining > 0) {
        readyCancelRemaining -= 1;
        notifyListeners();
      }
      if (readyCancelRemaining <= 0) {
        t.cancel();
        _readyCancelTimer = null;
        notifyListeners();
      }
    });
  }

  void _stopReadyCancelCountdown() {
    _readyCancelTimer?.cancel();
    _readyCancelTimer = null;
    if (readyCancelRemaining != 0) {
      readyCancelRemaining = 0;
      notifyListeners();
    }
  }

  // Clear all escrow-related client state so user can open a fresh escrow
  // after a game ends or when leaving a room.
  void clearEscrowState() {
    // Archive the session key with escrow info before clearing
    if (lastMatchId.isNotEmpty && escrowId.isNotEmpty) {
      final missing = <String>[];
      if (escrowFundingTxid.isEmpty) missing.add('funding txid');
      if (escrowFundingVout < 0) missing.add('funding vout');
      if (escrowRedeemScriptHex.isEmpty) missing.add('redeem script');
      if (escrowPkScriptHex.isEmpty) missing.add('pk script');
      if (escrowFundingValueAtoms <= 0) missing.add('funded amount');
      if (missing.isNotEmpty) {
        final msg =
            'Cannot archive escrow $escrowId. Missing: ${missing.join(', ')}';
        developer.log(msg, name: 'escrow');
        notificationModel.showNotification(msg);
      } else {
        final escrowInfo = {
          'escrow_id': escrowId,
          'funding_txid': escrowFundingTxid,
          'funding_vout': escrowFundingVout,
          'funded_amount': escrowFundingValueAtoms,
          'redeem_script_hex': escrowRedeemScriptHex,
          'pk_script_hex': escrowPkScriptHex,
          'csv_blocks': escrowCsvBlocks,
          'archived_at': DateTime.now().millisecondsSinceEpoch,
        };
        unawaited(
          Golib.archiveSettlementSessionKeyWithEscrow(
            lastMatchId,
            escrowInfo,
          ).catchError((e) {
            developer.log(
              'Failed to archive settlement session: $e',
              name: 'escrow',
            );
            notificationModel.showNotification(
              'Failed to archive escrow history: $e',
            );
          }),
        );
      }
    }

    escrowId = '';
    escrowDepositAddress = '';
    escrowPkScriptHex = '';
    escrowRedeemScriptHex = '';
    escrowBetAtoms = 0;
    escrowCsvBlocks = CSV_BLOCKS;
    escrowFundingTxid = '';
    escrowFundingVout = -1;
    escrowFundingValueAtoms = 0;
    escrowInfoPersisted = false;
    escrowInfoError = '';
    escrowRefundSessionValid = false;
    escrowRefundSessionError = '';
    escrowFunded = false;
    escrowConfirmed = false;
    escrowConfs = 0;
    fundingStatus = '';
    lastMatchId = '';
    notifyListeners();
  }

  void clearErrorMessage() {
    errorMessage = '';
    notifyListeners();
  }

  Future<void> createWaitingRoom() async {
    try {
      final sanitizedBet = betAmt >= 0 ? betAmt : 0;
      if (!serverIsF2P) {
        if (sanitizedBet <= 0) {
          errorMessage = "Set a bet amount before creating a room";
          notifyListeners();
          return;
        }
        if (escrowId.isEmpty) {
          errorMessage = "Open escrow first in Settings → Settlement panel";
          notifyListeners();
          return;
        }
        if (!escrowFunded) {
          errorMessage =
              "Wait until escrow deposit is seen before creating a room";
          notifyListeners();
          return;
        }
      }

      CreateWaitingRoomArgs createRoomArgs = CreateWaitingRoomArgs(
        clientId,
        sanitizedBet,
        escrowId: serverIsF2P ? null : escrowId,
      );

      developer.log("CreateWaitingRoom args: $createRoomArgs");
      var roomInfo = await Golib.CreateWaitingRoom(createRoomArgs);

      // Update the model state immediately
      currentWR = roomInfo;
      // Ensure the new room appears in the list without waiting for ntfn refresh
      final idx = waitingRooms.indexWhere((r) => r.id == roomInfo.id);
      if (idx == -1) {
        waitingRooms = [roomInfo, ...waitingRooms];
      } else {
        waitingRooms[idx] = roomInfo;
      }
      _currentGameState = GameState.inWaitingRoom;
      errorMessage = '';
      notifyListeners();

      notificationModel.showNotification(
        "Waiting room created with Bet Amount: ${roomInfo.betAmt}",
      );
    } catch (e) {
      errorMessage = "Error creating waiting room: $e";
      developer.log("Error creating waiting room: $e");
      notifyListeners();
    }
  }

  Future<void> joinWaitingRoom(String id) async {
    try {
      if (!serverIsF2P) {
        if (escrowId.isEmpty) {
          errorMessage = "Open escrow first in Settings → Settlement panel";
          notifyListeners();
          return;
        }
        if (!escrowFunded) {
          errorMessage = "Fund escrow before joining a room";
          notifyListeners();
          return;
        }
      }
      // Use the returned room info to immediately update UI state.
      final roomInfo = await Golib.JoinWaitingRoom(id,
          escrowId: serverIsF2P ? null : escrowId);

      // Set current room and ensure list reflects joined state without
      // waiting for async notifications.
      currentWR = roomInfo;
      final idx = waitingRooms.indexWhere((r) => r.id == roomInfo.id);
      if (idx == -1) {
        waitingRooms = [roomInfo, ...waitingRooms];
      } else {
        waitingRooms[idx] = roomInfo;
      }

      _currentGameState = GameState.inWaitingRoom;
      errorMessage = '';
      notifyListeners();
    } catch (e) {
      errorMessage = "Error joining waiting room: $e";
      notifyListeners();
    }
  }

  void toggleReady() async {
    if (currentWR == null) {
      var error = "Need to get into a waiting room to get ready.";
      errorMessage = error;
      notifyListeners();
      throw Exception(error);
    }

    if (_currentGameState != GameState.waitingRoomReady) {
      // Player is getting ready - start game stream via golib
      try {
        await Golib.startGameStream();
        _currentGameState = GameState.waitingRoomReady;
        // Game updates will come via golib notifications
        // Start render loop now that the stream is active
        renderLoop.start();
      } catch (error) {
        developer.log("Error starting game stream: $error");
        errorMessage = "Error starting game stream: $error";
        notifyListeners();
        return;
      }
    } else {
      // Player is unreadying
      try {
        await Golib.unreadyGameStream();
        _currentGameState = GameState.inWaitingRoom;
        _stopGameStreamAndRenderLoop();
      } catch (error) {
        developer.log("Error in unready game stream: $error");
        errorMessage = "Error in unready game stream: $error";
        notifyListeners();
        return;
      }
    }

    notifyListeners();
  }

  Future<void> leaveWaitingRoom() async {
    if (currentWR == null) {
      return;
    }

    try {
      await Golib.LeaveWaitingRoom(currentWR!.id);

      // Reset waiting room state and escrow so a new match can be started
      currentWR = null;
      _currentGameState = GameState.idle;
      errorMessage = '';
      _stopGameStreamAndRenderLoop();
      notifyListeners();

      notificationModel.showNotification("Left waiting room successfully");
    } catch (e) {
      errorMessage = "Error leaving waiting room: $e";
      developer.log("Error leaving waiting room: $e");
      notifyListeners();
    }
  }

  // Signal that the player is ready to play
  Future<void> signalReadyToPlay() async {
    try {
      if (currentGameId.isEmpty) {
        errorMessage = "No active game found";
        notifyListeners();
        return;
      }

      final success = await Golib.signalReadyToPlay(currentGameId);

      if (success) {
        _currentGameState = GameState.readyToPlay;
        notificationModel.showNotification("You are ready to play!");
      } else {
        errorMessage = "Failed to signal ready to play";
      }

      notifyListeners();
    } catch (e) {
      errorMessage = "Error signaling ready to play: $e";
      notifyListeners();
    }
  }

  // Sample current interpolated frame for rendering
  GameUpdate sampleInterpolatedGameState() {
    // Sample interpolated state with a small render delay to hide jitter.
    final s = interpolator.sample();
    if (s.gameWidth > 0 && s.gameHeight > 0) return s;
    // Fallback before first frame
    return GameUpdate()
      ..gameWidth = 800
      ..gameHeight = 600;
  }

  void _stopGameStreamAndRenderLoop() {
    renderLoop.stop();
    gameState = null;
  }

  // Refund-related methods
  Future<void> loadHistoricEscrows() async {
    isLoadingHistoricEscrows = true;
    historicEscrowsError = '';
    notifyListeners();

    try {
      // Ensure the minimal client is initialized (prelogin mode is enough).
      await ensurePreloginInitialized();
      developer.log(
        'loadHistoricEscrows: start',
        name: 'refunds',
      );
      List<Map<String, dynamic>> allEscrows = [];
      try {
        allEscrows = await Golib.listHistoricEscrows();
      } catch (e) {
        rethrow;
      }

      allEscrows.sort((a, b) {
        final aTime =
            (a['archived_at'] is num) ? (a['archived_at'] as num).toInt() : 0;
        final bTime =
            (b['archived_at'] is num) ? (b['archived_at'] as num).toInt() : 0;
        return bTime.compareTo(aTime);
      });

      historicEscrows = allEscrows
          .map((escrow) => Map<String, dynamic>.from(escrow))
          .toList();

      developer.log(
        'loadHistoricEscrows: found ${historicEscrows.length} historic escrows',
        name: 'refunds',
      );
      isLoadingHistoricEscrows = false;
      notifyListeners();
    } catch (e) {
      developer.log('loadHistoricEscrows error: $e', name: 'refunds');
      isLoadingHistoricEscrows = false;
      historicEscrowsError = 'Error loading historic escrows: $e';
      notifyListeners();
    }
  }

  Future<Map<String, dynamic>> buildRefundTransaction(
      String escrowId, String destAddr,
      {int feeAtoms = 20000, int? csvBlocks, int? utxoValue}) async {
    try {
      final result = await Golib.refundEscrow(
        escrowId: escrowId,
        destAddr: destAddr,
        feeAtoms: feeAtoms,
        csvBlocks: csvBlocks ?? CSV_BLOCKS,
        utxoValue: utxoValue,
      );

      developer.log(
        'buildRefundTransaction: escrow=$escrowId can_refund=${result['can_refund']}',
        name: 'refunds',
      );

      return result;
    } catch (e) {
      throw Exception('Failed to build refund transaction: $e');
    }
  }

  // Update escrow funding transaction info in historic file
  Future<void> updateEscrowFundingTx(
      String escrowId, String txid, int vout) async {
    try {
      developer.log(
        'updateEscrowFundingTx: escrow=$escrowId txid=$txid vout=$vout',
        name: 'refunds',
      );

      await Golib.updateHistoricEscrow({
        'escrow_id': escrowId,
        'funding_txid': txid,
        'funding_vout': vout,
      });

      developer.log(
        'updateEscrowFundingTx: successfully updated escrow $escrowId',
        name: 'refunds',
      );
    } catch (e) {
      developer.log(
        'updateEscrowFundingTx error: $e',
        name: 'refunds',
      );
      throw Exception('Failed to update escrow funding transaction: $e');
    }
  }

  Future<void> deleteHistoricEscrow(String escrowId) async {
    try {
      developer.log('deleteHistoricEscrow: escrow=$escrowId', name: 'refunds');

      await Golib.deleteHistoricEscrow(escrowId);

      historicEscrows.removeWhere((escrow) {
        final id = escrow['escrow_id']?.toString() ?? '';
        return id == escrowId;
      });

      notifyListeners();

      developer.log('deleteHistoricEscrow: deleted $escrowId', name: 'refunds');
    } catch (e) {
      developer.log('deleteHistoricEscrow error: $e', name: 'refunds');
      throw Exception('Failed to delete escrow: $e');
    }
  }
}
