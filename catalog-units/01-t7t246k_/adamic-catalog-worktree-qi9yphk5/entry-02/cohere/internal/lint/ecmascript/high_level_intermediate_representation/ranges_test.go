package high_level_intermediate_representation

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// rangesFor lowers one source, constructs single-assignment form, and returns the outermost
// function together with its ranges.
//
// Typed rather than plain, because a lowering built without a checker resolves every free reference
// as a global and never emits `StoreContext`, which is the one widening rule this pass can apply.
// A plain harness would make `TestRangesWidenAStoreIntoACapturedBinding` pass vacuously; that is
// measured rather than assumed, and the corpus probe that found it is recorded in the report.
func rangesFor(t *testing.T, source string) (*Function, *mutation_aliasing.MutableRanges) {
	t.Helper()

	var outermost *Function
	probe := rule.Rule{
		Name:             "ranges-harness",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						t.Fatal("the typed harness handed this probe a nil checker, so every " +
							"assertion below would pass vacuously")
					}
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						if outermost != nil {
							return
						}
						function := Lower(functionNode, ctx.TypeChecker)
						if function == nil {
							return
						}
						Construct(function)
						outermost = function
					})
				},
			}
		},
	}
	rule_testing.RunTypedFiles(t, probe, map[string]string{"fixture.ts": source}, "fixture.ts")
	if outermost == nil {
		t.Fatal("no function was lowered, so this test measured nothing")
	}
	return outermost, InferMutableRanges(outermost)
}

// rangeOfName returns the range of the value a named binding takes, choosing the LAST definition.
//
// Single-assignment form gives one binding several identifiers, so a name alone is ambiguous. The
// last one is chosen because these tests ask about a value after its final write, which is the
// question a consumer asks.
func rangeOfName(t *testing.T, function *Function, name string) mutation_aliasing.MutableRange {
	t.Helper()
	ranges := InferMutableRanges(function)
	var best mutation_aliasing.MutableRange
	var found bool
	for _, identifier := range function.Identifiers {
		if identifier == nil || identifier.Name != name {
			continue
		}
		r := ranges.Get(identifier.Id)
		if !r.IsSet() {
			continue
		}
		if !found || r.Start > best.Start {
			best = r
			found = true
		}
	}
	if !found {
		t.Fatalf("no value named %q carries a range; the lowering probably names it differently", name)
	}
	return best
}

// TestRangesOpenAtTheDefiningInstruction is the base case: a value's range starts where it is
// defined and ends one past it, which is upstream's unconditional lvalue rule.
func TestRangesOpenAtTheDefiningInstruction(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function f() {
  const a = 1;
  return a;
}
`)
	if ranges.Len() == 0 {
		t.Fatal("no value got a range, so this pass is inert rather than correctly silent")
	}
	r := rangeOfName(t, function, "a")
	if !r.IsSet() {
		t.Fatal("`a` has no range")
	}
	if r.End != r.Start+1 {
		t.Errorf("a value defined once and never mutated should span exactly one instruction, "+
			"got [%d,%d)", r.Start, r.End)
	}
}

// TestRangesAreHalfOpen pins React's `inRange`: the start is inside the range and the end is not.
//
// This is the property that separates an interval from a set of endpoints, and it is asserted
// directly rather than through a consumer because every consumer would hide it behind its own
// question.
func TestRangesAreHalfOpen(t *testing.T) {
	t.Parallel()

	r := mutation_aliasing.MutableRange{Start: 3, End: 6}
	for _, order := range []static_single_assignment.EvaluationOrder{3, 4, 5} {
		if !r.Contains(order) {
			t.Errorf("[3,6) should contain %d", order)
		}
	}
	for _, order := range []static_single_assignment.EvaluationOrder{2, 6, 7} {
		if r.Contains(order) {
			t.Errorf("[3,6) should not contain %d; the end is exclusive, which is React's inRange", order)
		}
	}
	if (mutation_aliasing.MutableRange{}).Contains(0) || (mutation_aliasing.MutableRange{}).Contains(5) {
		t.Error("the unset range must contain nothing, so a caller need not check IsSet first")
	}
}

// TestRangesSpanTheGapBetweenDisjointWrites is the interval-versus-set proof.
//
// A value written at two positions with untouched instructions between them has ONE interval
// covering both plus the gap, not two disjoint spans. Upstream's representation is `{start, end}`
// and its membership test is a comparison, so the gap is inside the range by construction.
//
// This is asserted over a hand-built range rather than over lowered source, deliberately. The
// widening that produces a two-write range is `AliasingState.mutate`, which needs effect facts this
// tree does not have, so no source input can produce the shape today. Building it by hand is what
// makes the REPRESENTATION testable independently of the pass that would populate it, and it is the
// case that would silently pass under a set-of-indices implementation while answering upstream's
// actual question wrongly.
func TestRangesSpanTheGapBetweenDisjointWrites(t *testing.T) {
	t.Parallel()

	// Written at 3, written again at 20, nothing between. Upstream's end is one past the last
	// write.
	r := mutation_aliasing.MutableRange{Start: 3, End: 21}

	if !r.Contains(10) {
		t.Error("an instruction in the GAP between two writes must be inside the range: upstream " +
			"represents this as one interval, so a reactive scope covering the value covers " +
			"everything between its writes. A set of write positions would answer false here and " +
			"would pass every straight-line test in this file")
	}
	if !r.Contains(3) || !r.Contains(20) {
		t.Error("both write positions must be inside the range")
	}
	if r.Contains(21) {
		t.Error("the end stays exclusive even for a widened range")
	}
}

// TestRangesRejectAnInvalidInterval pins upstream's `validateMutableRange` invariant: a range is
// either entirely unset or genuinely non-empty.
func TestRangesRejectAnInvalidInterval(t *testing.T) {
	t.Parallel()

	valid := []mutation_aliasing.MutableRange{{}, {Start: 1, End: 2}, {Start: 5, End: 100}}
	for _, r := range valid {
		if !r.IsValid() {
			t.Errorf("[%d,%d) should be valid", r.Start, r.End)
		}
	}
	invalid := []mutation_aliasing.MutableRange{{Start: 3, End: 3}, {Start: 5, End: 2}, {Start: 4, End: 0}}
	for _, r := range invalid {
		if r.IsValid() {
			t.Errorf("[%d,%d) should be invalid: upstream asserts (start==0 && end==0) || end>start",
				r.Start, r.End)
		}
	}
}

// TestRangesProduceNoInvalidIntervalsOnRealCode runs the invariant over a function with branches,
// a loop, and phis, which is where an off-by-one in the phi rule would show up.
func TestRangesProduceNoInvalidIntervalsOnRealCode(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function f(c: boolean, items: number[]) {
  let total = 0;
  for (const item of items) {
    if (c) {
      total = total + item;
    } else {
      total = total - item;
    }
  }
  return total;
}
`)
	if ranges.Len() == 0 {
		t.Fatal("no value got a range over a function with a loop and a branch; this is inert")
	}
	if invalid := mutation_aliasing.ValidateMutableRanges(ranges); len(invalid) != 0 {
		t.Errorf("%d values carry an invalid range: %v", len(invalid), invalid)
	}
	_ = function
}

