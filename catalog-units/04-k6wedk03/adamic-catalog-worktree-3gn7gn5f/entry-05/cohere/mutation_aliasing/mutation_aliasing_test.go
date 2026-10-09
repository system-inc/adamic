package mutation_aliasing

import (
	"strings"
	"testing"

	ssa "github.com/system-inc/cohere/static_single_assignment"
)

// The pass is tested here on an IR of the tests' own, as small as Graph allows, so the module proves
// itself without either real IR. Functions are built already in single assignment form, one value
// per id, and the effects are written by hand, which is what lets a test pin one seam at a time.
// cohere's high-level IR tests the same pass through its adapter over lowered source, and Adamic's
// suite through its own.

type testPlace struct{ id ssa.IdentifierId }

type testInstruction struct {
	uses, defines []testPlace
	effects       []AliasingEffect[testPlace]
	// storedContext is the value a store into a captured binding stores, when this is one.
	storedContext *ssa.IdentifierId
	order         ssa.EvaluationOrder
}

type testBlock struct {
	id            ssa.BlockId
	instructions  []*testInstruction
	successors    []ssa.BlockId
	returns       *testPlace
	predecessors  []ssa.BlockId
	phis          []*ssa.Phi[testPlace]
	terminalOrder ssa.EvaluationOrder
}

type testFunction struct {
	entry            ssa.BlockId
	blocks           []*testBlock
	table            map[ssa.BlockId]*testBlock
	next             ssa.BlockId
	params           []testPlace
	context          []testPlace
	returns          *testPlace
	parametersFrozen bool
}

func newTestFunction() *testFunction {
	return &testFunction{table: map[ssa.BlockId]*testBlock{}, next: 1}
}

func (f *testFunction) block() *testBlock {
	block := &testBlock{id: f.next}
	f.next++
	f.blocks = append(f.blocks, block)
	f.table[block.id] = block
	if f.entry == 0 {
		f.entry = block.id
	}
	return block
}

// add appends an instruction to a block.
func (b *testBlock) add(instruction *testInstruction) *testInstruction {
	b.instructions = append(b.instructions, instruction)
	return instruction
}

func place(id ssa.IdentifierId) testPlace { return testPlace{id: id} }

type testGraph struct{}

func (testGraph) Entry(f *testFunction) ssa.BlockId { return f.entry }
func (testGraph) BlockBound(f *testFunction) int    { return int(f.next) }
func (testGraph) Block(f *testFunction, id ssa.BlockId) (*testBlock, bool) {
	block, ok := f.table[id]
	return block, ok
}
func (testGraph) Blocks(f *testFunction) []*testBlock            { return f.blocks }
func (testGraph) SetBlocks(f *testFunction, blocks []*testBlock) { f.blocks = blocks }
func (testGraph) Retain(f *testFunction, keep func(ssa.BlockId) bool) {
	for id := range f.table {
		if !keep(id) {
			delete(f.table, id)
		}
	}
}
func (testGraph) Placeholder(f *testFunction, block *testBlock) *testBlock { return block }
func (testGraph) Id(block *testBlock) ssa.BlockId                          { return block.id }
func (testGraph) Predecessors(block *testBlock) []ssa.BlockId              { return block.predecessors }
func (testGraph) SetPredecessors(block *testBlock, predecessors []ssa.BlockId) {
	block.predecessors = predecessors
}
func (testGraph) Phis(block *testBlock) []*ssa.Phi[testPlace]          { return block.phis }
func (testGraph) SetPhis(block *testBlock, phis []*ssa.Phi[testPlace]) { block.phis = phis }
func (testGraph) EachEdge(block *testBlock, visit func(ssa.BlockId, ssa.Edge)) {
	for _, successor := range block.successors {
		visit(successor, ssa.Real)
	}
}
func (testGraph) EndsInReturn(block *testBlock) bool { return block.returns != nil }
func (testGraph) InstructionCount(f *testFunction, block *testBlock) int {
	return len(block.instructions)
}
func (testGraph) EachInstructionPlace(f *testFunction, block *testBlock, index int, visit func(*testPlace, ssa.Role)) {
	instruction := block.instructions[index]
	for i := range instruction.uses {
		visit(&instruction.uses[i], ssa.Use)
	}
	for i := range instruction.defines {
		visit(&instruction.defines[i], ssa.Define)
	}
}
func (testGraph) IsContextStore(f *testFunction, block *testBlock, index int) bool { return false }
func (testGraph) ContextStoreDefines(f *testFunction, block *testBlock, index int, place testPlace) bool {
	return false
}
func (testGraph) SetInstructionOrder(f *testFunction, block *testBlock, index int, order ssa.EvaluationOrder) {
	block.instructions[index].order = order
}
func (testGraph) EachTerminalPlace(block *testBlock, visit func(*testPlace, ssa.Role)) {}
func (testGraph) SetTerminalOrder(block *testBlock, order ssa.EvaluationOrder) {
	block.terminalOrder = order
}
func (testGraph) Params(f *testFunction) []testPlace                                 { return f.params }
func (testGraph) Returns(f *testFunction) *testPlace                                 { return f.returns }
func (testGraph) Declaration(f *testFunction, id ssa.IdentifierId) ssa.DeclarationId { return 0 }
func (testGraph) Contextual(f *testFunction, declaration ssa.DeclarationId) bool     { return false }
func (testGraph) Mint(f *testFunction, original ssa.IdentifierId) ssa.IdentifierId   { return original }
func (testGraph) Named(f *testFunction, id ssa.IdentifierId) bool                    { return true }
func (testGraph) PlaceString(f *testFunction, id ssa.IdentifierId) string            { return "" }
func (testGraph) IdentifierOf(place testPlace) ssa.IdentifierId                      { return place.id }
func (testGraph) WithIdentifier(place testPlace, id ssa.IdentifierId) testPlace {
	return testPlace{id: id}
}

