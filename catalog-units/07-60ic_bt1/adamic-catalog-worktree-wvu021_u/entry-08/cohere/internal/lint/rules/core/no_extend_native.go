package core

import (
	"fmt"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// nativePrototypeBuiltins is the set of global names whose prototype this rule protects.
//
// Upstream computes it rather than writing it down: it takes the keys of the `globals` package's
// latest ECMAScript set and keeps the ones whose first character is already uppercase. That filter
// is why `parseFloat.prototype.x = 1` is a clean case in the corpus, and it is also why `Infinity`
// and `NaN` are in here despite having no prototype worth extending -- the filter is spelling, not
// semantics, and reproducing the decision means reproducing the spelling test's output.
//
// Derived from the `globals` package this repository actually installs (16.4.0, key `es2026`, 62
// globals of which 49 survive the filter) and then VERIFIED against the installed eslint 10.8.1
// build: all 49 report on `<name>.prototype.p = 0` and the 9 lowercase or non-global controls
// (`parseFloat`, `parseInt`, `escape`, `decodeURI`, `isNaN`, `undefined`, `globalThis`, `x`, `Foo`)
// are clean. Written out rather than computed because we have no equivalent of that package, and a
// list that was measured beats a filter that was reimplemented.
var nativePrototypeBuiltins = map[string]bool{
	"AggregateError":       true,
	"Array":                true,
	"ArrayBuffer":          true,
	"Atomics":              true,
	"BigInt":               true,
	"BigInt64Array":        true,
	"BigUint64Array":       true,
	"Boolean":              true,
	"DataView":             true,
	"Date":                 true,
	"Error":                true,
	"EvalError":            true,
	"FinalizationRegistry": true,
	"Float16Array":         true,
	"Float32Array":         true,
	"Float64Array":         true,
	"Function":             true,
	"Infinity":             true,
	"Int16Array":           true,
	"Int32Array":           true,
	"Int8Array":            true,
	"Intl":                 true,
	"Iterator":             true,
	"JSON":                 true,
	"Map":                  true,
	"Math":                 true,
	"NaN":                  true,
	"Number":               true,
	"Object":               true,
	"Promise":              true,
	"Proxy":                true,
	"RangeError":           true,
	"ReferenceError":       true,
	"Reflect":              true,
	"RegExp":               true,
	"Set":                  true,
	"SharedArrayBuffer":    true,
	"String":               true,
	"Symbol":               true,
	"SyntaxError":          true,
	"TypeError":            true,
	"URIError":             true,
	"Uint16Array":          true,
	"Uint32Array":          true,
	"Uint8Array":           true,
	"Uint8ClampedArray":    true,
	"WeakMap":              true,
	"WeakRef":              true,
	"WeakSet":              true,
}

// NoExtendNativeOptions is the decoded option object.
//
// `exceptions` names builtins whose prototype may be extended anyway. The zero value is an empty
// list, which is upstream's default and also Go's, so a rule configured as a bare `"error"` and
// handed nil options lands on the strict setting the same way upstream does.
type NoExtendNativeOptions struct {
	Exceptions []string `json:"exceptions"`
}

// NoExtendNative flags adding a property to a native type's prototype.
//
//	valid:   x.prototype.p = 0
//	valid:   Object.defineProperty(x, 'p', {value: 0})
//	valid:   parseFloat.prototype.x = 1
//	valid:   function foo() { var Object = function() {}; Object.prototype.p = 0 }
//	invalid: Object.prototype.p = 0
//	invalid: Function.prototype['p'] = 0
//	invalid: Object.defineProperty(Array.prototype, 'p', {value: 0})
//
// Extending a native prototype changes an object every other script on the page shares, so a
// property added here appears in every `for...in` over a plain object, collides silently with a
// future standard method of the same name, and is invisible at the call site that breaks.
//
// # Two ways to extend, and the finding points at a different node for each
//
// Upstream recognises exactly two shapes. An assignment whose target is a property of `*.prototype`,
// where the finding is the whole `AssignmentExpression`; and a call to `Object.defineProperty` or
// `Object.defineProperties` whose FIRST argument is `*.prototype`, where the finding is the whole
// `CallExpression`. Measured on eslint 10.8.1:
//
//	Object.prototype.p = 0                                    columns 1 through 23, the assignment
//	Object.defineProperty(Array.prototype, 'p', {value: 0})   columns 1 through 56, the whole call
//
// # What stays silent, each measured rather than reasoned about
//
//	Object.prototype.p++          clean, because upstream matches an assignment and not an update
//	delete Object.prototype.p     clean, same reason
//	Object.prototype = 0          clean, the target is `prototype` itself rather than a property of it
//	Object.prototype.p.q = 0      clean, the assignment target's object is `...p` and not `...prototype`
//	Object.freeze(Array.prototype)  clean, the call is not one of the two define methods
//	Object.defineProperty(x, Array.prototype)  clean, the prototype is not the FIRST argument
//
// The first two read like gaps and they are upstream's, reproduced rather than improved on. The
// logical assignments `&&=`, `||=` and `??=` DO report, and the corpus asserts all three, so the
// assignment test has to accept every assignment operator rather than only `=`.
//
// `Object.defineProperty(Array.prototype)` with no further arguments reports, which is worth stating
// because it looks like an incomplete call the rule should decline: only the first argument is
// examined and the corpus's own `Object.defineProperty()` clean case is clean for having NO first
// argument rather than for being incomplete.
//
// # The shadow question, and why this reads the checker
//
// Upstream asks its scope analysis for the global variable of each builtin name and walks that
// variable's references, so a local binding of the same name never enters the loop. Two of the
// clean cases are exactly that, in two different declaration forms and two different scopes:
//
//	function foo() { var Object = function() {}; Object.prototype.p = 0 }   clean
//	{ let Object = function() {}; Object.prototype.p = 0 }                  clean
//
// `resolvesToAGlobal`, shared with `no-new-native-nonconstructor` and four other rules here, asks the
// equivalent question of the checker. Both clean cases were driven through the installed build and
// through this rule before it was written.
//
// A name that merely evaluates to the builtin is also clean, and that is a real limit of the
// approach rather than an accident: `o = Object; o.prototype.toString = 0` is silent upstream
// because `o` is a different variable, and `global.Object.prototype.toString = 0` is silent because
// the receiver is a member access rather than the global identifier.
var NoExtendNative = rule.Rule{
	Name: "no-extend-native",

	// See the doc above: half the clean cases are a local binding shadowing a builtin name, and
	// nothing structural separates those from the real global.
	NeedsTypeChecker: true,
	TypeReach:        rule.TypeReachShapes,

	Run: func(ctx rule.Context, options any) rule.Listeners {
		// A rule configured as a bare `"error"` is handed NIL rather than a decoded struct, because
		// `DecodeOptionsInto` errors on empty input and the config turns that into nil. The zero
		// value is the right default here, and naming it makes that a decision rather than an
		// accident. `TestNoExtendNativeDefaultsWithoutTheDecoder` bypasses the decoder to pin it.
		settings := NoExtendNativeOptions{}
		if decoded, ok := rule.OptionsAs[NoExtendNativeOptions](options); ok {
			settings = decoded
		}
		excepted := make(map[string]bool, len(settings.Exceptions))
		for _, name := range settings.Exceptions {
			excepted[name] = true
		}

		// builtinExtendedBy answers, for a node that might be `<Builtin>.prototype`, which builtin's
		// prototype it names. The empty string means it names none, which covers a non-global
		// receiver, a shadowed one, an excepted one, and a property that is not `prototype`.
		builtinExtendedBy := func(node *ast.Node) string {
			if node == nil {
				return ""
			}
			// Fidelity to a parser difference rather than a correctness improvement: ESTree has no
			// node for a parenthesis and wraps an optional access in a ChainExpression that upstream
			// reads through. `(Object?.prototype).p = 0` is an upstream corpus case and reports.
			node = ast.SkipParentheses(node)
			if node == nil {
				return ""
			}
			if name, named := property.AccessedName(node, property.Static); !named ||
				name != "prototype" {
				return ""
			}
			receiver := accessedObject(node)
			if receiver == nil {
				return ""
			}
			receiver = ast.SkipParentheses(receiver)
			if receiver == nil || receiver.Kind != ast.KindIdentifier {
				return ""
			}
			name := receiver.Text()
			if !nativePrototypeBuiltins[name] || excepted[name] {
				return ""
			}
			if !resolvesToAGlobal(ctx, receiver) {
				return ""
			}
			return name
		}

		report := func(node *ast.Node, builtin string) {
			ctx.ReportNode(node, rule.Message{
				Id: "unexpected",
				Description: fmt.Sprintf(
					"This adds a property to `%s.prototype`, which every object of that type in "+
						"the program shares, including ones created by code that has never heard "+
						"of this file. The property shows up in every `for...in` over such an "+
						"object, it silently loses to a future standard method of the same name, "+
						"and nothing at the call site says where it came from. Put the function "+
						"somewhere it can be imported instead.", builtin),
			})
		}

		return rule.Listeners{
			// `<Builtin>.prototype.p = 0`. The listener anchors on the assignment so the finding is
			// the whole assignment, which is where upstream points.
			ast.KindBinaryExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				binary := node.AsBinaryExpression()
				// Every assignment operator, not just `=`: the corpus asserts `&&=`, `||=` and
				// `??=` all report.
				if !ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
					return
				}
				target := ast.SkipParentheses(binary.Left)
				if target == nil {
					return
				}
				// The assignment target must be a property OF the prototype, which is what makes
				// `Object.prototype = 0` clean and `Object.prototype.p.q = 0` clean too.
				if builtin := builtinExtendedBy(accessedObject(target)); builtin != "" {
					report(node, builtin)
				}
			},

			// `Object.defineProperty(<Builtin>.prototype, ...)` and `defineProperties`.
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				call := node.AsCallExpression()
				if call.Arguments == nil || len(call.Arguments.Nodes) == 0 {
					return
				}
				if !isObjectDefinePropertyCallee(ctx, call.Expression) {
					return
				}
				// Only the FIRST argument, which is why `Object.defineProperty(x, Array.prototype)`
				// is clean on the installed build.
				if builtin := builtinExtendedBy(call.Arguments.Nodes[0]); builtin != "" {
					report(node, builtin)
				}
			},
		}
	},
}

