package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers block-scoped-var.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No `Decode`, because the rule takes no options at all: upstream's `meta.schema` is the empty
// array, which is the one shape that is checkable in a line and is worth checking rather than
// inheriting from the inventory column.
func init() {
	rule.Register(rule.Registration{Rule: BlockScopedVar})
}
