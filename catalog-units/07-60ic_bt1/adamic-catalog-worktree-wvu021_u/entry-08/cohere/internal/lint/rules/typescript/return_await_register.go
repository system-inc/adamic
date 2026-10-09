package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/return-await.
//
// Written immediately after the rule declaration compiled rather than last: between the rule file
// and this one, TestEveryRuleIsRegistered fails for the whole package and names this rule to every
// sibling agent, which is indistinguishable from an abandoned port.
//
// The Decode entry is required. The option is a bare string and its default is `in-try-catch`, so a
// rule configured as a plain "error" arrives with nil options and must fall back to that rather
// than to the zero value, which would be an empty mode matching no arm.
func init() {
	rule.Register(rule.Registration{
		Rule:   ReturnAwait,
		Decode: DecodeReturnAwaitOptions,
	})
}
