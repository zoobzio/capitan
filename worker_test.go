package capitan

import (
	"sync"
	"testing"
	"time"
)

func TestWorkerCreatedLazily(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.worker.lazy")
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

func TestWorkerProcessesEventsAsync(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.worker.async")
	key := NewIntKey("value")

	const numEvents = 100
	received := make([]int, 0, numEvents)
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(numEvents)

	c.Hook(sig, func(e *Event) {
		field := e.Get(key).(IntField)
		mu.Lock()
		received = append(received, field.Int())
		mu.Unlock()
		wg.Done()
	})

	// Emit many events rapidly - should not block caller
	start := time.Now()
	for i := 0; i < numEvents; i++ {
		c.Emit(sig, key.Field(i))
	}
	emitDuration := time.Since(start)

	// Emissions should be very fast (not blocking on listener execution)
	if emitDuration > 100*time.Millisecond {
		t.Errorf("emissions took too long: %v", emitDuration)
	}

	wg.Wait()

	if len(received) != numEvents {
		t.Errorf("expected %d events, got %d", numEvents, len(received))
	}
}

func TestWorkerInvokesAllListeners(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.worker.all")
	key := NewStringKey("value")

	count1, count2, count3 := 0, 0, 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(3)

	c.Hook(sig, func(_ *Event) {
		mu.Lock()
		count1++
		mu.Unlock()
		wg.Done()
	})

	c.Hook(sig, func(_ *Event) {
		mu.Lock()
		count2++
		mu.Unlock()
		wg.Done()
	})

	c.Hook(sig, func(_ *Event) {
		mu.Lock()
		count3++
		mu.Unlock()
		wg.Done()
	})

	c.Emit(sig, key.Field("test"))

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if count1 != 1 || count2 != 1 || count3 != 1 {
		t.Errorf("expected all listeners invoked once, got %d, %d, %d", count1, count2, count3)
	}
}

func TestWorkerPanicRecovery(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.worker.panic")
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

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case <-done:
		// Success - second listener executed despite first panic
	case <-time.After(time.Second):
		t.Fatal("second listener never executed - panic not recovered")
	}
}

func TestWorkerShutdownDrainsQueue(t *testing.T) {
	c := New()

	sig := Signal("test.worker.drain")
	key := NewStringKey("value")

	processed := 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(5)

	c.Hook(sig, func(_ *Event) {
		time.Sleep(10 * time.Millisecond)
		mu.Lock()
		processed++
		mu.Unlock()
		wg.Done()
	})

	// Emit multiple events
	for i := 0; i < 5; i++ {
		c.Emit(sig, key.Field("test"))
	}

	// Shutdown should wait for all events to be processed
	c.Shutdown()

	mu.Lock()
	finalCount := processed
	mu.Unlock()

	if finalCount != 5 {
		t.Errorf("expected 5 events processed, got %d", finalCount)
	}
}

func TestWorkerNoListeners(_ *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.worker.none")
	key := NewStringKey("value")

	// Should not panic or hang
	c.Emit(sig, key.Field("test"))

	time.Sleep(10 * time.Millisecond)
}

func TestWorkerMultipleSignalsIsolated(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig1 := Signal("test.worker.iso1")
	sig2 := Signal("test.worker.iso2")
	key := NewStringKey("value")

	count1, count2 := 0, 0
	var mu sync.Mutex
	var wg sync.WaitGroup
	wg.Add(3)

	c.Hook(sig1, func(_ *Event) {
		mu.Lock()
		count1++
		mu.Unlock()
		wg.Done()
	})

	c.Hook(sig2, func(_ *Event) {
		mu.Lock()
		count2++
		mu.Unlock()
		wg.Done()
	})

	c.Emit(sig1, key.Field("first"))
	c.Emit(sig1, key.Field("second"))
	c.Emit(sig2, key.Field("third"))

	wg.Wait()

	mu.Lock()
	defer mu.Unlock()

	if count1 != 2 {
		t.Errorf("sig1: expected 2 events, got %d", count1)
	}
	if count2 != 1 {
		t.Errorf("sig2: expected 1 event, got %d", count2)
	}
}
