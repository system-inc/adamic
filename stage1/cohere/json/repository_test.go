package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
	"github.com/system-inc/adamic/internal/tracked"

	"github.com/system-inc/adamic/stage1/cohere/json/oracles"
)

type repositoryJSON struct {
	paths map[string]bool
	dirty []string
}

var repositoryJSONPatterns = []string{"*.json", ":(exclude)cohere/**", ":(exclude,glob)**/.git/**"}

func gitCorpus(root string, arguments ...string) ([]byte, error) {
	return gitCorpusInput(root, nil, arguments...)
}

func gitCorpusInput(root string, input []byte, arguments ...string) ([]byte, error) {
	command := toolchain.GitCommand(append([]string{"-C", root}, arguments...)...)
	command.Stdin = bytes.NewReader(input)
	output, combinedOutputError := childguard.CombinedOutput(command, jsonGuard)
	if combinedOutputError != nil {
		return nil, fmt.Errorf(
			"JSON corpus requires usable Git metadata: git %s: %w: %s",
			strings.Join(arguments, " "),
			combinedOutputError,
			strings.TrimSpace(string(output)),
		)
	}
	return output, nil
}

func repositoryJSONAtHEAD(root string) (repositoryJSON, error) {
	if !tracked.Git(root) {
		return repositoryJSONInManifest(root)
	}
	state := repositoryJSON{paths: map[string]bool{}}
	top, gitCorpusError := gitCorpus(root, "rev-parse", "--show-toplevel")
	if gitCorpusError != nil {
		return state, gitCorpusError
	}
	expected, resolveSymlinksError := filepath.EvalSymlinks(root)
	if resolveSymlinksError != nil {
		return state, resolveSymlinksError
	}
	actual, resolveSymlinksError := filepath.EvalSymlinks(strings.TrimSpace(string(top)))
	if resolveSymlinksError != nil {
		return state, resolveSymlinksError
	}
	if actual != expected {
		return state, fmt.Errorf(
			"JSON corpus requires usable Git metadata at walk root %s; Git root is %s",
			expected,
			actual,
		)
	}
	if _, gitCorpusError := gitCorpus(root, "rev-parse", "--verify", "HEAD"); gitCorpusError != nil {
		return state, gitCorpusError
	}
	paths, gitCorpusError := gitCorpus(
		root,
		append([]string{"ls-files", "--cached", "-z", "--"}, repositoryJSONPatterns...)...,
	)
	if gitCorpusError != nil {
		return state, gitCorpusError
	}
	for _, path := range strings.Split(string(paths), "\x00") {
		if path != "" {
			state.paths[path] = true
		}
	}
	dirtyPaths := map[string]bool{}
	// Check index versus HEAD and worktree versus index separately. Comparing
	// worktree directly with HEAD can hide a staged change which was undone
	// only in the worktree.
	for _, arguments := range [][]string{
		{"diff", "--cached", "--name-only", "-z", "HEAD", "--"},
		{"diff", "--name-only", "-z", "--"},
	} {
		dirty, gitCorpusError := gitCorpus(root, append(arguments, repositoryJSONPatterns...)...)
		if gitCorpusError != nil {
			return state, gitCorpusError
		}
		for _, path := range strings.Split(string(dirty), "\x00") {
			if path != "" {
				dirtyPaths[path] = true
			}
		}
	}
	for path := range dirtyPaths {
		state.dirty = append(state.dirty, path)
	}
	sort.Strings(state.dirty)

	return state, nil
}

// repositoryJSONInManifest is repositoryJSONAtHEAD for a source with no .git, as a Loom runner unpacks it: the tree's
// manifest (internal/tracked) must record root as the tree's own repository, its *.json entries outside cohere/ are
// the tracked paths, and each one missing or not as recorded is dirty. The manifest is HEAD and index at once.
func repositoryJSONInManifest(root string) (repositoryJSON, error) {
	state := repositoryJSON{paths: map[string]bool{}}
	if _, headError := tracked.Head(root); headError != nil {
		return state, fmt.Errorf("JSON corpus requires usable Git metadata or the tree's manifest: %w", headError)
	}
	entries, filesError := tracked.Files(root)
	if filesError != nil {
		return state, fmt.Errorf("JSON corpus requires usable Git metadata or the tree's manifest: %w", filesError)
	}
	var selected []tracked.Entry
	for _, entry := range entries {
		// The pathspecs repositoryJSONPatterns names: *.json at any depth, never under cohere/ or a .git directory.
		if entry.Submodule() ||
			!strings.HasSuffix(entry.Path, ".json") ||
			strings.HasPrefix(entry.Path, "cohere/") ||
			strings.HasPrefix(entry.Path, ".git/") ||
			strings.Contains(entry.Path, "/.git/") {
			continue
		}
		selected = append(selected, entry)
		state.paths[entry.Path] = true
	}
	missing, changed, changedError := tracked.Changed(root, selected)
	if changedError != nil {
		return state, changedError
	}
	state.dirty = append(missing, changed...)
	sort.Strings(state.dirty)
	return state, nil
}

