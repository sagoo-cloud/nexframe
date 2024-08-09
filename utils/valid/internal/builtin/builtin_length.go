package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strconv"
	"strings"
)

// RuleLength implements `length` rule:
// Length between :min and :max.
// The length is calculated using unicode string, which means one chinese character or letter both has the length of 1.
//
// Format: length:min,max
type RuleLength struct{}

func init() {
	Register(RuleLength{})
}

func (r RuleLength) Name() string {
	return "length"
}

func (r RuleLength) Message() string {
	return "The {field} value `{value}` length must be between {min} and {max}"
}

func (r RuleLength) Run(in RunInput) error {
	var (
		valueRunes = convert.Runes(in.Value)
		valueLen   = len(valueRunes)
	)
	var (
		min   = 0
		max   = 0
		array = strings.Split(in.RulePattern, ",")
	)
	if len(array) > 0 {
		if v, err := strconv.Atoi(strings.TrimSpace(array[0])); err == nil {
			min = v
		}
	}
	if len(array) > 1 {
		if v, err := strconv.Atoi(strings.TrimSpace(array[1])); err == nil {
			max = v
		}
	}
	if valueLen < min || valueLen > max {
		return errors.New(utils.ReplaceByMap(in.Message, map[string]string{
			"{min}": strconv.Itoa(min),
			"{max}": strconv.Itoa(max),
		}))
	}
	return nil
}
