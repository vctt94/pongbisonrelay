//
//  Generated code. Do not modify.
//  source: pong.proto
//
// @dart = 2.12

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names, library_prefixes
// ignore_for_file: non_constant_identifier_names, prefer_final_fields
// ignore_for_file: unnecessary_import, unnecessary_this, unused_import

import 'dart:convert' as $convert;
import 'dart:core' as $core;
import 'dart:typed_data' as $typed_data;

@$core.Deprecated('Use notificationTypeDescriptor instead')
const NotificationType$json = {
  '1': 'NotificationType',
  '2': [
    {'1': 'UNKNOWN', '2': 0},
    {'1': 'MESSAGE', '2': 1},
    {'1': 'GAME_START', '2': 2},
    {'1': 'GAME_END', '2': 3},
    {'1': 'OPPONENT_DISCONNECTED', '2': 4},
    {'1': 'BET_AMOUNT_UPDATE', '2': 5},
    {'1': 'PLAYER_JOINED_WR', '2': 6},
    {'1': 'ON_WR_CREATED', '2': 7},
    {'1': 'ON_PLAYER_READY', '2': 8},
    {'1': 'ON_WR_REMOVED', '2': 9},
    {'1': 'PLAYER_LEFT_WR', '2': 10},
    {'1': 'COUNTDOWN_UPDATE', '2': 11},
    {'1': 'GAME_READY_TO_PLAY', '2': 12},
    {'1': 'MATCH_ALLOCATED', '2': 13},
  ],
};

/// Descriptor for `NotificationType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List notificationTypeDescriptor = $convert.base64Decode(
    'ChBOb3RpZmljYXRpb25UeXBlEgsKB1VOS05PV04QABILCgdNRVNTQUdFEAESDgoKR0FNRV9TVE'
    'FSVBACEgwKCEdBTUVfRU5EEAMSGQoVT1BQT05FTlRfRElTQ09OTkVDVEVEEAQSFQoRQkVUX0FN'
    'T1VOVF9VUERBVEUQBRIUChBQTEFZRVJfSk9JTkVEX1dSEAYSEQoNT05fV1JfQ1JFQVRFRBAHEh'
    'MKD09OX1BMQVlFUl9SRUFEWRAIEhEKDU9OX1dSX1JFTU9WRUQQCRISCg5QTEFZRVJfTEVGVF9X'
    'UhAKEhQKEENPVU5URE9XTl9VUERBVEUQCxIWChJHQU1FX1JFQURZX1RPX1BMQVkQDBITCg9NQV'
    'RDSF9BTExPQ0FURUQQDQ==');

@$core.Deprecated('Use clientMsgDescriptor instead')
const ClientMsg$json = {
  '1': 'ClientMsg',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'hello', '3': 10, '4': 1, '5': 11, '6': '.pong.Hello', '9': 0, '10': 'hello'},
    {'1': 'ack', '3': 12, '4': 1, '5': 11, '6': '.pong.Ack', '9': 0, '10': 'ack'},
    {'1': 'verify_ok', '3': 13, '4': 1, '5': 11, '6': '.pong.VerifyOk', '9': 0, '10': 'verifyOk'},
  ],
  '8': [
    {'1': 'kind'},
  ],
};

/// Descriptor for `ClientMsg`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List clientMsgDescriptor = $convert.base64Decode(
    'CglDbGllbnRNc2cSGQoIbWF0Y2hfaWQYASABKAlSB21hdGNoSWQSIwoFaGVsbG8YCiABKAsyCy'
    '5wb25nLkhlbGxvSABSBWhlbGxvEh0KA2FjaxgMIAEoCzIJLnBvbmcuQWNrSABSA2FjaxItCgl2'
    'ZXJpZnlfb2sYDSABKAsyDi5wb25nLlZlcmlmeU9rSABSCHZlcmlmeU9rQgYKBGtpbmQ=');

@$core.Deprecated('Use serverMsgDescriptor instead')
const ServerMsg$json = {
  '1': 'ServerMsg',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'req', '3': 11, '4': 1, '5': 11, '6': '.pong.NeedPreSigs', '9': 0, '10': 'req'},
    {'1': 'info', '3': 13, '4': 1, '5': 11, '6': '.pong.Info', '9': 0, '10': 'info'},
    {'1': 'ok', '3': 14, '4': 1, '5': 11, '6': '.pong.ServerOk', '9': 0, '10': 'ok'},
  ],
  '8': [
    {'1': 'kind'},
  ],
};

/// Descriptor for `ServerMsg`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serverMsgDescriptor = $convert.base64Decode(
    'CglTZXJ2ZXJNc2cSGQoIbWF0Y2hfaWQYASABKAlSB21hdGNoSWQSJQoDcmVxGAsgASgLMhEucG'
    '9uZy5OZWVkUHJlU2lnc0gAUgNyZXESIAoEaW5mbxgNIAEoCzIKLnBvbmcuSW5mb0gAUgRpbmZv'
    'EiAKAm9rGA4gASgLMg4ucG9uZy5TZXJ2ZXJPa0gAUgJva0IGCgRraW5k');

@$core.Deprecated('Use helloDescriptor instead')
const Hello$json = {
  '1': 'Hello',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'comp_pubkey', '3': 2, '4': 1, '5': 12, '10': 'compPubkey'},
    {'1': 'client_version', '3': 3, '4': 1, '5': 9, '10': 'clientVersion'},
  ],
};

