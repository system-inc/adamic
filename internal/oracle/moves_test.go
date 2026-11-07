package oracle

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"slices"
	"time"

	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/childguard"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"github.com/system-inc/adamic/internal/javascript"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

// objects.a has quadratic lowering and large native builds. Its dedicated products
// and TestMovesAgreesWithNode variants hold it to source and backend Node; the
// determinism product compares two independent compiler processes. Keep it out of
// every generic fixture family, including native, WASI, counts and determinism.
func init() {
	fixtures = slices.DeleteFunc(fixtures, func(fixture struct {
		path    string
		lowers  bool
		checked bool
	}) bool {
		return fixture.path == moveObjectsPath
	})
}

// Each build is an independent unit; both thread modes use its one binary.
func TestMovesAgreesWithNode_asan(test *testing.T) {
	test.Parallel()

	moveNativeVariant(test, "asan")
}

func TestMovesAgreesWithNode_release(test *testing.T) {
	test.Parallel()

	moveNativeVariant(test, "release")
}

func TestMovesAgreesWithNode_asan_slabs(test *testing.T) {
	test.Parallel()

	moveNativeVariant(test, "asan_slabs")
}

func TestMovesAgreesWithNode_tsan(test *testing.T) {
	test.Parallel()

	if runtime.GOOS != "linux" {
		test.Skip("platform-specific allocator or sanitizer")
	}
	moveNativeVariant(test, "tsan")
}

