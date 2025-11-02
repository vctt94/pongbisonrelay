// This is a generated file - do not edit.
//
// Generated from pong.proto.

// @dart = 3.3

// ignore_for_file: annotate_overrides, camel_case_types, comment_references
// ignore_for_file: constant_identifier_names
// ignore_for_file: curly_braces_in_flow_control_structures
// ignore_for_file: deprecated_member_use_from_same_package, library_prefixes
// ignore_for_file: non_constant_identifier_names

import 'dart:async' as $async;
import 'dart:core' as $core;

import 'package:grpc/service_api.dart' as $grpc;
import 'package:protobuf/protobuf.dart' as $pb;

import 'pong.pb.dart' as $0;

export 'pong.pb.dart';

/// Wallet authentication (Decred signmessage over gRPC)
@$pb.GrpcServiceName('pong.PongAuth')
class PongAuthClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  PongAuthClient(super.channel, {super.options, super.interceptors});

  /// Request a one-time nonce to sign in the wallet
  $grpc.ResponseFuture<$0.RequestNonceResponse> requestNonce(
    $0.RequestNonceRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$requestNonce, request, options: options);
  }

  /// Verify a wallet signature and establish a session
  $grpc.ResponseFuture<$0.VerifyLoginResponse> verifyLogin(
    $0.VerifyLoginRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$verifyLogin, request, options: options);
  }

  // method descriptors

  static final _$requestNonce =
      $grpc.ClientMethod<$0.RequestNonceRequest, $0.RequestNonceResponse>(
          '/pong.PongAuth/RequestNonce',
          ($0.RequestNonceRequest value) => value.writeToBuffer(),
          $0.RequestNonceResponse.fromBuffer);
  static final _$verifyLogin =
      $grpc.ClientMethod<$0.VerifyLoginRequest, $0.VerifyLoginResponse>(
          '/pong.PongAuth/VerifyLogin',
          ($0.VerifyLoginRequest value) => value.writeToBuffer(),
          $0.VerifyLoginResponse.fromBuffer);
}

@$pb.GrpcServiceName('pong.PongAuth')
abstract class PongAuthServiceBase extends $grpc.Service {
  $core.String get $name => 'pong.PongAuth';

  PongAuthServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$0.RequestNonceRequest, $0.RequestNonceResponse>(
            'RequestNonce',
            requestNonce_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.RequestNonceRequest.fromBuffer(value),
            ($0.RequestNonceResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.VerifyLoginRequest, $0.VerifyLoginResponse>(
            'VerifyLogin',
            verifyLogin_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.VerifyLoginRequest.fromBuffer(value),
            ($0.VerifyLoginResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.RequestNonceResponse> requestNonce_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.RequestNonceRequest> $request) async {
    return requestNonce($call, await $request);
  }

  $async.Future<$0.RequestNonceResponse> requestNonce(
      $grpc.ServiceCall call, $0.RequestNonceRequest request);

  $async.Future<$0.VerifyLoginResponse> verifyLogin_Pre($grpc.ServiceCall $call,
      $async.Future<$0.VerifyLoginRequest> $request) async {
    return verifyLogin($call, await $request);
  }

  $async.Future<$0.VerifyLoginResponse> verifyLogin(
      $grpc.ServiceCall call, $0.VerifyLoginRequest request);
}

@$pb.GrpcServiceName('pong.PongGame')
class PongGameClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  PongGameClient(super.channel, {super.options, super.interceptors});

  /// pong game
  $grpc.ResponseFuture<$0.GameUpdate> sendInput(
    $0.PlayerInput request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$sendInput, request, options: options);
  }

  $grpc.ResponseStream<$0.GameUpdateBytes> startGameStream(
    $0.StartGameStreamRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(
        _$startGameStream, $async.Stream.fromIterable([request]),
        options: options);
  }

  $grpc.ResponseStream<$0.NtfnStreamResponse> startNtfnStream(
    $0.StartNtfnStreamRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(
        _$startNtfnStream, $async.Stream.fromIterable([request]),
        options: options);
  }

  $grpc.ResponseFuture<$0.UnreadyGameStreamResponse> unreadyGameStream(
    $0.UnreadyGameStreamRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$unreadyGameStream, request, options: options);
  }

