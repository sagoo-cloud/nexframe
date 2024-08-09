package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strconv"
)

// RuleFloat implements `float` rule:
// Float. Note that an integer is actually a float number.
//
// Format: float
type RuleFloat struct{}

func init() {
	Register(RuleFloat{})
}

func (r RuleFloat) Name() string {
	return "float"
}

func (r RuleFloat) Message() string {
	return "The {field} value `{value}` is not of valid float type"
}

func (r RuleFloat) Run(in RunInput) error {
	if _, err := strconv.ParseFloat(convert.String(in.Value), 10); err == nil {
		return nil
	}
	return errors.New(in.Message)
}
