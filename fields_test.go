package capitan

import (
	"sync"
	"testing"
)

func TestStringKey(t *testing.T) {
	key := NewStringKey("test")

	if key.Name() != "test" {
		t.Errorf("expected name %q, got %q", "test", key.Name())
	}

	if key.Variant() != VariantString {
		t.Errorf("expected variant %v, got %v", VariantString, key.Variant())
	}
}

func TestStringField(t *testing.T) {
	key := NewStringKey("name")
	field := key.Field("hello")

	if field.Key().Name() != "name" {
		t.Errorf("expected key name %q, got %q", "name", field.Key().Name())
	}

	if field.Variant() != VariantString {
		t.Errorf("expected variant %v, got %v", VariantString, field.Variant())
	}

	if field.Value() != "hello" {
		t.Errorf("expected value %q, got %v", "hello", field.Value())
	}

	sf, ok := field.(StringField)
	if !ok {
		t.Fatal("field not StringField type")
	}

	if sf.String() != "hello" {
		t.Errorf("expected %q, got %q", "hello", sf.String())
	}
}

func TestIntKey(t *testing.T) {
	key := NewIntKey("count")

	if key.Name() != "count" {
		t.Errorf("expected name %q, got %q", "count", key.Name())
	}

	if key.Variant() != VariantInt {
		t.Errorf("expected variant %v, got %v", VariantInt, key.Variant())
	}
}

func TestIntField(t *testing.T) {
	key := NewIntKey("count")
	field := key.Field(42)

	if field.Key().Name() != "count" {
		t.Errorf("expected key name %q, got %q", "count", field.Key().Name())
	}

	if field.Variant() != VariantInt {
		t.Errorf("expected variant %v, got %v", VariantInt, field.Variant())
	}

	if field.Value() != 42 {
		t.Errorf("expected value %d, got %v", 42, field.Value())
	}

	inf, ok := field.(IntField)
	if !ok {
		t.Fatal("field not IntField type")
	}

	if inf.Int() != 42 {
		t.Errorf("expected %d, got %d", 42, inf.Int())
	}
}

func TestFloat64Key(t *testing.T) {
	key := NewFloat64Key("ratio")

	if key.Name() != "ratio" {
		t.Errorf("expected name %q, got %q", "ratio", key.Name())
	}

	if key.Variant() != VariantFloat64 {
		t.Errorf("expected variant %v, got %v", VariantFloat64, key.Variant())
	}
}

func TestFloat64Field(t *testing.T) {
	key := NewFloat64Key("ratio")
	field := key.Field(3.14)

	if field.Key().Name() != "ratio" {
		t.Errorf("expected key name %q, got %q", "ratio", field.Key().Name())
	}

	if field.Variant() != VariantFloat64 {
		t.Errorf("expected variant %v, got %v", VariantFloat64, field.Variant())
	}

	if field.Value() != 3.14 {
		t.Errorf("expected value %f, got %v", 3.14, field.Value())
	}

	ff, ok := field.(Float64Field)
	if !ok {
		t.Fatal("field not Float64Field type")
	}

	if ff.Float64() != 3.14 {
		t.Errorf("expected %f, got %f", 3.14, ff.Float64())
	}
}

func TestBoolKey(t *testing.T) {
	key := NewBoolKey("active")

	if key.Name() != "active" {
		t.Errorf("expected name %q, got %q", "active", key.Name())
	}

	if key.Variant() != VariantBool {
		t.Errorf("expected variant %v, got %v", VariantBool, key.Variant())
	}
}

func TestBoolField(t *testing.T) {
	key := NewBoolKey("active")
	field := key.Field(true)

	if field.Key().Name() != "active" {
		t.Errorf("expected key name %q, got %q", "active", field.Key().Name())
	}

	if field.Variant() != VariantBool {
		t.Errorf("expected variant %v, got %v", VariantBool, field.Variant())
	}

	if field.Value() != true {
		t.Errorf("expected value %v, got %v", true, field.Value())
	}

	bf, ok := field.(BoolField)
	if !ok {
		t.Fatal("field not BoolField type")
	}

	if bf.Bool() != true {
		t.Errorf("expected %v, got %v", true, bf.Bool())
	}
}

