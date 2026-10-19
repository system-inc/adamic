package css

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Every sanitized build the closed gaps' tests, TestTheCanonicalRangeChecksCanFail and TestCSSPrinterByteSlicesMatchGo
// run is a keyed product built ahead by a TestProduct_ of its own, so none of them compiles C (#tze35sy, Loom's build
// law): seven clang builds, four closed gaps, the raw and composed range guards and the byte slices, were made inside
// the tests.

func TestProduct_CSSGapsSanitized(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"gaps/1_regex_and_value_tree.ts", "gaps/3_optional_boolean_condition.ts", "gaps/4_empty_array_union.ts", "gaps/6_printer_regex_tree.ts"} {
		requirePort(t, programSanitized(t, path))
	}
}

func TestProduct_CSSRangeGuardsSanitized(t *testing.T) {
	t.Parallel()
	for _, path := range []string{"testdata/range_guard.ts", "testdata/composition_range_guard.ts"} {
		requirePort(t, programSanitized(t, path))
	}
}

func TestProduct_CSSByteSlicesSanitized(t *testing.T) {
	t.Parallel()
	requirePort(t, programSanitized(t, "testdata/byte_slices.ts"))
}

func requirePort(t *testing.T, directory string) {
	t.Helper()
	buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "port"})
}

// sanitizedPrograms are the programs this package's ordinary tests run sanitized, each by its path relative to this
// package, with the sources it is lowered from relative to the repository: the program and its import closure.
var sanitizedPrograms = map[string][]string{
	"gaps/1_regex_and_value_tree.ts": {
		"stage1/cohere/css/gaps/1_regex_and_value_tree.ts",
		"stage1/cohere/selector/nodes.ts", "stage1/cohere/selector/parser.ts", "stage1/cohere/selector/source.ts", "stage1/cohere/selector/tokenize.ts",
		"stage1/cohere/values/nodes.ts", "stage1/cohere/values/parser.ts", "stage1/cohere/values/tokenize.ts", "stage1/cohere/values/values.ts",
	},
	"gaps/3_optional_boolean_condition.ts": {"stage1/cohere/css/gaps/3_optional_boolean_condition.ts"},
	"gaps/4_empty_array_union.ts":          {"stage1/cohere/css/gaps/4_empty_array_union.ts"},
	"gaps/6_printer_regex_tree.ts": {
		"stage1/cohere/css/gaps/6_printer_regex_tree.ts",
		"stage1/cohere/mediaquery/index.ts", "stage1/cohere/mediaquery/nodes.ts", "stage1/cohere/mediaquery/parsers.ts", "stage1/cohere/mediaquery/whitespace.ts",
	},
	"testdata/range_guard.ts": {
		"stage1/cohere/css/testdata/range_guard.ts",
		"stage1/cohere/css/input.ts", "stage1/cohere/css/nodes.ts", "stage1/cohere/css/parser.ts", "stage1/cohere/css/tokenize.ts",
		"stage1/cohere/selector/nodes.ts", "stage1/cohere/selector/source.ts",
	},
	"testdata/composition_range_guard.ts": {
		"stage1/cohere/css/testdata/composition_range_guard.ts",
		"stage1/cohere/css/byte_view.ts", "stage1/cohere/css/compose.ts", "stage1/cohere/css/convert.ts", "stage1/cohere/css/input.ts",
		"stage1/cohere/css/location.ts", "stage1/cohere/css/nodes.ts", "stage1/cohere/css/parse_value.ts", "stage1/cohere/css/parser.ts",
		"stage1/cohere/css/tokenize.ts", "stage1/cohere/css/tree.ts",
		"stage1/cohere/mediaquery/index.ts", "stage1/cohere/mediaquery/nodes.ts", "stage1/cohere/mediaquery/parsers.ts", "stage1/cohere/mediaquery/whitespace.ts",
		"stage1/cohere/selector/nodes.ts", "stage1/cohere/selector/parser.ts", "stage1/cohere/selector/source.ts", "stage1/cohere/selector/tokenize.ts",
		"stage1/cohere/values/nodes.ts", "stage1/cohere/values/parser.ts", "stage1/cohere/values/tokenize.ts", "stage1/cohere/values/values.ts",
	},
	"testdata/byte_slices.ts": {
		"stage1/cohere/css/testdata/byte_slices.ts",
		"stage1/cohere/css/byte_view.ts",
		"stage1/cohere/selector/nodes.ts", "stage1/cohere/selector/source.ts",
	},
}

// programSanitized is the program at path, relative to this package and named in sanitizedPrograms, lowered in this
// process and built by clang under the sanitizers, keyed by the compiler's paths, the program's sources and the native
// build.
func programSanitized(t *testing.T, path string) string {
	t.Helper()
	sources, ok := sanitizedPrograms[path]
	if !ok {
		t.Fatalf("%s is not among sanitizedPrograms", path)
	}
	inputs := sanitizedInputs("css "+path+" native sanitized", sources...)
	return buildcache.Product(t, inputs, func(directory string) error {
		entry, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		return buildSanitized(entry, directory)
	})
}

// sanitizedInputs keys a sanitized build lowered in this process: the compiler is linked into this test binary, where
// no trace sees it read, so its paths are named (compilerinputs.Lowering) beside the sources and this file, with the
// native options and flags, the environment native.Build and clang read, and clang's version.
func sanitizedInputs(name string, sources ...string) buildcache.Inputs {
	options := native.Options{Sanitize: true}
	flags := append([]string{"load.Load", "lower.Lower", "native.C", fmt.Sprintf("options=%+v", options)}, native.Flags(options)...)
	for _, variable := range []string{"ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		value, set := os.LookupEnv(variable)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", variable, set, value))
	}
	files := append(compilerinputs.Lowering(), sources...)
	return buildcache.Inputs{
		Name:      name,
		Files:     append(files, "stage1/cohere/css/native_products_test.go"),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}
}

// buildSanitized lowers the program at entry and builds it under the sanitizers as directory/port.
func buildSanitized(entry, directory string) error {
	program, err := load.Load([]string{entry})
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return fmt.Errorf("Lower: %w", err)
	}
	return native.Build(native.C(lowered), filepath.Join(directory, "port"), native.Options{Sanitize: true})
}

// sanitizedRun runs a sanitized build with leak detection off, as the oracle does; leaks is its own run.
func sanitizedRun(t *testing.T, binary string, arguments ...string) run {
	t.Helper()
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return execute(t, environment, binary, arguments...)
}
