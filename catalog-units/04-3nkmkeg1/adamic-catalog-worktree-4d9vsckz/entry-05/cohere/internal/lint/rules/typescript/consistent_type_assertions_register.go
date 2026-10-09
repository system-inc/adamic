package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/consistent-type-assertions.
//
// The Decode entry is required rather than defensive: none of this rule's three defaults is a zero
// value, so a decoder handing back a zero struct would give an assertion style matching no branch
// and the rule would report nothing on every input.
func init() {
	rule.Register(rule.Registration{
		Rule:   ConsistentTypeAssertions,
		Decode: DecodeConsistentTypeAssertionsOptions,
	})
}
