package main

import (
	"bytes"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
)

// `adamic build <file> -o <relative path>` leaves the binary at that path under the working directory. native links
// from inside its own temporary build directory, so a relative -o used to land in there and go with it, exiting 0
// with nothing written (Language, Oct 26). Mutant: drop buildWithNativeExports's filepath.Abs, and the binary is
// missing.
func TestBuildWritesARelativeOutputUnderTheWorkingDirectory(t *testing.T) {
	t.Parallel()
	source, err := filepath.Abs("testdata/build_target.a")
	if err != nil {
		t.Fatal(err)
	}
	directory := t.TempDir()
	command := exec.Command(os.Args[0], "-test.run=^TestExplainChecksDriver$", "--", "build", source, "-o", "program")
	command.Dir = directory
	command.Env = append(os.Environ(), "ADAMIC_EXPLAIN_TEST_DRIVER=1")
	var output, diagnostic bytes.Buffer
	command.Stdout, command.Stderr = &output, &diagnostic
	if err := command.Run(); err != nil {
		t.Fatalf("adamic build -o program: %v, %s", err, diagnostic.Bytes())
	}
	binary := filepath.Join(directory, "program")
	result, err := exec.Command(binary).CombinedOutput()
	if err != nil || string(result) != "ready\n" {
		t.Fatalf("the binary a relative -o names: %v, %q; want it under the working directory, printing ready", err, result)
	}
}
