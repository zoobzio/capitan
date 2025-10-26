package capitan

import (
	"context"
	"testing"
	"time"
)

// TestWithBufferSize verifies custom buffer size configuration.
func TestWithBufferSize(t *testing.T) {
	c := New(WithBufferSize(128))

	if c.bufferSize != 128 {
		t.Errorf("expected bufferSize=128, got %d", c.bufferSize)
	}

	c.Shutdown()
}

// TestWithBufferSizeDefault verifies default buffer size.
func TestWithBufferSizeDefault(t *testing.T) {
	c := New()

	if c.bufferSize != 16 {
		t.Errorf("expected default bufferSize=16, got %d", c.bufferSize)
	}

	c.Shutdown()
}

// TestWithBufferSizeInvalid verifies invalid buffer size is rejected.
func TestWithBufferSizeInvalid(t *testing.T) {
	c := New(WithBufferSize(-1))

	if c.bufferSize != 16 {
		t.Errorf("expected bufferSize to remain default (16), got %d", c.bufferSize)
	}

	c.Shutdown()
}

// TestWithPanicHandler verifies panic handler is called on listener panic.
func TestWithPanicHandler(t *testing.T) {
	var panicSignal Signal
	var panicValue any

	handler := func(sig Signal, recovered any) {
		panicSignal = sig
		panicValue = recovered
	}

	c := New(WithPanicHandler(handler), WithSyncMode())
	defer c.Shutdown()

	sig := Signal("test.panic")
	key := NewStringKey("value")

	c.Hook(sig, func(_ context.Context, _ *Event) {
		panic("test panic")
	})

	c.Emit(context.Background(), sig, key.Field("test"))

	if panicSignal != sig {
		t.Errorf("expected panicSignal=%q, got %q", sig, panicSignal)
	}

	if panicValue != "test panic" {
		t.Errorf("expected panicValue=%q, got %v", "test panic", panicValue)
	}
}

// TestWithPanicHandlerNotSet verifies silent recovery when no handler set.
func TestWithPanicHandlerNotSet(t *testing.T) {
	c := New(WithSyncMode()) // No panic handler
	defer c.Shutdown()

	sig := Signal("test.panic.silent")
	key := NewStringKey("value")
	var called bool

	// First listener panics
	c.Hook(sig, func(_ context.Context, _ *Event) {
		panic("silent panic")
	})

	// Second listener should still run
	c.Hook(sig, func(_ context.Context, _ *Event) {
		called = true
	})

	c.Emit(context.Background(), sig, key.Field("test"))

	if !called {
		t.Error("second listener should have been called despite first listener panic")
	}
}

// TestStats verifies Stats() returns correct metrics.
func TestStats(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig1 := Signal("test.stats.1")
	sig2 := Signal("test.stats.2")
	key := NewStringKey("value")

	// Hook listeners
	c.Hook(sig1, func(_ context.Context, _ *Event) {
		time.Sleep(50 * time.Millisecond) // Slow listener
	})
	c.Hook(sig1, func(_ context.Context, _ *Event) {
		time.Sleep(50 * time.Millisecond)
	})
	c.Hook(sig2, func(_ context.Context, _ *Event) {
		time.Sleep(50 * time.Millisecond)
	})

	// Emit events to create workers
	c.Emit(context.Background(), sig1, key.Field("test1"))
	c.Emit(context.Background(), sig2, key.Field("test2"))

	// Give workers time to start
	time.Sleep(10 * time.Millisecond)

	stats := c.Stats()

	if stats.ActiveWorkers != 2 {
		t.Errorf("expected 2 active workers, got %d", stats.ActiveWorkers)
	}

	if stats.ListenerCounts[sig1] != 2 {
		t.Errorf("expected 2 listeners for sig1, got %d", stats.ListenerCounts[sig1])
	}

	if stats.ListenerCounts[sig2] != 1 {
		t.Errorf("expected 1 listener for sig2, got %d", stats.ListenerCounts[sig2])
	}

	// QueueDepths should be 0 or small (events being processed)
	if _, exists := stats.QueueDepths[sig1]; !exists {
		t.Error("expected QueueDepths to contain sig1")
	}
	if _, exists := stats.QueueDepths[sig2]; !exists {
		t.Error("expected QueueDepths to contain sig2")
	}
}

// TestMultipleOptions verifies multiple options can be combined.
func TestMultipleOptions(t *testing.T) {
	var handlerCalled bool

	c := New(
		WithBufferSize(256),
		WithPanicHandler(func(_ Signal, _ any) {
			handlerCalled = true
		}),
		WithSyncMode(),
	)
	defer c.Shutdown()

	if c.bufferSize != 256 {
		t.Errorf("expected bufferSize=256, got %d", c.bufferSize)
	}

	if c.panicHandler == nil {
		t.Error("expected panicHandler to be set")
	}

	// Verify handler works
	sig := Signal("test.multi")
	key := NewStringKey("value")

	c.Hook(sig, func(_ context.Context, _ *Event) {
		panic("test")
	})

	c.Emit(context.Background(), sig, key.Field("test"))

	if !handlerCalled {
		t.Error("expected panic handler to be called")
	}
}

// TestWithSyncMode verifies synchronous event processing.
func TestWithSyncMode(t *testing.T) {
	c := New(WithSyncMode())
	defer c.Shutdown()

	if !c.syncMode {
		t.Error("expected syncMode to be true")
	}

	// Verify events are processed synchronously (no timing needed)
	sig := Signal("test.sync")
	key := NewStringKey("value")
	var called bool

	c.Hook(sig, func(_ context.Context, e *Event) {
		called = true
		val, ok := key.From(e)
		if !ok || val != "sync-test" {
			t.Errorf("expected value='sync-test', got %v, %v", val, ok)
		}
	})

	c.Emit(context.Background(), sig, key.Field("sync-test"))

	// No sleep needed - should be processed immediately
	if !called {
		t.Error("expected listener to be called synchronously")
	}

	// Verify no workers were created
	stats := c.Stats()
	if stats.ActiveWorkers != 0 {
		t.Errorf("expected 0 active workers in sync mode, got %d", stats.ActiveWorkers)
	}
}
