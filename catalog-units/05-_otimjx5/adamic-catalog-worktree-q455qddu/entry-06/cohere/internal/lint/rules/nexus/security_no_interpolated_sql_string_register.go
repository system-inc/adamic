package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/security-no-interpolated-sql-string.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. An option naming trusted escaping functions would be the
// name heuristic its doc comment declines, and an option widening it to lowercase keywords would
// reach the English prose it stays out of.
func init() {
	rule.Register(rule.Registration{Rule: SecurityNoInterpolatedSqlString})
}
