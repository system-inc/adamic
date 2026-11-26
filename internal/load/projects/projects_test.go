package main

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"
)

// repository is the repository's root, from this package's directory. The projects' include directories are walked
// from it, and the two tsconfigs read there.
//
// Reads: tsconfig.json
// Reads: internal/load/projects/node/tsconfig.json
// Reads: dedication/*
// Reads: bench/*
// Reads: stage1/*
// Reads: internal/load/testdata/*
// Reads: internal/oracle/testdata/*
const repository = "../../.."

// The two tsconfigs are exactly what `go run ./internal/load/projects` writes from load's own predicate, asked of each
// source's first program: a source added under the include directories, moved, or given an import that changes whether
// Load gives it Node's declarations, with no rerun of the generator, fails this, naming each file that moved. Mutant: one
// source deleted from the Node project's files and the root's exclude, and this fails naming it.
func TestTheRootProjectsAreLoadsOwn(t *testing.T) {
	t.Parallel()
	placed, err := place(repository)
	if err != nil {
		t.Fatal(err)
	}
	rendered, err := render(placed.node)
	if err != nil {
		t.Fatal(err)
	}
	for _, name := range []string{rootFile, nodeFile} {
		committed, err := os.ReadFile(filepath.Join(repository, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(committed, rendered[name]) {
			// A head cut before another head added, moved or re-imported a source lands here through no fault of
			// its own, so the message names that race and the re-cut that ends it (Compiler's (c) on #kyq0pyn),
			// until Loom regenerates the projects on the candidate itself.
			t.Errorf("%s is stale against Load: sources under the include directories were added, moved or given other imports since it was generated, possibly by another head that landed after this one was cut. Re-cut on main's tip and run `go run ./internal/load/projects` from the repository's root. Only in the committed file:\n%s\nOnly in what the generator writes:\n%s",
				name, strings.Join(linesOnlyIn(committed, rendered[name]), "\n"), strings.Join(linesOnlyIn(rendered[name], committed), "\n"))
		}
	}
	for _, refusal := range placed.refused {
		t.Logf("standalone, since Load builds it no program: %s", refusal)
	}
	t.Logf("%d sources check under Node's declarations", len(placed.node))
}

// linesOnlyIn is each line of one file that the other doesn't hold, in its order.
func linesOnlyIn(one, other []byte) []string {
	held := map[string]bool{}
	for _, line := range strings.Split(string(other), "\n") {
		held[line] = true
	}
	only := []string{}
	for _, line := range strings.Split(string(one), "\n") {
		if !held[line] {
			only = append(only, line)
		}
	}
	return only
}

// committedProjects is the two tsconfigs as the repository holds them, the Node project's files by path from the
// repository's root.
func committedProjects(t *testing.T) (project, project) {
	t.Helper()
	read := func(name string) project {
		data, err := os.ReadFile(filepath.Join(repository, name))
		if err != nil {
			t.Fatal(err)
		}
		var read project
		if err := json.Unmarshal(data, &read); err != nil {
			t.Fatalf("%s: %v", name, err)
		}
		return read
	}
	root, nodeProject := read(rootFile), read(nodeFile)
	for index, file := range nodeProject.Files {
		if !strings.HasPrefix(file, nodeToRoot) {
			t.Fatalf("%s: %s doesn't start from the root, %s", nodeFile, file, nodeToRoot)
		}
		nodeProject.Files[index] = strings.TrimPrefix(file, nodeToRoot)
	}
	return root, nodeProject
}

// Each Adamic source under the include directories belongs to exactly one project, so cohere, which gives a file to
// the nearest tsconfig that includes it, never checks one twice or not at all: the Node project names it in its files,
// or the root includes it by directory and doesn't exclude it by name, never both and never neither. Both projects
// root the shared prelude.d.ts, as Load roots it in every program, and only the Node project its Node console. Mutant:
// one Node source dropped from the root's exclude, and this fails naming it as in both.
func TestEachSourceHasOneProject(t *testing.T) {
	t.Parallel()
	all, err := sources(repository)
	if err != nil {
		t.Fatal(err)
	}
	root, nodeProject := committedProjects(t)
	inNode := map[string]bool{}
	for _, file := range nodeProject.Files {
		inNode[file] = true
	}
	excluded := map[string]bool{}
	for _, file := range root.Exclude {
		excluded[file] = true
	}
	isSource := map[string]bool{}
	for _, source := range all {
		isSource[source] = true
		switch {
		case inNode[source] && !excluded[source]:
			t.Errorf("%s is in both projects: %s names it and %s includes it", source, nodeFile, rootFile)
		case !inNode[source] && excluded[source]:
			t.Errorf("%s is in neither project: %s excludes it and %s doesn't name it", source, rootFile, nodeFile)
		}
	}
	for _, file := range nodeProject.Files {
		if file != preludeNodeFile && file != preludeFile && !isSource[file] {
			t.Errorf("%s names %s, which isn't a source under %s's include directories", nodeFile, file, rootFile)
		}
	}
	t.Logf("%d sources: %d in %s, %d in %s", len(all), len(nodeProject.Files)-2, nodeFile, len(all)-len(nodeProject.Files)+2, rootFile)
}

// expectedDiagnostic is one row of expected_diagnostics.json: a type diagnostic cohere reports in one of the root's
// projects, how many times in that file, and why it's expected.
type expectedDiagnostic struct {
	Project string `json:"project"`
	File    string `json:"file"`
	Code    string `json:"code"`
	Count   int    `json:"count"`
	Cause   string `json:"cause"`
}

// Every row of the expected diagnostics names a real file in the project it says, a TS code, a count and a cause, and
// the rows are sorted and each (project, file, code) once, so a job's comparison against cohere's output reads one list.
// The comparison itself needs cohere built from its submodule and run over the whole tree, so it's the head's count
// job, never this test. Mutant: a Node project row's file moved to the root's project, and this fails naming it.
func TestTheExpectedDiagnosticsNameTheProjects(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("expected_diagnostics.json")
	if err != nil {
		t.Fatal(err)
	}
	var expected struct {
		About       string               `json:"about"`
		Diagnostics []expectedDiagnostic `json:"diagnostics"`
	}
	decoder := json.NewDecoder(bytes.NewReader(data))
	decoder.DisallowUnknownFields()
	if err := decoder.Decode(&expected); err != nil {
		t.Fatal(err)
	}
	_, nodeProject := committedProjects(t)
	code := regexp.MustCompile(`^TS[0-9]+$`)
	keys := []string{}
	for _, row := range expected.Diagnostics {
		key := row.Project + " " + row.File + " " + row.Code
		keys = append(keys, key)
		if _, err := os.Stat(filepath.Join(repository, filepath.FromSlash(row.File))); err != nil {
			t.Errorf("%s: %v", key, err)
		}
		if !code.MatchString(row.Code) || row.Count < 1 || row.Cause == "" {
			t.Errorf("%s: want a TS code, a count of at least 1 and a cause, got %+v", key, row)
		}
		inNode := slices.Contains(nodeProject.Files, row.File)
		switch row.Project {
		case nodeFile:
			if !inNode {
				t.Errorf("%s: the file isn't among %s's files", key, nodeFile)
			}
		case rootFile:
			if inNode {
				t.Errorf("%s: the file is one of %s's files, so %s leaves it out", key, nodeFile, rootFile)
			}
		default:
			t.Errorf("%s: want the project %s or %s", key, rootFile, nodeFile)
		}
	}
	if !slices.IsSorted(keys) || len(slices.Compact(slices.Clone(keys))) != len(keys) {
		t.Errorf("the rows aren't sorted by project, file and code, each once")
	}
}