func (testGraph) InstructionOrder(f *testFunction, block *testBlock, index int) (ssa.EvaluationOrder, bool) {
	return block.instructions[index].order, true
}
func (testGraph) TerminalOrder(block *testBlock) ssa.EvaluationOrder { return block.terminalOrder }
func (testGraph) Effects(f *testFunction, block *testBlock, index int) []AliasingEffect[testPlace] {
	return block.instructions[index].effects
}
func (testGraph) ParametersFrozen(f *testFunction) bool { return f.parametersFrozen }
func (testGraph) Context(f *testFunction) []testPlace   { return f.context }
func (testGraph) ReturnValue(block *testBlock) (testPlace, bool) {
	if block.returns == nil {
		return testPlace{}, false
	}
	return *block.returns, true
}
func (testGraph) StoredContextValue(f *testFunction, block *testBlock, index int) (ssa.IdentifierId, bool) {
	if stored := block.instructions[index].storedContext; stored != nil {
		return *stored, true
	}
	return 0, false
}
func (testGraph) Closure(f *testFunction, block *testBlock, index int, into ssa.IdentifierId,
	kinds map[ssa.IdentifierId]EffectValueKind) ([]testPlace, bool) {
	return nil, false
}

// finalize is the sequence both real IRs run before the pass: reverse postorder, predecessors, and
// evaluation order from 1.
func finalize(f *testFunction) {
	ssa.ReversePostorder(testGraph{}, f)
	ssa.MarkPredecessors(testGraph{}, f)
	ssa.MarkEvaluationOrder(testGraph{}, f)
}

func ranges(f *testFunction, options Options) *MutableRanges {
	finalize(f)
	return InferMutableRanges(testGraph{}, f, options)
}

func expectRange(t *testing.T, table *MutableRanges, id ssa.IdentifierId, want MutableRange, why string) {
	t.Helper()
	if got := table.Get(id); got != want {
		t.Errorf("value %d has range [%d,%d), want [%d,%d): %s", id, got.Start, got.End, want.Start, want.End, why)
	}
}

