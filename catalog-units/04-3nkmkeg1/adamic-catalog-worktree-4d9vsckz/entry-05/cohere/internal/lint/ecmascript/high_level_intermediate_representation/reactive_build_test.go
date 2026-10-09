package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
	"testing"
)

// TestBuildReactiveFunctionShapes asserts the tree each control-flow construct converts to.
//
// One source per construct, chosen so the four ways a block can leave the graph are each covered
// once: a branch with a join, a ternary and a logical (both value terminals), and a loop with a
// back edge.
//
// The assertion is instruction CONSERVATION rather than a printed shape. A golden string would fail
// on every cosmetic change to the printer and would not notice the one failure that matters, which
// is an instruction entering the graph and not arriving in the tree. Conservation is the property
// upstream's own conversion has and the property every downstream pass depends on: a pass that reads
// the tree to decide what a scope depends on cannot see an instruction the converter dropped.
func TestBuildReactiveFunctionShapes(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name           string
		valueTerminals int
		// A reconstructed value terminal introduces one reactive instruction for the HIR terminal
		// itself. Its test and arms remain nested values, so this is one structural instruction
		// beyond the graph's instruction table rather than a duplicated block.
		reactiveExtra int
		source        string
	}{
		{name: "branch with join", source: `function f(a) { if (a) { return 1; } return 2; }`},
		{name: "loop with back edge", source: `function f(xs) { let t = 0; for (const x of xs) { t = t + x; } return t; }`},
		{name: "ternary", valueTerminals: 1, reactiveExtra: 1,
			source: `function f(a) { const x = a ? 1 : 2; return x; }`},
		{name: "logical", valueTerminals: 1, reactiveExtra: 1,
			source: `function f(a) { const x = a && a.b; return x; }`},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, _ := rangesFor(t, testCase.source)
			if function == nil {
				t.Fatal("the source did not lower to a function, so every assertion below would " +
					"pass vacuously")
			}

			graphInstructions := 0
			valueTerminals := 0
			for _, block := range function.Blocks {
				if block == nil {
					continue
				}
				graphInstructions += len(block.Instructions)
				switch block.Terminal.(type) {
				case *Logical, *Ternary, *Optional, *Sequence:
					valueTerminals++
				}
			}
			if graphInstructions == 0 {
				t.Fatal("the graph holds no instructions, so conservation is trivially satisfied " +
					"and this case measures nothing")
			}

			// The case's own claim about its lowering is asserted, not assumed. If a change to
			// lowering stops emitting a ternary as a value terminal, this test would otherwise
			// quietly start applying the wrong assertion below and still pass.
			if valueTerminals != testCase.valueTerminals {
				t.Fatalf("this case claims valueTerminals=%d but its lowering says %d; the "+
					"claim decides which assertion applies, so a stale one hides a real change",
					testCase.valueTerminals, valueTerminals)
			}

			tree, result := BuildReactiveFunction(function)
			if tree == nil {
				t.Fatal("BuildReactiveFunction returned no tree for a function that lowered")
			}

			if result.DoubleEmitted != 0 {
				t.Errorf("%d block(s) emitted twice; a block reached from two places must be "+
					"scheduled by its parent and broken to, not walked again",
					result.DoubleEmitted)
			}
			if result.UnmatchedGotos != 0 {
				t.Errorf("%d goto(s) found no enclosing construct to break to, so the tree names a "+
					"target that is not on the control-flow stack", result.UnmatchedGotos)
			}

			wantInstructions := graphInstructions + testCase.reactiveExtra
			if result.Instructions != wantInstructions {
				t.Errorf("reactive tree holds %d instructions, want %d from %d graph instructions "+
					"plus %d reconstructed terminal instruction(s)", result.Instructions,
					wantInstructions, graphInstructions, testCase.reactiveExtra)
			}
		})
	}
}

