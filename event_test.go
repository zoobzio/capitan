package capitan

import (
	"testing"
	"time"
)

func TestEventGet(t *testing.T) {
	sig := Signal("test.get")
	strKey := NewStringKey("name")
	intKey := NewIntKey("count")

	event := newEvent(sig, strKey.Field("test"), intKey.Field(42))

	nameField := event.Get(strKey)
	if nameField == nil {
		t.Fatal("name field not found")
	}

	sf, ok := nameField.(StringField)
	if !ok {
		t.Fatal("name field wrong type")
	}
	if sf.String() != "test" {
		t.Errorf("expected %q, got %q", "test", sf.String())
	}

	countField := event.Get(intKey)
	if countField == nil {
		t.Fatal("count field not found")
	}

	cf, ok := countField.(IntField)
	if !ok {
		t.Fatal("count field wrong type")
	}
	if cf.Int() != 42 {
		t.Errorf("expected %d, got %d", 42, cf.Int())
	}
}

func TestEventGetMissingKey(t *testing.T) {
	sig := Signal("test.missing")
	strKey := NewStringKey("existing")
	missingKey := NewStringKey("missing")

	event := newEvent(sig, strKey.Field("test"))

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

	event := newEvent(sig,
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
		case StringField:
			if f.String() == "test" {
				foundStr = true
			}
		case IntField:
			if f.Int() == 42 {
				foundInt = true
			}
		case BoolField:
			if f.Bool() == true {
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

	event := newEvent(sig, key.Field("test"))

	if event.Signal() != sig {
		t.Errorf("expected signal %q, got %q", sig, event.Signal())
	}
}

func TestEventTimestamp(t *testing.T) {
	sig := Signal("test.timestamp")
	key := NewStringKey("value")

	before := time.Now()
	event := newEvent(sig, key.Field("test"))
	after := time.Now()

	if event.Timestamp().Before(before) || event.Timestamp().After(after) {
		t.Errorf("timestamp %v not between %v and %v", event.Timestamp(), before, after)
	}
}

func TestEventPooling(t *testing.T) {
	sig := Signal("test.pool")
	key := NewStringKey("value")

	// Create first event with "first" value
	event1 := newEvent(sig, key.Field("first"))
	field1 := event1.Get(key).(StringField)
	if field1.String() != "first" {
		t.Errorf("event1: expected %q, got %q", "first", field1.String())
	}

	// Return to pool
	eventPool.Put(event1)

	// Create second event with "second" value
	event2 := newEvent(sig, key.Field("second"))
	field2 := event2.Get(key).(StringField)
	if field2.String() != "second" {
		t.Errorf("event2: expected %q, got %q - pool not cleared properly", "second", field2.String())
	}

	// Verify field was cleared (whether pooled or not, fields should be correct)
	if field2.String() == "first" {
		t.Error("fields not cleared when reusing pooled event")
	}
}

func TestEventFieldsDefensiveCopy(t *testing.T) {
	sig := Signal("test.defensive")
	key := NewStringKey("value")

	event := newEvent(sig, key.Field("test"))

	fields1 := event.Fields()
	fields2 := event.Fields()

	// Should be different slices
	if &fields1[0] == &fields2[0] {
		t.Error("Fields() not returning defensive copy")
	}
}
