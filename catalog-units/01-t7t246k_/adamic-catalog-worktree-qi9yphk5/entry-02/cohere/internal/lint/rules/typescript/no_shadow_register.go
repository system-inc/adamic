package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-shadow.
//
// Its own file rather than a line in the package's register.go, so two ports landing at once do not
// conflict on a file neither is really changing.
//
// The decoder is the rule's own rather than `rule.DecodeOptionsInto`, because two of its options
// default to true and the generic helper cannot tell an absent key from an explicit false.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoShadow,
		Decode: DecodeNoShadowOptions,
	})
}
