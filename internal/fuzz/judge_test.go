package fuzz

import (
	"os"
	"path/filepath"
	"testing"
)

// stopped is an outcome where Node ran to the end, printing nodeStdout, and both of Adamic's backends
// stopped with exit 70 after printing stdout, writing stderr.
func stopped(nodeStdout string, stdout string, stderr string) Outcome {
	adamic := Run{Stdout: []byte(stdout), Stderr: []byte(stderr), ExitCode: 70}
	return Outcome{Node: Run{Stdout: []byte(nodeStdout)}, Native: adamic, Backend: adamic}
}

// A shared stop is an inserted check only when its panic line is one of the checks Adamic inserts
// where JavaScript runs on. Anything else both backends stop at while Node finishes is a finding,
// however little they printed first: an empty stdout is a prefix of everything.
func TestJudgeReadsThePanicLine(t *testing.T) {
	t.Parallel()
	checkout := &Checkout{}
	for _, test := range []struct {
		name    string
		outcome Outcome
		want    Verdict
	}{
		// What JavaScript throws, where Node didn't: both backends stopped where nothing should.
		{"a JavaScript error Node never threw", stopped("", "", "adamic: panic: TypeError: Cannot set properties of undefined\n"), Finding},
		{"a JavaScript error after shared output", stopped("a\nb\n", "a\n", "adamic: panic: TypeError: Cannot read properties of undefined (reading 'value')\n"), Finding},
		{"a ReferenceError Node never threw", stopped("a\n", "", "adamic: panic: ReferenceError: Cannot access 'total' before initialization\n"), Finding},
		{"a program's own panic Node never reached", stopped("a\n", "a\n", "adamic: panic: boom\n"), Finding},
		{"a runtime failure", stopped("a\n", "", "adamic: panic: compiler bug: a function ended without returning\n"), Finding},

		// The checks Adamic inserts where JavaScript goes on with a value the type can't hold.
		{"undefined put back after a narrowing", stopped("a\nundefined\n", "a\n", "adamic: panic: undefined where the checker narrowed it away: a call since the narrowing put it back\n"), Checked},
		{"null put back after a narrowing", stopped("a\nnull\n", "a\n", "adamic: panic: null where the checker narrowed it away: a call since the narrowing put it back\n"), Checked},
		{"map over an array that shrank", stopped("1,,\n", "", "adamic: panic: map: the array shrank while it was being mapped\n"), Checked},
		{"find over an array that shrank", stopped("-1\n", "", "adamic: panic: find: the array shrank while it was being searched\n"), Checked},
		{"findIndex over an array that shrank", stopped("-1\n", "", "adamic: panic: findIndex: the array shrank while it was being searched\n"), Checked},
		{"a write past the end", stopped("a\n4\n", "a\n", "adamic: panic: index 3 is outside an array of length 2\n"), Checked},
		{"a write at a fractional index", stopped("2\n", "", "adamic: panic: index 0.5 is outside an array of length 2\n"), Checked},
		{"a cast the discriminant refuses", stopped("circle\n", "", "adamic: panic: cast failed: this Shape is not a Square\n"), Checked},

		// An inserted check's words, but not a shared stop Node's run can explain.
		{"a check after output Node never printed", stopped("a\n", "b\n", "adamic: panic: index 3 is outside an array of length 2\n"), Finding},
		{"a check with more after it", stopped("a\n", "", "adamic: panic: index 3 is outside an array of length 2 and more\n"), Finding},
	} {
		got := checkout.judge(test.outcome, "", "")
		if got.Verdict != test.want {
			t.Errorf("%s: %s (%s), want %s", test.name, got.Verdict, got.Key, test.want)
		}
	}

	// The two backends stopped at different checks: they disagree with each other, whatever the words.
	split := stopped("a\n", "", "adamic: panic: undefined where the checker narrowed it away: a call since the narrowing put it back\n")
	split.Backend.Stderr = []byte("adamic: panic: index 3 is outside an array of length 2\n")
	if got := checkout.judge(split, "", ""); got.Verdict != Finding {
		t.Errorf("backends stopped at different checks: %s (%s), want %s", got.Verdict, got.Key, Finding)
	}
}

