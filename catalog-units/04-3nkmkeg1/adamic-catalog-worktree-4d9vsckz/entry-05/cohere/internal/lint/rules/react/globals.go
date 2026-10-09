package react

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/ecmascript/reference"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// messageGlobalReassignmentId is the finding's id, and messageGlobalReassignmentReason is
// everything after the sentence naming the binding.
//
// Split rather than held as one `rule.Message`, because this rule interpolates: the binding's name
// carries most of what a reader needs and it has to come first. The reason is the invariant half,
// so it lives here where a single edit moves every finding, and `TestGlobalsMessageNamesTheBinding`
// asserts the rendered result rather than either half.
const messageGlobalReassignmentId = "globalReassignment"

const messageGlobalReassignmentReason = "Render has to be able to run at any time and any number " +
	"of times, so a write that escapes it makes the result depend on how often the component " +
	"happened to render. Under Strict Mode and concurrent rendering that count is not something " +
	"the code controls, so the value drifts and the bug shows up somewhere else entirely. If the " +
	"value is used in rendering, hold it in useState. If it is not, write it from an effect or an " +
	"event handler, where running once is guaranteed."

// Globals flags a write, during render, to a variable declared outside the component or hook.
//
//	valid:   function Component() { let x = 1; x = 2; return <div />; }
//	valid:   function helper() { someGlobal = true; }                       (not a component)
//	valid:   function Component() { someGlobal.x = 1; return <div />; }     (property, not binding)
//	valid:   function Component() { g++; return <div />; }                  (update, see below)
//	valid:   function Component() { useEffect(() => { g = 1; }, []); ... }  (not during render)
//	invalid: let g = 0; function Component() { g = 1; return <div />; }
//	invalid: function Component() { undeclaredName = 1; return <div />; }
//	invalid: import {g} from 'm'; function Component() { g = 1; return <div />; }
//	invalid: function Component(props) { [a, b] = props.value; return <div />; }
//
// Ported from React's `globals` rule, which is `ErrorCategory.Globals` in the React Compiler. Both
// authorities were readable and both were run rather than reasoned about:
//
//	react 7.1.1   node_modules/eslint-plugin-react-hooks/cjs/eslint-plugin-react-hooks.development.js
//	              driven through the ESLint Linter API on 150-plus probe inputs
//	oxc           crates/oxc_linter/src/rules/react/globals.rs, a category filter over
//	              crates/oxc_react_compiler, plus its 1-pass 1-fail corpus and snapshot
//
// The two agree on every input measured here, including the corpus, the message text, and the span.
// No divergence between them was found and none is recorded.
//
// # What the rule is actually about, and why "global" is the wrong word for it
//
// The name says globals and the diagnostic says "declared outside of the component/hook", and the
// diagnostic is the accurate one. Measured against React's own rule, every one of these reports:
//
//	someUnknownGlobal = true;      no declaration anywhere in the program
//	let g = 0;   g = 1;            an ordinary module-scope binding
//	const g = 0; g = 1;            likewise, and it is also a const violation
//	import {g} from 'm'; g = 1;    an import binding
//	function g() {}  g = 1;        a function declaration
//	class g {}       g = 1;        a class declaration
//	Object = 1;                    a standard-library binding
//
// So a port keyed on "the name does not resolve" would have found only the first of those seven,
// which is one of the nine distinct cases in React's own corpus. Resolution is not the question;
// **where the resolved declaration lives** is. That was the single most load-bearing measurement in
// this port and it inverted the approach the dispatch proposed.
//
// The complement matters as much. A write to a binding the compiled function itself declares is
// silent, and so is a write to a binding an *enclosing* function inside the same compilation root
// declares:
//
//	function Component() { let x = 1; x = 2; ... }                       silent
//	function Component() { let x = 1; const f = () => { x = 2; }; f(); } silent
//
// The second is a closure capture, and it is the case the intermediate representation in
// `internal/utilities/hir` gets wrong. See the note on routes below.
//
// # Why this does not read the HIR, which lowers `StoreGlobal` directly
//
// `internal/utilities/hir` landed the same day as this rule and emits a `StoreGlobal` instruction,
// which reads as the whole detection already done. It was not used, and the reasons were measured
// rather than inferred. Both were re-probed at `19ac208`, the commit that taught lowering to resolve
// captures, because the first probe predated it and a stale measurement is worse than none.
//
// **A `StoreGlobal` still cannot separate an import from a true global.** Probed at current head,
// `import {g} from 'mod'; g = 2;` and a bare undeclared `someGlobal = 2;` lower to byte-identical
// instructions, and `GlobalBindingKind` is declared with several variants that lowering never sets
// to anything but `Global` (`lower.go:57` states this). For most consumers that is a small gap. For
// this rule it is the entire question, because the diagnostic is about **where a binding is
// declared** and an import and an unresolved name sit on opposite ends of that. Both happen to
// report here, so it would not have produced a wrong answer today, but the rule would have been
// resting on a coincidence rather than on the distinction it claims to make.
//
// **The capture gap is half closed, and the surviving half is the half this rule would meet.**
// Before `19ac208` a write to an enclosing function's local lowered to `StoreGlobal`, where React
// classifies it as a `StoreContext` and is silent. That is now fixed when the enclosing component is
// the lowering root. It is NOT fixed when the nested function is lowered as its own root, which is
// how a per-function rule reaches it. Measured at current head on one input, printing both:
//
//	function Component() { let x = 1; const f = () => { x = 2; }; f(); return null; }
//	  lowering the COMPONENT  ->  StoreContext reassign x$2 = $1     correct, and silent
//	  lowering the ARROW      ->  StoreGlobal x = $1                 still the old answer
//
// `high_level_intermediate_representation.Lower` takes one function and builds its own identifier table, and
// `lowerAssignmentTarget` emits `StoreGlobal` for any target not in that table, so which answer a
// consumer gets depends on which node it handed the lowering. A rule built on this would have had
// to be careful about its own traversal in a way nothing in the instruction set signals.
//
// The checker has neither gap: it resolves to a declaration, and the declaration's file and position
// are readable directly, so an import, a module binding, a standard-library binding, a capture and a
// local are all separable by the same test. So the route here is the checker. None of this is a
// criticism of that package, which is young and whose own header lists both gaps; it is a statement
// of why this particular rule could not consume it yet, and a later version plausibly can.
//
// # The third route, and why it was not taken either
//
// `internal/utilities/ecmascript/reference` already answers "does this identifier write to its
// binding", and `core/no-global-assign` already answers a question one word away from this one.
// Neither is the whole rule, but the structural half is genuinely the same question, and it is
// reused here rather than rewritten. Its doc comment records the 625-false-positive incident where
// `document.cookie = x` was read as a write to a binding named `cookie`; that guard is inside
// `WritesToBinding` and this rule inherits it. Writing a fourth implementation of the write test
// was the thing the shelf exists to prevent.
//
// # Which writes count, measured shape by shape
//
// Every result below was produced by running React's own rule on the input, holding everything but
// the operator fixed, with a control assignment in the same function proving the function was not
// simply bailing out. The reporting set:
//
//	g = 1                            =
//	g += -= *= /= %= **=             the arithmetic compounds
//	g <<= >>= >>>= &= |= ^=          the bitwise compounds
//	[g] = xs   ({g} = o)             destructuring assignment, every nesting
//	[...g] = xs   ({...g} = o)       including a rest element
//	[g = 5] = []                     including a default
//
// And the silent set, which is the surprising half:
//
//	g++  g--  ++g  --g               update expressions
//	g ||= 1   g &&= 1   g ??= 1      the logical compounds
//	g.x = 1   g[0] = 1               a write through a property of the global
//	for (g of xs) {}                 the head of a for-of or for-in
//
// The update expressions and the logical compounds are the ones a port will get wrong, because
// every other assignment form reports and these read as assignment forms. They are silent because
// lowering emits no `StoreGlobal` for them at all: the update path in `build_hir.rs` constructs a
// `PostfixUpdate`/`PrefixUpdate` whose lvalue is a place rather than a global store, and the
// logical compounds lower as branching structures. Both were measured with a control in the same
// function, so neither is a bailout. `error.update-global-should-bailout` in React's own corpus
// uses `+=` rather than `++` for exactly this reason.
//
// The `for (g of xs)` case is different and worth stating separately: it does not merely fail to
// report, it takes the **whole function** silent, control included. That is a lowering failure
// rather than a decision, so it is reproduced here as silence on the write without reproducing the
// whole-function bailout. Recorded as a stated divergence rather than smoothed over: a file with an
// undeclared for-of loop variable and a separate global write reports here and not upstream.
//
// A property write is not this rule at all. `wat.test = 1` on a module-scope `wat` reports under
// React's **immutability** rule with a different message ("This value cannot be modified"), which
// is why `error.store-property-in-global` and `error.mutate-property-from-global` are named for
// globals and carry no `Globals` diagnostic. Two fixtures that read as this rule's and are not.
//
// # Where the finding points, which is two different spans in one rule
//
// A simple assignment points at the **identifier**; a compound assignment points at the **whole
// assignment expression**. Measured, holding the name long enough to tell the two apart:
//
//	longGlobalName = 1     ^^^^^^^^^^^^^^          the name
//	longGlobalName += 1    ^^^^^^^^^^^^^^^^^^^     the name, the operator, and the right side
//
// That is not an accident of formatting. Lowering has two sites: the simple path passes
// `Some(ident_span)` (`build_hir.rs:4408`) and the compound path passes `span`, the assignment's
// own (`build_hir.rs:4534`). oxc's snapshot and React's goldens agree on both. A message-id
// assertion cannot see either, so both are asserted by span in the fixtures.
//
// A destructuring target points at the element rather than the statement, and a defaulted element
// points at the whole binding element (`[g = 5]` reports `g = 5`, not `g`). A parenthesized simple
// target reports the identifier **inside** the parens while a parenthesized compound target reports
// from the open paren, which follows from the same two spans and is pinned by a fixture.
//
// # The gate, which decides more inputs than the write test does
//
// None of this fires outside a function React Compiler compiles. The gate is shared with
// `unsupported-syntax` in this package, where it was measured over seventeen probe rounds against
// React's own rule, and every one of its properties was re-confirmed here on this rule's own
// diagnostic before it was reused:
//
//	function helper() { g = 1; }                          silent, name is not a component or hook
//	function Component() { g = 1; }                       silent, no JSX and no hook call
//	function Component() { g = 1; return <div />; }        reports
//	function Component(...props) { g = 1; return <div />; }        silent, rest parameter
//	function Component(a, b) { g = 1; return <div />; }            silent, second is not a ref
//	function Component(a, ref) { g = 1; return <div />; }          reports
//	function useFoo() { useState(0); g = 1; }                      reports, hook
//
// Reusing it rather than restating it is deliberate: two copies of a gate this intricate would
// drift, and the drift would be invisible because each copy would pass its own fixtures.
//
// # What this port does NOT reproduce, stated as a subset rather than left to be discovered
//
// A write inside a **nested function** is reported by React only when its effects pass can prove
// that function runs during render, and that proof is an inter-procedural aliasing analysis rather
// than anything syntactic. Measured, and the boundary is not a rule anybody could state:
//
//	const f = () => { g = 1; }; f();                       reports
//	const f = () => { g = 1; };                            silent, never called
//	const f = () => { g = 1; }; return <Foo cb={f} />;     silent, passed as a prop
//	const f = () => { g = 1; }; return <Foo>{f}</Foo>;     reports, passed as children
//	const a = () => { const b = () => { g = 1; }; b(); }; a();   silent, nested inline
//	const b = () => { g = 1; }; const a = () => { b(); }; a();   reports
//	function a() { g = 1; } a();                           silent when the call precedes the declaration
//
// The last three are the tell: the same call depth gives opposite answers depending on how the
// functions are spelled and ordered. That is the effects pass, and it is Tier C work. So this rule
// reports **only writes lexically inside the compilation root's own body**, not inside any nested
// function, and that is a stated subset: it misses findings React produces and adds none React does
// not. `TestGlobalsBoundary` pins the subset so it stays a decision rather than becoming a drift.
//
// Three further exclusions inside the root's own body, all measured with controls in the same
// function so that none of them is a bailout being misread:
//
//	try {} catch (e) { g = 1; }      silent, a catch body is not on the render path
//	try {} finally { g = 1; }        silent, likewise
//	return <div />; g = 1;           silent, unreachable after the return
//
// The unreachable case is reproduced by declining a write in a statement that follows a `return`,
// `throw`, `break` or `continue` in the same statement list. That is narrower than real
// reachability and deliberately so: it is the shape upstream's own corpus and this rule's probes
// exercise, and a full reachability answer belongs in the control-flow graph rather than here.
//
// No fix. The repair is a `useState` or an effect, which is a change of design rather than of text,
// and upstream proposes none either.
var Globals = rule.Rule{
	// No namespace prefix. The config writes `react/globals`, and the parity guard strips the
	// namespace on a `/` boundary, so `react-globals` would match no inventory entry while still
	// passing every fixture in this package. Measured in both spellings; see the test.
	Name: "react-hooks/globals",

	// The whole rule is "where is this name declared relative to the enclosing component", and
	// nothing structural answers it. `g = 1` is the same three characters whether `g` is a module
	// binding, a parameter, a local, an import, or nothing at all, and only resolution separates
	// them. This is the name-resolution half of the port brief's table rather than the scope-flag
	// half, established by probe: `GetSymbolAtLocation` answers a module-scope `KindVariableDeclaration`
	// for the reporting case and a `KindParameter` inside the component for the silent one.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindIdentifier: func(node *ast.Node) {
				// The engine hands a rule a nil checker when the program could not be built. This
				// rule can answer nothing without one, and the failure mode without the guard is
				// silence rather than a panic, because `GetSymbolAtLocation` tolerates a nil
				// receiver. Silence is the more dangerous of the two, so the guard is explicit and
				// `TestGlobalsRequiresTheTypedHarness` fails loudly if it is ever removed.
				if ctx.TypeChecker == nil {
					return
				}

				// The structural half first, because it is far cheaper than a checker call and this
				// listener sees every identifier in the file, most of them reads. This also carries
				// the property-name guard that kept `no-global-assign` from reporting every
				// `document.cookie = x` in the tree; see that package's note.
				if !reference.WritesToBinding(node) {
					return
				}
				// `WritesToBinding` is wider than this rule: it answers true for `g++` and for the
				// head of a `for (g of xs)`, both of which upstream lowers to something other than
				// a global store and both of which are silent. Measured with controls; see the rule
				// note. This is the narrowing, and it runs before the checker call.
				if !isGlobalStoreTarget(node) {
					return
				}

				root := enclosingCompilationRoot(node)
				if root == nil {
					return
				}
				// Only the root's own body, never a nested function. The subset is stated on the
				// rule and pinned by `TestGlobalsBoundary`.
				if crossesNestedFunctionBoundary(node, root) {
					return
				}
				if isInsideCatchOrFinally(node, root) {
					return
				}
				if followsAnAbruptCompletion(node, root) {
					return
				}

				if !declaredOutsideCompilationRoot(ctx, node, root) {
					return
				}

				ctx.ReportNode(globalStoreReportNode(node), rule.Message{
					Id: messageGlobalReassignmentId,
					Description: fmt.Sprintf(
						"This writes to `%s`, which is declared outside the component or hook, "+
							"while the component is rendering. %s",
						node.Text(), messageGlobalReassignmentReason),
				})
			},
		}
	},
}

