package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A fixed object never reaches record storage by a road recordStorageView didn't see (#mez92e7,
// Language's probes on #2msw6t8, main 219dbb6e). On each road a write through the record ran
// adamic_record_define on a fixed object: an ASan segmentation fault in map.c's find, where Node prints 1 3. Each now stops
// at its road as not yet, in recordStorageView's words ("fixed objects and records have different
// storage"): a shorthand property (f02), the right of ??=, ||= and &&= (f05, f05b, f05c), and a type
// argument bound to a record constraint, inferred (f07) or written (f07b). f01, a plain assignment, held
// on main and still stops. The controls, the same roads carrying a real record (c1 to c3), lower and agree
// with Node on release native, sanitized native and the JavaScript backend. A stop's .stop file holds its
// position; the message must start there and carry the storage rule's words, so another rule's wording
// moving never reds this. Mutants: internal/lower/record-storage-roads.mutants.md.
func TestRecordStorageRoadsArePinned(t *testing.T) {
	t.Parallel()
	const wall = "(fixed objects and records have different storage; copy explicitly)"
	directory := "internal/oracle/testdata/record_storage_roads"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if extension := filepath.Ext(entry.Name()); extension != ".a" && extension != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			path := filepath.Join(directory, entry.Name())
			absolute, err := filepath.Abs(filepath.Join(repository, path))
			if err != nil {
				t.Fatal(err)
			}
			expected, err := os.ReadFile(absolute + ".stop")
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, absolute)
			truth := onNode(t, absolute)
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					// What the write does once it lowers, so a mutant's red names the fault.
					sanitized, _ := natively(t, program)
					t.Fatalf("lowered, want a stop at %s: Node exit %d %q, sanitized native exit %d %q, stderr %q", stop, truth.exitCode, truth.stdout, sanitized.exitCode, sanitized.stdout, sanitized.stderr)
				}
				got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), "")
				if !strings.HasPrefix(got, stop+": ") || !strings.Contains(got, wall) {
					t.Fatalf("stop: got %q, want one at %s carrying %q", got, stop, wall)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			sanitized, _ := natively(t, program)
			for _, backend := range []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"sanitized", sanitized}, {"JavaScript", onJavaScriptBackend(t, program)}} {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Fatalf("%s: %s: Node %q, %s %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 10 {
		t.Fatalf("found %d programs, want 10", programs)
	}
}
