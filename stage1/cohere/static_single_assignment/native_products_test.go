package static_single_assignment

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

// Every sanitized build TestEachGapStandsWhereGapsMdSaysItDoes and TestThePortMakesWhatGoCohereMakes run is a keyed
// product built ahead by a TestProduct_ of its own, so neither test compiles C (#tze35sy, Loom's build law): thirteen
// clang builds, the port and the port with each of its twelve mutants, were made inside the tests, and a gap that
// closes would have been a fourteenth.

func TestProduct_StaticSingleAssignmentGapsSanitized(t *testing.T) {
	t.Parallel()
	for _, gap := range gaps {
		if gap.refused == "" && gap.notYet == "" {
			requirePort(t, gapSanitized(t, gap.path))
		}
	}
}

func TestProduct_StaticSingleAssignmentPortSanitized(t *testing.T) {
	t.Parallel()
	requirePort(t, portSanitized(t, nil))
}

func TestProduct_StaticSingleAssignmentMutantsSanitized(t *testing.T) {
	t.Parallel()
	for _, mutant := range mutants {
		requirePort(t, portSanitized(t, &mutant))
	}
}

func requirePort(t *testing.T, directory string) {
	t.Helper()
	buildcache.RequireArtifacts(t, directory, buildcache.Artifact{Name: "port"})
}

// gapSanitized is the closed gap at path, relative to this package, lowered in this process and built by clang under
// the sanitizers, keyed by the compiler's paths, the gap's source and the native build.
func gapSanitized(t *testing.T, path string) string {
	t.Helper()
	inputs := sanitizedInputs("static_single_assignment gap "+path+" native sanitized", "stage1/cohere/static_single_assignment/"+path)
	return buildcache.Product(t, inputs, func(directory string) error {
		entry, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		return buildSanitized(entry, directory)
	})
}

// portSanitized is the port with applied in it (none when applied is nil), lowered in this process and built by clang
// under the sanitizers, keyed by the compiler's paths, the port's files, the mutation and the native build. A mutant
// that doesn't change exactly one place in its file is refused.
func portSanitized(t *testing.T, applied *mutant) string {
	t.Helper()
	name := "static_single_assignment port native sanitized"
	if applied != nil {
		name = "static_single_assignment port mutant " + applied.name + " native sanitized"
	}
	var sources []string
	for _, file := range portFiles {
		sources = append(sources, "stage1/cohere/static_single_assignment/"+file)
	}
	inputs := sanitizedInputs(name, sources...)
	if applied != nil {
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("mutation=%s:%q=>%q", applied.file, applied.from, applied.to))
	}
	return buildcache.Product(t, inputs, func(directory string) error {
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
	flags := append([]string{"load.Load", "lower.Lower", "native.C", fmt.Sprintf("options=%+v", options)}, native.Flags(options)...)
	for _, variable := range []string{"ADAMIC_NATIVE_SPLIT", "CPATH", "C_INCLUDE_PATH", "LIBRARY_PATH", "SDKROOT", "MACOSX_DEPLOYMENT_TARGET"} {
		value, set := os.LookupEnv(variable)
		flags = append(flags, fmt.Sprintf("env:%s:set=%t:value=%s", variable, set, value))
	}
	files := append(compilerinputs.Lowering(), sources...)
	return buildcache.Inputs{
		Name:      name,
		Files:     append(files, "stage1/cohere/static_single_assignment/native_products_test.go"),
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
