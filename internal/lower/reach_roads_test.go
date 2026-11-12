package lower

import (
	"slices"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// ReachingAllocations' roads through optional arguments and absent values (reach_roads.mutants.md
// names the one-line change that turns each back): an omitted argument to an optional or defaulted
// parameter is undefined, undefined and null add no site, and a non-null assertion hands on its value.
// Each pin lowers a small TypeScript program and asks the flow graph which object literals a
// module-level variable can hold, by the literal's tag. tsc's ts.sys (src/compiler/sys.ts:1462), a
// let declared undefined and filled from getNodeSystem's literal, then read through `sys!`, is the
// shape they're for (#p2n17c4).

// reachProgram lowers source as main.ts and numbers its allocation sites.
func reachProgram(t *testing.T, source string) *ir.Program {
	t.Helper()
	program, err := lowerSourceAs(t, source, ".ts")
	if err != nil {
		t.Fatalf("lowering: %v", err)
	}
	assignAllocationSites(program)
	return program
}

// reachOfGlobal is what the flow graph says the module-level variable named name can hold.
func reachOfGlobal(t *testing.T, program *ir.Program, name string) ir.AllocationSet {
	t.Helper()
	local := -1
	for index, declared := range program.Locals {
		if declared.Name != name || !declared.Global {
			continue
		}
		if local >= 0 {
			t.Fatalf("two module-level variables are named %s", name)
		}
		local = index
	}
	if local < 0 {
		t.Fatalf("no module-level variable is named %s", name)
	}
	return newAllocationFlowGraph(program).ReachingAllocations(ir.Read{Local: local, Of: program.Locals[local].Type})
}

// literalSite is the allocation site of the one object literal whose tag field is the string given.
func literalSite(t *testing.T, program *ir.Program, tag string) int {
	t.Helper()
	sites := []int{}
	visit := func(node any) bool {
		literal, isLiteral := node.(ir.ObjectLiteral)
		if !isLiteral {
			return true
		}
		for _, field := range literal.Fields {
			if field.Name != "tag" {
				continue
			}
			walk(field.Value, func(inner any) bool {
				if constant, isString := inner.(ir.StringConstant); isString && program.Strings[constant.Index] == tag && len(literal.GraphTypes) > 0 {
					sites = append(sites, literal.GraphTypes[len(literal.GraphTypes)-1])
				}
				return true
			})
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	if len(sites) != 1 || sites[0] >= 0 {
		t.Fatalf("the literal tagged %s has sites %v, want one numbered site", tag, sites)
	}
	return sites[0]
}

// wantSites fails unless reach is known and is exactly the sites given.
func wantSites(t *testing.T, reach ir.AllocationSet, want ...int) {
	t.Helper()
	slices.Sort(want)
	if reach.Unknown || !slices.Equal(reach.Sites, want) {
		t.Fatalf("reach is %+v, want exactly the sites %v", reach, want)
	}
}

// wantUnknown fails unless reach is Unknown, for each of the reasons given.
func wantUnknown(t *testing.T, reach ir.AllocationSet, reasons ...string) {
	t.Helper()
	if !reach.Unknown {
		t.Fatalf("reach is %+v, want Unknown", reach)
	}
	for _, reason := range reasons {
		if !slices.Contains(reach.Reasons, reason) {
			t.Fatalf("reach is %+v, want Unknown for %q", reach, reason)
		}
	}
}

const reachBox = "interface Box { tag: string }\n"

// Road 1: a call that omits a defaulted argument hands the default in, and one that passes it hands
// in what it passes, so the parameter is both literals and nothing else. Mutants 1a and 1b.
func TestReachOmittedArgumentTakesTheDefault(t *testing.T) {
	t.Parallel()
	program := reachProgram(t, reachBox+`function make(box: Box = { tag: 'fallback' }): Box {
	return box;
}
const passed = make({ tag: 'passed' });
const omitted = make();
console.log(passed.tag + omitted.tag);
`)
	wantSites(t, reachOfGlobal(t, program, "omitted"), literalSite(t, program, "fallback"), literalSite(t, program, "passed"))
}

// Road 1: an omitted optional argument is undefined, which adds no site. Mutant 1a.
func TestReachOmittedOptionalArgumentIsUndefined(t *testing.T) {
	t.Parallel()
	program := reachProgram(t, reachBox+`function wrap(box?: Box): Box | undefined {
	return box;
}
const kept = wrap({ tag: 'given' });
const none = wrap();
if (kept !== undefined) {
	console.log(kept.tag);
}
if (none === undefined) {
	console.log('none');
}
`)
	wantSites(t, reachOfGlobal(t, program, "kept"), literalSite(t, program, "given"))
}

// Road 1's negatives, in hand-built IR, since TypeScript lets no direct call omit a required
// argument: make(first, box) returns box, one call passes a literal and one omits box. An omitted
// required argument stays unknown, and so does an omitted optional one where a spread may fill it.
// The optional row is the positive the other two are measured against. Mutants 1a, 1c and 1d.
func TestReachOmittedArgumentStaysUnknownUnlessOptional(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct {
		name             string
		optional, spread bool
		known            bool
	}{
		{name: "optional", optional: true, known: true},
		{name: "required", optional: false},
		{name: "optional under a spread", optional: true, spread: true},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			omitting := ir.Call{Function: 0, Arguments: []ir.Expression{ir.Undefined{Of: ir.Object}}, Returns: ir.Object}
			if probe.spread {
				omitting = ir.Call{Function: 0, Arguments: []ir.Expression{ir.Read{Local: 3, Of: ir.Array}}, Spread: []bool{true}, Returns: ir.Object}
			}
			program := &ir.Program{
				Locals: []ir.Local{
					{Name: "first", Type: ir.Object, Function: 0},
					{Name: "box", Type: ir.Object, Function: 0},
					{Name: "made", Type: ir.Object, Global: true, Function: -1},
					{Name: "items", Type: ir.Array, Global: true, Function: -1},
				},
				Functions: []ir.Function{{
					Name:               "make",
					Parameters:         []int{0, 1},
					OptionalParameters: ir.SortedMap[int, bool]{1: probe.optional},
					Returns:            ir.Object,
					Body:               []ir.Statement{ir.Return{Value: ir.Read{Local: 1, Of: ir.Object}}},
				}},
				Main: []ir.Statement{
					ir.Declare{Local: 3, Value: ir.ArrayLiteral{GraphTypes: []int{-2}}},
					ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.Undefined{Of: ir.Object}, ir.ObjectLiteral{GraphTypes: []int{-1}}}, Returns: ir.Object}},
					ir.Declare{Local: 2, Value: omitting},
				},
			}
			reach := newAllocationFlowGraph(program).ReachingAllocations(ir.Read{Local: 2, Of: ir.Object})
			if probe.known {
				wantSites(t, reach, -1)
			} else {
				wantUnknown(t, reach, "parameter may receive untracked or omitted arguments")
			}
		})
	}
}

