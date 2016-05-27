package cord

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"sync"
	"sync/atomic"
	"time"

	"github.com/WatchBeam/cord/model"
	"github.com/cenk/backoff"
	"github.com/gorilla/websocket"
)

// Identifying bytes for gzip that are used to signal we need to decompress
// the payload.
var gzipSignature = []byte{0x1f, 0x8b}

type WsOptions struct {
	// Handshake packet to send to the server. Note that `compress` and
	// `properties` will be filled for you.
	Handshake *model.Handshake

	// How often to send ping frames to the server if we don't get any
	// other messages. Defaults to 5 seconds.
	PingInterval time.Duration

	// How long to wait without frames or acknowledgment before we consider
	// the server to be dead. Should be longer than the PingInterval.
	// Defaults to twice the PingInterval.
	Timeout time.Duration

	// Backoff determines how long to wait between reconnections to the
	// websocket server. Defaults to an exponential backoff.
	Backoff backoff.BackOff

	// Dialer to use for the websocket. Defaults to a dialer with the
	// `timeout` duration.
	Dialer *websocket.Dialer

	// The retriever to get the gateway to connect to. Defaults to the
	// HTTPGatewayRetriever with the given `timeout`.
	Gateway GatewayRetriever

	// Headers to send in the websocket handshake.
	Header http.Header
}

func (w *WsOptions) fillDefaults(token string) {
	if w.PingInterval == 0 {
		w.PingInterval = time.Second * 5
	}

	if w.Timeout == 0 {
		w.Timeout = w.PingInterval * 2
	}

	if w.Backoff == nil {
		eb := backoff.NewExponentialBackOff()
		eb.InitialInterval = time.Millisecond * 500
		eb.RandomizationFactor = 1
		eb.Multiplier = 2
		eb.MaxInterval = time.Second * 10
		w.Backoff = eb
	}

	if w.Dialer == nil {
		w.Dialer = &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: w.Timeout,
		}
	}

	if w.Gateway == nil {
		w.Gateway = HTTPGatewayRetriever{
			Client:  &http.Client{Timeout: w.Timeout},
			BaseURL: "https://discordapp.com/api",
		}
	}

	if w.Handshake == nil {
		w.Handshake = &model.Handshake{}
	}

	w.Handshake.Compress = true
	w.Handshake.Token = token
	w.Handshake.Properties = model.HandshakeProperties{
		OS:      runtime.GOOS,
		Browser: "Cord 1.0",
	}
}

// Websocket is an implementation of the Socket interface.
type Websocket struct {
	opts   *WsOptions
	events events

	wsMu      sync.Mutex
	ws        *websocket.Conn
	sessionID string
	lastSeq   uint64

	outgoing chan []byte
	errs     chan error
	closer   chan struct{}
}

// start boots the websocket asynchronously.
func (w *Websocket) start() {
	w.events.On(Ready(func(r *model.Ready) {
		w.wsMu.Lock()
		defer w.wsMu.Unlock()
		w.sessionID = r.SessionID
	}))

	go w.restart(nil)
}

// restart closes the server and attempts to reconnect to Discord. It takes
// an optional error to log down.
func (w *Websocket) restart(err error) {
	// just return if we manually send .Close()
	if err == websocket.ErrCloseSent {
		return
	} else if err != nil {
		w.errs <- err
	}

	// Look up the websocket address to connect to.
	gateway, err := w.opts.Gateway.Gateway()
	if err != nil {
		w.restart(err)
		return
	}

	w.wsMu.Lock()
	if w.ws != nil {
		w.ws.Close()
		w.ws = nil
	}
	w.wsMu.Unlock()

	select {
	case <-time.After(w.opts.Backoff.NextBackOff()):
	case <-w.closer:
		return
	}

	w.establishSocketConnection(gateway)
}

func (w *Websocket) establishSocketConnection(gateway string) {
	ws, _, err := w.opts.Dialer.Dial(gateway, w.opts.Header)
	if err != nil {
		w.restart(err)
		return
	}

	w.wsMu.Lock()
	defer w.wsMu.Unlock()

	w.opts.Backoff.Reset()
	w.ws = ws
	go w.readPump(ws)
	go w.writePump(ws)

	if err := w.sendHandshake(); err != nil {
		w.errs <- err
	}
}

