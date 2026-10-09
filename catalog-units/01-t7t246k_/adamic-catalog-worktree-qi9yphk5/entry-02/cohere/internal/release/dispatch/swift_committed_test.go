package dispatch

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// The Swift engine a gate runs must come from the commit the gate's cohere was built from. These tests
// stand a script in for SwiftPM, because a real cold build takes minutes and what is under test is
// which tree gets built and when. The script records the tree it was handed, and reports
// sourceTreeModified the way the real manifest does, from git's view of the package it was built in.

const testToolchain = "swiftlang-6.4.0.34.1"

// Not parallel: newSwiftFixture swaps the package variable buildSwiftProduct for its own stand-in, and two
// fixtures at once would build with each other's.
func TestTheCommittedSwiftEngineIsBuiltFromTheCommitNotTheWorkingTree(t *testing.T) {
	fixture := newSwiftFixture(t)

	first, built, err := resolveCommittedSwiftEngine(fixture.paths, fixture.head(t), testToolchain, 2)
	if err != nil {
		t.Fatalf("building from the first commit: %v", err)
	}
	if !built {
		t.Fatal("an empty cache reported no build")
	}
	if output := run(t, first); output != "rule=one\n" {
		t.Fatalf("the first engine printed %q", output)
	}

	// An uncommitted edit in the checkout is not part of the commit, so the same engine serves.
	writeFile(t, fixture.rule(), "one, uncommitted\n")
	again, built, err := resolveCommittedSwiftEngine(fixture.paths, fixture.head(t), testToolchain, 2)
	if err != nil {
		t.Fatalf("resolving with an uncommitted edit on disk: %v", err)
	}
	if built || again != first {
		t.Fatalf("an uncommitted Swift edit changed the engine (built=%v)", built)
	}

	// A committed edit builds a new engine, and the build reads the commit even while the checkout
	// holds a different uncommitted edit. A cache hit above could not show that; this fresh build can.
	writeFile(t, fixture.rule(), "two\n")
	fixture.commit(t, "swift")
	writeFile(t, fixture.rule(), "two, uncommitted\n")
	second, built, err := resolveCommittedSwiftEngine(fixture.paths, fixture.head(t), testToolchain, 2)
	if err != nil {
		t.Fatalf("building from the second commit: %v", err)
	}
	if !built || second == first {
		t.Fatalf("a committed Swift edit did not build a new engine (built=%v)", built)
	}
	if output := run(t, second); output != "rule=two\n" {
		t.Fatalf("the engine for the second commit printed %q, so it was not built from that commit", output)
	}

	// A commit that changes only the engine's tests changes nothing compiled in, so it reuses the engine.
	writeFile(t, filepath.Join(fixture.repository, "swift", "Tests", "RuleTests.swift"), "// a new test\n")
	fixture.commit(t, "swift/Tests")
	third, built, err := resolveCommittedSwiftEngine(fixture.paths, fixture.head(t), testToolchain, 2)
	if err != nil {
		t.Fatalf("resolving a tests-only commit: %v", err)
	}
	if built || third != second {
		t.Fatalf("a tests-only commit rebuilt the engine (built=%v)", built)
	}

	if status := gitCommand(t, fixture.paths.SwiftSourceWorktree(), "status", "--porcelain"); status != "" {
		t.Fatalf("the worktree was left dirty:\n%s", status)
	}
}

// Not parallel: newSwiftFixture swaps the package variable buildSwiftProduct for its own stand-in, and two
// fixtures at once would build with each other's.
func TestAnOlderCommitGetsItsOwnEngineWhileTheCheckoutIsCleanAtANewerOne(t *testing.T) {
	// The case no provenance check can catch. A member commits, the checkout is clean at the new
	// commit, and a gate still running the cohere built from the previous one asks for its engine. A
	// working-tree build would hand it the newer engine, and that engine would honestly report a clean
	// tree. Only building the commit asked for gets this right.
	fixture := newSwiftFixture(t)
	older := fixture.head(t)
	writeFile(t, fixture.rule(), "two\n")
	fixture.commit(t, "swift")

	engine, _, err := resolveCommittedSwiftEngine(fixture.paths, older, testToolchain, 2)
	if err != nil {
		t.Fatalf("building the older commit's engine: %v", err)
	}
	if output := run(t, engine); output != "rule=one\n" {
		t.Fatalf("the older commit got an engine printing %q, which is the newer commit's", output)
	}
}

