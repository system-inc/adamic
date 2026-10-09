package typescript

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers default-param-last.
//
// Its own file rather than a line in the package's register.go: that shared list is one file every
// concurrent port has to edit, so two ports landing at once conflict on a file neither is really
// changing. rule.Register appends into a map keyed by name and is callable from any init, so a
// package may hold as many of these as it has rules, and a port becomes three new files and no
// shared edit.
func init() {
	rule.Register(rule.Registration{Rule: DefaultParamLast})
}
