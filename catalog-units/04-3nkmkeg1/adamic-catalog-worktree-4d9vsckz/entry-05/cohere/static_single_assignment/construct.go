// Single static assignment form: rename every definition so each value is written exactly once,
// and insert phi nodes where control from several predecessors rejoins.
//
// This is Braun et al., "Simple and Efficient Construction of Static Single Assignment Form"
// (CC 2013), in the sealed-block form upstream uses. React's own specification of the pass is at
// `compiler/packages/babel-plugin-react-compiler/docs/passes/02-enterSSA.md`, and the shape here
// follows it closely enough that the two can be read side by side.
//
// # Why Braun rather than Cytron
//
// The textbook construction computes the iterated dominance frontier, places empty phis at every
// block in it, then renames in a second walk over the dominator tree. Braun places a phi only when
// a lookup actually crosses a join with disagreeing predecessors, which means it never places one
// that redundancy elimination would immediately remove, and it needs no dominance frontier at all.
//
// Dominance is still what CORRECTNESS is stated against - see `VerifySSA` - but it is computed
// there, over this graph, for checking rather than for construction.
//
// # The one property everything rests on
//
// A block may be processed only after every predecessor that is not a back edge, so that a lookup
// into a predecessor finds a finished answer. The block slice is in reverse postorder, which is
// exactly that guarantee, and an IR's Finalize establishes it. A caller that restructured the graph
// without re-running it gets silent nonsense, which is why an IR's Construct re-runs it first.
//
// # Loops, which is where a construction that looks right on straight-line code breaks
//
// A loop header is reached from before the loop and from the back edge, and the back edge's block
// has not been processed when the header is. Braun's answer is the incomplete phi: when a lookup
// reaches a block with unprocessed predecessors, mint the phi's result immediately, record it as
// the block's definition so the loop body's reads bind to it, and leave the operands to be filled
// when the last predecessor lands. Recording the definition BEFORE recursing is also what stops the
// lookup recursing forever around the cycle.
//
// # What is deliberately not here
//
// Redundant-phi elimination is a separate pass, `EliminateRedundantPhis`, in eliminate.go, which
// Construct runs. Recursion into nested functions is each IR's own, around this.
package static_single_assignment

// Construct converts one function to single static assignment form, in place, then eliminates the
// redundant phis that placement produces. It does not re-establish the graph's invariants or recurse
// into nested functions; an IR's own Construct does both around it.
//
// After it returns: every identifier that a source binding takes is written exactly once, every use
// names the definition that actually reaches it, and each block's phis hold a phi wherever a
// binding's value depends on which predecessor control arrived from.
func Construct[G Graph[F, B, P], F any, B comparable, P any](graph G, function F) {
	// Phis from a previous run are dropped, because this pass APPENDS them and cannot reconcile
	// what it did not mint.
	//
	// A caller that restructures the graph and runs construction again is the case: inlining an
	// immediately invoked function expression does exactly that. The phis left behind name
	// identifiers from before the renumbering, so they define a value nothing produces, join no
	// class in the disjoint partition, and take no scope. Measured over the corpus, a second run
	// takes phis 1,548 to 3,096 -- every one duplicated -- and phis naming an undefined operand
	// 221 to 1,730.
	//
	// Same intent as recomputing reverse postorder first: re-establish the invariant rather than
	// trust it. Braun's algorithm derives every phi it needs from the graph, so nothing is lost by
	// discarding the previous answer, and `EliminateRedundantPhis` below then sees only phis this
	// run placed.
	var noBlock B
	for _, block := range graph.Blocks(function) {
		if block != noBlock {
			graph.SetPhis(block, nil)
		}
	}

	builder := &ssaBuilder[G, F, B, P]{
		graph:         graph,
		function:      function,
		states:        map[BlockId]*ssaState[P]{},
		unsealedPreds: map[BlockId]int{},
		unknown:       map[DeclarationId]bool{},
	}
	builder.run()

	// Elimination is part of construction rather than an optional follow-up, because Braun's
	// placement cannot avoid producing redundant phis and their share is not marginal. Measured over
	// 1,945 functions of real TypeScript: 24,418 phis before elimination, 2,746 after. Leaving them
	// in would mean 89% of the merge points a consumer sees are not merges, and every pass above
	// this would have to re-derive which ones are real.
	//
	// Both numbers moved when captures landed (from 21,654 and 2,770), in opposite directions, and
	// the split is worth knowing. Captures ADD raw phis, because a captured binding is now a real
	// value that gets merged at every join it crosses rather than an opaque name. They REMOVE
	// surviving ones, because a variable only reassigned from inside a closure now takes that write
	// as a `StoreContext` in the nested function instead of leaving the outer versions to be merged.
	EliminateRedundantPhis(graph, function)
}

