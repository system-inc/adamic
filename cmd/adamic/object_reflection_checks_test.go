package main

import (
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: build diagnostics use the process-wide stderr stream.
func TestObjectReflectionBuildReportsChecks(t *testing.T) {
	source, err := os.ReadFile("../../internal/oracle/testdata/object_reflection_ruling/scanner.a")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	path := filepath.Join(directory, "scanner.ts")
	if err = os.WriteFile(path, source, 0644); err != nil {
		t.Fatal(err)
	}
	saved := os.Stderr
	read, write, err := os.Pipe()
	if err != nil {
		t.Fatal(err)
	}
	os.Stderr = write
	code := build(path, filepath.Join(directory, "program"), []string{"--sanitize"})
	os.Stderr = saved
	write.Close()
	output, err := io.ReadAll(read)
	read.Close()
	if err != nil {
		t.Fatal(err)
	}
	if code != 0 || !strings.Contains(string(output), "adamic: checks: object reflection 1\n") {
		t.Fatalf("build code %d diagnostics %q", code, output)
	}
}
