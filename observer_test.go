package capitan

import (
	"context"
	"sync"
	"testing"
	"time"
)

// TestObserverDynamic verifies that observers receive events from signals
// created after the observer was registered.
func TestObserverDynamic(t *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	key := NewStringKey("msg")
	var received []Signal

	// Create observer BEFORE any signals exist
	observer := c.Observe(func(_ context.Context, e *Event) {
		received = append(received, e.Signal())
	})
	defer observer.Close()

	// Now create signals and emit - observer should see them
	sig1 := Signal("test.sig1")
	sig2 := Signal("test.sig2")

	c.Hook(sig1, func(_ context.Context, _ *Event) {}) // Create signal 1
	c.Hook(sig2, func(_ context.Context, _ *Event) {}) // Create signal 2

	c.Emit(context.Background(), sig1, key.Field("first"))
	c.Emit(context.Background(), sig2, key.Field("second"))

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
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig := Signal("test.emit")
	key := NewStringKey("value")

	var received int

	// Create observer before signal exists
	observer := c.Observe(func(_ context.Context, _ *Event) {
		received++
	})
	defer observer.Close()

	// Emit creates the signal lazily
	c.Emit(context.Background(), sig, key.Field("test"))

	if received != 1 {
		t.Errorf("expected 1 event, got %d", received)
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
			obs := c.Observe(func(_ context.Context, _ *Event) {})
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
			listener := c.Hook(sig, func(_ context.Context, _ *Event) {})
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
			c.Emit(context.Background(), sig, key.Field(i))
		}
	}()

	wg.Wait()
}

// TestObserverCloseIdempotentWithDynamic verifies Close() can be called
// multiple times safely with the dynamic observer implementation.
func TestObserverCloseIdempotentWithDynamic(_ *testing.T) {
	c := New()
	defer c.Shutdown()

	observer := c.Observe(func(_ context.Context, _ *Event) {})

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
	c := New(WithSyncMode())
	defer c.Shutdown()

	key := NewStringKey("msg")
	var count int

	// Create observer with no signals in registry
	observer := c.Observe(func(_ context.Context, _ *Event) {
		count++
	})
	defer observer.Close()

	// Create 5 new signals after observer exists
	for i := 0; i < 5; i++ {
		sig := Signal("test.future." + string(rune('a'+i)))
		c.Hook(sig, func(_ context.Context, _ *Event) {})
		c.Emit(context.Background(), sig, key.Field("test"))
	}

	if count != 5 {
		t.Errorf("expected 5 events, got %d", count)
	}
}

// TestObserverDoesNotReceiveAfterClose verifies that observers stop
// receiving events after Close() is called.
func TestObserverDoesNotReceiveAfterClose(t *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig := Signal("test.close")
	key := NewStringKey("value")

	var count int

	observer := c.Observe(func(_ context.Context, _ *Event) {
		count++
	})

	// Emit first event - should be received
	c.Hook(sig, func(_ context.Context, _ *Event) {})
	c.Emit(context.Background(), sig, key.Field("first"))

	// Close observer
	observer.Close()

	// Emit second event - should NOT be received
	sig2 := Signal("test.close2")
	c.Hook(sig2, func(_ context.Context, _ *Event) {})
	c.Emit(context.Background(), sig2, key.Field("second"))

	if count != 1 {
		t.Errorf("expected 1 event (before close), got %d", count)
	}
}

func TestObserverWithWhitelist(t *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig1 := Signal("test.whitelist.one")
	sig2 := Signal("test.whitelist.two")
	sig3 := Signal("test.whitelist.three")
	key := NewStringKey("value")

	// Hook all three signals
	c.Hook(sig1, func(_ context.Context, _ *Event) {})
	c.Hook(sig2, func(_ context.Context, _ *Event) {})
	c.Hook(sig3, func(_ context.Context, _ *Event) {})

	var received []Signal

	// Observer only watching sig1 and sig2 (whitelist)
	observer := c.Observe(func(_ context.Context, e *Event) {
		received = append(received, e.Signal())
	}, sig1, sig2)
	defer observer.Close()

	// Emit to all three
	c.Emit(context.Background(), sig1, key.Field("first"))
	c.Emit(context.Background(), sig2, key.Field("second"))
	c.Emit(context.Background(), sig3, key.Field("third"))

	// Should only receive sig1 and sig2
	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}

	found1, found2, found3 := false, false, false
	for _, sig := range received {
		if sig == sig1 {
			found1 = true
		}
		if sig == sig2 {
			found2 = true
		}
		if sig == sig3 {
			found3 = true
		}
	}

	if !found1 {
		t.Error("expected sig1 to be received")
	}
	if !found2 {
		t.Error("expected sig2 to be received")
	}
	if found3 {
		t.Error("sig3 should not be received (not in whitelist)")
	}
}

func TestObserverWhitelistFutureSignals(t *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig1 := Signal("test.future.one")
	sig2 := Signal("test.future.two")
	sig3 := Signal("test.future.three")
	key := NewStringKey("value")

	var received []Signal

	// Create observer BEFORE signals exist, with whitelist
	observer := c.Observe(func(_ context.Context, e *Event) {
		received = append(received, e.Signal())
	}, sig1, sig3)
	defer observer.Close()

	// Now create signals and emit
	c.Hook(sig1, func(_ context.Context, _ *Event) {})
	c.Hook(sig2, func(_ context.Context, _ *Event) {})
	c.Hook(sig3, func(_ context.Context, _ *Event) {})

	c.Emit(context.Background(), sig1, key.Field("first"))
	c.Emit(context.Background(), sig2, key.Field("second"))
	c.Emit(context.Background(), sig3, key.Field("third"))

	// Should only receive sig1 and sig3
	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}

	found1, found2, found3 := false, false, false
	for _, sig := range received {
		if sig == sig1 {
			found1 = true
		}
		if sig == sig2 {
			found2 = true
		}
		if sig == sig3 {
			found3 = true
		}
	}

	if !found1 {
		t.Error("expected sig1 to be received")
	}
	if found2 {
		t.Error("sig2 should not be received (not in whitelist)")
	}
	if !found3 {
		t.Error("expected sig3 to be received")
	}
}

func TestObserverNoWhitelistReceivesAll(t *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig1 := Signal("test.all.one")
	sig2 := Signal("test.all.two")
	key := NewStringKey("value")

	c.Hook(sig1, func(_ context.Context, _ *Event) {})
	c.Hook(sig2, func(_ context.Context, _ *Event) {})

	var received []Signal

	// Observer with NO whitelist (should receive all)
	observer := c.Observe(func(_ context.Context, e *Event) {
		received = append(received, e.Signal())
	})
	defer observer.Close()

	c.Emit(context.Background(), sig1, key.Field("first"))
	c.Emit(context.Background(), sig2, key.Field("second"))

	// Should receive both
	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}
}

func TestObserverContextPropagation(t *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	sig := Signal("test.observer.ctx")
	key := NewStringKey("value")

	type ctxKey string
	const traceKey ctxKey = "trace_id"
	expectedID := "trace-123"

	var received string

	c.Observe(func(ctx context.Context, e *Event) {
		if e.Signal() == sig {
			received = ctx.Value(traceKey).(string)
		}
	})

	c.Hook(sig, func(_ context.Context, _ *Event) {})

	ctx := context.WithValue(context.Background(), traceKey, expectedID)
	c.Emit(ctx, sig, key.Field("test"))

	if received != expectedID {
		t.Errorf("expected %q, got %q", expectedID, received)
	}
}
