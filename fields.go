package capitan

// GenericKey is a Key implementation for any type T.
// All built-in key types (StringKey, IntKey, etc.) are aliases of GenericKey[T].
type GenericKey[T any] struct {
	name    string
	variant Variant
}

// Name returns the semantic identifier.
func (k GenericKey[T]) Name() string { return k.name }

// Variant returns the type constraint.
func (k GenericKey[T]) Variant() Variant { return k.variant }

// Field creates a GenericField with this key and the given value.
func (k GenericKey[T]) Field(value T) Field {
	return GenericField[T]{key: k, value: value, variant: k.variant}
}

// From extracts the typed value for this key from the event.
// Returns the value and true if present, or zero value and false if not present or wrong type.
func (k GenericKey[T]) From(e *Event) (T, bool) {
	var zero T
	f := e.Get(k)
	if f == nil {
		return zero, false
	}
	if gf, ok := f.(GenericField[T]); ok {
		return gf.Get(), true
	}
	return zero, false
}

// NewKey creates a GenericKey for any type T with the given name and variant.
// Use a namespaced variant string to avoid collisions (e.g., "myapp.OrderInfo").
//
// Example:
//
//	type OrderInfo struct { ID string; Total float64 }
//	orderKey := capitan.NewKey[OrderInfo]("order", "myapp.OrderInfo")
//	capitan.Emit(sig, orderKey.Field(OrderInfo{ID: "123", Total: 99.99}))
func NewKey[T any](name string, variant Variant) GenericKey[T] {
	return GenericKey[T]{name: name, variant: variant}
}

// StringKey is a Key implementation for string values.
type StringKey = GenericKey[string]

// NewStringKey creates a StringKey with the given name.
func NewStringKey(name string) StringKey {
	return GenericKey[string]{name: name, variant: VariantString}
}

// IntKey is a Key implementation for int values.
type IntKey = GenericKey[int]

// NewIntKey creates an IntKey with the given name.
func NewIntKey(name string) IntKey {
	return GenericKey[int]{name: name, variant: VariantInt}
}

// Float64Key is a Key implementation for float64 values.
type Float64Key = GenericKey[float64]

// NewFloat64Key creates a Float64Key with the given name.
func NewFloat64Key(name string) Float64Key {
	return GenericKey[float64]{name: name, variant: VariantFloat64}
}

// BoolKey is a Key implementation for bool values.
type BoolKey = GenericKey[bool]

// NewBoolKey creates a BoolKey with the given name.
func NewBoolKey(name string) BoolKey {
	return GenericKey[bool]{name: name, variant: VariantBool}
}
