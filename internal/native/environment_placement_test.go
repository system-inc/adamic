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

func environmentPlacementCase(t *testing.T, name string, want environmentPlacement) {
	t.Helper()
	source, err := load.Load([]string{"../oracle/testdata/environment_" + name + ".a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), source)
	if err != nil {
		t.Fatal(err)
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
				t.Errorf("%s placement %d, want %d", function.Name, got, want)
			}
		}
	}
	if sites == 0 {
		t.Fatal("fixture has no explicit environment site")
	}
}
func TestEnvironmentPlacementArray(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "array", environmentHeap)
}
func TestEnvironmentPlacementCallbackEscape(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "callback_escape", environmentHeap)
}
func TestEnvironmentPlacementCallbacks(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "callbacks", environmentStack)
}
func TestEnvironmentPlacementCapture(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "capture", environmentHeap)
}
func TestEnvironmentPlacementCapturedParameters(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "captured_parameters", environmentStack)
}
func TestEnvironmentPlacementCoverageAliasJoin(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_alias_join", environmentHeap)
}
func TestEnvironmentPlacementCoverageBoundary64(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_boundary64", environmentStack)
}
func TestEnvironmentPlacementCoverageBoundedCall(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_bounded_call", environmentStack)
}
func TestEnvironmentPlacementCoverageCallbackExits(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_callback_exits", environmentStack)
}
func TestEnvironmentPlacementCoverageCallbackResult(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_callback_result", environmentHeap)
}
func TestEnvironmentPlacementCoverageCallbacksExtra(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_callbacks_extra", environmentStack)
}
func TestEnvironmentPlacementCoverageCapturedParameter(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_captured_parameter", environmentHeap)
}
func TestEnvironmentPlacementCoverageClosureSelf(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_closure_self", environmentHeap)
}
func TestEnvironmentPlacementCoverageCollections(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_collections", environmentStack)
}
func TestEnvironmentPlacementCoverageInitializerThrow(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_initializer_throw", environmentStack)
}
func TestEnvironmentPlacementCoverageLocalContainers(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_local_containers", environmentHeap)
}
func TestEnvironmentPlacementCoveragePrimitives(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_primitives", environmentStack)
}
func TestEnvironmentPlacementCoverageRegionValues(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_region_values", environmentRegion)
}
func TestEnvironmentPlacementCoverageReturnedValues(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_returned_values", environmentStack)
}
func TestEnvironmentPlacementCoverageValuesExits(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "coverage_values_exits", environmentStack)
}
func TestEnvironmentPlacementDirect(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "direct", environmentStack)
}
func TestEnvironmentPlacementExits(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "exits", environmentStack)
}
func TestEnvironmentPlacementField(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "field", environmentHeap)
}
func TestEnvironmentPlacementGlobal(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "global", environmentHeap)
}
func TestEnvironmentPlacementKeepingCall(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "keeping_call", environmentHeap)
}
func TestEnvironmentPlacementLarge(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "large", environmentRegion)
}
func TestEnvironmentPlacementLocalCall(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "local_call", environmentStack)
}
func TestEnvironmentPlacementLoop(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "loop", environmentStack)
}
func TestEnvironmentPlacementMap(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "map", environmentHeap)
}
func TestEnvironmentPlacementReturned(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "returned", environmentHeap)
}
func TestEnvironmentPlacementSet(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "set", environmentHeap)
}
func TestEnvironmentPlacementSiblings(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "siblings", environmentStack)
}
func TestEnvironmentPlacementUnknownCall(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "unknown_call", environmentHeap)
}
func TestEnvironmentPlacementUnknownCallback(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "unknown_callback", environmentHeap)
}
func TestEnvironmentPlacementVirtualCall(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "virtual_call", environmentHeap)
}

// A thrown carrier is an escape even when source throws only Error objects.
func TestThrownEnvironmentStaysOnHeap(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Number, Captured: true, EnvironmentCell: true}},
		Functions: []ir.Function{
			{FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}, ir.Throw{Value: ir.MakeClosure{Function: 1}}}},
			{Closure: true, Environment: []int{0}},
		},
	}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		t.Fatal("throw placed its carrier in the frame")
	}
}

func TestEnvironmentParameterAllTargets(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Closure}, {Type: ir.Closure}, {Type: ir.Closure, Global: true}},
		Functions: []ir.Function{
			{Parameters: []int{0}, Body: []ir.Statement{ir.Evaluate{Value: ir.CallClosure{Closure: ir.Read{Local: 0, Of: ir.Closure}, Returns: ir.String}}}},
			{Parameters: []int{1}, Body: []ir.Statement{ir.Assign{Local: 2, Value: ir.Read{Local: 1, Of: ir.Closure}}}},
		},
	}
	plan := planRegions(program)
	if plan.environmentCallEscapes(ir.FunctionTargets{Functions: []int{0}}, 0) {
		t.Error("synchronous helper retains its callback")
	}
	if !plan.environmentCallEscapes(ir.FunctionTargets{Functions: []int{0, 1}}, 0) {
		t.Error("keeping override ignored")
	}
	if !plan.environmentCallEscapes(ir.FunctionTargets{Unknown: true}, 0) {
		t.Error("unknown target treated as synchronous")
	}
}

// An operation added later (for example a pool handoff or async frame transfer)
// must get a lifetime proof before it can borrow an environment carrier.
type unknownEnvironmentTransfer struct{ Value ir.Expression }

func (unknownEnvironmentTransfer) Type() ir.Type { return ir.Closure }

