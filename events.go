package cord

import "sync"

type handlerList []Handler

func (h handlerList) Delete(handler Handler) []Handler {
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

// events is a simple eventemitter-like interface which contains Handler
// interfaces.
type events struct {
	mu       *sync.Mutex
	onces    map[string]handlerList
	handlers map[string]handlerList
}

func newEvents() events {
	return events{
		mu:       new(sync.Mutex),
		onces:    make(map[string]handlerList),
		handlers: make(map[string]handlerList),
	}
}

// On attaches a Handler so that it's called every time an event is received.
func (e *events) On(h Handler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.handlers[h.Name()] = append(e.handlers[h.Name()], h)
}

// Onces attaches a handler that's called the next time the event is received,
// then immediately removed.
func (e *events) Once(h Handler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.onces[h.Name()] = append(e.onces[h.Name()], h)
}

// Off removes a listening handler.
func (e *events) Off(h Handler) {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.handlers[h.Name()] = e.handlers[h.Name()].Delete(h)
	e.onces[h.Name()] = e.onces[h.Name()].Delete(h)
}

// Dispatch invokes all handlers listening on the event with the `b` bytes.
func (e *events) Dispatch(event string, b []byte) error {
	e.mu.Lock()
	l1, l2 := e.handlers[event], e.onces[event]
	e.onces[event] = nil

	list := make([]Handler, len(l1)+len(l2))
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
