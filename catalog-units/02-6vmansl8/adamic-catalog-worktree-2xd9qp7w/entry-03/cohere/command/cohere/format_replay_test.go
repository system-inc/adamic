package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
)

// TestANoFixFormatRunReplaysUntilWhatItsWalkReadMoves holds the replay of a check that includes formatting
// (#13a63n3): `--no-fix --format` replays an unchanged tree, and every way a tree can change what the format
// phase reports breaks the replay and reports the change.
//
// Each change is one the graph build never sees, which is the point: a file outside the program, a file
// added to a directory holding nothing the build reads, a .gitignore edit that lets a file into the walk,
// and a file inside a nested repository. A replay that recorded only the program's inputs would print the
// old clean verdict over each, so each is checked to report its file, and each is undone and replayed
// again so the next change starts from a replaying tree.
func TestANoFixFormatRunReplaysUntilWhatItsWalkReadMoves(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	home := t.TempDir()
	root := t.TempDir()
	// The program is the root's own TypeScript, so the build lists no subdirectory: quiet/ is a directory only
	// the format walk lists, as most of a real tree's are.
	writeTree(t, root, map[string]string{
		"tsconfig.json":       strings.Replace(fixScopeTsconfig, `"**/*.ts"`, `"*.ts"`, 1),
		"CohereSettings.json": "{ \"extends\": \"./NexusCohereSettings.json\", \"rules\": {} }\n",
		// The shared tsconfig is written in four-space JSON, which the formatter would rewrite.
		"NexusCohereSettings.json": "{ \"format\": { \"ignore\": [\"tsconfig.json\"] } }\n",
		".gitignore":               ".cache/\nignored/\n",
		"Tidy.ts":                  "export const tidy = 1;\n",
		"data.json":                "{ \"a\": 1 }\n",
		"quiet/readme.txt":         "nothing here is formatted\n",
		"ignored/loose.json":       "{ \"b\":   2 }\n",
		".gitmodules":              "[submodule \"library\"]\n\tpath = library\n\turl = ../library\n",
		"library/.git":             "gitdir: ../.git/modules/library\n",
		"library/inner.json":       "{ \"c\": 3 }\n",
	})
	run := func() (string, int) {
		t.Helper()
		command := exec.Command(binary, "--verbose", "--no-fix", "--format")
		command.Dir = root
		command.Env = append(os.Environ(), "HOME="+home, "XDG_CACHE_HOME=")
		output, err := command.CombinedOutput()
		if exitError, isExit := err.(*exec.ExitError); isExit {
			return string(output), childExitCode(t, exitError)
		}
		if err != nil {
			t.Fatalf("running cohere: %v\n%s", err, output)
		}
		return string(output), 0
	}
	replays := func(when string) {
		t.Helper()
		run()
		if output, code := run(); !strings.HasPrefix(output, "cached: ") || code != 0 {
			t.Fatalf("%s: a clean unchanged tree did not replay (exit %d):\n%s", when, code, output)
		}
	}
	reports := func(when string, fileName string) {
		t.Helper()
		output, code := run()
		if strings.HasPrefix(output, "cached: ") {
			t.Fatalf("%s: the run replayed the clean verdict over a tree that changed:\n%s", when, output)
		}
		if code == 0 || !strings.Contains(output, fileName) {
			t.Fatalf("%s: the run did not report %s (exit %d):\n%s", when, fileName, code, output)
		}
	}
	write := func(name string, contents string) {
		t.Helper()
		writeTree(t, root, map[string]string{name: contents})
	}

	replays("before any change")

	write("data.json", "{ \"a\":  1 }\n")
	reports("a one-byte whitespace edit to a file outside the program", "data.json")
	write("data.json", "{ \"a\": 1 }\n")
	replays("with the edit undone")

	write("quiet/added.json", "{ \"d\":   4 }\n")
	reports("an unformatted file added where the build reads nothing", "added.json")
	if err := os.Remove(filepath.Join(root, "quiet", "added.json")); err != nil {
		t.Fatal(err)
	}
	replays("with the added file removed")

	write(".gitignore", ".cache/\n")
	reports("a .gitignore edit that lets an unformatted file into the walk", "loose.json")
	write(".gitignore", ".cache/\nignored/\n")
	replays("with the .gitignore restored")

	write("library/inner.json", "{ \"c\":  3 }\n")
	reports("a whitespace edit inside a nested repository", "inner.json")
}

