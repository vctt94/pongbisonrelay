import 'dart:async';
import 'package:flutter/material.dart';
import 'package:golib_plugin/grpc/generated/pong.pb.dart';
import 'package:flutter/scheduler.dart';
import 'package:golib_plugin/golib_plugin.dart';

// A compact snapshot with timestamp for interpolation.
class _Snapshot {
  final GameUpdate s;
  final int tMicros; // monotonic-ish timestamp (DateTime.now().microsecondsSinceEpoch)
  _Snapshot(this.s, this.tMicros);
}

// Interpolates positions between server snapshots, rendering slightly in the past.
class SnapshotInterpolator {
  _Snapshot? _curr;

  // Keep only latest snapshot; no interpolation.
  void push(GameUpdate s) {
    final now = DateTime.now().microsecondsSinceEpoch;
    // Clone minimal fields to avoid external mutation issues.
    final snap = GameUpdate()
      ..gameWidth = s.gameWidth
      ..gameHeight = s.gameHeight
      ..p1X = s.p1X ..p1Y = s.p1Y ..p1Width = s.p1Width ..p1Height = s.p1Height
      ..p2X = s.p2X ..p2Y = s.p2Y ..p2Width = s.p2Width ..p2Height = s.p2Height
      ..ballX = s.ballX ..ballY = s.ballY ..ballWidth = s.ballWidth ..ballHeight = s.ballHeight
      ..p1Score = s.p1Score ..p2Score = s.p2Score;
    _curr = _Snapshot(snap, now);
  }

  GameUpdate sample() {
    final current = _curr;
    if (current == null) {
      return GameUpdate()
        ..gameWidth = 800
        ..gameHeight = 600;
    }
    return current.s;
  }
}

// Lightweight render ticker to repaint at ~60fps without a TickerProvider.
class RenderLoop extends ChangeNotifier {
  bool _running = false;

  void start() {
    if (_running) return;
    _running = true;
    SchedulerBinding.instance.scheduleFrameCallback(_tick);
  }

  void _tick(Duration ts) {
    if (!_running) return;
    notifyListeners();
    // Schedule next frame aligned with vsync
    SchedulerBinding.instance.scheduleFrameCallback(_tick);
  }

  void stop() {
    _running = false;
  }
}

class InputThrottler {
  String? _active; // 'ArrowUp' | 'ArrowDown' | null

  Future<void> update(double dy) async {
    final want = dy < 0 ? 'ArrowUp' : 'ArrowDown';
    if (want != _active) {
      // stop previous if any
      if (_active != null) {
        await Golib.sendInput(_active == 'ArrowUp' ? 'ArrowUpStop' : 'ArrowDownStop');
      }
      _active = want;
      await Golib.sendInput(want);
    }
  }

  Future<void> stop() async {
    if (_active != null) {
      await Golib.sendInput(_active == 'ArrowUp' ? 'ArrowUpStop' : 'ArrowDownStop');
      _active = null;
    }
  }
}
