package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/consistent-type-exports.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// `Decode` names the rule's own decoder rather than `rule.DecodeOptionsInto`. The one option
// defaults to false, so the generic helper's zero value happens to agree today; the hand-rolled
// decoder keeps an absent key distinguishable from an explicit one, which is what would matter if
// the default ever moved.
func init() {
	rule.Register(rule.Registration{
		Rule:   ConsistentTypeExports,
		Decode: DecodeConsistentTypeExportsOptions,
	})
}
