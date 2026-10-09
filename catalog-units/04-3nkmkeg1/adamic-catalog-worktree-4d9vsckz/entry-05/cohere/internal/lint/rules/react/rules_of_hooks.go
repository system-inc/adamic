package react

import (
	"strings"
	"unicode"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/control_flow_graph"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var (
	messageRulesOfHooksConditional = rule.Message{
		Id: "rulesOfHooksConditional",
		Description: "This Hook is not called on every render path, so React loses track of which " +
			"state belongs to which call. React identifies a Hook by its position in the call " +
			"order, not by its name, so a render that skips one shifts every Hook after it onto " +
			"the wrong slot. Move the call above the condition and branch on its result instead.",
	}
	messageRulesOfHooksLoop = rule.Message{
		Id: "rulesOfHooksLoop",
		Description: "This Hook can run more than once in a single render, which changes how many " +
			"Hooks the component calls from one render to the next. React matches Hooks up by " +
			"call order across renders, so a varying count desynchronizes all of them. Call it " +
			"once and put the repetition inside the value it returns.",
	}
	messageRulesOfHooksTopLevel = rule.Message{
		Id: "rulesOfHooksTopLevel",
		Description: "A Hook called at module scope has no component rendering it, so there is no " +
			"state for it to attach to and it runs once at import rather than once per render. " +
			"Move the call into a function component or a custom Hook.",
	}
	messageRulesOfHooksClassComponent = rule.Message{
		Id: "rulesOfHooksClassComponent",
		Description: "A class component does not render through the Hook dispatcher, so a Hook " +
			"called from one of its methods has nowhere to store state and throws at runtime. " +
			"Use the class's own lifecycle and instance state, or rewrite the component as a " +
			"function.",
	}
	messageRulesOfHooksCallback = rule.Message{
		Id: "rulesOfHooksCallback",
		Description: "This Hook is called inside a callback rather than during render, so it runs " +
			"at whatever moment the callback fires instead of in the render pass React is " +
			"tracking. Call the Hook in the component body and let the callback read what it " +
			"returned.",
	}
	messageRulesOfHooksNotComponent = rule.Message{
		Id: "rulesOfHooksNotComponent",
		Description: "This Hook is called from a plain function, which React never renders, so " +
			"there is no component instance to hold the state. React decides what is a component " +
			"or a custom Hook purely from the name: a component starts with a capital letter and " +
			"a custom Hook starts with `use`. Rename the function, or move the call to a caller " +
			"that is one.",
	}
	messageRulesOfHooksAsync = rule.Message{
		Id: "rulesOfHooksAsync",
		Description: "An async function body resumes after its awaits, outside the render pass, so " +
			"a Hook called there runs when no component is rendering. Keep the component " +
			"synchronous and do the asynchronous work inside an effect.",
	}
	messageRulesOfHooksTryCatchUse = rule.Message{
		Id: "rulesOfHooksTryCatchUse",
		Description: "`use` suspends by throwing a promise, which a surrounding `try` swallows " +
			"before React can see it, so the component never suspends and never resumes. Move " +
			"the call outside the `try`, and handle the failure with an error boundary.",
	}
	messageRulesOfHooksEffectEventEscape = rule.Message{
		Id: "rulesOfHooksEffectEventEscape",
		Description: "`useEffectEvent` returns a function that is only valid during the effect that " +
			"owns it, so passing the call's result straight into something else hands out a " +
			"reference that outlives it. Assign it to a local first and call that from an effect.",
	}
)

