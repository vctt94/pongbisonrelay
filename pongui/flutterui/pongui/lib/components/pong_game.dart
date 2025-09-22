import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:pongui/grpc/grpc_client.dart';
import 'package:golib_plugin/grpc/generated/pong.pb.dart';

class PongGame {
  final GrpcPongClient grpcClient; // gRPC client instance
  final String clientId;

  PongGame(this.clientId, this.grpcClient);

  Widget buildWidget(GameUpdate gameState, FocusNode focusNode, {VoidCallback? onReadyHotkey}) {
    return GestureDetector(
      onPanUpdate: handlePaddleMovement,
      onPanEnd: (details) {
        // Stop paddle movement when the user stops dragging
        stopPaddleMovement(clientId, 'ArrowUpStop');
        stopPaddleMovement(clientId, 'ArrowDownStop');
      },
      onTap: () => focusNode.requestFocus(),
      child: Focus(
        child: KeyboardListener(
          focusNode: focusNode..requestFocus(),
          onKeyEvent: (KeyEvent event) {
            if (event is KeyDownEvent || event is KeyRepeatEvent) {
              String keyLabel = event.logicalKey.keyLabel;
              // Ready-to-play hotkey: Space or 'R'
              if (onReadyHotkey != null) {
                if (event.logicalKey == LogicalKeyboardKey.space || keyLabel == 'r' || keyLabel == 'R') {
                  onReadyHotkey();
                  return; // don't also move paddles for space/r
                }
              }
              handleInput(clientId, keyLabel);
            } else if (event is KeyUpEvent) {
              // Handle key up events to stop paddle movement
              String keyLabel = event.logicalKey.keyLabel;
              if (keyLabel == 'W' || keyLabel == 'Arrow Up') {
                stopPaddleMovement(clientId, 'ArrowUpStop');
              } else if (keyLabel == 'S' || keyLabel == 'Arrow Down') {
                stopPaddleMovement(clientId, 'ArrowDownStop');
              }
            }
          },
          child: LayoutBuilder(
            builder: (context, constraints) {
              final gw = (gameState.gameWidth > 0) ? gameState.gameWidth : 800.0;
              final gh = (gameState.gameHeight > 0) ? gameState.gameHeight : 600.0;

              return FittedBox(
                fit: BoxFit.contain,
                child: SizedBox(
                  width: gw,
                  height: gh,
                  child: CustomPaint(
                    size: Size(gw, gh),
                    painter: PongPainter(gameState),
                  ),
                ),
              );
            },
          ),
        ),
      ),
    );
  }

