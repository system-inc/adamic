package leakcheck

import (
	"runtime"
	"strings"
	"testing"
)

func TestReportedLeakIsRejected(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux LeakSanitizer report control")
	}
	calls := 0
	report, err := Check(
		Program{Sanitized: "fixture", Execute: func(environment []string, name string, arguments ...string) Run {
			calls++
			// ADAMIC_LEAK_CHECK asks the runtime to mark an intentional exit (adamic.c), so the check
			// can tell process.exit's teardown from a leak.
			if len(environment) != 2 ||
				environment[0] != "ASAN_OPTIONS=detect_leaks=1" ||
				environment[1] != "ADAMIC_LEAK_CHECK=1" ||
				name != "fixture" {
				test.Fatal("leak detection was not requested", environment, name)
			}
			return Run{ExitCode: 1, Stderr: []byte("ERROR: LeakSanitizer: detected memory leaks\n")}
		}},
	)
	if err != nil || calls != 1 || !strings.Contains(report, "ERROR: LeakSanitizer") {
		test.Fatalf("reported leak was ignored: calls=%d report=%q error=%v", calls, report, err)
	}
}

func TestCountedBalance(test *testing.T) {
	test.Parallel()
	for _, testCase := range []struct {
		line   string
		leaked bool
	}{
		{"adamic: counts: allocations 3 frees 2 retains 0 releases 0 peak 3 regions 1\n", false},
		{"adamic: counts: allocations 3 frees 1 retains 0 releases 0 peak 3 regions 1\n", true},
		{"missing counts\n", true},
	} {
		if report := Unbalanced(Run{Stderr: []byte(testCase.line)}); (report != "") != testCase.leaked {
			test.Fatalf("counts %q: report %q", testCase.line, report)
		}
	}
}
