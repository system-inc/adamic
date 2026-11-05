package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Generators' first slice (#qk8rztp S1c, S2a's generic ones per instance, S2b-1a's new Set and
// new Map, S2b-1b's views, tsc's mapIterator and mapDefinedIterator stepping a collection's iterator
// in their bodies, and S2c-1's generator handed into a one-read Iterable<T> parameter, tsc's
// arrayFrom among them): a top-level function* consumed by a for-of over its call, a spread,
// Array.from, new Set or new Map of it, a for-of over the parameter it is passed straight into, or
// dropped by a call as a statement, whatever iterator type it is declared to return. Each fixture in testdata/generators agrees with Node natively,
// sanitized and on the JavaScript backend, and one that finishes leaks nothing: every way out of the
// consuming loop (exhaustion, break, return, a throw from the body or the loop, a labeled jump to an
// outer loop) and of the body (completion, an early return, a throw) lets go of the generator once
// and of what its frame holds. Strings are built, never only literals, so a missed release or an
// extra one is a real leak or a real use after free.
//
// Each fixture's mutant is recorded beside it, in testdata/generators/mutants.md.
func TestGeneratorsAgreeWithNode(t *testing.T) {
	t.Parallel()
	directory := "internal/oracle/testdata/generators"
	entries, err := os.ReadDir(filepath.Join(repository, directory))
	if err != nil {
		t.Fatal(err)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") {
			continue
		}
		programs++
		t.Run(entry.Name(), func(t *testing.T) {
			t.Parallel()
			generatorAgrees(t, directory+"/"+entry.Name())
		})
	}
	if programs != 40 {
		t.Fatalf("%d programs in %s, want 40: a program added or removed needs this count changed with it", programs, directory)
	}
}

// generatorAgrees holds one program to Node on every backend, whatever Node's exit, and, when Node
// finishes it, to leaking nothing.
func generatorAgrees(t *testing.T, relative string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, relative))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	sanitized, binary := natively(t, program)
	for _, candidate := range []struct {
		name string
		got  run
	}{{"native", released(t, program)}, {"sanitized", sanitized}, {"javascript", onJavaScriptBackend(t, program)}} {
		if difference := disagreement(want, candidate.got); difference != "" {
			t.Errorf("%s: %s, stdout %q, stderr %q", candidate.name, difference, candidate.got.stdout, candidate.got.stderr)
		}
	}
	if want.exitCode == 0 {
		if report := leaks(t, program, binary); report != "" {
			t.Errorf("leaked: %s", report)
		}
	}
}