// Tracked repository inputs follow Git HEAD. Only externally provisioned or
// generated inputs retain identity pins; arbitrary untracked files are refused.
func validateCorpus(root string, cases []textCase, expected corpusPin) (corpusPin, int, error) {
	state, repositoryJsonAtHeadError := repositoryJSONAtHEAD(root)
	if repositoryJsonAtHeadError != nil {
		return corpusPin{}, 0, repositoryJsonAtHeadError
	}
	return validateCorpusState(state, cases, expected)
}

func validateCorpusState(state repositoryJSON, cases []textCase, expected corpusPin) (corpusPin, int, error) {
	return oracles.ValidateCorpusState(state.paths, state.dirty, oracleCases(cases), expected)
}

type repositoryFixture struct {
	root string
	pin  corpusPin
}

func fixtureWrite(test *testing.T, root, path, text string) {
	test.Helper()
	absolute := filepath.Join(root, filepath.FromSlash(path))
	if createDirectoryError := os.MkdirAll(filepath.Dir(absolute), 0755); createDirectoryError != nil {
		test.Fatal(createDirectoryError)
	}
	if writeError := os.WriteFile(absolute, []byte(text), 0644); writeError != nil {
		test.Fatal(writeError)
	}
}

func fixtureGit(test *testing.T, root string, arguments ...string) {
	test.Helper()
	if _, gitCorpusError := gitCorpus(root, arguments...); gitCorpusError != nil {
		test.Fatal(gitCorpusError)
	}
}

func commitFixtureJSON(test *testing.T, root string, paths ...string) {
	test.Helper()
	fixtureGit(test, root, append([]string{"add", "--"}, paths...)...)
	fixtureGit(
		test,
		root,
		"-c",
		"user.name=Corpus fixture",
		"-c",
		"user.email=corpus@example.invalid",
		"-c",
		"commit.gpgsign=false",
		"commit",
		"-qm",
		"Scratch JSON landing",
	)
}

func newRepositoryFixture(test *testing.T) repositoryFixture {
	test.Helper()
	root := test.TempDir()
	fixtureGit(test, root, "init", "-q")
	fixtureWrite(test, root, "tracked.json", "{}")
	fixtureWrite(test, root, "nested/second.json", "null")
	commitFixtureJSON(test, root, "tracked.json", "nested/second.json")
	for _, item := range []textCase{
		{"cohere/upstream.json", "[1]"},
		{"cohere/TypeScript/upstream.json", "[]"},
		{"stage3/api/node_modules/typescript/package.json", "{}"},
	} {
		fixtureWrite(test, root, item.Name, item.Text)
	}
	fixture := repositoryFixture{root: root}
	cases := fixture.cases(test)
	var pinned []textCase
	for _, item := range cases {
		if item.Name != "tracked.json" && item.Name != "nested/second.json" {
			pinned = append(pinned, item)
		}
	}
	fixture.pin = pinForCases(pinned)
	if _, count, validateCorpusError := validateCorpus(
		root,
		cases,
		fixture.pin,
	); validateCorpusError != nil || count != 2 {
		test.Fatalf("fixture baseline: count %d: %v", count, validateCorpusError)
	}
	return fixture
}

func (fixture repositoryFixture) cases(test *testing.T) []textCase {
	test.Helper()
	cases, fileCorpusError := fileCorpus(fixture.root)
	if fileCorpusError != nil {
		test.Fatal(fileCorpusError)
	}
	return append(cases, textCase{"generated/0/probe.json", "null"})
}