  $grpc.ResponseFuture<$0.SignalReadyToPlayResponse> signalReadyToPlay(
    $0.SignalReadyToPlayRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$signalReadyToPlay, request, options: options);
  }

  // method descriptors

  static final _$sendInput = $grpc.ClientMethod<$0.PlayerInput, $0.GameUpdate>(
      '/pong.PongGame/SendInput',
      ($0.PlayerInput value) => value.writeToBuffer(),
      $0.GameUpdate.fromBuffer);
  static final _$startGameStream =
      $grpc.ClientMethod<$0.StartGameStreamRequest, $0.GameUpdateBytes>(
          '/pong.PongGame/StartGameStream',
          ($0.StartGameStreamRequest value) => value.writeToBuffer(),
          $0.GameUpdateBytes.fromBuffer);
  static final _$startNtfnStream =
      $grpc.ClientMethod<$0.StartNtfnStreamRequest, $0.NtfnStreamResponse>(
          '/pong.PongGame/StartNtfnStream',
          ($0.StartNtfnStreamRequest value) => value.writeToBuffer(),
          $0.NtfnStreamResponse.fromBuffer);
  static final _$unreadyGameStream = $grpc.ClientMethod<
          $0.UnreadyGameStreamRequest, $0.UnreadyGameStreamResponse>(
      '/pong.PongGame/UnreadyGameStream',
      ($0.UnreadyGameStreamRequest value) => value.writeToBuffer(),
      $0.UnreadyGameStreamResponse.fromBuffer);
  static final _$signalReadyToPlay = $grpc.ClientMethod<
          $0.SignalReadyToPlayRequest, $0.SignalReadyToPlayResponse>(
      '/pong.PongGame/SignalReadyToPlay',
      ($0.SignalReadyToPlayRequest value) => value.writeToBuffer(),
      $0.SignalReadyToPlayResponse.fromBuffer);
}

@$pb.GrpcServiceName('pong.PongGame')
abstract class PongGameServiceBase extends $grpc.Service {
  $core.String get $name => 'pong.PongGame';

  PongGameServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.PlayerInput, $0.GameUpdate>(
        'SendInput',
        sendInput_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.PlayerInput.fromBuffer(value),
        ($0.GameUpdate value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.StartGameStreamRequest, $0.GameUpdateBytes>(
            'StartGameStream',
            startGameStream_Pre,
            false,
            true,
            ($core.List<$core.int> value) =>
                $0.StartGameStreamRequest.fromBuffer(value),
            ($0.GameUpdateBytes value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.StartNtfnStreamRequest, $0.NtfnStreamResponse>(
            'StartNtfnStream',
            startNtfnStream_Pre,
            false,
            true,
            ($core.List<$core.int> value) =>
                $0.StartNtfnStreamRequest.fromBuffer(value),
            ($0.NtfnStreamResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.UnreadyGameStreamRequest,
            $0.UnreadyGameStreamResponse>(
        'UnreadyGameStream',
        unreadyGameStream_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.UnreadyGameStreamRequest.fromBuffer(value),
        ($0.UnreadyGameStreamResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.SignalReadyToPlayRequest,
            $0.SignalReadyToPlayResponse>(
        'SignalReadyToPlay',
        signalReadyToPlay_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.SignalReadyToPlayRequest.fromBuffer(value),
        ($0.SignalReadyToPlayResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.GameUpdate> sendInput_Pre(
      $grpc.ServiceCall $call, $async.Future<$0.PlayerInput> $request) async {
    return sendInput($call, await $request);
  }

  $async.Future<$0.GameUpdate> sendInput(
      $grpc.ServiceCall call, $0.PlayerInput request);

  $async.Stream<$0.GameUpdateBytes> startGameStream_Pre($grpc.ServiceCall $call,
      $async.Future<$0.StartGameStreamRequest> $request) async* {
    yield* startGameStream($call, await $request);
  }

  $async.Stream<$0.GameUpdateBytes> startGameStream(
      $grpc.ServiceCall call, $0.StartGameStreamRequest request);

  $async.Stream<$0.NtfnStreamResponse> startNtfnStream_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.StartNtfnStreamRequest> $request) async* {
    yield* startNtfnStream($call, await $request);
  }

  $async.Stream<$0.NtfnStreamResponse> startNtfnStream(
      $grpc.ServiceCall call, $0.StartNtfnStreamRequest request);

  $async.Future<$0.UnreadyGameStreamResponse> unreadyGameStream_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.UnreadyGameStreamRequest> $request) async {
    return unreadyGameStream($call, await $request);
  }

  $async.Future<$0.UnreadyGameStreamResponse> unreadyGameStream(
      $grpc.ServiceCall call, $0.UnreadyGameStreamRequest request);

  $async.Future<$0.SignalReadyToPlayResponse> signalReadyToPlay_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.SignalReadyToPlayRequest> $request) async {
    return signalReadyToPlay($call, await $request);
  }

  $async.Future<$0.SignalReadyToPlayResponse> signalReadyToPlay(
      $grpc.ServiceCall call, $0.SignalReadyToPlayRequest request);
}

@$pb.GrpcServiceName('pong.PongWaitingRoom')
class PongWaitingRoomClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  PongWaitingRoomClient(super.channel, {super.options, super.interceptors});

  /// waiting room
  $grpc.ResponseFuture<$0.WaitingRoomResponse> getWaitingRoom(
    $0.WaitingRoomRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getWaitingRoom, request, options: options);
  }

  $grpc.ResponseFuture<$0.WaitingRoomsResponse> getWaitingRooms(
    $0.WaitingRoomsRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getWaitingRooms, request, options: options);
  }

  $grpc.ResponseFuture<$0.CreateWaitingRoomResponse> createWaitingRoom(
    $0.CreateWaitingRoomRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$createWaitingRoom, request, options: options);
  }

  $grpc.ResponseFuture<$0.JoinWaitingRoomResponse> joinWaitingRoom(
    $0.JoinWaitingRoomRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$joinWaitingRoom, request, options: options);
  }

  $grpc.ResponseFuture<$0.LeaveWaitingRoomResponse> leaveWaitingRoom(
    $0.LeaveWaitingRoomRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$leaveWaitingRoom, request, options: options);
  }

  // method descriptors

  static final _$getWaitingRoom =
      $grpc.ClientMethod<$0.WaitingRoomRequest, $0.WaitingRoomResponse>(
          '/pong.PongWaitingRoom/GetWaitingRoom',
          ($0.WaitingRoomRequest value) => value.writeToBuffer(),
          $0.WaitingRoomResponse.fromBuffer);
  static final _$getWaitingRooms =
      $grpc.ClientMethod<$0.WaitingRoomsRequest, $0.WaitingRoomsResponse>(
          '/pong.PongWaitingRoom/GetWaitingRooms',
          ($0.WaitingRoomsRequest value) => value.writeToBuffer(),
          $0.WaitingRoomsResponse.fromBuffer);
  static final _$createWaitingRoom = $grpc.ClientMethod<
          $0.CreateWaitingRoomRequest, $0.CreateWaitingRoomResponse>(
      '/pong.PongWaitingRoom/CreateWaitingRoom',
      ($0.CreateWaitingRoomRequest value) => value.writeToBuffer(),
      $0.CreateWaitingRoomResponse.fromBuffer);
  static final _$joinWaitingRoom =
      $grpc.ClientMethod<$0.JoinWaitingRoomRequest, $0.JoinWaitingRoomResponse>(
          '/pong.PongWaitingRoom/JoinWaitingRoom',
          ($0.JoinWaitingRoomRequest value) => value.writeToBuffer(),
          $0.JoinWaitingRoomResponse.fromBuffer);
  static final _$leaveWaitingRoom = $grpc.ClientMethod<
          $0.LeaveWaitingRoomRequest, $0.LeaveWaitingRoomResponse>(
      '/pong.PongWaitingRoom/LeaveWaitingRoom',
      ($0.LeaveWaitingRoomRequest value) => value.writeToBuffer(),
      $0.LeaveWaitingRoomResponse.fromBuffer);
}

@$pb.GrpcServiceName('pong.PongWaitingRoom')
abstract class PongWaitingRoomServiceBase extends $grpc.Service {
  $core.String get $name => 'pong.PongWaitingRoom';

