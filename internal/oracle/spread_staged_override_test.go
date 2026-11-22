package oracle

import (
	"os"
	"path/filepath"
	"testing"
)

// A field a spread's literal writes into a slot its staged source reserved absent reads as written (#sx8pbw1's
// publish gap). The spread's copy keeps such a slot unwritten (adamic_object_copy_checked skips it), and on main
// f7fe5ce5 the literal's write filled it without publishing it: g01, g02 and g03, over a construction cast or a
// factory's, printed undefined natively, Object.keys and JSON.stringify left the field out and in said false, all
// with exit 0, where Node prints the value; g04, over an interface merged into a class, stopped with "read before
// assignment: field 'label' in m.label". The write now publishes the slot and makes it ready, in objectLiteral's
// spread arm (emit_objects.go), and on the reused path (reuse.go) as a defensive write: g06, its source a
// function's local, agrees with the reused path's publish removed, so no pin yet shows that path meeting a
// reserved slot. g05, its field written before the spread, is the control, and g08 writes the field through a
// nested spread, the spread arm again. Each agrees with Node on release and sanitized native and on JavaScript.
// Object.assign onto a spread's copy doesn't reach the gap: lowering stops it, since the copy's shape isn't
// proven by a plain literal.
func TestSpreadStagedOverride(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/spread_staged_override"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 7 {
		t.Fatalf("found %d programs, want 7", len(entries))
	}
	for _, entry := range entries {
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d, stdout %q, stderr %q", truth.exitCode, truth.stdout, truth.stderr)
			}
			program, err := lowered(t, absolute)
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Errorf("%s: %s: Node %q, got %q, stderr %q", backend.name, difference, truth.stdout, backend.got.stdout, backend.got.stderr)
				}
			}
		})
	}
}
