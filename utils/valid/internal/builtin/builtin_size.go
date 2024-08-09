package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"strconv"
	"strings"
)

// RuleSize implements `size` rule:
// Length must be :size.
// The length is calculated using unicode string, which means one chinese character or letter both has the length of 1.
//
// Format: size:size
type RuleSize struct{}

func init() {
	Register(RuleSize{})
}

func (r RuleSize) Name() string {
	return "size"
}

func (r RuleSize) Message() string {
	return "The {field} value `{value}` length must be {size}"
}

func (r RuleSize) Run(in RunInput) error {
	var (
		valueRunes = convert.Runes(in.Value)
		valueLen   = len(valueRunes)
	)
	size, err := strconv.Atoi(in.RulePattern)
	if valueLen != size || err != nil {
		return errors.New(strings.Replace(in.Message, "{size}", strconv.Itoa(size), -1))
	}
	return nil
}
