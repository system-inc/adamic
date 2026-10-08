package json

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
)

type repositoryJSON struct {
	paths map[string]bool
	dirty []string
}

var repositoryJSONPatterns = []string{"*.json", ":(exclude)cohere/**", ":(exclude,glob)**/.git/**"}

func gitCorpus(root string, arguments ...string) ([]byte, error) {
	command := exec.Command("git", append([]string{"-C", root}, arguments...)...)
	output, err := childguard.CombinedOutput(command, childguard.Options{})
	if err != nil {
		return nil, fmt.Errorf("JSON corpus requires usable Git metadata: git %s: %w: %s", strings.Join(arguments, " "), err, strings.TrimSpace(string(output)))
	}
	return output, nil
}

func repositoryJSONAtHEAD(root string) (repositoryJSON, error) {
	state := repositoryJSON{paths: map[string]bool{}}
	top, err := gitCorpus(root, "rev-parse", "--show-toplevel")
	if err != nil {
		return state, err
	}
	expected, err := filepath.EvalSymlinks(root)
	if err != nil {
		return state, err
	}
	actual, err := filepath.EvalSymlinks(strings.TrimSpace(string(top)))
	if err != nil {
		return state, err
	}
	if actual != expected {
		return state, fmt.Errorf("JSON corpus requires usable Git metadata at walk root %s; Git root is %s", expected, actual)
	}
	if _, err := gitCorpus(root, "rev-parse", "--verify", "HEAD"); err != nil {
		return state, err
	}
	paths, err := gitCorpus(root, append([]string{"ls-files", "--cached", "-z", "--"}, repositoryJSONPatterns...)...)
	if err != nil {
		return state, err
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
		dirty, err := gitCorpus(root, append(arguments, repositoryJSONPatterns...)...)
		if err != nil {
			return state, err
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

// Tracked repository inputs follow Git HEAD. Only externally provisioned or
// generated inputs retain identity pins; arbitrary untracked files are refused.
func validateCorpus(root string, cases []textCase, expected corpusPin) (corpusPin, int, error) {
	state, err := repositoryJSONAtHEAD(root)
	if err != nil {
		return corpusPin{}, 0, err
	}
	if len(state.dirty) > 0 {
		return corpusPin{}, 0, fmt.Errorf("repository dirty JSON: %s", strings.Join(state.dirty, ", "))
	}
	allowed := map[string]bool{}
	for _, item := range expected.Cases {
		allowed[item.Path] = true
	}
	seen, tracked := map[string]bool{}, map[string]bool{}
	var pinned []textCase
	for _, item := range cases {
		if seen[item.Name] {
			return corpusPin{}, len(tracked), fmt.Errorf("duplicate corpus path: %s", item.Name)
		}
		seen[item.Name] = true
		if state.paths[item.Name] {
			tracked[item.Name] = true
			continue
		}
		group := caseGroup(item.Name)
		if group == "cohere" || group == "TypeScript" || allowed[item.Name] {
			pinned = append(pinned, item)
			continue
		}
		return corpusPin{}, len(tracked), fmt.Errorf("unexpected untracked JSON: %s", item.Name)
	}
	var missing []string
	for path := range state.paths {
		if !tracked[path] {
			missing = append(missing, path)
		}
	}
	if len(missing) > 0 {
		sort.Strings(missing)
		return corpusPin{}, len(tracked), fmt.Errorf("repository missing tracked JSON from walk: %s", strings.Join(missing, ", "))
	}
	pin := pinForCases(pinned)
	return pin, len(tracked), corpusPinError(expected, pin)
}

type repositoryFixture struct {
	root string
	pin  corpusPin
}

func fixtureWrite(t *testing.T, root, path, text string) {
	t.Helper()
	absolute := filepath.Join(root, filepath.FromSlash(path))
	if err := os.MkdirAll(filepath.Dir(absolute), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(absolute, []byte(text), 0644); err != nil {
		t.Fatal(err)
	}
}

func fixtureGit(t *testing.T, root string, arguments ...string) {
	t.Helper()
	if _, err := gitCorpus(root, arguments...); err != nil {
		t.Fatal(err)
	}
}

func commitFixtureJSON(t *testing.T, root string, paths ...string) {
	t.Helper()
	fixtureGit(t, root, append([]string{"add", "--"}, paths...)...)
	fixtureGit(t, root, "-c", "user.name=Corpus fixture", "-c", "user.email=corpus@example.invalid", "-c", "commit.gpgsign=false", "commit", "-qm", "Scratch JSON landing")
}

func newRepositoryFixture(t *testing.T) repositoryFixture {
	t.Helper()
	root := t.TempDir()
	fixtureGit(t, root, "init", "-q")
	fixtureWrite(t, root, "tracked.json", "{}")
	fixtureWrite(t, root, "nested/second.json", "null")
	commitFixtureJSON(t, root, "tracked.json", "nested/second.json")
	for _, item := range []textCase{
		{"cohere/upstream.json", "[1]"},
		{"cohere/TypeScript/upstream.json", "[]"},
		{"stage3/api/node_modules/typescript/package.json", "{}"},
	} {
		fixtureWrite(t, root, item.Name, item.Text)
	}
	fixture := repositoryFixture{root: root}
	cases := fixture.cases(t)
	var pinned []textCase
	for _, item := range cases {
		if item.Name != "tracked.json" && item.Name != "nested/second.json" {
			pinned = append(pinned, item)
		}
	}
	fixture.pin = pinForCases(pinned)
	if _, count, err := validateCorpus(root, cases, fixture.pin); err != nil || count != 2 {
		t.Fatalf("fixture baseline: count %d: %v", count, err)
	}
	return fixture
}

func (f repositoryFixture) cases(t *testing.T) []textCase {
	t.Helper()
	cases, err := fileCorpus(f.root)
	if err != nil {
		t.Fatal(err)
	}
	return append(cases, textCase{"generated/0/probe.json", "null"})
}

func TestRepositoryCorpusMutants(t *testing.T) {
	t.Parallel()
	for _, mutant := range []struct {
		name, path, message string
		apply               func(*testing.T, repositoryFixture) []textCase
	}{
		{"walk drops tracked JSON", "tracked.json", "repository missing tracked JSON", func(t *testing.T, f repositoryFixture) []textCase {
			var kept []textCase
			for _, item := range f.cases(t) {
				if item.Name != "tracked.json" {
					kept = append(kept, item)
				}
			}
			return kept
		}},
		{"unexpected untracked JSON", "unexpected.json", "unexpected untracked JSON", func(t *testing.T, f repositoryFixture) []textCase {
			fixtureWrite(t, f.root, "unexpected.json", "{}")
			return f.cases(t)
		}},
		{"unlisted provisioned JSON", "stage3/api/node_modules/unexpected.json", "unexpected untracked JSON", func(t *testing.T, f repositoryFixture) []textCase {
			fixtureWrite(t, f.root, "stage3/api/node_modules/unexpected.json", "{}")
			return f.cases(t)
		}},
		{"missing provisioned file", "stage3/api/node_modules/typescript/package.json", "missing", func(t *testing.T, f repositoryFixture) []textCase {
			if err := os.Remove(filepath.Join(f.root, "stage3/api/node_modules/typescript/package.json")); err != nil {
				t.Fatal(err)
			}
			return f.cases(t)
		}},
		{"one byte upstream change", "cohere/upstream.json", "changed text", func(t *testing.T, f repositoryFixture) []textCase {
			fixtureWrite(t, f.root, "cohere/upstream.json", "[2]")
			return f.cases(t)
		}},
		{"staged change with restored worktree", "tracked.json", "repository dirty JSON", func(t *testing.T, f repositoryFixture) []textCase {
			fixtureWrite(t, f.root, "tracked.json", "{ }")
			fixtureGit(t, f.root, "add", "--", "tracked.json")
			fixtureWrite(t, f.root, "tracked.json", "{}")
			return f.cases(t)
		}},
		{"dirty tracked JSON", "tracked.json", "repository dirty JSON", func(t *testing.T, f repositoryFixture) []textCase {
			fixtureWrite(t, f.root, "tracked.json", "{ }")
			return f.cases(t)
		}},
	} {
		t.Run(mutant.name, func(t *testing.T) {
			t.Parallel()
			fixture := newRepositoryFixture(t)
			cases := mutant.apply(t, fixture)
			_, _, err := validateCorpus(fixture.root, cases, fixture.pin)
			if err == nil || !strings.Contains(err.Error(), mutant.message) || !strings.Contains(err.Error(), mutant.path) {
				t.Fatalf("mutant survived or lost its name: %v", err)
			}
			t.Logf("caught %s: %v", mutant.path, err)
		})
	}
}

func TestRepositoryLandingWithoutPinEdit(t *testing.T) {
	t.Parallel()
	fixture := newRepositoryFixture(t)
	before, err := json.Marshal(fixture.pin)
	if err != nil {
		t.Fatal(err)
	}
	// Only this test's private Git history receives the simulated later landing.
	fixtureWrite(t, fixture.root, "landing-added.json", "{}")
	commitFixtureJSON(t, fixture.root, "landing-added.json")
	actual, count, err := validateCorpus(fixture.root, fixture.cases(t), fixture.pin)
	if err != nil || count != 3 || actual.SHA256 != fixture.pin.SHA256 {
		t.Fatalf("landing changed pin or parity: count %d: %v", count, err)
	}
	after, err := json.Marshal(fixture.pin)
	if err != nil || !bytes.Equal(before, after) {
		t.Fatal("landing edited the pin")
	}
	t.Log("landing-added.json accepted at the later scratch HEAD; repository 2 -> 3; pin unchanged")
}

func TestRepositoryRequiresGit(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	fixtureWrite(t, root, "a.json", "{}")
	_, _, err := validateCorpus(root, []textCase{{"a.json", "{}"}}, corpusPin{})
	if err == nil || !strings.Contains(err.Error(), "requires usable Git metadata") {
		t.Fatalf("export without Git silently fell back to walking: %v", err)
	}
	t.Logf("caught checkout without Git: %v", err)
}
