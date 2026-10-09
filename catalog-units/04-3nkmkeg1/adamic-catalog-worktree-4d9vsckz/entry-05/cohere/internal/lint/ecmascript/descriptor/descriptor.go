// Package descriptor answers whether a function is the `get` or `set` of a property descriptor passed
// to one of the platform methods that read descriptors: `Object.defineProperty`,
// `Object.defineProperties`, `Object.create` and `Reflect.defineProperty`.
//
// Lifted out of `getter-return` because `no-setter-return` asks the same question about `set`, and a
// rule package may not import another. ESLint's two rules share it the same way, through
// `astUtils`. Before the lift no-setter-return left descriptors out entirely, because the port it
// followed did, and ESLint's 23 descriptor rows read as missing (#jjfa7qb).
package descriptor

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
)

// KeyName returns the key a property names when that is knowable statically.
//
// The accept set is the identifier, string and template spellings, bare or inside the brackets of a
// computed key, which is what ESLint's `getStaticPropertyName` reads. A descriptor's key is nearly
// always the plain identifier `get` or `set`; ESLint's corpus writes `'set'`, `['set']` and
// “ [`set`] “ as well. Numerics are excluded because this compares against fixed non-numeric names,
// so accepting them could change no verdict.
//
// A computed key naming a variable is declined by the shelf whatever the accept set says, since the
// property it names is whatever the variable holds.
func KeyName(name *ast.Node) string {
	text, _ := property.Name(name, property.Named|property.Quoted|property.Templated|property.Computed)
	return text
}

// IsFunctionUnder answers whether this function-like node is the value of the `key` property (`get` or
// `set`) of a property descriptor object.
//
// Two spellings reach one: `{ get: function () {} }` puts the function under a PropertyAssignment, and
// `{ get() {} }` makes it a method of the object literal itself.
//
// isGlobal says whether the identifier naming `Object` or `Reflect` in the call is the global rather
// than a local binding that shares its spelling. ESLint's corpus pins `let Object;
// Object.defineProperty(...)` as clean. A rule without the checker passes a function answering true,
// which reads the spelling alone.
//
// The descriptor must sit at the argument its method reads it from: the third of `defineProperty`, the
// second of `defineProperties` and `create`. ESLint's getter-return and no-setter-return share one
// `isPropertyDescriptor` that checks this, so `Object.defineProperty({ get() {} }, 'k', d)` is clean in
// both.
func IsFunctionUnder(node *ast.Node, key string, isGlobal func(identifier *ast.Node) bool) bool {
	parent := node.Parent
	if parent == nil {
		return false
	}

	var propertyName string
	var propertyHolder *ast.Node
	switch {
	case parent.Kind == ast.KindPropertyAssignment:
		propertyName = KeyName(parent.AsPropertyAssignment().Name())
		propertyHolder = parent.Parent
	case node.Kind == ast.KindMethodDeclaration && parent.Kind == ast.KindObjectLiteralExpression:
		propertyName = KeyName(node.AsMethodDeclaration().Name())
		propertyHolder = parent
	default:
		return false
	}

	if propertyName != key || propertyHolder == nil || propertyHolder.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	return isDescriptorObject(propertyHolder, isGlobal)
}

// isDescriptorObject answers whether this object literal is being passed somewhere that treats it as a
// property descriptor.
//
// Two arrangements reach one, and each belongs to its own methods, as ESLint pairs them:
//
//	Object.defineProperty(o, "k", DESC)          the descriptor is the call's third argument
//	Reflect.defineProperty(o, "k", DESC)         same
//	Object.defineProperties(o, { k: DESC })      the descriptor is a property of the second argument
//	Object.create(o, { k: DESC })                same
//
// So `Object.create(o, { set() {} })` is not a descriptor: the object create reads holds descriptors,
// it is not one. Counting kinds rather than parents is what is done here, because the parent count
// changes with parenthesization while the shape does not.
func isDescriptorObject(descriptor *ast.Node, isGlobal func(identifier *ast.Node) bool) bool {
	argument := descriptor
	nested := false
	// `{ k: { get: fn } }`: step out through the property to the object literal that is the
	// argument. One step only, since no watched method nests deeper than this.
	if container := descriptor.Parent; container != nil && container.Kind == ast.KindPropertyAssignment {
		argument = container.Parent
		if argument == nil || argument.Kind != ast.KindObjectLiteralExpression {
			return false
		}
		nested = true
	}

	call := skipParenthesesUpward(argument.Parent)
	if call == nil || call.Kind != ast.KindCallExpression {
		return false
	}
	object, member := descriptorCallee(call.AsCallExpression().Expression)
	if object == nil {
		return false
	}
	switch {
	case nested && object.Text() == "Object" && (member == "defineProperties" || member == "create"):
	case !nested && (object.Text() == "Object" || object.Text() == "Reflect") && member == "defineProperty":
	default:
		return false
	}
	wanted := 2
	if nested {
		wanted = 1
	}
	arguments := call.AsCallExpression().Arguments.Nodes
	if len(arguments) <= wanted || skipParenthesesDownward(arguments[wanted]) != argument {
		return false
	}
	return isGlobal(object)
}

// skipParenthesesUpward walks out of any parentheses wrapping a node.
//
// `(Object?.defineProperty)(o, "k", { get: fn })` is three of ESLint's rows, and without this the
// parenthesized callee shifts every parent index and the descriptor stops being found.
func skipParenthesesUpward(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	return node
}

// descriptorCallee reads a callee as `Object.member` or `Reflect.member`, answering the identifier and
// the member's name, or nil when it is anything else.
//
// The member may be written with a dot or as a static computed key, `Object['defineProperty']`, which
// ESLint reads through `getStaticPropertyName` and its corpus writes both ways. Which members count is
// the caller's question; `foo.defineProperty(...)` is never one, because a method with the right name
// on the wrong object is not the platform builtin.
func descriptorCallee(callee *ast.Node) (*ast.Node, string) {
	callee = skipParenthesesDownward(callee)
	if callee == nil {
		return nil, ""
	}
	var object *ast.Node
	var member string
	switch callee.Kind {
	case ast.KindPropertyAccessExpression:
		access := callee.AsPropertyAccessExpression()
		object = access.Expression
		member = KeyName(access.Name())
	case ast.KindElementAccessExpression:
		access := callee.AsElementAccessExpression()
		object = access.Expression
		argument := skipParenthesesDownward(access.ArgumentExpression)
		if argument == nil || (argument.Kind != ast.KindStringLiteral && argument.Kind != ast.KindNoSubstitutionTemplateLiteral) {
			return nil, ""
		}
		member = argument.Text()
	default:
		return nil, ""
	}
	object = skipParenthesesDownward(object)
	if object == nil || object.Kind != ast.KindIdentifier || (object.Text() != "Object" && object.Text() != "Reflect") {
		return nil, ""
	}
	return object, member
}

// skipParenthesesDownward unwraps parentheses around an expression.
func skipParenthesesDownward(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}
