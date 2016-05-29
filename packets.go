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
	// Dispatch is an operation used to dispatch  an event
	Dispatch Operation = iota
	// Heartbeat is an operation used for ping checking
	Heartbeat
	// Identify is an operation used for client handshake
	Identify
	// StatusUpdate is an operation used to update the client status
	StatusUpdate
	// VoiceStatusUpdate is an operation used to join/move/leave voice channels
	VoiceStatusUpdate
	// VoiceServerPing is an operation used for voice ping checking
	VoiceServerPing
	// Resume is an operation used to resume a closed connection
	Resume
	// Reconnect is an operation used to redirect clients to a new gateway
	Reconnect
	// RequestMembers is an operation used to request guild members
	RequestMembers
	// InvalidSession is an operation used to notify
	// client they have an invalid session id
	InvalidSession
)
