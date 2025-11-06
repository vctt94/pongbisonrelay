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
import 'package:golib_plugin/grpc/generated/pong.pbgrpc.dart';
import 'package:pongui/models/notifications.dart';
import 'package:path/path.dart' as path;

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
  bool escrowFunded = false;    // true when deposit seen (0-conf OK)
  bool escrowConfirmed = false; // true when at least 1 confirmation
  String errorMessage = '';
  List<LocalWaitingRoom> waitingRooms = [];
  LocalWaitingRoom? currentWR;
  GameUpdate? gameState;
  final SnapshotInterpolator interpolator = SnapshotInterpolator();
  final RenderLoop renderLoop = RenderLoop();
  StreamSubscription<GameUpdateBytes>? _gameStreamSub;
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

  PongModel(this.cfg, this.notificationModel);

  // Initialize golib PongClient after wallet authentication (requires clientId)
  Future<void> _initPongClient(Config cfg) async {
    try {
      if (clientId.isEmpty) {
        throw Exception("clientId is required - wallet authentication must be completed first");
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
        cfg.rpcPass,
      );

      developer.log("InitClient args: $initArgs");

      var localInfo = await Golib.initClient(initArgs);

      // clientId should match what we passed in
      if (localInfo.id != clientId) {
        throw Exception("clientId mismatch: expected $clientId, got ${localInfo.id}");
      }
      nick = localInfo.nick;

      // Query initial waiting rooms via golib
      waitingRooms = await Golib.getWaitingRooms();
      
      // Initialize game client (notifications come via golib now)
      print("Initializing game client with clientId: $clientId");
      pongGame = PongGame(clientId);

      isConnected = true;
      // Subscribe to UI notifications forwarded by golib (also carries structured events)
      _uiNtfnSub ??= Golib.uiNotifications().listen((n) {
        try {
          // Handle structured bet update events piggybacked via UINotification
          if (n.type == 'bet_update') {
            // Extras encoded inside the 'from' field as JSON
            try {
              final extras = n.from.isNotEmpty ? jsonDecode(n.from) : null;
              if (extras is Map<String, dynamic>) {
                final pid = (extras['player_id'] ?? '').toString();
                if (pid == clientId) {
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
              }
            } catch (_) {}
          } else if (n.type == 'game_started' || n.type == 'gamestarted') {
            // Structured game start event (extras in from)
            try {
              final extras = n.from.isNotEmpty ? jsonDecode(n.from) : null;
              if (extras is Map<String, dynamic>) {
                final gid = (extras['game_id'] ?? '').toString();
                if (gid.isNotEmpty) currentGameId = gid;
              }
            } catch (_) {}
            if (_currentGameState == GameState.idle ||
                _currentGameState == GameState.inWaitingRoom ||
                _currentGameState == GameState.waitingRoomReady) {
              _currentGameState = GameState.gameInitialized;
            }
            currentWR = null; // no longer in a waiting room
            notifyListeners();
          } else if (n.type == 'game_ready_to_play') {
            try {
              final extras = n.from.isNotEmpty ? jsonDecode(n.from) : null;
              if (extras is Map<String, dynamic>) {
                final gid = (extras['game_id'] ?? '').toString();
                if (gid.isNotEmpty) currentGameId = gid;
              }
            } catch (_) {}
            // Server says game is ready: show "I'm Ready!" overlay
            _currentGameState = GameState.gameInitialized;
            notificationModel.showNotification("Game is ready!");
            notifyListeners();
          } else if (n.type == 'countdown_update') {
            // Server countdown prior to playing
            String msg = n.text.trim();
            if (msg.isEmpty) {
              try {
                final extras = n.from.isNotEmpty ? jsonDecode(n.from) : null;
                if (extras is Map<String, dynamic>) {
                  msg = (extras['message'] ?? '').toString();
                }
              } catch (_) {}
            }
            countdownMessage = msg;
            _currentGameState = GameState.countdown;
            if (msg.contains('0')) {
              _currentGameState = GameState.playing;
            }
            notifyListeners();
          } else if (n.type == 'game_end') {
            // Stop local stream and show end-of-game overlay
            _stopGameStreamAndRenderLoop();
            gameEndingMessage = n.text.isNotEmpty ? n.text : 'Game ended';
            _currentGameState = GameState.gameEnded;
            notifyListeners();
          } else if (n.type == 'player_ready') {
            try {
              final extras = n.from.isNotEmpty ? jsonDecode(n.from) : null;
              if (extras is Map<String, dynamic>) {
                final pid = (extras['player_id'] ?? '').toString();
                final r = extras['ready'] == true;
                if (pid == clientId && r) {
                  _currentGameState = GameState.waitingRoomReady;
                  notifyListeners();
                }
              }
            } catch (_) {}
          } else if (n.type == 'wr_created') {
            try {
              final extras = n.from.isNotEmpty ? jsonDecode(n.from) : null;
              if (extras is Map<String, dynamic>) {
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
            } catch (_) {}
          } else if (n.type == 'wr_removed') {
            try {
              final extras = n.from.isNotEmpty ? jsonDecode(n.from) : null;
              if (extras is Map<String, dynamic>) {
                final rid = (extras['room_id'] ?? '').toString();
                if (rid.isNotEmpty) {
                  waitingRooms = waitingRooms.where((r) => r.id != rid).toList(growable: false);
                  if (currentWR?.id == rid) currentWR = null;
                  notifyListeners();
                }
              }
            } catch (_) {}
          } else if (n.type == 'player_joined_wr' || n.type == 'player_left_wr') {
            try {
              final extras = n.from.isNotEmpty ? jsonDecode(n.from) : null;
              if (extras is Map<String, dynamic>) {
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
            } catch (_) {}
          } else if (n.type == 'game_update') {
            try {
              final extras = n.from.isNotEmpty ? jsonDecode(n.from) : null;
              if (extras is Map<String, dynamic>) {
                final gu = GameUpdate()
                  ..gameWidth  = (extras['game_width'] as num?)?.toDouble() ?? gameState?.gameWidth ?? 800
                  ..gameHeight = (extras['game_height'] as num?)?.toDouble() ?? gameState?.gameHeight ?? 600
                  ..p1X = (extras['p1x'] as num?)?.toDouble() ?? 0
                  ..p1Y = (extras['p1y'] as num?)?.toDouble() ?? 0
                  ..p1Width  = (extras['p1w'] as num?)?.toDouble() ?? 0
                  ..p1Height = (extras['p1h'] as num?)?.toDouble() ?? 0
                  ..p2X = (extras['p2x'] as num?)?.toDouble() ?? 0
                  ..p2Y = (extras['p2y'] as num?)?.toDouble() ?? 0
                  ..p2Width  = (extras['p2w'] as num?)?.toDouble() ?? 0
                  ..p2Height = (extras['p2h'] as num?)?.toDouble() ?? 0
                  ..ballX = (extras['ballx'] as num?)?.toDouble() ?? 0
                  ..ballY = (extras['bally'] as num?)?.toDouble() ?? 0
                  ..ballWidth  = (extras['ballw'] as num?)?.toDouble() ?? 0
                  ..ballHeight = (extras['ballh'] as num?)?.toDouble() ?? 0
                  ..p1Score = (extras['p1score'] as num?)?.toInt() ?? (gameState?.p1Score ?? 0)
                  ..p2Score = (extras['p2score'] as num?)?.toInt() ?? (gameState?.p2Score ?? 0);
                gameState = gu;
                interpolator.push(gu);
                renderLoop.start();
              }
            } catch (_) {}
          } else if ((n.text).isNotEmpty) {
            // Show toast for human UI notifications
            notificationModel.showNotification(n.text);
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
  Future<void> applyWalletAuth({required String newClientId, String token = '', String address = '', String p2pkAddr = ''}) async {
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
    // Stop UI notifications subscription
    _uiNtfnSub?.cancel();
    _uiNtfnSub = null;
    notifyListeners();
  }

  // TODO: Notifications now come via golib automatically (forwarded from Go PongClient)
  // This notification handling logic needs to be connected to golib's notification system
  // For now, notifications from golib will be in raw protobuf format via the command result loop
  // The following method contains the logic for handling notifications when they arrive:
  /*
  void _handleIncomingNotification(dynamic ntfn) {
    // This method should be called when notifications arrive from golib
    //grpcClient.startNtfnStreamRequest(clientId).listen((ntfn) {
      developer.log("Notification Stream Response: $ntfn");

      isConnected = true;
      notifyListeners();

      switch (ntfn.notificationType) {
        case NotificationType.MESSAGE:
          // Avoid deriving funding state from free-form messages; show as toast only.
          notificationModel.showNotification(ntfn.message);
          notifyListeners();
          break;
        case NotificationType.BET_AMOUNT_UPDATE:
          if (ntfn.playerId == clientId) {
            betAmt = ntfn.betAmt.toInt();
            escrowConfs = ntfn.confs;
            // Consider escrow funded whenever a watcher-driven update arrives (may be 0-conf)
            escrowFunded = ntfn.confs >= 0;
            // Confirmed when at least 1 conf
            escrowConfirmed = ntfn.confs >= 1;
            // Optional: textual status for tooltip only, derived from structured confs
            fundingStatus = escrowConfirmed
                ? 'Deposit confirmed (${ntfn.confs})'
                : 'Deposit seen (mempool)';
            notifyListeners();
          }
          break;

        case NotificationType.ON_WR_CREATED:
          // Refresh rooms in background (can't await in this listener)
          Golib.getWaitingRooms().then((rooms) {
            waitingRooms = rooms;
            notifyListeners();
          }).catchError((_) {
            // Fallback: append from notification payload
            waitingRooms.add(LocalWaitingRoom.fromProto(ntfn.wr));
            notifyListeners();
          });
          notificationModel.showNotification(
            "Waiting room created by ${ntfn.wr.hostId}",
          );
          break;

        case NotificationType.GAME_START:
          if (_currentGameState == GameState.idle ||
              _currentGameState == GameState.inWaitingRoom ||
              _currentGameState == GameState.waitingRoomReady) {
            _currentGameState = GameState.gameInitialized;
          }
          // can set current wr as null after game starting
          currentWR = null;
          notificationModel.showNotification(
            "Game started with ID: ${ntfn.gameId}",
          );
          notifyListeners();
          break;

        case NotificationType.GAME_READY_TO_PLAY:
          // Store the game ID when we receive the ready to play notification
          currentGameId = ntfn.gameId;
          if (_currentGameState == GameState.idle ||
              _currentGameState == GameState.inWaitingRoom ||
              _currentGameState == GameState.waitingRoomReady) {
            _currentGameState = GameState.gameInitialized;
          }
          notificationModel.showNotification(
              "Game is ready! Signal when you're ready to play.");
          notifyListeners();
          break;

        case NotificationType.COUNTDOWN_UPDATE:
          countdownMessage = ntfn.message;
          _currentGameState = GameState.countdown;

          // When countdown reaches 0, transition to playing state
          if (ntfn.message.contains("0")) {
            _currentGameState = GameState.playing;
          }

          notificationModel.showNotification(ntfn.message);
          notifyListeners();
          break;

        case NotificationType.PLAYER_JOINED_WR:
          if (ntfn.playerId == clientId) {
            currentWR = LocalWaitingRoom.fromProto(ntfn.wr);
            _currentGameState = GameState.inWaitingRoom;
          }
          notificationModel
              .showNotification("A new player joined the waiting room");
          notifyListeners();
          break;

        case NotificationType.GAME_END:
          // Store the ending message and transition to game ended state
          gameEndingMessage = ntfn.message;
          _currentGameState = GameState.gameEnded;
          notificationModel.showNotification(ntfn.message);
          // Stop the game stream and render loop
          _stopGameStreamAndRenderLoop();
          notifyListeners();
          break;

        case NotificationType.ON_WR_REMOVED:
          // Handle the waiting room removal
          waitingRooms.removeWhere((room) => room.id == ntfn.roomId);

          // If we were in this waiting room, reset our state
          if (currentWR != null && currentWR!.id == ntfn.roomId) {
            currentWR = null;
            _currentGameState = GameState.idle;
          }

          notificationModel.showNotification(
            "Waiting room removed: ${ntfn.roomId}",
          );
          notifyListeners();
          break;

        case NotificationType.OPPONENT_DISCONNECTED:
          if (_currentGameState == GameState.playing) {
            _currentGameState = GameState.idle;
          }
          currentWR = LocalWaitingRoom.fromProto(ntfn.wr);
          notificationModel.showNotification(ntfn.message);
          notifyListeners();
          // Ensure we stop local rendering when opponent disconnects
          _stopGameStreamAndRenderLoop();
          break;

        case NotificationType.ON_PLAYER_READY:
          // Check if this is a ready to play notification for game
          if (ntfn.gameId.isNotEmpty) {
            String playerName =
                ntfn.playerId == clientId ? "You are" : "Opponent is";
            notificationModel.showNotification("$playerName ready to play!");

            // If this is our own ready signal, update our state
            if (ntfn.playerId == clientId) {
              _currentGameState = GameState.readyToPlay;
            }
          }
          // Otherwise handle waiting room ready state
          else if (currentWR != null) {
            // Find the player in the current waiting room and update their ready status
            for (var i = 0; i < currentWR!.players.length; i++) {
              if (currentWR!.players[i].uid == ntfn.playerId) {
                currentWR!.players[i].ready = ntfn.ready;

                // If this is our own ready signal, update our state
                if (ntfn.playerId == clientId) {
                  _currentGameState = ntfn.ready
                      ? GameState.waitingRoomReady
                      : GameState.inWaitingRoom;
                }
                break;
              }
            }

            // Show notification
            String playerName = ntfn.playerId;
            String readyStatus = ntfn.ready ? "ready" : "not ready";
            notificationModel.showNotification(
              "Player $playerName is now $readyStatus",
            );
          }
          notifyListeners();
          break;

        default:
          developer.log("Unknown notification type: ${ntfn.notificationType}");
      }
    //}, onError: (error) {
    //  errorMessage = "Error in notification stream: ${error.message}";
    //  developer.log("Error: $error");
    //  // XXX this is not correct, need to check if error is eof
    //  isConnected = false;
    //  notifyListeners();
    //});
  }
  */

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
      if (betAmt <= 0) {
        errorMessage = "bet amount needs to be higher than 0";
        notifyListeners();
        return;
      }
      if (escrowId.isEmpty) {
        errorMessage = "Open escrow first in Settings → Settlement panel";
        notifyListeners();
        return;
      }
      if (!escrowFunded) {
        errorMessage = "Wait until escrow deposit is seen before creating a room";
        notifyListeners();
        return;
      }

      CreateWaitingRoomArgs createRoomArgs =
        CreateWaitingRoomArgs(clientId, betAmt, escrowId: escrowId);

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
      await Golib.JoinWaitingRoom(id, escrowId: escrowId);
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
    return interpolator.sample();
  }

  void _stopGameStreamAndRenderLoop() {
    _gameStreamSub?.cancel();
    _gameStreamSub = null;
    renderLoop.stop();
  }
}
