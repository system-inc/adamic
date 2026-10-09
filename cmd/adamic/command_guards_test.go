package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"
)

func TestUsageExitStatus(t *testing.T) {
	t.Parallel()
	code, output, diagnostic := runGuardCommand(t, nil)
	if code != 2 || output != "" || diagnostic != usage+"\n" {
		t.Fatalf("bad usage: got exit %d, stdout %q, stderr %q; want exit 2 and usage text", code, output, diagnostic)
	}
}

func TestBuildWASISanitizeTargetBeforeSource(t *testing.T) {
	t.Parallel()
	source := "testdata/build_target.a"
	assertWASISanitizeRefusal(t, source, []string{"build", "--target", "wasm32-wasi", source, "-o", filepath.Join(t.TempDir(), "out.wasm"), "--sanitize"})
}

func TestBuildWASISanitizeTargetAfterSource(t *testing.T) {
	t.Parallel()
	source := "testdata/build_target.a"
	assertWASISanitizeRefusal(t, source, []string{"build", source, "-o", filepath.Join(t.TempDir(), "out.wasm"), "--target", "wasm32-wasi", "--sanitize"})
}

func assertWASISanitizeRefusal(t *testing.T, source string, arguments []string) {
	t.Helper()
	// A failed load must not stand in for the unsupported-options refusal.
	if code, _, diagnostic := runGuardCommand(t, []string{"types", source}); code != 0 || diagnostic != "" {
		t.Fatalf("fixture must load: got exit %d and stderr %q", code, diagnostic)
	}
	code, output, diagnostic := runGuardCommand(t, arguments)
	want := "adamic: native: sanitizers are not supported for wasm32-wasi\n"
	if code != 1 || output != "" || diagnostic != want {
		t.Fatalf("%v: got exit %d, stdout %q, stderr %q; want exit 1 and stderr %q", arguments, code, output, diagnostic, want)
	}
}

func runGuardCommand(t *testing.T, arguments []string) (int, string, string) {
	t.Helper()
	// Reuse the CLI test driver, which exits with run's status in an isolated process.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, arguments...)...)
	command.Env = append(os.Environ(), "ADAMIC_EXPLAIN_TEST_DRIVER=1")
	var output, diagnostic bytes.Buffer
	command.Stdout, command.Stderr = &output, &diagnostic
	code := 0
	if err := command.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			code = exit.ExitCode()
		} else {
			t.Fatal(err)
		}
	}
	return code, output.String(), diagnostic.String()
}
