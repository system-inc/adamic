package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// ReachingAllocations' roads through optional arguments, absent values and functions that never run
// (reach_roads.mutants.md names the one-line change that turns each back): an omitted argument to an
// optional or defaulted parameter is undefined, undefined and null add no site, a function that never
// runs makes no writes, a non-null assertion and a narrowed read hand on their value, and a cell made
// before its declaration holds nothing until it's assigned. Each pin lowers a small
// TypeScript program and asks the flow graph which object literals a module-level variable can hold,
// by the literal's tag. tsc's ts.sys (src/compiler/sys.ts:1462), a let declared undefined and filled
// from getNodeSystem's literal, then read through `sys!`, is the shape they're for (#p2n17c4), and
// the ts.sys pins near the end hold it and its twin.

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

// reachModules lowers several modules from main.ts and numbers their allocation sites.
func reachModules(t *testing.T, modules map[string]string) *ir.Program {
	t.Helper()
	program, err := lowerModules(t, modules)
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

// Road 5: x || y is x where x is truthy and y otherwise, so it reaches what either does. Mutant 5.
func TestReachLogicalOrIsEitherOperand(t *testing.T) {
	t.Parallel()
	program := reachProgram(t, reachBox+`function first(flag: boolean): Box | undefined {
	return flag ? { tag: 'first' } : undefined;
}
const picked = first(true) || { tag: 'fallback' };
console.log(picked.tag);
`)
	wantSites(t, reachOfGlobal(t, program, "picked"), literalSite(t, program, "first"), literalSite(t, program, "fallback"))
}

// An upcast keeps its value: lowering makes no node for it (cast.go's castValue hands back the
// operand), so the variable reached the literal before these roads too. This pins that `as` needs
// no road of its own; it has no mutant here.
func TestReachUpcastKeepsItsValue(t *testing.T) {
	t.Parallel()
	program := reachProgram(t, `interface Named { tag: string }
interface Labeled extends Named { label: string }
const made: Labeled = { tag: 'labeled', label: 'one' };
const named = made as Named;
console.log(named.tag);
`)
	wantSites(t, reachOfGlobal(t, program, "named"), literalSite(t, program, "labeled"))
}

// irHolds reports whether any statement of the program holds a node matching is.
func irHolds(program *ir.Program, is func(any) bool) bool {
	found := false
	visit := func(node any) bool {
		found = found || is(node)
		return !found
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	return found
}

// Road 6: a read the checker narrowed undefined out of is ir.Defined, checked, so it hands on what it
// reads or stops, and reaches what the read does. It's how ts.sys's arrow hands its let back in .a,
// which refuses tsc's `sys!` (oracle/testdata/tssys_presence_twin/system.a). The pin first holds that
// the narrowing lowered to ir.Defined at all, so it can't pass on some other road. Mutant 12.
func TestReachNarrowedReadHandsOnItsValue(t *testing.T) {
	t.Parallel()
	for _, extension := range []string{".ts", ".a"} {
		t.Run(extension, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSourceAs(t, reachBox+`function first(): Box | undefined {
	return { tag: 'narrowed' };
}
function sure(): Box {
	const box = first();
	if (box === undefined) {
		throw new Error('none');
	}
	return box;
}
const held = sure();
console.log(held.tag);
`, extension)
			if err != nil {
				t.Fatalf("lowering: %v", err)
			}
			assignAllocationSites(program)
			if !irHolds(program, func(node any) bool { _, defined := node.(ir.Defined); return defined }) {
				t.Fatalf("the narrowed return lowered to no ir.Defined, so this pins nothing about road 6")
			}
			wantSites(t, reachOfGlobal(t, program, "held"), literalSite(t, program, "narrowed"))
		})
	}
}

// Road 7: a function holding a nested function declaration gets its variables' cells before its body
// runs, so a hoisted function can capture one (nested_functions.go's preallocateNestedBindings). That
// early declaration holds no value: the variable is what its initializer assigns. It's getNodeSystem's
// shape, whose nodeSystem literal sits beside readFile and directoryExists, declared after its return
// (sys.ts:1468). Each row first holds that the cell was made early, valueless, so it can't pass on
// some other road; the second row's nested function captures the variable. Mutant 13.
func TestReachEarlyCellIsWhatItsDeclarationAssigns(t *testing.T) {
	t.Parallel()
	for _, probe := range []struct{ name, helper string }{
		{"uncaptured", "function helper(): string {\n\t\treturn 'uncaptured';\n\t}"},
		{"captured", "function helper(): string {\n\t\treturn box.tag;\n\t}"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program := reachProgram(t, reachBox+`function make(): Box {
	const box: Box = { tag: 'early' };
	console.log(helper());
	return box;
	`+probe.helper+`
}
const made = make();
console.log(made.tag);
`)
			if !irHolds(program, func(node any) bool {
				declared, isDeclare := node.(ir.Declare)
				return isDeclare && declared.Uninitialized && declared.Value == nil && program.Locals[declared.Local].Name == "box"
			}) {
				t.Fatalf("box's cell wasn't declared early and valueless, so this pins nothing about road 7")
			}
			wantSites(t, reachOfGlobal(t, program, "made"), literalSite(t, program, "early"))
		})
	}
}

