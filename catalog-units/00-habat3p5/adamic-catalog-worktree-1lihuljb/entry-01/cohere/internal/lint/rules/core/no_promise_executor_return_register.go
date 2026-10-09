package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-promise-executor-return.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common. Written immediately after the rule variable compiled rather than
// last: in the gap between the two, `TestEveryRuleIsRegistered` fails for the whole package and
// names this rule to every other agent, which is indistinguishable from an abandoned port.
//
// The decoder is hand-rolled rather than `rule.DecodeOptionsInto`: the generic helper errors on
// empty input and the config layer turns that error into nil, which the rule then reads as a zero
// value. That is the correct answer for this rule's one option today, because `allowVoid` defaults
// to false, and the explicit path is written anyway so a later default change cannot invert the rule
// without anything saying so.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoPromiseExecutorReturn,
		Decode: DecodeNoPromiseExecutorReturnOptions,
	})
}