  PongWaitingRoomServiceBase() {
    $addMethod(
        $grpc.ServiceMethod<$0.WaitingRoomRequest, $0.WaitingRoomResponse>(
            'GetWaitingRoom',
            getWaitingRoom_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.WaitingRoomRequest.fromBuffer(value),
            ($0.WaitingRoomResponse value) => value.writeToBuffer()));
    $addMethod(
        $grpc.ServiceMethod<$0.WaitingRoomsRequest, $0.WaitingRoomsResponse>(
            'GetWaitingRooms',
            getWaitingRooms_Pre,
            false,
            false,
            ($core.List<$core.int> value) =>
                $0.WaitingRoomsRequest.fromBuffer(value),
            ($0.WaitingRoomsResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.CreateWaitingRoomRequest,
            $0.CreateWaitingRoomResponse>(
        'CreateWaitingRoom',
        createWaitingRoom_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.CreateWaitingRoomRequest.fromBuffer(value),
        ($0.CreateWaitingRoomResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.JoinWaitingRoomRequest,
            $0.JoinWaitingRoomResponse>(
        'JoinWaitingRoom',
        joinWaitingRoom_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.JoinWaitingRoomRequest.fromBuffer(value),
        ($0.JoinWaitingRoomResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.LeaveWaitingRoomRequest,
            $0.LeaveWaitingRoomResponse>(
        'LeaveWaitingRoom',
        leaveWaitingRoom_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.LeaveWaitingRoomRequest.fromBuffer(value),
        ($0.LeaveWaitingRoomResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.WaitingRoomResponse> getWaitingRoom_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.WaitingRoomRequest> $request) async {
    return getWaitingRoom($call, await $request);
  }

  $async.Future<$0.WaitingRoomResponse> getWaitingRoom(
      $grpc.ServiceCall call, $0.WaitingRoomRequest request);

  $async.Future<$0.WaitingRoomsResponse> getWaitingRooms_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.WaitingRoomsRequest> $request) async {
    return getWaitingRooms($call, await $request);
  }

  $async.Future<$0.WaitingRoomsResponse> getWaitingRooms(
      $grpc.ServiceCall call, $0.WaitingRoomsRequest request);

  $async.Future<$0.CreateWaitingRoomResponse> createWaitingRoom_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.CreateWaitingRoomRequest> $request) async {
    return createWaitingRoom($call, await $request);
  }

  $async.Future<$0.CreateWaitingRoomResponse> createWaitingRoom(
      $grpc.ServiceCall call, $0.CreateWaitingRoomRequest request);

  $async.Future<$0.JoinWaitingRoomResponse> joinWaitingRoom_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.JoinWaitingRoomRequest> $request) async {
    return joinWaitingRoom($call, await $request);
  }

  $async.Future<$0.JoinWaitingRoomResponse> joinWaitingRoom(
      $grpc.ServiceCall call, $0.JoinWaitingRoomRequest request);

  $async.Future<$0.LeaveWaitingRoomResponse> leaveWaitingRoom_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.LeaveWaitingRoomRequest> $request) async {
    return leaveWaitingRoom($call, await $request);
  }

  $async.Future<$0.LeaveWaitingRoomResponse> leaveWaitingRoom(
      $grpc.ServiceCall call, $0.LeaveWaitingRoomRequest request);
}

/// Referee service coordinating escrow and settlement
@$pb.GrpcServiceName('pong.PongReferee')
class PongRefereeClient extends $grpc.Client {
  /// The hostname for this service.
  static const $core.String defaultHost = '';

  /// OAuth scopes needed for the client.
  static const $core.List<$core.String> oauthScopes = [
    '',
  ];

  PongRefereeClient(super.channel, {super.options, super.interceptors});

  /// Escrow-first funding
  $grpc.ResponseFuture<$0.OpenEscrowResponse> openEscrow(
    $0.OpenEscrowRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$openEscrow, request, options: options);
  }

  /// SettlementStream streams the settlement process
  $grpc.ResponseStream<$0.ServerMsg> settlementStream(
    $async.Stream<$0.ClientMsg> request, {
    $grpc.CallOptions? options,
  }) {
    return $createStreamingCall(_$settlementStream, request, options: options);
  }

  /// Winner fetches gamma and both presigs to finalize the exact winning draft
  $grpc.ResponseFuture<$0.GetFinalizeBundleResponse> getFinalizeBundle(
    $0.GetFinalizeBundleRequest request, {
    $grpc.CallOptions? options,
  }) {
    return $createUnaryCall(_$getFinalizeBundle, request, options: options);
  }

  // method descriptors

  static final _$openEscrow =
      $grpc.ClientMethod<$0.OpenEscrowRequest, $0.OpenEscrowResponse>(
          '/pong.PongReferee/OpenEscrow',
          ($0.OpenEscrowRequest value) => value.writeToBuffer(),
          $0.OpenEscrowResponse.fromBuffer);
  static final _$settlementStream =
      $grpc.ClientMethod<$0.ClientMsg, $0.ServerMsg>(
          '/pong.PongReferee/SettlementStream',
          ($0.ClientMsg value) => value.writeToBuffer(),
          $0.ServerMsg.fromBuffer);
  static final _$getFinalizeBundle = $grpc.ClientMethod<
          $0.GetFinalizeBundleRequest, $0.GetFinalizeBundleResponse>(
      '/pong.PongReferee/GetFinalizeBundle',
      ($0.GetFinalizeBundleRequest value) => value.writeToBuffer(),
      $0.GetFinalizeBundleResponse.fromBuffer);
}

@$pb.GrpcServiceName('pong.PongReferee')
abstract class PongRefereeServiceBase extends $grpc.Service {
  $core.String get $name => 'pong.PongReferee';

  PongRefereeServiceBase() {
    $addMethod($grpc.ServiceMethod<$0.OpenEscrowRequest, $0.OpenEscrowResponse>(
        'OpenEscrow',
        openEscrow_Pre,
        false,
        false,
        ($core.List<$core.int> value) => $0.OpenEscrowRequest.fromBuffer(value),
        ($0.OpenEscrowResponse value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.ClientMsg, $0.ServerMsg>(
        'SettlementStream',
        settlementStream,
        true,
        true,
        ($core.List<$core.int> value) => $0.ClientMsg.fromBuffer(value),
        ($0.ServerMsg value) => value.writeToBuffer()));
    $addMethod($grpc.ServiceMethod<$0.GetFinalizeBundleRequest,
            $0.GetFinalizeBundleResponse>(
        'GetFinalizeBundle',
        getFinalizeBundle_Pre,
        false,
        false,
        ($core.List<$core.int> value) =>
            $0.GetFinalizeBundleRequest.fromBuffer(value),
        ($0.GetFinalizeBundleResponse value) => value.writeToBuffer()));
  }

  $async.Future<$0.OpenEscrowResponse> openEscrow_Pre($grpc.ServiceCall $call,
      $async.Future<$0.OpenEscrowRequest> $request) async {
    return openEscrow($call, await $request);
  }

  $async.Future<$0.OpenEscrowResponse> openEscrow(
      $grpc.ServiceCall call, $0.OpenEscrowRequest request);

  $async.Stream<$0.ServerMsg> settlementStream(
      $grpc.ServiceCall call, $async.Stream<$0.ClientMsg> request);

  $async.Future<$0.GetFinalizeBundleResponse> getFinalizeBundle_Pre(
      $grpc.ServiceCall $call,
      $async.Future<$0.GetFinalizeBundleRequest> $request) async {
    return getFinalizeBundle($call, await $request);
  }

  $async.Future<$0.GetFinalizeBundleResponse> getFinalizeBundle(
      $grpc.ServiceCall call, $0.GetFinalizeBundleRequest request);
}
