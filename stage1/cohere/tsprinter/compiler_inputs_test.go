package tsprinter

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/toolchain"
)

// tsprinterCompilerInputs are the repository paths the compiler a port is lowered and built by links: every
// in-repository package internal/load, internal/lower and internal/native depend on, cohere/TypeScript/tsc's Go port
// among them, and each module file that pins one. That code is compiled into this test binary, so no trace of a build
// can see it: the key has to name it. TestMutantNativeInputsCoverWhatItReads keeps the list complete, and every entry
// needed, as markdownblocks' TestLoweringInputsCoverCompilerDependencies keeps its loweringCompilerInputs.
var tsprinterCompilerInputs = []string{
	"internal", "go.mod", "go.work", "bridge/tsgo", "cohere/go.mod", "cohere/go.sum", "cohere/TypeScript/tsc", "cohere/TypeScript-shim",
	"cohere/internal/lint/checking", "cohere/internal/lint/ecmascript/binding", "cohere/internal/lint/ecmascript/comments",
	"cohere/internal/lint/ecmascript/control_flow_graph", "cohere/internal/lint/ecmascript/imports", "cohere/internal/lint/ecmascript/jsx",
	"cohere/internal/lint/ecmascript/module", "cohere/internal/lint/ecmascript/module_roots", "cohere/internal/lint/ecmascript/nextjs",
	"cohere/internal/lint/ecmascript/property", "cohere/internal/lint/ecmascript/react", "cohere/internal/lint/ecmascript/reference",
	"cohere/internal/lint/ecmascript/regexsyntax", "cohere/internal/lint/ecmascript/scope", "cohere/internal/lint/ecmascript/text",
	"cohere/internal/lint/linter", "cohere/internal/lint/rule", "cohere/internal/lint/rules/adamic", "cohere/internal/lint/rules/nexus",
	"cohere/internal/types/sourcename", "cohere/policy", "cohere/rule_runner",
}

// mutantNativeRecipe are this package's files whose code makes a mutant native: the mutations and the port they
// plant into, the product and its key, lowering, and the build wrapper.
var mutantNativeRecipe = []string{"mutants_test.go", "mutant_products_test.go", "compiler_inputs_test.go", "helpers_test.go", "expressions_test.go"}

// mutantNativeBuildFiles are what a mutant native's lowering and build read, relative to the repository: the compiler
// (tsprinterCompilerInputs), the port's TypeScript, the parser's and the scanner's TypeScript it imports, and the
// recipe. Not cohere whole: its TypeScript submodule is 67,000 files no lowering reads, which a populated tree walked
// for every lookup (#4ycgm2n), while stage1/typescript/scanner, which the parser imports, was missing.
func mutantNativeBuildFiles(t *testing.T) []string {
	t.Helper()
	files := append([]string(nil), tsprinterCompilerInputs...)
	for _, directory := range []string{"stage1/cohere/tsprinter", "stage1/typescript/parser", "stage1/typescript/scanner"} {
		entries, err := os.ReadDir(filepath.Join(repository, directory))
		if err != nil {
			t.Fatal(err)
		}
		for _, entry := range entries {
			if !entry.IsDir() && strings.HasSuffix(entry.Name(), ".ts") {
				files = append(files, directory+"/"+entry.Name())
			}
		}
	}
	for _, file := range mutantNativeRecipe {
		files = append(files, "stage1/cohere/tsprinter/"+file)
	}
	return files
}

