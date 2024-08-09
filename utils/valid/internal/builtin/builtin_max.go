package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"github.com/sagoo-cloud/nexframe/utils/gstr"
	"strconv"
)

// RuleMax implements `max` rule:
// Equal or lesser than :max. It supports both integer and float.
//
// Format: max:max
type RuleMax struct{}

func init() {
	Register(RuleMax{})
}

func (r RuleMax) Name() string {
	return "max"
}

func (r RuleMax) Message() string {
	return "The {field} value `{value}` must be equal or lesser than {max}"
}

func (r RuleMax) Run(in RunInput) error {
	var (
		max, err1    = strconv.ParseFloat(in.RulePattern, 10)
		valueN, err2 = strconv.ParseFloat(convert.String(in.Value), 10)
	)
	if valueN > max || err1 != nil || err2 != nil {
		return errors.New(gstr.Replace(in.Message, "{max}", strconv.FormatFloat(max, 'f', -1, 64)))
	}
	return nil
}
