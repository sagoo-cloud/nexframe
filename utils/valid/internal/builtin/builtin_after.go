package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
)

// RuleAfter implements `after` rule:
// The datetime value should be after the value of field `field`.
//
// Format: after:field
type RuleAfter struct{}

func init() {
	Register(RuleAfter{})
}

func (r RuleAfter) Name() string {
	return "after"
}

func (r RuleAfter) Message() string {
	return "The {field} value `{value}` must be after field {field1} value `{value1}`"
}

func (r RuleAfter) Run(in RunInput) error {
	var (
		fieldName, fieldValue = utils.MapPossibleItemByKey(convert.Map(in.Data), in.RulePattern)
		valueDatetime         = convert.Time(in.Value)
		fieldDatetime         = convert.Time(fieldValue)
	)
	if valueDatetime.After(fieldDatetime) {
		return nil
	}
	return errors.New(utils.ReplaceByMap(in.Message, map[string]string{
		"{field1}": fieldName,
		"{value1}": convert.String(fieldValue),
	}))
}