func TestBuildReactiveFunctionOptionalValue(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   int
	}{
		{name: "single link", source: `function f(a) { return a?.b; }`, want: 1},
		{name: "nested links", source: `function f(a) { return a?.b?.c; }`, want: 2},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, _ := rangesFor(t, testCase.source)
			if function == nil {
				t.Fatal("source did not lower")
			}

			graphOptionals := 0
			joinPhis := map[static_single_assignment.IdentifierId]bool{}
			for _, block := range function.Blocks {
				if terminal, ok := block.Terminal.(*Optional); ok {
					graphOptionals++
					if join, ok := function.Block(terminal.Fallthrough); ok {
						for _, phi := range join.Phis {
							joinPhis[phi.Place.Identifier] = true
						}
					}
				}
			}
			if graphOptionals != testCase.want {
				t.Fatalf("graph holds %d Optional terminals, want %d; the test no longer reaches "+
					"the reconstruction path", graphOptionals, testCase.want)
			}

			tree, result := BuildReactiveFunction(function)
			if tree == nil {
				t.Fatal("BuildReactiveFunction returned no tree")
			}
			if result.DoubleEmitted != 0 {
				t.Fatalf("reconstructing the chain emitted %d block(s) twice", result.DoubleEmitted)
			}

			values, statementIfs := 0, 0
			defined := map[static_single_assignment.IdentifierId]bool{}
			VisitReactiveFunction(tree, ReactiveVisitor{
				Instruction: func(instruction *ReactiveInstruction, traverse func()) {
					if instruction.LValue != nil {
						defined[instruction.LValue.Identifier] = true
					}
					traverse()
				},
				Value: func(_ static_single_assignment.EvaluationOrder, value ReactiveValue, traverse func()) {
					if optional, ok := value.(*ReactiveOptionalValue); ok {
						values++
						if _, ok := optional.Value.(*ReactiveSequenceValue); !ok {
							t.Errorf("optional value wraps %T, want the test/consequent sequence", optional.Value)
						}
					}
					traverse()
				},
				Terminal: func(statement *ReactiveTerminalStatement, traverse func()) {
					if _, ok := statement.Terminal.(*ReactiveIf); ok {
						statementIfs++
					}
					traverse()
				},
			})
			if values != testCase.want {
				t.Errorf("tree holds %d ReactiveOptionalValue nodes, want %d", values, testCase.want)
			}
			if statementIfs != 0 {
				t.Errorf("optional expression became %d statement-level if(s)", statementIfs)
			}
			for identifier := range joinPhis {
				if !defined[identifier] {
					t.Errorf("optional join phi %d is read after the expression but never defined in the tree",
						identifier)
				}
			}
		})
	}
}

func TestBuildReactiveFunctionTernaryValue(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name   string
		source string
		want   int
	}{
		{name: "single conditional", source: `function f(a, b, c) { return a ? b : c; }`, want: 1},
		{name: "nested conditional", source: `function f(a, b, c, d, e) { return a ? (b ? c : d) : e; }`, want: 2},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			function, _ := rangesFor(t, testCase.source)
			if function == nil {
				t.Fatal("source did not lower")
			}

			graphTernaries := 0
			joinPhis := map[static_single_assignment.IdentifierId]bool{}
			for _, block := range function.Blocks {
				if terminal, ok := block.Terminal.(*Ternary); ok {
					graphTernaries++
					if join, ok := function.Block(terminal.Fallthrough); ok {
						for _, phi := range join.Phis {
							joinPhis[phi.Place.Identifier] = true
						}
					}
				}
			}
			if graphTernaries != testCase.want {
				t.Fatalf("graph holds %d Ternary terminals, want %d; the test no longer reaches "+
					"the reconstruction path", graphTernaries, testCase.want)
			}
			tree, result := BuildReactiveFunction(function)
			if tree == nil {
				t.Fatal("BuildReactiveFunction returned no tree")
			}
			if result.DoubleEmitted != 0 {
				t.Fatalf("reconstructing the conditional emitted %d block(s) twice", result.DoubleEmitted)
			}

			values, statementIfs := 0, 0
			defined := map[static_single_assignment.IdentifierId]bool{}
			VisitReactiveFunction(tree, ReactiveVisitor{
				Instruction: func(instruction *ReactiveInstruction, traverse func()) {
					if instruction.LValue != nil {
						defined[instruction.LValue.Identifier] = true
					}
					traverse()
				},
				Value: func(_ static_single_assignment.EvaluationOrder, value ReactiveValue, traverse func()) {
					if conditional, ok := value.(*ReactiveTernaryValue); ok {
						values++
						if conditional.Test == nil || conditional.Consequent == nil || conditional.Alternate == nil {
							t.Error("conditional value does not contain its test and both arms")
						}
					}
					traverse()
				},
				Terminal: func(statement *ReactiveTerminalStatement, traverse func()) {
					if _, ok := statement.Terminal.(*ReactiveIf); ok {
						statementIfs++
					}
					traverse()
				},
			})
			if values != testCase.want {
				t.Errorf("tree holds %d ReactiveTernaryValue nodes, want %d", values, testCase.want)
			}
			if statementIfs != 0 {
				t.Errorf("conditional expression became %d statement-level if(s)", statementIfs)
			}
			for identifier := range joinPhis {
				if !defined[identifier] {
					t.Errorf("conditional join phi %d is read after the expression but never defined in the tree",
						identifier)
				}
			}
		})
	}
}

