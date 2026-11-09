package oracle

import (
	"context"
	"encoding/gob"
	"os"
	"os/exec"
	"path/filepath"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
)

func hostBindingProcessOracleLower(t testing.TB) string {
	t.Helper()
	return buildcache.GoTest(t, "", "host-binding-process-oracle-lower.test", "./internal/lower", nil)
}

func TestProduct_HostBindingProcessOracleLower(t *testing.T) {
	t.Parallel()
	hostBindingProcessOracleLower(t)
}

// Typed declarations are synthetic host rows, admitted only by lower's isolated test builder. Keep
// their native observations uncached: the private test table is not a production compiler input.
func hostBindingTypedProcessAgrees(t *testing.T, of string) {
	t.Helper()
	output := filepath.Join(t.TempDir(), "program.gob")
	lowerBinary := hostBindingProcessOracleLower(t)
	ctx, cancel := context.WithTimeout(t.Context(), 900*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, lowerBinary, "-test.run=^TestHostBindingProcessOracleIR$", "-test.count=1", "-test.timeout=900s")
	command.Dir = filepath.Join(repository, "internal", "lower")
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
