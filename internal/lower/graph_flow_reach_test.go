package lower

import (
	"os"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// mayReachGraph on a handful of types, Node selected as graphTypes would: a scalar, an object or a
// container of scalars, and a class whose methods hold nothing reach none; a graph type, what holds one,
// and what this can't see to the end of (unknown, a type parameter, a function, an interface's method)
// reach one (#x490w4a). Mutant: mayReachGraph always true, and every "reaches none" row reds.
func TestMayReachGraph(t *testing.T) {
	t.Parallel()
	f, types := cycleIndexTypes(t, `interface Node { name: string; next: Node[] }
interface Visitor { visit(n: number): number }
class Counter { count = 0; bump(): void { this.count = this.count + 1; } }
declare const count: number;
declare const label: string;
declare const plain: { a: number };
declare const numbers: number[];
declare const pair: [number, string];
declare const table: Map<string, number>;
declare const counter: Counter;
declare const node: Node;
declare const holder: { n: Node };
declare const nodes: Node[];
declare const opaque: unknown;
declare const callback: (x: number) => number;
declare const visitor: Visitor;
function hold<T>(value: T): T { const held = value; return held; }
console.log(String(hold(count)));
`)
	f.l.result.GraphTypes = map[int]bool{int(types["node"].Id()): true}
	judged := map[*checker.Type]bool{}
	for _, each := range []struct {
		name    string
		reaches bool
	}{
		{"count", false},
		{"label", false},
		{"plain", false},
		{"numbers", false},
		{"pair", false},
		{"table", false},
		{"counter", false},
		{"node", true},
		{"holder", true},
		{"nodes", true},
		{"opaque", true},
		{"held", true},
		{"callback", true},
		{"visitor", true},
	} {
		proven := types[each.name]
		if proven == nil {
			t.Errorf("%s: the fixture has no such declaration", each.name)
			continue
		}
		if got := f.mayReachGraph(proven, judged); got != each.reaches {
			t.Errorf("%s, %s: mayReachGraph says %t, want %t", each.name, f.l.checker.TypeToString(proven), got, each.reaches)
		}
	}
	if !f.mayReachGraph(nil, judged) {
		t.Error("a type lowering didn't name reaches none: it must fail closed")
	}
}

// unknownClosureCalls is every function-value call of a lowered program whose targets are Unknown, Main's and
// every function's.
func unknownClosureCalls(program *ir.Program) []ir.CallClosure {
	calls := []ir.CallClosure{}
	visit := func(node any) bool {
		if call, isCall := node.(ir.CallClosure); isCall && program.ClosureTargets(call).Unknown {
			calls = append(calls, call)
		}
		return true
	}
	walk(program.Main, visit)
	for _, function := range program.Functions {
		walk(function.Body, visit)
	}
	return calls
}

// graphPlanned is whether graphFlows planned an allocation: its own site, the last of its GraphTypes, is set.
func graphPlanned(program *ir.Program, sites []int) bool {
	return len(sites) != 0 && program.GraphTypes[sites[len(sites)-1]]
}

// flow_unknown_scalar_arguments.a's plan: the interface method called through a parameter has Unknown targets,
// lowering names both its arguments' types, and the fresh { depth, flags } and [i & 3, i & 5] it's handed are
// left counted, since neither type can reach a graph type (#x490w4a). Performance measured them paying graph
// adoption on every call. internal/oracle's TestGraphOracleGraphRegionsFlowUnknownScalarArguments holds the
// program to Node and the counted build. Mutant 11 in graph-flow-exhaustive.mutants.md: mayReachGraph always
// true, and both are planned again.
func TestGraphFlowScalarArgumentsUnplanned(t *testing.T) {
	t.Parallel()
	source, err := os.ReadFile("../oracle/testdata/graph_regions/flow_unknown_scalar_arguments.a")
	if err != nil {
		t.Fatal(err)
	}
	program, err := lowerSource(t, string(source))
	if err != nil {
		t.Fatal(err)
	}
	if len(program.GraphTypes) == 0 {
		t.Fatal("no graph types selected: the fixture's cycle didn't make it a graph program")
	}
	visits := 0
	// The fixture's one call with Unknown targets is visitor.visit, so the count names it without reading its
	// callee, which only ClosureTargets may answer (internal/ir's TestCallTargetReadersLower).
	for _, call := range unknownClosureCalls(program) {
		visits++
		if len(call.ArgumentTypeIDs) != len(call.Arguments) {
			t.Fatalf("a call with %d arguments names %d argument types", len(call.Arguments), len(call.ArgumentTypeIDs))
		}
		for position, id := range call.ArgumentTypeIDs {
			if id == 0 {
				t.Errorf("argument %d of visitor.visit names no type", position)
			}
		}
	}
	if visits != 1 {
		t.Fatalf("found %d calls of visitor.visit with Unknown targets, want one: the fixture no longer reaches the quiet demand", visits)
	}
	contexts, pairs := 0, 0
	for _, function := range program.Functions {
		walk(function.Body, func(node any) bool {
			switch value := node.(type) {
			case ir.ObjectLiteral:
				for _, field := range value.Fields {
					if field.Name == "depth" {
						contexts++
						if graphPlanned(program, value.GraphTypes) {
							t.Error("the fresh { depth, flags } is planned into a graph region, though its type can't reach a graph type")
						}
					}
				}
			case ir.ArrayLiteral:
				if value.Element == ir.Number && len(value.Elements) == 2 {
					pairs++
					if graphPlanned(program, value.GraphTypes) {
						t.Error("the fresh [i & 3, i & 5] is planned into a graph region, though its type can't reach a graph type")
					}
				}
			}
			return true
		})
	}
	if contexts != 1 || pairs != 1 {
		t.Fatalf("found %d { depth, flags } literals and %d number pairs, want one of each", contexts, pairs)
	}
}

// A scalar-only literal passed as a parameter whose type reaches Node is planned: { n: undefined } holds nothing by
// its own type, and the callee sees it as a Holder, through which it can write a graph pointer into it, as this
// one does. Stored in the cycle too, a counted literal would close a cycle reference counting can't free. That
// leak can't be built today, since lowering stops the write into an optional field it would take (#avybwv1), so
// this holds the plan: argumentTypeID keeps the parameter's type beside the argument's own (#x490w4a). Mutant 12
// in graph-flow-exhaustive.mutants.md, the parameter's type dropped, leaves the literal counted.
func TestGraphFlowWiderParameterPlanned(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `interface Node { name: string; next: Node | undefined }
interface Holder { n: Node | undefined }
const a: Node = { name: 'a' + '', next: undefined };
const b: Node = { name: 'b' + '', next: a };
a.next = b;
function each(g: (x: Holder) => void): void { g({ n: undefined }); }
let seen = 0;
each((x: Holder): void => { x.n = a; seen += x.n === a ? 1 : 0; });
console.log(String(seen) + ' ' + String(a.next === b));
`)
	if err != nil {
		t.Fatal(err)
	}
	literals := 0
	for _, call := range unknownClosureCalls(program) {
		walk(call.Arguments, func(node any) bool {
			if literal, isLiteral := node.(ir.ObjectLiteral); isLiteral {
				literals++
				if !graphPlanned(program, literal.GraphTypes) {
					t.Error("the { n: undefined } passed as a Holder is left counted, though the callee can write a graph pointer into it")
				}
			}
			return true
		})
	}
	if literals != 1 {
		t.Fatalf("found %d literals made as an argument to a call with Unknown targets, want one", literals)
	}
}

