package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"github.com/sagoo-cloud/nexframe/utils/gstr"
	"strconv"
)

// RuleMinLength implements `min-length` rule:
// Length is equal or greater than :min.
// The length is calculated using unicode string, which means one chinese character or letter both has the length of 1.
//
// Format: min-length:min
type RuleMinLength struct{}

func init() {
	Register(RuleMinLength{})
}

func (r RuleMinLength) Name() string {
	return "min-length"
}

func (r RuleMinLength) Message() string {
	return "The {field} value `{value}` length must be equal or greater than {min}"
}

func (r RuleMinLength) Run(in RunInput) error {
	var (
		valueRunes = convert.Runes(in.Value)
		valueLen   = len(valueRunes)
	)
	minL, err := strconv.Atoi(in.RulePattern)
	if valueLen < minL || err != nil {
		return errors.New(gstr.Replace(in.Message, "{min}", strconv.Itoa(minL)))
	}
	return nil
}
