package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageDefaultParamLastShouldBeLast = rule.Message{
	Id: "shouldBeLast",
	Description: "A parameter with a default sits before a parameter without one, so the default " +
		"can never be taken without passing `undefined` explicitly at every call site. Move it " +
		"after the required parameters, where omitting it is what selects the default.",
}

// DefaultParamLast flags a defaulted or optional parameter written before a required one.
//
//	valid:   function foo(a: number, b = 1) {}
//	valid:   function foo(a: number, b?: number, c = 1) {}
//	valid:   function foo(a = 1, b = 2) {}
//	valid:   function foo(a = 1, ...b) {}
//	invalid: function foo(a = 1, b: number) {}
//	invalid: function foo(a?: number, b: number) {}
//	invalid: class Foo { constructor(public a = 0, private b: number) {} }
//
// Ported from `@typescript-eslint/default-param-last`, the defining implementation. No option
// surface: `meta.schema` is `[]` and no corpus case supplies a second tuple element. One message
// id, `shouldBeLast`, and `meta.fixable` is unset, so there is no repair to port. The rule is
// marked `frozen: true` upstream, meaning its behavior is closed to change.
//
// # The algorithm, and why it walks backwards
//
// Upstream sweeps the parameter list from the END toward the front, carrying one flag: has a plain
// parameter been seen yet. A plain parameter is one that is neither defaulted, nor rest, nor
// optional. When the sweep is standing on a defaulted or optional parameter and a plain one has
// already been passed, that plain parameter is to its right, so the default is unreachable without
// an explicit `undefined`, and the parameter reports.
//
// A rest parameter is neither plain nor reportable: it is skipped in both directions, which is why
// `function foo(a = 1, ...b) {}` is clean while `function foo(a = 1, b: number, ...c) {}` reports.
// The rest element sets no flag, so nothing behind it is condemned by it alone.
//
// The backwards sweep also means the findings are produced in reverse source order. ESLint sorts
// its diagnostics before reporting so this is invisible upstream, and `cohere` sorts too, but a
// forward sweep carrying "is a plain parameter still ahead" is the equivalent formulation and is
// what is written here, so that the findings come out in source order without depending on the
// sort. Both readings were checked against the corpus's four multi-finding cases, which pin the
// order directly.
//
// # The three parameter classifications collapse onto one node here
//
// This is the whole substance of the port, and no imported fixture can see it.
//
// The TypeScript ESTree that upstream reads splits a parameter into distinct node types:
// `AssignmentPattern` carries a default, `RestElement` is a rest, and `.optional` is a flag on
// `Identifier`, `ObjectPattern`, `ArrayPattern` and `RestElement` alike. typescript-go has one
// `KindParameter` carrying all three as optional fields, so the translation is direct:
//
//	AssignmentPattern     Initializer != nil
//	RestElement           DotDotDotToken != nil
//	node.optional         QuestionToken != nil
//
// Upstream's `isOptionalParam` additionally gates on the node type being one of five, which is
// every parameter shape that can carry `optional` at all, so the gate excludes nothing and is not
// reproduced. Its `isPlainParam` is then the negation of the three tests above.
//
// # A parameter property is a WRAPPER upstream and a modifier list here, and the span depends on it
//
// `constructor(public a = 0, ...)` gives upstream a `TSParameterProperty` node wrapping the real
// parameter. The classification questions are asked of the INNER node, and `context.report` is
// called on the OUTER one, so the finding spans the modifier as well as the parameter.
//
// typescript-go attaches the modifiers to the parameter itself, so there is no unwrapping to do and
// the span is the parameter's own token range, which already begins at the modifier. Measured
// against the installed rule, both agree:
//
//	class Foo { constructor(public a = 0, private b: number) {} }
//	    reports [1:25-1:37], span "public a = 0"
//	class Foo { constructor(readonly a = 0, b: number) {} }
//	    reports [1:25-1:39], span "readonly a = 0"
//
// So the absence of an unwrap step here is not a shortcut past upstream's, it is the same decision
// arriving through a different tree shape. The corpus pins four parameter-property spans and they
// are asserted directly.
//
// # A BODYLESS function-like construct is exempt, and that is a false-positive class of its own
//
// Upstream listens on exactly three node types: `ArrowFunctionExpression`, `FunctionDeclaration`
// and `FunctionExpression`. In the TypeScript ESTree, every function-like construct that has no
// body is a DIFFERENT node type, so all of them fall outside those three listeners:
//
//	declare function foo(a = 1, b: number): void;     TSDeclareFunction
//	class C { constructor(a = 1, b: number); }        TSEmptyBodyFunctionExpression
//	abstract class C { abstract m(a = 1, b: number); } TSEmptyBodyFunctionExpression
//	declare class C { m(a = 1, b: number): void; }    TSEmptyBodyFunctionExpression
//	interface I { m(a?: number, b: number): void }    TSMethodSignature
//	type T = (a?: number, b: number) => void;         TSFunctionType
//
// Our parser draws none of those distinctions for the first four: a `declare function` is an
// ordinary `KindFunctionDeclaration` whose `Body()` is nil, and an overload signature is an
// ordinary `KindConstructor` whose `Body()` is nil. So a port that simply listened for our
// function kinds would report every declaration file in the tree, and NOT ONE of the 41 clean
// corpus cases could see it, because the corpus writes no ambient declaration anywhere.
//
// The nil-body guard is what reproduces the silence. Measured against the installed rule, all six
// inputs above are clean, and the two that our parser gives their own kinds anyway
// (`KindMethodSignature`, `KindFunctionType`) are simply not listened for.
//
// The reason upstream is silent is worth stating, because it reads as an oversight and is not: a
// bodyless signature declares a shape rather than an implementation, so no default value is ever
// evaluated from it and there is nothing to be unreachable.
//
// # An accessor is a FunctionExpression upstream, and its arm is load-bearing
//
// `class C { set x(v = 1) {} }` gives upstream a `MethodDefinition` whose `value` is a
// `FunctionExpression`, so upstream's listener fires. Our parser gives accessors their own kinds,
// so each needs its own arm here.
//
// The reason this is written out at length: the obvious argument is that these arms can never
// report, because the grammar gives a setter exactly one parameter and a getter none, so no plain
// parameter can follow a defaulted one. That argument is about the GRAMMAR and it is wrong about
// the PARSER. Probed directly, `class C { set x(a = 1, b: number) {} }` parses to a SetAccessor
// carrying two parameters and a body: the source is illegal, so it is a parse error, and
// typescript-go recovers by attaching the parameters anyway.
//
// Upstream reports it, measured against the installed build, along with the getter form and both
// object-literal forms. Its parser recovers the same way and the node reaches its FunctionExpression
// listener unchanged.
//
// This was found by a mutation emptying the two accessor arms, which survived the entire imported
// corpus plus twelve added cases. The corpus writes no accessor at all, so nothing in it could see
// the arms being deleted, and the reasoning that would have justified deleting them was the wrong
// half of a grammar-versus-parser distinction. Four fixtures pin it now.
var DefaultParamLast = rule.Rule{
	Name: "@typescript-eslint/default-param-last",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		check := func(node *ast.Node) {
			// A bodyless construct declares a shape rather than an implementation, and upstream
			// gives every such construct a node type outside its three listeners. See the doc
			// comment: without this guard the rule reports every ambient declaration in the tree,
			// and the imported corpus cannot see it.
			if node.Body() == nil {
				return
			}
			parameters := node.Parameters()

			// Upstream sweeps backwards carrying "has a plain parameter been seen". Sweeping
			// forwards carrying "is a plain parameter still ahead" decides identically and emits
			// in source order. The last plain parameter's index is what separates the two regions.
			lastPlainIndex := -1
			for index, parameter := range parameters {
				if isPlainParameter(parameter) {
					lastPlainIndex = index
				}
			}
			for index, parameter := range parameters {
				// The boundary and the rest-skip below overlap by exactly one index, and that is
				// measured rather than assumed. A mutation widening this to `index >= lastPlainIndex+1`
				// SURVIVES the whole suite, because the parameter at lastPlainIndex is by
				// definition plain, so it carries no initializer and no question token and the
				// skip below declines it anyway. Widening by two is CAUGHT, on 51 lines. So the
				// break is a bound on the loop rather than a discrimination, and the skip is what
				// decides. Both are kept because both are upstream's: its backwards sweep carries
				// the same pair as a flag and a two-armed test.
				if index >= lastPlainIndex {
					break
				}
				declaration := parameter.AsParameterDeclaration()
				// A rest parameter is skipped rather than reported. It cannot precede a plain
				// parameter in valid source, but the parser accepts the shape under error
				// recovery, and upstream's condition names the two reportable classes rather than
				// negating the plain test, so the rest case falls through both arms.
				if declaration.Initializer == nil && declaration.QuestionToken == nil {
					continue
				}
				// The parameter's own token range already begins at its first modifier, which is
				// what makes a parameter property span its `public` or `readonly` keyword the way
				// upstream's outer TSParameterProperty node does.
				ctx.ReportNode(parameter, messageDefaultParamLastShouldBeLast)
			}
		}

		return rule.Listeners{
			ast.KindFunctionDeclaration: check,
			ast.KindFunctionExpression:  check,
			ast.KindArrowFunction:       check,
			// A class or object-literal method, a constructor, and both accessors are all a
			// `FunctionExpression` in the tree upstream reads, so upstream's FunctionExpression
			// listener reaches every one of them. Our parser gives each its own kind, so each
			// needs its own entry to decide the same cases.
			ast.KindMethodDeclaration: check,
			ast.KindConstructor:       check,
			ast.KindGetAccessor:       check,
			ast.KindSetAccessor:       check,
		}
	},
}

// isPlainParameter answers upstream's `isPlainParam`: a parameter that is neither defaulted, nor
// rest, nor optional.
//
// Upstream writes it as the negation of three tests over its own node types, and its `isOptionalParam`
// carries an additional gate on the node being one of five shapes. Those five are every parameter
// shape that can carry `optional`, so the gate excludes nothing and has no counterpart here.
func isPlainParameter(parameter *ast.Node) bool {
	declaration := parameter.AsParameterDeclaration()
	return declaration.Initializer == nil &&
		declaration.DotDotDotToken == nil &&
		declaration.QuestionToken == nil
}
