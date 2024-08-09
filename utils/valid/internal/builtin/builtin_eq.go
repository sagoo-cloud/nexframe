package builtin

// RuleEq implements `eq` rule:
// Value should be the same as value of field.
//
// This rule performs the same as rule `same`.
//
// Format: eq:field
type RuleEq struct{}

func init() {
	Register(RuleEq{})
}

func (r RuleEq) Name() string {
	return "eq"
}

func (r RuleEq) Message() string {
	return "The {field} value `{value}` must be equal to field {field1} value `{value1}`"
}

func (r RuleEq) Run(in RunInput) error {
	return RuleSame{}.Run(in)
}
