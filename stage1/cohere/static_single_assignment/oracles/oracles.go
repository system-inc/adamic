// Package oracles owns the compiler-free Go oracle products for this port.
package oracles

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"testing"
)

var producer bool

func StaticSingleAssignment(test testing.TB) (string, string) {
	test.Helper()
	directory := test.TempDir()
	module, absolutePathError := filepath.Abs(filepath.Join("../../..", "cohere", "static_single_assignment"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	side, absolutePathError := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	replace := map[string]string{filepath.Join(module, "adamic_port_side_test.go"): side}
	overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if writeError := os.WriteFile(overlayPath, overlay, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	// Built under GOWORK=off as before.
	binary := stock(
		test,
		"static_single_assignment.test-owner-v1",
		".",
		[]string{"-overlay=" + overlayPath},
		"GOWORK=off",
	)
	return binary, module
}

// stock is the identical stock request in owners and finders; only the owner builds.
func stock(test testing.TB, output, pkg string, arguments []string, environment ...string) string {
	test.Helper()
	if producer {
		return buildcache.GoTest(test, "cohere/static_single_assignment", output, pkg, arguments, environment...)
	}
	inputs, goTestInputsError := buildcache.GoTestInputs(
		"cohere/static_single_assignment",
		output,
		pkg,
		arguments,
		nil,
		environment...,
	)
	if goTestInputsError != nil {
		test.Fatal(goTestInputsError)
	}
	directory := buildcache.BuiltElsewhere(
		test,
		"Go oracle",
		inputs,
		"stage1/cohere/static_single_assignment/oracles owners",
		func(string) error { return fmt.Errorf("settle the oracle owner first") },
	)
	return filepath.Join(directory, output)
}
