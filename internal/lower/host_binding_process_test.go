package lower

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

const processDeclaration = "declare const process: { nextTick?: unknown; browser?: unknown; } | undefined;\n"

func processModules(body string) map[string]string {
	return map[string]string{
		"core.ts": processDeclaration + "export function run(): void {\nif (typeof process !== 'undefined') {\n" + body + "\n}\n}\n",
		"main.ts": "import 'node:process';\nimport { run } from './core.ts';\nrun();\n",
	}
}

// The real row clears both model admission walls. Plain reads must be IR undefined, never a slot.
func TestHostBindingProcessReads(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"const browser = process.browser; console.log(typeof browser);",
		"console.log(typeof process.browser);",
		"const browser = process?.browser; console.log(typeof browser);",
		"console.log(String(typeof process !== 'undefined' && !!process.nextTick && !process.browser));",
	} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			program, err := lowerModules(t, processModules(body))
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			undefined := false
			check := func(value any) bool {
				if _, ok := value.(ir.Undefined); ok {
					undefined = true
				}
				if property, ok := value.(ir.Property); ok && property.Name == "browser" {
					t.Error("browser read from a slot")
				}
				return true
			}
			walk(program.Main, check)
			for _, function := range program.Functions {
				walk(function.Body, check)
			}
			// typeof may simplify undefined to its string before IR is walked.
			if !undefined && !strings.Contains(body, "typeof process.browser") {
				t.Error("no IR undefined for browser")
			}
		})
	}
}

// Mutation, reflection and later views stay stopped with browser named, on the shared lowering.
func TestHostBindingProcessBrowserRoadsStop(t *testing.T) {
	t.Parallel()
	for _, body := range []string{
		"process.browser = 1;",
		"const p: { browser?: unknown } = process; p.browser = 1;",
		"Object.assign(process, { browser: 1 });",
		"Object.defineProperty(process, 'browser', { value: 1 });",
		"process['browser'] = 1;",
		"delete process.browser;",
		"console.log('browser' in process);",
		"const k = 'browser'; console.log(k in process);",
		"const k = 'browser'; console.log(process[k]);",
		"console.log(process['browser']);",
		"console.log(Reflect.get(process, 'browser'));",
		"const p: { browser?: unknown } = process; console.log(p.browser);",
		"const { browser } = process; console.log(browser);",
		"const { ['browser']: browser } = process; console.log(typeof browser);",
		"const p = { ...process }; console.log(p.browser);",
	} {
		t.Run(body, func(t *testing.T) {
			t.Parallel()
			_, err := lowerModules(t, processModules(body))
			var notYet *NotYet
			var refused *Refused
			what := ""
			switch {
			case errors.As(err, &notYet):
				what = notYet.What
			case errors.As(err, &refused):
				what = refused.What + "; " + refused.Fix
			}
			// The temporary path contains the subtest's browser spelling; only the payload counts.
			if !strings.Contains(what, "browser") {
				t.Fatalf("want a stop naming browser, got %v", err)
			}
			if strings.HasPrefix(body, "const k") && what != "a run-time key on process, which could name process.browser" {
				t.Fatalf("want the run-time key diagnostic, got %v", err)
			}
			if strings.HasPrefix(body, "const p:") && !strings.Contains(what, "a member no host registration names") {
				t.Fatalf("want the unregistered browser view stop, got %v", err)
			}
			t.Log(err)
		})
	}
	// Language's roads use the same program's core row and read it after the attempted write.
	for _, probe := range []struct{ name, body, prefix, module string }{
		{"q1 globalThis assign", "Object.assign(globalThis.process, { browser: true });", "", ""},
		{"q2 globalThis defineProperty", "Object.defineProperty(globalThis.process, 'browser', { value: true });", "", ""},
		{"q3 globalThis Reflect.set", "Reflect.set(globalThis.process, 'browser', true);", "", ""},
		{"q1 bracket assign", "Object.assign(globalThis['process'], { browser: true });", "", ""},
		{"q2 bracket defineProperty", "Object.defineProperty(globalThis['process'], 'browser', { value: true });", "", ""},
		{"q3 bracket Reflect.set", "Reflect.set(globalThis['process'], 'browser', true);", "", ""},
		{"q4 imported alias", "Object.assign(nodeProcess, { browser: true });", "import nodeProcess from 'node:process';\n", ""},
		{"q6 another module's global", "mutate();", "import { mutate } from './mutator.ts';\n", "export function mutate(): void { Object.assign(process, { browser: true }); }\n"},
		{"assert non-null assign", "Object.assign(process!, { browser: true });", "", ""},
		{"assert satisfies assign", "Object.assign((process satisfies object), { browser: true });", "", ""},
		{"assert angle assign", "Object.assign(<any>process, { browser: true });", "", ""},
		{"assert non-null write", "(process!).browser = true;", "", ""},
		{"q7 cast receiver", "(process as { browser?: unknown }).browser = true;", "", ""},
		{"q8 destructuring assignment", "({ browser: process.browser } = { browser: true });", "", ""},
		{"q9 for-of target", "for (process.browser of [true]) {}", "", ""},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			modules := processModules(probe.body + "\nconsole.log(String(process.browser));")
			modules["core.ts"] = probe.prefix + modules["core.ts"]
			if probe.module != "" {
				modules["mutator.ts"] = probe.module
			}
			_, err := lowerModules(t, modules)
			var notYet *NotYet
			var refused *Refused
			if !errors.As(err, &notYet) && !errors.As(err, &refused) {
				t.Fatalf("mutation road did not stop; compare its run with Node before changing lowering: %v", err)
			}
			if (strings.HasPrefix(probe.name, "q1 ") || strings.HasPrefix(probe.name, "q2 ") || strings.HasPrefix(probe.name, "q3 ") || strings.HasPrefix(probe.name, "q4 ") || strings.HasPrefix(probe.name, "assert ")) && (notYet == nil || notYet.What != "process.browser outside a plain property read") {
				t.Fatalf("want rule 70's browser stop, got %v", err)
			}
			t.Log(err)
		})
	}
	// s1 witnesses the process-as-a-value stop that the absent-member fold also depends on.
	t.Run("s1 parameter view read", func(t *testing.T) {
		t.Parallel()
		modules := processModules("read(process);\nconsole.log(String(process.browser));")
		modules["core.ts"] += "function read(p: { browser?: unknown }): void { console.log(String(p.browser)); }\n"
		_, err := lowerModules(t, modules)
		var notYet *NotYet
		if !errors.As(err, &notYet) || notYet.What != "process as a value (use its supported process operation directly)" {
			t.Fatalf("want the process-as-a-value stop, got %v", err)
		}
	})
}

