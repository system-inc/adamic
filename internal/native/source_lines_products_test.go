package native

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/buildcache/location"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/toolchain"
)

// Only witness tools map their disposable entry directory, through a keyed Go
// overlay. Shipped compilers and the ordinary command.test below keep the real
// directory, which TestDebugSourcePaths proves a debugger can open (#gd4csnt).
var debugWitnessMutations = []buildcache.Mutation{{
	File:   "internal/native/native.go",
	Before: `return []string{"-fdebug-compilation-dir=" + options.DebugDirectory}`,
	After:  `return append([]string{"-fdebug-compilation-dir=" + options.DebugDirectory}, prefixMaps(options.DebugDirectory, "/adamic-debug-witness")...)`,
}}

var debugBuildOptions = []Options{
	{Release: true},
	{Release: true, Debug: true},
	{Release: true, Debug: true, Sanitize: true},
}

// The compiler is linked into this test binary, so declare its inputs as values
// does. The fixtures and this recipe are under internal too. No assertion is
// cached: products hold executables only, and the consumer observes each run.
func debugProductInputs(name string, options Options) buildcache.Inputs {
	flags := append([]string{fmt.Sprintf("options=%+v", options)}, Flags(options)...)
	for _, variable := range []string{
		"ADAMIC_PROGRAM_REGION",
		"ADAMIC_NATIVE_SPLIT",
		"ADAMIC_NATIVE_JOBS",
		"CPATH",
		"C_INCLUDE_PATH",
		"SDKROOT",
		"MACOSX_DEPLOYMENT_TARGET",
	} {
		value, set := os.LookupEnv(variable)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", variable, set, value))
	}
	return buildcache.Inputs{
		Name:      name,
		Files:     append(compilerinputs.Lowering(), "cmd/adamic"),
		Flags:     flags,
		Toolchain: []string{goruntime.Version(), buildcache.Tool("clang", "--version")},
	}
}

// Toolchain commands deliberately drop ADAMIC_* variables. Pin both locations:
// Loom can hand a read-only cache and a separate writable output directory.
func debugCacheEnvironment(test *testing.T, environment []string) []string {
	test.Helper()
	if os.Getenv("ADAMIC_BUILD_CACHE_DIR") == "" {
		test.Fatal("debug products require the unit's ADAMIC_BUILD_CACHE_DIR; refusing the home cache (#phzhe68)")
	}
	cache, buildCacheError := location.BuildCache()
	if buildCacheError != nil {
		test.Fatal(buildCacheError)
	}
	output, outputError := location.Output()
	if outputError != nil {
		test.Fatal(outputError)
	}
	return append(environment, "ADAMIC_BUILD_CACHE_DIR="+cache, "ADAMIC_BUILD_OUTPUT_DIR="+output)
}

func debugProduct(test *testing.T, inputs buildcache.Inputs, build bool, recipe func(string) error) string {
	test.Helper()
	debugCacheEnvironment(test, nil)
	if build {
		return buildcache.Product(test, inputs, recipe)
	}
	directory, builtAheadError := buildcache.BuiltAhead(
		inputs,
		fmt.Errorf("debug product %s not built ahead; run TestProduct_Debug tests first", inputs.Name),
	)
	if builtAheadError != nil {
		test.Fatal(builtAheadError)
	}
	return directory
}

func debugReviewProduct(test *testing.T, path string, mode int, build bool) string {
	test.Helper()
	options := debugBuildOptions[mode]
	return debugProduct(test, debugProductInputs(
		fmt.Sprintf("debug-review-%d:%s", mode, filepath.ToSlash(path)),
		options,
	), build, func(directory string) error {
		program, debugProgramError := debugProgram(test, path, options.Debug)
		if debugProgramError != nil {
			return debugProgramError
		}
		return Build(CWithOptions(program, options), filepath.Join(directory, "program"), options)
	})
}

