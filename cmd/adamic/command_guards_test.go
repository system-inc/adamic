package main

import (
	"bytes"
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/toolchain"
)

func TestUsageExitStatus(t *testing.T) {
	t.Parallel()
	code, output, diagnostic := runGuardCommand(t, nil)
	if code != 2 || output != "" || diagnostic != usage+"\n" {
		t.Fatalf("bad usage: got exit %d, stdout %q, stderr %q; want exit 2 and usage text", code, output, diagnostic)
	}
}

// Usage asked for is no usage error: --help and -h print it to stdout and exit 0. Mutant: print it to stderr and
// return 2, as before #a733mfq.
func TestHelpPrintsUsageToStdout(t *testing.T) {
	t.Parallel()
	for _, flag := range []string{"--help", "-h"} {
		code, output, diagnostic := runGuardCommand(t, []string{flag})
		if code != 0 || output != usage+"\n" || diagnostic != "" {
			t.Errorf("%s: got exit %d, stdout %q, stderr %q; want exit 0 and usage on stdout", flag, code, output, diagnostic)
		}
	}
}

// --version names the commit, the Go and the clang, a line each, on stdout with exit 0. The clang line is read here
// apart from the command's own: the path toolchain.Path finds and the first line of its --version. Mutant: drop
// writeVersion's clang line.
func TestVersionNamesCommitGoAndClang(t *testing.T) {
	t.Parallel()
	information, ok := debug.ReadBuildInfo()
	clang := ""
	if compiler, err := toolchain.Path("clang"); err != nil {
		clang = "clang: not found (" + err.Error() + ")"
	} else if report, err := toolchain.Command(compiler, "--version").Output(); err != nil {
		clang = "clang: not found (" + err.Error() + ")"
	} else {
		clang = "clang " + compiler + ": " + strings.TrimSpace(strings.SplitN(string(report), "\n", 2)[0])
	}
	want := "adamic " + revision(information, ok) + "\ngo " + runtime.Version() + "\n" + clang + "\n"
	code, output, diagnostic := runGuardCommand(t, []string{"--version"})
	if code != 0 || output != want || diagnostic != "" {
		t.Fatalf("--version: got exit %d, stdout %q, stderr %q; want exit 0 and stdout %q", code, output, diagnostic, want)
	}
}

// --version is for bug reports, so a clang it can't find is a line, not a failure. Mutant: return 1 from --version
// when toolchain.Path fails.
func TestVersionWithoutClangStillExitsZero(t *testing.T) {
	t.Parallel()
	environment := []string{"PATH=" + t.TempDir(), "ADAMIC_TOOLS=" + t.TempDir()}
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if name != "PATH" && name != "ADAMIC_TOOLS" && name != "ADAMIC_SANDBOX" {
			environment = append(environment, variable)
		}
	}
	code, output, diagnostic := runGuardCommandIn(t, environment, []string{"--version"})
	lines := strings.Split(output, "\n")
	want := `clang: not found (exec: "clang": executable file not found in $PATH)`
	if code != 0 || len(lines) != 4 || lines[2] != want || diagnostic != "" {
		t.Fatalf("--version without clang: got exit %d, stdout %q, stderr %q; want exit 0 and a third line %q", code, output, diagnostic, want)
	}
}

// revision reads go build's stamp: the commit, (modified) for a changed tree, unknown with no stamp or no build info.
func TestRevisionReadsTheStamp(t *testing.T) {
	t.Parallel()
	stamped := func(settings ...debug.BuildSetting) *debug.BuildInfo { return &debug.BuildInfo{Settings: settings} }
	commit := debug.BuildSetting{Key: "vcs.revision", Value: "89c1c7958a9cc358b95fa641a97cca31dd3b718d"}
	for _, test := range []struct {
		information *debug.BuildInfo
		ok          bool
		want        string
	}{
		{stamped(commit), true, commit.Value},
		{stamped(commit, debug.BuildSetting{Key: "vcs.modified", Value: "false"}), true, commit.Value},
		{stamped(commit, debug.BuildSetting{Key: "vcs.modified", Value: "true"}), true, commit.Value + " (modified)"},
		{stamped(debug.BuildSetting{Key: "vcs.modified", Value: "true"}), true, "unknown"},
		{nil, false, "unknown"},
	} {
		if got := revision(test.information, test.ok); got != test.want {
			t.Errorf("revision(%v, %v) = %q, want %q", test.information, test.ok, got, test.want)
		}
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

// --debug is refused on wasm32-wasi rather than ignored: the module would carry no line table. Mutant: drop
// ValidateOptions' Debug refusal, and the build goes on to WASI_SYSROOT or the SDK.
func TestBuildWASIDebugTarget(t *testing.T) {
	t.Parallel()
	source := "testdata/build_target.a"
	assertWASITargetRefusal(t, source, []string{"build", "--target", "wasm32-wasi", source, "-o", filepath.Join(t.TempDir(), "out.wasm"), "--debug"}, "adamic: native: debug builds are not supported for wasm32-wasi\n")
}

func assertWASISanitizeRefusal(t *testing.T, source string, arguments []string) {
	t.Helper()
	assertWASITargetRefusal(t, source, arguments, "adamic: native: sanitizers are not supported for wasm32-wasi\n")
}

func assertWASITargetRefusal(t *testing.T, source string, arguments []string, want string) {
	t.Helper()
	// A failed load must not stand in for the unsupported-options refusal.
	if code, _, diagnostic := runGuardCommand(t, []string{"types", source}); code != 0 || diagnostic != "" {
		t.Fatalf("fixture must load: got exit %d and stderr %q", code, diagnostic)
	}
	code, output, diagnostic := runGuardCommand(t, arguments)
	if code != 1 || output != "" || diagnostic != want {
		t.Fatalf("%v: got exit %d, stdout %q, stderr %q; want exit 1 and stderr %q", arguments, code, output, diagnostic, want)
	}
}

func runGuardCommand(t *testing.T, arguments []string) (int, string, string) {
	t.Helper()
	return runGuardCommandIn(t, os.Environ(), arguments)
}

// runGuardCommandIn is runGuardCommand under an environment of the test's own.
func runGuardCommandIn(t *testing.T, environment []string, arguments []string) (int, string, string) {
	t.Helper()
	// Reuse the CLI test driver, which exits with run's status in an isolated process.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, os.Args[0], append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, arguments...)...)
	command.Env = append(environment, "ADAMIC_EXPLAIN_TEST_DRIVER=1")
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
