package capitan

import (
	"context"
	"testing"
	"time"
)

func TestEventGet(t *testing.T) {
	sig := Signal("test.get")
	strKey := NewStringKey("name")
	intKey := NewIntKey("count")

	event := newEvent(context.Background(), sig, strKey.Field("test"), intKey.Field(42))

	nameField := event.Get(strKey)
	if nameField == nil {
		t.Fatal("name field not found")
	}

	sf, ok := nameField.(GenericField[string])
	if !ok {
		t.Fatal("name field wrong type")
	}
	if sf.Get() != "test" {
		t.Errorf("expected %q, got %q", "test", sf.Get())
	}

	countField := event.Get(intKey)
	if countField == nil {
		t.Fatal("count field not found")
	}

	cf, ok := countField.(GenericField[int])
	if !ok {
		t.Fatal("count field wrong type")
	}
	if cf.Get() != 42 {
		t.Errorf("expected %d, got %d", 42, cf.Get())
	}
}

func TestEventGetMissingKey(t *testing.T) {
	sig := Signal("test.missing")
	strKey := NewStringKey("existing")
	missingKey := NewStringKey("missing")

	event := newEvent(context.Background(), sig, strKey.Field("test"))

	field := event.Get(missingKey)
	if field != nil {
		t.Errorf("expected nil for missing key, got %v", field)
	}
}

func TestEventFields(t *testing.T) {
	sig := Signal("test.fields")
	strKey := NewStringKey("name")
	intKey := NewIntKey("count")
	boolKey := NewBoolKey("active")

	event := newEvent(context.Background(), sig,
		strKey.Field("test"),
		intKey.Field(42),
		boolKey.Field(true),
	)

	fields := event.Fields()

	if len(fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(fields))
	}

	// Check all fields present (order not guaranteed)
	foundStr, foundInt, foundBool := false, false, false
	for _, field := range fields {
		switch f := field.(type) {
		case GenericField[string]:
			if f.Get() == "test" {
				foundStr = true
			}
		case GenericField[int]:
			if f.Get() == 42 {
				foundInt = true
			}
		case GenericField[bool]:
			if f.Get() == true {
				foundBool = true
			}
		}
	}

	if !foundStr {
		t.Error("string field not found in Fields()")
	}
	if !foundInt {
		t.Error("int field not found in Fields()")
	}
	if !foundBool {
		t.Error("bool field not found in Fields()")
	}
}

func TestEventSignal(t *testing.T) {
	sig := Signal("test.signal")
	key := NewStringKey("value")

	event := newEvent(context.Background(), sig, key.Field("test"))

	if event.Signal() != sig {
		t.Errorf("expected signal %q, got %q", sig, event.Signal())
	}
}

func TestEventTimestamp(t *testing.T) {
	sig := Signal("test.timestamp")
	key := NewStringKey("value")

	before := time.Now()
	event := newEvent(context.Background(), sig, key.Field("test"))
	after := time.Now()

	if event.Timestamp().Before(before) || event.Timestamp().After(after) {
		t.Errorf("timestamp %v not between %v and %v", event.Timestamp(), before, after)
	}
}

func TestEventPooling(t *testing.T) {
	sig := Signal("test.pool")
	key := NewStringKey("value")

	// Create first event with "first" value
	event1 := newEvent(context.Background(), sig, key.Field("first"))
	field1 := event1.Get(key).(GenericField[string])
	if field1.Get() != "first" {
		t.Errorf("event1: expected %q, got %q", "first", field1.Get())
	}

	// Return to pool
	eventPool.Put(event1)

	// Create second event with "second" value
	event2 := newEvent(context.Background(), sig, key.Field("second"))
	field2 := event2.Get(key).(GenericField[string])
	if field2.Get() != "second" {
		t.Errorf("event2: expected %q, got %q - pool not cleared properly", "second", field2.Get())
	}

	// Verify field was cleared (whether pooled or not, fields should be correct)
	if field2.Get() == "first" {
		t.Error("fields not cleared when reusing pooled event")
	}
}

func TestEventFieldsDefensiveCopy(t *testing.T) {
	sig := Signal("test.defensive")
	key := NewStringKey("value")

	event := newEvent(context.Background(), sig, key.Field("test"))

	fields1 := event.Fields()
	fields2 := event.Fields()

	// Should be different slices
	if &fields1[0] == &fields2[0] {
		t.Error("Fields() not returning defensive copy")
	}
}

func TestEventContext(t *testing.T) {
	sig := Signal("test.event.context")
	key := NewStringKey("value")

	type ctxKey string
	const testKey ctxKey = "test_key"
	expectedValue := "test_value"

	ctx := context.WithValue(context.Background(), testKey, expectedValue)
	event := newEvent(ctx, sig, key.Field("test"))

	eventCtx := event.Context()
	if eventCtx == nil {
		t.Fatal("Event.Context() returned nil")
	}

	receivedValue := eventCtx.Value(testKey)
	if receivedValue != expectedValue {
		t.Errorf("expected context value %q, got %v", expectedValue, receivedValue)
	}
}

func TestEventContextBackground(t *testing.T) {
	sig := Signal("test.background")
	key := NewStringKey("value")

	event := newEvent(context.Background(), sig, key.Field("test"))

	ctx := event.Context()
	if ctx == nil {
		t.Fatal("Event.Context() returned nil")
	}

	if ctx.Err() != nil {
		t.Errorf("background context should not have error, got %v", ctx.Err())
	}
}