// isGlobalStoreTarget reports whether an identifier's write is one upstream lowers to a global
// store, as opposed to one it lowers to something else and never reports.
//
// `reference.WritesToBinding` answers the general question "does this occurrence assign", and the
// answer is correct for every shape it covers. This rule needs the narrower question, because
// upstream reports a strict subset of the writes that exist, and the boundary is not intuitive.
//
// Every membership below was measured on React's own rule with a control assignment in the same
// function, so a silence here is a silence about the write rather than the function bailing out.
// See the table on the rule for the full result set.
func isGlobalStoreTarget(identifier *ast.Node) bool {
	parent := utilsreact.SkipParenthesesUpward(identifier)
	if parent == nil {
		return false
	}

	switch parent.Kind {
	case ast.KindBinaryExpression:
		// There is deliberately no test here that the identifier is on the LEFT of the operator,
		// and its absence is a measurement rather than an oversight. It was written first, reading
		// as obviously load-bearing, and a mutant disabling it entirely SURVIVED the whole suite.
		//
		// Probed rather than patched with another fixture. Over ten shapes putting an identifier on
		// the right of an assignment (`h = g`, `h += g`, `(h) = g`, `h ||= g`, `h = g = 1`,
		// `h = -g`, a call argument, an array element, a parenthesized right side), the guard's
		// answer and `reference.WritesToBinding`'s answer were the same everywhere: every one of
		// them is already declined by the write test one guard above, because a read position is
		// not a write access. Nothing can reach this point on the right of an operator.
		//
		// So it was a subsumed branch rather than a fixture gap, and the port brief's standing
		// advice is to delete a subsumed branch and record the reasoning at the line rather than
		// write a test that asserts nothing. The right-side shapes are still fixtures here, now
		// pinned by the guard that actually decides them.
		return isGlobalStoreOperator(parent.AsBinaryExpression().OperatorToken.Kind)

	case ast.KindPrefixUnaryExpression, ast.KindPostfixUnaryExpression:
		// `g++`, `++g`, `g--`, `--g`. Silent upstream, measured with a control. Lowering builds a
		// `PostfixUpdate`/`PrefixUpdate` whose lvalue is a place rather than a global store, so no
		// `StoreGlobal` is emitted and the effects pass has nothing to attach a diagnostic to.
		return false

	case ast.KindForOfStatement, ast.KindForInStatement:
		// `for (g of xs) {}` with no declaration in the head. Upstream takes the WHOLE FUNCTION
		// silent on this input, control included, which is a lowering failure rather than a
		// decision about the write. Reproduced here as silence on the write alone; the divergence
		// is stated on the rule.
		return false
	}

	// A destructuring assignment target reaches here through the literal that spells it, because a
	// destructuring target parses as an array or object literal rather than as a pattern. Every one
	// of those reports upstream, including a rest element and a defaulted element, so the walk out
	// to the assignment is the test: if `WritesToBinding` said yes and the shape is not one of the
	// non-reporting kinds above, it is a store.
	return true
}

