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
func Printer(test *testing.T) string {
	test.Helper()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	bridge, absolutePathError := filepath.Abs("testdata/bridge.go")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	driver, absolutePathError := filepath.Abs("testdata/go_driver.go")
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_stage_one/main.go")
	data, encodeError := json.Marshal(map[string]any{"Replace": map[string]string{
		filepath.Join(cohere, "internal/format/markdown/adamic_stage_one.go"): bridge,
		mainPath: driver,
	}})
	if encodeError != nil {
		test.Fatal(encodeError)
	}
	overlayPath := filepath.Join(test.TempDir(), "overlay.json")
	if writeError := os.WriteFile(overlayPath, data, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	const output = "go-printer-oracle-owner-v1"
	arguments := []string{"-overlay=" + overlayPath}
	if oracleProducer {
		return buildcache.GoBuildIn(test, "cohere", output, mainPath, arguments)
	}
	inputs, goBuildInInputsError := buildcache.GoBuildInInputs("cohere", output, mainPath, arguments, nil)
	if goBuildInInputsError != nil {
		test.Fatal(goBuildInInputsError)
	}
	product := buildcache.BuiltElsewhere(
		test,
		output,
		inputs,
		"stage1/cohere/markdowninline/oracles TestProduct_MarkdownInlineGoOverlay",
		func(string) error { return fmt.Errorf("settle the Go overlay owner first") },
	)
	return filepath.Join(product, output)
}
