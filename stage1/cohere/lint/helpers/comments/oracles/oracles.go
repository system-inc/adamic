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

func Comments(test *testing.T) string {
	test.Helper()
	root, _ := filepath.Abs("../../../../../cohere")
	main, _ := filepath.Abs("testdata/oracle.go")
	exports, _ := filepath.Abs("testdata/exports.go")
	virtual := filepath.Join(root, "adamic_comments_oracle.go")
	overlay, _ := json.Marshal(map[string]any{"Replace": map[string]string{
		virtual: main,
		filepath.Join(root, "internal/lint/ecmascript/comments/adamic_exports.go"): exports,
	}})
	path := filepath.Join(test.TempDir(), "overlay.json")
	if writeError := os.WriteFile(path, overlay, 0644); writeError != nil {
		test.Fatal(writeError)
	}
	// Cohere with the oracle laid over it is a product built ahead, keyed through the overlay (#5qykzj5).
	return stock(test, "go-oracle-owner-v1", virtual, []string{"-overlay=" + path})
}

// stock is the identical stock request in owners and finders; only the owner builds.
func stock(test testing.TB, output, pkg string, arguments []string, environment ...string) string {
	test.Helper()
	if producer {
		return buildcache.GoBuildIn(test, "cohere", output, pkg, arguments, environment...)
	}
	inputs, goBuildInInputsError := buildcache.GoBuildInInputs("cohere", output, pkg, arguments, nil, environment...)
	if goBuildInInputsError != nil {
		test.Fatal(goBuildInInputsError)
	}
	directory := buildcache.BuiltElsewhere(
		test,
		"Go oracle",
		inputs,
		"stage1/cohere/lint/helpers/comments/oracles owners",
		func(string) error { return fmt.Errorf("settle the oracle owner first") },
	)
	return filepath.Join(directory, output)
}
