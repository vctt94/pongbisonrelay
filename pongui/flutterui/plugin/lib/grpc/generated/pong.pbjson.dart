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

@$core.Deprecated('Use branchDescriptor instead')
const Branch$json = {
  '1': 'Branch',
  '2': [
    {'1': 'BRANCH_UNSPECIFIED', '2': 0},
    {'1': 'BRANCH_A', '2': 1},
    {'1': 'BRANCH_B', '2': 2},
  ],
};

/// Descriptor for `Branch`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List branchDescriptor = $convert.base64Decode(
    'CgZCcmFuY2gSFgoSQlJBTkNIX1VOU1BFQ0lGSUVEEAASDAoIQlJBTkNIX0EQARIMCghCUkFOQ0'
    'hfQhAC');

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
  ],
};

/// Descriptor for `NotificationType`. Decode as a `google.protobuf.EnumDescriptorProto`.
final $typed_data.Uint8List notificationTypeDescriptor = $convert.base64Decode(
    'ChBOb3RpZmljYXRpb25UeXBlEgsKB1VOS05PV04QABILCgdNRVNTQUdFEAESDgoKR0FNRV9TVE'
    'FSVBACEgwKCEdBTUVfRU5EEAMSGQoVT1BQT05FTlRfRElTQ09OTkVDVEVEEAQSFQoRQkVUX0FN'
    'T1VOVF9VUERBVEUQBRIUChBQTEFZRVJfSk9JTkVEX1dSEAYSEQoNT05fV1JfQ1JFQVRFRBAHEh'
    'MKD09OX1BMQVlFUl9SRUFEWRAIEhEKDU9OX1dSX1JFTU9WRUQQCRISCg5QTEFZRVJfTEVGVF9X'
    'UhAKEhQKEENPVU5URE9XTl9VUERBVEUQCxIWChJHQU1FX1JFQURZX1RPX1BMQVkQDA==');

@$core.Deprecated('Use clientMsgDescriptor instead')
const ClientMsg$json = {
  '1': 'ClientMsg',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'hello', '3': 10, '4': 1, '5': 11, '6': '.pong.Hello', '9': 0, '10': 'hello'},
    {'1': 'presigs', '3': 11, '4': 1, '5': 11, '6': '.pong.PreSigBatch', '9': 0, '10': 'presigs'},
    {'1': 'ack', '3': 12, '4': 1, '5': 11, '6': '.pong.Ack', '9': 0, '10': 'ack'},
  ],
  '8': [
    {'1': 'kind'},
  ],
};

/// Descriptor for `ClientMsg`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List clientMsgDescriptor = $convert.base64Decode(
    'CglDbGllbnRNc2cSGQoIbWF0Y2hfaWQYASABKAlSB21hdGNoSWQSIwoFaGVsbG8YCiABKAsyCy'
    '5wb25nLkhlbGxvSABSBWhlbGxvEi0KB3ByZXNpZ3MYCyABKAsyES5wb25nLlByZVNpZ0JhdGNo'
    'SABSB3ByZXNpZ3MSHQoDYWNrGAwgASgLMgkucG9uZy5BY2tIAFIDYWNrQgYKBGtpbmQ=');

@$core.Deprecated('Use serverMsgDescriptor instead')
const ServerMsg$json = {
  '1': 'ServerMsg',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'role', '3': 10, '4': 1, '5': 11, '6': '.pong.AssignRole', '9': 0, '10': 'role'},
    {'1': 'req', '3': 11, '4': 1, '5': 11, '6': '.pong.NeedPreSigs', '9': 0, '10': 'req'},
    {'1': 'reveal', '3': 12, '4': 1, '5': 11, '6': '.pong.RevealGamma', '9': 0, '10': 'reveal'},
    {'1': 'info', '3': 13, '4': 1, '5': 11, '6': '.pong.Info', '9': 0, '10': 'info'},
  ],
  '8': [
    {'1': 'kind'},
  ],
};

