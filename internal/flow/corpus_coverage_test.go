package flow

import (
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/system-inc/adamic/internal/load"
)

var flowProgramChecks = []func(*testing.T, string){
	checkSingleAssignmentProgram, checkMutationRangesProgram, checkGraphPathsProgram, checkLivenessProgram,
}

func checkAllFlowProgram(test *testing.T, path string) {
	test.Helper()
	// Keep the three independent Node observations. Only immutable setup is shared.
	for _, check := range flowProgramChecks {
		check(test, path)
	}
}

// beginFlowUnit holds a unit to the 30-second budget where it's measured: on the reference
// box (one Codex instance, 4 CPUs, cold), which sets ADAMIC_UNIT_BUDGET=1. Elsewhere a loaded
// machine only logs it, so the gate's correctness verdict never depends on its load.
func beginFlowUnit(test *testing.T) {
	test.Helper()
	began := time.Now()
	test.Cleanup(func() {
		if elapsed := time.Since(began); elapsed >= 30*time.Second {
			if os.Getenv("ADAMIC_UNIT_BUDGET") == "1" {
				test.Errorf("test unit exceeded 30 seconds: %s", elapsed)
			} else {
				test.Logf("test unit took %s, over the 30-second budget measured on the reference box", elapsed)
			}
		}
	})
}

func corpusTestName(prefix, path string) string {
	var label strings.Builder
	for _, character := range path {
		if character >= 'A' && character <= 'Z' ||
			character >= 'a' && character <= 'z' ||
			character >= '0' && character <= '9' {
			label.WriteRune(character)
		} else {
			label.WriteByte('_')
		}
	}
	digest := fmt.Sprintf("%x", sha256.Sum256([]byte(path)))
	return prefix + "_" + label.String() + "_" + digest[:12]
}

func functionName(function any) string {
	name := runtime.FuncForPC(reflect.ValueOf(function).Pointer()).Name()
	return name[strings.LastIndex(name, ".")+1:]
}

func TestFlowCorpusUnitsCoverEveryProgram(test *testing.T) {
	test.Parallel()
	expected := programs(test)
	if len(flowCorpusTests) != 4 || len(flowProgramChecks) != 4 {
		test.Fatalf(
			"corpus count: %d paths, %d expected, %d families, %d checks",
			len(flowCorpusPaths),
			len(expected),
			len(flowCorpusTests),
			len(flowProgramChecks),
		)
	}
	families := []string{"SingleAssignment", "MutationRanges", "GraphPaths", "Liveness"}
	checks := []string{
		"checkSingleAssignmentProgram",
		"checkMutationRangesProgram",
		"checkGraphPathsProgram",
		"checkLivenessProgram",
	}
	live := map[string]bool{}
	for _, path := range expected {
		live[filepath.ToSlash(path)] = true
	}
	seen := map[string]bool{}
	units := map[string]bool{}
	for index, path := range flowCorpusPaths {
		if !live[path] || seen[path] {
			test.Fatalf("corpus coverage at %d: %q is missing or duplicated", index, path)
		}
		seen[path] = true
		for family, bindings := range flowCorpusTests {
			if len(bindings) != len(flowCorpusPaths) {
				test.Fatalf("family %s covers %d of %d programs", families[family], len(bindings), len(expected))
			}
			prefix := "TestFlowProgram"
			if path == "../oracle/testdata/timsort.a" || path == "../oracle/testdata/node_fs_directory_system.a" {
				prefix = "TestFlow" + families[family]
			}
			want := corpusTestName(prefix, path)
			if got := functionName(bindings[index]); got != want {
				test.Fatalf("unit for %s/%s: %s, want %s", families[family], path, got, want)
			}
			units[want] = true
		}
	}
	for index, check := range flowProgramChecks {
		if got := functionName(check); got != checks[index] {
			test.Fatalf("check %d: %s, want %s", index, got, checks[index])
		}
	}
	test.Logf("%d programs, %d analysis pieces, %d selectable units", len(expected), len(expected)*4, len(units))
}

func TestFlowCorpusSetupIsShared(test *testing.T) {
	test.Parallel()
	path := "../../dedication/dedication.a"
	if lowered(test, path) != lowered(test, path) {
		test.Fatal("lowered program setup was rebuilt")
	}
	if tracePrepared(test, path) != tracePrepared(test, path) {
		test.Fatal("trace setup was rebuilt")
	}
}

// A lowering declaration excludes only the named witness and pins lowering's reason.
func TestFlowCorpusBoundaryDeclarations(test *testing.T) {
	test.Parallel()
	runnable := map[string]bool{}
	for _, path := range programs(test) {
		runnable[filepath.ToSlash(path)] = true
	}
	count := 0
	for _, pattern := range []string{
		"../oracle/testdata/coverage_regexp_*.a",
		"../oracle/testdata/coverage_view_*.a",
		"../oracle/testdata/json_decode_coverage_empty.a",
		"../oracle/testdata/json_decode_empty_object.a",
	} {
		paths, err := filepath.Glob(pattern)
		if err != nil {
			test.Fatal(err)
		}
		for _, path := range paths {
			reason := runtimeBoundaryLoweringReason(path)
			if reason == "" {
				test.Fatalf("missing lowering declaration for %s", path)
			}
			if runnable[filepath.ToSlash(path)] {
				test.Fatalf("refused fixture entered flow corpus: %s", path)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				test.Fatal(err)
			}
			program, err := load.Load([]string{absolute})
			if err != nil {
				test.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			if err == nil || !strings.Contains(err.Error(), reason) {
				test.Fatalf("%s: declared lowering refusal %q, got %v", path, reason, err)
			}
			count++
		}
	}
	if count != 10 {
		test.Fatalf("want ten boundary declarations, got %d", count)
	}
}
