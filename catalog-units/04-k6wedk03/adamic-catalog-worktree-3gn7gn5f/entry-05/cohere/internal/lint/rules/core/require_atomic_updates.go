package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageRequireAtomicUpdatesVariable = rule.Message{
	Id: "nonAtomicUpdate",
	Description: "This reassignment is not atomic. The value it is computed from was read before " +
		"an await or a yield, and the function is suspended in between, so anything else that runs " +
		"during the suspension can write the variable and have its write overwritten by this line. " +
		"The read and the write look adjacent in the source and are not adjacent in time. Read the " +
		"variable after the suspension, or keep the whole update on one side of it.",
}

var messageRequireAtomicUpdatesProperty = rule.Message{
	Id: "nonAtomicObjectUpdate",
	Description: "This property assignment is computed from an object that was read before an " +
		"await or a yield, so the object may have been replaced or mutated while the function was " +
		"suspended and this write lands on a stale view of it. The read and the write look " +
		"adjacent in the source and are not adjacent in time. Re-read the object after the " +
		"suspension, or keep the whole update on one side of it.",
}

// RequireAtomicUpdatesOptions configures whether property assignments are judged.
type RequireAtomicUpdatesOptions struct {
	// AllowProperties suppresses the property arm, leaving only assignments to the variable itself.
	//
	// Upstream's default is false, so both arms are on. The option exists because the property arm
	// is the noisier of the two: an object read before a suspension is very often a stable handle
	// whose identity nothing reassigns, and upstream added the escape hatch rather than narrow the
	// judgment. Named for what it permits rather than for what it disables, matching upstream's
	// spelling, so a config carried across from ESLint means the same thing here.
	AllowProperties bool `json:"allowProperties"`
}

