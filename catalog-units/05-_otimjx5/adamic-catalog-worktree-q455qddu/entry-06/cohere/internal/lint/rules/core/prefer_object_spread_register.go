package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers prefer-object-spread.
//
// A file of its own rather than a line in the package's shared register.go, so two authors landing
// at once touch nothing in common.
func init() {
	rule.Register(rule.Registration{Rule: PreferObjectSpread})
}