func TestAllFieldTypes(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.alltypes")

	strKey := NewStringKey("str")
	intKey := NewIntKey("int")
	floatKey := NewFloat64Key("float")
	boolKey := NewBoolKey("bool")

	var receivedStr string
	var receivedInt int
	var receivedFloat float64
	var receivedBool bool
	var wg sync.WaitGroup
	wg.Add(1)

	c.Hook(sig, func(e *Event) {
		receivedStr, _ = strKey.From(e)
		receivedInt, _ = intKey.From(e)
		receivedFloat, _ = floatKey.From(e)
		receivedBool, _ = boolKey.From(e)
		wg.Done()
	})

	c.Emit(sig,
		strKey.Field("hello"),
		intKey.Field(42),
		floatKey.Field(3.14),
		boolKey.Field(true),
	)

	wg.Wait()

	if receivedStr != "hello" {
		t.Errorf("string: expected %q, got %q", "hello", receivedStr)
	}
	if receivedInt != 42 {
		t.Errorf("int: expected %d, got %d", 42, receivedInt)
	}
	if receivedFloat != 3.14 {
		t.Errorf("float: expected %f, got %f", 3.14, receivedFloat)
	}
	if receivedBool != true {
		t.Errorf("bool: expected %v, got %v", true, receivedBool)
	}
}

func TestFromMethods(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.from")

	strKey := NewStringKey("str")
	intKey := NewIntKey("int")
	floatKey := NewFloat64Key("float")
	boolKey := NewBoolKey("bool")

	var wg sync.WaitGroup
	wg.Add(1)

	c.Hook(sig, func(e *Event) {
		defer wg.Done()

		// Test From() methods
		if got, ok := strKey.From(e); !ok || got != "test" {
			t.Errorf("StringKey.From: expected %q (ok=true), got %q (ok=%v)", "test", got, ok)
		}
		if got, ok := intKey.From(e); !ok || got != 100 {
			t.Errorf("IntKey.From: expected %d (ok=true), got %d (ok=%v)", 100, got, ok)
		}
		if got, ok := floatKey.From(e); !ok || got != 2.5 {
			t.Errorf("Float64Key.From: expected %f (ok=true), got %f (ok=%v)", 2.5, got, ok)
		}
		if got, ok := boolKey.From(e); !ok || got != true {
			t.Errorf("BoolKey.From: expected %v (ok=true), got %v (ok=%v)", true, got, ok)
		}

		// Test From() on missing fields (should return zero values and ok=false)
		missingStr := NewStringKey("missing_str")
		if got, ok := missingStr.From(e); ok || got != "" {
			t.Errorf("StringKey.From (missing): expected %q (ok=false), got %q (ok=%v)", "", got, ok)
		}

		missingInt := NewIntKey("missing_int")
		if got, ok := missingInt.From(e); ok || got != 0 {
			t.Errorf("IntKey.From (missing): expected %d (ok=false), got %d (ok=%v)", 0, got, ok)
		}

		missingFloat := NewFloat64Key("missing_float")
		if got, ok := missingFloat.From(e); ok || got != 0 {
			t.Errorf("Float64Key.From (missing): expected %f (ok=false), got %f (ok=%v)", 0.0, got, ok)
		}

		missingBool := NewBoolKey("missing_bool")
		if got, ok := missingBool.From(e); ok || got != false {
			t.Errorf("BoolKey.From (missing): expected %v (ok=false), got %v (ok=%v)", false, got, ok)
		}
	})

	c.Emit(sig,
		strKey.Field("test"),
		intKey.Field(100),
		floatKey.Field(2.5),
		boolKey.Field(true),
	)

	wg.Wait()
}
