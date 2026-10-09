package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/consistency-no-property-alias.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common. The rule moved here from the structure package (#kjnmdb1), because
// Reach Over Alias is language doctrine rather than React, and Base loads Nexus' ESLint plugin and not
// Structure's.
//
// No decoder: the rule takes no options.
func init() {
	rule.Register(rule.Registration{Rule: ConsistencyNoPropertyAlias})
}
