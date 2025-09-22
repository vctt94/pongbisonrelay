import 'package:flutter/material.dart';
import 'package:pongui/components/waiting_rooms.dart';
import 'package:pongui/models/pong.dart';

class MainContent extends StatelessWidget {
  final PongModel pongModel;

  MainContent({
    Key? key,
    required this.pongModel,
  }) : super(key: key);

  // Map the PongModel's game state to appropriate UI state
  Widget _buildContentForState(BuildContext context) {
    // Render the appropriate UI based on the current state
    switch (pongModel.currentGameState) {
      case GameState.idle:
        return _buildWelcomeState(context);

      case GameState.inWaitingRoom:
        return _buildWelcomeState(context);

      case GameState.waitingRoomReady:
        return _buildReadyToStartState();

      case GameState.gameInitialized:
        return _buildGameState(GameState.gameInitialized);

      case GameState.readyToPlay:
        return _buildGameState(GameState.readyToPlay);

      case GameState.countdown:
        return _buildGameState(GameState.countdown);

      case GameState.playing:
        return _buildGameState(GameState.playing);
    }
  }

  @override
  Widget build(BuildContext context) {
    return _buildContentForState(context);
  }

  // Welcome state UI
  Widget _buildWelcomeState(BuildContext context) {
    return SingleChildScrollView(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        children: [
          // Welcome section
          const SizedBox(height: 40),
          const Text(
            "Welcome to Pong!",
            style: TextStyle(
              fontSize: 32,
              fontWeight: FontWeight.bold,
              color: Colors.blueAccent,
            ),
          ),
          const SizedBox(height: 16),
          const Padding(
            padding: EdgeInsets.symmetric(horizontal: 24.0),
            child: Text(
              "Open escrow to place a bet. Configure payout in Settings.",
              textAlign: TextAlign.center,
              style: TextStyle(
                fontSize: 16,
                color: Colors.white,
                height: 1.4,
              ),
            ),
          ),

          // Waiting rooms or empty state
          if (pongModel.waitingRooms.isEmpty)
            Center(
              child: Column(
                mainAxisAlignment: MainAxisAlignment.center,
                children: [
                  const SizedBox(height: 40),
                  Container(
                    padding: const EdgeInsets.all(20),
                    decoration: const BoxDecoration(
                      color: Color.fromRGBO(27, 30, 44, 0.6),
                      shape: BoxShape.circle,
                    ),
                    child: Icon(
                      Icons.sports_esports,
                      size: 64,
                      color: Colors.grey.shade400,
                    ),
                  ),
                  const SizedBox(height: 24),
                  const Text(
                    'No active waiting rooms',
                    style: TextStyle(
                      fontSize: 22,
                      fontWeight: FontWeight.w500,
                      color: Colors.white,
                    ),
                  ),
                  const SizedBox(height: 8),
                  const Text(
                    'Create a room to start playing!',
                    style: TextStyle(
                      fontSize: 16,
                      color: Colors.white70,
                    ),
                  ),
                ],
              ),
            )
          else
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 12.0),
              child: WaitingRoomList(
                pongModel.waitingRooms,
                currentRoomId: pongModel.currentWR?.id,
                onJoinRoom: (roomId) => pongModel.joinWaitingRoom(roomId),
              ),
            ),

        ],
      ),
    );
  }

  // Ready to start state UI - waiting for game to begin
  Widget _buildReadyToStartState() {
    return const Center(
      child: Column(
        mainAxisAlignment: MainAxisAlignment.center,
        children: [
          Icon(
            Icons.sports_tennis,
            size: 60,
            color: Colors.blueAccent,
          ),
          SizedBox(height: 16),
          Text(
            "Waiting for players...",
            style: TextStyle(
              fontSize: 18,
              color: Colors.white,
              fontWeight: FontWeight.w500,
            ),
          ),
        ],
      ),
    );
  }

  // Game state UI with appropriate overlay based on the current sub-state
  Widget _buildGameState(GameState state) {
    // Basic check for game state availability
    if (pongModel.gameState == null) {
      return const Center(
        child: CircularProgressIndicator(
          valueColor: AlwaysStoppedAnimation<Color>(Colors.blueAccent),
        ),
      );
    }

    // For playing state, return a direct game widget without any overlays
    if (state == GameState.playing) {
      return LayoutBuilder(builder: (context, constraints) {
        final gw = pongModel.gameState!.gameWidth > 0 ? pongModel.gameState!.gameWidth : 800.0;
        final gh = pongModel.gameState!.gameHeight > 0 ? pongModel.gameState!.gameHeight : 600.0;
        final maxW = constraints.maxWidth;
        final maxH = constraints.maxHeight;
        final scale = (maxH.isFinite)
            ? (maxW / gw).clamp(0.0, double.infinity).compareTo((maxH / gh).clamp(0.0, double.infinity)) <= 0
                ? (maxW / gw)
                : (maxH / gh)
            : (maxW / gw);
        final childW = gw * scale;
        final childH = gh * scale;
        return SizedBox(
          width: maxW,
          height: maxH.isFinite ? maxH : childH,
          child: Center(
            child: SizedBox(
              width: childW,
              height: childH,
              child: Container(
                color: Colors.black,
                child: pongModel.pongGame.buildWidget(
                  pongModel.gameState!,
                  FocusNode(),
                  onReadyHotkey: () => pongModel.signalReadyToPlay(),
                ),
              ),
            ),
          ),
        );
      });
    }

    List<Widget> stackChildren = [
      // Black background layer
      Container(color: Colors.black),

      // Base game widget - always show the game state when available
      Positioned.fill(
        child: pongModel.pongGame.buildWidget(
          pongModel.gameState!,
          FocusNode(),
          onReadyHotkey: () => pongModel.signalReadyToPlay(),
        ),
      ),
    ];

    // Add the appropriate overlay based on the game state
    // ONLY add overlays for specific states, not for playing state
    if (state == GameState.gameInitialized) {
      stackChildren.add(_buildReadyToPlayOverlay(state));
    } else if (state == GameState.readyToPlay) {
      stackChildren.add(_buildWaitingForPlayersOverlay());
    } else if (state == GameState.countdown) {
      stackChildren.add(_buildCountdownOverlay(state));
    }

    return LayoutBuilder(builder: (context, constraints) {
      final gw = pongModel.gameState!.gameWidth > 0 ? pongModel.gameState!.gameWidth : 800.0;
      final gh = pongModel.gameState!.gameHeight > 0 ? pongModel.gameState!.gameHeight : 600.0;
      final maxW = constraints.maxWidth;
      final maxH = constraints.maxHeight;
      final scale = (maxH.isFinite)
          ? (maxW / gw).clamp(0.0, double.infinity).compareTo((maxH / gh).clamp(0.0, double.infinity)) <= 0
              ? (maxW / gw)
              : (maxH / gh)
          : (maxW / gw);
      final childW = gw * scale;
      final childH = gh * scale;
      return SizedBox(
        width: maxW,
        height: maxH.isFinite ? maxH : childH,
        child: Center(
          child: SizedBox(
            width: childW,
            height: childH,
            child: Stack(children: stackChildren),
          ),
        ),
      );
    });
  }

  // Ready to play overlay - shows "I'm Ready" button
  Widget _buildReadyToPlayOverlay(GameState state) {
    return Positioned.fill(
      child: Container(
        color: const Color.fromRGBO(0, 0, 0, 0.5),
        child: Material(
          type: MaterialType.transparency,
          child: Builder(
            builder: (context) => pongModel.pongGame.buildReadyToPlayOverlay(
              context,
              state == GameState.readyToPlay, // Only show as ready in readyToPlay state
              state == GameState.countdown,   // Only show countdown in countdown state
              '', // No countdown message
              () => pongModel.signalReadyToPlay(),
              pongModel.gameState!,
            ),
          ),
        ),
      ),
    );
  }

  // Waiting for players overlay - after pressing "I'm Ready"
  Widget _buildWaitingForPlayersOverlay() {
    return Positioned.fill(
      child: Container(
        color: const Color.fromRGBO(0, 0, 0, 0.5),
        child: Material(
          type: MaterialType.transparency,
          child: Center(
            child: Container(
              padding: const EdgeInsets.all(20),
              decoration: BoxDecoration(
                color: const Color(0xFF1B1E2C).withAlpha(230),
                borderRadius: BorderRadius.circular(15),
                boxShadow: [
                  BoxShadow(
                    color: Colors.blueAccent.withAlpha(76),
                    spreadRadius: 3,
                    blurRadius: 10,
                  ),
                ],
              ),
              child: Column(
                mainAxisSize: MainAxisSize.min,
                children: const [
                  Icon(
                    Icons.sports_esports,
                    size: 50,
                    color: Colors.blueAccent,
                  ),
                  SizedBox(height: 20),
                  Text(
                    "Waiting for players to get ready...",
                    style: TextStyle(
                      color: Colors.white,
                      fontSize: 24,
                      fontWeight: FontWeight.bold,
                    ),
                  ),
                  SizedBox(height: 20),
                  SizedBox(
                    width: 40,
                    height: 40,
                    child: CircularProgressIndicator(
                      color: Colors.blueAccent,
                      backgroundColor: Colors.grey,
                      strokeWidth: 4,
                    ),
                  ),
                ],
              ),
            ),
          ),
        ),
      ),
    );
  }

  // Countdown overlay
  Widget _buildCountdownOverlay(GameState state) {
    return Positioned.fill(
      child: Container(
        color: const Color.fromRGBO(0, 0, 0, 0.5),
        child: Material(
          type: MaterialType.transparency,
          child: Builder(
            builder: (context) => pongModel.pongGame.buildReadyToPlayOverlay(
              context,
              true, // Always show as ready during countdown
              true, // In countdown
              pongModel.countdownMessage,
              () {}, // No action during countdown
              pongModel.gameState!,
            ),
          ),
        ),
      ),
    );
  }
}
