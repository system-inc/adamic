package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-loop-func.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// No `Decode`: upstream's `meta.schema` is `[]`, so the rule takes no options.
func init() {
	rule.Register(rule.Registration{Rule: NoLoopFunc})
}
