package high_level_intermediate_representation

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// disjointFor lowers one source and returns the outermost function with its classes.
//
// Typed for the same reason `rangesFor` is: a checker-less lowering never emits `StoreContext`, and
// this pass reads ranges that rule widens. A plain harness would make several assertions below pass
// vacuously.
func disjointFor(t *testing.T, source string) (*Function, *DisjointSet) {
	t.Helper()
	function, ranges := rangesFor(t, source)
	return function, FindDisjointMutableValuesWithRanges(function, ranges)
}

// classOfName returns the sorted class containing the LAST value a named binding takes.
func classOfName(t *testing.T, function *Function, set *DisjointSet, name string) []static_single_assignment.IdentifierId {
	t.Helper()
	var target static_single_assignment.IdentifierId
	found := false
	for id, identifier := range function.Identifiers {
		if identifier != nil && identifier.Name == name {
			if !found || static_single_assignment.IdentifierId(id) > target {
				target = static_single_assignment.IdentifierId(id)
				found = true
			}
		}
	}
	if !found {
		t.Fatalf("no identifier named %q, so this test measured nothing", name)
	}
	for _, class := range set.Sets() {
		for _, member := range class {
			if member == target {
				return class
			}
		}
	}
	return nil
}

// ---------------------------------------------------------------------------
// DisjointSet, measured against React's own implementation
// ---------------------------------------------------------------------------

// TestDisjointSetMatchesReact pins the union-find behaviour against React's real DisjointSet.
//
// These nine cases were not invented. They were run through React 7.1.1's own `DisjointSet`, reached
// by copying the development bundle and appending an export inside its IIFE, and this table is the
// output that produced. All nine agreed on the partition AND on the representative, which is the
// half that would otherwise differ silently: a union-find that picks a different root still answers
// "are these two together" correctly while handing every consumer a different key.
//
// The cases are chosen to separate implementations rather than to confirm one. `twoRootsMerge` and
// `longChainMerge` merge two ALREADY-ESTABLISHED classes, which is the only path that exercises the
// chain-rewriting loop inside `Union`; a naive implementation that simply repoints the item and not
// its ancestors passes every other case here.
func TestDisjointSetMatchesReact(t *testing.T) {
	t.Parallel()

	const (
		a static_single_assignment.IdentifierId = 1
		b static_single_assignment.IdentifierId = 2
		c static_single_assignment.IdentifierId = 3
		d static_single_assignment.IdentifierId = 4
		e static_single_assignment.IdentifierId = 5
		f static_single_assignment.IdentifierId = 6
	)

	tests := []struct {
		name     string
		unions   [][]static_single_assignment.IdentifierId
		wantSets [][]static_single_assignment.IdentifierId
		wantRoot map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId
	}{
		{
			name:     "chain",
			unions:   [][]static_single_assignment.IdentifierId{{a, b}, {b, c}, {c, d}},
			wantSets: [][]static_single_assignment.IdentifierId{{a, b, c, d}},
			wantRoot: map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{a: a, b: a, c: a, d: a},
		},
		{
			name:     "two established classes merge",
			unions:   [][]static_single_assignment.IdentifierId{{a, b}, {c, d}, {a, c}},
			wantSets: [][]static_single_assignment.IdentifierId{{a, b, c, d}},
			wantRoot: map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{a: a, b: a, c: a, d: a},
		},
		{
			name:     "singleton",
			unions:   [][]static_single_assignment.IdentifierId{{a}},
			wantSets: [][]static_single_assignment.IdentifierId{{a}},
			wantRoot: map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{a: a},
		},
		{
			name:     "repeated union is idempotent",
			unions:   [][]static_single_assignment.IdentifierId{{a, b}, {a, b}},
			wantSets: [][]static_single_assignment.IdentifierId{{a, b}},
			wantRoot: map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{a: a, b: a},
		},
		{
			// The first element is already a NON-ROOT member of a class, so the root the union
			// adopts under is not the element passed first.
			name:     "first element is not a root",
			unions:   [][]static_single_assignment.IdentifierId{{a, b}, {c, d}, {b, d}},
			wantSets: [][]static_single_assignment.IdentifierId{{a, b, c, d}},
			wantRoot: map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{a: a, b: a, c: a, d: a},
		},
		{
			name:     "three way merge",
			unions:   [][]static_single_assignment.IdentifierId{{a, b, c}, {d, e, f}, {c, d}},
			wantSets: [][]static_single_assignment.IdentifierId{{a, b, c, d, e, f}},
			wantRoot: map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{a: a, b: a, c: a, d: a, e: a, f: a},
		},
		{
			name:     "self union",
			unions:   [][]static_single_assignment.IdentifierId{{a, a}},
			wantSets: [][]static_single_assignment.IdentifierId{{a}},
			wantRoot: map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{a: a},
		},
		{
			name:     "two chains merged at their tails",
			unions:   [][]static_single_assignment.IdentifierId{{a, b}, {b, c}, {d, e}, {e, f}, {a, f}},
			wantSets: [][]static_single_assignment.IdentifierId{{a, b, c, d, e, f}},
			wantRoot: map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{a: a, b: a, c: a, d: a, e: a, f: a},
		},
		{
			name:     "star",
			unions:   [][]static_single_assignment.IdentifierId{{a, b}, {a, c}, {a, d}},
			wantSets: [][]static_single_assignment.IdentifierId{{a, b, c, d}},
			wantRoot: map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{a: a, b: a, c: a, d: a},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			set := &DisjointSet{}
			for _, union := range test.unions {
				set.Union(union)
			}

			got := set.Sets()
			if len(got) != len(test.wantSets) {
				t.Fatalf("got %d classes, want %d: %v", len(got), len(test.wantSets), got)
			}
			for i, want := range test.wantSets {
				if len(got[i]) != len(want) {
					t.Fatalf("class %d is %v, want %v", i, got[i], want)
				}
				for j := range want {
					if got[i][j] != want[j] {
						t.Fatalf("class %d is %v, want %v", i, got[i], want)
					}
				}
			}
			for member, wantRoot := range test.wantRoot {
				root, ok := set.Find(member)
				if !ok {
					t.Fatalf("%d is absent from the set", member)
				}
				if root != wantRoot {
					t.Errorf("representative of %d is %d, want %d (React picks %d)",
						member, root, wantRoot, wantRoot)
				}
			}
		})
	}
}

