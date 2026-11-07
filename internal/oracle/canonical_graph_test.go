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

func TestGraphParallelMapRefusesRegions(test *testing.T) {
	test.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/concurrency/refused/graph_region.a"))
	if err != nil {
		test.Fatal(err)
	}
	_, err = lowered(test, path)
	var refused *lower.Refused
	if !errors.As(err, &refused) ||
		!strings.Contains(refused.What, "Ring.next") ||
		!strings.Contains(refused.Fix, "part 2 (#p286ycm)") {
		test.Fatalf("want named graph ownership refusal and move fix, got %v", err)
	}
	test.Log(err)
}

func TestGraphTaskLocalThreadSanitizer(test *testing.T) {
	test.Parallel()
	if runtime.GOOS != "linux" {
		test.Skip("requires Linux ThreadSanitizer")
	}
	path, err := filepath.Abs(
		filepath.Join(repository, "internal/oracle/testdata/concurrency/accepted/graph_regions_per_task.a"),
	)
	if err != nil {
		test.Fatal(err)
	}
	truth := onNode(test, path)
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	if len(program.GraphTypes) == 0 {
		test.Fatal("task did not use graph regions")
	}
	binary := filepath.Join(test.TempDir(), "task-local")
	if err := native.Build(native.C(program), binary, native.Options{ThreadSanitize: true}); err != nil {
		test.Fatal(err)
	}
	result := executeParallel(test, "4", false, binary)
	if difference := disagreement(truth, result); difference != "" {
		test.Fatalf("ThreadSanitizer: %s, %s", difference, result.stderr)
	}
}

// These mutants compile; a sanitizer or graph invariant must kill them at run time.
func graphOracleMutant(test *testing.T, fixture string, change func(string) string) {
	test.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata", fixture))
	if err != nil {
		test.Fatal(err)
	}
	truth := onNode(test, path)
	program, err := lowered(test, path)
	if err != nil {
		test.Fatal(err)
	}
	code := native.C(program)
	changed := change(code)
	if changed == code {
		test.Fatal("mutant changed no emitted C")
	}
	binary := filepath.Join(test.TempDir(), "mutant")
	if err := native.Build(changed, binary, native.Options{Sanitize: true}); err != nil {
		test.Fatalf("mutant must compile: %v", err)
	}
	result := executeWith(test, []string{
		"ASAN_OPTIONS=detect_leaks=1",
		"LSAN_OPTIONS=use_stacks=0:use_registers=0",
		"UBSAN_OPTIONS=halt_on_error=1",
	}, binary)
	if disagreement(truth, result) == "" {
		test.Fatal("ownership mutant survived")
	}
	diagnostic := string(result.stderr)
	if !strings.Contains(diagnostic, "Sanitizer") &&
		!strings.Contains(diagnostic, "runtime error:") &&
		!strings.Contains(diagnostic, "compiler bug: graph") {
		test.Fatalf("unexpected mutant death: %d %s", result.exitCode, result.stderr)
	}
	test.Logf("%s mutant caught: %s", fixture, diagnostic)
}
func TestGraphOracleAdoptionMutant(test *testing.T) {
	test.Parallel()
	graphOracleMutant(test, "graph_regions_list.a", func(code string) string {
		return regexp.MustCompile(`adamic_graph_adopt_owned\(([^,]+), [^;]+\)`).ReplaceAllString(code, `$1`)
	})
}
func TestGraphOracleSlotHoldMutant(test *testing.T) {
	test.Parallel()
	graphOracleMutant(test, "graph_regions_list.a", func(code string) string {
		return regexp.MustCompile(`adamic_graph_hold\([^,]+, ([^)]+)\)`).ReplaceAllString(code, `adamic_retain($1)`)
	})
}
func TestGraphOracleCanonicalCacheMutant(test *testing.T) {
	test.Parallel()
	graphOracleMutant(test, "canonical_graph.a", func(code string) string {
		return strings.ReplaceAll(code, "adamic_closure_canonical_graph(", "adamic_closure_canonical(")
	})
}
func TestGraphOracleFullObjectBytesMutant(test *testing.T) {
	test.Parallel()
	graphOracleMutant(test, "graph_regions_list.a", func(code string) string {
		return regexp.MustCompile(`adamic_object_allocation_size\([^)]*\)`).ReplaceAllString(
			code,
			`sizeof(adamic_object)`,
		)
	})
}