// TestBuildReactiveFunctionLogicalKeepsRightPrefix is the preservation-memoization golden that
// exposed the wrong logical arm. The `?? []` right block contains both ArrayExpression and a final
// LoadLocal; rebuilding only the final value drops the allocation, which in turn removes the scope
// declaration that the validator needs.
func TestBuildReactiveFunctionLogicalKeepsRightPrefix(t *testing.T) {
	t.Parallel()

	function, _ := rangesFor(t, `
		function Component(props) {
			const data = useMemo(() => {
				return props.items.edges.nodes ?? [];
			}, [props.items?.edges?.nodes]);
			return data;
		}
	`)
	if len(function.Functions) != 1 {
		t.Fatalf("component holds %d nested functions, want the useMemo callback", len(function.Functions))
	}
	callback := function.Functions[0]

	graphArrays, graphLogicals := 0, 0
	for _, block := range callback.Blocks {
		if _, ok := block.Terminal.(*Logical); ok {
			graphLogicals++
		}
		for _, instructionId := range block.Instructions {
			if _, ok := callback.Instructions[instructionId].Value.(*ArrayExpression); ok {
				graphArrays++
			}
		}
	}
	if graphArrays != 1 || graphLogicals != 1 {
		t.Fatalf("callback graph has %d arrays and %d logical terminals, want one of each; the "+
			"fixture no longer pins the lost-prefix shape", graphArrays, graphLogicals)
	}

	tree, result := BuildReactiveFunction(callback)
	if tree == nil {
		t.Fatal("BuildReactiveFunction returned no callback tree")
	}
	if result.DoubleEmitted != 0 {
		t.Fatalf("reconstructing the logical emitted %d block(s) twice", result.DoubleEmitted)
	}

	treeArrays, treeLogicals := 0, 0
	VisitReactiveFunction(tree, ReactiveVisitor{
		Value: func(_ static_single_assignment.EvaluationOrder, value ReactiveValue, traverse func()) {
			switch shape := value.(type) {
			case *ReactiveLogicalValue:
				treeLogicals++
			case *ReactiveInstructionValue:
				if _, ok := shape.Value.(*ArrayExpression); ok {
					treeArrays++
				}
			}
			traverse()
		},
	})
	if treeLogicals != graphLogicals {
		t.Errorf("tree holds %d logical values, want %d", treeLogicals, graphLogicals)
	}
	if treeArrays != graphArrays {
		t.Errorf("tree holds %d array allocations, want all %d from the graph", treeArrays, graphArrays)
	}
}