// TestParametersDefinedOnEntryBothWays pins Options.ParametersDefinedOnEntry by exact range, both
// ways, since until Adamic switches to this module it is the only gate on that option.
//
// The shape is the one that took Adamic's rule (its 4748a636): a parameter mutated through another
// value before its first read. Here p is mutated at order 1 by an instruction that reads only q (as
// Adamic's escaped-values rule does for a write into an escaped container q reaches), read at 2, and
// mutated again at 3.
//
//   - Without the option, the widening takes p's end to 4, and p first appears as an operand at order
//     2, where the operand loop opens its start: [2,4). The mutation at 1 is outside it.
//   - With it, p is defined on entry: [1,4), and the mutation at 1 is inside. r, a parameter nothing
//     touches, is unset without the option and is the first instruction alone with it, [1,2).
func TestParametersDefinedOnEntryBothWays(t *testing.T) {
	t.Parallel()
	build := func() (*testFunction, ssa.IdentifierId, ssa.IdentifierId, ssa.IdentifierId) {
		f := newTestFunction()
		p, q, r := ssa.IdentifierId(1), ssa.IdentifierId(2), ssa.IdentifierId(3)
		f.params = []testPlace{place(p), place(q), place(r)}
		entry := f.block()
		entry.add(&testInstruction{uses: []testPlace{place(q)},
			effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutateTransitiveConditionally, place(p))}})
		entry.add(&testInstruction{uses: []testPlace{place(p)}})
		entry.add(&testInstruction{uses: []testPlace{place(p)},
			effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutate, place(p))}})
		return f, p, q, r
	}

	f, p, q, r := build()
	without := ranges(f, Options{})
	if order := f.blocks[0].instructions[0].order; order != 1 {
		t.Fatalf("the first instruction is at order %d, so the ranges below would not mean what they say", order)
	}
	expectRange(t, without, p, MutableRange{Start: 2, End: 4},
		"React's rule opens a parameter's range at its first read, after the mutation at order 1")
	if without.Contains(p, 1) {
		t.Error("without the option, the mutation at order 1 must fall outside p's range: that is the hole the option closes")
	}
	expectRange(t, without, q, MutableRange{}, "q is read and never mutated")
	expectRange(t, without, r, MutableRange{}, "r is never touched")

	f, p, q, r = build()
	with := ranges(f, Options{ParametersDefinedOnEntry: true})
	expectRange(t, with, p, MutableRange{Start: 1, End: 4}, "a parameter is defined on entry")
	if !with.Contains(p, 1) {
		t.Error("with the option, the mutation at order 1 must be inside p's range")
	}
	expectRange(t, with, q, MutableRange{Start: 1, End: 2}, "a parameter nothing widened is the first instruction alone")
	expectRange(t, with, r, MutableRange{Start: 1, End: 2}, "a parameter nothing widened is the first instruction alone")
}

// TestAMutationReachesWhatItIsAnAliasOf: Alias is "mutating Into implies mutating From", so mutating
// the alias widens the source to the mutation.
func TestAMutationReachesWhatItIsAnAliasOf(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	source, alias := ssa.IdentifierId(1), ssa.IdentifierId(2)
	entry := f.block()
	entry.add(&testInstruction{defines: []testPlace{place(source)},
		effects: []AliasingEffect[testPlace]{CreateEffect(place(source), EffectValueMutable)}})
	entry.add(&testInstruction{uses: []testPlace{place(source)}, defines: []testPlace{place(alias)},
		effects: []AliasingEffect[testPlace]{
			CreateEffect(place(alias), EffectValueMutable),
			FlowEffect(AliasingEffectAlias, place(source), place(alias)),
		}})
	entry.add(&testInstruction{uses: []testPlace{place(alias)},
		effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutate, place(alias))}})

	table := ranges(f, Options{})
	expectRange(t, table, source, MutableRange{Start: 1, End: 4}, "the alias's mutation at 3 reaches its source")
	expectRange(t, table, alias, MutableRange{Start: 2, End: 4}, "mutated at 3")
}

