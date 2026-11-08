package oracle

import (
	"path/filepath"
	"testing"
)

// timerSourceFixtures are setTimeout and clearTimeout called from source (#995spev), held to Node by
// the ordinary families: native under the sanitizers and the leak check, the JavaScript backend,
// release, gob and counted. Each schedules a timer, so wasm32-wasi refuses it at setTimeout. The
// delay_ programs give delays Node warns for and makes 1 ms (#zhgmbwz): one of each kind, two of each
// (NaN and a negative warn once a process, an overflow every time), every kind with the hint line
// once, the warning's place from main, a timer's callback and a promise job, and delays inside the
// range that don't warn. Their warning lines differ only in the process id, which disagreement's
// timeoutWarningLine rule alone makes a token.
var timerSourceFixtures = []string{"forwarded", "fewer_parameters", "clear_then_set", "order", "callback_throw", "clear_kinds",
	"delay_nan", "delay_negative", "delay_overflow", "delay_nan_twice", "delay_negative_twice", "delay_overflow_twice", "delay_kinds",
	"delay_order_main", "delay_order_callback", "delay_order_job", "delay_computed"}

func init() {
	for _, name := range timerSourceFixtures {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{"internal/oracle/testdata/timers_source/" + name + ".a", true, false})
	}
}

// clearTimeoutForeignStop is the run-time NotYet clearTimeout stops in for an object that isn't a timer
// but holds a truthy _onTimeout, which Node would unenroll (Adamic's ruling on #995spev, Nov 7):
// timers_impl.h's words, and the JavaScript backend's adamicClearTimeout's.
const clearTimeoutForeignStop = "clearTimeout of an object that isn't a timer but holds _onTimeout isn't built yet; pass clearTimeout only what setTimeout or setInterval returned, or undefined"

// timerSourceRun is a timers_source program's run on each backend that can't be one of fixtures: its
// source on Node must print node, and natively (sanitized, with the leak check), released and on the
// JavaScript backend it must print want.
func timerSourceRun(t *testing.T, path, node string, want run) {
	t.Helper()
	if truth := onNode(t, path); disagreement(run{stdout: []byte(node)}, truth) != "" {
		t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("the program must lower: %v", err)
	}
	native, _ := natively(t, program)
	for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
		}
	}
}

func timerSourcePath(t *testing.T, name string) string {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/timers_source", name))
	if err != nil {
		t.Fatal(err)
	}
	return path
}

// Each stop_foreign_ program stops alike natively, released and on the JavaScript backend, after
// what Node printed before the clear, while Node, running the source, goes on. Each reaches the check
// by a road the compile-time guard doesn't read (Lower's must on #995spev): unknown, an intersection,
// a Record<string, unknown>'s key and a type parameter. They aren't among fixtures, whose families
// hold a program to Node's whole run. Mutant: timers_impl.h's check dropped, so every native leg runs
// on as Node's no-op would; clear_kinds.a's falsy and plain rows pin the other side.
func TestClearTimeoutForeignHandleStops(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"stop_foreign_handle.a", "stop_foreign_intersection.a", "stop_foreign_record.a", "stop_foreign_generic.a"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			want := run{stdout: []byte("live held\n"), stderr: []byte("adamic: panic: " + clearTimeoutForeignStop + "\n"), exitCode: 70}
			timerSourceRun(t, timerSourcePath(t, name), "live held\ncleared the object\nlive fired\n", want)
		})
	}
}

// clear_held.a clears a Timeout held as unknown and as NodeJS.Timeout | boolean, each a union whose
// reference is the Timeout, so the runtime cancels it by its shape (Library's must on #995spev).
// Mutant: native's nodeTimer handing a union's value over as NULL, so the unknown timer fires.
func TestClearTimeoutHeldAsUnionCancels(t *testing.T) {
	t.Parallel()
	timerSourceRun(t, timerSourcePath(t, "clear_held.a"), "cleared\nkept fired\n", run{stdout: []byte("cleared\nkept fired\n")})
}

