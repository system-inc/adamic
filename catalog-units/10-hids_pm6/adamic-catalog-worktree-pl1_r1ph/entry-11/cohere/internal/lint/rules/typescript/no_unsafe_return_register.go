package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-unsafe-return.
//
// Its own file rather than a line in the package's register.go, so two ports landing at once do not
// conflict on a file neither is really changing.
//
// No Decode entry: upstream's `schema: []` means the rule has no option surface at all.
func init() {
	rule.Register(rule.Registration{Rule: NoUnsafeReturn})
}
