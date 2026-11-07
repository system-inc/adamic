package oracle

import (
	"bytes"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Regular destinations always run. Only the terminal overlay requires a host
// with script(1) and /dev/ptmx; setup.sh does not promise either to Loom runners.
func stdoutIsTTYAgreement(test *testing.T, transport string) {
	test.Helper()
	path, binary, js := sanitized(test, "internal/oracle/testdata/host_stdout/istty.a")
	loader, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	node, err := toolchain.Path("node")
	if err != nil {
		test.Fatal(err)
	}
	variants := []struct {
		name string
		args []string
	}{
		{node, []string{"--disable-warning=ExperimentalWarning", loader, path}},
		{binary, nil},
		{node, []string{"--disable-warning=ExperimentalWarning", loader, js}},
	}
	// The stock units always compare Node/native/JS. WASI is the existing
	// opt-in oracle overlay, rather than a tool prerequisite for those units.
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		program, err := lowered(test, path)
		if err != nil {
			test.Fatal(err)
		}
		wasm := filepath.Join(test.TempDir(), "program.wasm")
		if err := native.Build(native.C(program), wasm, native.Options{Target: "wasm32-wasi"}); err != nil {
			test.Fatal(err)
		}
		wasi, _ := filepath.Abs(filepath.Join(repository, "oracle/wasi.mjs"))
		variants = append(variants, struct {
			name string
			args []string
		}{node, []string{"--disable-warning=ExperimentalWarning", wasi, wasm}})
	}
	test.Logf(
		"%s: comparing %d backends (WASI overlay %t)",
		transport,
		len(variants),
		os.Getenv("ADAMIC_ORACLE_WASI") == "1",
	)
	var truth run
	for i, v := range variants {
		got := stdoutIsTTYRun(test, transport, v.name, v.args...)
		if i == 0 {
			truth = got
			expected := "undefined false true\n"
			if transport == "pty" {
				expected = "true true false\n"
			}
			// stderr observes the value even when stdout is /dev/null.
			if truth.exitCode != 0 || string(truth.stderr) != expected {
				test.Fatalf("Node %s witness changed: %+v", transport, truth)
			}
		} else if d := disagreement(truth, got); d != "" {
			test.Fatalf("backend %d, %s: %s; Node %+v; got %+v", i, transport, d, truth, got)
		}
	}
}

func stdoutIsTTYRun(test *testing.T, transport, name string, args ...string) run {
	test.Helper()
	if transport == "pty" {
		quote := func(s string) string { return "'" + strings.ReplaceAll(s, "'", "'\"'\"'") + "'" }
		words := []string{quote(name)}
		for _, s := range args {
			words = append(words, quote(s))
		}
		// Separate stderr before script joins its child's streams into the pty.
		// This leaves the same data channels as the other transport comparisons.
		errorFile := filepath.Join(test.TempDir(), "stderr")
		command := strings.Join(words, " ") + " 2>" + quote(errorFile)
		ctx, cmd := bounded(test, "script", "-qec", command, "/dev/null")
		cmd.Env = []string{
			"PATH=/usr/bin:/bin",
			"TERM=dumb",
			"ASAN_OPTIONS=detect_leaks=1",
			"UBSAN_OPTIONS=halt_on_error=1",
		}
		out, errout, err := childguard.Capture(ctx, cmd)
		var exited *exec.ExitError
		if err != nil && !errors.As(err, &exited) {
			test.Fatal(err)
		}
		if len(errout) != 0 {
			test.Fatalf("script stderr: %s", errout)
		}
		data, err := os.ReadFile(errorFile)
		if err != nil {
			test.Fatal(err)
		}
		return run{stdout: out, stderr: data, exitCode: cmd.ProcessState.ExitCode()}
	}
	_, cmd := bounded(test, name, args...)
	cmd.Env = []string{"TERM=dumb", "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}
	var out, errout bytes.Buffer
	cmd.Stdout = &out
	cmd.Stderr = &errout
	var destination *os.File
	if transport != "pipe" {
		var err error
		if transport == "null" {
			destination, err = os.OpenFile(os.DevNull, os.O_WRONLY, 0)
		} else {
			destination, err = os.Create(filepath.Join(test.TempDir(), "stdout"))
		}
		if err != nil {
			test.Fatal(err)
		}
		defer destination.Close()
		cmd.Stdout = destination
	}
	err := childguard.Run(cmd, childguard.Options{})
	var exited *exec.ExitError
	if err != nil && !errors.As(err, &exited) {
		test.Fatal(err)
	}
	if transport == "file" {
		b, err := os.ReadFile(destination.Name())
		if err != nil {
			test.Fatal(err)
		}
		out.Write(b)
	}
	return run{stdout: out.Bytes(), stderr: errout.Bytes(), exitCode: cmd.ProcessState.ExitCode()}
}

func TestStdoutIsTTYMatchesNodeOnPipe(test *testing.T) {
	test.Parallel()
	stdoutTTYColumns(test, "pipe")
}
func TestStdoutIsTTYMatchesNodeOnFile(test *testing.T) {
	test.Parallel()
	stdoutTTYColumns(test, "file")
}
func TestStdoutIsTTYMatchesNodeOnNull(test *testing.T) {
	test.Parallel()
	stdoutTTYColumns(test, "null")
}
func TestStdoutIsTTYMatchesNodeOnPty(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_STDOUT_PTY") != "1" {
		test.Skip(
			"set ADAMIC_ORACLE_STDOUT_PTY=1 to run the stdout isTTY pty oracle (Platforms runs this lane by hand)",
		)
	}
	stdoutIsTTYAgreement(test, "pty")
	stdoutColumnsAgreement(test, "pty")
	stdoutColumnsLifecycle(test)
}
