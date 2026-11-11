package lower

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// S3c-1's nested callee (#qk8rztp S3c-1): a generator's direct call handed into the one-read
// Iterable<T> parameter of a function declared inside the same parent, whose call is an
// ir.CallClosure, lowers where its ClosureTargets is exactly that one function and never Unknown,
// as a module callee's ir.Call does where its CallTargets is (Adamic's ruling 2 on #qk8rztp S3c).
// The oracle holds what the fixtures run to (internal/oracle/testdata/generators/nested_callee_*.a);
// this holds that they and the inline shapes lower, and what the lowered program holds.

// nestedCalleeFixtures are the oracle's S3c-1 programs: the parent's, a sibling's, an arrow's and the
// callee's own recursive calls with a Map's values() beside them (Language's roads 2, 3 and 4 where
// they lower), a grandchild's (road 5), an overload set (road 6) and a generic callee (road 7).
var nestedCalleeFixtures = []string{"nested_callee_view.a", "nested_callee_grandchild.a", "nested_callee_overloads.a", "nested_callee_generic.a"}

// nestedCalleeHost wraps body in a parent function, with a generator of numbers whose TReturn is
// undefined, so a road the one-read rule or the view proof closes reaches lowering's stop rather than
// the void rules'.
func nestedCalleeHost(body string) string {
	return "function host(): void {\nfunction* numbers(): Generator<number, undefined, undefined> { yield 1; yield 2; }\n" + body + "\n}\nhost();\n"
}

// nestedCalleeTotal is a nested callee whose one use of its parameter is one for-of over it.
const nestedCalleeTotal = "function total(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\n"

// Every fixture lowers, and each lowered program hands a generator to a nested callee through a
// CallClosure whose ClosureTargets is one function, a nested one, with checkGeneratorRoads holding.
// The inline rows are the shapes the fixtures don't carry: a module generator and one written in
// place into a nested callee (S3b's row that S3c-1 flips), a plain function*, whose void TReturn the
// parameter's one for-of never reads, so the void rules let it by one call deeper through the nested
// callee too (rulings.md's exemption, extended), an IterableIterator<T> parameter, and an
// `Iterable<T> | undefined` one. Mutant: iterableCallee's nested arm taken out, and every row stops,
// the plain function* rows refused by the void rules (adamic/no-void).
func TestNestedIterableCalleeLowers(t *testing.T) {
	t.Parallel()
	for _, name := range nestedCalleeFixtures {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			loaded, err := load.Load([]string{"../oracle/testdata/generators/" + name})
			if err != nil {
				t.Fatalf("Load: %v", err)
			}
			program, err := Lower(context.Background(), loaded)
			if err != nil {
				t.Fatal(err)
			}
			if handed := nestedCalleeHandings(program); handed == 0 {
				t.Error("no generator handed to a nested callee through a CallClosure")
			}
			if err := checkGeneratorRoads(program); err != nil {
				t.Errorf("the lowered program's roads: %v", err)
			}
		})
	}
	for _, probe := range []struct{ name, source string }{
		{"a module generator into a nested callee", "function* numbers(): Generator<number, undefined, undefined> { yield 1; yield 2; }\nfunction host(): void {\n" + nestedCalleeTotal + "console.log(String(total(numbers())));\n}\nhost();\n"},
		{"written in place into a nested callee (S3b's row)", "function host(flag: boolean): void {\nfunction inner(items: Iterable<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(inner((function* (): Generator<number, undefined, undefined> { yield 1; })())));\n}\nhost(true);\n"},
		{"a plain nested function* into a nested callee", "function host(): void {\nfunction* numbers() { yield 1; yield 2; }\n" + nestedCalleeTotal + "console.log(String(total(numbers())));\n}\nhost();\n"},
		{"a plain function* written in place into a nested callee", "function host(): void {\n" + nestedCalleeTotal + "console.log(String(total((function* () { yield 3; })())));\n}\nhost();\n"},
		{"an IterableIterator parameter", nestedCalleeHost("function total(items: IterableIterator<number>): number { let sum = 0; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(numbers())));")},
		{"an Iterable<T> | undefined parameter", nestedCalleeHost("function first(items: Iterable<number> | undefined): number { if (items !== undefined) { for (const value of items) { return value; } } return -1; }\nconsole.log(String(first(numbers())));")},
		{"a sibling calling the callee", nestedCalleeHost(nestedCalleeTotal + "function run(): number { return total(numbers()); }\nconsole.log(String(run()));")},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			program, err := lowerSource(t, probe.source)
			if err != nil {
				t.Fatalf("want it to lower, got %v", err)
			}
			if handed := nestedCalleeHandings(program); handed == 0 {
				t.Error("no generator handed to a nested callee through a CallClosure")
			}
		})
	}
}