// TestDisjointSetAbsentValueIsNotClassZero pins that absence is reported separately from the id.
//
// `IdentifierId` zero is a usable value, so a `Find` returning only an id would make an absent value
// indistinguishable from a member of class zero. Upstream returns null; the boolean is that null.
func TestDisjointSetAbsentValueIsNotClassZero(t *testing.T) {
	t.Parallel()

	set := &DisjointSet{}
	set.Union([]static_single_assignment.IdentifierId{0, 1})

	if root, ok := set.Find(0); !ok || root != 0 {
		t.Fatalf("value zero should be a member with root 0, got (%d, %v)", root, ok)
	}
	if _, ok := set.Find(99); ok {
		t.Fatal("value 99 was never unioned, so Find must report it absent")
	}
	if got := set.RepresentativeOf(99); got != 99 {
		t.Fatalf("RepresentativeOf(99) = %d, want 99: an unmerged value is its own representative", got)
	}
	if set.Has(99) {
		t.Fatal("Has must agree with Find about membership")
	}
}

// TestDisjointSetEmptyUnionIsANoOp pins the one place this deliberately differs from upstream.
//
// React raises an invariant on an empty union. Every call site in this file builds the operand list
// conditionally, so an empty list is the ordinary outcome for an instruction that entangles nothing
// rather than a bug, and raising would turn the common case into a crash.
func TestDisjointSetEmptyUnionIsANoOp(t *testing.T) {
	t.Parallel()

	set := &DisjointSet{}
	set.Union(nil)
	set.Union([]static_single_assignment.IdentifierId{})
	if set.Size() != 0 {
		t.Fatalf("an empty union added %d members", set.Size())
	}
	if set.Sets() != nil {
		t.Fatal("an empty set has no classes")
	}
}

// TestDisjointSetDoesNotMutateTheCallersSlice pins the other difference from upstream.
//
// React's `union` calls `items.shift()`, consuming the caller's array. This takes the slice by value
// and never writes to it. The difference is in the implementation rather than in what it decides,
// and it is pinned so a later "optimization" that reuses a buffer cannot silently reintroduce the
// aliasing hazard.
func TestDisjointSetDoesNotMutateTheCallersSlice(t *testing.T) {
	t.Parallel()

	items := []static_single_assignment.IdentifierId{7, 8, 9}
	set := &DisjointSet{}
	set.Union(items)

	if len(items) != 3 || items[0] != 7 || items[1] != 8 || items[2] != 9 {
		t.Fatalf("Union modified the caller's slice: %v", items)
	}
}

// TestDisjointFindTerminatesOnALongChain proves the recursion is bounded, empirically.
//
// A union-find whose path compression is wrong does not give a wrong answer, it exhausts the stack,
// which is a silent-failure shape rather than a visible one. This builds a 10,000-long chain by
// hand -- longer than any real function produces -- and requires the walk to return. It also asserts
// compression actually happened, by requiring the SECOND walk to see a flat structure: without
// compression the chain would still be 10,000 deep on the way back.
func TestDisjointFindTerminatesOnALongChain(t *testing.T) {
	t.Parallel()

	const length = 10000
	set := &DisjointSet{parent: map[static_single_assignment.IdentifierId]static_single_assignment.IdentifierId{}}

	// A degenerate chain: 1 -> 2 -> 3 -> ... -> length, with length its own root. Built directly
	// rather than through Union, because Union's own rewriting would flatten it and the point here
	// is to hand Find the worst case.
	set.parent[static_single_assignment.IdentifierId(length)] = static_single_assignment.IdentifierId(length)
	for i := 1; i < length; i++ {
		set.parent[static_single_assignment.IdentifierId(i)] = static_single_assignment.IdentifierId(i + 1)
	}

	root, ok := set.Find(1)
	if !ok || root != static_single_assignment.IdentifierId(length) {
		t.Fatalf("Find(1) = (%d, %v), want (%d, true)", root, ok, length)
	}

	// Compression: every node on the walked path now points straight at the root.
	for i := 1; i < length; i++ {
		if set.parent[static_single_assignment.IdentifierId(i)] != static_single_assignment.IdentifierId(length) {
			t.Fatalf("path compression left %d pointing at %d rather than the root %d",
				i, set.parent[static_single_assignment.IdentifierId(i)], length)
		}
	}
}

// ---------------------------------------------------------------------------
// The pass
// ---------------------------------------------------------------------------

