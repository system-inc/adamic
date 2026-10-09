package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/consistency-require-matching-return-type.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. The principle is one spelling per kind of function, and a
// knob that let a project prefer the other spelling would be two conventions under one name.
func init() {
	rule.Register(rule.Registration{Rule: ConsistencyRequireMatchingReturnType})
}
