// Package oracles owns the Go reference products without linking the compiler.
// Both test binaries resolve relative paths from the parent package directory.
package oracles

import (
	"encoding/json"
	"fmt"
	"github.com/system-inc/adamic/internal/buildcache"
	"os"
	"path/filepath"
	"testing"
)

const repository = "../../.."

var oracleProducer bool

// EnableProduction enables product recipes in the owner test binary's TestMain.
func EnableProduction() { oracleProducer = true }
func Parser(t testing.TB) (string, string) {
	t.Helper()
	directory := t.TempDir()
	cohere, err := filepath.Abs(filepath.Join(repository, "cohere"))
	if err != nil {
		t.Fatal(err)
	}
	side, err := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if err != nil {
		t.Fatal(err)
	}
	packageDirectory := filepath.Join(cohere, "internal", "format", "css", "mediaquery")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, err := json.Marshal(map[string]any{"Replace": replace})
	if err != nil {
		t.Fatal(err)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(overlayPath, overlay, 0o644); err != nil {
		t.Fatal(err)
	}
	const output = "mediaquery.test-oracle-owner-v1"
	arguments := []string{"-overlay=" + overlayPath}
	var binary string
	if oracleProducer {
		binary = buildcache.GoTest(t, "cohere", output, "./internal/format/css/mediaquery", arguments)
	} else {
		inputs, err := buildcache.GoTestInputs("cohere", output, "./internal/format/css/mediaquery", arguments, nil)
		if err != nil {
			t.Fatal(err)
		}
		product := buildcache.BuiltElsewhere(t, output, inputs, "stage1/cohere/mediaquery/oracles TestProduct_MediaQueryGoOracle", func(string) error { return fmt.Errorf("settle the Go oracle owner first") })
		binary = filepath.Join(product, output)
	}
	return binary, packageDirectory
}
