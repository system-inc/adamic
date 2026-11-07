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

func ownershipAgreement(test *testing.T, path string) {
	test.Helper()
	checked, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	old, oldError := LowerWithOptions(context.Background(), checked, Options{})
	checked, loadError = load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	_, newError := LowerWithOptions(context.Background(), checked, Options{OwnershipQuery: true})
	if (oldError == nil) != (newError == nil) {
		test.Fatalf("move/query disagreement: old=%v query=%v", oldError, newError)
	}
	if oldError != nil && oldError.Error() != newError.Error() {
		test.Fatalf("refusal changed: %v vs %v", oldError, newError)
	}
	if oldError == nil {
		for _, answer := range fresh.OwnershipTransfers(old) {
			if answer.Answer.Verdict != fresh.Proven {
				test.Fatalf("admitted move not Proven: %+v", answer)
			}
		}
	}
}

func TestOwnershipQuerySiteDemand(test *testing.T) {
	test.Parallel()
	path := "../oracle/testdata/ownership-query/flat.a"
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	entry := loaded.Files()[0]
	checker, release := loaded.Checker(context.Background(), entry)
	defer release()
	lowering := &lowering{program: loaded, checker: checker}
	var call *ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression &&
			ast.IsIdentifier(node.AsCallExpression().Expression) &&
			node.AsCallExpression().Expression.Text() == "parallelMap" {
			call = node
			return false
		}
		return node.ForEachChild(visit)
	}
	entry.AsNode().ForEachChild(visit)
	if call == nil {
		test.Fatal("missing parallel site")
	}
	if !lowering.parallelMoves(call) || len(lowering.parallelMoveDemand) != 1 {
		test.Fatal("site demand not recorded")
	}
	lowering.checker = nil
	if !lowering.parallelMoves(call) {
		test.Fatal("cached demand changed")
	}
	lowering.checker = checker
	actual := checker.GetTypeAtLocation(call.AsCallExpression().Arguments.Nodes[0])
	lowering.typeMapper = newTypeMapper([]*checkerapi.Type{actual}, []*checkerapi.Type{actual})
	if !lowering.parallelMoves(call) || len(lowering.parallelMoveDemand) != 2 {
		test.Fatal("instantiations shared demand")
	}
}

func TestOwnershipQueryValidation(test *testing.T) {
	test.Parallel()
	loaded, loadError := load.Load([]string{"../oracle/testdata/ownership-query/flat.a"})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerWithOptionsError := LowerWithOptions(context.Background(), loaded, Options{OwnershipQuery: true})
	if lowerWithOptionsError != nil {
		test.Fatal(lowerWithOptionsError)
	}
	answers := fresh.QueryOwnershipTransfers(program)
	if len(answers) != 1 || answers[0].Site < 1 || answers[0].At == nil || answers[0].Answer.Verdict != fresh.Proven {
		test.Fatalf("validation was not exercised: %+v", answers)
	}
	snapshot := answers[0].Answer.Snapshot
	snapshot.Roots = append(snapshot.Roots, fresh.OwnershipRoot{Name: "second owner", Object: snapshot.Roots[0].Object})
	if fresh.QueryOwnership(snapshot, snapshot.Roots[0].Name).Verdict != fresh.Refused {
		test.Fatal("second surviving owner accepted")
	}
}

func BenchmarkOwnershipQueryLowering(benchmark *testing.B) {
	path := os.Getenv("ADAMIC_OWNERSHIP_BENCH_SOURCE")
	if path == "" {
		benchmark.Fatal("set ADAMIC_OWNERSHIP_BENCH_SOURCE")
	}
	for _, enabled := range []bool{false, true} {
		name := "off"
		if enabled {
			name = "on"
		}
		benchmark.Run(name, func(benchmark *testing.B) {
			for index := 0; index < benchmark.N; index++ {
				benchmark.StopTimer()
				checked, loadError := load.Load([]string{path})
				if loadError != nil {
					benchmark.Fatal(loadError)
				}
				checked.EnableTSGo()
				benchmark.StartTimer()
				program, lowerWithOptionsError := LowerWithOptions(
					context.Background(),
					checked,
					Options{OwnershipQuery: enabled},
				)
				if lowerWithOptionsError != nil {
					benchmark.Fatal(lowerWithOptionsError)
				}
				benchmark.StopTimer()
				benchmark.Logf("functions=%d locals=%d", len(program.Functions), len(program.Locals))
			}
		})
	}
}

