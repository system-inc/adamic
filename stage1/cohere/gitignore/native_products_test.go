package gitignore

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

// Every sanitized build TestEachGapStandsWhereGapsMdSaysItDoes and TestPathAnswersAsGosPathPackage run is a keyed
// product built ahead by a TestProduct_ of its own, so neither test compiles C (#tze35sy, Loom's build law): seventeen
// clang builds, ten closed gaps and the path sweep with each of its six mutants, were made inside the tests.

func TestProduct_GitignoreGapsSanitized(t *testing.T) {
	t.Parallel()
	for _, gap := range gaps {
		if !gap.open {
			requirePort(t, gapSanitized(t, gap.path))
		}
	}
}

func TestProduct_GitignorePathSweepsSanitized(t *testing.T) {
	t.Parallel()
	requirePort(t, pathSweepSanitized(t, pathMutant{}))
	for _, mutant := range pathMutants {
		requirePort(t, pathSweepSanitized(t, mutant))
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
	inputs := sanitizedInputs("gitignore gap "+path+" native sanitized", "stage1/cohere/gitignore/"+path)
	return buildcache.Product(t, inputs, func(directory string) error {
		entry, err := filepath.Abs(path)
		if err != nil {
			return err
		}
		return buildSanitized(entry, directory)
	})
}

// pathSweepSanitized is pathsweep.ts over path.ts with mutant applied (none when mutant.from is empty), lowered in
// this process and built by clang under the sanitizers, keyed by the compiler's paths, both sources, the mutation and
// the native build. A mutation that doesn't change exactly one place in path.ts is refused.
func pathSweepSanitized(t *testing.T, mutant pathMutant) string {
	t.Helper()
	name := "gitignore path sweep native sanitized"
	if mutant.from != "" {
		name = "gitignore path sweep mutant " + mutant.name + " native sanitized"
	}
	inputs := sanitizedInputs(name, "stage1/cohere/gitignore/path.ts", "stage1/cohere/gitignore/pathsweep.ts")
	inputs.Flags = append(inputs.Flags, fmt.Sprintf("mutation=path.ts:%q=>%q", mutant.from, mutant.to))
	return buildcache.Product(t, inputs, func(directory string) error {
		source := filepath.Join(directory, "source")
		if err := os.Mkdir(source, 0o755); err != nil {
			return err
		}
		for _, name := range []string{"path.ts", "pathsweep.ts"} {
			contents, err := os.ReadFile(name)
			if err != nil {
				return err
			}
			text := string(contents)
			if mutant.from != "" && name == "path.ts" {
				if strings.Count(text, mutant.from) != 1 {
					return fmt.Errorf("the mutant must change exactly one place in path.ts: %q", mutant.from)
				}
				text = strings.Replace(text, mutant.from, mutant.to, 1)
			}
			if err := os.WriteFile(filepath.Join(source, name), []byte(text), 0o644); err != nil {
				return err
			}
		}
		return buildSanitized(filepath.Join(source, "pathsweep.ts"), directory)
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
		Files:     append(files, "stage1/cohere/gitignore/native_products_test.go"),
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