/// Descriptor for `ServerMsg`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List serverMsgDescriptor = $convert.base64Decode(
    'CglTZXJ2ZXJNc2cSGQoIbWF0Y2hfaWQYASABKAlSB21hdGNoSWQSJgoEcm9sZRgKIAEoCzIQLn'
    'BvbmcuQXNzaWduUm9sZUgAUgRyb2xlEiUKA3JlcRgLIAEoCzIRLnBvbmcuTmVlZFByZVNpZ3NI'
    'AFIDcmVxEisKBnJldmVhbBgMIAEoCzIRLnBvbmcuUmV2ZWFsR2FtbWFIAFIGcmV2ZWFsEiAKBG'
    'luZm8YDSABKAsyCi5wb25nLkluZm9IAFIEaW5mb0IGCgRraW5k');

@$core.Deprecated('Use helloDescriptor instead')
const Hello$json = {
  '1': 'Hello',
  '2': [
    {'1': 'comp_pubkey', '3': 1, '4': 1, '5': 12, '10': 'compPubkey'},
    {'1': 'client_version', '3': 2, '4': 1, '5': 9, '10': 'clientVersion'},
  ],
};

/// Descriptor for `Hello`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List helloDescriptor = $convert.base64Decode(
    'CgVIZWxsbxIfCgtjb21wX3B1YmtleRgBIAEoDFIKY29tcFB1YmtleRIlCg5jbGllbnRfdmVyc2'
    'lvbhgCIAEoCVINY2xpZW50VmVyc2lvbg==');

@$core.Deprecated('Use assignRoleDescriptor instead')
const AssignRole$json = {
  '1': 'AssignRole',
  '2': [
    {'1': 'role', '3': 1, '4': 1, '5': 14, '6': '.pong.AssignRole.Role', '10': 'role'},
    {'1': 'required_atoms', '3': 2, '4': 1, '5': 4, '10': 'requiredAtoms'},
    {'1': 'deposit_pkscript_hex', '3': 3, '4': 1, '5': 9, '10': 'depositPkscriptHex'},
  ],
  '4': [AssignRole_Role$json],
};

@$core.Deprecated('Use assignRoleDescriptor instead')
const AssignRole_Role$json = {
  '1': 'Role',
  '2': [
    {'1': 'UNKNOWN', '2': 0},
    {'1': 'A', '2': 1},
    {'1': 'B', '2': 2},
  ],
};

/// Descriptor for `AssignRole`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List assignRoleDescriptor = $convert.base64Decode(
    'CgpBc3NpZ25Sb2xlEikKBHJvbGUYASABKA4yFS5wb25nLkFzc2lnblJvbGUuUm9sZVIEcm9sZR'
    'IlCg5yZXF1aXJlZF9hdG9tcxgCIAEoBFINcmVxdWlyZWRBdG9tcxIwChRkZXBvc2l0X3Brc2Ny'
    'aXB0X2hleBgDIAEoCVISZGVwb3NpdFBrc2NyaXB0SGV4IiEKBFJvbGUSCwoHVU5LTk9XThAAEg'
    'UKAUEQARIFCgFCEAI=');

@$core.Deprecated('Use needPreSigsDescriptor instead')
const NeedPreSigs$json = {
  '1': 'NeedPreSigs',
  '2': [
    {'1': 'branches_to_presign', '3': 1, '4': 3, '5': 14, '6': '.pong.Branch', '10': 'branchesToPresign'},
    {'1': 'draft_awins_tx_hex', '3': 2, '4': 1, '5': 9, '10': 'draftAwinsTxHex'},
    {'1': 'draft_bwins_tx_hex', '3': 3, '4': 1, '5': 9, '10': 'draftBwinsTxHex'},
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
    {'1': 'm_awins_hex', '3': 3, '4': 1, '5': 9, '10': 'mAwinsHex'},
    {'1': 'T_awins_compressed', '3': 4, '4': 1, '5': 12, '10': 'TAwinsCompressed'},
    {'1': 'm_bwins_hex', '3': 5, '4': 1, '5': 9, '10': 'mBwinsHex'},
    {'1': 'T_bwins_compressed', '3': 6, '4': 1, '5': 12, '10': 'TBwinsCompressed'},
  ],
};

