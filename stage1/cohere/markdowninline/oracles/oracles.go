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
	root, err := filepath.Abs(repository)
	if err != nil {
		test.Fatal(err)
	}
	bridge, err := filepath.Abs("testdata/bridge.go")
	if err != nil {
		test.Fatal(err)
	}
	driver, err := filepath.Abs("testdata/go_driver.go")
	if err != nil {
		test.Fatal(err)
	}
	cohere := filepath.Join(root, "cohere")
	mainPath := filepath.Join(cohere, "cmd/adamic_stage_one/main.go")
	data, err := json.Marshal(map[string]any{"Replace": map[string]string{
		filepath.Join(cohere, "internal/format/markdown/adamic_stage_one.go"): bridge,
		mainPath: driver,
	}})
	if err != nil {
		test.Fatal(err)
	}
	overlayPath := filepath.Join(test.TempDir(), "overlay.json")
	if err := os.WriteFile(overlayPath, data, 0644); err != nil {
		test.Fatal(err)
	}
	const output = "go-printer-oracle-owner-v1"
	arguments := []string{"-overlay=" + overlayPath}
	if oracleProducer {
		return buildcache.GoBuildIn(test, "cohere", output, mainPath, arguments)
	}
	inputs, err := buildcache.GoBuildInInputs("cohere", output, mainPath, arguments, nil)
	if err != nil {
		test.Fatal(err)
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
