package main

import (
	"runtime"
	"strings"
	"testing"
)

// The runner's shared default runtime must not execute a program with a different layout.
func TestRunnerFeatureMismatchStopsAtLink(t *testing.T) {
	t.Parallel()
	engine, err := prepare("../..", "testdata/mini", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	result := engine.attempt(classified{Path: "feature-set-link.a", Directory: "features", Program: `function count(...items:number[]):number {return arguments.length;}
const call=count;
console.log(call(1,2).toString());
`})
	// The runner records only the first stderr line. Darwin puts symbol names below it.
	undefined := strings.Contains(result.Reason, "undefined reference") && strings.Contains(result.Reason, "adamic_runtime_features_closure_convention")
	if runtime.GOOS == "darwin" {
		undefined = strings.Contains(result.Reason, "Undefined symbols")
	}
	if result.Kind != outcomeCrashed || !strings.Contains(result.Reason, "clang:") || !undefined {
		t.Fatalf("want feature-set link failure, got %#v", result)
	}
	t.Log(result.Reason)
}