// TestDisjointUnifiesAMutatedObjectWithItsAlias is the case the whole pass exists for.
//
// `items` is built, aliased into `alias`, and then mutated through the alias. The two names refer to
// one object, so memoizing them independently would hand back a stale value, and upstream's answer
// is to put them in one class.
func TestDisjointUnifiesAMutatedObjectWithItsAlias(t *testing.T) {
	t.Parallel()

	function, set := disjointFor(t, `
export function build(seed: number) {
  const items: number[] = [];
  const alias = items;
  alias.push(seed);
  return items;
}
`)

	class := classOfName(t, function, set, "items")
	if len(class) < 2 {
		t.Fatalf("items is in a class of %d, so nothing was unified with it: %v", len(class), class)
	}

	aliasClass := classOfName(t, function, set, "alias")
	if len(aliasClass) == 0 {
		t.Fatal("alias is in no class at all")
	}
	itemsRoot := set.RepresentativeOf(class[0])
	aliasRoot := set.RepresentativeOf(aliasClass[0])
	if itemsRoot != aliasRoot {
		t.Errorf("items and alias have representatives %d and %d; a value mutated through its "+
			"alias must share one class with it", itemsRoot, aliasRoot)
	}
}

// TestDisjointLeavesIndependentValuesApart is the two-sided half of the test above.
//
// Without this, a pass that unioned EVERYTHING would pass the aliasing test and read as correct. Two
// values that never meet must land in different classes.
func TestDisjointLeavesIndependentValuesApart(t *testing.T) {
	t.Parallel()

	function, set := disjointFor(t, `
export function separate() {
  const left: number[] = [];
  left.push(1);
  const right: number[] = [];
  right.push(2);
  return [left, right];
}
`)

	leftClass := classOfName(t, function, set, "left")
	rightClass := classOfName(t, function, set, "right")
	if len(leftClass) == 0 || len(rightClass) == 0 {
		t.Fatal("one of the two values is in no class, so this test measured nothing")
	}
	if set.RepresentativeOf(leftClass[0]) == set.RepresentativeOf(rightClass[0]) {
		t.Error("left and right never alias, so they must not share a class")
	}
}

