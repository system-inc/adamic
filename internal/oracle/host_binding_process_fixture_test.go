package oracle

import (
	"encoding/gob"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

// Typed declarations are synthetic host rows, admitted only by lower's isolated test builder. Keep
// their native observations uncached: the private test table is not a production compiler input.
func hostBindingTypedProcessAgrees(t *testing.T, of string) {
	t.Helper()
	output := filepath.Join(t.TempDir(), "program.gob")
	command := exec.CommandContext(t.Context(), "timeout", "900", "go", "test", "./internal/lower", "-count=1", "-run", "^TestHostBindingProcessOracleIR$")
	command.Dir = repository
	command.Env = append(os.Environ(), "ADAMIC_PROCESS_ORACLE_TYPE="+of, "ADAMIC_PROCESS_ORACLE_IR="+output)
	if log, err := command.CombinedOutput(); err != nil {
		t.Fatalf("typed-row builder: %v\n%s", err, log)
	}
	file, err := os.Open(output)
	if err != nil {
		t.Fatal(err)
	}
	defer file.Close()
	ir.RegisterGob()
	var program ir.Program
	if err := gob.NewDecoder(file).Decode(&program); err != nil {
		t.Fatal(err)
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/host_bindings/process_browser_"+of+"/main.ts"))
	if err != nil {
		t.Fatal(err)
	}
	want := onNode(t, path)
	sanitized, binary := nativelyUncached(t, &program)
	for _, candidate := range []struct {
		name string
		got  run
	}{{"native", releasedUncached(t, &program)}, {"sanitized", sanitized}, {"javascript", onJavaScriptBackend(t, &program)}} {
		if difference := disagreement(want, candidate.got); difference != "" {
			t.Errorf("%s: %s, Node %+v, got %+v", candidate.name, difference, want, candidate.got)
		}
	}
	if report := leaksUncached(t, &program, binary); report != "" {
		t.Errorf("leaked: %s", report)
	}
	if os.Getenv("ADAMIC_ORACLE_WASI") == "1" {
		if got := onWASI(t, native.C(&program)); disagreement(want, got) != "" {
			t.Fatalf("Node %+v WASI %+v", want, got)
		}
	}
}
