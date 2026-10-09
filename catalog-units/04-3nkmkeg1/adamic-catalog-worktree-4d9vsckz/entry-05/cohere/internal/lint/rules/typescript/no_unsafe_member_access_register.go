package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-unsafe-member-access.
//
// Its own file rather than a line in the package's register.go, so two ports landing at once do not
// conflict on a file neither is really changing.
//
// `Decode` names the rule's own decoder rather than `rule.DecodeOptionsInto`, so an explicitly
// configured `false` stays distinguishable from an absent key and the default lives in one place.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoUnsafeMemberAccess,
		Decode: DecodeNoUnsafeMemberAccessOptions,
	})
}
