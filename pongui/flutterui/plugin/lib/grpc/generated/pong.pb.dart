// This is a generated file - do not edit.
//
// Generated from pong.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'pong.pbenum.dart';

export 'package:protobuf/protobuf.dart' show GeneratedMessageGenericExtensions;

export 'pong.pbenum.dart';

class RequestNonceRequest extends $pb.GeneratedMessage {
  factory RequestNonceRequest({
    $core.String? userId,
  }) {
    final result = create();
    if (userId != null) result.userId = userId;
    return result;
  }

  RequestNonceRequest._();

  factory RequestNonceRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RequestNonceRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RequestNonceRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'userId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RequestNonceRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RequestNonceRequest copyWith(void Function(RequestNonceRequest) updates) =>
      super.copyWith((message) => updates(message as RequestNonceRequest))
          as RequestNonceRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RequestNonceRequest create() => RequestNonceRequest._();
  @$core.override
  RequestNonceRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static RequestNonceRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RequestNonceRequest>(create);
  static RequestNonceRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get userId => $_getSZ(0);
  @$pb.TagNumber(1)
  set userId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUserId() => $_has(0);
  @$pb.TagNumber(1)
  void clearUserId() => $_clearField(1);
}

class RequestNonceResponse extends $pb.GeneratedMessage {
  factory RequestNonceResponse({
    $core.String? nonce,
    $core.int? ttlSec,
    $core.String? addressHint,
  }) {
    final result = create();
    if (nonce != null) result.nonce = nonce;
    if (ttlSec != null) result.ttlSec = ttlSec;
    if (addressHint != null) result.addressHint = addressHint;
    return result;
  }

  RequestNonceResponse._();

