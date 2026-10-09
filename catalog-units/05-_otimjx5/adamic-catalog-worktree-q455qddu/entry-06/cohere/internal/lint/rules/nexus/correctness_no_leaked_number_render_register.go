package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-leaked-number-render.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. The set of types it reports is the set whose falsy value is
// a number and nothing else that renders, and an option widening it to `string | number` or
// `ReactNode` would bring back the findings the untyped rule was declined for.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoLeakedNumberRender})
}