// A reduce whose callee is parenthesized, (a.next.reduce)(f, [b]), is a.next.reduce(f, [b]): it lowers, both of what
// its callback is handed are numbered, and Initial's [b], a Node[], is planned as it is without the parentheses
// (#x490w4a). Reading the receiver off the unskipped callee panicked on the ParenthesizedExpression.
func TestArrayReduceParenthesizedCallee(t *testing.T) {
	t.Parallel()
	planned := map[string]bool{}
	for _, call := range []string{"a.next.reduce(f, [b])", "(a.next.reduce)(f, [b])"} {
		program, err := lowerSource(t, strings.ReplaceAll(`interface Node { name: string; next: Node[] }
const a: Node = { name: 'a' + '', next: [] };
const b: Node = { name: 'b' + '', next: [a] };
a.next.push(b);
function fold(f: (acc: Node[], n: Node) => Node[]): Node[] { return REDUCE; }
console.log(String(fold((acc: Node[], n: Node): Node[] => (n === b ? acc : n.next)).length));
`, "REDUCE", call))
		if err != nil {
			t.Fatalf("%s: %v", call, err)
		}
		reduces := 0
		for _, function := range program.Functions {
			walk(function.Body, func(node any) bool {
				if reduce, isReduce := node.(ir.ArrayReduce); isReduce {
					reduces++
					if reduce.InitialTypeID == 0 || reduce.ArrayTypeID == 0 {
						t.Errorf("%s: Initial is numbered %d and Array %d, want both named", call, reduce.InitialTypeID, reduce.ArrayTypeID)
					}
					if literal, isLiteral := reduce.Initial.(ir.ArrayLiteral); isLiteral {
						planned[call] = graphPlanned(program, literal.GraphTypes)
					} else {
						t.Errorf("%s: Initial is a %T, want the [b] literal", call, reduce.Initial)
					}
				}
				return true
			})
		}
		if reduces != 1 {
			t.Fatalf("%s: found %d reduces, want one", call, reduces)
		}
	}
	if !planned["a.next.reduce(f, [b])"] || planned["(a.next.reduce)(f, [b])"] != planned["a.next.reduce(f, [b])"] {
		t.Errorf("Initial's [b] planned: %v, want planned alike with and without the parentheses", planned)
	}
}

