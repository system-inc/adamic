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

var debugBuildOptions = []Options{{Release: true}, {Release: true, Debug: true}, {Release: true, Debug: true, Sanitize: true}}

// The compiler is linked into this test binary, so declare its inputs as values
// does. The fixtures and this recipe are under internal too. No assertion is
// cached: products hold executables only, and the consumer observes each run.
func debugProductInputs(name string, options Options) buildcache.Inputs {
	flags := append([]string{fmt.Sprintf("options=%+v", options)}, Flags(options)...)
	for _, variable := range []string{"ADAMIC_PROGRAM_REGION", "ADAMIC_NATIVE_SPLIT", "ADAMIC_NATIVE_JOBS", "CPATH", "C_INCLUDE_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		value, set := os.LookupEnv(variable)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", variable, set, value))
	}
	return buildcache.Inputs{Name: name, Files: append(compilerinputs.Lowering(), "cmd/adamic"), Flags: flags, Toolchain: []string{goruntime.Version(), buildcache.Tool("clang", "--version")}}
}

// Toolchain commands deliberately drop ADAMIC_* variables. Pin both locations:
// Loom can hand a read-only cache and a separate writable output directory.
func debugCacheEnvironment(t *testing.T, environment []string) []string {
	t.Helper()
	if os.Getenv("ADAMIC_BUILD_CACHE_DIR") == "" {
		t.Fatal("debug products require the unit's ADAMIC_BUILD_CACHE_DIR; refusing the home cache (#phzhe68)")
	}
	cache, err := location.BuildCache()
	if err != nil {
		t.Fatal(err)
	}
	output, err := location.Output()
	if err != nil {
		t.Fatal(err)
	}
	return append(environment, "ADAMIC_BUILD_CACHE_DIR="+cache, "ADAMIC_BUILD_OUTPUT_DIR="+output)
}

func debugProduct(t *testing.T, inputs buildcache.Inputs, build bool, recipe func(string) error) string {
	t.Helper()
	debugCacheEnvironment(t, nil)
	if build {
		return buildcache.Product(t, inputs, recipe)
	}
	directory, err := buildcache.Get(inputs, func(string) error {
		return fmt.Errorf("debug product %s not built ahead; run TestProduct_Debug tests first", inputs.Name)
	})
	if err != nil {
		t.Fatal(err)
	}
	return directory
}

func debugReviewProduct(t *testing.T, path string, mode int, build bool) string {
	t.Helper()
	options := debugBuildOptions[mode]
	return debugProduct(t, debugProductInputs(fmt.Sprintf("debug-review-%d:%s", mode, filepath.ToSlash(path)), options), build, func(directory string) error {
		program, err := debugProgram(t, path, options.Debug)
		if err != nil {
			return err
		}
		return Build(CWithOptions(program, options), filepath.Join(directory, "program"), options)
	})
}

func runDebugReviewProducts(t *testing.T, shard, mode int) {
	for _, path := range debugFamilyPaths(reviewPrograms(t), shard, false) {
		if filepath.Dir(path) != "../oracle/testdata/review/agree" {
			continue
		}
		if _, err := os.Stat(strings.TrimSuffix(path, filepath.Ext(path)) + ".pending"); err == nil {
			continue
		}
		t.Run(filepath.Base(path), func(t *testing.T) {
			directory := debugReviewProduct(t, path, mode, true)
			buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "program"})
		})
	}
}

func debugLineProduct(t *testing.T, build bool) string {
	t.Helper()
	return debugProduct(t, debugProductInputs("debug-line-table", Options{Debug: true, Release: true}), build, func(directory string) error {
		compiler := buildcache.GoBuild(t, "adamic", "./cmd/adamic", nil)
		path := filepath.Join(directory, "fixture.a")
		if err := os.WriteFile(path, []byte(debugLineFixture), 0644); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		command := toolchain.CommandContextIn(ctx, "../..", compiler, "build", path, "-o", filepath.Join(directory, "program"), "--debug")
		command.Env = debugCacheEnvironment(t, command.Env)
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("build --debug: %w: %s", err, output)
		}
		return nil
	})
}
func TestProduct_DebugLineTable(t *testing.T) { t.Parallel(); debugLineProduct(t, true) }

