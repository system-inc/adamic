package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers use-unknown-in-catch-callback-variable.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing.
//
// No Decode entry, because upstream declares `schema: []` and the rule reads no configuration. That
// was read off the installed build's metadata as well as the clone, and the two agree.
func init() {
	rule.Register(rule.Registration{Rule: UseUnknownInCatchCallbackVariable})
}
