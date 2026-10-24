package lower

import (
	"context"
	"github.com/system-inc/adamic/internal/buildcache"
	"github.com/system-inc/adamic/internal/compilerinputs"
	"os"
	"path/filepath"
	"runtime"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	checkerapi "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/fresh"
	"github.com/system-inc/adamic/internal/load"
)

func ownershipAgreement(t *testing.T, path string) {
	t.Helper()
	checked, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	old, oldErr := LowerWithOptions(context.Background(), checked, Options{})
	checked, err = load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	_, newErr := LowerWithOptions(context.Background(), checked, Options{OwnershipQuery: true})
	if (oldErr == nil) != (newErr == nil) {
		t.Fatalf("move/query disagreement: old=%v query=%v", oldErr, newErr)
	}
	if oldErr != nil && oldErr.Error() != newErr.Error() {
		t.Fatalf("refusal changed: %v vs %v", oldErr, newErr)
	}
	if oldErr == nil {
		for _, answer := range fresh.OwnershipTransfers(old) {
			if answer.Answer.Verdict != fresh.Proven {
				t.Fatalf("admitted move not Proven: %+v", answer)
			}
		}
	}
}

func TestOwnershipQuerySiteDemand(t *testing.T) {
	t.Parallel()
	path := "../oracle/testdata/ownership-query/flat.a"
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	entry := loaded.Files()[0]
	checker, release := loaded.Checker(context.Background(), entry)
	defer release()
	l := &lowering{program: loaded, checker: checker}
	var call *ast.Node
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindCallExpression && ast.IsIdentifier(n.AsCallExpression().Expression) && n.AsCallExpression().Expression.Text() == "parallelMap" {
			call = n
			return false
		}
		return n.ForEachChild(visit)
	}
	entry.AsNode().ForEachChild(visit)
	if call == nil {
		t.Fatal("missing parallel site")
	}
	if !l.parallelMoves(call) || len(l.parallelMoveDemand) != 1 {
		t.Fatal("site demand not recorded")
	}
	l.checker = nil
	if !l.parallelMoves(call) {
		t.Fatal("cached demand changed")
	}
	l.checker = checker
	actual := checker.GetTypeAtLocation(call.AsCallExpression().Arguments.Nodes[0])
	l.typeMapper = newTypeMapper([]*checkerapi.Type{actual}, []*checkerapi.Type{actual})
	if !l.parallelMoves(call) || len(l.parallelMoveDemand) != 2 {
		t.Fatal("instantiations shared demand")
	}
}

func TestOwnershipQueryValidation(t *testing.T) {
	t.Parallel()
	loaded, err := load.Load([]string{"../oracle/testdata/ownership-query/flat.a"})
	if err != nil {
		t.Fatal(err)
	}
	p, err := LowerWithOptions(context.Background(), loaded, Options{OwnershipQuery: true})
	if err != nil {
		t.Fatal(err)
	}
	answers := fresh.QueryOwnershipTransfers(p)
	if len(answers) != 1 || answers[0].Site < 1 || answers[0].At == nil || answers[0].Answer.Verdict != fresh.Proven {
		t.Fatalf("validation was not exercised: %+v", answers)
	}
	s := answers[0].Answer.Snapshot
	s.Roots = append(s.Roots, fresh.OwnershipRoot{Name: "second owner", Object: s.Roots[0].Object})
	if fresh.QueryOwnership(s, s.Roots[0].Name).Verdict != fresh.Refused {
		t.Fatal("second surviving owner accepted")
	}
}

func BenchmarkOwnershipQueryLowering(b *testing.B) {
	path := os.Getenv("ADAMIC_OWNERSHIP_BENCH_SOURCE")
	if path == "" {
		b.Fatal("set ADAMIC_OWNERSHIP_BENCH_SOURCE")
	}
	for _, enabled := range []bool{false, true} {
		name := "off"
		if enabled {
			name = "on"
		}
		b.Run(name, func(b *testing.B) {
			for i := 0; i < b.N; i++ {
				b.StopTimer()
				checked, err := load.Load([]string{path})
				if err != nil {
					b.Fatal(err)
				}
				checked.EnableTSGo()
				b.StartTimer()
				p, err := LowerWithOptions(context.Background(), checked, Options{OwnershipQuery: enabled})
				if err != nil {
					b.Fatal(err)
				}
				b.StopTimer()
				b.Logf("functions=%d locals=%d", len(p.Functions), len(p.Locals))
			}
		})
	}
}

