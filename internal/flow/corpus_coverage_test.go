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

func checkAllFlowProgram(t *testing.T, path string) {
	t.Helper()
	// Keep the three independent Node observations. Only immutable setup is shared.
	for _, check := range flowProgramChecks {
		check(t, path)
	}
}

// beginFlowUnit holds a unit to the 30-second budget where it's measured: on the reference
// box (one Codex instance, 4 CPUs, cold), which sets ADAMIC_UNIT_BUDGET=1. Elsewhere a loaded
// machine only logs it, so the gate's correctness verdict never depends on its load.
func beginFlowUnit(t *testing.T) {
	t.Helper()
	began := time.Now()
	t.Cleanup(func() {
		if elapsed := time.Since(began); elapsed >= 30*time.Second {
			if os.Getenv("ADAMIC_UNIT_BUDGET") == "1" {
				t.Errorf("test unit exceeded 30 seconds: %s", elapsed)
			} else {
				t.Logf("test unit took %s, over the 30-second budget measured on the reference box", elapsed)
			}
		}
	})
}

func corpusTestName(prefix, path string) string {
	var label strings.Builder
	for _, character := range path {
		if character >= 'A' && character <= 'Z' || character >= 'a' && character <= 'z' || character >= '0' && character <= '9' {
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

func TestFlowCorpusUnitsCoverEveryProgram(t *testing.T) {
	t.Parallel()
	expected := programs(t)
	if len(flowCorpusTests) != 4 || len(flowProgramChecks) != 4 {
		t.Fatalf("corpus count: %d paths, %d expected, %d families, %d checks", len(flowCorpusPaths), len(expected), len(flowCorpusTests), len(flowProgramChecks))
	}
	families := []string{"SingleAssignment", "MutationRanges", "GraphPaths", "Liveness"}
	checks := []string{"checkSingleAssignmentProgram", "checkMutationRangesProgram", "checkGraphPathsProgram", "checkLivenessProgram"}
	live := map[string]bool{}
	for _, path := range expected {
		live[filepath.ToSlash(path)] = true
	}
	seen := map[string]bool{}
	units := map[string]bool{}
	for index, path := range flowCorpusPaths {
		if !live[path] || seen[path] {
			t.Fatalf("corpus coverage at %d: %q is missing or duplicated", index, path)
		}
		seen[path] = true
		for family, bindings := range flowCorpusTests {
			if len(bindings) != len(flowCorpusPaths) {
				t.Fatalf("family %s covers %d of %d programs", families[family], len(bindings), len(expected))
			}
			prefix := "TestFlowProgram"
			if path == "../oracle/testdata/timsort.a" {
				prefix = "TestFlow" + families[family]
			}
			want := corpusTestName(prefix, path)
			if got := functionName(bindings[index]); got != want {
				t.Fatalf("unit for %s/%s: %s, want %s", families[family], path, got, want)
			}
			units[want] = true
		}
	}
	for index, check := range flowProgramChecks {
		if got := functionName(check); got != checks[index] {
			t.Fatalf("check %d: %s, want %s", index, got, checks[index])
		}
	}
	t.Logf("%d programs, %d analysis pieces, %d selectable units", len(expected), len(expected)*4, len(units))
}

func TestFlowCorpusSetupIsShared(t *testing.T) {
	t.Parallel()
	path := "../../dedication/dedication.a"
	if lowered(t, path) != lowered(t, path) {
		t.Fatal("lowered program setup was rebuilt")
	}
	if tracePrepared(t, path) != tracePrepared(t, path) {
		t.Fatal("trace setup was rebuilt")
	}
}

// Keep new programs covered in bounded units until the selectable corpus is regenerated.
func TestFlowCorpusRemainder0(t *testing.T) {
	t.Parallel()
	checkFlowCorpusRemainder(t, 0)
}
func TestFlowCorpusRemainder1(t *testing.T) {
	t.Parallel()
	checkFlowCorpusRemainder(t, 1)
}
func TestFlowCorpusRemainder2(t *testing.T) {
	t.Parallel()
	checkFlowCorpusRemainder(t, 2)
}
func TestFlowCorpusRemainder3(t *testing.T) {
	t.Parallel()
	checkFlowCorpusRemainder(t, 3)
}

func checkFlowCorpusRemainder(t *testing.T, shard byte) {
	t.Helper()
	beginFlowUnit(t)
	generated := map[string]bool{}
	for _, path := range flowCorpusPaths {
		generated[path] = true
	}
	for _, unit := range asyncFlowUnits {
		generated[unit.path] = true
	}
	count := 0
	for _, path := range programs(t) {
		canonical := filepath.ToSlash(path)
		if generated[canonical] || sha256.Sum256([]byte(canonical))[0]%4 != shard {
			continue
		}
		count++
		t.Run(filepath.ToSlash(path), func(t *testing.T) {
			checkAllFlowProgram(t, path)
		})
	}
	t.Logf("remainder ran %d programs", count)
}

// A lowering declaration excludes only the named witness and pins lowering's reason.
func TestFlowCorpusBoundaryDeclarations(t *testing.T) {
	t.Parallel()
	runnable := map[string]bool{}
	for _, path := range programs(t) {
		runnable[filepath.ToSlash(path)] = true
	}
	count := 0
	for _, pattern := range []string{"../oracle/testdata/coverage_regexp_*.a", "../oracle/testdata/coverage_view_*.a", "../oracle/testdata/json_decode_coverage_empty.a", "../oracle/testdata/json_decode_empty_object.a"} {
		paths, err := filepath.Glob(pattern)
		if err != nil {
			t.Fatal(err)
		}
		for _, path := range paths {
			reason := runtimeBoundaryLoweringReason(path)
			if reason == "" {
				t.Fatalf("missing lowering declaration for %s", path)
			}
			if runnable[filepath.ToSlash(path)] {
				t.Fatalf("refused fixture entered flow corpus: %s", path)
			}
			absolute, err := filepath.Abs(path)
			if err != nil {
				t.Fatal(err)
			}
			program, err := load.Load([]string{absolute})
			if err != nil {
				t.Fatal(err)
			}
			_, err = Lower(context.Background(), program)
			if err == nil || !strings.Contains(err.Error(), reason) {
				t.Fatalf("%s: declared lowering refusal %q, got %v", path, reason, err)
			}
			count++
		}
	}
	if count != 10 {
		t.Fatalf("want ten boundary declarations, got %d", count)
	}
}
