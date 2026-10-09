package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/consistency-no-iso-string-date-cut.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. An option exempting files or helpers would bring back the
// implicit zone the rule exists to make explicit, and the one helper that defines the shape,
// `dateIso8601`, is fixed rather than exempted.
func init() {
	rule.Register(rule.Registration{Rule: ConsistencyNoIsoStringDateCut})
}
