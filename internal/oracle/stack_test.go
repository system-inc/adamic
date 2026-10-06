package oracle

import (
	"bytes"
	"fmt"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/native"
)

// The stack's limit counts from its top, and above the first frame sit the program's arguments and
// its environment. Three arguments of 120 KB, or three variables that long, are stack the limit has to
// leave room for, or deep recursion crashes before the check fires (stack.c).
func TestLongArgumentsLeaveTheStackItsLimit(t *testing.T) {
	t.Parallel()
	long := strings.Repeat("x", 120_000)
	for _, setting := range []struct {
		name        string
		arguments   []string
		environment []string

		// alone runs with no environment at all: any one variable's string sits above argv and would
		// set the top, so only an empty environment leaves the arguments alone to set it. (A panic exits
		// through _exit, so LeakSanitizer, on by default there, never runs.)
		alone bool
	}{
		{"arguments", []string{long, long, long}, nil, false},
		{"arguments alone", []string{long, long, long}, nil, true},
		{"environment", nil, []string{"ADAMIC_LONG_1=" + long, "ADAMIC_LONG_2=" + long, "ADAMIC_LONG_3=" + long}, false},
	} {
		t.Run(setting.name, func(t *testing.T) {
			t.Parallel()
			cacheProbe(t, "internal/oracle/testdata/stack_over.a", nil, "", func() {

				path, binary, _ := sanitized(t, "internal/oracle/testdata/stack_over.a")
				// And built as adamic build builds it: under ASan the stack's layout hides the case of
				// arguments alone, which crashed the -O2 build that ships.
				plain := filepath.Join(t.TempDir(), "plain")
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				if err := native.Build(native.C(program), plain, native.Options{}); err != nil {
					t.Fatal(err)
				}
				ran := func(name string, arguments ...string) run {
					t.Helper()
					environment := append([]string{}, setting.environment...)
					environment = append(environment, "ASAN_OPTIONS=detect_leaks=0")
					if !setting.alone {
						return executeWith(t, environment, name, append(arguments, setting.arguments...)...)
					}
					command := bounded(t, name, append(arguments, setting.arguments...)...)
					command.Env = []string{}
					var stdout, stderr bytes.Buffer
					command.Stdout, command.Stderr = &stdout, &stderr
					if err := command.Run(); err != nil && command.ProcessState == nil {
						t.Fatalf("running %s: %v", name, err)
					}
					result := run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
					rememberRun(t, result)
					return result
				}
				runner := repository + "/oracle/node.mjs"
				node := ran("node", "--disable-warning=ExperimentalWarning", runner, path)
				if node.exitCode != 70 {
					t.Fatalf("want Node to run out of stack, exit 70, got %d, stderr %q", node.exitCode, node.stderr)
				}
				for _, built := range []string{binary, plain} {
					native := ran(built)
					if difference := disagreement(node, native); difference != "" {
						t.Errorf("%s: %s\nnode:   exit %d, stdout %q, stderr %q\nnative: exit %d, stdout %q, stderr %.300q", filepath.Base(built), difference, node.exitCode, node.stdout, node.stderr, native.exitCode, native.stdout, native.stderr)
					}
				}

			})
		})
	}
}

// A small stack still gets a limit (stack.c). At 1 MiB Node runs out of stack and panics, and native
// must too; below that Node itself crashes, so what's held there is docs/0.1.md's own rule, that
// running out of stack is a panic and never a bare crash. At 256 KiB a quarter of the stack is less
// than the 128 KiB Linux still allows arguments, so a 100 KB one is passed there. Each runs under sh's
// ulimit, sanitized and as adamic build builds it.
func TestSmallStacksStillPanic(t *testing.T) {
	t.Parallel()
	for _, setting := range []struct {
		kibibytes int
		arguments []string
	}{
		{1024, nil},
		{512, nil},
		{256, []string{strings.Repeat("x", 100_000)}},
	} {
		kibibytes := setting.kibibytes
		t.Run(strconv.Itoa(kibibytes)+" KiB", func(t *testing.T) {
			t.Parallel()
			cacheProbe(t, "internal/oracle/testdata/stack_over.a", nil, "", func() {

				path, binary, _ := sanitized(t, "internal/oracle/testdata/stack_over.a")
				plain := filepath.Join(t.TempDir(), "plain")
				program, err := lowered(t, path)
				if err != nil {
					t.Fatal(err)
				}
				if err := native.Build(native.C(program), plain, native.Options{}); err != nil {
					t.Fatal(err)
				}
				limited := func(name string, arguments ...string) run {
					t.Helper()
					script := fmt.Sprintf(`ulimit -s %d && exec "$0" "$@"`, kibibytes)
					arguments = append(append([]string{"-c", script, name}, arguments...), setting.arguments...)
					return executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=0"}, "/bin/sh", arguments...)
				}
				want := run{stdout: []byte("start\n"), stderr: []byte("adamic: panic: RangeError: Maximum call stack size exceeded\n"), exitCode: 70}
				if kibibytes >= 1024 {
					node := limited("node", "--disable-warning=ExperimentalWarning", repository+"/oracle/node.mjs", path)
					if difference := disagreement(want, node); difference != "" {
						t.Fatalf("want Node to run out of stack and panic at %d KiB: exit %d, stderr %q", kibibytes, node.exitCode, node.stderr)
					}
				}
				for _, built := range []string{binary, plain} {
					native := limited(built)
					if difference := disagreement(want, native); difference != "" {
						t.Errorf("%s at %d KiB: %s: exit %d, stdout %q, stderr %.300q", filepath.Base(built), kibibytes, difference, native.exitCode, native.stdout, native.stderr)
					}
				}

			})
		})
	}
}
