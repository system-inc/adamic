package oracle

import (
	"fmt"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/toolchain"
)

func TestStackLongArguments(test *testing.T) {
	test.Parallel()
	stackProbe(test, 0, longStackArguments(120000, 3), nil, false)
}
func TestStackArgumentsAlone(test *testing.T) {
	test.Parallel()
	stackProbe(test, 0, longStackArguments(120000, 3), nil, true)
}
func TestStackLongEnvironment(test *testing.T) {
	test.Parallel()
	stackProbe(test, 0, nil, []string{
		"ADAMIC_LONG_1=" + strings.Repeat("x", 120000),
		"ADAMIC_LONG_2=" + strings.Repeat("x", 120000),
		"ADAMIC_LONG_3=" + strings.Repeat("x", 120000),
	}, false)
}
func TestStackLimit1024(test *testing.T) { test.Parallel(); stackProbe(test, 1024, nil, nil, false) }
func TestStackLimit512(test *testing.T)  { test.Parallel(); stackProbe(test, 512, nil, nil, false) }
func TestStackLimit256(test *testing.T) {
	test.Parallel()
	stackProbe(test, 256, longStackArguments(100000, 1), nil, false)
}
func TestStackLimit1024LongArguments(test *testing.T) {
	test.Parallel()
	size, count := 136000, 3
	// Linux limits each string to 128 KiB and all strings to a quarter of the stack.
	if runtime.GOOS == "linux" {
		size, count = 100000, 2
	}
	stackProbe(test, 1024, longStackArguments(size, count), nil, false)
}
func TestStackLimit2048LongArguments(test *testing.T) {
	test.Parallel()
	size, count := 307000, 3
	if runtime.GOOS == "linux" {
		size, count = 120000, 4
	}
	stackProbe(test, 2048, longStackArguments(size, count), nil, false)
}
func longStackArguments(size, count int) []string {
	arguments := make([]string, count)
	for index := range arguments {
		arguments[index] = strings.Repeat("x", size)
	}
	return arguments
}

// Each limit and argument layout is a separate Loom unit, including setup.
func stackProbe(test *testing.T, limit int, arguments, environment []string, alone bool) {
	test.Helper()
	path, sanitizedBinary, _ := sanitized(test, "internal/oracle/testdata/stack_over.a")
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	plain := filepath.Join(test.TempDir(), "plain")
	if err := native.Build(native.C(program), plain, native.Options{}); err != nil {
		test.Fatal(err)
	}
	ran := func(name string, prefix ...string) run {
		test.Helper()
		resolved, err := toolchain.Path(name)
		if err != nil {
			test.Fatal(err)
		}
		args := append(append([]string{}, prefix...), arguments...)
		if limit > 0 {
			args = append([]string{"-c", fmt.Sprintf(`ulimit -s %d && exec "$0" "$@"`, limit), resolved}, args...)
			resolved = "/bin/sh"
		}
		ctx, command := bounded(test, resolved, args...)
		command.Env = append(command.Env, environment...)
		command.Env = append(command.Env, "ASAN_OPTIONS=detect_leaks=0")
		if alone {
			command.Env = []string{}
		}
		stdout, stderr, err := childguard.Capture(ctx, command)
		if err != nil && command.ProcessState == nil {
			test.Fatal(err)
		}
		result := run{stdout: stdout, stderr: stderr, exitCode: command.ProcessState.ExitCode()}
		rememberRun(test, result)
		return result
	}
	want := run{
		stdout:   []byte("start\n"),
		stderr:   []byte("adamic: panic: RangeError: Maximum call stack size exceeded in sum\n"),
		exitCode: 70,
	}
	if limit == 0 || limit >= 1024 {
		nodeArguments := []string{"--disable-warning=ExperimentalWarning"}
		// V8's default recursion budget exceeds the usable stack after long argv at 1 MiB.
		// Depth may differ; both backends must still report the same RangeError panic.
		if limit == 1024 && len(arguments) > 0 {
			nodeArguments = append(nodeArguments, "--stack-size=256")
		}
		nodeArguments = append(nodeArguments, repository+"/oracle/node.mjs", path)
		node := ran("node", nodeArguments...)
		if difference := disagreement(want, node); difference != "" {
			test.Fatalf("Node: %s: exit %d stderr %q", difference, node.exitCode, node.stderr)
		}
	}
	for _, binary := range []string{sanitizedBinary, plain} {
		got := ran(binary)
		if difference := disagreement(want, got); difference != "" {
			test.Errorf("%s: %s: exit %d stderr %.300q", filepath.Base(binary), difference, got.exitCode, got.stderr)
		}
	}
}
