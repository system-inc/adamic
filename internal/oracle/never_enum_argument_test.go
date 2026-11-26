package oracle

import (
	"errors"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
)

// A numeric enum's never read passed to a never parameter is taken as the argument, through value()'s
// enum trap (enumNeverTrapped, #xfpsqtn): the four shapes tsc writes, a const from a call inside an arrow
// (binder, binder.ts:405), a parameter (reverse, utilities.ts:8142), a const declared in an arrow
// (preferred, moduleSpecifiers.ts:231) and a field read (statistic, executeCommandLine.ts:1285), and
// Compiler's row, one assertNever call reached by a switch and by a chain of ifs (two_paths). Each
// program in range agrees with Node on the JavaScript backend and natively under the leak check. Each
// _out program runs the same calls and then one with `n as E` outside the enum, which an open enum holds:
// Node's assertNever throws there, and Adamic traps at the read, before the call. Node's words differ, so
// stdout and the exit are held to Node and stderr to the trap's own words. The stops are pinned below,
// and the mutants in testdata/never_enum_argument/mutants.md.
func TestNeverEnumArgumentAgreesWithNode(t *testing.T) {
	t.Parallel()
	for name, stdout := range map[string]string{
		"binder":    "2 1 0\n",
		"reverse":   "1 0 2\n",
		"preferred": "2,0,1 1,0,2 3,0,2,1\n",
		"statistic": "Files: 12\nCheck time: 1.23s\nMemory used: 57K\n",
		"two_paths": "left right l r\n",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_enum_argument", name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 0 || string(truth.stdout) != stdout {
				t.Fatalf("Node must print %q: exit %d, stdout %q, stderr %q", stdout, truth.exitCode, truth.stdout, truth.stderr)
			}
			backend := onJavaScriptBackend(t, program)
			if difference := disagreement(truth, backend); difference != "" {
				t.Fatalf("javascript: %s; stdout %q stderr %q exit %d", difference, backend.stdout, backend.stderr, backend.exitCode)
			}
			actual, leaked, comparable := nativelyCheckingLeaks(t, program, true)
			if leaked != "" {
				t.Fatalf("leaks: %s", leaked)
			}
			if comparable {
				if difference := disagreement(truth, actual); difference != "" {
					t.Fatalf("native: %s; stdout %q stderr %q exit %d", difference, actual.stdout, actual.stderr, actual.exitCode)
				}
			}
		})
	}
}

func TestNeverEnumArgumentTrapsBeforeTheCall(t *testing.T) {
	t.Parallel()
	for _, pin := range []struct {
		name, stdout, value, identity string
	}{
		{"binder_out", "2 1 0\n", "7", "ModuleInstanceState"},
		{"reverse_out", "1 0 2\n", "7", "AccessKind"},
		{"preferred_out", "2,0,1 1,0,2 3,0,2,1\n", "9", "ModuleSpecifierEnding"},
		{"statistic_out", "Files: 12\nCheck time: 1.23s\nMemory used: 57K\n", "7", "StatisticType"},
		{"two_paths_out_switch", "left right l r\n", "5", "Side"},
		{"two_paths_out_if", "left right l r\n", "6", "Side"},
	} {
		t.Run(pin.name, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_enum_argument", pin.name+".ts"))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if truth.exitCode != 70 || string(truth.stdout) != pin.stdout || !strings.Contains(string(truth.stderr), "Illegal value: "+pin.value+"\n") {
				t.Fatalf("Node must print %q and throw assertNever's Illegal value: %s: exit %d, stdout %q, stderr %q", pin.stdout, pin.value, truth.exitCode, truth.stdout, truth.stderr)
			}
			// The trap's own words, which assertNever's throw can't print: the read stopped before the call.
			want := run{stdout: truth.stdout, stderr: []byte("adamic: panic: unreachable value " + pin.value + " for numeric enum " + pin.identity + "\n"), exitCode: 70}
			native, _, _ := nativelyCheckingLeaks(t, program, false)
			for backend, result := range map[string]run{"native": native, "javascript": onJavaScriptBackend(t, program)} {
				if difference := disagreement(want, result); difference != "" {
					t.Errorf("%s: %s; exit %d, stdout %q, stderr %q", backend, difference, result.exitCode, result.stdout, result.stderr)
				}
			}
		})
	}
}

// The stops keep today's words: a never read with a write between it and its narrowing (write_between,
// enumNeverReadGap), and an argument that is no never read at all (not_a_read).
func TestNeverEnumArgumentStops(t *testing.T) {
	t.Parallel()
	const condition = "a call whose argument for a never parameter isn't a rulings.md:76 never read"
	for name, location := range map[string]string{
		"write_between": "20:21",
		"not_a_read":    "9:13",
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			file := name + ".ts"
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/never_enum_argument", file))
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, path)
			var notYet *lower.NotYet
			if !errors.As(err, &notYet) || !strings.Contains(err.Error(), file+":"+location+": stage 0 can't lower "+condition+" yet") {
				t.Fatalf("want a stop at %s:%s with %q, got %v", file, location, condition, err)
			}
		})
	}
}
