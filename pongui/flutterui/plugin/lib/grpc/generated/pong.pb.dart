//
//  Generated code. Do not modify.
//  source: pong.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:core' as $core;

import 'package:fixnum/fixnum.dart' as $fixnum;
import 'package:protobuf/protobuf.dart' as $pb;

import 'pong.pbenum.dart';

export 'pong.pbenum.dart';

enum ClientMsg_Kind {
  hello, 
  presigs, 
  ack, 
  notSet
}

/// === Phase 1 streaming messages ===
class ClientMsg extends $pb.GeneratedMessage {
  factory ClientMsg({
    $core.String? matchId,
    Hello? hello,
    PreSigBatch? presigs,
    Ack? ack,
  }) {
    final $result = create();
    if (matchId != null) {
      $result.matchId = matchId;
    }
    if (hello != null) {
      $result.hello = hello;
    }
    if (presigs != null) {
      $result.presigs = presigs;
    }
    if (ack != null) {
      $result.ack = ack;
    }
    return $result;
  }
  ClientMsg._() : super();
  factory ClientMsg.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ClientMsg.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static const $core.Map<$core.int, ClientMsg_Kind> _ClientMsg_KindByTag = {
    10 : ClientMsg_Kind.hello,
    11 : ClientMsg_Kind.presigs,
    12 : ClientMsg_Kind.ack,
    0 : ClientMsg_Kind.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ClientMsg', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..oo(0, [10, 11, 12])
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..aOM<Hello>(10, _omitFieldNames ? '' : 'hello', subBuilder: Hello.create)
    ..aOM<PreSigBatch>(11, _omitFieldNames ? '' : 'presigs', subBuilder: PreSigBatch.create)
    ..aOM<Ack>(12, _omitFieldNames ? '' : 'ack', subBuilder: Ack.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ClientMsg clone() => ClientMsg()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ClientMsg copyWith(void Function(ClientMsg) updates) => super.copyWith((message) => updates(message as ClientMsg)) as ClientMsg;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ClientMsg create() => ClientMsg._();
  ClientMsg createEmptyInstance() => create();
  static $pb.PbList<ClientMsg> createRepeated() => $pb.PbList<ClientMsg>();
  @$core.pragma('dart2js:noInline')
  static ClientMsg getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ClientMsg>(create);
  static ClientMsg? _defaultInstance;

  ClientMsg_Kind whichKind() => _ClientMsg_KindByTag[$_whichOneof(0)]!;
  void clearKind() => clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => clearField(1);

  @$pb.TagNumber(10)
  Hello get hello => $_getN(1);
  @$pb.TagNumber(10)
  set hello(Hello v) { setField(10, v); }
  @$pb.TagNumber(10)
  $core.bool hasHello() => $_has(1);
  @$pb.TagNumber(10)
  void clearHello() => clearField(10);
  @$pb.TagNumber(10)
  Hello ensureHello() => $_ensure(1);

  @$pb.TagNumber(11)
  PreSigBatch get presigs => $_getN(2);
  @$pb.TagNumber(11)
  set presigs(PreSigBatch v) { setField(11, v); }
  @$pb.TagNumber(11)
  $core.bool hasPresigs() => $_has(2);
  @$pb.TagNumber(11)
  void clearPresigs() => clearField(11);
  @$pb.TagNumber(11)
  PreSigBatch ensurePresigs() => $_ensure(2);

  @$pb.TagNumber(12)
  Ack get ack => $_getN(3);
  @$pb.TagNumber(12)
  set ack(Ack v) { setField(12, v); }
  @$pb.TagNumber(12)
  $core.bool hasAck() => $_has(3);
  @$pb.TagNumber(12)
  void clearAck() => clearField(12);
  @$pb.TagNumber(12)
  Ack ensureAck() => $_ensure(3);
}

enum ServerMsg_Kind {
  role, 
  req, 
  reveal, 
  info, 
  notSet
}

class ServerMsg extends $pb.GeneratedMessage {
  factory ServerMsg({
    $core.String? matchId,
    AssignRole? role,
    NeedPreSigs? req,
    RevealGamma? reveal,
    Info? info,
  }) {
    final $result = create();
    if (matchId != null) {
      $result.matchId = matchId;
    }
    if (role != null) {
      $result.role = role;
    }
    if (req != null) {
      $result.req = req;
    }
    if (reveal != null) {
      $result.reveal = reveal;
    }
    if (info != null) {
      $result.info = info;
    }
    return $result;
  }
  ServerMsg._() : super();
  factory ServerMsg.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory ServerMsg.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static const $core.Map<$core.int, ServerMsg_Kind> _ServerMsg_KindByTag = {
    10 : ServerMsg_Kind.role,
    11 : ServerMsg_Kind.req,
    12 : ServerMsg_Kind.reveal,
    13 : ServerMsg_Kind.info,
    0 : ServerMsg_Kind.notSet
  };
  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'ServerMsg', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..oo(0, [10, 11, 12, 13])
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..aOM<AssignRole>(10, _omitFieldNames ? '' : 'role', subBuilder: AssignRole.create)
    ..aOM<NeedPreSigs>(11, _omitFieldNames ? '' : 'req', subBuilder: NeedPreSigs.create)
    ..aOM<RevealGamma>(12, _omitFieldNames ? '' : 'reveal', subBuilder: RevealGamma.create)
    ..aOM<Info>(13, _omitFieldNames ? '' : 'info', subBuilder: Info.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  ServerMsg clone() => ServerMsg()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  ServerMsg copyWith(void Function(ServerMsg) updates) => super.copyWith((message) => updates(message as ServerMsg)) as ServerMsg;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static ServerMsg create() => ServerMsg._();
  ServerMsg createEmptyInstance() => create();
  static $pb.PbList<ServerMsg> createRepeated() => $pb.PbList<ServerMsg>();
  @$core.pragma('dart2js:noInline')
  static ServerMsg getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<ServerMsg>(create);
  static ServerMsg? _defaultInstance;

  ServerMsg_Kind whichKind() => _ServerMsg_KindByTag[$_whichOneof(0)]!;
  void clearKind() => clearField($_whichOneof(0));

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => clearField(1);

  @$pb.TagNumber(10)
  AssignRole get role => $_getN(1);
  @$pb.TagNumber(10)
  set role(AssignRole v) { setField(10, v); }
  @$pb.TagNumber(10)
  $core.bool hasRole() => $_has(1);
  @$pb.TagNumber(10)
  void clearRole() => clearField(10);
  @$pb.TagNumber(10)
  AssignRole ensureRole() => $_ensure(1);

  @$pb.TagNumber(11)
  NeedPreSigs get req => $_getN(2);
  @$pb.TagNumber(11)
  set req(NeedPreSigs v) { setField(11, v); }
  @$pb.TagNumber(11)
  $core.bool hasReq() => $_has(2);
  @$pb.TagNumber(11)
  void clearReq() => clearField(11);
  @$pb.TagNumber(11)
  NeedPreSigs ensureReq() => $_ensure(2);

  @$pb.TagNumber(12)
  RevealGamma get reveal => $_getN(3);
  @$pb.TagNumber(12)
  set reveal(RevealGamma v) { setField(12, v); }
  @$pb.TagNumber(12)
  $core.bool hasReveal() => $_has(3);
  @$pb.TagNumber(12)
  void clearReveal() => clearField(12);
  @$pb.TagNumber(12)
  RevealGamma ensureReveal() => $_ensure(3);

  @$pb.TagNumber(13)
  Info get info => $_getN(4);
  @$pb.TagNumber(13)
  set info(Info v) { setField(13, v); }
  @$pb.TagNumber(13)
  $core.bool hasInfo() => $_has(4);
  @$pb.TagNumber(13)
  void clearInfo() => clearField(13);
  @$pb.TagNumber(13)
  Info ensureInfo() => $_ensure(4);
}

class Hello extends $pb.GeneratedMessage {
  factory Hello({
    $core.List<$core.int>? compPubkey,
    $core.String? clientVersion,
  }) {
    final $result = create();
    if (compPubkey != null) {
      $result.compPubkey = compPubkey;
    }
    if (clientVersion != null) {
      $result.clientVersion = clientVersion;
    }
    return $result;
  }
  Hello._() : super();
  factory Hello.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Hello.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Hello', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, _omitFieldNames ? '' : 'compPubkey', $pb.PbFieldType.OY)
    ..aOS(2, _omitFieldNames ? '' : 'clientVersion')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Hello clone() => Hello()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Hello copyWith(void Function(Hello) updates) => super.copyWith((message) => updates(message as Hello)) as Hello;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Hello create() => Hello._();
  Hello createEmptyInstance() => create();
  static $pb.PbList<Hello> createRepeated() => $pb.PbList<Hello>();
  @$core.pragma('dart2js:noInline')
  static Hello getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Hello>(create);
  static Hello? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get compPubkey => $_getN(0);
  @$pb.TagNumber(1)
  set compPubkey($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasCompPubkey() => $_has(0);
  @$pb.TagNumber(1)
  void clearCompPubkey() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get clientVersion => $_getSZ(1);
  @$pb.TagNumber(2)
  set clientVersion($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasClientVersion() => $_has(1);
  @$pb.TagNumber(2)
  void clearClientVersion() => clearField(2);
}

class AssignRole extends $pb.GeneratedMessage {
  factory AssignRole({
    AssignRole_Role? role,
    $fixnum.Int64? requiredAtoms,
    $core.String? depositPkscriptHex,
  }) {
    final $result = create();
    if (role != null) {
      $result.role = role;
    }
    if (requiredAtoms != null) {
      $result.requiredAtoms = requiredAtoms;
    }
    if (depositPkscriptHex != null) {
      $result.depositPkscriptHex = depositPkscriptHex;
    }
    return $result;
  }
  AssignRole._() : super();
  factory AssignRole.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AssignRole.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'AssignRole', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..e<AssignRole_Role>(1, _omitFieldNames ? '' : 'role', $pb.PbFieldType.OE, defaultOrMaker: AssignRole_Role.UNKNOWN, valueOf: AssignRole_Role.valueOf, enumValues: AssignRole_Role.values)
    ..a<$fixnum.Int64>(2, _omitFieldNames ? '' : 'requiredAtoms', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(3, _omitFieldNames ? '' : 'depositPkscriptHex')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AssignRole clone() => AssignRole()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AssignRole copyWith(void Function(AssignRole) updates) => super.copyWith((message) => updates(message as AssignRole)) as AssignRole;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AssignRole create() => AssignRole._();
  AssignRole createEmptyInstance() => create();
  static $pb.PbList<AssignRole> createRepeated() => $pb.PbList<AssignRole>();
  @$core.pragma('dart2js:noInline')
  static AssignRole getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AssignRole>(create);
  static AssignRole? _defaultInstance;

  @$pb.TagNumber(1)
  AssignRole_Role get role => $_getN(0);
  @$pb.TagNumber(1)
  set role(AssignRole_Role v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasRole() => $_has(0);
  @$pb.TagNumber(1)
  void clearRole() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get requiredAtoms => $_getI64(1);
  @$pb.TagNumber(2)
  set requiredAtoms($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasRequiredAtoms() => $_has(1);
  @$pb.TagNumber(2)
  void clearRequiredAtoms() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get depositPkscriptHex => $_getSZ(2);
  @$pb.TagNumber(3)
  set depositPkscriptHex($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasDepositPkscriptHex() => $_has(2);
  @$pb.TagNumber(3)
  void clearDepositPkscriptHex() => clearField(3);
}

class NeedPreSigs_PerInput extends $pb.GeneratedMessage {
  factory NeedPreSigs_PerInput({
    $core.String? inputId,
    $core.String? redeemScriptHex,
    $core.String? mAwinsHex,
    $core.List<$core.int>? tAwinsCompressed,
    $core.String? mBwinsHex,
    $core.List<$core.int>? tBwinsCompressed,
  }) {
    final $result = create();
    if (inputId != null) {
      $result.inputId = inputId;
    }
    if (redeemScriptHex != null) {
      $result.redeemScriptHex = redeemScriptHex;
    }
    if (mAwinsHex != null) {
      $result.mAwinsHex = mAwinsHex;
    }
    if (tAwinsCompressed != null) {
      $result.tAwinsCompressed = tAwinsCompressed;
    }
    if (mBwinsHex != null) {
      $result.mBwinsHex = mBwinsHex;
    }
    if (tBwinsCompressed != null) {
      $result.tBwinsCompressed = tBwinsCompressed;
    }
    return $result;
  }
  NeedPreSigs_PerInput._() : super();
  factory NeedPreSigs_PerInput.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory NeedPreSigs_PerInput.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'NeedPreSigs.PerInput', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'inputId')
    ..aOS(2, _omitFieldNames ? '' : 'redeemScriptHex')
    ..aOS(3, _omitFieldNames ? '' : 'mAwinsHex')
    ..a<$core.List<$core.int>>(4, _omitFieldNames ? '' : 'TAwinsCompressed', $pb.PbFieldType.OY, protoName: 'T_awins_compressed')
    ..aOS(5, _omitFieldNames ? '' : 'mBwinsHex')
    ..a<$core.List<$core.int>>(6, _omitFieldNames ? '' : 'TBwinsCompressed', $pb.PbFieldType.OY, protoName: 'T_bwins_compressed')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  NeedPreSigs_PerInput clone() => NeedPreSigs_PerInput()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  NeedPreSigs_PerInput copyWith(void Function(NeedPreSigs_PerInput) updates) => super.copyWith((message) => updates(message as NeedPreSigs_PerInput)) as NeedPreSigs_PerInput;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static NeedPreSigs_PerInput create() => NeedPreSigs_PerInput._();
  NeedPreSigs_PerInput createEmptyInstance() => create();
  static $pb.PbList<NeedPreSigs_PerInput> createRepeated() => $pb.PbList<NeedPreSigs_PerInput>();
  @$core.pragma('dart2js:noInline')
  static NeedPreSigs_PerInput getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<NeedPreSigs_PerInput>(create);
  static NeedPreSigs_PerInput? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get inputId => $_getSZ(0);
  @$pb.TagNumber(1)
  set inputId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasInputId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInputId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get redeemScriptHex => $_getSZ(1);
  @$pb.TagNumber(2)
  set redeemScriptHex($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasRedeemScriptHex() => $_has(1);
  @$pb.TagNumber(2)
  void clearRedeemScriptHex() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get mAwinsHex => $_getSZ(2);
  @$pb.TagNumber(3)
  set mAwinsHex($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasMAwinsHex() => $_has(2);
  @$pb.TagNumber(3)
  void clearMAwinsHex() => clearField(3);

  @$pb.TagNumber(4)
  $core.List<$core.int> get tAwinsCompressed => $_getN(3);
  @$pb.TagNumber(4)
  set tAwinsCompressed($core.List<$core.int> v) { $_setBytes(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasTAwinsCompressed() => $_has(3);
  @$pb.TagNumber(4)
  void clearTAwinsCompressed() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get mBwinsHex => $_getSZ(4);
  @$pb.TagNumber(5)
  set mBwinsHex($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasMBwinsHex() => $_has(4);
  @$pb.TagNumber(5)
  void clearMBwinsHex() => clearField(5);

  @$pb.TagNumber(6)
  $core.List<$core.int> get tBwinsCompressed => $_getN(5);
  @$pb.TagNumber(6)
  set tBwinsCompressed($core.List<$core.int> v) { $_setBytes(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasTBwinsCompressed() => $_has(5);
  @$pb.TagNumber(6)
  void clearTBwinsCompressed() => clearField(6);
}

class NeedPreSigs extends $pb.GeneratedMessage {
  factory NeedPreSigs({
    $core.Iterable<Branch>? branchesToPresign,
    $core.String? draftAwinsTxHex,
    $core.String? draftBwinsTxHex,
    $core.Iterable<NeedPreSigs_PerInput>? inputs,
  }) {
    final $result = create();
    if (branchesToPresign != null) {
      $result.branchesToPresign.addAll(branchesToPresign);
    }
    if (draftAwinsTxHex != null) {
      $result.draftAwinsTxHex = draftAwinsTxHex;
    }
    if (draftBwinsTxHex != null) {
      $result.draftBwinsTxHex = draftBwinsTxHex;
    }
    if (inputs != null) {
      $result.inputs.addAll(inputs);
    }
    return $result;
  }
  NeedPreSigs._() : super();
  factory NeedPreSigs.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory NeedPreSigs.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'NeedPreSigs', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..pc<Branch>(1, _omitFieldNames ? '' : 'branchesToPresign', $pb.PbFieldType.KE, valueOf: Branch.valueOf, enumValues: Branch.values, defaultEnumValue: Branch.BRANCH_UNSPECIFIED)
    ..aOS(2, _omitFieldNames ? '' : 'draftAwinsTxHex')
    ..aOS(3, _omitFieldNames ? '' : 'draftBwinsTxHex')
    ..pc<NeedPreSigs_PerInput>(4, _omitFieldNames ? '' : 'inputs', $pb.PbFieldType.PM, subBuilder: NeedPreSigs_PerInput.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  NeedPreSigs clone() => NeedPreSigs()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  NeedPreSigs copyWith(void Function(NeedPreSigs) updates) => super.copyWith((message) => updates(message as NeedPreSigs)) as NeedPreSigs;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static NeedPreSigs create() => NeedPreSigs._();
  NeedPreSigs createEmptyInstance() => create();
  static $pb.PbList<NeedPreSigs> createRepeated() => $pb.PbList<NeedPreSigs>();
  @$core.pragma('dart2js:noInline')
  static NeedPreSigs getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<NeedPreSigs>(create);
  static NeedPreSigs? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Branch> get branchesToPresign => $_getList(0);

  @$pb.TagNumber(2)
  $core.String get draftAwinsTxHex => $_getSZ(1);
  @$pb.TagNumber(2)
  set draftAwinsTxHex($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasDraftAwinsTxHex() => $_has(1);
  @$pb.TagNumber(2)
  void clearDraftAwinsTxHex() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get draftBwinsTxHex => $_getSZ(2);
  @$pb.TagNumber(3)
  set draftBwinsTxHex($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasDraftBwinsTxHex() => $_has(2);
  @$pb.TagNumber(3)
  void clearDraftBwinsTxHex() => clearField(3);

  @$pb.TagNumber(4)
  $core.List<NeedPreSigs_PerInput> get inputs => $_getList(3);
}

class PreSigBatch_Sig extends $pb.GeneratedMessage {
  factory PreSigBatch_Sig({
    $core.String? inputId,
    $core.List<$core.int>? rprime32,
    $core.List<$core.int>? sprime32,
  }) {
    final $result = create();
    if (inputId != null) {
      $result.inputId = inputId;
    }
    if (rprime32 != null) {
      $result.rprime32 = rprime32;
    }
    if (sprime32 != null) {
      $result.sprime32 = sprime32;
    }
    return $result;
  }
  PreSigBatch_Sig._() : super();
  factory PreSigBatch_Sig.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PreSigBatch_Sig.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'PreSigBatch.Sig', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'inputId')
    ..a<$core.List<$core.int>>(2, _omitFieldNames ? '' : 'rprime32', $pb.PbFieldType.OY)
    ..a<$core.List<$core.int>>(3, _omitFieldNames ? '' : 'sprime32', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PreSigBatch_Sig clone() => PreSigBatch_Sig()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PreSigBatch_Sig copyWith(void Function(PreSigBatch_Sig) updates) => super.copyWith((message) => updates(message as PreSigBatch_Sig)) as PreSigBatch_Sig;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PreSigBatch_Sig create() => PreSigBatch_Sig._();
  PreSigBatch_Sig createEmptyInstance() => create();
  static $pb.PbList<PreSigBatch_Sig> createRepeated() => $pb.PbList<PreSigBatch_Sig>();
  @$core.pragma('dart2js:noInline')
  static PreSigBatch_Sig getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PreSigBatch_Sig>(create);
  static PreSigBatch_Sig? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get inputId => $_getSZ(0);
  @$pb.TagNumber(1)
  set inputId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasInputId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInputId() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<$core.int> get rprime32 => $_getN(1);
  @$pb.TagNumber(2)
  set rprime32($core.List<$core.int> v) { $_setBytes(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasRprime32() => $_has(1);
  @$pb.TagNumber(2)
  void clearRprime32() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<$core.int> get sprime32 => $_getN(2);
  @$pb.TagNumber(3)
  set sprime32($core.List<$core.int> v) { $_setBytes(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasSprime32() => $_has(2);
  @$pb.TagNumber(3)
  void clearSprime32() => clearField(3);
}

class PreSigBatch extends $pb.GeneratedMessage {
  factory PreSigBatch({
    Branch? branch,
    $core.Iterable<PreSigBatch_Sig>? presigs,
  }) {
    final $result = create();
    if (branch != null) {
      $result.branch = branch;
    }
    if (presigs != null) {
      $result.presigs.addAll(presigs);
    }
    return $result;
  }
  PreSigBatch._() : super();
  factory PreSigBatch.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PreSigBatch.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'PreSigBatch', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..e<Branch>(1, _omitFieldNames ? '' : 'branch', $pb.PbFieldType.OE, defaultOrMaker: Branch.BRANCH_UNSPECIFIED, valueOf: Branch.valueOf, enumValues: Branch.values)
    ..pc<PreSigBatch_Sig>(2, _omitFieldNames ? '' : 'presigs', $pb.PbFieldType.PM, subBuilder: PreSigBatch_Sig.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PreSigBatch clone() => PreSigBatch()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PreSigBatch copyWith(void Function(PreSigBatch) updates) => super.copyWith((message) => updates(message as PreSigBatch)) as PreSigBatch;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PreSigBatch create() => PreSigBatch._();
  PreSigBatch createEmptyInstance() => create();
  static $pb.PbList<PreSigBatch> createRepeated() => $pb.PbList<PreSigBatch>();
  @$core.pragma('dart2js:noInline')
  static PreSigBatch getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PreSigBatch>(create);
  static PreSigBatch? _defaultInstance;

  @$pb.TagNumber(1)
  Branch get branch => $_getN(0);
  @$pb.TagNumber(1)
  set branch(Branch v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasBranch() => $_has(0);
  @$pb.TagNumber(1)
  void clearBranch() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<PreSigBatch_Sig> get presigs => $_getList(1);
}

class RevealGamma extends $pb.GeneratedMessage {
  factory RevealGamma({
    $core.List<$core.int>? gamma32,
  }) {
    final $result = create();
    if (gamma32 != null) {
      $result.gamma32 = gamma32;
    }
    return $result;
  }
  RevealGamma._() : super();
  factory RevealGamma.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RevealGamma.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'RevealGamma', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, _omitFieldNames ? '' : 'gamma32', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RevealGamma clone() => RevealGamma()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RevealGamma copyWith(void Function(RevealGamma) updates) => super.copyWith((message) => updates(message as RevealGamma)) as RevealGamma;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RevealGamma create() => RevealGamma._();
  RevealGamma createEmptyInstance() => create();
  static $pb.PbList<RevealGamma> createRepeated() => $pb.PbList<RevealGamma>();
  @$core.pragma('dart2js:noInline')
  static RevealGamma getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RevealGamma>(create);
  static RevealGamma? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get gamma32 => $_getN(0);
  @$pb.TagNumber(1)
  set gamma32($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasGamma32() => $_has(0);
  @$pb.TagNumber(1)
  void clearGamma32() => clearField(1);
}

class Ack extends $pb.GeneratedMessage {
  factory Ack({
    $core.String? note,
  }) {
    final $result = create();
    if (note != null) {
      $result.note = note;
    }
    return $result;
  }
  Ack._() : super();
  factory Ack.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Ack.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Ack', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'note')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Ack clone() => Ack()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Ack copyWith(void Function(Ack) updates) => super.copyWith((message) => updates(message as Ack)) as Ack;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Ack create() => Ack._();
  Ack createEmptyInstance() => create();
  static $pb.PbList<Ack> createRepeated() => $pb.PbList<Ack>();
  @$core.pragma('dart2js:noInline')
  static Ack getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Ack>(create);
  static Ack? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get note => $_getSZ(0);
  @$pb.TagNumber(1)
  set note($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasNote() => $_has(0);
  @$pb.TagNumber(1)
  void clearNote() => clearField(1);
}

class Info extends $pb.GeneratedMessage {
  factory Info({
    $core.String? text,
  }) {
    final $result = create();
    if (text != null) {
      $result.text = text;
    }
    return $result;
  }
  Info._() : super();
  factory Info.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Info.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Info', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'text')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Info clone() => Info()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Info copyWith(void Function(Info) updates) => super.copyWith((message) => updates(message as Info)) as Info;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Info create() => Info._();
  Info createEmptyInstance() => create();
  static $pb.PbList<Info> createRepeated() => $pb.PbList<Info>();
  @$core.pragma('dart2js:noInline')
  static Info getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Info>(create);
  static Info? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get text => $_getSZ(0);
  @$pb.TagNumber(1)
  set text($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasText() => $_has(0);
  @$pb.TagNumber(1)
  void clearText() => clearField(1);
}

class CreateMatchRequest extends $pb.GeneratedMessage {
  factory CreateMatchRequest({
    $core.String? aC,
    $core.String? bC,
    $core.int? csv,
  }) {
    final $result = create();
    if (aC != null) {
      $result.aC = aC;
    }
    if (bC != null) {
      $result.bC = bC;
    }
    if (csv != null) {
      $result.csv = csv;
    }
    return $result;
  }
  CreateMatchRequest._() : super();
  factory CreateMatchRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateMatchRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateMatchRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'aC')
    ..aOS(2, _omitFieldNames ? '' : 'bC')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'csv', $pb.PbFieldType.OU3)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateMatchRequest clone() => CreateMatchRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateMatchRequest copyWith(void Function(CreateMatchRequest) updates) => super.copyWith((message) => updates(message as CreateMatchRequest)) as CreateMatchRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateMatchRequest create() => CreateMatchRequest._();
  CreateMatchRequest createEmptyInstance() => create();
  static $pb.PbList<CreateMatchRequest> createRepeated() => $pb.PbList<CreateMatchRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateMatchRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateMatchRequest>(create);
  static CreateMatchRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get aC => $_getSZ(0);
  @$pb.TagNumber(1)
  set aC($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAC() => $_has(0);
  @$pb.TagNumber(1)
  void clearAC() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get bC => $_getSZ(1);
  @$pb.TagNumber(2)
  set bC($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasBC() => $_has(1);
  @$pb.TagNumber(2)
  void clearBC() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get csv => $_getIZ(2);
  @$pb.TagNumber(3)
  set csv($core.int v) { $_setUnsignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasCsv() => $_has(2);
  @$pb.TagNumber(3)
  void clearCsv() => clearField(3);
}

class CreateMatchResponse extends $pb.GeneratedMessage {
  factory CreateMatchResponse({
    $core.String? matchId,
    $core.String? sC,
    $core.String? aA,
    $core.String? aB,
    $core.String? aS,
    $core.String? xA,
    $core.String? xB,
    $core.int? csv,
    $core.String? escrowTemplateA,
    $core.String? escrowTemplateB,
  }) {
    final $result = create();
    if (matchId != null) {
      $result.matchId = matchId;
    }
    if (sC != null) {
      $result.sC = sC;
    }
    if (aA != null) {
      $result.aA = aA;
    }
    if (aB != null) {
      $result.aB = aB;
    }
    if (aS != null) {
      $result.aS = aS;
    }
    if (xA != null) {
      $result.xA = xA;
    }
    if (xB != null) {
      $result.xB = xB;
    }
    if (csv != null) {
      $result.csv = csv;
    }
    if (escrowTemplateA != null) {
      $result.escrowTemplateA = escrowTemplateA;
    }
    if (escrowTemplateB != null) {
      $result.escrowTemplateB = escrowTemplateB;
    }
    return $result;
  }
  CreateMatchResponse._() : super();
  factory CreateMatchResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateMatchResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateMatchResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..aOS(2, _omitFieldNames ? '' : 'sC')
    ..aOS(3, _omitFieldNames ? '' : 'aA')
    ..aOS(4, _omitFieldNames ? '' : 'aB')
    ..aOS(5, _omitFieldNames ? '' : 'aS')
    ..aOS(6, _omitFieldNames ? '' : 'xA')
    ..aOS(7, _omitFieldNames ? '' : 'xB')
    ..a<$core.int>(8, _omitFieldNames ? '' : 'csv', $pb.PbFieldType.OU3)
    ..aOS(9, _omitFieldNames ? '' : 'escrowTemplateA')
    ..aOS(10, _omitFieldNames ? '' : 'escrowTemplateB')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateMatchResponse clone() => CreateMatchResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateMatchResponse copyWith(void Function(CreateMatchResponse) updates) => super.copyWith((message) => updates(message as CreateMatchResponse)) as CreateMatchResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateMatchResponse create() => CreateMatchResponse._();
  CreateMatchResponse createEmptyInstance() => create();
  static $pb.PbList<CreateMatchResponse> createRepeated() => $pb.PbList<CreateMatchResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateMatchResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateMatchResponse>(create);
  static CreateMatchResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get sC => $_getSZ(1);
  @$pb.TagNumber(2)
  set sC($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasSC() => $_has(1);
  @$pb.TagNumber(2)
  void clearSC() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get aA => $_getSZ(2);
  @$pb.TagNumber(3)
  set aA($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasAA() => $_has(2);
  @$pb.TagNumber(3)
  void clearAA() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get aB => $_getSZ(3);
  @$pb.TagNumber(4)
  set aB($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasAB() => $_has(3);
  @$pb.TagNumber(4)
  void clearAB() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get aS => $_getSZ(4);
  @$pb.TagNumber(5)
  set aS($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasAS() => $_has(4);
  @$pb.TagNumber(5)
  void clearAS() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get xA => $_getSZ(5);
  @$pb.TagNumber(6)
  set xA($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasXA() => $_has(5);
  @$pb.TagNumber(6)
  void clearXA() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get xB => $_getSZ(6);
  @$pb.TagNumber(7)
  set xB($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasXB() => $_has(6);
  @$pb.TagNumber(7)
  void clearXB() => clearField(7);

  @$pb.TagNumber(8)
  $core.int get csv => $_getIZ(7);
  @$pb.TagNumber(8)
  set csv($core.int v) { $_setUnsignedInt32(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasCsv() => $_has(7);
  @$pb.TagNumber(8)
  void clearCsv() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get escrowTemplateA => $_getSZ(8);
  @$pb.TagNumber(9)
  set escrowTemplateA($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasEscrowTemplateA() => $_has(8);
  @$pb.TagNumber(9)
  void clearEscrowTemplateA() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get escrowTemplateB => $_getSZ(9);
  @$pb.TagNumber(10)
  set escrowTemplateB($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasEscrowTemplateB() => $_has(9);
  @$pb.TagNumber(10)
  void clearEscrowTemplateB() => clearField(10);
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
    final $result = create();
    if (txid != null) {
      $result.txid = txid;
    }
    if (vout != null) {
      $result.vout = vout;
    }
    if (value != null) {
      $result.value = value;
    }
    if (redeemScriptHex != null) {
      $result.redeemScriptHex = redeemScriptHex;
    }
    if (pkScriptHex != null) {
      $result.pkScriptHex = pkScriptHex;
    }
    if (owner != null) {
      $result.owner = owner;
    }
    return $result;
  }
  EscrowUTXO._() : super();
  factory EscrowUTXO.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory EscrowUTXO.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'EscrowUTXO', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'txid')
    ..a<$core.int>(2, _omitFieldNames ? '' : 'vout', $pb.PbFieldType.OU3)
    ..a<$fixnum.Int64>(3, _omitFieldNames ? '' : 'value', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(4, _omitFieldNames ? '' : 'redeemScriptHex')
    ..aOS(5, _omitFieldNames ? '' : 'pkScriptHex')
    ..aOS(6, _omitFieldNames ? '' : 'owner')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  EscrowUTXO clone() => EscrowUTXO()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  EscrowUTXO copyWith(void Function(EscrowUTXO) updates) => super.copyWith((message) => updates(message as EscrowUTXO)) as EscrowUTXO;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static EscrowUTXO create() => EscrowUTXO._();
  EscrowUTXO createEmptyInstance() => create();
  static $pb.PbList<EscrowUTXO> createRepeated() => $pb.PbList<EscrowUTXO>();
  @$core.pragma('dart2js:noInline')
  static EscrowUTXO getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<EscrowUTXO>(create);
  static EscrowUTXO? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get txid => $_getSZ(0);
  @$pb.TagNumber(1)
  set txid($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasTxid() => $_has(0);
  @$pb.TagNumber(1)
  void clearTxid() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get vout => $_getIZ(1);
  @$pb.TagNumber(2)
  set vout($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasVout() => $_has(1);
  @$pb.TagNumber(2)
  void clearVout() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get value => $_getI64(2);
  @$pb.TagNumber(3)
  set value($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasValue() => $_has(2);
  @$pb.TagNumber(3)
  void clearValue() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get redeemScriptHex => $_getSZ(3);
  @$pb.TagNumber(4)
  set redeemScriptHex($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasRedeemScriptHex() => $_has(3);
  @$pb.TagNumber(4)
  void clearRedeemScriptHex() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get pkScriptHex => $_getSZ(4);
  @$pb.TagNumber(5)
  set pkScriptHex($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasPkScriptHex() => $_has(4);
  @$pb.TagNumber(5)
  void clearPkScriptHex() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get owner => $_getSZ(5);
  @$pb.TagNumber(6)
  set owner($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasOwner() => $_has(5);
  @$pb.TagNumber(6)
  void clearOwner() => clearField(6);
}

class SubmitFundingRequest extends $pb.GeneratedMessage {
  factory SubmitFundingRequest({
    $core.String? matchId,
    $core.Iterable<EscrowUTXO>? escrows,
  }) {
    final $result = create();
    if (matchId != null) {
      $result.matchId = matchId;
    }
    if (escrows != null) {
      $result.escrows.addAll(escrows);
    }
    return $result;
  }
  SubmitFundingRequest._() : super();
  factory SubmitFundingRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SubmitFundingRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SubmitFundingRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..pc<EscrowUTXO>(2, _omitFieldNames ? '' : 'escrows', $pb.PbFieldType.PM, subBuilder: EscrowUTXO.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SubmitFundingRequest clone() => SubmitFundingRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SubmitFundingRequest copyWith(void Function(SubmitFundingRequest) updates) => super.copyWith((message) => updates(message as SubmitFundingRequest)) as SubmitFundingRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SubmitFundingRequest create() => SubmitFundingRequest._();
  SubmitFundingRequest createEmptyInstance() => create();
  static $pb.PbList<SubmitFundingRequest> createRepeated() => $pb.PbList<SubmitFundingRequest>();
  @$core.pragma('dart2js:noInline')
  static SubmitFundingRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SubmitFundingRequest>(create);
  static SubmitFundingRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<EscrowUTXO> get escrows => $_getList(1);
}

class SubmitFundingResponse extends $pb.GeneratedMessage {
  factory SubmitFundingResponse({
    $core.bool? ok,
  }) {
    final $result = create();
    if (ok != null) {
      $result.ok = ok;
    }
    return $result;
  }
  SubmitFundingResponse._() : super();
  factory SubmitFundingResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SubmitFundingResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SubmitFundingResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'ok')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SubmitFundingResponse clone() => SubmitFundingResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SubmitFundingResponse copyWith(void Function(SubmitFundingResponse) updates) => super.copyWith((message) => updates(message as SubmitFundingResponse)) as SubmitFundingResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SubmitFundingResponse create() => SubmitFundingResponse._();
  SubmitFundingResponse createEmptyInstance() => create();
  static $pb.PbList<SubmitFundingResponse> createRepeated() => $pb.PbList<SubmitFundingResponse>();
  @$core.pragma('dart2js:noInline')
  static SubmitFundingResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SubmitFundingResponse>(create);
  static SubmitFundingResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get ok => $_getBF(0);
  @$pb.TagNumber(1)
  set ok($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasOk() => $_has(0);
  @$pb.TagNumber(1)
  void clearOk() => clearField(1);
}

class RevealAdaptorEntry extends $pb.GeneratedMessage {
  factory RevealAdaptorEntry({
    $core.String? inputId,
    $core.String? gamma,
  }) {
    final $result = create();
    if (inputId != null) {
      $result.inputId = inputId;
    }
    if (gamma != null) {
      $result.gamma = gamma;
    }
    return $result;
  }
  RevealAdaptorEntry._() : super();
  factory RevealAdaptorEntry.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RevealAdaptorEntry.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'RevealAdaptorEntry', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'inputId')
    ..aOS(2, _omitFieldNames ? '' : 'gamma')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RevealAdaptorEntry clone() => RevealAdaptorEntry()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RevealAdaptorEntry copyWith(void Function(RevealAdaptorEntry) updates) => super.copyWith((message) => updates(message as RevealAdaptorEntry)) as RevealAdaptorEntry;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RevealAdaptorEntry create() => RevealAdaptorEntry._();
  RevealAdaptorEntry createEmptyInstance() => create();
  static $pb.PbList<RevealAdaptorEntry> createRepeated() => $pb.PbList<RevealAdaptorEntry>();
  @$core.pragma('dart2js:noInline')
  static RevealAdaptorEntry getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RevealAdaptorEntry>(create);
  static RevealAdaptorEntry? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get inputId => $_getSZ(0);
  @$pb.TagNumber(1)
  set inputId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasInputId() => $_has(0);
  @$pb.TagNumber(1)
  void clearInputId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get gamma => $_getSZ(1);
  @$pb.TagNumber(2)
  set gamma($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasGamma() => $_has(1);
  @$pb.TagNumber(2)
  void clearGamma() => clearField(2);
}

class RevealAdaptorsRequest extends $pb.GeneratedMessage {
  factory RevealAdaptorsRequest({
    $core.String? matchId,
    Branch? branch,
  }) {
    final $result = create();
    if (matchId != null) {
      $result.matchId = matchId;
    }
    if (branch != null) {
      $result.branch = branch;
    }
    return $result;
  }
  RevealAdaptorsRequest._() : super();
  factory RevealAdaptorsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RevealAdaptorsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'RevealAdaptorsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..e<Branch>(2, _omitFieldNames ? '' : 'branch', $pb.PbFieldType.OE, defaultOrMaker: Branch.BRANCH_UNSPECIFIED, valueOf: Branch.valueOf, enumValues: Branch.values)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RevealAdaptorsRequest clone() => RevealAdaptorsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RevealAdaptorsRequest copyWith(void Function(RevealAdaptorsRequest) updates) => super.copyWith((message) => updates(message as RevealAdaptorsRequest)) as RevealAdaptorsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RevealAdaptorsRequest create() => RevealAdaptorsRequest._();
  RevealAdaptorsRequest createEmptyInstance() => create();
  static $pb.PbList<RevealAdaptorsRequest> createRepeated() => $pb.PbList<RevealAdaptorsRequest>();
  @$core.pragma('dart2js:noInline')
  static RevealAdaptorsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RevealAdaptorsRequest>(create);
  static RevealAdaptorsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => clearField(1);

  @$pb.TagNumber(2)
  Branch get branch => $_getN(1);
  @$pb.TagNumber(2)
  set branch(Branch v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasBranch() => $_has(1);
  @$pb.TagNumber(2)
  void clearBranch() => clearField(2);
}

class RevealAdaptorsResponse extends $pb.GeneratedMessage {
  factory RevealAdaptorsResponse({
    Branch? branch,
    $core.Iterable<RevealAdaptorEntry>? entries,
  }) {
    final $result = create();
    if (branch != null) {
      $result.branch = branch;
    }
    if (entries != null) {
      $result.entries.addAll(entries);
    }
    return $result;
  }
  RevealAdaptorsResponse._() : super();
  factory RevealAdaptorsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory RevealAdaptorsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'RevealAdaptorsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..e<Branch>(1, _omitFieldNames ? '' : 'branch', $pb.PbFieldType.OE, defaultOrMaker: Branch.BRANCH_UNSPECIFIED, valueOf: Branch.valueOf, enumValues: Branch.values)
    ..pc<RevealAdaptorEntry>(2, _omitFieldNames ? '' : 'entries', $pb.PbFieldType.PM, subBuilder: RevealAdaptorEntry.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  RevealAdaptorsResponse clone() => RevealAdaptorsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  RevealAdaptorsResponse copyWith(void Function(RevealAdaptorsResponse) updates) => super.copyWith((message) => updates(message as RevealAdaptorsResponse)) as RevealAdaptorsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static RevealAdaptorsResponse create() => RevealAdaptorsResponse._();
  RevealAdaptorsResponse createEmptyInstance() => create();
  static $pb.PbList<RevealAdaptorsResponse> createRepeated() => $pb.PbList<RevealAdaptorsResponse>();
  @$core.pragma('dart2js:noInline')
  static RevealAdaptorsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<RevealAdaptorsResponse>(create);
  static RevealAdaptorsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  Branch get branch => $_getN(0);
  @$pb.TagNumber(1)
  set branch(Branch v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasBranch() => $_has(0);
  @$pb.TagNumber(1)
  void clearBranch() => clearField(1);

  @$pb.TagNumber(2)
  $core.List<RevealAdaptorEntry> get entries => $_getList(1);
}

/// v0-min POC messages
class AllocateMatchRequest extends $pb.GeneratedMessage {
  factory AllocateMatchRequest({
    $core.String? aC,
    $core.String? bC,
    $fixnum.Int64? betAtoms,
  }) {
    final $result = create();
    if (aC != null) {
      $result.aC = aC;
    }
    if (bC != null) {
      $result.bC = bC;
    }
    if (betAtoms != null) {
      $result.betAtoms = betAtoms;
    }
    return $result;
  }
  AllocateMatchRequest._() : super();
  factory AllocateMatchRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AllocateMatchRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'AllocateMatchRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'aC')
    ..aOS(2, _omitFieldNames ? '' : 'bC')
    ..a<$fixnum.Int64>(3, _omitFieldNames ? '' : 'betAtoms', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AllocateMatchRequest clone() => AllocateMatchRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AllocateMatchRequest copyWith(void Function(AllocateMatchRequest) updates) => super.copyWith((message) => updates(message as AllocateMatchRequest)) as AllocateMatchRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AllocateMatchRequest create() => AllocateMatchRequest._();
  AllocateMatchRequest createEmptyInstance() => create();
  static $pb.PbList<AllocateMatchRequest> createRepeated() => $pb.PbList<AllocateMatchRequest>();
  @$core.pragma('dart2js:noInline')
  static AllocateMatchRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AllocateMatchRequest>(create);
  static AllocateMatchRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get aC => $_getSZ(0);
  @$pb.TagNumber(1)
  set aC($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasAC() => $_has(0);
  @$pb.TagNumber(1)
  void clearAC() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get bC => $_getSZ(1);
  @$pb.TagNumber(2)
  set bC($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasBC() => $_has(1);
  @$pb.TagNumber(2)
  void clearBC() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get betAtoms => $_getI64(2);
  @$pb.TagNumber(3)
  set betAtoms($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasBetAtoms() => $_has(2);
  @$pb.TagNumber(3)
  void clearBetAtoms() => clearField(3);
}

class AllocateMatchResponse extends $pb.GeneratedMessage {
  factory AllocateMatchResponse({
    $core.String? matchId,
    $core.int? csv,
    $fixnum.Int64? betAtoms,
    $fixnum.Int64? feeAtoms,
    $core.String? aRedeemScriptHex,
    $core.String? aPkScriptHex,
    $core.String? aDepositAddress,
    $core.String? bRedeemScriptHex,
    $core.String? bPkScriptHex,
    $core.String? bDepositAddress,
  }) {
    final $result = create();
    if (matchId != null) {
      $result.matchId = matchId;
    }
    if (csv != null) {
      $result.csv = csv;
    }
    if (betAtoms != null) {
      $result.betAtoms = betAtoms;
    }
    if (feeAtoms != null) {
      $result.feeAtoms = feeAtoms;
    }
    if (aRedeemScriptHex != null) {
      $result.aRedeemScriptHex = aRedeemScriptHex;
    }
    if (aPkScriptHex != null) {
      $result.aPkScriptHex = aPkScriptHex;
    }
    if (aDepositAddress != null) {
      $result.aDepositAddress = aDepositAddress;
    }
    if (bRedeemScriptHex != null) {
      $result.bRedeemScriptHex = bRedeemScriptHex;
    }
    if (bPkScriptHex != null) {
      $result.bPkScriptHex = bPkScriptHex;
    }
    if (bDepositAddress != null) {
      $result.bDepositAddress = bDepositAddress;
    }
    return $result;
  }
  AllocateMatchResponse._() : super();
  factory AllocateMatchResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AllocateMatchResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'AllocateMatchResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..a<$core.int>(2, _omitFieldNames ? '' : 'csv', $pb.PbFieldType.OU3)
    ..a<$fixnum.Int64>(3, _omitFieldNames ? '' : 'betAtoms', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$fixnum.Int64>(4, _omitFieldNames ? '' : 'feeAtoms', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..aOS(5, _omitFieldNames ? '' : 'aRedeemScriptHex')
    ..aOS(6, _omitFieldNames ? '' : 'aPkScriptHex')
    ..aOS(7, _omitFieldNames ? '' : 'aDepositAddress')
    ..aOS(8, _omitFieldNames ? '' : 'bRedeemScriptHex')
    ..aOS(9, _omitFieldNames ? '' : 'bPkScriptHex')
    ..aOS(10, _omitFieldNames ? '' : 'bDepositAddress')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AllocateMatchResponse clone() => AllocateMatchResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AllocateMatchResponse copyWith(void Function(AllocateMatchResponse) updates) => super.copyWith((message) => updates(message as AllocateMatchResponse)) as AllocateMatchResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AllocateMatchResponse create() => AllocateMatchResponse._();
  AllocateMatchResponse createEmptyInstance() => create();
  static $pb.PbList<AllocateMatchResponse> createRepeated() => $pb.PbList<AllocateMatchResponse>();
  @$core.pragma('dart2js:noInline')
  static AllocateMatchResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AllocateMatchResponse>(create);
  static AllocateMatchResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get csv => $_getIZ(1);
  @$pb.TagNumber(2)
  set csv($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasCsv() => $_has(1);
  @$pb.TagNumber(2)
  void clearCsv() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get betAtoms => $_getI64(2);
  @$pb.TagNumber(3)
  set betAtoms($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasBetAtoms() => $_has(2);
  @$pb.TagNumber(3)
  void clearBetAtoms() => clearField(3);

  @$pb.TagNumber(4)
  $fixnum.Int64 get feeAtoms => $_getI64(3);
  @$pb.TagNumber(4)
  set feeAtoms($fixnum.Int64 v) { $_setInt64(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasFeeAtoms() => $_has(3);
  @$pb.TagNumber(4)
  void clearFeeAtoms() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get aRedeemScriptHex => $_getSZ(4);
  @$pb.TagNumber(5)
  set aRedeemScriptHex($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasARedeemScriptHex() => $_has(4);
  @$pb.TagNumber(5)
  void clearARedeemScriptHex() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get aPkScriptHex => $_getSZ(5);
  @$pb.TagNumber(6)
  set aPkScriptHex($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasAPkScriptHex() => $_has(5);
  @$pb.TagNumber(6)
  void clearAPkScriptHex() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get aDepositAddress => $_getSZ(6);
  @$pb.TagNumber(7)
  set aDepositAddress($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasADepositAddress() => $_has(6);
  @$pb.TagNumber(7)
  void clearADepositAddress() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get bRedeemScriptHex => $_getSZ(7);
  @$pb.TagNumber(8)
  set bRedeemScriptHex($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasBRedeemScriptHex() => $_has(7);
  @$pb.TagNumber(8)
  void clearBRedeemScriptHex() => clearField(8);

  @$pb.TagNumber(9)
  $core.String get bPkScriptHex => $_getSZ(8);
  @$pb.TagNumber(9)
  set bPkScriptHex($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasBPkScriptHex() => $_has(8);
  @$pb.TagNumber(9)
  void clearBPkScriptHex() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get bDepositAddress => $_getSZ(9);
  @$pb.TagNumber(10)
  set bDepositAddress($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasBDepositAddress() => $_has(9);
  @$pb.TagNumber(10)
  void clearBDepositAddress() => clearField(10);
}

class WaitFundingRequest extends $pb.GeneratedMessage {
  factory WaitFundingRequest({
    $core.String? matchId,
  }) {
    final $result = create();
    if (matchId != null) {
      $result.matchId = matchId;
    }
    return $result;
  }
  WaitFundingRequest._() : super();
  factory WaitFundingRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WaitFundingRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WaitFundingRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WaitFundingRequest clone() => WaitFundingRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WaitFundingRequest copyWith(void Function(WaitFundingRequest) updates) => super.copyWith((message) => updates(message as WaitFundingRequest)) as WaitFundingRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitFundingRequest create() => WaitFundingRequest._();
  WaitFundingRequest createEmptyInstance() => create();
  static $pb.PbList<WaitFundingRequest> createRepeated() => $pb.PbList<WaitFundingRequest>();
  @$core.pragma('dart2js:noInline')
  static WaitFundingRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WaitFundingRequest>(create);
  static WaitFundingRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => clearField(1);
}

class WaitFundingResponse extends $pb.GeneratedMessage {
  factory WaitFundingResponse({
    $core.bool? confirmed,
    $core.int? confsA,
    $core.int? confsB,
    EscrowUTXO? utxoA,
    EscrowUTXO? utxoB,
  }) {
    final $result = create();
    if (confirmed != null) {
      $result.confirmed = confirmed;
    }
    if (confsA != null) {
      $result.confsA = confsA;
    }
    if (confsB != null) {
      $result.confsB = confsB;
    }
    if (utxoA != null) {
      $result.utxoA = utxoA;
    }
    if (utxoB != null) {
      $result.utxoB = utxoB;
    }
    return $result;
  }
  WaitFundingResponse._() : super();
  factory WaitFundingResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WaitFundingResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WaitFundingResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'confirmed')
    ..a<$core.int>(2, _omitFieldNames ? '' : 'confsA', $pb.PbFieldType.OU3)
    ..a<$core.int>(3, _omitFieldNames ? '' : 'confsB', $pb.PbFieldType.OU3)
    ..aOM<EscrowUTXO>(4, _omitFieldNames ? '' : 'utxoA', subBuilder: EscrowUTXO.create)
    ..aOM<EscrowUTXO>(5, _omitFieldNames ? '' : 'utxoB', subBuilder: EscrowUTXO.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WaitFundingResponse clone() => WaitFundingResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WaitFundingResponse copyWith(void Function(WaitFundingResponse) updates) => super.copyWith((message) => updates(message as WaitFundingResponse)) as WaitFundingResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitFundingResponse create() => WaitFundingResponse._();
  WaitFundingResponse createEmptyInstance() => create();
  static $pb.PbList<WaitFundingResponse> createRepeated() => $pb.PbList<WaitFundingResponse>();
  @$core.pragma('dart2js:noInline')
  static WaitFundingResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WaitFundingResponse>(create);
  static WaitFundingResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get confirmed => $_getBF(0);
  @$pb.TagNumber(1)
  set confirmed($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasConfirmed() => $_has(0);
  @$pb.TagNumber(1)
  void clearConfirmed() => clearField(1);

  @$pb.TagNumber(2)
  $core.int get confsA => $_getIZ(1);
  @$pb.TagNumber(2)
  set confsA($core.int v) { $_setUnsignedInt32(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasConfsA() => $_has(1);
  @$pb.TagNumber(2)
  void clearConfsA() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get confsB => $_getIZ(2);
  @$pb.TagNumber(3)
  set confsB($core.int v) { $_setUnsignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasConfsB() => $_has(2);
  @$pb.TagNumber(3)
  void clearConfsB() => clearField(3);

  @$pb.TagNumber(4)
  EscrowUTXO get utxoA => $_getN(3);
  @$pb.TagNumber(4)
  set utxoA(EscrowUTXO v) { setField(4, v); }
  @$pb.TagNumber(4)
  $core.bool hasUtxoA() => $_has(3);
  @$pb.TagNumber(4)
  void clearUtxoA() => clearField(4);
  @$pb.TagNumber(4)
  EscrowUTXO ensureUtxoA() => $_ensure(3);

  @$pb.TagNumber(5)
  EscrowUTXO get utxoB => $_getN(4);
  @$pb.TagNumber(5)
  set utxoB(EscrowUTXO v) { setField(5, v); }
  @$pb.TagNumber(5)
  $core.bool hasUtxoB() => $_has(4);
  @$pb.TagNumber(5)
  void clearUtxoB() => clearField(5);
  @$pb.TagNumber(5)
  EscrowUTXO ensureUtxoB() => $_ensure(4);
}

/// Server-managed deposit allocation
class AllocateEscrowRequest extends $pb.GeneratedMessage {
  factory AllocateEscrowRequest({
    $core.String? playerId,
    $core.String? aC,
    $fixnum.Int64? betAtoms,
    $core.int? csv,
  }) {
    final $result = create();
    if (playerId != null) {
      $result.playerId = playerId;
    }
    if (aC != null) {
      $result.aC = aC;
    }
    if (betAtoms != null) {
      $result.betAtoms = betAtoms;
    }
    if (csv != null) {
      $result.csv = csv;
    }
    return $result;
  }
  AllocateEscrowRequest._() : super();
  factory AllocateEscrowRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AllocateEscrowRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'AllocateEscrowRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'playerId')
    ..aOS(2, _omitFieldNames ? '' : 'aC')
    ..a<$fixnum.Int64>(3, _omitFieldNames ? '' : 'betAtoms', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..a<$core.int>(4, _omitFieldNames ? '' : 'csv', $pb.PbFieldType.OU3)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AllocateEscrowRequest clone() => AllocateEscrowRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AllocateEscrowRequest copyWith(void Function(AllocateEscrowRequest) updates) => super.copyWith((message) => updates(message as AllocateEscrowRequest)) as AllocateEscrowRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AllocateEscrowRequest create() => AllocateEscrowRequest._();
  AllocateEscrowRequest createEmptyInstance() => create();
  static $pb.PbList<AllocateEscrowRequest> createRepeated() => $pb.PbList<AllocateEscrowRequest>();
  @$core.pragma('dart2js:noInline')
  static AllocateEscrowRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AllocateEscrowRequest>(create);
  static AllocateEscrowRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get playerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set playerId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPlayerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlayerId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get aC => $_getSZ(1);
  @$pb.TagNumber(2)
  set aC($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasAC() => $_has(1);
  @$pb.TagNumber(2)
  void clearAC() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get betAtoms => $_getI64(2);
  @$pb.TagNumber(3)
  set betAtoms($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasBetAtoms() => $_has(2);
  @$pb.TagNumber(3)
  void clearBetAtoms() => clearField(3);

  @$pb.TagNumber(4)
  $core.int get csv => $_getIZ(3);
  @$pb.TagNumber(4)
  set csv($core.int v) { $_setUnsignedInt32(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasCsv() => $_has(3);
  @$pb.TagNumber(4)
  void clearCsv() => clearField(4);
}

class AllocateEscrowResponse extends $pb.GeneratedMessage {
  factory AllocateEscrowResponse({
    $core.String? matchId,
    Branch? role,
    $core.String? aA,
    $core.String? aB,
    $core.String? aS,
    $core.String? xA,
    $core.String? xB,
    $core.int? csv,
    $core.String? depositAddress,
    $core.String? redeemScriptHex,
    $core.String? pkScriptHex,
    $fixnum.Int64? requiredAtoms,
  }) {
    final $result = create();
    if (matchId != null) {
      $result.matchId = matchId;
    }
    if (role != null) {
      $result.role = role;
    }
    if (aA != null) {
      $result.aA = aA;
    }
    if (aB != null) {
      $result.aB = aB;
    }
    if (aS != null) {
      $result.aS = aS;
    }
    if (xA != null) {
      $result.xA = xA;
    }
    if (xB != null) {
      $result.xB = xB;
    }
    if (csv != null) {
      $result.csv = csv;
    }
    if (depositAddress != null) {
      $result.depositAddress = depositAddress;
    }
    if (redeemScriptHex != null) {
      $result.redeemScriptHex = redeemScriptHex;
    }
    if (pkScriptHex != null) {
      $result.pkScriptHex = pkScriptHex;
    }
    if (requiredAtoms != null) {
      $result.requiredAtoms = requiredAtoms;
    }
    return $result;
  }
  AllocateEscrowResponse._() : super();
  factory AllocateEscrowResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory AllocateEscrowResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'AllocateEscrowResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..e<Branch>(2, _omitFieldNames ? '' : 'role', $pb.PbFieldType.OE, defaultOrMaker: Branch.BRANCH_UNSPECIFIED, valueOf: Branch.valueOf, enumValues: Branch.values)
    ..aOS(3, _omitFieldNames ? '' : 'aA')
    ..aOS(4, _omitFieldNames ? '' : 'aB')
    ..aOS(5, _omitFieldNames ? '' : 'aS')
    ..aOS(6, _omitFieldNames ? '' : 'xA')
    ..aOS(7, _omitFieldNames ? '' : 'xB')
    ..a<$core.int>(8, _omitFieldNames ? '' : 'csv', $pb.PbFieldType.OU3)
    ..aOS(9, _omitFieldNames ? '' : 'depositAddress')
    ..aOS(10, _omitFieldNames ? '' : 'redeemScriptHex')
    ..aOS(11, _omitFieldNames ? '' : 'pkScriptHex')
    ..a<$fixnum.Int64>(12, _omitFieldNames ? '' : 'requiredAtoms', $pb.PbFieldType.OU6, defaultOrMaker: $fixnum.Int64.ZERO)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  AllocateEscrowResponse clone() => AllocateEscrowResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  AllocateEscrowResponse copyWith(void Function(AllocateEscrowResponse) updates) => super.copyWith((message) => updates(message as AllocateEscrowResponse)) as AllocateEscrowResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static AllocateEscrowResponse create() => AllocateEscrowResponse._();
  AllocateEscrowResponse createEmptyInstance() => create();
  static $pb.PbList<AllocateEscrowResponse> createRepeated() => $pb.PbList<AllocateEscrowResponse>();
  @$core.pragma('dart2js:noInline')
  static AllocateEscrowResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<AllocateEscrowResponse>(create);
  static AllocateEscrowResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => clearField(1);

  @$pb.TagNumber(2)
  Branch get role => $_getN(1);
  @$pb.TagNumber(2)
  set role(Branch v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasRole() => $_has(1);
  @$pb.TagNumber(2)
  void clearRole() => clearField(2);

  /// Coefficients and branch keys for client-side checks
  @$pb.TagNumber(3)
  $core.String get aA => $_getSZ(2);
  @$pb.TagNumber(3)
  set aA($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasAA() => $_has(2);
  @$pb.TagNumber(3)
  void clearAA() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get aB => $_getSZ(3);
  @$pb.TagNumber(4)
  set aB($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasAB() => $_has(3);
  @$pb.TagNumber(4)
  void clearAB() => clearField(4);

  @$pb.TagNumber(5)
  $core.String get aS => $_getSZ(4);
  @$pb.TagNumber(5)
  set aS($core.String v) { $_setString(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasAS() => $_has(4);
  @$pb.TagNumber(5)
  void clearAS() => clearField(5);

  @$pb.TagNumber(6)
  $core.String get xA => $_getSZ(5);
  @$pb.TagNumber(6)
  set xA($core.String v) { $_setString(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasXA() => $_has(5);
  @$pb.TagNumber(6)
  void clearXA() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get xB => $_getSZ(6);
  @$pb.TagNumber(7)
  set xB($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasXB() => $_has(6);
  @$pb.TagNumber(7)
  void clearXB() => clearField(7);

  @$pb.TagNumber(8)
  $core.int get csv => $_getIZ(7);
  @$pb.TagNumber(8)
  set csv($core.int v) { $_setUnsignedInt32(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasCsv() => $_has(7);
  @$pb.TagNumber(8)
  void clearCsv() => clearField(8);

  /// Deposit destination
  @$pb.TagNumber(9)
  $core.String get depositAddress => $_getSZ(8);
  @$pb.TagNumber(9)
  set depositAddress($core.String v) { $_setString(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasDepositAddress() => $_has(8);
  @$pb.TagNumber(9)
  void clearDepositAddress() => clearField(9);

  @$pb.TagNumber(10)
  $core.String get redeemScriptHex => $_getSZ(9);
  @$pb.TagNumber(10)
  set redeemScriptHex($core.String v) { $_setString(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasRedeemScriptHex() => $_has(9);
  @$pb.TagNumber(10)
  void clearRedeemScriptHex() => clearField(10);

  @$pb.TagNumber(11)
  $core.String get pkScriptHex => $_getSZ(10);
  @$pb.TagNumber(11)
  set pkScriptHex($core.String v) { $_setString(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasPkScriptHex() => $_has(10);
  @$pb.TagNumber(11)
  void clearPkScriptHex() => clearField(11);

  @$pb.TagNumber(12)
  $fixnum.Int64 get requiredAtoms => $_getI64(11);
  @$pb.TagNumber(12)
  set requiredAtoms($fixnum.Int64 v) { $_setInt64(11, v); }
  @$pb.TagNumber(12)
  $core.bool hasRequiredAtoms() => $_has(11);
  @$pb.TagNumber(12)
  void clearRequiredAtoms() => clearField(12);
}

class FinalizeWinnerRequest extends $pb.GeneratedMessage {
  factory FinalizeWinnerRequest({
    $core.String? matchId,
    Branch? branch,
  }) {
    final $result = create();
    if (matchId != null) {
      $result.matchId = matchId;
    }
    if (branch != null) {
      $result.branch = branch;
    }
    return $result;
  }
  FinalizeWinnerRequest._() : super();
  factory FinalizeWinnerRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory FinalizeWinnerRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'FinalizeWinnerRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'matchId')
    ..e<Branch>(2, _omitFieldNames ? '' : 'branch', $pb.PbFieldType.OE, defaultOrMaker: Branch.BRANCH_UNSPECIFIED, valueOf: Branch.valueOf, enumValues: Branch.values)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  FinalizeWinnerRequest clone() => FinalizeWinnerRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  FinalizeWinnerRequest copyWith(void Function(FinalizeWinnerRequest) updates) => super.copyWith((message) => updates(message as FinalizeWinnerRequest)) as FinalizeWinnerRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static FinalizeWinnerRequest create() => FinalizeWinnerRequest._();
  FinalizeWinnerRequest createEmptyInstance() => create();
  static $pb.PbList<FinalizeWinnerRequest> createRepeated() => $pb.PbList<FinalizeWinnerRequest>();
  @$core.pragma('dart2js:noInline')
  static FinalizeWinnerRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<FinalizeWinnerRequest>(create);
  static FinalizeWinnerRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get matchId => $_getSZ(0);
  @$pb.TagNumber(1)
  set matchId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasMatchId() => $_has(0);
  @$pb.TagNumber(1)
  void clearMatchId() => clearField(1);

  @$pb.TagNumber(2)
  Branch get branch => $_getN(1);
  @$pb.TagNumber(2)
  set branch(Branch v) { setField(2, v); }
  @$pb.TagNumber(2)
  $core.bool hasBranch() => $_has(1);
  @$pb.TagNumber(2)
  void clearBranch() => clearField(2);
}

class FinalizeWinnerResponse extends $pb.GeneratedMessage {
  factory FinalizeWinnerResponse({
    $core.String? broadcastTxid,
  }) {
    final $result = create();
    if (broadcastTxid != null) {
      $result.broadcastTxid = broadcastTxid;
    }
    return $result;
  }
  FinalizeWinnerResponse._() : super();
  factory FinalizeWinnerResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory FinalizeWinnerResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'FinalizeWinnerResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'broadcastTxid')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  FinalizeWinnerResponse clone() => FinalizeWinnerResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  FinalizeWinnerResponse copyWith(void Function(FinalizeWinnerResponse) updates) => super.copyWith((message) => updates(message as FinalizeWinnerResponse)) as FinalizeWinnerResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static FinalizeWinnerResponse create() => FinalizeWinnerResponse._();
  FinalizeWinnerResponse createEmptyInstance() => create();
  static $pb.PbList<FinalizeWinnerResponse> createRepeated() => $pb.PbList<FinalizeWinnerResponse>();
  @$core.pragma('dart2js:noInline')
  static FinalizeWinnerResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<FinalizeWinnerResponse>(create);
  static FinalizeWinnerResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get broadcastTxid => $_getSZ(0);
  @$pb.TagNumber(1)
  set broadcastTxid($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasBroadcastTxid() => $_has(0);
  @$pb.TagNumber(1)
  void clearBroadcastTxid() => clearField(1);
}

class UnreadyGameStreamRequest extends $pb.GeneratedMessage {
  factory UnreadyGameStreamRequest({
    $core.String? clientId,
  }) {
    final $result = create();
    if (clientId != null) {
      $result.clientId = clientId;
    }
    return $result;
  }
  UnreadyGameStreamRequest._() : super();
  factory UnreadyGameStreamRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UnreadyGameStreamRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UnreadyGameStreamRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UnreadyGameStreamRequest clone() => UnreadyGameStreamRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UnreadyGameStreamRequest copyWith(void Function(UnreadyGameStreamRequest) updates) => super.copyWith((message) => updates(message as UnreadyGameStreamRequest)) as UnreadyGameStreamRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UnreadyGameStreamRequest create() => UnreadyGameStreamRequest._();
  UnreadyGameStreamRequest createEmptyInstance() => create();
  static $pb.PbList<UnreadyGameStreamRequest> createRepeated() => $pb.PbList<UnreadyGameStreamRequest>();
  @$core.pragma('dart2js:noInline')
  static UnreadyGameStreamRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UnreadyGameStreamRequest>(create);
  static UnreadyGameStreamRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);
}

class UnreadyGameStreamResponse extends $pb.GeneratedMessage {
  factory UnreadyGameStreamResponse() => create();
  UnreadyGameStreamResponse._() : super();
  factory UnreadyGameStreamResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory UnreadyGameStreamResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'UnreadyGameStreamResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  UnreadyGameStreamResponse clone() => UnreadyGameStreamResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  UnreadyGameStreamResponse copyWith(void Function(UnreadyGameStreamResponse) updates) => super.copyWith((message) => updates(message as UnreadyGameStreamResponse)) as UnreadyGameStreamResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static UnreadyGameStreamResponse create() => UnreadyGameStreamResponse._();
  UnreadyGameStreamResponse createEmptyInstance() => create();
  static $pb.PbList<UnreadyGameStreamResponse> createRepeated() => $pb.PbList<UnreadyGameStreamResponse>();
  @$core.pragma('dart2js:noInline')
  static UnreadyGameStreamResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<UnreadyGameStreamResponse>(create);
  static UnreadyGameStreamResponse? _defaultInstance;
}

class StartNtfnStreamRequest extends $pb.GeneratedMessage {
  factory StartNtfnStreamRequest({
    $core.String? clientId,
  }) {
    final $result = create();
    if (clientId != null) {
      $result.clientId = clientId;
    }
    return $result;
  }
  StartNtfnStreamRequest._() : super();
  factory StartNtfnStreamRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory StartNtfnStreamRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'StartNtfnStreamRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  StartNtfnStreamRequest clone() => StartNtfnStreamRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  StartNtfnStreamRequest copyWith(void Function(StartNtfnStreamRequest) updates) => super.copyWith((message) => updates(message as StartNtfnStreamRequest)) as StartNtfnStreamRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StartNtfnStreamRequest create() => StartNtfnStreamRequest._();
  StartNtfnStreamRequest createEmptyInstance() => create();
  static $pb.PbList<StartNtfnStreamRequest> createRepeated() => $pb.PbList<StartNtfnStreamRequest>();
  @$core.pragma('dart2js:noInline')
  static StartNtfnStreamRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<StartNtfnStreamRequest>(create);
  static StartNtfnStreamRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);
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
  }) {
    final $result = create();
    if (notificationType != null) {
      $result.notificationType = notificationType;
    }
    if (started != null) {
      $result.started = started;
    }
    if (gameId != null) {
      $result.gameId = gameId;
    }
    if (message != null) {
      $result.message = message;
    }
    if (betAmt != null) {
      $result.betAmt = betAmt;
    }
    if (playerNumber != null) {
      $result.playerNumber = playerNumber;
    }
    if (playerId != null) {
      $result.playerId = playerId;
    }
    if (roomId != null) {
      $result.roomId = roomId;
    }
    if (wr != null) {
      $result.wr = wr;
    }
    if (ready != null) {
      $result.ready = ready;
    }
    return $result;
  }
  NtfnStreamResponse._() : super();
  factory NtfnStreamResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory NtfnStreamResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'NtfnStreamResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..e<NotificationType>(1, _omitFieldNames ? '' : 'notificationType', $pb.PbFieldType.OE, defaultOrMaker: NotificationType.UNKNOWN, valueOf: NotificationType.valueOf, enumValues: NotificationType.values)
    ..aOB(2, _omitFieldNames ? '' : 'started')
    ..aOS(3, _omitFieldNames ? '' : 'gameId')
    ..aOS(4, _omitFieldNames ? '' : 'message')
    ..aInt64(5, _omitFieldNames ? '' : 'betAmt', protoName: 'betAmt')
    ..a<$core.int>(6, _omitFieldNames ? '' : 'playerNumber', $pb.PbFieldType.O3)
    ..aOS(7, _omitFieldNames ? '' : 'playerId')
    ..aOS(8, _omitFieldNames ? '' : 'roomId')
    ..aOM<WaitingRoom>(9, _omitFieldNames ? '' : 'wr', subBuilder: WaitingRoom.create)
    ..aOB(10, _omitFieldNames ? '' : 'ready')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  NtfnStreamResponse clone() => NtfnStreamResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  NtfnStreamResponse copyWith(void Function(NtfnStreamResponse) updates) => super.copyWith((message) => updates(message as NtfnStreamResponse)) as NtfnStreamResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static NtfnStreamResponse create() => NtfnStreamResponse._();
  NtfnStreamResponse createEmptyInstance() => create();
  static $pb.PbList<NtfnStreamResponse> createRepeated() => $pb.PbList<NtfnStreamResponse>();
  @$core.pragma('dart2js:noInline')
  static NtfnStreamResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<NtfnStreamResponse>(create);
  static NtfnStreamResponse? _defaultInstance;

  @$pb.TagNumber(1)
  NotificationType get notificationType => $_getN(0);
  @$pb.TagNumber(1)
  set notificationType(NotificationType v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasNotificationType() => $_has(0);
  @$pb.TagNumber(1)
  void clearNotificationType() => clearField(1);

  @$pb.TagNumber(2)
  $core.bool get started => $_getBF(1);
  @$pb.TagNumber(2)
  set started($core.bool v) { $_setBool(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasStarted() => $_has(1);
  @$pb.TagNumber(2)
  void clearStarted() => clearField(2);

  @$pb.TagNumber(3)
  $core.String get gameId => $_getSZ(2);
  @$pb.TagNumber(3)
  set gameId($core.String v) { $_setString(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasGameId() => $_has(2);
  @$pb.TagNumber(3)
  void clearGameId() => clearField(3);

  @$pb.TagNumber(4)
  $core.String get message => $_getSZ(3);
  @$pb.TagNumber(4)
  set message($core.String v) { $_setString(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasMessage() => $_has(3);
  @$pb.TagNumber(4)
  void clearMessage() => clearField(4);

  @$pb.TagNumber(5)
  $fixnum.Int64 get betAmt => $_getI64(4);
  @$pb.TagNumber(5)
  set betAmt($fixnum.Int64 v) { $_setInt64(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasBetAmt() => $_has(4);
  @$pb.TagNumber(5)
  void clearBetAmt() => clearField(5);

  @$pb.TagNumber(6)
  $core.int get playerNumber => $_getIZ(5);
  @$pb.TagNumber(6)
  set playerNumber($core.int v) { $_setSignedInt32(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasPlayerNumber() => $_has(5);
  @$pb.TagNumber(6)
  void clearPlayerNumber() => clearField(6);

  @$pb.TagNumber(7)
  $core.String get playerId => $_getSZ(6);
  @$pb.TagNumber(7)
  set playerId($core.String v) { $_setString(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasPlayerId() => $_has(6);
  @$pb.TagNumber(7)
  void clearPlayerId() => clearField(7);

  @$pb.TagNumber(8)
  $core.String get roomId => $_getSZ(7);
  @$pb.TagNumber(8)
  set roomId($core.String v) { $_setString(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasRoomId() => $_has(7);
  @$pb.TagNumber(8)
  void clearRoomId() => clearField(8);

  @$pb.TagNumber(9)
  WaitingRoom get wr => $_getN(8);
  @$pb.TagNumber(9)
  set wr(WaitingRoom v) { setField(9, v); }
  @$pb.TagNumber(9)
  $core.bool hasWr() => $_has(8);
  @$pb.TagNumber(9)
  void clearWr() => clearField(9);
  @$pb.TagNumber(9)
  WaitingRoom ensureWr() => $_ensure(8);

  @$pb.TagNumber(10)
  $core.bool get ready => $_getBF(9);
  @$pb.TagNumber(10)
  set ready($core.bool v) { $_setBool(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasReady() => $_has(9);
  @$pb.TagNumber(10)
  void clearReady() => clearField(10);
}

/// Waiting Room Messages
class WaitingRoomsRequest extends $pb.GeneratedMessage {
  factory WaitingRoomsRequest({
    $core.String? roomId,
  }) {
    final $result = create();
    if (roomId != null) {
      $result.roomId = roomId;
    }
    return $result;
  }
  WaitingRoomsRequest._() : super();
  factory WaitingRoomsRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WaitingRoomsRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WaitingRoomsRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'roomId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WaitingRoomsRequest clone() => WaitingRoomsRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WaitingRoomsRequest copyWith(void Function(WaitingRoomsRequest) updates) => super.copyWith((message) => updates(message as WaitingRoomsRequest)) as WaitingRoomsRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoomsRequest create() => WaitingRoomsRequest._();
  WaitingRoomsRequest createEmptyInstance() => create();
  static $pb.PbList<WaitingRoomsRequest> createRepeated() => $pb.PbList<WaitingRoomsRequest>();
  @$core.pragma('dart2js:noInline')
  static WaitingRoomsRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WaitingRoomsRequest>(create);
  static WaitingRoomsRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get roomId => $_getSZ(0);
  @$pb.TagNumber(1)
  set roomId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasRoomId() => $_has(0);
  @$pb.TagNumber(1)
  void clearRoomId() => clearField(1);
}

class WaitingRoomsResponse extends $pb.GeneratedMessage {
  factory WaitingRoomsResponse({
    $core.Iterable<WaitingRoom>? wr,
  }) {
    final $result = create();
    if (wr != null) {
      $result.wr.addAll(wr);
    }
    return $result;
  }
  WaitingRoomsResponse._() : super();
  factory WaitingRoomsResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WaitingRoomsResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WaitingRoomsResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..pc<WaitingRoom>(1, _omitFieldNames ? '' : 'wr', $pb.PbFieldType.PM, subBuilder: WaitingRoom.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WaitingRoomsResponse clone() => WaitingRoomsResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WaitingRoomsResponse copyWith(void Function(WaitingRoomsResponse) updates) => super.copyWith((message) => updates(message as WaitingRoomsResponse)) as WaitingRoomsResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoomsResponse create() => WaitingRoomsResponse._();
  WaitingRoomsResponse createEmptyInstance() => create();
  static $pb.PbList<WaitingRoomsResponse> createRepeated() => $pb.PbList<WaitingRoomsResponse>();
  @$core.pragma('dart2js:noInline')
  static WaitingRoomsResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WaitingRoomsResponse>(create);
  static WaitingRoomsResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<WaitingRoom> get wr => $_getList(0);
}

class JoinWaitingRoomRequest extends $pb.GeneratedMessage {
  factory JoinWaitingRoomRequest({
    $core.String? roomId,
    $core.String? clientId,
  }) {
    final $result = create();
    if (roomId != null) {
      $result.roomId = roomId;
    }
    if (clientId != null) {
      $result.clientId = clientId;
    }
    return $result;
  }
  JoinWaitingRoomRequest._() : super();
  factory JoinWaitingRoomRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory JoinWaitingRoomRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'JoinWaitingRoomRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'roomId')
    ..aOS(2, _omitFieldNames ? '' : 'clientId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  JoinWaitingRoomRequest clone() => JoinWaitingRoomRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  JoinWaitingRoomRequest copyWith(void Function(JoinWaitingRoomRequest) updates) => super.copyWith((message) => updates(message as JoinWaitingRoomRequest)) as JoinWaitingRoomRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static JoinWaitingRoomRequest create() => JoinWaitingRoomRequest._();
  JoinWaitingRoomRequest createEmptyInstance() => create();
  static $pb.PbList<JoinWaitingRoomRequest> createRepeated() => $pb.PbList<JoinWaitingRoomRequest>();
  @$core.pragma('dart2js:noInline')
  static JoinWaitingRoomRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<JoinWaitingRoomRequest>(create);
  static JoinWaitingRoomRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get roomId => $_getSZ(0);
  @$pb.TagNumber(1)
  set roomId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasRoomId() => $_has(0);
  @$pb.TagNumber(1)
  void clearRoomId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get clientId => $_getSZ(1);
  @$pb.TagNumber(2)
  set clientId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasClientId() => $_has(1);
  @$pb.TagNumber(2)
  void clearClientId() => clearField(2);
}

class JoinWaitingRoomResponse extends $pb.GeneratedMessage {
  factory JoinWaitingRoomResponse({
    WaitingRoom? wr,
  }) {
    final $result = create();
    if (wr != null) {
      $result.wr = wr;
    }
    return $result;
  }
  JoinWaitingRoomResponse._() : super();
  factory JoinWaitingRoomResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory JoinWaitingRoomResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'JoinWaitingRoomResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOM<WaitingRoom>(1, _omitFieldNames ? '' : 'wr', subBuilder: WaitingRoom.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  JoinWaitingRoomResponse clone() => JoinWaitingRoomResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  JoinWaitingRoomResponse copyWith(void Function(JoinWaitingRoomResponse) updates) => super.copyWith((message) => updates(message as JoinWaitingRoomResponse)) as JoinWaitingRoomResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static JoinWaitingRoomResponse create() => JoinWaitingRoomResponse._();
  JoinWaitingRoomResponse createEmptyInstance() => create();
  static $pb.PbList<JoinWaitingRoomResponse> createRepeated() => $pb.PbList<JoinWaitingRoomResponse>();
  @$core.pragma('dart2js:noInline')
  static JoinWaitingRoomResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<JoinWaitingRoomResponse>(create);
  static JoinWaitingRoomResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WaitingRoom get wr => $_getN(0);
  @$pb.TagNumber(1)
  set wr(WaitingRoom v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasWr() => $_has(0);
  @$pb.TagNumber(1)
  void clearWr() => clearField(1);
  @$pb.TagNumber(1)
  WaitingRoom ensureWr() => $_ensure(0);
}

class CreateWaitingRoomRequest extends $pb.GeneratedMessage {
  factory CreateWaitingRoomRequest({
    $core.String? hostId,
    $fixnum.Int64? betAmt,
  }) {
    final $result = create();
    if (hostId != null) {
      $result.hostId = hostId;
    }
    if (betAmt != null) {
      $result.betAmt = betAmt;
    }
    return $result;
  }
  CreateWaitingRoomRequest._() : super();
  factory CreateWaitingRoomRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateWaitingRoomRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateWaitingRoomRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'hostId')
    ..aInt64(2, _omitFieldNames ? '' : 'betAmt', protoName: 'betAmt')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateWaitingRoomRequest clone() => CreateWaitingRoomRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateWaitingRoomRequest copyWith(void Function(CreateWaitingRoomRequest) updates) => super.copyWith((message) => updates(message as CreateWaitingRoomRequest)) as CreateWaitingRoomRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateWaitingRoomRequest create() => CreateWaitingRoomRequest._();
  CreateWaitingRoomRequest createEmptyInstance() => create();
  static $pb.PbList<CreateWaitingRoomRequest> createRepeated() => $pb.PbList<CreateWaitingRoomRequest>();
  @$core.pragma('dart2js:noInline')
  static CreateWaitingRoomRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateWaitingRoomRequest>(create);
  static CreateWaitingRoomRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get hostId => $_getSZ(0);
  @$pb.TagNumber(1)
  set hostId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasHostId() => $_has(0);
  @$pb.TagNumber(1)
  void clearHostId() => clearField(1);

  @$pb.TagNumber(2)
  $fixnum.Int64 get betAmt => $_getI64(1);
  @$pb.TagNumber(2)
  set betAmt($fixnum.Int64 v) { $_setInt64(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasBetAmt() => $_has(1);
  @$pb.TagNumber(2)
  void clearBetAmt() => clearField(2);
}

class CreateWaitingRoomResponse extends $pb.GeneratedMessage {
  factory CreateWaitingRoomResponse({
    WaitingRoom? wr,
  }) {
    final $result = create();
    if (wr != null) {
      $result.wr = wr;
    }
    return $result;
  }
  CreateWaitingRoomResponse._() : super();
  factory CreateWaitingRoomResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory CreateWaitingRoomResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'CreateWaitingRoomResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOM<WaitingRoom>(1, _omitFieldNames ? '' : 'wr', subBuilder: WaitingRoom.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  CreateWaitingRoomResponse clone() => CreateWaitingRoomResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  CreateWaitingRoomResponse copyWith(void Function(CreateWaitingRoomResponse) updates) => super.copyWith((message) => updates(message as CreateWaitingRoomResponse)) as CreateWaitingRoomResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static CreateWaitingRoomResponse create() => CreateWaitingRoomResponse._();
  CreateWaitingRoomResponse createEmptyInstance() => create();
  static $pb.PbList<CreateWaitingRoomResponse> createRepeated() => $pb.PbList<CreateWaitingRoomResponse>();
  @$core.pragma('dart2js:noInline')
  static CreateWaitingRoomResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<CreateWaitingRoomResponse>(create);
  static CreateWaitingRoomResponse? _defaultInstance;

  @$pb.TagNumber(1)
  WaitingRoom get wr => $_getN(0);
  @$pb.TagNumber(1)
  set wr(WaitingRoom v) { setField(1, v); }
  @$pb.TagNumber(1)
  $core.bool hasWr() => $_has(0);
  @$pb.TagNumber(1)
  void clearWr() => clearField(1);
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
    final $result = create();
    if (id != null) {
      $result.id = id;
    }
    if (hostId != null) {
      $result.hostId = hostId;
    }
    if (players != null) {
      $result.players.addAll(players);
    }
    if (betAmt != null) {
      $result.betAmt = betAmt;
    }
    return $result;
  }
  WaitingRoom._() : super();
  factory WaitingRoom.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WaitingRoom.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WaitingRoom', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'id')
    ..aOS(2, _omitFieldNames ? '' : 'hostId')
    ..pc<Player>(3, _omitFieldNames ? '' : 'players', $pb.PbFieldType.PM, subBuilder: Player.create)
    ..aInt64(4, _omitFieldNames ? '' : 'betAmt')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WaitingRoom clone() => WaitingRoom()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WaitingRoom copyWith(void Function(WaitingRoom) updates) => super.copyWith((message) => updates(message as WaitingRoom)) as WaitingRoom;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoom create() => WaitingRoom._();
  WaitingRoom createEmptyInstance() => create();
  static $pb.PbList<WaitingRoom> createRepeated() => $pb.PbList<WaitingRoom>();
  @$core.pragma('dart2js:noInline')
  static WaitingRoom getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WaitingRoom>(create);
  static WaitingRoom? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get id => $_getSZ(0);
  @$pb.TagNumber(1)
  set id($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasId() => $_has(0);
  @$pb.TagNumber(1)
  void clearId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get hostId => $_getSZ(1);
  @$pb.TagNumber(2)
  set hostId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasHostId() => $_has(1);
  @$pb.TagNumber(2)
  void clearHostId() => clearField(2);

  @$pb.TagNumber(3)
  $core.List<Player> get players => $_getList(2);

  @$pb.TagNumber(4)
  $fixnum.Int64 get betAmt => $_getI64(3);
  @$pb.TagNumber(4)
  set betAmt($fixnum.Int64 v) { $_setInt64(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasBetAmt() => $_has(3);
  @$pb.TagNumber(4)
  void clearBetAmt() => clearField(4);
}

class WaitingRoomRequest extends $pb.GeneratedMessage {
  factory WaitingRoomRequest() => create();
  WaitingRoomRequest._() : super();
  factory WaitingRoomRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WaitingRoomRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WaitingRoomRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WaitingRoomRequest clone() => WaitingRoomRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WaitingRoomRequest copyWith(void Function(WaitingRoomRequest) updates) => super.copyWith((message) => updates(message as WaitingRoomRequest)) as WaitingRoomRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoomRequest create() => WaitingRoomRequest._();
  WaitingRoomRequest createEmptyInstance() => create();
  static $pb.PbList<WaitingRoomRequest> createRepeated() => $pb.PbList<WaitingRoomRequest>();
  @$core.pragma('dart2js:noInline')
  static WaitingRoomRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WaitingRoomRequest>(create);
  static WaitingRoomRequest? _defaultInstance;
}

class WaitingRoomResponse extends $pb.GeneratedMessage {
  factory WaitingRoomResponse({
    $core.Iterable<Player>? players,
  }) {
    final $result = create();
    if (players != null) {
      $result.players.addAll(players);
    }
    return $result;
  }
  WaitingRoomResponse._() : super();
  factory WaitingRoomResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory WaitingRoomResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'WaitingRoomResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..pc<Player>(1, _omitFieldNames ? '' : 'players', $pb.PbFieldType.PM, subBuilder: Player.create)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  WaitingRoomResponse clone() => WaitingRoomResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  WaitingRoomResponse copyWith(void Function(WaitingRoomResponse) updates) => super.copyWith((message) => updates(message as WaitingRoomResponse)) as WaitingRoomResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static WaitingRoomResponse create() => WaitingRoomResponse._();
  WaitingRoomResponse createEmptyInstance() => create();
  static $pb.PbList<WaitingRoomResponse> createRepeated() => $pb.PbList<WaitingRoomResponse>();
  @$core.pragma('dart2js:noInline')
  static WaitingRoomResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<WaitingRoomResponse>(create);
  static WaitingRoomResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<Player> get players => $_getList(0);
}

/// Game Messages
class Player extends $pb.GeneratedMessage {
  factory Player({
    $core.String? uid,
    $core.String? nick,
    $fixnum.Int64? betAmt,
    $core.int? number,
    $core.int? score,
    $core.bool? ready,
  }) {
    final $result = create();
    if (uid != null) {
      $result.uid = uid;
    }
    if (nick != null) {
      $result.nick = nick;
    }
    if (betAmt != null) {
      $result.betAmt = betAmt;
    }
    if (number != null) {
      $result.number = number;
    }
    if (score != null) {
      $result.score = score;
    }
    if (ready != null) {
      $result.ready = ready;
    }
    return $result;
  }
  Player._() : super();
  factory Player.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory Player.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'Player', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'uid')
    ..aOS(2, _omitFieldNames ? '' : 'nick')
    ..aInt64(3, _omitFieldNames ? '' : 'betAmt')
    ..a<$core.int>(4, _omitFieldNames ? '' : 'number', $pb.PbFieldType.O3)
    ..a<$core.int>(5, _omitFieldNames ? '' : 'score', $pb.PbFieldType.O3)
    ..aOB(6, _omitFieldNames ? '' : 'ready')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  Player clone() => Player()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  Player copyWith(void Function(Player) updates) => super.copyWith((message) => updates(message as Player)) as Player;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static Player create() => Player._();
  Player createEmptyInstance() => create();
  static $pb.PbList<Player> createRepeated() => $pb.PbList<Player>();
  @$core.pragma('dart2js:noInline')
  static Player getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<Player>(create);
  static Player? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get uid => $_getSZ(0);
  @$pb.TagNumber(1)
  set uid($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasUid() => $_has(0);
  @$pb.TagNumber(1)
  void clearUid() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get nick => $_getSZ(1);
  @$pb.TagNumber(2)
  set nick($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasNick() => $_has(1);
  @$pb.TagNumber(2)
  void clearNick() => clearField(2);

  @$pb.TagNumber(3)
  $fixnum.Int64 get betAmt => $_getI64(2);
  @$pb.TagNumber(3)
  set betAmt($fixnum.Int64 v) { $_setInt64(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasBetAmt() => $_has(2);
  @$pb.TagNumber(3)
  void clearBetAmt() => clearField(3);

  @$pb.TagNumber(4)
  $core.int get number => $_getIZ(3);
  @$pb.TagNumber(4)
  set number($core.int v) { $_setSignedInt32(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasNumber() => $_has(3);
  @$pb.TagNumber(4)
  void clearNumber() => clearField(4);

  @$pb.TagNumber(5)
  $core.int get score => $_getIZ(4);
  @$pb.TagNumber(5)
  set score($core.int v) { $_setSignedInt32(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasScore() => $_has(4);
  @$pb.TagNumber(5)
  void clearScore() => clearField(5);

  @$pb.TagNumber(6)
  $core.bool get ready => $_getBF(5);
  @$pb.TagNumber(6)
  set ready($core.bool v) { $_setBool(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasReady() => $_has(5);
  @$pb.TagNumber(6)
  void clearReady() => clearField(6);
}

/// SignalReadyRequest contains information about the client signaling readiness
class StartGameStreamRequest extends $pb.GeneratedMessage {
  factory StartGameStreamRequest({
    $core.String? clientId,
  }) {
    final $result = create();
    if (clientId != null) {
      $result.clientId = clientId;
    }
    return $result;
  }
  StartGameStreamRequest._() : super();
  factory StartGameStreamRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory StartGameStreamRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'StartGameStreamRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  StartGameStreamRequest clone() => StartGameStreamRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  StartGameStreamRequest copyWith(void Function(StartGameStreamRequest) updates) => super.copyWith((message) => updates(message as StartGameStreamRequest)) as StartGameStreamRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static StartGameStreamRequest create() => StartGameStreamRequest._();
  StartGameStreamRequest createEmptyInstance() => create();
  static $pb.PbList<StartGameStreamRequest> createRepeated() => $pb.PbList<StartGameStreamRequest>();
  @$core.pragma('dart2js:noInline')
  static StartGameStreamRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<StartGameStreamRequest>(create);
  static StartGameStreamRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);
}

class GameUpdateBytes extends $pb.GeneratedMessage {
  factory GameUpdateBytes({
    $core.List<$core.int>? data,
  }) {
    final $result = create();
    if (data != null) {
      $result.data = data;
    }
    return $result;
  }
  GameUpdateBytes._() : super();
  factory GameUpdateBytes.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GameUpdateBytes.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GameUpdateBytes', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..a<$core.List<$core.int>>(1, _omitFieldNames ? '' : 'data', $pb.PbFieldType.OY)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GameUpdateBytes clone() => GameUpdateBytes()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GameUpdateBytes copyWith(void Function(GameUpdateBytes) updates) => super.copyWith((message) => updates(message as GameUpdateBytes)) as GameUpdateBytes;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GameUpdateBytes create() => GameUpdateBytes._();
  GameUpdateBytes createEmptyInstance() => create();
  static $pb.PbList<GameUpdateBytes> createRepeated() => $pb.PbList<GameUpdateBytes>();
  @$core.pragma('dart2js:noInline')
  static GameUpdateBytes getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GameUpdateBytes>(create);
  static GameUpdateBytes? _defaultInstance;

  @$pb.TagNumber(1)
  $core.List<$core.int> get data => $_getN(0);
  @$pb.TagNumber(1)
  set data($core.List<$core.int> v) { $_setBytes(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasData() => $_has(0);
  @$pb.TagNumber(1)
  void clearData() => clearField(1);
}

class PlayerInput extends $pb.GeneratedMessage {
  factory PlayerInput({
    $core.String? playerId,
    $core.String? input,
    $core.int? playerNumber,
  }) {
    final $result = create();
    if (playerId != null) {
      $result.playerId = playerId;
    }
    if (input != null) {
      $result.input = input;
    }
    if (playerNumber != null) {
      $result.playerNumber = playerNumber;
    }
    return $result;
  }
  PlayerInput._() : super();
  factory PlayerInput.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory PlayerInput.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'PlayerInput', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'playerId')
    ..aOS(2, _omitFieldNames ? '' : 'input')
    ..a<$core.int>(3, _omitFieldNames ? '' : 'playerNumber', $pb.PbFieldType.O3)
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  PlayerInput clone() => PlayerInput()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  PlayerInput copyWith(void Function(PlayerInput) updates) => super.copyWith((message) => updates(message as PlayerInput)) as PlayerInput;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static PlayerInput create() => PlayerInput._();
  PlayerInput createEmptyInstance() => create();
  static $pb.PbList<PlayerInput> createRepeated() => $pb.PbList<PlayerInput>();
  @$core.pragma('dart2js:noInline')
  static PlayerInput getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<PlayerInput>(create);
  static PlayerInput? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get playerId => $_getSZ(0);
  @$pb.TagNumber(1)
  set playerId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasPlayerId() => $_has(0);
  @$pb.TagNumber(1)
  void clearPlayerId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get input => $_getSZ(1);
  @$pb.TagNumber(2)
  set input($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasInput() => $_has(1);
  @$pb.TagNumber(2)
  void clearInput() => clearField(2);

  @$pb.TagNumber(3)
  $core.int get playerNumber => $_getIZ(2);
  @$pb.TagNumber(3)
  set playerNumber($core.int v) { $_setSignedInt32(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasPlayerNumber() => $_has(2);
  @$pb.TagNumber(3)
  void clearPlayerNumber() => clearField(3);
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
    final $result = create();
    if (ballX != null) {
      $result.ballX = ballX;
    }
    if (ballY != null) {
      $result.ballY = ballY;
    }
    if (p1X != null) {
      $result.p1X = p1X;
    }
    if (p1Y != null) {
      $result.p1Y = p1Y;
    }
    if (p2X != null) {
      $result.p2X = p2X;
    }
    if (p2Y != null) {
      $result.p2Y = p2Y;
    }
    if (p1YVelocity != null) {
      $result.p1YVelocity = p1YVelocity;
    }
    if (p2YVelocity != null) {
      $result.p2YVelocity = p2YVelocity;
    }
    if (ballXVelocity != null) {
      $result.ballXVelocity = ballXVelocity;
    }
    if (ballYVelocity != null) {
      $result.ballYVelocity = ballYVelocity;
    }
    if (fps != null) {
      $result.fps = fps;
    }
    if (tps != null) {
      $result.tps = tps;
    }
    if (gameWidth != null) {
      $result.gameWidth = gameWidth;
    }
    if (gameHeight != null) {
      $result.gameHeight = gameHeight;
    }
    if (p1Width != null) {
      $result.p1Width = p1Width;
    }
    if (p1Height != null) {
      $result.p1Height = p1Height;
    }
    if (p2Width != null) {
      $result.p2Width = p2Width;
    }
    if (p2Height != null) {
      $result.p2Height = p2Height;
    }
    if (ballWidth != null) {
      $result.ballWidth = ballWidth;
    }
    if (ballHeight != null) {
      $result.ballHeight = ballHeight;
    }
    if (p1Score != null) {
      $result.p1Score = p1Score;
    }
    if (p2Score != null) {
      $result.p2Score = p2Score;
    }
    if (error != null) {
      $result.error = error;
    }
    if (debug != null) {
      $result.debug = debug;
    }
    return $result;
  }
  GameUpdate._() : super();
  factory GameUpdate.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory GameUpdate.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'GameUpdate', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..a<$core.double>(1, _omitFieldNames ? '' : 'ballX', $pb.PbFieldType.OD, protoName: 'ballX')
    ..a<$core.double>(2, _omitFieldNames ? '' : 'ballY', $pb.PbFieldType.OD, protoName: 'ballY')
    ..a<$core.double>(3, _omitFieldNames ? '' : 'p1X', $pb.PbFieldType.OD, protoName: 'p1X')
    ..a<$core.double>(4, _omitFieldNames ? '' : 'p1Y', $pb.PbFieldType.OD, protoName: 'p1Y')
    ..a<$core.double>(5, _omitFieldNames ? '' : 'p2X', $pb.PbFieldType.OD, protoName: 'p2X')
    ..a<$core.double>(6, _omitFieldNames ? '' : 'p2Y', $pb.PbFieldType.OD, protoName: 'p2Y')
    ..a<$core.double>(7, _omitFieldNames ? '' : 'p1YVelocity', $pb.PbFieldType.OD, protoName: 'p1YVelocity')
    ..a<$core.double>(8, _omitFieldNames ? '' : 'p2YVelocity', $pb.PbFieldType.OD, protoName: 'p2YVelocity')
    ..a<$core.double>(9, _omitFieldNames ? '' : 'ballXVelocity', $pb.PbFieldType.OD, protoName: 'ballXVelocity')
    ..a<$core.double>(10, _omitFieldNames ? '' : 'ballYVelocity', $pb.PbFieldType.OD, protoName: 'ballYVelocity')
    ..a<$core.double>(11, _omitFieldNames ? '' : 'fps', $pb.PbFieldType.OD)
    ..a<$core.double>(12, _omitFieldNames ? '' : 'tps', $pb.PbFieldType.OD)
    ..a<$core.double>(13, _omitFieldNames ? '' : 'gameWidth', $pb.PbFieldType.OD, protoName: 'gameWidth')
    ..a<$core.double>(14, _omitFieldNames ? '' : 'gameHeight', $pb.PbFieldType.OD, protoName: 'gameHeight')
    ..a<$core.double>(15, _omitFieldNames ? '' : 'p1Width', $pb.PbFieldType.OD, protoName: 'p1Width')
    ..a<$core.double>(16, _omitFieldNames ? '' : 'p1Height', $pb.PbFieldType.OD, protoName: 'p1Height')
    ..a<$core.double>(17, _omitFieldNames ? '' : 'p2Width', $pb.PbFieldType.OD, protoName: 'p2Width')
    ..a<$core.double>(18, _omitFieldNames ? '' : 'p2Height', $pb.PbFieldType.OD, protoName: 'p2Height')
    ..a<$core.double>(19, _omitFieldNames ? '' : 'ballWidth', $pb.PbFieldType.OD, protoName: 'ballWidth')
    ..a<$core.double>(20, _omitFieldNames ? '' : 'ballHeight', $pb.PbFieldType.OD, protoName: 'ballHeight')
    ..a<$core.int>(21, _omitFieldNames ? '' : 'p1Score', $pb.PbFieldType.O3, protoName: 'p1Score')
    ..a<$core.int>(22, _omitFieldNames ? '' : 'p2Score', $pb.PbFieldType.O3, protoName: 'p2Score')
    ..aOS(23, _omitFieldNames ? '' : 'error')
    ..aOB(24, _omitFieldNames ? '' : 'debug')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  GameUpdate clone() => GameUpdate()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  GameUpdate copyWith(void Function(GameUpdate) updates) => super.copyWith((message) => updates(message as GameUpdate)) as GameUpdate;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static GameUpdate create() => GameUpdate._();
  GameUpdate createEmptyInstance() => create();
  static $pb.PbList<GameUpdate> createRepeated() => $pb.PbList<GameUpdate>();
  @$core.pragma('dart2js:noInline')
  static GameUpdate getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<GameUpdate>(create);
  static GameUpdate? _defaultInstance;

  @$pb.TagNumber(1)
  $core.double get ballX => $_getN(0);
  @$pb.TagNumber(1)
  set ballX($core.double v) { $_setDouble(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasBallX() => $_has(0);
  @$pb.TagNumber(1)
  void clearBallX() => clearField(1);

  @$pb.TagNumber(2)
  $core.double get ballY => $_getN(1);
  @$pb.TagNumber(2)
  set ballY($core.double v) { $_setDouble(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasBallY() => $_has(1);
  @$pb.TagNumber(2)
  void clearBallY() => clearField(2);

  @$pb.TagNumber(3)
  $core.double get p1X => $_getN(2);
  @$pb.TagNumber(3)
  set p1X($core.double v) { $_setDouble(2, v); }
  @$pb.TagNumber(3)
  $core.bool hasP1X() => $_has(2);
  @$pb.TagNumber(3)
  void clearP1X() => clearField(3);

  @$pb.TagNumber(4)
  $core.double get p1Y => $_getN(3);
  @$pb.TagNumber(4)
  set p1Y($core.double v) { $_setDouble(3, v); }
  @$pb.TagNumber(4)
  $core.bool hasP1Y() => $_has(3);
  @$pb.TagNumber(4)
  void clearP1Y() => clearField(4);

  @$pb.TagNumber(5)
  $core.double get p2X => $_getN(4);
  @$pb.TagNumber(5)
  set p2X($core.double v) { $_setDouble(4, v); }
  @$pb.TagNumber(5)
  $core.bool hasP2X() => $_has(4);
  @$pb.TagNumber(5)
  void clearP2X() => clearField(5);

  @$pb.TagNumber(6)
  $core.double get p2Y => $_getN(5);
  @$pb.TagNumber(6)
  set p2Y($core.double v) { $_setDouble(5, v); }
  @$pb.TagNumber(6)
  $core.bool hasP2Y() => $_has(5);
  @$pb.TagNumber(6)
  void clearP2Y() => clearField(6);

  @$pb.TagNumber(7)
  $core.double get p1YVelocity => $_getN(6);
  @$pb.TagNumber(7)
  set p1YVelocity($core.double v) { $_setDouble(6, v); }
  @$pb.TagNumber(7)
  $core.bool hasP1YVelocity() => $_has(6);
  @$pb.TagNumber(7)
  void clearP1YVelocity() => clearField(7);

  @$pb.TagNumber(8)
  $core.double get p2YVelocity => $_getN(7);
  @$pb.TagNumber(8)
  set p2YVelocity($core.double v) { $_setDouble(7, v); }
  @$pb.TagNumber(8)
  $core.bool hasP2YVelocity() => $_has(7);
  @$pb.TagNumber(8)
  void clearP2YVelocity() => clearField(8);

  @$pb.TagNumber(9)
  $core.double get ballXVelocity => $_getN(8);
  @$pb.TagNumber(9)
  set ballXVelocity($core.double v) { $_setDouble(8, v); }
  @$pb.TagNumber(9)
  $core.bool hasBallXVelocity() => $_has(8);
  @$pb.TagNumber(9)
  void clearBallXVelocity() => clearField(9);

  @$pb.TagNumber(10)
  $core.double get ballYVelocity => $_getN(9);
  @$pb.TagNumber(10)
  set ballYVelocity($core.double v) { $_setDouble(9, v); }
  @$pb.TagNumber(10)
  $core.bool hasBallYVelocity() => $_has(9);
  @$pb.TagNumber(10)
  void clearBallYVelocity() => clearField(10);

  @$pb.TagNumber(11)
  $core.double get fps => $_getN(10);
  @$pb.TagNumber(11)
  set fps($core.double v) { $_setDouble(10, v); }
  @$pb.TagNumber(11)
  $core.bool hasFps() => $_has(10);
  @$pb.TagNumber(11)
  void clearFps() => clearField(11);

  @$pb.TagNumber(12)
  $core.double get tps => $_getN(11);
  @$pb.TagNumber(12)
  set tps($core.double v) { $_setDouble(11, v); }
  @$pb.TagNumber(12)
  $core.bool hasTps() => $_has(11);
  @$pb.TagNumber(12)
  void clearTps() => clearField(12);

  @$pb.TagNumber(13)
  $core.double get gameWidth => $_getN(12);
  @$pb.TagNumber(13)
  set gameWidth($core.double v) { $_setDouble(12, v); }
  @$pb.TagNumber(13)
  $core.bool hasGameWidth() => $_has(12);
  @$pb.TagNumber(13)
  void clearGameWidth() => clearField(13);

  @$pb.TagNumber(14)
  $core.double get gameHeight => $_getN(13);
  @$pb.TagNumber(14)
  set gameHeight($core.double v) { $_setDouble(13, v); }
  @$pb.TagNumber(14)
  $core.bool hasGameHeight() => $_has(13);
  @$pb.TagNumber(14)
  void clearGameHeight() => clearField(14);

  @$pb.TagNumber(15)
  $core.double get p1Width => $_getN(14);
  @$pb.TagNumber(15)
  set p1Width($core.double v) { $_setDouble(14, v); }
  @$pb.TagNumber(15)
  $core.bool hasP1Width() => $_has(14);
  @$pb.TagNumber(15)
  void clearP1Width() => clearField(15);

  @$pb.TagNumber(16)
  $core.double get p1Height => $_getN(15);
  @$pb.TagNumber(16)
  set p1Height($core.double v) { $_setDouble(15, v); }
  @$pb.TagNumber(16)
  $core.bool hasP1Height() => $_has(15);
  @$pb.TagNumber(16)
  void clearP1Height() => clearField(16);

  @$pb.TagNumber(17)
  $core.double get p2Width => $_getN(16);
  @$pb.TagNumber(17)
  set p2Width($core.double v) { $_setDouble(16, v); }
  @$pb.TagNumber(17)
  $core.bool hasP2Width() => $_has(16);
  @$pb.TagNumber(17)
  void clearP2Width() => clearField(17);

  @$pb.TagNumber(18)
  $core.double get p2Height => $_getN(17);
  @$pb.TagNumber(18)
  set p2Height($core.double v) { $_setDouble(17, v); }
  @$pb.TagNumber(18)
  $core.bool hasP2Height() => $_has(17);
  @$pb.TagNumber(18)
  void clearP2Height() => clearField(18);

  @$pb.TagNumber(19)
  $core.double get ballWidth => $_getN(18);
  @$pb.TagNumber(19)
  set ballWidth($core.double v) { $_setDouble(18, v); }
  @$pb.TagNumber(19)
  $core.bool hasBallWidth() => $_has(18);
  @$pb.TagNumber(19)
  void clearBallWidth() => clearField(19);

  @$pb.TagNumber(20)
  $core.double get ballHeight => $_getN(19);
  @$pb.TagNumber(20)
  set ballHeight($core.double v) { $_setDouble(19, v); }
  @$pb.TagNumber(20)
  $core.bool hasBallHeight() => $_has(19);
  @$pb.TagNumber(20)
  void clearBallHeight() => clearField(20);

  @$pb.TagNumber(21)
  $core.int get p1Score => $_getIZ(20);
  @$pb.TagNumber(21)
  set p1Score($core.int v) { $_setSignedInt32(20, v); }
  @$pb.TagNumber(21)
  $core.bool hasP1Score() => $_has(20);
  @$pb.TagNumber(21)
  void clearP1Score() => clearField(21);

  @$pb.TagNumber(22)
  $core.int get p2Score => $_getIZ(21);
  @$pb.TagNumber(22)
  set p2Score($core.int v) { $_setSignedInt32(21, v); }
  @$pb.TagNumber(22)
  $core.bool hasP2Score() => $_has(21);
  @$pb.TagNumber(22)
  void clearP2Score() => clearField(22);

  /// Optional: if you want to send error messages or debug information
  @$pb.TagNumber(23)
  $core.String get error => $_getSZ(22);
  @$pb.TagNumber(23)
  set error($core.String v) { $_setString(22, v); }
  @$pb.TagNumber(23)
  $core.bool hasError() => $_has(22);
  @$pb.TagNumber(23)
  void clearError() => clearField(23);

  @$pb.TagNumber(24)
  $core.bool get debug => $_getBF(23);
  @$pb.TagNumber(24)
  set debug($core.bool v) { $_setBool(23, v); }
  @$pb.TagNumber(24)
  $core.bool hasDebug() => $_has(23);
  @$pb.TagNumber(24)
  void clearDebug() => clearField(24);
}

class LeaveWaitingRoomRequest extends $pb.GeneratedMessage {
  factory LeaveWaitingRoomRequest({
    $core.String? clientId,
    $core.String? roomId,
  }) {
    final $result = create();
    if (clientId != null) {
      $result.clientId = clientId;
    }
    if (roomId != null) {
      $result.roomId = roomId;
    }
    return $result;
  }
  LeaveWaitingRoomRequest._() : super();
  factory LeaveWaitingRoomRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LeaveWaitingRoomRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'LeaveWaitingRoomRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..aOS(2, _omitFieldNames ? '' : 'roomId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LeaveWaitingRoomRequest clone() => LeaveWaitingRoomRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LeaveWaitingRoomRequest copyWith(void Function(LeaveWaitingRoomRequest) updates) => super.copyWith((message) => updates(message as LeaveWaitingRoomRequest)) as LeaveWaitingRoomRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LeaveWaitingRoomRequest create() => LeaveWaitingRoomRequest._();
  LeaveWaitingRoomRequest createEmptyInstance() => create();
  static $pb.PbList<LeaveWaitingRoomRequest> createRepeated() => $pb.PbList<LeaveWaitingRoomRequest>();
  @$core.pragma('dart2js:noInline')
  static LeaveWaitingRoomRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LeaveWaitingRoomRequest>(create);
  static LeaveWaitingRoomRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get roomId => $_getSZ(1);
  @$pb.TagNumber(2)
  set roomId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasRoomId() => $_has(1);
  @$pb.TagNumber(2)
  void clearRoomId() => clearField(2);
}

class LeaveWaitingRoomResponse extends $pb.GeneratedMessage {
  factory LeaveWaitingRoomResponse({
    $core.bool? success,
    $core.String? message,
  }) {
    final $result = create();
    if (success != null) {
      $result.success = success;
    }
    if (message != null) {
      $result.message = message;
    }
    return $result;
  }
  LeaveWaitingRoomResponse._() : super();
  factory LeaveWaitingRoomResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory LeaveWaitingRoomResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'LeaveWaitingRoomResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'success')
    ..aOS(2, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  LeaveWaitingRoomResponse clone() => LeaveWaitingRoomResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  LeaveWaitingRoomResponse copyWith(void Function(LeaveWaitingRoomResponse) updates) => super.copyWith((message) => updates(message as LeaveWaitingRoomResponse)) as LeaveWaitingRoomResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static LeaveWaitingRoomResponse create() => LeaveWaitingRoomResponse._();
  LeaveWaitingRoomResponse createEmptyInstance() => create();
  static $pb.PbList<LeaveWaitingRoomResponse> createRepeated() => $pb.PbList<LeaveWaitingRoomResponse>();
  @$core.pragma('dart2js:noInline')
  static LeaveWaitingRoomResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<LeaveWaitingRoomResponse>(create);
  static LeaveWaitingRoomResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get success => $_getBF(0);
  @$pb.TagNumber(1)
  set success($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSuccess() => $_has(0);
  @$pb.TagNumber(1)
  void clearSuccess() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get message => $_getSZ(1);
  @$pb.TagNumber(2)
  set message($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasMessage() => $_has(1);
  @$pb.TagNumber(2)
  void clearMessage() => clearField(2);
}

/// SignalReadyToPlayRequest contains information about the client signaling readiness
class SignalReadyToPlayRequest extends $pb.GeneratedMessage {
  factory SignalReadyToPlayRequest({
    $core.String? clientId,
    $core.String? gameId,
  }) {
    final $result = create();
    if (clientId != null) {
      $result.clientId = clientId;
    }
    if (gameId != null) {
      $result.gameId = gameId;
    }
    return $result;
  }
  SignalReadyToPlayRequest._() : super();
  factory SignalReadyToPlayRequest.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SignalReadyToPlayRequest.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SignalReadyToPlayRequest', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOS(1, _omitFieldNames ? '' : 'clientId')
    ..aOS(2, _omitFieldNames ? '' : 'gameId')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SignalReadyToPlayRequest clone() => SignalReadyToPlayRequest()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SignalReadyToPlayRequest copyWith(void Function(SignalReadyToPlayRequest) updates) => super.copyWith((message) => updates(message as SignalReadyToPlayRequest)) as SignalReadyToPlayRequest;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SignalReadyToPlayRequest create() => SignalReadyToPlayRequest._();
  SignalReadyToPlayRequest createEmptyInstance() => create();
  static $pb.PbList<SignalReadyToPlayRequest> createRepeated() => $pb.PbList<SignalReadyToPlayRequest>();
  @$core.pragma('dart2js:noInline')
  static SignalReadyToPlayRequest getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SignalReadyToPlayRequest>(create);
  static SignalReadyToPlayRequest? _defaultInstance;

  @$pb.TagNumber(1)
  $core.String get clientId => $_getSZ(0);
  @$pb.TagNumber(1)
  set clientId($core.String v) { $_setString(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasClientId() => $_has(0);
  @$pb.TagNumber(1)
  void clearClientId() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get gameId => $_getSZ(1);
  @$pb.TagNumber(2)
  set gameId($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasGameId() => $_has(1);
  @$pb.TagNumber(2)
  void clearGameId() => clearField(2);
}

/// SignalReadyToPlayResponse contains the result of the ready signal
class SignalReadyToPlayResponse extends $pb.GeneratedMessage {
  factory SignalReadyToPlayResponse({
    $core.bool? success,
    $core.String? message,
  }) {
    final $result = create();
    if (success != null) {
      $result.success = success;
    }
    if (message != null) {
      $result.message = message;
    }
    return $result;
  }
  SignalReadyToPlayResponse._() : super();
  factory SignalReadyToPlayResponse.fromBuffer($core.List<$core.int> i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromBuffer(i, r);
  factory SignalReadyToPlayResponse.fromJson($core.String i, [$pb.ExtensionRegistry r = $pb.ExtensionRegistry.EMPTY]) => create()..mergeFromJson(i, r);

  static final $pb.BuilderInfo _i = $pb.BuilderInfo(_omitMessageNames ? '' : 'SignalReadyToPlayResponse', package: const $pb.PackageName(_omitMessageNames ? '' : 'pong'), createEmptyInstance: create)
    ..aOB(1, _omitFieldNames ? '' : 'success')
    ..aOS(2, _omitFieldNames ? '' : 'message')
    ..hasRequiredFields = false
  ;

  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.deepCopy] instead. '
  'Will be removed in next major version')
  SignalReadyToPlayResponse clone() => SignalReadyToPlayResponse()..mergeFromMessage(this);
  @$core.Deprecated(
  'Using this can add significant overhead to your binary. '
  'Use [GeneratedMessageGenericExtensions.rebuild] instead. '
  'Will be removed in next major version')
  SignalReadyToPlayResponse copyWith(void Function(SignalReadyToPlayResponse) updates) => super.copyWith((message) => updates(message as SignalReadyToPlayResponse)) as SignalReadyToPlayResponse;

  $pb.BuilderInfo get info_ => _i;

  @$core.pragma('dart2js:noInline')
  static SignalReadyToPlayResponse create() => SignalReadyToPlayResponse._();
  SignalReadyToPlayResponse createEmptyInstance() => create();
  static $pb.PbList<SignalReadyToPlayResponse> createRepeated() => $pb.PbList<SignalReadyToPlayResponse>();
  @$core.pragma('dart2js:noInline')
  static SignalReadyToPlayResponse getDefault() => _defaultInstance ??= $pb.GeneratedMessage.$_defaultFor<SignalReadyToPlayResponse>(create);
  static SignalReadyToPlayResponse? _defaultInstance;

  @$pb.TagNumber(1)
  $core.bool get success => $_getBF(0);
  @$pb.TagNumber(1)
  set success($core.bool v) { $_setBool(0, v); }
  @$pb.TagNumber(1)
  $core.bool hasSuccess() => $_has(0);
  @$pb.TagNumber(1)
  void clearSuccess() => clearField(1);

  @$pb.TagNumber(2)
  $core.String get message => $_getSZ(1);
  @$pb.TagNumber(2)
  set message($core.String v) { $_setString(1, v); }
  @$pb.TagNumber(2)
  $core.bool hasMessage() => $_has(1);
  @$pb.TagNumber(2)
  void clearMessage() => clearField(2);
}


const _omitFieldNames = $core.bool.fromEnvironment('protobuf.omit_field_names');
const _omitMessageNames = $core.bool.fromEnvironment('protobuf.omit_message_names');
