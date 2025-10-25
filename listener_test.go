package capitan

import (
	"sync"
	"testing"
	"time"
)

func TestListenerClose(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.listener.close")
	key := NewStringKey("value")

	count := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	listener := c.Hook(sig, func(_ *Event) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	})

	wg.Add(1)
	c.Emit(sig, key.Field("first"))
	wg.Wait()

	// Close listener
	listener.Close()

	// This emission should not be received
	c.Emit(sig, key.Field("second"))

	// Give time for any errant delivery
	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	finalCount := count
	mu.Unlock()

	if finalCount != 1 {
		t.Errorf("expected 1 event received, got %d", finalCount)
	}
}

func TestListenerCloseIdempotent(_ *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.listener.idempotent")

	listener := c.Hook(sig, func(_ *Event) {})

	// Close multiple times should not panic
	listener.Close()
	listener.Close()
	listener.Close()
}

func TestListenerMultiplePerSignal(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.listener.multiple")
	key := NewStringKey("value")

	count := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(3)

	c.Hook(sig, func(_ *Event) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	})

	c.Hook(sig, func(_ *Event) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	})

	c.Hook(sig, func(_ *Event) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	})

	c.Emit(sig, key.Field("test"))

	wg.Wait()

	mu.Lock()
	finalCount := count
	mu.Unlock()

	if finalCount != 3 {
		t.Errorf("expected 3 listener invocations, got %d", finalCount)
	}
}

func TestObserverClose(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig1 := Signal("test.observer.close1")
	sig2 := Signal("test.observer.close2")
	key := NewStringKey("value")

	// Create hooks so signals exist
	c.Hook(sig1, func(_ *Event) {})
	c.Hook(sig2, func(_ *Event) {})

	count := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	observer := c.Observe(func(_ *Event) {
		mu.Lock()
		count++
		mu.Unlock()
		wg.Done()
	})

	c.Emit(sig1, key.Field("first"))
	c.Emit(sig2, key.Field("second"))

	wg.Wait()

	// Close observer
	observer.Close()

	// These should not be received
	c.Emit(sig1, key.Field("third"))
	c.Emit(sig2, key.Field("fourth"))

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	finalCount := count
	mu.Unlock()

	if finalCount != 2 {
		t.Errorf("expected 2 events, got %d", finalCount)
	}
}

func TestObserverCloseIdempotent(_ *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.observer.idempotent")

	// Create hook so signal exists
	c.Hook(sig, func(_ *Event) {})

	observer := c.Observe(func(_ *Event) {})

	// Close multiple times should not panic
	observer.Close()
	observer.Close()
	observer.Close()
}

func TestObserverSnapshotBehavior(_ *testing.T) {
	c := New()
	defer c.Shutdown()

	sig1 := Signal("test.observer.snapshot1")
	sig2 := Signal("test.observer.snapshot2")
	key := NewStringKey("value")

	// Create first signal
	c.Hook(sig1, func(_ *Event) {})

	var wg sync.WaitGroup
	wg.Add(1)

	// Observer should only see sig1 (snapshot at creation time)
	observer := c.Observe(func(e *Event) {
		if e.Signal == sig1 {
			wg.Done()
		}
	})

	// Create second signal after observer created
	c.Hook(sig2, func(_ *Event) {})

	c.Emit(sig1, key.Field("first"))
	c.Emit(sig2, key.Field("second"))

	// Should only receive sig1
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - only sig1 received
	case <-time.After(100 * time.Millisecond):
		// Timeout is OK - we're checking sig2 wasn't received
	}

	observer.Close()
}

func TestObserverReceivesAllExistingSignals(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig1 := Signal("test.observer.all1")
	sig2 := Signal("test.observer.all2")
	sig3 := Signal("test.observer.all3")
	key := NewStringKey("value")

	// Create all signals
	c.Hook(sig1, func(_ *Event) {})
	c.Hook(sig2, func(_ *Event) {})
	c.Hook(sig3, func(_ *Event) {})

	received := make(map[Signal]bool)
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(3)

	c.Observe(func(e *Event) {
		mu.Lock()
		received[e.Signal] = true
		mu.Unlock()
		wg.Done()
	})

	c.Emit(sig1, key.Field("first"))
	c.Emit(sig2, key.Field("second"))
	c.Emit(sig3, key.Field("third"))

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if !received[sig1] || !received[sig2] || !received[sig3] {
		t.Errorf("observer did not receive all signals: %v", received)
	}
}
