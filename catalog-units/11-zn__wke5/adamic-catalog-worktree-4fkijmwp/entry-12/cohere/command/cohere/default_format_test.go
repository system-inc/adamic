package main

import (
	"path/filepath"
	"strings"
	"testing"
)

// TestABareRunFormats is Kirk's ruling through the binary (#b1sjy7b, 2026-10-04: "running cohere should
// type check lint fix format all in one call"): a bare cohere formats, --fix formats, --no-fix reports
// formatting and exits on it, and --no-format leaves it out of all three.
func TestABareRunFormats(t *testing.T) {
	t.Parallel()
	binary := buildCohere(t)
	unformatted := "export const ugly   =   1;\n"
	formatted := "export const ugly = 1;\n"
	fixture := func(t *testing.T) string {
		t.Helper()
		root := t.TempDir()
		writeTree(t, root, map[string]string{
			"tsconfig.json":            fixScopeTsconfig,
			"CohereSettings.json":      "{ \"extends\": \"./NexusCohereSettings.json\", \"rules\": { \"no-debugger\": \"error\" } }\n",
			"NexusCohereSettings.json": "{ \"format\": { \"ignore\": [\"tsconfig.json\"] } }\n",
			".gitignore":               ".cache/\n",
			"Ugly.ts":                  unformatted,
			"Debugged.ts":              "export function value(): number {\n  debugger;\n  return 1;\n}\n",
		})
		return root
	}
	ugly := func(root string) string { return filepath.Join(root, "Ugly.ts") }

	for _, arguments := range [][]string{{}, {"--fix"}} {
		t.Run("cohere "+strings.Join(arguments, " ")+" formats", func(t *testing.T) {
			t.Parallel()
			root := fixture(t)
			output, _ := runCohere(t, binary, root, arguments...)
			if got := readForTest(t, ugly(root)); got != formatted {
				t.Fatalf("the file was left unformatted:\n%q\n%s", got, output)
			}
			if got := readForTest(t, filepath.Join(root, "Debugged.ts")); strings.Contains(got, "debugger") {
				t.Fatalf("the fix did not land beside the format:\n%s", got)
			}
		})
	}

	// Formatting by default keeps a warm run replayable: once the tree is formatted, an unchanged bare run is
	// recorded and the next one replays it.
	t.Run("a warm bare run still replays", func(t *testing.T) {
		t.Parallel()
		root := fixture(t)
		runCohere(t, binary, root)
		runCohere(t, binary, root)
		output, code := runCohere(t, binary, root)
		if code != 0 || !strings.HasPrefix(output, "cached: ") {
			t.Fatalf("a warm bare run over a formatted tree did not replay, exit %d:\n%s", code, output)
		}
	})

	t.Run("cohere --no-fix reports the misformat and exits on it", func(t *testing.T) {
		t.Parallel()
		root := fixture(t)
		output, code := runCohere(t, binary, root, "--no-fix")
		if code == 0 || !strings.Contains(output, ugly(root)+":1:1 - --fix would rewrite this file: format [fix/would-change]") {
			t.Fatalf("exit %d, and the misformat was not a finding:\n%s", code, output)
		}
		if got := readForTest(t, ugly(root)); got != unformatted {
			t.Fatalf("--no-fix wrote the file:\n%s", got)
		}
	})

	for _, arguments := range [][]string{{"--no-format"}, {"--fix", "--no-format"}, {"--no-fix", "--no-format"}} {
		t.Run("cohere "+strings.Join(arguments, " ")+" leaves formatting out", func(t *testing.T) {
			t.Parallel()
			root := fixture(t)
			output, _ := runCohere(t, binary, root, arguments...)
			if got := readForTest(t, ugly(root)); got != unformatted {
				t.Fatalf("--no-format formatted the file:\n%s", got)
			}
			if strings.Contains(output, ugly(root)) {
				t.Fatalf("--no-format named the misformat:\n%s", output)
			}
			if !strings.Contains(output, "(formatting left out (--no-format))") {
				t.Fatalf("the phase line does not say formatting was left out:\n%s", output)
			}
		})
	}

	t.Run("--no-format refuses a flag that asks for formatting", func(t *testing.T) {
		t.Parallel()
		root := fixture(t)
		for _, other := range []string{"--format", "--format-all", "--format-only"} {
			output, code := runCohere(t, binary, root, "--no-format", other)
			if code == 0 || !strings.Contains(output, "--no-format and "+other+" contradict each other") {
				t.Fatalf("--no-format with %s was not refused by name, exit %d:\n%s", other, code, output)
			}
		}
	})
}
