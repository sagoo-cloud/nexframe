//数据校验处理

package valid

import (
	"fmt"
	"github.com/sagoo-cloud/nexframe/utils/rwmutex"
	"github.com/sagoo-cloud/nexframe/utils/valid/internal/builtin"
	"reflect"
	"regexp"
	"strings"
	"sync"
)

var (
	regexCache = make(map[string]*regexp.Regexp)
	regexMutex sync.RWMutex
)

func getRegex(pattern string) (*regexp.Regexp, error) {
	regexMutex.RLock()
	re, ok := regexCache[pattern]
	regexMutex.RUnlock()
	if ok {
		return re, nil
	}

	regexMutex.Lock()
	defer regexMutex.Unlock()
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, err
	}
	regexCache[pattern] = re
	return re, nil
}

func ValidateStruct(s interface{}) error {
	if s == nil {
		return fmt.Errorf("validateStruct: input is nil")
	}

	val := reflect.ValueOf(s)
	if val.Kind() == reflect.Ptr {
		if val.IsNil() {
			return fmt.Errorf("validateStruct: input is nil")
		}
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return fmt.Errorf("validateStruct: expected struct, got %T", s)
	}

	typ := val.Type()
	var allErrors []string
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		typeField := typ.Field(i)

		// 跳过 Meta 字段
		if typeField.Type.Name() == "Meta" && typeField.Type.PkgPath() == "github.com/sagoo-cloud/nexframe/nfwork" {
			continue
		}

		if v := typeField.Tag.Get("v"); v != "" {
			rules := strings.Split(v, ";")
			for _, rule := range rules {
				parts := strings.SplitN(rule, ":", 2)
				if len(parts) < 1 {
					allErrors = append(allErrors, fmt.Sprintf("Invalid rule '%s' for field '%s'", rule, typeField.Name))
					continue
				}

				ruleName := parts[0]
				rulePattern := ""
				if len(parts) > 1 {
					rulePattern = parts[1]
				} else {
					partsStr := strings.SplitN(rule, "#", 2)
					ruleName = partsStr[0]
					rulePattern = partsStr[1]
				}

				ruleInstance := builtin.GetRule(ruleName)
				if ruleInstance == nil {
					allErrors = append(allErrors, fmt.Sprintf("Unknown rule '%s' for field '%s'", ruleName, typeField.Name))
					continue
				}

				var fieldValue = field.Interface()
				runInput := builtin.RunInput{
					RuleKey:     ruleName,
					RulePattern: rulePattern,
					Field:       typeField.Name,
					ValueType:   field.Type(),
					Value:       &fieldValue,
					Data:        &s,
					Message:     ruleInstance.Message(),
					Option:      builtin.RunOption{},
				}

				if err := ruleInstance.Run(runInput); err != nil {
					allErrors = append(allErrors, err.Error())
				}
			}
		}
	}

	if len(allErrors) > 0 {
		return fmt.Errorf("%s", strings.Join(allErrors, "; "))
	}
	return nil
}

type StrSet struct {
	mu   rwmutex.RWMutex
	data map[string]struct{}
}

// NewStrSet create and returns a new set, which contains un-repeated items.
// The parameter `safe` is used to specify whether using set in concurrent-safety,
// which is false in default.
func NewStrSet(safe ...bool) *StrSet {
	return &StrSet{
		mu:   rwmutex.Create(safe...),
		data: make(map[string]struct{}),
	}
}

// AddIfNotExist checks whether item exists in the set,
// it adds the item to set and returns true if it does not exist in the set,
// or else it does nothing and returns false.
func (set *StrSet) AddIfNotExist(item string) bool {
	if !set.Contains(item) {
		set.mu.Lock()
		defer set.mu.Unlock()
		if set.data == nil {
			set.data = make(map[string]struct{})
		}
		if _, ok := set.data[item]; !ok {
			set.data[item] = struct{}{}
			return true
		}
	}
	return false
}

// Contains checks whether the set contains `item`.
func (set *StrSet) Contains(item string) bool {
	var ok bool
	set.mu.RLock()
	if set.data != nil {
		_, ok = set.data[item]
	}
	set.mu.RUnlock()
	return ok
}
