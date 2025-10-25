package capitan

import (
	"context"
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

	sf, ok := field.(GenericField[string])
	if !ok {
		t.Fatal("field not GenericField[string] type")
	}

	if sf.Get() != "hello" {
		t.Errorf("expected %q, got %q", "hello", sf.Get())
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

	inf, ok := field.(GenericField[int])
	if !ok {
		t.Fatal("field not GenericField[int] type")
	}

	if inf.Get() != 42 {
		t.Errorf("expected %d, got %d", 42, inf.Get())
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

	ff, ok := field.(GenericField[float64])
	if !ok {
		t.Fatal("field not GenericField[float64] type")
	}

	if ff.Get() != 3.14 {
		t.Errorf("expected %f, got %f", 3.14, ff.Get())
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

	bf, ok := field.(GenericField[bool])
	if !ok {
		t.Fatal("field not GenericField[bool] type")
	}

	if bf.Get() != true {
		t.Errorf("expected %v, got %v", true, bf.Get())
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

	c.Hook(sig, func(_ context.Context, e *Event) {
		receivedStr, _ = strKey.From(e)
		receivedInt, _ = intKey.From(e)
		receivedFloat, _ = floatKey.From(e)
		receivedBool, _ = boolKey.From(e)
		wg.Done()
	})

	c.Emit(context.Background(), sig,
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

	c.Hook(sig, func(_ context.Context, e *Event) {
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

	c.Emit(context.Background(), sig,
		strKey.Field("test"),
		intKey.Field(100),
		floatKey.Field(2.5),
		boolKey.Field(true),
	)

	wg.Wait()
}

// Custom struct type for testing custom fields.
type OrderInfo struct {
	ID     string
	Total  float64
	Items  int
	Active bool
}

// Custom variant.
const VariantOrderInfo Variant = "test.OrderInfo"

// Custom key implementation.
type OrderInfoKey struct {
	name string
}

func NewOrderInfoKey(name string) OrderInfoKey {
	return OrderInfoKey{name: name}
}

func (k OrderInfoKey) Name() string   { return k.name }
func (OrderInfoKey) Variant() Variant { return VariantOrderInfo }
func (k OrderInfoKey) Field(value OrderInfo) Field {
	return GenericField[OrderInfo]{
		key:     k,
		value:   value,
		variant: k.Variant(),
	}
}

func (k OrderInfoKey) From(e *Event) (OrderInfo, bool) {
	f := e.Get(k)
	if f == nil {
		return OrderInfo{}, false
	}
	if gf, ok := f.(GenericField[OrderInfo]); ok {
		return gf.Get(), true
	}
	return OrderInfo{}, false
}

func TestCustomStructField(t *testing.T) {
	c := New()
	defer c.Shutdown()

	sig := Signal("test.custom")
	orderKey := NewOrderInfoKey("order")

	expectedOrder := OrderInfo{
		ID:     "ORDER-123",
		Total:  99.99,
		Items:  3,
		Active: true,
	}

	var receivedOrder OrderInfo
	var wg sync.WaitGroup
	wg.Add(1)

	c.Hook(sig, func(_ context.Context, e *Event) {
		order, ok := orderKey.From(e)
		if !ok {
			t.Error("failed to extract order from event")
		}
		receivedOrder = order
		wg.Done()
	})

	c.Emit(context.Background(), sig, orderKey.Field(expectedOrder))

	wg.Wait()

	if receivedOrder.ID != expectedOrder.ID {
		t.Errorf("ID: expected %q, got %q", expectedOrder.ID, receivedOrder.ID)
	}
	if receivedOrder.Total != expectedOrder.Total {
		t.Errorf("Total: expected %.2f, got %.2f", expectedOrder.Total, receivedOrder.Total)
	}
	if receivedOrder.Items != expectedOrder.Items {
		t.Errorf("Items: expected %d, got %d", expectedOrder.Items, receivedOrder.Items)
	}
	if receivedOrder.Active != expectedOrder.Active {
		t.Errorf("Active: expected %v, got %v", expectedOrder.Active, receivedOrder.Active)
	}
}

func TestCustomStructFieldVariant(t *testing.T) {
	orderKey := NewOrderInfoKey("order")
	field := orderKey.Field(OrderInfo{ID: "TEST"})

	if field.Variant() != VariantOrderInfo {
		t.Errorf("expected variant %v, got %v", VariantOrderInfo, field.Variant())
	}
}

func TestCustomStructFieldTypeAssertion(t *testing.T) {
	orderKey := NewOrderInfoKey("order")
	expectedOrder := OrderInfo{
		ID:     "ORDER-456",
		Total:  199.99,
		Items:  5,
		Active: false,
	}

	field := orderKey.Field(expectedOrder)

	// Test type assertion
	gf, ok := field.(GenericField[OrderInfo])
	if !ok {
		t.Fatal("field not GenericField[OrderInfo] type")
	}

	receivedOrder := gf.Get()

	if receivedOrder.ID != expectedOrder.ID {
		t.Errorf("ID: expected %q, got %q", expectedOrder.ID, receivedOrder.ID)
	}
	if receivedOrder.Total != expectedOrder.Total {
		t.Errorf("Total: expected %.2f, got %.2f", expectedOrder.Total, receivedOrder.Total)
	}
	if receivedOrder.Items != expectedOrder.Items {
		t.Errorf("Items: expected %d, got %d", expectedOrder.Items, receivedOrder.Items)
	}
	if receivedOrder.Active != expectedOrder.Active {
		t.Errorf("Active: expected %v, got %v", expectedOrder.Active, receivedOrder.Active)
	}
}

func TestNewKeyGeneric(t *testing.T) {
	// Test NewKey[T] with custom struct type
	type CustomData struct {
		Value string
		Count int
	}

	customKey := NewKey[CustomData]("custom", "test.CustomData")

	if customKey.Name() != "custom" {
		t.Errorf("expected name %q, got %q", "custom", customKey.Name())
	}

	if customKey.Variant() != "test.CustomData" {
		t.Errorf("expected variant %q, got %v", "test.CustomData", customKey.Variant())
	}

	// Test creating field and extracting value
	c := New()
	defer c.Shutdown()

	sig := Signal("test.newkey")
	expectedData := CustomData{Value: "test", Count: 42}

	var received CustomData
	var wg sync.WaitGroup
	wg.Add(1)

	c.Hook(sig, func(_ context.Context, e *Event) {
		data, ok := customKey.From(e)
		if !ok {
			t.Error("failed to extract custom data from event")
		}
		received = data
		wg.Done()
	})

	c.Emit(context.Background(), sig, customKey.Field(expectedData))

	wg.Wait()

	if received.Value != expectedData.Value {
		t.Errorf("Value: expected %q, got %q", expectedData.Value, received.Value)
	}
	if received.Count != expectedData.Count {
		t.Errorf("Count: expected %d, got %d", expectedData.Count, received.Count)
	}
}

func TestFromTypeMismatch(t *testing.T) {
	// Test that From() returns false when types don't match
	c := New()
	defer c.Shutdown()

	sig := Signal("test.mismatch")
	strKey := NewStringKey("data")
	intKey := NewIntKey("data") // Same name, different type

	var wg sync.WaitGroup
	wg.Add(1)

	c.Hook(sig, func(_ context.Context, e *Event) {
		defer wg.Done()

		// Event has string field named "data"
		str, ok := strKey.From(e)
		if !ok || str != "hello" {
			t.Errorf("StringKey.From: expected %q (ok=true), got %q (ok=%v)", "hello", str, ok)
		}

		// Trying to extract as int with same name should fail
		intVal, ok := intKey.From(e)
		if ok || intVal != 0 {
			t.Errorf("IntKey.From (type mismatch): expected 0 (ok=false), got %d (ok=%v)", intVal, ok)
		}
	})

	c.Emit(context.Background(), sig, strKey.Field("hello"))

	wg.Wait()
}
