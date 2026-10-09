package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUseMemoCallbackReturnsNothing = rule.Message{
	Id: "useMemoCallbackReturnsNothing",
	Description: "This `useMemo` callback does not return a value, so there is nothing to " +
		"memoize. React calls the callback to compute a value and caches what comes back; a " +
		"callback that only performs side effects hands back `undefined`, which React then " +
		"caches faithfully and pointlessly. Worse, the side effect now runs on an unpredictable " +
		"schedule, since React is free to re-run or skip a memo computation whenever it likes. " +
		"Return the computed value, or move the side effect into an event handler or an effect, " +
		"which is where work that is not a computation belongs.",
}

var messageUseMemoResultUnused = rule.Message{
	Id: "useMemoResultUnused",
	Description: "This `useMemo` computes a value and then nothing reads it. Memoizing exists to " +
		"let a later render reuse a result, so a result no caller consumes buys nothing and " +
		"costs a cache entry plus a dependency comparison on every render. A call written for " +
		"its side effect rather than its value is the usual reason this appears, and the side " +
		"effect is the part that is unsafe: React may re-run or skip the computation at will. " +
		"Use the value, or move the work into an event handler or an effect.",
}

// VoidUseMemo flags a `useMemo` whose callback returns nothing, or whose result nothing consumes.
//
//	valid:   const x = useMemo(() => 1, []);
//	valid:   const x = useMemo(() => { return; }, []);            (a bare return counts)
//	valid:   const x = useMemo(() => { foo(); }, []); ... in a plain lowercase function
//	valid:   const cb = () => { foo(); }; const x = useMemo(cb, []);   (not inline)
//	valid:   const x = useMemo(() => { return 1; }, []);          (never read, still consumed)
//	invalid: const x = useMemo(() => { foo(); }, []);
//	invalid: useMemo(() => { return 1; }, []);
//
// Ported from React's `void-use-memo`. Both authorities implement it and they agree completely:
//
//	react 7.1.1   node_modules/eslint-plugin-react-hooks/cjs/eslint-plugin-react-hooks.development.js
//	              `validateUseMemo` at line 45992 and `hasNonVoidReturn` at 46139, unminified
//	oxc compiler  crates/oxc_react_compiler/src/react_compiler_validation/validate_use_memo.rs
//	              a line-for-line transcription of the same function
//	oxc linter    crates/oxc_linter/src/rules/react/void_use_memo.rs, one pass and one fail case
//
// The two corpus cases were run through React itself and it reports identically to oxc's snapshot,
// span included, so there is no divergence to record between the implementations. Every behavioral
// claim below was measured by driving React's own rule through the ESLint Linter API on the input
// named, rather than derived from reading. That distinction earns its place twice over, because
// reading produced the wrong answer twice and one of the two wrong answers came from React's own
// documentation.
//
// # Whether this needs the intermediate representation, and why it does not
//
// It does not, and that is a measurement rather than a preference.
//
// Upstream runs this over its lowered representation, walking every block's instructions and keying
// three side tables on identifier ids: which temporaries hold `useMemo`, which hold `React`, and
// which hold a function expression. That reads like dataflow, and the pipeline position invites the
// same conclusion from the other direction: it runs at position 3, *before* single-assignment form,
// which is the strongest available evidence it needs no value numbering.
//
// The decisive evidence is neither of those. It is that **every input where dataflow could change
// the answer comes back clean**, and clean for a reason that makes the rule strictly weaker than it
// looks. `functions` is keyed on the temporary the `FunctionExpression` instruction produced. Bind
// that callback to a variable first and a `StoreLocal` plus a `LoadLocal` interpose, the identifier
// at the call site is a different one, and the lookup misses. Measured, all clean:
//
//	const cb = () => { foo(); }; useMemo(cb, []);       callback through a variable
//	const m = useMemo; m(() => { foo(); }, []);          callee through a variable
//	const R = React; R.useMemo(() => { foo(); }, []);    namespace through a variable
//	let cb = ...; cb = () => { foo(); }; useMemo(cb, []) reassigned callback
//	const cb = a ? f : g; useMemo(cb, []);               conditional callback
//
// So the rule only ever sees a callback written **syntactically inline at the call site**, and a
// callee that is syntactically `useMemo` or `React.useMemo`. Those are properties of the text. This
// is the same verdict `error-boundaries` reached and for the same reason: a syntactic property
// wearing a control-flow costume, where the costume exists because upstream has no syntax tree left
// by the time the validator runs. Building a graph to recover what the parent chain already states
// is the cargo cult the port brief names, and here it would additionally have to *reproduce a
// lookup failure* to stay faithful, which is a strange thing to build machinery for.
//
// The intermediate representation in `internal/utilities/hir` was read before this was written, and
// nothing in it is used. Recorded plainly so the next reader does not assume the omission was an
// oversight: this rule was the first candidate consumer after that package landed, and it declined.
//
// # The two findings, stated exactly
//
// **A callback returns nothing.** Upstream's `hasNonVoidReturn` asks whether *any* block in the
// lowered callback ends in a return terminal whose variant is `Explicit` or `Implicit`. Only the
// synthetic fallthrough terminal every function gets is `Void`. So the question is whether the
// callback contains a reachable `return` statement at all, or is a concise arrow body.
//
// **A result is unused.** Upstream records the call's own temporary and deletes it when any later
// instruction or terminal takes it as an operand. Since the very next instruction for `const x =
// call()` is a store consuming that temporary, this is not liveness: it is whether the call's value
// is syntactically discarded. Measured boundary, and it is exactly three shapes:
//
//	useMemo(...);                  reports    a bare expression statement
//	(useMemo(...));                reports    parenthesized, still discarded
//	(useMemo(...), 0);             reports    the LEFT operand of a comma is discarded
//	(0, useMemo(...));             clean      the right operand becomes the value
//	void useMemo(...);             clean      `void` takes it as an operand
//	const x = useMemo(...);        clean      even when `x` is never read afterwards
//
// The `const x` case is the one that decides the design and it is the natural wrong guess. A rule
// built on "is the result ever read" would report it, and upstream does not.
//
// # Where reading was wrong, both times
//
// **A bare `return;` is CLEAN.** React's `ReturnStatement` lowering hardcodes `returnVariant:
// 'Explicit'` whether or not an argument is present, minting `undefined` as the value, so a
// callback whose only return is `return;` satisfies `hasNonVoidReturn` and does not report. This
// contradicts React's own pass documentation, which states at
// `docs/passes/40-validateUseMemo.md` that "A function with only `return;` statements (void
// returns) will trigger the 'must return a value' error." The executable is the authority and the
// documentation is wrong about it. Pinned by a fixture, because it is the single most likely place
// for a port to helpfully improve on upstream and be wrong.
//
// **The `Void` variant is unreachable for a real return statement.** It is minted only for the
// synthetic terminal closing a function body, which is why the check reads as a three-way test and
// behaves as a two-way one.
//
// # What the name `LoadGlobal` does not mean
//
// It does not mean "resolves to nothing". It means "not bound inside the function being compiled",
// and module scope is outside it. Measured, and the pair is the whole of the shadowing rule:
//
//	const useMemo = ...; function Component() { useMemo(...) }     reports  (module scope)
//	function Component() { const useMemo = ...; useMemo(...) }     clean    (function scope)
//	function outer() { const useMemo = ...; function Component() { useMemo(...) } }   clean
//
// So the shadow test walks up from the call site to the enclosing compiled function and stops
// there. A binding in any enclosing *function* shadows; one at module scope does not. An import
// does not shadow either, which is why `import {useMemo} from 'react'` still reports, and why an
// alias `import {useMemo as useM}` escapes entirely: the test is on the written name.
//
// # Which functions are eligible
//
// The same driver gate `error-boundaries` reproduces, and it is shared with that rule rather than
// restated. Verified against this rule specifically on fifteen inputs, all agreeing: a lowercase
// name is clean, `_Private` and `Éomponent` are clean, `usething` is clean, `use2Things` reports, a
// nested component is clean, a class method is clean, three parameters is clean, a rest parameter
// is clean, `(props, ref)` reports, `(props, other)` is clean, and returning an object is clean.
//
// One clause differs from `error-boundaries` and it is worth stating because it runs the opposite
// way. That rule could omit `callsHooksOrCreatesJsx` on the grounds that it can only fire on a
// function containing JSX. This rule can fire on a function containing no JSX at all, and measured,
// `function Component() { const x = useMemo(() => { foo(); }, []); return x; }` **reports**. The
// clause is satisfied anyway, because a `useMemo` call is itself a hook call, so the gate is
// satisfied everywhere this rule could report and is likewise not written out.
//
// Bare `use` is the one place this rule's name test differs from the package's shared
// `isHookIdentifierName`. That predicate answers true for `use`, correctly, because
// `rules-of-hooks` needs it to. Measured here, `function use()` is **clean**, so the driver gate
// wants `use` followed by an uppercase letter or a digit and the bare form is excluded below.
//
// # Where the findings point
//
// Two different places, both byte-measured. The void-return finding points at the **callback**, and
// oxc's snapshot independently confirms it by logging column 25 for `() => {}`. The unused-result
// finding points at the **callee**, which is the bare identifier for `useMemo(...)` and the whole
// member expression for `React.useMemo(...)`.
//
// Parentheses are transparent in both positions and excluded from both spans, which the port brief
// makes a mandatory measurement. Measured on React rather than guessed: a parenthesized callback
// reports at column 22 rather than 21, and a parenthesized callee at column 4 rather than 3.
//
// # Divergences recorded rather than smoothed over
//
// **`void useMemo(...)` is silent and arguably should not be.** The `void` operator discards the
// value as plainly as an expression statement does, but upstream emits an instruction taking it as
// an operand, so the temporary is consumed and the finding is suppressed. Reproduced as silence.
//
// **An aliased import escapes the rule entirely.** `import {useMemo as useM} from 'react'` is a
// real `useMemo` and neither authority reports it, because both test the written name rather than
// resolving it. cohere has a whole-program checker and could resolve it. Deliberately not done: it
// would report a class of finding neither authority produces, and a differential would show it.
var VoidUseMemo = rule.Rule{
	// No namespace prefix. The config writes `react/void-use-memo`, and the parity guard strips the
	// namespace on a `/` boundary, so a self-namespaced `react-void-use-memo` would match no
	// inventory entry and lint nothing.
	Name: "react-hooks/void-use-memo",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				call := node.AsCallExpression()

				enclosing := enclosingFunctionOf(node)
				if enclosing == nil || !isUseMemoCallee(call.Expression, enclosing) {
					return
				}
				if !isVoidUseMemoCompiledFunction(enclosing) {
					return
				}

				callback := inlineUseMemoCallback(call)
				if callback == nil {
					return
				}

				// Upstream's `if (!hasNonVoidReturn) { report } else { record as maybe-unused }`.
				// The two findings are exclusive for one call, which is why this is an else rather
				// than a second test: a callback that returns nothing is never also reported as
				// unused. Measured on React, which gives one finding for
				// `useMemo(() => { foo(); }, []);` and not two.
				if !hasNonVoidReturn(callback) {
					ctx.ReportNode(callback, messageUseMemoCallbackReturnsNothing)
					return
				}
				if isValueDiscarded(node) {
					ctx.ReportNode(ast.SkipParentheses(call.Expression), messageUseMemoResultUnused)
				}
			},
		}
	},
}

