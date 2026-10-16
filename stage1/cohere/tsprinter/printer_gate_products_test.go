package tsprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

func printerGateLowered(t *testing.T, family string) string {
	t.Helper()
	return oracles.PrinterGateOnce(t, "lowered-"+family, func() string {
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
	return oracles.PrinterGateOnce(t, name, func() string {
		lowered := printerGateLowered(t, family)
		source, err := os.ReadFile(lowered + "/port.c")
		if err != nil {
			t.Fatal(err)
		}
		options := native.Options{Sanitize: sanitize, Split: true, Jobs: 4}
		flags := append([]string{"C sha256=" + oracles.StatementBytesHash(source), "split=true", "jobs=4"}, native.Flags(options)...)
		for _, variable := range []string{"ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
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