/// Descriptor for `Hello`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List helloDescriptor = $convert.base64Decode(
    'CgVIZWxsbxIZCghtYXRjaF9pZBgBIAEoCVIHbWF0Y2hJZBIfCgtjb21wX3B1YmtleRgCIAEoDF'
    'IKY29tcFB1YmtleRIlCg5jbGllbnRfdmVyc2lvbhgDIAEoCVINY2xpZW50VmVyc2lvbg==');

@$core.Deprecated('Use needPreSigsDescriptor instead')
const NeedPreSigs$json = {
  '1': 'NeedPreSigs',
  '2': [
    {'1': 'draft_tx_hex', '3': 2, '4': 1, '5': 9, '10': 'draftTxHex'},
    {'1': 'inputs', '3': 4, '4': 3, '5': 11, '6': '.pong.NeedPreSigs.PerInput', '10': 'inputs'},
  ],
  '3': [NeedPreSigs_PerInput$json],
};

@$core.Deprecated('Use needPreSigsDescriptor instead')
const NeedPreSigs_PerInput$json = {
  '1': 'PerInput',
  '2': [
    {'1': 'input_id', '3': 1, '4': 1, '5': 9, '10': 'inputId'},
    {'1': 'redeem_script_hex', '3': 2, '4': 1, '5': 9, '10': 'redeemScriptHex'},
    {'1': 'm_hex', '3': 3, '4': 1, '5': 9, '10': 'mHex'},
    {'1': 'T_compressed', '3': 4, '4': 1, '5': 12, '10': 'TCompressed'},
  ],
};

/// Descriptor for `NeedPreSigs`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List needPreSigsDescriptor = $convert.base64Decode(
    'CgtOZWVkUHJlU2lncxIgCgxkcmFmdF90eF9oZXgYAiABKAlSCmRyYWZ0VHhIZXgSMgoGaW5wdX'
    'RzGAQgAygLMhoucG9uZy5OZWVkUHJlU2lncy5QZXJJbnB1dFIGaW5wdXRzGokBCghQZXJJbnB1'
    'dBIZCghpbnB1dF9pZBgBIAEoCVIHaW5wdXRJZBIqChFyZWRlZW1fc2NyaXB0X2hleBgCIAEoCV'
    'IPcmVkZWVtU2NyaXB0SGV4EhMKBW1faGV4GAMgASgJUgRtSGV4EiEKDFRfY29tcHJlc3NlZBgE'
    'IAEoDFILVENvbXByZXNzZWQ=');

@$core.Deprecated('Use verifyOkDescriptor instead')
const VerifyOk$json = {
  '1': 'VerifyOk',
  '2': [
    {'1': 'ack_digest', '3': 1, '4': 1, '5': 12, '10': 'ackDigest'},
    {'1': 'presigs', '3': 2, '4': 3, '5': 11, '6': '.pong.PreSig', '10': 'presigs'},
  ],
};

/// Descriptor for `VerifyOk`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List verifyOkDescriptor = $convert.base64Decode(
    'CghWZXJpZnlPaxIdCgphY2tfZGlnZXN0GAEgASgMUglhY2tEaWdlc3QSJgoHcHJlc2lncxgCIA'
    'MoCzIMLnBvbmcuUHJlU2lnUgdwcmVzaWdz');

@$core.Deprecated('Use preSigDescriptor instead')
const PreSig$json = {
  '1': 'PreSig',
  '2': [
    {'1': 'input_id', '3': 1, '4': 1, '5': 9, '10': 'inputId'},
    {'1': 'RLine_compressed', '3': 2, '4': 1, '5': 12, '10': 'RLineCompressed'},
    {'1': 'sLine32', '3': 3, '4': 1, '5': 12, '10': 'sLine32'},
  ],
};

/// Descriptor for `PreSig`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List preSigDescriptor = $convert.base64Decode(
    'CgZQcmVTaWcSGQoIaW5wdXRfaWQYASABKAlSB2lucHV0SWQSKQoQUkxpbmVfY29tcHJlc3NlZB'
    'gCIAEoDFIPUkxpbmVDb21wcmVzc2VkEhgKB3NMaW5lMzIYAyABKAxSB3NMaW5lMzI=');

@$core.Deprecated('Use ackDescriptor instead')
const Ack$json = {
  '1': 'Ack',
  '2': [
    {'1': 'note', '3': 1, '4': 1, '5': 9, '10': 'note'},
  ],
};

/// Descriptor for `Ack`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List ackDescriptor = $convert.base64Decode(
    'CgNBY2sSEgoEbm90ZRgBIAEoCVIEbm90ZQ==');

@$core.Deprecated('Use infoDescriptor instead')
const Info$json = {
  '1': 'Info',
  '2': [
    {'1': 'text', '3': 1, '4': 1, '5': 9, '10': 'text'},
  ],
};

/// Descriptor for `Info`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List infoDescriptor = $convert.base64Decode(
    'CgRJbmZvEhIKBHRleHQYASABKAlSBHRleHQ=');

@$core.Deprecated('Use serverOkDescriptor instead')
const ServerOk$json = {
  '1': 'ServerOk',
  '2': [
    {'1': 'ack_digest', '3': 1, '4': 1, '5': 12, '10': 'ackDigest'},
  ],
};