/// Descriptor for `NeedPreSigs`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List needPreSigsDescriptor = $convert.base64Decode(
    'CgtOZWVkUHJlU2lncxI8ChNicmFuY2hlc190b19wcmVzaWduGAEgAygOMgwucG9uZy5CcmFuY2'
    'hSEWJyYW5jaGVzVG9QcmVzaWduEisKEmRyYWZ0X2F3aW5zX3R4X2hleBgCIAEoCVIPZHJhZnRB'
    'd2luc1R4SGV4EisKEmRyYWZ0X2J3aW5zX3R4X2hleBgDIAEoCVIPZHJhZnRCd2luc1R4SGV4Ej'
    'IKBmlucHV0cxgEIAMoCzIaLnBvbmcuTmVlZFByZVNpZ3MuUGVySW5wdXRSBmlucHV0cxrtAQoI'
    'UGVySW5wdXQSGQoIaW5wdXRfaWQYASABKAlSB2lucHV0SWQSKgoRcmVkZWVtX3NjcmlwdF9oZX'
    'gYAiABKAlSD3JlZGVlbVNjcmlwdEhleBIeCgttX2F3aW5zX2hleBgDIAEoCVIJbUF3aW5zSGV4'
    'EiwKElRfYXdpbnNfY29tcHJlc3NlZBgEIAEoDFIQVEF3aW5zQ29tcHJlc3NlZBIeCgttX2J3aW'
    '5zX2hleBgFIAEoCVIJbUJ3aW5zSGV4EiwKElRfYndpbnNfY29tcHJlc3NlZBgGIAEoDFIQVEJ3'
    'aW5zQ29tcHJlc3NlZA==');

@$core.Deprecated('Use preSigBatchDescriptor instead')
const PreSigBatch$json = {
  '1': 'PreSigBatch',
  '2': [
    {'1': 'branch', '3': 1, '4': 1, '5': 14, '6': '.pong.Branch', '10': 'branch'},
    {'1': 'presigs', '3': 2, '4': 3, '5': 11, '6': '.pong.PreSigBatch.Sig', '10': 'presigs'},
  ],
  '3': [PreSigBatch_Sig$json],
};

@$core.Deprecated('Use preSigBatchDescriptor instead')
const PreSigBatch_Sig$json = {
  '1': 'Sig',
  '2': [
    {'1': 'input_id', '3': 1, '4': 1, '5': 9, '10': 'inputId'},
    {'1': 'rprime32', '3': 2, '4': 1, '5': 12, '10': 'rprime32'},
    {'1': 'sprime32', '3': 3, '4': 1, '5': 12, '10': 'sprime32'},
  ],
};

/// Descriptor for `PreSigBatch`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List preSigBatchDescriptor = $convert.base64Decode(
    'CgtQcmVTaWdCYXRjaBIkCgZicmFuY2gYASABKA4yDC5wb25nLkJyYW5jaFIGYnJhbmNoEi8KB3'
    'ByZXNpZ3MYAiADKAsyFS5wb25nLlByZVNpZ0JhdGNoLlNpZ1IHcHJlc2lncxpYCgNTaWcSGQoI'
    'aW5wdXRfaWQYASABKAlSB2lucHV0SWQSGgoIcnByaW1lMzIYAiABKAxSCHJwcmltZTMyEhoKCH'
    'NwcmltZTMyGAMgASgMUghzcHJpbWUzMg==');

@$core.Deprecated('Use revealGammaDescriptor instead')
const RevealGamma$json = {
  '1': 'RevealGamma',
  '2': [
    {'1': 'gamma32', '3': 1, '4': 1, '5': 12, '10': 'gamma32'},
  ],
};

