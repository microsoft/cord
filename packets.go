package cord

import "encoding/json"

// A Payload structure is the basic structure in which information is sent
// to and from the Discord gateway.
type Payload struct {
	Operation Operation       `json:"op"`
	Data      json.RawMessage `json:"d"`
	// Provided only for Dispatch operations:
	Sequence uint64 `json:"s"`
	Event    string `json:"t"`
}

// gatewayResponse is returned from /gateway on Discord's API.
type gatewayResponse struct {
	URL string `json:"url"`
}

// An Operation is contained in a Payload and defines what should occur
// as a result of that payload.
type Operation uint8

const (
	// dispatches an event
	Dispatch Operation = iota
	// used for ping checking
	Heartbeat
	// used for client handshake
	Identify
	// used to update the client status
	StatusUpdate
	// used to join/move/leave voice channels
	VoiceStatusUpdate
	// used for voice ping checking
	VoiceServerPing
	// used to resume a closed connection
	Resume
	// used to redirect clients to a new gateway
	Reconnect
	// used to request guild members
	RequestMembers
	// used to notify client they have an invalid session id
	InvalidSession
)
