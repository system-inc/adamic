package react

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/checking"
	utilsreact "github.com/system-inc/cohere/internal/lint/ecmascript/react"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageUnsupportedEval = rule.Message{
	Id: "unsupportedEval",
	Description: "This component or hook calls `eval`, so React Compiler stops optimizing it. " +
		"The compiler works by tracking where every value comes from and what touches it, and " +
		"`eval` runs code that only exists as a string at runtime, so there is nothing to track " +
		"through. The component still works, it just never gets memoized. Replace the eval with " +
		"the operation it stands in for: `JSON.parse` for data, bracket access for a dynamic " +
		"property, a lookup table for a dynamic function.",
}

var messageUnsupportedWith = rule.Message{
	Id: "unsupportedWith",
	Description: "This component or hook uses a `with` statement, so React Compiler stops " +
		"optimizing it. Inside a `with` block, a bare name might be a property of the object or " +
		"might be an outer variable, and which one it is cannot be known until the code runs, so " +
		"the compiler cannot say where any value came from. The statement is also forbidden in " +
		"strict mode and in modules. Destructure the object or read its properties explicitly.",
}

var messageUnsupportedInlineClass = rule.Message{
	Id: "unsupportedInlineClass",
	Description: "This component or hook declares a `class` in its body, so React Compiler stops " +
		"optimizing it. A class declared during render is a brand new class on every render, so " +
		"every instance of it is a fresh identity that nothing downstream can memoize against, " +
		"and the compiler declines to reason about it rather than guessing. Move the class to " +
		"module scope, where it is declared once.",
}