// RequireAtomicUpdates flags an assignment whose value was computed from a read taken before an
// await or a yield suspended the function.
//
//	valid:   let foo; async function x() { foo += bar; }
//	valid:   let foo; async function x() { foo = await bar + foo; }
//	valid:   async function x() { let foo; foo += await bar; }
//	valid:   const foo = {}; async function x() { foo.bar = await baz; }
//	invalid: let foo; async function x() { foo += await amount; }
//	invalid: let foo; async function x() { foo = foo + await amount; }
//	invalid: const foo = {}; async function x() { foo.bar += await baz }
//
// A suspended function is not holding the world still. Between the read on the right and the store
// on the left, every other task queued in the same runtime gets to run, and any of them may write
// the same variable. The write here then lands carrying a value computed from what the variable held
// before the suspension, silently discarding whatever happened during it. The classic shape is a
// counter, `count += await next()`, which loses increments under concurrency and is correct in every
// single-stepped test.
//
// # The judgment is a flow question and cannot be answered from the assignment
//
// `foo = foo + await bar` reports and `foo = await bar + foo` does not, and the two differ only in
// which side of the suspension the read of `foo` sits on. Nothing about the assignment node, the
// symbol, or the enclosing function separates them: the discrimination is the ORDER of a read, a
// suspension, and a write along a path, which is exactly what a control-flow graph carries and what
// a syntactic walk cannot see. Upstream reaches this with ESLint's code path analysis, tracking two
// per-segment sets and moving variables between them.
//
// # The lattice, and why it is a forward analysis over one set
//
// Upstream carries `freshReadVariables` and `outdatedReadVariables` per segment, but the second is
// the only one whose value is asked at a report site: `isOutdated` consults it and nothing consults
// the fresh set directly. The fresh set exists to be moved into the outdated one at a suspension and
// to be cleared by a later read, so its whole contribution is bookkeeping between two events inside
// one block.
//
// That collapses. This runs a FORWARD dataflow over a single set, "variables read before a
// suspension that no later read has refreshed", met by union across predecessors. Inside a block the
// transfer walks events in order: a read removes the variable from the outdated set and records it
// as freshly read, a suspension moves every freshly read variable into the outdated set, and a write
// is judged against the set as it stands. Union is upstream's own meet, reached in
// `SegmentInfo#initialize` by folding every previous segment's two sets together, and `isOutdated`
// answering true if ANY current segment holds the variable is the same disjunction on the other
// side.
//
// The block-local fresh set is not part of the lattice value because it cannot survive a block
// boundary in upstream either: `initialize` copies a predecessor's fresh set forward, but the only
// thing that ever reads a fresh variable is `makeOutdated` at a suspension, and a suspension always
// starts a new segment upstream. Measured rather than argued: the whole 63-case corpus reproduces
// with the fresh set block-local, including case 2's loop and case 26's try/catch, which are the two
// shapes where a cross-block fresh set would show.
//
// # Which references are candidates, and the three filters upstream applies
//
// Upstream builds a reference map per function scope and only for scopes that can suspend, then
// applies three filters at the report site. All three are reproduced:
//
//	the function suspends       only an async function or a generator is judged at all. A plain
//	                            function has no suspension point, so no read can go outdated.
//	the reference resolves      an unresolved identifier is skipped by `createReferenceMap`
//	                            outright. See the divergence note below, because this is the one
//	                            place our substrate answers a different question than ESLint's.
//	the variable escapes        a variable only this function can observe cannot be written by
//	                            anything else during the suspension, so no race is possible.
//	                            `isLocalVariableWithoutEscape`.
//
// # The escape filter, measured against the installed rule rather than read
//
// This is the filter most likely to be ported slightly wrong, so each row below is a command that
// was run against ESLint 10.8.1 rather than a reading of the source:
//
//	async function x() { let foo; foo += await bar; }                       clean   declared inside,
//	                                                                               never captured
//	async function x() { let foo; bar(() => foo); foo += await amount; }    reports captured by a
//	                                                                               closure
//	async function x() { let foo; bar(() => baz += 1); foo += await b; }    clean   a closure that
//	                                                                               does not name it
//	let foo; async function x() { foo += await amount; }                    reports declared outside
//	async function x() { foo += await bar; }                                clean   unresolved
//	async function f(foo) { foo = await bar; }                              clean   a parameter, for
//	                                                                               the variable arm
//	async function f(foo) { let b = await get(foo.id); foo.bar = b.bar; }   reports a parameter, for
//	                                                                               the property arm
//	async function f() { let foo = {}; let b = await get(foo.id);
//	                     foo.prop = b.prop; }                               clean   a local object
//
// The last two are the parameter special case and they are not symmetric, which reads as an
// inconsistency and is deliberate upstream: `isMemberAccess && variable.defs.some(Parameter)`
// returns false EARLY, before the escape scan runs. A parameter's object is reachable by the caller
// whatever this function does with the binding, so a property write on it is observable outside; the
// binding itself is not, so a plain reassignment of it is local. Two questions, one predicate,
// opposite answers, and the corpus pins both.
//
// # Two divergences from ESLint, both measured on the whole tree and both kept
//
// Run end to end over `~/Projects/ahra` on 2026-08-26, cohere reports 39 and ESLint 31, and the
// difference is entirely one-way: cohere is a strict SUPERSET, with 8 findings ESLint declines and
// zero that ESLint reports and cohere misses. Both classes are recorded below with the reduction
// that isolates them, because each is a place a later reader would otherwise "fix" cohere back to
// ESLint's answer.
//
// # Divergence one, resolution: 7 of the 8
//
// Upstream skips a reference whose `resolved` is null, so an identifier naming nothing declared is
// never judged. ESLint decides "declared" from its own scope analysis plus the configured globals
// list, and cohere has neither: our answer comes from the checker, which resolves against the
// program's type declarations.
//
// The two disagree in one direction, and upstream's own corpus contains the case. Case 26 assigns
// `process.exitCode` and upstream ships it with `globals: { process: "readonly" }` in its language
// options. Measured on the installed rule: with that global declared it reports twice, and with the
// SAME source and no global declared it is completely silent. In this tree `process` resolves
// through the ambient node declarations, so we report where a bare ESLint run on an untyped file
// would not.
//
//	node drive2.js require-atomic-updates '[{"code": <case 26>, "globals": {"process": "readonly"}}]'
//	  -> REPORT[nonAtomicObjectUpdate,nonAtomicObjectUpdate]
//	node drive2.js require-atomic-updates '[{"code": <case 26>}]'
//	  -> clean
//
// The divergence is a strict superset in favour of reporting, and it is the right direction here:
// the reason upstream skips an unresolved name is that it cannot tell a global from a typo, not that
// a global cannot race. A global is in fact the MOST racy thing in the file. `TestRequireAtomicUpdatesResolutionDivergence`
// pins it so the next reader sees a measurement rather than inherits an argument.
//
// On the real tree this accounts for 7 of the 8 extras, all of them the same shape: a handler saved
// into a local, replaced, and restored in a `finally` after an await.
//
//	const originalLog = console.log;
//	console.log = function () { ... };
//	try { return await callback(); }
//	finally { console.log = originalLog; }
//
// `modules/art/ArtTerminal.ts:414-416` and `modules/phi/social/PhiSocialTerminal.ts:480-482` are
// that pattern over `console.log`, `console.info` and `process.stdout.write`. Measured on the
// installed rule, the SAME source reports when `console` is a declared global and is silent when it
// is not.
//
// Why those two names are undeclared HERE is narrower than "the config has no globals", and the
// narrow version is the one that survives a grep. `StructureLintConfiguration.ts` declares globals
// in two places and the two lists differ:
//
//	line 91   the JavaScript block's own list: __dirname, __filename, module, require,
//	          process, console, global, Buffer. Spread into the JS config at line 441,
//	          which applies to `**/*.{mjs,js,jsx}`.
//	line 158  StructureJavaScriptAndTypeScriptGlobals: React, document, window, navigator,
//	          setTimeout, clearTimeout, setInterval, clearInterval. This is the ONLY list
//	          the TypeScript block gets, at line 460.
//
// So `console` and `process` are declared for JavaScript files and not for TypeScript ones, and
// every finding in this class is in a `.ts` file. A reader who checks "no globals anywhere" will
// find eight of them and reasonably doubt the rest of this note; the true statement is that these
// two specific names are absent from the list `.ts` files receive.
//
// Two controls, because a zero here has three plausible causes and reading the config separates
// none of them. A seeded `alert()` in `ArtTerminal.ts` was reported by the same ESLint run, so the
// file is linted and the silence is the rule declining rather than the file being skipped. And a
// seeded pair in that same file, two structurally identical writes in one `finally`, one over a
// locally declared object and one over `console`, produced exactly one ESLint finding: the local
// one. Same file, same shape, same run, opposite verdicts, which is the resolution difference
// isolated to the one variable that changed.
//
// These are true positives. Two overlapping calls to such a wrapper restore in the wrong order and
// the second restore installs a filter that was already torn down, which is the last-writer-wins
// hazard this rule exists to name.
//
// # Divergence two, a refresh credited across a throwing edge: the 8th
//
// ESLint clears a variable's outdated mark at a read, and carries that clearing into the `catch` of
// an enclosing `try`. So a read at the END of a try block silences a write in the catch, even though
// the catch is reached precisely on the paths where that read did not run.
//
// Reduced to the pair that isolates it, measured on eslint 10.8.1:
//
//	async function f(e) { if (e.s !== 1) return;
//	    try { await a(); e.p = await b(); } catch (x) { e.err = 1; } }
//	  -> ESLint reports BOTH writes
//
//	async function f(e) { if (e.s !== 1) return;
//	    try { await a(); e.p = await b(); use(e.p); } catch (x) { e.err = 1; } }
//	  -> ESLint reports only `e.p`, and goes silent on `e.err`
//
// The two differ by the single read `use(e.p)`. Reading an UNRELATED variable there does not change
// the verdict, and adding a further await after the read restores it, which together show the
// mechanism is the refresh rather than statement count or position.
//
// The catch runs when `await b()` rejects. On that path `use(e.p)` never executes, so the refresh
// ESLint credits did not happen and `e` is still built from a pre-suspension read. Cohere reports
// it; ESLint does not. This is `modules/kingdom/KingdomShadeController.ts:203`, and it is the only
// one of the 8 where the two implementations disagree about a JUDGMENT rather than about what
// resolves.
//
// Kept deliberately. A refresh that only happens on the non-throwing path cannot make the throwing
// path safe, and our graph forks to the handler as a real successor edge, which is why we see it and
// upstream's segment bookkeeping does not.
// `TestRequireAtomicUpdatesCatchDoesNotInheritTryRefresh` pins both directions.
//
// # No fix and no suggestion, matching upstream
//
// `meta.fixable` is null and there is no `suggest` anywhere in the rule. The repair is a
// restructuring rather than a rewrite: moving the read past the suspension changes which value the
// expression sees, which is the entire point and is not something a linter may do unattended. There
// is also no single right answer, since the author may have meant to re-read, to lock, or to hold
// the value deliberately.
var RequireAtomicUpdates = rule.Rule{
	Name: "require-atomic-updates",

	// Symbol identity is what makes two occurrences the same binding, and it is what separates a
	// shadow in an inner block from the outer variable this rule is judging. There is no syntactic
	// substitute: `let foo` inside the async function and `let foo` outside it produce identical
	// identifier text and opposite verdicts.
	//
	// A mutant flipping this to false survives every fixture, and that is a property of the HARNESS
	// rather than a gap in the fixtures: `rule_testing.RunTyped` supplies a checker whatever the rule
	// declares, so no rule test can observe the declaration. What the field actually controls is
	// whether a real run acquires the checker for this file at all, so under-declaring it makes the
	// rule go silent in production while staying green here. Recorded rather than left as an
	// unexplained survivor; the guard that can see it is the dry run against the real tree.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as a bare severity is handed nil options, which `options.(T)` turns into
		// the zero value. The zero value is the right default here (upstream defaults
		// `allowProperties` to false) but the fallback is written explicitly rather than relied on,
		// because a rule whose default arrives by accident is one field rename away from silently
		// inverting.
		settings := RequireAtomicUpdatesOptions{AllowProperties: false}
		if decoded, configured := rule.OptionsAs[RequireAtomicUpdatesOptions](options); configured {
			settings = decoded
		}

		return rule.Listeners{
			ast.KindSourceFile: func(sourceFile *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				analyzeNonAtomicUpdates(ctx, sourceFile, settings)
			},
		}
	},
}

