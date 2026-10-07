package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

func init() {
	for _, path := range []string{
		"internal/oracle/testdata/import_cycles/order/main.a",
		"internal/oracle/testdata/import_cycles/runtime/even.a",
		"stage3/drivers/scanner/probes/cyclic-initialized-value/main.a",
		"stage3/drivers/scanner/probes/cyclic-premature-value/main.a",
	} {
		fixtures = append(fixtures, struct {
			path            string
			lowers, checked bool
		}{path, true, false})
	}
}

func TestImportCycleRuntimeCalls(t *testing.T) {
	// Not parallel: the cohere runner caches one program process-wide.
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/import_cycles/runtime/even.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	// Invoke the exports only after all module bodies finish. This keeps the fixture
	// inside the interim policy while exercising actual cross-module function calls.
	runner := filepath.Join(t.TempDir(), "after.mjs")
	source := "import { even } from " + quoteJS(path) + ";\nimport { odd } from " + quoteJS(filepath.Join(filepath.Dir(path), "odd.a")) + ";\nconsole.log(`${even(8)}`);\nconsole.log(`${odd(9)}`);\n"
	if err := os.WriteFile(runner, []byte(source), 0o644); err != nil {
		t.Fatal(err)
	}
	// The generated runner imports fixture source. Run it fresh: the generic .mjs
	// cache hashes only the runner's bytes, not its imported .a modules.
	truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle", "node.mjs"), runner)
	if string(truth.stdout) != "odd body\neven body\n1\n1\n" || truth.exitCode != 0 {
		t.Fatalf("Node: %+v", truth)
	}
	for _, name := range []string{"even", "odd"} {
		index := -1
		for candidate, function := range program.Functions {
			if function.Name == name {
				index = candidate
				break
			}
		}
		if index < 0 {
			t.Fatalf("missing function %s", name)
		}
		argument := 8.0
		if name == "odd" {
			argument = 9
		}
		program.Main = append(program.Main, ir.WriteLine{Stream: ir.Stdout, Value: ir.NumberToString{Value: ir.Call{Function: index, Arguments: []ir.Expression{ir.NumberConstant{Value: argument}}, Returns: ir.Number}}})
	}
	generated := onJavaScriptBackend(t, program)
	native, sanitized := natively(t, program)
	if report := leaks(t, program, sanitized); report != "" {
		t.Fatal(report)
	}
	for _, result := range []run{generated, native, released(t, program)} {
		if difference := disagreement(truth, result); difference != "" {
			t.Fatal(difference)
		}
	}
}

func quoteJS(value string) string {
	// File names here contain no control characters; Go's quoted strings are also JS strings.
	return "'" + strings.ReplaceAll(strings.ReplaceAll(value, "\\", "\\\\"), "'", "\\'") + "'"
}

func TestImportCycleLoadTimeReads(t *testing.T) {
	// Not parallel: the cohere runner caches one program process-wide.
	for _, probe := range []struct{ path, output, nodeError string }{
		{"../imported_const_cases/cycle/main.a", "", "ReferenceError: Cannot access 'value' before initialization"},
		{"read/a.a", "", "ReferenceError: Cannot access 'value' before initialization"},
		{"classes/c.a", "", "ReferenceError: Cannot access 'Base' before initialization"},
		// Entering at Leaf evaluates Base then Middle, so ESM succeeds.
		{"classes/a.a", "base\nmiddle\nleaf\n", ""},
		{"indirect/b.a", "", "ReferenceError: Cannot access 'value' before initialization"},
		{"hoisted/a.a", "1\n", ""},
	} {
		t.Run(probe.path, func(t *testing.T) {
			path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/import_cycles", probe.path))
			if err != nil {
				t.Fatal(err)
			}
			truth := onNode(t, path)
			if string(truth.stdout) != probe.output {
				t.Fatalf("Node stdout: %q", truth.stdout)
			}
			if probe.nodeError != "" {
				if truth.exitCode == 0 || !strings.Contains(string(truth.stderr), probe.nodeError) {
					t.Fatalf("Node: %+v", truth)
				}
			} else if truth.exitCode != 0 {
				t.Fatalf("Node: %+v", truth)
			}
			program, err := lowered(t, path)
			if err != nil {
				t.Fatal(err)
			}
			generated := onJavaScriptBackend(t, program)
			compiled, _ := natively(t, program)
			for _, result := range []run{generated, compiled} {
				if difference := disagreement(truth, result); difference != "" {
					t.Fatal(difference)
				}
			}
			t.Logf("Node stdout=%q exit=%d; native and JavaScript agree", truth.stdout, truth.exitCode)
		})
	}
}