// RulesOfHooks enforces React's Rules of Hooks: a Hook runs on every render, in the same order, in
// something React actually renders.
//
//	valid:   function Component() { useHook(); }
//	valid:   function Component() { if (useHook()) { a(); } }        (in the test, so it always runs)
//	valid:   function useCustomHook() { useState(); }
//	valid:   function Component() { return useThing() && <A />; }    (left operand always evaluates)
//	valid:   const C = React.memo(props => { useHook(); });
//	invalid: function Component() { if (cond) { useHook(); } }
//	invalid: function Component() { while (cond) { useHook(); } }
//	invalid: function notAComponent() { useHook(); }
//	invalid: useHook();                                              (module scope)
//	invalid: class C { m() { useHook(); } }
//	invalid: async function Component() { useHook(); }
//	invalid: function Component() { useEffect(() => { useHook(); }); }
//
// # Which implementation this reproduces, and why the answer is not obvious here
//
// Two live implementations exist and they are **not** the same algorithm, which is unusual for this
// catalog. oxc's `rules_of_hooks.rs` is not a transliteration of ESLint's rule; it is an independent
// reimplementation reaching the same verdicts by different machinery, and on three shapes it reaches
// **different** verdicts. This port reproduces oxc, because oxc is what the differential gate runs,
// and every divergence is named below with the command that established it.
//
// Both were read in full. ESLint's is the original and lives in this tree as a compiled but
// unminified bundle at `node_modules/eslint-plugin-react-hooks/cjs/eslint-plugin-react-hooks.development.js`
// (version 7.1.1, the rule at lines 54935-55400), and the original TypeScript was fetched from
// `facebook/react`, `packages/eslint-plugin-react-hooks/src/rules/RulesOfHooks.ts`. Both are
// executable here, so every claim in this comment is a measurement rather than a reading: ESLint's
// Linter API runs from `~/Projects/ahra`, and the release oxlint binary is at
// `~/Projects/system/oxc/target/release/oxlint`.
//
// # The three measured divergences between oxc HEAD and ESLint 7.1.1
//
// The dispatch that commissioned this port expected oxc to be pinned at 2024 and therefore behind.
// It is not: `git log` on `rules_of_hooks.rs` shows its last change on 2026-08-14, ten days before
// this was written, and the commit is titled "report hooks inside try blocks". So oxc is the newer
// behaviour on the axis that matters most, and the drift runs the opposite way from what was
// expected.
//
//	try { const v = 1; useHook(v); } catch {}
//	    oxc     reports "called conditionally"
//	    eslint  clean
//
// oxc treats **any** Hook inside a `try` block as conditional, syntactically, without asking whether
// the `try` can throw. The case above cannot throw at all and still reports. ESLint asks its code
// path whether the segment is on every path, and a non-throwing `try` is, so it stays clean. This is
// oxc's newer rule and it is reproduced here: `isInsideTryStatement` below is a syntactic ancestor
// walk, deliberately, and the control-flow graph is not consulted for it.
//
//	function Component() { if (a) { return 1; } else { return 2; } useHook(); }
//	    oxc     silent
//	    eslint  reports "called conditionally"
//
// The opposite direction. oxc declines a Hook in code nothing reaches; ESLint's segment is
// unreachable but its `reactHooksMap` still holds it, and the path arithmetic reports. Reproduced as
// oxc: an unreachable Hook is not a Hook that runs, so this rule stays quiet and `no-unreachable`
// owns the finding. The same applies after a bare `throw`.
//
//	function notAComponent() { useÉtat(); }
//	    oxc     reports
//	    eslint  clean
//
// oxc's `is_react_hook_name` tests the fourth character with `char::is_uppercase`, which is Unicode.
// ESLint's is `/^use[A-Z0-9]/`, which is not. Reproduced as oxc, which means this rule's hook
// predicate is Unicode-uppercase-or-ASCII-digit, and that is why it cannot use
// `internal/utilities/react.IsHookName` — see below.
//
// One more difference is cosmetic rather than behavioural and is recorded so nobody re-derives it:
// for `Namespace.useThing()` oxc renders the hook name as `useThing` and ESLint renders it as
// `Namespace.useThing`. Our messages carry no interpolated name at all, so nothing here depends on
// it.
//
// # The shelf helper that is wrong for this rule, measured in three directions
//
// `internal/utilities/react.IsHookName` is exactly the shape this reaches for and would be wrong on
// every case in this rule's corpus that discriminates. Its body requires a name longer than `use`
// and tests `unicode.IsUpper` alone. Against both oracles, run on `Hook.use()`, `Hook.use42()` and
// `Hook.useHook()` from upstream's own case 133:
//
//	use          shelf says no; oxc and eslint both report it as a Hook
//	use2FAMutation, use42, use3DEngine
//	             shelf says no; oxc and eslint both report them as Hooks
//	useÉtat      shelf says yes; eslint says no, oxc says yes
//
// So the shelf disagrees with **both** implementations on the first two rows, and the corpus
// exercises them directly: `use2FAMutation` and `use3DEngine` are upstream pass cases whose whole
// point is that they are recognized as Hooks. `isHookIdentifierName` below is this rule's own
// predicate for that reason, and the divergence is not a defect in the shelf function so much as
// three sources with three answers, of which the shelf matches our own TypeScript house layer.
// Adjusting the shelf would break its thirteen other callers, so this rule owns its own.
//
// # Reached differently than ESLint, because we have a program and it does not
//
// ESLint's rule is built on `eslint-scope`. It records every `useEffectEvent` binding by walking
// scope references, and it resolves function names by inspecting the enclosing declarator through
// scope analysis. We have a resident type checker and a whole-program graph and could answer both by
// resolution. This port answers neither by resolution, and that is a fidelity decision rather than a
// capability one: oxc's `get_declaration_identifier` reads the **syntactic** parent — a variable
// declarator, an assignment target, an assignment pattern, an object property key — and never
// resolves anything. Resolving would report on shapes upstream is silent for. The rule therefore
// declares no type checker, and the checker is not consulted anywhere.
//
// # The control-flow graph, and the one place it does not answer the question
//
// This is the first rule in the catalog to consume `internal/lint/ecmascript/control_flow_graph`, which is a
// basic-block graph vendored from rslint whose block layout deliberately mirrors ESLint's own code
// path analysis. Whether that mirroring actually holds for the shapes this rule turns on had never
// been tested, so it was probed against both oracles before anything was built on it. Sixteen
// shapes, one Hook each, comparing `PathAnalysis.IsOnEveryFinalPath` and `IsCyclic` against what the
// two binaries print:
//
//	straight line, if-test, ternary-test, `&&`-left        onEveryFinal=true    both silent
//	if-body, `&&`-right, ternary-branch, switch-case       onEveryFinal=false   both conditional
//	statement after an early return                        onEveryFinal=false   both conditional
//	while body, for body, while-in-try                     cyclic=true          both loop
//
// Twelve of sixteen match exactly, which is a real result: `IsOnEveryFinalPath` is precisely
// ESLint's `countPathsFromStart * countPathsToEnd !== allPathsFromStartToEnd` comparison computed as
// dominance instead of as path arithmetic, and it agrees on every shape where the two can be
// compared. The four that do not are the three try-shapes above, which are a genuine oxc-versus-
// ESLint disagreement rather than a graph defect, and one graph shape that needs stating:
//
//	do { useHook(); } while (cond);
//	    our graph  cyclic=FALSE, onEveryFinal=TRUE
//	    oxc        reports "may be executed more than once"
//	    eslint     reports the same
//
// A `do…while` body genuinely runs on the way to the test, and the back edge re-enters the **test**
// rather than the body header, so no cycle passes through the body block on the first iteration and
// the dominance answer is correct about the graph while being useless for the rule. **ESLint hits
// this too and solves it exactly the way this rule does**, with a syntactic `isInsideDoWhileLoop`
// check sitting beside the code-path answer rather than derived from it (bundle line 54943). That is
// not a workaround for a defect in our graph; it is the same shape upstream, independently. The
// dispatch for this port predicted the `do…while` divergence would arrive through `labelsOf` keeping
// outer labels on a `continue`; the divergence is real and its cause is this instead.
//
// # What the corpus said that reading the sources did not
//
// Two things. `function useUnreachable() { return; useHook(); }` is an upstream **pass** case, which
// is the only statement anywhere that a Hook nothing reaches is not this rule's problem; both
// sources' code reads as though it would report. And case 133 enumerates the hook-name predicate
// exhaustively against a namespace: of `Hook.use()`, `Hook._use()`, `Hook.useState()`,
// `Hook._useState()`, `Hook.use42()`, `Hook.useHook()`, `Hook.use_hook()`, exactly four report. No
// prose in either implementation says that; the corpus does.
//
// # Scope
//
// The `useEffectEvent` **reference-escape** family is deliberately not ported. Upstream tracks every
// reference to a binding initialized by `useEffectEvent(...)` and reports each one that is not
// lexically inside an effect call, which needs a symbol-reference index this rule would otherwise
// have no use for. It is 12 of the 124 upstream diagnostics and a distinct feature rather than a
// missing branch of this one. The **inline**-escape arm is ported, because it needs only the call's
// own parent. Stated rather than silently missing: a `useEffectEvent` result assigned to a variable
// and then passed to a child is a finding this port gives up.
var RulesOfHooks = rule.Rule{
	Name: "react-hooks/rules-of-hooks",
	Run:  runRulesOfHooks,
}

