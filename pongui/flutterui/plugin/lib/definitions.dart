// ignore_for_file: constant_identifier_names

import 'dart:async';
import 'dart:convert';

import 'package:flutter/cupertino.dart';
import 'package:json_annotation/json_annotation.dart';
import 'package:golib_plugin/grpc/generated/pong.pbgrpc.dart';

part 'definitions.g.dart';

@JsonSerializable()
class InitClient {
  @JsonKey(name: 'server_addr')
  final String serverAddr;
  @JsonKey(name: 'grpc_cert_path')
  final String grpcCertPath;
  @JsonKey(name: 'datadir')
  final String dataDir;
  @JsonKey(name: 'log_file')
  final String logFile;
  @JsonKey(name: "msgs_root")
  final String msgsRoot;
  @JsonKey(name: 'debug_level')
  final String debugLevel;
  @JsonKey(name: 'wants_log_ntfns')
  final bool wantsLogNtfns;

  // rpc fields
  @JsonKey(name: 'rpc_websocket_url')
  final String rpcWebsockeURL;
  @JsonKey(name: 'rpc_cert_path')
  final String rpcCertPath;
  @JsonKey(name: 'rpc_client_cert_path')
  final String rpcClientCertpath;
  @JsonKey(name: 'rpc_client_key_path')
  final String rpcClientKeypath;
  @JsonKey(name: 'rpc_user')
  final String rpcUser;
  @JsonKey(name: 'rpc_pass')
  final String rpcPass;

  InitClient(
    this.serverAddr,
    this.grpcCertPath,
    this.dataDir,
    this.logFile,
    this.msgsRoot,
    this.debugLevel,
    this.wantsLogNtfns,
    this.rpcWebsockeURL,
    this.rpcCertPath,
    this.rpcClientCertpath,
    this.rpcClientKeypath,
    this.rpcUser,
    this.rpcPass,
  );

  Map<String, dynamic> toJson() => _$InitClientToJson(this);
}

@JsonSerializable()
class IDInit {
  @JsonKey(name: 'id')
  final String uid;
  @JsonKey(name: 'nick')
  final String nick;
  IDInit(this.uid, this.nick);
  factory IDInit.fromJson(Map<String, dynamic> json) => _$IDInitFromJson(json);

  Map<String, dynamic> toJson() => _$IDInitToJson(this);
}

@JsonSerializable()
class GetUserNickArgs {
  @JsonKey(name: 'uid')
  final String uid;

  GetUserNickArgs(this.uid);
  Map<String, dynamic> toJson() => _$GetUserNickArgsToJson(this);
}

@JsonSerializable()
class LocalPlayer {
  @JsonKey(name: 'uid')
  final String uid;
  @JsonKey(name: 'nick')
  final String? nick;
  @JsonKey(name: 'bet_amt')
  final int betAmount;
  @JsonKey(name: 'ready')
  bool ready;

  LocalPlayer(
    this.uid,
    this.nick,
    this.betAmount, {
    this.ready = false,
  });

  factory LocalPlayer.fromJson(Map<String, dynamic> json) =>
      _$LocalPlayerFromJson(json);
  Map<String, dynamic> toJson() => _$LocalPlayerToJson(this);

  factory LocalPlayer.fromProto(Player player) {
    return LocalPlayer(
      player.uid,
      player.nick,
      player.betAmt.toInt(),
      ready: player.ready,
    );
  }
}

@JsonSerializable()
class LocalWaitingRoom {
  @JsonKey(name: 'id')
  final String id;
  @JsonKey(name: 'host_id')
  final String host;
  @JsonKey(name: 'bet_amt')
  final int betAmt;
  @JsonKey(name: 'players', defaultValue: [])
  final List<LocalPlayer> players;

  const LocalWaitingRoom(
    this.id,
    this.host,
    this.betAmt, {
    this.players = const [],
  });

  factory LocalWaitingRoom.fromJson(Map<String, dynamic> json) =>
      _$LocalWaitingRoomFromJson(json);
  Map<String, dynamic> toJson() => _$LocalWaitingRoomToJson(this);

  factory LocalWaitingRoom.fromProto(WaitingRoom wr) {
    return LocalWaitingRoom(
      wr.id,
      wr.hostId,
      wr.betAmt.toInt(),
      players: wr.players
          .map((player) => LocalPlayer(
                player.uid,
                player.nick,
                player.betAmt.toInt(),
                ready: player.ready,
              ))
          .toList(),
    );
  }
}