// isGlobalStoreOperator reports whether an assignment operator lowers to a global store.
//
// Plain assignment and the arithmetic and bitwise compounds do. The three logical compounds
// (`||=`, `&&=`, `??=`) do NOT, and that is the surprising half: they are assignment operators, they
// sit in the same `KindBinaryExpression` shape, and they are silent upstream. Measured on React's
// own rule with a control assignment in the same function, so it is not a bailout. They lower as
// branching structures rather than as a store, which is why nothing is emitted for them.
//
// Written as an allow list rather than as "every assignment operator except three". A deny list
// silently admits any operator nobody thought of, and admitting one wrongly makes this rule report
// on source upstream passes, which is the expensive direction.
func isGlobalStoreOperator(operator ast.Kind) bool {
	switch operator {
	case ast.KindEqualsToken,
		ast.KindPlusEqualsToken,
		ast.KindMinusEqualsToken,
		ast.KindAsteriskEqualsToken,
		ast.KindAsteriskAsteriskEqualsToken,
		ast.KindSlashEqualsToken,
		ast.KindPercentEqualsToken,
		ast.KindLessThanLessThanEqualsToken,
		ast.KindGreaterThanGreaterThanEqualsToken,
		ast.KindGreaterThanGreaterThanGreaterThanEqualsToken,
		ast.KindAmpersandEqualsToken,
		ast.KindBarEqualsToken,
		ast.KindCaretEqualsToken:
		return true
	}
	return false
}

