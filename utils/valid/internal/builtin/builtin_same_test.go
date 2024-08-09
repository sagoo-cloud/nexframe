package builtin

import (
	"testing"
)

func TestRuleSame_Run(t *testing.T) {
	rule := RuleSame{}

	tests := []struct {
		name            string
		value           interface{}
		data            map[string]interface{}
		pattern         string
		caseInsensitive bool
		wantErr         bool
	}{
		{
			name:    "Same values",
			value:   "abc",
			data:    map[string]interface{}{"other": "abc"},
			pattern: "other",
			wantErr: false,
		},
		{
			name:    "Different values",
			value:   "abc",
			data:    map[string]interface{}{"other": "def"},
			pattern: "other",
			wantErr: true,
		},
		{
			name:            "Case insensitive - same",
			value:           "abc",
			data:            map[string]interface{}{"other": "ABC"},
			pattern:         "other",
			caseInsensitive: true,
			wantErr:         false,
		},
		{
			name:            "Case insensitive - different",
			value:           "abc",
			data:            map[string]interface{}{"other": "DEF"},
			pattern:         "other",
			caseInsensitive: true,
			wantErr:         true,
		},
		{
			name:    "Field not found",
			value:   "abc",
			data:    map[string]interface{}{"other": "abc"},
			pattern: "nonexistent",
			wantErr: true,
		},
		{
			name:    "Nil field value",
			value:   "abc",
			data:    map[string]interface{}{"other": nil},
			pattern: "other",
			wantErr: true,
		},
		{
			name:    "Non-string values - same",
			value:   123,
			data:    map[string]interface{}{"other": 123},
			pattern: "other",
			wantErr: false,
		},
		{
			name:    "Non-string values - different",
			value:   123,
			data:    map[string]interface{}{"other": 456},
			pattern: "other",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			data := any(tt.data)
			input := RunInput{
				RuleKey:     "same",
				RulePattern: tt.pattern,
				Field:       "testField",
				Value:       &tt.value,
				Data:        &data,
				Message:     rule.Message(),
				Option: RunOption{
					CaseInsensitive: tt.caseInsensitive,
				},
			}

			err := rule.Run(input)

			if (err != nil) != tt.wantErr {
				t.Errorf("RuleSame.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}

	// Test with nil value
	t.Run("Nil value", func(t *testing.T) {
		value := map[string]interface{}{"other": "abc"}
		data := any(value)
		input := RunInput{
			RuleKey:     "same",
			RulePattern: "other",
			Field:       "testField",
			Value:       nil,
			Data:        &data,
			Message:     rule.Message(),
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleSame.Run() error = nil, want error for nil value")
		}
	})

	// Test with nil data
	t.Run("Nil data", func(t *testing.T) {
		value := any("abc")
		input := RunInput{
			RuleKey:     "same",
			RulePattern: "other",
			Field:       "testField",
			Value:       &value,
			Data:        nil,
			Message:     rule.Message(),
		}

		err := rule.Run(input)

		if err == nil {
			t.Errorf("RuleSame.Run() error = nil, want error for nil data")
		}
	})
}
