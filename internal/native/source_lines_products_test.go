package native

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	goruntime "runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
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

func debugProduct(t *testing.T, inputs buildcache.Inputs, build bool, recipe func(string) error) string {
	t.Helper()
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
		if output, err := command.CombinedOutput(); err != nil {
			return fmt.Errorf("build --debug: %w: %s", err, output)
		}
		return nil
	})
}
func TestProduct_DebugLineTable(t *testing.T) { t.Parallel(); debugLineProduct(t, true) }

func debugDirectoriesProduct(t *testing.T, build bool) string {
	t.Helper()
	return debugProduct(t, debugProductInputs("debug-two-directories", Options{Debug: true}), build, func(directory string) error {
		for _, copy := range []string{"one", "two"} {
			root := filepath.Join(directory, copy)
			if err := os.MkdirAll(filepath.Join(root, "lib"), 0755); err != nil {
				return err
			}
			path := filepath.Join(root, "fixture.a")
			if err := os.WriteFile(filepath.Join(root, "lib/value.a"), []byte("export function f(n: number): number { return n + 1; }\n"), 0644); err != nil {
				return err
			}
			if err := os.WriteFile(path, []byte("import { f } from './lib/value.a';\nconsole.log((f(4)).toString());\n"), 0644); err != nil {
				return err
			}
			program, err := debugProgram(t, path, true)
			if err != nil {
				return err
			}
			if err := Build(CWithOptions(program, Options{Debug: true}), filepath.Join(root, "program"), Options{Debug: true}); err != nil {
				return err
			}
		}
		return nil
	})
}
func TestProduct_DebugTwoDirectories(t *testing.T) { t.Parallel(); debugDirectoriesProduct(t, true) }
