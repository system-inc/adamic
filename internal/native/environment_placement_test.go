package native

import (
	"context"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"github.com/system-inc/adamic/internal/lower"
)

func environmentPlacementCase(test *testing.T, name string, want environmentPlacement) {
	test.Helper()
	source, loadError := load.Load([]string{"../oracle/testdata/environment_" + name + ".a"})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), source)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	plan := planRegions(program)
	sites := 0
	for _, function := range program.Functions {
		for index := range function.Body {
			if _, ok := function.Body[index].(ir.AllocateEnvironment); !ok {
				continue
			}
			sites++
			if got := plan.environments[&function.Body[index]]; got != want {
				test.Errorf("%s placement %d, want %d", function.Name, got, want)
			}
		}
	}
	if sites == 0 {
		test.Fatal("fixture has no explicit environment site")
	}
}
func TestEnvironmentPlacementArray(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "array", environmentHeap)
}
func TestEnvironmentPlacementCallbackEscape(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "callback_escape", environmentHeap)
}
func TestEnvironmentPlacementCallbacks(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "callbacks", environmentStack)
}
func TestEnvironmentPlacementCapture(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "capture", environmentHeap)
}
func TestEnvironmentPlacementCapturedParameters(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "captured_parameters", environmentStack)
}
func TestEnvironmentPlacementCoverageAliasJoin(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_alias_join", environmentHeap)
}
func TestEnvironmentPlacementCoverageBoundary64(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_boundary64", environmentStack)
}
func TestEnvironmentPlacementCoverageBoundedCall(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_bounded_call", environmentStack)
}
func TestEnvironmentPlacementCoverageCallbackExits(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_callback_exits", environmentStack)
}
func TestEnvironmentPlacementCoverageCallbackResult(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_callback_result", environmentHeap)
}
func TestEnvironmentPlacementCoverageCallbacksExtra(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_callbacks_extra", environmentStack)
}
func TestEnvironmentPlacementCoverageCapturedParameter(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_captured_parameter", environmentHeap)
}
func TestEnvironmentPlacementCoverageClosureSelf(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_closure_self", environmentHeap)
}
func TestEnvironmentPlacementCoverageCollections(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_collections", environmentStack)
}
func TestEnvironmentPlacementCoverageInitializerThrow(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_initializer_throw", environmentStack)
}
func TestEnvironmentPlacementCoverageLocalContainers(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_local_containers", environmentHeap)
}
func TestEnvironmentPlacementCoveragePrimitives(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_primitives", environmentStack)
}
func TestEnvironmentPlacementCoverageRegionValues(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_region_values", environmentRegion)
}
func TestEnvironmentPlacementCoverageReturnedValues(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_returned_values", environmentStack)
}
func TestEnvironmentPlacementCoverageValuesExits(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "coverage_values_exits", environmentStack)
}
func TestEnvironmentPlacementDirect(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "direct", environmentStack)
}
func TestEnvironmentPlacementExits(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "exits", environmentStack)
}
func TestEnvironmentPlacementField(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "field", environmentHeap)
}
func TestEnvironmentPlacementGlobal(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "global", environmentHeap)
}
func TestEnvironmentPlacementKeepingCall(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "keeping_call", environmentHeap)
}
func TestEnvironmentPlacementLarge(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "large", environmentRegion)
}
func TestEnvironmentPlacementLocalCall(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "local_call", environmentStack)
}
func TestEnvironmentPlacementLoop(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "loop", environmentStack)
}
func TestEnvironmentPlacementMap(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "map", environmentHeap)
}
func TestEnvironmentPlacementReturned(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "returned", environmentHeap)
}
func TestEnvironmentPlacementSet(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "set", environmentHeap)
}
func TestEnvironmentPlacementSiblings(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "siblings", environmentStack)
}
func TestEnvironmentPlacementUnknownCall(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "unknown_call", environmentHeap)
}
func TestEnvironmentPlacementUnknownCallback(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "unknown_callback", environmentHeap)
}
func TestEnvironmentPlacementVirtualCall(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "virtual_call", environmentHeap)
}

