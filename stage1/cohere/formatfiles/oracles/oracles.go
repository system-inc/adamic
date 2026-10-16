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
func GoOracle(t *testing.T) (string, string) {
	t.Helper()
	directory := t.TempDir()
	cohere, err := filepath.Abs(filepath.Join("../../..", "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	packageDirectory := filepath.Join(cohere, "internal", "format", "formatfiles")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	binary := stockOracle(t, []string{"-overlay=" + overlayPath})
	return binary, packageDirectory
}

// stockOracle uses the identical request for owner builds and parent lookups.
func stockOracle(t *testing.T, arguments []string) string {
	t.Helper()
	const output = "formatfiles.test-owner-v1"
	const module = "cohere"
	const pkg = "./internal/format/formatfiles"
	if producer {
		return buildcache.GoTest(t, module, output, pkg, arguments)
	}
	inputs, err := buildcache.GoTestInputs(module, output, pkg, arguments, nil)
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.BuiltElsewhere(t, "formatfiles Go oracle", inputs, "stage1/cohere/formatfiles/oracles TestProduct_FormatfilesGoOracle", func(string) error { return fmt.Errorf("settle the formatfiles Go oracle owner first") })
	return filepath.Join(directory, output)
}