// globalStoreReportNode returns the node upstream reports for a write.
//
// Two answers in one rule, and both are upstream's. A simple assignment reports the identifier; a
// compound assignment reports the whole assignment expression. The two lowering sites pass
// different spans (`build_hir.rs:4408` against `:4534`), React's ESLint output and oxc's snapshot
// agree on both, and a message-id assertion cannot see either, so both are asserted by span.
//
// A destructuring element reports the element itself, which is the identifier except when it
// carries a default, where the whole binding element is reported. That is the same rule as the
// compound case arriving through a different shape: the node is whichever one holds both the target
// and the value.
//
// A NODE rather than a range, so the report goes through `ctx.ReportNode` and picks up its leading
// trivia strip. Written as `ReportRange` first, and every span fixture failed by exactly the
// indentation: `Loc.Pos()` sits before leading trivia, so a statement at the start of a line reports
// from the previous line's newline. `ReportRange` is documented as the helper for a span that is not
// a node, and both spans here are nodes, so the node helper is the correct one and the trivia
// question is already solved there.
func globalStoreReportNode(identifier *ast.Node) *ast.Node {
	// A defaulted destructuring element reports the whole element: `[g = 5] = []` reports `g = 5`.
	// It is checked first because it wears the same `KindBinaryExpression` shape as a plain
	// assignment and would otherwise be answered by the branch below.
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindBinaryExpression &&
		identifier.Parent.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken &&
		identifier.Parent.AsBinaryExpression().Left == identifier &&
		isInsideDestructuringTarget(identifier.Parent) {
		return identifier.Parent
	}

	if parent := utilsreact.SkipParenthesesUpward(identifier); parent != nil &&
		parent.Kind == ast.KindBinaryExpression &&
		parent.AsBinaryExpression().OperatorToken.Kind != ast.KindEqualsToken {
		// A compound assignment reports the whole expression, parentheses on the target included:
		// `(g) += 1` reports from the open paren. Measured against React's own output.
		return parent
	}

	// A plain assignment reports the identifier alone, inside any parentheses on the target:
	// `(g) = 1` reports `g` rather than `(g)`. Measured.
	return identifier
}

