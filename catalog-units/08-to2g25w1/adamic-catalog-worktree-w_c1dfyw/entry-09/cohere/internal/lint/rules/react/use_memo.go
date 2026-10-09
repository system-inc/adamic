package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUseMemoMissingCallback = rule.Message{
	Id: "useMemoMissingCallback",
	Description: "This `useMemo` or `useCallback` call was given no arguments at all, so there " +
		"is nothing to memoize. The first argument is the function whose result gets cached " +
		"between renders, and without it the call computes nothing and returns undefined. Pass " +
		"the calculation as an inline function.",
}

var messageUseMemoCallbackNotInline = rule.Message{
	Id: "useMemoCallbackNotInline",
	Description: "The first argument here is not a function written in place. React Compiler " +
		"caches this value by reading the callback's body to learn what it depends on, and a " +
		"name pointing somewhere else gives it nothing to read, so the memoization silently " +
		"stops working. Move the function into the call itself.",
}

var messageUseMemoDependencyListNotArrayLiteral = rule.Message{
	Id: "useMemoDependencyListNotArrayLiteral",
	Description: "The dependency list is not an array literal written in place. The list tells " +
		"React when the cached value is stale, and it is read at compile time rather than at " +
		"runtime, so a variable, a conditional, or a spread leaves nothing to read and the " +
		"memoization cannot be checked. Write the dependencies as a literal array of names.",
}

var messageUseMemoDependencyNotSimple = rule.Message{
	Id: "useMemoDependencyNotSimple",
	Description: "This dependency is not a plain value reference. Each entry has to be something " +
		"React can compare across renders by reading it, which means a name or a chain of " +
		"property accesses off one, such as `x`, `x.y.z`, or `x?.y?.z`. A call, an arithmetic " +
		"expression, or a literal has no identity to track, so it cannot say whether the cached " +
		"value went stale. Compute it into a variable first and depend on that.",
}

var messageUseMemoCallbackHasParameters = rule.Message{
	Id: "useMemoCallbackHasParameters",
	Description: "This memoization callback declares a parameter. React calls it with no " +
		"arguments to compute the cached value, so every parameter is permanently undefined and " +
		"the callback is almost certainly not doing what it reads as doing. Drop the parameters " +
		"and reference the props, state, or local variables the calculation needs directly.",
}

var messageUseMemoCallbackAsyncOrGenerator = rule.Message{
	Id: "useMemoCallbackAsyncOrGenerator",
	Description: "This memoization callback is an async function or a generator. React calls it " +
		"once and stores whatever comes back, so an async callback caches a pending promise and " +
		"a generator caches an iterator, never the value the code is reaching for. Compute the " +
		"value synchronously here, and fetch asynchronous data in an effect instead.",
}

var messageUseMemoCallbackReassignsOuterVariable = rule.Message{
	Id: "useMemoCallbackReassignsOuterVariable",
	Description: "This memoization callback assigns to a variable declared outside of it. The " +
		"callback only runs when the cached value is recomputed, which is on some renders and " +
		"not others, so a write from inside it lands unpredictably and the value it wrote is " +
		"lost on every render that reuses the cache. Return the value instead of assigning it, " +
		"and hold anything that must survive across renders in a ref.",
}

