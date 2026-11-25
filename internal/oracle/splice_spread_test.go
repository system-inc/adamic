package oracle

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// TestSpliceSpread holds splice with a spread among its items (#p1qw4kf), which tsc's own
// insertStatementsAfterPrologue writes as to.splice(statementIndex, 0, ...from). The items are packed as
// push packs them, every spread expanded before the array changes, and array_splice_arguments resolves
// the insertion index before it removes. Each .ts program but the stop is held to Node on the release
// build, the sanitized build under the leak check and the JavaScript backend, and must lower through
// array_splice_arguments: the site's shape, called with objects; a negative start, a start and a count
// past the end; an empty spread beside the old roads; the array spread into itself; and plain items
// mixed with several spreads of numbers, strings and objects, in JavaScript's order. stop_representation.ts
// spreads functions into objects and stops in push's words. The mutants are in
// testdata/splice_spread/mutants.md.
func TestSpliceSpread(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/splice_spread"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if filepath.Ext(entry.Name()) != ".ts" {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			absolute, err := filepath.Abs(filepath.Join(repository, directory, entry.Name()))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, absolute)
			if truth.exitCode != 0 || len(truth.stderr) != 0 {
				t.Fatalf("source Node: exit %d stderr %q", truth.exitCode, truth.stderr)
			}
			program, err := lowered(t, absolute)
			if entry.Name() == "stop_representation.ts" {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) || !strings.Contains(err.Error(), "spreading elements with an incompatible representation into splice (convert the elements explicitly first)") {
					t.Fatalf("want the spread stopped in push's words, got %v", err)
				}
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			spliced := false
			for _, function := range program.Functions {
				spliced = spliced || function.Name == "array_splice_arguments"
			}
			if !spliced {
				t.Fatal("no splice lowered through array_splice_arguments")
			}
			sanitized, leaked, comparable := nativelyCheckingLeaks(t, program, true)
			if leaked != "" {
				t.Errorf("leaks: %s", leaked)
			}
			backends := []struct {
				name string
				got  run
			}{{"release", released(t, program)}, {"JavaScript", onJavaScriptBackend(t, program)}}
			if comparable {
				backends = append(backends, struct {
					name string
					got  run
				}{"sanitized", sanitized})
			}
			for _, backend := range backends {
				if difference := disagreement(truth, backend.got); difference != "" {
					t.Errorf("%s: %s: Node %q, %s %q, stderr %q", backend.name, difference, truth.stdout, backend.name, backend.got.stdout, backend.got.stderr)
				}
			}
		})
	}
	// A program lost from the directory must fail here, not shrink the walk.
	if programs != 6 {
		t.Fatalf("found %d programs, want 6", programs)
	}
}
