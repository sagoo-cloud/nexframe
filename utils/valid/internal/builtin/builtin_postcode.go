package builtin

import (
	"errors"
	"regexp"
)

// RulePostcode implements `postcode` rule:
// Postcode number.
//
// Format: postcode
type RulePostcode struct{}

func init() {
	Register(RulePostcode{})
}

func (r RulePostcode) Name() string {
	return "postcode"
}

func (r RulePostcode) Message() string {
	return "The {field} value `{value}` is not a valid postcode format"
}

func (r RulePostcode) Run(in RunInput) error {
	value, ok := (*in.Value).(string)
	if !ok {
		return errors.New("postcode rule requires a string value")
	}

	pattern := regexp.MustCompile(`^\d{6}$`)

	if pattern.MatchString(value) {
		return nil
	}
	return errors.New(in.Message)
}
