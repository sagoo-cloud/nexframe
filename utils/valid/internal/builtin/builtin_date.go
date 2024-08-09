package builtin

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

// RuleDate implements `date` rule:
// Standard date, like: 2006-01-02, 20060102, 2006.01.02.
//
// Format: date
type RuleDate struct{}

func init() {
	Register(RuleDate{})
}

func (r RuleDate) Name() string {
	return "date"
}

func (r RuleDate) Message() string {
	return "The {field} value `{value}` is not a valid date"
}

func (r RuleDate) Run(in RunInput) error {
	if in.Value == nil {
		return errors.New("input value is nil")
	}

	value := reflect.ValueOf(*in.Value)
	if value.Kind() != reflect.String {
		return fmt.Errorf("date rule only applies to string values, got %v", value.Kind())
	}

	dateStr := value.String()

	// 定义支持的日期格式
	formats := []string{
		"2006-01-02",
		"20060102",
		"2006.01.02",
	}

	var parseErr error
	for _, format := range formats {
		_, err := time.Parse(format, dateStr)
		if err == nil {
			// 如果能成功解析，说明是有效的日期格式
			return nil
		}
		parseErr = err
	}

	// 如果所有格式都无法解析，返回错误
	return fmt.Errorf("invalid date format: %v", parseErr)
}
