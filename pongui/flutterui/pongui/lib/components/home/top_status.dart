import 'package:flutter/material.dart';
import 'package:pongui/models/pong.dart';

class TopStatusCard extends StatelessWidget {
  final PongModel pongModel;
  final VoidCallback? onErrorDismissed;

  const TopStatusCard({
    Key? key,
    required this.pongModel,
    this.onErrorDismissed,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    return Column(
      crossAxisAlignment: CrossAxisAlignment.start,
      children: [
        // Compact Status Section
        Padding(
          padding: const EdgeInsets.all(16.0),
          child: Column(
            children: [
              Card(
                elevation: 2,
                shape: RoundedRectangleBorder(
                  borderRadius: BorderRadius.circular(12),
                ),
                child: Padding(
                  padding: const EdgeInsets.all(12.0),
                  child: Column(
                    crossAxisAlignment: CrossAxisAlignment.start,
                    children: [
                      Row(
                        mainAxisAlignment: MainAxisAlignment.spaceBetween,
                        children: [
                          Text(
                            "Bet: ${pongModel.betAmt / 1e8}",
                            style: Theme.of(context).textTheme.bodyMedium,
                          ),
                          Text(
                            pongModel.isReady
                                ? (pongModel.isGameStarted
                                    ? "In Game"
                                    : "Ready")
                                : "Not Ready",
                            style: Theme.of(context).textTheme.bodyMedium,
                          ),
                        ],
                      ),
                      // If game hasn't started, show waiting room info
                      if (!pongModel.isGameStarted) ...[
                        const SizedBox(height: 12),
                        Divider(color: Colors.grey.shade400),
                        const SizedBox(height: 12),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              "Current Waiting Room:",
                              style: Theme.of(context).textTheme.titleMedium,
                            ),
                            if (pongModel.betAmt > 0) ...[
                              if (pongModel.currentWR == null)
                                FilledButton(
                                  onPressed: pongModel.createWaitingRoom,
                                  child: const Text("Create Waiting Room"),
                                ),
                              if (pongModel.currentWR != null)
                                Row(
                                  children: [
                                    FilledButton(
                                      onPressed: pongModel.toggleReady,
                                      child: Text(
                                        pongModel.isReady
                                            ? "Cancel Ready"
                                            : "Ready",
                                      ),
                                    ),
                                    const SizedBox(width: 8),
                                    FilledButton(
                                      style: FilledButton.styleFrom(
                                        backgroundColor: Colors.redAccent,
                                      ),
                                      onPressed: () =>
                                          pongModel.leaveWaitingRoom(),
                                      child: const Text("Leave Room"),
                                    ),
                                  ],
                                ),
                            ],
                          ],
                        ),
                        const SizedBox(height: 8),
                        Row(
                          mainAxisAlignment: MainAxisAlignment.spaceBetween,
                          children: [
                            Text(
                              "Room ID: ${pongModel.currentWR?.id ?? ""}",
                              style: Theme.of(context).textTheme.bodyMedium,
                            ),
                          ],
                        ),
                        const SizedBox(height: 8),
                        Text(
                          "Players: ${pongModel.currentWR?.players.length ?? 0} / 2",
                          style: Theme.of(context).textTheme.bodyMedium,
                        ),
                      ],
                    ],
                  ),
                ),
              ),
            ],
          ),
        ),

        // Error Message
        if (pongModel.errorMessage.isNotEmpty)
          Padding(
            padding: const EdgeInsets.symmetric(horizontal: 16),
            child: Card(
              color: Colors.red.shade100,
              child: Padding(
                padding: const EdgeInsets.all(8.0),
                child: Row(
                  children: [
                    const Icon(Icons.error, color: Colors.red),
                    const SizedBox(width: 8),
                    Expanded(
                      child: Text(
                        pongModel.errorMessage,
                        style: Theme.of(context)
                            .textTheme
                            .bodyMedium
                            ?.copyWith(color: Colors.red),
                      ),
                    ),
                    Material(
                      color: Colors.transparent,
                      child: InkWell(
                        onTap: () {
                          onErrorDismissed?.call();
                        },
                        borderRadius: BorderRadius.circular(20),
                        child: const Padding(
                          padding: EdgeInsets.all(8.0),
                          child: Icon(Icons.close, color: Colors.red, size: 20),
                        ),
                      ),
                    ),
                  ],
                ),
              ),
            ),
          ),
      ],
    );
  }
}