// RulesOfHooksOptions is upstream's schema for this rule, which declares one option that changes
// nothing.
//
// The schema accepts `{additionalHooks: string}`, and the rule never reads it. Upstream's `create`
// reads only the shared `settings["react-hooks"].additionalEffectHooks`, which cohere does not carry,
// and never `context.options`; at 7.1.1 the option is accepted and inert, and React's own corpus has
// no row for it. Accepting means implementing, and implementing this option is accepting it: a config
// copied from ESLint loads, and the findings are what ESLint's are (#d21war2).
type RulesOfHooksOptions struct {
	// AdditionalHooks is accepted and, as upstream, read by nothing.
	AdditionalHooks string `json:"additionalHooks"`
}

func runRulesOfHooks(ctx rule.Context, options any) rule.Listeners {
	// One graph per code path root per file, built lazily and only for roots that actually contain a
	// Hook call. Most functions contain none, and building a graph for them would make this rule pay
	// per function rather than per Hook.
	graphs := map[*ast.Node]*hookPathFacts{}

	return rule.Listeners{
		ast.KindCallExpression: func(node *ast.Node) {
			call := node.AsCallExpression()
			if call == nil {
				return
			}
			callee := call.Expression
			if !isHookCallee(callee) {
				return
			}

			isUse := isReactFunctionCall(callee, "use")

			enclosing := enclosingFunctionForHook(node)
			if enclosing == nil {
				ctx.ReportNode(node, messageRulesOfHooksTopLevel)
				return
			}

			if isReactFunctionCall(callee, "useEffectEvent") {
				reportInlineEffectEventEscape(ctx, node)
			}

			// A method, a static block, or a field initializer means the enclosing function is a
			// class member, and a class never renders through the Hook dispatcher.
			if isClassMemberFunction(enclosing) {
				ctx.ReportNode(node, messageRulesOfHooksClassComponent)
				return
			}

			if reportEnclosingFunctionShape(ctx, node, callee, enclosing, isUse) {
				return
			}

			// `use(...)` is allowed to be conditional and allowed in a loop, so the whole path
			// analysis below is skipped for it. What is not allowed is a `try`, because `use`
			// suspends by throwing.
			if isUse {
				if isInsideTryStatement(node, enclosing) {
					ctx.ReportNode(node, messageRulesOfHooksTryCatchUse)
				}
				return
			}

			facts := factsForRoot(graphs, enclosing)
			if facts == nil {
				return
			}
			block, found := facts.blockOf(node)
			if !found {
				return
			}

			// `try` is answered syntactically and it is answered FIRST, ahead of every graph
			// question including reachability. That ordering is upstream's and it is observable
			// rather than incidental: `try { throw err; useState(); } catch {}` is upstream case 112
			// and it REPORTS, while the same unreachable Hook with no `try` around it is upstream
			// case 18 and is SILENT. Measured on the release binary both ways after this rule got it
			// backwards on its first run. A cyclic Hook inside a `try` is still a loop rather than a
			// condition, which is why the cycle test sits inside this arm as well.
			if isInsideTryStatement(node, enclosing) {
				if facts.analysis.IsCyclic(block) || isInsideDoWhileStatement(node, enclosing) {
					ctx.ReportNode(node, messageRulesOfHooksLoop)
					return
				}
				ctx.ReportNode(node, messageRulesOfHooksConditional)
				return
			}

			// A Hook nothing reaches is not a Hook that runs. Upstream declines it, and upstream's
			// own corpus pins the decision with a pass case whose body is `return; useHook();`.
			// `no-unreachable` owns that finding; reporting it here would say the wrong thing about
			// why the line is wrong.
			if !block.Reachable {
				return
			}

			if facts.analysis.IsCyclic(block) || isInsideDoWhileStatement(node, enclosing) {
				ctx.ReportNode(node, messageRulesOfHooksLoop)
				return
			}

			if !facts.analysis.IsOnEveryFinalPath(block) {
				ctx.ReportNode(node, messageRulesOfHooksConditional)
			}
		},
	}
}

