package client

import (
	"fmt"
	"regexp"
	"sync"
	"time"

	"github.com/companyzero/bisonrelay/zkidentity"
	"github.com/vctt94/pongbisonrelay/pongrpc/grpc/pong"
)

// Following are the notification types. Add new types at the bottom of this
// list, then add a notifyX() to NotificationManager and initialize a new
// container in NewNotificationManager().

const onWRCreatedfnType = "onWRCreated"

// onWRCreatedNtfn is the handler for received private messages.
type OnWRCreatedNtfn func(*pong.WaitingRoom, time.Time)

func (_ OnWRCreatedNtfn) typ() string { return onWRCreatedfnType }

const onBetAmtChangedFnType = "onBetAmtChanged"

// onBetAmtChangedNtfn is the handler for changes in the bet amount.
type OnBetAmtChangedNtfn func(string, int64, time.Time)

func (_ OnBetAmtChangedNtfn) typ() string { return onBetAmtChangedFnType }

const onGameStartedFnType = "onGameStarted"

// OnGameStartedNtfn is the handler for when a game starts.
type OnGameStartedNtfn func(string, time.Time)

func (_ OnGameStartedNtfn) typ() string { return onGameStartedFnType }

const OnPlayerJoinedNtfnType = "onPlayerJoinedWR"

// OnPlayerJoinedNtfn is the handler for when a player enters the wr.
type OnPlayerJoinedNtfn func(*pong.WaitingRoom, time.Time)

func (_ OnPlayerJoinedNtfn) typ() string { return OnPlayerJoinedNtfnType }

const onPlayerLeftNtfnType = "onPlayerLeftWR"

// OnPlayerLeftNtfn is the handler for when a player leaves the waiting room.
type OnPlayerLeftNtfn func(*pong.WaitingRoom, string, time.Time)

func (_ OnPlayerLeftNtfn) typ() string { return onPlayerLeftNtfnType }

const onGameEndedfnType = "onGameEnded"

// onWRCreatedNtfn is the handler for received private messages.
type OnGameEndedNtfn func(string, string, time.Time)

func (_ OnGameEndedNtfn) typ() string { return onGameEndedfnType }

// UINotificationsConfig is the configuration for how UI notifications are
// emitted.
type UINotificationsConfig struct {
	// GameStarted flag whether to emit notification after game starts.
	GameStarted bool

	WRCreated bool

	// MaxLength is the max length of messages emitted.
	MaxLength int

	// MentionRegexp is the regexp to detect mentions.
	MentionRegexp *regexp.Regexp

	// EmitInterval is the interval to wait for additional messages before
	// emitting a notification. Multiple messages received within this
	// interval will only generate a single UI notification.
	EmitInterval time.Duration

	// CancelEmissionChannel may be set to a Context.Done() channel to
	// cancel emission of notifications.
	CancelEmissionChannel <-chan struct{}
}

func (cfg *UINotificationsConfig) clip(msg string) string {
	if len(msg) < cfg.MaxLength {
		return msg
	}
	return msg[:cfg.MaxLength]
}

// UINotificationType is the type of notification.
type UINotificationType string

const (
	UINtfnGameStarted UINotificationType = "gamestarted"
	UINtfnWRCreated   UINotificationType = "wrcreated"
	UINtfnMultiple    UINotificationType = "multiple"
)

// UINotification is a notification that should be shown as an UI alert.
type UINotification struct {
	// Type of notification.
	Type UINotificationType `json:"type"`

	// Text of the notification.
	Text string `json:"text"`

	// Count will be greater than one when multiple notifications were
	// batched.
	Count int `json:"count"`

	// From is the original sender or GC of the notification.
	From zkidentity.ShortID `json:"from"`

	// FromNick is the nick of the sender.
	FromNick string `json:"from_nick"`

	// Timestamp is the unix timestamp in seconds of the first message.
	Timestamp int64 `json:"timestamp"`
}

// fromSame returns true if the notification is from the same ID.
func (n *UINotification) fromSame(id *zkidentity.ShortID) bool {
	if id == nil || n.From.IsEmpty() {
		return false
	}

	return *id == n.From
}

const onUINtfnType = "uintfn"

// OnUINotification is called when a notification should be shown by the UI to
// the user. This should usually take the form of an alert dialog about a
// received message.
type OnUINotification func(ntfn UINotification)

func (_ OnUINotification) typ() string { return onUINtfnType }

// The following is used only in tests.

const onTestNtfnType = "testNtfnType"

type onTestNtfn func()

func (_ onTestNtfn) typ() string { return onTestNtfnType }

// Following is the generic notification code.

type NotificationRegistration struct {
	unreg func() bool
}

func (reg NotificationRegistration) Unregister() bool {
	return reg.unreg()
}

type NotificationHandler interface {
	typ() string
}