// TestAnAliasMadeAfterAMutationDoesNotWiden is the model of time: a mutation flows only through edges
// that existed when it happened, so an alias made later does not reach back.
func TestAnAliasMadeAfterAMutationDoesNotWiden(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	source, alias := ssa.IdentifierId(1), ssa.IdentifierId(2)
	entry := f.block()
	entry.add(&testInstruction{defines: []testPlace{place(source)},
		effects: []AliasingEffect[testPlace]{CreateEffect(place(source), EffectValueMutable)}})
	entry.add(&testInstruction{defines: []testPlace{place(alias)},
		effects: []AliasingEffect[testPlace]{CreateEffect(place(alias), EffectValueMutable)}})
	entry.add(&testInstruction{uses: []testPlace{place(alias)},
		effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutate, place(alias))}})
	entry.add(&testInstruction{uses: []testPlace{place(source), place(alias)},
		effects: []AliasingEffect[testPlace]{FlowEffect(AliasingEffectAlias, place(source), place(alias))}})

	table := ranges(f, Options{})
	expectRange(t, table, source, MutableRange{Start: 1, End: 2}, "the alias came after the mutation, so the source is untouched")
	expectRange(t, table, alias, MutableRange{Start: 2, End: 4}, "mutated at 3")
}

// TestFrozenParametersDropAConditionalMutation pins Graph.ParametersFrozen: React's component
// parameters are Frozen, and a conditional mutation of a value that is not Mutable is dropped. The
// same function with mutable parameters widens.
func TestFrozenParametersDropAConditionalMutation(t *testing.T) {
	t.Parallel()
	for _, frozen := range []bool{true, false} {
		f := newTestFunction()
		f.parametersFrozen = frozen
		p := ssa.IdentifierId(1)
		f.params = []testPlace{place(p)}
		entry := f.block()
		entry.add(&testInstruction{uses: []testPlace{place(p)}})
		entry.add(&testInstruction{uses: []testPlace{place(p)},
			effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutateConditionally, place(p))}})

		table := ranges(f, Options{})
		want := MutableRange{Start: 1, End: 3}
		if frozen {
			want = MutableRange{}
		}
		expectRange(t, table, p, want, map[bool]string{true: "frozen parameters drop it", false: "mutable parameters widen"}[frozen])
	}
}

// TestAStoreIntoACapturedBindingWidensFromItsShape pins Graph.StoredContextValue: the stored value's
// range reaches past the store with no effect consulted.
func TestAStoreIntoACapturedBindingWidensFromItsShape(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	value := ssa.IdentifierId(1)
	entry := f.block()
	entry.add(&testInstruction{defines: []testPlace{place(value)}})
	entry.add(&testInstruction{uses: []testPlace{place(value)}})
	stored := value
	entry.add(&testInstruction{uses: []testPlace{place(value)}, storedContext: &stored})

	table := ranges(f, Options{})
	expectRange(t, table, value, MutableRange{Start: 1, End: 4}, "the store at 3 keeps it live past 3")
}

// TestMutatingAReturnedValueReachesTheReturnsPlace pins Graph.ReturnValue: a return aliases its value
// into the Returns place, and an IR without one skips both the node and the edge.
func TestMutatingAReturnedValueReachesTheReturnsPlace(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	value, returned := ssa.IdentifierId(1), ssa.IdentifierId(2)
	returns := place(returned)
	f.returns = &returns
	entry := f.block()
	entry.add(&testInstruction{defines: []testPlace{place(value)},
		effects: []AliasingEffect[testPlace]{CreateEffect(place(value), EffectValueMutable)}})
	entry.returns = &testPlace{id: value}

	finalize(f)
	graph := BuildAliasingGraph(testGraph{}, f, Options{})
	if node := graph.state.nodes[returned]; node == nil || len(node.aliasesOrder) != 1 || node.aliasesOrder[0] != value {
		t.Fatalf("the Returns place should alias the returned value, got %+v", node)
	}

	f.returns = nil
	graph = BuildAliasingGraph(testGraph{}, f, Options{})
	if _, ok := graph.state.nodes[returned]; ok {
		t.Error("an IR with no Returns place has no Returns node")
	}
}

