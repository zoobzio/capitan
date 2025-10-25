package capitan

import "sync"

// EventCallback is a function that handles an Event.
// Handlers are responsible for their own error handling and logging.
// Events must not be modified by listeners.
type EventCallback func(*Event)

// Listener represents an active subscription to a signal.
// Call Close() to unregister the listener and prevent further callbacks.
type Listener struct {
	signal   Signal
	callback EventCallback
	capitan  *Capitan
}

// Close removes this listener from the registry, preventing future callbacks.
func (l *Listener) Close() {
	l.capitan.unregister(l)
}

// Observer represents a subscription to all signals (dynamic).
// Call Close() to unregister all listeners.
type Observer struct {
	listeners []*Listener
	callback  EventCallback
	capitan   *Capitan
	active    bool
	signals   map[Signal]struct{} // nil = all signals, non-nil = whitelist
	mu        sync.Mutex
}

// Close removes all individual listeners from the registry.
func (o *Observer) Close() {
	// Lock observer first to mark inactive
	o.mu.Lock()
	if !o.active {
		o.mu.Unlock()
		return // Already closed
	}
	o.active = false
	listeners := o.listeners
	o.listeners = nil
	o.mu.Unlock()

	// Remove from capitan's observer list
	o.capitan.mu.Lock()
	for i, obs := range o.capitan.observers {
		if obs == o {
			// Swap with last element and truncate
			lastIdx := len(o.capitan.observers) - 1
			o.capitan.observers[i] = o.capitan.observers[lastIdx]
			o.capitan.observers = o.capitan.observers[:lastIdx]
			break
		}
	}
	o.capitan.mu.Unlock()

	// Close all listeners
	for _, l := range listeners {
		l.Close()
	}
}
