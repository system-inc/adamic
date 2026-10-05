package oracle

import (
	"bytes"
	"path/filepath"
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
				return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
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
	}
}