func TestRepositoryCorpusMutants(test *testing.T) {
	test.Parallel()
	for _, mutant := range []struct {
		name, path, message string
		apply               func(*testing.T, repositoryFixture) []textCase
	}{
		{"walk drops tracked JSON", "tracked.json", "repository missing tracked JSON", func(
			test *testing.T,
			fixture repositoryFixture,
		) []textCase {
			var kept []textCase
			for _, item := range fixture.cases(test) {
				if item.Name != "tracked.json" {
					kept = append(kept, item)
				}
			}
			return kept
		}},
		{"unexpected untracked JSON", "unexpected.json", "unexpected untracked JSON", func(
			test *testing.T,
			fixture repositoryFixture,
		) []textCase {
			fixtureWrite(test, fixture.root, "unexpected.json", "{}")
			return fixture.cases(test)
		}},
		{"unlisted provisioned JSON", "stage3/api/node_modules/unexpected.json", "unexpected untracked JSON", func(
			test *testing.T,
			fixture repositoryFixture,
		) []textCase {
			fixtureWrite(test, fixture.root, "stage3/api/node_modules/unexpected.json", "{}")
			return fixture.cases(test)
		}},
		{"missing provisioned file", "stage3/api/node_modules/typescript/package.json", "missing", func(
			test *testing.T,
			fixture repositoryFixture,
		) []textCase {
			if removeError := os.Remove(
				filepath.Join(fixture.root, "stage3/api/node_modules/typescript/package.json"),
			); removeError != nil {
				test.Fatal(removeError)
			}
			return fixture.cases(test)
		}},
		{"one byte upstream change", "cohere/upstream.json", "changed text", func(
			test *testing.T,
			fixture repositoryFixture,
		) []textCase {
			fixtureWrite(test, fixture.root, "cohere/upstream.json", "[2]")
			return fixture.cases(test)
		}},
		{"staged change with restored worktree", "tracked.json", "repository dirty JSON", func(
			test *testing.T,
			fixture repositoryFixture,
		) []textCase {
			fixtureWrite(test, fixture.root, "tracked.json", "{ }")
			fixtureGit(test, fixture.root, "add", "--", "tracked.json")
			fixtureWrite(test, fixture.root, "tracked.json", "{}")
			return fixture.cases(test)
		}},
		{"dirty tracked JSON", "tracked.json", "repository dirty JSON", func(
			test *testing.T,
			fixture repositoryFixture,
		) []textCase {
			fixtureWrite(test, fixture.root, "tracked.json", "{ }")
			return fixture.cases(test)
		}},
	} {
		for _, manifest := range []bool{false, true} {
			name := mutant.name
			if manifest {
				// The manifest is HEAD and index at once, so a staged change has nowhere to hide.
				if mutant.name == "staged change with restored worktree" {
					continue
				}
				name += " without .git"
			}
			test.Run(name, func(test *testing.T) {
				test.Parallel()
				fixture := newRepositoryFixture(test)
				if manifest {
					fixture = unpackRepositoryFixture(test, fixture)
				}
				cases := mutant.apply(test, fixture)
				_, _, validateCorpusError := validateCorpus(fixture.root, cases, fixture.pin)
				if validateCorpusError == nil ||
					!strings.Contains(validateCorpusError.Error(), mutant.message) ||
					!strings.Contains(validateCorpusError.Error(), mutant.path) {
					test.Fatalf("mutant survived or lost its name: %v", validateCorpusError)
				}
				test.Logf("caught %s: %v", mutant.path, validateCorpusError)
			})
		}
	}
}

// unpackRepositoryFixture is the fixture as a Loom runner holds a source: every file but .git, and the manifest
// build-tree writes (internal/tracked). It validates as the checkout does, with the same count.
func unpackRepositoryFixture(test *testing.T, fixture repositoryFixture) repositoryFixture {
	test.Helper()
	root := test.TempDir()
	walkError := filepath.WalkDir(fixture.root, func(name string, entry fs.DirEntry, pathError error) error {
		if pathError != nil || name == fixture.root {
			return pathError
		}
		relative, relativePathError := filepath.Rel(fixture.root, name)
		if relativePathError != nil {
			return relativePathError
		}
		if entry.IsDir() {
			if relative == ".git" {
				return filepath.SkipDir
			}
			return nil
		}
		data, readError := os.ReadFile(name)
		if readError != nil {
			return readError
		}
		fixtureWrite(test, root, filepath.ToSlash(relative), string(data))
		return nil
	})
	if walkError != nil {
		test.Fatal(walkError)
	}
	if writeError := tracked.Write(fixture.root, root); writeError != nil {
		test.Fatal(writeError)
	}
	unpacked := repositoryFixture{root: root, pin: fixture.pin}
	if _, count, validateCorpusError := validateCorpus(
		root,
		unpacked.cases(test),
		unpacked.pin,
	); validateCorpusError != nil || count != 2 {
		test.Fatalf("unpacked fixture baseline: count %d: %v", count, validateCorpusError)
	}
	return unpacked
}

func TestRepositoryLandingWithoutPinEdit(test *testing.T) {
	test.Parallel()
	fixture := newRepositoryFixture(test)
	before, encodeError := json.Marshal(fixture.pin)
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	// Only this test's private Git history receives the simulated later landing.
	fixtureWrite(test, fixture.root, "landing-added.json", "{}")
	commitFixtureJSON(test, fixture.root, "landing-added.json")
	actual, count, validateCorpusError := validateCorpus(fixture.root, fixture.cases(test), fixture.pin)
	if validateCorpusError != nil || count != 3 || actual.SHA256 != fixture.pin.SHA256 {
		test.Fatalf("landing changed pin or parity: count %d: %v", count, validateCorpusError)
	}
	after, encodeError := json.Marshal(fixture.pin)
	if encodeError != nil || !bytes.Equal(before, after) {
		test.Fatal("landing edited the pin")
	}
	test.Log("landing-added.json accepted at the later scratch HEAD; repository 2 -> 3; pin unchanged")
}

func TestRepositoryRequiresGit(test *testing.T) {
	test.Parallel()
	root := test.TempDir()
	fixtureWrite(test, root, "a.json", "{}")
	_, _, validateCorpusError := validateCorpus(root, []textCase{{"a.json", "{}"}}, corpusPin{})
	if validateCorpusError == nil || !strings.Contains(validateCorpusError.Error(), "requires usable Git metadata") {
		test.Fatalf("export without Git silently fell back to walking: %v", validateCorpusError)
	}
	test.Logf("caught checkout without Git: %v", validateCorpusError)
}