func runDebugReviewProducts(test *testing.T, shard, mode int) {
	for _, path := range debugFamilyPaths(reviewPrograms(test), shard, false) {
		if filepath.Dir(path) != "../oracle/testdata/review/agree" {
			continue
		}
		if _, statError := os.Stat(strings.TrimSuffix(path, filepath.Ext(path)) + ".pending"); statError == nil {
			continue
		}
		test.Run(filepath.Base(path), func(test *testing.T) {
			directory := debugReviewProduct(test, path, mode, true)
			buildcache.RequireArtifacts(test, directory, buildcache.Artifact{Name: "program"})
		})
	}
}

func debugLineProduct(test *testing.T, build bool) string {
	test.Helper()
	return debugProduct(test, debugProductInputs(
		"debug-line-table",
		Options{Debug: true, Release: true},
	), build, func(directory string) error {
		compiler := buildcache.GoBuildInMutated(
			test,
			"",
			"adamic-debug-lines",
			"./cmd/adamic",
			nil,
			debugWitnessMutations,
		)
		// Each copy builds a debug and a release program, so the witnesses below can
		// require both to match byte for byte across entry directories.
		for index, copy := range []string{"one", "two"} {
			entry := filepath.Join(directory, copy, "fixture.a")
			if createDirectoryError := os.MkdirAll(filepath.Dir(entry), 0755); createDirectoryError != nil {
				return createDirectoryError
			}
			if writeError := os.WriteFile(entry, []byte(debugLineFixture), 0644); writeError != nil {
				return writeError
			}
			for _, mode := range []struct {
				names []string
				flags []string
			}{
				{[]string{"program", "program-again"}, []string{"--debug"}},
				{[]string{"release", "release-again"}, nil},
			} {
				timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
				arguments := append(
					[]string{"build", entry, "-o", filepath.Join(directory, mode.names[index])},
					mode.flags...,
				)
				command := toolchain.CommandContextIn(timeoutContext, "../..", compiler, arguments...)
				command.Env = debugCacheEnvironment(test, command.Env)
				output, combinedOutputError := command.CombinedOutput()
				cancel()
				if combinedOutputError != nil {
					return fmt.Errorf("build %s: %w: %s", mode.names[index], combinedOutputError, output)
				}
			}
		}
		return nil
	})
}
func TestProduct_DebugLineTable(test *testing.T) { test.Parallel(); debugLineProduct(test, true) }

// Normalize only archive-irrelevant options. The program still receives its own
// absolute comp_dir, split setting and sanitizer fact unchanged.
func debugArchiveProduct(test *testing.T, options Options, build bool) string {
	test.Helper()
	options.DebugDirectory = ""
	options.Split, options.Jobs = false, 0
	if options.Sanitize || options.ThreadSanitize || options.Count {
		options.Release = false
	}
	request, runtimeRequestForError := RuntimeRequestFor(options)
	if runtimeRequestForError != nil {
		test.Fatal(runtimeRequestForError)
	}
	inputs := debugProductInputs(fmt.Sprintf("debug-program-runtime:%+v", options), options)
	inputs.Flags = append(inputs.Flags, request.Portable...)
	directory := debugProduct(
		test,
		inputs,
		build,
		func(directory string) error { return request.Build(directory, 4, nil) },
	)
	return filepath.Join(directory, "runtime.a")
}
func TestProduct_DebugRuntimeRelease(test *testing.T) {
	test.Parallel()
	debugArchiveProduct(test, Options{Release: true}, true)
}
func TestProduct_DebugRuntimeSanitized(test *testing.T) {
	test.Parallel()
	debugArchiveProduct(test, Options{Sanitize: true}, true)
}
func TestProduct_DebugRuntimeDebug(test *testing.T) {
	test.Parallel()
	debugArchiveProduct(test, Options{Debug: true}, true)
}
func TestProduct_DebugRuntimeDebugSanitized(test *testing.T) {
	test.Parallel()
	debugArchiveProduct(test, Options{Debug: true, Sanitize: true}, true)
}