// ssaState is one block's view: what each original binding currently resolves to, and the phis
// whose operands are still waiting on an unprocessed predecessor.
type ssaState[P any] struct {
	// defs maps a BINDING to the value it holds on entry to, or within, this block.
	//
	// The key is the DeclarationId rather than the IdentifierId, and that is the single most
	// important decision in this file. Lowering already mints a fresh IdentifierId for every store
	// to a variable, so `let y` reassigned twice arrives as three distinct identifiers sharing one
	// declaration. Keying on the identifier would make each store a different variable, and a
	// lookup would never find a predecessor's definition because the predecessor stored under a
	// different key. Keying on the declaration is what makes them one variable with several values,
	// which is the fact SSA exists to express.
	defs map[DeclarationId]IdentifierId

	// incompletePhis are phis minted before every predecessor was processed. Their result is already
	// bound in defs; only the operands are outstanding.
	incompletePhis []incompletePhi[P]
}

// incompletePhi is a phi awaiting operands.
type incompletePhi[P any] struct {
	// original is the pre-SSA identifier being merged, the key a lookup uses.
	original P
	// renamed is the phi's result, already minted and already visible in defs.
	renamed P
}

type ssaBuilder[G Graph[F, B, P], F any, B comparable, P any] struct {
	graph    G
	function F

	states map[BlockId]*ssaState[P]

	// unsealedPreds counts, per block, how many predecessors have not yet been processed. A block
	// whose count reaches zero is sealed and its incomplete phis can be filled.
	//
	// Absent from the map means "not yet decremented", which is not the same as zero; the read sites
	// initialise from the predecessor count on first touch for exactly that reason.
	unsealedPreds map[BlockId]int

	// unknown holds bindings a lookup walked off the entry block without finding. In cohere they are
	// globals, or captures from an enclosing function; in Adamic, which keeps both out of its graph, a
	// variable read before anything defines it. They are left un-renamed: there is no definition in
	// this function to merge, so a phi over them would be an invention.
	unknown map[DeclarationId]bool

	// visited records blocks already processed, so sealing only fills phis for a block whose body
	// has actually been walked.
	visited map[BlockId]bool

	// walking, walkingBlock and walkingContextStore are where run's walk is, read by visitPlace.
	walking             walking
	walkingBlock        BlockId
	walkingContextStore bool
}

