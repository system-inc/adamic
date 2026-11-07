// Package oracles owns Go reference products without linking the compiler.
// The owner TestMain keeps relative paths rooted in the parent package.
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

// EnableProduction permits recipes only in the owner's test binary.
func EnableProduction() { oracleProducer = true }
func Binary(test *testing.T) string {
	test.Helper()
	cohere, absolutePathError := filepath.Abs(filepath.Join(repository, "cohere"))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	side, absolutePathError := filepath.Abs("testdata/cohere_side_test.go")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	overlay, encodeError := json.Marshal(map[string]any{
		"Replace": map[string]string{filepath.Join(cohere, "internal", "gitignore", "adamic_port_side_test.go"): side},
	})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	// The overlay names this checkout's paths: GoTest keys its map and the side file's content, never its own path.
	path := filepath.Join(test.TempDir(), "overlay.json")
	if writeError := os.WriteFile(path, overlay, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	const output = "oracle-oracle-owner-v1"
	arguments := []string{"-overlay=" + path}
	if oracleProducer {
		return buildcache.GoTest(test, "cohere", output, "./internal/gitignore", arguments)
	}
	inputs, goTestInputsError := buildcache.GoTestInputs("cohere", output, "./internal/gitignore", arguments, nil)
	if goTestInputsError != nil {
		test.Fatal(goTestInputsError)
	}
	product := buildcache.BuiltElsewhere(
		test,
		output,
		inputs,
		"stage1/cohere/gitignore/oracles TestProduct_GitignoreOracle",
		func(string) error { return fmt.Errorf("settle the Go oracle owner first") },
	)
	return filepath.Join(product, output)
}
