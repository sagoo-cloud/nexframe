// Package reflection provides some reflection functions for internal usage.
package reflection

import (
	"reflect"
	"sync"
)

// OriginValueAndKindOutput holds the input and origin reflect value and kind.
type OriginValueAndKindOutput struct {
	InputValue  reflect.Value
	InputKind   reflect.Kind
	OriginValue reflect.Value
	OriginKind  reflect.Kind
}

var originValueAndKindPool = sync.Pool{
	New: func() interface{} {
		return new(OriginValueAndKindOutput)
	},
}

// OriginValueAndKind retrieves and returns the original reflect value and kind.
func OriginValueAndKind(value interface{}) OriginValueAndKindOutput {
	out := originValueAndKindPool.Get().(*OriginValueAndKindOutput)
	defer originValueAndKindPool.Put(out)

	if v, ok := value.(reflect.Value); ok {
		out.InputValue = v
	} else {
		out.InputValue = reflect.ValueOf(value)
	}
	out.InputKind = out.InputValue.Kind()
	out.OriginValue = out.InputValue
	out.OriginKind = out.InputKind
	for out.OriginKind == reflect.Ptr {
		out.OriginValue = out.OriginValue.Elem()
		out.OriginKind = out.OriginValue.Kind()
	}
	return *out
}

// OriginTypeAndKindOutput holds the input and origin reflect type and kind.
type OriginTypeAndKindOutput struct {
	InputType  reflect.Type
	InputKind  reflect.Kind
	OriginType reflect.Type
	OriginKind reflect.Kind
}

var originTypeAndKindPool = sync.Pool{
	New: func() interface{} {
		return new(OriginTypeAndKindOutput)
	},
}

// OriginTypeAndKind retrieves and returns the original reflect type and kind.
func OriginTypeAndKind(value interface{}) OriginTypeAndKindOutput {
	out := originTypeAndKindPool.Get().(*OriginTypeAndKindOutput)
	defer originTypeAndKindPool.Put(out)

	if value == nil {
		return OriginTypeAndKindOutput{}
	}
	switch v := value.(type) {
	case reflect.Type:
		out.InputType = v
	case reflect.Value:
		out.InputType = v.Type()
	default:
		out.InputType = reflect.TypeOf(value)
	}
	out.InputKind = out.InputType.Kind()
	out.OriginType = out.InputType
	out.OriginKind = out.InputKind
	for out.OriginKind == reflect.Ptr {
		out.OriginType = out.OriginType.Elem()
		out.OriginKind = out.OriginType.Kind()
	}
	return *out
}

// ValueToInterface converts reflect value to its interface type.
func ValueToInterface(v reflect.Value) (interface{}, bool) {
	if v.IsValid() && v.CanInterface() {
		return v.Interface(), true
	}
	switch v.Kind() {
	case reflect.Bool:
		return v.Bool(), true
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return v.Int(), true
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return v.Uint(), true
	case reflect.Float32, reflect.Float64:
		return v.Float(), true
	case reflect.Complex64, reflect.Complex128:
		return v.Complex(), true
	case reflect.String:
		return v.String(), true
	case reflect.Ptr, reflect.Interface:
		return ValueToInterface(v.Elem())
	default:
		return nil, false
	}
}