// isInsideDestructuringTarget reports whether a node sits inside a destructuring assignment target
// rather than being, or containing, the target of a plain assignment.
//
// The distinction matters only for the defaulted element, where `[g = 5] = []` has the same
// `KindBinaryExpression` shape as a plain `g = 5` and reports the whole element rather than the
// name. The separator is the array or object literal that spells the destructuring target: a
// destructuring target parses as a literal rather than as a pattern, so reaching one on the way out
// means this equals sign is a default rather than an assignment.
func isInsideDestructuringTarget(node *ast.Node) bool {
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindArrayLiteralExpression, ast.KindObjectLiteralExpression:
			return true
		case ast.KindBinaryExpression, ast.KindSpreadElement, ast.KindSpreadAssignment,
			ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment,
			ast.KindParenthesizedExpression:
			continue
		default:
			return false
		}
	}
	return false
}

// enclosingCompilationRoot returns the function React Compiler would compile that contains a node,
// or nil when the node sits outside every one.
//
// The gate itself is the react shelf's `IsInsideComponentOrHook`, written for `unsupported-syntax`
// and measured there over seventeen probe rounds. This returns the root node rather than a boolean,
// because this rule needs to ask two further questions about the node's position *relative to* the
// root: whether a nested function sits between them, and whether the declaration it writes to lives
// inside it.
//
// The innermost qualifying ancestor wins, which matches the gate's own walk.
func enclosingCompilationRoot(node *ast.Node) *ast.Node {
	for current := node.Parent; current != nil; current = current.Parent {
		if utilsreact.IsComponentOrHookLike(current) && utilsreact.IsReachableRootPosition(current) {
			return current
		}
	}
	return nil
}