// UseMemo flags a `useMemo` or `useCallback` call whose shape defeats memoization.
//
//	valid:   function Component(props) { const x = useMemo(() => props.a + 1, [props.a]); return <div>{x}</div>; }
//	valid:   function Component(props) { const x = useMemo(() => 1, [props.a.b.c]); return <div>{x}</div>; }
//	valid:   function widget(props) { const x = useMemo(props.fn, [props.a]); return <div>{x}</div>; }
//	invalid: function Component(props) { const x = useMemo(props.fn, [props.a]); return <div>{x}</div>; }
//	invalid: function Component(props) { const x = useMemo(async () => 1, []); return <div>{x}</div>; }
//	invalid: function Component(props) { const x = useMemo((c) => c, []); return <div>{x}</div>; }
//	invalid: function Component() { let x; const y = useMemo(() => { x = 1; return 1; }, []); return [x, y]; }
//
// Ported from React's `use-memo` rule, which is `ErrorCategory.UseMemo` in React Compiler. oxc ships
// a linter rule for it at `crates/oxc_linter/src/rules/react/use_memo.rs`, but that file is a
// ninety-three line category filter over `oxc_react_compiler` with a corpus of one passing and one
// failing case, so the readable authority on both sides is the compiler rather than the lint rule.
//
//	react 7.1.1   node_modules/eslint-plugin-react-hooks/cjs/eslint-plugin-react-hooks.development.js
//	              `extractManualMemoizationArgs` at 42500, `dropManualMemoization` at 42600,
//	              `validateUseMemo` at 45992, `collectTemporaries` at 42432
//	oxc           crates/oxc_react_compiler/src/react_compiler_optimization/drop_manual_memoization.rs
//	              and .../react_compiler_validation/, a Rust transcription of the same functions
//
// Every behavioral claim below was established by **running** React's rule through the ESLint Linter
// API on the input named, over roughly a hundred and ten probe inputs, rather than by reading. That
// distinction earns its place: reading produced a confidently wrong answer twice, both recorded.
//
// # Where the answer lives, and why this does not use the intermediate representation
//
// This rule was written after `internal/utilities/hir` landed, and it is the kind of rule that package
// exists for: upstream's seven diagnostic constructors sit in `dropManualMemoization`, an
// optimization pass, and in `validateUseMemo`, a validator, and both read an instruction stream
// rather than syntax. The obvious move was to lower and transcribe.
//
// **It was probed rather than assumed, and the probe said no.** Lowering
// `function Widget(props) { const x = useMemo(props.fn, [props.a]); return x; }` produces an
// instruction stream that is a name-for-name match with upstream's shapes, exactly as the package
// comment promises:
//
//	$13 = LoadGlobal useMemo
//	$15 = PropertyLoad $14.fn
//	$18 = Array [$17]
//	$19 = Call $13($15, $18)
//
// `collectTemporaries` switches on `LoadGlobal`, `PropertyLoad`, `FunctionExpression` and
// `ArrayExpression`, and all four exist here under those names. `Function.Kind` already carries the
// component-or-hook guess, and `FunctionExpression` reaches a nested `Function` carrying `IsAsync`,
// `IsGenerator` and `Params`, which is three of the seven diagnostics answered directly.
//
// **The obstruction is the one thing this rule needs most.** Upstream's entire callee test is
// `getGlobalDeclaration(value.binding)`, which switches on the binding's kind: a true global, a
// module-local, a named import, a default import, a namespace import. Our `LoadGlobal` declares that
// field as `BindingKind GlobalBindingKind` with five variants, and **lowering never sets it to
// anything but `GlobalBindingKindGlobal`**. Probed directly: `import {useMemo} from 'react'` and a
// bare undeclared `useMemo` both lower to `LoadGlobal name="useMemo" bindingKind=0 source="" imported=""`,
// which is indistinguishable. `internal/utilities/hir/lower.go:61` says so in its own words, and adds
// that `FunctionExpression.Captures` is always empty and that this "is the largest single gap and it
// matters most to the memoization rules". This is one of those rules.
//
// So the intermediate representation would have supplied the shape of the answer while withholding
// the discrimination, and the discrimination is not decoration: the callee test decides whether the
// rule looks at a call at all, and getting it wrong is the difference between reporting every
// `useMemo` in the tree and reporting none. Whatever answered it would have to walk the syntax tree
// to the import declaration regardless, at which point the lowering is a second representation
// carried alongside the first for no remaining question.
//
// The question this rule actually asks is **what is written at this call site**: is the callee that
// name, is the first argument a function written here, is the second an array literal here, does
// this callback declare a parameter, is it async. Those are properties of text. The one question
// with any depth, whether a write inside the callback targets a binding declared outside it, is a
// scope question the checker answers directly through symbol identity. None of them is a question
// about paths, so none of them wants a control-flow graph. This is `error-boundaries`'s finding
// arriving a second time in the same package: a syntactic property wearing a compiler-pass costume,
// and the costume is upstream's because upstream had no syntax tree left by the time it looked.
//
// **What would genuinely change the verdict**, stated so the next reader can check rather than
// re-derive: populating `LoadGlobal.BindingKind`, `Source` and `Imported` during lowering. Those
// three fields already exist and are already documented; only the assignment is missing. With them
// set, transcribing `collectTemporaries` becomes strictly better than this rule's callee walk,
// because it would answer the same question with upstream's own code shape. Until then this rule
// would be carrying a lowering to obtain nothing it does not already have.
//
// # Whether `dropManualMemoization`'s judgment is required, which is the question the dispatch asked
//
// Four of the seven constructors live in `dropManualMemoization` and `extractManualMemoizationArgs`,
// and it is an optimization pass rather than a validator, so the question is fair. The answer is no,
// and it is worth being precise about why, because "the diagnostics are in an optimization pass"
// sounds like it should mean something.
//
// That pass does two separable things. It **rewrites** the call, replacing `useMemo(fn, deps)` with
// a plain `fn()` and threading `StartMemoize`/`FinishMemoize` markers around it so later passes can
// check the memoization was preserved. And on the way through it **complains** about argument shapes
// it could not make sense of. Only the second half constructs `ErrorCategory.UseMemo`, and every one
// of those four constructors fires on a syntactic property of the argument list, before the rewrite
// it guards has happened. Nothing downstream of the rewrite feeds back into them. So what is needed
// here is the pass's **complaint**, which is local, and not its **judgment**, which is about what the
// instruction stream should become.
//
// The one place that reads as an exception is the fourth constructor, "Expected the first argument to
// be an inline function expression", which upstream guards behind `isValidationEnabled`, a disjunction
// of three compiler config flags. That gate is a compiler concern: the diagnostic exists to protect
// the memoization-preservation checks that only run when those flags are on. Measured, the ESLint rule
// reports it unconditionally, because the plugin turns those flags on. Reproduced as unconditional.
//
// # The seven diagnostics, stated precisely
//
// Seven, not six. The dispatch that sent me here said six, and grepping `category: ErrorCategory.UseMemo`
// across the bundle finds seven constructors, at lines 42537, 42564, 42580, 42634, 46048, 46060 and
// 46123. All seven are reachable through the ESLint rule and all seven are ported. `VoidUseMemo` is a
// separate category with its own rule name, `void-use-memo`, and is a sibling's port rather than part
// of this one.
//
//	missing callback     `useMemo()` with no arguments at all. Points at the whole call.
//	not inline           the first argument is not a function written at the call site. A name, a
//	                     member access, a literal, a local variable holding a function: all report.
//	                     Points at the first argument.
//	deps not a literal   a second argument that is not an array literal written at the call site.
//	                     Points at the second argument.
//	dep not simple       an element of that array that is not a name or a static property chain off
//	                     one. Points at the element, and reports once per bad element.
//	callback parameters  the inline callback declares at least one parameter. Points at the first
//	                     parameter.
//	async or generator   the inline callback is `async` or a generator. Points at the whole callback.
//	reassigns outer      an assignment inside the inline callback whose target is a binding declared
//	                     outside it. Points at the assignment target.
//
// The first four are shared with `useCallback`; the last three are `useMemo` only. Measured:
// `useCallback((c) => c, [])` and `useCallback(async () => 1, [])` are both silent, while
// `useCallback(props.fn, [])` reports. That asymmetry is upstream's, because `validateUseMemo` keys
// on `useMemos` alone while `dropManualMemoization` keys on both.
//
// # Which callee counts, which is the whole gate on the call
//
// Upstream's `collectTemporaries` accepts two shapes, and both resolve the binding rather than
// matching text. Measured over a thirteen-input matrix holding the body byte-identical:
//
//	useMemo(...)                  reports, when `useMemo` resolves to no local declaration
//	import {useMemo} from 'react' reports
//	import {useMemo as um}, um()  reports  (keyed on the IMPORTED name, not the local one)
//	import {somethingElse as useMemo} from 'react'   silent  (imported name is not `useMemo`)
//	import {useMemo} from 'not-react'                silent  (module is not a React module)
//	const useMemo = ...           silent  (a local declaration shadows it, at either scope)
//	React.useMemo(...)            reports, when `React` resolves to no local declaration
//	import React from 'react'     reports  (keyed on the LOCAL name being `React`)
//	import * as React from 'react'  reports
//	import Rct from 'react', Rct.useMemo()   silent  (local name is not `React`)
//	import React from 'preact/compat'        reports (the receiver has NO module test; see below)
//	Foo.useMemo(...)              silent
//	const React = {useMemo}       silent
//
// The asymmetry between the two import forms is upstream's and is visible in `getGlobalDeclaration`:
// the `ImportSpecifier` arm looks up `binding.imported` while the `ImportDefault` and
// `ImportNamespace` arms look up `binding.name`. A React module is `react` or `react-dom`,
// lowercased, which `isKnownReactModule` in this package already answers and which oxc's
// `is_known_react_module` matches exactly.
//
// **The module test applies to the CALLEE only, never to the receiver**, which is not symmetric and
// is easy to get backwards. The receiver is recognized in `collectTemporaries` by a bare
// `binding.name === 'React'` with no module component at all, so `import React from 'preact/compat'`
// followed by `React.useMemo(...)` reports. That was measured after a surviving mutant contradicted
// the first version of this port, which had the check in both places; see `reactModuleLocalName`.
//
// This is the one place the checker is needed, and only for the negative half: to tell an
// undeclared `useMemo` from one shadowed by a local. `GetSymbolAtLocation` returning a symbol whose
// declarations are all in this file is the shadow case.
//
// # The component gate, and why it is most of the rule
//
// React Compiler compiles components and hooks, nothing else, so none of these seven fires outside
// one. Measured, holding the call byte-identical and varying only the enclosing function:
// `function widget(props)` is silent, `function Component(props)` reports, `function usething` is
// silent, `function useThing` reports, module scope is silent, `function Component(...props)` is
// silent, `function Component(props, other)` is silent, `function Component(props, ref)` reports.
//
// That gate is exactly the react shelf's `IsInsideComponentOrHook`, written for
// `unsupported-syntax` and measured there over seventeen probe rounds, so it is reused rather than
// rewritten. One difference from `error-boundaries` was measured rather than inherited: a
// `Component` declared **inside**
// another function **does** report here, where `error-boundaries` measured its equivalent silent.
// `IsInsideComponentOrHook` already answers that correctly, since a nested function reachable
// through a component-or-hook-named ancestor stays a root candidate.
//
// # Two readings that were wrong, recorded because the next reader will make them
//
// **The imports.** An early probe round concluded that any import of `useMemo` is silent and that
// oxc, which explicitly handles `ImportSpecifier`, therefore diverges from React. That was going to
// be shipped as a recorded divergence. It was false: the inputs that measured silent imported from
// `'not-react'`, and the conclusion was drawn across two probe batches that also differed in the
// enclosing function's name. Re-run as one controlled matrix varying only the import, React and oxc
// agree completely. This is the port brief's "a zero is three different things" arriving through a
// changed control rather than a bad pattern, and the only reason it was caught is that the second
// batch contradicted the first.
//
// **The rest parameter.** `useMemo((...a) => a, [props.a])` is silent, and so is
// `useMemo((b, ...a) => b, [props.a])`, while `useMemo((b) => b, [props.a])` reports. Reading
// upstream says this should report: `lowerParams` pushes a `Spread` place for a rest element, so
// `params.length > 0` holds. Probed further, a rest parameter **anywhere in the component or in any
// function nested in it** silences the whole function for this rule, including calls that have
// nothing to do with the callback carrying it: with `const bad = useMemo(props.fn, [props.a])` held
// fixed, adding an unrelated `const helper = (...a) => a;` to the same component makes `bad` stop
// reporting, and the control without the helper reports. Running every `react-hooks` rule over those
// inputs shows no Todo and no other diagnostic, so it is a silent whole-function bailout somewhere
// in the compiler rather than a judgment about the call.
//
// **That is not reproduced**, and the choice is deliberate rather than an omission. It is not a
// property of the call under judgment, it has no user-visible explanation upstream, and reproducing
// it would mean going silent on correct findings because of an unrelated function elsewhere in the
// file. Recorded as a real divergence: a differential against React's own rule will show findings
// here that it does not report, on files containing a rest parameter, and `restParameterElsewhere`
// in the fixtures pins that case as reporting with the reasoning at the line.
//
// # Where the findings point
//
// Every span below was measured by slicing the probe input's own bytes at the offsets React's rule
// reported, and every one is asserted in the fixtures, because a message-id assertion cannot see a
// span. The whole-call span for the no-arguments case, the callback span for async and generator,
// the first-parameter span for parameters, and the assignment-target span for a reassignment are
// four different anchors in one rule, and nothing but a span assertion separates them.
var UseMemo = rule.Rule{
	// No namespace prefix. The config writes `react/use-memo`, and the parity guard strips the
	// namespace on a `/` boundary, so a self-namespaced `react-use-memo` would match no inventory
	// entry while still passing every fixture in this package.
	Name: "react-hooks/use-memo",

	// The callee test asks which declaration a name binds to, which is the resolution half of the
	// port brief's scope table rather than the scope-flag half: a bare `useMemo` and a locally
	// declared one are identical in the syntax tree and differ only in what they resolve to. The
	// reassignment check asks the same kind of question about the assignment target.
	NeedsTypeChecker: true,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				call := node.AsCallExpression()
				if call == nil {
					return
				}
				kind, isManualMemo := manualMemoKind(ctx, call.Expression)
				if !isManualMemo {
					return
				}
				if !utilsreact.IsInsideComponentOrHook(node) {
					return
				}

				arguments := callArguments(call)
				if len(arguments) == 0 {
					// Upstream reports the whole call here, not the callee, because
					// `extractManualMemoizationArgs` uses `instr.value.loc`.
					ctx.ReportNode(node, messageUseMemoMissingCallback)
					return
				}

				callback := arguments[0]
				inlineCallback := inlineFunctionArgument(callback)
				if inlineCallback == nil {
					// Upstream's ordering: `extractManualMemoizationArgs` runs first and reports a
					// spread first argument as a missing callback, then the not-inline check runs.
					// A spread has no place to point, which is why it is declined entirely; see
					// the note on `inlineFunctionArgument`.
					if callback.Kind == ast.KindSpreadElement {
						return
					}
					ctx.ReportNode(callback, messageUseMemoCallbackNotInline)
				}

				// The dependency check runs whether or not the callback was inline, which was
				// measured rather than assumed: `useMemo(props.fn, [f()])` reports twice.
				if len(arguments) >= 2 {
					checkDependencyList(ctx, kind, arguments[1])
				}

				if inlineCallback == nil {
					return
				}
				checkInlineCallback(ctx, kind, inlineCallback)
			},
		}
	},
}

