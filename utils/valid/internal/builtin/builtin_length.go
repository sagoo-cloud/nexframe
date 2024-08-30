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
		minL  = 0
		maxL  = 0
		array = strings.Split(in.RulePattern, ",")
	)
	if len(array) > 0 {
		if v, err := strconv.Atoi(strings.TrimSpace(array[0])); err == nil {
			minL = v
		}
	}
	if len(array) > 1 {
		if v, err := strconv.Atoi(strings.TrimSpace(array[1])); err == nil {
			maxL = v
		}
	}
	if valueLen < minL || valueLen > maxL {
		return errors.New(utils.ReplaceByMap(in.Message, map[string]string{
			"{min}": strconv.Itoa(minL),
			"{max}": strconv.Itoa(maxL),
		}))
	}
	return nil
}
