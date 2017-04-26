package slack

import "sync"

type Handler interface{
	ServeEvent(*Message, *Client)
}

// The HandlerFunc type is an adapter to allow the use of
// ordinary functions as Event handlers. If f is a function
// with the appropriate signature, HandlerFunc(f) is a
// Handler that calls f.
type HandlerFunc func(*Message, *Client)

// ServeEvent calls f(msg, slack).
func (f HandlerFunc) ServeEvent(msg *Message, slack *Client) {
	f(msg, slack)
}

// EventMux is an Event multiplexer.
// It matches the event type of each incoming event against a list of registered
// events and calls the handler for the event
type EventMux struct {
	mu sync.RWMutex
	m  map[string]muxEntry
}

type muxEntry struct {
	explicit bool
	handler  Handler
	event    string
}

// NewEventMux allocates and returns a new EventsMux.
func NewEventMux() *EventMux { return new(EventMux) }

// Handle registers the handler for the given pattern.
// If a handler already exists for pattern, Handle panics.
func (mux *EventMux) Handle(event string, handler Handler) {
	mux.mu.Lock()
	defer mux.mu.Unlock()

	if event == "" {
		panic("events: invalid event " + event)
	}
	if handler == nil {
		panic("events: nil handler")
	}
	if mux.m[event].explicit {
		panic("events: multiple registrations for " + event)
	}

	if mux.m == nil {
		mux.m = make(map[string]muxEntry)
	}
	mux.m[event] = muxEntry{explicit: true, handler: handler, event: event}
}

// Find a handler in a handler map given an event string.
func (mux *EventMux) match(event string) (h Handler, pattern string) {
	// Check for exact match first.
	v, ok := mux.m[event]
	if ok {
		return v.handler, v.event
	}
		panic("events: missing handler for " + event)
}

func eventMatch(event string, inputEvent string) bool {
	if len(event) == 0 {
		// should not happen
		return false
	}

	return inputEvent == event
}