/// Descriptor for `ServerOk`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serverOkDescriptor = $convert.base64Decode(
    'CghTZXJ2ZXJPaxIdCgphY2tfZGlnZXN0GAEgASgMUglhY2tEaWdlc3Q=');

@$core.Deprecated('Use getFinalizeBundleRequestDescriptor instead')
const GetFinalizeBundleRequest$json = {
  '1': 'GetFinalizeBundleRequest',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'winner_uid', '3': 2, '4': 1, '5': 9, '10': 'winnerUid'},
  ],
};

/// Descriptor for `GetFinalizeBundleRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getFinalizeBundleRequestDescriptor = $convert.base64Decode(
    'ChhHZXRGaW5hbGl6ZUJ1bmRsZVJlcXVlc3QSGQoIbWF0Y2hfaWQYASABKAlSB21hdGNoSWQSHQ'
    'oKd2lubmVyX3VpZBgCIAEoCVIJd2lubmVyVWlk');

@$core.Deprecated('Use finalizeInputDescriptor instead')
const FinalizeInput$json = {
  '1': 'FinalizeInput',
  '2': [
    {'1': 'input_id', '3': 1, '4': 1, '5': 9, '10': 'inputId'},
    {'1': 'redeem_script_hex', '3': 2, '4': 1, '5': 9, '10': 'redeemScriptHex'},
    {'1': 'RLine_compressed', '3': 3, '4': 1, '5': 12, '10': 'RLineCompressed'},
    {'1': 'sLine32', '3': 4, '4': 1, '5': 12, '10': 'sLine32'},
  ],
};

/// Descriptor for `FinalizeInput`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List finalizeInputDescriptor = $convert.base64Decode(
    'Cg1GaW5hbGl6ZUlucHV0EhkKCGlucHV0X2lkGAEgASgJUgdpbnB1dElkEioKEXJlZGVlbV9zY3'
    'JpcHRfaGV4GAIgASgJUg9yZWRlZW1TY3JpcHRIZXgSKQoQUkxpbmVfY29tcHJlc3NlZBgDIAEo'
    'DFIPUkxpbmVDb21wcmVzc2VkEhgKB3NMaW5lMzIYBCABKAxSB3NMaW5lMzI=');

@$core.Deprecated('Use getFinalizeBundleResponseDescriptor instead')
const GetFinalizeBundleResponse$json = {
  '1': 'GetFinalizeBundleResponse',
  '2': [
    {'1': 'draft_tx_hex', '3': 1, '4': 1, '5': 9, '10': 'draftTxHex'},
    {'1': 'gamma32', '3': 2, '4': 1, '5': 12, '10': 'gamma32'},
    {'1': 'inputs', '3': 3, '4': 3, '5': 11, '6': '.pong.FinalizeInput', '10': 'inputs'},
  ],
};

/// Descriptor for `GetFinalizeBundleResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List getFinalizeBundleResponseDescriptor = $convert.base64Decode(
    'ChlHZXRGaW5hbGl6ZUJ1bmRsZVJlc3BvbnNlEiAKDGRyYWZ0X3R4X2hleBgBIAEoCVIKZHJhZn'
    'RUeEhleBIYCgdnYW1tYTMyGAIgASgMUgdnYW1tYTMyEisKBmlucHV0cxgDIAMoCzITLnBvbmcu'
    'RmluYWxpemVJbnB1dFIGaW5wdXRz');

@$core.Deprecated('Use openEscrowRequestDescriptor instead')
const OpenEscrowRequest$json = {
  '1': 'OpenEscrowRequest',
  '2': [
    {'1': 'owner_uid', '3': 1, '4': 1, '5': 9, '10': 'ownerUid'},
    {'1': 'comp_pubkey', '3': 2, '4': 1, '5': 12, '10': 'compPubkey'},
    {'1': 'bet_atoms', '3': 3, '4': 1, '5': 4, '10': 'betAtoms'},
    {'1': 'csv_blocks', '3': 4, '4': 1, '5': 13, '10': 'csvBlocks'},
    {'1': 'payout_pubkey', '3': 5, '4': 1, '5': 12, '10': 'payoutPubkey'},
  ],
};

/// Descriptor for `OpenEscrowRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List openEscrowRequestDescriptor = $convert.base64Decode(
    'ChFPcGVuRXNjcm93UmVxdWVzdBIbCglvd25lcl91aWQYASABKAlSCG93bmVyVWlkEh8KC2NvbX'
    'BfcHVia2V5GAIgASgMUgpjb21wUHVia2V5EhsKCWJldF9hdG9tcxgDIAEoBFIIYmV0QXRvbXMS'
    'HQoKY3N2X2Jsb2NrcxgEIAEoDVIJY3N2QmxvY2tzEiMKDXBheW91dF9wdWJrZXkYBSABKAxSDH'
    'BheW91dFB1YmtleQ==');

@$core.Deprecated('Use openEscrowResponseDescriptor instead')
const OpenEscrowResponse$json = {
  '1': 'OpenEscrowResponse',
  '2': [
    {'1': 'escrow_id', '3': 1, '4': 1, '5': 9, '10': 'escrowId'},
    {'1': 'deposit_address', '3': 2, '4': 1, '5': 9, '10': 'depositAddress'},
    {'1': 'pk_script_hex', '3': 3, '4': 1, '5': 9, '10': 'pkScriptHex'},
  ],
};

