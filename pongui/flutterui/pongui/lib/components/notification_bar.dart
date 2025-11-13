import 'package:flutter/material.dart';
import 'package:pongui/models/notifications.dart';
import 'package:provider/provider.dart';

class NotificationBar extends StatelessWidget {
  @override
  Widget build(BuildContext context) {
    return Consumer<NotificationModel>(
      builder: (context, notificationModel, child) {
        if (notificationModel.notification.isEmpty) {
          return SizedBox.shrink(); // Hide when no notification
        }
        return Material(
          color: Colors.transparent,
          child: Container(
            width: double.infinity,
            color: Colors.blueAccent,
            padding: const EdgeInsets.symmetric(horizontal: 12.0, vertical: 8.0),
            child: Row(
              crossAxisAlignment: CrossAxisAlignment.center,
              children: [
                Expanded(
                  child: Text(
                    notificationModel.notification,
                    style: const TextStyle(color: Colors.white),
                    textAlign: TextAlign.center,
                  ),
                ),
                // Use tooltip only when an Overlay is available to avoid
                // debugCheckHasOverlay assertion when this bar is outside
                // the Navigator's overlay.
                Builder(builder: (ctx) {
                  final hasOverlay = Overlay.maybeOf(ctx) != null;
                  final closeBtn = IconButton(
                    icon: const Icon(Icons.close, color: Colors.white),
                    padding: EdgeInsets.zero,
                    constraints: const BoxConstraints(),
                    onPressed: () => notificationModel.hideNotification(),
                  );
                  return hasOverlay
                      ? Tooltip(message: 'Dismiss', child: closeBtn)
                      : closeBtn;
                }),
              ],
            ),
          ),
        );
      },
    );
  }
}
