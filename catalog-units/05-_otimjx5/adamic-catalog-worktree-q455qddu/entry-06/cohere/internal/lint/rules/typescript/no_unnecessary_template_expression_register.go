package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-unnecessary-template-expression.
//
// Written immediately after the rule declaration compiled rather than last: between the rule file
// and this one, TestEveryRuleIsRegistered fails for the whole package and names this rule to every
// sibling agent, which is indistinguishable from an abandoned port.
//
// No Decode entry: upstream's `schema: []` means the rule has no option surface at all.
func init() {
	rule.Register(rule.Registration{Rule: NoUnnecessaryTemplateExpression})
}