// ran is an outcome where all three ways printed "a" and "b"; each case changes what one of Adamic's
// backends did.
func ran() Outcome {
	clean := Run{Stdout: []byte("a\nb\n")}
	return Outcome{Node: clean, Native: clean, Backend: clean}
}

// What AddressSanitizer and UndefinedBehaviorSanitizer wrote on this repository's macOS host for a
// null read, a read after free and a bool holding 5, compiled with native.Flags' sanitizer flags.
// macOS aborts after every report (abort_on_error), so those runs also died by SIGABRT.
const (
	addressSanitizerSegv = "AddressSanitizer:DEADLYSIGNAL\n=================================================================\n==91336==ERROR: AddressSanitizer: SEGV on unknown address 0x000000000000 (pc 0x0001045c4860 bp 0x00016b83abc0 sp 0x00016b83abb0 T0)\n==91336==The signal is caused by a READ memory access.\n==91336==Hint: address points to the zero page.\n"
	addressSanitizerFree = "=================================================================\n==91412==ERROR: AddressSanitizer: heap-use-after-free on address 0x602000000950 at pc 0x000102668744 bp 0x00016d796bb0 sp 0x00016d796ba8\nREAD of size 4 at 0x602000000950 thread T0\n    #0 0x000102668740 in main main.c:2\n"
	undefinedBool        = "main.c:3:64: runtime error: load of value 5, which is not a valid value for type '_Bool'\nSUMMARY: UndefinedBehaviorSanitizer: undefined-behavior main.c:3:64 \n"
)

// A native run that dies by a signal or carries a sanitizer's report is a Crash, the most severe
// verdict, whatever Node did: ending cleanly, or throwing where native crashed. It's never folded
// into Finding, and its Key is the one line shrinking and reducing hold it to.
func TestJudgeCallsACrashACrash(t *testing.T) {
	t.Parallel()
	checkout := &Checkout{}
	for _, test := range []struct {
		name    string
		outcome func() Outcome
		want    Verdict
		key     string
	}{
		{"a segfault with Node clean", func() Outcome {
			outcome := ran()
			outcome.Native = Run{Stdout: []byte("a\n"), ExitCode: -1, Signal: "segmentation fault"}
			return outcome
		}, Crash, "native signal: segmentation fault"},
		{"an AddressSanitizer SEGV, exiting 1 as ASan does on Linux", func() Outcome {
			outcome := ran()
			outcome.Native = Run{Stdout: []byte("a\n"), Stderr: []byte(addressSanitizerSegv), ExitCode: 1}
			return outcome
		}, Crash, "native AddressSanitizer: SEGV"},
		{"an AddressSanitizer report, aborting as on macOS", func() Outcome {
			outcome := ran()
			outcome.Native = Run{Stdout: []byte("a\nb\n"), Stderr: []byte(addressSanitizerFree), ExitCode: -1, Signal: "abort trap"}
			return outcome
		}, Crash, "native AddressSanitizer: heap-use-after-free"},
		{"an UndefinedBehaviorSanitizer report", func() Outcome {
			outcome := ran()
			outcome.Native = Run{Stdout: []byte("a\n"), Stderr: []byte(undefinedBool), ExitCode: 1}
			return outcome
		}, Crash, "native UndefinedBehaviorSanitizer: runtime error: load of value N, which is not a valid value for type '_Bool'"},
		{"a ThreadSanitizer report", func() Outcome {
			outcome := ran()
			outcome.Native.Stderr = []byte("==================\nWARNING: ThreadSanitizer: data race (pid=4242)\n  Write of size 8 at 0x7b0400000010 by thread T1:\n")
			outcome.Native.ExitCode = 66
			return outcome
		}, Crash, "native ThreadSanitizer: data race"},
		{"a segfault where Node threw at the same point", func() Outcome {
			outcome := ran()
			outcome.Node = Run{Stdout: []byte("a\n"), Stderr: []byte("adamic: panic: TypeError: Cannot read properties of undefined (reading 'value')\n"), ExitCode: 70}
			outcome.Backend = outcome.Node
			outcome.Native = Run{Stdout: []byte("a\n"), Stderr: []byte(addressSanitizerSegv), ExitCode: -1, Signal: "abort trap"}
			return outcome
		}, Crash, "native AddressSanitizer: SEGV"},
		{"the JavaScript backend dying by a signal", func() Outcome {
			outcome := ran()
			outcome.Backend = Run{Stdout: []byte("a\n"), Stderr: []byte("FATAL ERROR: Reached heap limit Allocation failed - JavaScript heap out of memory\n"), ExitCode: -1, Signal: "abort trap"}
			return outcome
		}, Crash, "javascript backend signal: abort trap"},

		// Not crashes.
		{"a plain stdout difference", func() Outcome {
			outcome := ran()
			outcome.Native.Stdout = []byte("a\nc\n")
			return outcome
		}, Finding, "native stdout differs"},
		{"an inserted check", func() Outcome {
			return stopped("a\nundefined\n", "a\n", "adamic: panic: undefined where the checker narrowed it away: a call since the narrowing put it back\n")
		}, Checked, "inserted check"},
		{"a native run killed at its deadline", func() Outcome {
			outcome := ran()
			outcome.Native = Run{Stdout: []byte("a\n"), ExitCode: -1, Signal: "killed", TimedOut: true}
			return outcome
		}, Finding, "native never finished"},
		{"sanitizer words the program wrote itself", func() Outcome {
			wrote := Run{Stdout: []byte("a\n"), Stderr: []byte("runtime error: the program's own words\n"), ExitCode: 1}
			return Outcome{Node: wrote, Native: wrote, Backend: wrote}
		}, Agreed, ""},
	} {
		got := checkout.judge(test.outcome(), "", "")
		if got.Verdict != test.want || got.Key != test.key {
			t.Errorf("%s: %s (%s), want %s (%s)", test.name, got.Verdict, got.Key, test.want, test.key)
		}
	}
}

