package main

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func declareFixtureCensus(t *testing.T, root string) {
	t.Helper()
	marker := filepath.Join(root, "internal/skipcensus/census.go")
	if err := os.MkdirAll(filepath.Dir(marker), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(marker, []byte("//go:build census_marker\n\npackage skipcensus\n"), 0600); err != nil {
		t.Fatal(err)
	}

}

func TestSkipCensusChecksTreeAndLog(t *testing.T) {
	t.Parallel()
	root := t.TempDir()
	if err := os.MkdirAll(filepath.Join(root, "probe"), 0755); err != nil {
		t.Fatal(err)
	}
	file := filepath.Join(root, "probe/probe_test.go")
	source := `package probe
import "testing"
func TestRequired(t *testing.T){
 // census: required-input ADAMIC_TYPESCRIPT_SOURCE fixture provision
 t.Skip("source input missing")
}
func TestNA(t *testing.T){
 // census: not-applicable platform witness
 t.Skip("platform")
}
func TestMeasurement(t *testing.T){
 // census: measurement
 t.Skip("measurement")
}
func TestLane(t *testing.T){
 // census: opt-in-lane separate lane
 t.Skip("separate lane")
}
`
	if err := os.WriteFile(file, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	declareFixtureCensus(t, root)
	log := filepath.Join(t.TempDir(), "test.jsonl")
	allowed := `{"Action":"skip","Package":"github.com/system-inc/adamic/probe","Test":"TestNA"}
{"Action":"skip","Package":"github.com/system-inc/adamic/probe","Test":"TestMeasurement"}
{"Action":"skip","Package":"github.com/system-inc/adamic/probe","Test":"TestLane"}
`
	if err := os.WriteFile(log, []byte(allowed), 0600); err != nil {
		t.Fatal(err)
	}
	summary, err := checkSkipCensus(root, log)
	if err != nil || summary.Status != "checked" || len(summary.NotApplicable) != 1 || len(summary.Measurement) != 1 || len(summary.OptInLane) != 1 {
		t.Fatal(summary, err)
	}
	if err := os.WriteFile(log, []byte(allowed+`{"Action":"skip","Package":"github.com/system-inc/adamic/probe","Test":"TestRequired"}`), 0600); err != nil {
		t.Fatal(err)
	}
	summary, err = checkSkipCensus(root, log)
	if err == nil || len(summary.RequiredInput) != 1 || !strings.Contains(err.Error(), "TestRequired") || !strings.Contains(err.Error(), "ADAMIC_TYPESCRIPT_SOURCE") {
		t.Fatal("required input skipped green", summary, err)
	}
	if err := os.WriteFile(file, []byte(source+"\nfunc TestNew(t *testing.T){t.Skip(\"new\")}\n"), 0600); err != nil {
		t.Fatal(err)
	}
	if _, err := checkSkipCensus(root, log); err == nil || !strings.Contains(err.Error(), "missing census annotation") {
		t.Fatal("unannotated skip accepted", err)
	}
}
