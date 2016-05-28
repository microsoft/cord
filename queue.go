package cord

import "sync"

type queuedMessage struct {
	data   *Payload
	result chan error
}

type queue struct {
	mu     *sync.Cond
	forks  []*queue
	items  []*queuedMessage
	closer chan struct{}
}

func newQueue() *queue {
	return &queue{
		mu:     sync.NewCond(&sync.Mutex{}),
		closer: make(chan struct{}),
	}
}

// Push appends a new item to the queue. Handshakes packets, however, are
// always prepended.
func (q *queue) Push(msg *queuedMessage) {
	q.mu.L.Lock()
	defer q.mu.L.Unlock()

	for _, fork := range q.forks {
		fork.Push(msg)
	}

	q.items = append(q.items, msg)
	q.mu.Broadcast()
}

// Close signals that no further messages may be expected on this queue.
func (q *queue) Close() {
	close(q.closer)
}

// Poll returns a channel that blocks until a message is available on
// the queue closed.
func (q *queue) Poll() <-chan *queuedMessage {
	ch := make(chan *queuedMessage)
	go func() {
		q.mu.L.Lock()
		defer q.mu.L.Unlock()
		defer close(ch)

		for len(q.items) == 0 {
			select {
			case <-q.closer:
				return
			default:
			}

			q.mu.Wait()
		}

		head := q.items[0]
		q.items = q.items[1:]
		ch <- head
	}()

	return ch
}

// Fork creates a new queue that inherits all current *and future* items
// from this queue.
func (q *queue) Fork() *queue {
	q.mu.L.Lock()
	defer q.mu.L.Unlock()

	fork := newQueue()
	fork.items = make([]*queuedMessage, len(q.items))
	copy(fork.items, q.items)
	q.forks = append(q.forks, fork)

	return fork
}
