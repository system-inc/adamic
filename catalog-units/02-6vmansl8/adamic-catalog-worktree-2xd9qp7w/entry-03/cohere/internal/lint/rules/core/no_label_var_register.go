package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-label-var.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No `Decode`: upstream's `meta.schema` is `[]`, so the rule has no option surface at all.
func init() {
	rule.Register(rule.Registration{Rule: NoLabelVar})
}