// TestDisjointAllocationGetsItsOwnClass pins the mayAllocate half of the lvalue gate.
//
// An object literal allocates, so upstream gives it a class even when its range is a single
// instruction: re-running `{}` produces a new identity and defeats every downstream comparison. The
// range gate alone would leave it out, so this is what separates the two halves of that condition.
//
// # The allocation lands on the TEMPORARY, not on the named binding, and the first version of this
// # test asserted the wrong one
//
// Written first as "the class containing `value`", it failed, and the pass was right. Lowering emits
// `t = {a: 1}` and then `StoreLocal value = t` as two instructions, so the ObjectExpression's lvalue
// is an unnamed temporary and `value` is a separate identifier that never allocates anything. The
// measured shape, with a one-wide range on every value:
//
//	order 2  ObjectExpression  lvalue 7 (unnamed)  mayAllocate=true   in the set
//	order 3  StoreLocal        lvalue 8 (unnamed)  mayAllocate=false  not in the set
//
// Asserting on the name would have been a fixture encoding the same wrong belief as a port that
// looked for allocations on named bindings, and both would have agreed. So this anchors on the
// instruction VALUE rather than on a name.
func TestDisjointAllocationGetsItsOwnClass(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function make() {
  const value = {a: 1};
  return value;
}
`)
	set := FindDisjointMutableValuesWithRanges(function, ranges)

	var allocations int
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			if _, ok := instruction.Value.(*ObjectExpression); !ok {
				continue
			}
			allocations++

			// The range is one wide, so the FIRST half of the gate is false here and only
			// mayAllocate can put this value in a class. That is what makes this a test of the
			// allocation half rather than of the range half.
			lvalueRange := ranges.Get(instruction.LValue.Identifier)
			if lvalueRange.End > lvalueRange.Start+1 {
				t.Fatalf("this value's range is %+v, wider than one, so the range gate alone "+
					"would admit it and this test no longer measures mayAllocate", lvalueRange)
			}
			if !set.Has(instruction.LValue.Identifier) {
				t.Error("an object literal allocates, so mayAllocate must place its lvalue in a " +
					"class even though its mutable range is one instruction wide")
			}
		}
	}
	if allocations == 0 {
		t.Fatal("no ObjectExpression was lowered, so this test measured nothing")
	}
}

// TestDisjointPrimitiveDoesNotAllocate is the control for the test above.
//
// `mayAllocate` returns false for `Primitive`, so a plain number with a one-wide range joins no
// class. Without this, a mayAllocate that returned true unconditionally would pass the allocation
// test and read as correct.
func TestDisjointPrimitiveDoesNotAllocate(t *testing.T) {
	t.Parallel()

	source := `
export function plain() {
  const n = 1;
  return n;
}
`
	function, ranges := rangesFor(t, source)
	set := FindDisjointMutableValuesWithRanges(function, ranges)

	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			if _, ok := instruction.Value.(*Primitive); !ok {
				continue
			}
			if mayAllocate(function, newCalleeProducers(function), instruction) {
				t.Error("mayAllocate must be false for a Primitive; React returns false for it " +
					"and a true here would union every constant in the function")
			}
			_ = set
		}
	}
}

// TestDisjointMayAllocateMatchesReact pins every instruction-kind arm against upstream's answer.
//
// The expected values were produced by driving React 7.1.1's own `mayAllocate` with a synthetic
// instruction per kind. The call cases here have no known signature, corresponding to an unresolved
// inferred result type; represented primitive and mutable signatures are exercised separately.
func TestDisjointMayAllocateMatchesReact(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value InstructionValue
		want  bool
	}{
		// React returns false for these regardless of the lvalue type.
		{"LoadLocal", &LoadLocal{}, false},
		{"LoadContext", &LoadContext{}, false},
		{"StoreLocal", &StoreLocal{}, false},
		{"StoreContext", &StoreContext{}, false},
		{"StoreGlobal", &StoreGlobal{}, false},
		{"LoadGlobal", &LoadGlobal{}, false},
		{"DeclareLocal", &DeclareLocal{}, false},
		{"DeclareContext", &DeclareContext{}, false},
		{"PropertyLoad", &PropertyLoad{}, false},
		{"PropertyDelete", &PropertyDelete{}, false},
		{"ComputedLoad", &ComputedLoad{}, false},
		{"ComputedDelete", &ComputedDelete{}, false},
		{"Primitive", &Primitive{}, false},
		{"TemplateLiteral", &TemplateLiteral{}, false},
		{"JsxText", &JsxText{}, false},
		{"UnaryExpression", &UnaryExpression{}, false},
		{"BinaryExpression", &BinaryExpression{}, false},
		{"PrefixUpdate", &PrefixUpdate{}, false},
		{"PostfixUpdate", &PostfixUpdate{}, false},
		{"Await", &Await{}, false},
		{"GetIterator", &GetIterator{}, false},
		{"IteratorNext", &IteratorNext{}, false},
		{"NextPropertyOf", &NextPropertyOf{}, false},
		{"MetaProperty", &MetaProperty{}, false},
		{"TypeCastExpression", &TypeCastExpression{}, false},
		{"Debugger", &Debugger{}, false},
		{"StartMemoize", &StartMemoize{}, false},
		{"FinishMemoize", &FinishMemoize{}, false},

		// React returns true for these regardless of the lvalue type.
		{"ObjectExpression", &ObjectExpression{}, true},
		{"ArrayExpression", &ArrayExpression{}, true},
		{"NewExpression", &NewExpression{}, true},
		{"RegExpLiteral", &RegExpLiteral{}, true},
		{"PropertyStore", &PropertyStore{}, true},
		{"ComputedStore", &ComputedStore{}, true},
		{"JsxExpression", &JsxExpression{}, true},
		{"JsxFragment", &JsxFragment{}, true},
		{"ObjectMethod", &ObjectMethod{}, true},
		{"FunctionExpression", &FunctionExpression{}, true},
		{"UnsupportedNode", &UnsupportedNode{}, true},

		// With no known primitive signature these take upstream's non-primitive default.
		{"CallExpression", &CallExpression{}, true},
		{"MethodCall", &MethodCall{}, true},
		{"TaggedTemplateExpression", &TaggedTemplateExpression{}, true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			if got := mayAllocate(nil, newCalleeProducers(nil), &Instruction{Value: test.value}); got != test.want {
				t.Errorf("mayAllocate(%s) = %v, want %v", test.name, got, test.want)
			}
		})
	}
}

// TestDisjointCallAllocationUsesTheEffectSignature pins every side of the call-result gate.
// React scopes a call result unless type inference proves it primitive. The effect signature table
// is the source of that same result-kind fact in this IR: method, global, and tagged primitive
// results are excluded, while mutable and unresolved results allocate.
func TestDisjointCallAllocationUsesTheEffectSignature(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `
export function classify(list: number[]) {
  const methodPrimitive = list.includes(1);
  const globalPrimitive = Number(list.length);
  const tagPrimitive = String`+"`value`"+`;
  const mutable = list.map(value => value);
  const unknown = unknownFunction(list);
  return [methodPrimitive, globalPrimitive, tagPrimitive, mutable, unknown];
}
`)

	want := map[string]bool{
		"includes":        false,
		"Number":          false,
		"String":          false,
		"map":             true,
		"unknownFunction": true,
	}
	seen := map[string]bool{}
	for _, instruction := range function.Instructions {
		if instruction == nil {
			continue
		}
		name := calleeSyntaxName(instruction)
		if name == "" {
			name = taggedTemplateCalleeSyntaxName(instruction)
		}
		expected, relevant := want[name]
		if !relevant {
			continue
		}
		seen[name] = true
		if got := mayAllocate(function, newCalleeProducers(function), instruction); got != expected {
			t.Errorf("mayAllocate(%s) = %v, want %v", name, got, expected)
		}
	}
	for name := range want {
		if !seen[name] {
			t.Errorf("fixture did not lower a %s call", name)
		}
	}
}

// TestDisjointDestructureAllocatesOnlyWithSpread pins the one pattern-dependent arm.
//
// React's `mayAllocate` for a Destructure is `doesPatternContainSpreadElement`. Measured against the
// real function: a plain object pattern is false, a pattern with a rest element is true. A rest
// binding allocates a fresh object to hold the remainder; a plain binding only names existing values.
func TestDisjointDestructureAllocatesOnlyWithSpread(t *testing.T) {
	t.Parallel()

	place := Place{Identifier: 1}

	plain := &Destructure{Pattern: &ObjectPattern{
		Properties: []ObjectPatternProperty{{Key: "a", Value: &PlacePattern{Place: place}}},
	}}
	if mayAllocate(nil, newCalleeProducers(nil), &Instruction{Value: plain}) {
		t.Error("a destructure with no rest element does not allocate")
	}

	spread := &Destructure{Pattern: &ObjectPattern{
		Properties: []ObjectPatternProperty{{Key: "a", Value: &PlacePattern{Place: place}}},
		Rest:       &place,
	}}
	if !mayAllocate(nil, newCalleeProducers(nil), &Instruction{Value: spread}) {
		t.Error("a destructure with a rest element allocates the remainder object")
	}

	arraySpread := &Destructure{Pattern: &ArrayPattern{Rest: &place}}
	if !mayAllocate(nil, newCalleeProducers(nil), &Instruction{Value: arraySpread}) {
		t.Error("an array pattern with a rest element allocates too")
	}

	// Nested, because `const {a: {...rest}} = o` allocates just as much as the flat form.
	nested := &Destructure{Pattern: &ObjectPattern{
		Properties: []ObjectPatternProperty{{
			Key:   "a",
			Value: &ObjectPattern{Rest: &place},
		}},
	}}
	if !mayAllocate(nil, newCalleeProducers(nil), &Instruction{Value: nested}) {
		t.Error("a rest element nested inside a pattern still allocates")
	}
}

// TestDisjointMethodCallAddsThePropertyUnconditionally pins the one identity-only union.
//
// Every other operand joins a class only when its range says it is still mutable. A method call's
// property joins on its identity alone, with no range test at all (React line 32411). It is the
// single exception in the pass and a reader would otherwise assume the guard was forgotten.
func TestDisjointMethodCallAddsThePropertyUnconditionally(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function call(target: {run(): void}) {
  target.run();
}
`)
	set := FindDisjointMutableValuesWithRanges(function, ranges)

	var properties int
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			methodCall, ok := instruction.Value.(*MethodCall)
			if !ok {
				continue
			}
			properties++
			if !set.Has(methodCall.Property.Identifier) {
				t.Errorf("the property of a method call joins its class unconditionally, but %d "+
					"is absent from the set", methodCall.Property.Identifier)
			}
		}
	}
	if properties == 0 {
		t.Fatal("no MethodCall was lowered, so this test measured nothing")
	}
}

