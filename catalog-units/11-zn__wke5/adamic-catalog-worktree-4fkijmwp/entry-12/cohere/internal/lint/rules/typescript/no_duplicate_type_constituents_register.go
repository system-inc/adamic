package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-duplicate-type-constituents.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// The Decode entry names the rule's own decoder rather than the generic helper, for the reason
// recorded on the wire struct.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoDuplicateTypeConstituents,
		Decode: DecodeNoDuplicateTypeConstituentsOptions,
	})
}