// reportEnclosingFunctionShape reports whatever is wrong with the function the Hook sits in, and
// says whether that answer is final.
//
// The arms are ordered the way oxc orders them, because the order is observable: a named function
// that is neither component nor Hook reports as `notComponent` even when it is also async, and a
// callback inside a component reports as `callback` even when its own name would qualify.
func reportEnclosingFunctionShape(ctx rule.Context, hook *ast.Node, callee *ast.Node, enclosing *ast.Node, isUse bool) bool {
	// A function with its own name that is neither component-shaped nor Hook-shaped. This arm points
	// at the callee rather than at the whole call, which is upstream's own choice and was measured
	// on the release binary rather than read: for `useHook(1)` the label covers seven bytes here and
	// ten in every other arm.
	if name := ownFunctionName(enclosing); name != "" {
		if !isComponentOrHookName(name) {
			ctx.ReportNode(callee, messageRulesOfHooksNotComponent)
			return true
		}
		if isAsyncFunction(enclosing) && isDirectlyInsideComponentOrHook(enclosing) {
			ctx.ReportNode(hook, messageRulesOfHooksAsync)
			return true
		}
		return false
	}

	// An anonymous function passed as an argument, as a `new` argument, or as a JSX expression
	// container is a callback. `memo` and `forwardRef` are the two callers whose callback React does
	// render, so they are not callbacks for this purpose.
	if isNonReactFunctionArgument(enclosing) {
		if !isUse && isSomewhereInsideComponentOrHook(enclosing) {
			ctx.ReportNode(hook, messageRulesOfHooksCallback)
		}
		return true
	}

	async := isAsyncFunction(enclosing)
	if async && isDirectlyInsideComponentOrHook(enclosing) {
		ctx.ReportNode(hook, messageRulesOfHooksAsync)
		return true
	}

	// An anonymous function whose surrounding syntax gives it a name — a declarator, an assignment,
	// an object property, a destructuring default. The name is read syntactically, never resolved.
	if name, named := declarationIdentifierOf(enclosing); named && !isComponentOrHookName(name) {
		ctx.ReportNode(callee, messageRulesOfHooksNotComponent)
		return true
	}

	return false
}