// TestRangesWidenAStoreIntoACapturedBinding pins the ONE widening rule that survives the absence of
// effect facts.
//
// `n` is declared in `outer` and written from a nested arrow, which lowers to `StoreContext`. React
// widens the stored value's end from the instruction shape alone, with no effect consulted, at
// development-bundle lines 42002-42004 and oxc lines 960-967.
//
// This is also the test that catches a plain harness: with a nil checker `captureOf` cannot resolve
// and no `StoreContext` is emitted at all, so the assertion below would have nothing to find.
func TestRangesWidenAStoreIntoACapturedBinding(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `
export function outer() {
  let n = 0;
  const bump = () => { n = 1; };
  bump();
  return n;
}
`)

	var stores int
	var walk func(*Function)
	walk = func(f *Function) {
		ranges := InferMutableRanges(f)
		for _, block := range f.Blocks {
			for _, instructionId := range block.Instructions {
				instruction := f.Instructions[instructionId]
				if instruction == nil {
					continue
				}
				store, ok := instruction.Value.(*StoreContext)
				if !ok {
					continue
				}
				stores++
				r := ranges.Get(store.Value.Identifier)
				if !r.Contains(instruction.Order) {
					t.Errorf("the value stored into a captured binding at order %d is not inside "+
						"its own range [%d,%d); the StoreContext widening did not fire",
						instruction.Order, r.Start, r.End)
				}
			}
		}
		for _, nested := range f.Functions {
			walk(nested)
		}
	}
	walk(function)

	if stores == 0 {
		t.Fatal("no StoreContext was lowered, so this test measured nothing. That is what a " +
			"checker-less harness produces, and it is why rangesFor asserts the checker is present")
	}
}

// TestPhiNonMutabilityControlsUnknownCallScopes pins the reduced counterpart of upstream's
// `InferenceState.inferPhi`.
//
// A phi does not create a fresh mutable value. It names the union of its operand values, so an
// unknown call's conditional mutation is discarded when every arm is already frozen or primitive.
// One mutable arm is the control: it must keep the mutation and pull the phi into the call scope.
func TestPhiNonMutabilityControlsUnknownCallScopes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		source         string
		wantPhiInScope bool
	}{
		{
			name: "optional frozen and primitive arms",
			source: `
				function Component(input: {x?: {y: object}}) {
					return opaque(input?.x.y);
				}
			`,
		},
		{
			name: "optional destructured frozen and primitive arms",
			source: `
				function Component({input}: {input: {x?: {y: object}}}) {
					return opaque(input?.x.y);
				}
			`,
		},
		{
			name: "ternary frozen and primitive arms",
			source: `
				function Component(input: {flag: boolean, value: object}) {
					return opaque(input.flag ? input.value : undefined);
				}
			`,
		},
		{
			name: "ternary mutable arm",
			source: `
				function helper(input: {flag: boolean, value: object}) {
					return opaque(input.flag ? input.value : undefined);
				}
			`,
			wantPhiInScope: true,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, ranges := rangesFor(t, testCase.source)
			call, phi := unknownCallAndArgumentPhi(t, function, "opaque")
			set := FindDisjointMutableValuesWithRanges(function, ranges)
			scopes := AssignReactiveScopesWithSets(function, ranges, set)
			callScope := scopes.ScopeOf(call.LValue.Identifier)
			if callScope == 0 {
				t.Fatal("unknown call result received no allocation scope")
			}

			phiScope := scopes.ScopeOf(phi.Place.Identifier)
			if got := phiScope == callScope; got != testCase.wantPhiInScope {
				t.Errorf("phi and call share scope = %v, want %v; phi range=%v call scope=%v",
					got, testCase.wantPhiInScope, ranges.Get(phi.Place.Identifier),
					scopes.RangeOf(callScope))
			}
			if got := ranges.Contains(phi.Place.Identifier, call.Order); got != testCase.wantPhiInScope {
				t.Errorf("phi is mutable at call = %v, want %v", got, testCase.wantPhiInScope)
			}

			if !testCase.wantPhiInScope {
				callMembers := map[static_single_assignment.IdentifierId]bool{}
				for _, member := range scopes.MembersOf(callScope) {
					callMembers[member] = true
				}
				for _, block := range function.Blocks {
					if block == nil || block.Kind != BlockKindValue {
						continue
					}
					for _, instructionID := range block.Instructions {
						instruction := function.Instructions[instructionID]
						if instruction != nil && callMembers[instruction.LValue.Identifier] {
							t.Errorf("call scope contains value-block instruction %d (%T); upstream starts "+
								"the scope after the join", instructionID, instruction.Value)
						}
					}
				}
				if got := scopes.RangeOf(callScope); got.Start != call.Order || got.End != call.Order+1 {
					t.Errorf("call scope range=%v, want [%d,%d)", got, call.Order, call.Order+1)
				}
			}
		})
	}
}

