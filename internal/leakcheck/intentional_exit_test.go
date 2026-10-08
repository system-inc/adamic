package leakcheck

import (
	"strings"
	"testing"
)

func TestIntentionalExitIsNarrow(t *testing.T) {
	const counts = "adamic: counts: allocations 1 frees 0 retains 1 releases 0 peak 1 regions 0\n"
	const marker = "adamic: intentional exit: status 2\n"
	cases := []struct {
		name     string
		run      Run
		accepted bool
	}{
		{"reported teardown", Run{Stderr: []byte(marker + counts), ExitCode: 2}, true},
		{"ordinary leak", Run{Stderr: []byte(counts)}, false},
		{"nonzero unmarked", Run{Stderr: []byte(counts), ExitCode: 2}, false},
		{"wrong status mutant", Run{Stderr: []byte(marker + counts), ExitCode: 1}, false},
		{"trailing diagnostic mutant", Run{Stderr: []byte(marker + counts + "unexpected failure\n"), ExitCode: 2}, false},
		{"interposed diagnostic mutant", Run{Stderr: []byte(marker + "unexpected failure\n" + counts), ExitCode: 2}, false},
		{"partial counts mutant", Run{Stderr: []byte(marker + "adamic: counts: allocations 1\n"), ExitCode: 2}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			report := Unbalanced(c.run)
			if (report == "") != c.accepted {
				t.Fatalf("report %q accepted=%t", report, c.accepted)
			}
		})
	}
	// Existing failure paths stay failures: a nonzero status without the terminal runtime marker.
	report, err := Check(Program{Execute: func(env []string, name string, args ...string) Run {
		return Run{ExitCode: 2, Stderr: []byte("ordinary failure\n")}
	}})
	if err == nil && !strings.Contains(report, "ordinary failure") {
		t.Fatalf("unmarked failure accepted: %q", report)
	}
}
