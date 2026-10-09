package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

// TestCohereRunsNoGit holds that cohere runs no git process on any path, and that what it reports does
// not depend on whether git is there.
//
// Git was how cohere learned what changed. It is gone: what changed is what the format record does not
// hold (format_record.go), a nested repository is a `.git` on disk, and an ignore line is read from the
// file. The proof is two runs of the same sequence. One has a `git` on PATH that writes every
// invocation to a log and fails, so any call, even one whose failure is swallowed, leaves a line. The
// other has the real PATH. The log must be empty, and the two transcripts must be the same.
//
// The sequence is also the end-to-end check of the default format scope: an edited file is formatted,
// an unedited one is not, a file edited and put back is unchanged, a deleted file drops out, a new
// untracked one is seen, a declared submodule is walked, and a clone nobody declared is left alone.
func TestCohereRunsNoGit(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)

	fakeGit := t.TempDir()
	gitLog := filepath.Join(t.TempDir(), "git-calls.log")
	script := "#!/bin/sh\necho \"git $*\" >> '" + gitLog + "'\nexit 1\n"
	if err := os.WriteFile(filepath.Join(fakeGit, "git"), []byte(script), 0o755); err != nil {
		t.Fatal(err)
	}

	// The two sequences share nothing, each with its own project and home, so they run at once: one after
	// the other they were the longest test in the package (#nxgt2ca).
	var withoutGit, withGit []string
	// Not parallel: this group holds its two parallel sequences until both finish, so the comparison below
	// reads both.
	t.Run("sequences", func(t *testing.T) {
		t.Run("without git", func(t *testing.T) {
			t.Parallel()
			withoutGit = runNoGitSequence(t, binary, "PATH="+fakeGit)
			if calls, err := os.ReadFile(gitLog); err == nil && len(calls) > 0 {
				t.Fatalf("cohere ran git:\n%s", calls)
			}
		})
		t.Run("with git", func(t *testing.T) {
			t.Parallel()
			if _, err := exec.LookPath("git"); err != nil {
				t.Skip("no git on this machine to compare against; the no-git sequence still runs")
			}
			withGit = runNoGitSequence(t, binary, "PATH="+os.Getenv("PATH"))
		})
	})
	if t.Failed() || withGit == nil {
		return
	}
	for index := range withoutGit {
		if withoutGit[index] != withGit[index] {
			t.Errorf("step %d differs with git on PATH:\n--- without git\n%s\n--- with git\n%s",
				index+1, withoutGit[index], withGit[index])
		}
	}
}

