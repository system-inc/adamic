package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/security-no-interpolated-shell-command.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options. Its judgment is the type of every piece of a shell command,
// and an option widening it (to opaque commands, to an explicit `sh -c`) would reach the shapes the
// rule's doc comment declines on purpose.
func init() {
	rule.Register(rule.Registration{Rule: SecurityNoInterpolatedShellCommand})
}
