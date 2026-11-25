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
// from it, and the three tsconfigs read there.
//
// Reads: tsconfig.json
// Reads: tsconfig.standalone.json
// Reads: tsconfig.node.json
// Reads: dedication/*
// Reads: bench/*
// Reads: stage1/*
// Reads: internal/load/testdata/*
// Reads: internal/oracle/testdata/*
const repository = "../../.."

// The root's three tsconfigs are exactly what `go run ./internal/load/projects` writes from load's own predicate, asked
// of each source's first program: a source added under the include directories, moved, or given an import that changes
// whether Load gives it Node's declarations, with no rerun of the generator, fails this, naming each file that moved.
// Mutant: one source deleted from tsconfig.node.json's files and tsconfig.standalone.json's exclude, and this fails
// naming it.
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
	for _, name := range []string{solutionFile, standaloneFile, nodeFile} {
		committed, err := os.ReadFile(filepath.Join(repository, name))
		if err != nil {
			t.Fatal(err)
		}
		if !bytes.Equal(committed, rendered[name]) {
			t.Errorf("%s isn't what load's predicate gives; run `go run ./internal/load/projects` from the repository's root. Only in the committed file:\n%s\nOnly in what the generator writes:\n%s",
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
// job, never this test. Mutant: a node project row's file moved to tsconfig.standalone.json, and this fails naming it.
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
	nodeProject, err := os.ReadFile(filepath.Join(repository, nodeFile))
	if err != nil {
		t.Fatal(err)
	}
	var node project
	if err := json.Unmarshal(nodeProject, &node); err != nil {
		t.Fatal(err)
	}
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
		inNode := slices.Contains(node.Files, row.File)
		switch row.Project {
		case nodeFile:
			if !inNode {
				t.Errorf("%s: the file isn't among %s's files", key, nodeFile)
			}
		case standaloneFile:
			if inNode {
				t.Errorf("%s: the file is one of %s's files, so %s leaves it out", key, nodeFile, standaloneFile)
			}
		default:
			t.Errorf("%s: want the project %s or %s", key, standaloneFile, nodeFile)
		}
	}
	if !slices.IsSorted(keys) || len(slices.Compact(slices.Clone(keys))) != len(keys) {
		t.Errorf("the rows aren't sorted by project, file and code, each once")
	}
}