type handler[T any] struct {
	handler T
	async   bool
}

type handlersFor[T any] struct {
	mtx      sync.Mutex
	next     uint
	handlers map[uint]handler[T]
}

func (hn *handlersFor[T]) register(h T, async bool) NotificationRegistration {
	var id uint

	hn.mtx.Lock()
	id, hn.next = hn.next, hn.next+1
	if hn.handlers == nil {
		hn.handlers = make(map[uint]handler[T])
	}
	hn.handlers[id] = handler[T]{handler: h, async: async}
	registered := true
	hn.mtx.Unlock()

	return NotificationRegistration{
		unreg: func() bool {
			hn.mtx.Lock()
			res := registered
			if registered {
				delete(hn.handlers, id)
				registered = false
			}
			hn.mtx.Unlock()
			return res
		},
	}
}

func (hn *handlersFor[T]) visit(f func(T)) {
	hn.mtx.Lock()
	for _, h := range hn.handlers {
		if h.async {
			go f(h.handler)
		} else {
			f(h.handler)
		}
	}
	hn.mtx.Unlock()
}

func (hn *handlersFor[T]) Register(v interface{}, async bool) NotificationRegistration {
	if h, ok := v.(T); !ok {
		panic("wrong type")
	} else {
		return hn.register(h, async)
	}
}

func (hn *handlersFor[T]) AnyRegistered() bool {
	hn.mtx.Lock()
	res := len(hn.handlers) > 0
	hn.mtx.Unlock()
	return res
}

type handlersRegistry interface {
	Register(v interface{}, async bool) NotificationRegistration
	AnyRegistered() bool
}

type NotificationManager struct {
	handlers map[string]handlersRegistry

	uiMtx      sync.Mutex
	uiConfig   UINotificationsConfig
	uiNextNtfn UINotification
	uiTimer    *time.Timer
}

// UpdateUIConfig updates the config used to generate UI notifications about
// PMs, GCMs, etc.
func (nmgr *NotificationManager) UpdateUIConfig(cfg UINotificationsConfig) {
	nmgr.uiMtx.Lock()
	nmgr.uiConfig = cfg
	nmgr.uiMtx.Unlock()
}

func (nmgr *NotificationManager) register(handler NotificationHandler, async bool) NotificationRegistration {
	handlers := nmgr.handlers[handler.typ()]
	if handlers == nil {
		panic(fmt.Sprintf("forgot to init the handler type %T "+
			"in NewNotificationManager", handler))
	}

	return handlers.Register(handler, async)
}

// Register registers a callback notification function that is called
// asynchronously to the event (i.e. in a separate goroutine).
func (nmgr *NotificationManager) Register(handler NotificationHandler) NotificationRegistration {
	return nmgr.register(handler, true)
}

// RegisterSync registers a callback notification function that is called
// synchronously to the event. This callback SHOULD return as soon as possible,
// otherwise the client might hang.
//
// Synchronous callbacks are mostly intended for tests and when external
// callers need to ensure proper order of multiple sequential events. In
// general it is preferable to use callbacks registered with the Register call,
// to ensure the client will not deadlock or hang.
func (nmgr *NotificationManager) RegisterSync(handler NotificationHandler) NotificationRegistration {
	return nmgr.register(handler, false)
}

// AnyRegistered returns true if there are any handlers registered for the given
// handler type.
func (ngmr *NotificationManager) AnyRegistered(handler NotificationHandler) bool {
	return ngmr.handlers[handler.typ()].AnyRegistered()
}

func (nmgr *NotificationManager) waitAndEmitUINtfn(c <-chan time.Time, cancel <-chan struct{}) {
	select {
	case <-c:
	case <-cancel:
		return
	}

	nmgr.uiMtx.Lock()
	n := nmgr.uiNextNtfn
	nmgr.uiNextNtfn = UINotification{}
	nmgr.uiMtx.Unlock()

	nmgr.handlers[onUINtfnType].(*handlersFor[OnUINotification]).
		visit(func(h OnUINotification) { h(n) })
}

