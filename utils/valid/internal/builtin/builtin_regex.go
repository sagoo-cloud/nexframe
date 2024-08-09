package builtin

import (
	"errors"
	"fmt"
	"regexp"
)

// RuleRegex implements `regex` rule:
// Value should match custom regular expression pattern.
//
// Format: regex:pattern
type RuleRegex struct{}

func init() {
	Register(RuleRegex{})
}

func (r RuleRegex) Name() string {
	return "regex"
}

func (r RuleRegex) Message() string {
	return "The {field} value `{value}` must be in regex of: {pattern}"
}

func (r RuleRegex) Run(in RunInput) error {
	value, ok := (*in.Value).(string)
	if !ok {
		return errors.New("regex rule requires a string value")
	}

	pattern, err := regexp.Compile(in.RulePattern)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %v", err)
	}

	if !pattern.MatchString(value) {
		return errors.New(in.Message)
	}
	return nil
}