// isUseMemoCallee reports whether a call's callee is the `useMemo` upstream tracks.
//
// Two spellings and no others: a bare `useMemo` identifier, and a `useMemo` property of an object
// written `React`. Both must be unshadowed within the compiled function, which is what `LoadGlobal`
// means and what the boundary argument on the rule establishes.
//
// A computed access is excluded because upstream matches a string property literal rather than an
// arbitrary key, so `React['useMemo']` is clean. A deeper chain is excluded because only a direct
// property of a global `React` counts, so `A.React.useMemo` is clean. An optional *member* access
// is excluded and an optional *call* is not, which is upstream's asymmetry rather than an
// approximation here. All four measured.
func isUseMemoCallee(callee *ast.Node, enclosing *ast.Node) bool {
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return false
	}

	if callee.Kind == ast.KindIdentifier {
		return callee.Text() == "useMemo" && !isShadowedWithin("useMemo", callee, enclosing)
	}

	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	// `React?.useMemo(...)` is clean upstream while `useMemo?.(...)` reports, so the optional token
	// is declined here and never at the call. Measured both ways.
	if access.QuestionDotToken != nil {
		return false
	}
	name := access.Name()
	if name == nil || name.Kind != ast.KindIdentifier || name.Text() != "useMemo" {
		return false
	}
	object := ast.SkipParentheses(access.Expression)
	if object == nil || object.Kind != ast.KindIdentifier || object.Text() != "React" {
		return false
	}
	return !isShadowedWithin("React", object, enclosing)
}