func (b *ssaBuilder[G, F, B, P]) run() {
	b.visited = map[BlockId]bool{}

	// Parameters are definitions in the entry block. Renaming them is what makes a reassigned
	// parameter behave like any other binding rather than like a global.
	entryId := b.graph.Entry(b.function)
	if _, ok := b.graph.Block(b.function, entryId); !ok {
		return
	}
	b.states[entryId] = newSSAState[P]()
	params := b.graph.Params(b.function)
	for index := range params {
		b.defineIn(entryId, &params[index])
	}

	// One visitor for every place, made once with the builder, which reads what to do from
	// b.walking. A visitor handed to a Graph method escapes, since the compiler cannot see through a
	// type parameter's method, and so does every local it captures; one per instruction would be an
	// allocation per instruction.
	visit := func(place *P, role Role) { b.visitPlace(place, role) }
	edges := edgeScratchPool.Get().(*edgeScratch)
	defer edgeScratchPool.Put(edges)

	for _, block := range b.graph.Blocks(b.function) {
		blockId := b.graph.Id(block)
		b.walkingBlock = blockId
		if _, exists := b.states[blockId]; !exists {
			b.states[blockId] = newSSAState[P]()
		}
		b.visited[blockId] = true

		count := b.graph.InstructionCount(b.function, block)
		for index := 0; index < count; index++ {
			// Uses first, then definitions. `x = x + 1` must read the OLD x before the store mints
			// the new one; visiting in the other order would make the increment read itself.
			b.walking = walkingUses
			b.graph.EachInstructionPlace(b.function, block, index, visit)
			// Only a context store reuses its binding's definition. The guard keys on the
			// declaration, and a `FunctionExpression` assigned to the same declaration would
			// otherwise reuse the identifier too -- measured as a duplicate definition of a
			// function value on `let n = 1; const g = () => n; n = 2`.
			b.walking = walkingDefinitions
			b.walkingContextStore = b.graph.IsContextStore(b.function, block, index)
			b.graph.EachInstructionPlace(b.function, block, index, visit)
		}

		b.walking = walkingTerminal
		b.graph.EachTerminalPlace(block, visit)

		// Seal each successor that this block was the last unprocessed predecessor of.
		for _, edge := range edgesOf(b.graph, block, edges) {
			if edge.edge == Fallthrough {
				continue
			}
			successor, ok := b.graph.Block(b.function, edge.successor)
			if !ok {
				continue
			}
			remaining, seen := b.unsealedPreds[edge.successor]
			if !seen {
				remaining = len(b.graph.Predecessors(successor))
			}
			remaining--
			b.unsealedPreds[edge.successor] = remaining
			if remaining == 0 && b.visited[edge.successor] {
				b.fixIncompletePhis(edge.successor)
			}
		}
	}

	// The function's return value is a definition like any other and is read by nothing inside the
	// graph, so it is renamed to whatever reaches the end rather than left pointing at the original.
	b.renameReturns()
}

// walking is which of run's three visits a place is in.
type walking uint8

const (
	// walkingUses renames an instruction's uses, and passes over its definitions.
	walkingUses walking = iota
	// walkingDefinitions renames an instruction's definitions, and passes over its uses.
	walkingDefinitions
	// walkingTerminal renames a terminal's places, uses and definitions both.
	walkingTerminal
)

// visitPlace is run's one visitor, doing what b.walking says for the block b.walkingBlock.
func (b *ssaBuilder[G, F, B, P]) visitPlace(place *P, role Role) {
	switch b.walking {
	case walkingUses:
		if role != Define {
			b.useIn(b.walkingBlock, place)
		}
	case walkingDefinitions:
		if role == Define {
			b.defineInMaybeContext(b.walkingBlock, place, b.walkingContextStore)
		}
	case walkingTerminal:
		if role == Define {
			b.defineIn(b.walkingBlock, place)
			return
		}
		b.useIn(b.walkingBlock, place)
	}
}

func newSSAState[P any]() *ssaState[P] {
	return &ssaState[P]{defs: map[DeclarationId]IdentifierId{}}
}