// reportInlineEffectEventEscape reports a `useEffectEvent(...)` result that is handed straight to
// something else instead of being bound to a local first.
//
// A declarator and a bare expression statement are the two shapes upstream permits.
func reportInlineEffectEventEscape(ctx rule.Context, call *ast.Node) {
	parent := call.Parent
	if parent == nil {
		return
	}
	switch parent.Kind {
	case ast.KindVariableDeclaration, ast.KindExpressionStatement:
		return
	}
	ctx.ReportNode(call, messageRulesOfHooksEffectEventEscape)
}

// hookPathFacts is one code path root's graph and the analysis over it, plus a map from a Hook call
// to the block it was laid out in.
//
// Keyed by the call node's position rather than by the node pointer, for the reason
// `internal/unused` found the hard way: the graph lays a `finally` block out twice and both copies
// carry the same positions, so the first **reachable** block a position lands in is the one that
// answers for it.
type hookPathFacts struct {
	analysis *control_flow_graph.PathAnalysis[hookEvent]
	blocks   map[int]*control_flow_graph.Block[hookEvent]
}

type hookEvent struct {
	position int
}

func (f *hookPathFacts) blockOf(call *ast.Node) (*control_flow_graph.Block[hookEvent], bool) {
	block, found := f.blocks[call.Pos()]
	return block, found
}

