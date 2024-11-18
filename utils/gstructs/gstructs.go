// Package gstructs 提供结构体信息检索的功能
package gstructs

import (
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/sagoo-cloud/nexframe/utils/errors/gcode"
	"github.com/sagoo-cloud/nexframe/utils/errors/gerror"
)

// 全局变量，用于控制并发
var (
	activeProcesses int64         // 当前活跃的处理数
	maxProcesses    = int64(1000) // 最大并发处理数
)

// 结构体缓存
var structCache = struct {
	sync.RWMutex
	m map[reflect.Type]*Type
}{
	m: make(map[reflect.Type]*Type),
}

// Type 包装reflect.Type提供额外功能
type Type struct {
	reflect.Type
}

// Field 包含结构体字段的信息
type Field struct {
	Value    reflect.Value       // 字段的底层值
	Field    reflect.StructField // 字段的底层字段信息
	TagName  string              // 检索到的标签名
	TagValue string              // 检索到的标签值
}

// FieldsInput 定义Fields函数的输入参数
type FieldsInput struct {
	Pointer         interface{}     // 应该是结构体指针
	RecursiveOption RecursiveOption // 递归选项
}

// FieldMapInput 定义FieldMap函数的输入参数
type FieldMapInput struct {
	Pointer          interface{}
	PriorityTagArray []string
	RecursiveOption  RecursiveOption
}

// RecursiveOption 定义递归选项
type RecursiveOption int

const (
	RecursiveOptionNone RecursiveOption = iota
	RecursiveOptionEmbedded
	RecursiveOptionEmbeddedNoTag
)

// tryAcquireProcess 尝试获取处理资源
func tryAcquireProcess() bool {
	return atomic.AddInt64(&activeProcesses, 1) <= maxProcesses
}

// releaseProcess 释放处理资源
func releaseProcess() {
	atomic.AddInt64(&activeProcesses, -1)
}

// Fields 获取结构体的字段信息
func Fields(in FieldsInput) ([]Field, error) {
	if !tryAcquireProcess() {
		return nil, gerror.New("exceeded maximum concurrent processes")
	}
	defer releaseProcess()

	if in.Pointer == nil {
		return nil, gerror.New("input pointer is nil")
	}

	rangeFields, err := getFieldValues(in.Pointer)
	if err != nil {
		return nil, err
	}

	fieldFilterMap := make(map[string]struct{}, len(rangeFields))
	retrievedFields := make([]Field, 0, len(rangeFields))
	currentLevelFieldMap := make(map[string]Field, len(rangeFields))

	for _, field := range rangeFields {
		currentLevelFieldMap[field.Name()] = field
	}

	for _, field := range rangeFields {
		if !field.Value.IsValid() {
			continue
		}

		if _, ok := fieldFilterMap[field.Name()]; ok {
			continue
		}

		if field.IsEmbedded() {
			if err := processEmbeddedField(field, in.RecursiveOption, fieldFilterMap, currentLevelFieldMap, &retrievedFields); err != nil {
				return nil, err
			}
			continue
		}

		fieldFilterMap[field.Name()] = struct{}{}
		retrievedFields = append(retrievedFields, field)
	}

	return retrievedFields, nil
}

// processEmbeddedField 处理嵌入字段
func processEmbeddedField(
	field Field,
	recursiveOption RecursiveOption,
	fieldFilterMap map[string]struct{},
	currentLevelFieldMap map[string]Field,
	retrievedFields *[]Field,
) error {
	if recursiveOption == RecursiveOptionNone {
		return nil
	}

	switch recursiveOption {
	case RecursiveOptionEmbeddedNoTag:
		if field.TagStr() != "" {
			*retrievedFields = append(*retrievedFields, field)
			return nil
		}
		fallthrough

	case RecursiveOptionEmbedded:
		structFields, err := Fields(FieldsInput{
			Pointer:         field.Value,
			RecursiveOption: recursiveOption,
		})
		if err != nil {
			return err
		}

		for _, structField := range structFields {
			fieldName := structField.Name()
			if _, ok := fieldFilterMap[fieldName]; ok {
				continue
			}
			fieldFilterMap[fieldName] = struct{}{}

			if v, ok := currentLevelFieldMap[fieldName]; !ok {
				*retrievedFields = append(*retrievedFields, structField)
			} else {
				*retrievedFields = append(*retrievedFields, v)
			}
		}
	}

	return nil
}

