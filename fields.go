package capitan

// StringKey is a Key implementation for string values.
type StringKey struct {
	name string
}

// Name returns the semantic identifier.
func (k StringKey) Name() string { return k.name }

// Variant returns the type constraint.
func (StringKey) Variant() Variant { return VariantString }

// Field creates a GenericField with this key and the given value.
func (k StringKey) Field(value string) Field {
	return GenericField[string]{key: k, value: value, variant: k.Variant()}
}

// From extracts the typed string value for this key from the event.
// Returns the value and true if present, or zero value and false if not present or wrong type.
func (k StringKey) From(e *Event) (string, bool) {
	f := e.Get(k)
	if f == nil {
		return "", false
	}
	if sf, ok := f.(GenericField[string]); ok {
		return sf.Get(), true
	}
	return "", false
}

// NewStringKey creates a StringKey with the given name.
func NewStringKey(name string) StringKey {
	return StringKey{name: name}
}

// IntKey is a Key implementation for int values.
type IntKey struct {
	name string
}

// Name returns the semantic identifier.
func (k IntKey) Name() string { return k.name }

// Variant returns the type constraint.
func (IntKey) Variant() Variant { return VariantInt }

// Field creates a GenericField with this key and the given value.
func (k IntKey) Field(value int) Field {
	return GenericField[int]{key: k, value: value, variant: k.Variant()}
}

// From extracts the typed int value for this key from the event.
// Returns the value and true if present, or zero value and false if not present or wrong type.
func (k IntKey) From(e *Event) (int, bool) {
	f := e.Get(k)
	if f == nil {
		return 0, false
	}
	if intf, ok := f.(GenericField[int]); ok {
		return intf.Get(), true
	}
	return 0, false
}

// NewIntKey creates an IntKey with the given name.
func NewIntKey(name string) IntKey {
	return IntKey{name: name}
}

// Float64Key is a Key implementation for float64 values.
type Float64Key struct {
	name string
}

// Name returns the semantic identifier.
func (k Float64Key) Name() string { return k.name }

// Variant returns the type constraint.
func (Float64Key) Variant() Variant { return VariantFloat64 }

// Field creates a GenericField with this key and the given value.
func (k Float64Key) Field(value float64) Field {
	return GenericField[float64]{key: k, value: value, variant: k.Variant()}
}

// From extracts the typed float64 value for this key from the event.
// Returns the value and true if present, or zero value and false if not present or wrong type.
func (k Float64Key) From(e *Event) (float64, bool) {
	f := e.Get(k)
	if f == nil {
		return 0, false
	}
	if ff, ok := f.(GenericField[float64]); ok {
		return ff.Get(), true
	}
	return 0, false
}

// NewFloat64Key creates a Float64Key with the given name.
func NewFloat64Key(name string) Float64Key {
	return Float64Key{name: name}
}

// BoolKey is a Key implementation for bool values.
type BoolKey struct {
	name string
}

// Name returns the semantic identifier.
func (k BoolKey) Name() string { return k.name }

// Variant returns the type constraint.
func (BoolKey) Variant() Variant { return VariantBool }

// Field creates a GenericField with this key and the given value.
func (k BoolKey) Field(value bool) Field {
	return GenericField[bool]{key: k, value: value, variant: k.Variant()}
}

// From extracts the typed bool value for this key from the event.
// Returns the value and true if present, or zero value and false if not present or wrong type.
func (k BoolKey) From(e *Event) (value bool, ok bool) {
	f := e.Get(k)
	if f == nil {
		return false, false
	}
	if bf, fieldOk := f.(GenericField[bool]); fieldOk {
		return bf.Get(), true
	}
	return false, false
}

// NewBoolKey creates a BoolKey with the given name.
func NewBoolKey(name string) BoolKey {
	return BoolKey{name: name}
}
