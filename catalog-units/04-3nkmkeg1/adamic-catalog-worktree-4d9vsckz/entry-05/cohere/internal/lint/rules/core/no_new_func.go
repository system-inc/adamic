package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoFunctionConstructor = rule.Message{
	Id: "noFunctionConstructor",
	Description: "This builds a function by compiling a string at runtime, which is eval wearing " +
		"a different name. The body is invisible to the type checker, to the bundler, and to every " +
		"tool that would otherwise catch a mistake in it, it cannot close over anything in the " +
		"surrounding scope, and if any part of the string came from input it is a code injection. " +
		"Write a real function, or take one as a parameter.",
}

// functionConstructorMethods are the Function.prototype methods that invoke the constructor when
// applied to `Function` itself. `Function.call(null, 'return 1')` compiles a function exactly as
// `Function('return 1')` does.
var functionConstructorMethods = map[string]bool{
	"apply": true,
	"bind":  true,
	"call":  true,
}

// NoNewFunc flags building a function from a string through the `Function` constructor.
//
//	valid:   var a = new _function("b", "c", "return b+c");
//	valid:   class Function {}; new Function()
//	valid:   call(Function)
//	valid:   Function.toString()
//	invalid: var a = new Function("b", "c", "return b+c");
//	invalid: var a = Function("b", "c", "return b+c");
//	invalid: var a = Function.call(null, "b", "c", "return b+c");
//
// # Three ways to reach the constructor, and upstream catches all three
//
// `new Function(...)` and `Function(...)` are the same thing, since the constructor ignores whether
// it was invoked with `new`. The third is indirect: `apply`, `bind` and `call` on `Function` itself
// invoke it too, so `Function.call(null, 'return 1')` compiles a function. Upstream enumerates
// exactly those three method names and this reproduces the set rather than generalising it.
// `Function.toString()` is a clean case upstream, and it would report under any looser reading.
//
// # Where the finding points, measured rather than inferred
//
// The report is on the invoking expression, which for the indirect form is the `apply`/`bind`/`call`
// call and NOT the outer call it may be part of. Measured on eslint 10.8.1:
//
//	Function.bind(null, "b")()    reports columns 1 through 20, which ends at the bind call's
//	                              closing paren rather than at the trailing ()
//	Function.bind(null, "b")      reports the same span, uncalled
//	Function.bind                 clean, because nothing invoked it
//
// The uncalled `Function.bind(...)` reporting is the tell that the anchor is the method call itself:
// the value is a function that will compile a string whenever it is eventually called, so upstream
// reports at the point the constructor was bound.
//
// # The shadow question, and why this reads the checker
//
// Half of upstream's clean cases are a local binding named `Function`, and there is no structural
// answer: the shadow can be a class, a function declaration, a parameter, or a `var` in any
// enclosing scope. Upstream asks its scope analysis for the global `Function` variable and then
// walks that variable's own references, so a shadowed name never enters the loop.
//
// `resolvesToAGlobal`, shipped for `no-new-native-nonconstructor` and shared with three other rules
// here, asks the equivalent question of the checker: a global is declared in the standard library,
// which is a declaration file, while any binding written in source is a shadow. Probed against all
// thirteen of upstream's clean cases and all ten reporting ones before this was written, and the two
// agreed on every row. The row that matters most is the nested one:
//
//	class Function {}; new Function()                        silent, the class shadows
//	const fn = () => { class Function {} }; new Function()   reports, the class is in an inner scope
//
// A name-only test reports the first; a test that gives up whenever the name is declared anywhere in
// the file is silent on the second. Only resolution separates them, and both are upstream cases.
//
// Upstream additionally requires `variable.defs.length === 0`, which excludes a global that source
// redeclares. That is the same judgment from the other side and `resolvesToAGlobal` answers it
// alike, since a redeclaration in source puts a non-declaration-file declaration on the symbol.
var NoNewFunc = rule.Rule{
	Name: "no-new-func",

	// See the doc above: the whole discrimination is whether `Function` is the global or a local
	// shadowing it, and nothing structural answers that.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		reportIfFunctionConstructor := func(node *ast.Node) {
			if ctx.TypeChecker == nil {
				return
			}

			var callee *ast.Node
			switch node.Kind {
			case ast.KindNewExpression:
				callee = node.AsNewExpression().Expression
			case ast.KindCallExpression:
				callee = node.AsCallExpression().Expression
			default:
				return
			}
			if callee == nil {
				return
			}

			// Fidelity to a parser difference rather than a correctness improvement: ESTree has no
			// node for a parenthesis, so upstream reads through them for free. Measured on eslint
			// 10.8.1, `(Function?.call)(null, "b")` reports and it is upstream's own corpus case.
			callee = ast.SkipParentheses(callee)
			if callee == nil {
				return
			}

			// The direct forms: `Function(...)` and `new Function(...)`.
			if callee.Kind == ast.KindIdentifier {
				if callee.Text() == "Function" && resolvesToAGlobal(ctx, callee) {
					ctx.ReportNode(node, messageNoFunctionConstructor)
				}
				return
			}

			// The indirect form: `Function.call(...)`, `Function.bind(...)`, `Function.apply(...)`,
			// and the same three through a subscript, which upstream reaches because it asks for
			// the static property name rather than reading the dot.
			name, named := property.AccessedName(callee, property.Static)
			if !named || !functionConstructorMethods[name] {
				return
			}
			object := accessedObject(callee)
			if object == nil {
				return
			}
			object = ast.SkipParentheses(object)
			if object == nil || object.Kind != ast.KindIdentifier ||
				object.Text() != "Function" || !resolvesToAGlobal(ctx, object) {
				return
			}
			ctx.ReportNode(node, messageNoFunctionConstructor)
		}

		return rule.Listeners{
			ast.KindNewExpression:  reportIfFunctionConstructor,
			ast.KindCallExpression: reportIfFunctionConstructor,
		}
	},
}
