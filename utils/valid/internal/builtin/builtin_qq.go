package builtin

import (
	"errors"
	"regexp"
)

// RuleQQ implements `qq` rule:
// Tencent QQ number.
//
// Format: qq
type RuleQQ struct{}

func init() {
	Register(RuleQQ{})
}

func (r RuleQQ) Name() string {
	return "qq"
}

func (r RuleQQ) Message() string {
	return "The {field} value `{value}` is not a valid QQ number"
}

func (r RuleQQ) Run(in RunInput) error {
	value, ok := (*in.Value).(string)
	if !ok {
		return errors.New("QQ rule requires a string value")
	}

	pattern := regexp.MustCompile(`^[1-9][0-9]{4,}$`)

	if pattern.MatchString(value) {
		return nil
	}
	return errors.New(in.Message)
}