func factsForRoot(cache map[*ast.Node]*hookPathFacts, root *ast.Node) *hookPathFacts {
	if existing, cached := cache[root]; cached {
		return existing
	}

	blocks := map[int]*control_flow_graph.Block[hookEvent]{}
	graph := control_flow_graph.Build(root, control_flow_graph.Hooks[hookEvent]{
		Expression: func(builder *control_flow_graph.Builder[hookEvent], node *ast.Node) {
			if node == nil || node.Kind != ast.KindCallExpression {
				return
			}
			call := node.AsCallExpression()
			if call == nil || !isHookCallee(call.Expression) {
				return
			}
			block, _ := builder.Emit(hookEvent{position: node.Pos()})
			// Keyed by position rather than by node, and the LAST layout wins.
			//
			// The duplicate-layout trap is real and `internal/unused` was bitten by it: the graph
			// lays a `finally` body out twice, once for normal completion and once for the path that
			// leaves the `try` abruptly, and both copies carry the same source positions. That
			// package resolved it by preferring the reachable copy. This rule deliberately does not,
			// and the difference is worth stating because the obvious reading is that it should.
			//
			// Measured on this substrate. `try { f(); } finally { useHook(); }` lays the Hook out in
			// b4 (onEveryFinal=true) and then again in b1 (onEveryFinal=false), so which copy wins
			// decides the verdict — and neither answer is ever read, because a `finally` body is
			// syntactically inside a `TryStatement` and `isInsideTryStatement` answers above the
			// graph. Every shape that produces a duplicate layout is a `try` shape, so the
			// preference has no reachable effect at all.
			//
			// A preference was written here first and a mutation flipping it SURVIVED the whole
			// fixture set. The survivor was not a fixture gap; the branch was subsumed by an earlier
			// guard in the same function. Disabling the `try` check and re-running confirmed it: the
			// verdict is unchanged either way. So the preference is gone rather than kept with a
			// test that could not fail, and this comment is here so the next reader does not
			// helpfully add it back.
			blocks[node.Pos()] = block
		},
	})

	facts := &hookPathFacts{analysis: control_flow_graph.AnalyzePaths(graph), blocks: blocks}
	cache[root] = facts
	return facts
}

// isHookCallee reports whether a call's callee names a Hook, by React's own naming convention.
//
// Two spellings: a bare identifier, and a member access whose object is component-named. The member
// form is why `Hook.useState()` counts and `lowercase.useState()` does not.
func isHookCallee(callee *ast.Node) bool {
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return isHookIdentifierName(callee.Text())

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		object := access.Expression
		if object == nil || object.Kind != ast.KindIdentifier {
			return false
		}
		if !isComponentIdentifierName(object.Text()) {
			return false
		}
		name := access.Name()
		return name != nil && name.Kind == ast.KindIdentifier && isHookIdentifierName(name.Text())
	}
	return false
}

// isHookIdentifierName reports React's Hook naming shape: `use`, or `use` followed by an uppercase
// letter or an ASCII digit.
//
// Not `internal/utilities/react.IsHookName`, and the reason is at the rule's doc comment. Bare `use` and
// `use42` are Hooks to both upstream implementations and are not to the shelf function, and
// upstream's corpus turns on both.
func isHookIdentifierName(name string) bool {
	if !strings.HasPrefix(name, "use") {
		return false
	}
	rest := []rune(name[3:])
	if len(rest) == 0 {
		return true
	}
	return unicode.IsUpper(rest[0]) || (rest[0] >= '0' && rest[0] <= '9')
}

// isComponentIdentifierName reports a name starting with an ASCII capital.
//
// Not `internal/utilities/react.IsLikelyComponentName`, which tests `unicode.IsUpper`. oxc's
// `is_react_component_name` is `is_ascii_uppercase`, and the two disagree on `Éomponent`. The hook
// predicate above is deliberately the Unicode one and this one is deliberately not, because that is
// how oxc spells them: `utils/react.rs:761` is `is_uppercase`, `:785` is `is_ascii_uppercase`. One
// file, two answers, and copying either one to both places would be wrong.
func isComponentIdentifierName(name string) bool {
	if name == "" {
		return false
	}
	first := name[0]
	return first >= 'A' && first <= 'Z'
}

func isComponentOrHookName(name string) bool {
	return isComponentIdentifierName(name) || isHookIdentifierName(name)
}

// isReactFunctionCall reports whether a callee names a specific React function, bare or under the
// `React` namespace.
func isReactFunctionCall(callee *ast.Node, expected string) bool {
	if callee == nil {
		return false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text() == expected

	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		name := access.Name()
		if name == nil || name.Kind != ast.KindIdentifier || name.Text() != expected {
			return false
		}
		object := access.Expression
		return object != nil && object.Kind == ast.KindIdentifier && object.Text() == "React"
	}
	return false
}

// enclosingFunctionForHook returns the nearest function-like ancestor of a Hook call, or nil when the
// call is at module scope.
func enclosingFunctionForHook(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if ast.IsFunctionLike(current) {
			return current
		}
	}
	return nil
}

