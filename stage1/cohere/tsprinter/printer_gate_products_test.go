package tsprinter

import (
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/stage1/cohere/tsprinter/oracles"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/native"
)

// printerGateEntry is a family's port: expressions' main.ts, statements' statementsMain.ts, documents' docMain.ts.
func printerGateEntry(family string) string {
	switch family {
	case "statements":
		return "statementsMain.ts"
	case "documents":
		return "docMain.ts"
	}
	return "main.ts"
}

// printerGateLoweredFiles are what a family's lowering reads: the compiler, which runs in this process, so its code is
// in the key by name (compilerinputs.Lowering), not cohere whole, whose TypeScript submodule a populated tree hashed for
// every lookup (#4ycgm2n); then the port and the TypeScript it imports. TestPrinterGateLoweredKeyCoversWhatItReads
// holds it to what the loader reads.
func printerGateLoweredFiles() []string {
	return append(compilerinputs.Lowering(), "stage1/cohere/tsprinter", "stage1/typescript")
}

func printerGateLowered(t *testing.T, family string) string {
	t.Helper()
	return oracles.PrinterGateOnce(t, "lowered-"+family, func() string {
		entry := printerGateEntry(family)
		path, err := filepath.Abs(entry)
		if err != nil {
			t.Fatal(err)
		}
		return buildcache.Product(t, buildcache.Inputs{
			Name:  "tsprinter-lowered-" + family + "-v3",
			Files: printerGateLoweredFiles(),
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
	path, err := filepath.Abs(printerGateEntry(family))
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

// The document port's lowered program and natives are products, as expressions' and statements' are: building the
// sanitized native inside TestDocumentsAgainstGoAndPrettier was 5.8 of its 10.6 s alone on one core, and most of its
// 65.5 s at load 37 on Cloud, which reddened a lint-only change by the time budget (#g7a9yd3).
func TestProduct_TSPrinterDocumentsLowered(t *testing.T) {
	t.Parallel()
	printerGateLowered(t, "documents")
}
func TestProduct_TSPrinterDocumentsSanitized(t *testing.T) {
	t.Parallel()
	printerGateNative(t, "documents", true)
}
func TestProduct_TSPrinterDocumentsRelease(t *testing.T) {
	t.Parallel()
	printerGateNative(t, "documents", false)
}
