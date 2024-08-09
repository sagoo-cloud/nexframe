package builtin

import (
	"errors"
	"fmt"
	"github.com/sagoo-cloud/nexframe/utils/convert"
	"net"
)

// RuleIp implements `ip` rule:
// IPv4/IPv6.
//
// Format: ip
type RuleIp struct{}

func init() {
	Register(RuleIp{})
}

func (r RuleIp) Name() string {
	return "ip"
}

func (r RuleIp) Message() string {
	return "The {field} value `{value}` is not a valid IP address"
}

func (r RuleIp) Run(in RunInput) error {
	if in.Value == nil {
		return errors.New("input value is nil")
	}

	value := convert.String(in.Value)
	if value == "" {
		return fmt.Errorf("the %s value is empty", in.Field)
	}

	ip := net.ParseIP(value)
	if ip == nil {
		return fmt.Errorf(in.Message)
	}

	return nil
}
