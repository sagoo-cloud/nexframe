package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"github.com/sagoo-cloud/nexframe/utils/gstr"
	"strconv"
)

// RuleMin implements `min` rule:
// Equal or greater than :min. It supports both integer and float.
//
// Format: min:min
type RuleMin struct{}

func init() {
	Register(RuleMin{})
}

func (r RuleMin) Name() string {
	return "min"
}

func (r RuleMin) Message() string {
	return "The {field} value `{value}` must be equal or greater than {min}"
}

func (r RuleMin) Run(in RunInput) error {
	var (
		minL, err1   = strconv.ParseFloat(in.RulePattern, 10)
		valueN, err2 = strconv.ParseFloat(convert.String(in.Value), 10)
	)
	if valueN < minL || err1 != nil || err2 != nil {
		return errors.New(gstr.Replace(in.Message, "{min}", strconv.FormatFloat(minL, 'f', -1, 64)))
	}
	return nil
}
