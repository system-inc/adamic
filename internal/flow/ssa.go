// Single static assignment form: rename every definition so each value is written exactly once,
// and insert phi nodes where control from several predecessors rejoins.
//
// Lifted from cohere's ssa.go (high_level_intermediate_representation at 715ba94) for #5jck546. The
// algorithm is cohere's, line for line. What changed is what Adamic's graph doesn't have: context
// stores (a captured variable isn't a value here at all; see flow.go), nested functions inside one
// graph (each Adamic closure is its own), a Returns place (a return's value is read by an
// instruction), and a place's effect, reactivity and source range. Each change is marked "Adamic:".
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
// (cohere's comment here also covers its other control-flow graph and the doubled `finally`. Adamic
// has neither: 0.1 has no `try`.)
//
// # The one property everything rests on
//
// A block may be processed only after every predecessor that is not a back edge, so that a lookup
// into a predecessor finds a finished answer. `Function.Blocks` is in reverse postorder, which is
// exactly that guarantee, and `Finalize` establishes it. A caller that restructured the graph
// without re-running `Finalize` gets silent nonsense, which is why `Construct` re-runs it.
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
// Redundant-phi elimination is a separate pass, `EliminateRedundantPhis`, in ssa_eliminate.go.
package flow

import "sort"

// Construct converts a function to single static assignment form, in place, and recursively for
// every nested function.
//
// After it returns: every identifier that a source binding takes is written exactly once, every use
// names the definition that actually reaches it, and `BasicBlock.Phis` holds a phi wherever a
// binding's value depends on which predecessor control arrived from.
//
// Adamic: cohere's comment here is about variables lowering left unresolved and about captures.
// Neither arises: Build resolves every variable through the IR's own local numbering, and leaves
// globals and captured variables out of the graph (flow.go says why).
func Construct(function *Function) {
	if function == nil {
		return
	}

	// Re-establish the invariants rather than trusting them. Reverse postorder is the property the
	// whole algorithm rests on and it is cheap to recompute; a caller who restructured the graph and
	// forgot would otherwise get a wrong answer with no symptom.
	Finalize(function)

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
	// Same intent as recomputing reverse postorder above: re-establish the invariant rather than
	// trust it. Braun's algorithm derives every phi it needs from the graph, so nothing is lost by
	// discarding the previous answer, and `EliminateRedundantPhis` below then sees only phis this
	// run placed.
	for _, block := range function.Blocks {
		if block != nil {
			block.Phis = nil
		}
	}

	builder := &ssaBuilder{
		function:      function,
		states:        map[BlockId]*ssaState{},
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
	EliminateRedundantPhis(function)
}

// ssaState is one block's view: what each original binding currently resolves to, and the phis
// whose operands are still waiting on an unprocessed predecessor.
type ssaState struct {
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
	incompletePhis []incompletePhi
}

// incompletePhi is a phi awaiting operands.
type incompletePhi struct {
	// original is the pre-SSA identifier being merged, the key a lookup uses.
	original Place
	// renamed is the phi's result, already minted and already visible in defs.
	renamed Place
}

type ssaBuilder struct {
	function *Function

	states map[BlockId]*ssaState

	// unsealedPreds counts, per block, how many predecessors have not yet been processed. A block
	// whose count reaches zero is sealed and its incomplete phis can be filled.
	//
	// Absent from the map means "not yet decremented", which is not the same as zero; the read sites
	// initialise from len(Predecessors) on first touch for exactly that reason.
	unsealedPreds map[BlockId]int

	// unknown holds bindings a lookup walked off the entry block without finding, and they are left
	// un-renamed: there is no definition in this function to merge, so a phi over them would be an
	// invention. (Adamic: in cohere these are globals and captures. Here, where Build leaves both
	// out, it's a variable read before anything defines it, which the checker's definite assignment
	// rules out; so a nonempty set is worth a look.)
	unknown map[DeclarationId]bool

	// visited records blocks already processed, so sealing only fills phis for a block whose body
	// has actually been walked.
	visited map[BlockId]bool
}

func (b *ssaBuilder) run() {
	b.visited = map[BlockId]bool{}

	// Parameters are definitions in the entry block. Renaming them is what makes a reassigned
	// parameter behave like any other binding rather than like a global.
	entry, ok := b.function.Block(b.function.Entry)
	if !ok {
		return
	}
	b.states[entry.Id] = newSSAState()
	for index := range b.function.Params {
		b.defineIn(entry.Id, &b.function.Params[index])
	}

	for _, block := range b.function.Blocks {
		if _, exists := b.states[block.Id]; !exists {
			b.states[block.Id] = newSSAState()
		}
		b.visited[block.Id] = true

		for _, instructionId := range block.Instructions {
			instruction := b.function.Instructions[instructionId]
			// Uses first, then definitions. `x = x + 1` must read the OLD x before the store mints
			// the new one; visiting in the other order would make the increment read itself.
			EachInstructionPlacePointer(instruction, func(place *Place, role PlaceRole) {
				if role != PlaceRoleDefine {
					b.useIn(block.Id, place)
				}
			})
			// Adamic: no context store, so every definition is a fresh value.
			EachInstructionPlacePointer(instruction, func(place *Place, role PlaceRole) {
				if role == PlaceRoleDefine {
					b.defineIn(block.Id, place)
				}
			})
		}

		EachTerminalPlacePointer(block.Terminal, func(place *Place, role PlaceRole) {
			if role == PlaceRoleDefine {
				b.defineIn(block.Id, place)
				return
			}
			b.useIn(block.Id, place)
		})

		// Seal each successor that this block was the last unprocessed predecessor of.
		EachSuccessor(block.Terminal, func(successorId BlockId) {
			successor, ok := b.function.Block(successorId)
			if !ok {
				return
			}
			remaining, seen := b.unsealedPreds[successorId]
			if !seen {
				remaining = len(successor.Predecessors)
			}
			remaining--
			b.unsealedPreds[successorId] = remaining
			if remaining == 0 && b.visited[successorId] {
				b.fixIncompletePhis(successorId)
			}
		})
	}
}

func newSSAState() *ssaState {
	return &ssaState{defs: map[DeclarationId]IdentifierId{}}
}

// defineIn mints a fresh value for a definition and records it as the block's current answer.
//
// Adamic: cohere's defineIn reuses a context binding's first definition for every later write
// through a closure (upstream's `SSA/EnterSSA.ts:124`). No captured variable is in this graph, so
// every definition here mints.
func (b *ssaBuilder) defineIn(blockId BlockId, place *Place) {
	binding := b.function.Identifiers[place.Identifier].Declaration
	renamed := b.mint(place.Identifier)
	b.states[blockId].defs[binding] = renamed
	place.Identifier = renamed
}

// useIn rewrites a use to whatever value reaches this block.
func (b *ssaBuilder) useIn(blockId BlockId, place *Place) {
	place.Identifier = b.valueAt(place, blockId)
}

// mint creates a new value carrying the same source binding as the original.
//
// The DeclarationId is preserved, which is the whole point: after renaming, "which variable is
// this" is answered by the declaration and "which value is this" by the identifier, and a pass can
// ask either.
func (b *ssaBuilder) mint(original IdentifierId) IdentifierId {
	source := b.function.Identifiers[original]
	created := b.function.NewIdentifier(source.Name, source.Declaration)
	return created.Id
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
func (b *ssaBuilder) valueAt(place *Place, blockId BlockId) IdentifierId {
	original := place.Identifier
	binding := b.function.Identifiers[original].Declaration
	if b.unknown[binding] {
		return original
	}

	state, ok := b.states[blockId]
	if !ok {
		state = newSSAState()
		b.states[blockId] = state
	}
	if renamed, defined := state.defs[binding]; defined {
		return renamed
	}

	block, ok := b.function.Block(blockId)
	if !ok {
		return original
	}

	if len(block.Predecessors) == 0 {
		b.unknown[binding] = true
		return original
	}

	remaining, seen := b.unsealedPreds[blockId]
	if !seen {
		remaining = len(block.Predecessors)
	}
	if remaining > 0 {
		renamed := b.mint(original)
		state.defs[binding] = renamed
		state.incompletePhis = append(state.incompletePhis, incompletePhi{
			original: *place,
			renamed:  Place{Identifier: renamed},
		})
		return renamed
	}

	if len(block.Predecessors) == 1 {
		renamed := b.valueAt(place, block.Predecessors[0])
		state.defs[binding] = renamed
		return renamed
	}

	renamed := b.mint(original)
	state.defs[binding] = renamed
	b.addPhi(blockId, *place, Place{
		Identifier: renamed,
	})
	return renamed
}

// addPhi collects one operand per predecessor and attaches the phi to the block.
func (b *ssaBuilder) addPhi(blockId BlockId, original Place, renamed Place) {
	block, ok := b.function.Block(blockId)
	if !ok {
		return
	}

	operands := make(map[BlockId]Place, len(block.Predecessors))
	for _, predecessorId := range block.Predecessors {
		lookup := original
		operands[predecessorId] = Place{
			Identifier: b.valueAt(&lookup, predecessorId),
		}
	}

	block.Phis = append(block.Phis, &Phi{Place: renamed, Operands: operands})
}

// fixIncompletePhis fills in the operands of every phi minted before this block was sealed.
func (b *ssaBuilder) fixIncompletePhis(blockId BlockId) {
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

// PhiOperandsInOrder returns a phi's operands keyed by predecessor, in ascending block order.
//
// `Phi.Operands` is a map, and Go randomises map iteration deliberately. Any pass that prints,
// hashes, or compares phis must read them through here, or the same input produces different output
// between runs. `TestLowerIsDeterministic` in lower_corpus_test.go is what catches a caller that
// forgets.
func PhiOperandsInOrder(phi *Phi) []BlockId {
	blocks := make([]BlockId, 0, len(phi.Operands))
	for blockId := range phi.Operands {
		blocks = append(blocks, blockId)
	}
	sort.Slice(blocks, func(i, j int) bool { return blocks[i] < blocks[j] })
	return blocks
}
