package css

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// cssLoweringFiles are what lowering a css program reads, the files cssPrinterProduct declares.
var cssLoweringFiles = []string{
	"stage1/cohere/css",
	"stage1/cohere/selector",
	"stage1/cohere/values",
	"stage1/cohere/mediaquery",
	"stage1/cohere/cssstrings",
	"stage1/cohere/cssnumbers",
	"internal",
	"bridge/tsgo",
	"cohere/policy",
	"cohere/internal",
	"cohere/rule_runner",
	"cohere/go.mod",
	"cohere/go.sum",
	"cohere/TypeScript/tsc",
	"cohere/TypeScript-shim",
	"go.mod",
	"go.work",
}

// cssOptimizedMains are the programs whose unsanitized natives the optimized tests run: the printer and the parser.
var cssOptimizedMains = []string{"print_main.ts", "compose_main.ts"}

var cssOptimizedOnce [2]sync.Once
var cssOptimizedNatives [2]string

// cssOptimizedNative is one of cssOptimizedMains' unsanitized native, built ahead as a product: lowering it and
// compiling it took most of TestCSSPrinterOptimizedMatchesGo's 34.1 s and TestCSSParserOptimizedMatchesNode's 24.1 s
// alone on one core (Home). Its key is what lowering reads, as cssPrinterProduct's is, with the build's flags.
func cssOptimizedNative(test *testing.T, index int) string {
	test.Helper()
	cssOptimizedOnce[index].Do(func() {
		inputs := buildcache.Inputs{
			Name:      "css-optimized-native-" + cssOptimizedMains[index],
			Files:     cssLoweringFiles,
			Flags:     append(native.Flags(native.Options{}), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT")),
			Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version"), runtime.GOOS, runtime.GOARCH},
		}
		directory := buildcache.Product(test, inputs, func(dir string) error {
			main, absolutePathError := filepath.Abs(cssOptimizedMains[index])
			if absolutePathError != nil {
				return absolutePathError
			}
			checked, loadError := load.Load([]string{main})
			if loadError != nil {
				return loadError
			}
			program, lowerError := lower.Lower(test.Context(), checked)
			if lowerError != nil {
				return lowerError
			}
			return native.Build(native.C(program), filepath.Join(dir, "native"), native.Options{})
		})
		cssOptimizedNatives[index] = filepath.Join(directory, "native")
	})
	if cssOptimizedNatives[index] == "" {
		test.Fatalf(
			"css optimized native %s unavailable: its build failed earlier in this process (see that test's log)",
			cssOptimizedMains[index],
		)
	}
	return cssOptimizedNatives[index]
}

func TestProduct_CssPrinterOptimizedNative(test *testing.T) {
	test.Parallel()
	cssOptimizedNative(test, 0)
}
func TestProduct_CssParserOptimizedNative(test *testing.T) {
	test.Parallel()
	cssOptimizedNative(test, 1)
}

// The optimized artifact must match the full Go oracle too; sanitizer builds
// use a different optimizer setting and the timing checksum alone is weaker.
// One test per mode: both modes in one test took 25.1 s alone on one core (Home), with its native a product.
func cssPrinterOptimizedMatchesGo(test *testing.T, mode string) {
	test.Helper()
	cases, _ := askedCases(test)
	binary := cssOptimizedNative(test, 0)
	expected := printerAnswers(test, cases, mode)
	result := execute(test, nil, binary, cases, "output", "once", mode)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		test.Fatalf("optimized %s: %d %s", mode, result.exitCode, result.stderr)
	}
	if difference := firstDifference(string(result.stdout), expected); difference != "" {
		test.Fatal(difference)
	}
	test.Logf("optimized %s: %d byte-identical Go results", mode, strings.Count(expected, "\n")/2)
}

// Not parallel: askedCases writes the shared ADAMIC_CSS_KEEP and ADAMIC_CSS_KEEP_RAW paths when configured.
func TestCssPrinterOptimizedMatchesGo_Default(test *testing.T) {
	cssPrinterOptimizedMatchesGo(test, "default")
}

// Not parallel: askedCases writes the shared ADAMIC_CSS_KEEP and ADAMIC_CSS_KEEP_RAW paths when configured.
func TestCssPrinterOptimizedMatchesGo_Narrow(test *testing.T) {
	cssPrinterOptimizedMatchesGo(test, "narrow")
}

// Not parallel: askedCases writes the shared ADAMIC_CSS_KEEP and ADAMIC_CSS_KEEP_RAW paths when configured.
func TestCssParserOptimizedMatchesNode(test *testing.T) {
	cases, _ := askedCases(test)
	path, _ := filepath.Abs("compose_main.ts")
	binary := cssOptimizedNative(test, 1)
	expected := onNode(test, path, cases)
	result := execute(test, nil, binary, cases)
	for _, side := range []struct {
		name   string
		result run
	}{{"Node", expected}, {"optimized native", result}} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
			test.Fatalf("%s: %d %s", side.name, side.result.exitCode, side.result.stderr)
		}
	}
	if difference := firstDifference(string(result.stdout), string(expected.stdout)); difference != "" {
		test.Fatal(difference)
	}
	test.Logf(
		"optimized parser: %d byte-identical Node trees/errors, held to Go by the composed-tree test",
		strings.Count(string(expected.stdout), "\n")/2,
	)
}
