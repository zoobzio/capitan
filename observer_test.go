package capitan

import (
	"sync"
	"testing"
	"time"
)

// TestObserverDynamic verifies that observers receive events from signals
// created after the observer was registered.
func TestObserverDynamic(t *testing.T) {
	c := New()
	defer c.Shutdown()

	key := NewStringKey("msg")
	var received []Signal
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create observer BEFORE any signals exist
	observer := c.Observe(func(e *Event) {
		mu.Lock()
		received = append(received, e.Signal)
		mu.Unlock()
		wg.Done()
	})
	defer observer.Close()

	// Now create signals and emit - observer should see them
	sig1 := Signal("test.sig1")
	sig2 := Signal("test.sig2")

	wg.Add(2)
	c.Hook(sig1, func(_ *Event) {}) // Create signal 1
	c.Hook(sig2, func(_ *Event) {}) // Create signal 2

	c.Emit(sig1, key.Field("first"))
	c.Emit(sig2, key.Field("second"))

	wg.Wait()

	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}

	// Verify both signals received
	found1, found2 := false, false
	for _, sig := range received {
		if sig == sig1 {
			found1 = true
		}
		if sig == sig2 {
			found2 = true
		}
	}

	if !found1 || !found2 {
		t.Errorf("expected both signals, got %v", received)
	}
}

// TestObserverDynamicWithEmit verifies observers work when signals are
// created via Emit() rather than Hook().
func TestObserverDynamicWithEmit(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.emit")
	key := NewStringKey("value")

	var received int
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(1)

	// Create observer before signal exists
	observer := c.Observe(func(_ *Event) {
		mu.Lock()
		received++
		mu.Unlock()
		wg.Done()
	})
	defer observer.Close()

	// Emit creates the signal lazily
	c.Emit(sig, key.Field("test"))

	wg.Wait()

	// Observer should have received the event
	mu.Lock()
	count := received
	mu.Unlock()

	if count != 1 {
		t.Errorf("expected 1 event, got %d", count)
	}
}

// TestConcurrentObserverAndHook tests race conditions between
// creating observers and hooking new signals.
func TestConcurrentObserverAndHook(_ *testing.T) {
	c := New()
	defer c.Shutdown()

	const numSignals = 50
	const duration = 100 * time.Millisecond

	var wg sync.WaitGroup

	// Goroutine 1: Create observers continuously
	wg.Add(1)
	go func() {
		defer wg.Done()
		deadline := time.Now().Add(duration)
		for i := 0; time.Now().Before(deadline); i++ {
			obs := c.Observe(func(_ *Event) {})
			time.Sleep(time.Microsecond)
			obs.Close()
		}
	}()

	// Goroutine 2: Hook new signals continuously
	wg.Add(1)
	go func() {
		defer wg.Done()
		deadline := time.Now().Add(duration)
		for i := 0; time.Now().Before(deadline); i++ {
			sig := Signal("test.concurrent." + string(rune(i)))
			listener := c.Hook(sig, func(_ *Event) {})
			time.Sleep(time.Microsecond)
			listener.Close()
		}
	}()

	// Goroutine 3: Emit to random signals
	wg.Add(1)
	go func() {
		defer wg.Done()
		key := NewIntKey("value")
		deadline := time.Now().Add(duration)
		for i := 0; time.Now().Before(deadline); i++ {
			sig := Signal("test.concurrent." + string(rune(i%numSignals)))
			c.Emit(sig, key.Field(i))
		}
	}()

	wg.Wait()
}

// TestObserverCloseIdempotentWithDynamic verifies Close() can be called
// multiple times safely with the dynamic observer implementation.
func TestObserverCloseIdempotentWithDynamic(_ *testing.T) {
	c := New()
	defer c.Shutdown()

	observer := c.Observe(func(_ *Event) {})

	// Multiple closes should not panic or race
	var wg sync.WaitGroup
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			observer.Close()
		}()
	}
	wg.Wait()
}

// TestObserverReceivesFutureSignals ensures observer gets events from
// signals created after observer creation.
func TestObserverReceivesFutureSignals(t *testing.T) {
	c := New()
	defer c.Shutdown()

	key := NewStringKey("msg")
	var count int
	var mu sync.Mutex
	var wg sync.WaitGroup

	// Create observer with no signals in registry
	observer := c.Observe(func(_ *Event) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	})
	defer observer.Close()

	// Create 5 new signals after observer exists
	for i := 0; i < 5; i++ {
		wg.Add(1)
		sig := Signal("test.future." + string(rune('a'+i)))
		c.Hook(sig, func(_ *Event) {})
		c.Emit(sig, key.Field("test"))
	}

	wg.Wait()

	mu.Lock()
	finalCount := count
	mu.Unlock()

	if finalCount != 5 {
		t.Errorf("expected 5 events, got %d", finalCount)
	}
}

// TestObserverDoesNotReceiveAfterClose verifies that observers stop
// receiving events after Close() is called.
func TestObserverDoesNotReceiveAfterClose(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.close")
	key := NewStringKey("value")

	var count int
	var mu sync.Mutex
	var wg sync.WaitGroup

	wg.Add(1)
	observer := c.Observe(func(_ *Event) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	})

	// Emit first event - should be received
	c.Hook(sig, func(_ *Event) {})
	c.Emit(sig, key.Field("first"))
	wg.Wait()

	// Close observer
	observer.Close()

	// Emit second event - should NOT be received
	sig2 := Signal("test.close2")
	c.Hook(sig2, func(_ *Event) {})
	c.Emit(sig2, key.Field("second"))

	// Give time for any processing
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	finalCount := count
	mu.Unlock()

	if finalCount != 1 {
		t.Errorf("expected 1 event (before close), got %d", finalCount)
	}
}
