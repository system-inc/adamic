package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/consistent-indexed-object-style.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// The Decode entry names the rule's own decoder, which cannot be the generic helper: upstream's
// option is a bare enum string rather than an object, so there is no struct to decode into.
func init() {
	rule.Register(rule.Registration{
		Rule:   ConsistentIndexedObjectStyle,
		Decode: DecodeConsistentIndexedObjectStyleOptions,
	})
}