// manualMemoKindValue names which of the two manual memoization hooks a callee resolved to.
//
// Carried rather than collapsed to a boolean because three of the seven diagnostics are `useMemo`
// only, which is upstream's asymmetry and is measured at the rule.
type manualMemoKindValue uint8

const (
	manualMemoKindUseMemo manualMemoKindValue = iota
	manualMemoKindUseCallback
)

// manualMemoKind answers upstream's `collectTemporaries` callee test.
//
// Two shapes, both resolving the binding rather than matching text. See the matrix at `UseMemo` for
// the thirteen inputs that pin this, and note that the two import forms are keyed differently: a
// named import on the name it was exported under, a default or namespace import on the local name.
func manualMemoKind(ctx rule.Context, callee *ast.Node) (manualMemoKindValue, bool) {
	if callee == nil {
		return 0, false
	}
	switch callee.Kind {
	case ast.KindIdentifier:
		name, resolved := reactModuleBindingName(ctx, callee)
		if !resolved {
			return 0, false
		}
		return manualMemoKindForName(name)
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		if access == nil {
			return 0, false
		}
		property := access.Name()
		if property == nil || property.Kind != ast.KindIdentifier {
			return 0, false
		}
		kind, isManualMemo := manualMemoKindForName(property.Text())
		if !isManualMemo {
			return 0, false
		}
		object := access.Expression
		if object == nil || object.Kind != ast.KindIdentifier {
			return 0, false
		}
		// The receiver is keyed on the LOCAL name being `React`, for a global and for a default or
		// namespace import alike. Measured: `import Rct from 'react'` then `Rct.useMemo(...)` is
		// silent, because upstream's `ImportDefault` arm looks up `binding.name`.
		name, resolved := reactModuleLocalName(ctx, object)
		return kind, resolved && name == "React"
	}
	return 0, false
}