// Road 3: a function nothing calls, exports or makes a value of never runs, so its write is no write.
// Mutant 3.
func TestReachUncalledWriterMakesNoWrite(t *testing.T) {
	t.Parallel()
	program := reachProgram(t, reachBox+`let current: Box = { tag: 'first' };
function replace(next: Box): void {
	current = next;
}
console.log(current.tag);
`)
	wantSites(t, reachOfGlobal(t, program, "current"), literalSite(t, program, "first"))
}

// Road 3 reads a field that holds a function's index plus one as that, and one that holds the index
// as that. Each pin is a writer into a module-level let that one field alone names, and wants the
// literal it writes reached. Lowering itself never leaves a const arrow or a nested declaration named
// only by its local's one-based field (its binding is declared with a MakeClosure too), nor a method
// only by its literal's table (a class's Methods name it as well), so each program is built as IR.
// writerProgram is current, declared with literal -1, and function 1, which writes literal -2 into
// it; main is what Main does after the declaration, and locals the locals past current. Function 0
// does nothing and is never the writer: a field read as an index at its zero value names function 0,
// as every local's NestedFunction would under mutant 8, current's included, so a writer at 0 would
// be named by fields that don't name it.
func writerProgram(locals []ir.Local, main ...ir.Statement) *ir.Program {
	return &ir.Program{
		Locals: append([]ir.Local{{Name: "current", Type: ir.Object, Global: true, Function: -1}}, locals...),
		Functions: []ir.Function{
			{Name: "idle"},
			{
				Name:    "replace",
				Closure: true,
				Body:    []ir.Statement{ir.Assign{Local: 0, Value: ir.ObjectLiteral{GraphTypes: []int{-2}}}},
			},
		},
		Main: append([]ir.Statement{ir.Declare{Local: 0, Value: ir.ObjectLiteral{GraphTypes: []int{-1}}}}, main...),
	}
}

// A writer named only by its local's NestedFunction, one-based, and called through the local, as a
// nested declaration is. Mutant 8.
func TestReachWriterNamedOnlyByAOneBasedField(t *testing.T) {
	t.Parallel()
	program := writerProgram(
		[]ir.Local{{Name: "replace", Type: ir.Closure, Global: true, Function: -1, NestedFunction: 2}},
		ir.Evaluate{Value: ir.CallClosure{Closure: ir.Read{Local: 1, Of: ir.Closure}}},
	)
	wantSites(t, newAllocationFlowGraph(program).ReachingAllocations(ir.Read{Local: 0, Of: ir.Object}), -2, -1)
}

