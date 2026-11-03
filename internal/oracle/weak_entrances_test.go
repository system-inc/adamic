package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A record seen as a weak type, one whose every field is optional, is where tsc lets any value through:
// a { [key: string]: string } is assignable to { overrideImportMode?: "A" | "B" } without its values
// being checked against the field (#yf2kj41, Language's note on #e8fjpb0). So the string "esm" can sit
// in a field typed "A" | "B" with no cast, and Node prints string:esm. Each program here takes that
// record into a weak slot by one road and reads the field, in .a and in .ts. Each pins the stop it
// meets: recordStorageView's storage NotYet at the entrance, or, on a road that names no view site,
// rule (c)'s NotYet at the read (entries_records.go). None may lower, since the read would believe the
// lie; the honest literal (k16) lowers and agrees with Node. The roads: a local, a parameter, a return,
// a field write, an array element, a shorthand property, a property assignment, ??=, a conditional
// arm, a destructuring default, a parameter default, a class field, push, a spread argument and
// Map.set. When the storage stop falls (#6c4swvt B admits records seen as fixed views), every pin here
// must still stop, at the read's run-time check if not before: that's what keeps rule (c) honest
// (#1w8gvx9).
func TestWeakEntrancesArePinned(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/weak_entrances"
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
			if stop := strings.TrimSpace(string(expected)); stop != "lowers" {
				if err == nil {
					t.Fatalf("lowered, want %q: Node prints %q", stop, onNode(t, absolute).stdout)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			truth := onNode(t, absolute)
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
	if programs != 32 {
		t.Fatalf("found %d programs, want 32", programs)
	}
}
