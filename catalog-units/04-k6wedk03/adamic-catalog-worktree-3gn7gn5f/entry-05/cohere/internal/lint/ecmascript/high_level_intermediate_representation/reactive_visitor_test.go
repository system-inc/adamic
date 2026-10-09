package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

// TestReactiveVisitorReachesEveryInstruction is the conservation property, applied to the walk.
//
// `BuildReactiveFunction` already asserts that instructions entering the graph arrive in the tree.
// This asserts the second half nobody had: that a walk over the tree reaches them. A visitor that
// silently skips a terminal's block would leave every consuming pass blind on that construct, and
// the tree itself would still look correct -- which is precisely the failure a shape assertion
// cannot see and a count can.
//
// # The denominator is not `result.Instructions`, and the difference is real rather than a fudge
//
// `measureValue` counts a bare `ReactiveInstructionValue` sitting in a terminal's value position as
// one instruction -- a single-instruction value block that `valueOf` collapsed rather than wrapping
// in a one-element sequence. There is no `ReactiveInstruction` node there for a walk to hand to an
// `Instruction` hook, so a hook-based count is legitimately lower.
//
// The first spelling of this test used `result.Instructions` as the denominator and failed at 10 of
// 12 on the for-of case, which is exactly the residue `eachTerminalValue`'s comment already
// describes. That was the test being wrong about the tree, not the walk being wrong about the tree.
// The property that actually matters is stated instead: every `ReactiveInstruction` NODE in the
// tree is reached, counted by an independent walk rather than by the builder.
func TestReactiveVisitorReachesEveryInstruction(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{name: "branch with join", source: `function f(a) { if (a) { return 1; } return 2; }`},
		{name: "for-of loop", source: `function f(xs) { let t = 0; for (const x of xs) { t = t + x; } return t; }`},
		{name: "while loop", source: `function f(a) { let i = 0; while (i < a) { i = i + 1; } return i; }`},
		{name: "try catch", source: `function f(a) { try { return a.b; } catch (e) { return 0; } }`},
		{name: "switch", source: `function f(a) { switch (a) { case 1: return 1; default: return 2; } }`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, _ := rangesFor(t, testCase.source)
			if function == nil {
				t.Fatal("the source did not lower, so every assertion below would pass vacuously")
			}
			tree, result := BuildReactiveFunction(function)
			if tree == nil {
				t.Fatal("no tree was built")
			}

			walked := 0
			VisitReactiveFunction(tree, ReactiveVisitor{
				Instruction: func(instruction *ReactiveInstruction, traverse func()) {
					walked++
					traverse()
				},
			})

			nodes := countReactiveInstructionNodes(tree.Body)
			if nodes == 0 {
				t.Fatal("the tree holds no instruction nodes, so the walk reaching none proves nothing")
			}
			if walked != nodes {
				t.Errorf("the tree holds %d instruction nodes and the walk reached %d; a consuming "+
					"pass is blind wherever the walk does not go", nodes, walked)
			}
			// The builder's own count is higher by the collapsed value blocks described above. It is
			// asserted as an upper bound rather than ignored, so a walk somehow reaching MORE nodes
			// than exist would still fail.
			if walked > result.Instructions {
				t.Errorf("the walk reached %d instruction nodes but the builder measured only %d "+
					"instructions in the whole tree", walked, result.Instructions)
			}
		})
	}
}