// A thrown carrier is an escape even when source throws only Error objects.
func TestThrownEnvironmentStaysOnHeap(test *testing.T) {
	test.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Number, Captured: true, EnvironmentCell: true}},
		Functions: []ir.Function{
			{FrameEnvironment: []int{0}, Body: []ir.Statement{
				ir.AllocateEnvironment{Cells: []int{0}},
				ir.Throw{Value: ir.MakeClosure{Function: 1}},
			}},
			{Closure: true, Environment: []int{0}},
		},
	}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		test.Fatal("throw placed its carrier in the frame")
	}
}

func TestEnvironmentParameterAllTargets(test *testing.T) {
	test.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Closure}, {Type: ir.Closure}, {Type: ir.Closure, Global: true}},
		Functions: []ir.Function{
			{Parameters: []int{0}, Body: []ir.Statement{
				ir.Evaluate{Value: ir.CallClosure{Closure: ir.Read{Local: 0, Of: ir.Closure}, Returns: ir.String}},
			}},
			{Parameters: []int{1}, Body: []ir.Statement{ir.Assign{Local: 2, Value: ir.Read{Local: 1, Of: ir.Closure}}}},
		},
	}
	plan := planRegions(program)
	if plan.environmentCallEscapes(ir.FunctionTargets{Functions: []int{0}}, 0) {
		test.Error("synchronous helper retains its callback")
	}
	if !plan.environmentCallEscapes(ir.FunctionTargets{Functions: []int{0, 1}}, 0) {
		test.Error("keeping override ignored")
	}
	if !plan.environmentCallEscapes(ir.FunctionTargets{Unknown: true}, 0) {
		test.Error("unknown target treated as synchronous")
	}
}

// An operation added later (for example a pool handoff or async frame transfer)
// must get a lifetime proof before it can borrow an environment carrier.
type unknownEnvironmentTransfer struct{ Value ir.Expression }

func (unknownEnvironmentTransfer) Type() ir.Type { return ir.Closure }

func TestUnknownEnvironmentTransferStaysOnHeap(test *testing.T) {
	test.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Number, Captured: true, EnvironmentCell: true}},
		Functions: []ir.Function{
			{FrameEnvironment: []int{0}, Body: []ir.Statement{
				ir.AllocateEnvironment{Cells: []int{0}},
				ir.Evaluate{Value: unknownEnvironmentTransfer{Value: ir.MakeClosure{Function: 1}}},
			}},
			{Closure: true, Environment: []int{0}},
		},
	}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		test.Fatal("unknown handoff placed its carrier in the frame")
	}
}

func TestAsyncEnvironmentStaysCounted(test *testing.T) {
	test.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.String, Function: 0, Captured: true, EnvironmentCell: true}},
		Functions: []ir.Function{
			{Async: true, FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}}},
		},
	}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		test.Fatal("async cells placed in call storage")
	}
}

func asyncCarrierAndArgumentsEscape(test *testing.T, callback bool) {
	test.Helper()
	program := &ir.Program{
		Locals: []ir.Local{
			{Type: ir.String, Function: 0, Captured: true, EnvironmentCell: true},
			{Type: ir.Closure, Function: 1},
		},
		Functions: []ir.Function{
			{FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}}},
			{Async: true, Closure: true, Parameters: []int{1}, Environment: []int{0}},
		},
	}
	var call ir.Expression = ir.CallClosure{Closure: ir.MakeClosure{Function: 1}, Returns: ir.Promise}
	if callback {
		call = ir.ArrayVisit{Array: ir.ArrayLiteral{Element: ir.Number}, Callback: ir.MakeClosure{Function: 1}}
	}
	program.Functions[0].Body = append(program.Functions[0].Body, ir.Evaluate{Value: call})
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		test.Fatal("async carrier placed in call storage")
	}
	if !plan.environmentCallEscapes(ir.FunctionTargets{Functions: []int{1}}, 0) {
		test.Fatal("async argument treated as borrowed")
	}
}
func TestAsyncCarrierAndArgumentsEscape(test *testing.T) {
	test.Parallel()
	asyncCarrierAndArgumentsEscape(test, false)
}
func TestAsyncCallbackAndArgumentsEscape(test *testing.T) {
	test.Parallel()
	asyncCarrierAndArgumentsEscape(test, true)
}

