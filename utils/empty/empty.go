// Package empty provides functions for checking empty/nil variables.
package empty

import (
	"reflect"
	"time"
)

// IsEmpty checks whether given `value` empty.
// It returns true if `value` is in: 0, nil, false, "", len(slice/map/chan) == 0,
// or else it returns false.
func IsEmpty(value any, traceSource ...bool) bool {
	trace := len(traceSource) > 0 && traceSource[0]
	return isEmptyInternal(value, trace)
}

func isEmptyInternal(value any, traceSource bool) bool {
	if value == nil {
		return true
	}

	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return rv.Complex() == 0
	case reflect.Array, reflect.Chan, reflect.Map, reflect.Slice, reflect.String:
		return rv.Len() == 0
	case reflect.Ptr:
		if rv.IsNil() {
			return true
		}
		if traceSource {
			return isEmptyInternal(rv.Elem().Interface(), traceSource)
		}
		return false
	case reflect.Interface:
		if rv.IsNil() {
			return true
		}
		return isEmptyInternal(rv.Elem().Interface(), traceSource)
	case reflect.Struct:
		if rv.Type() == reflect.TypeOf(time.Time{}) {
			return rv.Interface().(time.Time).IsZero()
		}
		for i := 0; i < rv.NumField(); i++ {
			if !isEmptyInternal(rv.Field(i).Interface(), traceSource) {
				return false
			}
		}
		return true
	case reflect.Func, reflect.UnsafePointer:
		return rv.IsNil()
	}

	return false
}

// IsNil checks whether given `value` is nil.
func IsNil(value any, _ ...bool) bool {
	if value == nil {
		return true
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return rv.IsNil()
	}
	return false
}

// IsZero checks whether given `value` is zero value.
func IsZero(value any, _ ...bool) bool {
	rv := reflect.ValueOf(value)
	return isZeroReflect(rv)
}

func isZeroReflect(rv reflect.Value) bool {
	switch rv.Kind() {
	case reflect.Bool:
		return !rv.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return rv.Int() == 0
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return rv.Uint() == 0
	case reflect.Float32, reflect.Float64:
		return rv.Float() == 0
	case reflect.Complex64, reflect.Complex128:
		return rv.Complex() == 0
	case reflect.Array:
		for i := 0; i < rv.Len(); i++ {
			if !isZeroReflect(rv.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Ptr, reflect.Slice:
		return !rv.IsValid() || rv.IsNil()
	case reflect.String:
		return rv.Len() == 0
	case reflect.Struct:
		if rv.Type() == reflect.TypeOf(time.Time{}) {
			return rv.Interface().(time.Time).IsZero()
		}
		for i := 0; i < rv.NumField(); i++ {
			if !isZeroReflect(rv.Field(i)) {
				return false
			}
		}
		return true
	case reflect.UnsafePointer:
		return rv.Pointer() == 0
	}
	return false
}

// IsEmptyOrZero checks whether given `value` is empty or zero.
func IsEmptyOrZero(value any, traceSource ...bool) bool {
	return IsEmpty(value, traceSource...) || IsZero(value)
}
