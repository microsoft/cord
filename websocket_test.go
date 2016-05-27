package cord

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/WatchBeam/cord/model"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/suite"
)

type WebsocketSuite struct {
	suite.Suite
	ts        *httptest.Server
	retriever GatewayRetriever
	onConnect func(c *websocket.Conn)
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
		w.onConnect(c)
	}))

	w.closer = make(chan struct{})
	w.socket = New("tooken", &WsOptions{
		Gateway: testGatewayRetriever{strings.Replace(w.ts.URL, "http://", "ws://", 1)},
	}).(*Websocket)
	w.socket.opts.Handshake.Properties.OS = "darwin" // normalize for testing
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
	num := 0

	w.onConnect = func(c *websocket.Conn) {
		_, msg, err := c.ReadMessage()
		w.Nil(err)

		if num == 0 {
			// first connect
			w.Equal(`{"op":2,"d":{"token":"tooken","properties":{"$os":"darwin",`+
				`"$browser":"Cord 1.0","$device":"","$referer":"",`+
				`"$referring_domain":""},"large_threshold":0,"compress":true,`+
				`"shard":[]},"s":0,"t":""}`, string(msg))
		} else {
			// reconnect
			w.Equal(`{"op":6,"d":{"token":"tooken","session_id":"asdf",`+
				`"seq":1},"s":0,"t":""}`, string(msg))
		}
		num++

		c.WriteMessage(websocket.TextMessage, []byte(`{
            "op":0,
            "t": "READY",
            "s": 1,
            "d": {"session_id": "asdf"}
        }`))
		c.Close()
	}

	done := make(chan struct{})
	w.socket.Once(Ready(func(r *model.Ready) {
		w.Equal("asdf", w.socket.sessionID)
		w.Equal(uint64(1), w.socket.lastSeq)
		// closing the underlying connection will result in an EOF error
		w.socket.ws.UnderlyingConn().Close()
		w.IsType(&websocket.CloseError{}, <-w.socket.Errs())

		w.socket.Once(Ready(func(r *model.Ready) {
			close(done)
		}))
	}))

	<-done
}

func TestWebsocketSuite(t *testing.T) {
	suite.Run(t, new(WebsocketSuite))
}