/// Descriptor for `OpenEscrowResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List openEscrowResponseDescriptor = $convert.base64Decode(
    'ChJPcGVuRXNjcm93UmVzcG9uc2USGwoJZXNjcm93X2lkGAEgASgJUghlc2Nyb3dJZBInCg9kZX'
    'Bvc2l0X2FkZHJlc3MYAiABKAlSDmRlcG9zaXRBZGRyZXNzEiIKDXBrX3NjcmlwdF9oZXgYAyAB'
    'KAlSC3BrU2NyaXB0SGV4');

@$core.Deprecated('Use escrowUTXODescriptor instead')
const EscrowUTXO$json = {
  '1': 'EscrowUTXO',
  '2': [
    {'1': 'txid', '3': 1, '4': 1, '5': 9, '10': 'txid'},
    {'1': 'vout', '3': 2, '4': 1, '5': 13, '10': 'vout'},
    {'1': 'value', '3': 3, '4': 1, '5': 4, '10': 'value'},
    {'1': 'redeem_script_hex', '3': 4, '4': 1, '5': 9, '10': 'redeemScriptHex'},
    {'1': 'pk_script_hex', '3': 5, '4': 1, '5': 9, '10': 'pkScriptHex'},
    {'1': 'owner', '3': 6, '4': 1, '5': 9, '10': 'owner'},
  ],
};

/// Descriptor for `EscrowUTXO`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List escrowUTXODescriptor = $convert.base64Decode(
    'CgpFc2Nyb3dVVFhPEhIKBHR4aWQYASABKAlSBHR4aWQSEgoEdm91dBgCIAEoDVIEdm91dBIUCg'
    'V2YWx1ZRgDIAEoBFIFdmFsdWUSKgoRcmVkZWVtX3NjcmlwdF9oZXgYBCABKAlSD3JlZGVlbVNj'
    'cmlwdEhleBIiCg1wa19zY3JpcHRfaGV4GAUgASgJUgtwa1NjcmlwdEhleBIUCgVvd25lchgGIA'
    'EoCVIFb3duZXI=');

@$core.Deprecated('Use matchAllocatedNtfnDescriptor instead')
const MatchAllocatedNtfn$json = {
  '1': 'MatchAllocatedNtfn',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'room_id', '3': 2, '4': 1, '5': 9, '10': 'roomId'},
    {'1': 'bet_atoms', '3': 3, '4': 1, '5': 4, '10': 'betAtoms'},
    {'1': 'csv_blocks', '3': 4, '4': 1, '5': 13, '10': 'csvBlocks'},
    {'1': 'a_comp', '3': 5, '4': 1, '5': 12, '10': 'aComp'},
    {'1': 'b_comp', '3': 6, '4': 1, '5': 12, '10': 'bComp'},
  ],
};

/// Descriptor for `MatchAllocatedNtfn`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List matchAllocatedNtfnDescriptor = $convert.base64Decode(
    'ChJNYXRjaEFsbG9jYXRlZE50Zm4SGQoIbWF0Y2hfaWQYASABKAlSB21hdGNoSWQSFwoHcm9vbV'
    '9pZBgCIAEoCVIGcm9vbUlkEhsKCWJldF9hdG9tcxgDIAEoBFIIYmV0QXRvbXMSHQoKY3N2X2Js'
    'b2NrcxgEIAEoDVIJY3N2QmxvY2tzEhUKBmFfY29tcBgFIAEoDFIFYUNvbXASFQoGYl9jb21wGA'
    'YgASgMUgViQ29tcA==');

@$core.Deprecated('Use unreadyGameStreamRequestDescriptor instead')
const UnreadyGameStreamRequest$json = {
  '1': 'UnreadyGameStreamRequest',
  '2': [
    {'1': 'client_id', '3': 1, '4': 1, '5': 9, '10': 'clientId'},
  ],
};

/// Descriptor for `UnreadyGameStreamRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List unreadyGameStreamRequestDescriptor = $convert.base64Decode(
    'ChhVbnJlYWR5R2FtZVN0cmVhbVJlcXVlc3QSGwoJY2xpZW50X2lkGAEgASgJUghjbGllbnRJZA'
    '==');

@$core.Deprecated('Use unreadyGameStreamResponseDescriptor instead')
const UnreadyGameStreamResponse$json = {
  '1': 'UnreadyGameStreamResponse',
};

/// Descriptor for `UnreadyGameStreamResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List unreadyGameStreamResponseDescriptor = $convert.base64Decode(
    'ChlVbnJlYWR5R2FtZVN0cmVhbVJlc3BvbnNl');

@$core.Deprecated('Use startNtfnStreamRequestDescriptor instead')
const StartNtfnStreamRequest$json = {
  '1': 'StartNtfnStreamRequest',
  '2': [
    {'1': 'client_id', '3': 1, '4': 1, '5': 9, '10': 'clientId'},
  ],
};

/// Descriptor for `StartNtfnStreamRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startNtfnStreamRequestDescriptor = $convert.base64Decode(
    'ChZTdGFydE50Zm5TdHJlYW1SZXF1ZXN0EhsKCWNsaWVudF9pZBgBIAEoCVIIY2xpZW50SWQ=');

