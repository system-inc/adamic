package markdownblocks

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

// Every native build the gap witnesses and the parser stage probes run is a keyed product built ahead by a TestProduct_
// of its own, so no ordinary test compiles C (#tze35sy, Loom's build law): nine gaps under the sanitizers, and the
// chunks, front-matter and prefix probes sanitized and release, were built inside the tests.
// TestMdastIdentifierWitnesses runs the mdast grain's sanitized port, which TestProduct_GrainMarkdown_mdast_sanitized
// already owns.

func TestProduct_MarkdownblocksGapsSanitized(t *testing.T) {
	t.Parallel()
	for _, probe := range representationProbes {
		if probe.closed() {
			requirePort(t, gapSanitized(t, probe.path, probe.imports...))
		}
	}
	requirePort(t, gapSanitized(t, "gaps/13_array_growth.ts"))
	requirePort(t, gapSanitized(t, "gaps/6_uninitialized_optional_string.ts"))
}

func TestProduct_MarkdownblocksProbesSanitized(t *testing.T) {
	t.Parallel()
	for _, probe := range nativeProbes {
		requirePort(t, probeNative(t, probe, native.Options{Sanitize: true}))
	}
}

func TestProduct_MarkdownblocksProbesRelease(t *testing.T) {
	t.Parallel()
	for _, probe := range nativeProbes {
		requirePort(t, probeNative(t, probe, native.Options{}))
	}
}

func requirePort(t *testing.T, directory string) {
	t.Helper()
	buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "port"})
}

// A nativeProbe is a testdata probe and the package's sources it imports, together its whole import closure.
type nativeProbe struct {
	main    string
	sources []string
}

var (
	chunksProbe      = nativeProbe{"testdata/chunks_probe.ts", []string{"inputChunks.ts"}}
	frontMatterProbe = nativeProbe{"testdata/frontmatter_probe.ts", []string{"frontmatter.ts", "codec.ts"}}
	prefixProbe      = nativeProbe{"testdata/prefix_probe.ts", []string{"preprocess.ts", "quotePrefix.ts", "codec.ts"}}
)

// nativeProbes are the probes TestMicromarkInputChunks, TestFrontMatterStage and TestMarkdownParserPrefixes run
// natively, sanitized for parity and release for throughput.
var nativeProbes = []nativeProbe{chunksProbe, frontMatterProbe, prefixProbe}

// gapSanitized is the gap at path, relative to this package, lowered in this process and built by clang under the
// sanitizers, keyed by the compiler's paths, the gap and the package's sources it imports, and the native build.
func gapSanitized(t *testing.T, path string, imports ...string) string {
	t.Helper()
	options := native.Options{Sanitize: true}
	inputs := loweredNativeInputs("markdownblocks gap "+path+" native sanitized", options, append([]string{path}, imports...)...)
	return buildcache.Product(t, inputs, func(directory string) error {
		return buildLoweredNative(path, directory, options)
	})
}

// probeNative is probe lowered in this process and built by clang with options, keyed by the compiler's paths, the
// probe's import closure and the native build.
func probeNative(t *testing.T, probe nativeProbe, options native.Options) string {
	t.Helper()
	mode := "release"
	if options.Sanitize {
		mode = "sanitized"
	}
	inputs := loweredNativeInputs("markdownblocks probe "+probe.main+" native "+mode, options, append([]string{probe.main}, probe.sources...)...)
	return buildcache.Product(t, inputs, func(directory string) error {
		return buildLoweredNative(probe.main, directory, options)
	})
}

// loweredNativeInputs keys a native build lowered in this process: the compiler is linked into this test binary, where
// no trace sees it read, so its paths are named (compilerinputs.Lowering) beside the sources, relative to this package,
// and this file, with the native options and flags, the environment native.Build and clang read, and clang's version.
func loweredNativeInputs(name string, options native.Options, sources ...string) buildcache.Inputs {
	flags := append([]string{"load.Load", "lower.Lower", "native.C", fmt.Sprintf("options=%+v", options)}, native.Flags(options)...)
	for _, variable := range []string{"ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		value, set := os.LookupEnv(variable)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", variable, set, value))
	}
	files := compilerinputs.Lowering()
	for _, source := range sources {
		files = append(files, "stage1/cohere/markdownblocks/"+source)
	}
	return buildcache.Inputs{
		Name:      name,
		Files:     append(files, "stage1/cohere/markdownblocks/native_products_test.go"),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}
}

// buildLoweredNative lowers the program at path, relative to this package, and builds it with options as
// directory/port.
func buildLoweredNative(path, directory string, options native.Options) error {
	entry, err := filepath.Abs(path)
	if err != nil {
		return err
	}
	program, err := load.Load([]string{entry})
	if err != nil {
		return fmt.Errorf("Load: %w", err)
	}
	lowered, err := lower.Lower(context.Background(), program)
	if err != nil {
		return fmt.Errorf("Lower: %w", err)
	}
	return native.Build(native.C(lowered), filepath.Join(directory, "port"), options)
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
