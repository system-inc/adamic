package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"testing"
)

// TestReactiveTransformKeepDoesNotReallocate is the property the lazy rebuild exists for.
//
// A pass that keeps every statement must leave the block identical, not merely equal. Asserted by
// slice identity rather than by contents, because a copy with the same elements passes any contents
// check and still costs an allocation per block on every pass that touches the tree.
func TestReactiveTransformKeepDoesNotReallocate(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `function f(a) { if (a) { return 1; } return 2; }`)
	tree, _ := BuildReactiveFunction(function)
	if tree == nil {
		t.Fatal("no tree was built")
	}
	if len(tree.Body) == 0 {
		t.Fatal("the tree body is empty, so identity below is trivially preserved")
	}

	before := tree.Body
	seen := 0
	TransformReactiveFunction(tree, ReactiveTransformer{
		Terminal: func(statement *ReactiveTerminalStatement, traverse func()) ReactiveTransformed {
			seen++
			traverse()
			return KeepStatement()
		},
	})

	if seen == 0 {
		t.Fatal("the transform reached no terminal, so keeping everything proves nothing")
	}
	if &before[0] != &tree.Body[0] {
		t.Error("a transform that kept every statement reallocated the block; the lazy rebuild " +
			"exists so an unchanged block costs nothing")
	}
}

// TestReactiveTransformRemovesAndReplaces covers the three rewriting answers.
//
// `replace-many` is the one that matters most: it is how `pruneNonEscapingScopes` flattens a scope,
// splicing the body into the enclosing block and dropping the wrapper.
func TestReactiveTransformRemovesAndReplaces(t *testing.T) {
	t.Parallel()

	source := `function f(a) { const x = a + 1; const y = x + 2; return y; }`

	t.Run("remove drops the statement", func(t *testing.T) {
		t.Parallel()
		function, _ := rangesFor(t, source)
		tree, _ := BuildReactiveFunction(function)
		before := countStatements(tree.Body)
		if before < 2 {
			t.Fatalf("the tree holds %d statements; this case needs at least two", before)
		}

		removed := false
		TransformReactiveFunction(tree, ReactiveTransformer{
			Instruction: func(statement *ReactiveInstructionStatement, traverse func()) ReactiveTransformed {
				if removed {
					return KeepStatement()
				}
				removed = true
				return RemoveStatement()
			},
		})
		if !removed {
			t.Fatal("the transform reached no instruction statement")
		}
		if after := countStatements(tree.Body); after != before-1 {
			t.Errorf("removing one statement left %d of %d", after, before)
		}
	})

	t.Run("replace-many splices", func(t *testing.T) {
		t.Parallel()
		function, _ := rangesFor(t, source)
		tree, _ := BuildReactiveFunction(function)
		before := countStatements(tree.Body)

		replaced := false
		TransformReactiveFunction(tree, ReactiveTransformer{
			Instruction: func(statement *ReactiveInstructionStatement, traverse func()) ReactiveTransformed {
				if replaced {
					return KeepStatement()
				}
				replaced = true
				// One statement becomes two copies of itself: the count must rise by exactly one.
				return ReplaceStatements([]ReactiveStatement{statement, statement})
			},
		})
		if !replaced {
			t.Fatal("the transform reached no instruction statement")
		}
		if after := countStatements(tree.Body); after != before+1 {
			t.Errorf("replacing one statement with two left %d, want %d", after, before+1)
		}
	})

	t.Run("replace-many with nothing removes", func(t *testing.T) {
		t.Parallel()
		function, _ := rangesFor(t, source)
		tree, _ := BuildReactiveFunction(function)
		before := countStatements(tree.Body)

		done := false
		TransformReactiveFunction(tree, ReactiveTransformer{
			Instruction: func(statement *ReactiveInstructionStatement, traverse func()) ReactiveTransformed {
				if done {
					return KeepStatement()
				}
				done = true
				return ReplaceStatements(nil)
			},
		})
		if after := countStatements(tree.Body); after != before-1 {
			t.Errorf("replacing one statement with none left %d of %d", after, before)
		}
	})
}