@$core.Deprecated('Use ntfnStreamResponseDescriptor instead')
const NtfnStreamResponse$json = {
  '1': 'NtfnStreamResponse',
  '2': [
    {'1': 'notification_type', '3': 1, '4': 1, '5': 14, '6': '.pong.NotificationType', '10': 'notificationType'},
    {'1': 'started', '3': 2, '4': 1, '5': 8, '10': 'started'},
    {'1': 'game_id', '3': 3, '4': 1, '5': 9, '10': 'gameId'},
    {'1': 'message', '3': 4, '4': 1, '5': 9, '10': 'message'},
    {'1': 'betAmt', '3': 5, '4': 1, '5': 3, '10': 'betAmt'},
    {'1': 'player_number', '3': 6, '4': 1, '5': 5, '10': 'playerNumber'},
    {'1': 'player_id', '3': 7, '4': 1, '5': 9, '10': 'playerId'},
    {'1': 'room_id', '3': 8, '4': 1, '5': 9, '10': 'roomId'},
    {'1': 'wr', '3': 9, '4': 1, '5': 11, '6': '.pong.WaitingRoom', '10': 'wr'},
    {'1': 'ready', '3': 10, '4': 1, '5': 8, '10': 'ready'},
    {'1': 'match_alloc', '3': 11, '4': 1, '5': 11, '6': '.pong.MatchAllocatedNtfn', '10': 'matchAlloc'},
    {'1': 'confs', '3': 12, '4': 1, '5': 13, '10': 'confs'},
  ],
};

/// Descriptor for `NtfnStreamResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List ntfnStreamResponseDescriptor = $convert.base64Decode(
    'ChJOdGZuU3RyZWFtUmVzcG9uc2USQwoRbm90aWZpY2F0aW9uX3R5cGUYASABKA4yFi5wb25nLk'
    '5vdGlmaWNhdGlvblR5cGVSEG5vdGlmaWNhdGlvblR5cGUSGAoHc3RhcnRlZBgCIAEoCFIHc3Rh'
    'cnRlZBIXCgdnYW1lX2lkGAMgASgJUgZnYW1lSWQSGAoHbWVzc2FnZRgEIAEoCVIHbWVzc2FnZR'
    'IWCgZiZXRBbXQYBSABKANSBmJldEFtdBIjCg1wbGF5ZXJfbnVtYmVyGAYgASgFUgxwbGF5ZXJO'
    'dW1iZXISGwoJcGxheWVyX2lkGAcgASgJUghwbGF5ZXJJZBIXCgdyb29tX2lkGAggASgJUgZyb2'
    '9tSWQSIQoCd3IYCSABKAsyES5wb25nLldhaXRpbmdSb29tUgJ3chIUCgVyZWFkeRgKIAEoCFIF'
    'cmVhZHkSOQoLbWF0Y2hfYWxsb2MYCyABKAsyGC5wb25nLk1hdGNoQWxsb2NhdGVkTnRmblIKbW'
    'F0Y2hBbGxvYxIUCgVjb25mcxgMIAEoDVIFY29uZnM=');

@$core.Deprecated('Use waitingRoomsRequestDescriptor instead')
const WaitingRoomsRequest$json = {
  '1': 'WaitingRoomsRequest',
  '2': [
    {'1': 'room_id', '3': 1, '4': 1, '5': 9, '10': 'roomId'},
  ],
};

/// Descriptor for `WaitingRoomsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomsRequestDescriptor = $convert.base64Decode(
    'ChNXYWl0aW5nUm9vbXNSZXF1ZXN0EhcKB3Jvb21faWQYASABKAlSBnJvb21JZA==');

@$core.Deprecated('Use waitingRoomsResponseDescriptor instead')
const WaitingRoomsResponse$json = {
  '1': 'WaitingRoomsResponse',
  '2': [
    {'1': 'wr', '3': 1, '4': 3, '5': 11, '6': '.pong.WaitingRoom', '10': 'wr'},
  ],
};

/// Descriptor for `WaitingRoomsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomsResponseDescriptor = $convert.base64Decode(
    'ChRXYWl0aW5nUm9vbXNSZXNwb25zZRIhCgJ3chgBIAMoCzIRLnBvbmcuV2FpdGluZ1Jvb21SAn'
    'dy');

@$core.Deprecated('Use joinWaitingRoomRequestDescriptor instead')
const JoinWaitingRoomRequest$json = {
  '1': 'JoinWaitingRoomRequest',
  '2': [
    {'1': 'room_id', '3': 1, '4': 1, '5': 9, '10': 'roomId'},
    {'1': 'client_id', '3': 2, '4': 1, '5': 9, '10': 'clientId'},
    {'1': 'escrow_id', '3': 3, '4': 1, '5': 9, '10': 'escrowId'},
  ],
};

/// Descriptor for `JoinWaitingRoomRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List joinWaitingRoomRequestDescriptor = $convert.base64Decode(
    'ChZKb2luV2FpdGluZ1Jvb21SZXF1ZXN0EhcKB3Jvb21faWQYASABKAlSBnJvb21JZBIbCgljbG'
    'llbnRfaWQYAiABKAlSCGNsaWVudElkEhsKCWVzY3Jvd19pZBgDIAEoCVIIZXNjcm93SWQ=');

@$core.Deprecated('Use joinWaitingRoomResponseDescriptor instead')
const JoinWaitingRoomResponse$json = {
  '1': 'JoinWaitingRoomResponse',
  '2': [
    {'1': 'wr', '3': 1, '4': 1, '5': 11, '6': '.pong.WaitingRoom', '10': 'wr'},
  ],
};