// atomicEventKind separates the three things the control-flow walk records for this rule.
type atomicEventKind uint8

const (
	// atomicRead is a load of the binding's current value. It refreshes the binding, clearing any
	// outdated mark an earlier suspension left on it.
	atomicRead atomicEventKind = iota
	// atomicSuspend is an await or a yield. Everything freshly read up to here becomes outdated.
	atomicSuspend
	// atomicWrite is a store this rule may judge: the assignment's own left-hand side, at the point
	// the assigned expression has finished evaluating.
	atomicWrite
	// atomicStoreAnchor marks the object of a member store whose own occurrence is not a read,
	// which is the plain `=` case. It carries no meaning to the lattice and exists only so
	// `placeDeferredEvents` has a position to hang the store on; it never survives that pass.
	atomicStoreAnchor
)

// atomicEvent is one point in the flow this rule cares about.
type atomicEvent struct {
	kind   atomicEventKind
	symbol *ast.Symbol
	// target is the identifier the event was recorded for. Only a write uses it, to name the
	// assignment it reports and to render the property text.
	target *ast.Node
	// assignment is the enclosing assignment expression a write belongs to, which is the node the
	// finding is reported on. Upstream reports on `node.parent`, the whole assignment, rather than
	// on either side of it.
	assignment *ast.Node
	// isProperty records whether the write stores into a property of the binding rather than into
	// the binding itself. It selects the message and is what `allowProperties` suppresses.
	isProperty bool
}

