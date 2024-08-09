package builtin

import (
	"errors"
	"regexp"
)

// RulePassword3 implements `password3` rule:
// Universal password format rule3:
// Must meet password rule1, must contain lower and upper letters, numbers and special chars.
//
// Format: password3
type RulePassword3 struct{}

func init() {
	Register(RulePassword3{})
}

func (r RulePassword3) Name() string {
	return "password3"
}

func (r RulePassword3) Message() string {
	return "The {field} value `{value}` is not a valid password3 format"
}

func (r RulePassword3) Run(in RunInput) error {
	value, ok := (*in.Value).(string)
	if !ok {
		return errors.New("password3 rule requires a string value")
	}

	pattern1 := regexp.MustCompile(`^[\w\S]{6,18}$`)
	pattern2 := regexp.MustCompile(`[a-z]+`)
	pattern3 := regexp.MustCompile(`[A-Z]+`)
	pattern4 := regexp.MustCompile(`\d+`)
	pattern5 := regexp.MustCompile(`[^a-zA-Z0-9]+`)

	if pattern1.MatchString(value) &&
		pattern2.MatchString(value) &&
		pattern3.MatchString(value) &&
		pattern4.MatchString(value) &&
		pattern5.MatchString(value) {
		return nil
	}
	return errors.New(in.Message)
}