// TestBuildReactiveFunctionCorpusConservation is the same conservation property over real source.
//
// The four cases above prove the property on constructs written to exercise it; this proves it did
// not stop holding on code nobody wrote for this test. The two are not redundant: the hand-written
// cases localise a failure to one construct, and this one is the only thing that would notice a
// construct nobody thought to write a case for.
//
// # The input is the graph the PIPELINE hands this pass, and that is load-bearing
//
// `BuildReactiveScopeTerminals` runs before this stage and REWRITES the graph, so a measurement
// taken on a graph without scope terminals is not a measurement of this pass's real input. The
// difference is not cosmetic and it was measured rather than assumed: over the same 400 files the
// converter reports `doubleEmitted=0 unmatchedGotos=0` before terminals are built and
// `doubleEmitted=21 unmatchedGotos=61` after. An earlier spelling of this test omitted the terminal
// pass, passed clean on both counters, and would have pinned a green that hides every one of those
// 82 findings.
//
// The corpus numbers in this file's comments were taken at 400 files with the same walk
// `forEachCorpusFunction` performs, WITH terminals built. A number taken from a different
// denominator, or from a graph at a different pipeline stage, is not comparable to them.
func TestBuildReactiveFunctionCorpusConservation(t *testing.T) {
	t.Parallel()

	converted, lostWithValueTerminal, lostWithScopedLoopValue, lostWithout := 0, 0, 0, 0
	doubleEmitted, unmatchedGotos := 0, 0
	elidedScopeBreaks, nonImplicitScopeBreaks := 0, 0

	forEachCorpusFunction(t, 400, func(function *Function, ranges *mutation_aliasing.MutableRanges, scopes *ReactiveScopes) {
		// Bring the graph to the state the pipeline actually delivers before converting it.
		aligned, merged := AlignThenMergeReactiveScopes(function, scopes)
		BuildReactiveScopeTerminals(function, scopes, MergedScopeIdentity{Aligned: aligned, Merged: merged})

		graphInstructions := 0
		hasValueTerminal := false
		for _, block := range function.Blocks {
			if block == nil {
				continue
			}
			graphInstructions += len(block.Instructions)
			switch block.Terminal.(type) {
			case *Logical, *Ternary, *Optional, *Sequence:
				hasValueTerminal = true
			}
		}
		hasScopedLoopValue := hasScopeTerminalInLoopValueBlock(function)

		tree, result := BuildReactiveFunction(function)
		if tree == nil {
			return
		}
		converted++

		doubleEmitted += result.DoubleEmitted
		unmatchedGotos += result.UnmatchedGotos
		elidedScopeBreaks += result.ElidedScopeBreaks
		nonImplicitScopeBreaks += result.NonImplicitScopeBreaks

		if graphInstructions > 0 && result.Instructions < graphInstructions {
			switch {
			case hasValueTerminal:
				lostWithValueTerminal++
			case hasScopedLoopValue:
				lostWithScopedLoopValue++
			default:
				lostWithout++
			}
		}
	})

	if converted < 100 {
		t.Fatalf("only %d functions converted; the corpus walk is not reaching real source and "+
			"every count below would be a fact about the harness", converted)
	}

	// React-compatible reverse postorder closes the old break-target defect: unmatched gotos and
	// non-implicit scope breaks both fall to zero, and double emission falls from 21 to 4. The four
	// remaining double emissions all carry composite value terminals and stay under the declared
	// `ReactiveFunctionGapValueExpressions`.
	//
	// Correctly classifying for-of headers as loop blocks closes the seven scoped-loop loss cases
	// and one composite-value case. Alignment now encloses the loop instead of leaving a Scope
	// terminal inside an initializer or test that `ReactiveValue` cannot represent. The remaining
	// composite-value cases initially stayed unchanged. Immutable alias-edge refinement closes
	// findBottomBarDataKey's one-instruction loss; IMMUTABLE_ALIAS_EDGES.md attributes that movement.
	// The genuinely unexplained-loss count stays zero.
	const (
		knownLostWithValueTerminal = 4
		knownLostWithScopedLoop    = 0
		knownDoubleEmitted         = 4
	)
	if lostWithValueTerminal != knownLostWithValueTerminal ||
		lostWithScopedLoopValue != knownLostWithScopedLoop || lostWithout != 0 {
		t.Errorf("instruction loss: composite-value=%d (want %d), scoped-loop-value=%d (want %d), "+
			"unattributed=%d (want 0)", lostWithValueTerminal, knownLostWithValueTerminal,
			lostWithScopedLoopValue, knownLostWithScopedLoop, lostWithout)
	}
	if doubleEmitted != knownDoubleEmitted {
		t.Errorf("double-emitted blocks=%d, want %d; the remaining population is the declared "+
			"composite-value gap", doubleEmitted, knownDoubleEmitted)
	}
	if unmatchedGotos != 0 {
		t.Errorf("unmatched gotos=%d, want 0; React-compatible reverse postorder closed this gap",
			unmatchedGotos)
	}

	// The scope-fallthrough elision, asserted as a floor because a zero here would mean
	// `scopeFallthroughs` is never populated rather than never needed -- and that is exactly the
	// state this pass was in before the set existed, when it emitted 2,842 `break` statements the
	// source does not contain. Upstream omits every one of them.
	if elidedScopeBreaks == 0 {
		t.Errorf("no break to a scope fallthrough was elided across %d functions holding 2,874 "+
			"scope terminals; the set is not being populated, so the tree carries invented breaks",
			converted)
	}
	if nonImplicitScopeBreaks != 0 {
		t.Errorf("non-implicit scope breaks=%d, want 0; upstream treats this shape as impossible",
			nonImplicitScopeBreaks)
	}

	t.Logf("converted=%d lostWithValueTerminal=%d lostWithScopedLoopValue=%d lostWithout=%d "+
		"doubleEmitted=%d unmatchedGotos=%d elidedScopeBreaks=%d nonImplicitScopeBreaks=%d",
		converted, lostWithValueTerminal, lostWithScopedLoopValue, lostWithout, doubleEmitted, unmatchedGotos,
		elidedScopeBreaks, nonImplicitScopeBreaks)
}

