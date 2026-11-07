package markdownblocks

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
	"github.com/system-inc/adamic/internal/javascript"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Every native build the gap witnesses and the parser stage probes run is a keyed product built ahead by a TestProduct_
// of its own, so no ordinary test compiles C (#tze35sy, Loom's build law): nine gaps under the sanitizers, and the
// chunks, front-matter and prefix probes sanitized and release, were built inside the tests.
// TestMdastIdentifierWitnesses runs the mdast grain's sanitized port, which TestProduct_GrainMarkdown_mdast_sanitized
// already owns.

func TestProduct_MarkdownblocksGapsSanitized(test *testing.T) {
	test.Parallel()
	for _, probe := range representationProbes {
		if probe.closed() {
			requirePort(test, gapSanitized(test, probe.path))
		}
	}
	requirePort(test, gapSanitized(test, "gaps/13_array_growth.ts"))
	requirePort(test, gapSanitized(test, "gaps/6_uninitialized_optional_string.ts"))
}

func TestProduct_MarkdownblocksProbesSanitized(test *testing.T) {
	test.Parallel()
	for _, probe := range nativeProbes {
		requirePort(test, probeNative(test, probe, native.Options{Sanitize: true}))
	}
}

func TestProduct_MarkdownblocksProbesRelease(test *testing.T) {
	test.Parallel()
	for _, probe := range nativeProbes {
		requirePort(test, probeNative(test, probe, native.Options{}))
	}
}

func TestProduct_MarkdownblocksProbesEmitted(test *testing.T) {
	test.Parallel()
	for _, probe := range emittedProbes {
		buildcache.RequireArtifacts(test, probeEmitted(test, probe), buildcache.Artifact{Name: "port.mjs"})
	}
}

func requirePort(test *testing.T, directory string) {
	test.Helper()
	buildcache.RequireArtifacts(test, directory, buildcache.Artifact{Name: "port"})
}

const (
	chunksProbe      = "testdata/chunks_probe.ts"
	frontMatterProbe = "testdata/frontmatter_probe.ts"
	prefixProbe      = "testdata/prefix_probe.ts"
)

// nativeProbes are the probes TestMicromarkInputChunks, TestFrontMatterStage and TestMarkdownParserPrefixes run
// natively, sanitized for parity and release for throughput.
var nativeProbes = []string{chunksProbe, frontMatterProbe, prefixProbe}

// emittedProbes are the probes whose lowering a test runs on the JavaScript backend: TestMicromarkInputChunks, which
// lowered it inside its own clock.
var emittedProbes = []string{chunksProbe}

// gapSanitized is the gap at path, relative to this package, lowered in this process and built by clang under the
// sanitizers, keyed by the compiler's paths, the gap's import closure and the native build.
func gapSanitized(test *testing.T, path string) string {
	test.Helper()
	options := native.Options{Sanitize: true}
	inputs := loweredNativeInputs("markdownblocks gap "+path+" native sanitized", options, importClosure(test, path)...)
	return buildcache.Product(test, inputs, func(directory string) error {
		return buildLoweredNative(path, directory, options)
	})
}

// probeNative is the probe at path, relative to this package, lowered in this process and built by clang with options,
// keyed by the compiler's paths, the probe's import closure and the native build.
func probeNative(test *testing.T, path string, options native.Options) string {
	test.Helper()
	mode := "release"
	if options.Sanitize {
		mode = "sanitized"
	}
	inputs := loweredNativeInputs("markdownblocks probe "+path+" native "+mode, options, importClosure(test, path)...)
	return buildcache.Product(test, inputs, func(directory string) error {
		return buildLoweredNative(path, directory, options)
	})
}