// isShadowedWithin reports whether a name is bound in a scope that encloses the call site, up to
// and including the compiled function.
//
// This is the port of `LoadGlobal`. Upstream resolves each identifier through its own scope
// analysis, so a binding shadows only the references its scope actually covers, and that
// granularity is observable rather than academic. Measured on React, all four:
//
//	function Component() { const useMemo = ...; useMemo(...) }        clean
//	function Component() { { const useMemo = ...; } useMemo(...) }    reports  sibling block
//	function Component() { function inner() { const useMemo = ...; } useMemo(...) }  reports
//	function Component({useMemo}) { useMemo(...) }                    clean    parameter
//
// A coarse "does this function mention the name anywhere" test was written first and would have
// gone silent on the two reporting cases, so this walks *outward* from the call site instead,
// asking each enclosing scope whether it declares the name and stopping at the compiled function.
// A binding in a scope the call site is not inside is never consulted, which is what makes the
// sibling block and the nested function report.
//
// The outer boundary is the counterintuitive half and it is measured: a binding at MODULE scope
// does not shadow, because the compiled unit is the function rather than the file. The loop
// therefore stops at `enclosing` rather than walking to the source file.
func isShadowedWithin(name string, callee *ast.Node, enclosing *ast.Node) bool {
	for scope := callee; scope != nil; scope = scope.Parent {
		if scope.Kind == ast.KindBlock || isFunctionLike(scope) {
			if scopeDeclaresName(scope, name) {
				return true
			}
		}
		if scope == enclosing {
			// The compiled function is the last scope consulted. Anything above it, module scope
			// included, is a global as far as this rule is concerned.
			//
			// This `break` is REDUNDANT and is kept because it states the boundary where a reader
			// looks for it. Measured by mutation: removing it alone changes nothing, because the
			// kind gate above declines a source file (it is neither a block nor function-like) and
			// `scopeStatements` declines it a second time. Three guards defend the module-scope
			// boundary and any two of them suffice; a mutant removing all three is caught by
			// `voidCallbackModuleScopeShadowStillReports`. Recorded rather than deleted, because
			// the redundancy is the reason the boundary is hard to break by accident, and a reader
			// deleting this line on the grounds that a sweep calls it inert should know the other
			// two exist.
			break
		}
	}
	return false
}

