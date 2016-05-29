package cord

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestGatewayReadsGood(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.URL.Path, "/gateway")
		fmt.Fprintln(w, `{"url":"wss://gateway.discord.gg"}`)
	}))

	gw, err := HTTPGatewayRetriever{
		Client:  http.DefaultClient,
		BaseURL: ts.URL,
	}.Gateway()

	assert.Nil(t, err)
	assert.Equal(t, gw, "wss://gateway.discord.gg")
	ts.Close()
}

func TestGatewayErrorsOnBadPacket(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"url":"wss://ga`)
	}))

	_, err := HTTPGatewayRetriever{
		Client:  http.DefaultClient,
		BaseURL: ts.URL,
	}.Gateway()

	assert.NotNil(t, err)
	ts.Close()
}
func TestGatewayPropogateHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		fmt.Fprintln(w, `{"url":"wss://gateway.discord.gg"}`)
	}))

	_, err := HTTPGatewayRetriever{
		Client:  &http.Client{Timeout: time.Nanosecond},
		BaseURL: ts.URL,
	}.Gateway()

	assert.NotNil(t, err)
	ts.Close()
}