// manualMemoKindForName maps upstream's two recognized hook names.
//
// An exact match on both, rather than `isHookName`, because upstream compares string equality
// against these two literals in `collectTemporaries`. A custom hook whose name merely looks like a
// hook is not a manual memoization call and is silent, which was measured.
func manualMemoKindForName(name string) (manualMemoKindValue, bool) {
	switch name {
	case "useMemo":
		return manualMemoKindUseMemo, true
	case "useCallback":
		return manualMemoKindUseCallback, true
	}
	return 0, false
}

// reactModuleBindingName answers the name upstream's `getGlobalDeclaration` would look the callee up
// under, for an identifier callee.
//
// Three outcomes. An identifier with no declaration anywhere is a true global and answers its own
// text. An identifier declared by a named import from a React module answers the name it was
// exported under, which is what makes `import {useMemo as um}` report on `um(...)` and
// `import {somethingElse as useMemo}` stay silent. Anything else declared in source is a local
// binding that shadows, and answers nothing.
func reactModuleBindingName(ctx rule.Context, identifier *ast.Node) (string, bool) {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		// Nothing resolved, so the name is a true global. This is the common case in the fixtures
		// and in real source that relies on the automatic React runtime.
		return identifier.Text(), true
	}
	specifier := importSpecifierDeclaration(symbol)
	if specifier == nil {
		// Declared in source and not by an import, so a local binding shadows the global.
		//
		// This guard is SUBSUMED by the module lookup below and is kept for what it says rather
		// than for what it decides. Measured: a mutation neutralizing it to `specifier == nil &&
		// false` survives the whole fixture set, and a hunt for a distinguishing input over seven
		// shapes that reach this line, a local const, a module const, a function declaration, a
		// parameter, a default import, a namespace import and an import-equals, produced an
		// identical verdict on every one with and without the guard. The reason is mechanical:
		// `importSpecifierModuleName(nil)` walks parents from nil and answers `("", false)` at
		// once, so the `!hasModule` test below returns the same `("", false)` this line would.
		//
		// Recorded rather than deleted, because it states the precondition the two lines beneath it
		// assume and deleting it would leave a nil flowing into two accessors on the strength of an
		// argument only this comment carries.
		return "", false
	}
	moduleName, hasModule := importSpecifierModuleName(specifier)
	if !hasModule || !isKnownReactModule(moduleName) {
		return "", false
	}
	return importSpecifierImportedName(specifier), true
}

