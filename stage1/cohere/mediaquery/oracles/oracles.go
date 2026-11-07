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
func Parser(test testing.TB) (string, string) {
	test.Helper()
	directory := test.TempDir()
	cohere, absolutePathError := filepath.Abs(filepath.Join(repository, "cohere"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	side, absolutePathError := filepath.Abs(filepath.Join("testdata", "cohere_side_test.go"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	packageDirectory := filepath.Join(cohere, "internal", "format", "css", "mediaquery")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if writeError := os.WriteFile(overlayPath, overlay, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	const output = "mediaquery.test-oracle-owner-v1"
	arguments := []string{"-overlay=" + overlayPath}
	var binary string
	if oracleProducer {
		binary = buildcache.GoTest(test, "cohere", output, "./internal/format/css/mediaquery", arguments)
	} else {
		inputs, goTestInputsError := buildcache.GoTestInputs(
			"cohere",
			output,
			"./internal/format/css/mediaquery",
			arguments,
			nil,
		)
		if goTestInputsError != nil {
			test.Fatal(goTestInputsError)
		}
		product := buildcache.BuiltElsewhere(
			test,
			output,
			inputs,
			"stage1/cohere/mediaquery/oracles TestProduct_MediaQueryGoOracle",
			func(string) error { return fmt.Errorf("settle the Go oracle owner first") },
		)
		binary = filepath.Join(product, output)
	}
	return binary, packageDirectory
}
