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

func TestImportCycleRuntimeCalls(test *testing.T) {
	test.Parallel()
	path, absolutePathError := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/import_cycles/runtime/even.a"),
	)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	program, loweredError := lowered(test, path)
	if loweredError != nil {
		test.Fatal(loweredError)
	}
	// Invoke the exports only after all module bodies finish. This keeps the fixture
	// inside the interim policy while exercising actual cross-module function calls.
	runner := filepath.Join(test.TempDir(), "after.mjs")
	source := "import { even } from " +
		quoteJs(path) +
		";\nimport { odd } from " +
		quoteJs(filepath.Join(filepath.Dir(path), "odd.a")) +
		";\nconsole.log(`${even(8)}`);\nconsole.log(`${odd(9)}`);\n"
	if writeError := os.WriteFile(runner, []byte(source), 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	// The generated runner imports fixture source. Run it fresh: the generic .mjs
	// cache hashes only the runner's bytes, not its imported .a modules.
	truth := execute(
		test,
		"node",
		"--disable-warning=ExperimentalWarning",
		filepath.Join(repository, "oracle", "node.mjs"),
		runner,
	)
	if string(truth.stdout) != "odd body\neven body\n1\n1\n" || truth.exitCode != 0 {
		test.Fatalf("Node: %+v", truth)
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
			test.Fatalf("missing function %s", name)
		}
		argument := 8.0
		if name == "odd" {
			argument = 9
		}
		program.Main = append(program.Main, ir.WriteLine{Stream: ir.Stdout, Value: ir.NumberToString{Value: ir.Call{
			Function:  index,
			Arguments: []ir.Expression{ir.NumberConstant{Value: argument}},
			Returns:   ir.Number,
		}}})
	}
	generated := onJavaScriptBackend(test, program)
	native, sanitized := natively(test, program)
	if report := leaks(test, program, sanitized); report != "" {
		test.Fatal(report)
	}
	for _, result := range []run{generated, native, released(test, program)} {
		if difference := disagreement(truth, result); difference != "" {
			test.Fatal(difference)
		}
	}
}

func quoteJs(value string) string {
	// File names here contain no control characters; Go's quoted strings are also JS strings.
	return "'" + strings.ReplaceAll(strings.ReplaceAll(value, "\\", "\\\\"), "'", "\\'") + "'"
}

func TestImportCycleLoadTimeReads(test *testing.T) {
	test.Parallel()
	for _, probe := range []struct{ path, output, nodeError string }{
		{"read/a.a", "", "ReferenceError: Cannot access 'value' before initialization"},
		{"classes/c.a", "", "ReferenceError: Cannot access 'Base' before initialization"},
		// Entering at Leaf evaluates Base then Middle, so ESM succeeds.
		{"classes/a.a", "base\nmiddle\nleaf\n", ""},
		{"indirect/b.a", "", "ReferenceError: Cannot access 'value' before initialization"},
		{"hoisted/a.a", "1\n", ""},
	} {
		test.Run(probe.path, func(test *testing.T) {
			path, absolutePathError := filepath.Abs(
				filepath.Join(repository, "internal/oracle/testdata/import_cycles", probe.path),
			)
			if absolutePathError != nil {
				test.Fatal(absolutePathError)
			}
			truth := onNode(test, path)
			if string(truth.stdout) != probe.output {
				test.Fatalf("Node stdout: %q", truth.stdout)
			}
			if probe.nodeError != "" {
				if truth.exitCode == 0 || !strings.Contains(string(truth.stderr), probe.nodeError) {
					test.Fatalf("Node: %+v", truth)
				}
			} else if truth.exitCode != 0 {
				test.Fatalf("Node: %+v", truth)
			}
			program, loweredError := lowered(test, path)
			if loweredError != nil {
				test.Fatal(loweredError)
			}
			generated := onJavaScriptBackend(test, program)
			compiled, _ := natively(test, program)
			for _, result := range []run{generated, compiled} {
				if difference := disagreement(truth, result); difference != "" {
					test.Fatal(difference)
				}
			}
			test.Logf("Node stdout=%q exit=%d; native and JavaScript agree", truth.stdout, truth.exitCode)
		})
	}
}
