import 'dart:async';
import 'dart:convert';

import 'package:ffi/ffi.dart';
import 'package:flutter/cupertino.dart';
import 'package:golib_plugin/definitions.dart';
import 'dart:ffi';
import 'dart:isolate';
import 'desktop_dynlib.dart';

class _ReadStrData {
  SendPort sp;
  _ReadStrData(this.sp);
}

void _readStrIsolate(_ReadStrData data) async {
  final DynamicLibrary lib = DynamicLibrary.open(desktopLibPath());
  late final ReadStrNative readStr =
      lib.lookupFunction<ReadStrNative, ReadStrNative>('ReadStr');

  for (;;) {
    var s = readStr().toDartString();
    data.sp.send(s);
  }
}

void _readAsyncResultsIsolate(SendPort sp) async {
  final DynamicLibrary lib = DynamicLibrary.open(desktopLibPath());
  final NextCallResultNative nextCallResult =
      lib.lookupFunction<NextCallResultNative, NextCallResultNative>(
          'NextCallResult');
  final CopyCallResultFunc copyCallResult =
      lib.lookupFunction<CopyCallResultNative, CopyCallResultFunc>(
          'CopyCallResult');

  var buffSize = 1024 * 1024;
  var buff = calloc.allocate<Uint8>(buffSize);

  // Perf counters (logs once per second) — no behavioral change.
  var lastLog = DateTime.now();
  int framesForwarded = 0; // NTGameFrame forwarded to main isolate
  // Inter-arrival metrics for NTGameFrame (helps detect isolate stalls).
  int lastFrameUs = 0;
  int maxGapMs = 0;
  int sumGapMs = 0;
  int gapCount = 0;
  int gapsOver100 = 0;
  int gapsOver500 = 0;
  // Per-iteration timing metrics to localize stalls.
  int nativeMaxMs = 0, nativeSumMs = 0, nativeCount = 0;
  int copyMaxUs = 0, copySumUs = 0, copyCount = 0;
  int sendMaxUs = 0, sendSumUs = 0, sendCount = 0;
  int payloadMax = 0;
  // Whole-loop heartbeat to detect isolate stalls irrespective of producer output.
  int loopLastUs = 0;
  int loopMaxGapMs = 0;
  int loopGapsOver100 = 0;
  int loopGapsOver500 = 0;

  await Future.delayed(const Duration(seconds: 1));
  for (;;) {
    // Loop heartbeat at the very start of the iteration.
    final tLoopStartUs = DateTime.now().microsecondsSinceEpoch;
    if (loopLastUs != 0) {
      final loopGapMs = ((tLoopStartUs - loopLastUs) / 1000).round();
      if (loopGapMs > loopMaxGapMs) loopMaxGapMs = loopGapMs;
      if (loopGapMs >= 100) loopGapsOver100++;
      if (loopGapMs >= 500) loopGapsOver500++;
    }
    loopLastUs = tLoopStartUs;

    final tStartUs = DateTime.now().microsecondsSinceEpoch;
    var nr = nextCallResult();
    // Skip forwarding idle heartbeats to reduce message pressure on the main isolate.
    if (nr.cmdType == NTNOP) {
      continue;
    }
    final tAfterNativeUs = DateTime.now().microsecondsSinceEpoch;
    final nativeMs = ((tAfterNativeUs - tStartUs) / 1000).round();
    nativeSumMs += nativeMs;
    nativeCount++;
    if (nativeMs > nativeMaxMs) nativeMaxMs = nativeMs;

    // Resize response reading buffer if needed.
    if (nr.payloadLen > buffSize) {
      calloc.free(buff);
      buffSize = nr.payloadLen;
      buff = calloc.allocate<Uint8>(buffSize);
    }

    // Copy the payload.
    final tCopyStartUs = DateTime.now().microsecondsSinceEpoch;
    var rid = copyCallResult(nr.handle, buff.cast<Utf8>());
    final tCopyEndUs = DateTime.now().microsecondsSinceEpoch;
    final copyUs = (tCopyEndUs - tCopyStartUs);
    copySumUs += copyUs;
    copyCount++;
    if (copyUs > copyMaxUs) copyMaxUs = copyUs;
    // Decode payload according to cmdType
    dynamic payload;
    final view = buff.asTypedList(nr.payloadLen);
    if (nr.cmdType == NTGameFrame) {
      // Copy into a new list since buffer is reused
      // Send as TransferableTypedData to avoid cross-isolate payload copy
      // and avoid an extra intermediate allocation.
      payload = TransferableTypedData.fromList([view]);
      framesForwarded++;
      if (nr.payloadLen > payloadMax) payloadMax = nr.payloadLen;
      // Compute inter-arrival gap
      final nowUs = DateTime.now().microsecondsSinceEpoch;
      if (lastFrameUs != 0) {
        final gapMs = ((nowUs - lastFrameUs) / 1000).round();
        if (gapMs > maxGapMs) maxGapMs = gapMs;
        sumGapMs += gapMs;
        gapCount++;
        if (gapMs >= 100) gapsOver100++;
        if (gapMs >= 500) {
          gapsOver500++;
          // Immediate warning for large gaps can be noisy; rely on per-second summary.
        }
      }
      lastFrameUs = nowUs;
    } else {
      payload = utf8.decode(view);
    }

    // Send the response.
    final tSendStartUs = DateTime.now().microsecondsSinceEpoch;
    var res = [rid, nr.isErr == 1, nr.cmdType, payload];
    sp.send(res);
    final tSendEndUs = DateTime.now().microsecondsSinceEpoch;
    final sendUs = (tSendEndUs - tSendStartUs);
    sendSumUs += sendUs;
    sendCount++;
    if (sendUs > sendMaxUs) sendMaxUs = sendUs;

    // Periodic perf log (once per second) — helps track frame rates.
    final now = DateTime.now();
    if (now.difference(lastLog).inSeconds >= 1) {
      // final avgGap = gapCount > 0 ? (sumGapMs ~/ gapCount) : 0;
      // final avgNative = nativeCount > 0 ? (nativeSumMs ~/ nativeCount) : 0;
      // final avgCopyUs = copyCount > 0 ? (copySumUs ~/ copyCount) : 0;
      // final avgSendUs = sendCount > 0 ? (sendSumUs ~/ sendCount) : 0;
      // print('[ffi-isolate] NTGameFrame fwd=$framesForwarded '
      //     'gap_max=${maxGapMs}ms gap_avg=${avgGap}ms gap>=100ms=$gapsOver100 gap>=500ms=$gapsOver500 '
      //     'loop_gap_max=${loopMaxGapMs}ms loop_gap>=100ms=$loopGapsOver100 loop_gap>=500ms=$loopGapsOver500 '
      //     'native_max=${nativeMaxMs}ms native_avg=${avgNative}ms '
      //     'copy_max=${copyMaxUs}us copy_avg=${avgCopyUs}us '
      //     'send_max=${sendMaxUs}us send_avg=${avgSendUs}us '
      //     'payload_max=${payloadMax}B');
      // Also send a perf notification to the UI isolate.
      sp.send([0, false, NTPerfFwd, framesForwarded]);
      // Reset counters for next interval.
      framesForwarded = 0;
      maxGapMs = 0;
      sumGapMs = 0;
      gapCount = 0;
      gapsOver100 = 0;
      gapsOver500 = 0;
      loopMaxGapMs = 0;
      loopGapsOver100 = 0;
      loopGapsOver500 = 0;
      nativeMaxMs = 0;
      nativeSumMs = 0;
      nativeCount = 0;
      copyMaxUs = 0;
      copySumUs = 0;
      copyCount = 0;
      sendMaxUs = 0;
      sendSumUs = 0;
      sendCount = 0;
      payloadMax = 0;
      lastLog = now;
    }
  }
}