  // Build an overlay widget for the ready-to-play UI and countdown
  Widget buildReadyToPlayOverlay(
      BuildContext context,
      bool isReadyToPlay,
      bool countdownStarted,
      String countdownMessage,
      Function onReadyPressed,
      GameUpdate gameState) {
    // If countdown has started, show the countdown message in the center
    if (countdownStarted) {
      return Center(
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
            children: [
              const Icon(
                Icons.sports_score,
                size: 50,
                color: Colors.blueAccent,
              ),
              const SizedBox(height: 20),
              Text(
                countdownMessage,
                style: const TextStyle(
                  color: Colors.white,
                  fontSize: 40,
                  fontWeight: FontWeight.bold,
                ),
              ),
            ],
          ),
        ),
      );
    }

    // If not ready to play, show the ready button with game controls info
    if (!isReadyToPlay) {
      return Container(
        color: Color.fromRGBO(0, 0, 0, 0.65),
        child: Center(
          child: Column(
            mainAxisSize: MainAxisSize.min,
            children: [
              // Static paddle and ball visualization
              SizedBox(
                height: 80,
                child: Row(
                  mainAxisAlignment: MainAxisAlignment.center,
                  children: [
                    Container(
                      width: 10,
                      height: 60,
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(5),
                      ),
                    ),
                    SizedBox(width: 100),
                    Container(
                      width: 20,
                      height: 20,
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(10),
                      ),
                    ),
                    SizedBox(width: 100),
                    Container(
                      width: 10,
                      height: 60,
                      decoration: BoxDecoration(
                        color: Colors.white,
                        borderRadius: BorderRadius.circular(5),
                      ),
                    ),
                  ],
                ),
              ),
              const SizedBox(height: 40),
              Text(
                "Ready to play?",
                style: const TextStyle(
                  color: Colors.blueAccent,
                  fontSize: 32,
                  fontWeight: FontWeight.bold,
                ),
              ),
              const SizedBox(height: 40),
              ElevatedButton(
                onPressed: () => onReadyPressed(),
                style: ElevatedButton.styleFrom(
                  backgroundColor: Colors.blueAccent,
                  padding:
                      const EdgeInsets.symmetric(horizontal: 50, vertical: 15),
                  shape: RoundedRectangleBorder(
                    borderRadius: BorderRadius.circular(30),
                  ),
                ),
                child: const Text(
                  "I'm Ready!",
                  style: TextStyle(
                    fontSize: 20,
                    fontWeight: FontWeight.bold,
                    color: Colors.white,
                  ),
                ),
              ),
              const SizedBox(height: 50),
              Container(
                padding: const EdgeInsets.all(20),
                decoration: BoxDecoration(
                  color: const Color(0xFF1B1E2C),
                  borderRadius: BorderRadius.circular(12),
                  border: Border.all(color: Colors.blueAccent.withAlpha(76)),
                ),
                child: Column(
                  children: [
                    const Text(
                      "GAME CONTROLS",
                      style: TextStyle(
                        color: Colors.blueAccent,
                        fontSize: 16,
                        fontWeight: FontWeight.bold,
                      ),
                    ),
                    const SizedBox(height: 15),
                    Row(
                      mainAxisSize: MainAxisSize.min,
                      children: [
                        _controlKey("W", "Up"),
                        const SizedBox(width: 10),
                        _controlKey("S", "Down"),
                        const SizedBox(width: 25),
                        _controlKey("↑", "Up"),
                        const SizedBox(width: 10),
                        _controlKey("↓", "Down"),
                      ],
                    ),
                  ],
                ),
              ),
            ],
          ),
        ),
      );
    }

    // If ready but waiting for opponent
    return Center(
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
          children: [
            const Icon(
              Icons.sports_esports,
              size: 50,
              color: Colors.blueAccent,
            ),
            const SizedBox(height: 20),
            const Text(
              "Waiting for players to get ready...",
              style: TextStyle(
                color: Colors.white,
                fontSize: 24,
                fontWeight: FontWeight.bold,
              ),
            ),
            const SizedBox(height: 20),
            SizedBox(
              width: 40,
              height: 40,
              child: CircularProgressIndicator(
                color: Colors.blueAccent,
                backgroundColor: Colors.grey.withAlpha(51),
                strokeWidth: 4,
              ),
            ),
          ],
        ),
      ),
    );
  }

  // Helper widget for control key display
  Widget _controlKey(String key, String action) {
    return Column(
      children: [
        Container(
          width: 40,
          height: 40,
          decoration: BoxDecoration(
            color: Colors.grey.shade800,
            borderRadius: BorderRadius.circular(6),
            border: Border.all(color: Colors.grey.shade600),
          ),
          child: Center(
            child: Text(
              key,
              style: const TextStyle(
                color: Colors.white,
                fontSize: 18,
                fontWeight: FontWeight.bold,
              ),
            ),
          ),
        ),
        const SizedBox(height: 5),
        Text(
          action,
          style: const TextStyle(
            color: Colors.white70,
            fontSize: 12,
          ),
        ),
      ],
    );
  }

  void handlePaddleMovement(DragUpdateDetails details) {
    double deltaY = details.delta.dy;
    String data = deltaY < 0 ? 'ArrowUp' : 'ArrowDown';
    grpcClient.sendInput(clientId, data);
  }

  Future<void> handleInput(String clientId, String data) async {
    await _sendKeyInput(data);
  }

  Future<void> _sendKeyInput(String data) async {
    try {
      String action;

      if (data == 'W' || data == 'Arrow Up') {
        action = 'ArrowUp';
      } else if (data == 'S' || data == 'Arrow Down') {
        action = 'ArrowDown';
      } else {
        return;
      }
      await grpcClient.sendInput(clientId, action);
    } catch (e) {
      print(e);
    }
  }

  // New method to stop paddle movement
  Future<void> stopPaddleMovement(String clientId, String action) async {
    try {
      await grpcClient.sendInput(clientId, action);
    } catch (e) {
      print(e);
    }
  }

  @override
  String get name => 'Pong';
}

