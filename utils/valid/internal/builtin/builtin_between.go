package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strconv"
	"strings"
)

// RuleBetween implements `between` rule:
// Range between :min and :max. It supports both integer and float.
//
// Format: between:min,max
type RuleBetween struct{}

func init() {
	Register(RuleBetween{})
}

func (r RuleBetween) Name() string {
	return "between"
}

func (r RuleBetween) Message() string {
	return "The {field} value `{value}` must be between {min} and {max}"
}

func (r RuleBetween) Run(in RunInput) error {
	var (
		array  = strings.Split(in.RulePattern, ",")
		minNum = float64(0)
		maxNum = float64(0)
	)
	if len(array) > 0 {
		if v, err := strconv.ParseFloat(strings.TrimSpace(array[0]), 10); err == nil {
			minNum = v
		}
	}
	if len(array) > 1 {
		if v, err := strconv.ParseFloat(strings.TrimSpace(array[1]), 10); err == nil {
			maxNum = v
		}
	}
	valueF, err := strconv.ParseFloat(convert.String(in.Value), 10)
	if valueF < minNum || valueF > maxNum || err != nil {
		return errors.New(utils.ReplaceByMap(in.Message, map[string]string{
			"{min}": strconv.FormatFloat(minNum, 'f', -1, 64),
			"{max}": strconv.FormatFloat(maxNum, 'f', -1, 64),
		}))
	}
	return nil
}
