package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestContextualNamesAgree(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("testdata/contextual/*.ts.txt")
	if err != nil || len(paths) != 19 {
		t.Fatalf("keyword fixtures: %d, %v", len(paths), err)
	}
	oracle := goOracle(t)
	directory, err := filepath.Abs(".")
	if err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	if err != nil {
		t.Fatal(err)
	}
	binary := buildPort(t, directory, true)
	for _, fixture := range paths {
		t.Run(filepath.Base(fixture), func(t *testing.T) {
			input, err := os.ReadFile(fixture)
			if err != nil {
				t.Fatal(err)
			}
			path := filepath.Join(t.TempDir(), "input.ts")
			if err := os.WriteFile(path, input, 0644); err != nil {
				t.Fatal(err)
			}
			want := execute(t, "", oracle, path, "--whole", "--recovery")
			for _, side := range []struct {
				name, command string
				args          []string
			}{
				{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), path, "--whole", "--recovery"}},
				{"native", binary, []string{path, "--whole", "--recovery"}},
			} {
				got := execute(t, "", side.command, side.args...)
				if !bytes.Equal(got.output, want.output) {
					t.Errorf("%s %s: %s", side.name, fixture, difference(got.output, want.output))
				}
			}
		})
	}
}

func contextualInput(t *testing.T, keyword string) string {
	t.Helper()
	input, err := os.ReadFile(filepath.Join("testdata/contextual", keyword+".ts.txt"))
	if err != nil {
		t.Fatal(err)
	}
	return string(input)
}
