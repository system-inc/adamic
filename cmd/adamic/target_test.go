package main

import (
	"os"
	"strings"
	"testing"
)

// Not parallel: test.Setenv changes the process-wide WASI_SYSROOT environment variable.
func TestBuildTargetParsing(test *testing.T) {
	test.Setenv("WASI_SYSROOT", test.TempDir())
	for _, arguments := range [][]string{
		{"build", "--target", "unknown", "missing.a", "-o", "out.wasm"},
		{"build", "missing.a", "-o", "out", "--target"},
		{"build", "missing.a", "-o", "out", "--target", "wasm32-wasi", "--target", "wasm32-wasi"},
	} {
		if code := run(arguments); code != 2 {
			test.Fatalf("%v: got %d, want usage error (2)", arguments, code)
		}
	}
}

// Not parallel: this test captures the command's process-wide stderr.
func TestWASIRejectsTSGoArchive(test *testing.T) {
	test.Setenv("WASI_SYSROOT", test.TempDir())
	output, err := os.CreateTemp(test.TempDir(), "diagnostic")
	if err != nil {
		test.Fatal(err)
	}
	defer output.Close()
	original := os.Stderr
	os.Stderr = output
	defer func() { os.Stderr = original }()
	code := run(
		[]string{"build", "--target", "wasm32-wasi", "missing.a", "-o", "out.wasm", "--tsgo", "checker.archive"},
	)
	diagnostic, err := os.ReadFile(output.Name())
	if err != nil {
		test.Fatal(err)
	}
	if code != 1 || !strings.Contains(string(diagnostic), "--tsgo is not supported for wasm32-wasi") {
		test.Fatalf("got exit %d and %q", code, diagnostic)
	}
}