@JsonSerializable()
class LocalInfo {
  final String id;
  final String nick;
  LocalInfo(this.id, this.nick);
  factory LocalInfo.fromJson(Map<String, dynamic> json) =>
      _$LocalInfoFromJson(json);
}

@JsonSerializable()
class ServerCert {
  @JsonKey(name: "inner_fingerprint")
  final String innerFingerprint;
  @JsonKey(name: "outer_fingerprint")
  final String outerFingerprint;
  const ServerCert(this.innerFingerprint, this.outerFingerprint);

  factory ServerCert.fromJson(Map<String, dynamic> json) =>
      _$ServerCertFromJson(json);
}

const connStateOffline = 0;
const connStateCheckingWallet = 1;
const connStateOnline = 2;

@JsonSerializable()
class ServerInfo {
  final String innerFingerprint;
  final String outerFingerprint;
  final String serverAddr;
  const ServerInfo(
      {required this.innerFingerprint,
      required this.outerFingerprint,
      required this.serverAddr});
  const ServerInfo.empty()
      : this(innerFingerprint: "", outerFingerprint: "", serverAddr: "");

  factory ServerInfo.fromJson(Map<String, dynamic> json) =>
      _$ServerInfoFromJson(json);
}

@JsonSerializable()
class RemoteUser {
  final String uid;
  final String nick;

  const RemoteUser(this.uid, this.nick);

  factory RemoteUser.fromJson(Map<String, dynamic> json) =>
      _$RemoteUserFromJson(json);
}

@JsonSerializable()
class PublicIdentity {
  final String name;
  final String nick;
  final String identity;

  PublicIdentity(this.name, this.nick, this.identity);
  factory PublicIdentity.fromJson(Map<String, dynamic> json) =>
      _$PublicIdentityFromJson(json);
}

@JsonSerializable()
class Account {
  final String name;
  @JsonKey(name: "unconfirmed_balance")
  final int unconfirmedBalance;
  @JsonKey(name: "confirmed_balance")
  final int confirmedBalance;
  @JsonKey(name: "internal_key_count")
  final int internalKeyCount;
  @JsonKey(name: "external_key_count")
  final int externalKeyCount;

  Account(this.name, this.unconfirmedBalance, this.confirmedBalance,
      this.internalKeyCount, this.externalKeyCount);

  factory Account.fromJson(Map<String, dynamic> json) =>
      _$AccountFromJson(json);
}

@JsonSerializable()
class LogEntry {
  final String from;
  final String message;
  final bool internal;
  final int timestamp;
  LogEntry(this.from, this.message, this.internal, this.timestamp);

  factory LogEntry.fromJson(Map<String, dynamic> json) =>
      _$LogEntryFromJson(json);
}

@JsonSerializable()
class SendOnChain {
  final String addr;
  final int amount;
  @JsonKey(name: "from_account")
  final String fromAccount;

  SendOnChain(this.addr, this.amount, this.fromAccount);
  Map<String, dynamic> toJson() => _$SendOnChainToJson(this);
}

@JsonSerializable()
class LoadUserHistory {
  final String uid;
  @JsonKey(name: "is_gc")
  final bool isGC;
  final int page;
  @JsonKey(name: "page_num")
  final int pageNum;

  LoadUserHistory(this.uid, this.isGC, this.page, this.pageNum);
  Map<String, dynamic> toJson() => _$LoadUserHistoryToJson(this);
}

@JsonSerializable()
class WriteInvite {
  @JsonKey(name: "fund_amount")
  final int fundAmount;
  @JsonKey(name: "fund_account")
  final String fundAccount;
  @JsonKey(name: "gc_id")
  final String? gcid;
  final bool prepaid;

  WriteInvite(this.fundAmount, this.fundAccount, this.gcid, this.prepaid);
  Map<String, dynamic> toJson() => _$WriteInviteToJson(this);
}

@JsonSerializable()
class RedeemedInviteFunds {
  final String txid;
  final int total;

  RedeemedInviteFunds(this.txid, this.total);
  factory RedeemedInviteFunds.fromJson(Map<String, dynamic> json) =>
      _$RedeemedInviteFundsFromJson(json);
}

