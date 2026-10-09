package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers only-throw-error.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// The Decode entry is not optional here. This rule's three booleans all default to TRUE, and a
// rule registered without a decoder is handed nil options, so the defaulting in Run is what keeps
// an unconfigured rule permissive rather than reporting every `any` in the tree.
func init() {
	rule.Register(rule.Registration{Rule: OnlyThrowError, Decode: DecodeOnlyThrowErrorOptions})
}
