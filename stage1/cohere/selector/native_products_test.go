package selector

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

// Every sanitized build TestEachGapStandsWhereGapsMdSaysItDoes, TestThePortParsesAsGoCohereDoes and the comparison's
// mutant tests run is a keyed product built ahead by a TestProduct_ of its own, so none of them compiles C (#tze35sy,
// Loom's build law): five clang builds, the closed gap, the port and its three mutants, were made inside the tests.

func TestProduct_SelectorGapsSanitized(test *testing.T) {
	test.Parallel()
	for _, gap := range gaps {
		if gap.notYet == "" && gap.badC == "" {
			requirePort(test, gapSanitized(test, gap.path))
		}
	}
}

func TestProduct_SelectorPortSanitized(test *testing.T) {
	test.Parallel()
	requirePort(test, portSanitized(test, nil))
}

func TestProduct_SelectorPortMutantsSanitized(test *testing.T) {
	test.Parallel()
	for _, applied := range comparisonMutants {
		requirePort(test, portSanitized(test, &applied))
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
	inputs := sanitizedInputs("selector gap "+path+" native sanitized", "stage1/cohere/selector/"+path)
	return buildcache.Product(test, inputs, func(directory string) error {
		entry, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		return buildSanitized(entry, directory)
	})
}

// portSanitized is the port and its driver with applied made (none when applied is nil), lowered in this process and
// built by clang under the sanitizers, keyed by the compiler's paths, the port's sources, the mutation and the native
// build. A mutant that doesn't change exactly one place in its file is refused.
func portSanitized(test *testing.T, applied *mutant) string {
	test.Helper()
	name := "selector port native sanitized"
	mutation := "mutation=none"
	if applied != nil {
		name = "selector port mutant " + applied.name + " native sanitized"
		mutation = fmt.Sprintf("mutation=%s:%q=>%q", applied.file, applied.from, applied.to)
	}
	var sources []string
	for _, file := range portFiles {
		sources = append(sources, "stage1/cohere/selector/"+file)
	}
	inputs := sanitizedInputs(name, sources...)
	inputs.Flags = append(inputs.Flags, mutation)
	return buildcache.Product(test, inputs, func(directory string) error {
		source := filepath.Join(directory, "source")
		if err := os.Mkdir(source, 0o755); err != nil {
			return err
		}
		for _, file := range portFiles {
			contents, err := os.ReadFile(file)
			if err != nil {
				return err
			}
			text := string(contents)
			if applied != nil && applied.file == file {
				if strings.Count(text, applied.from) != 1 {
					return fmt.Errorf("the mutant %q must change exactly one place in %s", applied.name, file)
				}
				text = strings.Replace(text, applied.from, applied.to, 1)
			}
			if err := os.WriteFile(filepath.Join(source, file), []byte(text), 0o644); err != nil {
				return err
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
		Files:     append(files, "stage1/cohere/selector/native_products_test.go"),
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
func sanitizedRun(test *testing.T, binary string, arguments ...string) run {
	test.Helper()
	var environment []string
	if runtime.GOOS == "linux" {
		environment = []string{"ASAN_OPTIONS=detect_leaks=0"}
	}
	return execute(test, environment, binary, arguments...)
}
