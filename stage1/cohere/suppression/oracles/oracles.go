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

func Suppression(test testing.TB) (string, string) {
	test.Helper()
	directory := test.TempDir()
	cohere, absolutePathError := filepath.Abs(filepath.Join("../../..", "cohere"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	side, absolutePathError := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	packageDirectory := filepath.Join(cohere, "internal", "lint", "suppression")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if writeError := os.WriteFile(overlayPath, overlay, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	binary := stock(
		test,
		"suppression.test-owner-v1",
		"./internal/lint/suppression",
		[]string{"-overlay=" + overlayPath},
	)
	return binary, packageDirectory
}

// stock is the identical stock request in owners and finders; only the owner builds.
func stock(test testing.TB, output, pkg string, arguments []string, environment ...string) string {
	test.Helper()
	if producer {
		return buildcache.GoTest(test, "cohere", output, pkg, arguments, environment...)
	}
	inputs, goTestInputsError := buildcache.GoTestInputs("cohere", output, pkg, arguments, nil, environment...)
	if goTestInputsError != nil {
		test.Fatal(goTestInputsError)
	}
	directory := buildcache.BuiltElsewhere(
		test,
		"Go oracle",
		inputs,
		"stage1/cohere/suppression/oracles owners",
		func(string) error { return fmt.Errorf("settle the oracle owner first") },
	)
	return filepath.Join(directory, output)
}
