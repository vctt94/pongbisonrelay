import 'dart:async';

import 'package:flutter/material.dart';
import 'package:flutter/services.dart';
import 'package:pongui/grpc/grpc_client.dart';
import 'package:golib_plugin/grpc/generated/pong.pb.dart';

class BackgroundLogo extends StatelessWidget {
  const BackgroundLogo({super.key, this.frac = 0.70}); // 70% of min(w,h)
  final double frac;

  @override
  Widget build(BuildContext context) {
    return IgnorePointer(
      child: LayoutBuilder(
        builder: (context, constraints) {
          final shortest = constraints.biggest.shortestSide;
          final size = (shortest.isFinite && shortest > 0)
              ? shortest * frac
              : 300.0;

          return Center(
            child: Image.asset(
              'assets/images/dcrlogo.png',
              width: size,
              height: size,
              fit: BoxFit.contain,
              opacity: const AlwaysStoppedAnimation(0.50),
              errorBuilder: (_, err, __) {
                debugPrint('BackgroundLogo: failed to load -> $err');
                return const SizedBox.shrink();
              },
            ),
          );
        },
      ),
    );
  }
}

class PongGame {
  final GrpcPongClient grpcClient; // gRPC client instance
  final String clientId;

  PongGame(this.clientId, this.grpcClient);

  Widget buildWidget(GameUpdate gameState, FocusNode focusNode, {VoidCallback? onReadyHotkey}) {
    return GestureDetector(
      onPanUpdate: handlePaddleMovement,
      onPanEnd: (details) {
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
              if (onReadyHotkey != null) {
                if (event.logicalKey == LogicalKeyboardKey.space || keyLabel == 'r' || keyLabel == 'R') {
                  onReadyHotkey();
                  return;
                }
              }
              handleInput(clientId, keyLabel);
            } else if (event is KeyUpEvent) {
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

              return Center(
                child: SizedBox(
                  width: constraints.maxWidth,
                  child: AspectRatio(
                    aspectRatio: gw / gh,
                    child: RepaintBoundary(
                      child: Stack(
                        fit: StackFit.expand,
                        children: [
                          // Black backdrop (static)
                          const ColoredBox(color: Colors.black),

                          // Game canvas (repaints)
                          CustomPaint(
                            painter: PongPainter(gameState),
                            // Hints to the engine:
                            isComplex: true,
                            willChange: true,
                          ),

                          // Static logo layer (no repaints) - moved after game canvas
                          const BackgroundLogo(),

                          // Score overlay (no pointer intercept)
                          IgnorePointer(
                            child: Stack(
                              fit: StackFit.expand,
                              children: [
                                Align(
                                  alignment: const FractionalOffset(0.25, 0.10),
                                  child: Text(
                                    '${gameState.p1Score}',
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 24,
                                      fontWeight: FontWeight.bold,
                                      height: 1.0,
                                    ),
                                  ),
                                ),
                                Align(
                                  alignment: const FractionalOffset(0.75, 0.10),
                                  child: Text(
                                    '${gameState.p2Score}',
                                    style: const TextStyle(
                                      color: Colors.white,
                                      fontSize: 24,
                                      fontWeight: FontWeight.bold,
                                      height: 1.0,
                                    ),
                                  ),
                                ),
                              ],
                            ),
                          ),
                        ],
                      ),
                    ),
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
                        color: Colors.blue,
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
                        color: Colors.green,
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

  // Removed stray override; this isn't implementing a named interface
  String get name => 'Pong';
}

class PongPainter extends CustomPainter {
  final GameUpdate gameState;
  PongPainter(this.gameState);

  @override
  void paint(Canvas canvas, Size size) {
    final gw = (gameState.gameWidth > 0) ? gameState.gameWidth : 800.0;
    final gh = (gameState.gameHeight > 0) ? gameState.gameHeight : 600.0;
    final sx = size.width / gw;
    final sy = size.height / gh;

    // Paints
    final ballPaint = Paint()..color = Colors.white..isAntiAlias = true;
    final paddle1Paint = Paint()..color = Colors.blue..isAntiAlias = true;
    final paddle2Paint = Paint()..color = Colors.green..isAntiAlias = true;

    // P1
    final p1x = gameState.p1X * sx;
    final p1y = gameState.p1Y * sy;
    final p1w = gameState.p1Width * sx;
    final p1h = gameState.p1Height * sy;
    canvas.drawRect(Rect.fromLTWH(p1x, p1y, p1w, p1h), paddle1Paint);

    // P2
    final p2x = gameState.p2X * sx;
    final p2y = gameState.p2Y * sy;
    final p2w = gameState.p2Width * sx;
    final p2h = gameState.p2Height * sy;
    canvas.drawRect(Rect.fromLTWH(p2x, p2y, p2w, p2h), paddle2Paint);

    // Ball
    final bx = gameState.ballX * sx;
    final by = gameState.ballY * sy;
    final bw = gameState.ballWidth * sx;
    final bh = gameState.ballHeight * sy;
    final r = (bw + bh) / 4;
    canvas.drawCircle(Offset(bx + bw / 2, by + bh / 2), r, ballPaint);
  }

  // Only repaint when the instance (or relevant fields) change.
  @override
  bool shouldRepaint(covariant PongPainter old) => old.gameState != gameState;
}
