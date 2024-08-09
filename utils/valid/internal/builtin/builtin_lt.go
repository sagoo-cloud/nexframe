package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strconv"
)

// RuleLT implements `lt` rule:
// Lesser than `field`.
// It supports both integer and float.
//
// Format: lt:field
type RuleLT struct{}

func init() {
	Register(RuleLT{})
}

func (r RuleLT) Name() string {
	return "lt"
}

func (r RuleLT) Message() string {
	return "The {field} value `{value}` must be lesser than field {field1} value `{value1}`"
}

func (r RuleLT) Run(in RunInput) error {
	var (
		fieldName, fieldValue = utils.MapPossibleItemByKey(convert.Map(in.Data), in.RulePattern)
		fieldValueN, err1     = strconv.ParseFloat(convert.String(fieldValue), 10)
		valueN, err2          = strconv.ParseFloat(convert.String(in.Value), 10)
	)

	if valueN >= fieldValueN || err1 != nil || err2 != nil {
		return errors.New(utils.ReplaceByMap(in.Message, map[string]string{
			"{field1}": fieldName,
			"{value1}": convert.String(fieldValue),
		}))
	}
	return nil
}