// TestMutatingMethodReceiverSurvivesArgumentControlFlow pins receiver-effect continuity across
// argument lowering. A method receiver is evaluated before its arguments and retained on the
// MethodCall after any argument join; Array.push's direct Mutate(receiver) must therefore end the
// receiver range at call.Order+1 for straight-line, ternary, and optional-chain arguments alike.
func TestMutatingMethodReceiverSurvivesArgumentControlFlow(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name            string
		source          string
		wantArgumentPhi bool
	}{
		{
			name: "direct argument",
			source: `
				function helper(arg: {items: object}) {
					const x: object[] = [];
					x.push(arg.items);
					return x;
				}
			`,
		},
		{
			name:            "ternary argument",
			wantArgumentPhi: true,
			source: `
				function helper(arg: {items: object}, cond: boolean) {
					const x: (object | undefined)[] = [];
					x.push(cond ? arg.items : undefined);
					return x;
				}
			`,
		},
		{
			name:            "optional-chain argument",
			wantArgumentPhi: true,
			source: `
				function helper(arg: {items: object} | null) {
					const x: (object | undefined)[] = [];
					x.push(arg?.items);
					return x;
				}
			`,
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, ranges := rangesFor(t, testCase.source)
			var call *Instruction
			var method *MethodCall
			for _, instruction := range function.Instructions {
				if instruction == nil {
					continue
				}
				candidate, ok := instruction.Value.(*MethodCall)
				if !ok || calleeName(function, instruction, candidate.Property) != "push" {
					continue
				}
				call = instruction
				method = candidate
				break
			}
			if call == nil {
				t.Fatal("no Array.push MethodCall was lowered")
			}

			mutations := 0
			for _, effect := range InferAliasingEffects(function).Get(call.Id) {
				if effect.Kind == mutation_aliasing.AliasingEffectMutate &&
					effect.Into.Identifier == method.Receiver.Identifier {
					mutations++
				}
			}
			if mutations != 1 {
				t.Fatalf("push effects contain %d direct Mutate(receiver) entries, want exactly 1",
					mutations)
			}

			if len(method.Args) != 1 {
				t.Fatalf("push has %d arguments, want 1", len(method.Args))
			}
			argumentIsPhi := false
			argument := method.Args[0].Place.Identifier
			for _, block := range function.Blocks {
				if block == nil {
					continue
				}
				for _, phi := range block.Phis {
					if phi.Place.Identifier == argument {
						argumentIsPhi = true
					}
				}
			}
			if argumentIsPhi != testCase.wantArgumentPhi {
				t.Errorf("push argument is phi = %v, want %v; the test stopped distinguishing argument CFG",
					argumentIsPhi, testCase.wantArgumentPhi)
			}

			var initializer, storedX Place
			foundStore := false
			for _, instruction := range function.Instructions {
				if instruction == nil {
					continue
				}
				store, ok := instruction.Value.(*StoreLocal)
				if !ok || int(store.LValue.Identifier) >= len(function.Identifiers) {
					continue
				}
				identifier := function.Identifiers[store.LValue.Identifier]
				if identifier == nil || identifier.Name != "x" {
					continue
				}
				initializer = store.Value
				storedX = store.LValue
				foundStore = true
				break
			}
			if !foundStore {
				t.Fatal("no StoreLocal connected the array initializer to x")
			}
			receiverProducerFound := false
			for _, instruction := range function.Instructions {
				if instruction == nil || instruction.LValue.Identifier != method.Receiver.Identifier {
					continue
				}
				receiverProducerFound = true
				load, ok := instruction.Value.(*LoadLocal)
				if !ok {
					t.Fatalf("MethodCall receiver %d is produced by %T, want LoadLocal",
						method.Receiver.Identifier, instruction.Value)
				}
				if load.Place.Identifier != storedX.Identifier {
					t.Fatalf("MethodCall receiver loads %d, want stored x %d",
						load.Place.Identifier, storedX.Identifier)
				}
				break
			}
			if !receiverProducerFound {
				t.Fatalf("no instruction produces MethodCall receiver %d", method.Receiver.Identifier)
			}
			initializerIsArray := false
			for _, instruction := range function.Instructions {
				if instruction != nil && instruction.LValue.Identifier == initializer.Identifier {
					_, initializerIsArray = instruction.Value.(*ArrayExpression)
					break
				}
			}
			if !initializerIsArray {
				t.Fatalf("x initializer %d is not produced by an ArrayExpression", initializer.Identifier)
			}

			for _, value := range []struct {
				name  string
				place Place
			}{
				{name: "array initializer", place: initializer},
				{name: "stored x", place: storedX},
				{name: "method receiver", place: method.Receiver},
			} {
				got := ranges.Get(value.place.Identifier)
				if got.Start == 0 || got.Start >= call.Order {
					t.Errorf("%s range=%v does not begin before call order %d", value.name, got, call.Order)
				}
				if got.End != call.Order+1 {
					t.Errorf("%s range=%v, want end at push call + 1 (%d)", value.name, got, call.Order+1)
				}
			}
		})
	}
}