// reactModuleLocalName answers the same question for a member-expression receiver, where upstream
// keys on the LOCAL name rather than the imported one.
//
// A true global answers its own text, and so does a default or namespace import, **whatever module
// it came from**. A named import is not a receiver shape upstream recognizes and answers nothing,
// matching `getGlobalDeclaration`'s split arms.
//
// # There is deliberately no module test here, and its absence is a correction rather than an
// omission
//
// This function was first written with an `isKnownReactModule` check, by symmetry with the named
// import path above, and that was **wrong**. A mutation removing the check survived the whole
// fixture set, and probing the distinguishing input against React's own rule showed the mutant was
// right and the original was not: `import React from 'preact/compat'` followed by
// `React.useMemo(props.fn, [props.a])` **reports** upstream, and so does the same shape from
// `not-react`.
//
// The mechanism is visible in `collectTemporaries`, which is where the receiver is recognized and
// which does not consult `getGlobalDeclaration` for it at all. Its `LoadGlobal` arm reads
// `else if (value.binding.name === 'React') { sidemap.react.add(...) }`, a bare name comparison
// with no module component, and oxc's transcription is the same: `if !detected && binding.name() ==
// "React"`. Only the *hook* name goes through the module-aware `getHookDetectionName`. So the
// asymmetry between this function and `reactModuleBindingName` is upstream's, and reproducing the
// symmetry a reader expects would go silent on inputs both authorities report.
//
// This is the port brief's first survivor category arriving live: the mutant was right and the
// original was wrong, and no imported fixture could have said so, because upstream's corpus writes
// no non-React module at all.
func reactModuleLocalName(ctx rule.Context, identifier *ast.Node) (string, bool) {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil {
		return identifier.Text(), true
	}
	if reactNamespaceImportDeclaration(symbol) == nil {
		return "", false
	}
	return identifier.Text(), true
}