// BaseDesktopPlatform is a mixin that fulfills the GolibPluginPlatform interface
// by loading a dynamic library (.so, .dynlib, .dll) and redirecting all calls to
// that library.
mixin BaseDesktopPlatform on NtfStreams {
  String get majorPlatform => "desktop";
  int id = 1;

  final Map<int, Completer<dynamic>> calls = {};

  // Reference to the dynamic library.
  final DynamicLibrary _lib = DynamicLibrary.open(desktopLibPath());

  // The following fields are references to the dynamic library functions. They
  // are lazily initialized when first used.
  late final SetTagFunc _setTag =
      _lib.lookupFunction<SetTagNative, SetTagFunc>('SetTag');
  late final HelloFunc _hello =
      _lib.lookupFunction<HelloNative, HelloFunc>('Hello');
  late final GetURLNative _getURL =
      _lib.lookupFunction<GetURLNative, GetURLNative>('GetURL');
  late final NextTimeNative _nextTime =
      _lib.lookupFunction<NextTimeNative, NextTimeNative>('NextTime');
  late final WriteStrFunc _writeStr =
      _lib.lookupFunction<WriteStrNative, WriteStrFunc>('WriteStr');
  late final AsyncCallFunc _asyncCall =
      _lib.lookupFunction<AsyncCallNative, AsyncCallFunc>('AsyncCall');

  // From here on are the actual functions to fulfill the GolibPluginPlatform
  // interface by calling into the dynlib.

  Future<void> setTag(String tag) async => _setTag(tag.toNativeUtf8());
  Future<void> hello() async => _hello();
  Future<String> nextTime() async => _nextTime().toDartString();
  Future<void> writeStr(String s) async => _writeStr(s.toNativeUtf8());

  Stream<String> readStream() async* {
    var rp = ReceivePort();
    Isolate.spawn(_readStrIsolate, _ReadStrData(rp.sendPort));
    while (true) {
      await for (String msg in rp) {
        yield msg;
      }
    }
  }

  Future<String> getURL(String url) async {
    GetURLResultNative res = _getURL(url.toNativeUtf8());
    if (res.err.address != nullptr.address) {
      var errStr = res.err.toDartString();
      if (errStr != "") {
        throw errStr;
      }
    }

    return res.res.toDartString();
  }

  Future<dynamic> asyncCall(int cmd, dynamic payload) {
    // Use a fixed clientHandle as we currently only support a single client per UI.
    const clientHandle = 0x12131400;

    var p = jsonEncode(payload).toNativeUtf8();
    var cid = id == -1 ? 1 : id++; // skips 0 as id.
    var c = Completer<dynamic>();
    calls[cid] = c;
    _asyncCall(cmd, cid, clientHandle, p, p.length);
    calloc.free(p);
    return c.future;
  }

  void readAsyncResults() async {
    var rp = ReceivePort();
    Isolate.spawn(_readAsyncResultsIsolate, rp.sendPort);
    while (true) {
      await for (List cmdReply in rp) {
        if (cmdReply.length < 3) {
          debugPrint("Received wrong nb of elements from isolate: $cmdReply");
          continue;
        }
        int id = cmdReply[0];
        bool isError = cmdReply[1];
        int cmdType = cmdReply[2];
        dynamic payload = cmdReply[3];

        var c = calls[id];
        if (c == null) {
          if (id == 0 && cmdType >= notificationsStartID) {
            try {
              handleNotifications(cmdType, isError, payload);
            } catch (exception, trace) {
              // Probably a decode error. Keep handling stuff.
              var err =
                  "Unable to handle notification ${cmdType.toRadixString(16)}: $exception\n$trace";
              debugPrint(
                  "Error notification from golib: $err\nPayload: $payload");
              // ignore: use_rethrow_when_possible
              (() async => throw exception)();
            }
          } else {
            debugPrint("Received reply for unknown call $id - $cmdReply");
          }

          continue;
        }
        calls.remove(id);

        // Move JSON decode to microtask to avoid blocking the hot path
        if (payload is String && payload.isNotEmpty) {
          scheduleMicrotask(() {
            dynamic response;
            try {
              response = jsonDecode(payload);
            } catch (e) {
              // If decode fails, complete with error
              if (isError) {
                c.completeError(payload);
              } else {
                c.completeError(e);
              }
              return;
            }
            if (isError) {
              c.completeError(response);
            } else {
              c.complete(response);
            }
          });
        } else {
          // No JSON decode needed, complete immediately
          if (isError) {
            c.completeError(payload);
          } else {
            c.complete(payload);
          }
        }
      }
    }
  }
}
