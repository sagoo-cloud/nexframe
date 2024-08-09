package builtin

import (
	"errors"
	"regexp"
)

// RuleEmail implements `email` rule:
// Email address.
//
// Format: email
type RuleEmail struct{}

func init() {
	Register(RuleEmail{})
}

func (r RuleEmail) Name() string {
	return "email"
}

func (r RuleEmail) Message() string {
	return "The {field} value `{value}` is not a valid email address"
}

func (r RuleEmail) Run(in RunInput) error {
	value, ok := (*in.Value).(string)
	if !ok {
		return errors.New("email rule requires a string value")
	}

	ok, _ = regexp.MatchString(
		`^[a-zA-Z0-9_\-\.]+@[a-zA-Z0-9_\-]+(\.[a-zA-Z0-9_\-]+)+$`,
		value,
	)
	if ok {
		return nil
	}
	return errors.New(in.Message)
}
