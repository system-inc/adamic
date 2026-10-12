package compilerinputs

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
)

// The compiler's packages: the loader, lowering, the native build and the JavaScript emission a port is lowered and
// built through.
var compiler = []string{"./internal/load", "./internal/lower", "./internal/native", "./internal/javascript"}

// Lowering is what the compiler links, no more: every in-repository package go list says the compiler's packages
// depend on, and each module file pinning one (with its go.sum, and go.work, which picks the modules), is under an
// entry; every entry is needed, so dropping any leaves a gap; and none holds cohere/TypeScript/tsc/testdata, the
// submodule's 48,000 baselines, so tsc whole can't come back. Mutants: an entry dropped (the gap is named), and
// cohere/TypeScript/tsc whole added (it holds the baselines).
func TestLoweringIsWhatTheCompilerLinks(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	// toolchain's go, under its environment, in the repository root with PWD set to it, so go.work resolves there.
	command := toolchain.CommandIn(root, "go", append([]string{"list", "-deps", "-f", "{{if not .Standard}}{{.Dir}}\t{{with .Module}}{{.GoMod}}{{end}}{{end}}"}, compiler...)...)
	var stderr bytes.Buffer
	command.Stderr = &stderr
	listing, err := command.Output()
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, stderr.Bytes())
	}
	inside := func(path string) (string, bool) {
		relative, err := filepath.Rel(root, path)
		return filepath.ToSlash(relative), err == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
	}
	needed := map[string]bool{"go.work": true}
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
