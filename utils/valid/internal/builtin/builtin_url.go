package builtin

import (
	"errors"
	"regexp"
)

// RuleUrl implements `url` rule:
// URL.
//
// Format: url
type RuleUrl struct{}

func init() {
	Register(RuleUrl{})
}

func (r RuleUrl) Name() string {
	return "url"
}

func (r RuleUrl) Message() string {
	return "The {field} value `{value}` is not a valid URL address"
}

func (r RuleUrl) Run(in RunInput) error {
	value, ok := (*in.Value).(string)
	if !ok {
		return errors.New("URL rule requires a string value")
	}

	pattern := regexp.MustCompile(`(https?|ftp|file)://[-A-Za-z0-9+&@#/%?=~_|!:,.;]+[-A-Za-z0-9+&@#/%=~_|]`)

	if pattern.MatchString(value) {
		return nil
	}
	return errors.New(in.Message)
}