// Fixtures each own a bounded unit; no admission shortcut is enabled in production.
func TestOwnershipAgreement_concurrency_accepted_fresh(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/fresh.a")
}
func TestOwnershipAgreement_concurrency_accepted_global_capture(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/global_capture.a")
}
func TestOwnershipAgreement_concurrency_accepted_graph_regions_per_task(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/graph_regions_per_task.a")
}
func TestOwnershipAgreement_concurrency_accepted_identity(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/identity.a")
}
func TestOwnershipAgreement_concurrency_accepted_large(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/large.a")
}
func TestOwnershipAgreement_concurrency_accepted_map_capture(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/map_capture.a")
}
func TestOwnershipAgreement_concurrency_accepted_nested(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/nested.a")
}
func TestOwnershipAgreement_concurrency_accepted_numbers(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/numbers.a")
}
func TestOwnershipAgreement_concurrency_accepted_readonly_forms(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/readonly_forms.a")
}
func TestOwnershipAgreement_concurrency_accepted_readonly_function(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/readonly_function.a")
}
func TestOwnershipAgreement_concurrency_accepted_records(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/records.a")
}
func TestOwnershipAgreement_concurrency_accepted_recursive_tree(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/recursive_tree.a")
}
func TestOwnershipAgreement_concurrency_accepted_strings(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/strings.a")
}
func TestOwnershipAgreement_concurrency_accepted_throw_middle(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/accepted/throw_middle.a")
}
func TestOwnershipAgreement_concurrency_refused_async(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/async.a")
}
func TestOwnershipAgreement_concurrency_refused_await(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/await.a")
}
func TestOwnershipAgreement_concurrency_refused_capture_write(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/capture_write.a")
}
func TestOwnershipAgreement_concurrency_refused_captured_let(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/captured_let.a")
}
func TestOwnershipAgreement_concurrency_refused_class_mutable(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/class_mutable.a")
}
func TestOwnershipAgreement_concurrency_refused_console(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/console.a")
}
func TestOwnershipAgreement_concurrency_refused_date(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/date.a")
}
func TestOwnershipAgreement_concurrency_refused_directory(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/directory.a")
}
func TestOwnershipAgreement_concurrency_refused_dispatch(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/dispatch.a")
}
func TestOwnershipAgreement_concurrency_refused_file_read(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/file_read.a")
}
func TestOwnershipAgreement_concurrency_refused_file_write(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/file_write.a")
}
func TestOwnershipAgreement_concurrency_refused_fresh_alias(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/fresh_alias.a")
}
func TestOwnershipAgreement_concurrency_refused_global_write(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/global_write.a")
}
func TestOwnershipAgreement_concurrency_refused_graph_region(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/graph_region.a")
}
func TestOwnershipAgreement_concurrency_refused_hidden_subclass(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/hidden_subclass.a")
}
func TestOwnershipAgreement_concurrency_refused_item_write(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/item_write.a")
}
func TestOwnershipAgreement_concurrency_refused_map_field(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/map_field.a")
}
func TestOwnershipAgreement_concurrency_refused_mapped_map(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/mapped_map.a")
}
func TestOwnershipAgreement_concurrency_refused_mutable_array(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/mutable_array.a")
}
func TestOwnershipAgreement_concurrency_refused_mutable_field(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/mutable_field.a")
}
func TestOwnershipAgreement_concurrency_refused_mutable_set(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/mutable_set.a")
}
func TestOwnershipAgreement_concurrency_refused_nested_array(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/nested_array.a")
}
func TestOwnershipAgreement_concurrency_refused_process(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/process.a")
}
func TestOwnershipAgreement_concurrency_refused_random(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/random.a")
}
func TestOwnershipAgreement_concurrency_refused_reassigned_parameter(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/reassigned_parameter.a")
}
func TestOwnershipAgreement_concurrency_refused_union_mutable(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/union_mutable.a")
}
func TestOwnershipAgreement_concurrency_refused_unknown_call(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/unknown_call.a")
}
func TestOwnershipAgreement_concurrency_refused_unknown_type(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/unknown_type.a")
}
func TestOwnershipAgreement_concurrency_refused_weak(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/concurrency/refused/weak.a")
}
func TestOwnershipAgreement_moves_accepted_cleanup(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/accepted/cleanup.a")
}

