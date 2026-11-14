package compilerinputs

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	// The compiler's packages, linked so a change to anything they import selects this test wherever selection follows
	// Go's import graph: a cohere bump that adds a package to the link set then fails here, not at the next full run.
	_ "github.com/system-inc/adamic/internal/javascript"
	_ "github.com/system-inc/adamic/internal/load"
	_ "github.com/system-inc/adamic/internal/lower"
	_ "github.com/system-inc/adamic/internal/native"
)

// The compiler's packages: the loader, lowering, the native build and the JavaScript emission a port is lowered and
// built through.
var compiler = []string{"./internal/load", "./internal/lower", "./internal/native", "./internal/javascript"}

// Lowering is what the compiler links, no more: every in-repository package go list says the compiler's packages
// depend on, and each module file pinning one (with its go.sum, and go.work, which picks the modules), is under an
// entry; every entry is needed, so dropping any leaves a gap; and none holds cohere/TypeScript/tsc/testdata, the
// submodule's 48,000 baselines, so tsc whole can't come back. The committed table must equal the generator's output,
// so a hand edit (an entry dropped, added or reordered) reds here first. Mutants: a link-only cohere import (the table
// is stale until regenerated, and then every consumer's guard passes with it), and cohere/TypeScript/tsc whole added
// (it holds the baselines).
func TestLoweringIsWhatTheCompilerLinks(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	table, _, err := Generate(root)
	if err != nil {
		t.Fatal(err)
	}
	committed, err := os.ReadFile("compilerinputs_table.go")
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(committed, table) {
		t.Fatal("compiler input table is stale; run go run ./internal/compilerinputs/generate")
	}
	linked, err := Derive(root, compiler)
	if err != nil {
		t.Fatal(err)
	}
	needed := map[string]bool{}
	for _, input := range linked {
		needed[input] = true
	}
	declared := Lowering()
	if missing := gaps(declared, needed); len(missing) > 0 {
		t.Fatalf("Lowering misses what the compiler links: %s", strings.Join(missing, ", "))
	}
	for _, input := range declared {
		if baselines := "cohere/TypeScript/tsc/testdata"; baselines == input || strings.HasPrefix(baselines, input+"/") {
			t.Fatalf("Lowering's %s holds %s, which no compiler links", input, baselines)
		}
	}
	for index, dropped := range declared {
		rest := append(append([]string(nil), declared[:index]...), declared[index+1:]...)
		if len(gaps(rest, needed)) == 0 {
			t.Fatalf("dropping %s from Lowering leaves no gap: it names nothing the compiler links", dropped)
		}
	}
	t.Logf("%d linked directories and module files, each under one of Lowering's %d entries, each entry needed", len(needed), len(declared))
}

// A caller's appends to Lowering never reach the list another caller gets.
func TestLoweringIsACopy(t *testing.T) {
	t.Parallel()
	first := Lowering()
	first[0] = "changed"
	if Lowering()[0] == "changed" {
		t.Fatal("a change to one caller's Lowering reached the next caller's")
	}
}

// gaps returns the needed paths no declared path holds, sorted.
func gaps(declared []string, needed map[string]bool) []string {
	var missing []string
	for path := range needed {
		covered := false
		for _, file := range declared {
			if path == file || strings.HasPrefix(path, file+"/") {
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

// Each consumer receives its own slice, including when roots are reordered.
// Mutating one key's input list must never change another key's inputs.
func TestGeneratedInputsAreCopies(t *testing.T) {
	t.Parallel()
	for _, projection := range []func() []string{Markdownblocks, Admission} {
		first := projection()
		original := first[0]
		first[0] = "changed"
		if projection()[0] != original {
			t.Fatal("a caller changed a generated projection")
		}
	}
	roots := []string{"./internal/native", "./internal/lower", "./internal/load"}
	first, err := InputsFor(roots)
	if err != nil {
		t.Fatal(err)
	}
	original := first[0]
	first[0] = "changed"
	next, err := InputsFor([]string{"./internal/load", "./internal/lower", "./internal/native"})
	if err != nil {
		t.Fatal(err)
	}
	if next[0] != original {
		t.Fatal("a caller changed a generated dependency set")
	}
	if _, err := InputsFor([]string{"./not-generated"}); err == nil {
		t.Fatal("unknown roots borrowed another root set")
	}
}