@JsonSerializable()
class CreateWaitingRoomArgs {
  @JsonKey(name: 'client_id')
  final String clientId;
  @JsonKey(name: 'bet_amt')
  final int betAmt;
  @JsonKey(name: 'escrow_id')
  final String? escrowId;

  CreateWaitingRoomArgs(this.clientId, this.betAmt, {this.escrowId});

  Map<String, dynamic> toJson() => _$CreateWaitingRoomArgsToJson(this);

  factory CreateWaitingRoomArgs.fromJson(Map<String, dynamic> json) =>
      _$CreateWaitingRoomArgsFromJson(json);
}

@JsonSerializable()
class RunState {
  @JsonKey(name: "dcrlnd_running")
  final bool dcrlndRunning;
  @JsonKey(name: "client_running")
  final bool clientRunning;

  RunState({required this.dcrlndRunning, required this.clientRunning});
  factory RunState.fromJson(Map<String, dynamic> json) =>
      _$RunStateFromJson(json);
}

@JsonSerializable()
class ZipLogsArgs {
  @JsonKey(name: "include_golib")
  final bool includeGolib;
  @JsonKey(name: "include_ln")
  final bool includeLn;
  @JsonKey(name: "only_last_file")
  final bool onlyLastFile;
  @JsonKey(name: "dest_path")
  final String destPath;

  ZipLogsArgs(
      this.includeGolib, this.includeLn, this.onlyLastFile, this.destPath);
  Map<String, dynamic> toJson() => _$ZipLogsArgsToJson(this);
}

const UINtfnPM = "pm";
const UINtfnGCM = "gcm";
const UINtfnGCMMention = "gcmmention";
const UINtfnMultiple = "multiple";

@JsonSerializable()
class UINotification {
  final String type;
  final String text;
  final int count;
  final String from;

  UINotification(this.type, this.text, this.count, this.from);
  factory UINotification.fromJson(Map<String, dynamic> json) =>
      _$UINotificationFromJson(json);
}

@JsonSerializable()
class UINotificationsConfig {
  final bool pms;
  final bool gcms;
  @JsonKey(name: "gcmentions")
  final bool gcMentions;

  UINotificationsConfig(this.pms, this.gcms, this.gcMentions);
  factory UINotificationsConfig.disabled() =>
      UINotificationsConfig(false, false, false);
  factory UINotificationsConfig.fromJson(Map<String, dynamic> json) =>
      _$UINotificationsConfigFromJson(json);
  Map<String, dynamic> toJson() => _$UINotificationsConfigToJson(this);
}

mixin NtfStreams {
  StreamController<RemoteUser> ntfAcceptedInvites =
      StreamController<RemoteUser>();
  Stream<RemoteUser> acceptedInvites() => ntfAcceptedInvites.stream;

  StreamController<String> ntfLogLines = StreamController<String>();
  Stream<String> logLines() => ntfLogLines.stream;

  StreamController<int> ntfRescanProgress = StreamController<int>();
  Stream<int> rescanWalletProgress() => ntfRescanProgress.stream;

  StreamController<UINotification> ntfUINotifications =
      StreamController<UINotification>();
  Stream<UINotification> uiNotifications() => ntfUINotifications.stream;

  handleNotifications(int cmd, bool isError, String jsonPayload) {
    if (jsonPayload != "") {
      // ignore: unused_local_variable
      final dynamic _ = jsonDecode(jsonPayload);
    }

    switch (cmd) {
      case NTNOP:
        // NOP.
        break;
      // case NTPM:
      //   isError
      //       ? ntfChatEvents.addError(payload)
      //       : ntfChatEvents.add(PM.fromJson(payload));
      //   break;

      default:
        debugPrint("Received unknown notification ${cmd.toRadixString(16)}");
    }
  }
}

abstract class PluginPlatform {
  Future<String?> get platformVersion => throw "unimplemented";
  String get majorPlatform => "unknown-major-plat";
  String get minorPlatform => "unknown-minor-plat";
  Future<void> setTag(String tag) async => throw "unimplemented";
  Future<void> hello() async => throw "unimplemented";
  Future<String> getURL(String url) async => throw "unimplemented";
  Future<String> nextTime() async => throw "unimplemented";
  Future<void> writeStr(String s) async => throw "unimplemented";
  Stream<String> readStream() async* {
    throw "unimplemented";
  }