func TestOwnershipAgreement_moves_refused_after(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/after.a")
}
func TestOwnershipAgreement_moves_refused_alias(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/alias.a")
}
func TestOwnershipAgreement_moves_refused_array_capture(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/array_capture.a")
}
func TestOwnershipAgreement_moves_refused_array_owner(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/array_owner.a")
}
func TestOwnershipAgreement_moves_refused_async(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/async.a")
}
func TestOwnershipAgreement_moves_refused_bare_return(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/bare_return.a")
}
func TestOwnershipAgreement_moves_refused_callback_closure(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/callback_closure.a")
}
func TestOwnershipAgreement_moves_refused_capture(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/capture.a")
}
func TestOwnershipAgreement_moves_refused_closure(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/closure.a")
}
func TestOwnershipAgreement_moves_refused_cross_block(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/cross_block.a")
}
func TestOwnershipAgreement_moves_refused_default_parameter(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/default_parameter.a")
}
func TestOwnershipAgreement_moves_refused_escape(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/escape.a")
}
func TestOwnershipAgreement_moves_refused_expression_work(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/expression_work.a")
}
func TestOwnershipAgreement_moves_refused_field_store(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/field_store.a")
}
func TestOwnershipAgreement_moves_refused_global(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/global.a")
}
func TestOwnershipAgreement_moves_refused_helper(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/helper.a")
}
func TestOwnershipAgreement_moves_refused_item_destructure(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/item_destructure.a")
}
func TestOwnershipAgreement_moves_refused_let_owner(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/let_owner.a")
}
func TestOwnershipAgreement_moves_refused_main_block(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/main_block.a")
}
func TestOwnershipAgreement_moves_refused_nested(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/nested.a")
}
func TestOwnershipAgreement_moves_refused_nested_arrays(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/nested_arrays.a")
}
func TestOwnershipAgreement_moves_refused_nested_call(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/nested_call.a")
}
func TestOwnershipAgreement_moves_refused_nested_map(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/nested_map.a")
}
func TestOwnershipAgreement_moves_refused_nested_race(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/nested_race.a")
}
func TestOwnershipAgreement_moves_refused_overlap(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/overlap.a")
}
func TestOwnershipAgreement_moves_refused_race_alias(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/race_alias.a")
}
func TestOwnershipAgreement_moves_refused_rebind(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/rebind.a")
}
func TestOwnershipAgreement_moves_refused_result(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/result.a")
}
func TestOwnershipAgreement_moves_refused_spread(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/spread.a")
}
func TestOwnershipAgreement_moves_refused_unknown(test *testing.T) {
	test.Parallel()
	ownershipAgreement(test, "../oracle/testdata/moves/refused/unknown.a")
}

func ownershipObjectsAgreementProduct(test *testing.T) string {
	test.Helper()
	inputs := buildcache.Inputs{Name: "ownership-query-objects-agreement", Files: append(
		compilerinputs.Lowering(),
		"internal/lower/ownership_query_test.go",
		"internal/oracle/testdata/moves/accepted/objects.a",
	), Toolchain: []string{runtime.Version()}}
	return buildcache.Product(test, inputs, func(directory string) error {
		ownershipAgreement(test, "../oracle/testdata/moves/accepted/objects.a")
		return os.WriteFile(
			filepath.Join(directory, "agreement"),
			[]byte("admission and extracted query agree\n"),
			0600,
		)
	})
}
func TestProduct_OwnershipObjectsAgreement(test *testing.T) {
	test.Parallel()
	ownershipObjectsAgreementProduct(test)
}
func TestOwnershipObjectsAgreementRecorded(test *testing.T) {
	test.Parallel()
	directory := ownershipObjectsAgreementProduct(test)
	data, readError := os.ReadFile(filepath.Join(directory, "agreement"))
	if readError != nil || string(data) != "admission and extracted query agree\n" {
		test.Fatalf("agreement evidence: %v %q", readError, data)
	}
}

func TestOwnershipQueryDemandDoesNotCrossSites(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "sites.a")
	source := "import { parallelMap } from 'adamic'; const objects = [{value: 1}]; const numbers = [1]; parallelMap(objects, item => item); parallelMap(numbers, item => item);"
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	entry := loaded.Files()[0]
	checker, release := loaded.Checker(context.Background(), entry)
	defer release()
	lowering := &lowering{program: loaded, checker: checker}
	var calls []*ast.Node
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if node.Kind == ast.KindCallExpression &&
			ast.IsIdentifier(node.AsCallExpression().Expression) &&
			node.AsCallExpression().Expression.Text() == "parallelMap" {
			calls = append(calls, node)
		}
		return node.ForEachChild(visit)
	}
	entry.AsNode().ForEachChild(visit)
	if len(calls) != 2 {
		test.Fatalf("calls: %d", len(calls))
	}
	if !lowering.parallelMoves(calls[0]) || lowering.parallelMoves(calls[1]) || len(lowering.parallelMoveDemand) != 2 {
		test.Fatal("different sites shared their demand answer")
	}
	lowering.checker = nil
	if !lowering.parallelMoves(calls[0]) || lowering.parallelMoves(calls[1]) {
		test.Fatal("cached site answers changed")
	}
}