// Road 2: a variable declared without a value is undefined until a call fills it, so it holds the
// one literal written and nothing else. Mutant 2.
func TestReachUndefinedAddsNoSite(t *testing.T) {
	t.Parallel()
	program := reachProgram(t, reachBox+`let held: Box | undefined;
function fill(): void {
	held = { tag: 'filled' };
}
function read(): Box | undefined {
	return held;
}
fill();
const got = read();
console.log(got === undefined ? 'empty' : got.tag);
`)
	wantSites(t, reachOfGlobal(t, program, "held"), literalSite(t, program, "filled"))
}

// Road 2's negative: a variable that is only ever undefined reaches no allocation, which is Unknown,
// never an empty proof. Mutant 2n.
func TestReachOnlyUndefinedStaysUnknown(t *testing.T) {
	t.Parallel()
	program := reachProgram(t, reachBox+`let empty: Box | undefined;
function read(): Box | undefined {
	return empty;
}
console.log(read() === undefined ? 'empty' : 'held');
`)
	reach := reachOfGlobal(t, program, "empty")
	if len(reach.Sites) != 0 {
		t.Fatalf("reach is %+v, want no site", reach)
	}
	wantUnknown(t, reach, "no reaching allocation established")
}

// Road 4: x! is x ?? panic, which hands on x or stops, so it reaches what x does. Mutant 4.
func TestReachNonNullAssertionHandsOnItsValue(t *testing.T) {
	t.Parallel()
	program := reachProgram(t, reachBox+`function first(): Box | undefined {
	return { tag: 'asserted' };
}
const sure = first()!;
console.log(sure.tag);
`)
	wantSites(t, reachOfGlobal(t, program, "sure"), literalSite(t, program, "asserted"))
}

// Road 4's negative: an assertion makes no allocation of its own, so one over a value that's only
// ever undefined still reaches none. Mutant 2n.
func TestReachNonNullAssertionOfOnlyUndefinedStaysUnknown(t *testing.T) {
	t.Parallel()
	program := reachProgram(t, reachBox+`let empty: Box | undefined;
function force(): Box {
	return empty!;
}
const forced = force();
console.log(forced.tag);
`)
	reach := reachOfGlobal(t, program, "forced")
	if len(reach.Sites) != 0 {
		t.Fatalf("reach is %+v, want no site", reach)
	}
	wantUnknown(t, reach, "no reaching allocation established")
}