// analyzeNonAtomicUpdates reports every non-atomic assignment in one file.
func analyzeNonAtomicUpdates(ctx rule.Context, sourceFile *ast.Node, settings RequireAtomicUpdatesOptions) {
	for _, root := range atomicUpdateRoots(sourceFile) {
		analyzeRootNonAtomicUpdates(ctx, root, settings)
	}
}

// atomicUpdateRoots collects the code path roots that can suspend.
//
// A root that cannot suspend is skipped entirely rather than analyzed and found clean. That is
// upstream's `shouldVerify` gate, and it is what keeps this rule off the overwhelming majority of
// functions in a real tree: nothing before the first await can be outdated, so a function with no
// await and no yield has no work to do at all.
//
// # The gate is a cost filter, and that was measured rather than assumed
//
// A mutant widening this to every root SURVIVED the whole corpus, and the honest verdict is
// equivalence rather than a fixture gap. A root that cannot suspend emits no suspension, so nothing
// in it ever enters the outdated set and no write in it can be judged; admitting more such roots
// cannot change a verdict. The inverse mutation, closing the gate on everything, fails 24 lines,
// which proves the line is reached and that the survival is about the direction rather than about
// dead code. Naming a distinguishing input requires a non-suspending function containing a
// suspension, which is a contradiction.
//
// Upstream's own `shouldVerify` is the same shape and the same kind of filter. What the gate buys is
// that the overwhelming majority of functions in a real tree are never walked at all.
//
// The source file itself is never a root here even though top-level await exists in a module,
// because upstream gates on `scope.type === "function"` and a top-level await has no caller that
// could interleave in the way this rule reasons about. Recorded rather than assumed: upstream's
// gate is a function-scope test, so a module's top level is outside it by construction.
func atomicUpdateRoots(sourceFile *ast.Node) []*ast.Node {
	var roots []*ast.Node

	var visit func(node *ast.Node) bool
	visit = func(node *ast.Node) bool {
		if node == nil {
			return false
		}
		if control_flow_graph.IsRoot(node) && canSuspend(node) {
			roots = append(roots, node)
		}
		node.ForEachChild(visit)
		return false
	}
	sourceFile.ForEachChild(visit)

	return roots
}

// canSuspend reports whether a code path root is an async function or a generator.
//
// Both halves matter and upstream tests both: `scope.block.async || scope.block.generator`. An async
// generator is both and is judged once, which the corpus exercises directly at
// `let foo; async function* x() { foo = (yield foo) + await bar; }`.
func canSuspend(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindGetAccessor, ast.KindSetAccessor:
	default:
		return false
	}
	if node.ModifierFlags()&ast.ModifierFlagsAsync != 0 {
		return true
	}
	// An arrow function cannot be a generator, so only the function-like kinds carry an asterisk.
	// `AsteriskToken` is on the concrete node rather than on `ast.Node`, so each kind is asked
	// separately rather than through one accessor.
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().AsteriskToken != nil
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().AsteriskToken != nil
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().AsteriskToken != nil
	}
	return false
}

