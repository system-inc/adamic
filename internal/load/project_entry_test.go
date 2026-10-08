package load

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestProjectEntryMustBeConfigured(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for name, source := range map[string]string{
		"entry.a":       "export const entry = 1;",
		"outside.a":     "export const outside = 2;",
		"tsconfig.json": `{"files":["entry.a"]}`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	_, err := LoadProjectEntry(filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "outside.a"))
	if err == nil || !strings.Contains(err.Error(), "outside.a") || !strings.Contains(err.Error(), "add it to files/include") {
		t.Fatalf("outside entry: %v", err)
	}
	_, err = LoadProjectEntry(filepath.Join(directory, "tsconfig.json"), "")
	if err == nil || !strings.Contains(err.Error(), "--entry") {
		t.Fatalf("missing entry: %v", err)
	}
}

func TestProjectChecksUnimportedFiles(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	for name, source := range map[string]string{
		"entry.a":       "export const entry = 1;",
		"unimported.a":  "export const broken: number = 'wrong';",
		"tsconfig.json": `{"files":["entry.a","unimported.a"]}`,
	} {
		if err := os.WriteFile(filepath.Join(directory, name), []byte(source), 0644); err != nil {
			t.Fatal(err)
		}
	}
	_, err := LoadProjectEntry(filepath.Join(directory, "tsconfig.json"), filepath.Join(directory, "entry.a"))
	if err == nil || !strings.Contains(err.Error(), "unimported.a") || !strings.Contains(err.Error(), "TS2322") {
		t.Fatalf("unimported type error: %v", err)
	}
}

func TestProjectEntryNormalizesRelativePaths(t *testing.T) {
	t.Parallel()
	path := "../lower/testdata/project_entry/tsconfig.json"
	program, err := LoadProjectEntry(path, "../lower/testdata/project_entry/entry.a")
	if err != nil {
		t.Fatal(err)
	}
	if len(program.Entries()) != 1 || !strings.HasSuffix(program.FileName(program.Entries()[0]), "/entry.a") {
		t.Fatal("entry was not selected")
	}
}
