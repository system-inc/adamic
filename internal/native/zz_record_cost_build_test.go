package native

import (
	"os"
	"path/filepath"
	"testing"
)

// TestRecordCostBuild builds the record cost probe's harness (zz_record_cost_probe_test.go), release and
// counted, into RECORD_COST_OUT for measure/record-cost/run.sh to time on a quiet box (#chncwr5's
// baseline). Never lands.
func TestRecordCostBuild(t *testing.T) {
	out := os.Getenv("RECORD_COST_OUT")
	if out == "" {
		t.Skip("RECORD_COST_OUT is unset")
	}
	if err := Build(recordCostHarness, filepath.Join(out, "release"), Options{}); err != nil {
		t.Fatal(err)
	}
	if err := Build(recordCostHarness, filepath.Join(out, "counted"), Options{Sanitize: true, Count: true}); err != nil {
		t.Fatal(err)
	}
	t.Logf("release flags: %v", Flags(Options{}))
}