// A writer named only by an object literal's method table, Method.Function, and called through the
// object, as a call through an interface finds a class's method. Mutant 9.
func TestReachWriterNamedOnlyByAMethod(t *testing.T) {
	t.Parallel()
	program := writerProgram(
		[]ir.Local{{Name: "holder", Type: ir.Object, Global: true, Function: -1}},
		ir.Declare{Local: 1, Value: ir.ObjectLiteral{GraphTypes: []int{-3}, Methods: []ir.Method{{Name: "replace", Function: 1}}}},
		ir.Evaluate{Value: ir.CallClosure{Closure: ir.Property{Object: ir.Read{Local: 1, Of: ir.Object}, Name: "replace", Method: true, Of: ir.Closure}}},
	)
	wantSites(t, newAllocationFlowGraph(program).ReachingAllocations(ir.Read{Local: 0, Of: ir.Object}), -2, -1)
}

// Road 3's guard: liveness is read when a graph is built, so a reference a later pass adds to a
// function the graph took to never run is a writer the graph's consumers missed. checkFlowLiveness,
// run once lowering is done, makes that a compiler bug. Here replace is named by nothing when the
// graph is built, so current is only its literal; then Main calls it. Mutant 7.
func TestReachLateReferenceIsACompilerBug(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals: []ir.Local{
			{Name: "current", Type: ir.Object, Global: true, Function: -1},
			{Name: "next", Type: ir.Object, Function: 0},
		},
		Functions: []ir.Function{{
			Name:       "replace",
			Parameters: []int{1},
			Body:       []ir.Statement{ir.Assign{Local: 0, Value: ir.Read{Local: 1, Of: ir.Object}}},
		}},
		Main: []ir.Statement{ir.Declare{Local: 0, Value: ir.ObjectLiteral{GraphTypes: []int{-1}}}},
	}
	lowering := &lowering{result: program}
	graph := lowering.flowGraph(program)
	wantSites(t, graph.ReachingAllocations(ir.Read{Local: 0, Of: ir.Object}), -1)
	if err := checkFlowLiveness(program, lowering.flowDead); err != nil {
		t.Fatalf("nothing names replace yet, but the check says %v", err)
	}
	program.Main = append(program.Main, ir.Evaluate{Value: ir.Call{Function: 0, Arguments: []ir.Expression{ir.ObjectLiteral{GraphTypes: []int{-2}}}}})
	err := checkFlowLiveness(program, lowering.flowDead)
	if err == nil || err.Error() != "lower: compiler bug: function 0 was named after the flow graph was built" {
		t.Fatalf("got %v, want the compiler bug naming function 0", err)
	}
}