// Admission is by the model export and member name, with both optionality and undefined required.
// These call the declaration judgment directly so a changed type cannot stop at row text matching.
func TestHostBindingProcessModel(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, members, want string }{
		{"export equals", "nextTick?: unknown; browser?: unknown;", ""},
		{"unlisted bogus", "nextTick?: unknown; browser?: unknown; bogus?: unknown;", "bogus its model export node:process.process doesn't have"},
		{"required browser", "nextTick?: unknown; browser: unknown;", "browser its model export node:process.process doesn't have"},
		{"unlisted release", "nextTick?: unknown; browser?: unknown; release?: unknown;", "which reaches node:process.Process.release"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			directory := t.TempDir()
			file := filepath.Join(directory, "core.ts")
			source := "import 'node:process';\ndeclare const process: { " + probe.members + " } | undefined;\nconsole.log(typeof process);\n"
			if err := os.WriteFile(file, []byte(source), 0o644); err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{file})
			if err != nil {
				t.Fatal(err)
			}
			core := program.Files()[0]
			checked, release := program.Checker(t.Context(), core)
			defer release()
			l := &lowering{program: program, checker: checked}
			declaration := core.Statements.Nodes[1].AsVariableStatement().DeclarationList.AsVariableDeclarationList().Declarations.Nodes[0]
			row := hostBindingNamed("process")
			if row == nil {
				t.Fatal("no process row")
			}
			model := l.hostBindingModel(*row)
			if model == nil {
				t.Fatal("export= process model not found")
			}
			module := checked.TryFindAmbientModule("node:process")
			if module.Exports[ast.InternalSymbolNameExportEquals] == nil {
				t.Fatal("node:process has no export=")
			}
			if l.hostBindingExportEqualsModel(module, "other") != nil {
				t.Error("export= admitted another name")
			}
			err = l.hostBindingModelRefusal(declaration, *row)
			if probe.want == "" {
				if err != nil {
					t.Fatal(err)
				}
				view := checked.GetNonNullableType(checked.GetTypeAtLocation(declaration.Name()))
				member := checked.GetPropertyOfType(view, "browser")
				registered, host := l.nodeHostObject(view)
				if !host || registered(member) {
					t.Error("browser registered for a later view")
				}
			} else {
				var notYet *NotYet
				if !errors.As(err, &notYet) || !strings.Contains(notYet.What, probe.want) {
					t.Fatalf("want stop containing %q, got %v", probe.want, err)
				}
			}
		})
	}
}

// Node's own object, including its entire prototype chain, holds every table entry absent in both
// module modes. Adding an entry for version must go red here (#zqz01vh, Adamic and Lower, Nov 8).
func TestHostBindingAbsentMembersHeldToNode(t *testing.T) {
	t.Parallel()
	node, err := exec.LookPath("node")
	if err != nil {
		t.Fatal(err)
	}
	for model, members := range hostBindingAbsentMembers {
		for member := range members {
			for _, mode := range []string{"module", "commonjs"} {
				t.Run(model+"/"+member+"/"+mode, func(t *testing.T) {
					t.Parallel()
					at := strings.LastIndex(model, ".")
					module, _ := json.Marshal(model[:at])
					key, _ := json.Marshal(member)
					source := "import object from " + string(module) + ";\n"
					if mode == "commonjs" {
						source = "const object = require(" + string(module) + ");\n"
					}
					source += "const name = " + string(key) + "; console.log(Object.getOwnPropertyDescriptor(object, name)); console.log(Object.hasOwn(object, name)); console.log(name in object); if (Object.hasOwn(object, name) || name in object) process.exitCode = 1;"
					output, err := exec.CommandContext(t.Context(), node, "--input-type="+mode, "--eval", source).CombinedOutput()
					if err != nil || string(output) != "undefined\nfalse\nfalse\n" {
						t.Fatalf("Node: %v, output %s", err, output)
					}
					t.Logf("Node %s: %s", mode, output)
				})
			}
		}
	}
}