// crossesNestedFunctionBoundary reports whether a function sits between a node and its compilation
// root.
//
// This is the stated subset. Upstream reports a nested function's write when its effects pass can
// prove the function runs during render, and that proof is inter-procedural aliasing rather than
// anything readable from position. The measurements on the rule show the boundary giving opposite
// answers for the same call depth depending on spelling and ordering, so there is no narrower
// syntactic approximation worth shipping. Declining every nested function loses findings and adds
// none, which is the right direction to be short in.
func crossesNestedFunctionBoundary(node *ast.Node, root *ast.Node) bool {
	for current := node.Parent; current != nil && current != root; current = current.Parent {
		if ast.IsFunctionLike(current) {
			return true
		}
	}
	return false
}

// isInsideCatchOrFinally reports whether a node sits in a catch clause or a finally block of a try
// statement inside the compilation root.
//
// Both are silent upstream and both were measured with a control assignment in the same function,
// so neither is the function bailing out: `try { g = 1; } catch (e) {}` reports while
// `try {} catch (e) { g = 1; }` and `try {} finally { g = 1; }` do not, in files whose control
// assignment fires either way.
//
// The try BLOCK is deliberately not excluded, for the same reason.
func isInsideCatchOrFinally(node *ast.Node, root *ast.Node) bool {
	for current := node; current != nil && current != root; current = current.Parent {
		parent := current.Parent
		if parent == nil {
			return false
		}
		if parent.Kind == ast.KindCatchClause {
			return true
		}
		if parent.Kind == ast.KindTryStatement && parent.AsTryStatement().FinallyBlock == current {
			return true
		}
	}
	return false
}