// An argument the callee stores in the cycle stays planned: [] is a never[] where it's made, which the callee
// sees as the Node[] it pushes into and stores, and f3's toSorted copy is a Node[] by its own type (#41vb71c).
// Neither needs the parameter's type: never[] is a graph type in this program itself, related to Node[].
func TestGraphFlowUnknownArgumentSlotPlanned(t *testing.T) {
	t.Parallel()
	f3, err := os.ReadFile("../oracle/testdata/graph_regions/flow_unknown_argument.a")
	if err != nil {
		t.Fatal(err)
	}
	for _, each := range []struct {
		name, source string
	}{
		{"an empty literal passed as a Node[]", `interface Node { name: string; next: Node[]; kids: Node[][] }
const a: Node = { name: 'a' + '', next: [], kids: [] };
const b: Node = { name: 'b' + '', next: [a], kids: [] };
a.next.push(b);
function each(g: (x: Node[]) => void): void { g([]); }
each((x: Node[]): void => { x.push(b); a.kids.push(x); });
const k = a.kids[0];
console.log(` + "`${a.kids.length} ${k === undefined ? -1 : k.length}`" + `);
`},
		{"f3's toSorted copy", string(f3)},
	} {
		t.Run(each.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, each.source)
			if err != nil {
				t.Fatal(err)
			}
			// g's call is the one with Unknown targets whose argument is an array made there.
			arrays := 0
			for _, call := range unknownClosureCalls(program) {
				walk(call.Arguments, func(node any) bool {
					var sites []int
					switch value := node.(type) {
					case ir.ArrayLiteral:
						sites = value.GraphTypes
					case ir.ArraySlice:
						sites = value.GraphTypes
					default:
						return true
					}
					arrays++
					if !graphPlanned(program, sites) {
						t.Error("the array the callee stores in the cycle is left counted, a leak")
					}
					return false
				})
			}
			if arrays != 1 {
				t.Fatalf("found %d arrays made as an argument to a call with Unknown targets, want one", arrays)
			}
		})
	}
}
