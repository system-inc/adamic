package printer

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
)

// The Go oracle is one read-only product: every corpus (as-Go, mutant, preflight, whitespace) calls the same recipe
// through a process-local Once, and a product built by an earlier process is fetched, not rebuilt. The as-Go and mutant
// tests once kept products of their own, keyed by a walk of all of cohere, that held this binary's very bytes; of what
// theirs keyed, this one adds cohere's go.work, which go reads from cohere, and the build environment as go reports it
// (GOFLAGS, CGO_ENABLED, GOAMD64, GOEXPERIMENT and the rest).
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
		inputs := buildcache.Inputs{
			Name:      "graphql-printer-go-oracle",
			Files:     []string{"stage1/cohere/graphql/printer/testdata/cohere_side_test.go", "stage1/cohere/graphql/testdata/cohere_side_test.go", "cohere/internal", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "cohere/go.mod", "cohere/go.sum", "cohere/go.work", "cohere/TypeScript/tsc/go.mod", "cohere/TypeScript/tsc/go.sum"},
			Flags:     []string{"test", "-c", "-trimpath", "-ldflags=-buildid=", "./internal/format/graphql"},
			Toolchain: []string{buildcache.Tool("go", "version"), buildcache.Tool("go", "env", "-json", "GOVERSION", "GOOS", "GOARCH", "GOEXPERIMENT", "GOFLAGS", "CGO_ENABLED", "GOTOOLCHAIN", "GOAMD64", "GOARM64", "GOARM", "GO386", "GOMIPS", "GOMIPS64", "GOPPC64", "GORISCV64", "GOWASM"), runtime.GOOS, runtime.GOARCH},
		}
		directory := buildcache.Product(t, inputs, func(dir string) error {
			overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{
				cohere + "/internal/format/graphql/adamic_printer_test.go":   side,
				cohere + "/internal/format/graphql/adamic_generator_test.go": generator,
			}})
			if err != nil {
				return err
			}
			path := filepath.Join(dir, "overlay.json")
			if err := os.WriteFile(path, overlay, 0644); err != nil {
				return err
			}
			// The overlay names this checkout's paths: it builds the product and is never part of it (#tqrqx60).
			defer os.Remove(path)
			start := time.Now()
			// A product carries no deadline of its own: a cold Go build of cohere's graphql package compiles
			// typescript-go's AST and can pass 90 s on 4 CPUs, and the build phase's 10-minute ceiling bounds it (rule 10).
			command := exec.CommandContext(t.Context(), "go", "test", "-c", "-trimpath", "-ldflags=-buildid=", "-o="+filepath.Join(dir, "oracle"), "-overlay="+path, "./internal/format/graphql")
			command.Dir = cohere
			if output, err := command.CombinedOutput(); err != nil {
				return fmt.Errorf("Go GraphQL printer oracle: %w\n%s", err, output)
			}
			t.Logf("build Go GraphQL printer oracle cold wall %.3fs (overlay)", time.Since(start).Seconds())
			return nil
		})
		printerOraclePath = filepath.Join(directory, "oracle")
	})
	// A build that failed inside the Once leaves the path empty for every later caller in this process.
	// Say so, rather than running an empty command.
	if printerOraclePath == "" {
		t.Fatal("Go GraphQL printer oracle unavailable: its build failed earlier in this process (see that test's log)")
	}
	return printerOraclePath
}

func TestProduct_GraphQLPrinterGoOracle(t *testing.T) { t.Parallel(); printerOracle(t) }
