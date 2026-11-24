package lower

import (
	"bytes"
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/native"
	"github.com/system-inc/adamic/internal/native_runtime"
)

// firstLinkTwoFunctions calls a function that prints, then one whose body stops today: Math.random is
// NotYet (refusedRandom), at main.a:5:9.
const firstLinkTwoFunctions = "function first(): void {\n\tconsole.log('first ran');\n}\nfunction second(): number {\n\treturn Math.random();\n}\nfirst();\nconsole.log(`${second()}`);\n"

// firstLinkNested is a function whose own work prints before it calls a nested function whose body stops
// today, at main.a:4:10.
const firstLinkNested = "function outer(): number {\n\tconsole.log('outer ran');\n\tfunction inner(): number {\n\t\treturn Math.random();\n\t}\n\treturn inner();\n}\nconsole.log(`${outer()}`);\n"

// lowerFirstLink lowers a .a program, as a first-link lowering when traps is set.
func lowerFirstLink(t *testing.T, source string, traps bool) (*ir.Program, error) {
	t.Helper()
	program, err := load.Load([]string{writeAgreementSource(t, "main.a", source)})
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	return LowerWithOptions(context.Background(), program, Options{FirstLinkTraps: traps})
}

// firstLinkRun is one backend's run of a lowered program: what it printed and its exit code.
type firstLinkRun struct {
	backend        string
	code           int
	stdout, stderr string
}

// firstLinkRuns builds a lowered program as sanitized native and on the JavaScript backend, and runs each.
func firstLinkRuns(t *testing.T, program *ir.Program) []firstLinkRun {
	t.Helper()
	directory := t.TempDir()
	binary := filepath.Join(directory, "native")
	if err := native.Build(native.C(program), binary, native_runtime.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	runner, err := filepath.Abs("../../oracle/node.mjs")
	if err != nil {
		t.Fatal(err)
	}
	generated := filepath.Join(directory, "generated.mjs")
	if err := os.WriteFile(generated, []byte(javascript.JavaScript(program)), 0o644); err != nil {
		t.Fatal(err)
	}
	commands := []struct {
		backend string
		command *exec.Cmd
	}{
		{"native", exec.Command(binary)},
		{"javascript", exec.Command("node", "--disable-warning=ExperimentalWarning", runner, generated)},
	}
	runs := []firstLinkRun{}
	for _, one := range commands {
		one.command.Env = append(os.Environ(), "ASAN_OPTIONS=detect_leaks=1", "UBSAN_OPTIONS=halt_on_error=1")
		var out, errs bytes.Buffer
		one.command.Stdout, one.command.Stderr = &out, &errs
		code := 0
		if err := one.command.Run(); err != nil {
			var exit *exec.ExitError
			if !errors.As(err, &exit) {
				t.Fatal(err)
			}
			code = exit.ExitCode()
		}
		runs = append(runs, firstLinkRun{backend: one.backend, code: code, stdout: out.String(), stderr: errs.String()})
	}
	return runs
}

// firstLinkTrapped holds a program's runs to its trap: each prints stdout, then exits 75 with the
// first-link prefix and a NotYet naming its site.
func firstLinkTrapped(t *testing.T, program *ir.Program, stdout, site string) {
	t.Helper()
	for _, run := range firstLinkRuns(t, program) {
		if run.code != 75 || run.stdout != stdout || !strings.HasPrefix(run.stderr, "adamic: first-link trap: NotYet: ") || !strings.Contains(run.stderr, site) {
			t.Errorf("%s: exit %d stdout %q stderr %q; want 75, %q, and a first-link NotYet at %s", run.backend, run.code, run.stdout, run.stderr, stdout, site)
		}
	}
}

// A first-link lowering lowers a body that stops as its trap, and the program runs up to it: the first
// function prints, then the second traps with exit 75, on native and on the JavaScript backend, and the
// program's Traps holds the one row. Mutant: adamic_first_link_trap and firstLinkTrap exit 0, and this
// goes red.
func TestFirstLinkTrapRunsUpToTheStop(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLink(t, firstLinkTwoFunctions, true)
	if err != nil {
		t.Fatal(err)
	}
	want := ir.Trap{Kind: "NotYet", Reason: "Math.random", Site: "main.a:5:9", Unit: "main.a second", Phase: "Body"}
	if len(program.Traps) != 1 || program.Traps[0] != want {
		t.Fatalf("Traps: got %+v, want %+v", program.Traps, want)
	}
	firstLinkTrapped(t, program, "first ran\n", "main.a:5:9")
}

// Without the option, the same program stops at lowering with the same NotYet, word for word as Lower stops it.
func TestFirstLinkTrapOffStopsAtCompileTime(t *testing.T) {
	t.Parallel()
	_, err := lowerFirstLink(t, firstLinkTwoFunctions, false)
	var notYet *NotYet
	if !errors.As(err, &notYet) || !strings.HasSuffix(err.Error(), "main.a:5:9: stage 0 can't lower Math.random yet") {
		t.Fatalf("got %v, want the NotYet for Math.random at main.a:5:9", err)
	}
	_, stop := lowerSource(t, firstLinkTwoFunctions)
	if stop == nil || stop.Error() != err.Error() {
		t.Fatalf("got %v, want Lower's stop %v", err, stop)
	}
}

// A nested function whose body stops lowers as its own trap alone: the function it's nested in still runs
// its own work before it calls the nested one, which traps.
func TestFirstLinkTrapNestedStubsOnlyItself(t *testing.T) {
	t.Parallel()
	program, err := lowerFirstLink(t, firstLinkNested, true)
	if err != nil {
		t.Fatal(err)
	}
	want := ir.Trap{Kind: "NotYet", Reason: "Math.random", Site: "main.a:4:10", Unit: "main.a outer.inner", Phase: "Body"}
	if len(program.Traps) != 1 || program.Traps[0] != want {
		t.Fatalf("Traps: got %+v, want %+v", program.Traps, want)
	}
	firstLinkTrapped(t, program, "outer ran\n", "main.a:4:10")
}
