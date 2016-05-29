package cord

import (
	"bytes"
	"compress/zlib"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WatchBeam/cord/events"
	"github.com/WatchBeam/cord/model"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
)

var (
	readyPacket = []byte(`{
        "op":0,
        "t": "READY",
        "s": 1,
        "d": {"session_id": "asdf", "heartbeat_interval": 10000}
    }`)
)

type WebsocketSuite struct {
	suite.Suite
	ts        *httptest.Server
	retriever GatewayRetriever
	onConnect chan func(c *websocket.Conn)
	socket    *Websocket
	closer    chan struct{}
}

type testGatewayRetriever struct{ gateway string }

func (t testGatewayRetriever) Gateway() (string, error) { return t.gateway, nil }

func (w *WebsocketSuite) SetupTest() {
	w.ts = httptest.NewServer(http.HandlerFunc(func(rw http.ResponseWriter, r *http.Request) {
		c, err := (&websocket.Upgrader{}).Upgrade(rw, r, nil)
		if err != nil {
			panic(err)
		}
		(<-w.onConnect)(c)
	}))

	w.closer = make(chan struct{})
	w.onConnect = make(chan func(c *websocket.Conn), 16)
	w.socket = New("tooken", &WsOptions{
		Gateway: testGatewayRetriever{strings.Replace(w.ts.URL, "http://", "ws://", 1)},
		Handshake: &model.Handshake{
			Properties: model.HandshakeProperties{OS: "darwin"},
		},
	}).(*Websocket)
}

func (w *WebsocketSuite) panicOnError() {
	for {
		select {
		case err := <-w.socket.Errs():
			panic(err)
		case <-w.closer:
			return
		}
	}
}

func (w *WebsocketSuite) TeardownTest() {
	close(w.closer)
	w.socket.Close()
	w.ts.Close()
}

func (w *WebsocketSuite) TestHandshakesAndReconnectsCorrectly() {
	w.onConnect <- func(c *websocket.Conn) {
		_, msg, err := c.ReadMessage()
		w.Nil(err)
		w.Equal(`{"op":2,"d":{"token":"tooken","properties":{"$os":"darwin",`+
			`"$browser":"Cord 1.0","$device":"","$referer":"",`+
			`"$referring_domain":""},"compress":true,"large_threshold":0},`+
			`"s":0,"t":""}`, string(msg))
		c.WriteMessage(websocket.TextMessage, readyPacket)
		c.Close()
	}

	w.onConnect <- func(c *websocket.Conn) {
		_, msg, err := c.ReadMessage()
		w.Nil(err)
		w.Equal(`{"op":6,"d":{"token":"tooken","session_id":"asdf",`+
			`"seq":0},"s":0,"t":""}`, string(msg))
		c.WriteMessage(websocket.TextMessage, readyPacket)
		c.Close()
	}

	done := make(chan struct{})
	w.socket.Once(events.Ready(func(r *model.Ready) {
		w.Equal("asdf", r.SessionID)
		// closing the underlying connection will result in an EOF error
		w.IsType(DisruptionError{}, <-w.socket.Errs())

		w.socket.Once(events.Ready(func(r *model.Ready) {
			close(done)
		}))
	}))

	<-done
}

func (w *WebsocketSuite) TestLogsInvalidTokenAsFatalError() {
	w.onConnect <- func(c *websocket.Conn) {
		c.ReadMessage()
		c.WriteMessage(websocket.CloseMessage, websocket.FormatCloseMessage(4004, "Authentication"))
		c.Close()
	}

	w.IsType(FatalError{}, <-w.socket.Errs())
}

func (w *WebsocketSuite) TestRetriesTokenOnInvalidSession() {
	w.onConnect <- func(c *websocket.Conn) {
		// initially it'll send the token
		_, msg, err := c.ReadMessage()
		w.Nil(err)
		w.Contains(string(msg), `"op":2`)
		c.WriteMessage(websocket.TextMessage, readyPacket)
		c.Close()
	}

	w.onConnect <- func(c *websocket.Conn) {
		// when we restart we won't respond with a ready and close the socket
		// like Discord does.
		_, msg, err := c.ReadMessage()
		w.Contains(string(msg), `"op":6`)
		w.Nil(err)
		c.WriteMessage(websocket.TextMessage, []byte(`{"op":9}`))

		_, msg, err = c.ReadMessage()
		w.Nil(err)
		w.Contains(string(msg), `"op":2`)
		c.WriteMessage(websocket.TextMessage, readyPacket)

		c.Close()
	}

	done := make(chan struct{})
	w.socket.Once(events.Ready(func(r *model.Ready) {
		w.IsType(DisruptionError{}, <-w.socket.Errs())

		w.socket.Once(events.Ready(func(r *model.Ready) {
			close(done)
		}))
	}))

	<-done
}

func (w *WebsocketSuite) TestReadsGzippedData() {
	w.onConnect <- func(c *websocket.Conn) {
		_, _, err := c.ReadMessage()
		w.Nil(err)

		var b bytes.Buffer
		zw := zlib.NewWriter(&b)
		zw.Write(readyPacket)
		zw.Close()
		c.WriteMessage(websocket.BinaryMessage, b.Bytes())
	}

	go w.panicOnError()

	done := make(chan struct{})
	w.socket.Once(events.Ready(func(r *model.Ready) {
		w.Equal("asdf", r.SessionID)
		close(done)
	}))

	<-done
}

func TestWebsocketSuite(t *testing.T) {
	suite.Run(t, new(WebsocketSuite))
}