/// Descriptor for `JoinWaitingRoomResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List joinWaitingRoomResponseDescriptor = $convert.base64Decode(
    'ChdKb2luV2FpdGluZ1Jvb21SZXNwb25zZRIhCgJ3chgBIAEoCzIRLnBvbmcuV2FpdGluZ1Jvb2'
    '1SAndy');

@$core.Deprecated('Use createWaitingRoomRequestDescriptor instead')
const CreateWaitingRoomRequest$json = {
  '1': 'CreateWaitingRoomRequest',
  '2': [
    {'1': 'host_id', '3': 1, '4': 1, '5': 9, '10': 'hostId'},
    {'1': 'betAmt', '3': 2, '4': 1, '5': 3, '10': 'betAmt'},
    {'1': 'escrow_id', '3': 3, '4': 1, '5': 9, '10': 'escrowId'},
  ],
};

/// Descriptor for `CreateWaitingRoomRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWaitingRoomRequestDescriptor = $convert.base64Decode(
    'ChhDcmVhdGVXYWl0aW5nUm9vbVJlcXVlc3QSFwoHaG9zdF9pZBgBIAEoCVIGaG9zdElkEhYKBm'
    'JldEFtdBgCIAEoA1IGYmV0QW10EhsKCWVzY3Jvd19pZBgDIAEoCVIIZXNjcm93SWQ=');

@$core.Deprecated('Use createWaitingRoomResponseDescriptor instead')
const CreateWaitingRoomResponse$json = {
  '1': 'CreateWaitingRoomResponse',
  '2': [
    {'1': 'wr', '3': 1, '4': 1, '5': 11, '6': '.pong.WaitingRoom', '10': 'wr'},
  ],
};

/// Descriptor for `CreateWaitingRoomResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWaitingRoomResponseDescriptor = $convert.base64Decode(
    'ChlDcmVhdGVXYWl0aW5nUm9vbVJlc3BvbnNlEiEKAndyGAEgASgLMhEucG9uZy5XYWl0aW5nUm'
    '9vbVICd3I=');

@$core.Deprecated('Use waitingRoomDescriptor instead')
const WaitingRoom$json = {
  '1': 'WaitingRoom',
  '2': [
    {'1': 'id', '3': 1, '4': 1, '5': 9, '10': 'id'},
    {'1': 'host_id', '3': 2, '4': 1, '5': 9, '10': 'hostId'},
    {'1': 'players', '3': 3, '4': 3, '5': 11, '6': '.pong.Player', '10': 'players'},
    {'1': 'bet_amt', '3': 4, '4': 1, '5': 3, '10': 'betAmt'},
  ],
};

/// Descriptor for `WaitingRoom`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomDescriptor = $convert.base64Decode(
    'CgtXYWl0aW5nUm9vbRIOCgJpZBgBIAEoCVICaWQSFwoHaG9zdF9pZBgCIAEoCVIGaG9zdElkEi'
    'YKB3BsYXllcnMYAyADKAsyDC5wb25nLlBsYXllclIHcGxheWVycxIXCgdiZXRfYW10GAQgASgD'
    'UgZiZXRBbXQ=');

@$core.Deprecated('Use waitingRoomRequestDescriptor instead')
const WaitingRoomRequest$json = {
  '1': 'WaitingRoomRequest',
  '2': [
    {'1': 'room_id', '3': 1, '4': 1, '5': 9, '10': 'roomId'},
  ],
};

/// Descriptor for `WaitingRoomRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomRequestDescriptor = $convert.base64Decode(
    'ChJXYWl0aW5nUm9vbVJlcXVlc3QSFwoHcm9vbV9pZBgBIAEoCVIGcm9vbUlk');

@$core.Deprecated('Use waitingRoomResponseDescriptor instead')
const WaitingRoomResponse$json = {
  '1': 'WaitingRoomResponse',
  '2': [
    {'1': 'wr', '3': 1, '4': 1, '5': 11, '6': '.pong.WaitingRoom', '10': 'wr'},
  ],
};

/// Descriptor for `WaitingRoomResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomResponseDescriptor = $convert.base64Decode(
    'ChNXYWl0aW5nUm9vbVJlc3BvbnNlEiEKAndyGAEgASgLMhEucG9uZy5XYWl0aW5nUm9vbVICd3'
    'I=');

@$core.Deprecated('Use playerDescriptor instead')
const Player$json = {
  '1': 'Player',
  '2': [
    {'1': 'uid', '3': 1, '4': 1, '5': 9, '10': 'uid'},
    {'1': 'nick', '3': 2, '4': 1, '5': 9, '10': 'nick'},
    {'1': 'bet_amt', '3': 3, '4': 1, '5': 3, '10': 'betAmt'},
    {'1': 'number', '3': 4, '4': 1, '5': 5, '10': 'number'},
    {'1': 'score', '3': 5, '4': 1, '5': 5, '10': 'score'},
    {'1': 'ready', '3': 6, '4': 1, '5': 8, '10': 'ready'},
  ],
};