// TestAMutationOfAPhiReachesItsOperands: a phi aliases each operand, so mutating the merge mutates
// whichever value it may be. Operands are walked in ascending predecessor order.
func TestAMutationOfAPhiReachesItsOperands(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	left, right, merged := ssa.IdentifierId(1), ssa.IdentifierId(2), ssa.IdentifierId(3)
	entry, leftBlock, rightBlock, join := f.block(), f.block(), f.block(), f.block()
	entry.successors = []ssa.BlockId{leftBlock.id, rightBlock.id}
	leftBlock.successors = []ssa.BlockId{join.id}
	rightBlock.successors = []ssa.BlockId{join.id}
	leftBlock.add(&testInstruction{defines: []testPlace{place(left)},
		effects: []AliasingEffect[testPlace]{CreateEffect(place(left), EffectValueMutable)}})
	rightBlock.add(&testInstruction{defines: []testPlace{place(right)},
		effects: []AliasingEffect[testPlace]{CreateEffect(place(right), EffectValueMutable)}})
	phi := &ssa.Phi[testPlace]{Place: place(merged)}
	phi.Operands.Set(leftBlock.id, place(left))
	phi.Operands.Set(rightBlock.id, place(right))
	join.phis = []*ssa.Phi[testPlace]{phi}
	join.add(&testInstruction{uses: []testPlace{place(merged)},
		effects: []AliasingEffect[testPlace]{MutationEffect(AliasingEffectMutate, place(merged))}})

	table := ranges(f, Options{})
	mutation := join.instructions[0].order
	for _, operand := range []ssa.IdentifierId{left, right} {
		if got := table.Get(operand); got.End != mutation+1 {
			t.Errorf("operand %d ends at %d, want %d, past the phi's mutation", operand, got.End, mutation+1)
		}
	}
	if invalid := ValidateMutableRanges(table); len(invalid) != 0 {
		t.Errorf("invalid ranges: %v", invalid)
	}
}

// TestRangesPhiOpensBeforeItsBlockWhenWidened proves the phi branch works, by seeding the state
// that only Stage 1 can produce today.
//
// `phiOpensBefore` is inert on this tree because its input is a widened `End` that only
// `AliasingState.mutate` writes. The branch is kept live rather than deleted, so it needs a test
// that does not depend on the missing input. Seeding the range by hand is what separates "this
// branch is correct and waiting" from "this branch is untested".
func TestRangesPhiOpensBeforeItsBlockWhenWidened(t *testing.T) {
	t.Parallel()

	if phiOpensBefore(MutableRange{}, 10) {
		t.Error("an unset range must not open a phi early")
	}
	if phiOpensBefore(MutableRange{Start: 5, End: 10}, 10) {
		t.Error("a range ending exactly at the block's first instruction is not mutated AFTER it; " +
			"upstream's test is strictly greater")
	}
	if !phiOpensBefore(MutableRange{Start: 5, End: 11}, 10) {
		t.Error("a range ending past the block's first instruction means the phi is mutated after " +
			"creation, so it must open early. This is the branch Stage 1 will make reachable")
	}
	if phiOpensBefore(MutableRange{Start: 1, End: 99}, 0) {
		t.Error("a block with no assigned order must not open a phi at a negative position")
	}
}

// TestPhiOpensOneBeforeItsBlock pins the offset upstream uses.
//
// React writes `makeInstructionId(firstInstructionIdOfBlock - 1)`. The minus one is what puts the
// merge point itself inside the range: a phi's value exists at the TOP of the block, before its
// first instruction runs, so opening at the instruction would make `Contains` answer false at the
// position the phi is definitionally live.
//
// Tested on the helper rather than through lowered source, because the branch only runs on a range
// whose end Stage 1 widened. A mutation dropping the minus one survived every fixture over real
// code for exactly that reason.
func TestPhiOpensOneBeforeItsBlock(t *testing.T) {
	t.Parallel()

	widened := MutableRange{Start: 0, End: 20}
	opened, moved := phiOpenedRange(widened, 10)
	if !moved {
		t.Fatal("a phi whose end is past its block's first instruction must open early")
	}
	if opened.Start != 9 {
		t.Errorf("the phi opened at %d; upstream opens at the block's first instruction MINUS ONE, "+
			"which is 9 here, so that the merge point itself is inside the range", opened.Start)
	}
	if !opened.Contains(9) || !opened.Contains(10) {
		t.Error("both the merge point and the block's first instruction must be inside the range")
	}

	if _, moved := phiOpenedRange(MutableRange{Start: 5, End: 20}, 10); moved {
		t.Error("a phi whose start is already set must not be reopened")
	}
	if _, moved := phiOpenedRange(MutableRange{}, 10); moved {
		t.Error("an unset range must not open a phi early")
	}
}