// TestDisjointPhiUsesReactsTwoTermTest pins the phi predicate and the divergence from oxc.
//
// React's test is `start + 1 !== end && end > firstInstr`. The first term is written that way rather
// than as `end > start + 1` because on an UNSET range, both fields zero, `start + 1 !== end` is true
// while `end > start + 1` is false; the second term is what rejects it. Transcribing the tidier form
// would change which phis are considered, so both spellings are exercised here.
func TestDisjointPhiUsesReactsTwoTermTest(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		phiRange   mutation_aliasing.MutableRange
		firstOrder static_single_assignment.EvaluationOrder
		want       bool
	}{
		{"unset range is rejected by the second term", mutation_aliasing.MutableRange{Start: 0, End: 0}, 5, false},
		{"exactly one wide is rejected by the first term", mutation_aliasing.MutableRange{Start: 3, End: 4}, 1, false},
		{"wider than one and ending after the block fires", mutation_aliasing.MutableRange{Start: 3, End: 9}, 5, true},
		{"wider than one but ending before the block does not", mutation_aliasing.MutableRange{Start: 1, End: 3}, 5, false},
		{"ending exactly at the first order does not fire", mutation_aliasing.MutableRange{Start: 1, End: 5}, 5, false},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			got := test.phiRange.Start+1 != test.phiRange.End && test.phiRange.End > test.firstOrder
			if got != test.want {
				t.Errorf("React's two-term test on %+v with first order %d = %v, want %v",
					test.phiRange, test.firstOrder, got, test.want)
			}
			// The tidier spelling disagrees on the unset range, which is why upstream's is kept.
			tidier := test.phiRange.End > test.phiRange.Start+1 && test.phiRange.End > test.firstOrder
			if test.phiRange == (mutation_aliasing.MutableRange{Start: 0, End: 0}) && tidier == got {
				t.Log("note: on this input both spellings agree only because the second term rejects")
			}
		})
	}
}

// TestDisjointUnifiesALoopCarriedPhi is the phi predicate over real lowered code.
//
// A counter reassigned in a loop produces a phi whose operands are the pre-loop and post-loop
// values. When the phi's range is wider than one and outlives its block, upstream unions the phi,
// its declaration and every operand, so the loop's scope owns the whole variable rather than half.
func TestDisjointUnifiesALoopCarriedPhi(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function count(limit: number) {
  const seen: number[] = [];
  let total = 0;
  for (let index = 0; index < limit; index++) {
    total = total + index;
    seen.push(total);
  }
  return [total, seen];
}
`)
	set := FindDisjointMutableValuesWithRanges(function, ranges)

	var phis, fired int
	for _, block := range function.Blocks {
		firstOrder := mutation_aliasing.BlockFirstOrder(rangeGraph{}, function, block)
		for _, phi := range block.Phis {
			phis++
			phiRange := ranges.Get(phi.Place.Identifier)
			if phiRange.Start+1 != phiRange.End && phiRange.End > firstOrder {
				fired++
				if !set.Has(phi.Place.Identifier) {
					t.Errorf("phi %d satisfies React's test but is absent from the set",
						phi.Place.Identifier)
				}
				for _, blockId := range PhiOperandsInOrder(phi) {
					operand := phi.Operands.At(blockId).Identifier
					if set.RepresentativeOf(operand) != set.RepresentativeOf(phi.Place.Identifier) {
						t.Errorf("phi %d and its operand %d must share a class",
							phi.Place.Identifier, operand)
					}
				}
			}
		}
	}
	if phis == 0 {
		t.Fatal("the loop produced no phis, so this test measured nothing")
	}
	if fired == 0 {
		t.Fatal("no phi satisfied React's test, so the union branch was never exercised and this " +
			"test would pass over a pass that did nothing")
	}
}

// TestDisjointOperandGateNeedsBothHalves pins the paired operand test as a whole.
//
// `appendMutableOperand` is `isMutable(instr, operand) && operand.mutableRange.start > 0`, and a
// mutation sweep found that neither half was pinned by any other test here. Both survived, and they
// survived for OPPOSITE reasons, which is why they are separated below rather than covered by one
// assertion.
//
// Dropping `Start > 0` is EQUIVALENT on this tree, measured rather than argued: the two halves
// disagree in that direction on 0 of 26,006 operand reads across 200 corpus files, because `Contains`
// is already false for an unset range and this lowering opens no range at order zero. It is kept
// because upstream carries it and because a range legitimately starting at zero would separate them,
// but no fixture can catch its removal and inventing one would be a test asserting nothing.
//
// Dropping `Contains` is NOT equivalent: the two disagree in that direction on 11,105 of the same
// 26,006 reads, and it changes the answer. The check below is on the PARTITION rather than on any
// single membership, because that is the layer where the difference appears -- the loop fixture
// above already reaches the mutation and cannot see it, since it asserts phi membership and never
// the shape of the whole result. Measured on this input: the real pass produces 16 members in 2
// classes and the Contains-free variant produces 24 in 6.
func TestDisjointOperandGateNeedsBothHalves(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function loopy(limit: number) {
  const seen: number[] = [];
  let total = 0;
  for (let index = 0; index < limit; index++) {
    total = total + index;
    seen.push(total);
  }
  return [total, seen];
}
`)
	set := FindDisjointMutableValuesWithRanges(function, ranges)

	// An operand that is settled by this point must NOT be dragged into a class through this
	// instruction. Dropping `Contains` admits exactly those, which both inflates the member count
	// and fragments the partition into more, smaller classes. React's reverse-postorder traversal
	// puts loop bodies before their continuations. Immutable Assign refinement then leaves the two
	// primitive initializer copies settled, producing 12 members in 3 classes. The Contains-free
	// mutant produces 22 in 7; IMMUTABLE_ALIAS_EDGES.md records the removed members.
	classes := set.Sets()
	if got := len(classes); got != 3 {
		t.Errorf("this function forms %d classes, want 3. Dropping the `Contains` half of the "+
			"operand gate produces 7, because every settled operand joins a class it has no "+
			"business in: %v", got, classes)
	}
	if got := set.Size(); got != 12 {
		t.Errorf("this function unifies %d values, want 12. Dropping the `Contains` half of the "+
			"operand gate produces 22.", got)
	}

	// Two-sided: a pass that unified nothing would also fail the counts above, so require at least
	// one genuine multi-member class. A singleton is legitimate here: the corrected ordering leaves
	// the settled return aggregate in its own class.
	multiMember := false
	for _, class := range classes {
		if len(class) > 1 {
			multiMember = true
		}
	}
	if !multiMember {
		t.Errorf("all classes are singletons; the pass unified nothing: %v", classes)
	}
}

