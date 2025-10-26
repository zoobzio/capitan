package capitan

import (
	"context"
	"sync"
)

var (
	defaultCapitan *Capitan
	defaultOnce    sync.Once
)

// Capitan is an event coordination system.
type Capitan struct {
	registry     map[Signal][]*Listener
	workers      map[Signal]*workerState
	observers    []*Observer
	shutdown     chan struct{}
	shutdownOnce sync.Once
	wg           sync.WaitGroup
	mu           sync.RWMutex
	bufferSize   int
	panicHandler PanicHandler
	syncMode     bool
}

// New creates a new Capitan instance with optional configuration.
// If no options are provided, sensible defaults are used (bufferSize=16, no panic handler).
func New(opts ...Option) *Capitan {
	c := &Capitan{
		registry:   make(map[Signal][]*Listener),
		workers:    make(map[Signal]*workerState),
		shutdown:   make(chan struct{}),
		bufferSize: 16, // default buffer size
	}
	for _, opt := range opts {
		opt(c)
	}
	return c
}

// defaultInstance returns the default Capitan instance, creating it if necessary.
func defaultInstance() *Capitan {
	defaultOnce.Do(func() {
		defaultOptMu.Lock()
		opts := defaultOptions
		defaultOptMu.Unlock()
		defaultCapitan = New(opts...)
	})
	return defaultCapitan
}

// Hook registers a callback for the given signal on the default instance.
// Returns a Listener that can be closed to unregister.
func Hook(signal Signal, callback EventCallback) *Listener {
	return defaultInstance().Hook(signal, callback)
}

// Hook registers a callback for the given signal.
// Returns a Listener that can be closed to unregister.
func (c *Capitan) Hook(signal Signal, callback EventCallback) *Listener {
	c.mu.Lock()
	defer c.mu.Unlock()

	listener := &Listener{
		signal:   signal,
		callback: callback,
		capitan:  c,
	}

	// Check if this is a new signal
	_, exists := c.registry[signal]
	c.registry[signal] = append(c.registry[signal], listener)

	// If new signal, attach to all active observers
	if !exists {
		c.attachObservers(signal)
	}

	return listener
}

// Emit dispatches an event on the default instance.
func Emit(ctx context.Context, signal Signal, fields ...Field) {
	defaultInstance().Emit(ctx, signal, fields...)
}

// unregister removes a listener from the registry.
func (c *Capitan) unregister(listener *Listener) {
	c.mu.Lock()
	defer c.mu.Unlock()

	listeners := c.registry[listener.signal]
	for i, l := range listeners {
		if l == listener {
			// Swap with last element and truncate (efficient removal)
			lastIdx := len(listeners) - 1
			listeners[i] = listeners[lastIdx]
			c.registry[listener.signal] = listeners[:lastIdx]
			break
		}
	}

	// Clean up empty signal entries and signal worker to exit
	if len(c.registry[listener.signal]) == 0 {
		delete(c.registry, listener.signal)

		// Signal worker goroutine to drain and exit
		if worker, exists := c.workers[listener.signal]; exists {
			close(worker.done)
			delete(c.workers, listener.signal)
		}
	}
}

// Stats returns runtime metrics for the Capitan instance.
// Provides visibility into active workers, queue depths, and listener counts.
func (c *Capitan) Stats() Stats {
	c.mu.RLock()
	defer c.mu.RUnlock()

	stats := Stats{
		ActiveWorkers:  len(c.workers),
		QueueDepths:    make(map[Signal]int, len(c.workers)),
		ListenerCounts: make(map[Signal]int, len(c.registry)),
	}

	for signal, worker := range c.workers {
		stats.QueueDepths[signal] = len(worker.events)
	}

	for signal, listeners := range c.registry {
		stats.ListenerCounts[signal] = len(listeners)
	}

	return stats
}

// Shutdown gracefully stops all worker goroutines on the default instance.
func Shutdown() {
	defaultInstance().Shutdown()
}
