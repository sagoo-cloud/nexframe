package builtin

import (
	"errors"
	"regexp"
	"strconv"
	"strings"
)

// RuleResidentId implements `resident-id` rule:
// Resident id number.
//
// Format: resident-id
type RuleResidentId struct{}

func init() {
	Register(RuleResidentId{})
}

func (r RuleResidentId) Name() string {
	return "resident-id"
}

func (r RuleResidentId) Message() string {
	return "The {field} value `{value}` is not a valid resident id number"
}

func (r RuleResidentId) Run(in RunInput) error {
	if in.Value == nil {
		return errors.New("input value is nil")
	}
	value, ok := (*in.Value).(string)
	if !ok {
		return errors.New("resident-id rule requires a string value")
	}
	if r.checkResidentId(value) {
		return nil
	}
	return errors.New(in.Message)
}

// checkResidentId checks whether given id is a valid china resident id number.
func (r RuleResidentId) checkResidentId(id string) bool {
	id = strings.ToUpper(strings.TrimSpace(id))
	if len(id) != 18 {
		return false
	}

	// Weight factor and check code remain the same
	var (
		weightFactor = []int{7, 9, 10, 5, 8, 4, 2, 1, 6, 3, 7, 9, 10, 5, 8, 4, 2}
		checkCode    = []byte{'1', '0', 'X', '9', '8', '7', '6', '5', '4', '3', '2'}
		last         = id[17]
		num          = 0
	)
	for i := 0; i < 17; i++ {
		tmp, err := strconv.Atoi(string(id[i]))
		if err != nil {
			return false
		}
		num = num + tmp*weightFactor[i]
	}
	if checkCode[num%11] != last {
		return false
	}

	// Use Go's standard regexp package
	pattern := `(^[1-9]\d{5}(18|19|([23]\d))\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}[0-9Xx]$)|(^[1-9]\d{5}\d{2}((0[1-9])|(10|11|12))(([0-2][1-9])|10|20|30|31)\d{3}$)`
	matched, err := regexp.MatchString(pattern, id)
	if err != nil {
		return false
	}
	return matched
}
