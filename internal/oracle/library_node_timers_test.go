package oracle

import (
	"path/filepath"
	"strings"
	"testing"
)

// timerSourceFixtures are setTimeout and clearTimeout called from source (#995spev), held to Node by
// the ordinary families: native under the sanitizers and the leak check, the JavaScript backend,
// release, gob and counted. Each schedules a timer, so wasm32-wasi refuses it at setTimeout.
var timerSourceFixtures = []string{"forwarded", "fewer_parameters", "clear_then_set", "order", "callback_throw", "clear_kinds"}

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
const clearTimeoutForeignStop = "clearTimeout of an object that isn't a timer but holds _onTimeout isn't built yet"

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

// timerDelayStop is the run-time NotYet a delay Node warns for and makes 1 ms stops in (Adamic's
// ruling on #995spev): timers_impl.h's words, and the JavaScript backend's adamicSetTimeout's.
const timerDelayStop = "a timer delay outside Node's range (NaN, negative or past 2^31 - 1 ms) isn't built yet"

// A delay Node warns for, NaN, negative or past 2^31 - 1, becomes 1 ms on Node with a warning on
// stderr that carries Node's process id, which no backend can print byte for byte. Each delay_ program
// stops alike natively, released and on the JavaScript backend after what Node printed before the
// timer, while Node runs on: the warning, then the callback, as at 1 ms. delay_computed.a, a delay
// computed inside the range, runs as Node's does. Mutant: timers_impl.h's check made Node's clamp,
// so delay_nan.a runs on natively at 1 ms with no warning.
func TestTimerDelayNeedingWarningsStops(t *testing.T) {
	t.Parallel()
	for name, warning := range map[string]string{"delay_nan.a": "TimeoutNaNWarning", "delay_negative.a": "TimeoutNegativeWarning", "delay_overflow.a": "TimeoutOverflowWarning"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path := timerSourcePath(t, name)
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != "before\nscheduled\nfired\n" || !strings.Contains(string(truth.stderr), warning+": ") || !strings.Contains(string(truth.stderr), "Timeout duration was set to 1.") {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatalf("the program must lower: %v", err)
			}
			want := run{stdout: []byte("before\n"), stderr: []byte("adamic: panic: " + timerDelayStop + "\n"), exitCode: 70}
			native, _ := natively(t, program)
			for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, got); difference != "" {
					t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
				}
			}
		})
	}
	t.Run("delay_computed.a", func(t *testing.T) {
		t.Parallel()
		node := "scheduled\nhalf\none and a half\nthree\n"
		timerSourceRun(t, timerSourcePath(t, "delay_computed.a"), node, run{stdout: []byte(node)})
	})
}
