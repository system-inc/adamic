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
func ScannerOracleProduct(test *testing.T) string {
	test.Helper()
	inputs := oracleInputs()
	build := func(directory string) error { return buildOracle(test, directory) }
	if producer {
		return buildcache.Product(test, inputs, build)
	}
	return buildcache.BuiltElsewhere(
		test,
		"Scanner Go oracle",
		inputs,
		"stage1/typescript/scanner/oracles TestProduct_ScannerOracle",
		func(
			string,
		) error {
			return fmt.Errorf(
				"settle stage1/typescript/scanner/oracles TestProduct_ScannerOracle before reading the oracle",
			)
		},
	)
}

func oracleInputs() buildcache.Inputs {
	inputs := buildcache.Inputs{
		Name: "scanner-go-oracle",
		Files: []string{
			"cohere/TypeScript/tsc",
			"cohere/TypeScript/go.work",
			"cohere/TypeScript/tools/go.mod",
			"cohere/TypeScript/tools/go.sum",
			"go.mod",
			"go.work",
			"stage1/typescript/scanner/testdata/oracle.go",
		},
		Flags: []string{
			"go build " + strings.Join(buildcache.ReproducibleGoFlags(), " "),
			"GOFLAGS=" + os.Getenv("GOFLAGS"),
			"GOEXPERIMENT=" + os.Getenv("GOEXPERIMENT"),
			"GOAMD64=" + os.Getenv("GOAMD64"),
			"GOARM64=" + os.Getenv("GOARM64"),
			"CGO_ENABLED=" + os.Getenv("CGO_ENABLED"),
			"GOOS=" + runtime.GOOS,
			"GOARCH=" + runtime.GOARCH,
		},
		Toolchain: []string{buildcache.Tool("go", "version")},
	}
	inputs.Name += " oracle-owner-v1"
	return inputs
}

func buildOracle(test *testing.T, directory string) error {
	root, absolutePathError := filepath.Abs("../../../cohere/TypeScript/tsc")
	if absolutePathError != nil {
		return absolutePathError
	}
	source, absolutePathError := filepath.Abs("testdata/oracle.go")
	if absolutePathError != nil {
		return absolutePathError
	}
	virtual := filepath.Join(root, "adamic_scanner_oracle.go")
	overlay, encodeError := json.Marshal(map[string]any{"Replace": map[string]string{virtual: source}})
	if encodeError != nil {
		return encodeError
	}
	path := filepath.Join(directory, "overlay.json")
	if writeError := os.WriteFile(path, overlay, 0644); writeError != nil {
		return writeError
	}
	// Checkout paths are build inputs, never artifacts.
	defer os.Remove(path)
	command := toolchain.CommandContextIn(test.Context(), root, "go", append(
		append([]string{"build"}, buildcache.ReproducibleGoFlags()...),
		"-overlay="+path,
		"-o",
		filepath.Join(directory, "oracle"),
		virtual,
	)...)
	var output bytes.Buffer
	command.Stdout, command.Stderr = &output, &output
	if runError := childguard.Run(command, childguard.Options{Unbounded: true}); runError != nil {
		return fmt.Errorf("Go oracle: %w\n%s", runError, &output)
	}
	return nil
}