/// Descriptor for `RevealGamma`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List revealGammaDescriptor = $convert.base64Decode(
    'CgtSZXZlYWxHYW1tYRIYCgdnYW1tYTMyGAEgASgMUgdnYW1tYTMy');

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

@$core.Deprecated('Use createMatchRequestDescriptor instead')
const CreateMatchRequest$json = {
  '1': 'CreateMatchRequest',
  '2': [
    {'1': 'a_c', '3': 1, '4': 1, '5': 9, '10': 'aC'},
    {'1': 'b_c', '3': 2, '4': 1, '5': 9, '10': 'bC'},
    {'1': 'csv', '3': 3, '4': 1, '5': 13, '10': 'csv'},
  ],
};

/// Descriptor for `CreateMatchRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createMatchRequestDescriptor = $convert.base64Decode(
    'ChJDcmVhdGVNYXRjaFJlcXVlc3QSDwoDYV9jGAEgASgJUgJhQxIPCgNiX2MYAiABKAlSAmJDEh'
    'AKA2NzdhgDIAEoDVIDY3N2');

@$core.Deprecated('Use createMatchResponseDescriptor instead')
const CreateMatchResponse$json = {
  '1': 'CreateMatchResponse',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 's_c', '3': 2, '4': 1, '5': 9, '10': 'sC'},
    {'1': 'a_a', '3': 3, '4': 1, '5': 9, '10': 'aA'},
    {'1': 'a_b', '3': 4, '4': 1, '5': 9, '10': 'aB'},
    {'1': 'a_s', '3': 5, '4': 1, '5': 9, '10': 'aS'},
    {'1': 'x_a', '3': 6, '4': 1, '5': 9, '10': 'xA'},
    {'1': 'x_b', '3': 7, '4': 1, '5': 9, '10': 'xB'},
    {'1': 'csv', '3': 8, '4': 1, '5': 13, '10': 'csv'},
    {'1': 'escrow_template_a', '3': 9, '4': 1, '5': 9, '10': 'escrowTemplateA'},
    {'1': 'escrow_template_b', '3': 10, '4': 1, '5': 9, '10': 'escrowTemplateB'},
  ],
};

/// Descriptor for `CreateMatchResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createMatchResponseDescriptor = $convert.base64Decode(
    'ChNDcmVhdGVNYXRjaFJlc3BvbnNlEhkKCG1hdGNoX2lkGAEgASgJUgdtYXRjaElkEg8KA3NfYx'
    'gCIAEoCVICc0MSDwoDYV9hGAMgASgJUgJhQRIPCgNhX2IYBCABKAlSAmFCEg8KA2FfcxgFIAEo'
    'CVICYVMSDwoDeF9hGAYgASgJUgJ4QRIPCgN4X2IYByABKAlSAnhCEhAKA2NzdhgIIAEoDVIDY3'
    'N2EioKEWVzY3Jvd190ZW1wbGF0ZV9hGAkgASgJUg9lc2Nyb3dUZW1wbGF0ZUESKgoRZXNjcm93'
    'X3RlbXBsYXRlX2IYCiABKAlSD2VzY3Jvd1RlbXBsYXRlQg==');

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

@$core.Deprecated('Use submitFundingRequestDescriptor instead')
const SubmitFundingRequest$json = {
  '1': 'SubmitFundingRequest',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'escrows', '3': 2, '4': 3, '5': 11, '6': '.pong.EscrowUTXO', '10': 'escrows'},
  ],
};

/// Descriptor for `SubmitFundingRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List submitFundingRequestDescriptor = $convert.base64Decode(
    'ChRTdWJtaXRGdW5kaW5nUmVxdWVzdBIZCghtYXRjaF9pZBgBIAEoCVIHbWF0Y2hJZBIqCgdlc2'
    'Nyb3dzGAIgAygLMhAucG9uZy5Fc2Nyb3dVVFhPUgdlc2Nyb3dz');