// TestAliasingRefinementMatchesAbstractKinds pins the kind-sensitive edge filtering performed by
// upstream's `InferenceState.applyEffect` before the range graph sees an effect.
//
// Alias is pruned for a frozen/primitive source or destination. MaybeAlias is deliberately
// asymmetric: a frozen source is pruned, but a primitive source is retained. The latter is a
// control against simplifying both variants to the same broad "immutable means no edge" rule.
func TestAliasingRefinementMatchesAbstractKinds(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		edge        mutation_aliasing.AliasingEffectKind
		fromKind    mutation_aliasing.EffectValueKind
		intoKind    mutation_aliasing.EffectValueKind
		wantWidened bool
	}{
		{name: "alias frozen source", edge: mutation_aliasing.AliasingEffectAlias, fromKind: mutation_aliasing.EffectValueFrozen, intoKind: mutation_aliasing.EffectValueMutable},
		{name: "alias primitive source", edge: mutation_aliasing.AliasingEffectAlias, fromKind: mutation_aliasing.EffectValuePrimitive, intoKind: mutation_aliasing.EffectValueMutable},
		{name: "alias frozen destination", edge: mutation_aliasing.AliasingEffectAlias, fromKind: mutation_aliasing.EffectValueMutable, intoKind: mutation_aliasing.EffectValueFrozen},
		{name: "alias mutable values", edge: mutation_aliasing.AliasingEffectAlias, fromKind: mutation_aliasing.EffectValueMutable, intoKind: mutation_aliasing.EffectValueMutable, wantWidened: true},
		{name: "maybe-alias frozen source", edge: mutation_aliasing.AliasingEffectMaybeAlias, fromKind: mutation_aliasing.EffectValueFrozen, intoKind: mutation_aliasing.EffectValueMutable},
		{name: "maybe-alias primitive source", edge: mutation_aliasing.AliasingEffectMaybeAlias, fromKind: mutation_aliasing.EffectValuePrimitive, intoKind: mutation_aliasing.EffectValueMutable, wantWidened: true},
		{name: "maybe-alias global source", edge: mutation_aliasing.AliasingEffectMaybeAlias, fromKind: mutation_aliasing.EffectValueGlobal, intoKind: mutation_aliasing.EffectValueMutable, wantWidened: true},
		{name: "maybe-alias mixed source", edge: mutation_aliasing.AliasingEffectMaybeAlias, fromKind: mutation_aliasing.EffectValueMaybeFrozen, intoKind: mutation_aliasing.EffectValueMutable, wantWidened: true},
		{name: "alias mixed source", edge: mutation_aliasing.AliasingEffectAlias, fromKind: mutation_aliasing.EffectValueMaybeFrozen, intoKind: mutation_aliasing.EffectValueMutable, wantWidened: true},
		{name: "capture mixed source", edge: mutation_aliasing.AliasingEffectCapture, fromKind: mutation_aliasing.EffectValueMaybeFrozen, intoKind: mutation_aliasing.EffectValueMutable, wantWidened: true},
		{name: "capture global source", edge: mutation_aliasing.AliasingEffectCapture, fromKind: mutation_aliasing.EffectValueGlobal, intoKind: mutation_aliasing.EffectValueMutable},
		{name: "create-from mixed source", edge: mutation_aliasing.AliasingEffectCreateFrom, fromKind: mutation_aliasing.EffectValueMaybeFrozen, intoKind: mutation_aliasing.EffectValueMutable, wantWidened: true},
		{name: "assign mixed source", edge: mutation_aliasing.AliasingEffectAssign, fromKind: mutation_aliasing.EffectValueMaybeFrozen, intoKind: mutation_aliasing.EffectValueMutable, wantWidened: true},
		{name: "create-from frozen source", edge: mutation_aliasing.AliasingEffectCreateFrom, fromKind: mutation_aliasing.EffectValueFrozen, intoKind: mutation_aliasing.EffectValueMutable},
		{name: "create-from primitive source", edge: mutation_aliasing.AliasingEffectCreateFrom, fromKind: mutation_aliasing.EffectValuePrimitive, intoKind: mutation_aliasing.EffectValueMutable},
		{name: "create-from mutable source", edge: mutation_aliasing.AliasingEffectCreateFrom, fromKind: mutation_aliasing.EffectValueMutable, intoKind: mutation_aliasing.EffectValueMutable, wantWidened: true},
		{name: "assign frozen source", edge: mutation_aliasing.AliasingEffectAssign, fromKind: mutation_aliasing.EffectValueFrozen, intoKind: mutation_aliasing.EffectValueMutable},
		{name: "assign primitive source", edge: mutation_aliasing.AliasingEffectAssign, fromKind: mutation_aliasing.EffectValuePrimitive, intoKind: mutation_aliasing.EffectValueMutable},
		{name: "assign mutable source", edge: mutation_aliasing.AliasingEffectAssign, fromKind: mutation_aliasing.EffectValueMutable, intoKind: mutation_aliasing.EffectValueMutable, wantWidened: true},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function := NewFunction(nil, "helper", FunctionKindOther)
			block := function.NewBlock(BlockKindBlock)
			function.Entry = block.Id
			place := func() Place {
				return Place{Identifier: function.NewIdentifier("", nil, 0).Id}
			}

			from := place()
			fromInstruction := &Instruction{LValue: from, Value: &Primitive{Value: 1}}
			function.AddInstruction(block, fromInstruction)
			into := place()
			edgeInstruction := &Instruction{LValue: into, Value: &ObjectExpression{}}
			function.AddInstruction(block, edgeInstruction)
			result := place()
			mutationInstruction := &Instruction{LValue: result, Value: &Primitive{Value: 0}}
			function.AddInstruction(block, mutationInstruction)
			function.Returns = result
			block.Terminal = &Return{Value: result}
			Finalize(function)

			mutationKind := mutation_aliasing.AliasingEffectMutate
			if testCase.edge == mutation_aliasing.AliasingEffectCapture {
				mutationKind = mutation_aliasing.AliasingEffectMutateTransitive
			}
			effects := &AliasingEffects{byInstruction: map[InstructionId][]AliasingEffect{
				fromInstruction.Id:     {mutation_aliasing.CreateEffect(from, testCase.fromKind)},
				edgeInstruction.Id:     {mutation_aliasing.CreateEffect(into, testCase.intoKind), mutation_aliasing.FlowEffect(testCase.edge, from, into)},
				mutationInstruction.Id: {mutation_aliasing.CreateEffect(result, mutation_aliasing.EffectValuePrimitive), mutation_aliasing.MutationEffect(mutationKind, into)},
			}}
			ranges := InferMutableRangesWithEffects(function, effects)

			wantEnd := fromInstruction.Order + 1
			if testCase.wantWidened {
				wantEnd = mutationInstruction.Order + 1
			}
			if got := ranges.Get(from.Identifier); got.End != wantEnd {
				t.Errorf("source range=%v, want end %d (edge=%s from=%s into=%s)", got,
					wantEnd, testCase.edge, testCase.fromKind, testCase.intoKind)
			}
		})
	}
}

func unknownCallAndArgumentPhi(t *testing.T, function *Function,
	name string) (*Instruction, *Phi) {
	t.Helper()
	var call *Instruction
	var argument static_single_assignment.IdentifierId
	for _, instruction := range function.Instructions {
		if instruction == nil {
			continue
		}
		value, ok := instruction.Value.(*CallExpression)
		if !ok || calleeName(function, instruction, value.Callee) != name || len(value.Args) != 1 {
			continue
		}
		call = instruction
		argument = value.Args[0].Place.Identifier
		break
	}
	if call == nil {
		t.Fatalf("no one-argument call to %s", name)
	}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		for _, phi := range block.Phis {
			if phi.Place.Identifier == argument {
				return call, phi
			}
		}
	}
	t.Fatalf("call argument %d is not a phi", argument)
	return nil, nil
}

// TestRangesAreIdempotent pins that a second run produces an equal table.
//
// `Construct` is famously NOT idempotent in this package, and a pass built beside it inherits the
// suspicion. Every guard here tests the range this pass itself wrote, so a second run reads what
// the first wrote and changes nothing.
func TestRangesAreIdempotent(t *testing.T) {
	t.Parallel()

	function, first := rangesFor(t, `
export function f(c: boolean) {
  let v = 1;
  if (c) { v = 2; }
  return v;
}
`)
	if first.Len() == 0 {
		t.Fatal("the first run produced no ranges, so idempotence here is vacuous")
	}
	second := InferMutableRanges(function)
	if second.Len() != first.Len() {
		t.Fatalf("a second run produced %d ranges against %d", second.Len(), first.Len())
	}
	for _, identifier := range function.Identifiers {
		id, want := identifier.Id, first.Get(identifier.Id)
		if got := second.Get(id); got != want {
			t.Errorf("value %d moved from [%d,%d) to [%d,%d) on a second run",
				id, want.Start, want.End, got.Start, got.End)
		}
	}
}

