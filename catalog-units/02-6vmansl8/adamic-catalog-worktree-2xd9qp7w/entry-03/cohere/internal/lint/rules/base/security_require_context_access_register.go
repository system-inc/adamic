package base

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers base/security-require-context-access.
//
// A file of its own rather than a line in the package's shared register, so two authors landing at
// once touch nothing in common.
//
// The decoder is hand-rolled to enforce the two constraints the source states in its JSON schema and
// we have no schema layer for: both fields are required, and `requiresAny` carries `minItems: 1`. An
// empty `requiresAny` would be a requirement nothing could satisfy, reporting on every injection of
// that key with a message naming no decorator to add.
func init() {
	rule.Register(rule.Registration{
		Rule:   SecurityRequireContextAccess,
		Decode: DecodeContextRequiresAccessOptions,
	})
}