func TestMovesRefusals(test *testing.T) {
	test.Parallel()
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/moves/refused/*.a"))
	if err != nil || len(paths) < 30 {
		test.Fatalf("fixtures: %d %v", len(paths), err)
	}
	for _, path := range paths {
		test.Run(filepath.Base(path), func(test *testing.T) {
			test.Parallel()
			want, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".what")
			if err != nil {
				test.Fatal(err)
			}
			fix, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".fix")
			if err != nil {
				test.Fatal(err)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				test.Fatal(err)
			}
			_, err = lowered(test, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				test.Fatalf("want Refused %q, got %v", strings.TrimSpace(string(want)), err)
			}
			if refused.Fix != "return it through the results" && refused.Fix != "don't use it after the parallelMap" {
				test.Fatalf("unapproved move fix: %q", refused.Fix)
			}
			if refused.What != strings.TrimSpace(string(want)) ||
				refused.Where == "" ||
				refused.Fix != strings.TrimSpace(string(fix)) {
				test.Fatalf("want %q; got %+v", strings.TrimSpace(string(want)), refused)
			}
		})
	}
}

// Ordinary retain totals cannot distinguish plain from atomic count traffic.
// Count actual graph headers at join, independently of the lowering's move flag.
func TestMovesPlainCounts_Control(test *testing.T) { test.Parallel(); movesPlainCounts(test, false) }
func TestMovesPlainCounts_SharingMutant(test *testing.T) {
	test.Parallel()
	movesPlainCounts(test, true)
}

func moveCountedBinary(test *testing.T, mutant bool) string {
	test.Helper()
	variant, abi := "control", "adamic_parallel_map_move"
	if mutant {
		variant, abi = "shared-mutant", "adamic_parallel_map"
	}
	inputs := moveSourceInputs()
	inputs.Name = "moves-counted-" + variant
	inputs.Flags = native.Flags(native.Options{Count: true})
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	directory := buildcache.Product(test, inputs, func(directory string) error {
		generated := moveSourceFile(test, "program.c")
		if !strings.Contains(generated, "adamic_parallel_map_move(") {
			return fmt.Errorf("no move ABI")
		}
		source := fmt.Sprintf(`#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
adamic_array *%s(adamic_array *, adamic_closure *, enum adamic_ir_type);
adamic_array *moves_count_map(adamic_array *items, adamic_closure *work, enum adamic_ir_type element) {
 adamic_array *results = %s(items, work, element);
 if (results == NULL) { abort(); }
 size_t shared = adamic_is_shared(&items->heap) + adamic_is_shared(&results->heap);
 for (size_t index = 0; index < items->length; index++) {
  shared += adamic_is_shared(items->elements[index].reference);
 }
 fprintf(stderr, "move graph counts: plain %%zu shared %%zu\n", items->length + 2 - shared, shared);
 return results;
}
`, abi, abi) + strings.ReplaceAll(generated, "adamic_parallel_map_move(", "moves_count_map(")
		binary := filepath.Join(directory, "program")
		if err := native.Build(source, binary, native.Options{Count: true}); err != nil {
			return err
		}
		return nil
	})
	return filepath.Join(directory, "program")
}

func TestProduct_MovesCounted_Control(test *testing.T) {
	test.Parallel()
	moveCountedBinary(test, false)
}
func TestProduct_MovesCounted_SharingMutant(test *testing.T) {
	test.Parallel()
	moveCountedBinary(test, true)
}

func movesPlainCounts(test *testing.T, mutant bool) {
	test.Helper()
	binary := moveCountedBinary(test, mutant)
	want := "move graph counts: plain 2050 shared 0\n"
	if mutant {
		want = "move graph counts: plain 1 shared 2049\n"
	}
	for _, threads := range []string{"1", ""} {
		observed := executeParallel(test, threads, false, binary)
		if observed.exitCode != 0 || !strings.HasPrefix(string(observed.stderr), want) {
			test.Fatalf("want %q: exit %d stderr %s", want, observed.exitCode, observed.stderr)
		}
		test.Logf("threads=%q %s", threads, observed.stderr)
	}
}

// A small real allocation keeps the ownership cleanup mutant independent of compile time.
func TestMovesCleanupAgreesWithNode(test *testing.T) {
	test.Parallel()

	parallelFixtureUnit(test, "internal/oracle/testdata/moves/accepted/cleanup.a", "asan", "all", -1)
}

const moveObjectsPath = "internal/oracle/testdata/moves/accepted/objects.a"

func moveSourceInputs() buildcache.Inputs {
	return buildcache.Inputs{
		Name:      "moves-source",
		Files:     append(compilerinputs.Lowering(), moveObjectsPath),
		Toolchain: []string{runtime.Version(), runtime.GOOS, runtime.GOARCH},
	}
}

// Generation and native compilation are separate independent units. Neither drops a program.
func moveSource(test *testing.T) string {
	test.Helper()
	return buildcache.Product(test, moveSourceInputs(), func(directory string) error {
		path, err := filepath.Abs(filepath.Join(repository, moveObjectsPath))
		if err != nil {
			return err
		}
		program, err := lowered(test, path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "program.c"), []byte(native.C(program)), 0600); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "program.mjs"), []byte(javascript.JavaScript(program)), 0600)
	})
}

func moveSourceFile(test *testing.T, name string) string {
	test.Helper()
	bytes, err := os.ReadFile(filepath.Join(moveSource(test), name))
	if err != nil {
		test.Fatal(err)
	}
	return string(bytes)
}

func TestProduct_MovesSource(test *testing.T) {
	test.Parallel()
	moveSource(test)
}

func TestMovesSourceKeyNamesCompiler(test *testing.T) {
	test.Parallel()
	if err := compilerinputs.CheckKey([]string{moveObjectsPath}, func(root string) (string, error) {
		return buildcache.Key(root, moveSourceInputs())
	}); err != nil {
		test.Fatal(err)
	}
}

func moveBuildOptions(build string) (native.Options, string) {
	switch build {
	case "asan":
		return native.Options{Sanitize: true, Split: true, Jobs: 4}, ""
	case "asan_slabs":
		return native.Options{Sanitize: true}, "ADAMIC_SLABS"
	case "tsan":
		return native.Options{ThreadSanitize: true}, ""
	case "malloc":
		return native.Options{}, "ADAMIC_MALLOC"
	default:
		return native.Options{}, ""
	}
}

// Each native build is its own preparation unit, after the source product.
func moveBinary(test *testing.T, build string) string {
	test.Helper()
	options, allocator := moveBuildOptions(build)
	inputs := moveSourceInputs()
	inputs.Name = "moves-native-" + build
	inputs.Flags = append(native.Flags(options), "allocator="+allocator)
	if options.Split {
		inputs.Flags = append(inputs.Flags, fmt.Sprintf("split=true jobs=%d", options.Jobs))
	}
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	directory := buildcache.Product(test, inputs, func(directory string) error {
		buildParallelProgram(
			test,
			moveSourceFile(test, "program.c"),
			filepath.Join(directory, "program"),
			options,
			allocator,
		)
		return nil
	})
	return filepath.Join(directory, "program")
}

func TestProduct_MovesNative_asan(test *testing.T)    { test.Parallel(); moveBinary(test, "asan") }
func TestProduct_MovesNative_release(test *testing.T) { test.Parallel(); moveBinary(test, "release") }
func TestProduct_MovesNative_asan_slabs(test *testing.T) {
	test.Parallel()
	moveBinary(test, "asan_slabs")
}
func TestProduct_MovesNative_tsan(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("Linux TSan")
	}
	moveBinary(test, "tsan")
}
func moveNativeVariant(test *testing.T, build string) {
	test.Helper()
	directory := moveSource(test)
	path, err := filepath.Abs(filepath.Join(repository, moveObjectsPath))
	if err != nil {
		test.Fatal(err)
	}
	source := onNode(test, path)
	if source.exitCode != 0 {
		test.Fatalf("Node: %+v", source)
	}
	if difference := disagreement(source, onNode(test, filepath.Join(directory, "program.mjs"))); difference != "" {
		test.Fatal(difference)
	}
	binary := moveBinary(test, build)
	options, _ := moveBuildOptions(build)
	for _, threads := range []string{"1", ""} {
		repeats := 1
		if options.ThreadSanitize {
			repeats = 3
		}
		for attempt := 0; attempt < repeats; attempt++ {
			observed := executeParallel(test, threads, false, binary)
			if difference := disagreement(source, observed); difference != "" {
				test.Fatalf("%s: exit %d stderr %s", difference, observed.exitCode, observed.stderr)
			}
		}
		if runtime.GOOS == "linux" && options.Sanitize {
			observed := executeParallel(test, threads, true, binary)
			if difference := disagreement(source, observed); difference != "" {
				test.Fatalf("leak check: %s: exit %d stderr %s", difference, observed.exitCode, observed.stderr)
			}
		}
		if runtime.GOOS == "darwin" && build == "malloc" {
			observed := executeParallel(test, threads, false, "leaks", "--atExit", "--", binary)
			if observed.exitCode != 0 {
				test.Fatalf("leaks: %s %s", observed.stdout, observed.stderr)
			}
		}
	}
}

func moveDeterminism(test *testing.T) string {
	test.Helper()
	inputs := moveSourceInputs()
	inputs.Name = "moves-determinism"
	return buildcache.Product(test, inputs, func(directory string) error {
		binary := buildcache.AdamicBuiltElsewhere(test)
		// Independent processes need no ordering. Run both on this product's CPU
		// share so two quadratic lowerings do not consume two leaf budgets.
		emissions := make(chan emitted, 2)
		killed := make(chan bool, 2)
		ctx, cancel := moveEmissionContext(test)
		defer cancel()
		for range 2 {
			go func() {
				result, deadline := emitWithin(ctx, binary, moveObjectsPath)
				emissions <- result
				killed <- deadline
			}()
		}
		first, second := <-emissions, <-emissions
		if firstKilled, secondKilled := <-killed, <-killed; firstKilled || secondKilled {
			return fmt.Errorf(
				"stage 0 was killed at the product's deadline before it finished emitting %s (exit %d and %d): a slow box, never a difference",
				moveObjectsPath,
				first.code,
				second.code,
			)
		}
		if first.code != 0 || second.code != 0 || !bytes.Equal(first.output, second.output) {
			return fmt.Errorf(
				"stage 0 emitted different C: exit %d and %d, first difference at line %d",
				first.code,
				second.code,
				firstDifferingLine(first.output, second.output),
			)
		}
		return os.WriteFile(filepath.Join(directory, "program.c"), first.output, 0600)
	})
}

// The emissions are bounded by the product's own share (its test deadline, less time to report),
// not emit's wall-clock minute: two quadratic lowerings of objects.a take about 40 s together
// untraced, so a traced or loaded box can pass a minute without being wrong (#d0bzeyh).
func moveEmissionContext(test *testing.T) (context.Context, context.CancelFunc) {
	if deadline, ok := test.Deadline(); ok {
		return context.WithDeadline(context.Background(), deadline.Add(-10*time.Second))
	}
	return context.WithTimeout(context.Background(), 10*time.Minute)
}

// emitWithin is emit under ctx, and says whether ctx's deadline killed the process group,
// so the caller reports a kill as a kill. childguard.Capture runs it in a group of its own and
// kills the group when ctx is done.
func emitWithin(ctx context.Context, binary, path string) (emitted, bool) {
	command := exec.CommandContext(ctx, binary, "c", path)
	command.Dir = repository
	output, _, err := childguard.Capture(ctx, command)
	if ctx.Err() != nil {
		return emitted{code: -1}, true
	}
	var exit *exec.ExitError
	if err != nil && !errors.As(err, &exit) {
		return emitted{code: -1}, false
	}
	return emitted{code: command.ProcessState.ExitCode(), output: output}, false
}

func TestProduct_MovesDeterminism(test *testing.T)    { test.Parallel(); moveDeterminism(test) }
func TestMovesEmitsDeterministically(test *testing.T) { test.Parallel(); moveDeterminism(test) }

// A new generic family must not silently put the heavy lowering back in a leaf.
func TestMovesObjectsUseDedicatedProducts(test *testing.T) {
	test.Parallel()
	for _, keys := range [][]string{fixturePaths(), determinismKeys(), countKeys()} {
		if slices.Contains(keys, moveObjectsPath) {
			test.Fatal("objects.a must use its dedicated moves products")
		}
	}
}