// importSpecifierDeclaration returns the named-import specifier a symbol was declared by, or nil.
//
// The declarations are looped rather than indexed at zero, which the port brief calls out: a symbol
// can carry several declarations and the index-zero pattern goes silent on arrangements upstream
// reports. The question asked here is "is any declaration a named import", so a loop is the right
// shape for it.
func importSpecifierDeclaration(symbol *ast.Symbol) *ast.Node {
	if symbol == nil {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		if declaration != nil && declaration.Kind == ast.KindImportSpecifier {
			return declaration
		}
	}
	return nil
}

// reactNamespaceImportDeclaration returns the default-import clause or namespace import a symbol was
// declared by, or nil.
//
// Both are receiver shapes for a `React.useMemo(...)` call and upstream treats them identically, so
// they are answered together. A named import is deliberately not accepted here.
func reactNamespaceImportDeclaration(symbol *ast.Symbol) *ast.Node {
	if symbol == nil {
		return nil
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		if declaration.Kind == ast.KindImportClause || declaration.Kind == ast.KindNamespaceImport {
			return declaration
		}
	}
	return nil
}

// importSpecifierModuleName walks from an import declaration part to the module specifier text.
//
// The walk is upward by parent rather than by a typed accessor chain, because the three node kinds
// this is called with sit at different depths under the import declaration: an ImportSpecifier is
// two levels down through NamedImports and ImportClause, an ImportClause is one, a NamespaceImport
// is two. Bounded rather than unbounded so a malformed tree cannot loop.
func importSpecifierModuleName(node *ast.Node) (string, bool) {
	current := node
	for steps := 0; current != nil && steps < 5; steps++ {
		if current.Kind == ast.KindImportDeclaration {
			return importedModuleName(current)
		}
		current = current.Parent
	}
	return "", false
}

// importSpecifierImportedName returns the name a named import was exported under.
//
// `import {useMemo as um}` has a PropertyName of `useMemo` and a Name of `um`, and upstream keys on
// the former. Where there is no rename the PropertyName is absent and the two are the same.
func importSpecifierImportedName(specifier *ast.Node) string {
	imported := specifier.AsImportSpecifier()
	if imported == nil {
		return ""
	}
	if imported.PropertyName != nil && imported.PropertyName.Kind == ast.KindIdentifier {
		return imported.PropertyName.Text()
	}
	name := imported.Name()
	if name == nil || name.Kind != ast.KindIdentifier {
		return ""
	}
	return name.Text()
}

// callArguments returns a call's arguments, or nil.
//
// Guarded rather than reached through directly, because `Arguments` is nil for `f` with no argument
// list at all in a partially recovered tree, and the port brief's standing rule is that every typed
// accessor is assumed to panic off its kind.
func callArguments(call *ast.CallExpression) []*ast.Node {
	if call.Arguments == nil {
		return nil
	}
	return call.Arguments.Nodes
}

// inlineFunctionArgument returns the argument as a function written at the call site, or nil.
//
// Upstream's test is `sidemap.functions.has(fnPlace.identifier.id)`, which holds exactly when the
// value passed was produced by a `FunctionExpression` instruction in this same function, meaning it
// was written inline. Measured, both spellings qualify and a name never does, even one bound to a
// function one line above: `const cb = () => 1; useMemo(cb, [])` reports.
//
// A spread first argument is declined by the caller rather than here. `useMemo(...props.args)` is
// silent upstream, and the reason is visible in `extractManualMemoizationArgs`: a spread makes
// `fnPlace.kind` something other than `Identifier`, so it takes the missing-callback branch, but
// that branch's own `loc` is the call's and the case measured silent rather than reporting. Declined
// with a fixture pinning the silence.
func inlineFunctionArgument(argument *ast.Node) *ast.Node {
	if argument == nil {
		return nil
	}
	switch argument.Kind {
	case ast.KindArrowFunction, ast.KindFunctionExpression:
		return argument
	}
	return nil
}