// scopeDeclaresName reports whether one scope-introducing node declares name directly.
//
// Only the declarations written in this scope's own statement list are consulted, so a binding
// nested one block deeper is invisible here and is reached only when the walk is standing in that
// block. Function parameters belong to the function's own scope.
//
// `Node.Parameters()` is reached only for a function-like kind. It panics off its kind, which the
// port brief flags as a category rather than an incident, and calling it on the `Block` this walk
// also visits took the whole run down before the guard was added.
func scopeDeclaresName(scope *ast.Node, name string) bool {
	if isFunctionLike(scope) {
		if parameters := scope.Parameters(); parameters != nil {
			for _, parameter := range parameters {
				if parameter.Kind == ast.KindParameter &&
					bindingDeclaresName(parameter.AsParameterDeclaration().Name(), name) {
					return true
				}
			}
		}
	}

	statements := scopeStatements(scope)
	for _, statement := range statements {
		if statement == nil {
			continue
		}
		if statement.Kind == ast.KindVariableStatement {
			list := statement.AsVariableStatement().DeclarationList
			if list == nil {
				continue
			}
			for _, declaration := range list.AsVariableDeclarationList().Declarations.Nodes {
				if bindingDeclaresName(declaration.AsVariableDeclaration().Name(), name) {
					return true
				}
			}
			continue
		}
		if declaresName(statement, name) {
			return true
		}
	}
	return false
}

