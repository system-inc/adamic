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
// in the key by name (compilerinputs.Lowering), not cohere whole, whose TypeScript submodule a populated tree hashed
// for every lookup (#4ycgm2n); then the port and the TypeScript it imports. TestPrinterGateLoweredKeyCoversWhatItReads
// holds it to what the loader reads.
func printerGateLoweredFiles() []string {
	return append(compilerinputs.Lowering(), "stage1/cohere/tsprinter", "stage1/typescript")
}

func printerGateLowered(test *testing.T, family string) string {
	test.Helper()
	return oracles.PrinterGateOnce(test, "lowered-"+family, func() string {
		entry := printerGateEntry(family)
		path, absolutePathError := filepath.Abs(entry)
		if absolutePathError != nil {
			test.Fatal(absolutePathError)
		}
		return buildcache.Product(test, buildcache.Inputs{
			Name:      "tsprinter-lowered-" + family + "-v3",
			Files:     printerGateLoweredFiles(),
			Flags:     []string{"entry=" + entry},
			Toolchain: []string{buildcache.Tool("go", "version"), runtime.Version(), runtime.GOOS, runtime.GOARCH},
		}, func(dir string) error {
			program := lowered(test, path)
			if writeError := os.WriteFile(dir+"/port.c", []byte(native.C(program)), 0644); writeError != nil {
				return writeError
			}
			return os.WriteFile(dir+"/program.mjs", []byte(javascript.JavaScript(program)), 0644)
		})
	})
}

func printerGateNative(test *testing.T, family string, sanitize bool) string {
	test.Helper()
	name := fmt.Sprintf("native-%s-%t", family, sanitize)
	return oracles.PrinterGateOnce(test, name, func() string {
		lowered := printerGateLowered(test, family)
		source, readError := os.ReadFile(lowered + "/port.c")
		if readError != nil {
			test.Fatal(readError)
		}
		options := native.Options{Sanitize: sanitize, Split: true, Jobs: 4}
		flags := append(
			[]string{"C sha256=" + oracles.StatementBytesHash(source), "split=true", "jobs=4"},
			native.Flags(options)...,
		)
		for _, variable := range []string{
			"ADAMIC_NATIVE_SPLIT",
			"CPATH",
			"C_INCLUDE_PATH",
			"LIBRARY_PATH",
			"SDKROOT",
			"MACOSX_DEPLOYMENT_TARGET",
		} {
			flags = append(flags, variable+"="+os.Getenv(variable))
		}
		return buildcache.Product(test, buildcache.Inputs{
			Name:  "tsprinter-" + name + "-v2",
			Files: []string{"internal/native", "stage1/cohere/tsprinter/printer_gate_products_test.go"},
			Flags: flags,
			Toolchain: []string{
				buildcache.Tool("clang", "--version"),
				buildcache.Tool("ar", "--version"),
				buildcache.Tool("ld", "--version"),
				runtime.Version(),
				runtime.GOOS,
				runtime.GOARCH,
			},
		}, func(dir string) error { return native.Build(string(source), dir+"/port", options) })
	})
}

func printerGateFamilyProducts(test *testing.T, family string) tsPrinterProducts {
	test.Helper()
	path, absolutePathError := filepath.Abs(printerGateEntry(family))
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	return tsPrinterProducts{
		source:    path,
		backend:   printerGateLowered(test, family) + "/program.mjs",
		sanitized: printerGateNative(test, family, true) + "/port",
		release:   printerGateNative(test, family, false) + "/port",
	}
}

func TestProduct_TSPrinterExpressionsLowered(test *testing.T) {
	test.Parallel()
	printerGateLowered(test, "expressions")
}
func TestProduct_TSPrinterExpressionsSanitized(test *testing.T) {
	test.Parallel()
	printerGateNative(test, "expressions", true)
}
func TestProduct_TSPrinterExpressionsRelease(test *testing.T) {
	test.Parallel()
	printerGateNative(test, "expressions", false)
}
func TestProduct_TSPrinterStatementsLowered(test *testing.T) {
	test.Parallel()
	printerGateLowered(test, "statements")
}
func TestProduct_TSPrinterStatementsSanitized(test *testing.T) {
	test.Parallel()
	printerGateNative(test, "statements", true)
}
func TestProduct_TSPrinterStatementsRelease(test *testing.T) {
	test.Parallel()
	printerGateNative(test, "statements", false)
}

// The document port's lowered program and natives are products, as expressions' and statements' are: building the
// sanitized native inside TestDocumentsAgainstGoAndPrettier was 5.8 of its 10.6 s alone on one core, and most of its
// 65.5 s at load 37 on Cloud, which reddened a lint-only change by the time budget (#g7a9yd3).
func TestProduct_TSPrinterDocumentsLowered(test *testing.T) {
	test.Parallel()
	printerGateLowered(test, "documents")
}
func TestProduct_TSPrinterDocumentsSanitized(test *testing.T) {
	test.Parallel()
	printerGateNative(test, "documents", true)
}
func TestProduct_TSPrinterDocumentsRelease(test *testing.T) {
	test.Parallel()
	printerGateNative(test, "documents", false)
}
