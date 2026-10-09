package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-return-assign.
//
// A file of its own rather than a line in the package's `register.go`, which is the shared list
// every porter used to edit and conflict on. `rule.Register` appends into a map keyed by name and is
// callable from any file's init, so a rule is three new files and no shared edit.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoReturnAssign,
		Decode: rule.DecodeOptionsInto[NoReturnAssignOptions](),
	})
}
