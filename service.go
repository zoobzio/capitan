package capitan

import (
	"context"
	"sync"
)

var (
	defaultCapitan *Capitan
	defaultOnce    sync.Once
	defaultOptions []Option
	defaultOptMu   sync.Mutex
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

// Observe registers a callback for all signals on the default instance (dynamic).
// If signals are provided, only those signals will be observed (whitelist).
// If no signals are provided, all signals will be observed.
// The observer will receive events from both existing and future signals.
// Returns an Observer that can be closed to unregister.
func Observe(callback EventCallback, signals ...Signal) *Observer {
	return defaultInstance().Observe(callback, signals...)
}

// Observe registers a callback for all signals (dynamic).
// If signals are provided, only those signals will be observed (whitelist).
// If no signals are provided, all signals will be observed.
// The observer will receive events from both existing and future signals.
// Returns an Observer that can be closed to unregister all listeners.
func (c *Capitan) Observe(callback EventCallback, signals ...Signal) *Observer {
	c.mu.Lock()
	defer c.mu.Unlock()

	o := &Observer{
		listeners: make([]*Listener, 0, len(c.registry)),
		callback:  callback,
		capitan:   c,
		active:    true,
		signals:   nil, // nil = observe all
	}

	// Build whitelist if signals provided
	if len(signals) > 0 {
		o.signals = make(map[Signal]struct{}, len(signals))
		for _, sig := range signals {
			o.signals[sig] = struct{}{}
		}
	}

	// Hook existing signals (filtered by whitelist if present)
	for signal := range c.registry {
		// Skip if whitelist exists and signal not in it
		if o.signals != nil {
			if _, ok := o.signals[signal]; !ok {
				continue
			}
		}

		listener := &Listener{
			signal:   signal,
			callback: callback,
			capitan:  c,
		}
		c.registry[signal] = append(c.registry[signal], listener)
		o.listeners = append(o.listeners, listener)
	}

	// Add to observers list for future signals
	c.observers = append(c.observers, o)

	return o
}

// Emit dispatches an event on the default instance.
func Emit(ctx context.Context, signal Signal, fields ...Field) {
	defaultInstance().Emit(ctx, signal, fields...)
}

// attachObservers attaches all active observers to a signal.
// Must be called while holding c.mu write lock.
func (c *Capitan) attachObservers(signal Signal) {
	for _, obs := range c.observers {
		obs.mu.Lock()
		if obs.active {
			// Skip if observer has whitelist and signal not in it
			if obs.signals != nil {
				if _, ok := obs.signals[signal]; !ok {
					obs.mu.Unlock()
					continue
				}
			}

			obsListener := &Listener{
				signal:   signal,
				callback: obs.callback,
				capitan:  c,
			}
			c.registry[signal] = append(c.registry[signal], obsListener)
			obs.listeners = append(obs.listeners, obsListener)
		}
		obs.mu.Unlock()
	}
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
