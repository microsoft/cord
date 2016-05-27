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
	defer ts.Close()

	gw, err := HTTPGatewayRetriever{
		Client:  http.DefaultClient,
		BaseURL: ts.URL,
	}.Gateway()

	assert.Nil(t, err)
	assert.Equal(t, gw, "wss://gateway.discord.gg")
}

func TestGatewayErrorsOnBadPacket(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"url":"wss://ga`)
	}))
	defer ts.Close()

	_, err := HTTPGatewayRetriever{
		Client:  http.DefaultClient,
		BaseURL: ts.URL,
	}.Gateway()

	assert.NotNil(t, err)
}
func TestGatewayPropogateHTTPError(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprintln(w, `{"url":"wss://gateway.discord.gg"}`)
	}))
	defer ts.Close()

	_, err := HTTPGatewayRetriever{
		Client:  &http.Client{Timeout: time.Nanosecond},
		BaseURL: ts.URL,
	}.Gateway()

	assert.NotNil(t, err)
}
