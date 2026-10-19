package main

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"

	// The lowering companion's packages, linked so a change to anything they import selects this test wherever
	// selection follows Go's import graph: a cohere bump that adds a package to the closure then fails here.
	_ "github.com/system-inc/adamic/internal/load"
	_ "github.com/system-inc/adamic/internal/lower"
)

// The packages admission-lower runs: cmd/adamic's compile (with its check) calls the loader and the lowerer and stops,
// so the C emitter and the JavaScript backend it also links can't move an admission set.
var lowering = []string{"./internal/load", "./internal/lower"}

// inputs.txt is what the lowering companion links, no more: every in-repository package go list says internal/load and
// internal/lower depend on, cmd/adamic (whose compile and check the adapter calls), and each module file pinning one
// (with its go.sum, and go.work, which picks the modules), is under an entry; every entry is needed, so dropping any
// leaves a gap; and the list is sorted, once each. Mutants: internal/ir dropped (the gap is named), internal/native
// added (it names nothing the companion's lowering links).
func TestInputsAreWhatTheLoweringCompanionLinks(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	data, err := os.ReadFile("inputs.txt")
	if err != nil {
		t.Fatal(err)
	}
	declared := strings.Split(strings.TrimSuffix(string(data), "\n"), "\n")
	if !sort.StringsAreSorted(declared) {
		t.Fatal("inputs.txt is not sorted")
	}
	for index, input := range declared {
		if input == "" || strings.TrimSpace(input) != input || filepath.IsAbs(input) || filepath.ToSlash(filepath.Clean(input)) != input {
			t.Fatalf("inputs.txt line %d, %q, is not a clean repository path", index+1, input)
		}
		if index > 0 && declared[index-1] == input {
			t.Fatalf("inputs.txt names %s twice", input)
		}
	}
	// toolchain's go, in the repository root with PWD set to it, so go.work resolves there. Once for each system the
	// gate builds on, cgo on as a native build has it, and the union taken, so a build-constrained import still counts.
	var listing []byte
	for _, system := range []string{"linux", "darwin"} {
		command := toolchain.CommandIn(root, "go", append([]string{"list", "-deps", "-f", "{{if not .Standard}}{{.Dir}}\t{{with .Module}}{{.GoMod}}{{end}}{{end}}"}, lowering...)...)
		command.Env = append(command.Env, "GOOS="+system, "CGO_ENABLED=1")
		var stderr bytes.Buffer
		command.Stderr = &stderr
		output, err := command.Output()
		if err != nil {
			t.Fatalf("go list for %s: %v\n%s", system, err, stderr.Bytes())
		}
		listing = append(listing, output...)
	}
	inside := func(path string) (string, bool) {
		relative, err := filepath.Rel(root, path)
		return filepath.ToSlash(relative), err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
	}
	needed := map[string]bool{"go.work": true, "cmd/adamic": true}
	for _, line := range strings.Split(strings.TrimSpace(string(listing)), "\n") {
		directory, goMod, _ := strings.Cut(line, "\t")
		if relative, ok := inside(directory); ok {
			needed[relative] = true
		}
		if relative, ok := inside(goMod); ok && goMod != "" {
			needed[relative] = true
			if _, err := os.Stat(filepath.Join(filepath.Dir(goMod), "go.sum")); err == nil {
				needed[filepath.ToSlash(filepath.Join(filepath.Dir(relative), "go.sum"))] = true
			}
		}
	}
	if missing := uncovered(declared, needed); len(missing) > 0 {
		t.Fatalf("inputs.txt misses what the lowering companion links: %s", strings.Join(missing, ", "))
	}
	for index, dropped := range declared {
		rest := append(append([]string(nil), declared[:index]...), declared[index+1:]...)
		if len(uncovered(rest, needed)) == 0 {
			t.Fatalf("dropping %s from inputs.txt leaves no gap: it names nothing the lowering companion links", dropped)
		}
	}
	t.Logf("%d linked directories and module files, each under one of inputs.txt's %d entries, each entry needed", len(needed), len(declared))
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
