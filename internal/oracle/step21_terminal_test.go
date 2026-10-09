package oracle

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// Copies keep these admission tests independent of the fixture readers in other
// Go packages. The runner and runtime are copied byte for byte, with no test seam.
func step21TerminalOracleCopy(t *testing.T, sources ...string) (string, string) {
	t.Helper()
	root := t.TempDir()
	for _, path := range append(sources, "oracle/node.mjs", "oracle/adamic.mjs") {
		data, err := os.ReadFile(filepath.Join(repository, path))
		if err != nil {
			t.Fatal(err)
		}
		destination := filepath.Join(root, path)
		if err := os.MkdirAll(filepath.Dir(destination), 0755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(destination, data, 0644); err != nil {
			t.Fatal(err)
		}
	}
	return root, filepath.Join(root, "oracle/node.mjs")
}

func TestStep21TerminalOracleSourceHash(t *testing.T) {
	t.Parallel()
	name := "internal/oracle/testdata/from_code_point_fails.a"
	root, runner := step21TerminalOracleCopy(t, name)
	path := filepath.Join(root, name)
	run := func() run {
		return execute(t, "node", "--disable-warning=ExperimentalWarning", runner, path)
	}
	if got := run(); got.exitCode != 70 {
		t.Fatalf("pinned terminal convention: exit %d stderr %q", got.exitCode, got.stderr)
	}
	if err := os.WriteFile(path, []byte("throw new Error('changed program');\n"), 0644); err != nil {
		t.Fatal(err)
	}
	if got := run(); got.exitCode != 1 {
		t.Fatalf("changed source borrowed terminal convention: exit %d stderr %q", got.exitCode, got.stderr)
	}
}

func TestStep21TerminalOracleDependencyHash(t *testing.T) {
	t.Parallel()
	name := "stage3/fixtures/cycles/06_import_order_mutant/"
	root, runner := step21TerminalOracleCopy(t, name+"main.a", name+"core.a", name+"utilities.a", name+"_namespaces/ts.a")
	dependency := filepath.Join(root, name+"core.a")
	original, err := os.ReadFile(dependency)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dependency, append([]byte("throw new Error('changed dependency');\n"), original...), 0644); err != nil {
		t.Fatal(err)
	}
	got := execute(t, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(root, name+"main.a"))
	if got.exitCode != 1 {
		t.Fatalf("changed dependency borrowed terminal convention: exit %d stderr %q", got.exitCode, got.stderr)
	}
}

func TestStep21ReadinessIsCatchable(t *testing.T) {
	t.Parallel()
	// A language TDZ is catchable; placeholder representation failures stay terminal.
	path := filepath.Join(t.TempDir(), "readiness.a")
	source := "function read(): number { return value; } try { console.log(String(read())); } catch { console.log('caught'); } finally { console.log('finally'); } const value = 1;"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	if truth.exitCode != 0 || string(truth.stdout) != "caught\nfinally\n" || len(truth.stderr) != 0 {
		t.Fatalf("Node TDZ control: %+v", truth)
	}
	program := &ir.Program{
		Source:  path,
		Locals:  []ir.Local{{Name: "value", Type: ir.Number, Global: true}},
		Strings: []string{"caught", "finally"},
		Main: []ir.Statement{ir.Try{
			HasCatch: true, CatchLocal: -1, HasFinally: true,
			Body:    []ir.Statement{ir.WriteLine{Value: ir.NumberToString{Value: ir.Read{Local: 0, Of: ir.Number, Checked: true}}}},
			Catch:   []ir.Statement{ir.WriteLine{Value: ir.StringConstant{Index: 0}}},
			Finally: []ir.Statement{ir.WriteLine{Value: ir.StringConstant{Index: 1}}},
		}},
	}
	native, binary := natively(t, program)
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program), "release": released(t, program)} {
		if difference := disagreement(truth, got); difference != "" {
			t.Errorf("%s language TDZ: %s: %+v", name, difference, got)
		}
	}
	if report := leaks(t, program, binary); report != "" {
		t.Fatal(report)
	}
}
