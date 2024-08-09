package builtin

import "testing"

func TestRuleDifferent_Run(t *testing.T) {
	rule := RuleDifferent{}

	tests := []struct {
		name            string
		value           interface{}
		data            map[string]interface{}
		pattern         string
		caseInsensitive bool
		wantErr         bool
	}{
		{
			name:    "Different values",
			value:   "abc",
			data:    map[string]interface{}{"other": "def"},
			pattern: "other",
			wantErr: false,
		},
		{
			name:    "Same values",
			value:   "abc",
			data:    map[string]interface{}{"other": "abc"},
			pattern: "other",
			wantErr: true,
		},
		{
			name:            "Case insensitive - different",
			value:           "abc",
			data:            map[string]interface{}{"other": "ABC"},
			pattern:         "other",
			caseInsensitive: true,
			wantErr:         true,
		},
		{
			name:            "Case insensitive - same",
			value:           "abc",
			data:            map[string]interface{}{"other": "def"},
			pattern:         "other",
			caseInsensitive: true,
			wantErr:         false,
		},
		{
			name:    "Field not found",
			value:   "abc",
			data:    map[string]interface{}{"other": "def"},
			pattern: "nonexistent",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			value := tt.value
			data := any(tt.data)
			input := RunInput{
				RuleKey:     "different",
				RulePattern: tt.pattern,
				Field:       "testField",
				Value:       &value,
				Data:        &data,
				Message:     rule.Message(),
				Option: RunOption{
					CaseInsensitive: tt.caseInsensitive,
				},
			}

			err := rule.Run(input)

			if (err != nil) != tt.wantErr {
				t.Errorf("RuleDifferent.Run() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
