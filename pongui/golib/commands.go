package golib

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"net/http"
	_ "net/http/pprof"
	"os"
	"sync"
	"sync/atomic"
	"time"

	"github.com/companyzero/bisonrelay/client"
	"github.com/davecgh/go-spew/spew"
)

type CmdType = int32

const (
	CTUnknown           CmdType = 0x00
	CTHello                     = 0x01
	CTInitClient                = 0x02
	CTGetUserNick               = 0x03
	CTStopClient                = 0x04
	CTGetWRPlayers              = 0x05
	CTGetWaitingRooms           = 0x06
	CTJoinWaitingRoom           = 0x07
	CTCreateWaitingRoom         = 0x08
	CTLeaveWaitingRoom          = 0x09
	// Settlement-related commands
	CTGenerateSessionKey = 0x0a
	CTOpenEscrow         = 0x0b
	CTStartPreSign       = 0x0c
	// Archive current session key into historic dir using match_id
	CTArchiveSessionKey = 0x0e

	// Wallet-auth commands (migrated from Dart gRPC)
	CTRequestNonce = 0x0f
	CTVerifyLogin  = 0x10

	// Player action commands (migrated from Dart gRPC)
	CTSendInput         = 0x11
	CTSignalReadyToPlay = 0x12
	CTUnreadyGameStream = 0x13
	CTStartGameStream   = 0x14

	// Refund command for CSV-locked escrows
	CTRefundEscrow = 0x15

	// List all historic escrows for refund checking
	CTListHistoricEscrows = 0x16
	// Cache partial escrow info to disk to prevent fund loss
	CTCacheEscrowInfo      = 0x17
	CTUpdateHistoricEscrow = 0x18
	CTDeleteHistoricEscrow = 0x19

	// Config management (UI<->golib)
	CTGetClientConfig  = 0x1a
	CTSaveClientConfig = 0x1b
	// Refund session validation
	CTValidateRefundSession = 0x1c

	CTCreateLockFile        = 0x60
	CTCloseLockFile         = 0x61
	CTGetRunState           = 0x83
	CTEnableBackgroundNtfs  = 0x84
	CTDisableBackgroundNtfs = 0x85
	CTEnableProfiler        = 0x86
	CTZipTimedProfilingLogs = 0x87
	CTEnableTimedProfiling  = 0x89

	NTUINotification = 0x1001
	NTClientStopped  = 0x1002
	NTLogLine        = 0x1003
	NTNOP            = 0x1004
	NTWRCreated      = 0x1005
	// High-frequency game frames (raw bytes of pong.GameUpdate)
	NTGameFrame = 0x1011
)

type cmd struct {
	Type         CmdType
	ID           int32
	ClientHandle int32
	Payload      []byte
}

func (cmd *cmd) decode(to interface{}) error {
	return json.Unmarshal(cmd.Payload, to)
}

type CmdResult struct {
	ID      int32
	Type    CmdType
	Err     error
	Payload []byte
}

type CmdResultLoopCB interface {
	F(id int32, typ int32, payload string, err string)
	UINtfn(text string, nick string, ts int64)
}

// Buffered channel to prevent blocking when FFI isolate can't keep up with high-frequency frames.
// Buffer size of 200 allows ~3 seconds of frames at 60fps before blocking.
var cmdResultChan = make(chan *CmdResult, 200)

// Tracks dropped NTGameFrame notifications when cmdResultChan is full.
var ntGameFrameDrops atomic.Uint64

// GetAndResetNTGameFrameDrops returns the number of NTGameFrame drops since the last call and resets the counter.
func GetAndResetNTGameFrameDrops() uint64 {
	return ntGameFrameDrops.Swap(0)
}

// Shared NTNOP result to avoid allocations in the polling loop
var ntnopResult = &CmdResult{Type: NTNOP, Payload: []byte{}}

