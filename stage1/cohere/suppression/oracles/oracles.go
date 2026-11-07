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
	cohere, err := filepath.Abs(filepath.Join("../../..", "cohere"))
	if err != nil {
		test.Fatal(err)
	}
	side, err := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if err != nil {
		test.Fatal(err)
	}
	packageDirectory := filepath.Join(cohere, "internal", "lint", "suppression")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		test.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		test.Fatal(err)
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
	inputs, err := buildcache.GoTestInputs("cohere", output, pkg, arguments, nil, environment...)
	if err != nil {
		test.Fatal(err)
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
