// Package capitan provides type-safe event coordination for Go with zero dependencies.
//
// At its core, capitan offers three operations: Emit events with typed fields,
// Hook listeners to specific signals, and Observe all signals. Events are processed
// asynchronously with per-signal worker goroutines for isolation and performance.
//
// Context Support: All events carry a context.Context for cancellation, timeouts,
// and request-scoped values. Canceled contexts prevent event queueing and processing.
//
// Quick example:
//
//	sig := capitan.Signal("order.created")
//	orderID := capitan.NewStringKey("order_id")
//
//	capitan.Hook(sig, func(ctx context.Context, e *capitan.Event) {
//	    id, _ := orderID.From(e)
//	    // Process order...
//	})
//
//	capitan.Emit(context.Background(), sig, orderID.Field("ORDER-123"))
//	capitan.Shutdown() // Drain pending events
//
// See https://github.com/zoobzio/capitan for full documentation.
package capitan

// Signal represents an event type identifier used for routing events to listeners.
type Signal string

// Severity represents the logging severity level of an event.
type Severity string

const (
	SeverityDebug Severity = "DEBUG"
	SeverityInfo  Severity = "INFO"
	SeverityWarn  Severity = "WARN"
	SeverityError Severity = "ERROR"
)

// Key represents a typed semantic identifier for a field.
// Each Key implementation is bound to a specific Variant, ensuring type safety.
type Key interface {
	// Name returns the semantic identifier for this key.
	Name() string

	// Variant returns the type constraint for this key.
	Variant() Variant
}

// Variant is a discriminator for the Field interface implementation type.
type Variant string

const (
	VariantString   Variant = "string"
	VariantInt      Variant = "int"
	VariantInt32    Variant = "int32"
	VariantInt64    Variant = "int64"
	VariantUint     Variant = "uint"
	VariantUint32   Variant = "uint32"
	VariantUint64   Variant = "uint64"
	VariantFloat32  Variant = "float32"
	VariantFloat64  Variant = "float64"
	VariantBool     Variant = "bool"
	VariantTime     Variant = "time.Time"
	VariantDuration Variant = "time.Duration"
	VariantBytes    Variant = "[]byte"
	VariantError    Variant = "error"
)

// Field represents a typed value with semantic meaning in an Event.
// Library authors can implement custom Field types while maintaining type safety.
// Use type assertions to access concrete field types and their typed accessor methods.
type Field interface {
	// Variant returns the discriminator for this field's concrete type.
	Variant() Variant

	// Key returns the semantic identifier for this field.
	Key() Key

	// Value returns the underlying value as any.
	Value() any
}

// GenericField is a generic implementation of Field for typed values.
type GenericField[T any] struct {
	key     Key
	value   T
	variant Variant
}

// Variant returns the discriminator for this field's type.
func (f GenericField[T]) Variant() Variant { return f.variant }

// Key returns the semantic identifier for this field.
func (f GenericField[T]) Key() Key { return f.key }

// Value returns the underlying value as any.
func (f GenericField[T]) Value() any { return f.value }

// Get returns the typed value.
func (f GenericField[T]) Get() T { return f.value }

// workerState manages the lifecycle of a signal's worker goroutine.
type workerState struct {
	events chan *Event   // buffered channel for queuing events
	done   chan struct{} // signals worker to drain and exit
}

// Stats provides runtime metrics for a Capitan instance.
type Stats struct {
	// ActiveWorkers is the number of worker goroutines currently running.
	ActiveWorkers int

	// QueueDepths maps each signal to the number of events queued in its buffer.
	QueueDepths map[Signal]int

	// ListenerCounts maps each signal to the number of registered listeners.
	ListenerCounts map[Signal]int
}
