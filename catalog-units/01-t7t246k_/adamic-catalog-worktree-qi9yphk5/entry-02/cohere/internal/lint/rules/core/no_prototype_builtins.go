package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoPrototypeBuiltins = rule.Message{
	Id: "noPrototypeBuiltins",
	Description: "This calls an `Object.prototype` method directly on an object that may not have " +
		"inherited it. `Object.create(null)` produces objects with no prototype at all, and any " +
		"object can carry a property that shadows the builtin, so parsed JSON like " +
		"`{\"hasOwnProperty\": 1}` turns this call into a crash. Call it through the prototype " +
		"instead: `Object.prototype.hasOwnProperty.call(target, key)`.",
}

// prototypeBuiltinMethods are the three names the rule refuses on a receiver.
//
// This is upstream's set exactly, in both oxc and ESLint, and it is deliberately not every
// `Object.prototype` member: `toString` and `valueOf` are excluded because calling them on an
// arbitrary object is ordinary and near universal, so flagging them would report far more noise
// than bugs. Covering a subset of even these three would be the quiet failure here, so the set is
// pinned by a fixture per name in both directions.
var prototypeBuiltinMethods = map[string]struct{}{
	"hasOwnProperty":       {},
	"isPrototypeOf":        {},
	"propertyIsEnumerable": {},
}

// NoPrototypeBuiltins flags calling `Object.prototype` methods directly on a target object.
//
//	valid:   Object.prototype.hasOwnProperty.call(foo, "bar")
//	valid:   foo.hasOwnProperty
//	valid:   foo[hasOwnProperty]("bar")
//	valid:   foo["hasOwn" + "Property"]("bar")
//	invalid: foo.hasOwnProperty("bar")
//	invalid: foo["isPrototypeOf"]("bar")
//	invalid: foo?.propertyIsEnumerable("bar")
//
// Ported from `eslint/no-prototype-builtins`, cross-read against oxc's port of the same rule.
//
// # It is the call that is wrong, not the access
//
// `foo.hasOwnProperty` on its own is clean and `foo.hasOwnProperty("bar")` is not, so the listener
// is on the call rather than on the property access. `foo.hasOwnProperty.bar()` is clean for the
// same reason read from the other end: the property being *called* is `bar`, and `hasOwnProperty`
// is merely on the path to it.
//
// # Optional chains, which is where the two neighbouring rules got hurt
//
// Every optional spelling reports: `foo?.hasOwnProperty(k)`, `foo?.bar.hasOwnProperty(k)`,
// `foo.hasOwnProperty?.(k)`, and `(foo?.hasOwnProperty)(k)` are all invalid upstream. That makes
// this rule the opposite of `no-extra-non-null-assertion`, whose port documents `IsOptionalChain`
// over-reporting because it propagates down a whole chain. Here the propagation is harmless because
// the rule asks nothing about optionality at all: a call is a call whether or not a `?.` precedes
// it. The trap is real and the fixtures for all four shapes are what prove it does not apply, which
// is why they are kept rather than collapsed into one.
//
// # Parentheses stand in for ESTree's ChainExpression
//
// ESLint reaches through `skipChainExpression` before asking whether the callee is a member
// expression, because `(foo?.hasOwnProperty)(k)` wraps the callee in a `ChainExpression` node there.
// typescript-go has no such node, but it does materialize a `ParenthesizedExpression` for the same
// source, so `ast.SkipParentheses` covers that case and `(a,b).hasOwnProperty(k)` alike.
//
// # No repair is offered, and that is a choice rather than an omission
//
// ESLint attaches a *suggestion* (never an autofix) rewriting to `Object.prototype.<name>.call`,
// and it declines to produce even that in three situations: after an optional chain, on a chain
// expression callee, and whenever the global `Object` is shadowed or redeclared. oxc ships the rule
// with `pending` and offers nothing. The last of ESLint's three conditions is name resolution,
// which is the kind of question that needs the checker, so offering the repair would mean declaring
// `NeedsTypeChecker` for the repair alone while the judgment itself needs no types. Reporting
// without a repair keeps the rule out of the type phase. Note also that the repair upstream offers
// is `Object.prototype.<name>.call` and not `Object.hasOwn`, which reads more modern but resolves
// only `hasOwnProperty`, has no analogue for the other two names, and requires ES2022.
var NoPrototypeBuiltins = rule.Rule{
	Name: "no-prototype-builtins",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				// No nil guard on the callee: a call expression cannot parse without something to
				// call, and error recovery synthesizes a missing-identifier node rather than leaving
				// the field nil. Probed across malformed sources (`()`, `f(`, `(`) as well as
				// `super()`, `import()`, tagged templates and nested parentheses, and it was never
				// nil, so a guard here would be a branch no input can reach.
				callee := ast.SkipParentheses(node.AsCallExpression().Expression)

				name, static := property.AccessedName(callee, property.Static)
				if !static {
					return
				}
				if _, disallowed := prototypeBuiltinMethods[name]; !disallowed {
					return
				}

				// A private field named `#hasOwnProperty` is a different property that merely
				// spells the same word, and `obj.#hasOwnProperty(k)` cannot reach the prototype
				// builtin at all.
				//
				// This guard is deliberately redundant today, and that was measured rather than
				// assumed: a mutation deleting it survived the whole fixture set. `Name().Text()`
				// on a private identifier keeps the `#` sigil for all three names, and a bare
				// `hasOwnProperty` lexes as a plain identifier and never as a private one, so no
				// input can make the set lookup above answer true for a private field. It stays
				// because that safety is an accident of spelling in a helper this rule does not
				// own: were `staticPropertyName` ever to strip the sigil, the corpus case would
				// begin reporting silently. Declining by kind states the decision rather than
				// inheriting it from another file's formatting choice.
				if callee.Kind == ast.KindPropertyAccessExpression {
					property := callee.AsPropertyAccessExpression().Name()
					if property == nil || property.Kind == ast.KindPrivateIdentifier {
						return
					}
				}

				// The finding points at the member expression, matching oxc. ESLint points at the
				// property name alone; oxc's span is the wider of the two and is what the imported
				// corpus was snapshotted against.
				ctx.ReportNode(callee, messageNoPrototypeBuiltins)
			},
		}
	},
}