// runNoGitSequence runs every mode against a fresh project with its own user cache, and returns each
// step's normalized transcript, exit code included.
func runNoGitSequence(t *testing.T, binary string, path string) []string {
	t.Helper()
	root := t.TempDir()
	home := t.TempDir()
	formatted := "export const tidy = 1;\n"
	writeTree(t, root, map[string]string{
		"tsconfig.json":            fixScopeTsconfig,
		"CohereSettings.json":      `{"extends":"./NexusCohereSettings.json","rules":{"no-debugger":"error"}}`,
		"NexusCohereSettings.json": `{"format":{"ignore":[]}}`,
		".gitignore":               "ignored/\n",
		"Producer.ts":              "export function value(): number {\n  debugger;\n  return 1;\n}\n",
		"Tidy.ts":                  formatted,
		"Ugly.ts":                  "export const ugly   =   1\n",
		"ignored/Kept.ts":          "export const kept   =   1\n",
		".gitmodules":              "[submodule \"library\"]\n\tpath = library\n\turl = ../library\n",
		"library/.git":             "gitdir: ../.git/modules/library\n",
		// Unformatted, and carrying a fix no-debugger would apply: a whole-tree run formats it and must not
		// fix it, because it sits in a repository of its own.
		"library/Inner.ts":         "export function inner(): number {\n    debugger;\n    return 1;\n}\n",
		"projects/clone/.git/HEAD": "ref: refs/heads/main\n",
		"projects/clone/Outer.ts":  "export const outer   =   1\n",
	})

	// The temporary root reaches the output through a symbolic link on macOS (`/var` is `/private/var`),
	// so both spellings are normalized, the resolved one first.
	resolvedRoot, err := filepath.EvalSymlinks(root)
	if err != nil {
		t.Fatal(err)
	}
	// Gigabytes too: the memory line reads available memory live, so two runs a second apart differ there.
	durations := regexp.MustCompile(`\d+(\.\d+)?(ms|s|µs| GB)\b`)
	clock := regexp.MustCompile(`\d\d:\d\d:\d\d`)
	// The total line's shape depends on timing as well as its numbers: the types phase's check runs alongside
	// the fix walk, so whether the phases overlap is a property of the invocation, never of git's presence.
	totalLine := regexp.MustCompile(`(?m)^ *total .*\n`)
	run := func(arguments ...string) (string, int) {
		t.Helper()
		command := exec.Command(binary, verboseArguments(arguments)...)
		command.Dir = root
		command.Env = append(os.Environ(), path, "HOME="+home, "XDG_CACHE_HOME="+filepath.Join(home, ".cache"))
		// The streams are read apart and joined after, not combined as they arrive. A run the run cache
		// records tees both through pipes of their own, so the order their lines interleave in is the
		// scheduler's, and two runs that printed the same things would compare as different.
		var stdout, stderr bytes.Buffer
		command.Stdout, command.Stderr = &stdout, &stderr
		err := command.Run()
		output := stdout.String() + "\n--- stderr\n" + stderr.String()
		code := 0
		if exitError, isExit := err.(*exec.ExitError); isExit {
			code = childExitCode(t, exitError)
		} else if err != nil {
			t.Fatalf("running cohere: %v\n%s", err, output)
		}
		normalized := strings.ReplaceAll(output, resolvedRoot, "ROOT")
		normalized = strings.ReplaceAll(normalized, root, "ROOT")
		normalized = durations.ReplaceAllString(normalized, "D")
		normalized = clock.ReplaceAllString(normalized, "T")
		normalized = totalLine.ReplaceAllString(normalized, "")
		return normalized, code
	}
	wouldChange := func(output string) []string {
		names := []string{}
		for _, line := range strings.Split(output, "\n") {
			if strings.Contains(line, "[fix/would-change]") {
				names = append(names, strings.TrimPrefix(strings.SplitN(line, ":", 2)[0], "ROOT/"))
			}
		}
		return names
	}
	expectWouldChange := func(step string, output string, want ...string) {
		t.Helper()
		got := wouldChange(output)
		if strings.Join(got, " ") != strings.Join(want, " ") {
			t.Fatalf("%s: would rewrite %v, expected %v:\n%s", step, got, want, output)
		}
	}

	transcripts := []string{}
	step := func(arguments ...string) string {
		t.Helper()
		output, code := run(arguments...)
		transcripts = append(transcripts, strings.Join(arguments, " ")+"\nexit "+string(rune('0'+code))+"\n"+output)
		return output
	}

	// A check with formatting left out finds nothing to scope, and still reports the fixable finding.
	output := step("--no-fix", "--no-format")
	if !strings.Contains(output, "format scope: nothing: formatting left out (--no-format)") {
		t.Fatalf("a run that does not format still scoped something:\n%s", output)
	}
	expectWouldChange("a check without formatting", output, "Producer.ts")

	// With no record, every file the formatter handles is checked, the configs among them, never the
	// ignored file or the clone nobody declared. The declared submodule is read and not written: its
	// file is reported as its own run's drift (@system_cohere's ruling, writes stay inside one repository).
	output = step("--no-fix", "--format")
	if !strings.Contains(output, "because no earlier check is on record") {
		t.Fatalf("the first run did not say why it checks every file:\n%s", output)
	}
	expectWouldChange("the first format check", output,
		"CohereSettings.json", "NexusCohereSettings.json", "Producer.ts", "Ugly.ts", "tsconfig.json")

	if !strings.Contains(output, "library/Inner.ts:1:1 - the formatter would rewrite this file in nested repository library, which a run here never writes: run cohere there [format/nested-drift]") ||
		!strings.Contains(output, "nested repositories: 1 read, 1 files: 1 checked by the formatter, 1 would change under their own run") {
		t.Fatalf("the submodule's drift was not reported as its own:\n%s", output)
	}

	step("--fix", "--format")
	if inner := readForTest(t, filepath.Join(root, "library/Inner.ts")); inner != "export function inner(): number {\n    debugger;\n    return 1;\n}\n" {
		t.Fatalf("a run in the project wrote into the submodule:\n%s", inner)
	}
	for name, want := range map[string]string{
		"ignored/Kept.ts":         "export const kept   =   1\n",
		"projects/clone/Outer.ts": "export const outer   =   1\n",
	} {
		if got := readForTest(t, filepath.Join(root, name)); got != want {
			t.Fatalf("%s was outside the scope and was rewritten:\n%s", name, got)
		}
	}

	// Everything on record now, so nothing is in scope.
	output = step("--no-fix", "--format")
	if !strings.Contains(output, "0 of 6 files not on record as formatted") {
		t.Fatalf("a formatted tree's scope:\n%s", output)
	}
	expectWouldChange("a formatted tree", output)

	// An edit is in scope; the file put back to bytes on record is not.
	writeTree(t, root, map[string]string{"Tidy.ts": "export const tidy   =   1\n"})
	expectWouldChange("an edit", step("--no-fix", "--format"), "Tidy.ts")
	writeTree(t, root, map[string]string{"Tidy.ts": formatted})
	expectWouldChange("an edit put back", step("--no-fix", "--format"))

	// A deleted file drops out, and a new file nothing has tracked is seen.
	if err := os.Remove(filepath.Join(root, "Ugly.ts")); err != nil {
		t.Fatal(err)
	}
	writeTree(t, root, map[string]string{"Fresh.ts": "export const fresh   =   1\n"})
	output = step("--no-fix", "--format")
	expectWouldChange("a deletion and a new file", output, "Fresh.ts")
	if !strings.Contains(output, "1 of 6 files not on record as formatted") {
		t.Fatalf("the scope did not count the deletion and the new file:\n%s", output)
	}

	// Every other mode, so a git call anywhere on them would land in the log.
	step("--no-fix", "--format-all")
	step("--lint", "Tidy.ts")
	step("--types")
	step("--no-fix", "--unused")
	step("--no-fix", "ignored/Kept.ts")
	step("--no-fix", "--format", "Fresh.ts")
	return transcripts
}
