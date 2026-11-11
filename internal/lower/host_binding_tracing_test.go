package lower

import (
	"testing"

	"github.com/system-inc/adamic/internal/ir"
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
