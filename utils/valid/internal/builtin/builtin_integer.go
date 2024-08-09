package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strconv"
)

// RuleInteger implements `integer` rule:
// Integer.
//
// Format: integer
type RuleInteger struct{}

func init() {
	Register(RuleInteger{})
}

func (r RuleInteger) Name() string {
	return "integer"
}

func (r RuleInteger) Message() string {
	return "The {field} value `{value}` is not an integer"
}

func (r RuleInteger) Run(in RunInput) error {
	if _, err := strconv.Atoi(convert.String(in.Value)); err == nil {
		return nil
	}
	return errors.New(in.Message)
}
