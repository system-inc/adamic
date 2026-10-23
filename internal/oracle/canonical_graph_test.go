package oracle

import (
	"errors"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/lower"
	"github.com/system-inc/adamic/internal/native"
)

func TestGraphParallelMapRefusesRegions(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/refused/graph_region.a"))
	if err != nil {
		t.Fatal(err)
	}
	_, err = lowered(t, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) || !strings.Contains(refused.What, "Ring.next") || !strings.Contains(refused.Fix, "part 2 (#p286ycm)") {
		t.Fatalf("want named graph ownership refusal and move fix, got %v", err)
	}
	t.Log(err)
}

func TestGraphTaskLocalThreadSanitizer(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "linux" {
		t.Skip("requires Linux ThreadSanitizer")
	}
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/graph_regions_per_task.a"))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	if len(program.GraphTypes) == 0 {
		t.Fatal("task did not use graph regions")
	}
	binary := filepath.Join(t.TempDir(), "task-local")
	if err := native.Build(native.C(program), binary, native.Options{ThreadSanitize: true}); err != nil {
		t.Fatal(err)
	}
	result := executeParallel(t, "4", false, binary)
	if difference := disagreement(truth, result); difference != "" {
		t.Fatalf("ThreadSanitizer: %s, %s", difference, result.stderr)
	}
}

// These mutants compile; a sanitizer or graph invariant must kill them at run time.
func graphOracleMutant(t *testing.T, fixture string, change func(string) string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		t.Fatal(err)
	}
	truth := onNode(t, path)
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	code := native.C(program)
	changed := change(code)
	if changed == code {
		t.Fatal("mutant changed no emitted C")
	}
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatalf("mutant must compile: %v", err)
	}
	result := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1", "LSAN_OPTIONS=use_stacks=0:use_registers=0", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	if disagreement(truth, result) == "" {
		t.Fatal("ownership mutant survived")
	}
	diagnostic := string(result.stderr)
	if !strings.Contains(diagnostic, "Sanitizer") && !strings.Contains(diagnostic, "runtime error:") && !strings.Contains(diagnostic, "compiler bug: graph") {
		t.Fatalf("unexpected mutant death: %d %s", result.exitCode, result.stderr)
	}
	t.Logf("%s mutant caught: %s", fixture, diagnostic)
}
func TestGraphOracleAdoptionMutant(t *testing.T) {
	t.Parallel()
	graphOracleMutant(t, "graph_regions_list.a", func(code string) string {
		return regexp.MustCompile(`adamic_graph_adopt_owned\(([^,]+), [^;]+\)`).ReplaceAllString(code, `$1`)
	})
}
func TestGraphOracleSlotHoldMutant(t *testing.T) {
	t.Parallel()
	graphOracleMutant(t, "graph_regions_list.a", func(code string) string {
		return regexp.MustCompile(`adamic_graph_hold\([^,]+, ([^)]+)\)`).ReplaceAllString(code, `adamic_retain($1)`)
	})
}
func TestGraphOracleCanonicalCacheMutant(t *testing.T) {
	t.Parallel()
	graphOracleMutant(t, "canonical_graph.a", func(code string) string {
		return strings.ReplaceAll(code, "adamic_closure_canonical_graph(", "adamic_closure_canonical(")
	})
}
func TestGraphOracleFullObjectBytesMutant(t *testing.T) {
	t.Parallel()
	graphOracleMutant(t, "graph_regions_list.a", func(code string) string {
		return regexp.MustCompile(`adamic_object_size\([^)]*\)`).ReplaceAllString(code, `sizeof(adamic_object)`)
	})
}
