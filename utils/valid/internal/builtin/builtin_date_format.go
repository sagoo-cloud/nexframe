package builtin

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

// RuleDateFormat implements `date-format` rule:
// Custom date format.
//
// Format: date-format:format
type RuleDateFormat struct{}

func init() {
	Register(RuleDateFormat{})
}

func (r RuleDateFormat) Name() string {
	return "date-format"
}

func (r RuleDateFormat) Message() string {
	return "The {field} value `{value}` does not match the format: {pattern}"
}

func (r RuleDateFormat) Run(in RunInput) error {
	if in.Value == nil {
		return errors.New("input value is nil")
	}

	value := reflect.ValueOf(*in.Value)
	if value.Kind() != reflect.String {
		return fmt.Errorf("date-format rule only applies to string values, got %v", value.Kind())
	}

	dateStr := value.String()

	if in.RulePattern == "" {
		return errors.New("date format pattern is required")
	}

	// 使用提供的格式解析日期
	_, err := time.Parse(in.RulePattern, dateStr)
	if err != nil {
		return fmt.Errorf("failed to parse date '%s' with format '%s': %v", dateStr, in.RulePattern, err)
	}

	return nil
}
