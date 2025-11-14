import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';

class WaitingRoomList extends StatelessWidget {
  final List<LocalWaitingRoom> waitingRooms;
  final Function(String roomId) onJoinRoom;
  final String? currentRoomId;
  final bool canJoinRooms;
  final String? joinDisabledTooltip;

  const WaitingRoomList(
    this.waitingRooms, {
    this.currentRoomId,
    required this.onJoinRoom,
    this.canJoinRooms = true,
    this.joinDisabledTooltip,
    Key? key,
  }) : super(key: key);

  @override
  Widget build(BuildContext context) {
    // This is now handled in MainContent
    if (waitingRooms.isEmpty) {
      return const SizedBox.shrink();
    }

    return Center(
      child: SizedBox(
        width: MediaQuery.of(context).size.width * 0.85,
        child: ListView.builder(
          itemCount: waitingRooms.length,
          padding: const EdgeInsets.all(12),
          shrinkWrap: true,
          physics: const NeverScrollableScrollPhysics(),
          itemBuilder: (context, index) {
            final wr = waitingRooms[index];
            final bool isCurrentRoom = currentRoomId == wr.id;
            // Mark rooms that contain any disconnected players; this allows
            // users to see at a glance which rooms have opponents that are
            // currently offline.
            final bool hasDisconnectedPlayer =
                wr.players.any((p) => p.connected == false);
            final int playerCount = wr.players.length;
            const int maxPlayers = 2;
            final bool isRoomFull = playerCount >= maxPlayers;
            final bool canJoinThisRoom = canJoinRooms && !isRoomFull;

            return Card(
              elevation: 4,
              color: const Color(0xFF1B1E2C), // Dark card background
              shape: RoundedRectangleBorder(
                borderRadius: BorderRadius.circular(12),
                side: isCurrentRoom
                    ? const BorderSide(color: Colors.greenAccent, width: 2)
                    : BorderSide.none,
              ),
              margin: const EdgeInsets.symmetric(vertical: 8),
              child: ListTile(
                contentPadding:
                    const EdgeInsets.symmetric(horizontal: 16, vertical: 12),
                title: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    Text(
                      'Room ID: ${wr.id}',
                      style: const TextStyle(
                          color: Colors.white,
                          fontSize: 14,
                          fontWeight: FontWeight.bold),
                    ),
                    const SizedBox(height: 8),
                    Text(
                      'Players: $playerCount / $maxPlayers',
                      style: const TextStyle(color: Colors.white70),
                    ),
                    const SizedBox(height: 4),
                    if (hasDisconnectedPlayer)
                      const Text(
                        'Status: Opponent disconnected or left',
                        style: TextStyle(
                          color: Colors.redAccent,
                          fontStyle: FontStyle.italic,
                        ),
                      )
                    else
                      const SizedBox.shrink(),
                    const SizedBox(height: 4),
                    Text(
                      'Bet: ${wr.betAmt / 1e8} DCR',
                      style: const TextStyle(color: Colors.white70),
                    ),
                  ],
                ),
                trailing: currentRoomId != wr.id
                    ? currentRoomId == null
                        ? Builder(builder: (context) {
                            String? tooltipMessage;
                            if (!canJoinThisRoom) {
                              if (!canJoinRooms &&
                                  (joinDisabledTooltip ?? '').isNotEmpty) {
                                tooltipMessage = joinDisabledTooltip;
                              } else if (isRoomFull) {
                                tooltipMessage = 'Room is full';
                              }
                            }

                            final btn = ElevatedButton(
                              onPressed: canJoinThisRoom
                                  ? () => onJoinRoom(wr.id)
                                  : null,
                              style: ElevatedButton.styleFrom(
                                backgroundColor: canJoinThisRoom
                                    ? Colors.blueAccent
                                    : Colors.blueGrey,
                                foregroundColor: Colors.white,
                                elevation: 2,
                                shape: RoundedRectangleBorder(
                                  borderRadius: BorderRadius.circular(8),
                                ),
                                padding: const EdgeInsets.symmetric(
                                    horizontal: 16, vertical: 10),
                              ),
                              child: const Text(
                                "Join",
                                style: TextStyle(fontWeight: FontWeight.bold),
                              ),
                            );
                            if (tooltipMessage == null ||
                                tooltipMessage.isEmpty) {
                              return btn;
                            }
                            return Tooltip(
                              message: tooltipMessage,
                              child: AbsorbPointer(child: btn),
                            );
                          })
                        : null
                    : Container(
                        padding: const EdgeInsets.symmetric(
                            horizontal: 12, vertical: 6),
                        decoration: BoxDecoration(
                          color: Colors.green.withOpacity(0.3),
                          borderRadius: BorderRadius.circular(8),
                          border:
                              Border.all(color: Colors.greenAccent, width: 1),
                        ),
                        child: const Text(
                          "Joined",
                          style: TextStyle(
                            color: Colors.greenAccent,
                            fontWeight: FontWeight.bold,
                          ),
                        ),
                      ),
              ),
            );
          },
        ),
      ),
    );
  }
}