// Not parallel: newSwiftFixture swaps the package variable buildSwiftProduct for its own stand-in, and two
// fixtures at once would build with each other's.
func TestADirtySwiftWorktreeIsRefusedNotForced(t *testing.T) {
	fixture := newSwiftFixture(t)
	if _, _, err := resolveCommittedSwiftEngine(fixture.paths, fixture.head(t), testToolchain, 2); err != nil {
		t.Fatalf("first build: %v", err)
	}

	stray := filepath.Join(fixture.paths.SwiftSourceWorktree(), "swift", "Sources", "Stray.swift")
	writeFile(t, stray, "// someone wrote here\n")
	writeFile(t, fixture.rule(), "two\n")
	fixture.commit(t, "swift")

	_, _, err := resolveCommittedSwiftEngine(fixture.paths, fixture.head(t), testToolchain, 2)
	if err == nil || !strings.Contains(err.Error(), "changes nobody should have made") {
		t.Fatalf("a dirty worktree was moved or built from: %v", err)
	}
	if _, statErr := os.Stat(stray); statErr != nil {
		t.Fatalf("the refusal discarded the stray file, which forcing would do: %v", statErr)
	}
}

// Not parallel: newSwiftFixture swaps the package variable buildSwiftProduct for its own stand-in, and two
// fixtures at once would build with each other's.
func TestACommittedSwiftEngineReportingAModifiedTreeIsNotCached(t *testing.T) {
	fixture := newSwiftFixture(t)
	fixture.reportModified = true

	_, _, err := resolveCommittedSwiftEngine(fixture.paths, fixture.head(t), testToolchain, 2)
	if err == nil || !strings.Contains(err.Error(), "modified source tree") {
		t.Fatalf("an engine reporting a modified tree was accepted: %v", err)
	}
	entries, _ := os.ReadDir(fixture.paths.BinaryDirectory())
	for _, entry := range entries {
		t.Errorf("a refused engine left %s in the cache", entry.Name())
	}
}

// swiftFixture is a repository holding an engine package, with a stand-in for SwiftPM.
type swiftFixture struct {
	repository     string
	paths          Paths
	reportModified bool
}

func newSwiftFixture(t *testing.T) *swiftFixture {
	t.Helper()
	root := t.TempDir()
	repository := filepath.Join(root, "cohere")
	writeFile(t, filepath.Join(repository, "swift", "Package.swift"), "// swift-tools-version:6.0\n")
	writeFile(t, filepath.Join(repository, "swift", "Package.resolved"), "{}\n")
	writeFile(t, filepath.Join(repository, "swift", "Sources", "Rule.swift"), "one\n")
	writeFile(t, filepath.Join(repository, "swift", "Tests", "RuleTests.swift"), "// tests\n")
	writeFile(t, filepath.Join(repository, ".gitignore"), "/.cache/\n")
	gitCommand(t, repository, "init", "-q")
	gitCommand(t, repository, "add", "swift", ".gitignore")
	gitCommand(t, repository, "commit", "-q", "-m", "first")

	fixture := &swiftFixture{
		repository: repository,
		paths:      Paths{ModuleDirectory: repository, CacheDirectory: filepath.Join(repository, ".cache", "cohere")},
	}

	original := buildSwiftProduct
	buildSwiftProduct = fixture.fakeBuild
	t.Cleanup(func() { buildSwiftProduct = original })
	return fixture
}

// fakeBuild writes an engine that prints the rule it was built from and the provenance the real
// manifest would compute for packageDirectory.
func (fixture *swiftFixture) fakeBuild(packageDirectory string, scratchDirectory string, binaryPath string) error {
	rule, err := os.ReadFile(filepath.Join(packageDirectory, "Sources", "Rule.swift"))
	if err != nil {
		return err
	}
	status, err := gitOutput(packageDirectory, "status", "--porcelain")
	if err != nil {
		return err
	}
	modified := strings.TrimSpace(status) != "" || fixture.reportModified

	script := fmt.Sprintf("#!/bin/sh\nif [ \"$3\" = \"--version\" ]; then\n  echo '{\"kind\":\"provenance\",\"sourceTreeModified\":%t}'\n  exit 0\nfi\necho 'rule=%s'\n",
		modified, strings.TrimSpace(string(rule)))
	if err := os.MkdirAll(filepath.Dir(binaryPath), 0o755); err != nil {
		return err
	}
	return os.WriteFile(binaryPath, []byte(script), 0o755)
}

func (fixture *swiftFixture) rule() string {
	return filepath.Join(fixture.repository, "swift", "Sources", "Rule.swift")
}

func (fixture *swiftFixture) head(t *testing.T) string {
	t.Helper()
	return strings.TrimSpace(gitCommand(t, fixture.repository, "rev-parse", "HEAD"))
}

func (fixture *swiftFixture) commit(t *testing.T, paths ...string) {
	t.Helper()
	gitCommand(t, fixture.repository, append([]string{"add"}, paths...)...)
	gitCommand(t, fixture.repository, "commit", "-q", "-m", "next")
}
