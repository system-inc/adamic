package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-identical-branches.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. Its judgment is exact equality, and an option widening it
// (to `switch`, to empty branches, to two equal branches inside a longer chain) would reach the shapes
// the rule's doc comment declines on purpose.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoIdenticalBranches})
}