// analyzeRootNonAtomicUpdates runs the forward outdated-read dataflow over one suspending root.
func analyzeRootNonAtomicUpdates(ctx rule.Context, root *ast.Node, settings RequireAtomicUpdatesOptions) {
	// escapes is memoized per root rather than per candidate, because the escape scan walks the
	// whole root looking for cross-scope references and a root with several writes to one binding
	// would otherwise pay for it once per write.
	escapes := map[*ast.Symbol]bool{}

	graph := control_flow_graph.Build(root, control_flow_graph.Hooks[atomicEvent]{
		Read: func(builder *control_flow_graph.Builder[atomicEvent], node *ast.Node) {
			// The Read hook fires for a plain assignment target too, which `no-useless-assignment`
			// measured and documented: `patternReads` reads a `KindIdentifier` target before the
			// assigned expression. So what the occurrence IS comes from the AST, and the hook only
			// says where in the flow it sits.
			//
			// That distinction is load-bearing here and in the same direction as upstream's. Reading
			// upstream: a reference is marked fresh when `reference.isRead()` AND it is not the
			// target of a plain `=`. A compound target (`foo += x`) IS a read and does refresh; a
			// plain target (`foo = x`) is not. Reproducing that means asking the AST which kind of
			// assignment target this is.
			// A declarator's own name is a declaration rather than a reference, and upstream's
			// reference map holds only references. The graph emits a Read for it because
			// `patternReads` runs over a binding name the same way it runs over an assignment
			// target, so the AST has to say which it is.
			//
			// Measured: without this, `async function f() { let records; records = await a.records;
			// g(() => { records }); }` reports, and upstream ships it as a clean case. The bare
			// `let records` was being counted as a read, made stale by the await, and the plain `=`
			// store then judged against it.
			if isDeclarationName(node) {
				return
			}

			plainTarget := isPlainAssignmentTarget(node)
			symbol := resolveOccurrence(ctx, node)
			if symbol == nil {
				return
			}
			if !plainTarget {
				builder.Emit(atomicEvent{kind: atomicRead, symbol: symbol, target: node})
				return
			}
			// A plain `=` target refreshes nothing, but a member store still has to be judged, and
			// the store is anchored on this occurrence. So an anchor is emitted where the read would
			// have been, carrying no refresh of its own. `placeDeferredEvents` turns it into a write
			// and drops it.
			//
			// Without this, `function* g(foo) { baz = foo.bar; yield x; foo.bar = 1; }` goes silent:
			// the object is read in an earlier statement and made stale by the yield, and the store
			// in the later statement is a plain `=` whose own occurrence was the only thing the
			// store could hang on. Four corpus cases are that shape.
			if _, isProperty := propertyAssignmentHeadedBy(node); isProperty {
				builder.Emit(atomicEvent{kind: atomicStoreAnchor, symbol: symbol, target: node})
			}
		},
		Expression: func(builder *control_flow_graph.Builder[atomicEvent], node *ast.Node) {
			// A suspension is recorded as a marker HERE, before its operand, and moved to its real
			// position by `placeDeferredEvents`. The Expression hook is the only one the graph
			// offers for an await or a yield, and it is pre-order, which is the wrong end.
			//
			// The distinction is not cosmetic and the corpus decides it in both directions.
			// `foo = await foo` REPORTS upstream, because `makeOutdated` runs at the await's exit
			// and so the read inside its own operand is made stale by it. `foo = await bar + foo`
			// is CLEAN, because the read to the right of the `+` happens after the whole await.
			// Recording the suspension before the operand gets the first wrong; recording it at the
			// operand's end gets both right.
			switch node.Kind {
			case ast.KindAwaitExpression, ast.KindYieldExpression:
				builder.Emit(atomicEvent{kind: atomicSuspend, target: node})
			}
		},
		Write: func(builder *control_flow_graph.Builder[atomicEvent], node *ast.Node) {
			recordAtomicWrite(builder, ctx, node, escapes, root)
		},
	})

	// Two of the three event kinds are recorded at the wrong end of their construct, because the
	// graph offers a pre-order Expression hook and no post-order one while both judgments belong
	// after the construct has finished evaluating. Both are placed by source position here, and the
	// property arm is collected here too, since a member store emits no Write event at all. The
	// reasoning for each is at `placeDeferredEvents`.
	placeDeferredEvents(ctx, graph, escapes, root)

	solution := control_flow_graph.Solve[atomicReadState, atomicEvent](
		graph,
		control_flow_graph.Forward,
		outdatedReadLattice{},
	)

	// One assignment can be judged more than once, and reporting each judgment would double the
	// finding.
	//
	// The graph lays a `finally` block out TWICE, once for normal completion and once for the path
	// leaving the `try` through return, throw, or a suspended yield. Both copies carry the same
	// source positions, so a write inside a `finally` is walked twice and both copies can find it
	// stale. Measured on the shape that surfaced it, which came off the real tree rather than out of
	// the corpus:
	//
	//	function o() { let g = false; async function f() {
	//	    if (g) return; g = true; try { await s(); } finally { g = false; } } }
	//
	// The installed rule reports `g = false` ONCE, at column 106. Before this, cohere reported the
	// identical span twice. No imported case could see it: upstream's corpus writes no assignment
	// inside a `finally` at all, and its one try/catch case puts the writes in the arms.
	//
	// Deduplicating by the assignment's own position is the right key rather than by block, because
	// the copies ARE one write in the source. `no-useless-assignment` hit the same duplicate layout
	// and answered a different question about it, needing every copy to agree before reporting; here
	// any copy finding it stale is a real race, so the first wins and the rest are the same finding.
	reported := map[int]bool{}

	for _, block := range graph.Blocks {
		if !block.Reachable {
			continue
		}
		incoming, ok := solution.In(block)
		if !ok {
			continue
		}
		state := incoming.clone()
		applyAtomicTransfer(block, state, func(event atomicEvent) {
			if event.isProperty && settings.AllowProperties {
				return
			}
			if event.assignment == nil || reported[event.assignment.Pos()] {
				return
			}
			reported[event.assignment.Pos()] = true
			reportNonAtomicUpdate(ctx, event)
		})
	}
}

