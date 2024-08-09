package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"github.com/sagoo-cloud/nexframe/utils/json"
)

// RuleArray implements `array` rule:
// Value should be type of array.
//
// Format: array
type RuleArray struct{}

func init() {
	Register(RuleArray{})
}

func (r RuleArray) Name() string {
	return "array"
}

func (r RuleArray) Message() string {
	return "The {field} value `{value}` is not of valid array type"
}

func (r RuleArray) Run(in RunInput) error {
	if utils.IsSlice(in.Value) {
		return nil
	}
	if json.Valid(convert.Bytes(in.Value)) {
		value := convert.String(in.Value)
		if len(value) > 1 && value[0] == '[' && value[len(value)-1] == ']' {
			return nil
		}
	}
	return errors.New(in.Message)
}
