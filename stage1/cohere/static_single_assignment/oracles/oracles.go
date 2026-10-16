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

func StaticSingleAssignment(t testing.TB) (string, string) {
	t.Helper()
	directory := t.TempDir()
	module, err := filepath.Abs(filepath.Join("../../..", "cohere", "static_single_assignment"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	replace := map[string]string{filepath.Join(module, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	// Built under GOWORK=off as before.
	binary := stock(t, "static_single_assignment.test-owner-v1", ".", []string{"-overlay=" + overlayPath}, "GOWORK=off")
	return binary, module
}

// stock is the identical stock request in owners and finders; only the owner builds.
func stock(t testing.TB, output, pkg string, arguments []string, environment ...string) string {
	t.Helper()
	if producer {
		return buildcache.GoTest(t, "cohere/static_single_assignment", output, pkg, arguments, environment...)
	}
	inputs, err := buildcache.GoTestInputs("cohere/static_single_assignment", output, pkg, arguments, nil, environment...)
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.BuiltElsewhere(t, "Go oracle", inputs, "stage1/cohere/static_single_assignment/oracles owners", func(string) error { return fmt.Errorf("settle the oracle owner first") })
	return filepath.Join(directory, output)
}
