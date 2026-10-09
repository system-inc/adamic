package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// wrapperObjects are the three primitive wrapper constructors upstream names, in upstream's order.
//
// Verbatim from `eslint/lib/rules/no-new-wrappers.js`, where the list is rebuilt inside the
// `NewExpression` listener on every visit. Hoisting it to package scope is the "fidelity is to what
// the rule decides" allowance: the decision is identical and the allocation is not part of it.
var wrapperObjects = map[string]bool{
	"String":  true,
	"Number":  true,
	"Boolean": true,
}

// NoNewWrappers flags `new String()`, `new Number()`, and `new Boolean()`.
//
//	valid:   var a = new Object();
//	valid:   var a = String('test'), b = String.fromCharCode(32);
//	valid:   function test(Number) { return new Number; }
//	valid:   import String from "./string"; const str = new String(42);
//	invalid: var a = new String('hello');
//	invalid: var a = new Number(10);
//	invalid: var a = new Boolean(false);
//
// Constructed, each of these produces an object rather than a primitive, and the object is truthy
// and compares by reference. `new Boolean(false)` is truthy, `new Number(0) === 0` is false, and
// `typeof new String("x") === "object"`. Called without `new` the same three names are the standard
// conversions and are correct, so the defect is exactly the `new`.
//
// # Why this reads the checker
//
// Four of upstream's seven valid cases are a local binding shadowing the global: a parameter named
// `Number`, an imported default named `String`, a `var` hoisted out of an `else` branch, and a
// `const` in a nested block. `new Number` inside `function test(Number)` constructs whatever that
// parameter holds, and reporting it is a false positive on correct code.
//
// Upstream answers this with `getVariableByName(sourceCode.getScope(node), name)` and then
// `variable.identifiers.length === 0`, which is scope analysis rather than a spelling test: it walks
// out through the scope chain, and the zero-identifier test is what distinguishes the global (whose
// variable is synthesized from the environment and declares nothing in source) from any binding
// written in a file.
//
// We have no resolved-reference index, and there is no structural answer: the shadow can be a
// parameter, an import, a function declaration, a class, or a `var` in any enclosing scope, so no
// bounded walk finds it. The checker answers the same question directly, and the predicate is where
// the name's declaration lives. `String`, `Number` and `Boolean` are declared in the TypeScript
// standard library, which is a declaration file; anything written in source is a shadow. That is
// `resolvesToAGlobal`, shared with `no-new-native-nonconstructor`, which asks the identical question
// about `Symbol` and `BigInt`.
//
// # Two divergences, both measured against the installed rule
//
// **Parentheses on the callee.** ESTree has no parenthesized-expression node, so upstream's
// `node.callee` is the identifier however many parens wrap it, and `new (String)('x')` reports.
// Measured: it reports, and so does `new ((String))('x')`. Our parser gives the callee
// `KindParenthesizedExpression`, so `ast.SkipParentheses` here is fidelity rather than an
// improvement; without it both inputs go silent. Upstream's corpus writes no parenthesized form, so
// nothing imported can see this.
//
// **`globals: {String: "off"}` and `/* global Boolean:off */`.** Two of upstream's valid cases are
// clean only because the ESLint configuration removed the global, which makes `getVariableByName`
// return nothing. That is a configuration layer we do not have, and the checker still resolves the
// name to the standard library, so both report here. They are carried in the test file as reporting
// cases with the reasoning at the line, which is where the brief says to pin a case decided above
// the rule.
//
// # The span
//
// Upstream reports on `node`, the whole `NewExpression`, not on the callee. Its corpus states this
// explicitly: `var a = new String('hello');` is columns 9 through 28, which is `new String('hello')`.
// This differs from `no-new-native-nonconstructor` beside it, which reports on the callee, so the
// two rules are not interchangeable on this point and the fixtures assert the text.
var NoNewWrappers = rule.Rule{
	Name: "no-new-wrappers",

	// See the doc above: four of seven upstream valid cases are shadowing, and nothing structural
	// separates the global from a local of the same name.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindNewExpression: func(node *ast.Node) {
				// No nil guard on the callee, and that is measured rather than assumed.
				//
				// A mutant deleting a `callee == nil` guard survived the whole fixture set, so the
				// question was put to the parser instead of argued. Across ten sources including
				// error-recovery shapes that well-formed source cannot produce -- `new;`, `new ;`,
				// `new`, `new ()`, `new (;`, `new 'x'`, `new [1]`, `new new;`, and a bare `new;` in
				// a method body -- every `NewExpression` came back with a non-nil `Expression`.
				// The degenerate ones get a synthesized Identifier whose `Text()` is empty, which
				// fails the wrapper lookup below and is correctly silent. `ast.SkipParentheses`
				// never returned nil either, including on `new ()` and `new (())`.
				//
				// Callers enumerated when this verdict was taken: this listener only. If a caller
				// is added that can hand this a synthesized or detached node, the verdict is void
				// and the guard has to be re-argued rather than assumed to still hold.
				callee := node.AsNewExpression().Expression

				// Fidelity to a parser difference rather than a correctness improvement: ESTree has
				// no node for a parenthesis, so upstream reads through them for free and reports
				// `new (String)('x')`. Measured on the installed rule; see the doc comment.
				callee = ast.SkipParentheses(callee)
				if callee.Kind != ast.KindIdentifier {
					return
				}

				name := callee.Text()
				if !wrapperObjects[name] {
					return
				}
				if !resolvesToAGlobal(ctx, callee) {
					return
				}

				ctx.ReportNode(node, rule.Message{
					Id: "noNewWrappers",
					Description: fmt.Sprintf(
						"This calls `new %s()`. Constructed, %s returns an object rather than a "+
							"primitive, so the result is always truthy, compares equal to nothing "+
							"by `===`, and reports its `typeof` as \"object\". Drop the `new` and "+
							"call %s directly.", name, name, name),
				})
			},
		}
	},
}
