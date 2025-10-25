package capitan

import (
	"sync"
	"time"
)

var eventPool = sync.Pool{
	New: func() any {
		return &Event{
			fields: make(map[string]Field),
		}
	},
}

// Event represents a signal emission with typed fields.
type Event struct {
	// Signal identifies the event type for routing to listeners.
	Signal Signal

	// Timestamp records when the event was created.
	Timestamp time.Time

	// fields contains the event's data as key-value pairs, keyed by Key.Name().
	fields map[string]Field
}

// newEvent creates an Event with the given signal and fields.
// Events are pooled internally to reduce allocations.
func newEvent(signal Signal, fields ...Field) *Event {
	e := eventPool.Get().(*Event) //nolint:errcheck // Pool always returns *Event
	e.Signal = signal
	e.Timestamp = time.Now()

	// Clear existing fields
	for k := range e.fields {
		delete(e.fields, k)
	}

	// Add new fields, keyed by name
	for _, field := range fields {
		e.fields[field.Key().Name()] = field
	}

	return e
}

// Get retrieves a field by key, returning nil if not found.
func (e Event) Get(key Key) Field {
	return e.fields[key.Name()]
}

// Fields returns all fields as a slice.
// Returns a defensive copy; modifications don't affect the event.
func (e Event) Fields() []Field {
	result := make([]Field, 0, len(e.fields))
	for _, field := range e.fields {
		result = append(result, field)
	}
	return result
}
