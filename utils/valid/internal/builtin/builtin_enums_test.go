package builtin

import (
	"reflect"
	"testing"
)

// 为测试目的定义一个枚举类型
type TestEnum int

const (
	EnumValue1 TestEnum = iota
	EnumValue2
	EnumValue3
)

func init() {
	RegisterEnumType(reflect.TypeOf(TestEnum(0)))
}
func TestRuleEnums_Run(t *testing.T) {
	// 保存原始的 getEnumsByType 函数
	originalGetEnumsByType := getEnumsByType

	// 在测试结束后恢复原始函数
	defer func() {
		getEnumsByType = originalGetEnumsByType
	}()

	// 设置 mock 函数
	getEnumsByType = func(typeID string) string {
		if typeID == "github.com/sagoo-cloud/nexframe/utils/valid/internal/builtin.TestEnum" {
			return `["0", "1", "2"]`
		}
		return ""
	}

	rule := RuleEnums{}

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"Valid enum value 0", TestEnum(0), false},
		{"Valid enum value 1", TestEnum(1), false},
		{"Valid enum value 2", TestEnum(2), false},
		{"Invalid enum value", TestEnum(3), true},
		{"Non-enum type (string)", "invalid", true},
		{"Non-enum type (int)", 42, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := RunInput{
				RuleKey:   "enums",
				Field:     "testField",
				Value:     &tt.input,
				ValueType: reflect.TypeOf(tt.input),
				Message:   rule.Message(),
			}

			err := rule.Run(input)

			if (err != nil) != tt.wantErr {
				t.Errorf("RuleEnums.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// 测试 nil 值
	t.Run("Nil value", func(t *testing.T) {
		input := RunInput{
			RuleKey:   "enums",
			Field:     "testField",
			Value:     nil,
			ValueType: reflect.TypeOf((*TestEnum)(nil)).Elem(),
			Message:   rule.Message(),
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleEnums.Run() error = nil, want error for nil value")
		}
	})

	// 测试 nil ValueType
	t.Run("Nil ValueType", func(t *testing.T) {
		value := any(EnumValue1)
		input := RunInput{
			RuleKey:   "enums",
			Field:     "testField",
			Value:     &value,
			ValueType: nil,
			Message:   rule.Message(),
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleEnums.Run() error = nil, want error for nil ValueType")
		}
	})

	// 测试未知类型
	t.Run("Unknown type", func(t *testing.T) {
		type UnknownType struct{}
		value := any(UnknownType{})
		input := RunInput{
			RuleKey:   "enums",
			Field:     "testField",
			Value:     &value,
			ValueType: reflect.TypeOf(value),
			Message:   rule.Message(),
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleEnums.Run() error = nil, want error for unknown type")
		}
	})

	t.Run("Slice_type", func(t *testing.T) {
		value := []TestEnum{TestEnum(0), TestEnum(1)}
		anyValue := any(value)
		input := RunInput{
			RuleKey:   "enums",
			Field:     "testField",
			Value:     &anyValue,
			ValueType: reflect.TypeOf(value),
			Message:   rule.Message(),
		}

		err := rule.Run(input)

		if err != nil {
			t.Errorf("RuleEnums.Run() error = %v, want nil for valid slice type", err)
		}
	})

	// 添加一个新的测试用例，用于测试包含无效值的切片
	t.Run("Slice_type_with_invalid_value", func(t *testing.T) {
		value := []TestEnum{TestEnum(0), TestEnum(3)} // TestEnum(3) 是无效的
		anyValue := any(value)
		input := RunInput{
			RuleKey:   "enums",
			Field:     "testField",
			Value:     &anyValue,
			ValueType: reflect.TypeOf(value),
			Message:   rule.Message(),
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleEnums.Run() error = nil, want error for slice with invalid value")
		}
	})
}