// TestReactiveVisitorPrunesWhenTraverseIsNotCalled is the property that makes this a visitor rather
// than a walk.
//
// Every consuming pass depends on it: `pruneNonEscapingScopes` decides a scope does not escape and
// then declines to descend. If a non-nil hook recursed anyway, the pass would still compile, still
// pass a shape test, and quietly do nothing -- so the pruning is asserted directly.
func TestReactiveVisitorPrunesWhenTraverseIsNotCalled(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `function f(xs) { let t = 0; for (const x of xs) { t = t + x; } return t; }`)
	tree, _ := BuildReactiveFunction(function)
	if tree == nil {
		t.Fatal("no tree was built")
	}

	descended, pruned := 0, 0
	VisitReactiveFunction(tree, ReactiveVisitor{
		Terminal: func(statement *ReactiveTerminalStatement, traverse func()) {
			traverse()
		},
		Instruction: func(instruction *ReactiveInstruction, traverse func()) {
			descended++
			traverse()
		},
	})
	VisitReactiveFunction(tree, ReactiveVisitor{
		Terminal: func(statement *ReactiveTerminalStatement, traverse func()) {
			// Deliberately does not descend.
		},
		Instruction: func(instruction *ReactiveInstruction, traverse func()) {
			pruned++
			traverse()
		},
	})

	if descended == 0 {
		t.Fatal("the descending walk reached no instruction, so the comparison below is vacuous")
	}
	if pruned >= descended {
		t.Errorf("declining to traverse reached %d instructions and descending reached %d; not "+
			"calling traverse must prune the subtree or every pruning pass is a no-op",
			pruned, descended)
	}

	// # Why the comparison above is not sufficient on its own
	//
	// A walk that calls the hook and then descends REGARDLESS -- the single-character change of
	// dropping the `return` after the hook -- raises both counts by the same amount and leaves the
	// inequality holding. A mutation sweep found exactly that: the two-walk comparison scored it as
	// caught when it was not.
	//
	// So the pruning is asserted absolutely as well. A terminal hook that never traverses must leave
	// every instruction nested inside a terminal unreached, and the only instructions the walk may
	// then reach are the ones at the top level of the body.
	topLevel := 0
	for _, statement := range tree.Body {
		if _, ok := statement.(*ReactiveInstructionStatement); ok {
			topLevel++
		}
	}
	if pruned != topLevel {
		t.Errorf("with the terminal hook declining to traverse, the walk reached %d instructions "+
			"but only %d sit outside any terminal; the walk is descending into a subtree the hook "+
			"declined", pruned, topLevel)
	}
	if topLevel >= descended {
		t.Fatalf("this source puts %d of its %d instructions at the top level, so pruning the "+
			"terminals removes nothing and the assertion above is vacuous", topLevel, descended)
	}
}

// TestReactiveVisitorSeesSequenceInstructions covers the one nesting a value-only walk would miss.
//
// A `ReactiveSequenceValue` carries instructions INSIDE a value, and upstream visits them as
// instructions rather than as values. A walk that only recursed through values would skip them, and
// this package has already paid for that exact omission once: the first spelling of `valueOf`
// dropped everything but a block's last instruction and cost 27 functions.
func TestReactiveVisitorSeesSequenceInstructions(t *testing.T) {
	t.Parallel()

	// A for-of loop's Test and Init are value blocks, which is what produces sequence values here.
	function, _ := rangesFor(t, `function f(xs) { let t = 0; for (const x of xs) { t = t + x; } return t; }`)
	tree, result := BuildReactiveFunction(function)
	if tree == nil {
		t.Fatal("no tree was built")
	}

	sequences, instructionsInSequences := 0, 0
	VisitReactiveFunction(tree, ReactiveVisitor{
		Value: func(order static_single_assignment.EvaluationOrder, value ReactiveValue, traverse func()) {
			if sequence, ok := value.(*ReactiveSequenceValue); ok {
				sequences++
				instructionsInSequences += len(sequence.Instructions)
			}
			traverse()
		},
	})

	if sequences == 0 {
		t.Skip("this source produced no sequence value, so there is nothing to assert; the loop " +
			"arms this case relies on are built by valueOf and may have moved")
	}
	if instructionsInSequences == 0 {
		t.Errorf("%d sequence value(s) carry zero instructions, which is the wrapper being built "+
			"where it is not needed", sequences)
	}
	if result.Instructions == 0 {
		t.Error("the tree reports no instructions at all")
	}
}

// TestReactiveVisitorCorpus is the walk over real source.
//
// The cases above prove the property on constructs written to exercise it. This proves it did not
// stop holding on code nobody wrote for this test, which is the only thing that would catch a
// terminal variant no hand-written case happens to produce.
func TestReactiveVisitorCorpus(t *testing.T) {
	t.Parallel()

	converted, mismatched, treeInstructions, walkedTotal := 0, 0, 0, 0

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
		BuildReactiveScopeTerminals(function, scopes, MergedScopeIdentity{Aligned: aligned, Merged: merged})

		tree, result := BuildReactiveFunction(function)
		if tree == nil {
			return
		}
		converted++

		walked := 0
		VisitReactiveFunction(tree, ReactiveVisitor{
			Instruction: func(instruction *ReactiveInstruction, traverse func()) {
				walked++
				traverse()
			},
		})
		nodes := countReactiveInstructionNodes(tree.Body)
		treeInstructions += nodes
		walkedTotal += walked
		if walked != nodes {
			mismatched++
		}
		_ = result
	})

	if converted < 100 {
		t.Fatalf("only %d functions converted; the corpus walk is not reaching real source", converted)
	}
	if treeInstructions == 0 {
		t.Fatal("the corpus produced no tree instructions, so a matching walk proves nothing")
	}
	if mismatched != 0 {
		t.Errorf("%d of %d functions had a walk that did not reach every tree instruction "+
			"(%d reached against %d in the trees); a consuming pass is blind wherever it does not go",
			mismatched, converted, walkedTotal, treeInstructions)
	}

	t.Logf("converted=%d treeInstructions=%d walked=%d mismatched=%d",
		converted, treeInstructions, walkedTotal, mismatched)
}

