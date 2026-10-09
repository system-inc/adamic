package control_flow_graph

import (
	"testing"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/parser"
)

func TestPathAnalysis(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name          string
		code          string
		event         string
		wantReachable bool
		wantCyclic    bool
		wantEvery     bool
		wantEarly     bool
	}{
		{
			name:          "unconditional after branch",
			code:          `function f(x) { if (x) consume(); markHook(); }`,
			event:         "markHook",
			wantReachable: true,
			wantEvery:     true,
		},
		{
			name:          "conditional branch",
			code:          `function f(x) { if (x) markHook(); }`,
			event:         "markHook",
			wantReachable: true,
			wantEvery:     false,
		},
		{
			name:          "short circuit before early return",
			code:          `function f(x, y) { if (!x || y) return; markHook(); }`,
			event:         "markHook",
			wantReachable: true,
			wantEvery:     false,
			wantEarly:     true,
		},
		{
			name:          "loop condition",
			code:          `function f() { for (; markHook(); ) {} }`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    false,
			wantEvery:     true,
		},
		{
			name:          "short circuit inside loop condition",
			code:          `function f(x) { for (; x && markHook(); ) {} }`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    true,
			wantEvery:     false,
		},
		{
			name:          "empty for loop header",
			code:          `function f() { for (;;) { markHook(); } }`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    false,
			wantEvery:     true,
		},
		{
			name:          "branch inside empty for loop",
			code:          `function f(x) { for (;;) { if (x) markHook(); } }`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    true,
			wantEvery:     true,
		},
		{
			name:          "loop body",
			code:          `function f(x) { for (; x; ) { markHook(); } }`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    true,
			wantEvery:     false,
		},
		{
			name: "switch break stops outer cycle",
			code: `function f(xs) {
				for (const x of xs) switch (x) {
					case 0: break;
					case 1: markHook(); break;
				}
			}`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    false,
			wantEvery:     false,
			wantEarly:     true,
		},
		{
			name: "switch return keeps outer cycle",
			code: `function f(xs) {
				for (const x of xs) switch (x) {
					case 0: return;
					case 1: markHook(); break;
				}
			}`,
			event:         "markHook",
			wantReachable: true,
			wantCyclic:    true,
			wantEvery:     false,
			wantEarly:     true,
		},
		{
			name:          "uncaught throw is not a final path",
			code:          `function f(x) { if (x) throw x; markHook(); }`,
			event:         "markHook",
			wantReachable: true,
			wantEvery:     true,
		},
		{
			name:          "unreachable",
			code:          `function f() { return; markHook(); }`,
			event:         "markHook",
			wantReachable: false,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			graph, locations := buildPathTestGraph(t, testCase.code)
			analysis := AnalyzePaths(graph)
			blocks := locations[testCase.event]
			if len(blocks) != 1 {
				t.Fatalf("event %q was laid out in %d blocks, want 1", testCase.event, len(blocks))
			}
			block := blocks[0]
			if got := block.Reachable; got != testCase.wantReachable {
				t.Fatalf("reachable = %v, want %v", got, testCase.wantReachable)
			}
			if !block.Reachable {
				return
			}
			if got := analysis.IsCyclic(block); got != testCase.wantCyclic {
				t.Errorf("cyclic = %v, want %v", got, testCase.wantCyclic)
			}
			if got := analysis.IsOnEveryFinalPath(block); got != testCase.wantEvery {
				t.Errorf("on every final path = %v, want %v", got, testCase.wantEvery)
			}
			shortestFinal, hasFinal := analysis.ShortestExitPathFromStart()
			shortestEvent, reachable := analysis.ShortestPathFromStart(block)
			gotEarly := hasFinal && reachable && shortestFinal < shortestEvent
			if analysis.IsExit(block) {
				gotEarly = hasFinal && reachable && shortestFinal <= shortestEvent
			}
			gotEarly = gotEarly && !analysis.IsOnEveryFinalPath(block)
			if gotEarly != testCase.wantEarly {
				t.Errorf("early final = %v (final=%d, event=%d), want %v", gotEarly, shortestFinal, shortestEvent, testCase.wantEarly)
			}
		})
	}
}

func TestPossibleThrowThroughFinallyIsExcluded(t *testing.T) {
	t.Parallel()
	graph, locations := buildPathTestGraph(t, `
		function f() {
			try {
				mayThrow();
				markInside();
			} finally {
				cleanup();
			}
			markAfter();
		}
	`)
	analysis := AnalyzePaths(graph)
	if !analysis.IsOnEveryFinalPath(locations["markInside"][0]) {
		t.Fatal("the thrown path that bypasses markInside should not count as a final path")
	}
	if !analysis.IsOnEveryFinalPath(locations["markAfter"][0]) {
		t.Fatal("a possible throw through finally should rejoin before markAfter")
	}
}

func TestYieldDoesNotHideLoopBackEdge(t *testing.T) {
	t.Parallel()
	graph, locations := buildPathTestGraph(t, `
		function* f(condition) {
			while (condition) {
				markHook();
				yield condition;
			}
		}
	`)
	block := locations["markHook"][0]
	if analysis := AnalyzePaths(graph); !analysis.IsCyclic(block) {
		t.Fatal("a resumable yield must not hide the loop back edge")
	}
}