// nestedCalleeHandings counts the generators program hands to a nested callee: IterableOf arguments
// of a CallClosure whose ClosureTargets is one function with a nested parent.
func nestedCalleeHandings(program *ir.Program) int {
	handed := 0
	inspect := func(node any) bool {
		call, isCall := node.(ir.CallClosure)
		if !isCall {
			return true
		}
		targets := program.ClosureTargets(call)
		if targets.Unknown || len(targets.Functions) != 1 || program.Functions[targets.Functions[0]].NestedParent == 0 {
			return true
		}
		for _, argument := range call.Arguments {
			if _, isIterable := argument.(ir.IterableOf); isIterable {
				handed++
			}
		}
		return true
	}
	walk(program.Main, inspect)
	for _, function := range program.Functions {
		walk(function.Body, inspect)
	}
	return handed
}

// Language's roads on #qk8rztp S3c that stop, each in its own words, and the stops a nested callee
// keeps from a module one's (Lower's and Adamic's rulings on S3c). Road 1: the callee also passed as
// a callback, so something the graph doesn't see may call it, and the view proof can't bound its
// parameter. Road 2: the callee recursing with its own parameter, a second read. Road 3: a local of
// the callee's name shadowing it, so the call names no function declaration. And a callee held in a
// local, one whose parameter a closure captures, and two generator functions into one parameter,
// which is S3c-2's. Each row's words are a part of its stop, so a row that stops for another reason
// goes red too. Mutants (internal/oracle/testdata/generators/mutants.md, S3c-1): onlyCalledDirectly
// answering yes for a closure also passed as a callback, and road 1's first row lowers;
// checkGeneratorArguments' CallClosure arm taken out, and road 2's row lowers, its second read
// uncounted.
func TestNestedIterableCalleeRoadsAreNotYet(t *testing.T) {
	t.Parallel()
	const other = "a generator function whose generator is used other than by a for-of over its call or a call as a statement (#qk8rztp)"
	for _, probe := range []struct{ name, source, reason string }{
		{"road 1: the callee also passed as a callback", "function apply(run: (items: Iterable<number>) => number, items: Iterable<number>): number { return run(items); }\n" + nestedCalleeHost(nestedCalleeTotal+"const sizes = new Map<string, number>([[`a${String(1)}`, 1]]);\nconsole.log(String(total(numbers()) + apply(total, sizes.values())));"), "a view over an iterable of unproven origin"},
		{"road 1: the callee also held in a local", nestedCalleeHost(nestedCalleeTotal + "const held = total;\nconsole.log(String(total(numbers()) + held(new Set<number>([3]).values())));"), "a view over an iterable of unproven origin"},
		{"road 2: the callee recursing with its parameter", nestedCalleeHost("function total(items: Iterable<number>, depth: number): number { let sum = 0; for (const value of items) { sum += value; } return depth > 0 ? sum + total(items, depth - 1) : sum; }\nconsole.log(String(total(numbers(), 1)));"), "read 2 times"},
		{"road 3: a local of the callee's name", nestedCalleeHost(nestedCalleeTotal + "{\nconst total = (items: Iterable<number>): number => { let sum = 0; for (const value of items) { sum += value; } return sum; };\nconsole.log(String(total(numbers())));\n}"), other},
		{"a call through a local holding the callee", nestedCalleeHost(nestedCalleeTotal + "const held = total;\nconsole.log(String(held(numbers())));"), other},
		{"a closure capturing the parameter", nestedCalleeHost("function total(items: Iterable<number>): number { let sum = 0; const keep = (): Iterable<number> => items; for (const value of items) { sum += value; } return sum; }\nconsole.log(String(total(numbers())));"), "which a closure captures"},
		{"two generator functions into one parameter (S3c-2)", nestedCalleeHost(nestedCalleeTotal + "function* halves(): Generator<number, undefined, undefined> { yield 0.5; }\nconsole.log(String(total(numbers()) + total(halves())));"), "generators of more than one function"},
	} {
		t.Run(probe.name, func(t *testing.T) {
			t.Parallel()
			_, err := lowerSource(t, probe.source)
			var gap *NotYet
			if !errors.As(err, &gap) || !strings.Contains(gap.What, probe.reason) {
				t.Fatalf("got %v, want a NotYet saying %q", err, probe.reason)
			}
		})
	}
}

