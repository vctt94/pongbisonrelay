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
            final int playerCount = wr.players.length;
            final int readyCount = wr.players.where((p) => p.ready).length;
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
                leading: CircleAvatar(
                  backgroundColor: Colors.blueAccent.withOpacity(0.3),
                  child: const Icon(
                    Icons.person,
                    color: Colors.white,
                  ),
                ),
                title: Text(
                  wr.host,
                  style: const TextStyle(
                    color: Colors.white,
                    fontWeight: FontWeight.bold,
                    fontSize: 16,
                  ),
                ),
                subtitle: Column(
                  crossAxisAlignment: CrossAxisAlignment.start,
                  children: [
                    const SizedBox(height: 6),
                    Row(
                      children: [
                        const Icon(Icons.attach_money,
                            size: 16, color: Colors.amber),
                        const SizedBox(width: 4),
                        Text(
                          'Bet: ${wr.betAmt / 1e8} DCR',
                          style: const TextStyle(color: Colors.white70),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        const Icon(Icons.people,
                            size: 16, color: Colors.lightBlueAccent),
                        const SizedBox(width: 4),
                        Text(
                          'Players: $playerCount / $maxPlayers',
                          style: const TextStyle(color: Colors.white70),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        const Icon(Icons.check_circle,
                            size: 16, color: Colors.greenAccent),
                        const SizedBox(width: 4),
                        Text(
                          'Ready: $readyCount / $maxPlayers',
                          style: TextStyle(
                            color: readyCount == maxPlayers
                                ? Colors.greenAccent
                                : Colors.white70,
                          ),
                        ),
                      ],
                    ),
                    const SizedBox(height: 4),
                    Row(
                      children: [
                        const Icon(Icons.tag, size: 16, color: Colors.white54),
                        const SizedBox(width: 4),
                        Text(
                          'Room ID: ${wr.id}',
                          style: const TextStyle(
                              color: Colors.white60, fontSize: 12),
                        ),
                      ],
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
