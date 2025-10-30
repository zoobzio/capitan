package capitan

import (
	"context"
	"sync"
	"time"
)

var eventPool = sync.Pool{
	New: func() any {
		return &Event{
			fields:   make(map[string]Field),
			severity: SeverityInfo,
		}
	},
}

// Event represents a signal emission with typed fields.
// Events are immutable after creation - all fields are read-only.
type Event struct {
	// signal identifies the event type for routing to listeners.
	signal Signal

	// timestamp records when the event was created.
	timestamp time.Time

	// ctx is the context passed at emission time, used for cancellation and request-scoped values.
	ctx context.Context

	// fields contains the event's data as key-value pairs, keyed by Key.Name().
	fields map[string]Field

	// severity indicates the logging severity level of this event.
	severity Severity
}

// Signal returns the event's signal identifier.
func (e *Event) Signal() Signal {
	return e.signal
}

// Timestamp returns when the event was created.
func (e *Event) Timestamp() time.Time {
	return e.timestamp
}

// Context returns the context passed at emission time.
// Used for cancellation checks, timeouts, and request-scoped values.
func (e *Event) Context() context.Context {
	return e.ctx
}

// Severity returns the event's severity level.
func (e *Event) Severity() Severity {
	return e.severity
}

// newEvent creates an Event with the given context, signal, severity and fields.
// Events are pooled internally to reduce allocations.
func newEvent(ctx context.Context, signal Signal, severity Severity, fields ...Field) *Event {
	e := eventPool.Get().(*Event) //nolint:errcheck // Pool always returns *Event
	e.signal = signal
	e.timestamp = time.Now()
	e.ctx = ctx
	e.severity = severity

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
