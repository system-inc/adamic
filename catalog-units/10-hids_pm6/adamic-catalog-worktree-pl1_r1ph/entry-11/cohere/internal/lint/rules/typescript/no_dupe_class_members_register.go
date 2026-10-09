package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers @typescript-eslint/no-dupe-class-members.
//
// Its own file rather than a line in the package's register.go, so two ports landing at once do not
// conflict on a file neither is really changing.
//
// No Decode entry: the base rule's schema is empty, so the extension has no option surface either.
func init() {
	rule.Register(rule.Registration{Rule: NoDupeClassMembers})
}