// atomicReadState is the lattice value: the two sets upstream keeps per code path segment.
//
// # Why both sets travel, when only one is consulted at a report site
//
// `isOutdated` reads only the outdated set, so a first version of this rule kept the fresh set
// block-local and treated it as bookkeeping inside one block. That is wrong wherever a read and the
// suspension that stales it land in DIFFERENT blocks, which is what any forked right-hand side
// produces. Measured on upstream's own cases: `foo = foo + (bar ? baz : await amount)` lays the read
// out in the entry block and the await in one arm of the fork, so at the await the block-local fresh
// set is empty and nothing is ever marked outdated. Five corpus findings were silent that way, and
// the rule's own doc comment had asserted the collapse was safe, which is exactly the "documented a
// limit I did not measure" failure.
//
// Upstream carries both sets forward in `SegmentInfo#initialize`, unioning each from every previous
// segment. This is the same thing said as a lattice.
type atomicReadState struct {
	// fresh holds bindings read since the last suspension on this path. A suspension moves them all
	// into outdated.
	fresh map[*ast.Symbol]bool
	// outdated holds bindings whose last read was before a suspension and which no later read has
	// refreshed. A write to one of these is what this rule reports.
	outdated map[*ast.Symbol]bool
}

func (state atomicReadState) clone() atomicReadState {
	cloned := atomicReadState{
		fresh:    make(map[*ast.Symbol]bool, len(state.fresh)),
		outdated: make(map[*ast.Symbol]bool, len(state.outdated)),
	}
	for symbol := range state.fresh {
		cloned.fresh[symbol] = true
	}
	for symbol := range state.outdated {
		cloned.outdated[symbol] = true
	}
	return cloned
}

// outdatedReadLattice is the forward analysis over atomicReadState.
//
// # Union is the meet, and it is upstream's own
//
// `SegmentInfo#initialize` folds every previous segment's two sets together with
// `Set.prototype.add`, and `isOutdated` returns true if ANY current segment holds the variable. Both
// are disjunction, so a binding outdated on one incoming path is outdated at the join. That is the
// conservative direction for a race warning: a write reachable by a path where the read was stale is
// a write that can be stale.
type outdatedReadLattice struct{}

// Bottom is the identity for union, so both sets are empty.
func (outdatedReadLattice) Bottom() atomicReadState {
	return atomicReadState{fresh: map[*ast.Symbol]bool{}, outdated: map[*ast.Symbol]bool{}}
}

// Entry says nothing has been read and nothing is stale where the function begins. It coincides
// numerically with Bottom and means a different thing, which is why it is written separately.
func (outdatedReadLattice) Entry() atomicReadState {
	return atomicReadState{fresh: map[*ast.Symbol]bool{}, outdated: map[*ast.Symbol]bool{}}
}

func (outdatedReadLattice) Meet(left, right atomicReadState) atomicReadState {
	// Neither argument is mutated: Solve holds one as a settled boundary value and compares against
	// it to decide whether anything moved.
	merged := atomicReadState{
		fresh:    make(map[*ast.Symbol]bool, len(left.fresh)+len(right.fresh)),
		outdated: make(map[*ast.Symbol]bool, len(left.outdated)+len(right.outdated)),
	}
	for symbol := range left.fresh {
		merged.fresh[symbol] = true
	}
	for symbol := range right.fresh {
		merged.fresh[symbol] = true
	}
	for symbol := range left.outdated {
		merged.outdated[symbol] = true
	}
	for symbol := range right.outdated {
		merged.outdated[symbol] = true
	}
	return merged
}

func (outdatedReadLattice) Transfer(
	block *control_flow_graph.Block[atomicEvent],
	incoming atomicReadState,
) atomicReadState {
	state := incoming.clone()
	applyAtomicTransfer(block, state, nil)
	return state
}

