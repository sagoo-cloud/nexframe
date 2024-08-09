package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/errors/gcode"
	"github.com/sagoo-cloud/nexframe/errors/gerror"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strings"
)

// RuleRequiredIf implements `required-if` rule:
// Required if any of given field and its value are equal.
//
// Format:  required-if:field,value,...
// Example: required-if:id,1,age,18
type RuleRequiredIf struct{}

func init() {
	Register(RuleRequiredIf{})
}

func (r RuleRequiredIf) Name() string {
	return "required-if"
}

func (r RuleRequiredIf) Message() string {
	return "The {field} field is required"
}

func (r RuleRequiredIf) Run(in RunInput) error {
	var (
		required   = false
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
	// It supports multiple field and value pairs.
	for i := 0; i < len(array); {
		var (
			tk = array[i]
			tv = array[i+1]
		)
		_, foundValue = utils.MapPossibleItemByKey(dataMap, tk)
		if in.Option.CaseInsensitive {
			required = strings.EqualFold(tv, convert.String(foundValue))
		} else {
			required = strings.Compare(tv, convert.String(foundValue)) == 0
		}
		if required {
			break
		}
		i += 2
	}
	if required && isRequiredEmpty(in.Value) {
		return errors.New(in.Message)
	}
	return nil
}
