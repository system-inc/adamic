// Recovering optional chains from the control flow they lower to.
//
// This is React's `collectOptionalChainSidemap` (`HIR/CollectOptionalChainDependencies.ts`), which
// answers a question the instruction stream cannot: `a?.b.c` is ONE dependency path with one
// optional step, but by the time it reaches the dependency collector it is a tree of blocks -- an
// `Optional` terminal per link, a `Branch` per test, a consequent holding the load, and one shared
// alternate storing undefined.
//
// # Why the shape is the algorithm
//
// Upstream does not track optionality on the instruction. `PropagateScopeDependenciesHIR` passes a
// literal `false` at both of its collection sites (`:301` and `:696`), so a dependency it collects
// from a `PropertyLoad` is never optional. Every optional path in its output comes from here, by
// pattern-matching the block structure the lowering produced.
//
// That makes this pass load-bearing rather than an optimisation: without it, an optional access
// either goes uncollected or is collected without its marker, and
// `CompareManualMemoDependencies` compares the marker strictly against what the developer wrote.
package high_level_intermediate_representation

// OptionalChainSidemap is what one traversal recovered, upstream's three outputs.
type OptionalChainSidemap struct {
	// TemporariesReadInOptional maps a temporary written inside an optional chain to the full
	// dependency path it represents, with optionality recorded per step.
	TemporariesReadInOptional map[IdentifierId]ReactiveScopeDependency

	// ProcessedInstructions are the StoreLocal instructions this pass consumed. Property-load
	// lvalues are deferred through TemporariesReadInOptional, while test terminals are recorded in
	// ProcessedOptionalTests because terminals have no InstructionId in this IR.
	ProcessedInstructions map[InstructionId]bool

	// ProcessedOptionalTests are the blocks whose `Branch` terminal this pass recovered as an
	// optional chain's test. Upstream records the terminal itself (`processedInstrsInOptional` is
	// typed `Instruction | Terminal`) and skips its operand walk; keying by block is the same claim
	// in an IR where terminals have no stable id.
	ProcessedOptionalTests map[BlockId]bool

	// HoistableObjects records, per optional block, a base whose further property loads are safe to
	// read unconditionally. Set only for a non-optional link riding an outer chain, which is the
	// `.c` in `a?.b.c`.
	HoistableObjects map[BlockId]ReactiveScopeDependency
}

// optionalTraversal is the mutable state of one walk. Upstream's `OptionalTraversalContext`.
type optionalTraversal struct {
	function *Function
	seen     map[BlockId]bool
	result   *OptionalChainSidemap
}

// CollectOptionalChainSidemap recovers every optional chain in a function.
//
// Returns empty maps when the function holds no `Optional` terminal, which includes functions with
// no lowered dot-property optional access. Every consumer tolerates that; optional calls and
// computed links remain on the residual path declared by DependencyGapOptionalChains.
func CollectOptionalChainSidemap(function *Function) *OptionalChainSidemap {
	result := &OptionalChainSidemap{
		TemporariesReadInOptional: map[IdentifierId]ReactiveScopeDependency{},
		ProcessedInstructions:     map[InstructionId]bool{},
		ProcessedOptionalTests:    map[BlockId]bool{},
		HoistableObjects:          map[BlockId]ReactiveScopeDependency{},
	}
	if function == nil {
		return result
	}
	traversal := &optionalTraversal{
		function: function,
		seen:     map[BlockId]bool{},
		result:   result,
	}
	traversal.traverseFunction()
	return result
}

// recordOptionalChainJoinPhis maps a chain's join phi to the path its resolved operand carries.
//
// This is deliberately not part of CollectOptionalChainSidemap. Upstream's optional-chain sidemap
// records the property and store values, then dependency collection observes matching phi OPERANDS;
// it does not resolve the phi output back to the chain. Keeping the output local is load-bearing for
// values such as `propB?.x.y`: the resulting phi is declared inside the memo block and is therefore
// not a dependency the developer could have written.
//
// The Go pipeline does need this extra projection while reading the developer's dependency ARRAY.
// DropManualMemoization runs after SSA construction here, where an array element names the join phi;
// upstream runs that pass before SSA and reads the StoreLocal directly. The caller therefore applies
// this helper only to its source-dependency sidemap and never to inferred scope dependencies.
//
// Projection is restricted to actual Optional fallthrough blocks. One resolved operand is enough
// there: the other arm is the short circuit, which contributes no access. Applying the same rule to
// an enclosing ternary phi would incorrectly turn `cond ? value?.x : compute()` into `value?.x`.
func recordOptionalChainJoinPhis(function *Function, result *OptionalChainSidemap) {
	if function == nil || result == nil {
		return
	}
	joins := optionalChainJoinBlocks(function)
	for _, block := range function.Blocks {
		if block == nil || !joins[block.Id] {
			continue
		}
		for _, phi := range block.Phis {
			if _, known := result.TemporariesReadInOptional[phi.Place.Identifier]; known {
				continue
			}
			// Every operand that resolves must resolve to the SAME access. A join whose arms carry
			// different paths is not one chain's short circuit -- it is a real branch, and naming it
			// after either arm claims a dependency the other arm never reads.
			var resolved *ReactiveScopeDependency
			conflict := false
			for _, operand := range phi.Operands {
				dependency, ok := result.TemporariesReadInOptional[operand.Identifier]
				if !ok {
					continue
				}
				if resolved == nil {
					copied := dependency
					resolved = &copied
					continue
				}
				if resolved.Identifier != dependency.Identifier ||
					!equalPaths(resolved.Path, dependency.Path) {
					conflict = true
					break
				}
			}
			if resolved != nil && !conflict {
				result.TemporariesReadInOptional[phi.Place.Identifier] = *resolved
			}
		}
	}
}