// defineIn mints a fresh value for a definition and records it as the block's current answer.
//
// # A context binding is defined once and never redefined
//
// Upstream refuses every rename after the first for a binding it has marked as context:
//
//	// Do not redefine context references.
//	if (this.#context.has(oldId)) {
//	  return this.getPlace(oldPlace);
//	}
//
// `SSA/EnterSSA.ts:124`. That is what makes a later write reach the value a closure already
// captured. For `let x = []; useCallback(() => [x], [x]); x = makeArray();` the reassignment's
// `Mutate` lands on the identifier the callback captured, so the range widens across the memo
// marker and the rule can see that the dependency may be mutated later.
//
// Versioning that write instead mints a value at the reassignment's own order, and a mutation on a
// value whose range starts there widens nothing -- measured as `{20, 21}` against a memo block
// spanning 6 to 15, which is why the effect and the lowering were both faithful and the finding
// still did not fire.
func (b *ssaBuilder[G, F, B, P]) defineIn(blockId BlockId, place *P) {
	b.defineInMaybeContext(blockId, place, false)
}

// defineInMaybeContext is `defineIn`, told whether this definition is a context store.
func (b *ssaBuilder[G, F, B, P]) defineInMaybeContext(blockId BlockId, place *P, contextStore bool) {
	binding := b.graph.Declaration(b.function, b.graph.IdentifierOf(*place))
	// # A context binding is defined once and every later write reuses that definition
	//
	// Upstream's two writes to a reassigned-and-captured binding name the SAME identifier --
	// measured on the pinned build, `StoreContext@2 lvalueId=2 kind=Let` and
	// `StoreContext@17 lvalueId=2 kind=Reassign`. Lowering resolves both through one Babel binding
	// and nothing renumbers it afterwards.
	//
	// That shared identity is what puts the declaration and the reassignment in one disjoint class,
	// so the scope's hull spans both and covers the memo block between them. Versioning them
	// instead gives the reassignment a range starting at its own order, which widens nothing --
	// measured as `{20, 21}` against a memo block spanning 6 to 15.
	//
	// Gated on the instruction being a context store. The set is keyed by declaration, and a
	// `FunctionExpression` assigned to the same declaration would otherwise reuse the identifier
	// too, which is a real duplicate definition -- measured on
	// `let n = 1; const g = () => n; n = 2`.
	if contextStore && b.graph.Contextual(b.function, binding) {
		if existing, defined := b.states[blockId].defs[binding]; defined {
			*place = b.graph.WithIdentifier(*place, existing)
			return
		}
	}
	renamed := b.graph.Mint(b.function, b.graph.IdentifierOf(*place))
	b.states[blockId].defs[binding] = renamed
	*place = b.graph.WithIdentifier(*place, renamed)
}

// useIn rewrites a use to whatever value reaches this block.
func (b *ssaBuilder[G, F, B, P]) useIn(blockId BlockId, place *P) {
	*place = b.graph.WithIdentifier(*place, b.valueAt(*place, blockId))
}

// valueAt is Braun's lookup: which value does this binding hold on entry to this block.
//
// The order of the cases is load-bearing and is Braun's:
//
//  1. Defined here already - the local answer wins.
//  2. No predecessors - the entry block, and the binding was never defined. It is a global or a
//     capture; record it and hand back the original rather than inventing a definition.
//  3. Some predecessor unprocessed - a loop. Mint an incomplete phi. Recording it in defs BEFORE
//     recursing is what terminates the walk around the cycle.
//  4. Exactly one predecessor - no merge, so recurse and cache. A phi here would be redundant by
//     construction.
//  5. Several predecessors - a real join. Mint the result, record it, THEN collect operands; a
//     predecessor whose lookup comes back around to this block must find the result already bound.
func (b *ssaBuilder[G, F, B, P]) valueAt(place P, blockId BlockId) IdentifierId {
	original := b.graph.IdentifierOf(place)
	binding := b.graph.Declaration(b.function, original)
	if b.unknown[binding] {
		return original
	}

	state, ok := b.states[blockId]
	if !ok {
		state = newSSAState[P]()
		b.states[blockId] = state
	}
	if renamed, defined := state.defs[binding]; defined {
		return renamed
	}

	block, ok := b.graph.Block(b.function, blockId)
	if !ok {
		return original
	}

	predecessors := b.graph.Predecessors(block)
	if len(predecessors) == 0 {
		b.unknown[binding] = true
		return original
	}

	remaining, seen := b.unsealedPreds[blockId]
	if !seen {
		remaining = len(predecessors)
	}
	if remaining > 0 {
		renamed := b.graph.Mint(b.function, original)
		state.defs[binding] = renamed
		state.incompletePhis = append(state.incompletePhis, incompletePhi[P]{
			original: place,
			renamed:  b.graph.WithIdentifier(place, renamed),
		})
		return renamed
	}

	if len(predecessors) == 1 {
		renamed := b.valueFromPredecessor(place, predecessors[0], blockId)
		state.defs[binding] = renamed
		return renamed
	}

	renamed := b.graph.Mint(b.function, original)
	state.defs[binding] = renamed
	b.addPhi(blockId, place, b.graph.WithIdentifier(place, renamed))
	return renamed
}

