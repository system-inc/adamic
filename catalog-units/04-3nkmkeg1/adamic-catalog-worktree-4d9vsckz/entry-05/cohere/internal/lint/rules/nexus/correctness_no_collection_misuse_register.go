package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-collection-misuse.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. Each of its four checks reports only what the default
// library's types decide, so there is nothing for a configuration to widen or narrow.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoCollectionMisuse})
}
