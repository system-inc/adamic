package json

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

// Every native build json's ordinary tests run is a keyed product built ahead by a TestProduct_ of its own, so no
// ordinary test compiles C (#tze35sy, Loom's build law). Built inside the tests they were most of json#1/23's time:
// TestCachedWidthMutantIsCaught 33 to 40 s, TestAdditionalJSONBoundaries 20 to 22 s, TestClosedEmptyFallbackGap 9 to
// 10 s (Loom's uploaded test logs, Cloud and Home, Oct 18). The unmutated port's builds are portMatchesProduct's.

// cachedWidthMutant is TestCachedWidthMutantIsCaught's mutant, which it runs as a release build.
var cachedWidthMutant = printerMutation{"zero cached width", "doc.ts", "width: stringWidth(text)", "width: 0"}

// formatterMutants are TestThreePortMutantsAreCaught's mutants, which it runs as sanitized builds.
var formatterMutants = []printerMutation{
	{"missing final newline", "formatter.ts", "printer.docs.concat([document, printer.docs.line(false, true)])", "document"},
	{"missing colon space", "formatter.ts", "documents.text(': ')", "documents.text(':')"},
	{"wrong filename parser", "formatter.ts", "base === 'package.json'", "base === 'other-package.json'"},
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestProduct_JSONCachedWidthMutantRelease(t *testing.T) {
	requireNativePort(t, mutantPortProduct(t, cachedWidthMutant, native.Options{}))
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestProduct_JSONFormatterMutantSanitized_0(t *testing.T) {
	requireNativePort(t, mutantPortProduct(t, formatterMutants[0], native.Options{Sanitize: true}))
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestProduct_JSONFormatterMutantSanitized_1(t *testing.T) {
	requireNativePort(t, mutantPortProduct(t, formatterMutants[1], native.Options{Sanitize: true}))
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestProduct_JSONFormatterMutantSanitized_2(t *testing.T) {
	requireNativePort(t, mutantPortProduct(t, formatterMutants[2], native.Options{Sanitize: true}))
}

// Not parallel: native.Build writes the shared user cache directory adamic/runtime (and adamic/units when split builds are enabled).
func TestProduct_JSONEmptyFallbackGapSanitized(t *testing.T) {
	requireNativePort(t, gapPortProduct(t, "emptyFallback.ts", native.Options{Sanitize: true}))
}

func requireNativePort(t *testing.T, directory string) {
	t.Helper()
	buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "port"})
}

// mutantPortProduct is the port with mutation applied, lowered in this process and built by clang with options: keyed,
// as json's lowered port is, by the compiler and the port's sources, with this file, the mutation and the native build.
// buildJSONLoweredPortMutation refuses a mutation that doesn't change exactly one place.
func mutantPortProduct(t *testing.T, mutation printerMutation, options native.Options) string {
	t.Helper()
	inputs := jsonLoweredPortInputs()
	inputs.Name = fmt.Sprintf("json %s mutant native %s", mutation.name, nativeBuildName(options))
	inputs.Files = append(inputs.Files, "stage1/cohere/json/native_products_test.go")
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("mutation=%s:%q=>%q", mutation.file, mutation.from, mutation.to))
	inputs.Flags = append(inputs.Flags, nativeBuildFlags(options)...)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	return buildcache.Product(t, inputs, func(directory string) error {
		source := filepath.Join(directory, "source")
		if err := os.Mkdir(source, 0755); err != nil {
			return err
		}
		if err := buildJSONLoweredPortMutation(source, &mutation); err != nil {
			return err
		}
		c, err := os.ReadFile(filepath.Join(source, "main.c"))
		if err != nil {
			return err
		}
		return native.Build(string(c), filepath.Join(directory, "port"), options)
	})
}

// gapPortProduct is gaps/<file> lowered in this process and built by clang with options, keyed by the compiler, the
// gap's source and the native build.
func gapPortProduct(t *testing.T, file string, options native.Options) string {
	t.Helper()
	inputs := buildcache.Inputs{
		Name:      fmt.Sprintf("json gap %s native %s", file, nativeBuildName(options)),
		Files:     append(compilerinputs.Lowering(), "stage1/cohere/json/gaps/"+file, "stage1/cohere/json/native_products_test.go"),
		Flags:     append([]string{"load.Load", "lower.Lower", "native.C"}, nativeBuildFlags(options)...),
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}
	return buildcache.Product(t, inputs, func(directory string) error {
		path, err := filepath.Abs(filepath.Join("gaps", file))
		if err != nil {
			return err
		}
		program, err := load.Load([]string{path})
		if err != nil {
			return fmt.Errorf("Load: %w", err)
		}
		lowered, err := lower.Lower(context.Background(), program)
		if err != nil {
			return fmt.Errorf("Lower: %w", err)
		}
		return native.Build(native.C(lowered), filepath.Join(directory, "port"), options)
	})
}

func nativeBuildName(options native.Options) string {
	if options.Sanitize {
		return "sanitized"
	}
	return "release"
}

// nativeBuildFlags are what a native build's output depends on beyond its C: the options, clang's flags, and the
// environment native.Build and clang read.
func nativeBuildFlags(options native.Options) []string {
	flags := append([]string{fmt.Sprintf("options=%+v", options)}, native.Flags(options)...)
	return append(flags, jsonBuildEnvironment("ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET")...)
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