// TestDisjointPhiUnionIncludesTheDeclaration pins the declaration operand of the phi union.
//
// A firing phi unions itself, its operands, AND the value its source binding was first declared as
// (React line 32393). That third member is the one a reader is most likely to drop as redundant,
// because the declaration is usually already pulled into the class through some other instruction.
//
// It is NOT always redundant, and the measurement is what pins that. A mutation removing it survived
// every other test here. Across 200 corpus files the branch is reached by 60 firing phis, and on
// exactly ONE function of 343 the partition differs by the declaration member. That is a real
// distinguishing case rather than an equivalence, and this reproduces its shape minimally: a
// binding DECLARED without an initializer, assigned separately, then reassigned in a loop. The
// declared value is a member no other path contributes.
//
// Assert that relationship directly rather than pinning the size of the whole partition. Call
// allocation and alias refinement legitimately add or remove unrelated members; neither may make a
// phi stop sharing a class with the first value of its declaration.
func TestDisjointPhiUnionIncludesTheDeclaration(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		binding string
		source  string
	}{
		{
			name:    "declared without an initializer then reassigned in a for loop",
			binding: "acc",
			source: `
export function accumulate(limit: number) {
  let acc;
  acc = [];
  for (let index = 0; index < limit; index++) {
    acc.push(index);
    acc = acc.concat(index);
  }
  return acc;
}
`,
		},
		{
			name:    "declared without an initializer then reassigned in a while loop",
			binding: "list",
			source: `
export function gather(limit: number) {
  let list;
  list = [];
  let index = 0;
  while (index < limit) {
    list.push(index);
    list = list.concat(index);
    index = index + 1;
  }
  return list;
}
`,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			t.Parallel()
			function, ranges := rangesFor(t, test.source)
			set := FindDisjointMutableValuesWithRanges(function, ranges)

			var declaration static_single_assignment.IdentifierId
			var declarationFound bool
			for _, instruction := range function.Instructions {
				if instruction == nil {
					continue
				}
				value, ok := instruction.Value.(*DeclareLocal)
				if !ok {
					continue
				}
				identifier := function.Identifiers[value.LValue.Identifier]
				if identifier != nil && identifier.Name == test.binding {
					declaration = value.LValue.Identifier
					declarationFound = true
					break
				}
			}
			if !declarationFound {
				t.Fatalf("no DeclareLocal for %q", test.binding)
			}

			var phi static_single_assignment.IdentifierId
			var phiFound bool
			var phiBlockFirstOrder static_single_assignment.EvaluationOrder
			for _, block := range function.Blocks {
				if block == nil {
					continue
				}
				for _, candidate := range block.Phis {
					identifier := function.Identifiers[candidate.Place.Identifier]
					if identifier != nil && identifier.Name == test.binding {
						phi = candidate.Place.Identifier
						phiFound = true
						phiBlockFirstOrder = mutation_aliasing.BlockFirstOrder(rangeGraph{}, function, block)
						break
					}
				}
				if phiFound {
					break
				}
			}
			if !phiFound {
				t.Fatalf("no loop phi for %q", test.binding)
			}
			phiRange := ranges.Get(phi)
			if phiRange.Start+1 == phiRange.End || phiRange.End <= phiBlockFirstOrder {
				t.Fatalf("phi %d for %q does not reach the declaration-union branch: range=%v "+
					"blockFirstOrder=%d", phi, test.binding, phiRange, phiBlockFirstOrder)
			}

			if !set.Has(declaration) {
				t.Errorf("declared value %d for %q is absent from every class; the firing phi must "+
					"union the first value of its declaration", declaration, test.binding)
			}
			if !set.Has(phi) {
				t.Errorf("phi %d for %q is absent from every class", phi, test.binding)
			}
			if declarationRoot, phiRoot := set.RepresentativeOf(declaration),
				set.RepresentativeOf(phi); declarationRoot != phiRoot {
				t.Errorf("declared value %d and phi %d for %q have representatives %d and %d; "+
					"the phi union must include the declaration", declaration, phi, test.binding,
					declarationRoot, phiRoot)
			}
		})
	}
}

