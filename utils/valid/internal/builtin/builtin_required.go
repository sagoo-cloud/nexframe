package builtin

import (
	"errors"
	"reflect"
)

// RuleRequired implements `required` rule.
// Format: required
type RuleRequired struct{}

func init() {
	Register(RuleRequired{})
}

func (r RuleRequired) Name() string {
	return "required"
}

func (r RuleRequired) Message() string {
	return "The {field} field is required"
}

func (r RuleRequired) Run(in RunInput) error {
	if isRequiredEmpty(*in.Value) {
		return errors.New(in.Message)
	}
	return nil
}

// isRequiredEmpty checks and returns whether given value is empty.
// Note that if given value is a zero integer, it will be considered as not empty.
func isRequiredEmpty(value interface{}) bool {
	if value == nil {
		return true
	}
	reflectValue := reflect.ValueOf(value)
	for reflectValue.Kind() == reflect.Ptr {
		if reflectValue.IsNil() {
			return true
		}
		reflectValue = reflectValue.Elem()
	}
	switch reflectValue.Kind() {
	case reflect.String, reflect.Map, reflect.Array, reflect.Slice:
		return reflectValue.Len() == 0
	case reflect.Bool:
		return !reflectValue.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64,
		reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr,
		reflect.Float32, reflect.Float64:
		return false // Always consider numbers as non-empty
	case reflect.Interface:
		return reflectValue.IsNil()
	case reflect.Struct:
		return false // Consider all structs as non-empty
	default:
	}
	return false
}
