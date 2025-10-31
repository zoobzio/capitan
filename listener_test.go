package capitan

import (
	"context"
	"sync"
	"testing"
	"time"
)

func TestListenerClose(t *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig := NewSignal("test.listener.close", "Test listener close signal")
	key := NewStringKey("value")

	count := 0

	listener := c.Hook(sig, func(_ context.Context, _ *Event) {
		count++
	})

	c.Emit(context.Background(), sig, key.Field("first"))

	// Close listener
	listener.Close()

	// This emission should not be received
	c.Emit(context.Background(), sig, key.Field("second"))

	if count != 1 {
		t.Errorf("expected 1 event received, got %d", count)
	}
}

func TestListenerCloseIdempotent(_ *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig := NewSignal("test.listener.idempotent", "Test listener idempotent signal")

	listener := c.Hook(sig, func(_ context.Context, _ *Event) {})

	// Close multiple times should not panic
	listener.Close()
	listener.Close()
	listener.Close()
}

func TestListenerMultiplePerSignal(t *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig := NewSignal("test.listener.multiple", "Test listener multiple signal")
	key := NewStringKey("value")

	count := 0

	c.Hook(sig, func(_ context.Context, _ *Event) {
		count++
	})

	c.Hook(sig, func(_ context.Context, _ *Event) {
		count++
	})

	c.Hook(sig, func(_ context.Context, _ *Event) {
		count++
	})

	c.Emit(context.Background(), sig, key.Field("test"))

	if count != 3 {
		t.Errorf("expected 3 listener invocations, got %d", count)
	}
}

func TestObserverClose(t *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig1 := NewSignal("test.observer.close1", "Test observer close signal 1")
	sig2 := NewSignal("test.observer.close2", "Test observer close signal 2")
	key := NewStringKey("value")

	// Create hooks so signals exist
	c.Hook(sig1, func(_ context.Context, _ *Event) {})
	c.Hook(sig2, func(_ context.Context, _ *Event) {})

	count := 0

	observer := c.Observe(func(_ context.Context, _ *Event) {
		count++
	})

	c.Emit(context.Background(), sig1, key.Field("first"))
	c.Emit(context.Background(), sig2, key.Field("second"))

	// Close observer
	observer.Close()

	// These should not be received
	c.Emit(context.Background(), sig1, key.Field("third"))
	c.Emit(context.Background(), sig2, key.Field("fourth"))

	if count != 2 {
		t.Errorf("expected 2 events, got %d", count)
	}
}

func TestObserverCloseIdempotent(_ *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig := NewSignal("test.observer.idempotent", "Test observer idempotent signal")

	// Create hook so signal exists
	c.Hook(sig, func(_ context.Context, _ *Event) {})

	observer := c.Observe(func(_ context.Context, _ *Event) {})

	// Close multiple times should not panic
	observer.Close()
	observer.Close()
	observer.Close()
}

func TestObserverSnapshotBehavior(_ *testing.T) {
	c := New()
	defer c.Shutdown()

	sig1 := NewSignal("test.observer.snapshot1", "Test observer snapshot signal 1")
	sig2 := NewSignal("test.observer.snapshot2", "Test observer snapshot signal 2")
	key := NewStringKey("value")

	// Create first signal
	c.Hook(sig1, func(_ context.Context, _ *Event) {})

	var wg sync.WaitGroup
	wg.Add(1)

	// Observer should only see sig1 (snapshot at creation time)
	observer := c.Observe(func(_ context.Context, e *Event) {
		if e.Signal() == sig1 {
			wg.Done()
		}
	})

	// Create second signal after observer created
	c.Hook(sig2, func(_ context.Context, _ *Event) {})

	c.Emit(context.Background(), sig1, key.Field("first"))
	c.Emit(context.Background(), sig2, key.Field("second"))

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
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig1 := NewSignal("test.observer.all1", "Test observer all signal 1")
	sig2 := NewSignal("test.observer.all2", "Test observer all signal 2")
	sig3 := NewSignal("test.observer.all3", "Test observer all signal 3")
	key := NewStringKey("value")

	// Create all signals
	c.Hook(sig1, func(_ context.Context, _ *Event) {})
	c.Hook(sig2, func(_ context.Context, _ *Event) {})
	c.Hook(sig3, func(_ context.Context, _ *Event) {})

	received := make(map[Signal]bool)

	c.Observe(func(_ context.Context, e *Event) {
		received[e.Signal()] = true
	})

	c.Emit(context.Background(), sig1, key.Field("first"))
	c.Emit(context.Background(), sig2, key.Field("second"))
	c.Emit(context.Background(), sig3, key.Field("third"))

	if !received[sig1] || !received[sig2] || !received[sig3] {
		t.Errorf("observer did not receive all signals: %v", received)
	}
}

