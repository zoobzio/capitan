package capitan

import (
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
}

// New creates a new Capitan instance.
func New() *Capitan {
	return &Capitan{
		registry: make(map[Signal][]*Listener),
		workers:  make(map[Signal]*workerState),
		shutdown: make(chan struct{}),
	}
}

// defaultInstance returns the default Capitan instance, creating it if necessary.
func defaultInstance() *Capitan {
	defaultOnce.Do(func() {
		defaultCapitan = New()
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
// The observer will receive events from both existing and future signals.
// Returns an Observer that can be closed to unregister.
func Observe(callback EventCallback) *Observer {
	return defaultInstance().Observe(callback)
}

// Observe registers a callback for all signals (dynamic).
// The observer will receive events from both existing and future signals.
// Returns an Observer that can be closed to unregister all listeners.
func (c *Capitan) Observe(callback EventCallback) *Observer {
	c.mu.Lock()
	defer c.mu.Unlock()

	o := &Observer{
		listeners: make([]*Listener, 0, len(c.registry)),
		callback:  callback,
		capitan:   c,
		active:    true,
	}

	// Hook all existing signals
	for signal := range c.registry {
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
func Emit(signal Signal, fields ...Field) {
	defaultInstance().Emit(signal, fields...)
}

// attachObservers attaches all active observers to a signal.
// Must be called while holding c.mu write lock.
func (c *Capitan) attachObservers(signal Signal) {
	for _, obs := range c.observers {
		obs.mu.Lock()
		if obs.active {
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

// Shutdown gracefully stops all worker goroutines on the default instance.
func Shutdown() {
	defaultInstance().Shutdown()
}