func (outdatedReadLattice) Equal(left, right atomicReadState) bool {
	return sameAtomicSymbolSet(left.fresh, right.fresh) &&
		sameAtomicSymbolSet(left.outdated, right.outdated)
}

func sameAtomicSymbolSet(left, right map[*ast.Symbol]bool) bool {
	if len(left) != len(right) {
		return false
	}
	for symbol := range left {
		if !right[symbol] {
			return false
		}
	}
	return true
}

// applyAtomicTransfer walks one block's events in order, updating the state in place and handing
// each judged write to onWrite.
//
// The same function serves the lattice's Transfer and the reporting pass, so the two cannot drift: a
// reporting walk that disagreed with the transfer would report against a state the fixed point never
// produced. `no-useless-assignment` uses the same arrangement for the same reason.
func applyAtomicTransfer(
	block *control_flow_graph.Block[atomicEvent],
	state atomicReadState,
	onWrite func(atomicEvent),
) {
	for _, event := range block.Events {
		switch event.kind {
		case atomicRead:
			// A read refreshes the binding: whatever it held before the suspension is no longer what
			// this expression is built from. Upstream's `markAsRead` does both halves, adding to
			// fresh and deleting from outdated, and the delete is the half that decides
			// `foo = await bar + foo`.
			state.fresh[event.symbol] = true
			delete(state.outdated, event.symbol)

		case atomicSuspend:
			// Everything read since the last suspension is now stale. Upstream's `makeOutdated`.
			for symbol := range state.fresh {
				state.outdated[symbol] = true
				delete(state.fresh, symbol)
			}

		case atomicWrite:
			if onWrite != nil && state.outdated[event.symbol] {
				onWrite(event)
			}
		}
	}
}

// reportNonAtomicUpdate emits one finding on the assignment expression.
//
// The span is the whole assignment rather than either side of it, matching upstream's
// `context.report({ node: node.parent })` where `node` is the assigned expression. Both message arms
// report on the same node; only the text differs.
func reportNonAtomicUpdate(ctx rule.Context, event atomicEvent) {
	name := event.symbol.Name
	if event.isProperty {
		ctx.ReportNode(event.assignment, rule.Message{
			Id: messageRequireAtomicUpdatesProperty.Id,
			Description: fmt.Sprintf(
				"%s Here the assignment writes %s and the object %s was read before the suspension.",
				messageRequireAtomicUpdatesProperty.Description,
				assignmentTargetText(ctx, event.assignment),
				name),
		})
		return
	}
	ctx.ReportNode(event.assignment, rule.Message{
		Id: messageRequireAtomicUpdatesVariable.Id,
		Description: fmt.Sprintf(
			"%s Here the variable is %s.",
			messageRequireAtomicUpdatesVariable.Description, name),
	})
}

// assignmentTargetText renders the left-hand side of an assignment as it appears in the source.
//
// Upstream renders the same thing with `sourceCode.getText(node.parent.left)`, which is the raw
// slice rather than a reconstruction, so `foo[bar].baz` and `foo.#bar` come out exactly as written.
// Taking the slice rather than rebuilding it from the tree is what keeps those two right: a
// reconstruction has to decide how to print a computed member and a private name, and upstream never
// makes that decision.
func assignmentTargetText(ctx rule.Context, assignment *ast.Node) string {
	if assignment == nil || ctx.SourceFile == nil {
		return ""
	}
	left := assignmentLeftHandSide(assignment)
	if left == nil {
		return ""
	}
	text := ctx.SourceFile.Text()
	start, end := rule.TokenRange(ctx.SourceFile, left).Pos(), left.End()
	if start < 0 || end > len(text) || start >= end {
		return ""
	}
	return text[start:end]
}

// assignmentLeftHandSide returns the target of an assignment expression.
func assignmentLeftHandSide(assignment *ast.Node) *ast.Node {
	if assignment == nil || assignment.Kind != ast.KindBinaryExpression {
		return nil
	}
	return assignment.AsBinaryExpression().Left
}

// recordAtomicWrite files a store into the binding itself, if this rule may judge it.
//
// The write is emitted at the point the graph reached it, which is after the assigned expression has
// been evaluated. That is what upstream achieves by deferring verification to the `:expression:exit`
// of the assigned expression rather than checking at the identifier.
func recordAtomicWrite(
	builder *control_flow_graph.Builder[atomicEvent],
	ctx rule.Context,
	node *ast.Node,
	escapes map[*ast.Symbol]bool,
	root *ast.Node,
) {
	assignment := enclosingAssignmentFor(node)
	if assignment == nil {
		return
	}
	symbol := resolveOccurrence(ctx, node)
	if symbol == nil {
		return
	}
	// The variable arm asks whether the BINDING escapes, so isMemberAccess is false here. The
	// property arm asks the same predicate with it true, which is what makes a parameter answer
	// differently on the two sides. See the escape table in the rule's doc comment.
	if !escapesEnclosingFunction(ctx, symbol, root, false, escapes) {
		return
	}
	builder.Emit(atomicEvent{
		kind:       atomicWrite,
		symbol:     symbol,
		target:     node,
		assignment: assignment,
	})
}

