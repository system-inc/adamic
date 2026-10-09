package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers operator-assignment.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled because the option is a bare string enum whose default is `always`
// rather than the zero value.
func init() {
	rule.Register(rule.Registration{
		Rule:   OperatorAssignment,
		Decode: DecodeOperatorAssignmentOptions,
	})
}