// sendHandshake dispatches either an Identify or Resume packet on the
// connection, depending whether we were connected before.
func (w *Websocket) sendHandshake() error {
	if w.sessionID == "" {
		return w.Send(Identify, w.opts.Handshake)
	}

	return w.Send(Resume, &model.Resume{
		Token:     w.opts.Handshake.Token,
		SessionID: w.sessionID,
		Sequence:  atomic.LoadUint64(&w.lastSeq),
	})
}

// readPump reads off messages from the socket and dispatches them into the
// handleIncoming method.
func (w *Websocket) readPump(ws *websocket.Conn) {
	for {
		ws.SetReadDeadline(time.Now().Add(w.opts.Timeout))
		kind, message, err := ws.ReadMessage()
		if err != nil {
			w.restart(err)
			return
		}

		// Control frames won't have associated messages, only care about
		// binary or text messages.
		if kind == websocket.TextMessage || kind == websocket.BinaryMessage {
			go w.handleIncoming(message)
		}
	}
}

// writePump
func (w *Websocket) writePump(ws *websocket.Conn) {
	ticker := time.NewTicker(w.opts.PingInterval)
	defer ticker.Stop()

	for {
		var (
			err error
			msg []byte
		)

		select {
		case <-ticker.C:
			err = ws.WriteMessage(websocket.PingMessage, nil)
		case msg = <-w.outgoing:
			err = ws.WriteMessage(websocket.TextMessage, msg)
		}

		if err != nil {
			if msg != nil {
				select {
				case w.outgoing <- msg:
				case <-w.closer:
				}
			}

			return
		}
	}
}

// inflate decompresses the provided zlib-compressed bytes
func inflate(b []byte) ([]byte, error) {
	r, err := zlib.NewReader(bytes.NewReader(b))
	if err != nil {
		return nil, err
	}

	return ioutil.ReadAll(r)
}

// unmarshalPayload parses and extracts the payload from the byte slice.
func (w *Websocket) unmarshalPayload(b []byte) (*Payload, error) {
	if bytes.HasPrefix(b, gzipSignature) {
		var err error
		if b, err = inflate(b); err != nil {
			return nil, err
		}
	}

	wrapper := &Payload{}
	if err := wrapper.UnmarshalJSON(b); err != nil {
		return nil, err
	}

	return wrapper, nil
}

// handleIncoming processes a message from the websocket and dispatches
// it to clients.
func (w *Websocket) handleIncoming(b []byte) {
	wrapper, err := w.unmarshalPayload(b)
	if err != nil {
		w.errs <- fmt.Errorf("cord/websocket: error unpacking payload: %s", err)
		return
	}

	switch wrapper.Operation {
	case Dispatch:
		atomic.StoreUint64(&w.lastSeq, wrapper.Sequence)
		if err := w.events.Dispatch(wrapper.Event, wrapper.Data); err != nil {
			w.errs <- fmt.Errorf("cord/websocket: error dispatching event: %s", err)
		}
	case Reconnect:
		w.restart(nil)
	case InvalidSession:
		w.restart(fmt.Errorf("cord/websocket: invalid session detected"))
	default:
		w.errs <- fmt.Errorf("cord/websocket: unhandled op code %d", wrapper.Operation)
	}
}

// On implements Socket.On
func (w *Websocket) On(h Handler) { w.events.On(h) }

// Off implements Socket.Off
func (w *Websocket) Off(h Handler) { w.events.Off(h) }

// Once implements Socket.Once
func (w *Websocket) Once(h Handler) { w.events.Once(h) }

// Errs implements Socket.Errs
func (w *Websocket) Errs() <-chan error { return w.errs }

// Send implements Socket.Send
func (w *Websocket) Send(op Operation, data json.Marshaler) error {
	bytes, err := data.MarshalJSON()
	if err != nil {
		return err
	}

	wrapper, err := (&Payload{
		Operation: op,
		Data:      bytes,
	}).MarshalJSON()

	if err != nil {
		return err
	}

	w.outgoing <- wrapper
	return nil
}

// Close frees resources associated with the websocket.
func (w *Websocket) Close() error {
	w.wsMu.Lock()
	defer w.wsMu.Unlock()

	w.closer <- struct{}{}
	if w.ws != nil {
		return w.ws.Close()
	}

	return nil
}
