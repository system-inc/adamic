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
)

// The census on a stand-in repository: a package that imports the compiler, one that launches a process, one whose only
// marker is in a testdata program, one naming cohere's cmd/adamic_path harness, and a clean one. The first three and the
// compiler's own packages depend on Compiler; the last two don't. Mutants: cmd/adamic unbounded (adamic_path counts);
// testdata skipped (its owner is missed); launching not counted (the launcher is missed).
func TestDependenciesFollowTheMarkers(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
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
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}
	for _, arguments := range [][]string{{"init", "-q"}, {"add", "."}} {
		if output, err := toolchain.GitCommandIn(root, arguments...).CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", arguments, err, output)
		}
	}
	dependencies, err := Dependencies(root)
	if err != nil {
		t.Fatal(err)
	}
	want := append([]string{"fixture", "imports", "launches"}, Compiler...)
	sort.Strings(want)
	var got []string
	for directory, inputs := range dependencies {
		got = append(got, directory)
		if !slices.Equal(inputs, Compiler) {
			t.Errorf("%s depends on %v, want the whole compiler", directory, inputs)
		}
	}
	sort.Strings(got)
	if !slices.Equal(got, want) {
		t.Fatalf("consumers %v, want %v", got, want)
	}
}

// Every package the hand-kept map declares is one the census derives, so the readers can take the census in its place
// and select nothing less (#ng0ksed). Mutant: drop the launch evidence (cmd/adamic-gocacheprog, internal/tracked and
// internal/buildcache/cmd/traced, each declared, are missed).
func TestDependenciesCoverTheDeclaredMap(t *testing.T) {
	t.Parallel()
	root, err := filepath.Abs(filepath.Join("..", ".."))
	if err != nil {
		t.Fatal(err)
	}
	content, err := os.ReadFile(filepath.Join(root, "cloud", "fast-gate", "compiler-dependencies.json"))
	if os.IsNotExist(err) {
		t.Skip("no hand-kept map to cover: the census is the map")
	}
	if err != nil {
		t.Fatal(err)
	}
	var declared struct {
		Packages map[string][]string `json:"packages"`
	}
	if err := json.Unmarshal(content, &declared); err != nil {
		t.Fatal(err)
	}
	dependencies, err := Dependencies(root)
	if err != nil {
		t.Fatal(err)
	}
	var missed []string
	for directory := range declared.Packages {
		if _, derived := dependencies[directory]; !derived {
			missed = append(missed, directory)
		}
	}
	sort.Strings(missed)
	if len(missed) > 0 {
		t.Fatalf("the census misses declared compiler consumers: %s", strings.Join(missed, ", "))
	}
	t.Logf("%d declared, each derived among the census's %d", len(declared.Packages), len(dependencies))
}
