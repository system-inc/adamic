package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers require-await.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No `Decode`, because the rule takes no options: upstream's `meta.schema` is the empty array, which
// is checkable in one line and worth checking rather than inheriting from an inventory column.
func init() {
	rule.Register(rule.Registration{Rule: RequireAwait})
}
