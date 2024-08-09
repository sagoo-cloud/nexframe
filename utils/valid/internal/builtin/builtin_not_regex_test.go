package builtin

import (
	"strings"
	"testing"
)

func TestRuleNotRegex_Run(t *testing.T) {
	rule := RuleNotRegex{}

	tests := []struct {
		name        string
		value       interface{}
		pattern     string
		wantErr     bool
		errContains string
	}{
		{"Valid - not matching", "abc123", "^[0-9]+$", false, ""},
		{"Invalid - matching", "123", "^[0-9]+$", true, "should not be in regex"},
		{"Empty string", "", ".*", true, "should not be in regex"},
		{"Non-string input", 123, "^[0-9]+$", true, "should not be in regex"},
		{"Invalid regex", "abc", "[", true, "invalid regex pattern"},
		{"Empty pattern", "abc", "", true, "regex pattern is empty"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := RunInput{
				RuleKey:     "not-regex",
				RulePattern: tt.pattern,
				Field:       "testField",
				Value:       &tt.value,
				Message:     rule.Message(),
			}

			err := rule.Run(input)

			if (err != nil) != tt.wantErr {
				t.Errorf("RuleNotRegex.Run() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if err != nil && tt.errContains != "" && !strings.Contains(err.Error(), tt.errContains) {
				t.Errorf("RuleNotRegex.Run() error = %v, want error containing %v", err, tt.errContains)
			}
		})
	}

	// Test with nil value
	t.Run("Nil value", func(t *testing.T) {
		input := RunInput{
			RuleKey:     "not-regex",
			RulePattern: ".*",
			Field:       "testField",
			Value:       nil,
			Message:     rule.Message(),
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleNotRegex.Run() error = nil, want error for nil value")
		}
	})
}
