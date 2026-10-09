package react

import "github.com/system-inc/cohere/internal/lint/rule"

// Registered with a hand-written decoder rather than `rule.DecodeOptionsInto`.
//
// The generic helper errors on empty input, and a rule configured as a bare severity is handed nil.
// Answering the documented default explicitly makes the unconfigured path a decision.
func init() {
	rule.Register(rule.Registration{
		Rule:   PreferStatelessFunction,
		Decode: DecodePreferStatelessFunctionOptions,
	})
}
