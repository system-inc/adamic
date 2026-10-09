package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-warning-comments.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// The decoder is hand-rolled because a decoration entry must be exactly one non-whitespace
// character, which upstream states as a schema `pattern` and we have no schema layer to enforce.
func init() {
	rule.Register(rule.Registration{
		Rule:   NoWarningComments,
		Decode: DecodeNoWarningCommentsOptions,
	})
}