// The leak run, the same binary again with leak detection on: LeakSanitizer's report is a leak, a
// finding, never a crash, since a leak isn't memory corruption. A failure with no report is a leak
// too; that is how macOS's ASan fails there, aborting because it can't detect leaks, and a whole run
// of agreed programs must not turn into crashes on a Mac. A bad access the leak run reports is still
// a crash.
func TestJudgeReadsTheLeakRun(t *testing.T) {
	t.Parallel()
	checkout := &Checkout{}
	for _, test := range []struct {
		name   string
		script string
		want   Verdict
		key    string
	}{
		{"a LeakSanitizer report", "echo '==7==ERROR: LeakSanitizer: detected memory leaks' >&2\necho 'Direct leak of 16 byte(s) in 1 object(s) allocated from:' >&2\nexit 23\n", Finding, "leak"},
		{"a bad access only the leak run reports", "echo '==7==ERROR: AddressSanitizer: heap-use-after-free on address 0x602000000950 at pc 0x000102668744' >&2\nexit 1\n", Crash, "native AddressSanitizer: heap-use-after-free"},
		{"an abort with no report", "kill -ABRT $$\n", Finding, "leak"},
		{"nothing leaked", "exit 0\n", Agreed, ""},
	} {
		directory := t.TempDir()
		binary := filepath.Join(directory, "program")
		if err := os.WriteFile(binary, []byte("#!/bin/sh\n"+test.script), 0o755); err != nil {
			t.Fatal(err)
		}
		got := checkout.judge(ran(), binary, directory)
		if got.Verdict != test.want || got.Key != test.key {
			t.Errorf("%s: %s (%s), want %s (%s)\n%s", test.name, got.Verdict, got.Key, test.want, test.key, got.Detail)
		}
	}
}

// Every verdict is listed once, a crash first and a finding right after it, so a summary reads the
// worst news first.
func TestVerdictsGoMostSevereFirst(t *testing.T) {
	t.Parallel()
	if len(Verdicts) < 2 || Verdicts[0] != Crash || Verdicts[1] != Finding {
		t.Fatalf("verdicts in order %v, want crash then finding first", Verdicts)
	}
	seen := map[Verdict]bool{}
	for _, verdict := range Verdicts {
		if seen[verdict] {
			t.Errorf("%s listed twice", verdict)
		}
		seen[verdict] = true
	}
	for _, verdict := range []Verdict{Crash, Finding, Agreed, Checked, NotYet, Invalid, Unfit} {
		if !seen[verdict] {
			t.Errorf("%s isn't listed", verdict)
		}
	}
}
