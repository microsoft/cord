package cord

import (
	"bytes"
	"compress/zlib"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"runtime"
	"strconv"
	"sync/atomic"
	"time"
	"unsafe"

	"github.com/WatchBeam/cord/events"
	"github.com/WatchBeam/cord/model"
	"github.com/cenk/backoff"
	"github.com/gorilla/websocket"
)

type WsOptions struct {
	// Handshake packet to send to the server. Note that `compress` and
	// `properties` will be filled for you.
	Handshake *model.Handshake

	// How long to wait without frames or acknowledgment before we consider
	// the server to be dead. Defaults to ten seconds.
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

	// Debugger struct we log incoming/outgoing messages to.
	Debugger Debugger

	// Headers to send in the websocket handshake.
	Header http.Header
}

func (w *WsOptions) fillDefaults(token string) {
	if w.Timeout == 0 {
		w.Timeout = 10 * time.Second
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

	if w.Debugger == nil {
		w.Debugger = NilDebugger{}
	}

	w.Handshake.Compress = true
	w.Handshake.Token = token
	w.Handshake.Properties = model.HandshakeProperties{
		OS:      runtime.GOOS,
		Browser: "Cord 1.0",
	}
}

// wsConn is a struct atomically stored within a Websocket, containing a
// websocket connection and a queue of messages to send. When a restart
// happens, the queue is forked and the websocket connection is
// reestablished in a new wsConn struct.
type wsConn struct {
	ws    *websocket.Conn
	queue *queue
}

// Close closes the associated websocket and queue.
func (w *wsConn) Close() error {
	if w == nil {
		return nil
	}

	w.queue.Close()
	if w.ws != nil {
		return w.ws.Close()
	}

	return nil
}

// Fork creates a new wsConn whose queue inherits from the current one.
// The websocket itself will be nil.
func (w *wsConn) Fork() *wsConn {
	if w == nil {
		return &wsConn{queue: newQueue()}
	}

	return &wsConn{queue: w.queue.Fork()}
}

// Websocket is an implementation of the Socket interface.
type Websocket struct {
	opts   *WsOptions
	events emitter

	// ws points to a wsConn, atomically updated
	ws        unsafe.Pointer
	sessionID unsafe.Pointer
	lastSeq   uint64 // atomically updated
	errs      chan error
}

// start boots the websocket asynchronously.
func (w *Websocket) start() { go w.restart(nil, nil) }

// restart closes the server and attempts to reconnect to Discord. It takes
// an optional error to log down.
func (w *Websocket) restart(err error, prev *wsConn) {
	next := prev.Fork()

	// If someone already restarted or closed us, do nothing.
	if !atomic.CompareAndSwapPointer(&w.ws, unsafe.Pointer(prev), unsafe.Pointer(next)) {
		return
	}
	prev.Close()

	if err != nil {
		w.sendErr(err)
	}

	// Look up the websocket address to connect to.
	gateway, err := w.opts.Gateway.Gateway()
	if err != nil {
		w.restart(err, next)
		return
	}

	// Wait for a short while then reestablished the connection. Note that
	// the atomic playing we do with the *wsConn establishes a thread-safety
	// around the backoff interface.
	time.Sleep(w.opts.Backoff.NextBackOff())
	w.establishSocketConnection(gateway, next)
}

func (w *Websocket) establishSocketConnection(gateway string, cnx *wsConn) {
	ws, _, err := w.opts.Dialer.Dial(gateway, w.opts.Header)
	if err != nil {
		w.restart(err, cnx)
		return
	}

	ready, err := w.runHandshake(ws)
	if err != nil {
		w.restart(err, cnx)
		return
	}

	next := &wsConn{
		queue: cnx.queue,
		ws:    ws,
	}

	// Note: we store a new pointer rather than updating the cnx because
	// someone else might have read the wsConn pointer in the meantime.
	atomic.StorePointer(&w.ws, unsafe.Pointer(unsafe.Pointer(next)))
	w.opts.Backoff.Reset()

	atomic.StorePointer(&w.sessionID, unsafe.Pointer(&ready.SessionID))
	interval := time.Duration(ready.HeartbeatInterval) * time.Millisecond

	go w.readPump(next)
	go w.writePump(next, interval)
}

// sendHandshake dispatches either an Identify or Resume packet on the
// connection, depending whether we were connected before.
func (w *Websocket) runHandshake(ws *websocket.Conn) (*model.Ready, error) {
	var (
		sid   = (*string)(atomic.LoadPointer(&w.sessionID))
		data  *Payload
		err   error
		ready *model.Ready
	)

	if sid == nil {
		data, err = w.marshalPayload(Identify, w.opts.Handshake)
	} else {
		data, err = w.marshalPayload(Resume, &model.Resume{
			Token:     w.opts.Handshake.Token,
			SessionID: *sid,
			Sequence:  atomic.LoadUint64(&w.lastSeq),
		})
	}
	if err != nil {
		return nil, err
	}

	if err = w.writeMessage(ws, data); err != nil {
		return nil, err
	}

	ws.SetReadDeadline(time.Now().Add(w.opts.Timeout))
	_, message, err := ws.ReadMessage()
	if err != nil {
		return nil, err
	}

	payload, err := w.unmarshalPayload(message)
	if err != nil {
		return nil, err
	}
	if payload.Event != "READY" {
		return nil, fmt.Errorf("cord/websocket: expected to get READY event, got %s", payload)
	}

	err = events.Ready(func(r *model.Ready) { ready = r }).Invoke(payload.Data)
	go w.events.Dispatch(payload.Event, payload.Data)

	return ready, err
}

// readPump reads off messages from the socket and dispatches them into the
// handleIncoming method.
func (w *Websocket) readPump(cnx *wsConn) {
	cnx.ws.SetReadDeadline(time.Time{})

	for {
		kind, message, err := cnx.ws.ReadMessage()
		if err != nil {
			w.restart(err, cnx)
			return
		}

		// Control frames won't have associated messages, only care about
		// binary or text messages.
		if kind == websocket.TextMessage || kind == websocket.BinaryMessage {
			go w.handleIncoming(message, cnx)
		}
	}
}

func (w *Websocket) writeMessage(ws *websocket.Conn, data json.Marshaler) error {
	bytes, err := data.MarshalJSON()
	if err != nil {
		return err
	}

	ws.SetWriteDeadline(time.Now().Add(w.opts.Timeout))
	w.opts.Debugger.Outgoing(bytes)
	return ws.WriteMessage(websocket.TextMessage, bytes)
}

func (w *Websocket) writePump(cnx *wsConn, heartbeat time.Duration) {
	ticker := time.NewTicker(heartbeat)
	defer ticker.Stop()

	for {
		var err error

		select {
		case <-ticker.C:
			seq := atomic.LoadUint64(&w.lastSeq)
			err = w.writeMessage(cnx.ws, &Payload{
				Operation: Heartbeat,
				Data:      json.RawMessage(strconv.FormatUint(seq, 10)),
			})

		case msg, ok := <-cnx.queue.Poll():
			if !ok {
				return
			}
			err = w.writeMessage(cnx.ws, msg.data)
			msg.result <- err
		}

		if err != nil {
			w.restart(err, cnx)
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
	if len(b) > 0 && b[0] != '{' && b[0] != '[' {
		var err error
		if b, err = inflate(b); err != nil {
			return nil, err
		}
	}

	w.opts.Debugger.Incoming(b)

	wrapper := &Payload{}
	if err := wrapper.UnmarshalJSON(b); err != nil {
		return nil, err
	}

	return wrapper, nil
}

// sendErr dispatches an error on the socket and notifies the debugger.
func (w *Websocket) sendErr(err error) {
	w.opts.Debugger.Error(err)
	w.errs <- err
}

// handleIncoming processes a message from the websocket and dispatches
// it to clients.
func (w *Websocket) handleIncoming(b []byte, cnx *wsConn) {
	wrapper, err := w.unmarshalPayload(b)
	if err != nil {
		w.sendErr(fmt.Errorf("cord/websocket: error unpacking payload: %s", err))
		return
	}

	switch wrapper.Operation {
	case Dispatch:
		atomic.StoreUint64(&w.lastSeq, wrapper.Sequence)
		if err := w.events.Dispatch(wrapper.Event, wrapper.Data); err != nil {
			w.sendErr(fmt.Errorf("cord/websocket: error dispatching event: %s", err))
		}
	case Reconnect:
		w.restart(nil, cnx)
	case InvalidSession:
		atomic.StorePointer(&w.sessionID, unsafe.Pointer(nil))
		w.restart(fmt.Errorf("cord/websocket: invalid session detected"), cnx)
	default:
		w.sendErr(fmt.Errorf("cord/websocket: unhandled op code %d", wrapper.Operation))
	}
}

// On implements Socket.On
func (w *Websocket) On(h events.Handler) { w.events.On(h) }

// Off implements Socket.Off
func (w *Websocket) Off(h events.Handler) { w.events.Off(h) }

// Once implements Socket.Once
func (w *Websocket) Once(h events.Handler) { w.events.Once(h) }

// Errs implements Socket.Errs
func (w *Websocket) Errs() <-chan error { return w.errs }

// marshalPayload marshals the provided data for transport over the socket.
func (w *Websocket) marshalPayload(op Operation, data json.Marshaler) (*Payload, error) {
	bytes, err := data.MarshalJSON()
	if err != nil {
		return nil, err
	}

	return &Payload{
		Operation: op,
		Data:      bytes,
	}, nil
}

// Send implements Socket.Send
func (w *Websocket) Send(op Operation, data json.Marshaler) error {
	payload, err := w.marshalPayload(op, data)
	if err != nil {
		return err
	}

	result := make(chan error, 1)
	cnx := (*wsConn)(atomic.LoadPointer(&w.ws))
	cnx.queue.Push(&queuedMessage{payload, result})
	return <-result
}

// Close frees resources associated with the websocket.
func (w *Websocket) Close() error {
	cnx := (*wsConn)(atomic.SwapPointer(&w.ws, unsafe.Pointer(nil)))
	if cnx == nil {
		return nil
	}

	return cnx.Close()
}
