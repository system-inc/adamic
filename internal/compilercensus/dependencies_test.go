package compilercensus

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/internal/tracked"
)

// The census on a stand-in repository: a package that imports the compiler, one that launches a process, one whose only
// marker is in a testdata program, one naming cohere's cmd/adamic_path harness, and a clean one. The first three and
// the compiler's own packages depend on Compiler; the last two don't. The census is the same read from a copy with no
// .git, through its manifest, as a Loom runner's source is. Mutants: cmd/adamic unbounded (adamic_path counts);
// testdata skipped (its owner is missed); launching not counted (the launcher is missed); git's listing alone (the copy
// fails).
func TestDependenciesFollowTheMarkers(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	files := map[string]string{
		"go.mod":                          "module example.com/standin\n\ngo 1.24\n",
		"imports/imports.go":              "package imports\n\nconst compiler = \"github.com/system-inc/adamic/internal/lower\"\n",
		"launches/launches.go":            "package launches\n\nimport _ \"os/exec\"\n",
		"fixture/fixture.go":              "package fixture\n",
		"fixture/testdata/program/run.go": "package main\n\nconst driver = \"./cmd/adamic\"\n",
		"harness/harness.go":              "package harness\n\nconst command = \"cmd/adamic_path\"\n",
		"clean/clean.go":                  "package clean\n",
	}
	for _, directory := range Compiler {
		files[directory+"/package.go"] = "package " + filepath.Base(directory) + "\n"
	}
	for name, content := range files {
		path := filepath.Join(root, filepath.FromSlash(name))
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0o755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(path, []byte(content), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	for _, arguments := range [][]string{{"init", "-q"}, {"add", "."}, {
		"-c",
		"user.name=census",
		"-c",
		"user.email=census@example.com",
		"-c",
		"commit.gpgsign=false",
		"commit",
		"-qm",
		"stand-in",
	}} {
		if output, combinedOutputError := toolchain.GitCommandIn(
			root,
			arguments...,
		).CombinedOutput(); combinedOutputError != nil {
			test.Fatalf("git %v: %v\n%s", arguments, combinedOutputError, output)
		}
	}
	// A runner's source: the tracked files and the manifest build-tree writes, with no .git.
	source := test.TempDir()
	for name, content := range files {
		path := filepath.Join(source, filepath.FromSlash(name))
		if createDirectoryError := os.MkdirAll(filepath.Dir(path), 0o755); createDirectoryError != nil {
			test.Fatal(createDirectoryError)
		}
		if writeError := os.WriteFile(path, []byte(content), 0o644); writeError != nil {
			test.Fatal(writeError)
		}
	}
	if writeError := tracked.Write(root, source); writeError != nil {
		test.Fatal(writeError)
	}
	want := append([]string{"fixture", "imports", "launches"}, Compiler...)
	sort.Strings(want)
	for _, directory := range []string{root, source} {
		dependencies, dependenciesError := Dependencies(directory)
		if dependenciesError != nil {
			test.Fatal(dependenciesError)
		}
		var got []string
		for consumer, inputs := range dependencies {
			got = append(got, consumer)
			if !slices.Equal(inputs, Compiler) {
				test.Errorf("%s depends on %v, want the whole compiler", consumer, inputs)
			}
		}
		sort.Strings(got)
		if !slices.Equal(got, want) {
			test.Fatalf("consumers in %s: %v, want %v", directory, got, want)
		}
	}
}

// Every package the hand-kept map declares is one the census derives, so the readers can take the census in its place
// and select nothing less (#ng0ksed). Mutant: drop the launch evidence (cmd/adamic-gocacheprog, internal/tracked and
// internal/buildcache/cmd/traced, each declared, are missed).
func TestDependenciesCoverTheDeclaredMap(test *testing.T) {
	test.Parallel()
	root, absolutePathError := filepath.Abs(filepath.Join("..", ".."))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	content, readError := os.ReadFile(filepath.Join(root, "cloud", "fast-gate", "compiler-dependencies.json"))
	if os.IsNotExist(readError) {
		test.Skip("no hand-kept map to cover: the census is the map")
	}
	if readError != nil {
		test.Fatal(readError)
	}
	var declared struct {
		Packages map[string][]string `json:"packages"`
	}
	if decodeError := json.Unmarshal(content, &declared); decodeError != nil {
		test.Fatal(decodeError)
	}
	dependencies, dependenciesError := Dependencies(root)
	if dependenciesError != nil {
		test.Fatal(dependenciesError)
	}
	var missed []string
	for directory := range declared.Packages {
		if _, derived := dependencies[directory]; !derived {
			missed = append(missed, directory)
		}
	}
	sort.Strings(missed)
	if len(missed) > 0 {
		test.Fatalf("the census misses declared compiler consumers: %s", strings.Join(missed, ", "))
	}
	test.Logf("%d declared, each derived among the census's %d", len(declared.Packages), len(dependencies))
}