// followsAnAbruptCompletion reports whether a node sits in a statement that can never be reached
// because an earlier statement in the same list ends control flow.
//
// `return <div />; g = 1;` is silent upstream, measured with a control. This reproduces the shape
// the corpus and the probes exercise rather than real reachability: only a sibling statement in the
// same list is considered, and only the four completions that end a block. A full answer belongs in
// `internal/lint/ecmascript/control_flow_graph` rather than here, and this rule does not need one, since unreachable
// code after a return is the only form either upstream's corpus or these probes produce.
func followsAnAbruptCompletion(node *ast.Node, root *ast.Node) bool {
	for current := node; current != nil && current != root; current = current.Parent {
		parent := current.Parent
		if parent == nil {
			return false
		}
		statements, hasStatements := statementListOf(parent)
		if !hasStatements {
			continue
		}
		for _, statement := range statements {
			if statement == current {
				break
			}
			switch statement.Kind {
			case ast.KindReturnStatement, ast.KindThrowStatement,
				ast.KindBreakStatement, ast.KindContinueStatement:
				return true
			}
		}
	}
	return false
}

// statementListOf returns a node's statement list when it has one.
//
// The port brief's standing hazard is that every `Node.Xxx()` accessor panics off its kind, and
// `Node.Statements()` is named as one that panics off Block, SourceFile and ModuleBlock. The kind is
// therefore established by the switch before any typed accessor is reached.
func statementListOf(node *ast.Node) ([]*ast.Node, bool) {
	switch node.Kind {
	case ast.KindBlock:
		return node.AsBlock().Statements.Nodes, true
	case ast.KindCaseClause:
		return node.AsCaseOrDefaultClause().Statements.Nodes, true
	case ast.KindDefaultClause:
		return node.AsCaseOrDefaultClause().Statements.Nodes, true
	}
	return nil, false
}

