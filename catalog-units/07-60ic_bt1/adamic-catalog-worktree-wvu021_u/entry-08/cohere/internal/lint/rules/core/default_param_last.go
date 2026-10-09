package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageDefaultParamLastShouldBeLast = rule.Message{
	Id: "shouldBeLast",
	Description: "A parameter carrying a default or an optional marker sits before a parameter " +
		"without one, so the default can never be taken: reaching the later parameter obliges " +
		"every caller to pass something here. Move it after the required parameters, where " +
		"omitting it is what selects the default.",
}

// DefaultParamLast flags a defaulted, optional, or rest parameter written before a required one.
//
//	valid:   function f(a, b = 5) {}
//	valid:   function f(a, b = 5, ...c) {}
//	valid:   function f(a: number, b?: number, c = 1) {}
//	invalid: function f(a = 5, b) {}
//	invalid: function f(a?: number, b: number) {}
//	invalid: function f(...a: number[], b: number) {}
//	invalid: class C { constructor(public a = 0, private b: number) {} }
//
// Ported from ESLint core's `default-param-last`, the defining implementation. No option surface:
// `meta.schema` is `[]` and no corpus case supplies a second tuple element. One message id,
// `shouldBeLast`, and `meta.fixable` is unset, so there is no repair to port. Upstream marks the
// rule `frozen: true`, meaning its behavior is closed to change.
//
// # This is not a duplicate of the shipped namespaced rule, and the difference is measured
//
// `@typescript-eslint/default-param-last` is registered and enabled in this tree. It declares
// `extendsBaseRule: true` in its metadata but imports no `getESLintCoreRule`, so it is a standalone
// reimplementation rather than a wrapper, and the two disagree on one class of input.
//
// Core's `isRequiredParameter` answers false for a rest element, and core's report is then
// UNCONDITIONAL once a required parameter has been seen to the right. The typescript-eslint
// rewrite adds a second gate, reporting only when the parameter is optional or an assignment
// pattern, which excludes a rest element from ever reporting. Driving both installed builds over
// the same inputs:
//
//	function f(...a: number[], b: number) {}              core [1:12-1:26]  ts silent
//	function f(a: number, ...b: number[], c: number) {}   core [1:23-1:37]  ts silent
//	function f(a = 1, ...b: number[], c: number) {}       core 2 findings   ts 1 finding
//
// The same replay over all 96 cases of core's own corpus produced ZERO divergence, so no imported
// fixture can see this and the cases pinning it below are written from the measurement rather than
// from the corpus. A rest parameter before another parameter is illegal source that the parser
// accepts under error recovery, which is the only reason the shape is reachable at all.
//
// # The three parameter classifications collapse onto one node here
//
// The TypeScript ESTree that upstream reads splits a parameter into distinct node types:
// `AssignmentPattern` carries a default, `RestElement` is a rest, and `.optional` is a flag. Probed
// directly, typescript-go gives one flat `KindParameter` carrying all three as optional fields, so
// the translation is a field test apiece:
//
//	AssignmentPattern     Initializer != nil
//	RestElement           DotDotDotToken != nil
//	node.optional         QuestionToken != nil
//
// Nothing here reads a parameter's NAME, which matters: a destructured parameter's name node is a
// `KindObjectBindingPattern` or `KindArrayBindingPattern`, and `Node.Text()` panics on both. The
// corpus writes five such parameters and the walk recovers per file, so one call would have cost
// every rule in the package its verdict on that file.
//
// # A parameter property is a WRAPPER upstream and a modifier list here
//
// `constructor(public a = 0, ...)` gives upstream a `TSParameterProperty` node wrapping the real
// parameter. Core asks its classification questions of the INNER node and calls `context.report`
// on the OUTER one, so the finding spans the modifier as well as the parameter.
//
// typescript-go attaches modifiers to the parameter itself, so there is no wrapper to unwrap and
// the parameter's own token range already begins at the modifier. The corpus pins four such spans
// and they are asserted as text rather than as columns.
//
// # A BODYLESS function-like construct is exempt, and that is a false-positive class of its own
//
// Upstream listens on exactly three node types, and in the TypeScript ESTree every function-like
// construct WITHOUT a body is a different node type that falls outside all three. Our parser draws
// no such distinction: a `declare function` is an ordinary `KindFunctionDeclaration` whose `Body()`
// is nil, and an overload signature is an ordinary `KindConstructor` whose `Body()` is nil. Without
// the nil-body guard this rule would report every ambient declaration in the tree, and not one of
// the 55 clean corpus cases could see it, because the corpus writes no ambient declaration at all.
//
// Measured against the installed core build, all of these are clean, and the two our parser gives
// their own kinds anyway are simply not listened for:
//
//	declare function f(a = 1, b: number): void;          KindFunctionDeclaration, nil body
//	class C { constructor(a = 1, b: number); }           KindConstructor, nil body
//	interface I { m(a?: number, b: number): void }       KindMethodSignature, not listened for
//	type T = (a?: number, b: number) => void;            KindFunctionType, not listened for
//
// The silence is a judgment rather than an oversight: a bodyless signature declares a shape, so no
// default is ever evaluated from it and there is nothing to be unreachable.
//
// # An accessor is a FunctionExpression upstream, and its arm is load-bearing
//
// `class C { set x(v = 1) {} }` gives upstream a `MethodDefinition` whose `value` is a
// `FunctionExpression`, so upstream's listener reaches every method, constructor and accessor. Our
// parser gives each its own kind, so each needs its own arm.
//
// The obvious argument is that the accessor arms can never fire, because the grammar gives a setter
// exactly one parameter and a getter none. That argument is about the GRAMMAR and it is wrong about
// the PARSER: `class C { set x(a = 1, b: number) {} }` is illegal source, and typescript-go recovers
// by attaching both parameters and a body. The installed core build reports all four such shapes.
// This was found on the namespaced sibling by a mutation that emptied both arms and survived its
// entire corpus; the same measurement holds here and four fixtures pin it.
var DefaultParamLast = rule.Rule{
	Name: "default-param-last",
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

			// Upstream sweeps backwards carrying "has a required parameter been seen". Sweeping
			// forwards carrying "is a required parameter still ahead" decides identically and emits
			// in source order rather than reversed. The last required parameter's index is the
			// boundary between the two regions.
			lastRequiredIndex := -1
			for index, parameter := range parameters {
				if defaultParamLastIsRequired(parameter) {
					lastRequiredIndex = index
				}
			}
			for index, parameter := range parameters {
				// The boundary and the required-skip below overlap by exactly one index, and that
				// is measured rather than assumed. Widening this to `lastRequiredIndex+1` SURVIVES
				// the whole suite, and no fixture can catch it: the parameter at lastRequiredIndex
				// is by definition required, so the skip declines it anyway. Enumerating every
				// parameter list up to length six over the four classes, all 5,461 of them, the two
				// spellings produce identical output, while a `+2` control distinguishes on 4,095.
				// So the break is a bound on the loop rather than a discrimination, and the skip is
				// what decides. Both are kept because both are upstream's: its backwards sweep
				// carries the same pair as a flag and a continue.
				if index >= lastRequiredIndex {
					break
				}
				// A required parameter is never itself reported: upstream's backwards sweep
				// `continue`s on one after setting its flag, so the flag is only ever consulted for
				// a parameter that is not required. Without this skip every parameter in an
				// all-required list reports, which 9 of the corpus's clean cases catch directly.
				if defaultParamLastIsRequired(parameter) {
					continue
				}
				// What remains is defaulted, optional or rest, and core reports all three. This is
				// the whole difference from the namespaced sibling, which gates the report a second
				// time on optional-or-defaulted and so never reports a rest parameter.
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

// defaultParamLastIsRequired answers upstream's `isRequiredParameter`: a parameter is required when
// it carries neither a default, nor a rest marker, nor an optional marker.
//
// Named for the rule because a Go package is one namespace and a bare `isRequiredParameter` would
// be reachable from every other rule in this directory.
func defaultParamLastIsRequired(parameter *ast.Node) bool {
	declaration := parameter.AsParameterDeclaration()
	return declaration.Initializer == nil &&
		declaration.DotDotDotToken == nil &&
		declaration.QuestionToken == nil
}
