package capitan

// StringKey is a Key implementation for string values.
type StringKey struct {
	name string
}

// Name returns the semantic identifier.
func (k StringKey) Name() string { return k.name }

// Variant returns the type constraint.
func (StringKey) Variant() Variant { return VariantString }

// Field creates a StringField with this key and the given value.
func (k StringKey) Field(value string) Field {
	return StringField{key: k, value: value}
}

// From extracts the typed string value for this key from the event.
// Returns the value and true if present, or zero value and false if not present or wrong type.
func (k StringKey) From(e *Event) (string, bool) {
	f := e.Get(k)
	if f == nil {
		return "", false
	}
	if sf, ok := f.(StringField); ok {
		return sf.String(), true
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

// Field creates an IntField with this key and the given value.
func (k IntKey) Field(value int) Field {
	return IntField{key: k, value: value}
}

// From extracts the typed int value for this key from the event.
// Returns the value and true if present, or zero value and false if not present or wrong type.
func (k IntKey) From(e *Event) (int, bool) {
	f := e.Get(k)
	if f == nil {
		return 0, false
	}
	if intf, ok := f.(IntField); ok {
		return intf.Int(), true
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

// Field creates a Float64Field with this key and the given value.
func (k Float64Key) Field(value float64) Field {
	return Float64Field{key: k, value: value}
}

// From extracts the typed float64 value for this key from the event.
// Returns the value and true if present, or zero value and false if not present or wrong type.
func (k Float64Key) From(e *Event) (float64, bool) {
	f := e.Get(k)
	if f == nil {
		return 0, false
	}
	if ff, ok := f.(Float64Field); ok {
		return ff.Float64(), true
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

// Field creates a BoolField with this key and the given value.
func (k BoolKey) Field(value bool) Field {
	return BoolField{key: k, value: value}
}

// From extracts the typed bool value for this key from the event.
// Returns the value and true if present, or zero value and false if not present or wrong type.
func (k BoolKey) From(e *Event) (value bool, ok bool) {
	f := e.Get(k)
	if f == nil {
		return false, false
	}
	if bf, fieldOk := f.(BoolField); fieldOk {
		return bf.Bool(), true
	}
	return false, false
}

// NewBoolKey creates a BoolKey with the given name.
func NewBoolKey(name string) BoolKey {
	return BoolKey{name: name}
}

// StringField represents a string value with semantic meaning.
type StringField struct {
	key   Key
	value string
}

func (StringField) Variant() Variant { return VariantString }
func (f StringField) Key() Key       { return f.key }
func (f StringField) Value() any     { return f.value }

// String returns the typed string value.
func (f StringField) String() string { return f.value }

// IntField represents an int value with semantic meaning.
type IntField struct {
	key   Key
	value int
}

func (IntField) Variant() Variant { return VariantInt }
func (f IntField) Key() Key       { return f.key }
func (f IntField) Value() any     { return f.value }

// Int returns the typed int value.
func (f IntField) Int() int { return f.value }

// Float64Field represents a float64 value with semantic meaning.
type Float64Field struct {
	key   Key
	value float64
}

func (Float64Field) Variant() Variant { return VariantFloat64 }
func (f Float64Field) Key() Key       { return f.key }
func (f Float64Field) Value() any     { return f.value }

// Float64 returns the typed float64 value.
func (f Float64Field) Float64() float64 { return f.value }

// BoolField represents a bool value with semantic meaning.
type BoolField struct {
	key   Key
	value bool
}

func (BoolField) Variant() Variant { return VariantBool }
func (f BoolField) Key() Key       { return f.key }
func (f BoolField) Value() any     { return f.value }

// Bool returns the typed bool value.
func (f BoolField) Bool() bool { return f.value }