// TestReactiveTransformRewritesNestedBlocks is the assertion `transformTerminalBlocks` exists for.
//
// Each of its arms assigns the rewritten block back onto the terminal. A missing assignment leaves
// the terminal holding the pre-rewrite slice, which is invisible to any test that only counts what
// the walk reached -- the walk did reach it, and the result was thrown away.
func TestReactiveTransformRewritesNestedBlocks(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
	}{
		{name: "if arms", source: `function f(a) { if (a) { const x = a + 1; return x; } return 2; }`},
		{name: "for-of body", source: `function f(xs) { let t = 0; for (const x of xs) { t = t + x; } return t; }`},
		{name: "while body", source: `function f(a) { let i = 0; while (i < a) { i = i + 1; } return i; }`},
		{name: "try blocks", source: `function f(a) { try { const x = a.b; return x; } catch (e) { return 0; } }`},
		{name: "switch arms", source: `function f(a) { switch (a) { case 1: { const x = 1; return x; } default: return 2; } }`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, _ := rangesFor(t, testCase.source)
			tree, _ := BuildReactiveFunction(function)
			if tree == nil {
				t.Fatal("no tree was built")
			}

			before := countStatements(tree.Body)
			nested := before - len(tree.Body)
			if nested <= 0 {
				t.Fatalf("this source puts all %d statements at the top level, so nothing nested "+
					"is rewritten and the assertion is vacuous", before)
			}

			// Remove every instruction statement anywhere in the tree. Whatever survives must be
			// the non-instruction statements, so a nested block whose rewrite was discarded shows
			// up as a count that is too high.
			TransformReactiveFunction(tree, ReactiveTransformer{
				Instruction: func(statement *ReactiveInstructionStatement, traverse func()) ReactiveTransformed {
					return RemoveStatement()
				},
				Terminal: func(statement *ReactiveTerminalStatement, traverse func()) ReactiveTransformed {
					traverse()
					return KeepStatement()
				},
				Scope: func(scope *ReactiveScopeBlock, traverse func()) ReactiveTransformed {
					traverse()
					return KeepStatement()
				},
			})

			if remaining := countInstructionStatements(tree.Body); remaining != 0 {
				t.Errorf("%d instruction statement(s) survived a transform that removed every one; "+
					"a nested block's rewrite was computed and discarded", remaining)
			}
		})
	}
}

// TestReactiveTransformCorpus runs the same removal over real source.
//
// The hand-written cases above each hold one construct. This is the only thing that would catch a
// terminal arm no case happens to produce, which is exactly the failure `transformTerminalBlocks`
// invites by being written out by hand.
func TestReactiveTransformCorpus(t *testing.T) {
	t.Parallel()

	converted, survived, totalRemoved := 0, 0, 0

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
		BuildReactiveScopeTerminals(function, scopes, MergedScopeIdentity{Aligned: aligned, Merged: merged})

		tree, _ := BuildReactiveFunction(function)
		if tree == nil {
			return
		}
		converted++
		before := countInstructionStatements(tree.Body)

		TransformReactiveFunction(tree, ReactiveTransformer{
			Instruction: func(statement *ReactiveInstructionStatement, traverse func()) ReactiveTransformed {
				return RemoveStatement()
			},
			Terminal: func(statement *ReactiveTerminalStatement, traverse func()) ReactiveTransformed {
				traverse()
				return KeepStatement()
			},
			Scope: func(scope *ReactiveScopeBlock, traverse func()) ReactiveTransformed {
				traverse()
				return KeepStatement()
			},
		})

		totalRemoved += before
		if remaining := countInstructionStatements(tree.Body); remaining != 0 {
			survived++
		}
	})

	if converted < 100 {
		t.Fatalf("only %d functions converted; the corpus walk is not reaching real source", converted)
	}
	if totalRemoved == 0 {
		t.Fatal("no instruction statements were removed, so a zero survivor count proves nothing")
	}
	if survived != 0 {
		t.Errorf("%d of %d functions kept an instruction statement through a transform that "+
			"removed every one; a nested block's rewrite was discarded", survived, converted)
	}

	t.Logf("converted=%d removed=%d survived=%d", converted, totalRemoved, survived)
}

// countStatements counts every statement in a tree, at any depth.
//
// Written here rather than reusing the builder's measurement for the reason the visitor tests
// record: an expected value produced by the code under test asserts only self-agreement.
func countStatements(block ReactiveBlock) int {
	total := 0
	for _, statement := range block {
		total++
		switch shape := statement.(type) {
		case *ReactiveScopeBlock:
			total += countStatements(shape.Instructions)
		case *ReactiveTerminalStatement:
			eachNestedBlock(shape.Terminal, func(nested ReactiveBlock) {
				total += countStatements(nested)
			})
		}
	}
	return total
}

// countInstructionStatements counts only instruction statements, at any depth.
func countInstructionStatements(block ReactiveBlock) int {
	total := 0
	for _, statement := range block {
		switch shape := statement.(type) {
		case *ReactiveInstructionStatement:
			total++
		case *ReactiveScopeBlock:
			total += countInstructionStatements(shape.Instructions)
		case *ReactiveTerminalStatement:
			eachNestedBlock(shape.Terminal, func(nested ReactiveBlock) {
				total += countInstructionStatements(nested)
			})
		}
	}
	return total
}