func hasScopeTerminalInLoopValueBlock(function *Function) bool {
	if function == nil {
		return false
	}
	var valueBlocks []static_single_assignment.BlockId
	for _, block := range function.Blocks {
		switch terminal := block.Terminal.(type) {
		case *While:
			valueBlocks = append(valueBlocks, terminal.Test)
		case *DoWhile:
			valueBlocks = append(valueBlocks, terminal.Test)
		case *For:
			valueBlocks = append(valueBlocks, terminal.Init, terminal.Test)
			if HasBlock(terminal.Update) {
				valueBlocks = append(valueBlocks, terminal.Update)
			}
		case *ForOf:
			valueBlocks = append(valueBlocks, terminal.Init, terminal.Test)
		case *ForIn:
			valueBlocks = append(valueBlocks, terminal.Init)
		}
	}
	for _, blockID := range valueBlocks {
		block, ok := function.Block(blockID)
		if !ok {
			continue
		}
		if _, scoped := block.Terminal.(*Scope); scoped {
			return true
		}
	}
	return false
}

// TestReactiveFunctionGapsAreDeclared pins the gap list so closing one is a visible event.
//
// `ReactiveFunctionGaps` exists to be asserted on, for the reason its own comment gives: a gap that
// closes silently is an improvement nobody can point at, and a gap that opens silently is a
// regression nobody can either.
func TestReactiveFunctionGapsAreDeclared(t *testing.T) {
	t.Parallel()

	gaps := ReactiveFunctionGaps()
	if len(gaps) != 4 {
		t.Fatalf("expected 4 declared gaps, found %d; a gap was added or closed without this "+
			"test and its consumers being updated", len(gaps))
	}

	seen := map[ReactiveFunctionGap]bool{}
	for _, gap := range gaps {
		if seen[gap] {
			t.Errorf("gap %v is listed twice", gap)
		}
		seen[gap] = true
	}
	for _, required := range []ReactiveFunctionGap{
		ReactiveFunctionGapUnprunedLabels,
		ReactiveFunctionGapPrunedScopes,
		ReactiveFunctionGapUnbuiltTerminals,
		ReactiveFunctionGapValueExpressions,
	} {
		if !seen[required] {
			t.Errorf("gap %v is declared as a constant but not returned by ReactiveFunctionGaps, "+
				"so nothing asserts on it", required)
		}
	}
}
