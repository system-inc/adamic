package oracle

import (
	"path/filepath"
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

// stop_foreign_handle.a stops alike natively, released and on the JavaScript backend, after what Node
// printed before the clear, while Node, running the source, goes on. It isn't one of fixtures, whose
// families hold a program to Node's whole run. Mutant: timers_impl.h's check dropped, so the native
// legs run on as Node's no-op would; clear_kinds.a's falsy and plain rows pin the other side.
func TestClearTimeoutForeignHandleStops(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/timers_source/stop_foreign_handle.a"))
	if err != nil {
		t.Fatal(err)
	}
	if truth := onNode(t, path); disagreement(run{stdout: []byte("live held\ncleared the object\nlive fired\n")}, truth) != "" {
		t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatalf("the clear must lower and stop at run time: %v", err)
	}
	want := run{stdout: []byte("live held\n"), stderr: []byte("adamic: panic: " + clearTimeoutForeignStop + "\n"), exitCode: 70}
	native, _ := natively(t, program)
	backend := onJavaScriptBackend(t, program)
	for name, got := range map[string]run{"native": native, "release": released(t, program), "javascript": backend} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s: %s: exit %d, stdout %q, stderr %q", name, difference, got.exitCode, got.stdout, got.stderr)
		}
	}
}
