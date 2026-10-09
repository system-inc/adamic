package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-empty-function.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common. Written immediately after the rule variable compiled rather than
// last: in the gap between the two, `TestEveryRuleIsRegistered` fails for the whole package and
// names this rule to every sibling agent, which is indistinguishable from an abandoned port.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoEmptyFunction,
		Decode: DecodeNoEmptyFunctionOptions,
	})
}