@$core.Deprecated('Use submitFundingResponseDescriptor instead')
const SubmitFundingResponse$json = {
  '1': 'SubmitFundingResponse',
  '2': [
    {'1': 'ok', '3': 1, '4': 1, '5': 8, '10': 'ok'},
  ],
};

/// Descriptor for `SubmitFundingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List submitFundingResponseDescriptor = $convert.base64Decode(
    'ChVTdWJtaXRGdW5kaW5nUmVzcG9uc2USDgoCb2sYASABKAhSAm9r');

@$core.Deprecated('Use revealAdaptorEntryDescriptor instead')
const RevealAdaptorEntry$json = {
  '1': 'RevealAdaptorEntry',
  '2': [
    {'1': 'input_id', '3': 1, '4': 1, '5': 9, '10': 'inputId'},
    {'1': 'gamma', '3': 2, '4': 1, '5': 9, '10': 'gamma'},
  ],
};

/// Descriptor for `RevealAdaptorEntry`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List revealAdaptorEntryDescriptor = $convert.base64Decode(
    'ChJSZXZlYWxBZGFwdG9yRW50cnkSGQoIaW5wdXRfaWQYASABKAlSB2lucHV0SWQSFAoFZ2FtbW'
    'EYAiABKAlSBWdhbW1h');

@$core.Deprecated('Use revealAdaptorsRequestDescriptor instead')
const RevealAdaptorsRequest$json = {
  '1': 'RevealAdaptorsRequest',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'branch', '3': 2, '4': 1, '5': 14, '6': '.pong.Branch', '10': 'branch'},
  ],
};

/// Descriptor for `RevealAdaptorsRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List revealAdaptorsRequestDescriptor = $convert.base64Decode(
    'ChVSZXZlYWxBZGFwdG9yc1JlcXVlc3QSGQoIbWF0Y2hfaWQYASABKAlSB21hdGNoSWQSJAoGYn'
    'JhbmNoGAIgASgOMgwucG9uZy5CcmFuY2hSBmJyYW5jaA==');

@$core.Deprecated('Use revealAdaptorsResponseDescriptor instead')
const RevealAdaptorsResponse$json = {
  '1': 'RevealAdaptorsResponse',
  '2': [
    {'1': 'branch', '3': 1, '4': 1, '5': 14, '6': '.pong.Branch', '10': 'branch'},
    {'1': 'entries', '3': 2, '4': 3, '5': 11, '6': '.pong.RevealAdaptorEntry', '10': 'entries'},
  ],
};

/// Descriptor for `RevealAdaptorsResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List revealAdaptorsResponseDescriptor = $convert.base64Decode(
    'ChZSZXZlYWxBZGFwdG9yc1Jlc3BvbnNlEiQKBmJyYW5jaBgBIAEoDjIMLnBvbmcuQnJhbmNoUg'
    'ZicmFuY2gSMgoHZW50cmllcxgCIAMoCzIYLnBvbmcuUmV2ZWFsQWRhcHRvckVudHJ5UgdlbnRy'
    'aWVz');

@$core.Deprecated('Use allocateMatchRequestDescriptor instead')
const AllocateMatchRequest$json = {
  '1': 'AllocateMatchRequest',
  '2': [
    {'1': 'a_c', '3': 1, '4': 1, '5': 9, '10': 'aC'},
    {'1': 'b_c', '3': 2, '4': 1, '5': 9, '10': 'bC'},
    {'1': 'bet_atoms', '3': 3, '4': 1, '5': 4, '10': 'betAtoms'},
  ],
};

/// Descriptor for `AllocateMatchRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List allocateMatchRequestDescriptor = $convert.base64Decode(
    'ChRBbGxvY2F0ZU1hdGNoUmVxdWVzdBIPCgNhX2MYASABKAlSAmFDEg8KA2JfYxgCIAEoCVICYk'
    'MSGwoJYmV0X2F0b21zGAMgASgEUghiZXRBdG9tcw==');