// scopeStatements returns a scope's own statement list, or nil when it has none.
//
// `Node.Statements()` panics off Block, SourceFile and ModuleBlock, which the port brief flags, so
// every kind is established before the accessor is reached. A function's statements live one level
// down in its body block, and a concise arrow body has none at all.
func scopeStatements(scope *ast.Node) []*ast.Node {
	if scope.Kind == ast.KindBlock {
		return scope.AsBlock().Statements.Nodes
	}
	if isFunctionLike(scope) {
		body := scope.Body()
		if body == nil || body.Kind != ast.KindBlock {
			return nil
		}
		return body.AsBlock().Statements.Nodes
	}
	return nil
}

// declaresName reports whether a node introduces a binding for name.
//
// Variable declarations, function declarations and class declarations are the shapes that can
// shadow a global here. A nested function's own parameters are not consulted, because a call inside
// a nested function is not this rule's call: `enclosingFunctionOf` would have returned that inner
// function instead.
func declaresName(node *ast.Node, name string) bool {
	switch node.Kind {
	case ast.KindVariableDeclaration:
		return bindingDeclaresName(node.AsVariableDeclaration().Name(), name)
	case ast.KindFunctionDeclaration, ast.KindClassDeclaration:
		declared := node.Name()
		return declared != nil && declared.Kind == ast.KindIdentifier && declared.Text() == name
	}
	return false
}

// bindingDeclaresName reports whether a binding name, possibly destructured, introduces name.
//
// A destructuring pattern is walked rather than tested directly, because `function Component({
// useMemo })` shadows and is measured clean upstream. `Node.Text()` panics on a binding pattern,
// which the port brief flags, so the kind is established before any text is read.
func bindingDeclaresName(binding *ast.Node, name string) bool {
	if binding == nil {
		return false
	}
	if binding.Kind == ast.KindIdentifier {
		return binding.Text() == name
	}
	switch binding.Kind {
	case ast.KindObjectBindingPattern, ast.KindArrayBindingPattern:
		found := false
		binding.ForEachChild(func(child *ast.Node) bool {
			if child.Kind == ast.KindBindingElement {
				if bindingDeclaresName(child.AsBindingElement().Name(), name) {
					found = true
					return true
				}
			}
			return false
		})
		return found
	}
	return false
}

// inlineUseMemoCallback returns the callback upstream would find, or nil.
//
// Upstream looks up its `functions` table by the first argument's identifier, and that table is
// keyed on the temporary a `FunctionExpression` instruction produced. Only a callback written
// inline at the call site produces such a temporary in the same identifier, so this is a syntactic
// test and the rule's doc comment carries the five measurements establishing it.
//
// A spread in first position is declined explicitly by both implementations, and an empty argument
// list exits before the callback is read. `args.length === 0` rather than `< 2`, so a missing
// dependency array still reaches the check.
func inlineUseMemoCallback(call *ast.CallExpression) *ast.Node {
	if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
		return nil
	}
	first := call.Arguments.Nodes[0]
	// The spread decline is upstream's own branch, transcribed, and it is INERT here. Measured by
	// mutation three ways: replacing the kind, deleting the guard, and even adding
	// `KindSpreadElement` to the accepting switch below all leave every fixture green. The cause is
	// a parse shape rather than another branch: a spread argument is `KindSpreadElement`, which is
	// disjoint from both accepted kinds, so the switch has already declined it. Kept because it
	// records that both authorities decline a spread explicitly rather than incidentally, and a
	// reader who deletes it should know the switch is what is actually doing the work.
	if first == nil || first.Kind == ast.KindSpreadElement {
		return nil
	}
	// Parentheses around the callback are transparent and the span excludes them, measured at
	// column 22 rather than 21 on React.
	first = ast.SkipParentheses(first)
	if first == nil {
		return nil
	}
	switch first.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return first
	}
	return nil
}

