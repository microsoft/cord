package cord

import (
	"io/ioutil"
	"net/http"
)

// GatewayRetriever calls the Discord API and returns the socket URL to
// connect to.
type GatewayRetriever interface {
	// Gateway returns the gateway URL to connect to.
	Gateway() (url string, err error)
}

// HTTPGatewayRetriever is an implementation of the GatewayRetriever that
// looks up the gateway from Discord's REST API.
type HTTPGatewayRetriever struct {
	Client  *http.Client
	BaseURL string
}

// Gateway implements GatewayRetriever.Gateway
func (h HTTPGatewayRetriever) Gateway() (string, error) {
	res, err := h.Client.Get(h.BaseURL + "/gateway")
	if err != nil {
		return "", err
	}
	defer res.Body.Close()

	b, err := ioutil.ReadAll(res.Body)
	if err != nil {
		return "", err
	}

	data := &gatewayResponse{}
	if err := data.UnmarshalJSON(b); err != nil {
		return "", err
	}

	return data.URL, nil
}
