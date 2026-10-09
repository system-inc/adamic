package high_level_intermediate_representation

// branchingParentSource is the fixture both copy tests lower.
//
// Two properties are load-bearing and neither is incidental. The nested body branches, so it holds
// more than one block and its terminals name blocks -- a single-block body has no reference a
// skipped remap could leave stale. And the PARENT branches before it ever reaches the call, which
// pushes the parent's own block ids past the low numbers a freshly lowered nested body starts at.
// Measured: with a straight-line parent, the two id spaces coincided exactly, so a mutation that
// skipped the block remap left ids that were still inside the copied set and the test passed. The
// spread is what gives the assertion something to see.

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	shimchecker "github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/cohere/internal/lint/rule"
	"github.com/system-inc/cohere/internal/lint/testing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// TestCopyNestedBodyLeavesNoUnresolvedReference is the guard the whole file exists for.
//
// A copy that misses an id space produces a graph that builds, verifies, and points at a value from
// the child's table. Nothing downstream can tell that apart from a correct graph, so the assertion
// has to be made here or not at all.
//
// Four spaces cross the boundary and each is checked: every place resolves to an identifier the
// parent owns, every block reference names a block the parent holds, every instruction id indexes
// the parent's table, and every copied instruction is reachable from the copied entry.
func TestCopyNestedBodyLeavesNoUnresolvedReference(t *testing.T) {
	t.Parallel()

	// The closure carries a branch on purpose. A single-block body names no other block, so a
	// mutation skipping the block remap survives it: measured, and this fixture is what caught it.
	parent, nested, captures := loweredParentAndNested(t, branchingParentSource)
	if parent == nil || nested == nil {
		t.Fatal("the fixture produced no nested function, so this test asserts nothing about the " +
			"copy; the lowering changed rather than the copy being correct")
	}
	if len(nested.Blocks) < 2 {
		t.Fatalf("the nested body has %d block(s); with fewer than two, a terminal names no other "+
			"block and the block remap below is never exercised", len(nested.Blocks))
	}

	identifiersBefore := len(parent.Identifiers)
	blocksBefore := len(parent.Blocks)
	// Every identifier the parent already owns. A copied instruction naming one of these that is
	// NOT a capture means the remap was skipped and the id happened to be in range, which is the
	// failure an in-range check cannot see.
	preexisting := map[static_single_assignment.IdentifierId]bool{}
	for id := range parent.Identifiers {
		preexisting[static_single_assignment.IdentifierId(id)] = true
	}

	capturedIdentifiers := map[static_single_assignment.IdentifierId]bool{}
	for _, capture := range captures {
		capturedIdentifiers[capture.Identifier] = true
	}

	remap, ok := CopyNestedBodyInto(parent, nested, captures)
	if !ok {
		t.Fatal("the copy declined; with a matching context and capture list it should not")
	}
	if len(remap.Identifiers) == 0 || len(remap.Blocks) == 0 {
		t.Fatal("the remap is empty, so every assertion below passes vacuously")
	}
	if len(parent.Identifiers) <= identifiersBefore || len(parent.Blocks) <= blocksBefore {
		t.Error("the parent grew by nothing, so the body was not actually copied in")
	}

	copiedBlocks := map[static_single_assignment.BlockId]bool{}
	for _, mapped := range remap.Blocks {
		copiedBlocks[mapped] = true
	}

	for _, blockId := range sortedBlockIds(copiedBlocks) {
		block, found := parent.Block(blockId)
		if !found || block == nil {
			t.Fatalf("block %d is in the remap and not in the parent", blockId)
			continue
		}
		for _, instructionId := range block.Instructions {
			if int(instructionId) >= len(parent.Instructions) {
				t.Errorf("block %d names instruction %d, past the parent's table of %d",
					blockId, instructionId, len(parent.Instructions))
				continue
			}
			instruction := parent.Instructions[instructionId]
			if instruction == nil {
				t.Errorf("block %d names instruction %d, which the parent does not hold",
					blockId, instructionId)
				continue
			}
			EachInstructionPlacePointer(instruction, func(place *Place, role PlaceRole) {
				if int(place.Identifier) >= len(parent.Identifiers) ||
					parent.Identifiers[place.Identifier] == nil {
					t.Errorf("instruction %d names identifier %d, which the parent does not own; "+
						"an id space was missed by the copy", instructionId, place.Identifier)
					return
				}
				// In range is not enough. A skipped remap leaves the child's id, and the child's
				// ids start at zero, so they land inside the parent's table and read as valid.
				if preexisting[place.Identifier] && !capturedIdentifiers[place.Identifier] {
					t.Errorf("instruction %d names identifier %d, which the parent already owned "+
						"and which is not a capture; the remap was skipped for this place",
						instructionId, place.Identifier)
				}
			})
		}
		EachBlockReferencePointer(block.Terminal, func(reference *static_single_assignment.BlockId) {
			if _, found := parent.Block(*reference); !found {
				t.Errorf("block %d has a terminal naming block %d, which the parent does not hold",
					blockId, *reference)
				return
			}
			// Holding the block is not enough. The parent's ids and the nested body's both start
			// low, so a skipped remap leaves an id the parent happens to hold -- one of its OWN
			// blocks, not the copied one. Measured: a mutation skipping the block remap survived an
			// in-parent check entirely. Every reference out of a copied block must name a copied
			// block, since the body is self-contained until the caller splices it.
			if !copiedBlocks[*reference] {
				t.Errorf("copied block %d names block %d, which the parent holds but which is not "+
					"one of the copied blocks; the block remap was skipped for this reference",
					blockId, *reference)
			}
		})
	}

	// Phi operands are keyed by the predecessor block they arrive from, and that key is a block id
	// like any other -- it needs the same remap the terminals get, and it is reached by a different
	// code path, so the assertions above say nothing about it. A phi keyed by an unremapped
	// predecessor names a block in the nested body, which after the splice resolves to whatever the
	// parent happens to hold at that id: the operand is then read from the wrong edge.
	phiOperandsSeen := 0
	for _, blockId := range sortedBlockIds(copiedBlocks) {
		block, found := parent.Block(blockId)
		if !found || block == nil {
			continue
		}
		for _, phi := range block.Phis {
			for _, entry := range phi.Operands {
				predecessor, operand := entry.Predecessor, entry.Place
				phiOperandsSeen++
				if !copiedBlocks[predecessor] {
					t.Errorf("the phi in copied block %d takes an operand from block %d, which is "+
						"not one of the copied blocks; the predecessor remap was skipped",
						blockId, predecessor)
				}
				if preexisting[operand.Identifier] && !capturedIdentifiers[operand.Identifier] {
					t.Errorf("the phi in copied block %d takes identifier %d, which the parent "+
						"already owned and which is not a capture; the operand remap was skipped",
						blockId, operand.Identifier)
				}
			}
		}
	}
	if phiOperandsSeen == 0 {
		t.Error("no copied block holds a phi, so the operand assertions above passed vacuously; " +
			"the fixture must lower a nested body whose branches merge a value")
	}

	if _, found := parent.Block(remap.Entry); !found {
		t.Errorf("the remapped entry names block %d, which the parent does not hold", remap.Entry)
	}
}