// checkDependencyList reproduces the two dependency diagnostics.
//
// The split between them is upstream's `ArrayExpression` arm in `collectTemporaries`, which registers
// a value as a dependency list only when `value.elements.every(e => e.kind === 'Identifier')`. A
// spread element or a hole makes an element something other than an Identifier place, so the whole
// literal fails to register and `extractManualMemoizationArgs` then reports it as **not an array
// literal** rather than complaining about the element. Measured both ways: `[...props.a]` and
// `[, props.a]` report "not an array literal", while `[f()]` reports "not a simple expression".
// Reading the source alone would have put both in the second bucket.
func checkDependencyList(ctx rule.Context, kind manualMemoKindValue, argument *ast.Node) {
	if argument == nil {
		return
	}
	if argument.Kind != ast.KindArrayLiteralExpression {
		ctx.ReportNode(argument, messageUseMemoDependencyListNotArrayLiteral)
		return
	}
	elements := argument.AsArrayLiteralExpression().Elements
	if elements == nil {
		return
	}
	for _, element := range elements.Nodes {
		if element == nil {
			continue
		}
		if element.Kind == ast.KindSpreadElement || element.Kind == ast.KindOmittedExpression {
			// One finding for the whole list, not one per offending element, because the list
			// never became a list.
			ctx.ReportNode(argument, messageUseMemoDependencyListNotArrayLiteral)
			return
		}
	}
	for _, element := range elements.Nodes {
		if element != nil && !isSimpleDependencyExpression(element) {
			// One finding per bad element, measured: `[f(), g()]` reports twice.
			ctx.ReportNode(element, messageUseMemoDependencyNotSimple)
		}
	}
}

// isSimpleDependencyExpression answers upstream's `collectMaybeMemoDependencies`.
//
// A dependency is simple when it is a root the compiler can name plus a chain of statically named
// property loads off it. The root is a local variable or a global, meaning an identifier; the chain
// is non-computed property access, optional or not.
//
// **A computed access is where reading and measuring part company, in both directions.** Upstream's
// `PropertyLoad` case handles only a statically named property, so `props['a']` is not a PropertyLoad
// and should fail. It does, and reports. But `props[0]` measured **silent**, which no reading of
// `collectMaybeMemoDependencies` predicts. Both were re-run to confirm. The likeliest account is that
// a numeric index lowers through a different instruction than a string one, but that was not
// established, so it is recorded as measured behavior rather than explained: a numeric computed index
// is accepted here and a non-numeric one is not, because that is what upstream does.
//
// `this` is not an identifier and is not a root, so `[this.a]` reports, which was measured.
func isSimpleDependencyExpression(node *ast.Node) bool {
	current := node
	for {
		switch current.Kind {
		case ast.KindIdentifier:
			return true
		case ast.KindPropertyAccessExpression:
			access := current.AsPropertyAccessExpression()
			if access == nil || access.Expression == nil {
				return false
			}
			current = access.Expression
		case ast.KindElementAccessExpression:
			access := current.AsElementAccessExpression()
			if access == nil || access.Expression == nil {
				return false
			}
			// See the note above: a numeric index is accepted on measured behavior rather than on
			// a reading of upstream, which predicts it should report.
			if access.ArgumentExpression == nil ||
				access.ArgumentExpression.Kind != ast.KindNumericLiteral {
				return false
			}
			current = access.Expression
		case ast.KindNonNullExpression:
			// TypeScript-only syntax with no Babel counterpart, so upstream has no opinion. It is
			// transparent to what the expression names, so it is walked through rather than
			// rejected. No upstream fixture can cover this; it is our tree's own case.
			current = current.AsNonNullExpression().Expression
			if current == nil {
				return false
			}
		default:
			return false
		}
	}
}

// checkInlineCallback reproduces the three `useMemo`-only diagnostics.
//
// Gated on `useMemo` rather than running for both hooks, which is upstream's asymmetry:
// `validateUseMemo` builds its `useMemos` set from the `useMemo` name alone, so a `useCallback`
// callback may take parameters and may be async. Measured on all three.
func checkInlineCallback(ctx rule.Context, kind manualMemoKindValue, callback *ast.Node) {
	if kind != manualMemoKindUseMemo {
		return
	}

	parameters := utilsreact.FunctionParameters(callback)
	if len(parameters) > 0 {
		// Points at the first parameter, whatever its shape. Measured over an identifier, an object
		// pattern and a default: the span is the whole parameter node in each case.
		ctx.ReportNode(parameters[0], messageUseMemoCallbackHasParameters)
	}

	if isAsyncOrGeneratorFunction(callback) {
		// Points at the whole callback, not at the `async` keyword. Measured.
		ctx.ReportNode(callback, messageUseMemoCallbackAsyncOrGenerator)
	}

	reportOuterReassignments(ctx, callback)
}

// isAsyncOrGeneratorFunction reports whether a function-like node is async or a generator.
//
// An arrow cannot be a generator in the grammar, so the asterisk test only ever answers true for a
// function expression, and both are asked uniformly rather than split by kind.
func isAsyncOrGeneratorFunction(node *ast.Node) bool {
	if hasModifierKind(node, ast.KindAsyncKeyword) {
		return true
	}
	if node.Kind != ast.KindFunctionExpression {
		return false
	}
	return node.AsFunctionExpression().AsteriskToken != nil
}