// declaredOutsideCompilationRoot reports whether the binding an identifier writes to is declared
// outside the function React Compiler is compiling.
//
// This is the rule's actual judgment, and it reproduces `resolve_identifier` at
// `oxc_react_compiler/.../hir_builder.rs:760-810`, which classifies a binding as global or module
// local when it has no symbol, when it is a type-only declaration, when its scope is the program
// scope, or when its scope is not within the compiled function.
//
// # Why every declaration is examined rather than the first, and the survivor that says so
//
// The port brief's standing rule is never to index `symbol.Declarations[0]`, and this rule is the
// shape where a loop is right rather than a single choice: it asks "is this binding declared
// outside", so a symbol carrying several declarations is outside only when they all are. Measured,
// `Object` carries two declarations, an interface and a variable, both in `lib.es5.d.ts`; a merged
// `interface Foo` above a `var Foo` in source carries two in the source file.
//
// **A mutant replacing this loop with `symbol.Declarations[0]` SURVIVES the whole suite, and so does
// one taking the last declaration.** That was chased rather than accepted, because the brief calls
// equivalence the comfortable verdict and the one most likely to be wrong.
//
// The two versions can differ on exactly one kind of input: a symbol whose declarations straddle the
// compilation root, some inside and some outside. Fifteen shapes were probed looking for one, each
// printing every declaration's kind and its side of the boundary: a module `var` plus an inner
// `var`, a module `function` plus an inner `function`, a module `interface` plus an inner
// `function`, a module `type`/`namespace`/`enum` each plus an inner `var`, an `interface` and a
// `var` merged at module scope, an `interface` and a `var` merged inside the component, an ambient
// `declare var` plus a local, and `Object`.
//
// **Not one straddles**, and the reason is structural rather than a property of the sample.
// Declaration merging happens within a single scope; an inner declaration of the same name does not
// merge with an outer one, it SHADOWS it, and the checker mints a separate symbol. Every probe with
// an inner declaration came back `decls=1` pointing inside, and every probe with only outer ones
// came back with all of them outside. So no input can put declarations of one symbol on both sides,
// and the mutants are equivalent rather than unseen.
//
// The loop is kept anyway, and that is a deliberate choice rather than an oversight of the sweep.
// It is not slower in any measurable way, it states the question the rule is actually asking, and
// the equivalence rests on a checker behavior this rule does not own. `M14b`, which flips the loop's
// verdict rather than its extent, is caught by 53 subtests, so the loop body is genuinely reached
// and the fixtures see through it; only first-versus-any is unobservable.
//
// An unresolved name is outside by definition, and it is the case the rule's own name suggests and
// the smallest part of what it catches.
func declaredOutsideCompilationRoot(ctx rule.Context, identifier *ast.Node, root *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)

	// A shorthand property in a destructuring target resolves through the plain accessor to the
	// PROPERTY's own symbol, which is declared in source inside the component and therefore reads
	// exactly like a correctly declined local. `({g} = props)` is a case React reports, so the
	// plain accessor would drop it silently. TypeScript has a separate accessor for the value side,
	// and the same recovery is shipped in `core/no-global-assign` for the same shape.
	//
	// The substitution is UNCONDITIONAL, and that is the correction this rule needed. The sibling
	// rule keeps the value symbol only when it is non-nil, which is right for a rule asking "is this
	// a standard-library binding", where nil means "no evidence" and declining is correct. Here nil
	// means the opposite: the shorthand target resolves to NO binding at all, which is precisely the
	// reporting case, and `({a} = props)` on an undeclared `a` is one React reports. Written
	// conditionally first, and the fixture for that exact input was the only one of forty-two that
	// failed, because falling back to the plain symbol hands this function the property's own
	// declaration sitting inside the component and it correctly answers "inside".
	if identifier.Parent != nil && identifier.Parent.Kind == ast.KindShorthandPropertyAssignment {
		symbol = ctx.TypeChecker.GetShorthandAssignmentValueSymbol(identifier.Parent)
	}

	// No binding anywhere in the program. `someUnknownGlobal = true` is upstream's own first
	// corpus case and it lands here.
	if symbol == nil || len(symbol.Declarations) == 0 {
		return true
	}

	for _, declaration := range symbol.Declarations {
		if declarationIsInsideRoot(declaration, root) {
			return false
		}
	}
	return true
}

// declarationIsInsideRoot reports whether a declaration node sits lexically inside a compilation
// root.
//
// Walking the declaration's ancestors to look for the root is the whole test, and it answers all
// four of upstream's categories at once. A module-scope binding, an import, a standard-library
// declaration and an unresolved name all fail to find the root; a local, a parameter, and a capture
// from an enclosing function *inside* the root all find it.
//
// That last one is the case the intermediate representation gets wrong and the reason this rule
// reads the checker instead. See the note on the rule.
func declarationIsInsideRoot(declaration *ast.Node, root *ast.Node) bool {
	if declaration == nil {
		return false
	}
	for current := declaration; current != nil; current = current.Parent {
		if current == root {
			return true
		}
	}
	return false
}