// TestCopyNestedBodyKeepsCapturesPointingAtTheParent is the seeding half.
//
// A capture that gets a fresh identifier reads a different variable than the code around it, which
// is invisible in the graph's shape: the instruction is there, the value is wrong.
func TestCopyNestedBodyKeepsCapturesPointingAtTheParent(t *testing.T) {
	t.Parallel()

	const source = `
		function Component(properties: {value: number}) {
			const compute = () => {
				const doubled = properties.value * 2;
				return [doubled];
			};
			return compute();
		}
	`
	parent, nested, captures := loweredParentAndNested(t, source)
	if parent == nil || nested == nil {
		t.Skip("no nested function in this lowering")
	}
	if len(captures) == 0 {
		t.Fatal("the fixture captured nothing, so the seeding this test guards is never exercised")
	}

	remap, ok := CopyNestedBodyInto(parent, nested, captures)
	if !ok {
		t.Fatal("the copy declined")
	}
	for index, contextValue := range nested.Context {
		mapped, present := remap.Identifiers[contextValue.Identifier]
		if !present {
			t.Errorf("context value %d is not in the remap at all", contextValue.Identifier)
			continue
		}
		if mapped != captures[index].Identifier {
			t.Errorf("capture %d maps to %d, want the parent's %d; an inlined body reading a "+
				"fresh copy of a captured value reads a different variable",
				contextValue.Identifier, mapped, captures[index].Identifier)
		}
	}
}

