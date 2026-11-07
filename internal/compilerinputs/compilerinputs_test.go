package compilerinputs

import (
	"bytes"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"

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
// submodule's 48,000 baselines, so tsc whole can't come back. Mutants: an entry dropped (the gap is named), and
// cohere/TypeScript/tsc whole added (it holds the baselines).
func TestLoweringIsWhatTheCompilerLinks(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(filepath.Join("..", ".."))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	// toolchain's go, under its environment, in the repository root with PWD set to it, so go.work resolves there. Once
	// for each system the gate builds on, cgo on as a native build has it, and the union taken: a package only one of
	// them links (a build-constrained import) is still needed, wherever this test runs. On Oct 12 the two differ only
	// outside the repository (linux links golang.org/x/sys/unix too), so the union guards a difference to come.
	var listing []byte
	for _, system := range []string{"linux", "darwin"} {
		command := toolchain.CommandIn(root, "go", append(
			[]string{"list", "-deps", "-f", "{{if not .Standard}}{{.Dir}}\t{{with .Module}}{{.GoMod}}{{end}}{{end}}"},
			compiler...,
		)...)
		command.Env = append(command.Env, "GOOS="+system, "CGO_ENABLED=1")
		var stderr bytes.Buffer
		command.Stderr = &stderr
		output, outputError := command.Output()
		if outputError != nil {
			test.Fatalf("go list for %s: %v\n%s", system, outputError, stderr.Bytes())
		}
		listing = append(listing, output...)
	}
	inside := func(path string) (string, bool) {
		relative, relativePathError := filepath.Rel(root, path)
		return filepath.ToSlash(
			relative,
		), relativePathError == nil && relative != ".." && !strings.HasPrefix(relative, ".."+string(filepath.Separator))
	}
	needed := map[string]bool{"go.work": true}
	for _, line := range strings.Split(strings.TrimSpace(string(listing)), "\n") {
		directory, goMod, _ := strings.Cut(line, "\t")
		if relative, ok := inside(directory); ok {
			needed[relative] = true
		}
		if relative, ok := inside(goMod); ok && goMod != "" {
			needed[relative] = true
			if _, statError := os.Stat(filepath.Join(filepath.Dir(goMod), "go.sum")); statError == nil {
				needed[filepath.ToSlash(filepath.Join(filepath.Dir(relative), "go.sum"))] = true
			}
		}
	}
	declared := Lowering()
	if missing := gaps(declared, needed); len(missing) > 0 {
		test.Fatalf("Lowering misses what the compiler links: %s", strings.Join(missing, ", "))
	}
	for _, input := range declared {
		if baselines := "cohere/TypeScript/tsc/testdata"; baselines == input ||
			strings.HasPrefix(baselines, input+"/") {
			test.Fatalf("Lowering's %s holds %s, which no compiler links", input, baselines)
		}
	}
	for index, dropped := range declared {
		rest := append(append([]string(nil), declared[:index]...), declared[index+1:]...)
		if len(gaps(rest, needed)) == 0 {
			test.Fatalf("dropping %s from Lowering leaves no gap: it names nothing the compiler links", dropped)
		}
	}
	test.Logf(
		"%d linked directories and module files, each under one of Lowering's %d entries, each entry needed",
		len(needed),
		len(declared),
	)
}

// A caller's appends to Lowering never reach the list another caller gets.
func TestLoweringIsACopy(test *testing.T) {
	test.Parallel()
	first := Lowering()
	first[0] = "changed"
	if Lowering()[0] == "changed" {
		test.Fatal("a change to one caller's Lowering reached the next caller's")
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
