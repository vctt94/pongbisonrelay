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

import 'package:protobuf/protobuf.dart' as $pb;

class Branch extends $pb.ProtobufEnum {
  static const Branch BRANCH_UNSPECIFIED = Branch._(0, _omitEnumNames ? '' : 'BRANCH_UNSPECIFIED');
  static const Branch BRANCH_A = Branch._(1, _omitEnumNames ? '' : 'BRANCH_A');
  static const Branch BRANCH_B = Branch._(2, _omitEnumNames ? '' : 'BRANCH_B');

  static const $core.List<Branch> values = <Branch> [
    BRANCH_UNSPECIFIED,
    BRANCH_A,
    BRANCH_B,
  ];

  static final $core.Map<$core.int, Branch> _byValue = $pb.ProtobufEnum.initByValue(values);
  static Branch? valueOf($core.int value) => _byValue[value];

  const Branch._($core.int v, $core.String n) : super(v, n);
}

/// Notification Messages
class NotificationType extends $pb.ProtobufEnum {
  static const NotificationType UNKNOWN = NotificationType._(0, _omitEnumNames ? '' : 'UNKNOWN');
  static const NotificationType MESSAGE = NotificationType._(1, _omitEnumNames ? '' : 'MESSAGE');
  static const NotificationType GAME_START = NotificationType._(2, _omitEnumNames ? '' : 'GAME_START');
  static const NotificationType GAME_END = NotificationType._(3, _omitEnumNames ? '' : 'GAME_END');
  static const NotificationType OPPONENT_DISCONNECTED = NotificationType._(4, _omitEnumNames ? '' : 'OPPONENT_DISCONNECTED');
  static const NotificationType BET_AMOUNT_UPDATE = NotificationType._(5, _omitEnumNames ? '' : 'BET_AMOUNT_UPDATE');
  static const NotificationType PLAYER_JOINED_WR = NotificationType._(6, _omitEnumNames ? '' : 'PLAYER_JOINED_WR');
  static const NotificationType ON_WR_CREATED = NotificationType._(7, _omitEnumNames ? '' : 'ON_WR_CREATED');
  static const NotificationType ON_PLAYER_READY = NotificationType._(8, _omitEnumNames ? '' : 'ON_PLAYER_READY');
  static const NotificationType ON_WR_REMOVED = NotificationType._(9, _omitEnumNames ? '' : 'ON_WR_REMOVED');
  static const NotificationType PLAYER_LEFT_WR = NotificationType._(10, _omitEnumNames ? '' : 'PLAYER_LEFT_WR');
  static const NotificationType COUNTDOWN_UPDATE = NotificationType._(11, _omitEnumNames ? '' : 'COUNTDOWN_UPDATE');
  static const NotificationType GAME_READY_TO_PLAY = NotificationType._(12, _omitEnumNames ? '' : 'GAME_READY_TO_PLAY');

  static const $core.List<NotificationType> values = <NotificationType> [
    UNKNOWN,
    MESSAGE,
    GAME_START,
    GAME_END,
    OPPONENT_DISCONNECTED,
    BET_AMOUNT_UPDATE,
    PLAYER_JOINED_WR,
    ON_WR_CREATED,
    ON_PLAYER_READY,
    ON_WR_REMOVED,
    PLAYER_LEFT_WR,
    COUNTDOWN_UPDATE,
    GAME_READY_TO_PLAY,
  ];

  static final $core.Map<$core.int, NotificationType> _byValue = $pb.ProtobufEnum.initByValue(values);
  static NotificationType? valueOf($core.int value) => _byValue[value];

  const NotificationType._($core.int v, $core.String n) : super(v, n);
}

class AssignRole_Role extends $pb.ProtobufEnum {
  static const AssignRole_Role UNKNOWN = AssignRole_Role._(0, _omitEnumNames ? '' : 'UNKNOWN');
  static const AssignRole_Role A = AssignRole_Role._(1, _omitEnumNames ? '' : 'A');
  static const AssignRole_Role B = AssignRole_Role._(2, _omitEnumNames ? '' : 'B');

  static const $core.List<AssignRole_Role> values = <AssignRole_Role> [
    UNKNOWN,
    A,
    B,
  ];

  static final $core.Map<$core.int, AssignRole_Role> _byValue = $pb.ProtobufEnum.initByValue(values);
  static AssignRole_Role? valueOf($core.int value) => _byValue[value];

  const AssignRole_Role._($core.int v, $core.String n) : super(v, n);
}


const _omitEnumNames = $core.bool.fromEnvironment('protobuf.omit_enum_names');
