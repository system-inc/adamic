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
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			test.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			test.Fatal(err)
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
		if output, err := toolchain.GitCommandIn(root, arguments...).CombinedOutput(); err != nil {
			test.Fatalf("git %v: %v\n%s", arguments, err, output)
		}
	}
	// A runner's source: the tracked files and the manifest build-tree writes, with no .git.
	source := test.TempDir()
	for name, content := range files {
		path := filepath.Join(source, filepath.FromSlash(name))
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			test.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			test.Fatal(err)
		}
	}
	if err := tracked.Write(root, source); err != nil {
		test.Fatal(err)
	}
	want := append([]string{"fixture", "imports", "launches"}, Compiler...)
	sort.Strings(want)
	for _, directory := range []string{root, source} {
		dependencies, err := Dependencies(directory)
		if err != nil {
			test.Fatal(err)
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
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		test.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "cloud", "fast-gate", "compiler-dependencies.json"))
	if os.IsNotExist(err) {
		test.Skip("no hand-kept map to cover: the census is the map")
	}
	if err != nil {
		test.Fatal(err)
	}
	var declared struct {
		Packages map[string][]string `json:"packages"`
	}
	if err := json.Unmarshal(content, &declared); err != nil {
		test.Fatal(err)
	}
	dependencies, err := Dependencies(root)
	if err != nil {
		test.Fatal(err)
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