class PongPainter extends CustomPainter {
  final GameUpdate gameState;

  PongPainter(this.gameState);

  @override
  void paint(Canvas canvas, Size size) {
    // Extract game dimensions
    double gameWidth = gameState.gameWidth;
    double gameHeight = gameState.gameHeight;

    // Calculate scaling factors
    double scaleX = size.width / gameWidth;
    double scaleY = size.height / gameHeight;

    // Paint object for drawing
    var paint = Paint()
      ..color = Colors.white
      ..style = PaintingStyle.fill
      ..isAntiAlias = true;

    // Draw background
    canvas.drawRect(
      Rect.fromLTWH(0.0, 0.0, size.width, size.height),
      Paint()..color = Colors.black,
    );

    // Extract and scale paddle 1 properties (use real position from server)
    double paddle1X = gameState.p1X;
    double paddle1Y = gameState.p1Y;
    double paddle1Width = gameState.p1Width;
    double paddle1Height = gameState.p1Height;

    // Scale paddle 1 properties
    paddle1X *= scaleX;
    paddle1Y *= scaleY;
    paddle1Width *= scaleX;
    paddle1Height *= scaleY;

    // Extract and scale paddle 2 properties
    double paddle2X = gameState.p2X;
    double paddle2Y = gameState.p2Y;
    double paddle2Width = gameState.p2Width;
    double paddle2Height = gameState.p2Height;

    // Scale paddle 2 properties
    paddle2X *= scaleX;
    paddle2Y *= scaleY;
    paddle2Width *= scaleX;
    paddle2Height *= scaleY;

    // Extract and scale ball properties
    double ballX = gameState.ballX;
    double ballY = gameState.ballY;
    double ballWidth = gameState.ballWidth;
    double ballHeight = gameState.ballHeight;

    // Scale ball properties
    ballX *= scaleX;
    ballY *= scaleY;
    ballWidth *= scaleX;
    ballHeight *= scaleY;

    // Draw Paddle 1
    canvas.drawRect(
      Rect.fromLTWH(paddle1X, paddle1Y, paddle1Width, paddle1Height),
      paint,
    );

    // Draw Paddle 2
    canvas.drawRect(
      Rect.fromLTWH(paddle2X, paddle2Y, paddle2Width, paddle2Height),
      paint,
    );

    // Draw the ball as a circle for smoother motion
    double radius = (ballWidth + ballHeight) / 4;
    canvas.drawCircle(
      Offset(ballX + ballWidth / 2, ballY + ballHeight / 2),
      radius,
      paint,
    );

    // Draw scores
    int p1Score = gameState.p1Score;
    int p2Score = gameState.p2Score;

    // Create text painters for scores
    final p1ScoreTextPainter = TextPainter(
      text: TextSpan(
        text: '$p1Score',
        style: TextStyle(
            color: Colors.white, fontSize: 24, fontWeight: FontWeight.bold),
      ),
      textDirection: TextDirection.ltr,
    );

    final p2ScoreTextPainter = TextPainter(
      text: TextSpan(
        text: '$p2Score',
        style: TextStyle(
            color: Colors.white, fontSize: 24, fontWeight: FontWeight.bold),
      ),
      textDirection: TextDirection.ltr,
    );

    // Layout the text
    p1ScoreTextPainter.layout();
    p2ScoreTextPainter.layout();

    // Position and draw the scores at the top of the screen
    p1ScoreTextPainter.paint(
        canvas, Offset(size.width * 0.25 - p1ScoreTextPainter.width / 2, 20));
    p2ScoreTextPainter.paint(
        canvas, Offset(size.width * 0.75 - p2ScoreTextPainter.width / 2, 20));
  }

  @override
  bool shouldRepaint(PongPainter oldDelegate) {
    // Repaint whenever the game state changes
    return true;
  }
}