// UnsupportedSyntax flags syntax React Compiler will not support inside a component or hook.
//
//	valid:   function Component(props) { return <div>{props.text}</div>; }
//	valid:   function helper(a) { eval('a'); return a; }          (not a component or hook)
//	valid:   function Component(props) { const Foo = class {}; return <div />; }  (expression)
//	invalid: function Component(props) { eval('props.x = true'); return <div />; }
//	invalid: function Component(props) { with (props) {} return <div />; }
//	invalid: function Component(props) { class Foo {} return <div />; }
//
// Ported from React's `unsupported-syntax` rule, which is `ErrorCategory.UnsupportedSyntax` in the
// React Compiler. Unlike `gating` and `config`, oxc does ship a linter rule for this one, so both
// authorities were readable and both were run:
//
//	react 7.1.1   node_modules/eslint-plugin-react-hooks/cjs/eslint-plugin-react-hooks.development.js
//	              and driven through the ESLint Linter API on 60-plus probe inputs
//	oxc           crates/oxc_linter/src/rules/react/unsupported_syntax.rs, a thin category filter
//	              over crates/oxc_react_compiler, and driven through the 1.79.0 release binary
//
// # The category is exactly three diagnostics, and the dispatch that sent me here named a fourth
//
// The task named four sites in oxc's `build_hir.rs`: `with` at 6567, inline `class` at 6570, `eval`
// at 875, and a missing parameter binding at 714. **The fourth is not this rule.** Read at
// `diagnostics.rs:1525`, `missing_parameter_binding` constructs with `ErrorCategory::Invariant`,
// while the other three construct with `ErrorCategory::UnsupportedSyntax` at 1533, 1557 and 1566.
// Since oxc's lint rule is `run_react_compiler_rule(ctx, ErrorCategory::UnsupportedSyntax)`, a
// category filter and nothing else, an `Invariant` diagnostic can never surface under this rule's
// name. Grepping `ErrorCategory::UnsupportedSyntax` across the whole crate finds those three
// constructors and one test enumeration, and nothing else. So the count is three, not four.
//
// That diagnostic is also not syntactic in any case. It fires when a plain identifier parameter
// fails to resolve to a local binding after the compiler has already tried a descendant-scope
// lookup, which is a scope-resolution failure inside the compiler rather than a property of the
// source. Its own help text calls it `[BuildHIR] Could not find binding`, which is a note to a
// compiler author rather than to a user.
//
// # React has a fourth that oxc does not, and it is unreachable
//
// React constructs `UnsupportedSyntax` at four sites, not three. The extra one is in
// `resolveBinding` (bundle line 21155): an identifier literally named `this` reports "`this` is not
// supported syntax". oxc has no counterpart.
//
// It was probed rather than reasoned about, and it does not fire. `function Component(props) {
// return <div>{this.x}</div>; }` is silent under React's own rule, because ordinary `this` parses
// as a ThisExpression and never reaches `resolveBinding`, which takes an Identifier node. The one
// upstream fixture that produces a `this`-named identifier is
// `error.typescript-this-param-in-compiled-function`, a TypeScript `this` parameter, and its golden
// reports "Expected a non-reserved identifier name" under `ErrorCategory::Syntax` rather than under
// this category at all. So the site exists, is unreachable through this rule, and is not ported.
// Recorded rather than omitted silently, because the next reader will find it in the bundle.
//
// # The gate, which is the whole rule
//
// None of the three diagnostics fires on code outside a component or hook. Measured on React's own
// rule: `eval` at module top level, `eval` in a lowercase function, and `eval` in a capitalized
// function that returns a number are all silent, and only the component case reports. This is not
// an optimization. It is what the rule decides, and a port without it would report every `eval` in
// every file.
//
// The gate is React Compiler's `compilationMode: 'infer'` inference, `getComponentOrHookLike` at
// bundle line 51105, and it asks three things of the enclosing function:
//
//   - **A name that claims to be a component or a hook.** `/^[A-Z]/` for a component,
//     `/^use[A-Z0-9]/` for a hook. Both are ASCII regexes.
//   - **Evidence it is one**: `callsHooksOrCreatesJsx`, which walks the body for any JSX node or any
//     call whose callee reads as a hook name, and which **skips nested functions**, so JSX that only
//     appears inside an inner closure does not count.
//   - **For components only**, parameters that could be props: at most two, the first not a rest
//     element and not annotated as a primitive, and the second, if present, named with `ref` in it.
//
// Every one of those was measured on React's own rule rather than inferred. `function Component(...props)`
// is silent; `function Component(props, other)` is silent while `function Component(props, ref)`
// reports; `function Component(props: string)` is silent while `function Component(props: {a: number})`
// reports; and a component whose only JSX is inside a nested arrow is silent.
//
// # Where this port narrows the gate deliberately, and why
//
// React's inference also names a function through a variable declarator, an assignment, an object
// property, and a member expression, and it has `forwardRef` and `memo` callback paths. Measured
// against React's own rule, most of those are silent in practice: an object property with a
// capitalized key, an assignment to `ns.Component`, a bare `forwardRef(...)` callback and a
// `memo(...)` callback all produce no finding, while a plain `const Component = (props) => ...`
// does. So this port recognizes two naming forms, the function declaration and the variable
// declarator, and that is a measured subset rather than a guess about which paths matter.
//
// `returnsNonNode` is **not** implemented, and that is also measured rather than skipped. React
// reads it as a component disqualifier, so a function returning an object literal should be
// exempt, and it is not: `function Component(props) { eval('x'); if (props.a) return {}; return <div />; }`
// **reports** under React's own rule. Implementing the check as written would have made this port
// silent on an input React reports. That is the port brief's warning about a helper being more
// careful than upstream, arriving through the reference implementation's own source.
//
// The consequence of both narrowings is the same and it is stated rather than hidden: this rule
// reports a **subset** of what React reports, never a superset. A component named only through an
// object property is missed. Nothing clean is flagged. That is the direction a port should miss in,
// and every narrowing here was chosen after measuring which way upstream actually falls.
//
// # `eval` is resolved, not matched
//
// The `eval` diagnostic fires in `lowerIdentifier` on a binding that resolved to the **global**
// named `eval`, so it is a name-resolution question rather than a syntactic one, and two
// consequences follow that a text match would get wrong in opposite directions.
//
// A shadowed `eval` is silent: `function Component(eval) { eval('x'); return <div />; }` reports
// nothing under React's rule, because the binding resolves to the parameter. And a bare reference
// fires without any call: `const e = eval;` reports, because the identifier is lowered whether or
// not it is being invoked. Both were measured, and a rule keyed on `KindCallExpression` with a
// callee named `eval` would have been wrong on both.
//
// So this rule declares the type checker and asks `type_checking.IsSymbolFromDefaultLibrary`. Probed
// against this rule's own inputs rather than trusted: global `eval` resolves to a symbol declared
// in `lib.es5.d.ts`, a declaration file, while a shadowing parameter or local resolves to a
// declaration in the source file. The shelf helper and a hand-rolled `IsDeclarationFile` loop agree
// on all three shapes.
//
// # A syntactic filter on the eval identifier was written, measured subsumed, and deleted
//
// The first version of this rule carried an `isReadReference` predicate excluding a property name,
// a member access's name half, a declared binding's own identifier, an import specifier and a
// class member, on the reasoning that none of those is a reference to the global. It read as
// obviously load-bearing and a mutation disabling it entirely SURVIVED the whole fixture suite.
//
// Probed rather than patched with another fixture. Over six shapes, printing the guard's answer
// and the symbol lookup's answer side by side, the two columns were identical everywhere: every
// identifier the predicate declined also failed `IsSymbolFromDefaultLibrary`, because a property
// name resolves to the property's own symbol and a declared binding resolves to its declaration in
// the source file. Neither can resolve to `lib.es5.d.ts`.
//
// So it was a subsumed branch rather than a fixture gap, and the port brief's standing advice is to
// delete a subsumed branch and leave the reasoning at the line rather than write a test that
// asserts nothing. The clean cases it was protecting are still fixtures here, now pinned by the
// mechanism that actually decides them.
//
// # `with` is reachable for us where it was not for oxc, which is a parser difference
//
// Probing oxc's release binary on a `.tsx` file, `with` produced a bare `error: 'with' statements
// are not allowed` with no rule code, which is a parse error rather than a finding: TypeScript
// forbids the statement outright. Re-running the same body in a `.jsx` file produced
// `react(unsupported-syntax): JavaScript 'with' syntax is not supported`, so oxc and React agree
// and the earlier zero was an artifact of the extension. That is the port brief's "run a control
// alongside any zero" rule catching a divergence that was not one.
//
// Our parser recovers instead of refusing. Probed on typescript-go: `with (props) {}` produces a
// live `KindWithStatement` in both `.ts` and `.tsx`, so the arm is reachable here in files where
// oxc could not reach it. The rule is therefore wider than oxlint on TypeScript input and exactly
// as wide as React, which is the right side to land on.
//
// # An inline class is a DECLARATION, and the corpus says so
//
// `oxc::Statement::ClassDeclaration` and React's `case 'ClassDeclaration'` both sit in the statement
// dispatch, so a class **expression** is untouched: `const Foo = class {};` inside a component is
// silent at both upstreams and silent here, measured on React's rule and on the release binary.
//
// The corpus carries one case for this and it is a clean one that looks like a violation.
// `error.todo-kitchensink.js` declares `class Bar {}` inside `function foo`, and it is silent under
// React's own rule because `foo` is lowercase, so the gate declines the whole function. Its golden
// reports a different category entirely. That fixture is the reason the gate is not optional: a port
// without it would report an imported case upstream passes.
//
// # Where the findings point
//
// The identifier `eval` alone, the whole `with` statement, and the whole class declaration. Read off
// React's own golden for the eval case, which draws `^^^^` under four characters, and off the
// ESLint runs for the other two, which report columns 3 through 19 for `with (props) {}` and 3
// through 15 for `class Foo {}`. All three reproduce exactly through `ctx.ReportNode`, verified
// against the source bytes before this rule was written, and asserted in the fixtures because a
// message-id assertion cannot see a span.
var UnsupportedSyntax = rule.Rule{
	// No namespace prefix. The config writes `react/unsupported-syntax`, and the parity guard
	// strips the namespace on a `/` boundary, so `react-unsupported-syntax` would match no
	// inventory entry while still passing every fixture in this package.
	Name: "react-hooks/unsupported-syntax",

	// The `eval` arm asks which declaration a name binds to, which is the resolution half of the
	// port brief's table rather than the scope-flag half. Established by probe: `GetSymbolAtLocation`
	// on `eval` answers a `lib.es5.d.ts` declaration for the global and a source-file declaration
	// for a shadowing parameter, and nothing in the abstract syntax tree distinguishes them.
	NeedsTypeChecker: true,

	// Compiler options and the default library, through type_checking's builtin and specifier helpers.
	ProgramReads: rule.ReadsCompilerOptions | rule.ReadsDefaultLibrary,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindIdentifier: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				if node.Text() != "eval" {
					return
				}
				// There is deliberately no syntactic filter here excluding a property name or a
				// declared binding, and its absence is a measurement rather than an omission. See
				// the note below on why the symbol lookup already answers for all of them.
				if !utilsreact.IsInsideComponentOrHook(node) {
					return
				}
				symbol := ctx.TypeChecker.GetSymbolAtLocation(node)
				if !type_checking.IsSymbolFromDefaultLibrary(ctx.Program, symbol) {
					return
				}
				ctx.ReportNode(node, messageUnsupportedEval)
			},
			ast.KindWithStatement: func(node *ast.Node) {
				if !utilsreact.IsInsideComponentOrHook(node) {
					return
				}
				ctx.ReportNode(node, messageUnsupportedWith)
			},
			ast.KindClassDeclaration: func(node *ast.Node) {
				// A class declaration whose parent is the source file is module scope, which is
				// exactly where upstream tells the author to move it. The gate below already
				// declines it, since module scope has no enclosing function, but stating it here
				// keeps the arm readable next to the message that says "move it to module scope".
				if !utilsreact.IsInsideComponentOrHook(node) {
					return
				}
				ctx.ReportNode(node, messageUnsupportedInlineClass)
			},
		}
	},
}
