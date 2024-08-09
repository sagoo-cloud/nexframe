package builtin

import (
	"errors"
	"fmt"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"regexp"
)

// RuleNotRegex implements `not-regex` rule:
// Value should not match custom regular expression pattern.
//
// Format: not-regex:pattern
type RuleNotRegex struct{}

func init() {
	Register(RuleNotRegex{})
}

func (r RuleNotRegex) Name() string {
	return "not-regex"
}

func (r RuleNotRegex) Message() string {
	return "The {field} value `{value}` should not be in regex of: {pattern}"
}

func (r RuleNotRegex) Run(in RunInput) error {
	if in.Value == nil {
		return errors.New("input value is nil")
	}

	value := convert.String(in.Value)

	if in.RulePattern == "" {
		return errors.New("regex pattern is empty")
	}

	matched, err := regexp.MatchString(in.RulePattern, value)
	if err != nil {
		return fmt.Errorf("invalid regex pattern: %v", err)
	}

	if matched {
		return errors.New(in.Message)
	}

	return nil
}