@$core.Deprecated('Use allocateMatchResponseDescriptor instead')
const AllocateMatchResponse$json = {
  '1': 'AllocateMatchResponse',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'csv', '3': 2, '4': 1, '5': 13, '10': 'csv'},
    {'1': 'bet_atoms', '3': 3, '4': 1, '5': 4, '10': 'betAtoms'},
    {'1': 'fee_atoms', '3': 4, '4': 1, '5': 4, '10': 'feeAtoms'},
    {'1': 'a_redeem_script_hex', '3': 5, '4': 1, '5': 9, '10': 'aRedeemScriptHex'},
    {'1': 'a_pk_script_hex', '3': 6, '4': 1, '5': 9, '10': 'aPkScriptHex'},
    {'1': 'a_deposit_address', '3': 7, '4': 1, '5': 9, '10': 'aDepositAddress'},
    {'1': 'b_redeem_script_hex', '3': 8, '4': 1, '5': 9, '10': 'bRedeemScriptHex'},
    {'1': 'b_pk_script_hex', '3': 9, '4': 1, '5': 9, '10': 'bPkScriptHex'},
    {'1': 'b_deposit_address', '3': 10, '4': 1, '5': 9, '10': 'bDepositAddress'},
  ],
};

/// Descriptor for `AllocateMatchResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List allocateMatchResponseDescriptor = $convert.base64Decode(
    'ChVBbGxvY2F0ZU1hdGNoUmVzcG9uc2USGQoIbWF0Y2hfaWQYASABKAlSB21hdGNoSWQSEAoDY3'
    'N2GAIgASgNUgNjc3YSGwoJYmV0X2F0b21zGAMgASgEUghiZXRBdG9tcxIbCglmZWVfYXRvbXMY'
    'BCABKARSCGZlZUF0b21zEi0KE2FfcmVkZWVtX3NjcmlwdF9oZXgYBSABKAlSEGFSZWRlZW1TY3'
    'JpcHRIZXgSJQoPYV9wa19zY3JpcHRfaGV4GAYgASgJUgxhUGtTY3JpcHRIZXgSKgoRYV9kZXBv'
    'c2l0X2FkZHJlc3MYByABKAlSD2FEZXBvc2l0QWRkcmVzcxItChNiX3JlZGVlbV9zY3JpcHRfaG'
    'V4GAggASgJUhBiUmVkZWVtU2NyaXB0SGV4EiUKD2JfcGtfc2NyaXB0X2hleBgJIAEoCVIMYlBr'
    'U2NyaXB0SGV4EioKEWJfZGVwb3NpdF9hZGRyZXNzGAogASgJUg9iRGVwb3NpdEFkZHJlc3M=');

@$core.Deprecated('Use waitFundingRequestDescriptor instead')
const WaitFundingRequest$json = {
  '1': 'WaitFundingRequest',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
  ],
};

/// Descriptor for `WaitFundingRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitFundingRequestDescriptor = $convert.base64Decode(
    'ChJXYWl0RnVuZGluZ1JlcXVlc3QSGQoIbWF0Y2hfaWQYASABKAlSB21hdGNoSWQ=');

@$core.Deprecated('Use waitFundingResponseDescriptor instead')
const WaitFundingResponse$json = {
  '1': 'WaitFundingResponse',
  '2': [
    {'1': 'confirmed', '3': 1, '4': 1, '5': 8, '10': 'confirmed'},
    {'1': 'confs_a', '3': 2, '4': 1, '5': 13, '10': 'confsA'},
    {'1': 'confs_b', '3': 3, '4': 1, '5': 13, '10': 'confsB'},
    {'1': 'utxo_a', '3': 4, '4': 1, '5': 11, '6': '.pong.EscrowUTXO', '10': 'utxoA'},
    {'1': 'utxo_b', '3': 5, '4': 1, '5': 11, '6': '.pong.EscrowUTXO', '10': 'utxoB'},
  ],
};