// delay_kinds.a on Node prints what the rule was written from, Node v24's words: each warning's line,
// the duration line after each, and the hint line after the first alone, its pid made a token on both
// sides. Native and the JavaScript backend are held to Node by the families.
func TestTimerDelayWarningsInNodesWords(t *testing.T) {
	t.Parallel()
	const duration = "Timeout duration was set to 1.\n"
	want := run{
		stdout: []byte("scheduled\ninfinity fired\nnan fired\nnegative fired\njust past fired\n"),
		stderr: []byte("(node:1) TimeoutOverflowWarning: Infinity does not fit into a 32-bit signed integer.\n" + duration +
			"(Use `node --trace-warnings ...` to show where the warning was created)\n" +
			"(node:1) TimeoutNaNWarning: NaN is not a number.\n" + duration +
			"(node:1) TimeoutNegativeWarning: -1.5 is a negative number.\n" + duration +
			"(node:1) TimeoutOverflowWarning: 2147483647.5 does not fit into a 32-bit signed integer.\n" + duration),
	}
	if truth := onNode(t, timerSourcePath(t, "delay_kinds.a")); disagreement(want, truth) != "" {
		t.Fatalf("source Node: %s: exit %d, stdout %q, stderr %q", disagreement(want, truth), truth.exitCode, truth.stdout, truth.stderr)
	}
}

// The timeout warning rule lets two runs differ in the process id of one of Node's three warning
// lines and in nothing else (Adamic's ruling on #995spev, under Language's condition). Mutant: the
// rule as a general "(node:<digits>) ...Warning" pattern, so a misspelled warning word and another
// warning agree across two process ids.
func TestTimeoutWarningPidAloneIsNormalized(t *testing.T) {
	t.Parallel()
	const tail = "\nTimeout duration was set to 1.\n(Use `node --trace-warnings ...` to show where the warning was created)\n"
	for _, line := range []string{
		"TimeoutNaNWarning: NaN is not a number.",
		"TimeoutNegativeWarning: -2 is a negative number.",
		"TimeoutNegativeWarning: -1.5 is a negative number.",
		"TimeoutNegativeWarning: -1e-7 is a negative number.",
		"TimeoutNegativeWarning: -Infinity is a negative number.",
		"TimeoutOverflowWarning: 2147483648 does not fit into a 32-bit signed integer.",
		"TimeoutOverflowWarning: 2147483647.5 does not fit into a 32-bit signed integer.",
		"TimeoutOverflowWarning: 1e+300 does not fit into a 32-bit signed integer.",
		"TimeoutOverflowWarning: Infinity does not fit into a 32-bit signed integer.",
	} {
		node, native := run{stderr: []byte("before\n(node:48213) " + line + tail)}, run{stderr: []byte("before\n(node:7) " + line + tail)}
		if difference := disagreement(node, native); difference != "" {
			t.Errorf("%q under two process ids: %s", line, difference)
		}
	}
	for name, pair := range map[string][2]string{
		"a misspelled warning word":        {"(node:1) TimeoutNegaitveWarning: -2 is a negative number.\n", "(node:2) TimeoutNegaitveWarning: -2 is a negative number.\n"},
		"another warning":                  {"(node:1) DeprecationWarning: Buffer() is deprecated.\n", "(node:2) DeprecationWarning: Buffer() is deprecated.\n"},
		"another number":                   {"(node:1) TimeoutNegativeWarning: -2 is a negative number.\n", "(node:1) TimeoutNegativeWarning: -1.5 is a negative number.\n"},
		"a negative warning's positive":    {"(node:1) TimeoutNegativeWarning: 2 is a negative number.\n", "(node:2) TimeoutNegativeWarning: 2 is a negative number.\n"},
		"a number that isn't JavaScript's": {"(node:1) TimeoutOverflowWarning: 1E300 does not fit into a 32-bit signed integer.\n", "(node:2) TimeoutOverflowWarning: 1E300 does not fit into a 32-bit signed integer.\n"},
		"words after the line":             {"(node:1) TimeoutNaNWarning: NaN is not a number. twice\n", "(node:2) TimeoutNaNWarning: NaN is not a number. twice\n"},
		"words before the line":            {"say (node:1) TimeoutNaNWarning: NaN is not a number.\n", "say (node:2) TimeoutNaNWarning: NaN is not a number.\n"},
		"the duration line dropped":        {"(node:1) TimeoutNaNWarning: NaN is not a number.\nTimeout duration was set to 1.\n", "(node:2) TimeoutNaNWarning: NaN is not a number.\n"},
		"the hint line dropped":            {"(node:1) TimeoutNaNWarning: NaN is not a number.\n" + tail[1:], "(node:2) TimeoutNaNWarning: NaN is not a number.\nTimeout duration was set to 1.\n"},
	} {
		if difference := disagreement(run{stderr: []byte(pair[0])}, run{stderr: []byte(pair[1])}); difference != "stderr differs" {
			t.Errorf("%s: got %q, want \"stderr differs\"", name, difference)
		}
	}
}
