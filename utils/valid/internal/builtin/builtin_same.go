package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strings"
)

// RuleSame implements `same` rule:
// Value should be the same as value of field.
//
// Format: same:field
type RuleSame struct{}

func init() {
	Register(RuleSame{})
}

func (r RuleSame) Name() string {
	return "same"
}

func (r RuleSame) Message() string {
	return "The {field} value `{value}` must be the same as field {field1} value `{value1}`"
}

func (r RuleSame) Run(in RunInput) error {
	if in.Value == nil {
		return errors.New("input value is nil")
	}

	if in.Data == nil {
		return errors.New("input data is nil")
	}

	value := convert.String(in.Value)
	data := convert.Map(*in.Data)

	fieldName, fieldValue := utils.MapPossibleItemByKey(data, in.RulePattern)
	if fieldValue == nil {
		return errors.New("field not found or field value is nil")
	}

	fieldValueStr := convert.String(fieldValue)

	var ok bool
	if in.Option.CaseInsensitive {
		ok = strings.EqualFold(value, fieldValueStr)
	} else {
		ok = value == fieldValueStr
	}

	if !ok {
		return errors.New(utils.ReplaceByMap(in.Message, map[string]string{
			"{field1}": fieldName,
			"{value1}": fieldValueStr,
		}))
	}
	return nil
}
