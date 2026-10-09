package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-implicit-return.
//
// No decoder: the rule takes no options. It reports what TypeScript's noImplicitReturns would, and an
// option widening or narrowing that would make it something TypeScript does not say.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoImplicitReturn})
}
