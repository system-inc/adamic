package static_single_assignment_test

import (
	"fmt"
	"slices"
	"testing"

	ssa "github.com/system-inc/cohere/static_single_assignment"
)

// The passes are tested here on an IR of the tests' own, as small as Graph allows, so the module
// proves itself without either real IR. cohere's high-level IR and its React rules test the same
// passes through cohere's adapter, and Adamic's suite through its own.

// testPlace carries a tag beside its value, standing in for what a real IR's place carries (cohere's
// effect, reactivity and range), so a test can see it ride along through renaming.
type testPlace struct {
	id  ssa.IdentifierId
	tag string
}

type testInstruction struct {
	uses, defines []testPlace
	contextStore  bool
	// result is the definition a context store's own result is, the one the verifier counts.
	result ssa.IdentifierId
	order  ssa.EvaluationOrder
}

type testEdge struct {
	to   ssa.BlockId
	kind ssa.Edge
}

type testBlock struct {
	id            ssa.BlockId
	instructions  []*testInstruction
	edges         []testEdge
	returns       bool
	predecessors  []ssa.BlockId
	phis          []*ssa.Phi[testPlace]
	terminalOrder ssa.EvaluationOrder
	placeholder   bool
}

type testIdentifier struct {
	declaration ssa.DeclarationId
	name        string
}

type testFunction struct {
	entry       ssa.BlockId
	blocks      []*testBlock
	table       map[ssa.BlockId]*testBlock
	next        ssa.BlockId
	identifiers []testIdentifier
	params      []testPlace
	returns     *testPlace
	contextual  map[ssa.DeclarationId]bool
	// variables maps a name to its declaration, for building.
	variables map[string]ssa.DeclarationId
}

func newTestFunction() *testFunction {
	function := &testFunction{
		table:      map[ssa.BlockId]*testBlock{},
		next:       1,
		contextual: map[ssa.DeclarationId]bool{},
		variables:  map[string]ssa.DeclarationId{},
	}
	// Identifier 0 is never a variable's, as in both real IRs.
	function.identifiers = append(function.identifiers, testIdentifier{})
	return function
}

func (f *testFunction) block() *testBlock {
	block := &testBlock{id: f.next}
	f.next++
	f.blocks = append(f.blocks, block)
	f.table[block.id] = block
	return block
}

// value mints a new pre-assignment value of a variable, as lowering mints one per store and per read.
func (f *testFunction) value(name string) testPlace {
	declaration, ok := f.variables[name]
	if !ok {
		declaration = ssa.DeclarationId(len(f.variables) + 1)
		f.variables[name] = declaration
	}
	id := ssa.IdentifierId(len(f.identifiers))
	f.identifiers = append(f.identifiers, testIdentifier{declaration: declaration, name: name})
	return testPlace{id: id, tag: name}
}

func (f *testFunction) define(block *testBlock, name string) *testInstruction {
	instruction := &testInstruction{defines: []testPlace{f.value(name)}}
	block.instructions = append(block.instructions, instruction)
	return instruction
}

func (f *testFunction) use(block *testBlock, name string) *testInstruction {
	instruction := &testInstruction{uses: []testPlace{f.value(name)}}
	block.instructions = append(block.instructions, instruction)
	return instruction
}