// TestRangesRequireEvaluationOrder pins that an unnumbered graph yields nothing rather than
// nonsense.
//
// Opening a range at order zero would write the UNSET value, which reads back as "no range" — a
// wrong answer wearing the shape of a right one. The pass skips instead, so a caller who forgot
// `Finalize` gets an empty table they can notice.
func TestRangesRequireEvaluationOrder(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function f() {
  const a = 1;
  return a;
}
`)
	if ranges.Len() == 0 {
		t.Fatal("the numbered control produced nothing, so the comparison below is meaningless")
	}

	for _, instruction := range function.Instructions {
		if instruction != nil {
			instruction.Order = 0
		}
	}
	unnumbered := InferMutableRanges(function)
	if unnumbered.Len() != 0 {
		t.Errorf("an unnumbered graph produced %d ranges; opening at order zero writes the unset "+
			"value and reads back as absent, which is a wrong answer that looks right",
			unnumbered.Len())
	}
}

// TestRangeGapsAreNamed pins the declared gaps, so closing one is a visible event.
//
// It worked. Stage 2 declared mutation-extension and alias-propagation gaps, and this test failed
// when Stage 2.5 closed both. It failed again when property loads and destructuring began emitting
// CreateFrom, forcing that input gap out of the API too.
func TestRangeGapsAreNamed(t *testing.T) {
	t.Parallel()

	gaps := mutation_aliasing.RangeGaps()
	if len(gaps) != 1 {
		t.Fatalf("expected exactly the one declared gap, got %d", len(gaps))
	}
	if gaps[0] != mutation_aliasing.RangeGapLoopCarriedInversion {
		t.Errorf("the declared gaps changed; if one was closed, update the package comment and the "+
			"RangeGap constants together, got %v", gaps)
	}
}

// TestCreateFromMutationPropagatesTransitively pins the range path that closed the former
// CreateFrom gap. A direct mutation of the derived value must reach both its source and a value the
// source captured; reaching only the source would pass if createdFrom incorrectly preserved the
// mutation's non-transitive kind.
func TestCreateFromMutationPropagatesTransitively(t *testing.T) {
	t.Parallel()

	function := NewFunction(nil, "helper", FunctionKindOther)
	block := function.NewBlock(BlockKindBlock)
	function.Entry = block.Id
	place := func() Place {
		return Place{Identifier: function.NewIdentifier("", nil, 0).Id}
	}

	leaf := place()
	leafInstruction := &Instruction{LValue: leaf, Value: &ObjectExpression{}}
	function.AddInstruction(block, leafInstruction)
	source := place()
	sourceInstruction := &Instruction{LValue: source, Value: &ObjectExpression{}}
	function.AddInstruction(block, sourceInstruction)
	derived := place()
	derivedInstruction := &Instruction{LValue: derived, Value: &PropertyLoad{Object: source, Property: "value"}}
	function.AddInstruction(block, derivedInstruction)
	result := place()
	mutationInstruction := &Instruction{LValue: result, Value: &Primitive{Value: 0}}
	function.AddInstruction(block, mutationInstruction)
	function.Returns = result
	block.Terminal = &Return{Value: result}
	Finalize(function)

	effects := &AliasingEffects{byInstruction: map[InstructionId][]AliasingEffect{
		leafInstruction.Id:     {mutation_aliasing.CreateEffect(leaf, mutation_aliasing.EffectValueMutable)},
		sourceInstruction.Id:   {mutation_aliasing.CreateEffect(source, mutation_aliasing.EffectValueMutable), mutation_aliasing.FlowEffect(mutation_aliasing.AliasingEffectCapture, leaf, source)},
		derivedInstruction.Id:  {mutation_aliasing.FlowEffect(mutation_aliasing.AliasingEffectCreateFrom, source, derived)},
		mutationInstruction.Id: {mutation_aliasing.CreateEffect(result, mutation_aliasing.EffectValuePrimitive), mutation_aliasing.MutationEffect(mutation_aliasing.AliasingEffectMutate, derived)},
	}}
	ranges := InferMutableRangesWithEffects(function, effects)

	wantEnd := mutationInstruction.Order + 1
	for _, value := range []struct {
		name  string
		place Place
	}{
		{name: "leaf", place: leaf},
		{name: "source", place: source},
		{name: "derived", place: derived},
	} {
		if got := ranges.Get(value.place.Identifier); got.End != wantEnd {
			t.Errorf("%s range=%v, want end %d through transitive CreateFrom propagation", value.name,
				got, wantEnd)
		}
	}
}

// TestRangesWidenToCoverAMutatingCall is the case Stage 2 wrote as an under-approximation and
// Stage 2.5 turned positive.
//
// It was originally `TestRangesAreAnUnderApproximation`, asserting the SHORTER answer and carrying
// an error message instructing whoever made it fail to update the gap list and the package comment.
// It failed on the first run of the widening, which is the event it existed to detect, and both were
// updated. Kept as the positive assertion rather than deleted, because the input is the smallest one
// on which the whole chain runs: a call effect names the argument mutated, the graph reaches it, and
// the widening covers the call.
//
// Upstream's answer, and now this one: `o` is defined at its literal and still being written at the
// call, so the interval spans both rather than closing at the definition.
func TestRangesWidenToCoverAMutatingCall(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `
declare function mutate(target: number[]): void;
export function f() {
  const o: number[] = [];
  mutate(o);
  return o;
}
`)
	r := rangeOfName(t, function, "o")
	if !r.IsSet() {
		t.Fatal("`o` has no range at all, so this measured nothing")
	}
	if r.End <= r.Start+1 {
		t.Errorf("`o`'s range is [%d,%d), which closes at its definition point. The call to "+
			"`mutate` carries a mutation effect naming `o`, so the widening should have extended "+
			"the end past it; a range of width 1 here means the alias graph never reached the value",
			r.Start, r.End)
	}
}

// TestRangesDoNotWidenWithoutAMutation is the two-sided control for the test above.
//
// Without it, a widening that extended EVERY range to the end of the function would pass the
// positive assertion and read as correct.
//
// # The first version of this control was WRONG and the code was right
//
// It was written as the same source with `mutate(o)` replaced by `read(o)`, on the assumption that a
// call taking a `readonly` parameter would carry no mutation effect. It failed, and the failure was
// the fixture rather than the widening: `read` appears in no signature table, so it takes the
// DEFAULT path, which `effects.go` documents as `MutateTransitiveConditionally` on every operand
// plus a capture into every other one. An unknown call is assumed to mutate everything handed to it,
// which is the conservative direction and is upstream's behavior rather than ours.
//
// So a call cannot be the control at all: every unknown call widens, correctly. The control has to
// be a function with no call in it, where the mutation effect count is genuinely zero. Measured on
// this input: `mutationEffects=0`, and the widest range in the whole function is 1.
//
// Recorded at length because the failing version looked exactly like a defect in the widening and
// the brief's rule -- when a probe says shipped code is broken, the probe is wrong until a control
// says otherwise -- is what produced the right diagnosis.
func TestRangesDoNotWidenWithoutAMutation(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function f() {
  const o: number[] = [];
  const n = 1;
  return n;
}
`)
	effects := InferAliasingEffects(function)
	mutations := 0
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			for _, effect := range effects.Get(instructionId) {
				if effect.Kind.IsMutation() {
					mutations++
				}
			}
		}
	}
	if mutations != 0 {
		t.Fatalf("this control assumes the input carries no mutation effect and it carries %d; "+
			"the control is measuring something other than what it claims", mutations)
	}
	for _, identifier := range function.Identifiers {
		if identifier == nil {
			continue
		}
		r := ranges.Get(identifier.Id)
		if !r.IsSet() {
			continue
		}
		if r.End > r.Start+1 {
			t.Errorf("value %d has range [%d,%d) but the function carries no mutation effect at "+
				"all; a widening that fires here is firing on instruction shape rather than on an "+
				"effect", identifier.Id, r.Start, r.End)
		}
	}
}

