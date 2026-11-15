package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// TestSortSharedArguments holds the JavaScript backend's adamicSort, which hands every comparison of
// one sort the same two-element array, written in place, instead of a fresh one per call (#367nnz2).
// That is sound only while no closure's code keeps a reference to values past its entry: the entry
// copies each parameter, arguments.length is a number read from values.length, a rest is a slice taken
// at once, and the arguments object is refused (rule 2600). Each .a program is held to Node on the
// release build, the sanitized native build and the JavaScript backend: a plain comparator, captured
// ones, one that throws, one that answers NaN, and reentrant.a, the witness, whose comparator runs a
// second sort with its own closure comparator and then reads its own pair. arguments_kept.ts is the
// pin on the condition: a comparator that keeps its arguments object would see the pair rewritten, so
// it must stay refused. The planned mutant is in testdata/sort_shared_arguments/mutants.md.
func TestSortSharedArguments(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/sort_shared_arguments"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		extension := filepath.Ext(entry.Name())
		if extension != ".a" && extension != ".ts" {
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
			program, err := lowered(t, absolute)
			if entry.Name() == "arguments_kept.ts" {
				var notYet *lower.NotYet
				var refused *lower.Refused
				if !errors.As(err, &notYet) && !errors.As(err, &refused) {
					t.Fatalf("want a refusal of the arguments object, got %v", err)
				}
				if !strings.Contains(err.Error(), "arguments other than a read of arguments.length") {
					t.Fatalf("want the arguments object refused, got %v", err)
				}
				return
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
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
	if programs != 6 {
		t.Fatalf("found %d programs, want 6", programs)
	}
}
