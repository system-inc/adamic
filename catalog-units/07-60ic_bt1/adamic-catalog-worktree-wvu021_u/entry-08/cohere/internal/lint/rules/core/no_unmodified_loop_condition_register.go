package core

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers no-unmodified-loop-condition.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors
// landing at once touch nothing in common.
//
// No `Decode`: the installed eslint 10.8.1 build's `meta.schema` is `[]`. The clone at 10.9.1 adds
// `checkConditionalExpressions`, and the rule's doc comment records why this ports the former.
func init() {
	rule.Register(rule.Registration{Rule: NoUnmodifiedLoopCondition})
}
