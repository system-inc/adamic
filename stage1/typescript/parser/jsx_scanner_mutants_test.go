package parser

import (
	"path/filepath"
	"strings"
	"testing"
)

// Not parallel: each compiled scanner mutant is checked after its positive control.
func TestJsxScannerMutants(test *testing.T) {
	manifest, _ := jsxManifest(test)
	oracle := goOracle(test)
	want := execute(test, "", oracle, "--manifest", manifest, "--whole").output
	directory, _ := filepath.Abs(".")
	for _, got := range []execution{
		wholeNode(test, directory, manifest, false),
		execute(test, "", buildPort(test, directory, true), "--manifest", manifest, "--whole"),
	} {
		if diff := difference(got.output, want); diff != "" {
			test.Fatal(diff)
		}
	}
	for _, change := range jsxScannerChanges() {
		test.Run(change.name, func(test *testing.T) {
			mutant := jsxScannerMutant(test, change)
			node := wholeNode(test, mutant, manifest, false)
			native := execute(test, "", buildPort(test, mutant, true), "--manifest", manifest, "--whole")
			for _, side := range []struct {
				name string
				got  execution
			}{{"Node", node}, {"native", native}} {
				diff := difference(side.got.output, want)
				if diff == "" {
					test.Fatalf("%s scanner mutant survived", side.name)
				}
				test.Logf("%s compiled mutant caught: %s", side.name, strings.Split(diff, "\n")[0])
			}
		})
	}
}
