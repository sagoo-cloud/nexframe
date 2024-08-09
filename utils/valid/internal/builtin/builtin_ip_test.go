package builtin

import (
	"testing"
)

func TestRuleIp_Run(t *testing.T) {
	rule := RuleIp{}

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"Valid IPv4", "192.168.1.1", false},
		{"Valid IPv6", "2001:0db8:85a3:0000:0000:8a2e:0370:7334", false},
		{"Valid IPv6 shortened", "2001:db8:85a3::8a2e:370:7334", false},
		{"Invalid IP", "256.1.2.3", true},
		{"Invalid format", "192.168.1", true},
		{"Empty string", "", true},
		{"Non-IP string", "hello world", true},
		{"Number", 12345, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := RunInput{
				RuleKey: "ip",
				Field:   "testField",
				Value:   &tt.input,
				Message: rule.Message(),
			}

			err := rule.Run(input)

			if (err != nil) != tt.wantErr {
				t.Errorf("RuleIp.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test with nil value
	t.Run("Nil value", func(t *testing.T) {
		input := RunInput{
			RuleKey: "ip",
			Field:   "testField",
			Value:   nil,
			Message: rule.Message(),
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleIp.Run() error = nil, want error for nil value")
		}
	})
}