// TestAnIgnoreFileTheWalkReadBreaksTheFormatReplay holds the replay to every git ignore file the format walk
// reads (#z661dek), the two that no directory it records can see change.
//
// A nested .gitignore edited in place moves no directory's modification time, and .git/info/exclude sits in
// .git/info, which the walk never lists. A replay keyed only on directories would print the old verdict over
// either edit, so each is made from a replaying tree, must not replay, and must report what the edit changed.
func TestAnIgnoreFileTheWalkReadBreaksTheFormatReplay(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	home := t.TempDir()
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":            strings.Replace(fixScopeTsconfig, `"**/*.ts"`, `"*.ts"`, 1),
		"CohereSettings.json":      "{ \"extends\": \"./NexusCohereSettings.json\", \"rules\": {} }\n",
		"NexusCohereSettings.json": "{ \"format\": { \"ignore\": [\"tsconfig.json\"] } }\n",
		".gitignore":               ".cache/\n",
		".git/HEAD":                "ref: refs/heads/main\n",
		"Tidy.ts":                  "export const tidy = 1;\n",
		"quiet/.gitignore":         "loose.json\n",
		"quiet/loose.json":         "{ \"b\":   2 }\n",
		"stray.json":               "{ \"c\":   3 }\n",
	})
	run := func() (string, int) {
		t.Helper()
		command := exec.Command(binary, "--verbose", "--no-fix", "--format")
		command.Dir = root
		command.Env = append(os.Environ(), "HOME="+home, "XDG_CACHE_HOME=")
		output, err := command.CombinedOutput()
		if exitError, isExit := err.(*exec.ExitError); isExit {
			return string(output), childExitCode(t, exitError)
		}
		if err != nil {
			t.Fatalf("running cohere: %v\n%s", err, output)
		}
		return string(output), 0
	}
	replays := func(when string, reported string) {
		t.Helper()
		run()
		output, _ := run()
		if !strings.HasPrefix(output, "cached: ") || !strings.Contains(output, reported) {
			t.Fatalf("%s: an unchanged tree did not replay its verdict naming %s:\n%s", when, reported, output)
		}
	}
	fresh := func(when string) string {
		t.Helper()
		output, _ := run()
		if strings.HasPrefix(output, "cached: ") {
			t.Fatalf("%s: the run replayed the old verdict over a changed ignore file:\n%s", when, output)
		}
		return output
	}
	write := func(name string, contents string) {
		t.Helper()
		writeTree(t, root, map[string]string{name: contents})
	}

	replays("before any change", "stray.json")

	// (a) A nested .gitignore stops ignoring an unformatted file. The edit is in place, so quiet/'s
	// modification time does not move.
	write("quiet/.gitignore", "\n")
	if output := fresh("a nested .gitignore that stopped ignoring loose.json"); !strings.Contains(output, "loose.json") {
		t.Fatalf("the run did not report loose.json once quiet/.gitignore stopped ignoring it:\n%s", output)
	}
	write("quiet/.gitignore", "loose.json\n")
	replays("with quiet/.gitignore restored", "stray.json")

	// (b) info/exclude appears and ignores the reported file. Nothing the walk lists holds it.
	write(".git/info/exclude", "stray.json\n")
	if output := fresh("info/exclude created to ignore stray.json"); strings.Contains(output, "stray.json") {
		t.Fatalf("the run still reported stray.json, which info/exclude now ignores:\n%s", output)
	}
	if err := os.Remove(filepath.Join(root, ".git", "info", "exclude")); err != nil {
		t.Fatal(err)
	}
	if output := fresh("info/exclude removed again"); !strings.Contains(output, "stray.json") {
		t.Fatalf("the run did not report stray.json once info/exclude was gone:\n%s", output)
	}
	replays("with info/exclude removed", "stray.json")
}
