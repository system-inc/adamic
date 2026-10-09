package base

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers base/boundary-no-global-container.
//
// Its own file rather than a line in a shared list, matching how every rule in this tree registers
// now: a port is three new files and no shared edit, so two landing at once conflict on nothing.
//
// No Decode entry, because the source rule's `schema` is `[]` and it reads no options.
func init() {
	rule.Register(rule.Registration{Rule: BoundaryNoGlobalContainer})
}
