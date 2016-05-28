package cord

import (
	"sync"

	"github.com/WatchBeam/cord/events"
)

type handlerList []events.Handler

func (h handlerList) Delete(handler events.Handler) []events.Handler {
	for i, other := range h {
		if other != handler {
			continue
		}

		h[i] = h[len(h)-1]
		h[len(h)-1] = nil
		return h[:len(h)-1]
	}

	return h
}

// emitter is a simple eventemitter-like interface which
// contains events.Handler interfaces.
type emitter struct {
	mu       *sync.Mutex
	onces    map[string]handlerList
	handlers map[string]handlerList
}

func newEmitter() emitter {
	return emitter{
		mu:       new(sync.Mutex),
		onces:    make(map[string]handlerList),
		handlers: make(map[string]handlerList),
	}
}

// On attaches a events.Handler so that it's called every time an event is received.
func (e *emitter) On(h events.Handler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.handlers[h.Name()] = append(e.handlers[h.Name()], h)
}

// Onces attaches a handler that's called the next time the event is received,
// then immediately removed.
func (e *emitter) Once(h events.Handler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.onces[h.Name()] = append(e.onces[h.Name()], h)
}

// Off removes a listening handler.
func (e *emitter) Off(h events.Handler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.handlers[h.Name()] = e.handlers[h.Name()].Delete(h)
	e.onces[h.Name()] = e.onces[h.Name()].Delete(h)
}

// Dispatch invokes all handlers listening on the event with the `b` bytes.
func (e *emitter) Dispatch(event string, b []byte) error {
	e.mu.Lock()
	l1, l2 := e.handlers[event], e.onces[event]
	e.onces[event] = nil

	list := make([]events.Handler, len(l1)+len(l2))
	copy(list, l1)
	copy(list[len(l1):], l2)
	e.mu.Unlock()

	for _, handler := range list {
		if err := handler.Invoke(b); err != nil {
			return err
		}
	}

	return nil
}