func debugCommandProduct(test *testing.T, build bool) string {
	test.Helper()
	return debugProduct(test, debugProductInputs(
		"debug-command-driver",
		Options{Debug: true},
	), build, func(directory string) error {
		driver := buildcache.GoTest(test, "", "command.test", "./cmd/adamic", nil)
		contents, readError := os.ReadFile(driver)
		if readError != nil {
			return readError
		}
		if writeError := os.WriteFile(filepath.Join(directory, "command.test"), contents, 0755); writeError != nil {
			return writeError
		}
		// Read the runtime dependency in this build's window, so native's test units
		// fetch it as an upstream, rather than merely depending on producer order.
		witnessDriver := buildcache.GoTestMutated(
			test,
			"",
			"command-witness.test",
			"./cmd/adamic",
			nil,
			debugWitnessMutations,
		)
		for index, copy := range []string{"one", "two"} {
			entry := filepath.Join(directory, copy, "driver-witness.a")
			if createDirectoryError := os.MkdirAll(filepath.Dir(entry), 0755); createDirectoryError != nil {
				return createDirectoryError
			}
			if writeError := os.WriteFile(entry, []byte("console.log('driver');\n"), 0644); writeError != nil {
				return writeError
			}
			name := []string{"witness", "witness-again"}[index]
			timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			command := exec.CommandContext(
				timeoutContext,
				witnessDriver,
				"-test.run=^TestBuildDebugDriver$",
				"--",
				"build",
				entry,
				"-o",
				filepath.Join(directory, name),
				"--debug",
			)
			command.Dir = buildcache.Absolute("<repository>/cmd/adamic")
			command.Env = debugCacheEnvironment(test, append(os.Environ(), "ADAMIC_BUILD_DEBUG_DRIVER=1"))
			output, combinedOutputError := command.CombinedOutput()
			cancel()
			if combinedOutputError != nil {
				return fmt.Errorf("command driver: %w: %s", combinedOutputError, output)
			}
		}
		return nil
	})
}
func TestProduct_DebugCommandDriver(test *testing.T) {
	test.Parallel()
	debugCommandProduct(test, true)
}

// This proof intentionally starts with an empty raw runtime cache. Its expensive
// miss belongs to a built-ahead product; the unit rechecks the recorded log.
func debugSharingProduct(test *testing.T, build bool) string {
	test.Helper()
	return debugProduct(test, debugProductInputs(
		"debug-runtime-sharing",
		Options{Debug: true, Release: true},
	), build, func(directory string) error {
		compiler := buildcache.GoBuildInMutated(
			test,
			"",
			"adamic-debug-sharing",
			"./cmd/adamic",
			nil,
			debugWitnessMutations,
		)
		root := directory
		// The cold-cache witness stays under the unit's handed output, never HOME.
		cache := filepath.Join(directory, "raw-cache")
		defer os.RemoveAll(cache)
		for _, copy := range []string{"one", "two"} {
			entry := debugPathFixture(test, filepath.Join(root, copy))
			timeoutContext, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			command := toolchain.CommandContextIn(
				timeoutContext,
				"../..",
				compiler,
				"build",
				entry,
				"-o",
				filepath.Join(filepath.Dir(entry), "program"),
				"--debug",
			)
			command.Env = append(
				command.Env,
				"ADAMIC_BUILD_CACHE_DIR="+cache,
				"ADAMIC_BUILD_OUTPUT_DIR="+cache,
				"ADAMIC_BUILD_LOG="+filepath.Join(directory, "build.log"),
				"ADAMIC_NATIVE_SPLIT=0",
			)
			output, combinedOutputError := command.CombinedOutput()
			cancel()
			if combinedOutputError != nil {
				return fmt.Errorf("runtime sharing build: %w: %s", combinedOutputError, output)
			}
		}
		// Keep the observed archive outcomes and keys, which the consumer checks,
		// without publishing wall-clock durations as part of this product's bytes.
		path := filepath.Join(directory, "build.log")
		contents, readError := os.ReadFile(path)
		if readError != nil {
			return readError
		}
		var outcomes strings.Builder
		for _, line := range strings.Split(string(contents), "\n") {
			fields := strings.Fields(line)
			if len(fields) == 5 && fields[0] == "cArchive" {
				fmt.Fprintf(&outcomes, "%s 0.000\n", strings.Join(fields[:4], " "))
			}
		}
		return os.WriteFile(path, []byte(outcomes.String()), 0644)
	})
}
func TestProduct_DebugRuntimeSharing(test *testing.T) {
	test.Parallel()
	debugSharingProduct(test, true)
}
