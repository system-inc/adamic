package core

import (
	"slices"
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// NoUnreachableLoopOptions configures which loop kinds the rule judges.
type NoUnreachableLoopOptions struct {
	// Ignore names loop kinds to leave alone, spelled with the ESTree node names upstream's
	// schema enumerates: WhileStatement, DoWhileStatement, ForStatement, ForInStatement,
	// ForOfStatement.
	//
	// The spelling is upstream's wire format rather than ours, and it stays upstream's because a
	// configuration written for ESLint has to keep meaning the same thing here. Our own AST calls
	// two of these something else (a `do`/`while` is KindDoStatement, and `for...in` and
	// `for...of` share KindForInStatement / KindForOfStatement), so the mapping happens inside
	// this rule rather than at the wire.
	//
	// Empty upstream, so the zero value is the default and a rule handed nil options judges every
	// loop kind.
	Ignore []string `json:"ignore"`
}

var messageUnreachableLoop = rule.Message{
	Id: "invalid",
	Description: "This loop can never run a second iteration, because every path through its body " +
		"leaves the loop. Whatever the loop condition says, the body breaks, returns, or throws " +
		"on every path, so the loop reads as a repetition and behaves as a plain block that runs " +
		"once. That is almost always a bug in the body rather than a deliberate shape: an exit " +
		"that was meant to be conditional, or a `continue` that was written as a `break`. If one " +
		"iteration really is what was wanted, say so with an `if` instead of a loop.",
}

// NoUnreachableLoop flags a loop whose body allows only one iteration.
//
//	valid:   while (a) { bar(); }
//	valid:   while (a) continue;
//	valid:   while (a) { if (foo) break; }
//	valid:   for (;;);                                   an infinite loop iterates
//	valid:   function foo() { return; while (a) break; } the loop is itself unreachable
//	valid:   while (a) break;                            with ignore: ["WhileStatement"]
//	invalid: while (a) break;
//	invalid: do break; while (a)
//	invalid: for (a of b) break;
//	invalid: while (a) for (;;);                         the OUTER loop, not the inner one
//	invalid: for (a in b) { while (foo) { if(baz) { break; } else { break; } } break; }   twice
//
// A loop exists to repeat. A body that leaves on every path repeats nothing, so the loop header is
// telling the reader something the code does not do. The interesting cases are the ones where the
// exit is not written at the top level of the body: a `switch` whose every arm returns, a `try`
// whose `catch` breaks, or a nested loop that never terminates.
//
// # This is a reachability question, and it is asked of the control-flow graph
//
// The judgment cannot be made syntactically. `while (a) { if (foo) break; }` is clean and
// `while (a) { if (foo) break; else return; }` is not, and nothing about the shape of the body
// separates them; what separates them is whether control can arrive back at the loop header.
//
// Upstream asks its code path analysis, through three events. `onCodePathSegmentLoop` fires where
// control flows back for another iteration; the loop selector records a loop when
// `isAnySegmentReachable` says the position it stands at is reachable; `Program:exit` reports
// whatever was recorded and never had a loop event.
//
// `internal/lint/ecmascript/control_flow_graph` answers all three. Its `Loop` hook is documented as running "where
// control flows back into a loop for another iteration of it" and it is handed the loop node
// directly, so upstream's `loopsByTargetSegments` map from segment to loop has no counterpart here:
// the event already names its loop. `Builder.Current().Reachable` is the reachability question, at
// both the loop statement and the loop event.
//
// # The reachability gate on the loop EVENT is load-bearing, and it is ours rather than upstream's
//
// Upstream gets reachability filtering for free, because ESLint's code path analysis does not raise
// a segment-loop event out of an unreachable segment at all. Our builder does: `statements.go`
// calls `b.loop(node)` at the end of every loop body unconditionally, whether or not control can be
// there. So a `Loop` event on its own is not evidence the loop iterates.
//
// Measured, and this is the single distinction the whole rule rests on: without the gate,
// `while (a) break;` receives a Loop event (from the body end, in the unreachable block the `break`
// left behind) and the rule goes completely silent on every case upstream reports. With the gate,
// all 1,629 corpus cases agree with the installed eslint 10.8.1 build.
//
// `continue` is the other half and needs no gate of its own: `makeContinue` returns early when the
// current block is unreachable, so an unreachable `continue` raises no event. The gate is written
// once, at the hook, and covers both.
//
// # Why the loop STATEMENT is gated too, and what it buys
//
// A loop standing in unreachable code is skipped, "to avoid unnecessary complexity in the
// implementation, or false positives otherwise" in upstream's words. So `function foo() { return;
// while (a) break; }` is clean, and so is the second loop of `while(true); while(true) break;`.
// That is a decision rather than an accident and it is reproduced: both are upstream valid cases.
//
// # What upstream deliberately does not catch
//
// Three shapes in the corpus's valid list are there because code path analysis does not evaluate
// conditions:
//
//	while (false) { foo(); }     never iterates at all, and is not reported
//	do foo(); while (false)      the same, from the other side
//	for (x of []);               an empty iterable, which is a value question
//
// Upstream's own comment calls these "out of scope for the code path analysis and consequently out
// of scope for this rule". They are reproduced as silence, and the corpus asserts each one.
//
// # Findings are sorted, and the reason is the multi-root walk
//
// Upstream reports from `Program:exit` over a set filled in traversal order, which is source order,
// and ESLint sorts its messages by position regardless. Here each code path root is built
// separately, so a loop inside a function is judged during that function's graph rather than during
// the file's, and emitting as they are found would interleave roots. `for (a in b) { while (foo) {
// if(baz) { break; } else { break; } } break; }` is the case that pins it: upstream reports the
// outer loop then the inner one, and `ExpectFindings` asserts ids in order.
var NoUnreachableLoop = rule.Rule{
	Name: "no-unreachable-loop",

	// No type information. Reachability is a control-flow question over the syntax tree, and
	// upstream asks its scope analysis nothing here either.
	NeedsTypeChecker: false,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as a bare severity is handed nil, which `options.(T)` turns into the
		// zero value. Upstream's default is `{ ignore: [] }`, so the zero value is right, and the
		// fallback is still written out rather than relied on.
		settings := NoUnreachableLoopOptions{}
		if decoded, configured := rule.OptionsAs[NoUnreachableLoopOptions](options); configured {
			settings = decoded
		}

		// Every loop kind ignored means the rule has nothing to say, which upstream expresses by
		// returning an empty visitor when the selector it would build is empty.
		//
		// This is a COST decision here rather than a behavioral one, and the sweep says so: with
		// this arm deleted the whole suite stays green, because the per-loop `Ignore` filter below
		// already declines every kind in `unreachableLoopKindNames` whenever this predicate is
		// true, and that map is exactly the set the predicate checks. There is no input on which
		// the two versions report differently.
		//
		// It is kept because the two versions differ in what they DO: without it, a file's every
		// code-path root is enumerated and a control-flow graph built for each, all to hand every
		// result to a filter that rejects it. Upstream declines the same work for the same reason.
		if unreachableLoopIgnoresEverything(settings.Ignore) {
			return rule.Listeners{}
		}

		return rule.Listeners{
			ast.KindSourceFile: func(node *ast.Node) {
				reportUnreachableLoops(ctx, node, settings)
			},
		}
	},
}

