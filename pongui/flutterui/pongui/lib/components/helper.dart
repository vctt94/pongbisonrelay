import 'dart:async';
import 'package:flutter/material.dart';
import 'package:golib_plugin/grpc/generated/pong.pb.dart';
import 'package:flutter/scheduler.dart';
import 'package:golib_plugin/golib_plugin.dart';

// A compact snapshot with timestamp for interpolation.
class _Snapshot {
  final GameUpdate s;
  final int tMicros; // timestamp (DateTime.now().microsecondsSinceEpoch)
  _Snapshot(this.s, this.tMicros);
}

// Interpolates positions between snapshots, rendering slightly in the past.
class SnapshotInterpolator {
  // Target render delay to absorb arrival jitter (~16.6ms per frame @60fps).
  // 60–75ms hides occasional 80–130ms spikes while keeping latency low.
  int renderDelayMicros;
  final List<_Snapshot> _buf = <_Snapshot>[]; // keep last 2–3 snapshots

  SnapshotInterpolator({this.renderDelayMicros = 60000});

  void push(GameUpdate s) {
    final now = DateTime.now().microsecondsSinceEpoch;
    // Clone fields to avoid external mutation.
    final snap = GameUpdate()
      ..gameWidth = s.gameWidth
      ..gameHeight = s.gameHeight
      ..p1X = s.p1X ..p1Y = s.p1Y ..p1Width = s.p1Width ..p1Height = s.p1Height
      ..p2X = s.p2X ..p2Y = s.p2Y ..p2Width = s.p2Width ..p2Height = s.p2Height
      ..ballX = s.ballX ..ballY = s.ballY ..ballWidth = s.ballWidth ..ballHeight = s.ballHeight
      ..p1Score = s.p1Score ..p2Score = s.p2Score;
    _buf.add(_Snapshot(snap, now));
    if (_buf.length > 3) {
      _buf.removeAt(0);
    }
  }

  GameUpdate sample() {
    if (_buf.isEmpty) {
      return GameUpdate()
        ..gameWidth = 800
        ..gameHeight = 600;
    }

    final target = DateTime.now().microsecondsSinceEpoch - renderDelayMicros;

    // Single snapshot: nothing to interpolate.
    if (_buf.length == 1) return _buf.last.s;

    // Find bracketing snapshots around the target time.
    _Snapshot a = _buf.first;
    _Snapshot b = _buf.last;
    for (int i = 0; i < _buf.length - 1; i++) {
      final s0 = _buf[i];
      final s1 = _buf[i + 1];
      if (s0.tMicros <= target && target <= s1.tMicros) {
        a = s0; b = s1; break;
      }
    }
    // If target is before earliest, use earliest; if after latest, use latest.
    if (target <= _buf.first.tMicros) return _buf.first.s;
    if (target >= _buf.last.tMicros) return _buf.last.s;

    final dt = (b.tMicros - a.tMicros).toDouble();
    final t = dt > 0 ? ((target - a.tMicros).toDouble() / dt) : 1.0;
    double lerp(double x, double y) => x + (y - x) * t;

    final out = GameUpdate()
      ..gameWidth = b.s.gameWidth
      ..gameHeight = b.s.gameHeight
      ..p1X = lerp(a.s.p1X, b.s.p1X)
      ..p1Y = lerp(a.s.p1Y, b.s.p1Y)
      ..p1Width = b.s.p1Width
      ..p1Height = b.s.p1Height
      ..p2X = lerp(a.s.p2X, b.s.p2X)
      ..p2Y = lerp(a.s.p2Y, b.s.p2Y)
      ..p2Width = b.s.p2Width
      ..p2Height = b.s.p2Height
      ..ballX = lerp(a.s.ballX, b.s.ballX)
      ..ballY = lerp(a.s.ballY, b.s.ballY)
      ..ballWidth = b.s.ballWidth
      ..ballHeight = b.s.ballHeight
      ..p1Score = b.s.p1Score
      ..p2Score = b.s.p2Score;
    return out;
  }
}

// Lightweight render ticker to repaint at ~60fps without a TickerProvider.
class RenderLoop extends ChangeNotifier {
  bool _running = false;
  int _ticks = 0;
  int _lastLogMs = DateTime.now().millisecondsSinceEpoch;

  void start() {
    if (_running) return;
    _running = true;
    SchedulerBinding.instance.scheduleFrameCallback(_tick);
  }

  void _tick(Duration ts) {
    if (!_running) return;
    notifyListeners();
    // simple FPS counter
    _ticks += 1;
    final nowMs = DateTime.now().millisecondsSinceEpoch;
    if (nowMs - _lastLogMs >= 1000) {
      debugPrint('[render] fps=$_ticks');
      _ticks = 0;
      _lastLogMs = nowMs;
    }
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
