package builtin

import (
	"reflect"
	"testing"
)

func TestRuleDatetime_Run(t *testing.T) {
	rule := RuleDatetime{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid datetime YYYY-MM-DD HH:MM:SS", "2023-08-06 15:04:05", false},
		{"Valid datetime YYYY-MM-DDTHH:MM:SS", "2023-08-06T15:04:05", false},
		{"Valid datetime RFC3339", "2023-08-06T15:04:05Z", false},
		{"Valid datetime RFC3339 with timezone", "2023-08-06T15:04:05+08:00", false},
		{"Valid datetime with nanoseconds", "2023-08-06 15:04:05.123456789", false},
		{"Invalid datetime format", "2023/08/06 15:04:05", true},
		{"Invalid datetime", "2023-13-32 25:61:61", true},
		{"Empty string", "", true},
		{"Non-datetime string", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := any(tt.input)
			input := RunInput{
				RuleKey:     "datetime",
				RulePattern: "",
				Field:       "testField",
				ValueType:   reflect.TypeOf(tt.input),
				Value:       &value,
				Data:        nil,
				Message:     "",
				Option:      RunOption{},
			}

			err := rule.Run(input)

			if (err != nil) != tt.wantErr {
				t.Errorf("RuleDatetime.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test with non-string input
	t.Run("Non-string input", func(t *testing.T) {
		value := any(123)
		input := RunInput{
			RuleKey:     "datetime",
			RulePattern: "",
			Field:       "testField",
			ValueType:   reflect.TypeOf(123),
			Value:       &value,
			Data:        nil,
			Message:     "",
			Option:      RunOption{},
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleDatetime.Run() error = nil, want error for non-string input")
		}
	})

	// Test with nil input
	t.Run("Nil input", func(t *testing.T) {
		input := RunInput{
			RuleKey:     "datetime",
			RulePattern: "",
			Field:       "testField",
			ValueType:   nil,
			Value:       nil,
			Data:        nil,
			Message:     "",
			Option:      RunOption{},
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleDatetime.Run() error = nil, want error for nil input")
		}
	})
}
