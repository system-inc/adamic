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
var cssLoweringFiles = []string{"stage1/cohere/css", "stage1/cohere/selector", "stage1/cohere/values", "stage1/cohere/mediaquery", "stage1/cohere/cssstrings", "stage1/cohere/cssnumbers", "internal", "bridge/tsgo", "cohere/policy", "cohere/internal", "cohere/rule_runner", "cohere/go.mod", "cohere/go.sum", "cohere/TypeScript/tsc", "cohere/TypeScript-shim", "go.mod", "go.work"}

// cssOptimizedMains are the programs whose unsanitized natives the optimized tests run: the printer and the parser.
var cssOptimizedMains = []string{"print_main.ts", "compose_main.ts"}

var cssOptimizedOnce [2]sync.Once
var cssOptimizedNatives [2]string

// cssOptimizedNative is one of cssOptimizedMains' unsanitized native, built ahead as a product: lowering it and
// compiling it took most of TestCSSPrinterOptimizedMatchesGo's 34.1 s and TestCSSParserOptimizedMatchesNode's 24.1 s
// alone on one core (Home). Its key is what lowering reads, as cssPrinterProduct's is, with the build's flags.
func cssOptimizedNative(t *testing.T, index int) string {
	t.Helper()
	cssOptimizedOnce[index].Do(func() {
		inputs := buildcache.Inputs{
			Name:      "css-optimized-native-" + cssOptimizedMains[index],
			Files:     cssLoweringFiles,
			Flags:     append(native.Flags(native.Options{}), "ADAMIC_NATIVE_SPLIT="+os.Getenv("ADAMIC_NATIVE_SPLIT")),
			Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version"), runtime.GOOS, runtime.GOARCH},
		}
		directory := buildcache.Product(t, inputs, func(dir string) error {
			main, err := filepath.Abs(cssOptimizedMains[index])
			if err != nil {
				return err
			}
			checked, err := load.Load([]string{main})
			if err != nil {
				return err
			}
			program, err := lower.Lower(t.Context(), checked)
			if err != nil {
				return err
			}
			return native.Build(native.C(program), filepath.Join(dir, "native"), native.Options{})
		})
		cssOptimizedNatives[index] = filepath.Join(directory, "native")
	})
	if cssOptimizedNatives[index] == "" {
		t.Fatalf("css optimized native %s unavailable: its build failed earlier in this process (see that test's log)", cssOptimizedMains[index])
	}
	return cssOptimizedNatives[index]
}

func TestProduct_CSSPrinterOptimizedNative(t *testing.T) { t.Parallel(); cssOptimizedNative(t, 0) }
func TestProduct_CSSParserOptimizedNative(t *testing.T)  { t.Parallel(); cssOptimizedNative(t, 1) }

// The optimized artifact must match the full Go oracle too; sanitizer builds
// use a different optimizer setting and the timing checksum alone is weaker.
// One test per mode: both modes in one test took 25.1 s alone on one core (Home), with its native a product.
func cssPrinterOptimizedMatchesGo(t *testing.T, mode string) {
	t.Helper()
	cases, _ := askedCases(t)
	binary := cssOptimizedNative(t, 0)
	expected := printerAnswers(t, cases, mode)
	result := execute(t, nil, binary, cases, "output", "once", mode)
	if result.exitCode != 0 || len(result.stderr) != 0 {
		t.Fatalf("optimized %s: %d %s", mode, result.exitCode, result.stderr)
	}
	if difference := firstDifference(string(result.stdout), expected); difference != "" {
		t.Fatal(difference)
	}
	t.Logf("optimized %s: %d byte-identical Go results", mode, strings.Count(expected, "\n")/2)
}

// Not parallel: askedCases writes the shared ADAMIC_CSS_KEEP and ADAMIC_CSS_KEEP_RAW paths when configured.
func TestCSSPrinterOptimizedMatchesGo_Default(t *testing.T) {
	cssPrinterOptimizedMatchesGo(t, "default")
}

// Not parallel: askedCases writes the shared ADAMIC_CSS_KEEP and ADAMIC_CSS_KEEP_RAW paths when configured.
func TestCSSPrinterOptimizedMatchesGo_Narrow(t *testing.T) { cssPrinterOptimizedMatchesGo(t, "narrow") }

// Not parallel: askedCases writes the shared ADAMIC_CSS_KEEP and ADAMIC_CSS_KEEP_RAW paths when configured.
func TestCSSParserOptimizedMatchesNode(t *testing.T) {
	cases, _ := askedCases(t)
	path, _ := filepath.Abs("compose_main.ts")
	binary := cssOptimizedNative(t, 1)
	expected := onNode(t, path, cases)
	result := execute(t, nil, binary, cases)
	for _, side := range []struct {
		name   string
		result run
	}{{"Node", expected}, {"optimized native", result}} {
		if side.result.exitCode != 0 || len(side.result.stderr) != 0 {
			t.Fatalf("%s: %d %s", side.name, side.result.exitCode, side.result.stderr)
		}
	}
	if difference := firstDifference(string(result.stdout), string(expected.stdout)); difference != "" {
		t.Fatal(difference)
	}
	t.Logf("optimized parser: %d byte-identical Node trees/errors, held to Go by the composed-tree test", strings.Count(string(expected.stdout), "\n")/2)
}
