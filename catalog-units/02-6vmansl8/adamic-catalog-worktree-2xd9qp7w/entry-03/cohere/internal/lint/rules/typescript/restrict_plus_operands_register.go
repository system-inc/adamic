package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers restrict-plus-operands.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// The Decode entry is required and the reason is sharper than "the rule takes options". Five of the
// six default to TRUE, so a rule reaching the zero-valued struct forbids everything and reports most
// of a real tree. `rule.DecodeOptionsInto` binds the *bool fields, which is what keeps an absent key
// distinguishable from an explicit false, and the settings resolver applies the defaults on nil.
func init() {
	rule.Register(rule.Registration{Rule: RestrictPlusOperands, Decode: rule.DecodeOptionsInto[RestrictPlusOperandsOptions]()})
}
