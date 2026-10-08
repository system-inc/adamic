package parser

import (
	"bytes"
	"os"
	"path/filepath"
	"testing"
)

func TestIncompleteRecoveryCasesAgree(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob("testdata/incomplete_recovery/*.ts.txt")
	if err != nil || len(paths) != 30 {
		t.Fatalf("incomplete recovery fixtures: %d, %v", len(paths), err)
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
			want, err := recoveryRunLimit(t, incompleteDeadline(len(input)), path+".go", oracle, path, "--whole", "--recovery")
			if err != nil {
				t.Fatal(err)
			}
			for _, side := range []struct {
				name, command string
				args          []string
			}{
				{"Node", "node", []string{"--disable-warning=ExperimentalWarning", runner, filepath.Join(directory, "main.ts"), path, "--whole", "--recovery"}},
				{"native", binary, []string{path, "--whole", "--recovery"}},
			} {
				got, err := recoveryRunLimit(t, incompleteDeadline(len(input)), path+"."+side.name, side.command, side.args...)
				if err != nil {
					t.Fatal(err)
				}
				if !bytes.Equal(got, want) {
					t.Errorf("%s %s: %s", side.name, fixture, difference(got, want))
				}
			}
		})
	}
}