func TestUnknownEnvironmentTransferStaysOnHeap(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.Number, Captured: true, EnvironmentCell: true}},
		Functions: []ir.Function{
			{FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}, ir.Evaluate{Value: unknownEnvironmentTransfer{Value: ir.MakeClosure{Function: 1}}}}},
			{Closure: true, Environment: []int{0}},
		},
	}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		t.Fatal("unknown handoff placed its carrier in the frame")
	}
}

func TestAsyncEnvironmentStaysCounted(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.String, Function: 0, Captured: true, EnvironmentCell: true}},
		Functions: []ir.Function{
			{Async: true, FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}}},
		},
	}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		t.Fatal("async cells placed in call storage")
	}
}

func asyncCarrierAndArgumentsEscape(t *testing.T, callback bool) {
	t.Helper()
	program := &ir.Program{
		Locals: []ir.Local{{Type: ir.String, Function: 0, Captured: true, EnvironmentCell: true}, {Type: ir.Closure, Function: 1}},
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
		t.Fatal("async carrier placed in call storage")
	}
	if !plan.environmentCallEscapes(ir.FunctionTargets{Functions: []int{1}}, 0) {
		t.Fatal("async argument treated as borrowed")
	}
}
func TestAsyncCarrierAndArgumentsEscape(t *testing.T) {
	t.Parallel()
	asyncCarrierAndArgumentsEscape(t, false)
}
func TestAsyncCallbackAndArgumentsEscape(t *testing.T) {
	t.Parallel()
	asyncCarrierAndArgumentsEscape(t, true)
}

func TestEnvironmentPlacementCapturedParameterEscape(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "captured_parameter_escape", environmentHeap)
}
func TestEnvironmentPlacementCapturedParameterReassign(t *testing.T) {
	t.Parallel()
	environmentPlacementCase(t, "captured_parameter_reassign", environmentStack)
}
func TestEnvironmentCapturedParameterHasOneOwner(t *testing.T) {
	t.Parallel()
	source, err := load.Load([]string{"../oracle/testdata/environment_captured_parameter_escape.a"})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), source)
	if err != nil {
		t.Fatal(err)
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
			if local.Borrowed || strings.Count(code, "adamic_retain("+name+")") != 1 || strings.Contains(code, "adamic_release("+name+");") {
				t.Fatalf("captured parameter %s must have one owning cell and no raw entry count", local.Name)
			}
		}
	}
	if !found {
		t.Fatal("fixture lost owned captured input")
	}
}

func environmentMemberStaysHeap(t *testing.T, programMember bool) {
	t.Helper()
	local := ir.Local{Type: ir.String, Captured: true, EnvironmentCell: true, ProgramRegion: programMember, GraphCell: !programMember}
	program := &ir.Program{Locals: []ir.Local{local}, Functions: []ir.Function{{FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}}}}}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		t.Fatal("graph/Program member placed in call storage")
	}
}
func TestEnvironmentGraphMemberStaysHeap(t *testing.T) {
	t.Parallel()
	environmentMemberStaysHeap(t, false)
}
func TestEnvironmentProgramMemberStaysHeap(t *testing.T) {
	t.Parallel()
	environmentMemberStaysHeap(t, true)
}
func TestEnvironmentUnknownUnionAsyncCarrier(t *testing.T) {
	t.Parallel()
	program := &ir.Program{Locals: []ir.Local{{Type: ir.String, Function: 0, Captured: true, EnvironmentCell: true}, {Type: ir.Closure, Function: 0}}, Functions: []ir.Function{
		{FrameEnvironment: []int{0}, Body: []ir.Statement{ir.AllocateEnvironment{Cells: []int{0}}, ir.Declare{Local: 1, Value: ir.MakeClosure{Function: 1}}, ir.Evaluate{Value: ir.CallClosure{Closure: ir.Read{Local: 1, Of: ir.Closure}, Returns: ir.Union}}}},
		{Async: true, Closure: true, Environment: []int{0}},
	}}
	plan := planRegions(program)
	if plan.environments[&program.Functions[0].Body[0]] != environmentHeap {
		t.Fatal("unknown Union callback can retain its carrier in an async frame")
	}
}
func TestEnvironmentAdapterScopeNeedsNoRegionPlan(t *testing.T) {
	t.Parallel()
	adapter := emitter{scopes: [][]string{{"held"}}}
	adapter.releaseScopes(0)
	if adapter.out.String() != "adamic_release(held);\n" {
		t.Fatal(adapter.out.String())
	}
}

// Even a currently non-keeping override stays outside the call-storage proof.
func TestEnvironmentVirtualBorrowStaysCounted(t *testing.T) {
	t.Parallel()
	path := filepath.Join(t.TempDir(), "virtual-borrow.a")
	const source = `class User { use(read: () => string): string { return read(); } }
class Other extends User { use(read: () => string): string { return read(); } }
function fill(user: User, start: number): string {
 const text = 'scratch' + String(start);
 function read(): string { return text; }
 return user.use(read);
}
console.log(fill(new Other(), 3));`
	if err := os.WriteFile(path, []byte(source), 0600); err != nil {
		t.Fatal(err)
	}
	loaded, err := load.Load([]string{path})
	if err != nil {
		t.Fatal(err)
	}
	program, err := lower.Lower(context.Background(), loaded)
	if err != nil {
		t.Fatal(err)
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
				t.Errorf("virtual callback carrier placement %d, want counted heap", got)
			}
		}
	}
	if virtual == 0 || sites == 0 {
		t.Fatalf("probe lost virtual dispatch or environment: virtual=%d sites=%d", virtual, sites)
	}
}
