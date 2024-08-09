package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
)

// RuleBefore implements `before` rule:
// The datetime value should be after the value of field `field`.
//
// Format: before:field
type RuleBefore struct{}

func init() {
	Register(RuleBefore{})
}

func (r RuleBefore) Name() string {
	return "before"
}

func (r RuleBefore) Message() string {
	return "The {field} value `{value}` must be before field {field1} value `{value1}`"
}

func (r RuleBefore) Run(in RunInput) error {
	var (
		fieldName, fieldValue = utils.MapPossibleItemByKey(convert.Map(in.Data), in.RulePattern)
		valueDatetime         = convert.Time(in.Value)
		fieldDatetime         = convert.Time(fieldValue)
	)
	if valueDatetime.Before(fieldDatetime) {
		return nil
	}
	return errors.New(utils.ReplaceByMap(in.Message, map[string]string{
		"{field1}": fieldName,
		"{value1}": convert.String(fieldValue),
	}))
}