/// Descriptor for `Player`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List playerDescriptor = $convert.base64Decode(
    'CgZQbGF5ZXISEAoDdWlkGAEgASgJUgN1aWQSEgoEbmljaxgCIAEoCVIEbmljaxIXCgdiZXRfYW'
    '10GAMgASgDUgZiZXRBbXQSFgoGbnVtYmVyGAQgASgFUgZudW1iZXISFAoFc2NvcmUYBSABKAVS'
    'BXNjb3JlEhQKBXJlYWR5GAYgASgIUgVyZWFkeQ==');

@$core.Deprecated('Use startGameStreamRequestDescriptor instead')
const StartGameStreamRequest$json = {
  '1': 'StartGameStreamRequest',
  '2': [
    {'1': 'client_id', '3': 1, '4': 1, '5': 9, '10': 'clientId'},
  ],
};

/// Descriptor for `StartGameStreamRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List startGameStreamRequestDescriptor = $convert.base64Decode(
    'ChZTdGFydEdhbWVTdHJlYW1SZXF1ZXN0EhsKCWNsaWVudF9pZBgBIAEoCVIIY2xpZW50SWQ=');

@$core.Deprecated('Use gameUpdateBytesDescriptor instead')
const GameUpdateBytes$json = {
  '1': 'GameUpdateBytes',
  '2': [
    {'1': 'data', '3': 1, '4': 1, '5': 12, '10': 'data'},
  ],
};

/// Descriptor for `GameUpdateBytes`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List gameUpdateBytesDescriptor = $convert.base64Decode(
    'Cg9HYW1lVXBkYXRlQnl0ZXMSEgoEZGF0YRgBIAEoDFIEZGF0YQ==');

@$core.Deprecated('Use playerInputDescriptor instead')
const PlayerInput$json = {
  '1': 'PlayerInput',
  '2': [
    {'1': 'player_id', '3': 1, '4': 1, '5': 9, '10': 'playerId'},
    {'1': 'input', '3': 2, '4': 1, '5': 9, '10': 'input'},
    {'1': 'player_number', '3': 3, '4': 1, '5': 5, '10': 'playerNumber'},
  ],
};

/// Descriptor for `PlayerInput`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List playerInputDescriptor = $convert.base64Decode(
    'CgtQbGF5ZXJJbnB1dBIbCglwbGF5ZXJfaWQYASABKAlSCHBsYXllcklkEhQKBWlucHV0GAIgAS'
    'gJUgVpbnB1dBIjCg1wbGF5ZXJfbnVtYmVyGAMgASgFUgxwbGF5ZXJOdW1iZXI=');

@$core.Deprecated('Use gameUpdateDescriptor instead')
const GameUpdate$json = {
  '1': 'GameUpdate',
  '2': [
    {'1': 'gameWidth', '3': 13, '4': 1, '5': 1, '10': 'gameWidth'},
    {'1': 'gameHeight', '3': 14, '4': 1, '5': 1, '10': 'gameHeight'},
    {'1': 'p1Width', '3': 15, '4': 1, '5': 1, '10': 'p1Width'},
    {'1': 'p1Height', '3': 16, '4': 1, '5': 1, '10': 'p1Height'},
    {'1': 'p2Width', '3': 17, '4': 1, '5': 1, '10': 'p2Width'},
    {'1': 'p2Height', '3': 18, '4': 1, '5': 1, '10': 'p2Height'},
    {'1': 'ballWidth', '3': 19, '4': 1, '5': 1, '10': 'ballWidth'},
    {'1': 'ballHeight', '3': 20, '4': 1, '5': 1, '10': 'ballHeight'},
    {'1': 'p1Score', '3': 21, '4': 1, '5': 5, '10': 'p1Score'},
    {'1': 'p2Score', '3': 22, '4': 1, '5': 5, '10': 'p2Score'},
    {'1': 'ballX', '3': 1, '4': 1, '5': 1, '10': 'ballX'},
    {'1': 'ballY', '3': 2, '4': 1, '5': 1, '10': 'ballY'},
    {'1': 'p1X', '3': 3, '4': 1, '5': 1, '10': 'p1X'},
    {'1': 'p1Y', '3': 4, '4': 1, '5': 1, '10': 'p1Y'},
    {'1': 'p2X', '3': 5, '4': 1, '5': 1, '10': 'p2X'},
    {'1': 'p2Y', '3': 6, '4': 1, '5': 1, '10': 'p2Y'},
    {'1': 'p1YVelocity', '3': 7, '4': 1, '5': 1, '10': 'p1YVelocity'},
    {'1': 'p2YVelocity', '3': 8, '4': 1, '5': 1, '10': 'p2YVelocity'},
    {'1': 'ballXVelocity', '3': 9, '4': 1, '5': 1, '10': 'ballXVelocity'},
    {'1': 'ballYVelocity', '3': 10, '4': 1, '5': 1, '10': 'ballYVelocity'},
    {'1': 'fps', '3': 11, '4': 1, '5': 1, '10': 'fps'},
    {'1': 'tps', '3': 12, '4': 1, '5': 1, '10': 'tps'},
    {'1': 'error', '3': 23, '4': 1, '5': 9, '10': 'error'},
    {'1': 'debug', '3': 24, '4': 1, '5': 8, '10': 'debug'},
  ],
};