// valueFromPredecessor is the value a binding holds when control arrives at blockId from predecessor.
//
// Every lookup across an edge comes through here, and today it is the predecessor's value at its end
// whatever the edge, which is right for every edge but an exceptional one. An exceptional edge leaves
// its block from the throwing instruction, before that instruction's definitions happen, so the
// handler should see the definitions from just before it. That is #2yz9ra9, live in Adamic, whose
// MayThrow emits such edges, and latent in cohere, which emits none. Its fix lands here: snapshot the
// block's definitions before its throwing instruction while walking it, and answer an Exceptional
// edge from the snapshot.
func (b *ssaBuilder[G, F, B, P]) valueFromPredecessor(place P, predecessor BlockId, blockId BlockId) IdentifierId {
	return b.valueAt(place, predecessor)
}

// addPhi collects one operand per predecessor and attaches the phi to the block.
//
// original is the place a lookup reads, and each operand is that place naming the predecessor's
// value, so whatever else the place carries rides along on every operand.
func (b *ssaBuilder[G, F, B, P]) addPhi(blockId BlockId, original P, renamed P) {
	block, ok := b.graph.Block(b.function, blockId)
	if !ok {
		return
	}

	predecessors := b.graph.Predecessors(block)
	operands := make(PhiOperands[P], 0, len(predecessors))
	for _, predecessorId := range predecessors {
		operands.Set(predecessorId, b.graph.WithIdentifier(original,
			b.valueFromPredecessor(original, predecessorId, blockId)))
	}

	b.graph.SetPhis(block, append(b.graph.Phis(block), &Phi[P]{Place: renamed, Operands: operands}))
}

// fixIncompletePhis fills in the operands of every phi minted before this block was sealed.
func (b *ssaBuilder[G, F, B, P]) fixIncompletePhis(blockId BlockId) {
	state, ok := b.states[blockId]
	if !ok {
		return
	}
	pending := state.incompletePhis
	state.incompletePhis = nil
	for _, phi := range pending {
		b.addPhi(blockId, phi.original, phi.renamed)
	}
}

// renameReturns rewrites the function's Returns place to the value that reaches the exit.
//
// Every `return` stores into one identifier, so after renaming the stores there are several values
// and the Returns place still names the original. It is resolved against the blocks that actually
// end in a Return, which is where the value is observable. An IR with no Returns place skips this.
func (b *ssaBuilder[G, F, B, P]) renameReturns() {
	returns := b.graph.Returns(b.function)
	if returns == nil {
		return
	}
	binding := b.graph.Declaration(b.function, b.graph.IdentifierOf(*returns))
	if b.unknown[binding] {
		return
	}
	for _, block := range b.graph.Blocks(b.function) {
		if !b.graph.EndsInReturn(block) {
			continue
		}
		if state, ok := b.states[b.graph.Id(block)]; ok {
			if renamed, defined := state.defs[binding]; defined {
				*returns = b.graph.WithIdentifier(*returns, renamed)
				return
			}
		}
	}
}
