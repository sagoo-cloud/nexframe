package builtin

import (
	"reflect"
	"testing"
)

func TestRuleDate_Run(t *testing.T) {
	rule := RuleDate{}

	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"Valid date YYYY-MM-DD", "2023-08-06", false},
		{"Valid date YYYYMMDD", "20230806", false},
		{"Valid date YYYY.MM.DD", "2023.08.06", false},
		{"Invalid date format", "2023/08/06", true},
		{"Invalid date", "2023-13-32", true},
		{"Empty string", "", true},
		{"Non-date string", "hello", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := any(tt.input)
			input := RunInput{
				RuleKey:     "date",
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
				t.Errorf("RuleDate.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test with non-string input
	t.Run("Non-string input", func(t *testing.T) {
		value := any(123)
		input := RunInput{
			RuleKey:     "date",
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
			t.Errorf("RuleDate.Run() error = nil, want error for non-string input")
		}
	})

	// Test with nil input
	t.Run("Nil input", func(t *testing.T) {
		input := RunInput{
			RuleKey:     "date",
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
			t.Errorf("RuleDate.Run() error = nil, want error for nil input")
		}
	})
}
