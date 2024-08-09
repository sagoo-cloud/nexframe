package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
)

// RuleAfterEqual implements `after-equal` rule:
// The datetime value should be after or equal to the value of field `field`.
//
// Format: after-equal:field
type RuleAfterEqual struct{}

func init() {
	Register(RuleAfterEqual{})
}

func (r RuleAfterEqual) Name() string {
	return "after-equal"
}

func (r RuleAfterEqual) Message() string {
	return "The {field} value `{value}` must be after or equal to field {field1} value `{value1}`"
}

func (r RuleAfterEqual) Run(in RunInput) error {
	var (
		fieldName, fieldValue = utils.MapPossibleItemByKey(convert.Map(in.Data), in.RulePattern)
		valueDatetime         = convert.Time(in.Value)
		fieldDatetime         = convert.Time(fieldValue)
	)
	if valueDatetime.After(fieldDatetime) || valueDatetime.Equal(fieldDatetime) {
		return nil
	}
	return errors.New(utils.ReplaceByMap(in.Message, map[string]string{
		"{field1}": fieldName,
		"{value1}": convert.String(fieldValue),
	}))
}
