package builtin

import (
	"errors"
	"regexp"
)

// RulePassword2 implements `password2` rule:
// Universal password format rule2:
// Must meet password rule1, must contain lower and upper letters and numbers.
//
// Format: password2
type RulePassword2 struct{}

func init() {
	Register(RulePassword2{})
}

func (r RulePassword2) Name() string {
	return "password2"
}

func (r RulePassword2) Message() string {
	return "The {field} value `{value}` is not a valid password2 format"
}

func (r RulePassword2) Run(in RunInput) error {
	value, ok := (*in.Value).(string)
	if !ok {
		return errors.New("password2 rule requires a string value")
	}

	pattern1 := regexp.MustCompile(`^[\w\S]{6,18}$`)
	pattern2 := regexp.MustCompile(`[a-z]+`)
	pattern3 := regexp.MustCompile(`[A-Z]+`)
	pattern4 := regexp.MustCompile(`\d+`)

	if pattern1.MatchString(value) &&
		pattern2.MatchString(value) &&
		pattern3.MatchString(value) &&
		pattern4.MatchString(value) {
		return nil
	}
	return errors.New(in.Message)
}
