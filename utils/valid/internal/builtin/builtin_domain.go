package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"regexp"
)

// RuleDomain implements `domain` rule:
// Domain.
//
// Format: domain
type RuleDomain struct{}

func init() {
	Register(RuleDomain{})
}

func (r RuleDomain) Name() string {
	return "domain"
}

func (r RuleDomain) Message() string {
	return "The {field} value `{value}` is not a valid domain format"
}

func (r RuleDomain) Run(in RunInput) error {
	if in.Value == nil {
		return errors.New("input value is nil")
	}

	// 修改正则表达式以确保顶级域名不以连字符结尾
	pattern := `^([a-zA-Z0-9]([a-zA-Z0-9\-]{0,61}[a-zA-Z0-9])?\.)+[a-zA-Z]{2,63}$`
	matched, err := regexp.MatchString(pattern, convert.String(in.Value))
	if err != nil {
		return err
	}
	if matched {
		return nil
	}
	return errors.New(in.Message)
}