// TestMutationSitesNamesWhatStageOneWouldWiden pins the seam, two-sided: the definite shapes are
// found and an ordinary definition is not mistaken for one.
func TestMutationSitesNamesWhatStageOneWouldWiden(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `
declare function use(x: unknown): void;
export function f(o: {a: number}, k: string) {
  o.a = 1;
  (o as any)[k] = 2;
  use(o);
  const plain = 3;
  return plain;
}
`)
	counts := map[MutationSiteKind]int{}
	for _, site := range MutationSites(function) {
		counts[site.Kind]++
	}
	if counts[MutationSitePropertyStore] == 0 {
		t.Error("a property store was not named as a mutation site")
	}
	if counts[MutationSiteComputedStore] == 0 {
		t.Error("a computed store was not named as a mutation site")
	}
	if counts[MutationSiteCall] == 0 {
		t.Error("a call was not named as a mutation site; calls are the CEILING half of the gap " +
			"and are counted separately rather than omitted")
	}

	// The two-sided half: a function that mutates nothing names no definite site. Without this a
	// MutationSites that returned every instruction would pass the assertions above.
	clean, _ := rangesFor(t, `
export function g() {
  const a = 1;
  const b = a + 1;
  return b;
}
`)
	for _, site := range MutationSites(clean) {
		if site.Kind != MutationSiteCall {
			t.Errorf("a function with no stores named a %s site", site.Kind)
		}
	}
}

// TestRangesLeaveParametersUnset pins that a value this function did not define gets no range.
//
// Upstream opens a range only at an INSTRUCTION LVALUE. A parameter arrives already defined, so
// nothing opens one for it, and `state.create` records it in the aliasing graph without touching
// `mutableRange`. A pass that opened ranges at operand positions instead would give every parameter
// a one-instruction range at its first USE, which is a different and wrong analysis: it would claim
// the parameter is still being written at the point it is merely read.
//
// This is the fixture a mutation sweep demanded. Inverting the `PlaceRoleDefine` filter in the
// lvalue loop survived the whole suite, and the distinguishing input is exactly this shape:
// measured, parameter `p` moves from unset to [1,2) under that mutant while every other assertion
// in this file stays green, because every OTHER value is defined by an instruction before it is
// used and so its range is already open by the time the operand loop sees it.
func TestRangesLeaveParametersUnset(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function f(p: number) {
  const a = p + 1;
  return a;
}
`)
	if ranges.Len() == 0 {
		t.Fatal("no value got a range, so the negative assertion below is vacuous")
	}

	var checked int
	for _, identifier := range function.Identifiers {
		if identifier == nil || identifier.Name != "p" {
			continue
		}
		checked++
		if r := ranges.Get(identifier.Id); r.IsSet() {
			t.Errorf("parameter `p` (value %d) carries range [%d,%d); upstream opens a range only "+
				"at an instruction lvalue, and a parameter is defined by the call rather than by "+
				"any instruction in this function", identifier.Id, r.Start, r.End)
		}
	}
	if checked == 0 {
		t.Fatal("no value named `p` was found, so this test measured nothing")
	}
}

// TestRangeIsSetTreatsEitherFieldAsSet pins that a range with only an end is SET.
//
// This is a unit assertion rather than one over lowered source, deliberately. The state it
// describes - an end without a start - cannot be produced by this pass today: measured two-sided
// over the corpus, 145,113 values carry a set range and 0 carry an end without a start, because the
// two writers of `End` both set `Start` alongside it. Stage 1's mutation extension is a writer that
// does not, so the state becomes reachable then.
//
// Asserting it now is what keeps `IsSet` honest through that change. A mutation reducing it to
// `Start != 0` survived the whole suite, because nothing could construct the input that separates
// the two spellings; the unit test can, and it costs one line.
func TestRangeIsSetTreatsEitherFieldAsSet(t *testing.T) {
	t.Parallel()

	if (mutation_aliasing.MutableRange{}).IsSet() {
		t.Error("the zero range must not read as set")
	}
	if !(mutation_aliasing.MutableRange{Start: 3, End: 4}).IsSet() {
		t.Error("a fully set range must read as set")
	}
	if !(mutation_aliasing.MutableRange{Start: 0, End: 4}).IsSet() {
		t.Error("a range carrying only an END must read as set. Upstream's unset test is " +
			"`start === 0 && end === 0`, both fields, and a spelling that reads only Start would " +
			"report a value widened by Stage 1 as having no range at all")
	}
	if !(mutation_aliasing.MutableRange{Start: 3, End: 0}).IsSet() {
		t.Error("a range carrying only a START must read as set")
	}
}

// TestBlockFirstOrderFallsBackToTheTerminal pins upstream's `?? block.terminal.id`.
//
// An empty block is common in this graph - measured, 4,546 of 28,335 blocks, and 1,198 of those
// carry phis - so the fallback runs constantly. What it cannot yet change is an ANSWER, because the
// only reader is `phiOpensBefore`, whose own precondition needs a widened end that Stage 1 has not
// landed. A mutation replacing the fallback with zero therefore survived every fixture over lowered
// source.
//
// Tested directly against a hand-built block for that reason. Returning zero here would make
// `phiOpensBefore` decline every phi in an empty block once Stage 1 makes it live, which is a
// silent under-report in exactly the shape - a merge point with no instructions - that a loop
// header most often has.
func TestBlockFirstOrderFallsBackToTheTerminal(t *testing.T) {
	t.Parallel()

	function := &Function{}
	withInstruction := &BasicBlock{
		Id:           1,
		Instructions: []InstructionId{0},
		Terminal:     &Goto{Order: 99},
	}
	function.Instructions = []*Instruction{{Id: 0, Order: 7}}
	if got := mutation_aliasing.BlockFirstOrder(rangeGraph{}, function, withInstruction); got != 7 {
		t.Errorf("a block with instructions must report its FIRST instruction's order, got %d", got)
	}

	empty := &BasicBlock{Id: 2, Terminal: &Goto{Order: 42}}
	if got := mutation_aliasing.BlockFirstOrder(rangeGraph{}, function, empty); got != 42 {
		t.Errorf("an empty block must fall back to its TERMINAL's order, which is upstream's "+
			"`block.instructions.at(0)?.id ?? block.terminal.id`, got %d. Zero would read as "+
			"unassigned and make phiOpensBefore decline every phi in a merge block", got)
	}
}

// TestMutationSitesNameTheMutatedValueNotTheStoredOne pins WHICH value a site names.
//
// `o.a = v` mutates `o`, not `v`. Upstream's effect for it is `Mutate { value: object }`, and the
// widening Stage 1 will apply extends the RECEIVER's range. A site naming the stored value instead
// would hand Stage 1 the wrong identifier and widen a range that was never mutated, which is the
// unsafe direction: a value would look mutable past the point it settled.
//
// This exists because a mutation swapping `value.Object` for `value.Value` at the PropertyStore site
// survived every other assertion in this file. The kind assertions could not see it, which is the
// "fixtures assert the wrong layer" shape: they check that a site was found and never what it says.
func TestMutationSitesNameTheMutatedValueNotTheStoredOne(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `
export function f(o: {a: number}) {
  const v = 1;
  o.a = v;
  return o;
}
`)

	var checked int
	for _, site := range MutationSites(function) {
		if site.Kind != MutationSitePropertyStore {
			continue
		}
		instruction := function.Instructions[site.Instruction]
		store, ok := instruction.Value.(*PropertyStore)
		if !ok {
			t.Fatalf("a property-store site points at a %T", instruction.Value)
		}
		checked++
		if site.Target != store.Object.Identifier {
			t.Errorf("the site names value %d; a property store mutates its RECEIVER (%d), not the "+
				"value it stores (%d). Handing Stage 1 the stored value would widen the range of "+
				"something that was never mutated",
				site.Target, store.Object.Identifier, store.Value.Identifier)
		}
		if site.Target == store.Value.Identifier {
			t.Error("the site names the STORED value, which is the mutant this test exists for")
		}
	}
	if checked == 0 {
		t.Fatal("no property-store site was found, so this test measured nothing")
	}
}

// TestRangesStayValidAcrossALoopBackEdge pins the one divergence Stage 2.5 introduces.
//
// A loop-carried value is defined at a HIGH evaluation order by the back-edge store, while the
// mutation that reaches it through the loop's phi happened at a LOWER one. Both upstreams then write
// `start` and `end` from independently-guarded branches and produce an interval whose end is at or
// before its start, which React's own `validateMutableRange` declares invalid and which its config
// does not check by default.
//
// This input is the minimal reproduction, found by shrinking a real corpus case: before the clamp
// `i` came out as the empty interval [23,23), and 48 values across 400 corpus files were in that
// state. The assertion is the invariant itself rather than a specific interval, because the exact
// orders move whenever lowering changes and an assertion on them would fail for the wrong reason.
func TestRangesStayValidAcrossALoopBackEdge(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
declare function mutate(x: number[]): void;
export function f(items: number[][]) {
  for (let i = 0; i < items.length; i++) {
    const row = items[i];
    mutate(row);
  }
  return items;
}
`)
	if invalid := mutation_aliasing.ValidateMutableRanges(ranges); len(invalid) != 0 {
		for _, id := range invalid {
			r := ranges.Get(id)
			t.Errorf("value %d has the invalid range [%d,%d); a loop-carried value whose end was "+
				"widened below its own definition point must still come out non-empty", id, r.Start, r.End)
		}
	}

	// Two-sided: the loop body's value really is widened, so this input exercises the widening
	// rather than passing because nothing happened.
	widest := static_single_assignment.EvaluationOrder(0)
	for _, identifier := range function.Identifiers {
		if identifier == nil {
			continue
		}
		if r := ranges.Get(identifier.Id); r.IsSet() && r.End-r.Start > widest {
			widest = r.End - r.Start
		}
	}
	if widest <= 1 {
		t.Errorf("the widest range in this function is %d, so nothing was widened and the "+
			"invariant assertion above passed vacuously", widest)
	}

	// # The loop counter's range must cross the whole loop, which is what pins the BACK-EDGE PHI
	//
	// A phi operand whose predecessor block has not been walked yet is deferred to that block and
	// applied there with the index recorded at DEFERRAL, which is how a back edge gets an index low
	// enough for a mutation later in the loop to travel through it. Inverting that deferral test
	// leaves the loop counter at [3,4) instead of [3,23): the widening never crosses the back edge,
	// so every value defined before the loop reads as settled inside it.
	//
	// # Assert on the NARROWEST counter value, not the widest, and that distinction is the whole test
	//
	// A first version of this assertion took the WIDEST range among the values named `i` and
	// survived the mutation, because single-assignment form gives the counter several identifiers
	// and the mutant preserves the widest one. Measured on this input, base against mutant:
	//
	//	id 25  `i`  base [3,23)   mutant [3,4)    the value `i` is initialised to
	//	id 26  `i`  base [4,23)   mutant [4,23)   the loop-header phi
	//
	// The phi keeps its widening either way, because the mutation inside the loop reaches it
	// directly. What the back edge carries is the widening travelling BACK to the value the counter
	// held before the loop, and only the initial value can see that. So the narrowest is the
	// discriminating quantity and the widest is the one that cannot see the guard at all.
	//
	// Recorded because the brief's rule applied twice here: the first fixture written for this
	// survivor did not kill it, which means the hypothesis rather than the fixture was wrong, and
	// the fix was to read the two tables side by side instead of writing a second guess.
	// Counted rather than reduced to an extreme, because neither extreme discriminates: the widest
	// is preserved by the mutant and the narrowest belongs to the post-increment result, which is
	// legitimately one wide in both. Two of the four values named `i` cross the loop when the back
	// edge carries the widening and one does when it does not.
	wideCounters := 0
	var sawCounter bool
	for _, identifier := range function.Identifiers {
		if identifier == nil || identifier.Name != "i" {
			continue
		}
		r := ranges.Get(identifier.Id)
		if !r.IsSet() {
			continue
		}
		sawCounter = true
		if r.End-r.Start >= 10 {
			wideCounters++
		}
	}
	if !sawCounter {
		t.Fatal("no value named `i` carries a range, so this measured nothing")
	}
	if wideCounters < 2 {
		t.Errorf("only %d of the values the loop counter `i` takes stay mutable across the loop, "+
			"want at least 2. The loop-header phi learns it from the mutation directly; the value "+
			"feeding that phi can only learn it through the BACK EDGE. A phi operand applied with "+
			"a fresh index instead of its deferred one leaves exactly one", wideCounters)
	}
}

