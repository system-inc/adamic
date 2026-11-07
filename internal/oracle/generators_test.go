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
// arrayFrom among them, and S2c-3's into an Iterable<T> | undefined one, tsc's
// firstOrUndefinedIterator): a top-level function* consumed by a for-of over its call, a spread,
// Array.from, new Set or new Map of it, a for-of over the parameter it is passed straight into, or
// dropped by a call as a statement, whatever iterator type it is declared to return; and S3a's
// function* nested in a function's body, made through its parent's closure, whose frame keeps it,
// consumed by a for-of over its call, the parameter or a call as a statement (the nested_ fixtures,
// tsc's checker.ts shapes among them); and S3b's anonymous function* written in place and called, a
// closure of its own made for the call, consumed the same three ways (the in_place_ fixtures, tsc's
// elaborateJsxComponents shape among them). Each fixture in testdata/generators agrees with Node natively,
// sanitized and on the JavaScript backend, and one that finishes leaks nothing: every way out of the
// consuming loop (exhaustion, break, return, a throw from the body or the loop, a labeled jump to an
// outer loop) and of the body (completion, an early return, a throw) lets go of the generator once
// and of what its frame holds. Strings are built, never only literals, so a missed release or an
// extra one is a real leak or a real use after free.
//
// Each fixture's mutant is recorded beside it, in testdata/generators/mutants.md.
func TestGeneratorsAgreeWithNode(test *testing.T) {
	test.Parallel()
	directory := "internal/oracle/testdata/generators"
	entries, readDirectoryError := os.ReadDir(filepath.Join(repository, directory))
	if readDirectoryError != nil {
		test.Fatal(readDirectoryError)
	}
	programs := 0
	for _, entry := range entries {
		if !strings.HasSuffix(entry.Name(), ".a") {
			continue
		}
		programs++
		test.Run(entry.Name(), func(test *testing.T) {
			test.Parallel()
			generatorAgrees(test, directory+"/"+entry.Name())
		})
	}
	if programs != 54 {
		test.Fatalf(
			"%d programs in %s, want 54: a program added or removed needs this count changed with it",
			programs,
			directory,
		)
	}
}

// generatorAgrees holds one program to Node on every backend, whatever Node's exit, and, when Node
// finishes it, to leaking nothing.
func generatorAgrees(test *testing.T, relative string) {
	test.Helper()
	path, absolutePathError := filepath.Abs(filepath.Join(repository, relative))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	want := onNode(test, path)
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	sanitized, binary := natively(test, program)
	for _, candidate := range []struct {
		name string
		got  run
	}{
		{"native", released(test, program)},
		{"sanitized", sanitized},
		{"javascript", onJavaScriptBackend(test, program)},
	} {
		if difference := disagreement(want, candidate.got); difference != "" {
			test.Errorf(
				"%s: %s, stdout %q, stderr %q",
				candidate.name,
				difference,
				candidate.got.stdout,
				candidate.got.stderr,
			)
		}
	}
	if want.exitCode == 0 {
		if report := leaks(test, program, binary); report != "" {
			test.Errorf("leaked: %s", report)
		}
	}
}