func TestEnvironmentPlacementCapturedParameterEscape(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "captured_parameter_escape", environmentHeap)
}
func TestEnvironmentPlacementCapturedParameterReassign(test *testing.T) {
	test.Parallel()
	environmentPlacementCase(test, "captured_parameter_reassign", environmentStack)
}
func TestEnvironmentCapturedParameterHasOneOwner(test *testing.T) {
	test.Parallel()
	source, loadError := load.Load([]string{"../oracle/testdata/environment_captured_parameter_escape.a"})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), source)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	code := C(program)
	found := false
	for _, function := range program.Functions {
		for _, parameter := range function.Parameters {
			local := program.Locals[parameter]
			if !local.EnvironmentCell || !local.Type.IsReference() {
				continue
			}
			found = true
			name := (&emitter{program: program}).localName(parameter)
			if local.Borrowed ||
				strings.Count(code, "adamic_retain("+name+")") != 1 ||
				strings.Contains(code, "adamic_release("+name+");") {
				test.Fatalf("captured parameter %s must have one owning cell and no raw entry count", local.Name)
			}
		}
	}
	if !found {
		test.Fatal("fixture lost owned captured input")
	}
}

func environmentMemberStaysHeap(test *testing.T, programMember bool) {
	test.Helper()
	local := ir.Local{
		Type:            ir.String,
		Captured:        true,
		EnvironmentCell: true,
		ProgramRegion:   programMember,
		GraphCell:       !programMember,
	}
	program := &ir.Program{Locals: []ir.Local{local}, Functions: []ir.Function{
		{FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}}},
	}}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		test.Fatal("graph/Program member placed in call storage")
	}
}
func TestEnvironmentGraphMemberStaysHeap(test *testing.T) {
	test.Parallel()
	environmentMemberStaysHeap(test, false)
}
func TestEnvironmentProgramMemberStaysHeap(test *testing.T) {
	test.Parallel()
	environmentMemberStaysHeap(test, true)
}
func TestEnvironmentUnknownUnionAsyncCarrier(test *testing.T) {
	test.Parallel()
	program := &ir.Program{Locals: []ir.Local{
		{Type: ir.String, Function: 0, Captured: true, EnvironmentCell: true},
		{Type: ir.Closure, Function: 0},
	}, Functions: []ir.Function{
		{FrameEnvironment: []int{0}, Body: []ir.Statement{
			ir.AllocateEnvironment{Cells: []int{0}},
			ir.Declare{Local: 1, Value: ir.MakeClosure{Function: 1}},
			ir.Evaluate{Value: ir.CallClosure{Closure: ir.Read{Local: 1, Of: ir.Closure}, Returns: ir.Union}},
		}},
		{Async: true, Closure: true, Environment: []int{0}},
	}}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		test.Fatal("unknown Union callback can retain its carrier in an async frame")
	}
}
func TestEnvironmentAdapterScopeNeedsNoRegionPlan(test *testing.T) {
	test.Parallel()
	adapter := emitter{scopes: [][]string{{"held"}}}
	adapter.releaseScopes(0)
	if adapter.out.String() != "adamic_release(held);\n" {
		test.Fatal(adapter.out.String())
	}
}