// TestDisjointDeclarationKeepsTheFirstValue pins that the declarations map is first-writer-wins.
//
// `declareIdentifier` writes only when the key is absent (React line 32381), so a binding assigned
// several times records the value it took FIRST. A mutation making it last-writer-wins survived
// every other test here, and it is reachable: across 200 corpus files the map sees 1,753 first
// writes and 216 further attempts that the guard declines.
//
// The direction matters because the recorded value is what a firing phi unions in. Recording the
// last value would union the phi with a value defined AFTER it in evaluation order, which is a
// different claim about what the scope owns than upstream makes.
func TestDisjointDeclarationKeepsTheFirstValue(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function reassign(limit: number) {
  let acc;
  acc = [];
  acc = [];
  for (let index = 0; index < limit; index++) {
    acc.push(index);
    acc = acc.concat(index);
  }
  return acc;
}
`)

	// Rebuild the declarations map exactly as the pass does, and assert the guard holds.
	declarations := map[static_single_assignment.DeclarationId]static_single_assignment.IdentifierId{}
	rewriteAttempts := 0
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			var lvalue Place
			switch value := instruction.Value.(type) {
			case *DeclareLocal:
				lvalue = value.LValue
			case *DeclareContext:
				lvalue = value.LValue
			case *StoreLocal:
				lvalue = value.LValue
			case *StoreContext:
				lvalue = value.LValue
			default:
				continue
			}
			declaration := declarationOf(function, lvalue.Identifier)
			if existing, present := declarations[declaration]; present {
				rewriteAttempts++
				if existing > lvalue.Identifier {
					t.Errorf("the map holds %d for declaration %d while %d was seen earlier; "+
						"first-writer-wins must keep the lower id", existing, declaration,
						lvalue.Identifier)
				}
				continue
			}
			declarations[declaration] = lvalue.Identifier
		}
	}

	if rewriteAttempts == 0 {
		t.Fatal("no binding was assigned twice, so the first-writer-wins guard was never " +
			"exercised and this test would pass over a pass that overwrote every entry")
	}
	if len(declarations) == 0 {
		t.Fatal("no declaration was recorded, so this test measured nothing")
	}

	// And the pass itself still forms classes over this input.
	set := FindDisjointMutableValuesWithRanges(function, ranges)
	if set.Size() == 0 {
		t.Fatal("no value joined a class")
	}
}

// TestDisjointIsIdempotent pins that a second run produces the same partition.
//
// Union-find is monotone, so a second pass over the same function must not merge anything further.
// A pass that read its own output would drift, and the drift would be invisible to any single run.
func TestDisjointIsIdempotent(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function repeat(seed: number) {
  const values: number[] = [];
  values.push(seed);
  let total = 0;
  for (let index = 0; index < seed; index++) {
    total = total + index;
  }
  return [values, total];
}
`)

	first := FindDisjointMutableValuesWithRanges(function, ranges)
	second := FindDisjointMutableValuesWithRanges(function, ranges)

	firstSets, secondSets := first.Sets(), second.Sets()
	if len(firstSets) == 0 {
		t.Fatal("no classes were formed, so this test measured nothing")
	}
	if len(firstSets) != len(secondSets) {
		t.Fatalf("first run made %d classes, second made %d", len(firstSets), len(secondSets))
	}
	for i := range firstSets {
		if len(firstSets[i]) != len(secondSets[i]) {
			t.Fatalf("class %d: %v then %v", i, firstSets[i], secondSets[i])
		}
		for j := range firstSets[i] {
			if firstSets[i][j] != secondSets[i][j] {
				t.Fatalf("class %d: %v then %v", i, firstSets[i], secondSets[i])
			}
		}
	}
}

// TestDisjointIsDeterministic pins that the answer does not depend on Go's map iteration order.
//
// `DisjointSet.parent` is a Go map. The PARTITION is order-independent by construction, but the
// representative and the ordering of `Sets` are not, and a consumer keying a table on the
// representative would get a different table per run. Repeated here rather than argued, because a
// randomised map is exactly the input that makes a single run look stable.
func TestDisjointIsDeterministic(t *testing.T) {
	t.Parallel()

	source := `
export function shuffle(limit: number) {
  let total = 0;
  const seen: number[] = [];
  for (let index = 0; index < limit; index++) {
    total = total + index;
    seen.push(total);
  }
  return [total, seen];
}
`
	function, ranges := rangesFor(t, source)

	baseline := FindDisjointMutableValuesWithRanges(function, ranges).Sets()
	if len(baseline) == 0 {
		t.Fatal("no classes were formed, so this test measured nothing")
	}
	for run := 0; run < 25; run++ {
		got := FindDisjointMutableValuesWithRanges(function, ranges).Sets()
		if len(got) != len(baseline) {
			t.Fatalf("run %d produced %d classes, baseline had %d", run, len(got), len(baseline))
		}
		for i := range baseline {
			if len(got[i]) != len(baseline[i]) {
				t.Fatalf("run %d class %d is %v, baseline %v", run, i, got[i], baseline[i])
			}
			for j := range baseline[i] {
				if got[i][j] != baseline[i][j] {
					t.Fatalf("run %d class %d is %v, baseline %v", run, i, got[i], baseline[i])
				}
			}
		}
	}
}