// isClassMemberFunction reports whether a function-like node is a class member's body.
//
// A constructor, an accessor and a static block only exist in a class, so their kind settles it. A
// **method does not**, and that is the trap: `({ useHook() { ... } })` parses to the same
// `KindMethodDeclaration` as a class method does, so a kind test alone reads an object literal as a
// class and reports the wrong arm. Upstream's own corpus pins both sides of it — case 19 writes
// `({useHook() { useState(); }})` as a passing case and case 101 writes `({g() { useState(); }})` as
// a `notComponent` failure — and this rule reported `classComponent` for both on its first run. The
// parent is what separates them.
//
// A field initializer is an ordinary function whose parent is a property declaration, and a class is
// the only place a property declaration appears.
func isClassMemberFunction(fn *ast.Node) bool {
	switch fn.Kind {
	case ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindClassStaticBlockDeclaration:
		return true
	case ast.KindMethodDeclaration:
		parent := fn.Parent
		return parent != nil && parent.Kind == ast.KindClassDeclaration ||
			parent != nil && parent.Kind == ast.KindClassExpression
	}
	parent := fn.Parent
	return parent != nil && parent.Kind == ast.KindPropertyDeclaration
}

func isAsyncFunction(fn *ast.Node) bool {
	return ast.HasSyntacticModifier(fn, ast.ModifierFlagsAsync)
}

// ownFunctionName returns a function's own declared name, which is a function declaration's
// identifier or a named function expression's identifier.
//
// Empty for an arrow and for an anonymous function expression, whose names come from the syntax
// around them instead.
func ownFunctionName(fn *ast.Node) string {
	switch fn.Kind {
	case ast.KindFunctionDeclaration, ast.KindFunctionExpression:
		name := fn.Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.Text()
		}
	}
	return ""
}

// declarationIdentifierOf reads the name an anonymous function borrows from the syntax around it.
//
// Four shapes, matching oxc's `get_declaration_identifier` exactly, and read syntactically rather
// than resolved. The second result separates "no name is available" from "a name is available and
// it is empty", which matters because the caller's verdict differs between them: an unnamed default
// export is silent and a lowercase-named arrow reports.
func declarationIdentifierOf(fn *ast.Node) (string, bool) {
	// An object-literal method is its OWN name-bearer rather than borrowing one from a parent: the
	// function-like node IS the `KindMethodDeclaration`, and there is no anonymous function nested
	// inside it whose parent could be consulted. Every other shape below reads the parent, which is
	// why this one has to be answered before the switch rather than inside it. Missed on the first
	// pass, and upstream's case 101 hid it — that case reports nine findings and this rule reported
	// nine, one of them from a different arm on a different line.
	if fn.Kind == ast.KindMethodDeclaration {
		if text, readable := ast.TryGetTextOfPropertyName(fn.Name()); readable {
			return text, true
		}
		return "", false
	}

	parent := fn.Parent
	if parent == nil {
		return "", false
	}

	switch parent.Kind {
	case ast.KindVariableDeclaration:
		name := parent.AsVariableDeclaration().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.Text(), true
		}

	case ast.KindBinaryExpression:
		binary := parent.AsBinaryExpression()
		if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindEqualsToken {
			return "", false
		}
		if binary.Right != fn {
			return "", false
		}
		left := binary.Left
		if left != nil && left.Kind == ast.KindIdentifier {
			return left.Text(), true
		}

	case ast.KindPropertyAssignment:
		name := parent.AsPropertyAssignment().Name()
		if text, readable := ast.TryGetTextOfPropertyName(name); readable {
			return text, true
		}

	// `KindShorthandPropertyAssignment` is deliberately absent, and it is the one arm here that
	// looks like an oversight.
	//
	// Our parser reaches it from exactly one shape: an assignment-destructuring default,
	// `({k = () => { useState(); }} = {})`. Upstream is SILENT on that, measured standalone on the
	// release binary against the binding-declaration spelling beside it:
	//
	//	const {j = () => { useState(); }} = {};    REPORTS
	//	({k = () => { useState(); }} = {});        SILENT
	//
	// The two are one line apart in upstream case 101 and only nine of its ten statements report,
	// which is the only place in the whole corpus that says so. oxc reaches that silence through a
	// node kind rather than a decision: a destructuring default in an ASSIGNMENT is an
	// `AssignmentTargetWithDefault` there, which its `get_declaration_identifier` does not name, so
	// no name is available and the caller declines. Our parser calls the same shape a shorthand
	// property assignment, and reading a name off it would report where upstream does not.
	//
	// So the absence is a reproduced verdict rather than a gap. Adding the arm passes every other
	// fixture in this file and breaks exactly one.
	case ast.KindBindingElement:
		name := parent.AsBindingElement().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.Text(), true
		}
	}

	return "", false
}

