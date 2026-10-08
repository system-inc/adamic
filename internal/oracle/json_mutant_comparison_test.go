package oracle

import (
	"fmt"
	"testing"
)

// A stdout catch needs a working reference. Infrastructure disagreements are
// reference failures, while agreement with the mutant is an actual survival.
func jsonMutantComparisonFailure(truth, got run) string {
	diff := disagreement(truth, got)
	if diff == "" {
		return "mutant survived"
	}
	var failure string
	switch {
	case truth.exitCode < 0:
		failure = fmt.Sprintf("Node reference exit %d: signal or timeout, stderr %q", truth.exitCode, truth.stderr)
	case truth.exitCode != 0:
		failure = fmt.Sprintf("Node reference exit %d, stderr %q", truth.exitCode, truth.stderr)
	case len(truth.stdout) == 0 && len(got.stdout) > 0:
		failure = fmt.Sprintf("Node reference produced empty output; native printed %d bytes", len(got.stdout))
	case diff != "stdout differs":
		failure = fmt.Sprintf("Node reference exit %d, stderr %q", truth.exitCode, truth.stderr)
	default:
		return ""
	}
	return fmt.Sprintf("Node side differed: %s (%s)", diff, failure)
}

func requireJSONMutantCaught(t *testing.T, truth, got run) {
	t.Helper()
	if failure := jsonMutantComparisonFailure(truth, got); failure != "" {
		t.Fatal(failure)
	}
}

func TestJSONMutantComparisonDiagnostics(t *testing.T) {
	printed := run{stdout: []byte("mutated\n")}
	cases := []struct {
		name       string
		truth, got run
		want       string
	}{
		{"caught", run{stdout: []byte("reference\n")}, printed, ""},
		{"survived", printed, printed, "mutant survived"},
		{"reference_exit", run{exitCode: 1, stderr: []byte("schema failed")}, printed, `Node side differed: exit codes differ (Node reference exit 1, stderr "schema failed")`},
		{"reference_killed", run{exitCode: -1}, printed, `Node side differed: exit codes differ (Node reference exit -1: signal or timeout, stderr "")`},
		{"reference_empty", run{}, printed, "Node side differed: stdout differs (Node reference produced empty output; native printed 8 bytes)"},
		{"reference_stderr", run{stdout: printed.stdout, stderr: []byte("warning")}, printed, `Node side differed: stderr differs (Node reference exit 0, stderr "warning")`},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := jsonMutantComparisonFailure(c.truth, c.got); got != c.want {
				t.Fatalf("got %q, want %q", got, c.want)
			}
		})
	}
}