// isObjectDefinePropertyCallee reports whether a callee names `Object.defineProperty` or
// `Object.defineProperties` on the real global `Object`.
//
// Upstream spells this as `isSpecificMemberAccess(callee, "Object", /^definePropert(?:y|ies)$/u)`,
// which tests the receiver's NAME only. This additionally asks `resolvesToAGlobal` about that
// receiver, which is a deliberate narrowing and it is stated here rather than left to be discovered:
// a local `Object` calling `defineProperty` on a native prototype is a shape upstream reports and
// this declines. No corpus case writes it, and the rule already requires the global for the
// prototype receiver, so accepting a shadowed `Object` on the callee while rejecting one on the
// argument would be the inconsistent reading.
func isObjectDefinePropertyCallee(ctx rule.Context, callee *ast.Node) bool {
	if callee == nil {
		return false
	}
	callee = ast.SkipParentheses(callee)
	if callee == nil {
		return false
	}
	name, named := property.AccessedName(callee, property.Static)
	if !named || (name != "defineProperty" && name != "defineProperties") {
		return false
	}
	receiver := accessedObject(callee)
	if receiver == nil {
		return false
	}
	receiver = ast.SkipParentheses(receiver)
	if receiver == nil || receiver.Kind != ast.KindIdentifier || receiver.Text() != "Object" {
		return false
	}
	return resolvesToAGlobal(ctx, receiver)
}
