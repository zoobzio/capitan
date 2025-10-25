package capitan

import (
	"sync"
	"testing"
	"time"
)

func TestNew(t *testing.T) {
	c := New()
	if c == nil {
		t.Fatal("New() returned nil")
	}
	if c.registry == nil {
		t.Error("registry not initialized")
	}
	if c.workers == nil {
		t.Error("workers not initialized")
	}
	if c.shutdown == nil {
		t.Error("shutdown channel not initialized")
	}
}

func TestEmitCreatesWorkerLazily(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.lazy")
	key := NewStringKey("value")

	// Before emit, no worker exists
	c.mu.RLock()
	_, exists := c.workers[sig]
	c.mu.RUnlock()

	if exists {
		t.Error("worker should not exist before first emit")
	}

	var wg sync.WaitGroup
	wg.Add(1)

	c.Hook(sig, func(_ *Event) {
		wg.Done()
	})

	c.Emit(sig, key.Field("test"))

	// After emit, worker should exist
	c.mu.RLock()
	_, exists = c.workers[sig]
	c.mu.RUnlock()

	if !exists {
		t.Error("worker should exist after first emit")
	}

	wg.Wait()
}

func TestEmitAsyncProcessing(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.async")
	key := NewIntKey("value")

	const numEmissions = 100
	received := make([]int, 0, numEmissions)
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(numEmissions)

	c.Hook(sig, func(e *Event) {
		field := e.Get(key).(GenericField[int])
		mu.Lock()
		received = append(received, field.Get())
		mu.Unlock()
		wg.Done()
	})

	// Emit many events rapidly - should not block
	for i := 0; i < numEmissions; i++ {
		c.Emit(sig, key.Field(i))
	}

	wg.Wait()

	if len(received) != numEmissions {
		t.Errorf("expected %d events, got %d", numEmissions, len(received))
	}
}

func TestEmitWithMultipleListeners(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.multi")
	key := NewStringKey("value")

	var received1, received2 string
	var wg sync.WaitGroup
	wg.Add(2)

	c.Hook(sig, func(e *Event) {
		field := e.Get(key).(GenericField[string])
		received1 = field.Get()
		wg.Done()
	})

	c.Hook(sig, func(e *Event) {
		field := e.Get(key).(GenericField[string])
		received2 = field.Get()
		wg.Done()
	})

	c.Emit(sig, key.Field("hello"))

	wg.Wait()

	if received1 != "hello" {
		t.Errorf("listener 1: expected %q, got %q", "hello", received1)
	}
	if received2 != "hello" {
		t.Errorf("listener 2: expected %q, got %q", "hello", received2)
	}
}

func TestEmitWithNoListeners(_ *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.nolisteners")
	key := NewStringKey("value")

	// Should not panic or block
	c.Emit(sig, key.Field("test"))

	// Give time for any processing
	time.Sleep(10 * time.Millisecond)
}

func TestEmitWithPanicRecovery(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.panic")
	key := NewStringKey("value")

	var wg sync.WaitGroup
	wg.Add(1)

	// First listener panics
	c.Hook(sig, func(_ *Event) {
		panic("intentional panic")
	})

	// Second listener should still execute
	c.Hook(sig, func(_ *Event) {
		wg.Done()
	})

	c.Emit(sig, key.Field("test"))

	// If panic recovery doesn't work, this times out
	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success
	case <-time.After(time.Second):
		t.Fatal("second listener never executed - panic not recovered")
	}
}

func TestHook(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.hook")
	key := NewStringKey("value")

	var received *Event
	var wg sync.WaitGroup
	wg.Add(1)

	listener := c.Hook(sig, func(e *Event) {
		received = e
		wg.Done()
	})

	if listener == nil {
		t.Fatal("Hook returned nil")
	}

	c.Emit(sig, key.Field("test"))

	wg.Wait()

	if received == nil {
		t.Error("event not received")
	}
}

func TestObserve(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig1 := Signal("test.sig1")
	sig2 := Signal("test.sig2")
	key := NewStringKey("msg")

	// Create hooks first so signals exist in registry
	c.Hook(sig1, func(_ *Event) {})
	c.Hook(sig2, func(_ *Event) {})

	var received []Signal
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(2)

	observer := c.Observe(func(e *Event) {
		mu.Lock()
		received = append(received, e.Signal())
		mu.Unlock()
		wg.Done()
	})

	if observer == nil {
		t.Fatal("Observe returned nil")
	}

	c.Emit(sig1, key.Field("first"))
	c.Emit(sig2, key.Field("second"))

	wg.Wait()

	if len(received) != 2 {
		t.Fatalf("expected 2 events, got %d", len(received))
	}

	// Order not guaranteed, check both present
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

func TestShutdown(_ *testing.T) {
	c := New()

	sig := Signal("test.shutdown")
	key := NewStringKey("value")

	var wg sync.WaitGroup
	wg.Add(1)

	c.Hook(sig, func(_ *Event) {
		// Simulate some work
		time.Sleep(10 * time.Millisecond)
		wg.Done()
	})

	c.Emit(sig, key.Field("test"))

	// Shutdown should wait for in-flight events
	c.Shutdown()

	// If shutdown didn't wait, this would fail
	wg.Wait()
}

func TestModuleLevelAPI(t *testing.T) {
	sig := Signal("test.module")
	key := NewStringKey("value")

	var received *Event
	var wg sync.WaitGroup
	wg.Add(1)

	listener := Hook(sig, func(e *Event) {
		received = e
		wg.Done()
	})

	if listener == nil {
		t.Fatal("Hook returned nil")
	}

	Emit(sig, key.Field("test"))

	wg.Wait()

	if received == nil {
		t.Error("event not received")
	}
	if received.Signal() != sig {
		t.Errorf("expected signal %q, got %q", sig, received.Signal())
	}
}
