package builtin

import (
	"errors"
	"regexp"
)

// RuleMac implements `mac` rule:
// MAC.
//
// Format: mac
type RuleMac struct{}

func init() {
	Register(RuleMac{})
}

func (r RuleMac) Name() string {
	return "mac"
}

func (r RuleMac) Message() string {
	return "The {field} value `{value}` is not a valid MAC address"
}

func (r RuleMac) Run(in RunInput) error {
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