// FieldMap 将结构体字段信息转换为map
func FieldMap(in FieldMapInput) (map[string]Field, error) {
	if !tryAcquireProcess() {
		return nil, gerror.New("exceeded maximum concurrent processes")
	}
	defer releaseProcess()

	fields, err := getFieldValues(in.Pointer)
	if err != nil {
		return nil, err
	}

	mapField := make(map[string]Field, len(fields))

	for _, field := range fields {
		if !field.IsExported() {
			continue
		}

		tagValue := ""
		for _, p := range in.PriorityTagArray {
			tagValue = field.Tag(p)
			if tagValue != "" && tagValue != "-" {
				break
			}
		}

		tempField := field
		tempField.TagValue = tagValue

		if tagValue != "" {
			mapField[tagValue] = tempField
		} else if in.RecursiveOption != RecursiveOptionNone && field.IsEmbedded() {
			if err := processEmbeddedMap(&tempField, in, mapField); err != nil {
				return nil, err
			}
		} else {
			mapField[field.Name()] = tempField
		}
	}

	return mapField, nil
}

// processEmbeddedMap 处理嵌入字段的map转换
func processEmbeddedMap(field *Field, in FieldMapInput, mapField map[string]Field) error {
	switch in.RecursiveOption {
	case RecursiveOptionEmbeddedNoTag:
		if field.TagStr() != "" {
			mapField[field.Name()] = *field
			return nil
		}
		fallthrough

	case RecursiveOptionEmbedded:
		m, err := FieldMap(FieldMapInput{
			Pointer:          field.Value,
			PriorityTagArray: in.PriorityTagArray,
			RecursiveOption:  in.RecursiveOption,
		})
		if err != nil {
			return err
		}
		for k, v := range m {
			if _, ok := mapField[k]; !ok {
				mapField[k] = v
			}
		}
	}
	return nil
}

// getFieldValues 获取结构体的字段值
func getFieldValues(structObject interface{}) ([]Field, error) {
	if structObject == nil {
		return nil, gerror.New("input struct object is nil")
	}

	var (
		reflectValue reflect.Value
		reflectKind  reflect.Kind
	)

	if v, ok := structObject.(reflect.Value); ok {
		reflectValue = v
		reflectKind = reflectValue.Kind()
	} else {
		reflectValue = reflect.ValueOf(structObject)
		reflectKind = reflectValue.Kind()
	}

	// 处理指针和数组/切片类型
	for {
		switch reflectKind {
		case reflect.Ptr:
			if !reflectValue.IsValid() || reflectValue.IsNil() {
				reflectValue = reflect.New(reflectValue.Type().Elem()).Elem()
				reflectKind = reflectValue.Kind()
			} else {
				reflectValue = reflectValue.Elem()
				reflectKind = reflectValue.Kind()
			}

		case reflect.Array, reflect.Slice:
			reflectValue = reflect.New(reflectValue.Type().Elem()).Elem()
			reflectKind = reflectValue.Kind()

		default:
			goto exitLoop
		}
	}

exitLoop:
	if reflectKind != reflect.Struct {
		return nil, gerror.NewCodef(
			gcode.CodeInvalidParameter,
			"invalid object kind: %s, struct required",
			reflectKind,
		)
	}

	var (
		structType = reflectValue.Type()
		length     = reflectValue.NumField()
		fields     = make([]Field, length)
	)

	for i := 0; i < length; i++ {
		fields[i] = Field{
			Value: reflectValue.Field(i),
			Field: structType.Field(i),
		}
	}

	return fields, nil
}

// StructType 获取结构体类型信息
func StructType(object interface{}) (*Type, error) {
	if object == nil {
		return nil, gerror.New("input object is nil")
	}

	var (
		reflectValue reflect.Value
		reflectKind  reflect.Kind
	)

	// 检查缓存
	if t, ok := object.(*Type); ok {
		return t, nil
	}

	if rv, ok := object.(reflect.Value); ok {
		reflectValue = rv
	} else {
		reflectValue = reflect.ValueOf(object)
	}
	reflectKind = reflectValue.Kind()

	// 处理指针和数组/切片类型
	for {
		switch reflectKind {
		case reflect.Ptr:
			if !reflectValue.IsValid() || reflectValue.IsNil() {
				reflectValue = reflect.New(reflectValue.Type().Elem()).Elem()
				reflectKind = reflectValue.Kind()
			} else {
				reflectValue = reflectValue.Elem()
				reflectKind = reflectValue.Kind()
			}

		case reflect.Array, reflect.Slice:
			reflectValue = reflect.New(reflectValue.Type().Elem()).Elem()
			reflectKind = reflectValue.Kind()

		default:
			goto exitLoop
		}
	}

exitLoop:
	if reflectKind != reflect.Struct {
		return nil, gerror.NewCodef(
			gcode.CodeInvalidParameter,
			"invalid object kind: %s, struct required",
			reflectKind,
		)
	}

	// 使用缓存
	structCache.RLock()
	if t, ok := structCache.m[reflectValue.Type()]; ok {
		structCache.RUnlock()
		return t, nil
	}
	structCache.RUnlock()

	t := &Type{Type: reflectValue.Type()}

	// 更新缓存
	structCache.Lock()
	structCache.m[reflectValue.Type()] = t
	structCache.Unlock()

	return t, nil
}