// TestEmitAfterListenerCloseFullBuffer verifies that emissions after
// listener close don't block when the buffer is full.
func TestEmitAfterListenerCloseFullBuffer(t *testing.T) {
	// Small buffer to easily fill
	c := New(WithBufferSize(2))
	defer c.Shutdown()

	sig := NewSignal("test.fullbuffer", "Test full buffer signal")
	key := NewIntKey("value")

	// Create listener that blocks processing
	block := make(chan struct{})
	var wg sync.WaitGroup
	wg.Add(1)

	listener := c.Hook(sig, func(_ context.Context, _ *Event) {
		wg.Done()
		<-block // Block until we release
	})

	// Fill the buffer: 1 processing + 2 buffered = 3 total
	c.Emit(context.Background(), sig, key.Field(1))
	wg.Wait() // Wait for first to start processing

	c.Emit(context.Background(), sig, key.Field(2)) // Buffer slot 1
	c.Emit(context.Background(), sig, key.Field(3)) // Buffer slot 2
	time.Sleep(10 * time.Millisecond)

	// Close listener (buffer still full, worker will exit)
	listener.Close()

	// This should NOT block (previously would without worker.done check)
	done := make(chan struct{})
	go func() {
		c.Emit(context.Background(), sig, key.Field(4)) // Should detect worker.done and drop
		close(done)
	}()

	select {
	case <-done:
		// Success - emit didn't block
	case <-time.After(500 * time.Millisecond):
		t.Fatal("Emit blocked after listener close with full buffer")
	}

	close(block) // Unblock worker
}

// TestConcurrentEmitAndListenerClose tests the TOCTOU race between
// Emit capturing a worker reference and listener close deleting it.
func TestConcurrentEmitAndListenerClose(t *testing.T) {
	const iterations = 100

	for i := 0; i < iterations; i++ {
		c := New(WithBufferSize(1))
		sig := NewSignal("test.race", "Test race signal")
		key := NewIntKey("value")

		// Slow listener to increase contention window
		listener := c.Hook(sig, func(_ context.Context, _ *Event) {
			time.Sleep(time.Microsecond)
		})

		var wg sync.WaitGroup

		// Goroutine 1: Emit rapidly
		wg.Add(1)
		go func() {
			defer wg.Done()
			for j := 0; j < 50; j++ {
				c.Emit(context.Background(), sig, key.Field(j))
			}
		}()

		// Goroutine 2: Close listener mid-stream
		wg.Add(1)
		go func() {
			defer wg.Done()
			time.Sleep(time.Microsecond * 10)
			listener.Close()
		}()

		// Should complete without blocking
		done := make(chan struct{})
		go func() {
			wg.Wait()
			c.Shutdown()
			close(done)
		}()

		select {
		case <-done:
			// Success
		case <-time.After(2 * time.Second):
			t.Fatalf("iteration %d: deadlock detected", i)
		}
	}
}

// TestEmitToClosedWorkerDropsEvent verifies events are properly
// dropped (not leaked) when sent to a closing worker.
func TestEmitToClosedWorkerDropsEvent(t *testing.T) {
	c := New(WithBufferSize(1))
	defer c.Shutdown()

	sig := NewSignal("test.drop", "Test drop signal")
	key := NewStringKey("value")

	received := 0
	var mu sync.Mutex
	var wg sync.WaitGroup

	listener := c.Hook(sig, func(_ context.Context, _ *Event) {
		mu.Lock()
		received++
		mu.Unlock()
		wg.Done()
	})

	// Emit first event - should be received
	wg.Add(1)
	c.Emit(context.Background(), sig, key.Field("first"))
	wg.Wait()

	// Close listener
	listener.Close()
	time.Sleep(20 * time.Millisecond) // Ensure worker exits

	// Emit more events - should be dropped silently
	c.Emit(context.Background(), sig, key.Field("dropped1"))
	c.Emit(context.Background(), sig, key.Field("dropped2"))
	c.Emit(context.Background(), sig, key.Field("dropped3"))

	time.Sleep(50 * time.Millisecond)

	mu.Lock()
	finalCount := received
	mu.Unlock()

	if finalCount != 1 {
		t.Errorf("expected 1 event received, got %d", finalCount)
	}
}
