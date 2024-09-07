package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"github.com/sagoo-cloud/nexframe/utils/errors/gcode"
	"github.com/sagoo-cloud/nexframe/utils/errors/gerror"
	"strings"
)

// RuleRequiredIfAll implements `required-if-all` rule:
// Required if all given field and its value are equal.
//
// Format:  required-if-all:field,value,...
// Example: required-if-all:id,1,age,18
type RuleRequiredIfAll struct{}

func init() {
	Register(RuleRequiredIfAll{})
}

func (r RuleRequiredIfAll) Name() string {
	return "required-if-all"
}

func (r RuleRequiredIfAll) Message() string {
	return "The {field} field is required"
}

func (r RuleRequiredIfAll) Run(in RunInput) error {
	var (
		required   = true
		array      = strings.Split(in.RulePattern, ",")
		foundValue interface{}
		dataMap    = convert.Map(in.Data)
	)
	if len(array)%2 != 0 {
		return gerror.NewCodef(
			gcode.CodeInvalidParameter,
			`invalid "%s" rule pattern: %s`,
			r.Name(),
			in.RulePattern,
		)
	}
	for i := 0; i < len(array); {
		var (
			tk = array[i]
			tv = array[i+1]
			eq bool
		)
		_, foundValue = utils.MapPossibleItemByKey(dataMap, tk)
		if in.Option.CaseInsensitive {
			eq = strings.EqualFold(tv, convert.String(foundValue))
		} else {
			eq = strings.Compare(tv, convert.String(foundValue)) == 0
		}
		if !eq {
			required = false
			break
		}
		i += 2
	}
	if required && isRequiredEmpty(in.Value) {
		return errors.New(in.Message)
	}
	return nil
}