// TestRangesThirdLoopOpensAWidenedOperand pins the loop Stage 2 omitted and Stage 2.5 restored.
//
// # Why this needs a PARAMETER specifically
//
// Stage 2 omitted upstream's third loop, over operands rather than lvalues, after measuring that no
// value in the corpus reached the state it exists for -- `End > order && Start == 0`, a value whose
// end was widened while its start was never opened. That measurement was correct, and Stage 2 wrote
// down the condition on which it would expire: a widening pass reaching a value through the alias
// graph. Re-measured with the widening present, the state occurs 39,619 times over the same 400
// files, so the loop is doing real work rather than being restored on principle.
//
// A parameter is the clean case and the reason is structural. Every ordinary value is defined by an
// instruction lvalue, so the lvalue loop opens its start and the third loop has nothing to do. A
// parameter is never an lvalue of any instruction -- it arrives already defined -- so when a mutation
// widens its end, NOTHING else can give it a start. Measured on this input: the widening alone
// leaves `p` at [0,4), which `IsSet` reads as unset because both fields being zero is the unset
// range, and the third loop is what turns it into [2,4).
//
// So the assertion is that `p` carries a range at all, plus that the range is genuinely wide. A
// version of this test asserting only that some range exists would pass without the widening too.
func TestRangesThirdLoopOpensAWidenedOperand(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
declare function mutate(x: number[]): void;
export function f(p: number[]) {
  mutate(p);
  return p;
}
`)

	var parameter static_single_assignment.IdentifierId
	var found bool
	for _, place := range function.Params {
		for _, identifier := range function.Identifiers {
			if identifier != nil && identifier.Id == place.Identifier && identifier.Name == "p" {
				parameter = place.Identifier
				found = true
			}
		}
	}
	if !found {
		t.Fatal("the parameter `p` was not found among the function's params, so this measured nothing")
	}

	r := ranges.Get(parameter)
	if !r.IsSet() {
		t.Fatalf("the parameter `p` carries no range. Its end was widened by the mutation but only "+
			"the third loop can give a parameter a start, because a parameter is never an "+
			"instruction lvalue; without that loop the range stays [0,%d) and reads as unset", r.End)
	}
	if r.Start == 0 {
		t.Errorf("`p`'s range is [%d,%d) with an unopened start, which is exactly the state the "+
			"third loop exists to repair", r.Start, r.End)
	}
	if r.End <= r.Start+1 {
		t.Errorf("`p`'s range is [%d,%d), so it was never widened and this test would pass without "+
			"the third loop doing anything", r.Start, r.End)
	}
}

// TestRangesDoNotWidenThroughAnEdgeThatDidNotExistYet pins the sequence-index guard.
//
// # What the guard is and why a name-based assertion cannot see it
//
// Every edge and every mutation is stamped with a monotonically increasing index as the graph is
// built, and the walk refuses to traverse an edge whose index is at or after the mutation's own.
// That is upstream's model of time: a mutation flows only through aliases that already existed when
// it happened, so a value aliased AFTER being mutated is not retroactively widened.
//
// Removing that test does not change any NAMED value's range on a small input. It widens two
// unnamed temporaries that the faithful walk leaves alone, which is invisible to `rangeOfName` and
// is why a first version of this test, asserting on `inner` and `holder`, passed against the mutant.
// The assertion is therefore on the SIZE of the widened set, which is the level at which the
// difference actually appears: ten values here, twelve without the guard.
//
// Measured on the corpus rather than argued: over 959 functions in 200 files, dropping this guard
// changes 5,010 values, so it is load-bearing rather than an optimization.
func TestRangesDoNotWidenThroughAnEdgeThatDidNotExistYet(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `
declare function deep(x: unknown): void;
export function f() {
  const inner: number[] = [];
  const holder: unknown[] = [inner];
  deep(holder);
  return inner;
}
`)
	effects := InferAliasingEffects(function)
	widened := &mutation_aliasing.MutableRanges{}
	mutation_aliasing.BuildAliasingGraph(newRangeGraph(function, effects), function, mutation_aliasing.Options{}).Widen(widened)

	// The count is what separates the faithful walk from one that ignores the index. It is asserted
	// exactly rather than as a bound, because a bound would be satisfied by a walk that widened
	// nothing at all.
	//
	// Ten until the frozen-capture rule in `ranges.go`, eight after it, and the two that leave are
	// not values this test is about. Traced before the count was changed:
	//
	//	id 19  the `LoadGlobal` temporary holding `deep`, created frozen
	//	id 21  the result of `deep(holder)`
	//
	// Both named values still widen to exactly `[0,9)`: `inner` and `holder` are byte-identical
	// before and after. What stopped is the `Capture` from the frozen global into `holder`, which
	// upstream also prunes -- its `Capture` arm maps a `Global` source to a null `sourceType` that
	// matches none of its three branches (`InferMutationAliasingEffects.ts:901`), and a `Frozen`
	// source to `ImmutableCapture`. LoadGlobal now carries Global directly and takes the first route.
	//
	// Widening a global was never meaningful, and the sequence-index guard this test exists for is
	// untouched: without it the count rises by two from whatever the baseline is, which is still
	// what the assertion below catches.
	const wantWidened = 8
	if got := widened.Len(); got != wantWidened {
		t.Errorf("the widening reached %d values, want %d. More than %d means the walk followed an "+
			"edge created after the mutation it is propagating, which is the sequence-index guard "+
			"failing; fewer means it stopped early", got, wantWidened, wantWidened)
	}
}

// TestRangesFollowCapturesOnlyForATransitiveMutation pins the capture-direction guard.
//
// # The distinction the guard makes
//
// A capture edge records that information flowed from one value into another without aliasing them.
// Mutating a container does NOT mutate what it holds, so a plain mutation must not walk a capture
// edge backwards. A TRANSITIVE mutation does reach everything the value transitively holds, so it
// must. Upstream gates the backward capture walk on `entry.transitive` for exactly that reason.
//
// # Finding an input that can see it took a signature-table method, and that is the point
//
// Six hand-written shapes -- an object literal holding an array, a two-level nest, a closure
// capturing an accumulator -- all failed to distinguish the guarded walk from the unguarded one,
// because an unknown call takes the default path and emits `MutateTransitiveConditionally` on every
// operand, which makes the mutation transitive anyway and the guard vacuous.
//
// `Map.prototype.set` is in oxc's global signature table with a `Capture` effect, so it produces a
// real capture edge alongside a NON-transitive mutation, which is the only shape on which the two
// walks disagree. Measured here: `seed` ends at 15 with the guard and at 21 without it, so dropping
// the guard over-widens a captured parameter by six positions.
//
// Recorded at length because the brief's rule -- name the input on which the two versions produce
// different output, before writing a fixture -- is what turned this from a survivor into a test. The
// first four hypotheses about which shape would distinguish them were all wrong.
func TestRangesFollowCapturesOnlyForATransitiveMutation(t *testing.T) {
	t.Parallel()

	function, ranges := rangesFor(t, `
export function f(seed: number[]) {
  const m = new Map<string, number[]>();
  m.set("k", seed);
  const arr = Array.from(m.values());
  arr.push([1]);
  return m;
}
`)

	var seed static_single_assignment.IdentifierId
	var found bool
	for _, place := range function.Params {
		for _, identifier := range function.Identifiers {
			if identifier != nil && identifier.Id == place.Identifier && identifier.Name == "seed" {
				seed, found = place.Identifier, true
			}
		}
	}
	if !found {
		t.Fatal("the parameter `seed` was not found, so this measured nothing")
	}

	r := ranges.Get(seed)
	if !r.IsSet() {
		t.Fatal("`seed` carries no range, so the capture edge never reached it and this test " +
			"cannot see the guard it exists for")
	}

	// The later `arr.push` mutates a value `seed` was captured INTO, non-transitively. A walk that
	// followed the capture edge backwards anyway would carry that mutation back to `seed`.
	const wantEnd = 15
	if r.End != wantEnd {
		t.Errorf("`seed`'s range is [%d,%d), want an end of %d. A larger end means the walk "+
			"followed a capture edge backwards for a NON-transitive mutation, which widens a "+
			"contained value on a mutation of its container", r.Start, r.End, wantEnd)
	}
}