  // These are only implemented in android.
  Future<void> startForegroundSvc() => throw "unimplemented";
  Future<void> stopForegroundSvc() => throw "unimplemented";
  Future<void> setNtfnsEnabled(bool enabled) => throw "unimplemented";

  Future<dynamic> asyncCall(int cmd, dynamic payload) async =>
      throw "unimplemented";

  Future<String> asyncHello(String name) async {
    var r = await asyncCall(CTHello, name);
    return r as String;
  }

  Future<LocalInfo> initClient(InitClient args) async {
    var res = await asyncCall(CTInitClient, args);
    return LocalInfo.fromJson(res as Map<String, dynamic>);
  }

  Future<void> createLockFile(String rootDir) async =>
      await asyncCall(CTCreateLockFile, rootDir);
  Future<void> closeLockFile(String rootDir) async =>
      await asyncCall(CTCloseLockFile, rootDir);
  Future<String> userNick(String pid) async {
    return await asyncCall(CTGetUserNick, pid);
  }

  Future<List<LocalPlayer>> getWRPlayers() async {
    var res = await asyncCall(CTGetWRPlayers, "");
    if (res == null) {
      return [];
    }
    return (res as List)
        .map<LocalPlayer>((v) => LocalPlayer.fromJson(v))
        .toList();
  }

  Future<List<LocalWaitingRoom>> getWaitingRooms() async {
    var res = await asyncCall(CTGetWaitingRooms, "");
    if (res == null) {
      return [];
    }
    return (res as List).map<LocalWaitingRoom>((v) {
      return LocalWaitingRoom.fromJson(v);
    }).toList();
  }

  Future<LocalWaitingRoom> JoinWaitingRoom(String id, {String? escrowId}) async {
    try {
      // Always send JSON object so golib handler can consistently parse
      final payload = {
        'room_id': id,
        'escrow_id': escrowId ?? '',
      };
      final response = await asyncCall(CTJoinWaitingRoom, payload);

      if (response is Map<String, dynamic>) {
        return LocalWaitingRoom.fromJson(response);
      } else {
        throw Exception("Invalid response format: $response");
      }
    } catch (err) {
      throw Exception("Failed to join waiting room: $err");
    }
  }

  Future<LocalWaitingRoom> CreateWaitingRoom(CreateWaitingRoomArgs args) async {
    try {
      // Always ensure escrow_id is present in the payload
      final payload = {
        'client_id': args.clientId,
        'bet_amt': args.betAmt,
        'escrow_id': args.escrowId ?? '',
      };
      final response = await asyncCall(CTCreateWaitingRoom, payload);

      if (response is Map<String, dynamic>) {
        return LocalWaitingRoom.fromJson(response);
      } else {
        throw Exception("Invalid response format: $response");
      }
    } catch (err) {
      throw Exception("Failed to join waiting room: $err");
    }
  }

  Future<void> LeaveWaitingRoom(String id) async {
    await asyncCall(CTLeaveWaitingRoom, id);
  }

  // Escrow/Settlement methods
  Future<Map<String, String>> generateSettlementSessionKey() async {
    final res = await asyncCall(CTGenerateSessionKey, "");
    return Map<String, String>.from(res as Map);
  }

  Future<Map<String, dynamic>> openEscrow({required String payout, required int betAtoms, int csvBlocks = 64}) async {
    final payload = {
      'payout': payout,
      'bet_atoms': betAtoms,
      'csv_blocks': csvBlocks,
    };
    final res = await asyncCall(CTOpenEscrow, payload);
    return Map<String, dynamic>.from(res as Map);
  }

  Future<void> startPreSign(String matchId) async {
    await asyncCall(CTStartPreSign, { 'match_id': matchId });
  }
}

const int CTUnknown = 0x00;
const int CTHello = 0x01;
const int CTInitClient = 0x02;
const int CTGetUserNick = 0x03;
const int CTCreateLockFile = 0x04;
const int CTGetWRPlayers = 0x05;
const int CTGetWaitingRooms = 0x06;
const int CTJoinWaitingRoom = 0x07;
const int CTCreateWaitingRoom = 0x08;
const int CTLeaveWaitingRoom = 0x09;
const int CTGenerateSessionKey = 0x0a;
const int CTOpenEscrow        = 0x0b;
const int CTStartPreSign      = 0x0c;
const int CTCloseLockFile = 0x60;

const int notificationsStartID = 0x1000;
const int notificationClientStopped = 0x1001;
const int NTNOP = 0X1004;
