package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"github.com/sagoo-cloud/nexframe/utils/empty"
	"strings"
)

// RuleRequiredWithoutAll implements `required-without-all` rule:
// Required if all given fields are empty.
//
// Format:  required-without-all:field1,field2,...
// Example: required-without-all:id,name
type RuleRequiredWithoutAll struct{}

func init() {
	Register(RuleRequiredWithoutAll{})
}

func (r RuleRequiredWithoutAll) Name() string {
	return "required-without-all"
}

func (r RuleRequiredWithoutAll) Message() string {
	return "The {field} field is required"
}

func (r RuleRequiredWithoutAll) Run(in RunInput) error {
	var (
		required   = true
		array      = strings.Split(in.RulePattern, ",")
		foundValue interface{}
		dataMap    = convert.Map(in.Data)
	)

	for i := 0; i < len(array); i++ {
		_, foundValue = utils.MapPossibleItemByKey(dataMap, array[i])
		if !empty.IsEmpty(foundValue) {
			required = false
			break
		}
	}

	if required && isRequiredEmpty(in.Value) {
		return errors.New(in.Message)
	}
	return nil
}
