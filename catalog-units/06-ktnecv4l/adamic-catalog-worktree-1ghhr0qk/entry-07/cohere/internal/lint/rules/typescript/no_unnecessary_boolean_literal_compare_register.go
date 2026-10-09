package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-unnecessary-boolean-literal-compare.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// `Decode` names the rule's own hand-written decoder rather than `rule.DecodeOptionsInto`. Two of
// this rule's three options default to TRUE, so the generic helper's zero-valued struct would turn
// both allowances off and the rule would report two families of comparison upstream is silent on.
// Every fixture built from a struct rather than routed through the decoder would pass anyway.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoUnnecessaryBooleanLiteralCompare,
		Decode: DecodeNoUnnecessaryBooleanLiteralCompareOptions,
	})
}
