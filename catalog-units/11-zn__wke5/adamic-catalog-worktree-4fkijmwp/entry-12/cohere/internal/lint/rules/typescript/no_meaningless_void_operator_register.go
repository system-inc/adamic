package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-meaningless-void-operator.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// The Decode entry names the rule's own decoder. This rule's single option defaults to FALSE, so
// the generic helper would happen to produce the right zero value, and the hand-rolled one is used
// anyway: relying on that coincidence leaves the next person to add a default-true key inheriting a
// decoder that silently inverts their rule, which is a real defect this package already has one
// rule guarding against.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoMeaninglessVoidOperator,
		Decode: DecodeNoMeaninglessVoidOperatorOptions,
	})
}
