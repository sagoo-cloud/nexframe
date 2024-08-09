package builtin

import (
	"reflect"
	"testing"
)

func TestRuleDateFormat_Run(t *testing.T) {
	rule := RuleDateFormat{}

	tests := []struct {
		name       string
		input      string
		dateFormat string
		wantErr    bool
	}{
		{"Valid date YYYY-MM-DD", "2023-08-06", "2006-01-02", false},
		{"Valid date DD/MM/YYYY", "06/08/2023", "02/01/2006", false},
		{"Valid date with time", "2023-08-06 15:04:05", "2006-01-02 15:04:05", false},
		{"Invalid date for format", "2023-08-06", "02/01/2006", true},
		{"Invalid date", "2023-13-32", "2006-01-02", true},
		{"Empty string", "", "2006-01-02", true},
		{"Non-date string", "hello", "2006-01-02", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := any(tt.input)
			input := RunInput{
				RuleKey:     "date-format",
				RulePattern: tt.dateFormat,
				Field:       "testField",
				ValueType:   reflect.TypeOf(tt.input),
				Value:       &value,
				Data:        nil,
				Message:     "",
				Option:      RunOption{},
			}

			err := rule.Run(input)

			if (err != nil) != tt.wantErr {
				t.Errorf("RuleDateFormat.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test with non-string input
	t.Run("Non-string input", func(t *testing.T) {
		value := any(123)
		input := RunInput{
			RuleKey:     "date-format",
			RulePattern: "2006-01-02",
			Field:       "testField",
			ValueType:   reflect.TypeOf(123),
			Value:       &value,
			Data:        nil,
			Message:     "",
			Option:      RunOption{},
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleDateFormat.Run() error = nil, want error for non-string input")
		}
	})

	// Test with nil input
	t.Run("Nil input", func(t *testing.T) {
		input := RunInput{
			RuleKey:     "date-format",
			RulePattern: "2006-01-02",
			Field:       "testField",
			ValueType:   nil,
			Value:       nil,
			Data:        nil,
			Message:     "",
			Option:      RunOption{},
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleDateFormat.Run() error = nil, want error for nil input")
		}
	})

	// Test with empty format
	t.Run("Empty format", func(t *testing.T) {
		value := any("2023-08-06")
		input := RunInput{
			RuleKey:     "date-format",
			RulePattern: "",
			Field:       "testField",
			ValueType:   reflect.TypeOf(""),
			Value:       &value,
			Data:        nil,
			Message:     "",
			Option:      RunOption{},
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleDateFormat.Run() error = nil, want error for empty format")
		}
	})
}
