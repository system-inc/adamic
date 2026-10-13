package leakcheck

import (
	"runtime"
	"strings"
	"testing"
)

func TestReportedLeakIsRejected(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux LeakSanitizer report control")
	}
	calls := 0
	report, err := Check(Program{Sanitized: "fixture", Execute: func(environment []string, name string, arguments ...string) Run {
		calls++
		if len(environment) != 1 || environment[0] != "ASAN_OPTIONS=detect_leaks=1" || name != "fixture" {
			t.Fatal("leak detection was not requested", environment, name)
		}
		return Run{ExitCode: 1, Stderr: []byte("ERROR: LeakSanitizer: detected memory leaks\n")}
	}})
	if err != nil || calls != 1 || !strings.Contains(report, "ERROR: LeakSanitizer") {
		t.Fatalf("reported leak was ignored: calls=%d report=%q error=%v", calls, report, err)
	}
}

func TestCountedBalance(t *testing.T) {
	t.Parallel()
	for _, test := range []struct {
		line   string
		leaked bool
	}{
		{"adamic: counts: allocations 3 frees 2 retains 0 releases 0 peak 3 regions 1\n", false},
		{"adamic: counts: allocations 3 frees 1 retains 0 releases 0 peak 3 regions 1\n", true},
		{"missing counts\n", true},
	} {
		if report := Unbalanced(Run{Stderr: []byte(test.line)}); (report != "") != test.leaked {
			t.Fatalf("counts %q: report %q", test.line, report)
		}
	}
}
