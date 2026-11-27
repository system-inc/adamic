package main

import (
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/compilerinputs"

	// The lowering companion's packages, linked so a change to anything they import selects this test wherever
	// selection follows Go's import graph: a cohere bump that adds a package to the closure then fails here.
	_ "github.com/system-inc/adamic/internal/load"
	_ "github.com/system-inc/adamic/internal/lower"
)

// The packages admission-lower runs: cmd/adamic's compile (with its check) calls the loader and the lowerer and stops,
// so the C emitter and the JavaScript backend it also links can't move an admission set.
var lowering = []string{"./internal/load", "./internal/lower"}

// compilerinputs.Admission, the admission projection internal/compilerinputs/compilerinputs_table.go holds, is what
// the lowering companion links, no more: every in-repository package go list says internal/load and internal/lower
// depend on, cmd/adamic (whose compile and check the adapter calls), and each module file pinning one (with its go.sum,
// and go.work, which picks the modules), is under an entry; every entry is needed, so dropping any leaves a gap; and the
// list is sorted, once each. It's read live, so no committed copy of it can go stale (#w5zzdjp). Mutants, in the
// table's admission list: internal/ir dropped (the gap is named), internal/native added (it names nothing the
// companion's lowering links).
func TestInputsAreWhatTheLoweringCompanionLinks(t *testing.T) {
	t.Parallel()
	declared := compilerinputs.Admission()
	if len(declared) == 0 {
		t.Fatal("the admission projection is empty")
	}
	if !sort.StringsAreSorted(declared) {
		t.Fatal("the admission projection is not sorted")
	}
	for index, input := range declared {
		if input == "" || strings.TrimSpace(input) != input || filepath.IsAbs(input) || filepath.ToSlash(filepath.Clean(input)) != input {
			t.Fatalf("the admission projection's entry %d, %q, is not a clean repository path", index+1, input)
		}
		if index > 0 && declared[index-1] == input {
			t.Fatalf("the admission projection names %s twice", input)
		}
	}
	// The central exact-output guard holds this snapshot to live go list results.
	// Here both directions hold the file to its own lowering companion's roots.
	linked, err := compilerinputs.InputsFor(lowering)
	if err != nil {
		t.Fatal(err)
	}
	needed := map[string]bool{"cmd/adamic": true}
	for _, input := range linked {
		needed[input] = true
	}
	if missing := uncovered(declared, needed); len(missing) > 0 {
		t.Fatalf("the admission projection misses what the lowering companion links: %s", strings.Join(missing, ", "))
	}
	for index, dropped := range declared {
		rest := append(append([]string(nil), declared[:index]...), declared[index+1:]...)
		if len(uncovered(rest, needed)) == 0 {
			t.Fatalf("dropping %s from the admission projection leaves no gap: it names nothing the lowering companion links", dropped)
		}
	}
	t.Logf("%d linked directories and module files, each under one of the admission projection's %d entries, each entry needed", len(needed), len(declared))
}

// uncovered returns the needed paths no declared path holds, sorted.
func uncovered(declared []string, needed map[string]bool) []string {
	var missing []string
	for path := range needed {
		covered := false
		for _, input := range declared {
			if path == input || strings.HasPrefix(path, input+"/") {
				covered = true
				break
			}
		}
		if !covered {
			missing = append(missing, path)
		}
	}
	sort.Strings(missing)
	return missing
}