func TestPhiValueKindsPreserveMixedFrozenValues(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name        string
		left, right EffectValueKind
		want        EffectValueKind
		unseen      bool
	}{
		{"frozen mutable", EffectValueFrozen, EffectValueMutable, EffectValueMaybeFrozen, false},
		{"mutable frozen", EffectValueMutable, EffectValueFrozen, EffectValueMaybeFrozen, false},
		{"mixed primitive", EffectValueMaybeFrozen, EffectValuePrimitive, EffectValueMaybeFrozen, false},
		{"mixed global", EffectValueMaybeFrozen, EffectValueGlobal, EffectValueMaybeFrozen, false},
		{"mixed frozen", EffectValueMaybeFrozen, EffectValueFrozen, EffectValueMaybeFrozen, false},
		{"frozen primitive", EffectValueFrozen, EffectValuePrimitive, EffectValueFrozen, false},
		{"frozen global", EffectValueFrozen, EffectValueGlobal, EffectValueFrozen, false},
		{"global primitive", EffectValueGlobal, EffectValuePrimitive, EffectValueGlobal, false},
		{"global mutable", EffectValueGlobal, EffectValueMutable, EffectValueMutable, false},
		{"primitive mutable", EffectValuePrimitive, EffectValueMutable, EffectValueMutable, false},
		{"unvisited predecessor", EffectValueFrozen, EffectValueMutable, EffectValueMutable, true},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			state := newAliasingState()
			for index, kind := range []EffectValueKind{testCase.left, testCase.right} {
				id := ssa.IdentifierId(index + 1)
				state.create(id, aliasingNodeObject)
				if kind != EffectValueMutable {
					state.markImmutable(id, kind)
				}
			}
			phi := &ssa.Phi[testPlace]{Place: place(3), Operands: ssa.PhiOperands[testPlace]{
				{Predecessor: 1, Place: place(1)}, {Predecessor: 2, Place: place(2)},
			}}
			derivePhiImmutable(state, testGraph{}, phi, map[ssa.BlockId]bool{1: true, 2: !testCase.unseen})
			if got := state.immutable[3]; got != testCase.want {
				t.Fatalf("phi kind=%s, want %s", got, testCase.want)
			}
		})
	}
}

func TestFreezeRetainsAlreadyImmutableKinds(t *testing.T) {
	t.Parallel()
	for _, kind := range []EffectValueKind{EffectValuePrimitive, EffectValueGlobal, EffectValueFrozen} {
		t.Run(kind.String(), func(t *testing.T) {
			t.Parallel()
			state := newAliasingState()
			id := ssa.IdentifierId(1)
			state.create(id, aliasingNodeObject)
			state.markImmutable(id, kind)
			if state.freeze(id) || state.immutable[id] != kind {
				t.Fatalf("freeze changed %s to %s", kind, state.immutable[id])
			}
		})
	}
}

// TestTheVocabularyNamesEveryKind keeps the String methods total, since a kind printing "<unknown>"
// in a test failure hides which effect was meant.
func TestTheVocabularyNamesEveryKind(t *testing.T) {
	t.Parallel()
	for kind := AliasingEffectCreate; kind <= AliasingEffectApply; kind++ {
		if name := kind.String(); name == "<unknown>" || strings.TrimSpace(name) == "" {
			t.Errorf("effect kind %d has no name", kind)
		}
	}
	if AliasingEffectKind(AliasingEffectApply+1).String() != "<unknown>" {
		t.Error("a kind past the last should print <unknown>")
	}
}
