package main

import (
	"os"
	"strings"
	"testing"
)

func TestBuildTargetParsing(t *testing.T) {
	t.Setenv("WASI_SYSROOT", t.TempDir())
	for _, arguments := range [][]string{
		{"build", "--target", "wasm32-wasi", "missing.a", "-o", "out.wasm", "--sanitize"},
		{"build", "missing.a", "-o", "out.wasm", "--target", "wasm32-wasi", "--sanitize"},
	} {
		if code := run(arguments); code != 1 {
			t.Fatalf("%v: got %d, want unsupported options (1)", arguments, code)
		}
	}
	for _, arguments := range [][]string{
		{"build", "--target", "unknown", "missing.a", "-o", "out.wasm"},
		{"build", "missing.a", "-o", "out", "--target"},
		{"build", "missing.a", "-o", "out", "--target", "wasm32-wasi", "--target", "wasm32-wasi"},
	} {
		if code := run(arguments); code != 2 {
			t.Fatalf("%v: got %d, want usage error (2)", arguments, code)
		}
	}
}

// Not parallel: this test captures the command's process-wide stderr.
func TestWASIRejectsTSGoArchive(t *testing.T) {
	t.Setenv("WASI_SYSROOT", t.TempDir())
	output, err := os.CreateTemp(t.TempDir(), "diagnostic")
	if err != nil {
		t.Fatal(err)
	}
	defer output.Close()
	original := os.Stderr
	os.Stderr = output
	defer func() { os.Stderr = original }()
	code := run([]string{"build", "--target", "wasm32-wasi", "missing.a", "-o", "out.wasm", "--tsgo", "checker.archive"})
	diagnostic, err := os.ReadFile(output.Name())
	if err != nil {
		t.Fatal(err)
	}
	if code != 1 || !strings.Contains(string(diagnostic), "--tsgo is not supported for wasm32-wasi") {
		t.Fatalf("got exit %d and %q", code, diagnostic)
	}
}