// checkFlowLiveness holds only the graphs that noted their dead (flowGraph, iterableViewFlowGraph), so a
// pass that builds one with newAllocationFlowGraph or newIterableViewFlowGraph directly judges by a
// liveness nothing checks again: host_member_write.go did, from #fj0vqtf wave 1 until this test. This
// reads every source file of the package and allows those two calls only inside their wrappers, and
// collectAllocationFlowGraph, which both build on, only inside the two constructors (Lower's should).
// Mutant 11: host_member_write.go's graph built by newAllocationFlowGraph again.
func TestEveryLoweringGraphNotesItsDead(t *testing.T) {
	t.Parallel()
	files, err := filepath.Glob("*.go")
	if err != nil {
		t.Fatal(err)
	}
	callers := map[string]map[string]bool{
		"newAllocationFlowGraph":     {"flowGraph": true},
		"newIterableViewFlowGraph":   {"iterableViewFlowGraph": true},
		"collectAllocationFlowGraph": {"newAllocationFlowGraph": true, "newIterableViewFlowGraph": true},
	}
	set := token.NewFileSet()
	for _, file := range files {
		if strings.HasSuffix(file, "_test.go") {
			continue
		}
		parsed, err := parser.ParseFile(set, file, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range parsed.Decls {
			function, isFunction := declaration.(*ast.FuncDecl)
			if !isFunction || function.Body == nil {
				continue
			}
			ast.Inspect(function.Body, func(node ast.Node) bool {
				if call, isCall := node.(*ast.CallExpr); isCall {
					if name, isName := call.Fun.(*ast.Ident); isName && callers[name.Name] != nil && !callers[name.Name][function.Name.Name] {
						t.Errorf("%s: %s builds a flow graph with %s; build it through l.flowGraph or l.iterableViewFlowGraph, so checkFlowLiveness holds it", set.Position(call.Pos()), function.Name.Name, name.Name)
					}
				}
				return true
			})
		}
	}
}

// Road 3's negatives: a writer that's called keeps its write and what it's passed; one exported from
// the entry module can be called by outside code with anything, and one made a function value can be
// called through it with anything, so each keeps its write and the parameter it writes is unknown.
// The function value is held by a let, so MakeClosure is the only field that names it (a const
// would name it by Local.ConstantClosure too). Mutants 3a, 3b and 3c.
func TestReachRunningWriterKeepsItsWrite(t *testing.T) {
	t.Parallel()
	t.Run("called", func(t *testing.T) {
		t.Parallel()
		program := reachProgram(t, reachBox+`let current: Box = { tag: 'first' };
function replace(next: Box): void {
	current = next;
}
replace({ tag: 'second' });
console.log(current.tag);
`)
		wantSites(t, reachOfGlobal(t, program, "current"), literalSite(t, program, "first"), literalSite(t, program, "second"))
	})
	t.Run("exported from the entry module", func(t *testing.T) {
		t.Parallel()
		program := reachProgram(t, reachBox+`let current: Box = { tag: 'first' };
export function replace(next: Box): void {
	current = next;
}
console.log(current.tag);
`)
		wantUnknown(t, reachOfGlobal(t, program, "current"))
	})
	t.Run("made a function value", func(t *testing.T) {
		t.Parallel()
		program := reachProgram(t, reachBox+`let current: Box = { tag: 'first' };
let replace = (next: Box): void => {
	current = next;
};
console.log(current.tag);
`)
		wantUnknown(t, reachOfGlobal(t, program, "current"), "parameter may receive untracked or omitted arguments")
	})
}

// ts.sys, sys.ts:1462: a module's exported let, initialized by an arrow called where it's written,
// whose inner let is undefined until a call fills it with a const object literal, returned with !.
// setSys, exported from a module that isn't the entry and called by nothing, writes it too. The read
// is exactly the literal. Mutants 2, 3 and 4 each turn it Unknown.
const reachSystem = `export interface System {
	tag: string;
}

function isNodeLikeSystem(): boolean {
	return true;
}

function getNodeSystem(): System {
	const nodeSystem: System = { tag: 'node' };
	return nodeSystem;
}

export let sys: System = (() => {
	let sys: System | undefined;
	if (isNodeLikeSystem()) {
		sys = getNodeSystem();
	}
	return sys!;
})();

export function setSys(s: System): void {
	sys = s;
}
`

// reachOfTagRead is what the flow graph says the object Main's one read of a tag field reads from can
// be, asked of the expression lowering made for that object, as a consumer would ask it.
func reachOfTagRead(t *testing.T, program *ir.Program) ir.AllocationSet {
	t.Helper()
	objects := []ir.Expression{}
	walk(program.Main, func(node any) bool {
		if property, isProperty := node.(ir.Property); isProperty && property.Name == "tag" {
			objects = append(objects, property.Object)
		}
		return true
	})
	if len(objects) != 1 {
		t.Fatalf("Main reads a tag field %d times, want once", len(objects))
	}
	return newAllocationFlowGraph(program).ReachingAllocations(objects[0])
}

// wantKnownClosureCalls fails unless Main calls a function value and every such call's targets are
// known: the arrow sys's initializer calls where it's written is one, and its result is demanded only
// if ClosureTargets names it.
func wantKnownClosureCalls(t *testing.T, program *ir.Program) {
	t.Helper()
	calls := 0
	walk(program.Main, func(node any) bool {
		if call, isCall := node.(ir.CallClosure); isCall {
			calls++
			if program.ClosureTargets(call).Unknown {
				t.Errorf("a call through a function value in Main has Unknown targets")
			}
		}
		return true
	})
	if calls == 0 {
		t.Errorf("Main calls no function value: the arrow called where it's written lowered some other way")
	}
}

func TestReachSystemIsItsLiteral(t *testing.T) {
	t.Parallel()
	program := reachModules(t, map[string]string{
		"sys.ts":  reachSystem,
		"main.ts": "import { sys } from './sys';\nconsole.log(sys.tag);\n",
	})
	wantKnownClosureCalls(t, program)
	wantSites(t, reachOfGlobal(t, program, "sys"), literalSite(t, program, "node"))
	wantSites(t, reachOfTagRead(t, program), literalSite(t, program, "node"))
}

// tsc reads it as ts.sys, through a namespace import of a barrel that re-exports sys.ts with export *
// (src/compiler/_namespaces/ts.ts). That read is the module's global, not a field of a namespace
// object, so it's no projection, which a program with any call through a function value would leave
// Unknown (shape_flow.go's projectedSources).
func TestReachSystemThroughANamespaceBarrelIsItsLiteral(t *testing.T) {
	t.Parallel()
	program := reachModules(t, map[string]string{
		"sys.ts":  reachSystem,
		"ts.ts":   "export * from './sys';\n",
		"main.ts": "import * as ts from './ts';\nconsole.log(ts.sys.tag);\n",
	})
	wantKnownClosureCalls(t, program)
	wantSites(t, reachOfTagRead(t, program), literalSite(t, program, "node"))
}

// Its twin: setSys called once with another literal is a write that runs, so the read is both.
func TestReachSystemSetOnceIsBothLiterals(t *testing.T) {
	t.Parallel()
	program := reachModules(t, map[string]string{
		"sys.ts":  reachSystem,
		"main.ts": "import { setSys, sys } from './sys';\nsetSys({ tag: 'other' });\nconsole.log(sys.tag);\n",
	})
	wantSites(t, reachOfGlobal(t, program, "sys"), literalSite(t, program, "node"), literalSite(t, program, "other"))
	wantSites(t, reachOfTagRead(t, program), literalSite(t, program, "node"), literalSite(t, program, "other"))
}

// functionReferenceFields must name every IR field that can hold a function, or a function some node
// names would be taken for one that never runs and its writes dropped. A field can hold one only if
// it holds an integer: of any width, signed or not, the ir package's own named integer types (Type,
// ViewContractID, ViewKind) among them, or a slice, array, pointer or map of one, as key or value.
// The structs statements and expressions hold are local_uses.go's localFields' to classify, and
// TestFunctionFieldTablesAgree holds this table to it, so here such a field is asked about when its
// name says it may hold a function. Every other struct a program holds, read from Program by its
// exported fields (Class, Accessor, Local, Function, Export, ViewContract and what they hold), no
// table but this one covers, so each of its integer fields is asked about whatever it's named: it's
// counted in functionReferenceFields or listed here with why it names no function. Mutant 10.
func TestFunctionReferenceFieldsAreEveryFunctionField(t *testing.T) {
	t.Parallel()
	const (
		irType     = "an IR type"
		contractID = "a view contract's ID"
		typeID     = "a checker type ID"
		counted    = "a count"
	)
	notFunctions := map[string]string{
		// Fields statements and expressions hold whose names say they may hold a function.
		"CallClosure.FunctionType":  typeID,
		"FunctionTargets.Functions": "ClosureTargets' answer, which no program holds",
		// Every integer field of the structs a program holds outside its statements and expressions.
		"Class.Definition":                  "the class's erased identity, its definition's first layout plus one",
		"Class.Base":                        "the base class's index plus one",
		"Class.OwnStart":                    "the first of the class's own field slots",
		"Class.StaticParent":                "a one-based hidden parent slot of a class's static side",
		"Class.StaticFlags":                 "one-based presence slots of the static side's data slots",
		"CounterGuard.Dependencies":         "the counter locals the guard's integer path needs",
		"Export.Local":                      "the global an export held in one reads",
		"Function.SourceLength":             "ECMAScript length, a count of parameters",
		"Function.Parameters":               "the parameters' locals",
		"Function.ArgumentsCount":           "the hidden argument count's local plus one",
		"Function.OptionalParameters":       "keyed by the locals of optional parameters",
		"Function.RestElement":              irType,
		"Function.Returns":                  irType,
		"Function.AsyncSite":                "a diagnostic source site",
		"Function.AsyncReturns":             irType,
		"Function.Yields":                   irType,
		"Function.Environment":              "the captured locals' cells",
		"Function.FrameEnvironment":         "the locals the frame's one allocation holds",
		"Function.FrameIdentity":            "a synthetic cell's local plus one",
		"Local.Type":                        irType,
		"Local.Function":                    "the function that declares the local, which says nothing about whether it runs",
		"Local.Ready":                       "the readiness local's index plus one",
		"MainModule.Statements":             counted,
		"NonNullCheckCounts.Proven":         counted,
		"NonNullCheckCounts.Checked":        counted,
		"OverloadResultCheckCounts.Proven":  counted,
		"OverloadResultCheckCounts.Checked": counted,
		"OverloadResultCheckCounts.Erased":  counted,
		"PredicateCallCheck.Overload":       "an overload's ordinal",
		"PredicateCheckCounts.Proven":       counted,
		"PredicateCheckCounts.Checked":      counted,
		"PredicateCheckCounts.Unobservable": counted,
		"Program.ProgramTypes":              "keyed by checker type IDs",
		"Program.GraphTypes":                "keyed by checker type IDs",
		"Program.ViewContractTypes":         "checker type IDs, each to a view contract's ID",
		"Program.PlainReceivers":            "keyed by checker type IDs",
		"Program.FunctionTypeTargets":       "every function a checker function type could be by assignability alone; a function is a value only by a counted node",
		"Program.FixedArgumentSlots":        counted,
		"Program.RestArgumentSlots":         "rest layouts, each to a count of slots",
		"Program.RestGraphTypes":            "rest layouts, each to its packs' graph types",
		"RestArguments.Start":               "the position of a rest's first argument",
		"RestArguments.Element":             irType,
		"SourcePosition.Line":               "a source line",
		"SourcePosition.Column":             "a source column",
		"ViewArrayCast.OriginStart":         "the start of a range of Program.ViewOrigins",
		"ViewArrayCast.OriginEnd":           "the end of a range of Program.ViewOrigins",
		"ViewContract.ObjectPresent":        contractID,
		"ViewContract.RepresentationMask":   "a mask of a callable union's members",
		"ViewContract.Payload":              contractID,
		"ViewContract.PayloadTypeID":        typeID,
		"ViewContract.NominalClass":         "a class's erased identity, as Class.Definition",
		"ViewContract.Kind":                 "a view's kind",
		"ViewContract.Of":                   irType,
		"ViewContract.Members":              contractID,
		"ViewContract.Key":                  contractID,
		"ViewContract.Element":              contractID,
		"ViewContract.TupleMinimum":         "a tuple's least arity",
		"ViewContract.TupleRest":            contractID,
		"ViewContract.Tuple":                contractID,
		"ViewContract.Parameters":           contractID,
		"ViewContract.Result":               contractID,
		"ViewFieldContract.Contract":        contractID,
		"ViewLiteral.Of":                    irType,
	}
	named := regexp.MustCompile(`Function|Direct|Comparator|Factory|Constructor|Method|Getter|Setter|Entry|Parent|Closure|Thunk|Target|Forward`)
	// Reads: internal/ir/*.go
	paths, err := filepath.Glob("../ir/*.go")
	if err != nil {
		t.Fatal(err)
	}
	structs := map[string]*ast.StructType{}
	// defined is each of the ir package's named types that isn't a struct, an interface or a function
	// type, by the type it's defined as.
	defined := map[string]ast.Expr{}
	roots := []string{}
	for _, path := range paths {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		for _, declaration := range file.Decls {
			switch declaration := declaration.(type) {
			case *ast.GenDecl:
				for _, spec := range declaration.Specs {
					if typeSpec, isType := spec.(*ast.TypeSpec); isType {
						switch as := typeSpec.Type.(type) {
						case *ast.StructType:
							structs[typeSpec.Name.Name] = as
						case *ast.InterfaceType, *ast.FuncType:
							// A Statement or an Expression, or a function: none is followed, as a
							// struct an interface holds is reached from the roots below instead.
						default:
							defined[typeSpec.Name.Name] = as
						}
					}
				}
			case *ast.FuncDecl:
				// A statement has the unexported statement(), an expression Type() Type.
				if declaration.Recv != nil && len(declaration.Recv.List) == 1 && (declaration.Name.Name == "statement" || declaration.Name.Name == "Type") {
					receiver := declaration.Recv.List[0].Type
					if star, isStar := receiver.(*ast.StarExpr); isStar {
						receiver = star.X
					}
					if identifier, isIdentifier := receiver.(*ast.Ident); isIdentifier {
						roots = append(roots, identifier.Name)
					}
				}
			}
		}
	}
	// integers are the names of the types that hold an integer: Go's own, and each named type of the
	// ir package defined as one or as a slice, array, pointer or map holding one.
	integers := map[string]bool{"int": true, "int8": true, "int16": true, "int32": true, "int64": true, "uint": true, "uint8": true, "uint16": true, "uint32": true, "uint64": true, "uintptr": true, "byte": true, "rune": true}
	var holdsInteger func(ast.Expr) bool
	holdsInteger = func(of ast.Expr) bool {
		switch of := of.(type) {
		case *ast.Ident:
			return integers[of.Name]
		case *ast.ArrayType:
			return holdsInteger(of.Elt)
		case *ast.StarExpr:
			return holdsInteger(of.X)
		case *ast.MapType:
			return holdsInteger(of.Key) || holdsInteger(of.Value)
		case *ast.IndexExpr:
			// A generic type's instance, as SortedMap[K, V] is: it holds what its arguments do.
			return holdsInteger(of.Index)
		case *ast.IndexListExpr:
			return slices.ContainsFunc(of.Indices, holdsInteger)
		}
		return false
	}
	for grown := true; grown; {
		grown = false
		for name, as := range defined {
			if !integers[name] && holdsInteger(as) {
				integers[name] = true
				grown = true
			}
		}
	}
	// reach is every struct the types named reach through their fields, through named types too, and
	// through only exported fields when exportedOnly is set.
	reach := func(from []string, exportedOnly bool) map[string]bool {
		reached := map[string]bool{}
		queue := slices.Clone(from)
		follow := func(of ast.Expr) {
			ast.Inspect(of, func(node ast.Node) bool {
				if identifier, isIdentifier := node.(*ast.Ident); isIdentifier && !reached[identifier.Name] {
					queue = append(queue, identifier.Name)
				}
				return true
			})
		}
		for len(queue) != 0 {
			name := queue[0]
			queue = queue[1:]
			if reached[name] {
				continue
			}
			if as, isDefined := defined[name]; isDefined {
				reached[name] = true
				follow(as)
				continue
			}
			structure, isStruct := structs[name]
			if !isStruct {
				continue
			}
			reached[name] = true
			for _, field := range structure.Fields.List {
				if exportedOnly && len(field.Names) != 0 && !slices.ContainsFunc(field.Names, (*ast.Ident).IsExported) {
					continue
				}
				follow(field.Type)
			}
		}
		return reached
	}
	inExpressions := reach(roots, false)
	inProgram := reach([]string{"Program"}, true)
	found := map[string]bool{}
	for structName, structure := range structs {
		whole := inProgram[structName] && !inExpressions[structName]
		for _, field := range structure.Fields.List {
			for _, name := range field.Names {
				key := structName + "." + name.Name
				found[key] = true
				if !name.IsExported() || !holdsInteger(field.Type) || !whole && !named.MatchString(name.Name) {
					continue
				}
				_, isCounted := functionReferenceFields[key]
				switch {
				case isCounted && notFunctions[key] != "":
					t.Errorf("IR field %s is counted in functionReferenceFields and listed here as naming no function", key)
				case isCounted:
				case notFunctions[key] != "":
				case whole:
					t.Errorf("IR field %s holds an integer, in %s, which only this table covers: count it in functionReferenceFields if it may name a function, or say here why it doesn't", key, structName)
				default:
					t.Errorf("IR field %s may hold a function: count it in functionReferenceFields, or say here why it doesn't", key)
				}
			}
		}
	}
	for key := range functionReferenceFields {
		if !found[key] {
			t.Errorf("functionReferenceFields names %s, which no IR struct has", key)
		}
	}
	for key := range notFunctions {
		if !found[key] {
			t.Errorf("the list of fields that hold no function names %s, which no IR struct has", key)
		}
	}
}

// localFieldFunction is what local_uses.go's localFields says of a field it classes as naming a
// function: true, with its offset, one where its why says the index is plus one and zero otherwise.
func localFieldFunction(why string) (int, bool) {
	if !strings.Contains(why, "function") || !strings.Contains(why, "index") {
		return 0, false
	}
	if strings.Contains(why, "plus one") {
		return 1, true
	}
	return 0, true
}

// functionReferenceFields and local_uses.go's localFields each class IR fields that hold a function's
// index, the second by its why. On every field both name, they agree, offset included; each field one
// names as a function, the other names too, unless it's listed here with why the other doesn't.
// localFields covers the fields of statements and expressions and the structs they hold, so the
// fields of a program, class, local, function, export or view contract are functionReferenceFields'
// alone.
func TestFunctionFieldTablesAgree(t *testing.T) {
	t.Parallel()
	const outside = "not a field of a statement, an expression or a struct one holds, which localFields covers"
	notInLocalFields := map[string]string{
		"Local.NestedFunction":           outside,
		"Local.ConstantClosure":          outside,
		"Class.Constructor":              outside,
		"Class.Methods":                  outside,
		"Accessor.Getter":                outside,
		"Accessor.Setter":                outside,
		"ViewContract.Functions":         outside,
		"Program.MethodTargets":          outside,
		"Program.StructuralMethodThunks": outside,
		"Program.AsyncEntry":             outside,
		"Export.Function":                outside,
		"Function.ForwardsArguments":     outside,
		"Function.NestedParent":          outside,
		"Function.ForwardedNestedParent": outside,
		"Function.ReferenceParents":      outside,
	}
	for field, classed := range localFields {
		offset, names := localFieldFunction(classed.why)
		if !names {
			continue
		}
		counted, listed := functionReferenceFields[field]
		if !listed {
			t.Errorf("localFields says %s is %q, which functionReferenceFields doesn't count", field, classed.why)
		} else if counted != offset {
			t.Errorf("localFields says %s is %q, offset %d, but functionReferenceFields offsets it by %d", field, classed.why, offset, counted)
		}
	}
	for field := range functionReferenceFields {
		_, excluded := notInLocalFields[field]
		classed, listed := localFields[field]
		_, names := localFieldFunction(classed.why)
		switch {
		case excluded && listed:
			t.Errorf("%s is listed here as outside localFields, which classes it as %q", field, classed.why)
		case !excluded && !(listed && names):
			t.Errorf("functionReferenceFields counts %s, which localFields doesn't class as a function's index: class it there, or say here why not", field)
		}
	}
	for field := range notInLocalFields {
		if _, counted := functionReferenceFields[field]; !counted {
			t.Errorf("%s is listed here, but functionReferenceFields doesn't count it", field)
		}
	}
}
