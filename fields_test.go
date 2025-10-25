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
		receivedStr = e.Get(strKey).(StringField).String()
		receivedInt = e.Get(intKey).(IntField).Int()
		receivedFloat = e.Get(floatKey).(Float64Field).Float64()
		receivedBool = e.Get(boolKey).(BoolField).Bool()
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
