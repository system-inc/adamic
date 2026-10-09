package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers require-atomic-updates.
//
// A file of its own rather than a line in the package's shared `register.go`. `rule.Register`
// appends into a map keyed by name and is callable from any file's init, so a rule is three new
// files and no shared edit, and two authors landing at once touch nothing in common.
func init() {
	rule.Register(rule.Registration{
		Rule:   RequireAtomicUpdates,
		Decode: rule.DecodeOptionsInto[RequireAtomicUpdatesOptions](),
	})
}
