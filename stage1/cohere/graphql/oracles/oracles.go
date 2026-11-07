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
	packageDirectory := filepath.Join(cohere, "internal", "format", "graphql")
	replace := map[string]string{filepath.Join(packageDirectory, "adamic_port_side_test.go"): side}
	overlay, encodeError := json.Marshal(map[string]any{"Replace": replace})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	overlayPath := filepath.Join(directory, "overlay.json")
	if writeError := os.WriteFile(overlayPath, overlay, 0o644); writeError != nil {
		test.Fatal(writeError)
	}
	const output = "graphql.test-oracle-owner-v1"
	arguments := []string{"-overlay=" + overlayPath}
	var binary string
	if oracleProducer {
		binary = buildcache.GoTest(test, "cohere", output, "./internal/format/graphql", arguments)
	} else {
		inputs, goTestInputsError := buildcache.GoTestInputs(
			"cohere",
			output,
			"./internal/format/graphql",
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
			"stage1/cohere/graphql/oracles TestProduct_GraphQLGoOracle",
			func(string) error { return fmt.Errorf("settle the Go oracle owner first") },
		)
		binary = filepath.Join(product, output)
	}
	return binary, packageDirectory
}