// Construction runs before the Go inliner, so a nested body already contains several SSA values
// for one source binding. IdentifierIds must be renamed into the parent, while DeclarationId
// equivalence classes must survive the rename for the data-flow passes keyed by declarations.
func TestCopyNestedBodyPreservesDeclarationEquivalenceClasses(t *testing.T) {
	t.Parallel()

	parent, nested, captures := loweredParentAndNested(t, branchingParentSource)
	if parent == nil || nested == nil {
		t.Fatal("the fixture produced no nested function")
	}

	classes := map[static_single_assignment.DeclarationId][]static_single_assignment.IdentifierId{}
	for _, identifier := range nested.Identifiers {
		if identifier != nil && identifier.Declaration != 0 {
			classes[identifier.Declaration] = append(classes[identifier.Declaration], identifier.Id)
		}
	}
	multiValueClass := false
	for _, identifiers := range classes {
		if len(identifiers) > 1 {
			multiValueClass = true
			break
		}
	}
	if !multiValueClass {
		t.Fatal("the nested function has no declaration with multiple SSA values; the fixture no " +
			"longer exercises declaration equivalence")
	}

	remap, ok := CopyNestedBodyInto(parent, nested, captures)
	if !ok {
		t.Fatal("the copy declined")
	}
	parentClass := map[static_single_assignment.DeclarationId]static_single_assignment.DeclarationId{}
	for nestedDeclaration, identifiers := range classes {
		for _, identifier := range identifiers {
			mapped, found := remap.Identifiers[identifier]
			if !found {
				t.Fatalf("nested identifier %d has no parent mapping", identifier)
			}
			mappedIdentifier := parent.Identifiers[mapped]
			if mappedIdentifier == nil {
				t.Fatalf("mapped identifier %d is absent from the parent", mapped)
			}
			if declaration, seen := parentClass[nestedDeclaration]; seen {
				if mappedIdentifier.Declaration != declaration {
					t.Errorf("nested declaration %d split across parent declarations %d and %d",
						nestedDeclaration, declaration, mappedIdentifier.Declaration)
				}
			} else {
				parentClass[nestedDeclaration] = mappedIdentifier.Declaration
			}
		}
	}
}

// TestCopyNestedBodyDeclinesOnAContextMismatch pins the refusal.
func TestCopyNestedBodyDeclinesOnAContextMismatch(t *testing.T) {
	t.Parallel()

	parent := &Function{}
	nested := &Function{Context: []Place{{Identifier: 1}}}
	if _, ok := CopyNestedBodyInto(parent, nested, nil); ok {
		t.Error("the copy accepted a context of one against no captures; that mismatch means the " +
			"two sides disagree about what was closed over")
	}
}

func sortedBlockIds(set map[static_single_assignment.BlockId]bool) []static_single_assignment.BlockId {
	var ids []static_single_assignment.BlockId
	for id := range set {
		ids = append(ids, id)
	}
	for outer := 1; outer < len(ids); outer++ {
		for inner := outer; inner > 0 && ids[inner] < ids[inner-1]; inner-- {
			ids[inner], ids[inner-1] = ids[inner-1], ids[inner]
		}
	}
	return ids
}

// loweredParentAndNested lowers source and returns the first function holding a nested one.
func loweredParentAndNested(t *testing.T, source string) (*Function, *Function, []Place) {
	t.Helper()
	var parent, nested *Function
	var captures []Place
	probe := rule.Rule{
		Name:             "inline-remap-fixture",
		NeedsTypeChecker: true,
		Run: func(ctx rule.Context, options any) rule.Listeners {
			return rule.Listeners{
				ast.KindSourceFile: func(node *ast.Node) {
					if ctx.TypeChecker == nil {
						return
					}
					forEachFunctionLike(node, func(functionNode *ast.Node) {
						if parent != nil {
							return
						}
						lowered := Lower(functionNode, ctx.TypeChecker)
						if lowered == nil {
							return
						}
						Construct(lowered)
						for _, block := range lowered.Blocks {
							if block == nil {
								continue
							}
							for _, instructionId := range block.Instructions {
								instruction := lowered.Instructions[instructionId]
								if instruction == nil {
									continue
								}
								expression, isFunction := instruction.Value.(*FunctionExpression)
								if !isFunction {
									continue
								}
								if int(expression.Function) >= len(lowered.Functions) {
									continue
								}
								parent = lowered
								nested = lowered.Functions[expression.Function]
								captures = expression.Captures
								return
							}
						}
					})
				},
			}
		},
	}
	_ = shimchecker.Checker{}
	rule_testing.RunTypedFiles(t, probe, map[string]string{
		"/react.d.ts":  reactiveDeclarations,
		"/fixture.tsx": source,
	}, "/fixture.tsx")
	return parent, nested, captures
}

