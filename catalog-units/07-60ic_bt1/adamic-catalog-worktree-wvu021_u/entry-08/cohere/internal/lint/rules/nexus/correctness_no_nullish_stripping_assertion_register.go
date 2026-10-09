package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-nullish-stripping-assertion.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. Its boundary with
// `@typescript-eslint/non-nullable-type-assertion-style` is fixed by that rule's predicate, and an
// option widening or narrowing this one would let the two report the same cast.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoNullishStrippingAssertion})
}
