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
  int escrowBetAtoms = 0;
  String fundingStatus = '';
  int escrowConfs = 0;
  // Escrow funding flags derived from notifications
  bool escrowFunded = false; // true when deposit seen (0-conf OK)
  bool escrowConfirmed = false; // true when at least 1 confirmation
  String errorMessage = '';
  List<LocalWaitingRoom> waitingRooms = [];
  LocalWaitingRoom? currentWR;
  GameUpdate? gameState;
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

  // Track last settlement match id used for presign so we can archive the
  // session key safely after game completion.
  String lastMatchId = '';

  void setEscrowId(String id) {
    escrowId = id;
    notifyListeners();
  }

  void setEscrowDetails(String id, String depositAddr, [String? pkScriptHex]) {
    escrowId = id;
    escrowDepositAddress = depositAddr;
    escrowPkScriptHex = pkScriptHex ?? escrowPkScriptHex;
    notifyListeners();
  }

  void setEscrowBetAtoms(int atoms) {
    escrowBetAtoms = atoms;
    // Reflect intended bet in UI immediately
    betAmt = atoms;
    notifyListeners();
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
  Future<void> _initPongClient(Config cfg) async {
    try {
      if (clientId.isEmpty) {
        throw Exception(
            "clientId is required - wallet authentication must be completed first");
      }
      if (isConnected) {
        return; // Already initialized
      }

      final appDataDir = await defaultAppDataDir();
      final logFilePath = path.join(appDataDir, "logs", "pongui.log");

      // Let golib load the authoritative BR config from disk; pass UI config as overrides only.
      // Pass wallet-authenticated clientId as required parameter.
      InitClient initArgs = InitClient(
          clientId, // Wallet-authenticated clientID (required)
          cfg.serverAddr,
          cfg.grpcCertPath,
          appDataDir,
          logFilePath,
          "",
          cfg.debugLevel,
          cfg.wantsLogNtfns,
          cfg.rpcWebsocketURL,
          cfg.rpcCertPath,
          cfg.rpcClientCertPath,
          cfg.rpcClientKeyPath,
          cfg.rpcUser,
          cfg.rpcPass);

      developer.log("InitClient args: $initArgs");

      var localInfo = await Golib.initClient(initArgs);

      // clientId should match what we passed in
      if (localInfo.id != clientId) {
        throw Exception(
            "clientId mismatch: expected $clientId, got ${localInfo.id}");
      }
      nick = localInfo.nick;
      serverIsF2P = localInfo.serverIsF2P;
      serverVersion = localInfo.serverVersion ?? "";

      // Query initial waiting rooms via golib
      waitingRooms = await Golib.getWaitingRooms();

      // Initialize game client (notifications come via golib now)
      print("Initializing game client with clientId: $clientId");
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

  void logout() {
    isWalletAuthenticated = false;
    walletAddress = '';
    authToken = '';
    payoutAddressOrPubkey = '';
    escrowId = '';
    escrowDepositAddress = '';
    escrowPkScriptHex = '';
    escrowBetAtoms = 0;
    escrowFunded = false;
    escrowConfirmed = false;
    currentWR = null;
    _currentGameState = GameState.idle;
    serverIsF2P = false;
    // Stop UI notifications subscription
    _uiNtfnSub?.cancel();
    _uiNtfnSub = null;
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
    if (msg.contains('0')) {
      _currentGameState = GameState.playing;
    }
    notifyListeners();
  }

  void _handleNtfnGameEnd(UINotification n) {
    _stopGameStreamAndRenderLoop();
    gameEndingMessage = n.text.isNotEmpty ? n.text : 'Game ended';
    _currentGameState = GameState.gameEnded;
    notifyListeners();
  }

  void _handleNtfnPlayerReady(UINotification n) {
    final extras = _extrasFrom(n);
    final pid = (extras['player_id'] ?? '').toString();
    final r = extras['ready'] == true;
    if (pid == clientId && r) {
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

  // Removed JSON fallback for game updates. Binary-only via gameUpdates().

  void _handleGameUpdateFrame(GameUpdate gu) {
    // No interpolation: keep only the latest authoritative state.
    gameState = gu;
    renderLoop.start();
  }

  void resetGameState() {
    _currentGameState = GameState.idle;
    currentWR = null;
    betAmt = 0;
    currentGameId = '';
    countdownMessage = '';
    gameEndingMessage = '';
    clearEscrowState();
    _stopGameStreamAndRenderLoop();
    notifyListeners();
  }

  // Clear all escrow-related client state so user can open a fresh escrow
  // after a game ends or when leaving a room.
  void clearEscrowState() {
    escrowId = '';
    escrowDepositAddress = '';
    escrowPkScriptHex = '';
    escrowBetAtoms = 0;
    escrowFunded = false;
    escrowConfirmed = false;
    escrowConfs = 0;
    fundingStatus = '';
    // Also archive the persisted session key so a new escrow can use a new key
    // while retaining recovery data for this match.
    if (lastMatchId.isNotEmpty) {
      Golib.archiveSettlementSessionKey(lastMatchId);
    }
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
    // No interpolation: render latest directly, with safe fallback
    final gs = gameState;
    if (gs != null && gs.gameWidth > 0 && gs.gameHeight > 0) return gs;
    return GameUpdate()
      ..gameWidth = 800
      ..gameHeight = 600;
  }

  void _stopGameStreamAndRenderLoop() {
    renderLoop.stop();
    gameState = null;
  }
}