func call(cmd *cmd) *CmdResult {
	var v interface{}
	var err error

	decode := func(to interface{}) bool {
		err = cmd.decode(to)
		if err != nil {
			err = fmt.Errorf("unable to decode input payload: %v; full payload: %s", err, spew.Sdump(cmd.Payload))
		}
		return err == nil
	}

	// ctx := context.Background()
	// Handle calls that do not need a client.
	switch cmd.Type {
	case CTHello:
		var name string
		if decode(&name) {
			v, err = handleHello(name)
		}
	case CTInitClient:
		var initClient initClient
		if decode(&initClient) {
			v, err = handleInitClient(uint32(cmd.ClientHandle), initClient)
		}

	case CTRequestNonce:
		var args requestNonceArgs
		if decode(&args) {
			v, err = handleRequestNonce(uint32(cmd.ClientHandle), args)
		}

	case CTVerifyLogin:
		var args verifyLoginArgs
		if decode(&args) {
			v, err = handleVerifyLogin(uint32(cmd.ClientHandle), args)
		}

	case CTCreateLockFile:
		var args string
		decode(&args)
		err = handleCreateLockFile(args)

	case CTCloseLockFile:
		var args string
		decode(&args)
		err = handleCloseLockFile(args)

	case CTGetRunState:
		v = runState{
			ClientRunning: isClientRunning(uint32(cmd.ClientHandle)),
		}
		err = nil

	case CTEnableProfiler:
		var args string
		decode(&args)
		if args == "" {
			args = "0.0.0.0:8118"
		}
		fmt.Printf("Enabling profiler on %s\n", args)
		go func() {
			err := http.ListenAndServe(args, nil)
			if err != nil {
				fmt.Printf("Unable to listen on profiler %s: %v\n",
					args, err)
			}
		}()

	case CTEnableTimedProfiling:
		var args string
		decode(&args)
		go globalProfiler.Run(args)

	case CTZipTimedProfilingLogs:
		var args string
		decode(&args)
		err = globalProfiler.zipLogs(args)

	case CTGetClientConfig:
		// Prefer the running client's app config if available for this handle.
		// Accept an optional payload with data_dir provided by Flutter.
		var args struct {
			DataDir string `json:"data_dir"`
		}
		_ = decode(&args)
		v, err = handleGetClientConfigForHandle(uint32(cmd.ClientHandle), args.DataDir)

	case CTSaveClientConfig:
		var args saveClientConfigArgs
		if decode(&args) {
			v, err = handleSaveClientConfig(args)
		}
	default:
		// Calls that need a client. Figure out the client.
		cmtx.Lock()
		var client *clientCtx
		if cs != nil {
			client = cs[uint32(cmd.ClientHandle)]
		}
		cmtx.Unlock()

		if client == nil {
			err = fmt.Errorf("unknown client handle %d", cmd.ClientHandle)
		} else {
			v, err = handleClientCmd(client, cmd)
		}
	}

	var resPayload []byte
	if err == nil {
		resPayload, err = json.Marshal(v)
	}

	return &CmdResult{ID: cmd.ID, Type: cmd.Type, Err: err, Payload: resPayload}
}

func AsyncCall(typ CmdType, id, clientHandle int32, payload []byte) {
	cmd := &cmd{
		Type:         typ,
		ID:           id,
		ClientHandle: clientHandle,
		Payload:      payload,
	}
	go func() { cmdResultChan <- call(cmd) }()
}

func AsyncCallStr(typ CmdType, id, clientHandle int32, payload string) {
	cmd := &cmd{
		Type:         typ,
		ID:           id,
		ClientHandle: clientHandle,
		Payload:      []byte(payload),
	}
	go func() { cmdResultChan <- call(cmd) }()
}

func notify(typ CmdType, payload interface{}, err error) {
	var resPayload []byte
	if err == nil {
		switch v := payload.(type) {
		case []byte:
			// Pass-through raw bytes (used by NTGameFrame)
			resPayload = v
		default:
			resPayload, err = json.Marshal(payload)
		}
	}

	r := &CmdResult{Type: typ, Err: err, Payload: resPayload}

	// For high-frequency game frames, use non-blocking send to avoid blocking
	// the command handler when the FFI isolate is slow.
	if typ == NTGameFrame {
		select {
		case cmdResultChan <- r:
			// Frame sent successfully
		default:
			// Buffer full, drop frame to avoid blocking.
			// Track drop for per-second visibility on the producer side.
			ntGameFrameDrops.Add(1)
			fmt.Println("cmdResultChan full, dropping frame") // keep existing visibility
			// Buffer full, drop frame to avoid blocking
			// This prevents the command handler from stalling when FFI isolate is slow
		}
	} else {
		// For other notifications, blocking send is fine (they're low-frequency)
		cmdResultChan <- r
	}
}

func NextCmdResult() *CmdResult {
	select {
	case r := <-cmdResultChan:
		return r
	case <-time.After(16 * time.Millisecond):
		// Return shared NTNOP result to avoid allocations in the polling loop
		return ntnopResult
	}
}

var (
	cmdResultLoopsMtx   sync.Mutex
	cmdResultLoops      = map[int32]chan struct{}{}
	cmdResultLoopsLive  atomic.Int32
	cmdResultLoopsCount int32
)

// emitBackgroundNtfns emits background notifications to the callback object.
func emitBackgroundNtfns(r *CmdResult, cb CmdResultLoopCB) {
	switch r.Type {
	case NTUINotification:
		var n client.UINotification
		err := json.Unmarshal(r.Payload, &n)
		if err != nil {
			return
		}

		cb.UINtfn(n.Text, n.FromNick, n.Timestamp)

	default:
		// Ignore every other notification.
	}
}

