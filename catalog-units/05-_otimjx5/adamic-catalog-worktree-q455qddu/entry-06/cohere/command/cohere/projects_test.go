//go:build unix

// Every project here is checked through a stand-in Swift engine, a /bin/sh script, so the file builds on
// Unix, as swift_engine_test.go does (#tejf9bc).

package main

import (
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// mixedRepository writes a repository holding a TypeScript program at its root and a Swift package below
// it, the shape a run from the root checks both of (#f9nftxz).
//
// Its files are in the house format, since a run formats them and the Swift engine refuses --no-format
// (#nqb3mjv), so these runs check formatting too.
func mixedRepository(t *testing.T, swiftDirectory string) string {
	t.Helper()
	root := newRepository(t, map[string]string{
		"tsconfig.json":       incrementalFixtureConfig + "\n",
		"CohereSettings.json": "{ \"rules\": {} }\n",
		"index.ts":            "export const value = 1;\n",
		".gitignore":          "tsconfig.tsbuildinfo\n",
	})
	writeTree(t, filepath.Join(root, swiftDirectory), map[string]string{
		"Package.swift":         discoveryPackage,
		"Sources/Toy/Toy.swift": "let toy = 1\n",
	})
	return root
}

// A mixed repository is checked whole from its root: each project in its own section under its own
// engine, one line per project, and one summary. The run is green only when both are, a Swift finding
// turns it red, and its exit code is the one the failing project's own run gives.
func TestAMixedRepositoryIsCheckedWholeAndTheWorstExitWins(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	clean := fakeSwiftEngine(t, "cat "+filepath.Join(contractDirectory(t), "Clean.jsonl"))
	findings := fakeSwiftEngine(t, "cat "+filepath.Join(contractDirectory(t), "Findings.jsonl")+"\nexit 1")

	root := mixedRepository(t, "apple/Toy")
	output, exitCode := runWithEngine(t, binary, clean, root, "--no-fix")
	if exitCode != 0 {
		t.Fatalf("a clean mixed repository exited %d:\n%s", exitCode, output)
	}
	for _, want := range []string{
		"== . (TypeScript) ==",
		"== apple/Toy (Swift) ==",
		"  . (TypeScript): green",
		"  apple/Toy (Swift): green",
		"projects: 2 checked (1 TypeScript, 1 Swift), every one green",
	} {
		if !strings.Contains(output, want) {
			t.Errorf("the report lacks %q:\n%s", want, output)
		}
	}

	output, exitCode = runWithEngine(t, binary, findings, root, "--no-fix")
	_, alone := runWithEngine(t, binary, findings, root, "--no-fix", "--directory", filepath.Join(root, "apple", "Toy"))
	if exitCode == 0 || exitCode != alone {
		t.Fatalf("a Swift finding exited %d from the root and %d from the package alone, want the same nonzero code:\n%s", exitCode, alone, output)
	}
	if !strings.Contains(output, "apple/Toy (Swift): failed, exit") || !strings.Contains(output, "1 failed") {
		t.Errorf("the report does not name the failing project:\n%s", output)
	}
	if !strings.Contains(output, "  . (TypeScript): green") {
		t.Errorf("the TypeScript project's own verdict is lost:\n%s", output)
	}
}

// Both markers in one directory are two projects, each checked by its own engine; and projects nested two
// levels down are found from the root.
func TestBothMarkersInOneDirectoryAndProjectsNestedDeepAreEachChecked(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	clean := fakeSwiftEngine(t, "cat "+filepath.Join(contractDirectory(t), "Clean.jsonl"))

	together := mixedRepository(t, ".")
	output, exitCode := runWithEngine(t, binary, clean, together, "--no-fix")
	if exitCode != 0 || !strings.Contains(output, "== . (TypeScript) ==") || !strings.Contains(output, "== . (Swift) ==") {
		t.Fatalf("a directory holding both markers was not checked by both engines (exit %d):\n%s", exitCode, output)
	}

	deep := newRepository(t, nil)
	writeTree(t, filepath.Join(deep, "web", "app"), map[string]string{
		"tsconfig.json":       incrementalFixtureConfig + "\n",
		"CohereSettings.json": "{ \"rules\": {} }\n",
		"index.ts":            "export const value = 1;\n",
	})
	writeTree(t, filepath.Join(deep, "apple", "Toy"), map[string]string{"Package.swift": discoveryPackage, "Sources/Toy/Toy.swift": "let toy = 1\n"})
	output, exitCode = runWithEngine(t, binary, clean, deep, "--no-fix")
	if exitCode != 0 || !strings.Contains(output, "== apple/Toy (Swift) ==") || !strings.Contains(output, "== web/app (TypeScript) ==") {
		t.Fatalf("projects two levels down were not both checked (exit %d):\n%s", exitCode, output)
	}
}

// A project whose run fails without a verdict, here an engine that dies before reporting, fails the run
// from the root; and a root holding no project at all fails loudly rather than passing over nothing.
func TestACrashedProjectOrNoProjectFailsTheRun(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	crashed := fakeSwiftEngine(t, "exit 2")
	root := mixedRepository(t, "apple/Toy")
	output, exitCode := runWithEngine(t, binary, crashed, root, "--no-fix")
	if exitCode == 0 || !strings.Contains(output, "apple/Toy (Swift): failed") {
		t.Fatalf("a Swift engine that died passed the run (exit %d):\n%s", exitCode, output)
	}

	empty := newRepository(t, map[string]string{"README.md": "nothing to check\n"})
	output, exitCode = runWithEngine(t, binary, crashed, empty)
	if exitCode == 0 || !strings.Contains(output, "no project here to check") {
		t.Fatalf("a repository with no project passed (exit %d):\n%s", exitCode, output)
	}
}

// Under --format-only a Swift package is not run, since the engine has no format-only mode and would
// type-check and lint it as well (#nqb3mjv): the engine never starts, and the final line names the gap
// rather than passing over it. Beside a TypeScript program the run checks that and passes; alone, it has
// checked nothing and fails (#71a0ts8). The same flag against the package alone is refused by name.
func TestFormatOnlySkipsASwiftPackageAndNamesTheGap(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	started := filepath.Join(t.TempDir(), "started")
	engine := fakeSwiftEngine(t, "touch '"+started+"'\nexit 3")
	formatOnly := func(root string) (string, int) {
		command := exec.Command(binary, "--no-fix", "--format-only")
		command.Dir = root
		command.Env = append(os.Environ(), "COHERE_SWIFT_ENGINE="+engine, "COHERE_VERDICT_FD=")
		output, err := command.CombinedOutput()
		var exited *exec.ExitError
		if errors.As(err, &exited) {
			return string(output), childExitCode(t, exited)
		} else if err != nil {
			t.Fatalf("running cohere: %v\n%s", err, output)
		}
		return string(output), 0
	}
	const gap = "⚠ apple/Toy not format-checked: the Swift engine has no format-only mode yet"

	output, exitCode := formatOnly(mixedRepository(t, "apple/Toy"))
	if exitCode != 0 || !strings.Contains(output, "✓ 💎") || !strings.Contains(output, "1 project") || !strings.Contains(output, gap) {
		t.Errorf("beside a TypeScript program, the run did not pass naming the Swift gap (exit %d):\n%s", exitCode, output)
	}

	swiftOnly := newRepository(t, nil)
	writeTree(t, filepath.Join(swiftOnly, "apple", "Toy"), map[string]string{"Package.swift": discoveryPackage, "Sources/Toy/Toy.swift": "let toy = 1\n"})
	output, exitCode = formatOnly(swiftOnly)
	// The root as cohere names it, through any symlink in the temporary directory's path.
	resolved, err := filepath.EvalSymlinks(swiftOnly)
	if err != nil {
		t.Fatal(err)
	}
	if exitCode == 0 || !strings.Contains(output, "✗ ☠️") || !strings.Contains(output, "nothing checked under "+resolved) || !strings.Contains(output, gap) {
		t.Errorf("a run that skipped its only project did not fail naming what it skipped (exit %d):\n%s", exitCode, output)
	}
	if _, err := os.Stat(started); err == nil {
		t.Errorf("the Swift engine ran under --format-only")
	}

	output, exitCode = runWithEngine(t, binary, engine, swiftOnly, "--no-fix", "--format-only", "--directory", filepath.Join(swiftOnly, "apple", "Toy"))
	if exitCode == 0 || !strings.Contains(output, "--format-only is not implemented for Swift yet") {
		t.Errorf("--format-only against the package alone was not refused by name (exit %d):\n%s", exitCode, output)
	}
	if _, err := os.Stat(started); err == nil {
		t.Errorf("the Swift engine ran under a refused --format-only")
	}
}

// contractDirectory is where the Swift contract's fixtures are.
func contractDirectory(t *testing.T) string {
	t.Helper()
	directory, err := filepath.Abs(filepath.Join("..", "..", "swift", "Contract"))
	if err != nil {
		t.Fatal(err)
	}
	return directory
}
