package lower

import (
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"regexp"
	"slices"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
)

// ReachingAllocations' four roads (reach_roads.mutants.md names the one-line change that turns each
// back): an omitted argument to an optional or defaulted parameter is undefined, undefined and null
// add no site, a function that never runs makes no writes, and a non-null assertion hands on its
// value. Each pin lowers a small TypeScript program and asks the flow graph which object literals a
// module-level variable can hold, by the literal's tag. tsc's ts.sys (src/compiler/sys.ts:1462) is
// the shape they're for: the last pair holds it and its twin.

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

// An entry module's exported function is called by outside code with anything, so its parameter is
// unknown though the program's own call passes it a literal; the same function not exported is that
// literal. Mutant 6.
func TestReachExportedEntryFunctionParameterIsUnknown(t *testing.T) {
	t.Parallel()
	const body = `function keep(box: Box): Box {
	return box;
}
const kept = keep({ tag: 'inside' });
console.log(kept.tag);
`
	t.Run("not exported", func(t *testing.T) {
		t.Parallel()
		program := reachProgram(t, reachBox+body)
		wantSites(t, reachOfGlobal(t, program, "kept"), literalSite(t, program, "inside"))
	})
	t.Run("exported from the entry module", func(t *testing.T) {
		t.Parallel()
		program := reachProgram(t, reachBox+"export "+body)
		wantUnknown(t, reachOfGlobal(t, program, "kept"), "parameter may receive untracked or omitted arguments")
	})
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
				t.Errorf("a call through a function value in Main has Unknown targets: %#v", call.Closure)
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
// names would be taken for one that never runs and its writes dropped. Every field whose name says it
// may hold one, of an integer type, a slice of one or a map keyed by one, is in that table or in the
// list here of fields that don't, each with why.
func TestFunctionReferenceFieldsAreEveryFunctionField(t *testing.T) {
	t.Parallel()
	notFunctions := map[string]string{
		"Local.Function":              "the function that declares the local, which says nothing about whether it runs",
		"CallClosure.FunctionType":    "a checker type ID",
		"Program.FunctionTypeTargets": "every function a checker function type could be by assignability alone; a function is a value only by a counted node",
		"Class.StaticParent":          "a one-based hidden parent slot of a class's static side",
		"FunctionTargets.Functions":   "ClosureTargets' answer, which no program holds",
	}
	named := regexp.MustCompile(`Function|Direct|Comparator|Factory|Constructor|Method|Getter|Setter|Entry|Parent|Closure|Thunk|Target|Forward`)
	integer := regexp.MustCompile(`\bint\b`)
	// Reads: internal/ir/*.go
	files, err := filepath.Glob("../ir/*.go")
	if err != nil {
		t.Fatal(err)
	}
	found := map[string]bool{}
	for _, path := range files {
		if strings.HasSuffix(path, "_test.go") {
			continue
		}
		file, err := parser.ParseFile(token.NewFileSet(), path, nil, 0)
		if err != nil {
			t.Fatal(err)
		}
		ast.Inspect(file, func(node ast.Node) bool {
			declaration, isType := node.(*ast.TypeSpec)
			if !isType {
				return true
			}
			structure, isStruct := declaration.Type.(*ast.StructType)
			if !isStruct {
				return true
			}
			for _, field := range structure.Fields.List {
				for _, name := range field.Names {
					key := declaration.Name.Name + "." + name.Name
					found[key] = true
					if !name.IsExported() || !named.MatchString(name.Name) || !integer.MatchString(types.ExprString(field.Type)) {
						continue
					}
					if _, counted := functionReferenceFields[key]; !counted && notFunctions[key] == "" {
						t.Errorf("IR field %s may hold a function: count it in functionReferenceFields, or say here why it doesn't", key)
					}
				}
			}
			return true
		})
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
