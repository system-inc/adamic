package react

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers this package's `display-name` rule.
//
// A file of its own rather than a line in `register.go`, so a port is three owned files and no
// shared edit. Written second rather than last, because the gap between the rule file landing and
// its registration makes `TestEveryRuleIsRegistered` name this rule to every other agent in the
// tree.
func init() {
	rule.Register(rule.Registration{Rule: DisplayName, Decode: DecodeDisplayNameOptions})
}