func (nmgr *NotificationManager) addUINtfn(from zkidentity.ShortID, typ UINotificationType, msg string, ts time.Time) {
	nmgr.uiMtx.Lock()

	n := &nmgr.uiNextNtfn
	cfg := &nmgr.uiConfig

	// // Remove embeds.
	// msg = mdembeds.ReplaceEmbeds(msg, func(args mdembeds.EmbeddedArgs) string {
	// 	if strings.HasPrefix(args.Typ, "image/") {
	// 		return "[image]"
	// 	}
	// 	return ""
	// })

	switch {
	case typ == UINtfnWRCreated && !cfg.WRCreated:

		// Ignore
		nmgr.uiMtx.Unlock()
		return

	case typ == UINtfnWRCreated && n.Type == UINtfnWRCreated:
		// First PM.
		n.Type = typ
		n.Count = 1
		n.From = from
		n.Timestamp = ts.Unix()
		n.Text = fmt.Sprintf("wr created by %s: %s", from,
			cfg.clip(msg))

	default:
		// Multiple types.
		n.Type = UINtfnMultiple
		n.FromNick = "multiple"
		n.Count += 1
		n.Text = fmt.Sprintf("%d messages received", n.Count)
	}

	// The first notification starts the timer to emit the actual UI
	// notification. Other notifications will get batched.
	if n.Count == 1 {
		nmgr.uiTimer.Reset(cfg.EmitInterval)
		c, cancel := nmgr.uiTimer.C, cfg.CancelEmissionChannel
		go nmgr.waitAndEmitUINtfn(c, cancel)
	}

	nmgr.uiMtx.Unlock()
}

// Following are the notifyX() calls (one for each type of notification).

func (nmgr *NotificationManager) notifyTest() {
	nmgr.handlers[onTestNtfnType].(*handlersFor[onTestNtfn]).
		visit(func(h onTestNtfn) { h() })
}

func (nmgr *NotificationManager) notifyOnWRCreated(wr *pong.WaitingRoom, ts time.Time) {
	nmgr.handlers[onWRCreatedfnType].(*handlersFor[OnWRCreatedNtfn]).
		visit(func(h OnWRCreatedNtfn) { h(wr, ts) })

	var id zkidentity.ShortID
	id.FromString(wr.HostId)
	// nmgr.addUINtfn(id, id.Nick(), UINtfnPM, pm.Message, ts)
}

func (nmgr *NotificationManager) notifyBetAmtChanged(playerID string, betAmt int64, ts time.Time) {
	nmgr.handlers[onBetAmtChangedFnType].(*handlersFor[OnBetAmtChangedNtfn]).
		visit(func(h OnBetAmtChangedNtfn) { h(playerID, betAmt, ts) })

	var id zkidentity.ShortID
	id.FromString(playerID)
	// nmgr.addUINtfn(id, player.Nick(), UINtfnBetChange, fmt.Sprintf("New bet amount: %d", betAmt), ts)
}

func (nmgr *NotificationManager) notifyGameStarted(gameID string, ts time.Time) {
	nmgr.handlers[onGameStartedFnType].(*handlersFor[OnGameStartedNtfn]).
		visit(func(h OnGameStartedNtfn) { h(gameID, ts) })

	// XXX add ui ntfn for game started
	// nmgr.addUINtfn(id, player.Nick(), UINtfnBetChange, fmt.Sprintf("New bet amount: %d", betAmt), ts)
}

func (nmgr *NotificationManager) notifyPlayerJoinedWR(wr *pong.WaitingRoom, ts time.Time) {
	nmgr.handlers[OnPlayerJoinedNtfnType].(*handlersFor[OnPlayerJoinedNtfn]).
		visit(func(h OnPlayerJoinedNtfn) { h(wr, ts) })
}

func (nmgr *NotificationManager) notifyGameEnded(gameID string, msg string, ts time.Time) {
	nmgr.handlers[onGameEndedfnType].(*handlersFor[OnGameEndedNtfn]).
		visit(func(h OnGameEndedNtfn) { h(gameID, msg, ts) })
}

func (nmgr *NotificationManager) notifyPlayerLeftWR(wr *pong.WaitingRoom, playerID string, ts time.Time) {
	nmgr.handlers[onPlayerLeftNtfnType].(*handlersFor[OnPlayerLeftNtfn]).
		visit(func(h OnPlayerLeftNtfn) { h(wr, playerID, ts) })
}

func NewNotificationManager() *NotificationManager {
	nmgr := &NotificationManager{
		uiConfig: UINotificationsConfig{
			MaxLength:    255,
			EmitInterval: 30 * time.Second,
		},
		uiTimer: time.NewTimer(time.Hour * 24),
		handlers: map[string]handlersRegistry{
			onTestNtfnType:         &handlersFor[onTestNtfn]{},
			onWRCreatedfnType:      &handlersFor[OnWRCreatedNtfn]{},
			onBetAmtChangedFnType:  &handlersFor[OnBetAmtChangedNtfn]{},
			onGameStartedFnType:    &handlersFor[OnGameStartedNtfn]{},
			OnPlayerJoinedNtfnType: &handlersFor[OnPlayerJoinedNtfn]{},
			onGameEndedfnType:      &handlersFor[OnGameEndedNtfn]{},
			onPlayerLeftNtfnType:   &handlersFor[OnPlayerLeftNtfn]{},

			onUINtfnType: &handlersFor[OnUINotification]{},
		},
	}
	if !nmgr.uiTimer.Stop() {
		<-nmgr.uiTimer.C
	}

	return nmgr
}