// hasModifierKind reports whether a node carries a given modifier token.
//
// Written against `ast.Node.Modifiers()`, which is nil rather than panicking for a node with no
// modifier list, so the guard is on the list rather than on the kind.
func hasModifierKind(node *ast.Node, kind ast.Kind) bool {
	modifiers := node.Modifiers()
	if modifiers == nil {
		return false
	}
	for _, modifier := range modifiers.Nodes {
		if modifier != nil && modifier.Kind == kind {
			return true
		}
	}
	return false
}

// reportOuterReassignments reproduces `validateNoContextVariableAssignment`.
//
// Upstream keys on the `StoreContext` instruction, which lowering emits for a write to a binding the
// function closed over. The equivalent question in the syntax tree is: is the assignment target a
// plain name whose declaration sits outside this callback. Symbol identity answers it, and it is the
// second and last place this rule needs the checker.
//
// Four measurements shape what is walked and what is not, and three of them are asymmetries a
// reading would not predict:
//
//	x = 1          reports.   A simple assignment to an outer `let`.
//	x += 1         reports.   A compound assignment is still a write.
//	x++            SILENT.    An update expression is not a StoreContext, measured both prefix and
//	                          postfix. Upstream's own gap and reproduced.
//	x.a = 1        SILENT.    A property write is a PropertyStore, not a StoreContext.
//	globalThing=1  SILENT.    A true global is a StoreGlobal, not a StoreContext.
//	const x; x=1   SILENT.    The write is a TypeScript error and never lowers.
//
// Nested functions inside the callback are not descended into, measured: a write to an outer
// variable from a function declared inside the callback is silent, because that write closes over
// the callback's scope rather than the component's from the callback's own point of view.
func reportOuterReassignments(ctx rule.Context, callback *ast.Node) {
	body := utilsreact.FunctionBody(callback)
	if body == nil {
		return
	}
	var walk func(current *ast.Node)
	walk = func(current *ast.Node) {
		if current == nil {
			return
		}
		// A nested function is a barrier: its writes are not this callback's context writes.
		if current != callback && isFunctionLike(current) {
			return
		}
		if current.Kind == ast.KindBinaryExpression {
			binary := current.AsBinaryExpression()
			if binary != nil && binary.OperatorToken != nil &&
				ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
				reportIfOuterBinding(ctx, callback, binary.Left)
			}
		}
		current.ForEachChild(func(child *ast.Node) bool {
			walk(child)
			return false
		})
	}
	body.ForEachChild(func(child *ast.Node) bool {
		walk(child)
		return false
	})
}

// reportIfOuterBinding reports an assignment target that names a binding declared outside the
// callback.
//
// A target that is not a plain identifier is declined, which covers the property-write and
// element-write cases measured silent. A target resolving to no symbol is a true global, also
// measured silent. A target whose declarations all sit inside the callback is a local and is fine.
func reportIfOuterBinding(ctx rule.Context, callback *ast.Node, target *ast.Node) {
	// The identifier test is SUBSUMED by the symbol lookup below, established rather than assumed.
	// Measured: a mutation dropping it to `target == nil` survives the whole fixture set, and a
	// hunt over nine assignment-target shapes chosen to reach it produced an identical verdict on
	// every one with and without the test. A property write, an element write in both spellings, a
	// nested property write, an array and an object destructuring pattern, a `this` write and a
	// parenthesized target all answer zero either way, and the plain-identifier control answers one
	// either way. The reason is that `GetSymbolAtLocation` on a non-identifier target either
	// answers nothing or answers a property symbol with no declaration in this file, which the
	// guard below already declines.
	//
	// Kept rather than deleted, because it states what the checker call beneath it is being asked
	// about and because the two guards fail for unrelated reasons: one is about syntax and one is
	// about resolution, and collapsing them would leave the resolution guard carrying an argument
	// only this comment records.
	if target == nil || target.Kind != ast.KindIdentifier {
		return
	}
	symbol := ctx.TypeChecker.GetSymbolAtLocation(target)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return
	}
	for _, declaration := range symbol.Declarations {
		if declaration == nil {
			continue
		}
		if isNodeWithin(declaration, callback) {
			return
		}
	}
	ctx.ReportNode(target, messageUseMemoCallbackReassignsOuterVariable)
}

// isNodeWithin reports whether a node sits inside another, by walking parents.
//
// Position comparison was the alternative and was declined: two nodes from different source files
// have comparable positions and no containment relationship, and this walks a chain that cannot
// cross a file boundary.
func isNodeWithin(node *ast.Node, ancestor *ast.Node) bool {
	for current := node; current != nil; current = current.Parent {
		if current == ancestor {
			return true
		}
	}
	return false
}
