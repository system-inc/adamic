package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// A write to a property that only an in test added to a single object's type stops at the write as not
// yet, in .a and .ts (#yqbg9vz). `'extra' in view` narrows view: Plain to Plain & Record<"extra", unknown>,
// and tsc takes any value for view.extra, so a string (i3) or an object (i4) went into real's number
// slot: Node prints text1, and natively it faulted under ASan. Every write form stops by the same words:
// a compound assignment (w1), ++ on the read narrowed further by typeof, since tsc refuses ++ on unknown
// (w2), ??= (w3), a destructuring target (w4) and a literal key (w5). Their controls lower and agree with
// Node on release native, sanitized native and the JavaScript backend: the in test alone (i0), a typeof
// read of the added property (i2), which stays dynamic, and a write in the same block to a property Plain
// declares (c1). (i1, a function written there, stops on an unrelated not yet first, so it isn't pinned.)
// Mutant: inNarrowedWriteGap's stop dropped, and every write here lowers; i3 then faults under the
// sanitizers or prints other than Node's text1, which the failure below shows beside Node's.
func TestInNarrowedWritesArePinned(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/in_narrowed_write"
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
					t.Fatalf("lowered, want %q: Node exit %d %q, sanitized native exit %d %q, stderr %q", stop, truth.exitCode, truth.stdout, sanitized.exitCode, sanitized.stdout, sanitized.stderr)
				}
				if got := strings.ReplaceAll(err.Error(), strings.TrimSuffix(absolute, path), ""); got != stop {
					t.Fatalf("stop: got %q, want %q", got, stop)
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
	if programs != 20 {
		t.Fatalf("found %d programs, want 20", programs)
	}
}
