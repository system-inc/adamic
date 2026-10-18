package oracle

import (
	"bytes"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"slices"

	"github.com/system-inc/adamic/internal/buildcache"
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
func TestMovesAgreesWithNode_asan(t *testing.T) {
	t.Parallel()

	moveNativeVariant(t, "asan")
}

func TestMovesAgreesWithNode_release(t *testing.T) {
	t.Parallel()

	moveNativeVariant(t, "release")
}

func TestMovesAgreesWithNode_asan_slabs(t *testing.T) {
	t.Parallel()

	moveNativeVariant(t, "asan_slabs")
}

func TestMovesAgreesWithNode_tsan(t *testing.T) {
	t.Parallel()

	if runtime.GOOS != "linux" {
		t.Skip("platform-specific allocator or sanitizer")
	}
	moveNativeVariant(t, "tsan")
}

func TestMovesRefusals(t *testing.T) {
	t.Parallel()
	paths, err := filepath.Glob(filepath.Join(repository, "internal/oracle/testdata/moves/refused/*.a"))
	if err != nil || len(paths) < 30 {
		t.Fatalf("fixtures: %d %v", len(paths), err)
	}
	for _, path := range paths {
		t.Run(filepath.Base(path), func(t *testing.T) {
			t.Parallel()
			want, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".what")
			if err != nil {
				t.Fatal(err)
			}
			fix, err := os.ReadFile(strings.TrimSuffix(path, ".a") + ".fix")
			if err != nil {
				t.Fatal(err)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			_, err = lowered(t, absolute)
			var refused *lower.Refused
			if !errors.As(err, &refused) {
				t.Fatalf("want Refused %q, got %v", strings.TrimSpace(string(want)), err)
			}
			if refused.Fix != "return it through the results" && refused.Fix != "don't use it after the parallelMap" {
				t.Fatalf("unapproved move fix: %q", refused.Fix)
			}
			if refused.What != strings.TrimSpace(string(want)) || refused.Where == "" || refused.Fix != strings.TrimSpace(string(fix)) {
				t.Fatalf("want %q; got %+v", strings.TrimSpace(string(want)), refused)
			}
		})
	}
}

// Ordinary retain totals cannot distinguish plain from atomic count traffic.
// Count actual graph headers at join, independently of the lowering's move flag.
func TestMovesPlainCounts_Control(t *testing.T) { t.Parallel(); movesPlainCounts(t, false) }
func TestMovesPlainCounts_SharingMutant(t *testing.T) {
	t.Parallel()
	movesPlainCounts(t, true)
}

func moveCountedBinary(t *testing.T, mutant bool) string {
	t.Helper()
	variant, abi := "control", "adamic_parallel_map_move"
	if mutant {
		variant, abi = "shared-mutant", "adamic_parallel_map"
	}
	inputs := moveSourceInputs()
	inputs.Name = "moves-counted-" + variant
	inputs.Flags = native.Flags(native.Options{Count: true})
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	directory := buildcache.Product(t, inputs, func(directory string) error {
		generated := moveSourceFile(t, "program.c")
		if !strings.Contains(generated, "adamic_parallel_map_move(") {
			return fmt.Errorf("no move ABI")
		}
		source := fmt.Sprintf(`#include "adamic.h"
#include <stdio.h>
#include <stdlib.h>
adamic_array *%s(adamic_array *, adamic_closure *, bool);
adamic_array *moves_count_map(adamic_array *items, adamic_closure *work, bool references) {
 adamic_array *results = %s(items, work, references);
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

func TestProduct_MovesCounted_Control(t *testing.T)       { t.Parallel(); moveCountedBinary(t, false) }
func TestProduct_MovesCounted_SharingMutant(t *testing.T) { t.Parallel(); moveCountedBinary(t, true) }

func movesPlainCounts(t *testing.T, mutant bool) {
	t.Helper()
	binary := moveCountedBinary(t, mutant)
	want := "move graph counts: plain 2050 shared 0\n"
	if mutant {
		want = "move graph counts: plain 1 shared 2049\n"
	}
	for _, threads := range []string{"1", ""} {
		observed := executeParallel(t, threads, false, binary)
		if observed.exitCode != 0 || !strings.HasPrefix(string(observed.stderr), want) {
			t.Fatalf("want %q: exit %d stderr %s", want, observed.exitCode, observed.stderr)
		}
		t.Logf("threads=%q %s", threads, observed.stderr)
	}
}

// A small real allocation keeps the ownership cleanup mutant independent of compile time.
func TestMovesCleanupAgreesWithNode(t *testing.T) {
	t.Parallel()

	parallelFixtureUnit(t, "internal/oracle/testdata/moves/accepted/cleanup.a", "asan", "all", -1)
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
func moveSource(t *testing.T) string {
	t.Helper()
	return buildcache.Product(t, moveSourceInputs(), func(directory string) error {
		path, err := filepath.Abs(filepath.Join(repository, moveObjectsPath))
		if err != nil {
			return err
		}
		program, err := lowered(t, path)
		if err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(directory, "program.c"), []byte(native.C(program)), 0600); err != nil {
			return err
		}
		return os.WriteFile(filepath.Join(directory, "program.mjs"), []byte(javascript.JavaScript(program)), 0600)
	})
}

func moveSourceFile(t *testing.T, name string) string {
	t.Helper()
	bytes, err := os.ReadFile(filepath.Join(moveSource(t), name))
	if err != nil {
		t.Fatal(err)
	}
	return string(bytes)
}

func TestProduct_MovesSource(t *testing.T) {
	t.Parallel()
	moveSource(t)
}

func TestMovesSourceKeyNamesCompiler(t *testing.T) {
	t.Parallel()
	if err := compilerinputs.CheckKey([]string{moveObjectsPath}, func(root string) (string, error) {
		return buildcache.Key(root, moveSourceInputs())
	}); err != nil {
		t.Fatal(err)
	}
}

func moveBuildOptions(build string) (native.Options, string) {
	switch build {
	case "asan":
		return native.Options{Sanitize: true}, ""
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
func moveBinary(t *testing.T, build string) string {
	t.Helper()
	options, allocator := moveBuildOptions(build)
	inputs := moveSourceInputs()
	inputs.Name = "moves-native-" + build
	inputs.Flags = append(native.Flags(options), "allocator="+allocator)
	inputs.Toolchain = append(inputs.Toolchain, buildcache.Tool("clang", "--version"))
	directory := buildcache.Product(t, inputs, func(directory string) error {
		buildParallelProgram(t, moveSourceFile(t, "program.c"), filepath.Join(directory, "program"), options, allocator)
		return nil
	})
	return filepath.Join(directory, "program")
}

func TestProduct_MovesNative_asan(t *testing.T)       { t.Parallel(); moveBinary(t, "asan") }
func TestProduct_MovesNative_release(t *testing.T)    { t.Parallel(); moveBinary(t, "release") }
func TestProduct_MovesNative_asan_slabs(t *testing.T) { t.Parallel(); moveBinary(t, "asan_slabs") }
func TestProduct_MovesNative_tsan(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("Linux TSan")
	}
	moveBinary(t, "tsan")
}
func moveNativeVariant(t *testing.T, build string) {
	t.Helper()
	directory := moveSource(t)
	path, err := filepath.Abs(filepath.Join(repository, moveObjectsPath))
	if err != nil {
		t.Fatal(err)
	}
	source := onNode(t, path)
	if source.exitCode != 0 {
		t.Fatalf("Node: %+v", source)
	}
	if difference := disagreement(source, onNode(t, filepath.Join(directory, "program.mjs"))); difference != "" {
		t.Fatal(difference)
	}
	binary := moveBinary(t, build)
	options, _ := moveBuildOptions(build)
	for _, threads := range []string{"1", ""} {
		repeats := 1
		if options.ThreadSanitize {
			repeats = 3
		}
		for attempt := 0; attempt < repeats; attempt++ {
			observed := executeParallel(t, threads, false, binary)
			if difference := disagreement(source, observed); difference != "" {
				t.Fatalf("%s: exit %d stderr %s", difference, observed.exitCode, observed.stderr)
			}
		}
		if runtime.GOOS == "linux" && options.Sanitize {
			observed := executeParallel(t, threads, true, binary)
			if difference := disagreement(source, observed); difference != "" {
				t.Fatalf("leak check: %s: exit %d stderr %s", difference, observed.exitCode, observed.stderr)
			}
		}
		if runtime.GOOS == "darwin" && build == "malloc" {
			observed := executeParallel(t, threads, false, "leaks", "--atExit", "--", binary)
			if observed.exitCode != 0 {
				t.Fatalf("leaks: %s %s", observed.stdout, observed.stderr)
			}
		}
	}
}

func moveDeterminism(t *testing.T) string {
	t.Helper()
	inputs := moveSourceInputs()
	inputs.Name = "moves-determinism"
	return buildcache.Product(t, inputs, func(directory string) error {
		binary := buildcache.AdamicBuiltElsewhere(t)
		// Independent processes need no ordering. Run both on this product's CPU
		// share so two quadratic lowerings do not consume two leaf budgets.
		emissions := make(chan emitted, 2)
		for range 2 {
			go func() { emissions <- emit(t, binary, moveObjectsPath) }()
		}
		first, second := <-emissions, <-emissions
		if first.code != 0 || second.code != 0 || !bytes.Equal(first.output, second.output) {
			return fmt.Errorf("stage 0 emitted different C: exit %d and %d, first difference at line %d", first.code, second.code, firstDifferingLine(first.output, second.output))
		}
		return os.WriteFile(filepath.Join(directory, "program.c"), first.output, 0600)
	})
}

func TestProduct_MovesDeterminism(t *testing.T)    { t.Parallel(); moveDeterminism(t) }
func TestMovesEmitsDeterministically(t *testing.T) { t.Parallel(); moveDeterminism(t) }

// A new generic family must not silently put the heavy lowering back in a leaf.
func TestMovesObjectsUseDedicatedProducts(t *testing.T) {
	t.Parallel()
	for _, keys := range [][]string{fixturePaths(), determinismKeys(), countKeys()} {
		if slices.Contains(keys, moveObjectsPath) {
			t.Fatal("objects.a must use its dedicated moves products")
		}
	}
}
