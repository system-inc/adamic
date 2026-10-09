package nexus

import "github.com/system-inc/cohere/internal/lint/rule"

// init registers nexus/correctness-no-uncleared-race-timeout.
//
// A file of its own rather than a line in the package's shared `register.go`, so two authors landing
// at once touch nothing in common.
//
// No decoder: the rule takes no options, because its shape has nothing a configuration could tune; the
// producers are the default library's `Promise.race` and the global `setTimeout`, both fixed.
func init() {
	rule.Register(rule.Registration{Rule: CorrectnessNoUnclearedRaceTimeout})
}