// CmdResultLoop runs the loop that fetches async results in a goroutine and
// calls cb.F() with the results. Returns an ID that may be passed to
// StopCmdResultLoop to stop this goroutine.
//
// If onlyBgNtfns is specified, only background notifications are sent.
func CmdResultLoop(cb CmdResultLoopCB, onlyBgNtfns bool) int32 {
	cmdResultLoopsMtx.Lock()
	id := cmdResultLoopsCount + 1
	cmdResultLoopsCount += 1
	ch := make(chan struct{})
	cmdResultLoops[id] = ch
	cmdResultLoopsLive.Add(1)
	cmdResultLoopsMtx.Unlock()

	// onlyBgNtfns == true when this is called from the native plugin
	// code while the flutter engine is _not_ attached to it.
	deliverBackgroundNtfns := onlyBgNtfns

	cmtx.Lock()
	if cs != nil && cs[0x12131400] != nil {
		cc := cs[0x12131400]
		cc.log.Infof("CmdResultLoop: starting new run for pid %d id %d",
			os.Getpid(), id)
	}
	cmtx.Unlock()

	go func() {
		minuteTicker := time.NewTicker(time.Minute)
		defer minuteTicker.Stop()
		startTime := time.Now()
		wallStartTime := startTime.Round(0)
		lastTime := startTime
		lastCPUTimes := make([]cpuTime, 6)

		defer func() {
			cmtx.Lock()
			if cs != nil && cs[0x12131400] != nil {
				elapsed := time.Since(startTime).Truncate(time.Millisecond)
				elapsedWall := time.Now().Round(0).Sub(wallStartTime).Truncate(time.Millisecond)
				cc := cs[0x12131400]
				cc.log.Infof("CmdResultLoop: finishing "+
					"goroutine for pid %d id %d after %s (wall %s)",
					os.Getpid(), id, elapsed, elapsedWall)
			}
			cmtx.Unlock()
		}()

		for {
			var r *CmdResult
			select {
			case r = <-cmdResultChan:
			case <-minuteTicker.C:
				// This is being used to debug background issues
				// on mobile. It may be removed in the future.
				go reportCmdResultLoop(startTime, lastTime, id, lastCPUTimes)
				lastTime = time.Now()
				continue

			case <-ch:
				return
			}

			// Process the special commands that toggle calling
			// native code with background ntfn events.
			switch r.Type {
			case CTEnableBackgroundNtfs:
				deliverBackgroundNtfns = true
				continue
			case CTDisableBackgroundNtfs:
				deliverBackgroundNtfns = false
				continue
			}

			// If the flutter engine is attached to the process,
			// deliver the event so that it can be processed.
			if !onlyBgNtfns {
				var errMsg, payload string
				if r.Err != nil {
					errMsg = r.Err.Error()
				}
				if len(r.Payload) > 0 {
					if r.Type == NTGameFrame {
						// Encode as base64 for mobile transports that only support string payloads.
						payload = base64.StdEncoding.EncodeToString(r.Payload)
					} else {
						payload = string(r.Payload)
					}
				}
				cb.F(r.ID, r.Type, payload, errMsg)
			}

			// Emit a background ntfn if the flutter engine is
			// deatched or if it is attached but paused/on
			// background.
			if deliverBackgroundNtfns {
				emitBackgroundNtfns(r, cb)
			}
		}
	}()

	return id
}

// StopCmdResultLoop stops an async goroutine created with CmdResultLoop. Does
// nothing if this goroutine is already stopped.
func StopCmdResultLoop(id int32) {
	cmdResultLoopsMtx.Lock()
	ch := cmdResultLoops[id]
	delete(cmdResultLoops, id)
	cmdResultLoopsLive.Add(-1)
	cmdResultLoopsMtx.Unlock()
	if ch != nil {
		close(ch)
	}
}

// StopAllCmdResultLoops stops all async goroutines created by CmdResultLoop.
func StopAllCmdResultLoops() {
	cmdResultLoopsMtx.Lock()
	chans := cmdResultLoops
	cmdResultLoops = map[int32]chan struct{}{}
	cmdResultLoopsLive.Store(0)
	cmdResultLoopsMtx.Unlock()
	for _, ch := range chans {
		close(ch)
	}
}

// ClientExists returns true if the client with the specified handle is running.
func ClientExists(handle int32) bool {
	cmtx.Lock()
	exists := cs != nil && cs[uint32(handle)] != nil
	cmtx.Unlock()
	return exists
}

func LogInfo(id int32, s string) {
	cmtx.Lock()
	if cs != nil && cs[uint32(id)] != nil {
		cs[uint32(id)].log.Info(s)
	} else {
		fmt.Println(s)
	}
	cmtx.Unlock()
}
