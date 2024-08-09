package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
)

// RuleBeforeEqual implements `before-equal` rule:
// The datetime value should be after or equal to the value of field `field`.
//
// Format: before-equal:field
type RuleBeforeEqual struct{}

func init() {
	Register(RuleBeforeEqual{})
}

func (r RuleBeforeEqual) Name() string {
	return "before-equal"
}

func (r RuleBeforeEqual) Message() string {
	return "The {field} value `{value}` must be before or equal to field {pattern}"
}

func (r RuleBeforeEqual) Run(in RunInput) error {
	var (
		fieldName, fieldValue = utils.MapPossibleItemByKey(convert.Map(in.Data), in.RulePattern)
		valueDatetime         = convert.Time(in.Value)
		fieldDatetime         = convert.Time(fieldValue)
	)
	if valueDatetime.Before(fieldDatetime) || valueDatetime.Equal(fieldDatetime) {
		return nil
	}
	return errors.New(utils.ReplaceByMap(in.Message, map[string]string{
		"{field1}": fieldName,
		"{value1}": convert.String(fieldValue),
	}))
}