// optionalChainJoinBlocks returns the fallthroughs at which optional short-circuit arms rejoin.
// Phi projection is valid only at these blocks; a later control-flow join may have one optional
// operand without itself representing that optional chain.
func optionalChainJoinBlocks(function *Function) map[BlockId]bool {
	result := map[BlockId]bool{}
	if function == nil {
		return result
	}
	for _, block := range function.Blocks {
		if block == nil {
			continue
		}
		if optional, ok := block.Terminal.(*Optional); ok {
			result[optional.Fallthrough] = true
		}
	}
	return result
}

func (t *optionalTraversal) traverseFunction() {
	for _, block := range t.function.Blocks {
		if block == nil {
			continue
		}
		if _, isOptional := block.Terminal.(*Optional); isOptional && !t.seen[block.Id] {
			t.traverseOptionalBlock(block, nil)
		}
	}
}

// optionalTestMatch is what `matchOptionalTestBlock` recovers from a branch's consequent.
type optionalTestMatch struct {
	consequentId   IdentifierId
	property       string
	propertyId     IdentifierId
	storeLocalId   InstructionId
	consequentGoto BlockId
}

// matchOptionalTestBlock matches the consequent and alternate of an optional's test.
//
// Upstream's function of the same name (`:163`). The shape it requires is exactly what the lowering
// emits: a consequent of exactly `PropertyLoad` then `StoreLocal` ending in `Goto{Break}`, against
// an alternate of exactly `Primitive` then `StoreLocal`.
//
// Upstream raises invariants where this returns nil. The difference is deliberate and is this
// tree's standing choice: a shape this pass cannot read is a chain it declines to collect, which
// leaves the access to be collected un-chained rather than failing the compilation.
func (t *optionalTraversal) matchOptionalTestBlock(branch *Branch) *optionalTestMatch {
	consequent, found := t.function.Block(branch.Consequent)
	if !found || consequent == nil || len(consequent.Instructions) != 2 {
		return nil
	}
	first := t.function.Instructions[consequent.Instructions[0]]
	second := t.function.Instructions[consequent.Instructions[1]]
	if first == nil || second == nil {
		return nil
	}
	load, isLoad := first.Value.(*PropertyLoad)
	store, isStore := second.Value.(*StoreLocal)
	if !isLoad || !isStore {
		return nil
	}
	// The load must be OF the value the branch tested, and the store must be OF the load. Upstream
	// asserts both; declining is the same claim without the crash.
	if load.Object.Identifier != branch.Test.Identifier {
		return nil
	}
	if store.Value.Identifier != first.LValue.Identifier {
		return nil
	}
	goToNext, isGoto := consequent.Terminal.(*Goto)
	if !isGoto || goToNext.Variant != GotoVariantBreak {
		return nil
	}
	alternate, foundAlternate := t.function.Block(branch.Alternate)
	if !foundAlternate || alternate == nil || len(alternate.Instructions) != 2 {
		return nil
	}
	alternateFirst := t.function.Instructions[alternate.Instructions[0]]
	alternateSecond := t.function.Instructions[alternate.Instructions[1]]
	if alternateFirst == nil || alternateSecond == nil {
		return nil
	}
	if _, isPrimitive := alternateFirst.Value.(*Primitive); !isPrimitive {
		return nil
	}
	if _, isStoreLocal := alternateSecond.Value.(*StoreLocal); !isStoreLocal {
		return nil
	}
	return &optionalTestMatch{
		consequentId:   store.LValue.Identifier,
		property:       load.Property,
		propertyId:     first.LValue.Identifier,
		storeLocalId:   second.Id,
		consequentGoto: goToNext.Block,
	}
}

