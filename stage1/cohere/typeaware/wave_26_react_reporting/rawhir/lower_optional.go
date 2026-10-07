// Optional chains: `a?.b`, lowered to the block shape upstream produces.
//
// Upstream lowers `a?.b` to an `Optional` terminal with a value block per link and a join, so the
// chain becomes real control flow and each join carries a phi. The optionality lives on the
// TERMINAL, not on the `PropertyLoad` inside the block: the load is an ordinary one, and what makes
// the access optional is the branch that guarded reaching it. `optional_chains.go` reads
// `terminal.Optional` when it recovers the path, which is why that is the faithful place for it.
package high_level_intermediate_representation

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// isOptionalChainLink reports whether an access participates in an optional chain.
//
// A link participates when it carries `?.` itself OR when its object does, because `a?.b.c` is one
// chain: the second access rides the same short circuit even though it has no `?.` of its own. That
// is the distinction `Optional.Optional` records -- "whether this specific link is the one that
// tests" -- and getting it wrong here would lower `a?.b.c` as a chain plus a stray load.
func isOptionalChainLink(node *ast.Node) bool {
	for node != nil {
		switch node.Kind {
		case ast.KindPropertyAccessExpression:
			expression := node.AsPropertyAccessExpression()
			if expression.QuestionDotToken != nil {
				return true
			}
			node = expression.Expression
		case ast.KindElementAccessExpression:
			expression := node.AsElementAccessExpression()
			if expression.QuestionDotToken != nil {
				return true
			}
			node = expression.Expression
		default:
			return false
		}
	}
	return false
}

