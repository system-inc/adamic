package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/require-array-sort-compare.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// The Decode entry names the rule's OWN decoder rather than `rule.DecodeOptionsInto`, and that is
// load-bearing rather than stylistic. The rule's single option defaults to TRUE, so the generic
// helper would yield a zero-value struct on an absent key and invert the rule silently. The
// hand-rolled decoder keeps an absent key distinguishable from an explicit false.
func init() {
	rule.Register(rule.Registration{
		Rule:   RequireArraySortCompare,
		Decode: DecodeRequireArraySortCompareOptions,
	})
}
