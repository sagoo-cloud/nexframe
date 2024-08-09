package builtin

import (
	"testing"
)

func TestRuleRequired(t *testing.T) {
	rule := RuleRequired{}

	tests := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{"Empty string", "", true},
		{"Non-empty string", "hello", false},
		{"Zero integer", 0, false},
		{"Non-zero integer", 42, false},
		{"Zero float", 0.0, false},
		{"Non-zero float", 3.14, false},
		{"Nil pointer", (*string)(nil), true},
		{"Non-nil pointer to zero value", func() interface{} { v := 0; return &v }(), false},
		{"Empty slice", []int{}, true},
		{"Non-empty slice", []int{1, 2, 3}, false},
		{"Empty map", map[string]int{}, true},
		{"Non-empty map", map[string]int{"a": 1}, false},
		{"False boolean", false, true},
		{"True boolean", true, false},
		{"Nil interface", (interface{})(nil), true},
		{"Empty struct", struct{}{}, false},
		{"Non-empty struct", struct{ Name string }{Name: "test"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := RunInput{
				Value:   &tt.input,
				Message: "Test message",
			}
			err := rule.Run(input)
			if tt.expected && err == nil {
				t.Errorf("Expected an error for input %v, but got none", tt.input)
			} else if !tt.expected && err != nil {
				t.Errorf("Unexpected error for input %v: %v", tt.input, err)
			}
		})
	}
}
func TestRuleRequiredName(t *testing.T) {
	rule := RuleRequired{}
	if rule.Name() != "required" {
		t.Errorf("Expected rule name to be 'required', got %s", rule.Name())
	}
}

func TestRuleRequiredMessage(t *testing.T) {
	rule := RuleRequired{}
	expectedMsg := "The {field} field is required"
	if rule.Message() != expectedMsg {
		t.Errorf("Expected message to be '%s', got '%s'", expectedMsg, rule.Message())
	}
}