  factory RequestNonceResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory RequestNonceResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'RequestNonceResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'nonce')
    ..aI(2, _omitFieldNames ? '' : 'ttlSec')
    ..aOS(3, _omitFieldNames ? '' : 'addressHint')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RequestNonceResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  RequestNonceResponse copyWith(void Function(RequestNonceResponse) updates) =>
      super.copyWith((message) => updates(message as RequestNonceResponse))
          as RequestNonceResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RequestNonceResponse create() => RequestNonceResponse._();
  @$core.override
  RequestNonceResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static RequestNonceResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<RequestNonceResponse>(create);
  static RequestNonceResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get nonce => $_getSZ(0);
  @$pb.TagNumber(1)
  set nonce($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasNonce() => $_has(0);
  @$pb.TagNumber(1)
  void clearNonce() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get ttlSec => $_getIZ(1);
  @$pb.TagNumber(2)
  set ttlSec($core.int value) => $_setSignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasTtlSec() => $_has(1);
  @$pb.TagNumber(2)
  void clearTtlSec() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get addressHint => $_getSZ(2);
  @$pb.TagNumber(3)
  set addressHint($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasAddressHint() => $_has(2);
  @$pb.TagNumber(3)
  void clearAddressHint() => $_clearField(3);
}

class VerifyLoginRequest extends $pb.GeneratedMessage {
  factory VerifyLoginRequest({
    $core.String? address,
    $core.String? nonce,
    $core.String? signature,
  }) {
    final result = create();
    if (address != null) result.address = address;
    if (nonce != null) result.nonce = nonce;
    if (signature != null) result.signature = signature;
    return result;
  }

  VerifyLoginRequest._();

  factory VerifyLoginRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory VerifyLoginRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VerifyLoginRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'address')
    ..aOS(2, _omitFieldNames ? '' : 'nonce')
    ..aOS(3, _omitFieldNames ? '' : 'signature')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VerifyLoginRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VerifyLoginRequest copyWith(void Function(VerifyLoginRequest) updates) =>
      super.copyWith((message) => updates(message as VerifyLoginRequest))
          as VerifyLoginRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static VerifyLoginRequest create() => VerifyLoginRequest._();
  @$core.override
  VerifyLoginRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static VerifyLoginRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<VerifyLoginRequest>(create);
  static VerifyLoginRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get address => $_getSZ(0);
  @$pb.TagNumber(1)
  set address($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAddress() => $_has(0);
  @$pb.TagNumber(1)
  void clearAddress() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get nonce => $_getSZ(1);
  @$pb.TagNumber(2)
  set nonce($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNonce() => $_has(1);
  @$pb.TagNumber(2)
  void clearNonce() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get signature => $_getSZ(2);
  @$pb.TagNumber(3)
  set signature($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSignature() => $_has(2);
  @$pb.TagNumber(3)
  void clearSignature() => $_clearField(3);
}

class VerifyLoginResponse extends $pb.GeneratedMessage {
  factory VerifyLoginResponse({
    $core.bool? ok,
    $core.String? token,
    $core.String? clientId,
    $core.List<$core.int>? compPubkey,
    $core.String? p2pkAddr,
  }) {
    final result = create();
    if (ok != null) result.ok = ok;
    if (token != null) result.token = token;
    if (clientId != null) result.clientId = clientId;
    if (compPubkey != null) result.compPubkey = compPubkey;
    if (p2pkAddr != null) result.p2pkAddr = p2pkAddr;
    return result;
  }

  VerifyLoginResponse._();

  factory VerifyLoginResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory VerifyLoginResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VerifyLoginResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'ok')
    ..aOS(2, _omitFieldNames ? '' : 'token')
    ..aOS(3, _omitFieldNames ? '' : 'clientId')
    ..a<$core.List<$core.int>>(
        4, _omitFieldNames ? '' : 'compPubkey', $pb.PbFieldType.OY)
    ..aOS(5, _omitFieldNames ? '' : 'p2pkAddr')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VerifyLoginResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VerifyLoginResponse copyWith(void Function(VerifyLoginResponse) updates) =>
      super.copyWith((message) => updates(message as VerifyLoginResponse))
          as VerifyLoginResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static VerifyLoginResponse create() => VerifyLoginResponse._();
  @$core.override
  VerifyLoginResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static VerifyLoginResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<VerifyLoginResponse>(create);
  static VerifyLoginResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get ok => $_getBF(0);
  @$pb.TagNumber(1)
  set ok($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasOk() => $_has(0);
  @$pb.TagNumber(1)
  void clearOk() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get token => $_getSZ(1);
  @$pb.TagNumber(2)
  set token($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasToken() => $_has(1);
  @$pb.TagNumber(2)
  void clearToken() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get clientId => $_getSZ(2);
  @$pb.TagNumber(3)
  set clientId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasClientId() => $_has(2);
  @$pb.TagNumber(3)
  void clearClientId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.List<$core.int> get compPubkey => $_getN(3);
  @$pb.TagNumber(4)
  set compPubkey($core.List<$core.int> value) => $_setBytes(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCompPubkey() => $_has(3);
  @$pb.TagNumber(4)
  void clearCompPubkey() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get p2pkAddr => $_getSZ(4);
  @$pb.TagNumber(5)
  set p2pkAddr($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasP2pkAddr() => $_has(4);
  @$pb.TagNumber(5)
  void clearP2pkAddr() => $_clearField(5);
}

enum ClientMsg_Kind { hello, ack, verifyOk, notSet }

/// === Phase 1 streaming messages ===
class ClientMsg extends $pb.GeneratedMessage {
  factory ClientMsg({
    $core.String? matchId,
    Hello? hello,
    Ack? ack,
    VerifyOk? verifyOk,
  }) {
    final result = create();
    if (matchId != null) result.matchId = matchId;
    if (hello != null) result.hello = hello;
    if (ack != null) result.ack = ack;
    if (verifyOk != null) result.verifyOk = verifyOk;
    return result;
  }

  ClientMsg._();

  factory ClientMsg.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ClientMsg.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, ClientMsg_Kind> _ClientMsg_KindByTag = {
    10: ClientMsg_Kind.hello,
    12: ClientMsg_Kind.ack,
    13: ClientMsg_Kind.verifyOk,
    0: ClientMsg_Kind.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ClientMsg',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..oo(0, [10, 12, 13])
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..aOM<Hello>(10, _omitFieldNames ? '' : 'hello', subBuilder: Hello.create)
    ..aOM<Ack>(12, _omitFieldNames ? '' : 'ack', subBuilder: Ack.create)
    ..aOM<VerifyOk>(13, _omitFieldNames ? '' : 'verifyOk',
        subBuilder: VerifyOk.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ClientMsg clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ClientMsg copyWith(void Function(ClientMsg) updates) =>
      super.copyWith((message) => updates(message as ClientMsg)) as ClientMsg;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ClientMsg create() => ClientMsg._();
  @$core.override
  ClientMsg createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ClientMsg getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ClientMsg>(create);
  static ClientMsg? _defaultInstance;

  @$pb.TagNumber(10)
  @$pb.TagNumber(12)
  @$pb.TagNumber(13)
  ClientMsg_Kind whichKind() => _ClientMsg_KindByTag[$_whichOneof(0)]!;
  @$pb.TagNumber(10)
  @$pb.TagNumber(12)
  @$pb.TagNumber(13)
  void clearKind() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => $_clearField(1);

  @$pb.TagNumber(10)
  Hello get hello => $_getN(1);
  @$pb.TagNumber(10)
  set hello(Hello value) => $_setField(10, value);
  @$pb.TagNumber(10)
  $core.bool hasHello() => $_has(1);
  @$pb.TagNumber(10)
  void clearHello() => $_clearField(10);
  @$pb.TagNumber(10)
  Hello ensureHello() => $_ensure(1);

  @$pb.TagNumber(12)
  Ack get ack => $_getN(2);
  @$pb.TagNumber(12)
  set ack(Ack value) => $_setField(12, value);
  @$pb.TagNumber(12)
  $core.bool hasAck() => $_has(2);
  @$pb.TagNumber(12)
  void clearAck() => $_clearField(12);
  @$pb.TagNumber(12)
  Ack ensureAck() => $_ensure(2);

  /// minimal handshake message carrying ack_digest and presigs.
  @$pb.TagNumber(13)
  VerifyOk get verifyOk => $_getN(3);
  @$pb.TagNumber(13)
  set verifyOk(VerifyOk value) => $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasVerifyOk() => $_has(3);
  @$pb.TagNumber(13)
  void clearVerifyOk() => $_clearField(13);
  @$pb.TagNumber(13)
  VerifyOk ensureVerifyOk() => $_ensure(3);
}

enum ServerMsg_Kind { req, info, ok, notSet }

class ServerMsg extends $pb.GeneratedMessage {
  factory ServerMsg({
    $core.String? matchId,
    NeedPreSigs? req,
    Info? info,
    ServerOk? ok,
  }) {
    final result = create();
    if (matchId != null) result.matchId = matchId;
    if (req != null) result.req = req;
    if (info != null) result.info = info;
    if (ok != null) result.ok = ok;
    return result;
  }

  ServerMsg._();

  factory ServerMsg.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ServerMsg.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static const $core.Map<$core.int, ServerMsg_Kind> _ServerMsg_KindByTag = {
    11: ServerMsg_Kind.req,
    13: ServerMsg_Kind.info,
    14: ServerMsg_Kind.ok,
    0: ServerMsg_Kind.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ServerMsg',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..oo(0, [11, 13, 14])
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..aOM<NeedPreSigs>(11, _omitFieldNames ? '' : 'req',
        subBuilder: NeedPreSigs.create)
    ..aOM<Info>(13, _omitFieldNames ? '' : 'info', subBuilder: Info.create)
    ..aOM<ServerOk>(14, _omitFieldNames ? '' : 'ok',
        subBuilder: ServerOk.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerMsg clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerMsg copyWith(void Function(ServerMsg) updates) =>
      super.copyWith((message) => updates(message as ServerMsg)) as ServerMsg;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ServerMsg create() => ServerMsg._();
  @$core.override
  ServerMsg createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ServerMsg getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ServerMsg>(create);
  static ServerMsg? _defaultInstance;

  @$pb.TagNumber(11)
  @$pb.TagNumber(13)
  @$pb.TagNumber(14)
  ServerMsg_Kind whichKind() => _ServerMsg_KindByTag[$_whichOneof(0)]!;
  @$pb.TagNumber(11)
  @$pb.TagNumber(13)
  @$pb.TagNumber(14)
  void clearKind() => $_clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => $_clearField(1);

  @$pb.TagNumber(11)
  NeedPreSigs get req => $_getN(1);
  @$pb.TagNumber(11)
  set req(NeedPreSigs value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasReq() => $_has(1);
  @$pb.TagNumber(11)
  void clearReq() => $_clearField(11);
  @$pb.TagNumber(11)
  NeedPreSigs ensureReq() => $_ensure(1);

  @$pb.TagNumber(13)
  Info get info => $_getN(2);
  @$pb.TagNumber(13)
  set info(Info value) => $_setField(13, value);
  @$pb.TagNumber(13)
  $core.bool hasInfo() => $_has(2);
  @$pb.TagNumber(13)
  void clearInfo() => $_clearField(13);
  @$pb.TagNumber(13)
  Info ensureInfo() => $_ensure(2);

  /// New handshake completion ack from server.
  @$pb.TagNumber(14)
  ServerOk get ok => $_getN(3);
  @$pb.TagNumber(14)
  set ok(ServerOk value) => $_setField(14, value);
  @$pb.TagNumber(14)
  $core.bool hasOk() => $_has(3);
  @$pb.TagNumber(14)
  void clearOk() => $_clearField(14);
  @$pb.TagNumber(14)
  ServerOk ensureOk() => $_ensure(3);
}

class Hello extends $pb.GeneratedMessage {
  factory Hello({
    $core.String? matchId,
    $core.List<$core.int>? compPubkey,
    $core.String? clientVersion,
  }) {
    final result = create();
    if (matchId != null) result.matchId = matchId;
    if (compPubkey != null) result.compPubkey = compPubkey;
    if (clientVersion != null) result.clientVersion = clientVersion;
    return result;
  }

  Hello._();

  factory Hello.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Hello.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Hello',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..a<$core.List<$core.int>>(
        2, _omitFieldNames ? '' : 'compPubkey', $pb.PbFieldType.OY)
    ..aOS(3, _omitFieldNames ? '' : 'clientVersion')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Hello clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Hello copyWith(void Function(Hello) updates) =>
      super.copyWith((message) => updates(message as Hello)) as Hello;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Hello create() => Hello._();
  @$core.override
  Hello createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Hello getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Hello>(create);
  static Hello? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get compPubkey => $_getN(1);
  @$pb.TagNumber(2)
  set compPubkey($core.List<$core.int> value) => $_setBytes(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCompPubkey() => $_has(1);
  @$pb.TagNumber(2)
  void clearCompPubkey() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get clientVersion => $_getSZ(2);
  @$pb.TagNumber(3)
  set clientVersion($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasClientVersion() => $_has(2);
  @$pb.TagNumber(3)
  void clearClientVersion() => $_clearField(3);
}

class NeedPreSigs_PerInput extends $pb.GeneratedMessage {
  factory NeedPreSigs_PerInput({
    $core.String? inputId,
    $core.String? redeemScriptHex,
    $core.String? mHex,
    $core.List<$core.int>? tCompressed,
  }) {
    final result = create();
    if (inputId != null) result.inputId = inputId;
    if (redeemScriptHex != null) result.redeemScriptHex = redeemScriptHex;
    if (mHex != null) result.mHex = mHex;
    if (tCompressed != null) result.tCompressed = tCompressed;
    return result;
  }

  NeedPreSigs_PerInput._();

  factory NeedPreSigs_PerInput.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory NeedPreSigs_PerInput.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'NeedPreSigs.PerInput',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'inputId')
    ..aOS(2, _omitFieldNames ? '' : 'redeemScriptHex')
    ..aOS(3, _omitFieldNames ? '' : 'mHex')
    ..a<$core.List<$core.int>>(
        4, _omitFieldNames ? '' : 'TCompressed', $pb.PbFieldType.OY,
        protoName: 'T_compressed')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NeedPreSigs_PerInput clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NeedPreSigs_PerInput copyWith(void Function(NeedPreSigs_PerInput) updates) =>
      super.copyWith((message) => updates(message as NeedPreSigs_PerInput))
          as NeedPreSigs_PerInput;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static NeedPreSigs_PerInput create() => NeedPreSigs_PerInput._();
  @$core.override
  NeedPreSigs_PerInput createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static NeedPreSigs_PerInput getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<NeedPreSigs_PerInput>(create);
  static NeedPreSigs_PerInput? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get inputId => $_getSZ(0);
  @$pb.TagNumber(1)
  set inputId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasInputId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInputId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get redeemScriptHex => $_getSZ(1);
  @$pb.TagNumber(2)
  set redeemScriptHex($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRedeemScriptHex() => $_has(1);
  @$pb.TagNumber(2)
  void clearRedeemScriptHex() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get mHex => $_getSZ(2);
  @$pb.TagNumber(3)
  set mHex($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasMHex() => $_has(2);
  @$pb.TagNumber(3)
  void clearMHex() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.List<$core.int> get tCompressed => $_getN(3);
  @$pb.TagNumber(4)
  set tCompressed($core.List<$core.int> value) => $_setBytes(3, value);
  @$pb.TagNumber(4)
  $core.bool hasTCompressed() => $_has(3);
  @$pb.TagNumber(4)
  void clearTCompressed() => $_clearField(4);
}

class NeedPreSigs extends $pb.GeneratedMessage {
  factory NeedPreSigs({
    $core.String? draftTxHex,
    $core.Iterable<NeedPreSigs_PerInput>? inputs,
  }) {
    final result = create();
    if (draftTxHex != null) result.draftTxHex = draftTxHex;
    if (inputs != null) result.inputs.addAll(inputs);
    return result;
  }

  NeedPreSigs._();

  factory NeedPreSigs.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory NeedPreSigs.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'NeedPreSigs',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(2, _omitFieldNames ? '' : 'draftTxHex')
    ..pPM<NeedPreSigs_PerInput>(4, _omitFieldNames ? '' : 'inputs',
        subBuilder: NeedPreSigs_PerInput.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NeedPreSigs clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NeedPreSigs copyWith(void Function(NeedPreSigs) updates) =>
      super.copyWith((message) => updates(message as NeedPreSigs))
          as NeedPreSigs;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static NeedPreSigs create() => NeedPreSigs._();
  @$core.override
  NeedPreSigs createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static NeedPreSigs getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<NeedPreSigs>(create);
  static NeedPreSigs? _defaultInstance;

  @$pb.TagNumber(2)
  $core.String get draftTxHex => $_getSZ(0);
  @$pb.TagNumber(2)
  set draftTxHex($core.String value) => $_setString(0, value);
  @$pb.TagNumber(2)
  $core.bool hasDraftTxHex() => $_has(0);
  @$pb.TagNumber(2)
  void clearDraftTxHex() => $_clearField(2);

  @$pb.TagNumber(4)
  $pb.PbList<NeedPreSigs_PerInput> get inputs => $_getList(1);
}

/// Client VERIFY_OK message: verifies draft, builds presigs, and includes ack digest.
class VerifyOk extends $pb.GeneratedMessage {
  factory VerifyOk({
    $core.List<$core.int>? ackDigest,
    $core.Iterable<PreSig>? presigs,
  }) {
    final result = create();
    if (ackDigest != null) result.ackDigest = ackDigest;
    if (presigs != null) result.presigs.addAll(presigs);
    return result;
  }

  VerifyOk._();

  factory VerifyOk.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory VerifyOk.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'VerifyOk',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..a<$core.List<$core.int>>(
        1, _omitFieldNames ? '' : 'ackDigest', $pb.PbFieldType.OY)
    ..pPM<PreSig>(2, _omitFieldNames ? '' : 'presigs',
        subBuilder: PreSig.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VerifyOk clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  VerifyOk copyWith(void Function(VerifyOk) updates) =>
      super.copyWith((message) => updates(message as VerifyOk)) as VerifyOk;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static VerifyOk create() => VerifyOk._();
  @$core.override
  VerifyOk createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static VerifyOk getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<VerifyOk>(create);
  static VerifyOk? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get ackDigest => $_getN(0);
  @$pb.TagNumber(1)
  set ackDigest($core.List<$core.int> value) => $_setBytes(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAckDigest() => $_has(0);
  @$pb.TagNumber(1)
  void clearAckDigest() => $_clearField(1);

  @$pb.TagNumber(2)
  $pb.PbList<PreSig> get presigs => $_getList(1);
}

/// Per-input pre-signature using minus variant and normalized R'.
class PreSig extends $pb.GeneratedMessage {
  factory PreSig({
    $core.String? inputId,
    $core.List<$core.int>? rLineCompressed,
    $core.List<$core.int>? sLine32,
  }) {
    final result = create();
    if (inputId != null) result.inputId = inputId;
    if (rLineCompressed != null) result.rLineCompressed = rLineCompressed;
    if (sLine32 != null) result.sLine32 = sLine32;
    return result;
  }

  PreSig._();

  factory PreSig.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory PreSig.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PreSig',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'inputId')
    ..a<$core.List<$core.int>>(
        2, _omitFieldNames ? '' : 'RLineCompressed', $pb.PbFieldType.OY,
        protoName: 'RLine_compressed')
    ..a<$core.List<$core.int>>(
        3, _omitFieldNames ? '' : 'sLine32', $pb.PbFieldType.OY,
        protoName: 'sLine32')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PreSig clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PreSig copyWith(void Function(PreSig) updates) =>
      super.copyWith((message) => updates(message as PreSig)) as PreSig;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PreSig create() => PreSig._();
  @$core.override
  PreSig createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static PreSig getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PreSig>(create);
  static PreSig? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get inputId => $_getSZ(0);
  @$pb.TagNumber(1)
  set inputId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasInputId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInputId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get rLineCompressed => $_getN(1);
  @$pb.TagNumber(2)
  set rLineCompressed($core.List<$core.int> value) => $_setBytes(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRLineCompressed() => $_has(1);
  @$pb.TagNumber(2)
  void clearRLineCompressed() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get sLine32 => $_getN(2);
  @$pb.TagNumber(3)
  set sLine32($core.List<$core.int> value) => $_setBytes(2, value);
  @$pb.TagNumber(3)
  $core.bool hasSLine32() => $_has(2);
  @$pb.TagNumber(3)
  void clearSLine32() => $_clearField(3);
}

class Ack extends $pb.GeneratedMessage {
  factory Ack({
    $core.String? note,
  }) {
    final result = create();
    if (note != null) result.note = note;
    return result;
  }

  Ack._();

  factory Ack.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Ack.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Ack',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'note')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Ack clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Ack copyWith(void Function(Ack) updates) =>
      super.copyWith((message) => updates(message as Ack)) as Ack;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Ack create() => Ack._();
  @$core.override
  Ack createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Ack getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Ack>(create);
  static Ack? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get note => $_getSZ(0);
  @$pb.TagNumber(1)
  set note($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasNote() => $_has(0);
  @$pb.TagNumber(1)
  void clearNote() => $_clearField(1);
}

class Info extends $pb.GeneratedMessage {
  factory Info({
    $core.String? text,
  }) {
    final result = create();
    if (text != null) result.text = text;
    return result;
  }

  Info._();

  factory Info.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Info.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Info',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'text')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Info clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Info copyWith(void Function(Info) updates) =>
      super.copyWith((message) => updates(message as Info)) as Info;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Info create() => Info._();
  @$core.override
  Info createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Info getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Info>(create);
  static Info? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get text => $_getSZ(0);
  @$pb.TagNumber(1)
  set text($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasText() => $_has(0);
  @$pb.TagNumber(1)
  void clearText() => $_clearField(1);
}

/// Server SERVER_OK message: acknowledges successful verification.
class ServerOk extends $pb.GeneratedMessage {
  factory ServerOk({
    $core.List<$core.int>? ackDigest,
  }) {
    final result = create();
    if (ackDigest != null) result.ackDigest = ackDigest;
    return result;
  }

  ServerOk._();

  factory ServerOk.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory ServerOk.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'ServerOk',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..a<$core.List<$core.int>>(
        1, _omitFieldNames ? '' : 'ackDigest', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerOk clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  ServerOk copyWith(void Function(ServerOk) updates) =>
      super.copyWith((message) => updates(message as ServerOk)) as ServerOk;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ServerOk create() => ServerOk._();
  @$core.override
  ServerOk createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static ServerOk getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ServerOk>(create);
  static ServerOk? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get ackDigest => $_getN(0);
  @$pb.TagNumber(1)
  set ackDigest($core.List<$core.int> value) => $_setBytes(0, value);
  @$pb.TagNumber(1)
  $core.bool hasAckDigest() => $_has(0);
  @$pb.TagNumber(1)
  void clearAckDigest() => $_clearField(1);
}

/// === Finalization bundle for winner ===
class GetFinalizeBundleRequest extends $pb.GeneratedMessage {
  factory GetFinalizeBundleRequest({
    $core.String? matchId,
    $core.String? winnerUid,
  }) {
    final result = create();
    if (matchId != null) result.matchId = matchId;
    if (winnerUid != null) result.winnerUid = winnerUid;
    return result;
  }

  GetFinalizeBundleRequest._();

  factory GetFinalizeBundleRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetFinalizeBundleRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetFinalizeBundleRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..aOS(2, _omitFieldNames ? '' : 'winnerUid')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetFinalizeBundleRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetFinalizeBundleRequest copyWith(
          void Function(GetFinalizeBundleRequest) updates) =>
      super.copyWith((message) => updates(message as GetFinalizeBundleRequest))
          as GetFinalizeBundleRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetFinalizeBundleRequest create() => GetFinalizeBundleRequest._();
  @$core.override
  GetFinalizeBundleRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetFinalizeBundleRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetFinalizeBundleRequest>(create);
  static GetFinalizeBundleRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get winnerUid => $_getSZ(1);
  @$pb.TagNumber(2)
  set winnerUid($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasWinnerUid() => $_has(1);
  @$pb.TagNumber(2)
  void clearWinnerUid() => $_clearField(2);
}

class FinalizeInput extends $pb.GeneratedMessage {
  factory FinalizeInput({
    $core.String? inputId,
    $core.String? redeemScriptHex,
    $core.List<$core.int>? rLineCompressed,
    $core.List<$core.int>? sLine32,
  }) {
    final result = create();
    if (inputId != null) result.inputId = inputId;
    if (redeemScriptHex != null) result.redeemScriptHex = redeemScriptHex;
    if (rLineCompressed != null) result.rLineCompressed = rLineCompressed;
    if (sLine32 != null) result.sLine32 = sLine32;
    return result;
  }

  FinalizeInput._();

  factory FinalizeInput.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory FinalizeInput.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'FinalizeInput',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'inputId')
    ..aOS(2, _omitFieldNames ? '' : 'redeemScriptHex')
    ..a<$core.List<$core.int>>(
        3, _omitFieldNames ? '' : 'RLineCompressed', $pb.PbFieldType.OY,
        protoName: 'RLine_compressed')
    ..a<$core.List<$core.int>>(
        4, _omitFieldNames ? '' : 'sLine32', $pb.PbFieldType.OY,
        protoName: 'sLine32')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FinalizeInput clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  FinalizeInput copyWith(void Function(FinalizeInput) updates) =>
      super.copyWith((message) => updates(message as FinalizeInput))
          as FinalizeInput;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static FinalizeInput create() => FinalizeInput._();
  @$core.override
  FinalizeInput createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static FinalizeInput getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<FinalizeInput>(create);
  static FinalizeInput? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get inputId => $_getSZ(0);
  @$pb.TagNumber(1)
  set inputId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasInputId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInputId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get redeemScriptHex => $_getSZ(1);
  @$pb.TagNumber(2)
  set redeemScriptHex($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRedeemScriptHex() => $_has(1);
  @$pb.TagNumber(2)
  void clearRedeemScriptHex() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get rLineCompressed => $_getN(2);
  @$pb.TagNumber(3)
  set rLineCompressed($core.List<$core.int> value) => $_setBytes(2, value);
  @$pb.TagNumber(3)
  $core.bool hasRLineCompressed() => $_has(2);
  @$pb.TagNumber(3)
  void clearRLineCompressed() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.List<$core.int> get sLine32 => $_getN(3);
  @$pb.TagNumber(4)
  set sLine32($core.List<$core.int> value) => $_setBytes(3, value);
  @$pb.TagNumber(4)
  $core.bool hasSLine32() => $_has(3);
  @$pb.TagNumber(4)
  void clearSLine32() => $_clearField(4);
}

class GetFinalizeBundleResponse extends $pb.GeneratedMessage {
  factory GetFinalizeBundleResponse({
    $core.String? draftTxHex,
    $core.List<$core.int>? gamma32,
    $core.Iterable<FinalizeInput>? inputs,
  }) {
    final result = create();
    if (draftTxHex != null) result.draftTxHex = draftTxHex;
    if (gamma32 != null) result.gamma32 = gamma32;
    if (inputs != null) result.inputs.addAll(inputs);
    return result;
  }

  GetFinalizeBundleResponse._();

  factory GetFinalizeBundleResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GetFinalizeBundleResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GetFinalizeBundleResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'draftTxHex')
    ..a<$core.List<$core.int>>(
        2, _omitFieldNames ? '' : 'gamma32', $pb.PbFieldType.OY)
    ..pPM<FinalizeInput>(3, _omitFieldNames ? '' : 'inputs',
        subBuilder: FinalizeInput.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetFinalizeBundleResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GetFinalizeBundleResponse copyWith(
          void Function(GetFinalizeBundleResponse) updates) =>
      super.copyWith((message) => updates(message as GetFinalizeBundleResponse))
          as GetFinalizeBundleResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GetFinalizeBundleResponse create() => GetFinalizeBundleResponse._();
  @$core.override
  GetFinalizeBundleResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GetFinalizeBundleResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GetFinalizeBundleResponse>(create);
  static GetFinalizeBundleResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get draftTxHex => $_getSZ(0);
  @$pb.TagNumber(1)
  set draftTxHex($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasDraftTxHex() => $_has(0);
  @$pb.TagNumber(1)
  void clearDraftTxHex() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get gamma32 => $_getN(1);
  @$pb.TagNumber(2)
  set gamma32($core.List<$core.int> value) => $_setBytes(1, value);
  @$pb.TagNumber(2)
  $core.bool hasGamma32() => $_has(1);
  @$pb.TagNumber(2)
  void clearGamma32() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<FinalizeInput> get inputs => $_getList(2);
}

class OpenEscrowRequest extends $pb.GeneratedMessage {
  factory OpenEscrowRequest({
    $core.String? ownerUid,
    $core.List<$core.int>? compPubkey,
    $fixnum.Int64? betAtoms,
    $core.int? csvBlocks,
    $core.List<$core.int>? payoutPubkey,
  }) {
    final result = create();
    if (ownerUid != null) result.ownerUid = ownerUid;
    if (compPubkey != null) result.compPubkey = compPubkey;
    if (betAtoms != null) result.betAtoms = betAtoms;
    if (csvBlocks != null) result.csvBlocks = csvBlocks;
    if (payoutPubkey != null) result.payoutPubkey = payoutPubkey;
    return result;
  }

  OpenEscrowRequest._();

  factory OpenEscrowRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory OpenEscrowRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'OpenEscrowRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'ownerUid')
    ..a<$core.List<$core.int>>(
        2, _omitFieldNames ? '' : 'compPubkey', $pb.PbFieldType.OY)
    ..a<$fixnum.Int64>(
        3, _omitFieldNames ? '' : 'betAtoms', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..aI(4, _omitFieldNames ? '' : 'csvBlocks', fieldType: $pb.PbFieldType.OU3)
    ..a<$core.List<$core.int>>(
        5, _omitFieldNames ? '' : 'payoutPubkey', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OpenEscrowRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OpenEscrowRequest copyWith(void Function(OpenEscrowRequest) updates) =>
      super.copyWith((message) => updates(message as OpenEscrowRequest))
          as OpenEscrowRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static OpenEscrowRequest create() => OpenEscrowRequest._();
  @$core.override
  OpenEscrowRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static OpenEscrowRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<OpenEscrowRequest>(create);
  static OpenEscrowRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get ownerUid => $_getSZ(0);
  @$pb.TagNumber(1)
  set ownerUid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasOwnerUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearOwnerUid() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get compPubkey => $_getN(1);
  @$pb.TagNumber(2)
  set compPubkey($core.List<$core.int> value) => $_setBytes(1, value);
  @$pb.TagNumber(2)
  $core.bool hasCompPubkey() => $_has(1);
  @$pb.TagNumber(2)
  void clearCompPubkey() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get betAtoms => $_getI64(2);
  @$pb.TagNumber(3)
  set betAtoms($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasBetAtoms() => $_has(2);
  @$pb.TagNumber(3)
  void clearBetAtoms() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get csvBlocks => $_getIZ(3);
  @$pb.TagNumber(4)
  set csvBlocks($core.int value) => $_setUnsignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCsvBlocks() => $_has(3);
  @$pb.TagNumber(4)
  void clearCsvBlocks() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.List<$core.int> get payoutPubkey => $_getN(4);
  @$pb.TagNumber(5)
  set payoutPubkey($core.List<$core.int> value) => $_setBytes(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPayoutPubkey() => $_has(4);
  @$pb.TagNumber(5)
  void clearPayoutPubkey() => $_clearField(5);
}

class OpenEscrowResponse extends $pb.GeneratedMessage {
  factory OpenEscrowResponse({
    $core.String? escrowId,
    $core.String? depositAddress,
    $core.String? pkScriptHex,
  }) {
    final result = create();
    if (escrowId != null) result.escrowId = escrowId;
    if (depositAddress != null) result.depositAddress = depositAddress;
    if (pkScriptHex != null) result.pkScriptHex = pkScriptHex;
    return result;
  }

  OpenEscrowResponse._();

  factory OpenEscrowResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory OpenEscrowResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'OpenEscrowResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'escrowId')
    ..aOS(2, _omitFieldNames ? '' : 'depositAddress')
    ..aOS(3, _omitFieldNames ? '' : 'pkScriptHex')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OpenEscrowResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  OpenEscrowResponse copyWith(void Function(OpenEscrowResponse) updates) =>
      super.copyWith((message) => updates(message as OpenEscrowResponse))
          as OpenEscrowResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static OpenEscrowResponse create() => OpenEscrowResponse._();
  @$core.override
  OpenEscrowResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static OpenEscrowResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<OpenEscrowResponse>(create);
  static OpenEscrowResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get escrowId => $_getSZ(0);
  @$pb.TagNumber(1)
  set escrowId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasEscrowId() => $_has(0);
  @$pb.TagNumber(1)
  void clearEscrowId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get depositAddress => $_getSZ(1);
  @$pb.TagNumber(2)
  set depositAddress($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasDepositAddress() => $_has(1);
  @$pb.TagNumber(2)
  void clearDepositAddress() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get pkScriptHex => $_getSZ(2);
  @$pb.TagNumber(3)
  set pkScriptHex($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPkScriptHex() => $_has(2);
  @$pb.TagNumber(3)
  void clearPkScriptHex() => $_clearField(3);
}

class EscrowUTXO extends $pb.GeneratedMessage {
  factory EscrowUTXO({
    $core.String? txid,
    $core.int? vout,
    $fixnum.Int64? value,
    $core.String? redeemScriptHex,
    $core.String? pkScriptHex,
    $core.String? owner,
  }) {
    final result = create();
    if (txid != null) result.txid = txid;
    if (vout != null) result.vout = vout;
    if (value != null) result.value = value;
    if (redeemScriptHex != null) result.redeemScriptHex = redeemScriptHex;
    if (pkScriptHex != null) result.pkScriptHex = pkScriptHex;
    if (owner != null) result.owner = owner;
    return result;
  }

  EscrowUTXO._();

  factory EscrowUTXO.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory EscrowUTXO.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'EscrowUTXO',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'txid')
    ..aI(2, _omitFieldNames ? '' : 'vout', fieldType: $pb.PbFieldType.OU3)
    ..a<$fixnum.Int64>(3, _omitFieldNames ? '' : 'value', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(4, _omitFieldNames ? '' : 'redeemScriptHex')
    ..aOS(5, _omitFieldNames ? '' : 'pkScriptHex')
    ..aOS(6, _omitFieldNames ? '' : 'owner')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EscrowUTXO clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  EscrowUTXO copyWith(void Function(EscrowUTXO) updates) =>
      super.copyWith((message) => updates(message as EscrowUTXO)) as EscrowUTXO;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static EscrowUTXO create() => EscrowUTXO._();
  @$core.override
  EscrowUTXO createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static EscrowUTXO getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<EscrowUTXO>(create);
  static EscrowUTXO? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get txid => $_getSZ(0);
  @$pb.TagNumber(1)
  set txid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasTxid() => $_has(0);
  @$pb.TagNumber(1)
  void clearTxid() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.int get vout => $_getIZ(1);
  @$pb.TagNumber(2)
  set vout($core.int value) => $_setUnsignedInt32(1, value);
  @$pb.TagNumber(2)
  $core.bool hasVout() => $_has(1);
  @$pb.TagNumber(2)
  void clearVout() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get value => $_getI64(2);
  @$pb.TagNumber(3)
  set value($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasValue() => $_has(2);
  @$pb.TagNumber(3)
  void clearValue() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get redeemScriptHex => $_getSZ(3);
  @$pb.TagNumber(4)
  set redeemScriptHex($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasRedeemScriptHex() => $_has(3);
  @$pb.TagNumber(4)
  void clearRedeemScriptHex() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.String get pkScriptHex => $_getSZ(4);
  @$pb.TagNumber(5)
  set pkScriptHex($core.String value) => $_setString(4, value);
  @$pb.TagNumber(5)
  $core.bool hasPkScriptHex() => $_has(4);
  @$pb.TagNumber(5)
  void clearPkScriptHex() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.String get owner => $_getSZ(5);
  @$pb.TagNumber(6)
  set owner($core.String value) => $_setString(5, value);
  @$pb.TagNumber(6)
  $core.bool hasOwner() => $_has(5);
  @$pb.TagNumber(6)
  void clearOwner() => $_clearField(6);
}

class MatchAllocatedNtfn extends $pb.GeneratedMessage {
  factory MatchAllocatedNtfn({
    $core.String? matchId,
    $core.String? roomId,
    $fixnum.Int64? betAtoms,
    $core.int? csvBlocks,
    $core.List<$core.int>? aComp,
    $core.List<$core.int>? bComp,
  }) {
    final result = create();
    if (matchId != null) result.matchId = matchId;
    if (roomId != null) result.roomId = roomId;
    if (betAtoms != null) result.betAtoms = betAtoms;
    if (csvBlocks != null) result.csvBlocks = csvBlocks;
    if (aComp != null) result.aComp = aComp;
    if (bComp != null) result.bComp = bComp;
    return result;
  }

  MatchAllocatedNtfn._();

  factory MatchAllocatedNtfn.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory MatchAllocatedNtfn.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'MatchAllocatedNtfn',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..aOS(2, _omitFieldNames ? '' : 'roomId')
    ..a<$fixnum.Int64>(
        3, _omitFieldNames ? '' : 'betAtoms', $pb.PbFieldType.OU6,
        defaultOrMaker: $fixnum.Int64.ZERO)
    ..aI(4, _omitFieldNames ? '' : 'csvBlocks', fieldType: $pb.PbFieldType.OU3)
    ..a<$core.List<$core.int>>(
        5, _omitFieldNames ? '' : 'aComp', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(
        6, _omitFieldNames ? '' : 'bComp', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MatchAllocatedNtfn clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  MatchAllocatedNtfn copyWith(void Function(MatchAllocatedNtfn) updates) =>
      super.copyWith((message) => updates(message as MatchAllocatedNtfn))
          as MatchAllocatedNtfn;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static MatchAllocatedNtfn create() => MatchAllocatedNtfn._();
  @$core.override
  MatchAllocatedNtfn createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static MatchAllocatedNtfn getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<MatchAllocatedNtfn>(create);
  static MatchAllocatedNtfn? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get roomId => $_getSZ(1);
  @$pb.TagNumber(2)
  set roomId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRoomId() => $_has(1);
  @$pb.TagNumber(2)
  void clearRoomId() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get betAtoms => $_getI64(2);
  @$pb.TagNumber(3)
  set betAtoms($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasBetAtoms() => $_has(2);
  @$pb.TagNumber(3)
  void clearBetAtoms() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get csvBlocks => $_getIZ(3);
  @$pb.TagNumber(4)
  set csvBlocks($core.int value) => $_setUnsignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasCsvBlocks() => $_has(3);
  @$pb.TagNumber(4)
  void clearCsvBlocks() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.List<$core.int> get aComp => $_getN(4);
  @$pb.TagNumber(5)
  set aComp($core.List<$core.int> value) => $_setBytes(4, value);
  @$pb.TagNumber(5)
  $core.bool hasAComp() => $_has(4);
  @$pb.TagNumber(5)
  void clearAComp() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.List<$core.int> get bComp => $_getN(5);
  @$pb.TagNumber(6)
  set bComp($core.List<$core.int> value) => $_setBytes(5, value);
  @$pb.TagNumber(6)
  $core.bool hasBComp() => $_has(5);
  @$pb.TagNumber(6)
  void clearBComp() => $_clearField(6);
}

class UnreadyGameStreamRequest extends $pb.GeneratedMessage {
  factory UnreadyGameStreamRequest({
    $core.String? clientId,
  }) {
    final result = create();
    if (clientId != null) result.clientId = clientId;
    return result;
  }

  UnreadyGameStreamRequest._();

  factory UnreadyGameStreamRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UnreadyGameStreamRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UnreadyGameStreamRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnreadyGameStreamRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnreadyGameStreamRequest copyWith(
          void Function(UnreadyGameStreamRequest) updates) =>
      super.copyWith((message) => updates(message as UnreadyGameStreamRequest))
          as UnreadyGameStreamRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UnreadyGameStreamRequest create() => UnreadyGameStreamRequest._();
  @$core.override
  UnreadyGameStreamRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static UnreadyGameStreamRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UnreadyGameStreamRequest>(create);
  static UnreadyGameStreamRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => $_clearField(1);
}

class UnreadyGameStreamResponse extends $pb.GeneratedMessage {
  factory UnreadyGameStreamResponse() => create();

  UnreadyGameStreamResponse._();

  factory UnreadyGameStreamResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory UnreadyGameStreamResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'UnreadyGameStreamResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnreadyGameStreamResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  UnreadyGameStreamResponse copyWith(
          void Function(UnreadyGameStreamResponse) updates) =>
      super.copyWith((message) => updates(message as UnreadyGameStreamResponse))
          as UnreadyGameStreamResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UnreadyGameStreamResponse create() => UnreadyGameStreamResponse._();
  @$core.override
  UnreadyGameStreamResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static UnreadyGameStreamResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<UnreadyGameStreamResponse>(create);
  static UnreadyGameStreamResponse? _defaultInstance;
}

class StartNtfnStreamRequest extends $pb.GeneratedMessage {
  factory StartNtfnStreamRequest({
    $core.String? clientId,
  }) {
    final result = create();
    if (clientId != null) result.clientId = clientId;
    return result;
  }

  StartNtfnStreamRequest._();

  factory StartNtfnStreamRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory StartNtfnStreamRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartNtfnStreamRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartNtfnStreamRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartNtfnStreamRequest copyWith(
          void Function(StartNtfnStreamRequest) updates) =>
      super.copyWith((message) => updates(message as StartNtfnStreamRequest))
          as StartNtfnStreamRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StartNtfnStreamRequest create() => StartNtfnStreamRequest._();
  @$core.override
  StartNtfnStreamRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static StartNtfnStreamRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StartNtfnStreamRequest>(create);
  static StartNtfnStreamRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => $_clearField(1);
}

class NtfnStreamResponse extends $pb.GeneratedMessage {
  factory NtfnStreamResponse({
    NotificationType? notificationType,
    $core.bool? started,
    $core.String? gameId,
    $core.String? message,
    $fixnum.Int64? betAmt,
    $core.int? playerNumber,
    $core.String? playerId,
    $core.String? roomId,
    WaitingRoom? wr,
    $core.bool? ready,
    MatchAllocatedNtfn? matchAlloc,
    $core.int? confs,
    $core.bool? serverIsF2p,
  }) {
    final result = create();
    if (notificationType != null) result.notificationType = notificationType;
    if (started != null) result.started = started;
    if (gameId != null) result.gameId = gameId;
    if (message != null) result.message = message;
    if (betAmt != null) result.betAmt = betAmt;
    if (playerNumber != null) result.playerNumber = playerNumber;
    if (playerId != null) result.playerId = playerId;
    if (roomId != null) result.roomId = roomId;
    if (wr != null) result.wr = wr;
    if (ready != null) result.ready = ready;
    if (matchAlloc != null) result.matchAlloc = matchAlloc;
    if (confs != null) result.confs = confs;
    if (serverIsF2p != null) result.serverIsF2p = serverIsF2p;
    return result;
  }

  NtfnStreamResponse._();

  factory NtfnStreamResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory NtfnStreamResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'NtfnStreamResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aE<NotificationType>(1, _omitFieldNames ? '' : 'notificationType',
        enumValues: NotificationType.values)
    ..aOB(2, _omitFieldNames ? '' : 'started')
    ..aOS(3, _omitFieldNames ? '' : 'gameId')
    ..aOS(4, _omitFieldNames ? '' : 'message')
    ..aInt64(5, _omitFieldNames ? '' : 'betAmt', protoName: 'betAmt')
    ..aI(6, _omitFieldNames ? '' : 'playerNumber')
    ..aOS(7, _omitFieldNames ? '' : 'playerId')
    ..aOS(8, _omitFieldNames ? '' : 'roomId')
    ..aOM<WaitingRoom>(9, _omitFieldNames ? '' : 'wr',
        subBuilder: WaitingRoom.create)
    ..aOB(10, _omitFieldNames ? '' : 'ready')
    ..aOM<MatchAllocatedNtfn>(11, _omitFieldNames ? '' : 'matchAlloc',
        subBuilder: MatchAllocatedNtfn.create)
    ..aI(12, _omitFieldNames ? '' : 'confs', fieldType: $pb.PbFieldType.OU3)
    ..aOB(13, _omitFieldNames ? '' : 'serverIsF2p')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NtfnStreamResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  NtfnStreamResponse copyWith(void Function(NtfnStreamResponse) updates) =>
      super.copyWith((message) => updates(message as NtfnStreamResponse))
          as NtfnStreamResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static NtfnStreamResponse create() => NtfnStreamResponse._();
  @$core.override
  NtfnStreamResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static NtfnStreamResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<NtfnStreamResponse>(create);
  static NtfnStreamResponse? _defaultInstance;

  @$pb.TagNumber(1)
  NotificationType get notificationType => $_getN(0);
  @$pb.TagNumber(1)
  set notificationType(NotificationType value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasNotificationType() => $_has(0);
  @$pb.TagNumber(1)
  void clearNotificationType() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get started => $_getBF(1);
  @$pb.TagNumber(2)
  set started($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasStarted() => $_has(1);
  @$pb.TagNumber(2)
  void clearStarted() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get gameId => $_getSZ(2);
  @$pb.TagNumber(3)
  set gameId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasGameId() => $_has(2);
  @$pb.TagNumber(3)
  void clearGameId() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.String get message => $_getSZ(3);
  @$pb.TagNumber(4)
  set message($core.String value) => $_setString(3, value);
  @$pb.TagNumber(4)
  $core.bool hasMessage() => $_has(3);
  @$pb.TagNumber(4)
  void clearMessage() => $_clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get betAmt => $_getI64(4);
  @$pb.TagNumber(5)
  set betAmt($fixnum.Int64 value) => $_setInt64(4, value);
  @$pb.TagNumber(5)
  $core.bool hasBetAmt() => $_has(4);
  @$pb.TagNumber(5)
  void clearBetAmt() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.int get playerNumber => $_getIZ(5);
  @$pb.TagNumber(6)
  set playerNumber($core.int value) => $_setSignedInt32(5, value);
  @$pb.TagNumber(6)
  $core.bool hasPlayerNumber() => $_has(5);
  @$pb.TagNumber(6)
  void clearPlayerNumber() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.String get playerId => $_getSZ(6);
  @$pb.TagNumber(7)
  set playerId($core.String value) => $_setString(6, value);
  @$pb.TagNumber(7)
  $core.bool hasPlayerId() => $_has(6);
  @$pb.TagNumber(7)
  void clearPlayerId() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.String get roomId => $_getSZ(7);
  @$pb.TagNumber(8)
  set roomId($core.String value) => $_setString(7, value);
  @$pb.TagNumber(8)
  $core.bool hasRoomId() => $_has(7);
  @$pb.TagNumber(8)
  void clearRoomId() => $_clearField(8);

  @$pb.TagNumber(9)
  WaitingRoom get wr => $_getN(8);
  @$pb.TagNumber(9)
  set wr(WaitingRoom value) => $_setField(9, value);
  @$pb.TagNumber(9)
  $core.bool hasWr() => $_has(8);
  @$pb.TagNumber(9)
  void clearWr() => $_clearField(9);
  @$pb.TagNumber(9)
  WaitingRoom ensureWr() => $_ensure(8);

  @$pb.TagNumber(10)
  $core.bool get ready => $_getBF(9);
  @$pb.TagNumber(10)
  set ready($core.bool value) => $_setBool(9, value);
  @$pb.TagNumber(10)
  $core.bool hasReady() => $_has(9);
  @$pb.TagNumber(10)
  void clearReady() => $_clearField(10);

  @$pb.TagNumber(11)
  MatchAllocatedNtfn get matchAlloc => $_getN(10);
  @$pb.TagNumber(11)
  set matchAlloc(MatchAllocatedNtfn value) => $_setField(11, value);
  @$pb.TagNumber(11)
  $core.bool hasMatchAlloc() => $_has(10);
  @$pb.TagNumber(11)
  void clearMatchAlloc() => $_clearField(11);
  @$pb.TagNumber(11)
  MatchAllocatedNtfn ensureMatchAlloc() => $_ensure(10);

  /// Number of confirmations for the relevant escrow deposit (if applicable)
  @$pb.TagNumber(12)
  $core.int get confs => $_getIZ(11);
  @$pb.TagNumber(12)
  set confs($core.int value) => $_setUnsignedInt32(11, value);
  @$pb.TagNumber(12)
  $core.bool hasConfs() => $_has(11);
  @$pb.TagNumber(12)
  void clearConfs() => $_clearField(12);

  /// Server-wide free-to-play flag so clients can auto-toggle escrow UI gating.
  @$pb.TagNumber(13)
  $core.bool get serverIsF2p => $_getBF(12);
  @$pb.TagNumber(13)
  set serverIsF2p($core.bool value) => $_setBool(12, value);
  @$pb.TagNumber(13)
  $core.bool hasServerIsF2p() => $_has(12);
  @$pb.TagNumber(13)
  void clearServerIsF2p() => $_clearField(13);
}

class InitConnectionRequest extends $pb.GeneratedMessage {
  factory InitConnectionRequest({
    $core.String? clientVersion,
  }) {
    final result = create();
    if (clientVersion != null) result.clientVersion = clientVersion;
    return result;
  }

  InitConnectionRequest._();

  factory InitConnectionRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory InitConnectionRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InitConnectionRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientVersion')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InitConnectionRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InitConnectionRequest copyWith(
          void Function(InitConnectionRequest) updates) =>
      super.copyWith((message) => updates(message as InitConnectionRequest))
          as InitConnectionRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static InitConnectionRequest create() => InitConnectionRequest._();
  @$core.override
  InitConnectionRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static InitConnectionRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InitConnectionRequest>(create);
  static InitConnectionRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientVersion => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientVersion($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasClientVersion() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientVersion() => $_clearField(1);
}

class InitConnectionResponse extends $pb.GeneratedMessage {
  factory InitConnectionResponse({
    $core.String? serverVersion,
    $core.bool? isF2p,
  }) {
    final result = create();
    if (serverVersion != null) result.serverVersion = serverVersion;
    if (isF2p != null) result.isF2p = isF2p;
    return result;
  }

  InitConnectionResponse._();

  factory InitConnectionResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory InitConnectionResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'InitConnectionResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'serverVersion')
    ..aOB(2, _omitFieldNames ? '' : 'isF2p')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InitConnectionResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  InitConnectionResponse copyWith(
          void Function(InitConnectionResponse) updates) =>
      super.copyWith((message) => updates(message as InitConnectionResponse))
          as InitConnectionResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static InitConnectionResponse create() => InitConnectionResponse._();
  @$core.override
  InitConnectionResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static InitConnectionResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<InitConnectionResponse>(create);
  static InitConnectionResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get serverVersion => $_getSZ(0);
  @$pb.TagNumber(1)
  set serverVersion($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasServerVersion() => $_has(0);
  @$pb.TagNumber(1)
  void clearServerVersion() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.bool get isF2p => $_getBF(1);
  @$pb.TagNumber(2)
  set isF2p($core.bool value) => $_setBool(1, value);
  @$pb.TagNumber(2)
  $core.bool hasIsF2p() => $_has(1);
  @$pb.TagNumber(2)
  void clearIsF2p() => $_clearField(2);
}

/// Waiting Room Messages
class WaitingRoomsRequest extends $pb.GeneratedMessage {
  factory WaitingRoomsRequest({
    $core.String? roomId,
  }) {
    final result = create();
    if (roomId != null) result.roomId = roomId;
    return result;
  }

  WaitingRoomsRequest._();

  factory WaitingRoomsRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory WaitingRoomsRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WaitingRoomsRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'roomId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WaitingRoomsRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WaitingRoomsRequest copyWith(void Function(WaitingRoomsRequest) updates) =>
      super.copyWith((message) => updates(message as WaitingRoomsRequest))
          as WaitingRoomsRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoomsRequest create() => WaitingRoomsRequest._();
  @$core.override
  WaitingRoomsRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static WaitingRoomsRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WaitingRoomsRequest>(create);
  static WaitingRoomsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get roomId => $_getSZ(0);
  @$pb.TagNumber(1)
  set roomId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRoomId() => $_has(0);
  @$pb.TagNumber(1)
  void clearRoomId() => $_clearField(1);
}

class WaitingRoomsResponse extends $pb.GeneratedMessage {
  factory WaitingRoomsResponse({
    $core.Iterable<WaitingRoom>? wr,
  }) {
    final result = create();
    if (wr != null) result.wr.addAll(wr);
    return result;
  }

  WaitingRoomsResponse._();

  factory WaitingRoomsResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory WaitingRoomsResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WaitingRoomsResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..pPM<WaitingRoom>(1, _omitFieldNames ? '' : 'wr',
        subBuilder: WaitingRoom.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WaitingRoomsResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WaitingRoomsResponse copyWith(void Function(WaitingRoomsResponse) updates) =>
      super.copyWith((message) => updates(message as WaitingRoomsResponse))
          as WaitingRoomsResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoomsResponse create() => WaitingRoomsResponse._();
  @$core.override
  WaitingRoomsResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static WaitingRoomsResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WaitingRoomsResponse>(create);
  static WaitingRoomsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $pb.PbList<WaitingRoom> get wr => $_getList(0);
}

class JoinWaitingRoomRequest extends $pb.GeneratedMessage {
  factory JoinWaitingRoomRequest({
    $core.String? roomId,
    $core.String? clientId,
    $core.String? escrowId,
  }) {
    final result = create();
    if (roomId != null) result.roomId = roomId;
    if (clientId != null) result.clientId = clientId;
    if (escrowId != null) result.escrowId = escrowId;
    return result;
  }

  JoinWaitingRoomRequest._();

  factory JoinWaitingRoomRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory JoinWaitingRoomRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'JoinWaitingRoomRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'roomId')
    ..aOS(2, _omitFieldNames ? '' : 'clientId')
    ..aOS(3, _omitFieldNames ? '' : 'escrowId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  JoinWaitingRoomRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  JoinWaitingRoomRequest copyWith(
          void Function(JoinWaitingRoomRequest) updates) =>
      super.copyWith((message) => updates(message as JoinWaitingRoomRequest))
          as JoinWaitingRoomRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static JoinWaitingRoomRequest create() => JoinWaitingRoomRequest._();
  @$core.override
  JoinWaitingRoomRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static JoinWaitingRoomRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<JoinWaitingRoomRequest>(create);
  static JoinWaitingRoomRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get roomId => $_getSZ(0);
  @$pb.TagNumber(1)
  set roomId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRoomId() => $_has(0);
  @$pb.TagNumber(1)
  void clearRoomId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get clientId => $_getSZ(1);
  @$pb.TagNumber(2)
  set clientId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasClientId() => $_has(1);
  @$pb.TagNumber(2)
  void clearClientId() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get escrowId => $_getSZ(2);
  @$pb.TagNumber(3)
  set escrowId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEscrowId() => $_has(2);
  @$pb.TagNumber(3)
  void clearEscrowId() => $_clearField(3);
}

class JoinWaitingRoomResponse extends $pb.GeneratedMessage {
  factory JoinWaitingRoomResponse({
    WaitingRoom? wr,
  }) {
    final result = create();
    if (wr != null) result.wr = wr;
    return result;
  }

  JoinWaitingRoomResponse._();

  factory JoinWaitingRoomResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory JoinWaitingRoomResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'JoinWaitingRoomResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOM<WaitingRoom>(1, _omitFieldNames ? '' : 'wr',
        subBuilder: WaitingRoom.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  JoinWaitingRoomResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  JoinWaitingRoomResponse copyWith(
          void Function(JoinWaitingRoomResponse) updates) =>
      super.copyWith((message) => updates(message as JoinWaitingRoomResponse))
          as JoinWaitingRoomResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static JoinWaitingRoomResponse create() => JoinWaitingRoomResponse._();
  @$core.override
  JoinWaitingRoomResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static JoinWaitingRoomResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<JoinWaitingRoomResponse>(create);
  static JoinWaitingRoomResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WaitingRoom get wr => $_getN(0);
  @$pb.TagNumber(1)
  set wr(WaitingRoom value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWr() => $_has(0);
  @$pb.TagNumber(1)
  void clearWr() => $_clearField(1);
  @$pb.TagNumber(1)
  WaitingRoom ensureWr() => $_ensure(0);
}

class CreateWaitingRoomRequest extends $pb.GeneratedMessage {
  factory CreateWaitingRoomRequest({
    $core.String? hostId,
    $fixnum.Int64? betAmt,
    $core.String? escrowId,
  }) {
    final result = create();
    if (hostId != null) result.hostId = hostId;
    if (betAmt != null) result.betAmt = betAmt;
    if (escrowId != null) result.escrowId = escrowId;
    return result;
  }

  CreateWaitingRoomRequest._();

  factory CreateWaitingRoomRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateWaitingRoomRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateWaitingRoomRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'hostId')
    ..aInt64(2, _omitFieldNames ? '' : 'betAmt', protoName: 'betAmt')
    ..aOS(3, _omitFieldNames ? '' : 'escrowId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWaitingRoomRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWaitingRoomRequest copyWith(
          void Function(CreateWaitingRoomRequest) updates) =>
      super.copyWith((message) => updates(message as CreateWaitingRoomRequest))
          as CreateWaitingRoomRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateWaitingRoomRequest create() => CreateWaitingRoomRequest._();
  @$core.override
  CreateWaitingRoomRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CreateWaitingRoomRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateWaitingRoomRequest>(create);
  static CreateWaitingRoomRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get hostId => $_getSZ(0);
  @$pb.TagNumber(1)
  set hostId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasHostId() => $_has(0);
  @$pb.TagNumber(1)
  void clearHostId() => $_clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get betAmt => $_getI64(1);
  @$pb.TagNumber(2)
  set betAmt($fixnum.Int64 value) => $_setInt64(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBetAmt() => $_has(1);
  @$pb.TagNumber(2)
  void clearBetAmt() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.String get escrowId => $_getSZ(2);
  @$pb.TagNumber(3)
  set escrowId($core.String value) => $_setString(2, value);
  @$pb.TagNumber(3)
  $core.bool hasEscrowId() => $_has(2);
  @$pb.TagNumber(3)
  void clearEscrowId() => $_clearField(3);
}

class CreateWaitingRoomResponse extends $pb.GeneratedMessage {
  factory CreateWaitingRoomResponse({
    WaitingRoom? wr,
  }) {
    final result = create();
    if (wr != null) result.wr = wr;
    return result;
  }

  CreateWaitingRoomResponse._();

  factory CreateWaitingRoomResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory CreateWaitingRoomResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'CreateWaitingRoomResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOM<WaitingRoom>(1, _omitFieldNames ? '' : 'wr',
        subBuilder: WaitingRoom.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWaitingRoomResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  CreateWaitingRoomResponse copyWith(
          void Function(CreateWaitingRoomResponse) updates) =>
      super.copyWith((message) => updates(message as CreateWaitingRoomResponse))
          as CreateWaitingRoomResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateWaitingRoomResponse create() => CreateWaitingRoomResponse._();
  @$core.override
  CreateWaitingRoomResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static CreateWaitingRoomResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<CreateWaitingRoomResponse>(create);
  static CreateWaitingRoomResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WaitingRoom get wr => $_getN(0);
  @$pb.TagNumber(1)
  set wr(WaitingRoom value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWr() => $_has(0);
  @$pb.TagNumber(1)
  void clearWr() => $_clearField(1);
  @$pb.TagNumber(1)
  WaitingRoom ensureWr() => $_ensure(0);
}

class WaitingRoom extends $pb.GeneratedMessage {
  factory WaitingRoom({
    $core.String? id,
    $core.String? hostId,
    $core.Iterable<Player>? players,
    $fixnum.Int64? betAmt,
  }) {
    final result = create();
    if (id != null) result.id = id;
    if (hostId != null) result.hostId = hostId;
    if (players != null) result.players.addAll(players);
    if (betAmt != null) result.betAmt = betAmt;
    return result;
  }

  WaitingRoom._();

  factory WaitingRoom.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory WaitingRoom.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WaitingRoom',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'hostId')
    ..pPM<Player>(3, _omitFieldNames ? '' : 'players',
        subBuilder: Player.create)
    ..aInt64(4, _omitFieldNames ? '' : 'betAmt')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WaitingRoom clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WaitingRoom copyWith(void Function(WaitingRoom) updates) =>
      super.copyWith((message) => updates(message as WaitingRoom))
          as WaitingRoom;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoom create() => WaitingRoom._();
  @$core.override
  WaitingRoom createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static WaitingRoom getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WaitingRoom>(create);
  static WaitingRoom? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get hostId => $_getSZ(1);
  @$pb.TagNumber(2)
  set hostId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasHostId() => $_has(1);
  @$pb.TagNumber(2)
  void clearHostId() => $_clearField(2);

  @$pb.TagNumber(3)
  $pb.PbList<Player> get players => $_getList(2);

  @$pb.TagNumber(4)
  $fixnum.Int64 get betAmt => $_getI64(3);
  @$pb.TagNumber(4)
  set betAmt($fixnum.Int64 value) => $_setInt64(3, value);
  @$pb.TagNumber(4)
  $core.bool hasBetAmt() => $_has(3);
  @$pb.TagNumber(4)
  void clearBetAmt() => $_clearField(4);
}

class WaitingRoomRequest extends $pb.GeneratedMessage {
  factory WaitingRoomRequest({
    $core.String? roomId,
  }) {
    final result = create();
    if (roomId != null) result.roomId = roomId;
    return result;
  }

  WaitingRoomRequest._();

  factory WaitingRoomRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory WaitingRoomRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WaitingRoomRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'roomId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WaitingRoomRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WaitingRoomRequest copyWith(void Function(WaitingRoomRequest) updates) =>
      super.copyWith((message) => updates(message as WaitingRoomRequest))
          as WaitingRoomRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoomRequest create() => WaitingRoomRequest._();
  @$core.override
  WaitingRoomRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static WaitingRoomRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WaitingRoomRequest>(create);
  static WaitingRoomRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get roomId => $_getSZ(0);
  @$pb.TagNumber(1)
  set roomId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasRoomId() => $_has(0);
  @$pb.TagNumber(1)
  void clearRoomId() => $_clearField(1);
}

class WaitingRoomResponse extends $pb.GeneratedMessage {
  factory WaitingRoomResponse({
    WaitingRoom? wr,
  }) {
    final result = create();
    if (wr != null) result.wr = wr;
    return result;
  }

  WaitingRoomResponse._();

  factory WaitingRoomResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory WaitingRoomResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'WaitingRoomResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOM<WaitingRoom>(1, _omitFieldNames ? '' : 'wr',
        subBuilder: WaitingRoom.create)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WaitingRoomResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  WaitingRoomResponse copyWith(void Function(WaitingRoomResponse) updates) =>
      super.copyWith((message) => updates(message as WaitingRoomResponse))
          as WaitingRoomResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoomResponse create() => WaitingRoomResponse._();
  @$core.override
  WaitingRoomResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static WaitingRoomResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<WaitingRoomResponse>(create);
  static WaitingRoomResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WaitingRoom get wr => $_getN(0);
  @$pb.TagNumber(1)
  set wr(WaitingRoom value) => $_setField(1, value);
  @$pb.TagNumber(1)
  $core.bool hasWr() => $_has(0);
  @$pb.TagNumber(1)
  void clearWr() => $_clearField(1);
  @$pb.TagNumber(1)
  WaitingRoom ensureWr() => $_ensure(0);
}

class Player extends $pb.GeneratedMessage {
  factory Player({
    $core.String? uid,
    $core.String? nick,
    $fixnum.Int64? betAmt,
    $core.int? number,
    $core.int? score,
    $core.bool? ready,
  }) {
    final result = create();
    if (uid != null) result.uid = uid;
    if (nick != null) result.nick = nick;
    if (betAmt != null) result.betAmt = betAmt;
    if (number != null) result.number = number;
    if (score != null) result.score = score;
    if (ready != null) result.ready = ready;
    return result;
  }

  Player._();

  factory Player.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory Player.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'Player',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..aOS(2, _omitFieldNames ? '' : 'nick')
    ..aInt64(3, _omitFieldNames ? '' : 'betAmt')
    ..aI(4, _omitFieldNames ? '' : 'number')
    ..aI(5, _omitFieldNames ? '' : 'score')
    ..aOB(6, _omitFieldNames ? '' : 'ready')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Player clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  Player copyWith(void Function(Player) updates) =>
      super.copyWith((message) => updates(message as Player)) as Player;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Player create() => Player._();
  @$core.override
  Player createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static Player getDefault() =>
      _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Player>(create);
  static Player? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get nick => $_getSZ(1);
  @$pb.TagNumber(2)
  set nick($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasNick() => $_has(1);
  @$pb.TagNumber(2)
  void clearNick() => $_clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get betAmt => $_getI64(2);
  @$pb.TagNumber(3)
  set betAmt($fixnum.Int64 value) => $_setInt64(2, value);
  @$pb.TagNumber(3)
  $core.bool hasBetAmt() => $_has(2);
  @$pb.TagNumber(3)
  void clearBetAmt() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.int get number => $_getIZ(3);
  @$pb.TagNumber(4)
  set number($core.int value) => $_setSignedInt32(3, value);
  @$pb.TagNumber(4)
  $core.bool hasNumber() => $_has(3);
  @$pb.TagNumber(4)
  void clearNumber() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.int get score => $_getIZ(4);
  @$pb.TagNumber(5)
  set score($core.int value) => $_setSignedInt32(4, value);
  @$pb.TagNumber(5)
  $core.bool hasScore() => $_has(4);
  @$pb.TagNumber(5)
  void clearScore() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.bool get ready => $_getBF(5);
  @$pb.TagNumber(6)
  set ready($core.bool value) => $_setBool(5, value);
  @$pb.TagNumber(6)
  $core.bool hasReady() => $_has(5);
  @$pb.TagNumber(6)
  void clearReady() => $_clearField(6);
}

/// SignalReadyRequest contains information about the client signaling readiness
class StartGameStreamRequest extends $pb.GeneratedMessage {
  factory StartGameStreamRequest({
    $core.String? clientId,
  }) {
    final result = create();
    if (clientId != null) result.clientId = clientId;
    return result;
  }

  StartGameStreamRequest._();

  factory StartGameStreamRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory StartGameStreamRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'StartGameStreamRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartGameStreamRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  StartGameStreamRequest copyWith(
          void Function(StartGameStreamRequest) updates) =>
      super.copyWith((message) => updates(message as StartGameStreamRequest))
          as StartGameStreamRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StartGameStreamRequest create() => StartGameStreamRequest._();
  @$core.override
  StartGameStreamRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static StartGameStreamRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<StartGameStreamRequest>(create);
  static StartGameStreamRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => $_clearField(1);
}

class GameUpdateBytes extends $pb.GeneratedMessage {
  factory GameUpdateBytes({
    $core.List<$core.int>? data,
  }) {
    final result = create();
    if (data != null) result.data = data;
    return result;
  }

  GameUpdateBytes._();

  factory GameUpdateBytes.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GameUpdateBytes.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GameUpdateBytes',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..a<$core.List<$core.int>>(
        1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.OY)
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GameUpdateBytes clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GameUpdateBytes copyWith(void Function(GameUpdateBytes) updates) =>
      super.copyWith((message) => updates(message as GameUpdateBytes))
          as GameUpdateBytes;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GameUpdateBytes create() => GameUpdateBytes._();
  @$core.override
  GameUpdateBytes createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GameUpdateBytes getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GameUpdateBytes>(create);
  static GameUpdateBytes? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get data => $_getN(0);
  @$pb.TagNumber(1)
  set data($core.List<$core.int> value) => $_setBytes(0, value);
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => $_clearField(1);
}

class PlayerInput extends $pb.GeneratedMessage {
  factory PlayerInput({
    $core.String? playerId,
    $core.String? input,
    $core.int? playerNumber,
  }) {
    final result = create();
    if (playerId != null) result.playerId = playerId;
    if (input != null) result.input = input;
    if (playerNumber != null) result.playerNumber = playerNumber;
    return result;
  }

  PlayerInput._();

  factory PlayerInput.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory PlayerInput.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'PlayerInput',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'playerId')
    ..aOS(2, _omitFieldNames ? '' : 'input')
    ..aI(3, _omitFieldNames ? '' : 'playerNumber')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PlayerInput clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  PlayerInput copyWith(void Function(PlayerInput) updates) =>
      super.copyWith((message) => updates(message as PlayerInput))
          as PlayerInput;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PlayerInput create() => PlayerInput._();
  @$core.override
  PlayerInput createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static PlayerInput getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<PlayerInput>(create);
  static PlayerInput? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get playerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set playerId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasPlayerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlayerId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get input => $_getSZ(1);
  @$pb.TagNumber(2)
  set input($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasInput() => $_has(1);
  @$pb.TagNumber(2)
  void clearInput() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.int get playerNumber => $_getIZ(2);
  @$pb.TagNumber(3)
  set playerNumber($core.int value) => $_setSignedInt32(2, value);
  @$pb.TagNumber(3)
  $core.bool hasPlayerNumber() => $_has(2);
  @$pb.TagNumber(3)
  void clearPlayerNumber() => $_clearField(3);
}

class GameUpdate extends $pb.GeneratedMessage {
  factory GameUpdate({
    $core.double? ballX,
    $core.double? ballY,
    $core.double? p1X,
    $core.double? p1Y,
    $core.double? p2X,
    $core.double? p2Y,
    $core.double? p1YVelocity,
    $core.double? p2YVelocity,
    $core.double? ballXVelocity,
    $core.double? ballYVelocity,
    $core.double? fps,
    $core.double? tps,
    $core.double? gameWidth,
    $core.double? gameHeight,
    $core.double? p1Width,
    $core.double? p1Height,
    $core.double? p2Width,
    $core.double? p2Height,
    $core.double? ballWidth,
    $core.double? ballHeight,
    $core.int? p1Score,
    $core.int? p2Score,
    $core.String? error,
    $core.bool? debug,
  }) {
    final result = create();
    if (ballX != null) result.ballX = ballX;
    if (ballY != null) result.ballY = ballY;
    if (p1X != null) result.p1X = p1X;
    if (p1Y != null) result.p1Y = p1Y;
    if (p2X != null) result.p2X = p2X;
    if (p2Y != null) result.p2Y = p2Y;
    if (p1YVelocity != null) result.p1YVelocity = p1YVelocity;
    if (p2YVelocity != null) result.p2YVelocity = p2YVelocity;
    if (ballXVelocity != null) result.ballXVelocity = ballXVelocity;
    if (ballYVelocity != null) result.ballYVelocity = ballYVelocity;
    if (fps != null) result.fps = fps;
    if (tps != null) result.tps = tps;
    if (gameWidth != null) result.gameWidth = gameWidth;
    if (gameHeight != null) result.gameHeight = gameHeight;
    if (p1Width != null) result.p1Width = p1Width;
    if (p1Height != null) result.p1Height = p1Height;
    if (p2Width != null) result.p2Width = p2Width;
    if (p2Height != null) result.p2Height = p2Height;
    if (ballWidth != null) result.ballWidth = ballWidth;
    if (ballHeight != null) result.ballHeight = ballHeight;
    if (p1Score != null) result.p1Score = p1Score;
    if (p2Score != null) result.p2Score = p2Score;
    if (error != null) result.error = error;
    if (debug != null) result.debug = debug;
    return result;
  }

  GameUpdate._();

  factory GameUpdate.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory GameUpdate.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'GameUpdate',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aD(1, _omitFieldNames ? '' : 'ballX', protoName: 'ballX')
    ..aD(2, _omitFieldNames ? '' : 'ballY', protoName: 'ballY')
    ..aD(3, _omitFieldNames ? '' : 'p1X', protoName: 'p1X')
    ..aD(4, _omitFieldNames ? '' : 'p1Y', protoName: 'p1Y')
    ..aD(5, _omitFieldNames ? '' : 'p2X', protoName: 'p2X')
    ..aD(6, _omitFieldNames ? '' : 'p2Y', protoName: 'p2Y')
    ..aD(7, _omitFieldNames ? '' : 'p1YVelocity', protoName: 'p1YVelocity')
    ..aD(8, _omitFieldNames ? '' : 'p2YVelocity', protoName: 'p2YVelocity')
    ..aD(9, _omitFieldNames ? '' : 'ballXVelocity', protoName: 'ballXVelocity')
    ..aD(10, _omitFieldNames ? '' : 'ballYVelocity', protoName: 'ballYVelocity')
    ..aD(11, _omitFieldNames ? '' : 'fps')
    ..aD(12, _omitFieldNames ? '' : 'tps')
    ..aD(13, _omitFieldNames ? '' : 'gameWidth', protoName: 'gameWidth')
    ..aD(14, _omitFieldNames ? '' : 'gameHeight', protoName: 'gameHeight')
    ..aD(15, _omitFieldNames ? '' : 'p1Width', protoName: 'p1Width')
    ..aD(16, _omitFieldNames ? '' : 'p1Height', protoName: 'p1Height')
    ..aD(17, _omitFieldNames ? '' : 'p2Width', protoName: 'p2Width')
    ..aD(18, _omitFieldNames ? '' : 'p2Height', protoName: 'p2Height')
    ..aD(19, _omitFieldNames ? '' : 'ballWidth', protoName: 'ballWidth')
    ..aD(20, _omitFieldNames ? '' : 'ballHeight', protoName: 'ballHeight')
    ..aI(21, _omitFieldNames ? '' : 'p1Score', protoName: 'p1Score')
    ..aI(22, _omitFieldNames ? '' : 'p2Score', protoName: 'p2Score')
    ..aOS(23, _omitFieldNames ? '' : 'error')
    ..aOB(24, _omitFieldNames ? '' : 'debug')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GameUpdate clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  GameUpdate copyWith(void Function(GameUpdate) updates) =>
      super.copyWith((message) => updates(message as GameUpdate)) as GameUpdate;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GameUpdate create() => GameUpdate._();
  @$core.override
  GameUpdate createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static GameUpdate getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<GameUpdate>(create);
  static GameUpdate? _defaultInstance;

  @$pb.TagNumber(1)
  $core.double get ballX => $_getN(0);
  @$pb.TagNumber(1)
  set ballX($core.double value) => $_setDouble(0, value);
  @$pb.TagNumber(1)
  $core.bool hasBallX() => $_has(0);
  @$pb.TagNumber(1)
  void clearBallX() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.double get ballY => $_getN(1);
  @$pb.TagNumber(2)
  set ballY($core.double value) => $_setDouble(1, value);
  @$pb.TagNumber(2)
  $core.bool hasBallY() => $_has(1);
  @$pb.TagNumber(2)
  void clearBallY() => $_clearField(2);

  @$pb.TagNumber(3)
  $core.double get p1X => $_getN(2);
  @$pb.TagNumber(3)
  set p1X($core.double value) => $_setDouble(2, value);
  @$pb.TagNumber(3)
  $core.bool hasP1X() => $_has(2);
  @$pb.TagNumber(3)
  void clearP1X() => $_clearField(3);

  @$pb.TagNumber(4)
  $core.double get p1Y => $_getN(3);
  @$pb.TagNumber(4)
  set p1Y($core.double value) => $_setDouble(3, value);
  @$pb.TagNumber(4)
  $core.bool hasP1Y() => $_has(3);
  @$pb.TagNumber(4)
  void clearP1Y() => $_clearField(4);

  @$pb.TagNumber(5)
  $core.double get p2X => $_getN(4);
  @$pb.TagNumber(5)
  set p2X($core.double value) => $_setDouble(4, value);
  @$pb.TagNumber(5)
  $core.bool hasP2X() => $_has(4);
  @$pb.TagNumber(5)
  void clearP2X() => $_clearField(5);

  @$pb.TagNumber(6)
  $core.double get p2Y => $_getN(5);
  @$pb.TagNumber(6)
  set p2Y($core.double value) => $_setDouble(5, value);
  @$pb.TagNumber(6)
  $core.bool hasP2Y() => $_has(5);
  @$pb.TagNumber(6)
  void clearP2Y() => $_clearField(6);

  @$pb.TagNumber(7)
  $core.double get p1YVelocity => $_getN(6);
  @$pb.TagNumber(7)
  set p1YVelocity($core.double value) => $_setDouble(6, value);
  @$pb.TagNumber(7)
  $core.bool hasP1YVelocity() => $_has(6);
  @$pb.TagNumber(7)
  void clearP1YVelocity() => $_clearField(7);

  @$pb.TagNumber(8)
  $core.double get p2YVelocity => $_getN(7);
  @$pb.TagNumber(8)
  set p2YVelocity($core.double value) => $_setDouble(7, value);
  @$pb.TagNumber(8)
  $core.bool hasP2YVelocity() => $_has(7);
  @$pb.TagNumber(8)
  void clearP2YVelocity() => $_clearField(8);

  @$pb.TagNumber(9)
  $core.double get ballXVelocity => $_getN(8);
  @$pb.TagNumber(9)
  set ballXVelocity($core.double value) => $_setDouble(8, value);
  @$pb.TagNumber(9)
  $core.bool hasBallXVelocity() => $_has(8);
  @$pb.TagNumber(9)
  void clearBallXVelocity() => $_clearField(9);

  @$pb.TagNumber(10)
  $core.double get ballYVelocity => $_getN(9);
  @$pb.TagNumber(10)
  set ballYVelocity($core.double value) => $_setDouble(9, value);
  @$pb.TagNumber(10)
  $core.bool hasBallYVelocity() => $_has(9);
  @$pb.TagNumber(10)
  void clearBallYVelocity() => $_clearField(10);

  @$pb.TagNumber(11)
  $core.double get fps => $_getN(10);
  @$pb.TagNumber(11)
  set fps($core.double value) => $_setDouble(10, value);
  @$pb.TagNumber(11)
  $core.bool hasFps() => $_has(10);
  @$pb.TagNumber(11)
  void clearFps() => $_clearField(11);

  @$pb.TagNumber(12)
  $core.double get tps => $_getN(11);
  @$pb.TagNumber(12)
  set tps($core.double value) => $_setDouble(11, value);
  @$pb.TagNumber(12)
  $core.bool hasTps() => $_has(11);
  @$pb.TagNumber(12)
  void clearTps() => $_clearField(12);

  @$pb.TagNumber(13)
  $core.double get gameWidth => $_getN(12);
  @$pb.TagNumber(13)
  set gameWidth($core.double value) => $_setDouble(12, value);
  @$pb.TagNumber(13)
  $core.bool hasGameWidth() => $_has(12);
  @$pb.TagNumber(13)
  void clearGameWidth() => $_clearField(13);

  @$pb.TagNumber(14)
  $core.double get gameHeight => $_getN(13);
  @$pb.TagNumber(14)
  set gameHeight($core.double value) => $_setDouble(13, value);
  @$pb.TagNumber(14)
  $core.bool hasGameHeight() => $_has(13);
  @$pb.TagNumber(14)
  void clearGameHeight() => $_clearField(14);

  @$pb.TagNumber(15)
  $core.double get p1Width => $_getN(14);
  @$pb.TagNumber(15)
  set p1Width($core.double value) => $_setDouble(14, value);
  @$pb.TagNumber(15)
  $core.bool hasP1Width() => $_has(14);
  @$pb.TagNumber(15)
  void clearP1Width() => $_clearField(15);

  @$pb.TagNumber(16)
  $core.double get p1Height => $_getN(15);
  @$pb.TagNumber(16)
  set p1Height($core.double value) => $_setDouble(15, value);
  @$pb.TagNumber(16)
  $core.bool hasP1Height() => $_has(15);
  @$pb.TagNumber(16)
  void clearP1Height() => $_clearField(16);

  @$pb.TagNumber(17)
  $core.double get p2Width => $_getN(16);
  @$pb.TagNumber(17)
  set p2Width($core.double value) => $_setDouble(16, value);
  @$pb.TagNumber(17)
  $core.bool hasP2Width() => $_has(16);
  @$pb.TagNumber(17)
  void clearP2Width() => $_clearField(17);

  @$pb.TagNumber(18)
  $core.double get p2Height => $_getN(17);
  @$pb.TagNumber(18)
  set p2Height($core.double value) => $_setDouble(17, value);
  @$pb.TagNumber(18)
  $core.bool hasP2Height() => $_has(17);
  @$pb.TagNumber(18)
  void clearP2Height() => $_clearField(18);

  @$pb.TagNumber(19)
  $core.double get ballWidth => $_getN(18);
  @$pb.TagNumber(19)
  set ballWidth($core.double value) => $_setDouble(18, value);
  @$pb.TagNumber(19)
  $core.bool hasBallWidth() => $_has(18);
  @$pb.TagNumber(19)
  void clearBallWidth() => $_clearField(19);

  @$pb.TagNumber(20)
  $core.double get ballHeight => $_getN(19);
  @$pb.TagNumber(20)
  set ballHeight($core.double value) => $_setDouble(19, value);
  @$pb.TagNumber(20)
  $core.bool hasBallHeight() => $_has(19);
  @$pb.TagNumber(20)
  void clearBallHeight() => $_clearField(20);

  @$pb.TagNumber(21)
  $core.int get p1Score => $_getIZ(20);
  @$pb.TagNumber(21)
  set p1Score($core.int value) => $_setSignedInt32(20, value);
  @$pb.TagNumber(21)
  $core.bool hasP1Score() => $_has(20);
  @$pb.TagNumber(21)
  void clearP1Score() => $_clearField(21);

  @$pb.TagNumber(22)
  $core.int get p2Score => $_getIZ(21);
  @$pb.TagNumber(22)
  set p2Score($core.int value) => $_setSignedInt32(21, value);
  @$pb.TagNumber(22)
  $core.bool hasP2Score() => $_has(21);
  @$pb.TagNumber(22)
  void clearP2Score() => $_clearField(22);

  /// Optional: if you want to send error messages or debug information
  @$pb.TagNumber(23)
  $core.String get error => $_getSZ(22);
  @$pb.TagNumber(23)
  set error($core.String value) => $_setString(22, value);
  @$pb.TagNumber(23)
  $core.bool hasError() => $_has(22);
  @$pb.TagNumber(23)
  void clearError() => $_clearField(23);

  @$pb.TagNumber(24)
  $core.bool get debug => $_getBF(23);
  @$pb.TagNumber(24)
  set debug($core.bool value) => $_setBool(23, value);
  @$pb.TagNumber(24)
  $core.bool hasDebug() => $_has(23);
  @$pb.TagNumber(24)
  void clearDebug() => $_clearField(24);
}

class LeaveWaitingRoomRequest extends $pb.GeneratedMessage {
  factory LeaveWaitingRoomRequest({
    $core.String? clientId,
    $core.String? roomId,
  }) {
    final result = create();
    if (clientId != null) result.clientId = clientId;
    if (roomId != null) result.roomId = roomId;
    return result;
  }

  LeaveWaitingRoomRequest._();

  factory LeaveWaitingRoomRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LeaveWaitingRoomRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LeaveWaitingRoomRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..aOS(2, _omitFieldNames ? '' : 'roomId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LeaveWaitingRoomRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LeaveWaitingRoomRequest copyWith(
          void Function(LeaveWaitingRoomRequest) updates) =>
      super.copyWith((message) => updates(message as LeaveWaitingRoomRequest))
          as LeaveWaitingRoomRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LeaveWaitingRoomRequest create() => LeaveWaitingRoomRequest._();
  @$core.override
  LeaveWaitingRoomRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static LeaveWaitingRoomRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LeaveWaitingRoomRequest>(create);
  static LeaveWaitingRoomRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get roomId => $_getSZ(1);
  @$pb.TagNumber(2)
  set roomId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasRoomId() => $_has(1);
  @$pb.TagNumber(2)
  void clearRoomId() => $_clearField(2);
}

class LeaveWaitingRoomResponse extends $pb.GeneratedMessage {
  factory LeaveWaitingRoomResponse({
    $core.bool? success,
    $core.String? message,
  }) {
    final result = create();
    if (success != null) result.success = success;
    if (message != null) result.message = message;
    return result;
  }

  LeaveWaitingRoomResponse._();

  factory LeaveWaitingRoomResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory LeaveWaitingRoomResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'LeaveWaitingRoomResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'success')
    ..aOS(2, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LeaveWaitingRoomResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  LeaveWaitingRoomResponse copyWith(
          void Function(LeaveWaitingRoomResponse) updates) =>
      super.copyWith((message) => updates(message as LeaveWaitingRoomResponse))
          as LeaveWaitingRoomResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LeaveWaitingRoomResponse create() => LeaveWaitingRoomResponse._();
  @$core.override
  LeaveWaitingRoomResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static LeaveWaitingRoomResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<LeaveWaitingRoomResponse>(create);
  static LeaveWaitingRoomResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get success => $_getBF(0);
  @$pb.TagNumber(1)
  set success($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSuccess() => $_has(0);
  @$pb.TagNumber(1)
  void clearSuccess() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get message => $_getSZ(1);
  @$pb.TagNumber(2)
  set message($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMessage() => $_has(1);
  @$pb.TagNumber(2)
  void clearMessage() => $_clearField(2);
}

/// SignalReadyToPlayRequest contains information about the client signaling readiness
class SignalReadyToPlayRequest extends $pb.GeneratedMessage {
  factory SignalReadyToPlayRequest({
    $core.String? clientId,
    $core.String? gameId,
  }) {
    final result = create();
    if (clientId != null) result.clientId = clientId;
    if (gameId != null) result.gameId = gameId;
    return result;
  }

  SignalReadyToPlayRequest._();

  factory SignalReadyToPlayRequest.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SignalReadyToPlayRequest.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SignalReadyToPlayRequest',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..aOS(2, _omitFieldNames ? '' : 'gameId')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SignalReadyToPlayRequest clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SignalReadyToPlayRequest copyWith(
          void Function(SignalReadyToPlayRequest) updates) =>
      super.copyWith((message) => updates(message as SignalReadyToPlayRequest))
          as SignalReadyToPlayRequest;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SignalReadyToPlayRequest create() => SignalReadyToPlayRequest._();
  @$core.override
  SignalReadyToPlayRequest createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static SignalReadyToPlayRequest getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SignalReadyToPlayRequest>(create);
  static SignalReadyToPlayRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String value) => $_setString(0, value);
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get gameId => $_getSZ(1);
  @$pb.TagNumber(2)
  set gameId($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasGameId() => $_has(1);
  @$pb.TagNumber(2)
  void clearGameId() => $_clearField(2);
}

/// SignalReadyToPlayResponse contains the result of the ready signal
class SignalReadyToPlayResponse extends $pb.GeneratedMessage {
  factory SignalReadyToPlayResponse({
    $core.bool? success,
    $core.String? message,
  }) {
    final result = create();
    if (success != null) result.success = success;
    if (message != null) result.message = message;
    return result;
  }

  SignalReadyToPlayResponse._();

  factory SignalReadyToPlayResponse.fromBuffer($core.List<$core.int> data,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromBuffer(data, registry);
  factory SignalReadyToPlayResponse.fromJson($core.String json,
          [$pb.ExtensionRegistry registry = $pb.ExtensionRegistry.EMPTY]) =>
      create()..mergeFromJson(json, registry);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(
      _omitMessageNames ? '' : 'SignalReadyToPlayResponse',
      package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'),
      createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'success')
    ..aOS(2, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false;

  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SignalReadyToPlayResponse clone() => deepCopy();
  @$core.Deprecated('See https://github.com/google/protobuf.dart/issues/998.')
  SignalReadyToPlayResponse copyWith(
          void Function(SignalReadyToPlayResponse) updates) =>
      super.copyWith((message) => updates(message as SignalReadyToPlayResponse))
          as SignalReadyToPlayResponse;

  @$core.override
  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SignalReadyToPlayResponse create() => SignalReadyToPlayResponse._();
  @$core.override
  SignalReadyToPlayResponse createEmptyInstance() => create();
  @$core.pragma('dart2js:noInline')
  static SignalReadyToPlayResponse getDefault() => _defaultInstance ??=
      $pb.GeneratedMessage.$_defaultFor<SignalReadyToPlayResponse>(create);
  static SignalReadyToPlayResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get success => $_getBF(0);
  @$pb.TagNumber(1)
  set success($core.bool value) => $_setBool(0, value);
  @$pb.TagNumber(1)
  $core.bool hasSuccess() => $_has(0);
  @$pb.TagNumber(1)
  void clearSuccess() => $_clearField(1);

  @$pb.TagNumber(2)
  $core.String get message => $_getSZ(1);
  @$pb.TagNumber(2)
  set message($core.String value) => $_setString(1, value);
  @$pb.TagNumber(2)
  $core.bool hasMessage() => $_has(1);
  @$pb.TagNumber(2)
  void clearMessage() => $_clearField(2);
}

const $core.bool _omitFieldNames =
    $core.bool.fromEnvironment('protobuf.omit_field_names');
const $core.bool _omitMessageNames =
    $core.bool.fromEnvironment('protobuf.omit_message_names');
