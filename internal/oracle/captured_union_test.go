package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/leakcheck"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// capturedUnionPins are #n7sx9gz's programs: a let whose members are held differently (an object, a
// string literal, undefined), captured by the closures that write and read it. Each is held to Node,
// and to the pinned output too, so an agreement on the wrong text can't pass.
var capturedUnionPins = []struct{ name, want string }{
	{"class", "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"},
	{"object", "start undefined false\nstep 0 session first true object\nstep 1 stopping true string\nstep 2 undefined false undefined\nenabled\nenable while profiled\ndisabled profiled during stopping\nnothing to disable\nenable while stopping\nenabled\nafter a write the cell says string\nstopping\n"},
	{"by_value", "step 0 false false false false true\nstep 1 true true true true false\nstep 2 false false false false true\nstep 3 true true true true false\nthe cell holds the word\nthe cell holds something else\nthe cell holds the word\n"},
	{"released", "first 0 of 3, stopping, undefined, second 0 of 3, third 0 of 3, stopping, last 0 of 3\nfirst 1 of 3, stopping, undefined, second 1 of 3, third 1 of 3, stopping, last 1 of 3\nfirst 2 of 3, stopping, undefined, second 2 of 3, third 2 of 3, stopping, last 2 of 3\n"},
	{"nested", "no base\none base ./base.json\n3 bases ./a.json+./b.json+./c.json\n2 bases ./a.json+./b.json\none base ./base.json\nno base\nno base\n"},
	{"alternating", "session 0.0 with 2, stopping, undefined, session 0.3 with 2, stopping, undefined, session 0.6 with 2, stopping\nsession 1.0 with 2, stopping, undefined, session 1.3 with 2, stopping, undefined, session 1.6 with 2, stopping\n"},
	{"graph", "held undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\nheld undefined, held first, held stopping, held undefined, held again 3, held again 4, held stopping\n"},
}

// Each pin is committed as .ts and run as .a too, from the same text, so both dialects are held to
// one program. Every build is asked whether it let go of everything: the sanitized one by the shared
// leak check, and the counted one by its own counts, where an overwritten member never released reads
// as a value leaked.
func TestCapturedUnionAgrees(t *testing.T) {
	t.Parallel()
	for _, pin := range capturedUnionPins {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(pin.name+extension, func(t *testing.T) {
				t.Parallel()
				path := capturedUnionSource(t, pin.name, extension)
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				if pin.name == "graph" && !capturedUnionGraphCell(program, "active") {
					t.Fatal("active's cell isn't a graph cell, so the graph pin doesn't pin one")
				}
				capturedUnionRuns(t, path, program, pin.want)
			})
		}
	}
}

// Memory's m3: the Session the cell holds keeps a closure that keeps the cell, a cycle through the
// Union. The rule holds it, agreeing with Node and letting go of everything, or refuses it as a cycle;
// either is an answer, and anything else fails.
func TestCapturedUnionCycleHeldOrRefused(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".ts", ".a"} {
		t.Run("cycle"+extension, func(t *testing.T) {
			t.Parallel()
			path := capturedUnionSource(t, "cycle", extension)
			program, err := lowered(t, path)
			var refused *lower.Refused
			var notYet *lower.NotYet
			switch {
			case err == nil:
				t.Logf("held: graph cell %t", capturedUnionGraphCell(program, "active"))
				capturedUnionRuns(t, path, program, "undefined, session first 0, stopping, session second 0, undefined, session last 0\nundefined, session first 1, stopping, session second 1, undefined, session last 1\n")
			case (errors.As(err, &refused) || errors.As(err, &notYet)) && strings.Contains(err.Error(), "cycle"):
				t.Logf("refused: %v", err)
			default:
				t.Fatalf("neither held nor refused as a cycle: %v", err)
			}
		})
	}
}

// capturedUnionSource writes a pin's committed .ts text to a file of the given extension.
func capturedUnionSource(t *testing.T, name, extension string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/captured_union", name+".ts"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name+extension)
	if err := os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}

// capturedUnionGraphCell is whether the captured local of that name lives in a graph cell.
func capturedUnionGraphCell(program *ir.Program, name string) bool {
	for _, local := range program.Locals {
		if local.Name == name && local.Captured && local.GraphCell {
			return true
		}
	}
	return false
}

// capturedUnionRuns holds a lowered pin to Node and to want, natively, sanitized with the leak check,
// on the counted build with its counts, and on JavaScript.
func capturedUnionRuns(t *testing.T, path string, program *ir.Program, want string) {
	t.Helper()
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != want {
		t.Fatalf("Node: exit %d, stdout %q, want %q, stderr %q", truth.exitCode, truth.stdout, want, truth.stderr)
	}
	t.Run("native", func(t *testing.T) {
		got := releasedUncached(t, program)
		if difference := disagreement(truth, got); difference != "" {
			t.Errorf("%s: got stdout %q, stderr %q", difference, got.stdout, got.stderr)
		}
	})
	t.Run("sanitized", func(t *testing.T) {
		got, leaked, comparable := nativelyCheckingLeaks(t, program, true)
		if leaked != "" {
			t.Errorf("leaks: %s", leaked)
		}
		if !comparable {
			return
		}
		if difference := disagreement(truth, got); difference != "" {
			t.Errorf("%s: got stdout %q, stderr %q", difference, got.stdout, got.stderr)
		}
	})
	t.Run("counted", func(t *testing.T) {
		binary := filepath.Join(t.TempDir(), "counted")
		if err := native.Build(native.C(program), binary, native.Options{Count: true}); err != nil {
			t.Fatal(err)
		}
		got := execute(t, binary)
		if report := leakcheck.Unbalanced(leakRun(got)); report != "" {
			t.Errorf("counts: %s", report)
		}
		if string(got.stdout) != want {
			t.Errorf("got stdout %q", got.stdout)
		}
		if match := leakcheck.CountsLine.FindSubmatch(got.stderr); match != nil {
			t.Logf("counts: allocations %s, frees %s, in regions %s", match[1], match[2], match[6])
		}
	})
	t.Run("javascript", func(t *testing.T) {
		got := onJavaScriptBackend(t, program)
		if difference := disagreement(truth, got); difference != "" {
			t.Errorf("%s: got stdout %q, stderr %q", difference, got.stdout, got.stderr)
		}
	})
}
