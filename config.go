package capitan

import "sync"

var (
	defaultOptions []Option
	defaultOptMu   sync.Mutex
)

// Option configures a Capitan instance.
type Option func(*Capitan)

// PanicHandler is called when a listener panics during event processing.
// Receives the signal being processed and the recovered panic value.
type PanicHandler func(signal Signal, recovered any)

// Configure sets options for the default Capitan instance.
// Must be called before any module-level functions (Hook, Emit, Observe, Shutdown).
// Subsequent calls have no effect once the default instance is created.
func Configure(opts ...Option) {
	defaultOptMu.Lock()
	defaultOptions = opts
	defaultOptMu.Unlock()
}

// WithBufferSize sets the event queue buffer size for each signal's worker.
// Default is 16. Larger buffers reduce backpressure but increase memory usage.
func WithBufferSize(size int) Option {
	return func(c *Capitan) {
		if size > 0 {
			c.bufferSize = size
		}
	}
}

// WithPanicHandler sets a callback to be invoked when a listener panics.
// The handler receives the signal and the recovered panic value.
// By default, panics are recovered silently to prevent system crashes.
func WithPanicHandler(handler PanicHandler) Option {
	return func(c *Capitan) {
		c.panicHandler = handler
	}
}

// WithSyncMode enables synchronous event processing for testing.
// When enabled, Emit() calls listeners directly instead of queueing to workers.
// This eliminates timing dependencies and makes tests deterministic.
// Should only be used in tests, not production code.
func WithSyncMode() Option {
	return func(c *Capitan) {
		c.syncMode = true
	}
}
