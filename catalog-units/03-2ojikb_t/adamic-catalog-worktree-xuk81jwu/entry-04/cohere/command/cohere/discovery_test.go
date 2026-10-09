package main

import (
	"os"
	"path/filepath"
	"reflect"
	"testing"
)

const (
	discoveryTsconfig = `{"compilerOptions":{"strict":true,"noEmit":true},"include":["**/*.ts"]}`
	discoveryPackage  = "// swift-tools-version:6.0\nimport PackageDescription\nlet package = Package(name: \"Toy\")\n"
)

// newRepository makes a directory that is a repository's root, by the .git directory discovery reads.
func newRepository(t *testing.T, files map[string]string) string {
	t.Helper()
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	writeTree(t, root, files)
	return root
}

// Every tsconfig.json and Package.swift the repository keeps is a project, nested at any depth, and a
// directory holding both is two (#f9nftxz). Parents come before children and TypeScript before Swift.
func TestDiscoveryFindsEveryProjectTheRepositoryKeeps(t *testing.T) {
	t.Parallel()
	root := newRepository(t, map[string]string{
		"tsconfig.json":               discoveryTsconfig,
		"Package.swift":               discoveryPackage,
		"apps/web/tsconfig.json":      discoveryTsconfig,
		"packages/a/b/Package.swift":  discoveryPackage,
		"packages/a-b/tsconfig.json":  discoveryTsconfig,
		"packages/a/b/Sources/x.txt":  "not a marker\n",
		"apps/web/tsconfig.base.json": discoveryTsconfig,
	})
	found, err := discoverProjects(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	want := []discoveredProject{
		{Directory: ".", Engine: engineTypeScript},
		{Directory: ".", Engine: engineSwift},
		{Directory: "apps/web", Engine: engineTypeScript},
		{Directory: "packages/a/b", Engine: engineSwift},
		{Directory: "packages/a-b", Engine: engineTypeScript},
	}
	if !reflect.DeepEqual(found.Projects, want) {
		t.Fatalf("found %+v, want %+v", found.Projects, want)
	}
}

// What discovery must never enter, each with a marker planted inside it: dependencies, build output,
// caches, fixtures, what .gitignore ignores, a repository of its own, a submodule, and a marker the root's
// ignorePatterns refuse. Positive control: the same tree walked with every refusal lifted finds every
// planted marker, so the pass is the refusals at work and not markers that were never there.
// Not parallel: it swaps the package-level neverDescended and newIgnoreScope, which discoverProjects reads
func TestDiscoveryNeverEntersWhatItMustNot(t *testing.T) {
	root := newRepository(t, map[string]string{
		"Package.swift":                         discoveryPackage,
		".gitignore":                            "generated/\n",
		"node_modules/dependency/tsconfig.json": discoveryTsconfig,
		".build/checkouts/dep/Package.swift":    discoveryPackage,
		".cache/cohere/tsconfig.json":           discoveryTsconfig,
		"internal/testdata/tsconfig.json":       discoveryTsconfig,
		"generated/tsconfig.json":               discoveryTsconfig,
		"apps/.gitignore":                       "legacy/\n",
		"apps/legacy/tsconfig.json":             discoveryTsconfig,
		"clone/Package.swift":                   discoveryPackage,
		"submodule/tsconfig.json":               discoveryTsconfig,
		"submodule/.git":                        "gitdir: ../.git/modules/submodule\n",
		"refused/tsconfig.json":                 discoveryTsconfig,
	})
	if err := os.Mkdir(filepath.Join(root, "clone", ".git"), 0o755); err != nil {
		t.Fatal(err)
	}

	found, err := discoverProjects(root, []string{"refused/**"})
	if err != nil {
		t.Fatal(err)
	}
	if want := []discoveredProject{{Directory: ".", Engine: engineSwift}}; !reflect.DeepEqual(found.Projects, want) {
		t.Fatalf("found %+v, want only the root's package", found.Projects)
	}
	// node_modules, .build, .cache and testdata; the root's own .git is passed over without counting.
	if found.NeverDescended != 4 || found.Ignored != 2 || found.Submodules != 1 {
		t.Errorf("never descended %d (want 4), ignored %d (want 2: one by the root .gitignore, one by apps/.gitignore), submodules %d (want 1)", found.NeverDescended, found.Ignored, found.Submodules)
	}
	if want := []string{"clone"}; !reflect.DeepEqual(found.NestedRepositories, want) {
		t.Errorf("nested repositories %v, want %v", found.NestedRepositories, want)
	}
	if want := []string{"refused/tsconfig.json"}; !reflect.DeepEqual(found.Refused, want) {
		t.Errorf("refused %v, want %v", found.Refused, want)
	}

	// The control: with nothing never-descended and nothing ignored, the planted markers are found.
	previousNames, previousScope := neverDescended, newIgnoreScope
	t.Cleanup(func() { neverDescended, newIgnoreScope = previousNames, previousScope })
	neverDescended = map[string]bool{".git": true}
	newIgnoreScope = func(string) (ignoreScope, error) { return ignoresNothing{}, nil }
	lifted, err := discoverProjects(root, nil)
	if err != nil {
		t.Fatal(err)
	}
	directories := map[string]bool{}
	for _, project := range lifted.Projects {
		directories[project.Directory] = true
	}
	for _, planted := range []string{"node_modules/dependency", ".build/checkouts/dep", ".cache/cohere", "internal/testdata", "generated", "apps/legacy", "refused"} {
		if !directories[planted] {
			t.Errorf("with the refusals lifted, %s was still not found, so its refusal above proves nothing", planted)
		}
	}
}

// ignoresNothing is a scope that ignores nothing, for the control above.
type ignoresNothing struct{}

func (ignoresNothing) Enter(string) (ignoreScope, error) { return ignoresNothing{}, nil }
func (ignoresNothing) Ignored(string, bool) bool         { return false }

// Discovery starts where the walk up found a project, unless the caller stands in a repository with no
// project at or above them inside it: then the repository's root, which is how cohere's own repository
// finds swift/. A project above the repository belongs to some other tree.
func TestDiscoveryRootIsTheRepositoryTheCallerStandsIn(t *testing.T) {
	t.Parallel()
	repository := newRepository(t, map[string]string{
		"swift/Package.swift": discoveryPackage,
		"docs/readme.md":      "docs\n",
	})
	start := filepath.Join(repository, "docs")
	location, locateError := locateProject(locationRequest{WorkingDirectory: start, ConfigFileName: projectMarker})
	if locateError == nil {
		t.Fatalf("the walk up found %s, so this does not test a repository with no project above the caller", location.Root)
	}
	if root := discoveryRoot(location, locateError, start); root != repository {
		t.Errorf("discovery starts at %s, want the repository %s", root, repository)
	}

	outer := t.TempDir()
	writeTree(t, outer, map[string]string{"tsconfig.json": discoveryTsconfig, "inner/swift/Package.swift": discoveryPackage})
	inner := filepath.Join(outer, "inner")
	if err := os.Mkdir(filepath.Join(inner, ".git"), 0o755); err != nil {
		t.Fatal(err)
	}
	location, locateError = locateProject(locationRequest{WorkingDirectory: inner, ConfigFileName: projectMarker})
	if locateError != nil || location.Root != outer {
		t.Fatalf("the walk up should find the outer project, got %s, %v", location.Root, locateError)
	}
	if root := discoveryRoot(location, locateError, inner); root != inner {
		t.Errorf("discovery starts at %s, want the repository %s rather than the project above it", root, inner)
	}

	project := t.TempDir()
	writeTree(t, project, map[string]string{"tsconfig.json": discoveryTsconfig, "sub/file.ts": "export {};\n"})
	start = filepath.Join(project, "sub")
	location, locateError = locateProject(locationRequest{WorkingDirectory: start, ConfigFileName: projectMarker})
	if root := discoveryRoot(location, locateError, start); root != project {
		t.Errorf("discovery starts at %s, want the project the walk up found, %s", root, project)
	}
}

// A run that names one project, by directory, tsconfig or path, checks that project alone, as it always
// did, and so do the listings and explanations; a bare run discovers, and so does one naming a lint config,
// which says which rules apply and not which project (#wvgxtey).
func TestDiscoveryAppliesOnlyToARunThatNamesNoProject(t *testing.T) {
	t.Parallel()
	if !discoveryApplies(map[string]bool{"no-fix": true, "format-all": true}, nil) {
		t.Error("a bare --no-fix --format-all run does not discover")
	}
	if !discoveryApplies(map[string]bool{"lint-config": true}, nil) {
		t.Error("a run naming --lint-config does not discover")
	}
	for _, named := range []string{"directory", "tsconfig", "rules", "print-config", "explain", "stdin-filepath", "version"} {
		if discoveryApplies(map[string]bool{named: true}, nil) {
			t.Errorf("a run naming --%s discovers", named)
		}
	}
	if discoveryApplies(map[string]bool{}, []string{"src/index.ts"}) {
		t.Error("a run naming a path discovers")
	}
}
