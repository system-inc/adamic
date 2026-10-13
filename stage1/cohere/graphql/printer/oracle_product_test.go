package printer

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
)

// The Go oracle is one read-only product (buildcache.GoTest): cohere's graphql package with this package's and the
// generator's harnesses laid over it, keyed by every file go list says the test binary compiles, each module's go.mod,
// go.sum and go.work, the overlay's map and sources, and the build environment as go resolves it. Every corpus (as-Go,
// mutant, preflight, whitespace) calls it through a process-local Once, and a product built by an earlier process is
// fetched, not rebuilt. The as-Go and mutant tests once kept products of their own that held this binary's very bytes.
var printerOracleOnce sync.Once
var printerOraclePath string

func printerOracle(t *testing.T) string {
	t.Helper()
	printerOracleOnce.Do(func() {
		root, err := filepath.Abs(repository)
		if err != nil {
			t.Fatal(err)
		}
		side, _ := filepath.Abs("testdata/cohere_side_test.go")
		generator, _ := filepath.Abs("../testdata/cohere_side_test.go")
		cohere := filepath.Join(root, "cohere")
		overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
			cohere + "/internal/format/graphql/adamic_printer_test.go":   side,
			cohere + "/internal/format/graphql/adamic_generator_test.go": generator,
		}})
		if err != nil {
			t.Fatal(err)
		}
		// The overlay names this checkout's paths: it keys and builds the product and is never part of it (#tqrqx60).
		path := filepath.Join(t.TempDir(), "overlay.json")
		if err := os.WriteFile(path, overlay, 0644); err != nil {
			t.Fatal(err)
		}
		// The binary keeps the name oracle: each corpus keys it by that name and its bytes (printerBuild).
		printerOraclePath = buildcache.GoTest(t, "cohere", "oracle", "./internal/format/graphql", []string{"-overlay=" + path})
	})
	// A build that failed inside the Once leaves the path empty for every later caller in this process.
	// Say so, rather than running an empty command.
	if printerOraclePath == "" {
		t.Fatal("Go GraphQL printer oracle unavailable: its build failed earlier in this process (see that test's log)")
	}
	return printerOraclePath
}

func TestProduct_GraphQLPrinterGoOracle(t *testing.T) { t.Parallel(); printerOracle(t) }