// A mutant native's key names everything its build reads. The compiler's side: every in-repository package
// internal/load, internal/lower and internal/native link, and each module file pinning one, is under
// tsprinterCompilerInputs, and each of its entries is needed. The TypeScript side: every source file the loader reads
// for each entry a mutation plants into is in mutantNativeBuildFiles. Mutants: stage1/typescript/scanner dropped from
// the build files (the scanner the parser imports goes unkeyed); a compiler input dropped (a gap the guard names).
func TestMutantNativeInputsCoverWhatItReads(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	// toolchain's go, under its environment, in the repository root with PWD set to it, so go.work resolves there.
	command := toolchain.CommandIn(root, "go", "list", "-deps", "-f", "{{if not .Standard}}{{.Dir}}\t{{with .Module}}{{.GoMod}}{{end}}{{end}}", "./internal/load", "./internal/lower", "./internal/native")
	listing, err := combinedOutput(command)
	if err != nil {
		t.Fatalf("go list: %v\n%s", err, listing)
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
	declared := mutantNativeBuildFiles(t)
	if gaps := compilerInputGaps(declared, needed); len(gaps) > 0 {
		t.Fatalf("mutant natives' key misses compiler inputs: %s", strings.Join(gaps, ", "))
	}
	for index, dropped := range tsprinterCompilerInputs {
		planted := append(append([]string(nil), tsprinterCompilerInputs[:index]...), tsprinterCompilerInputs[index+1:]...)
		if len(compilerInputGaps(planted, needed)) == 0 {
			t.Fatalf("dropping %s leaves no gap: it keys nothing the compiler links", dropped)
		}
	}
	read := map[string]bool{}
	for _, change := range mutations {
		if read[change.entry] {
			continue
		}
		read[change.entry] = true
		program, err := load.Load([]string{filepath.Join(root, "stage1/cohere/tsprinter", change.entry)})
		if err != nil {
			t.Fatalf("%s: %v", change.entry, err)
		}
		// Every source the checked program holds, its imports' imports included. A bundled library (bundled:///libs/...)
		// and the prelude (/adamic-prelude, internal/load's go:embed) are compiled into the binary, so the compiler inputs
		// key them; any other must be a repository file the key names.
		sources, compiledIn := map[string]bool{}, 0
		for _, file := range program.CompilerProgram().GetSourceFiles() {
			name := string(file.FileName())
			if strings.HasPrefix(name, "bundled://") || strings.HasPrefix(name, "/adamic-prelude/") {
				compiledIn++
				continue
			}
			relative, ok := inside(name)
			if !ok {
				t.Fatalf("%s reads %s, outside the repository", change.entry, name)
			}
			sources[relative] = true
		}
		if gaps := compilerInputGaps(declared, sources); len(gaps) > 0 {
			t.Fatalf("lowering %s reads files the mutant natives' key misses: %s", change.entry, strings.Join(gaps, ", "))
		}
		t.Logf("%s: %d repository source files, each keyed, and %d compiled in", change.entry, len(sources), compiledIn)
	}
	t.Logf("%d compiler directories and module files covered; each of %d declared compiler inputs needed", len(needed), len(tsprinterCompilerInputs))
}

// A mutant native's key moves with the scanner the parser imports and stands still for cohere's TypeScript submodule,
// which no lowering reads: a stand-in tree holds every declared input, the scanner's first source is edited by one
// byte, then a submodule file. Keying builds nothing. Mutant: stage1/typescript/scanner dropped from
// mutantNativeBuildFiles (the scanner edit keeps the key, as it did under the old cohere-wide key).
func TestMutantNativeKeyMovesWithTheScanner(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	write := func(name, content string) {
		t.Helper()
		path := filepath.Join(root, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	declared := mutantNativeBuildFiles(t)
	scanner := ""
	for _, name := range declared {
		if filepath.Ext(name) == "" {
			// A declared directory: one source inside it stands for the package.
			write(name+"/source.go", name)
			continue
		}
		write(name, name)
		if scanner == "" && strings.HasPrefix(name, "stage1/typescript/scanner/") {
			scanner = name
		}
	}
	if scanner == "" {
		t.Fatal("mutant natives' key names no scanner source, though the parser imports it")
	}
	write("cohere/TypeScript/README.md", "submodule")
	key := func() string {
		t.Helper()
		key, err := buildcache.Key(root, buildcache.Inputs{Name: "mutant native probe", Files: declared})
		if err != nil {
			t.Fatal(err)
		}
		return key
	}
	before := key()
	write("cohere/TypeScript/README.md", "submodule, edited")
	if key() != before {
		t.Fatal("an edit under cohere/TypeScript, which no lowering reads, moved the key")
	}
	write(scanner, scanner+"!")
	if key() == before {
		t.Fatalf("a one-byte edit to %s kept the key", scanner)
	}
	t.Logf("%s moves the key; cohere/TypeScript doesn't", scanner)
}

// compilerInputGaps returns the needed paths no declared path holds, sorted.
func compilerInputGaps(declared []string, needed map[string]bool) []string {
	var gaps []string
	for path := range needed {
		covered := false
		for _, file := range declared {
			if path == file || strings.HasPrefix(path, file+"/") {
				covered = true
				break
			}
		}
		if !covered {
			gaps = append(gaps, path)
		}
	}
	sort.Strings(gaps)
	return gaps
}
