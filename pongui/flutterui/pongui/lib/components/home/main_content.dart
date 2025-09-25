// main_content.dart
import 'package:flutter/material.dart';
import 'package:golib_plugin/grpc/generated/pong.pb.dart';
import 'package:pongui/components/waiting_rooms.dart';
import 'package:pongui/models/pong.dart';

class MainContent extends StatefulWidget {
  final PongModel pongModel;

  const MainContent({Key? key, required this.pongModel}) : super(key: key);

  @override
  State<MainContent> createState() => _MainContentState();
}

class _MainContentState extends State<MainContent> {
  final FocusNode _gameFocus = FocusNode();

  @override
  void dispose() {
    _gameFocus.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return _buildContentForState(context);
  }

  // Map the PongModel's game state to appropriate UI state
  Widget _buildContentForState(BuildContext context) {
    switch (widget.pongModel.currentGameState) {
      case GameState.idle:
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
      case GameState.gameEnded:
        return _buildGameEndedState();
    }
  }

  // Welcome state UI
  Widget _buildWelcomeState(BuildContext context) {
    final model = widget.pongModel;
    return SingleChildScrollView(
      padding: const EdgeInsets.only(bottom: 16),
      child: Column(
        children: [
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
          if (model.waitingRooms.isEmpty)
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
                model.waitingRooms,
                currentRoomId: model.currentWR?.id,
                onJoinRoom: (roomId) => model.joinWaitingRoom(roomId),
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
          Icon(Icons.sports_tennis, size: 60, color: Colors.blueAccent),
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

  // Game ended state UI - shows the final result
  Widget _buildGameEndedState() {
    final model = widget.pongModel;
    final isWin = model.gameEndingMessage.contains("won");
    final isDraw = model.gameEndingMessage.contains("draw");
    
    return Center(
      child: Container(
        padding: const EdgeInsets.all(32),
        margin: const EdgeInsets.symmetric(horizontal: 24),
        decoration: BoxDecoration(
          color: const Color(0xFF1B1E2C).withAlpha(240),
          borderRadius: BorderRadius.circular(20),
          boxShadow: [
            BoxShadow(
              color: (isWin ? Colors.green : isDraw ? Colors.orange : Colors.red)
                  .withAlpha(76),
              spreadRadius: 4,
              blurRadius: 15,
            ),
          ],
        ),
        child: Column(
          mainAxisSize: MainAxisSize.min,
          children: [
            // Game over icon
            Icon(
              isWin ? Icons.emoji_events : isDraw ? Icons.handshake : Icons.sports_tennis,
              size: 80,
              color: isWin ? Colors.green : isDraw ? Colors.orange : Colors.red,
            ),
            const SizedBox(height: 24),
            
            // Game over title
            Text(
              "Game End!",
              style: TextStyle(
                fontSize: 32,
                fontWeight: FontWeight.bold,
                color: isWin ? Colors.green : isDraw ? Colors.orange : Colors.red,
              ),
            ),
            const SizedBox(height: 16),
            
            // Result message
            Text(
              model.gameEndingMessage,
              style: const TextStyle(
                fontSize: 20,
                color: Colors.white,
                fontWeight: FontWeight.w500,
              ),
              textAlign: TextAlign.center,
            ),
            const SizedBox(height: 32),
            
            // Action buttons
            Row(
              mainAxisAlignment: MainAxisAlignment.spaceEvenly,
              children: [
                ElevatedButton.icon(
                  onPressed: () {
                    // Reset game state and return to main menu
                    model.resetGameState();
                  },
                  icon: const Icon(Icons.home),
                  label: const Text("Main Menu"),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.blueAccent,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                  ),
                ),
                ElevatedButton.icon(
                  onPressed: () {
                    // Reset game state but keep escrow for quick rematch
                    model.resetGameState();
                  },
                  icon: const Icon(Icons.refresh),
                  label: const Text("Play Again"),
                  style: ElevatedButton.styleFrom(
                    backgroundColor: Colors.green,
                    foregroundColor: Colors.white,
                    padding: const EdgeInsets.symmetric(horizontal: 24, vertical: 12),
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }

  // Game state UI with appropriate overlay based on the current sub-state
  Widget _buildGameState(GameState state) {
    final model = widget.pongModel;
    // Use a safe fallback before first frame arrives
    final initial = model.gameState ??
        (GameUpdate()
          ..gameWidth = 800
          ..gameHeight = 600);

    // Base game canvas (centered, preserves aspect). We rebuild only when
    // the render loop ticks, sampling the interpolated state to avoid
    // invalidating the entire page per network update.
    Widget gameCanvas = AnimatedBuilder(
      animation: model.renderLoop,
      builder: (context, _) {
        final interpolated = model.sampleInterpolatedGameState();
        final use = (interpolated.gameWidth > 0 && interpolated.gameHeight > 0)
            ? interpolated
            : initial;
        final gw = use.gameWidth > 0 ? use.gameWidth : 800.0;
        final gh = use.gameHeight > 0 ? use.gameHeight : 600.0;
        return Center(
          child: AspectRatio(
            aspectRatio: gw / gh,
            child: Container(
              color: Colors.black,
              child: model.pongGame.buildWidget(
                use,
                _gameFocus,
                onReadyHotkey: () => model.signalReadyToPlay(),
              ),
            ),
          ),
        );
      },
    );

    // Playing: raw canvas only (no overlays)
    if (state == GameState.playing) {
      return SizedBox.expand(child: gameCanvas);
    }

    // Overlays for non-playing sub-states
    final List<Widget> stackChildren = [
      // Base game
      Positioned.fill(child: gameCanvas),

      // Specific overlays
      if (state == GameState.gameInitialized) _buildReadyToPlayOverlay(state, model),
      if (state == GameState.readyToPlay) _buildWaitingForPlayersOverlay(),
      if (state == GameState.countdown) _buildCountdownOverlay(model),
    ];

    return SizedBox.expand(
      child: Stack(
        fit: StackFit.expand,
        children: stackChildren,
      ),
    );
  }

  // Ready to play overlay - shows "I'm Ready" button
  Widget _buildReadyToPlayOverlay(GameState state, PongModel model) {
    return Positioned.fill(
      child: Container(
        color: const Color.fromRGBO(0, 0, 0, 0.5),
        child: Material(
          type: MaterialType.transparency,
          child: Builder(
            builder: (context) => model.pongGame.buildReadyToPlayOverlay(
              context,
              state == GameState.readyToPlay,  // not ready yet in this state
              false,                           // no countdown here
              '',
              () => model.signalReadyToPlay(),
              model.gameState ??
                  (GameUpdate()
                    ..gameWidth = 800
                    ..gameHeight = 600),
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
        child: const Material(
          type: MaterialType.transparency,
          child: Center(
            child: _WaitingForPlayersCard(),
          ),
        ),
      ),
    );
  }

  // Countdown overlay
  Widget _buildCountdownOverlay(PongModel model) {
    return Positioned.fill(
      child: Container(
        color: const Color.fromRGBO(0, 0, 0, 0.5),
        child: Material(
          type: MaterialType.transparency,
          child: Builder(
            builder: (context) => model.pongGame.buildReadyToPlayOverlay(
              context,
              true, // shows as ready during countdown
              true, // in countdown
              model.countdownMessage,
              () {}, // no action during countdown
              model.gameState ??
                  (GameUpdate()
                    ..gameWidth = 800
                    ..gameHeight = 600),
            ),
          ),
        ),
      ),
    );
  }
}

// Small presentational widget for the waiting overlay
class _WaitingForPlayersCard extends StatelessWidget {
  const _WaitingForPlayersCard();

  @override
  Widget build(BuildContext context) {
    return Container(
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
      child: const Column(
        mainAxisSize: MainAxisSize.min,
        children: [
          Icon(Icons.sports_esports, size: 50, color: Colors.blueAccent),
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
    );
  }
}