// TestCopyNestedBodyLeavesTheNestedFunctionUntouched pins that the copy is a copy.
//
// The remap rewrites through pointers handed out by the visitors, so a copied block that shares a
// terminal with the block it was copied from is renamed in place: the parent gets its ids and the
// nested function loses its own. That reads as working, because the parent is what the caller
// looks at, and the damage only surfaces when something later walks the nested body and finds it
// pointing at blocks in another function. The nested function is the control here.
func TestCopyNestedBodyLeavesTheNestedFunctionUntouched(t *testing.T) {
	t.Parallel()

	parent, nested, captures := loweredParentAndNested(t, branchingParentSource)
	if nested == nil {
		t.Fatal("the fixture lowered no nested function")
	}
	if len(nested.Blocks) < 2 {
		t.Fatalf("the fixture lowered %d nested blocks; a single-block body cannot catch a "+
			"reference that was rewritten in place", len(nested.Blocks))
	}

	type reference struct {
		block static_single_assignment.BlockId
		index int
	}
	before := map[reference]static_single_assignment.BlockId{}
	for _, block := range nested.Blocks {
		if block == nil {
			continue
		}
		index := 0
		EachBlockReferencePointer(block.Terminal, func(target *static_single_assignment.BlockId) {
			before[reference{block: block.Id, index: index}] = *target
			index++
		})
	}
	if len(before) == 0 {
		t.Fatal("the nested body names no blocks from any terminal; nothing here could catch a " +
			"rewrite of the nested function")
	}

	if _, ok := CopyNestedBodyInto(parent, nested, captures); !ok {
		t.Fatal("the copy declined a fixture it should accept")
	}

	for _, block := range nested.Blocks {
		if block == nil {
			continue
		}
		index := 0
		EachBlockReferencePointer(block.Terminal, func(target *static_single_assignment.BlockId) {
			key := reference{block: block.Id, index: index}
			index++
			if was, ok := before[key]; ok && was != *target {
				t.Errorf("nested block %d reference %d named block %d before the copy and names "+
					"block %d after it; the copy rewrote the nested function instead of the copy "+
					"it made of it", key.block, key.index, was, *target)
			}
		})
	}
}

const branchingParentSource = `
	function Component(properties: {value: number; flag: boolean; ready: boolean}) {
		if (!properties.ready) {
			return null;
		}
		if (properties.value < 0) {
			return [0];
		}
		const compute = () => {
			let doubled = properties.value * 2;
			if (properties.flag) { doubled = doubled + 1; }
			return [doubled];
		};
		return compute();
	}
`

// overlappingIdsParentSource is a memo callback long enough that its own identifier ids run past the
// first fresh id the parent mints for the copy, so some fresh parent id is also a nested id the remap
// holds. That overlap is what a place mapped twice needs in order to land somewhere wrong; without
// it a second mapping finds no entry and leaves the id alone. It is www's StackIngredients shape cut
// down, a `||` default feeding a second `||` (#p67vev4).
const overlappingIdsParentSource = `
	import React from 'react';
	declare const Day: { title: string }[];
	declare const Night: { title: string }[];
	export function useFilter(properties: { kind: 'A' | 'B' }) {
		const value = React.useMemo(function() {
			const x = properties.kind || 'A';
			const y = x === 'A' || x === 'B';
			return [y ? Day : Night];
		}, [properties.kind]);
		return value;
	}
`

// TestCopyNestedBodyMapsEveryPlaceOnce pins that each copied place is mapped exactly once, from its
// nested id. Mapping an lvalue a second time sent `$26` to `$51` and on to `$76`: a definition no
// operand reads, an id defined twice in the parent, and an inferred dependency rooted at nothing, which
// is how preserve-manual-memoization came to report a memoization React Compiler keeps (#p67vev4).
func TestCopyNestedBodyMapsEveryPlaceOnce(t *testing.T) {
	t.Parallel()

	parent, nested, captures := loweredParentAndNested(t, overlappingIdsParentSource)
	if parent == nil || nested == nil {
		t.Fatal("the fixture produced no nested function, so this test asserts nothing about the copy")
	}
	remap, ok := CopyNestedBodyInto(parent, nested, captures)
	if !ok {
		t.Fatal("the copy declined; with a matching context and capture list it should not")
	}

	overlapping := false
	for from, to := range remap.Identifiers {
		if _, isNestedId := remap.Identifiers[to]; isNestedId && to != from {
			overlapping = true
			break
		}
	}
	if !overlapping {
		t.Fatal("no fresh parent id is also a nested id, so a place mapped twice would land where it " +
			"should and this test could not fail; the fixture no longer overlaps the id ranges")
	}

	defined := map[static_single_assignment.IdentifierId]InstructionId{}
	for _, block := range nested.Blocks {
		if block == nil {
			continue
		}
		for _, sourceId := range block.Instructions {
			source := nested.Instructions[sourceId]
			copiedId, found := remap.Instructions[sourceId]
			if source == nil || !found {
				continue
			}
			copied := parent.Instructions[copiedId]
			if want := remap.Identifiers[source.LValue.Identifier]; copied.LValue.Identifier != want {
				t.Errorf("instruction %d defines identifier %d where the remap names %d for its nested "+
					"lvalue %d; the lvalue was mapped more than once",
					copiedId, copied.LValue.Identifier, want, source.LValue.Identifier)
			}
			if earlier, twice := defined[copied.LValue.Identifier]; twice {
				t.Errorf("identifier %d is defined by instruction %d and again by instruction %d",
					copied.LValue.Identifier, earlier, copiedId)
			}
			defined[copied.LValue.Identifier] = copiedId
		}
	}
	if len(defined) == 0 {
		t.Fatal("no copied instruction was checked, so the assertions above passed vacuously")
	}
}