// probeEmitted is the probe at path, relative to this package, lowered in this process and emitted as JavaScript into
// directory/port.mjs, keyed by the compiler's paths and the probe's import closure.
func probeEmitted(test *testing.T, path string) string {
	test.Helper()
	inputs := loweredEmittedInputs("markdownblocks probe "+path+" emitted JavaScript", importClosure(test, path)...)
	return buildcache.Product(test, inputs, func(directory string) error {
		entry, absolutePathError := filepath.Abs(path)
		if absolutePathError != nil {
			return absolutePathError
		}
		program, loweredResultError := loweredResult(entry)
		if loweredResultError != nil {
			return loweredResultError
		}
		return os.WriteFile(filepath.Join(directory, "port.mjs"), []byte(javascript.JavaScript(program)), 0644)
	})
}

// relativeImport is a relative specifier a source imports or re-exports from.
var relativeImport = regexp.MustCompile(`(?:from|import)\s+['"](\.[^'"]*)['"]`)

// importClosure is the program at path, relative to this package, and every source it imports, followed import by
// import, relative to the repository root and sorted: the sources its lowering reads beyond the compiler. It is read
// from the tree each time it keys, so a source that gains an import moves the key.
func importClosure(test *testing.T, path string) []string {
	test.Helper()
	root, absolutePathError := filepath.Abs(repository)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
	}
	entry, absolutePathError := filepath.Abs(path)
	if absolutePathError != nil {
		test.Fatal(absolutePathError)
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
		contents, readError := os.ReadFile(source)
		if readError != nil {
			test.Fatal(readError)
		}
		for _, match := range relativeImport.FindAllStringSubmatch(string(contents), -1) {
			pending = append(pending, filepath.Join(filepath.Dir(source), match[1]))
		}
	}
	var sources []string
	for source := range seen {
		relative, relativePathError := filepath.Rel(root, source)
		if relativePathError != nil {
			test.Fatal(relativePathError)
		}
		sources = append(sources, filepath.ToSlash(relative))
	}
	sort.Strings(sources)
	return sources
}

// loweredNativeInputs keys a native build lowered in this process: the compiler is linked into this test binary, where
// no trace sees it read, so its paths are named (compilerinputs.Lowering) beside the sources, relative to the
// repository, and this file, with the native options and flags, the environment native.Build and clang read, and
// clang's version.
func loweredNativeInputs(name string, options native.Options, sources ...string) buildcache.Inputs {
	flags := append(
		[]string{"load.Load", "lower.Lower", "native.C", fmt.Sprintf("options=%+v", options)},
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
		value, set := os.LookupEnv(variable)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", variable, set, value))
	}
	files := append(compilerinputs.Lowering(), sources...)
	return buildcache.Inputs{
		Name:      name,
		Files:     append(files, "stage1/cohere/markdownblocks/native_products_test.go"),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}
}

// loweredEmittedInputs keys a lowering emitted as JavaScript in this process: as loweredNativeInputs, without the
// native build's options, the environment clang reads, or clang.
func loweredEmittedInputs(name string, sources ...string) buildcache.Inputs {
	files := append(compilerinputs.Lowering(), sources...)
	return buildcache.Inputs{
		Name:      name,
		Files:     append(files, "stage1/cohere/markdownblocks/native_products_test.go"),
		Flags:     []string{"load.Load", "lower.Lower", "javascript.JavaScript"},
		Toolchain: []string{runtime.Version()},
	}
}

// buildLoweredNative lowers the program at path, relative to this package, and builds it with options as
// directory/port.
func buildLoweredNative(path, directory string, options native.Options) error {
	entry, absolutePathError := filepath.Abs(path)
	if absolutePathError != nil {
		return absolutePathError
	}
	program, loadError := load.Load([]string{entry})
	if loadError != nil {
		return fmt.Errorf("Load: %w", loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		return fmt.Errorf("Lower: %w", lowerError)
	}
	return native.Build(native.C(lowered), filepath.Join(directory, "port"), options)
}

// sanitizedRun runs a sanitized build with leak detection off, as the oracle does; leaks is its own run.
func sanitizedRun(test *testing.T, binary string, arguments ...string) run {
	test.Helper()
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return execute(test, environment, binary, arguments...)
}
