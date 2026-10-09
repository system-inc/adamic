package main

import (
	"os"
	"os/exec"
	"strings"
	"testing"
)

// TestADefaultRunPrintsNoNotes is the default view end to end, golden: a clean run whose cache table was
// just thrown away prints its footer and nothing else. Notes are --verbose's, so the line that says the
// table was discarded, and the one that says .gitignore does not ignore the cache, print only there.
//
// This is the run ahra saw after a cache format bump: two `note: cache table discarded` lines above the
// footer. A table from another format and a table that is not one take the same path and print the same
// note, so the table here is overwritten rather than written by an older build. Both halves are held: the
// default run read the table and rewrote it, so the note had something to say, and --verbose says it.
func TestADefaultRunPrintsNoNotes(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	writeTree(t, root, map[string]string{
		"tsconfig.json":       `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["source"]}`,
		"CohereSettings.json": `{"rules":{}}`,
		"source/index.ts":     "export const value = 1;\n",
		// Left without `.cache/` on purpose: that note is one of the two a default run must not print.
		".gitignore": "node_modules/\n",
	})
	for _, arguments := range [][]string{
		{"init", "--quiet"},
		{"-c", "user.email=fixture@example.com", "-c", "user.name=fixture", "commit", "--quiet", "--allow-empty", "-m", "initial"},
	} {
		command := exec.Command("git", arguments...)
		command.Dir = root
		if output, err := command.CombinedOutput(); err != nil {
			t.Fatalf("git %v: %v\n%s", arguments, err, output)
		}
	}

	// The fixture's JSON is not in the house format, and a run that rewrites it lists the files it cohered
	// above the footer. Formatting is not this test's subject.
	run := func(arguments ...string) (string, int) {
		t.Helper()
		command := exec.Command(binary, append([]string{"--no-format"}, arguments...)...)
		command.Dir = root
		output, err := command.CombinedOutput()
		if err == nil {
			return string(output), 0
		}
		exitError, isExit := err.(*exec.ExitError)
		if !isExit {
			t.Fatalf("running cohere: %v\n%s", err, output)
		}
		return string(output), childExitCode(t, exitError)
	}
	discardTable := func() {
		t.Helper()
		tables := cacheTableFiles(t, root)
		if len(tables) == 0 {
			t.Fatalf("no cache table in %s, so there is nothing for a run to discard", root)
		}
		for _, table := range tables {
			if err := os.WriteFile(table, []byte("not a cache table"), 0o600); err != nil {
				t.Fatal(err)
			}
		}
	}

	if output, code := run(); code != 0 {
		t.Fatalf("exit %d from the first run of a clean project:\n%s", code, output)
	}
	discardTable()

	output, code := run()
	if code != 0 {
		t.Fatalf("exit %d from the run after the discard:\n%s", code, output)
	}
	lines := strings.Split(strings.TrimRight(output, "\n"), "\n")
	if len(lines) != 1 || !strings.HasPrefix(lines[0], "✓ 💎 ") {
		t.Errorf("the default run after a discard printed more than its footer:\n%s", output)
	}
	if strings.Contains(output, "note:") {
		t.Errorf("the default run printed a note:\n%s", output)
	}
	for _, table := range cacheTableFiles(t, root) {
		if contents, err := os.ReadFile(table); err == nil && string(contents) == "not a cache table" {
			t.Fatalf("the run never read the discarded table, so its silence proves nothing: %s", table)
		}
	}

	// The control: the same discard under --verbose says both notes, so the default run had them to hide.
	discardTable()
	verbose, _ := run("--verbose")
	for _, want := range []string{"note: cache table discarded", "does not ignore .cache/"} {
		if !strings.Contains(verbose, want) {
			t.Errorf("--verbose did not say %q after a discard:\n%s", want, verbose)
		}
	}
}
