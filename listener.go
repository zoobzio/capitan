package capitan

import "context"

// EventCallback is a function that handles an Event.
// The context is inherited from the Emit call and can be used for cancellation,
// timeouts, and accessing request-scoped values.
// Handlers are responsible for their own error handling and logging.
// Events must not be modified by listeners.
type EventCallback func(context.Context, *Event)

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