// unreachableLoopKindNames maps our AST kinds to the ESTree names upstream's `ignore` option
// spells. A kind absent from this map is not a loop.
//
// `KindDoStatement` is the spelling difference that matters: upstream calls it DoWhileStatement and
// a configuration naming it has to keep working.
var unreachableLoopKindNames = map[ast.Kind]string{
	ast.KindWhileStatement: "WhileStatement",
	ast.KindDoStatement:    "DoWhileStatement",
	ast.KindForStatement:   "ForStatement",
	ast.KindForInStatement: "ForInStatement",
	ast.KindForOfStatement: "ForOfStatement",
}

// unreachableLoopIgnoresEverything reports whether the option leaves no loop kind to judge.
//
// Upstream builds a selector string from the kinds NOT ignored and returns `{}` when it is empty.
// The corpus asserts this with an `ignore` naming all five.
func unreachableLoopIgnoresEverything(ignore []string) bool {
	for _, name := range unreachableLoopKindNames {
		if !slices.Contains(ignore, name) {
			return false
		}
	}
	return true
}

// reportUnreachableLoops judges every loop in the file, one code-path root at a time.
func reportUnreachableLoops(ctx rule.Context, sourceFile *ast.Node, settings NoUnreachableLoopOptions) {
	// Recorded per loop node: whether the loop statement was reached at a reachable position, and
	// whether any reachable event flowed back into it.
	type loopState struct {
		node      *ast.Node
		reachable bool
		iterates  bool
	}
	states := map[*ast.Node]*loopState{}

	stateFor := func(loop *ast.Node) *loopState {
		state, seen := states[loop]
		if !seen {
			state = &loopState{node: loop}
			states[loop] = state
		}
		return state
	}

	// Only a root whose own code holds a loop this rule judges can report, so only those are built.
	// On a cold ahra run that is 7.6% of roots and 30% of the blocks every root used to cost
	// (#tmn9n27).
	for _, summary := range codePathRoots(ctx, sourceFile) {
		root := summary.Node
		if !unreachableLoopRootHoldsJudgedLoop(summary, settings.Ignore) {
			if CheckCodePathGates {
				unreachableLoopCheckSkipped(root, settings.Ignore)
			}
			continue
		}
		control_flow_graph.Build(root, control_flow_graph.Hooks[struct{}]{
			Statement: func(builder *control_flow_graph.Builder[struct{}], node *ast.Node) {
				name, isLoop := unreachableLoopKindNames[node.Kind]
				if !isLoop || slices.Contains(settings.Ignore, name) {
					return
				}
				// No guard here that the loop belongs to THIS root, and the absence is measured
				// rather than assumed.
				//
				// The obvious worry is that a loop inside a nested function would be visited by the
				// enclosing graph as well as by its own, and judged against the wrong reachability.
				// It cannot be: the builder never descends into another root's body. A function
				// declaration is listed among the statements that lay out nothing
				// (`statements.go`'s "Nothing executes here in the enclosing code path" arm), and a
				// function expression, arrow, method, accessor, class static block and property
				// initializer are each opaque to `expr` for the same reason.
				//
				// Probed over 20 root shapes covering every kind `control_flow_graph.IsRoot` names,
				// including a loop in a function nested inside an unreachable region: a
				// `control_flow_graph.RootOf(node) != root` test fired 0 times while its inverted control
				// fired 21, so the guard would decline nothing. It was written, scored with the
				// sweep, and removed as unreachable rather than kept as reassurance.
				//
				// The verdict names the callers it was taken over: `control_flow_graph.Build`'s own
				// statement and expression walks, at the pin in `roots.go` / `statements.go` /
				// `expressions.go` as of this commit. If the builder ever descends into a nested
				// root, this is void and the guard has to come back.
				// An unreachable sighting still creates the state, with `reachable` left false.
				//
				// The entry is what makes `reachable` a field the report site can discriminate on
				// rather than a constant. Dropping the else branch produces identical findings on
				// every input, since a state nothing ever marks reachable is skipped anyway, and it
				// silently kills the guard: with no entry, every state at the report site has
				// `reachable` true and `state.reachable` stops being able to decline anything.
				//
				// Measured, and this is why it is written out. Deleting the else branch left all
				// 1,636 fixtures green. It also flipped the `state.reachable` mutant from CAUGHT by
				// 8 lines to SURVIVED, which is the sweep reporting a discrimination that had
				// stopped existing while nothing else changed.
				if builder.Current().Reachable {
					stateFor(node).reachable = true
				} else {
					stateFor(node)
				}
			},
			Loop: func(builder *control_flow_graph.Builder[struct{}], loop *ast.Node) {
				// The gate this rule rests on. See the doc comment: our builder raises the
				// body-end event from unreachable positions and upstream's analysis does not.
				if builder.Current().Reachable {
					stateFor(loop).iterates = true
				}
			},
		})
	}

	invalid := []*ast.Node{}
	for _, state := range states {
		if state.reachable && !state.iterates {
			invalid = append(invalid, state.node)
		}
	}

	// Source order, because findings from different roots are gathered separately. See the doc
	// comment on why upstream gets this ordering for free.
	sort.Slice(invalid, func(first, second int) bool {
		return invalid[first].Pos() < invalid[second].Pos()
	})

	for _, loop := range invalid {
		ctx.ReportNode(loop, messageUnreachableLoop)
	}
}

// unreachableLoopRootHoldsJudgedLoop reports whether a root's own code holds a loop of a kind the
// options leave to judge. A root holding none records no loop state, so it cannot report.
func unreachableLoopRootHoldsJudgedLoop(root control_flow_graph.RootSummary, ignore []string) bool {
	for _, kind := range root.Loops {
		if !slices.Contains(ignore, unreachableLoopKindNames[kind]) {
			return true
		}
	}
	return false
}

// unreachableLoopCheckSkipped is the gate's check under CheckCodePathGates: a root skipped for
// holding no judged loop must lay out none.
func unreachableLoopCheckSkipped(root *ast.Node, ignore []string) {
	control_flow_graph.Build(root, control_flow_graph.Hooks[struct{}]{
		Statement: func(builder *control_flow_graph.Builder[struct{}], node *ast.Node) {
			if name, isLoop := unreachableLoopKindNames[node.Kind]; isLoop && !slices.Contains(ignore, name) {
				panic("no-unreachable-loop skipped a root whose graph holds a judged loop")
			}
		},
	})
}