func TestFinalPathCountSaturates(t *testing.T) {
	t.Parallel()
	for _, testCase := range []struct {
		name string
		code string
		want bool
	}{
		{name: "return", code: `function f() { return; }`, want: true},
		{name: "throw", code: `function f() { throw value; }`, want: false},
		{name: "continue plus loop exit", code: `function f(x) { while (x) { continue; } }`, want: true},
		{name: "break plus loop exit", code: `function f(x) { while (x) { break; } }`, want: false},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			graph, _ := buildPathTestGraph(t, testCase.code)
			if got := AnalyzePaths(graph).HasSingleFinalPath(); got != testCase.want {
				t.Fatalf("single final path = %v, want %v", got, testCase.want)
			}
		})
	}
}

func TestPathAnalysisRejectsProvisionallyReachableLayout(t *testing.T) {
	t.Parallel()
	graph, locations := buildPathTestGraph(t, `
		function f(condition) {
			for (; condition; markHook()) {
				return;
			}
		}
	`)
	block := locations["markHook"][0]
	if !block.Reachable {
		t.Fatal("the incrementor should be provisionally reachable while it is laid out")
	}
	if _, reachable := AnalyzePaths(graph).ShortestPathFromStart(block); reachable {
		t.Fatal("the incrementor should have no actual path from the graph entry")
	}
}

func buildPathTestGraph(t *testing.T, code string) (*Graph[string], map[string][]*Block[string]) {
	t.Helper()
	source := parser.ParseSourceFile(ast.SourceFileParseOptions{
		FileName: "/test.ts",
		PathKey:  "/test.ts",
	}, code, core.ScriptKindTS)
	var root *ast.Node
	source.AsNode().ForEachChild(func(node *ast.Node) bool {
		if ast.IsFunctionLike(node) {
			root = node
			return true
		}
		return false
	})
	if root == nil {
		t.Fatal("test source has no function root")
	}

	graph := Build(root, Hooks[string]{
		Expression: func(builder *Builder[string], node *ast.Node) {
			if node.Kind != ast.KindCallExpression {
				return
			}
			callee := node.AsCallExpression().Expression
			if callee != nil && callee.Kind == ast.KindIdentifier {
				name := callee.AsIdentifier().Text
				if len(name) >= 4 && name[:4] == "mark" {
					builder.Emit(name)
				}
			}
		},
	})
	locations := make(map[string][]*Block[string])
	for _, block := range graph.Blocks {
		for _, event := range block.Events {
			locations[event] = append(locations[event], block)
		}
	}
	return graph, locations
}

/*
 * An implicit exception inside a `catch` body leaves through the thrown path, not the normal one.
 *
 * `firstThrowableFork` had a branch that, for a throwable node inside a `catch` of a `try` with a
 * `finally`, recorded the current block as an *implicit* source of the statement's ordinary
 * continuation. That made the `finally`'s normal entry reachable from a path that only exists
 * because something threw, and so made the enclosing function's end reachable when every path
 * through it returns or throws.
 *
 * ESLint does not have that branch. `makeFirstThrowablePathInTryOrCatchBlock`
 * (`code-path-analysis/code-path-state.js:1780`) adds the head to `thrownForkContext` for both the
 * `try` and `catch` positions, with `position !== "catch" || !hasFinalizer` only deciding whether
 * to fork at all rather than where the fork goes.
 *
 * The shape below is the minimal one. A declaration in the `try` supplies the throwable node that
 * triggers the fork; without it there is nothing that can throw before the `return`, which is why
 * `noDeclaration` was correct even while this was broken. The `finally` supplies the frame the
 * branch keyed on. The `catch` rethrowing is what makes the end genuinely unreachable, so it is the
 * case where a spurious normal edge is observable.
 *
 * On the ahra tree this took `consistent-return` from 141 findings to 131, matching eslint exactly,
 * and moved no other rule that reads this graph.
 */
func TestEndIsUnreachableWhenACatchRethrowsBesideAFinally(t *testing.T) {
	t.Parallel()

	for _, testCase := range []struct {
		name string
		code string
		want bool
	}{
		{
			name: "declarationInTryWithRethrowingCatchAndFinally",
			code: "function f() { try { const r = g(); return r; } catch (e) { throw e; } finally { k(); } }",
			want: false,
		},
		{
			// No throwable node before the `return`, so no fork is created at all. Correct before
			// the repair and after it, and kept so a regression that removed the fork entirely
			// could not pass this table.
			name: "noDeclarationInTry",
			code: "function f() { try { return 1; } catch (e) { throw e; } finally { k(); } }",
			want: false,
		},
		{
			// No `finally`, so the branch that was removed never applied here.
			name: "noFinally",
			code: "function f() { try { const r = g(); return r; } catch (e) { throw e; } }",
			want: false,
		},
		{
			// The `catch` returns rather than rethrows, so every path still leaves with a value.
			name: "catchReturns",
			code: "function f() { try { const r = g(); return r; } catch (e) { return 2; } finally { k(); } }",
			want: false,
		},
		{
			// The end genuinely is reachable here: the `catch` neither returns nor throws, so
			// control falls out of the statement and off the end of the function. This is the
			// control that fails if a repair reaches the other four by declaring the end
			// unreachable whenever a `finally` is present.
			name: "catchFallsThroughSoEndIsReachable",
			code: "function f() { try { const r = g(); return r; } catch (e) { h(e); } finally { k(); } }",
			want: true,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			t.Parallel()
			graph, _ := buildPathTestGraph(t, testCase.code)
			if graph == nil {
				t.Fatal("no graph")
			}
			if graph.EndReachable != testCase.want {
				t.Errorf("EndReachable = %v, want %v", graph.EndReachable, testCase.want)
			}
		})
	}
}
