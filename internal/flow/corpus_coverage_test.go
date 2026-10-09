package flow

import (
	"bytes"
	"context"
	"crypto/sha256"
	"fmt"
	"os"
	"os/exec"
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

// A long remainder is the signal to regenerate the selectable corpus units.
// Until then it keeps new programs covered without requiring regeneration.
func TestFlowCorpusRemainder(t *testing.T) {
	t.Parallel()
	beginFlowUnit(t)
	generated := map[string]bool{}
	for _, path := range flowCorpusPaths {
		generated[path] = true
	}
	count := 0
	for _, path := range programs(t) {
		if generated[filepath.ToSlash(path)] {
			continue
		}
		count++
		t.Run(filepath.ToSlash(path), func(t *testing.T) {
			checkAllFlowProgram(t, path)
		})
	}
	t.Logf("remainder ran %d programs", count)
}

// The accessor source is a pinned lowering refusal, not a graph-bearing program.
// Keep its old selectable unit, but require the exact refusal and Node behavior.
func checkRefusedAccessorProgram(t *testing.T, path string) {
	t.Helper()
	source, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	canonical, err := os.ReadFile(filepath.Join(filepath.Dir(path), "class_features_refused", filepath.Base(path)))
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(source, canonical) {
		t.Fatal("accessor source differs from pinned refusal")
	}
	program, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, err = Lower(context.Background(), program)
	if err == nil || !strings.Contains(err.Error(), "can't lower spreading an accessor literal whose getter may throw yet") {
		t.Fatalf("expected throwing accessor spread gap, got %v", err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	command := exec.CommandContext(ctx, "node", "--disable-warning=ExperimentalWarning", "../../oracle/node.mjs", path)
	output, err := command.CombinedOutput()
	const want = "6\nGgSsGg\n5/5/GgSsGgrwr\namount,label\n5/literal/amount,label\ngetter1\nsetterinput2\n2,10,label\n2,10,label/ab\n3/3\n1/3\n"
	if err != nil || string(output) != want {
		t.Fatalf("source Node = %v, %q, want %q", err, output, want)
	}
	t.Log("source Node agrees; throwing accessor spread remains explicitly refused")
}
