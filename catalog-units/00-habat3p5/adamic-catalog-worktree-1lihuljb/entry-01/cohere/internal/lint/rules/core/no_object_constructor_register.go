package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-object-constructor.
//
// A file of its own rather than a line in the package's shared `register.go`, which is the file
// every porter used to conflict on.
func init() {
	rule.Register(rule.Registration{Rule: NoObjectConstructor})
}
