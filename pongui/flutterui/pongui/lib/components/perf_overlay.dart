import 'dart:async';
import 'package:flutter/material.dart';
import 'package:golib_plugin/definitions.dart';
import 'package:golib_plugin/golib_plugin.dart';

class PerfOverlay extends StatefulWidget {
  const PerfOverlay({super.key});

  @override
  State<PerfOverlay> createState() => _PerfOverlayState();
}

class _PerfOverlayState extends State<PerfOverlay> {
  StreamSubscription<PerfStats>? _sub;
  PerfStats? _last;

  @override
  void initState() {
    super.initState();
    _sub = Golib.perfStats.listen((s) {
      setState(() => _last = s);
    }, onError: (_) {});
  }

  @override
  void dispose() {
    _sub?.cancel();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    final s = _last;
    if (s == null) return const SizedBox.shrink();
    final warn = s.sinceLastEmitMs > 500;
    
    return Container(
      margin: const EdgeInsets.all(8),
      padding: const EdgeInsets.symmetric(horizontal: 8, vertical: 6),
      decoration: BoxDecoration(
        color: Colors.black.withOpacity(0.5),
        borderRadius: BorderRadius.circular(6),
        border: Border.all(
          color: warn ? Colors.redAccent.withOpacity(0.6) : Colors.blueGrey.withOpacity(0.4),
          width: 1,
        ),
      ),
      child: DefaultTextStyle(
        style: const TextStyle(fontSize: 11, color: Colors.white70),
        child: Column(
          crossAxisAlignment: CrossAxisAlignment.start,
          mainAxisSize: MainAxisSize.min,
          children: [
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text('fwd ${s.ffiFwd}'),
                const SizedBox(width: 8),
                Text('in ${s.framesIn}'),
                const SizedBox(width: 8),
                Text('out ${s.framesOut}'),
                const SizedBox(width: 8),
                Text('dec ${s.decodeLastMs}ms'),
              ],
            ),
            const SizedBox(height: 3),
            Row(
              mainAxisSize: MainAxisSize.min,
              children: [
                Text('q ${s.queueLen}/${s.queueMax}'),
                const SizedBox(width: 8),
                Text('Δ ${s.inDtAvg}/${s.outDtAvg}ms'),
                const SizedBox(width: 8),
                Text(
                  'idle ${s.sinceLastEmitMs}ms',
                  style: TextStyle(
                    color: warn ? Colors.redAccent : Colors.white70,
                  ),
                ),
              ],
            ),
          ],
        ),
      ),
    );
  }
}
