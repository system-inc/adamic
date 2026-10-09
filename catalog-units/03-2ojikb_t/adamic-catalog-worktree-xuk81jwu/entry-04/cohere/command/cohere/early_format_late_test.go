package main

import (
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"
)

// A bare run's walk starts without waiting for the early format pass's scope, and takes the scope once the walk
// is done (#g3046x5). A scope that arrives late must find exactly what a prompt one does: the same unformatted
// files, the same findings, the same exit. COHERE_TEST_LATE_EARLY_FORMAT holds the early pass before it
// enumerates, so its scope is ready only after this small tree's walk has finished, as a cold ahra run's
// often was. The held run has to take longer than the hold, or the instrument did nothing and the comparison
// proves nothing.
func TestALateEarlyFormatScopeFindsWhatAPromptOneDoes(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	root := t.TempDir()
	files := map[string]string{
		"tsconfig.json":       `{"compilerOptions":{"strict":true,"noEmit":true,"target":"es2022","module":"esnext","moduleResolution":"bundler"},"include":["source"]}`,
		"CohereSettings.json": `{"rules":{"no-debugger":"error"}}`,
		// Unformatted TypeScript, in the program.
		"source/Spaced.ts": "export const spaced   =   1\n",
		// A lint finding, formatted.
		"source/Halt.ts": "export function halt(): void {\n    debugger;\n}\n",
		// Unformatted markdown, outside the program: the scope is its only way into the run.
		"notes.md": "#   Title\n\n\n\nText.\n",
	}
	for name, contents := range files {
		path := filepath.Join(root, name)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(contents), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	const hold = 600 * time.Millisecond
	timing := regexp.MustCompile(`[0-9]+(\.[0-9]+)?(ms|µs|s)\b`)
	run := func(held bool) (string, int) {
		t.Helper()
		command := exec.Command(binary, "--no-fix", "--no-cache")
		command.Dir = root
		command.Env = append(os.Environ(), "COHERE_TEST_LATE_EARLY_FORMAT=")
		if held {
			command.Env = append(command.Env, "COHERE_TEST_LATE_EARLY_FORMAT="+hold.String())
		}
		started := time.Now()
		output, _ := command.CombinedOutput()
		if held && time.Since(started) < hold {
			t.Fatalf("the held run took %s, under the %s hold, so the early pass was not held:\n%s", time.Since(started), hold, output)
		}
		return timing.ReplaceAllString(string(output), "T"), command.ProcessState.ExitCode()
	}
	prompt, promptExit := run(false)
	late, lateExit := run(true)

	for _, expected := range []string{"Spaced.ts", "notes.md", "no-debugger"} {
		if !strings.Contains(prompt, expected) {
			t.Fatalf("the prompt run does not mention %s, so the fixture proves nothing:\n%s", expected, prompt)
		}
	}
	if late != prompt || lateExit != promptExit {
		t.Errorf("a late early-format scope changed the run (exit %d against %d)\n--- late\n%s\n--- prompt\n%s", lateExit, promptExit, late, prompt)
	}
}