func (f *testFunction) name(id ssa.IdentifierId) string { return f.identifiers[id].name }

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
func (testGraph) Placeholder(f *testFunction, block *testBlock) *testBlock {
	placeholder := &testBlock{id: block.id, placeholder: true,
		predecessors: append([]ssa.BlockId(nil), block.predecessors...)}
	f.table[block.id] = placeholder
	return placeholder
}
func (testGraph) Id(block *testBlock) ssa.BlockId             { return block.id }
func (testGraph) Predecessors(block *testBlock) []ssa.BlockId { return block.predecessors }
func (testGraph) SetPredecessors(block *testBlock, predecessors []ssa.BlockId) {
	block.predecessors = predecessors
}
func (testGraph) Phis(block *testBlock) []*ssa.Phi[testPlace]          { return block.phis }
func (testGraph) SetPhis(block *testBlock, phis []*ssa.Phi[testPlace]) { block.phis = phis }
func (testGraph) EachEdge(block *testBlock, visit func(ssa.BlockId, ssa.Edge)) {
	for _, edge := range block.edges {
		visit(edge.to, edge.kind)
	}
}
func (testGraph) EndsInReturn(block *testBlock) bool { return block.returns }
func (testGraph) InstructionCount(f *testFunction, block *testBlock) int {
	return len(block.instructions)
}
func (testGraph) EachInstructionPlace(f *testFunction, block *testBlock, index int,
	visit func(*testPlace, ssa.Role)) {
	instruction := block.instructions[index]
	for i := range instruction.uses {
		visit(&instruction.uses[i], ssa.Use)
	}
	for i := range instruction.defines {
		visit(&instruction.defines[i], ssa.Define)
	}
}
func (testGraph) IsContextStore(f *testFunction, block *testBlock, index int) bool {
	return block.instructions[index].contextStore
}
func (testGraph) ContextStoreDefines(f *testFunction, block *testBlock, index int, place testPlace) bool {
	return place.id == block.instructions[index].result
}
func (testGraph) SetInstructionOrder(f *testFunction, block *testBlock, index int, order ssa.EvaluationOrder) {
	block.instructions[index].order = order
}
func (testGraph) EachTerminalPlace(block *testBlock, visit func(*testPlace, ssa.Role)) {}
func (testGraph) SetTerminalOrder(block *testBlock, order ssa.EvaluationOrder) {
	block.terminalOrder = order
}
func (testGraph) Params(f *testFunction) []testPlace { return f.params }
func (testGraph) Returns(f *testFunction) *testPlace { return f.returns }
func (testGraph) Declaration(f *testFunction, id ssa.IdentifierId) ssa.DeclarationId {
	return f.identifiers[id].declaration
}
func (testGraph) Contextual(f *testFunction, declaration ssa.DeclarationId) bool {
	return f.contextual[declaration]
}
func (testGraph) Mint(f *testFunction, original ssa.IdentifierId) ssa.IdentifierId {
	source := f.identifiers[original]
	id := ssa.IdentifierId(len(f.identifiers))
	f.identifiers = append(f.identifiers, source)
	return id
}
func (testGraph) Named(f *testFunction, id ssa.IdentifierId) bool {
	return f.identifiers[id].name != ""
}
func (testGraph) PlaceString(f *testFunction, id ssa.IdentifierId) string {
	return fmt.Sprintf("%s$%d", f.identifiers[id].name, id)
}
func (testGraph) IdentifierOf(place testPlace) ssa.IdentifierId { return place.id }
func (testGraph) WithIdentifier(place testPlace, id ssa.IdentifierId) testPlace {
	place.id = id
	return place
}

// finalize and construct are the sequence both real IRs run.
func finalize(f *testFunction) {
	ssa.ReversePostorder(testGraph{}, f)
	ssa.MarkPredecessors(testGraph{}, f)
	ssa.MarkEvaluationOrder(testGraph{}, f)
}

func construct(t *testing.T, f *testFunction) {
	t.Helper()
	finalize(f)
	ssa.Construct(testGraph{}, f)
	if violations := ssa.VerifySSA(testGraph{}, f); len(violations) != 0 {
		t.Fatalf("not in single assignment form: %v", violations)
	}
}

func goTo(from *testBlock, to ...*testBlock) {
	for _, block := range to {
		from.edges = append(from.edges, testEdge{to: block.id, kind: ssa.Real})
	}
}

// TestDiamondPlacesOnePhiAtTheJoin: x is written before an if and on one arm, so the join merges two
// values, and the read after it names the phi.
func TestDiamondPlacesOnePhiAtTheJoin(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	entry, left, right, join := f.block(), f.block(), f.block(), f.block()
	f.entry = entry.id
	f.define(entry, "x")
	f.define(left, "x")
	read := f.use(join, "x")
	goTo(entry, left, right)
	goTo(left, join)
	goTo(right, join)

	construct(t, f)

	if len(join.phis) != 1 {
		t.Fatalf("join has %d phis, want 1", len(join.phis))
	}
	phi := join.phis[0]
	if got := ssa.PhiOperandsInOrder(phi); !slices.Equal(got, []ssa.BlockId{left.id, right.id}) {
		t.Errorf("operands from %v, want [%d %d]", got, left.id, right.id)
	}
	if read.uses[0].id != phi.Place.id {
		t.Errorf("the read after the join names $%d, want the phi's $%d", read.uses[0].id, phi.Place.id)
	}
	if left.instructions[0].defines[0].id == entry.instructions[0].defines[0].id {
		t.Errorf("the two writes of x share a value")
	}
	// The place's other fields ride along on the phi and its operands.
	for _, operand := range phi.Operands {
		if operand.Place.tag != "x" {
			t.Errorf("operand from bb%d lost its tag: %+v", operand.Predecessor, operand.Place)
		}
	}
	if phi.Place.tag != "x" {
		t.Errorf("phi lost its tag: %+v", phi.Place)
	}
}