// Normalize only archive-irrelevant options. The program still receives its own
// absolute comp_dir, split setting and sanitizer fact unchanged.
func debugArchiveProduct(t *testing.T, options Options, build bool) string {
	t.Helper()
	options.DebugDirectory = ""
	options.Split, options.Jobs = false, 0
	if options.Sanitize || options.ThreadSanitize || options.Count {
		options.Release = false
	}
	request, err := RuntimeRequestFor(options)
	if err != nil {
		t.Fatal(err)
	}
	inputs := debugProductInputs(fmt.Sprintf("debug-program-runtime:%+v", options), options)
	inputs.Flags = append(inputs.Flags, request.Portable...)
	directory := debugProduct(t, inputs, build, func(directory string) error { return request.Build(directory, 4, nil) })
	return filepath.Join(directory, "runtime.a")
}
func TestProduct_DebugRuntimeRelease(t *testing.T) {
	t.Parallel()
	debugArchiveProduct(t, Options{Release: true}, true)
}
func TestProduct_DebugRuntimeSanitized(t *testing.T) {
	t.Parallel()
	debugArchiveProduct(t, Options{Sanitize: true}, true)
}
func TestProduct_DebugRuntimeDebug(t *testing.T) {
	t.Parallel()
	debugArchiveProduct(t, Options{Debug: true}, true)
}
func TestProduct_DebugRuntimeDebugSanitized(t *testing.T) {
	t.Parallel()
	debugArchiveProduct(t, Options{Debug: true, Sanitize: true}, true)
}

func debugCommandProduct(t *testing.T, build bool) string {
	t.Helper()
	return debugProduct(t, debugProductInputs("debug-command-driver", Options{Debug: true}), build, func(directory string) error {
		driver := buildcache.GoTest(t, "", "command.test", "./cmd/adamic", nil)
		contents, err := os.ReadFile(driver)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "command.test"), contents, 0755); err != nil {
			return err
		}
		// Read the runtime dependency in this build's window, so native's test units
		// fetch it as an upstream, rather than merely depending on producer order.
		entry := filepath.Join(directory, "driver-witness.a")
		if err := os.WriteFile(entry, []byte("console.log('driver');\n"), 0644); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
		defer cancel()
		command := exec.CommandContext(ctx, driver, "-test.run=^TestBuildDebugDriver$", "--", "build", entry, "-o", filepath.Join(directory, "witness"), "--debug")
		command.Dir = buildcache.Absolute("<repository>/cmd/adamic")
		command.Env = debugCacheEnvironment(t, append(os.Environ(), "ADAMIC_BUILD_DEBUG_DRIVER=1"))
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("command driver: %w: %s", err, output)
		}
		return nil
	})
}
func TestProduct_DebugCommandDriver(t *testing.T) { t.Parallel(); debugCommandProduct(t, true) }

// This proof intentionally starts with an empty raw runtime cache. Its expensive
// miss belongs to a built-ahead product; the unit rechecks the recorded log.
func debugSharingProduct(t *testing.T, build bool) string {
	t.Helper()
	return debugProduct(t, debugProductInputs("debug-runtime-sharing", Options{Debug: true, Release: true}), build, func(directory string) error {
		compiler := buildcache.Adamic(t)
		root := t.TempDir()
		// The cold-cache witness stays under the unit's handed output, never HOME.
		cache := filepath.Join(directory, "raw-cache")
		defer os.RemoveAll(cache)
		for _, copy := range []string{"one", "two"} {
			entry := debugPathFixture(t, filepath.Join(root, copy))
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
			command := toolchain.CommandContextIn(ctx, "../..", compiler, "build", entry, "-o", filepath.Join(filepath.Dir(entry), "program"), "--debug")
			command.Env = append(command.Env, "ADAMIC_BUILD_CACHE_DIR="+cache, "ADAMIC_BUILD_OUTPUT_DIR="+cache, "ADAMIC_BUILD_LOG="+filepath.Join(directory, "build.log"), "ADAMIC_NATIVE_SPLIT=0")
			output, err := command.CombinedOutput()
			cancel()
			if err != nil {
				return fmt.Errorf("runtime sharing build: %w: %s", err, output)
			}
		}
		return nil
	})
}
func TestProduct_DebugRuntimeSharing(t *testing.T) { t.Parallel(); debugSharingProduct(t, true) }
