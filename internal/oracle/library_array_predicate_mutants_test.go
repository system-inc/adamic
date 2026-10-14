package oracle

import (
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/native"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
)

func TestArrayFilterNodeMutant(t *testing.T) {
	t.Parallel()
	arrayPredicateNodeMutant(t, "filter", false)
}
func TestArraySomeNodeMutant(t *testing.T) { t.Parallel(); arrayPredicateNodeMutant(t, "some", false) }
func TestArrayEveryNodeMutant(t *testing.T) {
	t.Parallel()
	arrayPredicateNodeMutant(t, "every", false)
}
func TestArrayNaNTruthNodeMutant(t *testing.T) {
	t.Parallel()
	arrayPredicateNodeMutant(t, "filter", true)
}

func arrayPredicateNodeMutant(t *testing.T, method string, nan bool) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_predicates/library_array_predicate.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for i := range program.Functions {
		f := &program.Functions[i]
		if f.Name != "array_predicate_"+method {
			continue
		}
		for j, statement := range f.Body {
			loop, ok := statement.(ir.Loop)
			if !ok {
				continue
			}
			present := loop.Body[0].(ir.If)
			last := len(present.Then) - 1
			selected := present.Then[last].(ir.If)
			if nan {
				condition, ok := selected.Condition.(ir.Binary)
				if !ok || condition.Operator != ir.And {
					continue
				}
				selected.Condition = condition.Left
			} else {
				selected.Condition = ir.Unary{Operator: ir.Not, Operand: selected.Condition}
			}
			present.Then[last] = selected
			loop.Body[0] = present
			f.Body[j] = loop
			changed = true
		}
	}
	if !changed {
		t.Fatal("mutant changed nothing")
	}
	arrayNodeOnlyMutant(t, native.C(program), path)
}

func TestArrayFindLastNodeMutant(t *testing.T) {
	t.Parallel()
	arrayFindLastNodeMutant(t, "findLast", "search_shrink_findLast_maybe.a")
}
func TestArrayFindLastIndexNodeMutant(t *testing.T) {
	t.Parallel()
	arrayFindLastNodeMutant(t, "findLastIndex", "search_shrink_findLastIndex_maybe.a")
}
func arrayFindLastNodeMutant(t *testing.T, method, fixture string) {
	t.Helper()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_predicates", fixture))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	changed := false
	for i := range program.Functions {
		f := &program.Functions[i]
		if f.Name != "array_"+method {
			continue
		}
		for j, statement := range f.Body {
			loop, ok := statement.(ir.Loop)
			if !ok {
				continue
			}
			update := loop.Update[0].(ir.Assign)
			decrement := update.Value.(ir.Binary)
			decrement.Right = ir.NumberConstant{Value: 2}
			update.Value = decrement
			loop.Update[0] = update
			f.Body[j] = loop
			changed = true
		}
	}
	if !changed {
		t.Fatal("mutant changed nothing")
	}
	arrayNodeOnlyMutant(t, native.C(program), path)
}

func TestArrayShrinkingSearchNodeMutant(t *testing.T) {
	t.Parallel()
	path, err := filepath.Abs(filepath.Join(repository, "internal/oracle/testdata/array_predicates/search_shrink_find_maybe.a"))
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowered(t, path)
	if err != nil {
		t.Fatal(err)
	}
	source := native.C(program)
	pattern := regexp.MustCompile(`bool (\w+) = (\w+) < (\w+)->length;`)
	mutant := pattern.ReplaceAllString(source, `$0 if (!$1) continue;`)
	if mutant == source {
		t.Fatal("mutant changed nothing")
	}
	arrayNodeOnlyMutant(t, mutant, path)
}

// Sanitizers, leaks, exit status and stderr must all stay clean. Only Node's
// independent source output is allowed to kill a behavioral mutation.
func arrayNodeOnlyMutant(t *testing.T, source, path string) {
	t.Helper()
	binary := filepath.Join(t.TempDir(), "mutant")
	if err := native.Build(source, binary, native.Options{Sanitize: true}); err != nil {
		t.Fatal(err)
	}
	got := executeWith(t, []string{"ASAN_OPTIONS=detect_leaks=1:halt_on_error=1", "UBSAN_OPTIONS=halt_on_error=1"}, binary)
	truth := onNode(t, path)
	if got.exitCode != 0 || len(got.stderr) != 0 || strings.Contains(string(got.stderr), "Sanitizer") || disagreement(truth, got) != "stdout differs" {
		t.Fatalf("mutant failed outside Node stdout: truth %+v got %+v", truth, got)
	}
	t.Log("caught only by Node stdout; exit 0, empty stderr, ASan/UBSan/LeakSanitizer clean")
}