// lowerOptionalChain lowers one link of an optional chain into upstream's block shape.
//
// This is `lowerOptionalMemberExpression` (`BuildHIR.ts:2820`), and the shape is the reason it
// exists rather than a preference. `PropertyLoad.Optional` was carrying optionality as a flag on the
// instruction, which is not what the analyses downstream read: `collectOptionalChainSidemap` matches
// on an `Optional` TERMINAL and the branching around it, so with none it returned empty and the
// optional half of dependency collection never ran. This closes the dot-property portion of
// `DependencyGapOptionalChains`; computed links and optional calls remain declared there.
//
// The block shape, which is upstream's:
//
//	alternate    stores undefined into the result, then breaks to the continuation
//	test         evaluates the object, then branches on it
//	consequent   performs the load, stores into the result, breaks to the continuation
//	terminal     Optional{Optional: this link tests, Test: test, Fallthrough: continuation}
//
// `parentAlternate` is what makes a nested chain ONE chain. Upstream's comment: "we only create an
// alternate when first entering an optional subtree of the ast: if this is a child of an optional
// node, we use the alternate created by the parent". Both links of `a?.b?.c` short circuit to the
// same place, and creating a second alternate would lower them as two independent chains.
func (b *builder) lowerOptionalChain(node *ast.Node, parentAlternate *BasicBlock) Place {
	result := b.newTemporary(node)

	object, name, computed, tests := optionalChainParts(node)

	// The continuation INHERITS the current block's kind rather than forcing `BlockKindBlock`.
	// Upstream reserves `builder.currentBlockKind()` here (`BuildHIR.ts:2825`), where its `Logical`
	// and `Ternary` lowerings reserve a literal `'block'`.
	//
	// The difference is load-bearing and was measured. `AlignReactiveScopesToBlockScopes` widens a
	// scope to the program block containing it (`merge_scopes.go:626` reads
	// `Kind == BlockKindBlock`), so a chain lowered into block-kind continuations stretches every
	// scope inside it to the whole function. On
	// `repro-slow-validate-preserve-memo.ts`, whose value is a chain of `?.push(...)` calls, three
	// scopes went from `[2,44)`, `[17,20)` and `[31,41)` to `[2,48)`, `[2,46)` and `[2,46)` -- two
	// with identical ranges, which the reactive-tree builder cannot express as nested blocks, so
	// only one reached the tree and `PruneNonEscapingScopes` removed it.
	continuationKind := BlockKindValue
	if b.currentBlock != nil {
		continuationKind = b.currentBlock.Kind
	}
	continuation := b.reserve(continuationKind)

	// The alternate is filled before the chain is entered so a nested link can be handed the same
	// one, and it is reached only from the branch below rather than by falling into it. Filling it
	// through a terminal would orphan every block after it: the terminal would have to name a test
	// block that does not exist yet, and `ReversePostorder` drops what nothing reaches. Measured
	// when this was written the other way round -- `props?.a?.b` lowered to two blocks holding one
	// `Optional` and no `Branch`, where upstream produces two of each.
	alternate := parentAlternate
	if alternate == nil {
		alternate = b.reserve(BlockKindValue)
		saved := b.currentBlock
		b.enter(alternate)
		undefined := b.emit(&Primitive{Value: nil}, node)
		b.emit(&StoreLocal{LValue: result, Value: undefined, Kind: InstructionKindConst}, node)
		b.gotoBlock(continuation.Id, GotoVariantBreak)
		b.enter(saved)
	}

	testBlock := b.reserve(BlockKindValue)
	b.terminateAndEnter(&Optional{
		Optional:    tests,
		Test:        testBlock.Id,
		Fallthrough: continuation.Id,
	}, testBlock)

	// A nested link lowers into its own blocks and leaves the builder in ITS continuation, not in
	// the test block reserved above. The branch has to close whichever block the object's lowering
	// ended in, or it closes an empty one and the test never runs there. Measured on `props?.a?.b`
	// before this: the inner test block held no instructions and carried a `Branch` anyway, which
	// is a shape `matchOptionalTestBlock` rejects outright.
	var objectPlace Place
	if isOptionalChainLink(object) {
		objectPlace = b.lowerOptionalChain(object, alternate)
	} else {
		objectPlace = b.lowerExpressionToPlace(object)
	}
	consequent := b.reserve(BlockKindValue)
	b.terminateWith(&Branch{
		Test:        objectPlace,
		Consequent:  consequent.Id,
		Alternate:   alternate.Id,
		Fallthrough: continuation.Id,
	})

	b.enter(consequent)
	var loaded Place
	if computed != nil {
		property := b.lowerExpressionToPlace(computed)
		loaded = b.emit(&ComputedLoad{Object: objectPlace, Property: property}, node)
	} else {
		loaded = b.emit(&PropertyLoad{Object: objectPlace, Property: name}, node)
	}
	// `StoreLocal` rather than a load into the result, and the alternate matches. This is the shape
	// `matchOptionalTestBlock` keys on (`CollectOptionalChainDependencies.ts:176`): a consequent of
	// exactly `PropertyLoad` then `StoreLocal` ending in `Goto{Break}`, against an alternate of
	// exactly `Primitive` then `StoreLocal`. A load there would be rejected by the matcher and the
	// chain would go uncollected, which is the whole reason this lowering exists.
	b.emit(&StoreLocal{LValue: result, Value: loaded, Kind: InstructionKindConst}, node)
	b.gotoBlock(continuation.Id, GotoVariantBreak)

	b.enter(continuation)
	return result
}

// optionalChainParts reads one link's object, property name, computed key and whether it tests.
func optionalChainParts(node *ast.Node) (object *ast.Node, name string, computed *ast.Node,
	tests bool) {
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		expression := node.AsPropertyAccessExpression()
		if expression.Name() != nil {
			name = expression.Name().Text()
		}
		return expression.Expression, name, nil, expression.QuestionDotToken != nil
	case ast.KindElementAccessExpression:
		expression := node.AsElementAccessExpression()
		return expression.Expression, "", expression.ArgumentExpression,
			expression.QuestionDotToken != nil
	}
	return nil, "", nil, false
}