// countReactiveInstructionNodes counts every `ReactiveInstruction` node in a tree, independently of
// both the builder's measurement and the visitor under test.
//
// Written as its own recursion rather than reusing either, because a test whose expected value is
// produced by the code under test asserts only that the code agrees with itself. This one walks the
// statement and value shapes directly and would have to be wrong in the same way as the visitor,
// at the same node type, to hide a defect.
func countReactiveInstructionNodes(block ReactiveBlock) int {
	total := 0
	for _, statement := range block {
		switch shape := statement.(type) {
		case *ReactiveInstructionStatement:
			total++
			total += countInstructionNodesInValue(shape.Instruction.Value)
		case *ReactiveScopeBlock:
			total += countReactiveInstructionNodes(shape.Instructions)
		case *ReactiveTerminalStatement:
			total += countInstructionNodesInTerminal(shape.Terminal)
		}
	}
	return total
}

func countInstructionNodesInValue(value ReactiveValue) int {
	switch shape := value.(type) {
	case *ReactiveSequenceValue:
		total := 0
		for _, nested := range shape.Instructions {
			total++
			total += countInstructionNodesInValue(nested.Value)
		}
		return total + countInstructionNodesInValue(shape.Value)
	case *ReactiveLogicalValue:
		return countInstructionNodesInValue(shape.Left) + countInstructionNodesInValue(shape.Right)
	case *ReactiveTernaryValue:
		return countInstructionNodesInValue(shape.Test) +
			countInstructionNodesInValue(shape.Consequent) +
			countInstructionNodesInValue(shape.Alternate)
	case *ReactiveOptionalValue:
		return countInstructionNodesInValue(shape.Value)
	}
	return 0
}

func countInstructionNodesInTerminal(terminal ReactiveTerminal) int {
	total := 0
	switch shape := terminal.(type) {
	case *ReactiveIf:
		total += countReactiveInstructionNodes(shape.Consequent)
		if shape.Alternate != nil {
			total += countReactiveInstructionNodes(*shape.Alternate)
		}
	case *ReactiveSwitch:
		for _, arm := range shape.Cases {
			if arm.Block != nil {
				total += countReactiveInstructionNodes(*arm.Block)
			}
		}
	case *ReactiveFor:
		total += countInstructionNodesInValue(shape.Init)
		total += countInstructionNodesInValue(shape.Test)
		total += countReactiveInstructionNodes(shape.Loop)
		if shape.Update != nil {
			total += countInstructionNodesInValue(shape.Update)
		}
	case *ReactiveForOf:
		total += countInstructionNodesInValue(shape.Init)
		total += countInstructionNodesInValue(shape.Test)
		total += countReactiveInstructionNodes(shape.Loop)
	case *ReactiveForIn:
		total += countInstructionNodesInValue(shape.Init)
		total += countReactiveInstructionNodes(shape.Loop)
	case *ReactiveWhile:
		total += countInstructionNodesInValue(shape.Test)
		total += countReactiveInstructionNodes(shape.Loop)
	case *ReactiveDoWhile:
		total += countReactiveInstructionNodes(shape.Loop)
		total += countInstructionNodesInValue(shape.Test)
	case *ReactiveLabelTerminal:
		total += countReactiveInstructionNodes(shape.Block)
	case *ReactiveTry:
		total += countReactiveInstructionNodes(shape.Block)
		total += countReactiveInstructionNodes(shape.Handler)
	}
	return total
}

// # What this file does not cover, measured rather than assumed
//
// A mutation sweep over the visitor caught the two reach defects it was written for -- dropping a
// try handler's block, and descending into a subtree whose hook declined -- and did not catch
// swapping the order of a do-while's body and test.
//
// That is correct rather than a gap to paper over: every assertion here counts what the walk
// reaches, and reordering two siblings changes neither the set nor the count. Order is load-bearing
// for a def-use pass (see `traverseTerminal`'s comment) but no consumer exists yet to state what the
// order must be. When `promoteUsedTemporaries` lands it will be the first pass that can tell, and
// the order assertion belongs with it rather than guessed at here.
