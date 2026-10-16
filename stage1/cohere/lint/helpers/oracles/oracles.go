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

func Helpers(t *testing.T) string {
	t.Helper()
	root, _ := filepath.Abs("../../../../cohere")
	side, _ := filepath.Abs("testdata/oracle.go")
	catalog, _ := filepath.Abs("testdata/catalog.go")
	descriptors, _ := filepath.Abs("testdata/descriptors.go")
	virtual := filepath.Join(root, "adamic_helper_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{virtual: side, filepath.Join(root, "policy/adamic_helpers.go"): catalog, filepath.Join(root, "adamic_helper_descriptors.go"): descriptors}})
	path := filepath.Join(t.TempDir(), "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		t.Fatal(err)
	}
	// Cohere with the oracle laid over it is a product built ahead, keyed through the overlay (#5qykzj5).
	return stock(t, "go-oracle-owner-v1", virtual+" "+filepath.Join(root, "adamic_helper_descriptors.go"), []string{"-overlay=" + path})
}

// stock is the identical stock request in owners and finders; only the owner builds.
func stock(t testing.TB, output, pkg string, arguments []string, environment ...string) string {
	t.Helper()
	if producer {
		return buildcache.GoBuildIn(t, "cohere", output, pkg, arguments, environment...)
	}
	inputs, err := buildcache.GoBuildInInputs("cohere", output, pkg, arguments, nil, environment...)
	if err != nil {
		t.Fatal(err)
	}
	directory := buildcache.BuiltElsewhere(t, "Go oracle", inputs, "stage1/cohere/lint/helpers/oracles owners", func(string) error { return fmt.Errorf("settle the oracle owner first") })
	return filepath.Join(directory, output)
}
