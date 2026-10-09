package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messageNoProto = rule.Message{
	Id: "unexpectedProto",
	Description: "This reads or writes `__proto__`, which reaches an object's prototype through a " +
		"deprecated accessor. It is normative only for web compatibility, it is slow because " +
		"writing it deoptimizes the object's shape, and on an object parsed from untrusted input " +
		"it is the prototype pollution vector. Use `Object.getPrototypeOf` and " +
		"`Object.setPrototypeOf`, or `Object.create(null)` for a map that has no prototype to reach.",
}

// NoProto flags a property access naming `__proto__`.
//
//	valid:   var a = test[__proto__];
//	valid:   var __proto__ = null;
//	valid:   foo[`__proto`] = null;
//	invalid: var a = test.__proto__;
//	invalid: var a = test['__proto__'];
//	invalid: var a = test[`__proto__`];
//
// # The predicate is the static property name, and the clean cases say why
//
// Upstream asks the member expression for its static property name and compares it against one
// fixed string. That answers alike for the dotted form, a string-literal subscript, and a
// no-substitution template, and answers nothing for a computed access whose key is a variable. Each
// of upstream's clean cases fails a different way if the rule is written as a text match:
// `test[__proto__]` reads a *variable* of that name, so the property is whatever it holds; a
// declaration binding the name is not an access at all; and `foo[`__proto__\n`]` is a different key
// whose text merely contains the name.
//
// This is the same shape as `no-iterator` one file over, and it reaches the same shelf helper. The
// two rules are a matched pair upstream as well, both being pre-standard property names kept alive
// only by the web's refusal to break.
//
// # What this deliberately does not catch, measured rather than assumed
//
// The rule anchors on member access alone, so a `__proto__` written as an object-literal key or a
// class field is silent even though the literal key is the form that actually sets a prototype.
// That reads like an oversight and it is upstream's decision. Measured against eslint 10.8.1:
//
//	({ __proto__: 1 });          clean
//	({ [`__proto__`]: 1 });      clean
//	class C { __proto__ = 1; }   clean
//
// Core ships `no-proto-builtins`-adjacent coverage of the literal form nowhere; the object-literal
// spelling is left to `no-dupe-keys` and to nothing else. Reproduced as silence because fidelity is
// the authority here, not what the rule looks like it should do.
//
// Optional chaining reports: `test?.__proto__` and `test?.[`__proto__`]` both fired on the same
// build. Our parser hangs the question off the same two access kinds, so no arm is needed for it.
var NoProto = rule.Rule{
	Name: "no-proto",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			// Both access kinds, because the shelf helper answers for both and a rule watching only
			// the dotted form is silent on `test['__proto__']`, which upstream reports.
			ast.KindPropertyAccessExpression: reportIfProtoAccess(ctx),
			ast.KindElementAccessExpression:  reportIfProtoAccess(ctx),
		}
	},
}

// reportIfProtoAccess reports an access whose statically-known property is `__proto__`.
//
// `property.Static` rather than `property.Textual`, because upstream resolves a no-substitution
// template subscript and `Textual` declines templates. The set is wider than the comparison needs in
// one direction that cannot matter: no numeric key renders as `__proto__`.
//
// A private name cannot collide either, and that is a property of the shelf rather than of this
// comparison: `Text()` on a `KindPrivateIdentifier` carries the leading hash, so `#__proto__`
// answers `"#__proto__"`. Upstream's `class C { #__proto__; ... }` clean case pins it.
func reportIfProtoAccess(ctx rule.Context) func(*ast.Node) {
	return func(node *ast.Node) {
		name, _ := property.AccessedName(node, property.Static)
		if name != "__proto__" {
			return
		}
		ctx.ReportNode(node, messageNoProto)
	}
}
