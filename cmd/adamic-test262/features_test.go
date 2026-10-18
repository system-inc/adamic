package main

import (
	"testing"
)

// Main builds a runtime for the emitted features; retain this rest-argument control.
func TestRunnerRestArgumentFeatureSetAgrees(t *testing.T) {
	t.Parallel()
	engine, err := prepare(t, "testdata/mini", t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	result := engine.attempt(classified{Path: "feature-set-link.a", Directory: "features", Program: `function count(...items:number[]):number {return arguments.length;}
const call=count;
console.log(call(1,2).toString());
`})
	if result.Kind != outcomePass {
		t.Fatalf("Node/native comparison: %#v", result)
	}
}