// isNonReactFunctionArgument reports whether a function is being handed to something as a callback,
// excluding the two callers React itself renders.
func isNonReactFunctionArgument(fn *ast.Node) bool {
	parent := fn.Parent
	if parent == nil {
		return false
	}
	switch parent.Kind {
	case ast.KindCallExpression:
		callee := parent.AsCallExpression().Expression
		return !isReactFunctionCall(callee, "forwardRef") && !isReactFunctionCall(callee, "memo")

	case ast.KindNewExpression, ast.KindJsxExpression:
		return true
	}
	return false
}

// isMemoOrForwardRefCallback reports whether any enclosing call is `memo` or `forwardRef`.
//
// Bare and namespaced spellings both count, and unlike isNonReactFunctionArgument this looks all the
// way up rather than at the immediate parent, matching oxc's `is_memo_or_forward_ref_callback`.
func isMemoOrForwardRefCallback(fn *ast.Node) bool {
	for current := fn.Parent; current != nil; current = current.Parent {
		if current.Kind != ast.KindCallExpression {
			continue
		}
		callee := current.AsCallExpression().Expression
		if calleeAccessedName(callee) == "forwardRef" || calleeAccessedName(callee) == "memo" {
			return true
		}
	}
	return false
}

// calleeAccessedName reads the final name of a callee, ignoring any namespace.
//
// oxc's `callee_name()` is namespace-blind here, which is why `Whatever.memo(...)` counts as a memo
// callback upstream. Reproduced rather than tightened.
func calleeAccessedName(callee *ast.Node) string {
	if callee == nil {
		return ""
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		return callee.Text()
	case ast.KindPropertyAccessExpression:
		name := callee.AsPropertyAccessExpression().Name()
		if name != nil && name.Kind == ast.KindIdentifier {
			return name.Text()
		}
	}
	return ""
}

// isDirectlyInsideComponentOrHook asks whether this one function is itself a component or a Hook.
func isDirectlyInsideComponentOrHook(fn *ast.Node) bool {
	if name := ownFunctionName(fn); name != "" {
		if isComponentOrHookName(name) {
			return true
		}
	} else if name, named := declarationIdentifierOf(fn); named && isComponentOrHookName(name) {
		return true
	}
	return isMemoOrForwardRefCallback(fn)
}

// isSomewhereInsideComponentOrHook asks whether this function or any function enclosing it is a
// component or a Hook.
//
// This is what separates a Hook in a callback inside a component, which reports, from a Hook in a
// callback inside a plain function, which is silent — upstream's judgment being that the second is
// somebody else's library code and not a React mistake.
func isSomewhereInsideComponentOrHook(fn *ast.Node) bool {
	for current := fn; current != nil; current = current.Parent {
		if !ast.IsFunctionLike(current) {
			continue
		}
		if isDirectlyInsideComponentOrHook(current) {
			return true
		}
	}
	return false
}

// isInsideTryStatement reports whether a Hook sits inside a `try` block or a `catch` clause, stopping
// at the enclosing function so a nested function's own `try` does not count.
//
// Syntactic on purpose. oxc reports a Hook in a `try` that provably cannot throw, so consulting the
// control-flow graph here would be silent on a case upstream reports.
func isInsideTryStatement(node *ast.Node, stop *ast.Node) bool {
	for current := node; current != nil && current != stop; current = current.Parent {
		switch current.Kind {
		case ast.KindTryStatement, ast.KindCatchClause:
			return true
		}
	}
	return false
}

// isInsideDoWhileStatement reports whether a Hook sits inside a `do…while`, stopping at the enclosing
// function.
//
// This exists because the graph cannot answer it, and ESLint has the identical check for the
// identical reason. A `do…while` body runs before its test, so the back edge enters the test rather
// than the body and no cycle passes through the body block. The body is genuinely on every path to
// the first iteration and genuinely repeats, and only the syntax says the second part.
func isInsideDoWhileStatement(node *ast.Node, stop *ast.Node) bool {
	for current := node; current != nil && current != stop; current = current.Parent {
		if current.Kind == ast.KindDoStatement {
			return true
		}
	}
	return false
}
