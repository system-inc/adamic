// Package oracles owns formatfiles' compiler-free Go oracle product.
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

// GoOracle returns the overlaid Go test binary and the directory it runs in.
func GoOracle(test *testing.T) (string, string) {
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
	packageDirectory := filepath.Join(cohere, "internal", "format", "formatfiles")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if writeError := os.WriteFile(overlayPath, overlay, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	binary := stockOracle(test, []string{"-overlay=" + overlayPath})
	return binary, packageDirectory
}

// stockOracle uses the identical request for owner builds and parent lookups.
func stockOracle(test *testing.T, arguments []string) string {
	test.Helper()
	const output = "formatfiles.test-owner-v1"
	const module = "cohere"
	const pkg = "./internal/format/formatfiles"
	if producer {
		return buildcache.GoTest(test, module, output, pkg, arguments)
	}
	inputs, goTestInputsError := buildcache.GoTestInputs(module, output, pkg, arguments, nil)
	if goTestInputsError != nil {
		test.Fatal(goTestInputsError)
	}
	directory := buildcache.BuiltElsewhere(
		test,
		"formatfiles Go oracle",
		inputs,
		"stage1/cohere/formatfiles/oracles TestProduct_FormatfilesGoOracle",
		func(string) error { return fmt.Errorf("settle the formatfiles Go oracle owner first") },
	)
	return filepath.Join(directory, output)
}