/// Descriptor for `GameUpdate`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List gameUpdateDescriptor = $convert.base64Decode(
    'CgpHYW1lVXBkYXRlEhwKCWdhbWVXaWR0aBgNIAEoAVIJZ2FtZVdpZHRoEh4KCmdhbWVIZWlnaH'
    'QYDiABKAFSCmdhbWVIZWlnaHQSGAoHcDFXaWR0aBgPIAEoAVIHcDFXaWR0aBIaCghwMUhlaWdo'
    'dBgQIAEoAVIIcDFIZWlnaHQSGAoHcDJXaWR0aBgRIAEoAVIHcDJXaWR0aBIaCghwMkhlaWdodB'
    'gSIAEoAVIIcDJIZWlnaHQSHAoJYmFsbFdpZHRoGBMgASgBUgliYWxsV2lkdGgSHgoKYmFsbEhl'
    'aWdodBgUIAEoAVIKYmFsbEhlaWdodBIYCgdwMVNjb3JlGBUgASgFUgdwMVNjb3JlEhgKB3AyU2'
    'NvcmUYFiABKAVSB3AyU2NvcmUSFAoFYmFsbFgYASABKAFSBWJhbGxYEhQKBWJhbGxZGAIgASgB'
    'UgViYWxsWRIQCgNwMVgYAyABKAFSA3AxWBIQCgNwMVkYBCABKAFSA3AxWRIQCgNwMlgYBSABKA'
    'FSA3AyWBIQCgNwMlkYBiABKAFSA3AyWRIgCgtwMVlWZWxvY2l0eRgHIAEoAVILcDFZVmVsb2Np'
    'dHkSIAoLcDJZVmVsb2NpdHkYCCABKAFSC3AyWVZlbG9jaXR5EiQKDWJhbGxYVmVsb2NpdHkYCS'
    'ABKAFSDWJhbGxYVmVsb2NpdHkSJAoNYmFsbFlWZWxvY2l0eRgKIAEoAVINYmFsbFlWZWxvY2l0'
    'eRIQCgNmcHMYCyABKAFSA2ZwcxIQCgN0cHMYDCABKAFSA3RwcxIUCgVlcnJvchgXIAEoCVIFZX'
    'Jyb3ISFAoFZGVidWcYGCABKAhSBWRlYnVn');

@$core.Deprecated('Use leaveWaitingRoomRequestDescriptor instead')
const LeaveWaitingRoomRequest$json = {
  '1': 'LeaveWaitingRoomRequest',
  '2': [
    {'1': 'client_id', '3': 1, '4': 1, '5': 9, '10': 'clientId'},
    {'1': 'room_id', '3': 2, '4': 1, '5': 9, '10': 'roomId'},
  ],
};

/// Descriptor for `LeaveWaitingRoomRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List leaveWaitingRoomRequestDescriptor = $convert.base64Decode(
    'ChdMZWF2ZVdhaXRpbmdSb29tUmVxdWVzdBIbCgljbGllbnRfaWQYASABKAlSCGNsaWVudElkEh'
    'cKB3Jvb21faWQYAiABKAlSBnJvb21JZA==');

@$core.Deprecated('Use leaveWaitingRoomResponseDescriptor instead')
const LeaveWaitingRoomResponse$json = {
  '1': 'LeaveWaitingRoomResponse',
  '2': [
    {'1': 'success', '3': 1, '4': 1, '5': 8, '10': 'success'},
    {'1': 'message', '3': 2, '4': 1, '5': 9, '10': 'message'},
  ],
};

/// Descriptor for `LeaveWaitingRoomResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List leaveWaitingRoomResponseDescriptor = $convert.base64Decode(
    'ChhMZWF2ZVdhaXRpbmdSb29tUmVzcG9uc2USGAoHc3VjY2VzcxgBIAEoCFIHc3VjY2VzcxIYCg'
    'dtZXNzYWdlGAIgASgJUgdtZXNzYWdl');

@$core.Deprecated('Use signalReadyToPlayRequestDescriptor instead')
const SignalReadyToPlayRequest$json = {
  '1': 'SignalReadyToPlayRequest',
  '2': [
    {'1': 'client_id', '3': 1, '4': 1, '5': 9, '10': 'clientId'},
    {'1': 'game_id', '3': 2, '4': 1, '5': 9, '10': 'gameId'},
  ],
};

/// Descriptor for `SignalReadyToPlayRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List signalReadyToPlayRequestDescriptor = $convert.base64Decode(
    'ChhTaWduYWxSZWFkeVRvUGxheVJlcXVlc3QSGwoJY2xpZW50X2lkGAEgASgJUghjbGllbnRJZB'
    'IXCgdnYW1lX2lkGAIgASgJUgZnYW1lSWQ=');

@$core.Deprecated('Use signalReadyToPlayResponseDescriptor instead')
const SignalReadyToPlayResponse$json = {
  '1': 'SignalReadyToPlayResponse',
  '2': [
    {'1': 'success', '3': 1, '4': 1, '5': 8, '10': 'success'},
    {'1': 'message', '3': 2, '4': 1, '5': 9, '10': 'message'},
  ],
};

/// Descriptor for `SignalReadyToPlayResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List signalReadyToPlayResponseDescriptor = $convert.base64Decode(
    'ChlTaWduYWxSZWFkeVRvUGxheVJlc3BvbnNlEhgKB3N1Y2Nlc3MYASABKAhSB3N1Y2Nlc3MSGA'
    'oHbWVzc2FnZRgCIAEoCVIHbWVzc2FnZQ==');

