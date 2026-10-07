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
// That last point is why this does not consume `internal/lint/ecmascript/control_flow_graph`'s dominator tree.
// `AnalyzeDominators` is generic over `control_flow_graph.Graph[E]` and `control_flow_graph.Block[E]`, which are
// that package's own types; this graph is `Function` and `BasicBlock`. There is no conversion, and
// writing one would mean materialising a second graph purely to compute a frontier this algorithm
// does not use. Dominance is still what CORRECTNESS is stated against - see `VerifySSA` - but it is
// computed there, over this graph, for checking rather than for construction.
//
// Stated plainly for the next pass built on this IR, because it is easy to plan around the wrong
// answer: NOTHING in `internal/lint/ecmascript/control_flow_graph` applies to this graph without a conversion that
// does not exist. Its dominators, its dataflow solver, and its path analysis are all generic over
// `Graph[E]`. A pass here that wants any of them either writes the conversion or, as this does,
// computes what it needs over `Function.Blocks`, which is already in reverse postorder.
//
// # The doubled finally, and why it does not reach here
//
// `controlflow` lays a `finally` body out TWICE, and three consumers have been bitten by it. For
// this pass it would be a correctness bug rather than a nuisance: two layouts of one body give a
// binding two definitions that are not a merge, and construction over that mints a phi where the
// source has no join.
//
// It does not reach here, and that was verified rather than assumed. Lowering emits the finally
// body ONCE and jumps to it from every path that must run it; see `lowerTryStatement`. Checked
// empirically on `try/finally` and `try/catch/finally`, both of which produce a single finally
// block whose predecessors are the normal and abrupt paths. That is a real join, and a phi placed
// there is correct.
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
// Reclassifying `const`/`let` after renaming - upstream's
// `rewrite_instruction_kinds_based_on_reassignment` - is NOT implemented, because nothing in cohere
// reads `InstructionKind` yet and a reclassification no pass consumes is untested by construction.
package high_level_intermediate_representation

import "sort"

// Construct converts a function to single static assignment form, in place, and recursively for
// every nested function.
//
// After it returns: every identifier that a source binding takes is written exactly once, every use
// names the definition that actually reaches it, and `BasicBlock.Phis` holds a phi wherever a
// binding's value depends on which predecessor control arrived from.
//
// # What this cannot do for you, and how to tell
//
// A variable reference lowering left as `LoadGlobal` carries a name rather than a value, so there
// is nothing here to rename and no phi to place for it. That is correct for a true global and wrong
// for a local, and lowering produces the latter for EVERY reference when it was given no type
// checker. `Construct` cannot distinguish the two cases. `Lower`'s comment carries the measurement;
// the short version is that SSA over an unresolved function is well-formed, empty of phis, and
// meaningless.
//
// # Captures, and what renaming does and does not do at a function boundary
//
// A variable a nested function closes over IS tracked: lowering populates `FunctionExpression.
// Captures` and `Function.Context`, and the nested body reads it as `LoadContext` rather than
// `LoadGlobal`. `Construct` recurses into nested functions, so both sides are in single-assignment
// form.
//
// What renaming does NOT do is unify the two sides, and a pass must not assume it did. The capture
// and the captured value are separate identifiers in separate tables, because a `Place` is only
// meaningful against the table of the function holding it. The edge between them is positional:
// `FunctionExpression.Captures[i]` is the enclosing function's value and `nested.Context[i]` is the
// nested function's, naming one source binding from the two sides. A pass following a value into a
// closure walks that pairing.
//
// The capture is bound at the point the closure is CREATED, which is what makes it a value rather
// than a name. For `let n = 1; if (c) { n = 2; } const g = () => n;` the capture names the phi that
// merges the two versions, so `g` closes over what actually reaches it. A later write through
// `StoreContext` is a definition inside the nested function and is versioned there; it is NOT
// reflected back into the enclosing function's versions, so a pass asking what the OUTER `n` holds
// after calling `g` must still treat that as unknown.
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
		contextual:    function.ContextDeclarations,
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
	EliminateRedundantPhis(function)

	for _, nested := range function.Functions {
		Construct(nested)
	}
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

	// unknown holds bindings a lookup walked off the entry block without finding. They are globals,
	// or captures from an enclosing function, and they are left un-renamed: there is no definition
	// in this function to merge, so a phi over them would be an invention.
	unknown map[DeclarationId]bool

	// contextual holds bindings a nested function captures, which are defined once and never
	// redefined. See `defineIn` for why, and for what versioning them instead costs.
	contextual map[DeclarationId]bool

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
			// Only a context store reuses its binding's definition. The guard keys on the
			// declaration, and a `FunctionExpression` assigned to the same declaration would
			// otherwise reuse the identifier too -- measured as a duplicate definition of a
			// function value on `let n = 1; const g = () => n; n = 2`.
			_, isContextStore := instruction.Value.(*StoreContext)
			EachInstructionPlacePointer(instruction, func(place *Place, role PlaceRole) {
				if role == PlaceRoleDefine {
					b.defineInMaybeContext(block.Id, place, isContextStore)
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

	// The function's return value is a definition like any other and is read by nothing inside the
	// graph, so it is renamed to whatever reaches the end rather than left pointing at the original.
	b.renameReturns()
}

func newSSAState() *ssaState {
	return &ssaState{defs: map[DeclarationId]IdentifierId{}}
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
func (b *ssaBuilder) defineIn(blockId BlockId, place *Place) {
	b.defineInMaybeContext(blockId, place, false)
}

// defineInMaybeContext is `defineIn`, told whether this definition is a context store.
func (b *ssaBuilder) defineInMaybeContext(blockId BlockId, place *Place, contextStore bool) {
	binding := b.function.Identifiers[place.Identifier].Declaration
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
	if contextStore && b.contextual[binding] {
		if existing, defined := b.states[blockId].defs[binding]; defined {
			place.Identifier = existing
			return
		}
	}
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
	created := b.function.NewIdentifier(source.Name, source.Node, source.Declaration)
	created.Type = source.Type
	created.Scope = source.Scope
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
			renamed:  Place{Identifier: renamed, Effect: place.Effect, Reactive: place.Reactive, Range: place.Range},
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
		Effect:     place.Effect,
		Reactive:   place.Reactive,
		Range:      place.Range,
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
			Effect:     original.Effect,
			Reactive:   original.Reactive,
			Range:      original.Range,
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

// renameReturns rewrites the function's Returns place to the value that reaches the exit.
//
// Every `return` stores into one identifier, so after renaming the stores there are several values
// and `Function.Returns` still names the original. It is resolved against the blocks that actually
// end in a Return, which is where the value is observable.
func (b *ssaBuilder) renameReturns() {
	binding := b.function.Identifiers[b.function.Returns.Identifier].Declaration
	if b.unknown[binding] {
		return
	}
	for _, block := range b.function.Blocks {
		if _, ends := block.Terminal.(*Return); !ends {
			continue
		}
		if state, ok := b.states[block.Id]; ok {
			if renamed, defined := state.defs[binding]; defined {
				b.function.Returns.Identifier = renamed
				return
			}
		}
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
