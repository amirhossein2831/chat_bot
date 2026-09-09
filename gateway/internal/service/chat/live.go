package chat

import (
	"context"
	"sync"

	"github.com/google/uuid"

	"github.com/amirhossein2831/chat_bot/gateway/internal/client/provider"
)

const subscriberBuffer = 64

// LiveStream fans one generation out to zero or more subscribers and keeps what
// has already been emitted, so a client that reconnects mid-answer sees the
// whole reply rather than only the remainder.
type LiveStream struct {
	cancel context.CancelFunc
	done   chan struct{}

	mu          sync.Mutex
	emitted     []provider.Delta
	subscribers map[chan provider.Delta]struct{}
	closed      bool
}

func newLiveStream(cancel context.CancelFunc) *LiveStream {
	return &LiveStream{
		cancel:      cancel,
		done:        make(chan struct{}),
		subscribers: make(map[chan provider.Delta]struct{}),
	}
}

func (l *LiveStream) Subscribe() <-chan provider.Delta {
	l.mu.Lock()
	defer l.mu.Unlock()

	out := make(chan provider.Delta, subscriberBuffer+len(l.emitted))
	for _, delta := range l.emitted {
		out <- delta
	}

	if l.closed {
		close(out)
		return out
	}

	l.subscribers[out] = struct{}{}
	return out
}

func (l *LiveStream) publish(delta provider.Delta) {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return
	}
	l.emitted = append(l.emitted, delta)

	for subscriber := range l.subscribers {
		select {
		case subscriber <- delta:
		default:
			// A subscriber that cannot keep up is dropped rather than allowed
			// to stall the generation for everyone else.
			delete(l.subscribers, subscriber)
			close(subscriber)
		}
	}
}

func (l *LiveStream) close() {
	l.mu.Lock()
	defer l.mu.Unlock()

	if l.closed {
		return
	}
	l.closed = true
	for subscriber := range l.subscribers {
		close(subscriber)
	}
	l.subscribers = nil
	close(l.done)
}

// Done is closed once the generation has been written away. Callers that must
// not observe the message before then wait on it.
func (l *LiveStream) Done() <-chan struct{} { return l.done }

// Stop cancels the generation. This is the explicit stop button, and it is the
// only thing that ends a generation early.
func (l *LiveStream) Stop() { l.cancel() }

type liveRegistry struct {
	mu      sync.RWMutex
	streams map[uuid.UUID]*LiveStream
}

func newLiveRegistry() *liveRegistry {
	return &liveRegistry{streams: make(map[uuid.UUID]*LiveStream)}
}

func (r *liveRegistry) add(id uuid.UUID, stream *LiveStream) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.streams[id] = stream
}

func (r *liveRegistry) get(id uuid.UUID) (*LiveStream, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	stream, ok := r.streams[id]
	return stream, ok
}

func (r *liveRegistry) remove(id uuid.UUID) {
	r.mu.Lock()
	defer r.mu.Unlock()
	delete(r.streams, id)
}
