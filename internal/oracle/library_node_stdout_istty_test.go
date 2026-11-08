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
func stdoutIsTtyAgreement(test *testing.T, transport string) {
	test.Helper()
	path, binary, js := sanitized(test, "internal/oracle/testdata/host_stdout/istty.a")
	loader, _ := filepath.Abs(filepath.Join(repository, "oracle/node.mjs"))
	node, pathError := toolchain.Path("node")
	if pathError != nil {
		test.Fatal(pathError)
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
		program, loweredError := lowered(test, path)
		if loweredError != nil {
			test.Fatal(loweredError)
		}
		wasm := filepath.Join(test.TempDir(), "program.wasm")
		if buildError := native.Build(
			native.C(program),
			wasm,
			native.Options{Target: "wasm32-wasi"},
		); buildError != nil {
			test.Fatal(buildError)
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
	for index, variant := range variants {
		got := stdoutIsTtyRun(test, transport, variant.name, variant.args...)
		if index == 0 {
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
			test.Fatalf("backend %d, %s: %s; Node %+v; got %+v", index, transport, d, truth, got)
		}
	}
}

func stdoutIsTtyRun(test *testing.T, transport, name string, args ...string) run {
	test.Helper()
	if transport == "pty" {
		quote := func(text string) string { return "'" + strings.ReplaceAll(text, "'", "'\"'\"'") + "'" }
		words := []string{quote(name)}
		for _, text := range args {
			words = append(words, quote(text))
		}
		// Separate stderr before script joins its child's streams into the pty.
		// This leaves the same data channels as the other transport comparisons.
		errorFile := filepath.Join(test.TempDir(), "stderr")
		command := strings.Join(words, " ") + " 2>" + quote(errorFile)
		ctx, osCommand := bounded(test, "script", "-qec", command, "/dev/null")
		osCommand.Env = []string{
			"PATH=/usr/bin:/bin",
			"TERM=dumb",
			"ASAN_OPTIONS=detect_leaks=1",
			"UBSAN_OPTIONS=halt_on_error=1",
		}
		out, errout, captureError := childguard.Capture(ctx, osCommand)
		var exited *exec.ExitError
		if captureError != nil && !errors.As(captureError, &exited) {
			test.Fatal(captureError)
		}
		if len(errout) != 0 {
			test.Fatalf("script stderr: %s", errout)
		}
		data, readError := os.ReadFile(errorFile)
		if readError != nil {
			test.Fatal(readError)
		}
		return run{stdout: out, stderr: data, exitCode: osCommand.ProcessState.ExitCode()}
	}
	_, command := bounded(test, name, args...)
	command.Env = []string{"TERM=dumb", "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1"}
	var out, errout bytes.Buffer
	command.Stdout = &out
	command.Stderr = &errout
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
		command.Stdout = destination
	}
	runError := childguard.Run(command, childguard.Options{})
	var exited *exec.ExitError
	if runError != nil && !errors.As(runError, &exited) {
		test.Fatal(runError)
	}
	if transport == "file" {
		content, readError := os.ReadFile(destination.Name())
		if readError != nil {
			test.Fatal(readError)
		}
		out.Write(content)
	}
	return run{stdout: out.Bytes(), stderr: errout.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

func TestStdoutIsTtyMatchesNodeOnPipe(test *testing.T) {
	test.Parallel()
	stdoutTtyColumns(test, "pipe")
}
func TestStdoutIsTtyMatchesNodeOnFile(test *testing.T) {
	test.Parallel()
	stdoutTtyColumns(test, "file")
}
func TestStdoutIsTtyMatchesNodeOnNull(test *testing.T) {
	test.Parallel()
	stdoutTtyColumns(test, "null")
}
func TestStdoutIsTtyMatchesNodeOnPty(test *testing.T) {
	test.Parallel()
	if os.Getenv("ADAMIC_ORACLE_STDOUT_PTY") != "1" {
		test.Skip(
			"set ADAMIC_ORACLE_STDOUT_PTY=1 to run the stdout isTTY pty oracle (Platforms runs this lane by hand)",
		)
	}
	stdoutIsTtyAgreement(test, "pty")
	stdoutColumnsAgreement(test, "pty")
	stdoutColumnsLifecycle(test)
}
