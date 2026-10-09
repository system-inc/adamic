package tsprinter

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

type printerGateProductState struct {
	once      sync.Once
	directory string
}

var printerGateProducts sync.Map

// Only process-local initialization is held here. Every persistent product is
// owned, keyed and published by internal/buildcache.
func printerGateOnce(t *testing.T, name string, build func() string) string {
	t.Helper()
	value, _ := printerGateProducts.LoadOrStore(name, &printerGateProductState{})
	state := value.(*printerGateProductState)
	state.once.Do(func() { state.directory = build() })
	if state.directory == "" {
		t.Fatalf("product %s preparation failed", name)
	}
	return state.directory
}

func printerGateGoOracle(t *testing.T) string {
	t.Helper()
	return printerGateOnce(t, "go-oracle", func() string {
		flags := []string{"go test -c -trimpath -ldflags=-buildid= -overlay=<scratch>/overlay.json ./internal/format/javascript"}
		for _, name := range []string{"GOFLAGS", "GOTOOLCHAIN", "GOEXPERIMENT", "CGO_ENABLED", "GOOS", "GOARCH", "CC", "CXX", "CGO_CFLAGS", "CGO_CPPFLAGS", "CGO_CXXFLAGS", "CGO_LDFLAGS"} {
			flags = append(flags, name+"="+os.Getenv(name))
		}
		return buildcache.Product(t, buildcache.Inputs{
			Name: "tsprinter-go-oracle-v2", Files: []string{"go.mod", "go.work", "cohere", "stage1/cohere/tsprinter/testdata/expressions_side_test.go", "stage1/cohere/tsprinter/testdata/statements_side_test.go", "stage1/cohere/tsprinter/printer_gate_products_test.go"},
			Flags: flags, Toolchain: []string{buildcache.Tool("go", "version"), runtime.GOOS, runtime.GOARCH},
		}, func(dir string) error {
			root, err := filepath.Abs(repository)
			if err != nil {
				return err
			}
			replace := map[string]string{}
			for _, family := range []string{"expressions", "statements"} {
				replace[root+"/cohere/internal/format/javascript/adamic_"+family+"_test.go"] = root + "/stage1/cohere/tsprinter/testdata/" + family + "_side_test.go"
			}
			overlay, err := json.Marshal(map[string]any{"Replace": replace})
			if err != nil {
				return err
			}
			scratch := t.TempDir()
			path := filepath.Join(scratch, "overlay.json")
			if err := os.WriteFile(path, overlay, 0644); err != nil {
				return err
			}
			ctx, cancel := context.WithCancel(context.Background())
			defer cancel()
			command := statementContextCommand(ctx, "go", "test", "-c", "-trimpath", "-ldflags=-buildid=", "-overlay="+path, "-o="+dir+"/oracle", "./internal/format/javascript")
			command.Dir = root + "/cohere"
			output, err := command.CombinedOutput()
			if err != nil {
				return fmt.Errorf("Go printer oracle: %w\n%s", err, output)
			}
			return nil
		})
	})
}

func printerGateLowered(t *testing.T, family string) string {
	t.Helper()
	return printerGateOnce(t, "lowered-"+family, func() string {
		entry := "main.ts"
		if family == "statements" {
			entry = "statementsMain.ts"
		}
		path, err := filepath.Abs(entry)
		if err != nil {
			t.Fatal(err)
		}
		return buildcache.Product(t, buildcache.Inputs{
			Name:  "tsprinter-lowered-" + family + "-v2",
			Files: []string{"go.mod", "go.work", "cohere", "internal", "stage1/cohere/tsprinter", "stage1/typescript"},
			Flags: []string{"entry=" + entry}, Toolchain: []string{buildcache.Tool("go", "version"), runtime.Version(), runtime.GOOS, runtime.GOARCH},
		}, func(dir string) error {
			program := lowered(t, path)
			if err := os.WriteFile(dir+"/port.c", []byte(native.C(program)), 0644); err != nil {
				return err
			}
			return os.WriteFile(dir+"/program.mjs", []byte(javascript.JavaScript(program)), 0644)
		})
	})
}

func printerGateNative(t *testing.T, family string, sanitize bool) string {
	t.Helper()
	name := fmt.Sprintf("native-%s-%t", family, sanitize)
	return printerGateOnce(t, name, func() string {
		lowered := printerGateLowered(t, family)
		source, err := os.ReadFile(lowered + "/port.c")
		if err != nil {
			t.Fatal(err)
		}
		options := native.Options{Sanitize: sanitize, Split: true, Jobs: 4}
		flags := append([]string{"C sha256=" + statementBytesHash(source), "split=true", "jobs=4"}, native.Flags(options)...)
		for _, variable := range []string{"ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
			flags = append(flags, variable+"="+os.Getenv(variable))
		}
		return buildcache.Product(t, buildcache.Inputs{
			Name: "tsprinter-" + name + "-v2", Files: []string{"internal/native", "stage1/cohere/tsprinter/printer_gate_products_test.go"}, Flags: flags,
			Toolchain: []string{buildcache.Tool("clang", "--version"), buildcache.Tool("ar", "--version"), buildcache.Tool("ld", "--version"), runtime.Version(), runtime.GOOS, runtime.GOARCH},
		}, func(dir string) error { return native.Build(string(source), dir+"/port", options) })
	})
}

func printerGateFamilyProducts(t *testing.T, family string) tsPrinterProducts {
	t.Helper()
	entry := "main.ts"
	if family == "statements" {
		entry = "statementsMain.ts"
	}
	path, err := filepath.Abs(entry)
	if err != nil {
		t.Fatal(err)
	}
	return tsPrinterProducts{source: path, backend: printerGateLowered(t, family) + "/program.mjs", sanitized: printerGateNative(t, family, true) + "/port", release: printerGateNative(t, family, false) + "/port"}
}

func TestProduct_TSPrinterGoOracle(t *testing.T) { t.Parallel(); printerGateGoOracle(t) }
func TestProduct_TSPrinterExpressionsLowered(t *testing.T) {
	t.Parallel()
	printerGateLowered(t, "expressions")
}
func TestProduct_TSPrinterExpressionsSanitized(t *testing.T) {
	t.Parallel()
	printerGateNative(t, "expressions", true)
}
func TestProduct_TSPrinterExpressionsRelease(t *testing.T) {
	t.Parallel()
	printerGateNative(t, "expressions", false)
}
func TestProduct_TSPrinterStatementsLowered(t *testing.T) {
	t.Parallel()
	printerGateLowered(t, "statements")
}
func TestProduct_TSPrinterStatementsSanitized(t *testing.T) {
	t.Parallel()
	printerGateNative(t, "statements", true)
}
func TestProduct_TSPrinterStatementsRelease(t *testing.T) {
	t.Parallel()
	printerGateNative(t, "statements", false)
}
