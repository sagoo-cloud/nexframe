package builtin

import (
	"errors"
	"github.com/sagoo-cloud/nexframe/utils/convert"
)

// RuleBankCard implements `bank-card` rule:
// Bank card number.
//
// Format: bank-card
type RuleBankCard struct{}

func init() {
	Register(RuleBankCard{})
}

func (r RuleBankCard) Name() string {
	return "bank-card"
}

func (r RuleBankCard) Message() string {
	return "The {field} value `{value}` is not a valid bank card number"
}

func (r RuleBankCard) Run(in RunInput) error {
	if r.checkLuHn(convert.String(in.Value)) {
		return nil
	}
	return errors.New(in.Message)
}

// checkLuHn checks `value` with LUHN algorithm.
// It's usually used for bank card number validation.
func (r RuleBankCard) checkLuHn(value string) bool {
	var (
		sum     = 0
		nDigits = len(value)
		parity  = nDigits % 2
	)
	for i := 0; i < nDigits; i++ {
		var digit = int(value[i] - 48)
		if i%2 == parity {
			digit *= 2
			if digit > 9 {
				digit -= 9
			}
		}
		sum += digit
	}
	return sum%10 == 0
}
