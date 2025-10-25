package capitan

const (
	// defaultBufferSize is the default channel buffer size for event queues.
	// Each signal gets its own buffered channel with this capacity.
	defaultBufferSize = 16
)

// Emit dispatches an event with the given signal and fields.
// Queues the event for asynchronous processing by the signal's worker goroutine.
// Creates a worker goroutine lazily on first emission to this signal.
func (c *Capitan) Emit(signal Signal, fields ...Field) {
	// Fast path: check if worker already exists (read lock)
	c.mu.RLock()
	worker, exists := c.workers[signal]
	c.mu.RUnlock()

	if !exists {
		// Slow path: create worker (write lock)
		c.mu.Lock()

		// Double-check: another goroutine may have created it
		worker, exists = c.workers[signal]
		if !exists {
			worker = make(chan *Event, defaultBufferSize)
			c.workers[signal] = worker
			c.wg.Add(1)
			go c.processEvents(signal, worker)
		}

		// Check if this is a new signal in registry
		_, registryExists := c.registry[signal]
		if !registryExists {
			// Initialize registry entry for this signal
			c.registry[signal] = nil

			// Attach to all active observers
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

		c.mu.Unlock()
	}

	// Create event from pool
	event := newEvent(signal, fields...)

	// Send to worker (blocks if queue full - backpressure)
	worker <- event
}

// processEvents is the worker goroutine for a specific signal.
// Processes events from the queue and invokes all registered listeners.
func (c *Capitan) processEvents(signal Signal, events chan *Event) {
	defer c.wg.Done()

	for {
		select {
		case event := <-events:
			// Copy listener slice while holding lock to prevent data race
			c.mu.RLock()
			listeners := make([]*Listener, len(c.registry[signal]))
			copy(listeners, c.registry[signal])
			c.mu.RUnlock()

			// Invoke all listeners with panic recovery
			for _, listener := range listeners {
				func() {
					defer func() {
						_ = recover() //nolint:errcheck // Intentionally discard panic value
					}()
					listener.callback(event)
				}()
			}

			// Return event to pool
			eventPool.Put(event)

		case <-c.shutdown:
			// Process remaining events before shutting down
			for {
				select {
				case event := <-events:
					// Copy listener slice while holding lock to prevent data race
					c.mu.RLock()
					listeners := make([]*Listener, len(c.registry[signal]))
					copy(listeners, c.registry[signal])
					c.mu.RUnlock()

					// Invoke all listeners with panic recovery
					for _, listener := range listeners {
						func() {
							defer func() {
								_ = recover() //nolint:errcheck // Intentionally discard panic value
							}()
							listener.callback(event)
						}()
					}

					// Return event to pool
					eventPool.Put(event)
				default:
					return
				}
			}
		}
	}
}

// Shutdown gracefully stops all worker goroutines, draining pending events.
// Safe to call multiple times; subsequent calls are no-ops.
func (c *Capitan) Shutdown() {
	c.shutdownOnce.Do(func() {
		close(c.shutdown)
	})
	c.wg.Wait()
}
