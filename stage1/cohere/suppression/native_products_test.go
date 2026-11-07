package suppression

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// Every sanitized build TestEachGapStandsWhereGapsMdSaysItDoes and TestThePortParsesAsGoCohereDoes run is a keyed
// product built ahead by a TestProduct_ of its own, so neither test compiles C (#tze35sy, Loom's build law): fifteen
// clang builds, two closed gaps, the port and the port with each of its twelve mutants, were made inside the tests.

func TestProduct_SuppressionGapsSanitized(test *testing.T) {
	test.Parallel()
	for _, gap := range gaps {
		if gap.notYet == "" && gap.badC == "" {
			requirePort(test, gapSanitized(test, gap.path))
		}
	}
}

func TestProduct_SuppressionPortSanitized(test *testing.T) {
	test.Parallel()
	requirePort(test, portSanitized(test, nil))
}

func TestProduct_SuppressionMutantsSanitized(test *testing.T) {
	test.Parallel()
	for _, mutant := range mutants {
		requirePort(test, portSanitized(test, &mutant))
	}
}

func requirePort(test *testing.T, directory string) {
	test.Helper()
	buildcache.RequireArtifacts(test, directory, buildcache.Artifact{Name: "port"})
}

// gapSanitized is the closed gap at path, relative to this package, lowered in this process and built by clang under
// the sanitizers, keyed by the compiler's paths, the gap's source and the native build.
func gapSanitized(test *testing.T, path string) string {
	test.Helper()
	inputs := sanitizedInputs("suppression gap "+path+" native sanitized", "stage1/cohere/suppression/"+path)
	return buildcache.Product(test, inputs, func(directory string) error {
		entry, absolutePathError := filepath.Abs(path)
		if absolutePathError != nil {
			return absolutePathError
		}
		return buildSanitized(entry, directory)
	})
}

// portSanitized is the port with applied in it (none when applied is nil), lowered in this process and built by clang
// under the sanitizers, keyed by the compiler's paths, the port's files, the mutation and the native build. A mutant
// that doesn't change exactly one place in its file is refused.
func portSanitized(test *testing.T, applied *mutant) string {
	test.Helper()
	name := "suppression port native sanitized"
	if applied != nil {
		name = "suppression port mutant " + applied.name + " native sanitized"
	}
	var sources []string
	for _, file := range portFiles {
		sources = append(sources, "stage1/cohere/suppression/"+file)
	}
	inputs := sanitizedInputs(name, sources...)
	if applied != nil {
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("mutation=%s:%q=>%q", applied.file, applied.from, applied.to))
	}
	return buildcache.Product(test, inputs, func(directory string) error {
		source := filepath.Join(directory, "source")
		if createDirectoryError := os.Mkdir(source, 0o755); createDirectoryError != nil {
			return createDirectoryError
		}
		for _, file := range portFiles {
			contents, readError := os.ReadFile(file)
			if readError != nil {
				return readError
			}
			text := string(contents)
			if applied != nil && applied.file == file {
				if strings.Count(text, applied.from) != 1 {
					return fmt.Errorf("the mutant %q must change exactly one place in %s", applied.name, file)
				}
				text = strings.Replace(text, applied.from, applied.to, 1)
			}
			if writeError := os.WriteFile(filepath.Join(source, file), []byte(text), 0o644); writeError != nil {
				return writeError
			}
		}
		return buildSanitized(filepath.Join(source, "main.ts"), directory)
	})
}

// sanitizedInputs keys a sanitized build lowered in this process: the compiler is linked into this test binary, where
// no trace sees it read, so its paths are named (compilerinputs.Lowering) beside the sources and this file, with the
// native options and flags, the environment native.Build and clang read, and clang's version.
func sanitizedInputs(name string, sources ...string) buildcache.Inputs {
	options := native.Options{Sanitize: true}
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
		Files:     append(files, "stage1/cohere/suppression/native_products_test.go"),
		Flags:     flags,
		Toolchain: []string{runtime.Version(), buildcache.Tool("clang", "--version")},
	}
}

// buildSanitized lowers the program at entry and builds it under the sanitizers as directory/port.
func buildSanitized(entry, directory string) error {
	program, loadError := load.Load([]string{entry})
	if loadError != nil {
		return fmt.Errorf("Load: %w", loadError)
	}
	lowered, lowerError := lower.Lower(context.Background(), program)
	if lowerError != nil {
		return fmt.Errorf("Lower: %w", lowerError)
	}
	return native.Build(native.C(lowered), filepath.Join(directory, "port"), native.Options{Sanitize: true})
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