// hasNonVoidReturn reports whether a callback contains a return that upstream counts.
//
// Upstream asks whether any block of the lowered callback ends in a return terminal whose variant
// is `Explicit` or `Implicit`. `Implicit` is minted for a concise arrow body and `Explicit` for
// every `return` statement, so the syntactic question is whether the callback is a concise arrow or
// contains a return statement of its own.
//
// **A bare `return;` counts**, because the lowering hardcodes `Explicit` whether or not an argument
// is present. That is the case React's own pass documentation gets wrong, and it is pinned by a
// fixture. Writing this as "returns a value" rather than "contains a return" would be the sensible
// reading and would disagree with the implementation on exactly that input.
//
// A `throw` is not a return and produces no return terminal, so a callback that only throws
// reports. Measured.
//
// Nested functions are skipped: their returns belong to them. A return inside a nested function
// cannot terminate the callback, and counting one would silence a genuinely void callback that
// happens to define a helper.
func hasNonVoidReturn(callback *ast.Node) bool {
	body := callback.Body()
	if body == nil {
		return false
	}
	if callback.Kind == ast.KindArrowFunction && body.Kind != ast.KindBlock {
		// A concise arrow body is upstream's `Implicit` return.
		return true
	}

	found := false
	var walk func(node *ast.Node) bool
	walk = func(node *ast.Node) bool {
		if node == nil || found {
			return false
		}
		if isFunctionLike(node) {
			return false
		}
		if node.Kind == ast.KindReturnStatement {
			found = true
			return true
		}
		node.ForEachChild(walk)
		return false
	}
	body.ForEachChild(walk)
	return found
}

// isValueDiscarded reports whether a call expression's own value is thrown away.
//
// This is the port of upstream's unused-result bookkeeping, and it is a syntactic test rather than
// a liveness one. The measured boundary is exactly three shapes and the rule's doc comment lists
// them: a bare expression statement, that statement parenthesized, and the left operand of a comma
// expression in statement position.
//
// Every other position consumes the value, including a store into a variable that is never read
// afterwards, which is the case that decides this rule needs no dataflow. `void` also consumes it,
// which is a divergence recorded on the rule rather than corrected here.
func isValueDiscarded(call *ast.Node) bool {
	current := call
	for parent := current.Parent; parent != nil; parent = parent.Parent {
		switch parent.Kind {
		case ast.KindParenthesizedExpression:
			// Transparent: keep climbing and let the real parent decide.
			current = parent
		case ast.KindExpressionStatement:
			return true
		case ast.KindBinaryExpression:
			// Only the comma operator discards, and only its left operand. The right operand
			// becomes the sequence's value and is consumed. Measured both ways.
			binary := parent.AsBinaryExpression()
			if binary.OperatorToken == nil || binary.OperatorToken.Kind != ast.KindCommaToken {
				return false
			}
			if binary.Left != current {
				return false
			}
			current = parent
		default:
			return false
		}
	}
	return false
}

// isVoidUseMemoCompiledFunction reports whether React's compiler would compile this function.
//
// The driver gate, shared with `error-boundaries` rather than restated, since both rules ask the
// identical question and that rule's implementation was measured across its own fifteen inputs. It
// is verified here against fifteen inputs of this rule's own, listed on the rule.
//
// The one difference is the hook name test. `isReactCompiledFunction` accepts bare `use` through
// this package's `isHookIdentifierName`, which is correct for `rules-of-hooks` and wrong for the
// driver: measured, `function use()` is clean under this rule. So the bare form is excluded here,
// and the exclusion is written as an extra decline rather than by changing the shared predicate,
// which several rules depend on.
func isVoidUseMemoCompiledFunction(fn *ast.Node) bool {
	if !isReactCompiledFunction(fn) {
		return false
	}
	name, named := reactFunctionNameOf(fn)
	if named && name == "use" {
		return false
	}
	return true
}
