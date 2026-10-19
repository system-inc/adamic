package css

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
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

// programSanitized is the program at path, relative to this package, lowered in this process and built by clang under
// the sanitizers, keyed by the compiler's paths, the program's import closure and the native build.
func programSanitized(t *testing.T, path string) string {
	t.Helper()
	inputs := sanitizedInputs("css "+path+" native sanitized", importClosure(t, path)...)
	return buildcache.Product(t, inputs, func(directory string) error {
		entry, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		return buildSanitized(entry, directory)
	})
}

// relativeImport is a relative specifier a source imports or re-exports from.
var relativeImport = regexp.MustCompile(`(?:from|import)\s+['"](\.[^'"]*)['"]`)

// importClosure is the program at path, relative to this package, and every source it imports, followed import by
// import, relative to the repository root and sorted: the sources its lowering reads beyond the compiler. It is read
// from the tree each time it keys, so a source that gains an import moves the key.
func importClosure(t *testing.T, path string) []string {
	t.Helper()
	root, err := filepath.Abs(repository)
	if err != nil {
		t.Fatal(err)
	}
	entry, err := filepath.Abs(path)
	if err != nil {
		t.Fatal(err)
	}
	seen := map[string]bool{}
	pending := []string{entry}
	for len(pending) > 0 {
		source := pending[len(pending)-1]
		pending = pending[:len(pending)-1]
		if seen[source] {
			continue
		}
		seen[source] = true
		contents, err := os.ReadFile(source)
		if err != nil {
			t.Fatal(err)
		}
		for _, match := range relativeImport.FindAllStringSubmatch(string(contents), -1) {
			pending = append(pending, filepath.Join(filepath.Dir(source), match[1]))
		}
	}
	var sources []string
	for source := range seen {
		relative, err := filepath.Rel(root, source)
		if err != nil {
			t.Fatal(err)
		}
		sources = append(sources, filepath.ToSlash(relative))
	}
	sort.Strings(sources)
	return sources
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
