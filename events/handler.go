package events

// Handler defines a type that can be passed into a Socket to listen for
// an event being broadcasted.
type Handler interface {
	// Name returns the name of the packet that this handler process, the
	// "t" key in Discord payloads.
	Name() string
	// Invoke is called with the raw, still-marshalled byte payload from
	// the socket. It may return an error if unmarshalling fails.
	Invoke(b []byte) error
}
