package builtin

import (
	"errors"
	"regexp"
)

// RuleTelephone implements `telephone` rule:
// "XXXX-XXXXXXX"
// "XXXX-XXXXXXXX"
// "XXX-XXXXXXX"
// "XXX-XXXXXXXX"
// "XXXXXXX"
// "XXXXXXXX"
//
// Format: telephone
type RuleTelephone struct{}

func init() {
	Register(RuleTelephone{})
}

func (r RuleTelephone) Name() string {
	return "telephone"
}

func (r RuleTelephone) Message() string {
	return "The {field} value `{value}` is not a valid telephone number"
}

func (r RuleTelephone) Run(in RunInput) error {
	value, ok := (*in.Value).(string)
	if !ok {
		return errors.New("telephone rule requires a string value")
	}

	pattern := regexp.MustCompile(`^((\d{3,4})|\d{3,4}-)?\d{7,8}$`)

	if pattern.MatchString(value) {
		return nil
	}
	return errors.New(in.Message)
}
