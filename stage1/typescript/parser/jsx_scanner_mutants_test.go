package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: each compiled scanner mutant is checked after its positive control.
func TestJsxScannerMutants(t *testing.T) {
	manifest, _ := jsxManifest(t)
	oracle := goOracle(t)
	want := execute(t, "", oracle, "--manifest", manifest, "--whole").output
	directory, _ := filepath.Abs(".")
	for _, got := range []execution{wholeNode(t, directory, manifest, false), execute(t, "", buildPort(t, directory, true), "--manifest", manifest, "--whole")} {
		if diff := difference(got.output, want); diff != "" {
			t.Fatal(diff)
		}
	}
	for _, change := range jsxScannerChanges() {
		t.Run(change.name, func(t *testing.T) {
			mutant := jsxScannerMutant(t, change)
			node := wholeNode(t, mutant, manifest, false)
			native := execute(t, "", buildPort(t, mutant, true), "--manifest", manifest, "--whole")
			for _, side := range []struct {
				name string
				got  execution
			}{{"Node", node}, {"native", native}} {
				diff := difference(side.got.output, want)
				if diff == "" {
					t.Fatalf("%s scanner mutant survived", side.name)
				}
				t.Logf("%s compiled mutant caught: %s", side.name, strings.Split(diff, "\n")[0])
			}
		})
	}
}
