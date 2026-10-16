// Package oracles owns the compiler-free Go reference product.
package oracles

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/toolchain"
)

var producer bool

// ScannerOracleProduct returns the shared reference executable's product directory. Only this
// package's product test may build it; consumers must find the owner's settlement.
func ScannerOracleProduct(t *testing.T) string {
	t.Helper()
	inputs := oracleInputs()
	build := func(directory string) error { return buildOracle(t, directory) }
	if producer {
		return buildcache.Product(t, inputs, build)
	}
	return buildcache.BuiltElsewhere(t, "Scanner Go oracle", inputs, "stage1/typescript/scanner/oracles TestProduct_ScannerOracle", func(string) error {
		return fmt.Errorf("settle stage1/typescript/scanner/oracles TestProduct_ScannerOracle before reading the oracle")
	})
}

func oracleInputs() buildcache.Inputs {
	inputs := buildcache.Inputs{
		Name:      "scanner-go-oracle",
		Files:     []string{"cohere/TypeScript/tsc", "cohere/TypeScript/go.work", "cohere/TypeScript/tools/go.mod", "cohere/TypeScript/tools/go.sum", "go.mod", "go.work", "stage1/typescript/scanner/testdata/oracle.go"},
		Flags:     []string{"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "), "GOFLAGS=" + os.Getenv("GOFLAGS"), "GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT"), "GOAMD64=" + os.Getenv("GOAMD64"), "GOARM64=" + os.Getenv("GOARM64"), "CGO_ENABLED=" + os.Getenv("CGO_ENABLED"), "GOOS=" + runtime.GOOS, "GOARCH=" + runtime.GOARCH},
		Toolchain: []string{buildcache.Tool("go", "version")},
	}
	inputs.Name += " oracle-owner-v1"
	return inputs
}

func buildOracle(t *testing.T, directory string) error {
	root, err := filepath.Abs("../../../cohere/TypeScript/tsc")
	if err != nil {
		return err
	}
	source, err := filepath.Abs("testdata/oracle.go")
	if err != nil {
		return err
	}
	virtual := filepath.Join(root, "adamic_scanner_oracle.go")
	overlay, err := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
	if err != nil {
		return err
	}
	path := filepath.Join(directory, "overlay.json")
	if err := os.WriteFile(path, overlay, 0644); err != nil {
		return err
	}
	// Checkout paths are build inputs, never artifacts.
	defer os.Remove(path)
	command := toolchain.CommandContextIn(t.Context(), root, "go", append(append([]string{"build"}, buildcache.ReproducibleGoFlags()...), "-overlay="+path, "-o", filepath.Join(directory, "oracle"), virtual)...)
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	if err := childguard.Run(command, childguard.Options{Unbounded: true}); err != nil {
		return fmt.Errorf("Go oracle: %w\n%s", err, &output)
	}
	return nil
}