// TestDisjointHandlesANilFunction pins the guard every entry point carries.
func TestDisjointHandlesANilFunction(t *testing.T) {
	t.Parallel()

	if set := FindDisjointMutableValues(nil); set == nil || set.Size() != 0 {
		t.Fatal("a nil function must produce an empty set rather than a nil pointer")
	}
	if set := FindDisjointMutableValuesWithRanges(nil, nil); set == nil || set.Size() != 0 {
		t.Fatal("a nil function must produce an empty set rather than a nil pointer")
	}
}

// TestDisjointGapsAreDeclared pins the gap list.
//
// A test reading this list fails when the set changes, which is what makes closing a gap a visible
// event rather than a silent improvement. Same mechanism as `RangeGaps` and `ReactiveGaps`.
func TestDisjointGapsAreDeclared(t *testing.T) {
	t.Parallel()

	gaps := DisjointGaps()
	if len(gaps) != 1 || gaps[0] != DisjointGapPrimitiveCallResult {
		t.Fatalf("DisjointGaps() = %v; the only declared gap is the partial, name-keyed projection "+
			"of inferred call-result types. It closes when mayAllocate can read identifier types or "+
			"a complete shape-directed signature lookup.", gaps)
	}
}

// ---------------------------------------------------------------------------
// The corpus, which is the measurement that decides whether this pass is worth having
// ---------------------------------------------------------------------------

// TestDisjointClassSizesAreNotAllSingletons is this stage's headline, pinned as a test.
//
// Stage 3 was dispatched once before and DECLINED, correctly, on the measurement that the input was
// uniformly degenerate: every mutable range had width one, so every gate was false for a structural
// reason and the pass produced 23,256 members in 23,256 singleton classes. That is an identity
// function, and it would have compiled, passed every hand-written fixture above, and told every
// consumer that nothing is entangled with anything.
//
// So the non-degeneracy is a TEST rather than a number in a report. If a future change to
// `InferMutableRanges` narrows ranges back toward width one, this fails and names what happened,
// instead of the pass quietly reverting to an identity function that still passes its unit tests.
//
// Measured when written, over 400 corpus files and 685 outermost functions:
//
//	members                     26,962
//	classes                      3,306
//	singletons                   1,983   60.0% of classes
//	largest class                  527   in a 672-instruction hook
//	mean class size               8.16
//
// The thresholds below are deliberately far looser than those numbers. This guards the SHAPE -- that
// real multi-member classes are being formed -- not the exact corpus, which moves when the tree does.
func TestDisjointClassSizesAreNotAllSingletons(t *testing.T) {
	t.Parallel()

	skipWithoutCorpus(t)

	var files []string
	err := filepath.Walk(corpusRoot(t), func(path string, info os.FileInfo, err error) error {
		if err != nil || info.IsDir() {
			return nil
		}
		if strings.HasSuffix(path, ".ts") || strings.HasSuffix(path, ".tsx") {
			files = append(files, path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walking the corpus: %v", err)
	}
	sort.Strings(files)
	if len(files) > 200 {
		files = files[:200]
	}
	if len(files) < 10 {
		t.Fatalf("the corpus holds only %d files; the path is probably wrong", len(files))
	}

	var functions, members, classes, singletons, largest int

	for _, path := range files {
		contents, err := os.ReadFile(path)
		if err != nil {
			continue
		}
		fileName := "/corpus/" + filepath.Base(path)
		probe := rule.Rule{
			Name:             "disjoint-corpus",
			NeedsTypeChecker: true,
			Run: func(ctx rule.Context, options any) rule.Listeners {
				return rule.Listeners{
					ast.KindSourceFile: func(node *ast.Node) {
						if ctx.TypeChecker == nil {
							t.Fatal("the typed harness handed this probe a nil checker, so every " +
								"count below would be a fact about the harness rather than the code")
						}
						forEachFunctionLike(node, func(functionNode *ast.Node) {
							function := Lower(functionNode, ctx.TypeChecker)
							if function == nil {
								return
							}
							Construct(function)
							functions++
							set := FindDisjointMutableValues(function)
							members += set.Size()
							for _, class := range set.Sets() {
								classes++
								if len(class) == 1 {
									singletons++
								}
								if len(class) > largest {
									largest = len(class)
								}
							}
						})
					},
				}
			},
		}
		rule_testing.RunTypedFiles(t, probe, map[string]string{fileName: string(contents)}, fileName)
	}

	t.Logf("functions=%d members=%d classes=%d singletons=%d largest=%d",
		functions, members, classes, singletons, largest)

	if functions < 100 {
		t.Fatalf("only %d functions were lowered; this test measured almost nothing", functions)
	}
	if members == 0 {
		t.Fatal("no value joined any class, so the pass is inert")
	}
	if classes == singletons {
		t.Fatal("every class is a singleton, which is the identity function this stage exists to " +
			"avoid. The input has become degenerate for these predicates: check that " +
			"InferMutableRanges still produces ranges wider than one.")
	}
	if largest < 5 {
		t.Errorf("the largest class holds %d values; real code entangles more than that, so this "+
			"suggests the union is not propagating", largest)
	}
	// Two-sided: a pass that unioned everything would satisfy every check above.
	if singletons == 0 {
		t.Error("not one class is a singleton, which suggests over-merging: most values in real " +
			"code are independent")
	}
}
