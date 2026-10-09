package fuzz

import (
	"path/filepath"
	"strings"
	"testing"
)

// This tool still shares a default runtime. A feature mismatch must stop before execution.
func TestFuzzerFeatureMismatchStopsAtLink(t *testing.T) {
	t.Parallel()
	directory := t.TempDir()
	checkout, err := Prepare("../..", filepath.Join(directory, "checkout"))
	if err != nil {
		t.Fatal(err)
	}
	outcome := checkout.Try(`function count(...items:number[]):number {return arguments.length;}
const call=count;
console.log(call(1,2).toString());
`, filepath.Join(directory, "program"))
	undefined := strings.Contains(outcome.Detail, "undefined reference") || strings.Contains(outcome.Detail, "Undefined symbols")
	if outcome.Verdict != Finding || outcome.Key != "clang refused the C" || !undefined || !strings.Contains(outcome.Detail, "adamic_runtime_features_closure_convention") {
		t.Fatalf("want feature-set link failure, got %#v", outcome)
	}
	t.Log(outcome.Detail)
}
