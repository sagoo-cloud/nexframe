package builtin

import (
	"testing"
)

func TestRuleDomain_Run(t *testing.T) {
	rule := RuleDomain{}

	tests := []struct {
		name    string
		input   interface{}
		wantErr bool
	}{
		{"Valid domain", "example.com", false},
		{"Valid subdomain", "sub.example.com", false},
		{"Valid domain with hyphen", "my-domain.com", false},
		{"Valid domain with numbers", "123domain.com", false},
		{"Invalid domain - starts with hyphen", "-domain.com", true},
		{"Invalid domain - ends with hyphen", "domain-.com", true},
		{"Invalid domain - double dot", "domain..com", true},
		{"Invalid domain - no TLD", "domain", true},
		{"Invalid domain - space", "domain .com", true},
		{"Invalid domain - special characters", "domain@.com", true},
		{"Empty string", "", true},
		{"Nil input", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			input := RunInput{
				RuleKey: "domain",
				Field:   "testField",
				Value:   &tt.input,
				Message: rule.Message(),
			}

			err := rule.Run(input)

			if (err != nil) != tt.wantErr {
				t.Errorf("RuleDomain.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test with non-string input
	t.Run("Non-string input", func(t *testing.T) {
		value := any(12345)
		input := RunInput{
			RuleKey: "domain",
			Field:   "testField",
			Value:   &value,
			Message: rule.Message(),
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleDomain.Run() error = nil, want error for non-string input")
		}
	})
}
