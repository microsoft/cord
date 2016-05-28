package cord

// A Debugger can be passed into the options to be notified of all socket
// sends and receives.
type Debugger interface {
	// Incoming is called with the raw packet string sent to cord, after
	// inflation for gzipped strings.
	Incoming(b []byte)

	// Outgoing is called with data when a packet is sent on cord.
	Outgoing(b []byte)

	// Error is called when an error occurs on the socket. The error
	// is ALSO sent down the Errs() channel for your
	Error(error)
}

// NilDebugger is the default debugger with noops.
type NilDebugger struct{}

// Incoming implements Debugger.Incoming
func (n NilDebugger) Incoming(b []byte) {}

// Outgoing implements Debugger.Outgoing
func (n NilDebugger) Outgoing(b []byte) {}

// Error implements Debugger.Error
func (n NilDebugger) Error(e error) {}