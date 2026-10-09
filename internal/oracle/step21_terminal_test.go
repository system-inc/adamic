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
	name := "internal/oracle/testdata/dead_zone.a"
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
	name := "internal/oracle/testdata/import_cycles/read/"
	root, runner := step21TerminalOracleCopy(t, name+"a.a", name+"b.a")
	dependency := filepath.Join(root, name+"b.a")
	original, err := os.ReadFile(dependency)
	if err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(dependency, append([]byte("throw new Error('changed dependency');\n"), original...), 0644); err != nil {
		t.Fatal(err)
	}
	got := execute(t, "node", "--disable-warning=ExperimentalWarning", runner, filepath.Join(root, name+"a.a"))
	if got.exitCode != 1 {
		t.Fatalf("changed dependency borrowed terminal convention: exit %d stderr %q", got.exitCode, got.stderr)
	}
}

func TestStep21ReadinessIsTerminal(t *testing.T) {
	t.Parallel()
	// Node's language TDZ is catchable; Adamic's inserted readiness check is terminal.
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
	native, _ := natively(t, program)
	want := run{stderr: []byte("adamic: panic: ReferenceError: Cannot access 'value' before initialization\n"), exitCode: 70}
	for name, got := range map[string]run{"native": native, "JavaScript": onJavaScriptBackend(t, program), "release": released(t, program)} {
		if difference := disagreement(want, got); difference != "" {
			t.Errorf("%s readiness guard ran catch/finally: %s: %+v", name, difference, got)
		}
	}
}
