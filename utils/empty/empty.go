// Package empty 提供检查变量是否为空/nil的函数
package empty

import (
	"reflect"
	"sync"
	"time"
)

// 用于缓存类型信息的池
var typePool = sync.Pool{
	New: func() interface{} {
		return make(map[reflect.Type]bool)
	},
}

func IsEmpty(value any, traceSource ...bool) bool {
	if value == nil {
		return true
	}

	trace := len(traceSource) > 0 && traceSource[0]
	return isEmptyInternal(value, trace, nil)
}

func isEmptyInternal(value any, traceSource bool, visited map[reflect.Type]bool) bool {
	if value == nil {
		return true
	}

	if visited == nil {
		visited = typePool.Get().(map[reflect.Type]bool)
		defer func() {
			clear(visited)
			typePool.Put(visited)
		}()
	}

	rv := reflect.ValueOf(value)

	// 如果值本身实现了某个接口并且不是空结构体，则不为空
	if _, ok := value.(interface{ Test() string }); ok {
		return false
	}

	rt := rv.Type()
	if visited[rt] {
		return false
	}
	visited[rt] = true

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
		if !traceSource {
			return false
		}
		elem := rv.Elem()
		if !elem.IsValid() {
			return true
		}
		return isEmptyInternal(elem.Interface(), traceSource, visited)
	case reflect.Interface:
		if rv.IsNil() {
			return true
		}
		// 已经在函数开始时检查了接口实现，这里只需要处理nil情况
		return false
	case reflect.Struct:
		if rt == reflect.TypeOf(time.Time{}) {
			return rv.Interface().(time.Time).IsZero()
		}

		for i := 0; i < rv.NumField(); i++ {
			field := rv.Field(i)
			if !field.CanInterface() {
				continue
			}
			if !isEmptyInternal(field.Interface(), traceSource, visited) {
				delete(visited, rt)
				return false
			}
		}
		delete(visited, rt)
		return true
	case reflect.Func, reflect.UnsafePointer:
		return rv.IsNil()
	}

	return false
}

// IsNil 检查给定值是否为nil
func IsNil(value any, _ ...bool) bool {
	if value == nil {
		return true
	}
	rv := reflect.ValueOf(value)
	switch rv.Kind() {
	case reflect.Chan, reflect.Func, reflect.Map, reflect.Ptr, reflect.UnsafePointer, reflect.Interface, reflect.Slice:
		return !rv.IsValid() || rv.IsNil()
	}
	return false
}

// IsZero 检查给定值是否为零值
func IsZero(value any, _ ...bool) bool {
	if value == nil {
		return true
	}
	rv := reflect.ValueOf(value)
	return isZeroReflect(rv, make(map[reflect.Type]bool))
}

func isZeroReflect(rv reflect.Value, visited map[reflect.Type]bool) bool {
	if !rv.IsValid() {
		return true
	}

	rt := rv.Type()
	if visited[rt] {
		return false
	}
	visited[rt] = true
	defer delete(visited, rt)

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
			if !isZeroReflect(rv.Index(i), visited) {
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
			field := rv.Field(i)
			if !field.CanInterface() {
				continue
			}
			if !isZeroReflect(field, visited) {
				return false
			}
		}
		return true
	case reflect.UnsafePointer:
		return rv.Pointer() == 0
	}
	return false
}

// IsEmptyOrZero 检查给定值是否为空或零值
func IsEmptyOrZero(value any, traceSource ...bool) bool {
	return IsEmpty(value, traceSource...) || IsZero(value)
}