/// Descriptor for `WaitFundingResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitFundingResponseDescriptor = $convert.base64Decode(
    'ChNXYWl0RnVuZGluZ1Jlc3BvbnNlEhwKCWNvbmZpcm1lZBgBIAEoCFIJY29uZmlybWVkEhcKB2'
    'NvbmZzX2EYAiABKA1SBmNvbmZzQRIXCgdjb25mc19iGAMgASgNUgZjb25mc0ISJwoGdXR4b19h'
    'GAQgASgLMhAucG9uZy5Fc2Nyb3dVVFhPUgV1dHhvQRInCgZ1dHhvX2IYBSABKAsyEC5wb25nLk'
    'VzY3Jvd1VUWE9SBXV0eG9C');

@$core.Deprecated('Use allocateEscrowRequestDescriptor instead')
const AllocateEscrowRequest$json = {
  '1': 'AllocateEscrowRequest',
  '2': [
    {'1': 'player_id', '3': 1, '4': 1, '5': 9, '10': 'playerId'},
    {'1': 'a_c', '3': 2, '4': 1, '5': 9, '10': 'aC'},
    {'1': 'bet_atoms', '3': 3, '4': 1, '5': 4, '10': 'betAtoms'},
    {'1': 'csv', '3': 4, '4': 1, '5': 13, '10': 'csv'},
  ],
};

/// Descriptor for `AllocateEscrowRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List allocateEscrowRequestDescriptor = $convert.base64Decode(
    'ChVBbGxvY2F0ZUVzY3Jvd1JlcXVlc3QSGwoJcGxheWVyX2lkGAEgASgJUghwbGF5ZXJJZBIPCg'
    'NhX2MYAiABKAlSAmFDEhsKCWJldF9hdG9tcxgDIAEoBFIIYmV0QXRvbXMSEAoDY3N2GAQgASgN'
    'UgNjc3Y=');

@$core.Deprecated('Use allocateEscrowResponseDescriptor instead')
const AllocateEscrowResponse$json = {
  '1': 'AllocateEscrowResponse',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'role', '3': 2, '4': 1, '5': 14, '6': '.pong.Branch', '10': 'role'},
    {'1': 'a_a', '3': 3, '4': 1, '5': 9, '10': 'aA'},
    {'1': 'a_b', '3': 4, '4': 1, '5': 9, '10': 'aB'},
    {'1': 'a_s', '3': 5, '4': 1, '5': 9, '10': 'aS'},
    {'1': 'x_a', '3': 6, '4': 1, '5': 9, '10': 'xA'},
    {'1': 'x_b', '3': 7, '4': 1, '5': 9, '10': 'xB'},
    {'1': 'csv', '3': 8, '4': 1, '5': 13, '10': 'csv'},
    {'1': 'deposit_address', '3': 9, '4': 1, '5': 9, '10': 'depositAddress'},
    {'1': 'redeem_script_hex', '3': 10, '4': 1, '5': 9, '10': 'redeemScriptHex'},
    {'1': 'pk_script_hex', '3': 11, '4': 1, '5': 9, '10': 'pkScriptHex'},
    {'1': 'required_atoms', '3': 12, '4': 1, '5': 4, '10': 'requiredAtoms'},
  ],
};

/// Descriptor for `AllocateEscrowResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List allocateEscrowResponseDescriptor = $convert.base64Decode(
    'ChZBbGxvY2F0ZUVzY3Jvd1Jlc3BvbnNlEhkKCG1hdGNoX2lkGAEgASgJUgdtYXRjaElkEiAKBH'
    'JvbGUYAiABKA4yDC5wb25nLkJyYW5jaFIEcm9sZRIPCgNhX2EYAyABKAlSAmFBEg8KA2FfYhgE'
    'IAEoCVICYUISDwoDYV9zGAUgASgJUgJhUxIPCgN4X2EYBiABKAlSAnhBEg8KA3hfYhgHIAEoCV'
    'ICeEISEAoDY3N2GAggASgNUgNjc3YSJwoPZGVwb3NpdF9hZGRyZXNzGAkgASgJUg5kZXBvc2l0'
    'QWRkcmVzcxIqChFyZWRlZW1fc2NyaXB0X2hleBgKIAEoCVIPcmVkZWVtU2NyaXB0SGV4EiIKDX'
    'BrX3NjcmlwdF9oZXgYCyABKAlSC3BrU2NyaXB0SGV4EiUKDnJlcXVpcmVkX2F0b21zGAwgASgE'
    'Ug1yZXF1aXJlZEF0b21z');