// Fixtures each own a bounded unit; no admission shortcut is enabled in production.
func TestOwnershipAgreement_concurrency_accepted_fresh(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/fresh.a")
}
func TestOwnershipAgreement_concurrency_accepted_global_capture(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/global_capture.a")
}
func TestOwnershipAgreement_concurrency_accepted_graph_regions_per_task(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/graph_regions_per_task.a")
}
func TestOwnershipAgreement_concurrency_accepted_identity(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/identity.a")
}
func TestOwnershipAgreement_concurrency_accepted_large(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/large.a")
}
func TestOwnershipAgreement_concurrency_accepted_map_capture(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/map_capture.a")
}
func TestOwnershipAgreement_concurrency_accepted_nested(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/nested.a")
}
func TestOwnershipAgreement_concurrency_accepted_numbers(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/numbers.a")
}
func TestOwnershipAgreement_concurrency_accepted_readonly_forms(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/readonly_forms.a")
}
func TestOwnershipAgreement_concurrency_accepted_readonly_function(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/readonly_function.a")
}
func TestOwnershipAgreement_concurrency_accepted_records(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/records.a")
}
func TestOwnershipAgreement_concurrency_accepted_recursive_tree(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/recursive_tree.a")
}
func TestOwnershipAgreement_concurrency_accepted_strings(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/strings.a")
}
func TestOwnershipAgreement_concurrency_accepted_throw_middle(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/accepted/throw_middle.a")
}
func TestOwnershipAgreement_concurrency_refused_async(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/async.a")
}
func TestOwnershipAgreement_concurrency_refused_await(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/await.a")
}
func TestOwnershipAgreement_concurrency_refused_capture_write(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/capture_write.a")
}
func TestOwnershipAgreement_concurrency_refused_captured_let(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/captured_let.a")
}
func TestOwnershipAgreement_concurrency_refused_class_mutable(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/class_mutable.a")
}
func TestOwnershipAgreement_concurrency_refused_console(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/console.a")
}
func TestOwnershipAgreement_concurrency_refused_date(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/date.a")
}
func TestOwnershipAgreement_concurrency_refused_directory(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/directory.a")
}
func TestOwnershipAgreement_concurrency_refused_dispatch(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/dispatch.a")
}
func TestOwnershipAgreement_concurrency_refused_file_read(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/file_read.a")
}
func TestOwnershipAgreement_concurrency_refused_file_write(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/file_write.a")
}
func TestOwnershipAgreement_concurrency_refused_fresh_alias(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/fresh_alias.a")
}
func TestOwnershipAgreement_concurrency_refused_global_write(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/global_write.a")
}
func TestOwnershipAgreement_concurrency_refused_graph_region(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/graph_region.a")
}
func TestOwnershipAgreement_concurrency_refused_hidden_subclass(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/hidden_subclass.a")
}
func TestOwnershipAgreement_concurrency_refused_item_write(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/item_write.a")
}
func TestOwnershipAgreement_concurrency_refused_map_field(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/map_field.a")
}
func TestOwnershipAgreement_concurrency_refused_mapped_map(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/mapped_map.a")
}
func TestOwnershipAgreement_concurrency_refused_mutable_array(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/mutable_array.a")
}
func TestOwnershipAgreement_concurrency_refused_mutable_field(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/mutable_field.a")
}
func TestOwnershipAgreement_concurrency_refused_mutable_set(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/mutable_set.a")
}
func TestOwnershipAgreement_concurrency_refused_nested_array(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/nested_array.a")
}
func TestOwnershipAgreement_concurrency_refused_process(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/process.a")
}
func TestOwnershipAgreement_concurrency_refused_random(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/random.a")
}
func TestOwnershipAgreement_concurrency_refused_reassigned_parameter(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/reassigned_parameter.a")
}
func TestOwnershipAgreement_concurrency_refused_union_mutable(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/union_mutable.a")
}
func TestOwnershipAgreement_concurrency_refused_unknown_call(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/unknown_call.a")
}
func TestOwnershipAgreement_concurrency_refused_unknown_type(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/unknown_type.a")
}
func TestOwnershipAgreement_concurrency_refused_weak(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/concurrency/refused/weak.a")
}
func TestOwnershipAgreement_moves_accepted_cleanup(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/accepted/cleanup.a")
}

