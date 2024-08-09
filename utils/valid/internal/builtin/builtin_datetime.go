package builtin

import (
	"errors"
	"fmt"
	"reflect"
	"time"
)

// RuleDatetime implements `datetime` rule:
// Standard datetime, like: 2006-01-02 12:00:00.
//
// Format: datetime
type RuleDatetime struct{}

func init() {
	Register(RuleDatetime{})
}

func (r RuleDatetime) Name() string {
	return "datetime"
}

func (r RuleDatetime) Message() string {
	return "The {field} value `{value}` is not a valid datetime"
}

func (r RuleDatetime) Run(in RunInput) error {
	if in.Value == nil {
		return errors.New("input value is nil")
	}

	value := reflect.ValueOf(*in.Value)
	if value.Kind() != reflect.String {
		return fmt.Errorf("datetime rule only applies to string values, got %v", value.Kind())
	}

	datetimeStr := value.String()

	// 定义支持的日期时间格式
	formats := []string{
		"2006-01-02 15:04:05",
		"2006-01-02T15:04:05",
		"2006-01-02 15:04:05.999999999 -0700 MST",
		time.RFC3339,
		time.RFC3339Nano,
	}

	var parseErr error
	for _, format := range formats {
		_, err := time.Parse(format, datetimeStr)
		if err == nil {
			// 如果能成功解析，说明是有效的日期时间格式
			return nil
		}
		parseErr = err
	}

	// 如果所有格式都无法解析，返回错误
	return fmt.Errorf("invalid datetime format: %v", parseErr)
}