// TestLoopPhiTakesTheBackEdge: i is written before a loop and in its body, so the header merges the
// entry's value with the back edge's, an incomplete phi filled when the body is sealed.
func TestLoopPhiTakesTheBackEdge(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	entry, header, body, exit := f.block(), f.block(), f.block(), f.block()
	f.entry = entry.id
	f.define(entry, "i")
	test := f.use(header, "i")
	f.define(body, "i")
	after := f.use(exit, "i")
	goTo(entry, header)
	goTo(header, body, exit)
	goTo(body, header)

	construct(t, f)

	if len(header.phis) != 1 {
		t.Fatalf("header has %d phis, want 1", len(header.phis))
	}
	phi := header.phis[0]
	if got := phi.Operands.At(body.id).id; got != body.instructions[0].defines[0].id {
		t.Errorf("the back edge's operand is $%d, want the body's write $%d", got, body.instructions[0].defines[0].id)
	}
	if test.uses[0].id != phi.Place.id || after.uses[0].id != phi.Place.id {
		t.Errorf("reads in and after the loop don't name the header's phi")
	}
}

// TestLoopThatNeverWritesLeavesNoPhi: the header's incomplete phi takes the entry's value and its own
// result, which is redundant, and elimination removes it and rewrites its reads.
func TestLoopThatNeverWritesLeavesNoPhi(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	entry, header, body, exit := f.block(), f.block(), f.block(), f.block()
	f.entry = entry.id
	f.define(entry, "i")
	test := f.use(header, "i")
	f.use(body, "i")
	goTo(entry, header)
	goTo(header, body, exit)
	goTo(body, header)

	construct(t, f)

	if len(header.phis) != 0 {
		t.Fatalf("header kept %d phis, want none", len(header.phis))
	}
	if want := entry.instructions[0].defines[0].id; test.uses[0].id != want {
		t.Errorf("the loop's read names $%d, want the entry's write $%d", test.uses[0].id, want)
	}
}

// TestUnreachableBlockIsDropped: a block no edge reaches leaves the slice and the table.
func TestUnreachableBlockIsDropped(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	entry, exit, orphan := f.block(), f.block(), f.block()
	f.entry = entry.id
	goTo(entry, exit)
	goTo(orphan, exit)

	finalize(f)

	for _, block := range f.blocks {
		if block.id == orphan.id {
			t.Fatalf("the unreachable block is still in the slice")
		}
	}
	if _, ok := (testGraph{}).Block(f, orphan.id); ok {
		t.Errorf("the unreachable block is still in the table")
	}
	if !slices.Equal(exit.predecessors, []ssa.BlockId{entry.id}) {
		t.Errorf("exit's predecessors are %v, want only the entry", exit.predecessors)
	}
}

// TestFallthroughNothingReachesBecomesAPlaceholder: a structured terminal names a continuation no
// real edge reaches, so it stays, empty, and is nobody's successor.
func TestFallthroughNothingReachesBecomesAPlaceholder(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	entry, body, continuation := f.block(), f.block(), f.block()
	f.entry = entry.id
	entry.edges = []testEdge{{to: continuation.id, kind: ssa.Fallthrough}, {to: body.id, kind: ssa.Real}}
	body.returns = true

	finalize(f)

	var kept *testBlock
	for _, block := range f.blocks {
		if block.id == continuation.id {
			kept = block
		}
	}
	if kept == nil || !kept.placeholder {
		t.Fatalf("the fallthrough is %+v, want a placeholder in the slice", kept)
	}
	if len(kept.predecessors) != 0 {
		t.Errorf("a fallthrough made a predecessor: %v", kept.predecessors)
	}
	if !slices.Equal(body.predecessors, []ssa.BlockId{entry.id}) {
		t.Errorf("body's predecessors are %v", body.predecessors)
	}
}

