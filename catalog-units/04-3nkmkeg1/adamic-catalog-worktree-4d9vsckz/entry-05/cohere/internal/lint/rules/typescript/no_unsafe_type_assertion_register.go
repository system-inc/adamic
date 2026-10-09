package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-unsafe-type-assertion.
//
// Registered but deliberately NOT enabled in the live config. Measured against the ahra tree with
// the installed 8.67.0 build, this rule reports 2,846 findings across 672 files, it ships no fixer,
// and each site is a judgment about what the code actually guarantees. Turning it on is a decision
// for Kirk rather than a line a porter adds.
//
// No Decode entry: upstream's `schema: []` means the rule has no option surface at all.
func init() {
	rule.Register(rule.Registration{Rule: NoUnsafeTypeAssertion})
}