// traverseOptionalBlock walks one optional link and everything it transitively references.
//
// Returns the identifier the chain's value was stored into, or zero when any part of it is not a
// plain property-load chain. Upstream's `traverseOptionalBlock` (`:240`).
func (t *optionalTraversal) traverseOptionalBlock(block *BasicBlock,
	outerAlternate *BlockId) IdentifierId {
	t.seen[block.Id] = true
	terminal, isOptional := block.Terminal.(*Optional)
	if !isOptional {
		return 0
	}
	maybeTest, found := t.function.Block(terminal.Test)
	if !found || maybeTest == nil {
		return 0
	}

	var baseObject ReactiveScopeDependency
	var test *Branch
	var testBlockId BlockId

	switch testTerminal := maybeTest.Terminal.(type) {
	case *Branch:
		// The base case. Upstream only matches a base expression that is a straightforward
		// `LoadLocal` followed by a chain of property loads, and its reason is worth keeping:
		// "Optional base expressions are currently within value blocks which cannot be interrupted
		// by scope boundaries. As such, the only dependencies we can hoist out of optional chains
		// are property load chains with no intervening instructions."
		if len(maybeTest.Instructions) == 0 {
			return 0
		}
		firstInstruction := t.function.Instructions[maybeTest.Instructions[0]]
		if firstInstruction == nil {
			return 0
		}
		// `LoadContext` as well as `LoadLocal`, which is a DIVERGENCE and is forced.
		//
		// Upstream's base case accepts only `LoadLocal` because its dependency traversal runs after
		// memo callbacks are inlined and captures have become ordinary locals. This pass also runs
		// earlier from DropManualMemoization to recover the developer's dependency array, when a
		// nested callback still reads captures through LoadContext.
		//
		// Measured on `error.hoist-optional-member-expression-with-conditional.js`: the base test
		// block holds exactly one instruction and it is a `LoadContext`, so accepting only
		// `LoadLocal` declined every chain in the fixtures this pass exists for.
		var base Place
		switch load := firstInstruction.Value.(type) {
		case *LoadLocal:
			base = load.Place
		case *LoadContext:
			base = load.Place
		default:
			return 0
		}
		path := make([]DependencyPathEntry, 0, len(maybeTest.Instructions)-1)
		for index := 1; index < len(maybeTest.Instructions); index++ {
			instruction := t.function.Instructions[maybeTest.Instructions[index]]
			previous := t.function.Instructions[maybeTest.Instructions[index-1]]
			if instruction == nil || previous == nil {
				return 0
			}
			load, isPropertyLoad := instruction.Value.(*PropertyLoad)
			if !isPropertyLoad || load.Object.Identifier != previous.LValue.Identifier {
				return 0
			}
			path = append(path, DependencyPathEntry{Property: load.Property})
		}
		baseObject = ReactiveScopeDependency{
			Identifier: base.Identifier,
			Reactive:   base.Reactive,
			Path:       path,
		}
		test = testTerminal
		testBlockId = maybeTest.Id

	case *Optional:
		// A nested link. Its fallthrough is where the outer test lands, and upstream requires that
		// block to terminate in a branch: "Fallthrough of the inner optional should be a block with
		// no instructions, terminating with Test($<temporary written to from StoreLocal>)".
		testBlock, foundTest := t.function.Block(testTerminal.Fallthrough)
		if !foundTest || testBlock == nil {
			return 0
		}
		branch, isBranch := testBlock.Terminal.(*Branch)
		if !isBranch {
			return 0
		}
		inner := t.traverseOptionalBlock(maybeTest, &branch.Alternate)
		if inner == 0 {
			return 0
		}
		// The inner chain has to be the value this test reads, or the two are unrelated chains that
		// merely nest. Upstream's example is `a(c?.d)?.d`, where the inner optional belongs to the
		// argument rather than to the outer chain.
		if branch.Test.Identifier != inner {
			return 0
		}
		innerDependency, hasInner := t.result.TemporariesReadInOptional[inner]
		if !hasInner {
			return 0
		}
		if !terminal.Optional {
			// A non-optional load riding an outer chain, the `.c` in `a?.b.c`. Loads from the inner
			// value are hoistable, because the short circuit already guaranteed it is non-null.
			t.result.HoistableObjects[block.Id] = innerDependency
		}
		baseObject = innerDependency
		test = branch
		testBlockId = testBlock.Id

	default:
		return 0
	}

	if outerAlternate != nil && test.Alternate == *outerAlternate &&
		len(block.Instructions) != 0 {
		// Upstream raises here: instructions in an inner optional block indicate two unrelated
		// chains being concatenated. Declining keeps the same guarantee without the crash.
		return 0
	}

	match := t.matchOptionalTestBlock(test)
	if match == nil {
		// Not hoistable, e.g. `a?.[computed()]`.
		return 0
	}
	if match.consequentGoto != terminal.Fallthrough {
		return 0
	}

	load := ReactiveScopeDependency{
		Identifier: baseObject.Identifier,
		Reactive:   baseObject.Reactive,
		Path: append(append([]DependencyPathEntry{}, baseObject.Path...),
			DependencyPathEntry{Property: match.property, Optional: terminal.Optional}),
	}
	t.result.ProcessedInstructions[match.storeLocalId] = true
	t.result.ProcessedOptionalTests[testBlockId] = true
	t.result.TemporariesReadInOptional[match.consequentId] = load
	t.result.TemporariesReadInOptional[match.propertyId] = load
	return match.consequentId
}