// TestContextStoreReusesItsDefinition: a binding a closure captures is defined once, and a second
// store to it in the same block names the same value, which the verifier accepts on the store's own
// result alone.
func TestContextStoreReusesItsDefinition(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	entry := f.block()
	f.entry = entry.id
	first := f.define(entry, "n")
	first.contextStore = true
	f.contextual[f.variables["n"]] = true
	second := f.define(entry, "n")
	second.contextStore = true

	construct(t, f)

	if first.defines[0].id != second.defines[0].id {
		t.Errorf("the context stores define $%d and $%d, want one value", first.defines[0].id, second.defines[0].id)
	}
}

// TestExceptionalEdgeIsAnEdge: MayThrow's handler is reached by an exceptional edge, which is a real
// predecessor today, so the handler's read resolves through it.
//
// It resolves to the block's END: x$2, written by the throwing instruction itself, which never
// completed when the handler runs. That is #2yz9ra9. Its fix belongs in valueFromPredecessor, where
// every lookup across an edge now goes, and flips the want below to the write before the throw.
func TestExceptionalEdgeIsAnEdge(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	entry, next, handler := f.block(), f.block(), f.block()
	f.entry = entry.id
	before := f.define(entry, "x")
	throwing := f.define(entry, "x")
	read := f.use(handler, "x")
	entry.edges = []testEdge{{to: next.id, kind: ssa.Real}, {to: handler.id, kind: ssa.Exceptional}}

	construct(t, f)

	if !slices.Equal(handler.predecessors, []ssa.BlockId{entry.id}) {
		t.Fatalf("the handler's predecessors are %v, want the throwing block", handler.predecessors)
	}
	if got, want := read.uses[0].id, throwing.defines[0].id; got != want {
		t.Errorf("the handler reads $%d, want $%d (the end of the block, #2yz9ra9; $%d once fixed)",
			got, want, before.defines[0].id)
	}
}

// TestEvaluationOrderFollowsTheSlice: instructions and terminals are numbered from 1 in reverse
// postorder, with nothing left at zero.
func TestEvaluationOrderFollowsTheSlice(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	entry, exit := f.block(), f.block()
	f.entry = entry.id
	f.define(entry, "a")
	f.use(exit, "a")
	goTo(entry, exit)

	finalize(f)

	var orders []ssa.EvaluationOrder
	for _, block := range f.blocks {
		for _, instruction := range block.instructions {
			orders = append(orders, instruction.order)
		}
		orders = append(orders, block.terminalOrder)
	}
	if !slices.Equal(orders, []ssa.EvaluationOrder{1, 2, 3, 4}) {
		t.Errorf("orders are %v, want [1 2 3 4]", orders)
	}
}

// TestVerifierCatchesAnUndominatedUse: a hand-broken function, a read on one arm of a diamond naming
// a value written on the other, is a violation, so a clean result elsewhere is a real one.
func TestVerifierCatchesAnUndominatedUse(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	entry, left, right := f.block(), f.block(), f.block()
	f.entry = entry.id
	write := f.define(left, "x")
	read := f.use(right, "x")
	read.uses[0].id = write.defines[0].id
	goTo(entry, left, right)
	finalize(f)

	violations := ssa.VerifySSA(testGraph{}, f)
	if len(violations) != 1 || violations[0].Kind != ssa.SSAViolationUseNotDominated {
		t.Fatalf("violations are %v, want one undominated use", violations)
	}
	if stats := ssa.CollectSSAStats(testGraph{}, f); stats.Uses != 1 || stats.NamedValues != 1 {
		t.Errorf("stats are %+v, want one use and one named value", stats)
	}
}

// TestDominanceIsTheEntrysAndTheJoins: the entry dominates every block, a branch arm dominates
// neither the other arm nor the join.
func TestDominanceIsTheEntrysAndTheJoins(t *testing.T) {
	t.Parallel()
	f := newTestFunction()
	entry, left, right, join := f.block(), f.block(), f.block(), f.block()
	f.entry = entry.id
	goTo(entry, left, right)
	goTo(left, join)
	goTo(right, join)
	finalize(f)

	dominance := ssa.ComputeDominance(testGraph{}, f)
	for _, block := range []*testBlock{entry, left, right, join} {
		if !dominance.Dominates(entry.id, block.id) {
			t.Errorf("the entry does not dominate bb%d", block.id)
		}
	}
	if dominance.Dominates(left.id, join.id) || dominance.Dominates(left.id, right.id) {
		t.Errorf("an arm dominates the join or the other arm")
	}
}
