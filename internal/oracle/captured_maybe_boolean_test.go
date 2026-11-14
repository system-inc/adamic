package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// capturedMaybeBooleanPins are #fds729j's programs: a boolean | undefined local or parameter a
// function value captures, held in its cell as adamic_value's maybe_boolean byte. local is the
// captured local each pin is about. Each is held to Node, and to the pinned output too, so an
// agreement on the wrong text can't pass. Every pin prints a captured undefined somewhere, so a cell
// read as a plain boolean shows false there (internal/lower/captured_maybe_boolean.mutants.md).
var capturedMaybeBooleanPins = []struct{ name, local, want string }{
	{"uninitialized", "seen", "fresh undefined\nfresh false\nearly undefined\nearly true\n"},
	{"written", "seen", "step 0 true\nstep 1 false\nstep 2 undefined\nstep 3 true\n"},
	{"read_after", "seen", "start undefined true true true off\nstep 0 true true false false on\nstep 1 false false false true off\nstep 2 undefined true true true off\nstart false false false true off\nstep 0 true true false false on\nstep 1 false false false true off\nstep 2 undefined true true true off\n"},
	{"optional_parameter", "recursive", "nothing undefined false true\nundefined undefined false true\ntrue true true false\nfalse false false false\n"},
	{"narrowed", "seen", "unset, not off\non, not off\noff, off for sure\nunset, not off\n"},
	{"async", "seen", "before undefined true\nafter true true\nlast undefined true\nbefore undefined false\nafter false false\nlast undefined false\nbefore undefined undefined\nafter undefined undefined\nlast undefined undefined\n"},
	{"reporter", "pretty", "12:00 - built, pretty undefined true\n[12:01] built, pretty true false\n12:02 - built, pretty false false\n12:03 - built, pretty undefined true\n"},
}

// Each pin is committed as .ts and run as .a too, from the same text, so both dialects are held to
// one program, natively, sanitized with the leak check, on the counted build and on JavaScript
// (capturedUnionRuns). uninitialized measures the two things the cell's start has to get right: a
// let declared without an initializer reads undefined from its cell, and a cell made before its
// declaration (early's) is never read before the declaration runs. How each cell is made is logged.
func TestCapturedMaybeBooleanAgrees(t *testing.T) {
	t.Parallel()
	for _, pin := range capturedMaybeBooleanPins {
		for _, extension := range []string{".ts", ".a"} {
			t.Run(pin.name+extension, func(t *testing.T) {
				t.Parallel()
				path := capturedMaybeBooleanSource(t, pin.name, extension)
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				found := false
				for _, local := range program.Locals {
					if local.Name != pin.local {
						continue
					}
					if local.Type != ir.MaybeBoolean || !local.Captured {
						t.Fatalf("%s: Type %v Captured %t, want a captured boolean | undefined", local.Name, local.Type, local.Captured)
					}
					found = true
					t.Logf("%s: preallocated %t, environment cell %t, graph cell %t", local.Name, local.Preallocated, local.EnvironmentCell, local.GraphCell)
				}
				if !found {
					t.Fatalf("no local named %s", pin.local)
				}
				capturedUnionRuns(t, path, program, pin.want)
			})
		}
	}
}

// capturedMaybeBooleanSource writes a pin's committed .ts text to a file of the given extension.
func capturedMaybeBooleanSource(t *testing.T, name, extension string) string {
	t.Helper()
	source, err := os.ReadFile(filepath.Join(repository, "internal/oracle/testdata/captured_maybe_boolean", name+".ts"))
	if err != nil {
		t.Fatal(err)
	}
	path := filepath.Join(t.TempDir(), name+extension)
	if err := os.WriteFile(path, source, 0o644); err != nil {
		t.Fatal(err)
	}
	return path
}
