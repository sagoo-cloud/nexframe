package deepcopy

import (
	"reflect"
	"time"
)

// Interface 定义深拷贝接口
type Interface interface {
	DeepCopy() interface{}
}

// Iface 是Copy的别名，为保持向后兼容性而存在
func Iface(iface interface{}) interface{} {
	return Copy(iface)
}

// Copy 创建传入值的深度拷贝并返回interface{}
func Copy(src interface{}) interface{} {
	if src == nil {
		return nil
	}

	// 将interface转换为reflect.Value
	original := reflect.ValueOf(src)

	// 创建与原值相同类型的拷贝
	cpy := reflect.New(original.Type()).Elem()

	// 初始化visited map用于检测循环引用
	visited := make(map[uintptr]bool)

	// 递归拷贝
	copyRecursive(original, cpy, visited)

	return cpy.Interface()
}

// copyRecursive 执行实际的递归拷贝
func copyRecursive(original, cpy reflect.Value, visited map[uintptr]bool) {
	// 处理实现了deepcopy.Interface的类型
	if original.CanInterface() {
		if copier, ok := original.Interface().(Interface); ok {
			cpy.Set(reflect.ValueOf(copier.DeepCopy()))
			return
		}
	}

	switch original.Kind() {
	case reflect.Ptr:
		// 处理nil指针
		if original.IsNil() {
			return
		}

		// 获取被指向的值
		originalValue := original.Elem()

		// 检查是否是有效值
		if !originalValue.IsValid() {
			return
		}

		// 检测循环引用
		if originalValue.CanAddr() {
			ptr := originalValue.UnsafeAddr()
			if visited[ptr] {
				return
			}
			visited[ptr] = true
		}

		// 创建新的指针并设置其值
		cpy.Set(reflect.New(originalValue.Type()))
		copyRecursive(originalValue, cpy.Elem(), visited)

	case reflect.Interface:
		// 处理nil接口
		if original.IsNil() {
			return
		}

		// 获取接口内的值
		originalValue := original.Elem()

		// 创建新值并递归复制
		copyValue := reflect.New(originalValue.Type()).Elem()
		copyRecursive(originalValue, copyValue, visited)
		cpy.Set(copyValue)

	case reflect.Struct:
		// 特殊处理time.Time类型
		if t, ok := original.Interface().(time.Time); ok {
			cpy.Set(reflect.ValueOf(t))
			return
		}

		// 复制结构体的每个字段
		for i := 0; i < original.NumField(); i++ {
			// 跳过未导出字段
			if original.Type().Field(i).PkgPath != "" {
				continue
			}
			copyRecursive(original.Field(i), cpy.Field(i), visited)
		}

	case reflect.Slice:
		// 处理nil切片
		if original.IsNil() {
			return
		}

		// 创建新切片
		cpy.Set(reflect.MakeSlice(original.Type(), original.Len(), original.Cap()))

		// 复制每个元素
		for i := 0; i < original.Len(); i++ {
			copyRecursive(original.Index(i), cpy.Index(i), visited)
		}

	case reflect.Map:
		// 处理nil map
		if original.IsNil() {
			return
		}

		// 创建新map
		cpy.Set(reflect.MakeMap(original.Type()))

		// 复制每个键值对
		for _, key := range original.MapKeys() {
			// 深拷贝map的key
			copyKey := reflect.New(key.Type()).Elem()
			copyRecursive(key, copyKey, visited)

			// 深拷贝map的value
			originalValue := original.MapIndex(key)
			copyValue := reflect.New(originalValue.Type()).Elem()
			copyRecursive(originalValue, copyValue, visited)

			cpy.SetMapIndex(copyKey, copyValue)
		}

	default:
		// 对于基本类型，直接设置值
		if cpy.CanSet() {
			cpy.Set(original)
		}
	}
}
