package builtin

import (
	"errors"
	"regexp"
)

// RulePassword implements `password` rule:
// Universal password format rule1:
// Containing any visible chars, length between 6 and 18.
//
// Format: password
type RulePassword struct{}

func init() {
	Register(RulePassword{})
}

func (r RulePassword) Name() string {
	return "password"
}

func (r RulePassword) Message() string {
	return "The {field} value `{value}` is not a valid password format"
}

func (r RulePassword) Run(in RunInput) error {
	value, ok := (*in.Value).(string)
	if !ok {
		return errors.New("password rule requires a string value")
	}

	matched, _ := regexp.MatchString(`^[\w\S]{6,18}$`, value)
	if !matched {
		return errors.New(in.Message)
	}
	return nil
}
