package builtin

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
)

// RuleEnums implements `enums` rule:
// Value should be in enums of its constant type.
//
// Format: enums
type RuleEnums struct{}

func init() {
	Register(RuleEnums{})
}

func (r RuleEnums) Name() string {
	return "enums"
}

func (r RuleEnums) Message() string {
	return "The {field} value `{value}` should be in enums of: {enums}"
}

func (r RuleEnums) Run(in RunInput) error {
	if in.Value == nil {
		return errors.New("input value is nil")
	}

	if in.ValueType == nil {
		return fmt.Errorf("value type cannot be empty to use validation rule \"enums\"")
	}

	var (
		pkgPath  = in.ValueType.PkgPath()
		typeName = in.ValueType.Name()
	)

	isSlice := false
	if in.ValueType.Kind() == reflect.Slice {
		isSlice = true
		pkgPath = in.ValueType.Elem().PkgPath()
		typeName = in.ValueType.Elem().Name()
	}

	if pkgPath == "" {
		return fmt.Errorf("no pkg path found for type \"%s\"", in.ValueType.String())
	}

	typeID := fmt.Sprintf("%s.%s", pkgPath, typeName)

	enumsStr := getEnumsByType(typeID)
	if enumsStr == "" {
		return fmt.Errorf("no enums found for type \"%s\", missing enums definition?", typeID)
	}

	var enumsValues []interface{}
	if err := json.Unmarshal([]byte(enumsStr), &enumsValues); err != nil {
		return fmt.Errorf("failed to unmarshal enums: %v", err)
	}

	// 将 *any 转换为实际值
	value := reflect.ValueOf(*in.Value)

	if isSlice {
		if value.Kind() != reflect.Slice {
			return fmt.Errorf("expected slice, got %v", value.Kind())
		}
		for i := 0; i < value.Len(); i++ {
			elemValue := fmt.Sprintf("%v", value.Index(i).Interface())
			if !containsValue(enumsValues, elemValue) {
				return fmt.Errorf(strings.Replace(in.Message, "{enums}", enumsStr, -1))
			}
		}
	} else {
		inputValue := fmt.Sprintf("%v", value.Interface())
		if !containsValue(enumsValues, inputValue) {
			return fmt.Errorf(strings.Replace(in.Message, "{enums}", enumsStr, -1))
		}
	}

	return nil
}

// containsValue checks if a value is in the slice of enum values
func containsValue(enumsValues []interface{}, value string) bool {
	for _, v := range enumsValues {
		if fmt.Sprintf("%v", v) == value {
			return true
		}
	}
	return false
}

// getEnumsByType is a function that returns the enum values for a given type
var getEnumsByType = func(typeID string) string {
	// 分割类型ID以获取包路径和类型名
	parts := strings.Split(typeID, ".")
	if len(parts) < 2 {
		return ""
	}

	// 使用反射获取类型
	var t reflect.Type
	if tt, ok := typeRegistry[typeID]; ok {
		t = tt
	} else {
		return ""
	}

	// 确保类型是一个整型
	if t.Kind() != reflect.Int {
		return ""
	}

	// 获取类型的所有字段（常量）
	var enumValues []interface{}
	for i := 0; i < t.NumField(); i++ {
		field := t.Field(i)
		if field.Type == t && field.IsExported() {
			// 假设字段的值就是枚举值
			enumValues = append(enumValues, field.Name)
		}
	}

	// 将枚举值转换为JSON字符串
	jsonBytes, err := json.Marshal(enumValues)
	if err != nil {
		return ""
	}

	return string(jsonBytes)
}

// typeRegistry 是一个用于注册枚举类型的映射
var typeRegistry = make(map[string]reflect.Type)

// RegisterEnumType 用于注册枚举类型
func RegisterEnumType(t reflect.Type) {
	typeID := fmt.Sprintf("%s.%s", t.PkgPath(), t.Name())
	typeRegistry[typeID] = t
}
