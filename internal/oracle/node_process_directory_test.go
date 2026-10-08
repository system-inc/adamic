package oracle

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// Runtime and emitter proof only: shared declaration and source-lowering integration is pending.
func TestNodeProcessDirectoryRuntime(t *testing.T) {
	t.Parallel()
	// Canonical, so cwd() reads back what chdir was given: macOS's /tmp is a link to /private/tmp.
	scratch, err := filepath.EvalSymlinks(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	child, missing := filepath.Join(scratch, "child"), filepath.Join(scratch, "absent")
	if err := os.Mkdir(child, 0700); err != nil {
		t.Fatal(err)
	}
	source, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/node_process_runtime/node_process_directory_mutation.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := execute(t, "node", "--disable-warning=ExperimentalWarning", filepath.Join(repository, "oracle/node.mjs"), source, scratch, child, missing)
	if truth.exitCode != 0 || len(truth.stderr) != 0 || !strings.HasPrefix(string(truth.stdout), "true\ntrue\ntrue\nENOENT\nENOENT:") {
		t.Fatalf("unexpected Node result: %+v", truth)
	}
	program := &ir.Program{Source: filepath.Base(source), Strings: []string{scratch, child, missing}, Locals: []ir.Local{{Name: "memoized", Type: ir.String, Function: -1}, {Name: "error", Type: ir.Object, Function: -1}}}
	cwd := ir.ProcessCall{Operation: "cwd", Of: ir.String}
	change := func(index int) ir.Statement {
		return ir.Evaluate{Value: ir.ProcessCall{Operation: "chdir", Of: ir.Object, Arguments: []ir.Expression{ir.StringConstant{Index: index}}}}
	}
	equal := func(value ir.Expression, index int) ir.Statement {
		return ir.WriteLine{Stream: ir.Stdout, Value: ir.BooleanToString{Value: ir.Binary{Operator: ir.Equal, Left: value, Right: ir.StringConstant{Index: index}}}}
	}
	errorValue := ir.Read{Local: 1, Of: ir.Object}
	program.Main = []ir.Statement{
		ir.Evaluate{Value: cwd}, change(0), equal(cwd, 0), ir.Declare{Local: 0, Value: cwd}, change(1), equal(cwd, 1), equal(ir.Read{Local: 0, Of: ir.String}, 0),
		ir.Try{HasCatch: true, CatchLocal: 1, Body: []ir.Statement{change(2)}, Catch: []ir.Statement{
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: errorValue, Name: "code", Of: ir.String}},
			ir.WriteLine{Stream: ir.Stdout, Value: ir.Property{Object: errorValue, Name: "message", Of: ir.String}},
		}},
	}
	binary := filepath.Join(t.TempDir(), "directory")
	if err := native.Build(native.C(program), binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	script := filepath.Join(t.TempDir(), "directory.mjs")
	if err := os.WriteFile(script, []byte(javascript.JavaScript(program)), 0600); err != nil {
		t.Fatal(err)
	}
	for _, got := range []run{execute(t, binary), onNode(t, script)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Fatalf("%s\nNode %q\ngot %q %q", difference, truth.stdout, got.stdout, got.stderr)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
	data, err := os.ReadFile(filepath.Join(repository, "internal/native/runtime/node_process.c"))
	if err != nil {
		t.Fatal(err)
	}
	for name, anchor := range map[string]string{"cache invalidation": "adamic_release(adamic_library_cwd_value); adamic_library_cwd_value = NULL;", "error code": "case ENOENT: code = \"ENOENT\";"} {
		t.Run(name, func(t *testing.T) {
			replacement := "(void)adamic_library_cwd_value;"
			if name == "error code" {
				replacement = "case ENOENT: code = \"ENOTDIR\";"
			}
			code := native.C(program)
			if name == "error code" {
				code = strings.ReplaceAll(string(data)+"\n"+code, "adamic_node_", "mutant_node_")
			}
			if strings.Count(code, anchor) < 1 {
				t.Fatal("mutant anchor changed")
			}
			code = strings.Replace(code, anchor, replacement, 1)
			mutant := filepath.Join(t.TempDir(), "mutant")
			if err := native.Build(code, mutant, native.Options{Sanitize: true}); err != nil {
				t.Fatal(err)
			}
			bad := execute(t, mutant)
			if bad.exitCode != 0 || len(bad.stderr) != 0 || disagreement(truth, bad) != "stdout differs" {
				t.Fatal("mutant not caught only by Node stdout comparison")
			}
			t.Log("clean runtime mutant caught by Node stdout comparison")
		})
	}
}