@$core.Deprecated('Use finalizeWinnerRequestDescriptor instead')
const FinalizeWinnerRequest$json = {
  '1': 'FinalizeWinnerRequest',
  '2': [
    {'1': 'match_id', '3': 1, '4': 1, '5': 9, '10': 'matchId'},
    {'1': 'branch', '3': 2, '4': 1, '5': 14, '6': '.pong.Branch', '10': 'branch'},
  ],
};

/// Descriptor for `FinalizeWinnerRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List finalizeWinnerRequestDescriptor = $convert.base64Decode(
    'ChVGaW5hbGl6ZVdpbm5lclJlcXVlc3QSGQoIbWF0Y2hfaWQYASABKAlSB21hdGNoSWQSJAoGYn'
    'JhbmNoGAIgASgOMgwucG9uZy5CcmFuY2hSBmJyYW5jaA==');

@$core.Deprecated('Use finalizeWinnerResponseDescriptor instead')
const FinalizeWinnerResponse$json = {
  '1': 'FinalizeWinnerResponse',
  '2': [
    {'1': 'broadcast_txid', '3': 1, '4': 1, '5': 9, '10': 'broadcastTxid'},
  ],
};

/// Descriptor for `FinalizeWinnerResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List finalizeWinnerResponseDescriptor = $convert.base64Decode(
    'ChZGaW5hbGl6ZVdpbm5lclJlc3BvbnNlEiUKDmJyb2FkY2FzdF90eGlkGAEgASgJUg1icm9hZG'
    'Nhc3RUeGlk');

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
    'cmVhZHk=');

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
  ],
};

/// Descriptor for `JoinWaitingRoomRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List joinWaitingRoomRequestDescriptor = $convert.base64Decode(
    'ChZKb2luV2FpdGluZ1Jvb21SZXF1ZXN0EhcKB3Jvb21faWQYASABKAlSBnJvb21JZBIbCgljbG'
    'llbnRfaWQYAiABKAlSCGNsaWVudElk');

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
  ],
};

/// Descriptor for `CreateWaitingRoomRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List createWaitingRoomRequestDescriptor = $convert.base64Decode(
    'ChhDcmVhdGVXYWl0aW5nUm9vbVJlcXVlc3QSFwoHaG9zdF9pZBgBIAEoCVIGaG9zdElkEhYKBm'
    'JldEFtdBgCIAEoA1IGYmV0QW10');

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
};

/// Descriptor for `WaitingRoomRequest`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomRequestDescriptor = $convert.base64Decode(
    'ChJXYWl0aW5nUm9vbVJlcXVlc3Q=');

@$core.Deprecated('Use waitingRoomResponseDescriptor instead')
const WaitingRoomResponse$json = {
  '1': 'WaitingRoomResponse',
  '2': [
    {'1': 'players', '3': 1, '4': 3, '5': 11, '6': '.pong.Player', '10': 'players'},
  ],
};

/// Descriptor for `WaitingRoomResponse`. Decode as a `google.protobuf.DescriptorProto`.
final $typed_data.Uint8List waitingRoomResponseDescriptor = $convert.base64Decode(
    'ChNXYWl0aW5nUm9vbVJlc3BvbnNlEiYKB3BsYXllcnMYASADKAsyDC5wb25nLlBsYXllclIHcG'
    'xheWVycw==');

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

