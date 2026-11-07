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

func TestUsageExitStatus(test *testing.T) {
	test.Parallel()
	code, output, diagnostic := runGuardCommand(test, nil)
	if code != 2 || output != "" || diagnostic != usage+"\n" {
		test.Fatalf(
			"bad usage: got exit %d, stdout %q, stderr %q; want exit 2 and usage text",
			code,
			output,
			diagnostic,
		)
	}
}

// Usage asked for is no usage error: --help and -h print it to stdout and exit 0. Mutant: print it to stderr and
// return 2, as before #a733mfq.
func TestHelpPrintsUsageToStdout(test *testing.T) {
	test.Parallel()
	for _, flag := range []string{"--help", "-h"} {
		code, output, diagnostic := runGuardCommand(test, []string{flag})
		if code != 0 || output != usage+"\n" || diagnostic != "" {
			test.Errorf(
				"%s: got exit %d, stdout %q, stderr %q; want exit 0 and usage on stdout",
				flag,
				code,
				output,
				diagnostic,
			)
		}
	}
}

// --version names the commit, the Go and the clang, a line each, on stdout with exit 0. The clang line is read here
// apart from the command's own: the path toolchain.Path finds and the first line of its --version. Mutant: drop
// writeVersion's clang line.
func TestVersionNamesCommitGoAndClang(test *testing.T) {
	test.Parallel()
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
	code, output, diagnostic := runGuardCommand(test, []string{"--version"})
	if code != 0 || output != want || diagnostic != "" {
		test.Fatalf(
			"--version: got exit %d, stdout %q, stderr %q; want exit 0 and stdout %q",
			code,
			output,
			diagnostic,
			want,
		)
	}
}

// --version is for bug reports, so a clang it can't find is a line, not a failure. Mutant: return 1 from --version
// when toolchain.Path fails.
func TestVersionWithoutClangStillExitsZero(test *testing.T) {
	test.Parallel()
	environment := []string{"PATH=" + test.TempDir(), "ADAMIC_TOOLS=" + test.TempDir()}
	for _, variable := range os.Environ() {
		name, _, _ := strings.Cut(variable, "=")
		if name != "PATH" && name != "ADAMIC_TOOLS" && name != "ADAMIC_SANDBOX" {
			environment = append(environment, variable)
		}
	}
	code, output, diagnostic := runGuardCommandIn(test, environment, []string{"--version"})
	lines := strings.Split(output, "\n")
	want := `clang: not found (exec: "clang": executable file not found in $PATH)`
	if code != 0 || len(lines) != 4 || lines[2] != want || diagnostic != "" {
		test.Fatalf(
			"--version without clang: got exit %d, stdout %q, stderr %q; want exit 0 and a third line %q",
			code,
			output,
			diagnostic,
			want,
		)
	}
}

// revision reads go build's stamp: the commit, (modified) for a changed tree, unknown with no stamp or no build info.
func TestRevisionReadsTheStamp(test *testing.T) {
	test.Parallel()
	stamped := func(settings ...debug.BuildSetting) *debug.BuildInfo { return &debug.BuildInfo{Settings: settings} }
	commit := debug.BuildSetting{Key: "vcs.revision", Value: "89c1c7958a9cc358b95fa641a97cca31dd3b718d"}
	for _, testCase := range []struct {
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
		if got := revision(testCase.information, testCase.ok); got != testCase.want {
			test.Errorf("revision(%v, %v) = %q, want %q", testCase.information, testCase.ok, got, testCase.want)
		}
	}
}

func TestBuildWASISanitizeTargetBeforeSource(test *testing.T) {
	test.Parallel()
	source := "testdata/build_target.a"
	assertWASISanitizeRefusal(test, source, []string{
		"build",
		"--target",
		"wasm32-wasi",
		source,
		"-o",
		filepath.Join(test.TempDir(), "out.wasm"),
		"--sanitize",
	})
}

func TestBuildWASISanitizeTargetAfterSource(test *testing.T) {
	test.Parallel()
	source := "testdata/build_target.a"
	assertWASISanitizeRefusal(test, source, []string{
		"build",
		source,
		"-o",
		filepath.Join(test.TempDir(), "out.wasm"),
		"--target",
		"wasm32-wasi",
		"--sanitize",
	})
}

// --debug is refused on wasm32-wasi rather than ignored: the module would carry no line table. Mutant: drop
// ValidateOptions' Debug refusal, and the build goes on to WASI_SYSROOT or the SDK.
func TestBuildWASIDebugTarget(test *testing.T) {
	test.Parallel()
	source := "testdata/build_target.a"
	assertWASITargetRefusal(test, source, []string{
		"build",
		"--target",
		"wasm32-wasi",
		source,
		"-o",
		filepath.Join(test.TempDir(), "out.wasm"),
		"--debug",
	}, "adamic: native: debug builds are not supported for wasm32-wasi\n")
}

func assertWASISanitizeRefusal(test *testing.T, source string, arguments []string) {
	test.Helper()
	assertWASITargetRefusal(test, source, arguments, "adamic: native: sanitizers are not supported for wasm32-wasi\n")
}

func assertWASITargetRefusal(test *testing.T, source string, arguments []string, want string) {
	test.Helper()
	// A failed load must not stand in for the unsupported-options refusal.
	if code, _, diagnostic := runGuardCommand(test, []string{"types", source}); code != 0 || diagnostic != "" {
		test.Fatalf("fixture must load: got exit %d and stderr %q", code, diagnostic)
	}
	code, output, diagnostic := runGuardCommand(test, arguments)
	if code != 1 || output != "" || diagnostic != want {
		test.Fatalf(
			"%v: got exit %d, stdout %q, stderr %q; want exit 1 and stderr %q",
			arguments,
			code,
			output,
			diagnostic,
			want,
		)
	}
}

func runGuardCommand(test *testing.T, arguments []string) (int, string, string) {
	test.Helper()
	return runGuardCommandIn(test, os.Environ(), arguments)
}

// runGuardCommandIn is runGuardCommand under an environment of the test's own.
func runGuardCommandIn(test *testing.T, environment []string, arguments []string) (int, string, string) {
	test.Helper()
	// Reuse the CLI test driver, which exits with run's status in an isolated process.
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(
		ctx,
		os.Args[0],
		append([]string{"-test.run=^TestExplainChecksDriver$", "--"}, arguments...)...,
	)
	command.Env = append(environment, "ADAMIC_EXPLAIN_TEST_DRIVER=1")
	var output, diagnostic bytes.Buffer
	command.Stdout, command.Stderr = &output, &diagnostic
	code := 0
	if err := command.Run(); err != nil {
		if exit, ok := err.(*exec.ExitError); ok {
			code = exit.ExitCode()
		} else {
			test.Fatal(err)
		}
	}
	return code, output.String(), diagnostic.String()
}
