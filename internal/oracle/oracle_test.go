// Package oracle holds Adamic's differential test: every fixture runs twice, from source on Node and
// as a native binary stage 0 built, and the two must agree byte for byte on stdout and stderr and
// exactly on the exit code. A disagreement is a miscompile until proven otherwise.
package oracle

import (
	"bytes"
	"context"
	"errors"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// repository is the repository's root, from this package's directory.
const repository = "../.."

// fixtures is every program the oracle knows, and whether stage 0 lowers it yet. A program that
// doesn't lower yet must still be refused loudly, with where and what, so all of them are run every
// time and the list says exactly how far stage 0 has come.
var fixtures = []struct {
	path   string
	lowers bool
}{
	{"dedication/dedication.a", true},
	{"internal/load/testdata/0.1/compile/01_hello.ts", true},
	{"internal/load/testdata/0.1/compile/02_fizzbuzz.ts", true},
	{"internal/load/testdata/0.1/compile/03_shapes.ts", true},
	{"internal/load/testdata/0.1/compile/04_closures.ts", false},
	{"internal/load/testdata/0.1/compile/05_wordcount.ts", false},
	{"internal/load/testdata/0.1/compile/06_stack.ts", false},
	{"internal/load/testdata/0.1/compile/07_modules/main.ts", true},
	{"internal/load/testdata/0.1/compile/08_results.ts", true},
	{"internal/load/testdata/0.1/compile/09_tree.ts", true},
	{"internal/load/testdata/0.1/compile/10_unicode.ts", false},
	{"internal/oracle/testdata/strings.a", true},
	{"internal/oracle/testdata/numbers.a", true},
	{"internal/oracle/testdata/loops.a", true},
	{"internal/oracle/testdata/booleans.a", true},
	{"internal/oracle/testdata/shadowing.a", true},
	{"internal/oracle/testdata/functions.a", true},
	{"internal/oracle/testdata/effects.a", true},
	{"internal/oracle/testdata/dead_zone.a", true},
	{"internal/oracle/testdata/objects.a", true},
	{"internal/oracle/testdata/modules/main.a", true},
	{"internal/oracle/testdata/panic.a", true},
	{"internal/oracle/testdata/maps_and_text.a", true},
}

// run is one execution's observable behavior: what the oracle compares.
type run struct {
	stdout   []byte
	stderr   []byte
	exitCode int
}

func execute(t *testing.T, name string, arguments ...string) run {
	t.Helper()
	command := exec.Command(name, arguments...)
	var stdout, stderr bytes.Buffer
	command.Stdout = &stdout
	command.Stderr = &stderr
	err := command.Run()
	var exitError *exec.ExitError
	if err != nil && !errors.As(err, &exitError) {
		t.Fatalf("running %s: %v", name, err)
	}
	return run{stdout: stdout.Bytes(), stderr: stderr.Bytes(), exitCode: command.ProcessState.ExitCode()}
}

// onNode runs a program's source on Node: the oracle.
func onNode(t *testing.T, path string) run {
	t.Helper()
	return execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), path)
}

// lowered checks and lowers a program, or returns stage 0's refusal.
func lowered(t *testing.T, path string) (*ir.Program, error) {
	t.Helper()
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return lower.Lower(context.Background(), program)
}

// natively builds a lowered program under the sanitizers and runs it.
func natively(t *testing.T, program *ir.Program) run {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	return execute(t, binary)
}

// leaks builds a lowered program without sanitizers (they and macOS's leaks tool don't mix), runs it
// under leaks --atExit, and returns its report when anything leaked. No garbage collector means every
// reference the compiler hands out has to come back; this is where a missing release shows.
func leaks(t *testing.T, program *ir.Program) string {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "program")
	if err := native.Build(native.C(program), binary, native.Options{}); err != nil {
		t.Fatal(err)
	}
	report := execute(t, "leaks", "--atExit", "--", binary)
	if report.exitCode == 0 {
		return ""
	}
	return string(report.stdout)
}

// disagreement says how two runs differ, or "" when they don't.
func disagreement(oracle run, native run) string {
	switch {
	case oracle.exitCode != native.exitCode:
		return "exit codes differ"
	case !bytes.Equal(oracle.stdout, native.stdout):
		return "stdout differs"
	case !bytes.Equal(oracle.stderr, native.stderr):
		return "stderr differs"
	}
	return ""
}

func TestNativeAgreesWithNode(t *testing.T) {
	t.Parallel()
	for _, fixture := range fixtures {
		t.Run(fixture.path, func(t *testing.T) {
			t.Parallel()
			path, err := filepath.Abs(filepath.Join(repository, fixture.path))
			if err != nil {
				t.Fatal(err)
			}
			program, err := lowered(t, path)
			if !fixture.lowers {
				var notYet *lower.NotYet
				if !errors.As(err, &notYet) {
					t.Fatalf("want stage 0 to refuse with where and what, got %v", err)
				}
				t.Logf("not yet: %v", notYet)
				return
			}
			if err != nil {
				t.Fatalf("Lower: %v", err)
			}
			oracle, native := onNode(t, path), natively(t, program)
			if difference := disagreement(oracle, native); difference != "" {
				t.Errorf("%s\nnode:   exit %d, stdout %q, stderr %q\nnative: exit %d, stdout %q, stderr %q",
					difference, oracle.exitCode, oracle.stdout, oracle.stderr, native.exitCode, native.stdout, native.stderr)
			}
			// A program that panicked stopped where it stood, as Node's does, so what it held then
			// isn't a leak; every program that finishes must have let go of everything.
			if oracle.exitCode == 0 {
				if leaked := leaks(t, program); leaked != "" {
					t.Errorf("leaks:\n%s", leaked)
				}
			}
		})
	}
}

// The oracle has to be able to fail. One byte added to the dedication's string, after lowering, must
// read as a disagreement.
func TestTheOracleCatchesOneByte(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "dedication", "dedication.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	program.Strings[0] += "!"
	if difference := disagreement(onNode(t, path), natively(t, program)); difference != "stdout differs" {
		t.Errorf("got %q, want the mutant caught as \"stdout differs\"", difference)
	}
}
