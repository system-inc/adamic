package lower

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// tracingProcessDeclaration is the last statement adaptation 65 leaves in tsc's src/compiler/tracing.ts
// (stage3/adapt/65-temporary-node-builtins/adapt.cjs), the declaration of tracing.ts's process row in
// internal/load/host_bindings.json.
const tracingProcessDeclaration = "declare const process: { readonly pid: number };\n"

// tracingPid is a function reading process.pid, as tsc's startTracing does to name its trace files
// (tracing.ts:82-83).
const tracingPid = "export function traceName(): string {\n\treturn \"trace.\" + String(process.pid) + \".json\";\n}\n"

// tracingModules is a program whose file holds tracingPid and then declares process as tracing.ts's
// last statement does, on the line after it, and whose main.ts imports node:process, the row's model, and
// calls traceName.
func tracingModules(file string) map[string]string {
	return map[string]string{
		file:      tracingPid + tracingProcessDeclaration,
		"main.ts": "import 'node:process';\nimport { traceName } from './" + file + "';\nconsole.log(traceName());\n",
	}
}

// pidRouted reports whether a program reads process.pid through node:process's model, the host's pid.
func pidRouted(program *ir.Program) bool {
	routed := false
	find := func(value any) bool {
		if call, ok := value.(ir.ProcessCall); ok && call.Operation == "pid" {
			routed = true
		}
		return true
	}
	walk(program.Main, find)
	for _, function := range program.Functions {
		walk(function.Body, find)
	}
	return routed
}

// A name has a row in each file that declares it (#00qn5az, Adamic on Nov 11), so tracing.ts's process
// is its own row beside core.ts's, Provided through node:process, and its pid read, in a function above
// the declaration as tsc's startTracing is, is the host's. Each declaration is matched only to its own
// file's row: tracing.ts's text in core.ts, or in a file no row names, stays refused as an ambient value
// declaration (TestHostBindingUnmatchedDeclarationIsRefused's last two rows). core.ts's own process
// still lowers by its row: TestHostBindingProcessReads (host_binding_process_test.go) pins it alone, and
// the program here declaring both pins it beside tracing.ts's. Mutant (n),
// internal/oracle/testdata/host_bindings/mutants.md: hostBindingNamed taking the row by name alone again
// finds core.ts's row, the first process row, for tracing.ts's declaration, whose type text isn't
// core.ts's, so tracing.ts's declaration is refused as an ambient value declaration and both rows here go
// red.
func TestHostBindingTracingProcessIsItsOwnRow(t *testing.T) {
	t.Parallel()
	t.Run("tracing.ts's process", func(t *testing.T) {
		t.Parallel()
		program, err := lowerModules(t, tracingModules("tracing.ts"))
		if err != nil {
			t.Fatalf("want tracing.ts's process to lower by its row, got %v", err)
		}
		if !pidRouted(program) {
			t.Error("process.pid didn't lower to the host's pid")
		}
	})
	t.Run("core.ts's process beside tracing.ts's", func(t *testing.T) {
		t.Parallel()
		modules := tracingModules("tracing.ts")
		modules["core.ts"] = processDeclaration + "export function nodeLike(): boolean {\n\treturn typeof process !== \"undefined\" && !!process.nextTick;\n}\n"
		modules["main.ts"] = "import 'node:process';\nimport { nodeLike } from './core.ts';\nimport { traceName } from './tracing.ts';\nconsole.log(String(nodeLike()));\nconsole.log(traceName());\n"
		program, err := lowerModules(t, modules)
		if err != nil {
			t.Fatalf("want each process to lower by its own file's row, got %v", err)
		}
		if !pidRouted(program) {
			t.Error("tracing.ts's process.pid didn't lower to the host's pid")
		}
	})
}

// The text adaptation 65 wrote before Nov 11, `{ pid: number }`, isn't the row's: @types/node declares pid
// read only, and a writable pid over the model's read only one is a view that can write the host's pid.
// In tracing.ts it's refused at the declaration, since the row holds the read only text. Judged against
// the model directly, as TestHostBindingProcessModel judges core.ts's, it stops on that field, the stop
// performanceCore.ts's writable timeOrigin met (host_binding_route_test.go's a2 and b2). Mutant (m2) of
// host_binding_route_test.go, hostBindingModelRefusal's widened check dropped, admits it there.
func TestHostBindingTracingWritablePidStops(t *testing.T) {
	t.Parallel()
	const writable = "declare const process: { pid: number };\n"
	t.Run("the row's text", func(t *testing.T) {
		t.Parallel()
		modules := tracingModules("tracing.ts")
		modules["tracing.ts"] = tracingPid + writable
		_, err := lowerModules(t, modules)
		const want = "tracing.ts:4:1: Adamic 0.1 refuses an ambient value declaration; define the values in a module and import them; use Adamic-generated platform bindings for host values"
		var refused *Refused
		if !errors.As(err, &refused) || !strings.HasSuffix(err.Error(), want) {
			t.Errorf("got %v, want %s", err, want)
		}
	})
	t.Run("the model's read only pid", func(t *testing.T) {
		t.Parallel()
		directory := t.TempDir()
		file := filepath.Join(directory, "tracing.ts")
		if err := os.WriteFile(file, []byte("import 'node:process';\n"+writable+"console.log(typeof process);\n"), 0o644); err != nil {
			t.Fatal(err)
		}
		program, err := load.Load([]string{file})
		if err != nil {
			t.Fatal(err)
		}
		tracing := program.Files()[0]
		checked, release := program.Checker(t.Context(), tracing)
		defer release()
		l := &lowering{program: program, checker: checked}
		declaration := tracing.Statements.Nodes[1].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0]
		row := hostBindingNamed("process", "tracing.ts")
		if row == nil {
			t.Fatal("no tracing.ts process row")
		}
		const want = "whose field pid is writable where its model export node:process.process's is read only,"
		var notYet *NotYet
		if err := l.hostBindingModelRefusal(declaration, *row); !errors.As(err, &notYet) || !strings.Contains(notYet.What, want) {
			t.Errorf("got %v, want a stop containing %q", err, want)
		}
	})
}
