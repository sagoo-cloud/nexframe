package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strings"
)

// RuleDifferent implements `different` rule:
// Value should be different from value of field.
//
// Format: different:field
type RuleDifferent struct{}

func init() {
	Register(RuleDifferent{})
}

func (r RuleDifferent) Name() string {
	return "different"
}

func (r RuleDifferent) Message() string {
	return "The {field} value `{value}` must be different from field {field1} value `{value1}`"
}

func (r RuleDifferent) Run(in RunInput) error {
	if in.Value == nil {
		return errors.New("input value is nil")
	}

	value := convert.String(*in.Value)
	data := convert.Map(*in.Data)
	fieldName, fieldValue := utils.MapPossibleItemByKey(data, in.RulePattern)

	if fieldValue == nil {
		return nil // 如果找不到比较的字段，则认为是不同的
	}

	fieldValueStr := convert.String(fieldValue)

	var isDifferent bool
	if in.Option.CaseInsensitive {
		isDifferent = !strings.EqualFold(value, fieldValueStr)
	} else {
		isDifferent = value != fieldValueStr
	}

	if isDifferent {
		return nil
	}

	return errors.New(utils.ReplaceByMap(in.Message, map[string]string{
		"{field1}": fieldName,
		"{value1}": fieldValueStr,
	}))
}