func TestOwnershipAgreement_moves_refused_after(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/after.a")
}
func TestOwnershipAgreement_moves_refused_alias(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/alias.a")
}
func TestOwnershipAgreement_moves_refused_array_capture(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/array_capture.a")
}
func TestOwnershipAgreement_moves_refused_array_owner(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/array_owner.a")
}
func TestOwnershipAgreement_moves_refused_async(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/async.a")
}
func TestOwnershipAgreement_moves_refused_bare_return(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/bare_return.a")
}
func TestOwnershipAgreement_moves_refused_callback_closure(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/callback_closure.a")
}
func TestOwnershipAgreement_moves_refused_capture(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/capture.a")
}
func TestOwnershipAgreement_moves_refused_closure(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/closure.a")
}
func TestOwnershipAgreement_moves_refused_cross_block(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/cross_block.a")
}
func TestOwnershipAgreement_moves_refused_default_parameter(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/default_parameter.a")
}
func TestOwnershipAgreement_moves_refused_escape(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/escape.a")
}
func TestOwnershipAgreement_moves_refused_expression_work(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/expression_work.a")
}
func TestOwnershipAgreement_moves_refused_field_store(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/field_store.a")
}
func TestOwnershipAgreement_moves_refused_global(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/global.a")
}
func TestOwnershipAgreement_moves_refused_helper(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/helper.a")
}
func TestOwnershipAgreement_moves_refused_item_destructure(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/item_destructure.a")
}
func TestOwnershipAgreement_moves_refused_let_owner(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/let_owner.a")
}
func TestOwnershipAgreement_moves_refused_main_block(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/main_block.a")
}
func TestOwnershipAgreement_moves_refused_nested(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/nested.a")
}
func TestOwnershipAgreement_moves_refused_nested_arrays(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/nested_arrays.a")
}
func TestOwnershipAgreement_moves_refused_nested_call(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/nested_call.a")
}
func TestOwnershipAgreement_moves_refused_nested_map(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/nested_map.a")
}
func TestOwnershipAgreement_moves_refused_nested_race(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/nested_race.a")
}
func TestOwnershipAgreement_moves_refused_overlap(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/overlap.a")
}
func TestOwnershipAgreement_moves_refused_race_alias(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/race_alias.a")
}
func TestOwnershipAgreement_moves_refused_rebind(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/rebind.a")
}
func TestOwnershipAgreement_moves_refused_result(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/result.a")
}
func TestOwnershipAgreement_moves_refused_spread(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/spread.a")
}
func TestOwnershipAgreement_moves_refused_unknown(t *testing.T) {
	t.Parallel()
	ownershipAgreement(t, "../oracle/testdata/moves/refused/unknown.a")
}

func ownershipObjectsAgreementProduct(t *testing.T) string {
	t.Helper()
	inputs := buildcache.Inputs{Name: "ownership-query-objects-agreement", Files: append(compilerinputs.Lowering(), "internal/lower/ownership_query_test.go", "internal/oracle/testdata/moves/accepted/objects.a"), Toolchain: []string{runtime.Version()}}
	return buildcache.Product(t, inputs, func(directory string) error {
		ownershipAgreement(t, "../oracle/testdata/moves/accepted/objects.a")
		return os.WriteFile(filepath.Join(directory, "agreement"), []byte("admission and extracted query agree\n"), 0600)
	})
}
func TestProduct_OwnershipObjectsAgreement(t *testing.T) {
	t.Parallel()
	ownershipObjectsAgreementProduct(t)
}
func TestOwnershipObjectsAgreementRecorded(t *testing.T) {
	t.Parallel()
	directory := ownershipObjectsAgreementProduct(t)
	data, err := os.ReadFile(filepath.Join(directory, "agreement"))
	if err != nil || string(data) != "admission and extracted query agree\n" {
		t.Fatalf("agreement evidence: %v %q", err, data)
	}
}

func TestOwnershipQueryDemandDoesNotCrossSites(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "sites.a")
	source := "import { parallelMap } from 'adamic'; const objects = [{value: 1}]; const numbers = [1]; parallelMap(objects, item => item); parallelMap(numbers, item => item);"
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	entry := loaded.Files()[0]
	checker, release := loaded.Checker(context.Background(), entry)
	defer release()
	l := &lowering{program: loaded, checker: checker}
	var calls []*ast.Node
	var visit ast.Visitor
	visit = func(n *ast.Node) bool {
		if n.Kind == ast.KindCallExpression && ast.IsIdentifier(n.AsCallExpression().Expression) && n.AsCallExpression().Expression.Text() == "parallelMap" {
			calls = append(calls, n)
		}
		return n.ForEachChild(visit)
	}
	entry.AsNode().ForEachChild(visit)
	if len(calls) != 2 {
		t.Fatalf("calls: %d", len(calls))
	}
	if !l.parallelMoves(calls[0]) || l.parallelMoves(calls[1]) || len(l.parallelMoveDemand) != 2 {
		t.Fatal("different sites shared their demand answer")
	}
	l.checker = nil
	if !l.parallelMoves(calls[0]) || l.parallelMoves(calls[1]) {
		t.Fatal("cached site answers changed")
	}
}