// iterableCallTarget's nested arm admits a CallClosure only where ClosureTargets is exactly one
// function and never Unknown (Adamic's ruling 2 on #qk8rztp S3c): a sibling's call, direct; the
// parent's, through its binding; a grandchild's or a generic instance's, through a closure made for
// it. A call through a function value, Unknown, a spread, a bound closure's kept call and an optional
// call each answer no. No lowered program reaches the Unknown row, since iterableCallee admits only
// a name that resolves to a function declaration, so this holds the second line on its own. Mutant
// (internal/oracle/testdata/generators/mutants.md, S3c-1): its target test made `if false {`, and the
// Unknown row answers yes, or panics reading a target an Unknown answer doesn't have.
func TestNestedIterableCalleeIsExactlyOneBody(t *testing.T) {
	t.Parallel()
	program := &ir.Program{
		Locals:    []ir.Local{{Name: "consume", Type: ir.Closure, NestedFunction: 1}, {Name: "run", Type: ir.Closure}, {Name: "items", Type: ir.Object}},
		Functions: []ir.Function{{Name: "consume", Parameters: []int{2}, Closure: true, NestedParent: 2}, {Name: "host"}},
	}
	for _, row := range []struct {
		name string
		call ir.Expression
		one  bool
	}{
		{"a sibling's direct call", ir.CallClosure{Closure: ir.ClosureSelf{}, Direct: 1}, true},
		{"the parent's call through its binding", ir.CallClosure{Closure: ir.Read{Local: 0, Of: ir.Closure}}, true},
		{"a call through a closure made for it", ir.CallClosure{Closure: ir.MakeClosure{Function: 0}}, true},
		{"a call through a function value", ir.CallClosure{Closure: ir.Read{Local: 1, Of: ir.Closure}}, false},
		{"a spread", ir.CallClosure{Closure: ir.ClosureSelf{}, Direct: 1, Spread: []bool{true}}, false},
		{"a bound closure's kept call", ir.CallClosure{Closure: ir.ClosureSelf{}, Direct: 1, Bound: true}, false},
		{"an optional call", ir.CallClosure{Closure: ir.MakeClosure{Function: 0}, Optional: true}, false},
		{"a module function's call", ir.Call{Function: 1}, true},
		{"a virtual call", ir.Call{Function: 1, Virtual: 1}, false},
	} {
		t.Run(row.name, func(t *testing.T) {
			t.Parallel()
			answer := func() (one bool, err error) {
				defer func() {
					if recovered := recover(); recovered != nil {
						err = fmt.Errorf("panicked: %v", recovered)
					}
				}()
				_, one = iterableCallTarget(program, row.call)
				return one, nil
			}
			one, err := answer()
			if err != nil || one != row.one {
				t.Fatalf("one body %t (%v), want %t", one, err, row.one)
			}
		})
	}
}

// onlyCalledDirectly holds for a nested function every call of which is an edge the graph collects,
// and not for one also passed as a value (Lower's predicate on #qk8rztp S3c), and only the view
// proof's graph reads it (Compiler's scope): newAllocationFlowGraph, which graph regions, shapes and
// library_spread build, still marks every nested function's parameters unknown, as it did before
// S3c-1. Mutants (internal/oracle/testdata/generators/mutants.md, S3c-1): onlyCalledDirectly
// answering yes for a closure also passed as a callback, and the passed row goes red; the exemption
// read by the shared graph, and the shared rows go red.
func TestOnlyCalledDirectlyIsTheViewProofsAlone(t *testing.T) {
	t.Parallel()
	program, err := lowerSource(t, `function host(seed: number): number {
	function inc(n: number): number {
		return n + seed;
	}
	function twice(n: number): number {
		return inc(inc(n));
	}
	function passed(n: number): number {
		return n * 2;
	}
	function down(n: number): number {
		return n > 0 ? down(n - 1) : seed;
	}
	const apply = (f: (n: number) => number): number => f(seed);
	return twice(1) + inc(2) + apply(passed) + passed(3) + down(2);
}
console.log(String(host(1)));
`)
	if err != nil {
		t.Fatal(err)
	}
	named := map[string]int{}
	for index, function := range program.Functions {
		named[function.Name] = index
	}
	calls := closureCallsOf(program)
	shared := newAllocationFlowGraph(program)
	proof := newIterableViewFlowGraph(program)
	for _, row := range []struct {
		name   string
		direct bool
	}{{"inc", true}, {"twice", true}, {"down", true}, {"passed", false}} {
		function, found := named[row.name]
		if !found {
			t.Fatalf("no function %s", row.name)
		}
		parameter := program.Functions[function].Parameters[0] + 1
		if got := calls.onlyCalledDirectly(function); got != row.direct {
			t.Errorf("%s: onlyCalledDirectly %t, want %t", row.name, got, row.direct)
		}
		if !shared.unknownParameters[parameter] {
			t.Errorf("%s: the shared graph's parameter is known, where a MakeClosure leaves it unknown", row.name)
		}
		if proof.unknownParameters[parameter] == row.direct {
			t.Errorf("%s: the view proof's parameter unknown %t, want %t", row.name, proof.unknownParameters[parameter], !row.direct)
		}
	}
}