// enclosingAssignmentFor returns the assignment expression an identifier is the target of, or nil.
//
// # The left-side test is subsumed by this function's only caller
//
// A mutant disabling `binary.Left != identifier` survived the whole corpus, and the cause is the
// caller rather than the fixtures. `recordAtomicWrite` is the single call site and it runs only from
// the graph's Write hook, which fires where a binding is STORED into, so an identifier on the right
// of an assignment never arrives here. Measured on four shapes chosen to separate them, including
// `let foo; let bar; async function x() { foo; await q; bar = foo; }` where `foo` is read on the
// right after a suspension: mutant and pristine produce identical findings on all four.
//
// The test is kept because the verdict names one caller, and a second caller handing this an
// arbitrary identifier voids it. That is the enumeration the argument rests on, so it is written
// down rather than left implicit.
//
// Upstream excludes a variable declarator explicitly, with the comment "exclude variable
// declarations", by requiring `writeExpr.parent.right === writeExpr`. A declarator's initializer is
// not the right operand of an assignment expression, so requiring a `KindBinaryExpression` parent
// with an assignment operator reproduces that exclusion by construction rather than by a separate
// test. `let foo = await bar` cannot be a race: the binding did not exist before the suspension.
func enclosingAssignmentFor(identifier *ast.Node) *ast.Node {
	parent := identifier.Parent
	if parent == nil || parent.Kind != ast.KindBinaryExpression {
		return nil
	}
	binary := parent.AsBinaryExpression()
	if binary.Left != identifier {
		return nil
	}
	if !isAssignmentOperatorToken(binary.OperatorToken) {
		return nil
	}
	return parent
}

// isAssignmentOperatorToken reports whether a binary operator stores into its left operand.
func isAssignmentOperatorToken(operator *ast.Node) bool {
	if operator == nil {
		return false
	}
	switch operator.Kind {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken, ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken, ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken, ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken, ast.KindBarEqualsToken, ast.KindCaretEqualsToken,
		ast.KindBarBarEqualsToken, ast.KindAmpersandAmpersandEqualsToken,
		ast.KindQuestionQuestionEqualsToken:
		return true
	}
	return false
}

// isPlainAssignmentTarget reports whether an identifier heads the target of a plain `=`, and so is
// a write that does not also refresh the binding.
//
// Upstream's condition is `reference.isRead() && !(writeExpr && writeExpr.parent.operator === "=")`,
// where `getWriteExpr` walks up through enclosing member expressions. So the test is about the whole
// member CHAIN the identifier heads, not only about the identifier being the target itself, and both
// halves of that are pinned by the corpus:
//
//	foo = await bar             clean    a plain `=` target is not a read
//	foo += await bar            reports  a compound target IS a read, and goes stale
//	foo.bar = await baz         clean    the object of a plain `=` target is likewise not a read
//	foo.bar += await baz        reports  the object of a compound target is
//
// The third row is the one that separates this from the naive version. Reading only the identifier's
// own parent makes `foo.bar = await baz` report, and upstream ships it as a clean case. It is clean
// for the same reason `foo = await bar` is: nothing has been read yet at the point the store is set
// up, so there is no stale value for the store to be built from.
func isPlainAssignmentTarget(identifier *ast.Node) bool {
	node := identifier
	for node != nil {
		parent := node.Parent
		if parent == nil {
			return false
		}
		switch parent.Kind {
		case ast.KindPropertyAccessExpression:
			if parent.AsPropertyAccessExpression().Expression != node {
				return false
			}
			node = parent

		case ast.KindElementAccessExpression:
			if parent.AsElementAccessExpression().Expression != node {
				return false
			}
			node = parent

		case ast.KindBinaryExpression:
			binary := parent.AsBinaryExpression()
			if binary.Left != node {
				return false
			}
			return binary.OperatorToken != nil && binary.OperatorToken.Kind == ast.KindEqualsToken

		default:
			return false
		}
	}
	return false
}

// isDeclarationName reports whether an identifier is the name a declaration introduces rather than a
// reference to one.
//
// A binding's own declarator name is not a reference and upstream's `createReferenceMap` never holds
// it. Our graph emits a Read for it because `patternReads` treats a binding name and an assignment
// target alike, so the distinction has to come from the AST.
func isDeclarationName(identifier *ast.Node) bool {
	parent := identifier.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration, ast.KindBindingElement, ast.KindParameter:
		return parent.Name() == identifier
	}
	return false
}
