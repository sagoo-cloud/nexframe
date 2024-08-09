package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strings"
)

// RuleRequiredUnless implements `required-unless` rule:
// Required unless all given field and its value are not equal.
//
// Format:  required-unless:field,value,...
// Example: required-unless:id,1,age,18
type RuleRequiredUnless struct{}

func init() {
	Register(RuleRequiredUnless{})
}

func (r RuleRequiredUnless) Name() string {
	return "required-unless"
}

func (r RuleRequiredUnless) Message() string {
	return "The {field} field is required"
}

func (r RuleRequiredUnless) Run(in RunInput) error {
	var (
		required   = true
		array      = strings.Split(in.RulePattern, ",")
		foundValue interface{}
		dataMap    = convert.Map(in.Data)
	)

	// It supports multiple field and value pairs.
	if len(array)%2 == 0 {
		for i := 0; i < len(array); {
			tk := array[i]
			tv := array[i+1]
			_, foundValue = utils.MapPossibleItemByKey(dataMap, tk)
			if in.Option.CaseInsensitive {
				required = !strings.EqualFold(tv, convert.String(foundValue))
			} else {
				required = strings.Compare(tv, convert.String(foundValue)) != 0
			}
			if !required {
				break
			}
			i += 2
		}
	}

	if required && isRequiredEmpty(in.Value) {
		return errors.New(in.Message)
	}
	return nil
}