// Even a currently non-keeping override stays outside the call-storage proof.
func TestEnvironmentVirtualBorrowStaysCounted(test *testing.T) {
	test.Parallel()
	path := filepath.Join(test.TempDir(), "virtual-borrow.a")
	const source = `class User { use(read: () => string): string { return read(); } }
class Other extends User { use(read: () => string): string { return read(); } }
function fill(user: User, start: number): string {
 const text = 'scratch' + String(start);
 function read(): string { return text; }
 return user.use(read);
}
console.log(fill(new Other(), 3));`
	if writeError := os.WriteFile(path, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	loaded, loadError := load.Load([]string{path})
	if loadError != nil {
		test.Fatal(loadError)
	}
	program, lowerError := lower.Lower(context.Background(), loaded)
	if lowerError != nil {
		test.Fatal(lowerError)
	}
	plan := planRegions(program)
	virtual, sites := 0, 0
	var scan func([]ir.Statement)
	scan = func(body []ir.Statement) {
		for _, statement := range body {
			walkStatement(statement, func(e ir.Expression) {
				if c, ok := e.(ir.Call); ok && c.Virtual != 0 {
					virtual++
				}
			}, scan)
		}
	}
	for _, f := range program.Functions {
		scan(f.Body)
		if f.Name != "fill" {
			continue
		}
		for i := range f.Body {
			if _, ok := f.Body[i].(ir.AllocateEnvironment); !ok {
				continue
			}
			sites++
			if got := plan.environments[&f.Body[i]]; got != environmentHeap {
				test.Errorf("virtual callback carrier placement %d, want counted heap", got)
			}
		}
	}
	if virtual == 0 || sites == 0 {
		test.Fatalf("probe lost virtual dispatch or environment: virtual=%d sites=%d", virtual, sites)
	}
}

// A nested generator's call keeps its carrier in the frame it makes (#qk8rztp S3a), so a parent
// whose one escape is that call has its environment counted on the heap, never in its own C frame
// (#s6tt5xt, Memory's seam): the make-frame call is a CallClosure direct to a function that
// suspends, which environmentCalleeEscapes counts. The control is the same parent with an ordinary
// nested function in the generator's place, which the proof keeps in call storage, so the fixture's
// parent is one it would otherwise have put on the stack. Mutant: environmentCalleeEscapes reads
// Async where it reads Suspends(); the generator's parent reads stack. Only this test sees that
// mutant: every S3a consumer steps the generator inside its parent's call, so the frame lets go of
// its carrier before a C frame holding the cells could end, and nested_stack_parent.a runs clean
// either way under the sanitizers. The placement is what keeps it sound once a generator can be held
// past its parent's return (S2c-4).
func TestEnvironmentPlacementNestedGenerator(test *testing.T) {
	test.Parallel()
	placements := func(path string) []environmentPlacement {
		test.Helper()
		loaded, loadError := load.Load([]string{path})
		if loadError != nil {
			test.Fatal(loadError)
		}
		program, lowerError := lower.Lower(context.Background(), loaded)
		if lowerError != nil {
			test.Fatal(lowerError)
		}
		plan := planRegions(program)
		var found []environmentPlacement
		for _, function := range program.Functions {
			if function.Name != "host" {
				continue
			}
			for index := range function.Body {
				if _, ok := function.Body[index].(ir.AllocateEnvironment); ok {
					found = append(found, plan.environments[&function.Body[index]])
				}
			}
		}
		if len(found) != 1 {
			test.Fatalf("%s: %d environment sites in host, want 1", path, len(found))
		}
		return found
	}
	if got := placements("../oracle/testdata/generators/nested_stack_parent.a"); got[0] != environmentHeap {
		test.Errorf("the nested generator's parent placed %d, want counted heap", got[0])
	}
	// A generator written in place reading the parent's parameters (#qk8rztp S3b): its carrier is a
	// closure of its own, made for a call direct to it, which suspends, so the same rule keeps the
	// environment on the heap.
	if got := placements("../oracle/testdata/generators/in_place_environment.a"); got[0] != environmentHeap {
		test.Errorf("the parent of a generator written in place placed %d, want counted heap", got[0])
	}
	control := filepath.Join(test.TempDir(), "nested-function.a")
	const source = "function host(first: string, second: string): string {\n" +
		"    const pair = `${first}+${second}`;\n" +
		"    function both(): string {\n" +
		"        return `${pair}<1${pair}<2`;\n" +
		"    }\n" +
		"    return both();\n" +
		"}\n" +
		"console.log(host(`a${String(1)}`, `b${String(2)}`));\n"
	if writeError := os.WriteFile(control, []byte(source), 0600); writeError != nil {
		test.Fatal(writeError)
	}
	if got := placements(control); got[0] != environmentStack {
		test.Errorf(
			"the control, an ordinary nested function's parent, placed %d, want stack: the fixture's parent proves nothing",
			got[0],
		)
	}
}
