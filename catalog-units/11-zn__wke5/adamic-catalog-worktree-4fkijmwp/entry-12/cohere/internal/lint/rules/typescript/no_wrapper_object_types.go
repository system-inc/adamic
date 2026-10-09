package typescript

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/internal/lint/ecmascript/module"
	"github.com/system-inc/cohere/internal/lint/rule"
)

// wrapperObjectTypes are the six object types that wrap a primitive.
//
// Each names a boxed object rather than the primitive value: `String` is the object `new String("x")`
// produces, `string` is the value `"x"` is. Annotating with the wrapper accepts the object and
// rejects the primitive under some operations, which is the reverse of what almost every author
// means, and the two read almost identically at a glance.
//
// `Function` is a seventh member of the same family and lives in its own rule upstream, so it stays
// in its own rule here.
var wrapperObjectTypes = map[string]string{
	"BigInt":  "bigint",
	"Boolean": "boolean",
	"Number":  "number",
	"Object":  "object",
	"String":  "string",
	"Symbol":  "symbol",
}

var messageBannedWrapperObjectType = rule.Message{
	Id: "bannedWrapperObjectType",
	Description: "This names the boxed object rather than the primitive: `String` is what " +
		"`new String('x')` produces, and `string` is what `'x'` is. The wrapper accepts the object " +
		"and behaves differently from the primitive under comparison and truthiness, which is the " +
		"reverse of what the annotation almost always means. Use the lowercase primitive.",
}

// NoWrapperObjectTypes flags `BigInt`, `Boolean`, `Number`, `Object`, `String`, and `Symbol` in a
// type position.
//
//	valid:   let value: number
//	valid:   let value: NumberLike
//	valid:   class Thing extends Number {}
//	invalid: let value: Number
//	invalid: interface Thing extends Number {}
//	invalid: class Thing implements Number {}
//
// Ported from oxc's `no_wrapper_object_types`, which is what the gate runs.
//
// The `extends` case is the one worth reading, and it came from oxc's own corpus rather than from
// reasoning: `class Thing extends Number {}` is valid and `class Thing implements Number {}` is not.
// A class extends a value, and `Number` the value is a real constructor, so extending it is legal
// TypeScript that means what it says. A class implements a *type*, which puts the wrapper back in a
// type position.
//
// typescript-go draws that line in the parse, which took a probe to learn rather than a reading:
//
//	class C extends Number     ExpressionWithTypeArguments
//	interface I extends Number TypeReference
//	class C implements Number  TypeReference
//
// So listening to type references alone gets the distinction for free, and an explicit exemption for
// class-extends is unreachable. One was written here first, from the assumption that all three
// heritage forms parse alike, and removing it changed no fixture. That assumption also produced a
// dead listener in `no-unsafe-function-type` earlier, in the opposite direction: there the original
// registered three visitors where one suffices, here the rule guarded a case that never arrives.
//
// A fix is offered only in a plain type-reference position. oxc makes the same distinction: replacing
// `Number` with `number` inside `implements` would produce code that does not compile, since a class
// cannot implement a primitive.
var NoWrapperObjectTypes = rule.Rule{
	Name: "@typescript-eslint/no-wrapper-object-types",
	Run: func(ctx rule.Context, options any) rule.Listeners {
		if ctx.SourceFile == nil {
			return nil
		}

		// The same shadow guard as `no-unsafe-function-type`, and for the same reason: a file that
		// declares its own `Number` has redefined the name, and oxc's corpus has `type Number = 0 | 1;
		// let value: Number;` as a passing case. The walk is whole-file because a declaration nested
		// in a block is still a declaration, which is a false positive this package already shipped
		// once.
		wrapperNames := make(map[string]bool, len(wrapperObjectTypes))
		for name := range wrapperObjectTypes {
			wrapperNames[name] = true
		}
		shadowed := module.DeclaredTypeNames(ctx.SourceFile, wrapperNames)

		return rule.Listeners{
			ast.KindTypeReference: func(node *ast.Node) {
				typeReference := node.AsTypeReferenceNode()
				if typeReference == nil {
					return
				}

				name := typeReference.TypeName
				if name == nil || name.Kind != ast.KindIdentifier {
					return
				}

				primitive, isWrapper := wrapperObjectTypes[name.Text()]
				if !isWrapper || shadowed[name.Text()] {
					return
				}

				if isImplementsHeritage(node) {
					// No fix. `class Thing implements number {}` does not compile, so the repair a
					// reader would expect is not available here.
					ctx.ReportNode(name, messageBannedWrapperObjectType)
					return
				}

				ctx.ReportNodeWithFixes(name, messageBannedWrapperObjectType, ctx.ReplaceNode(name, primitive))
			},
		}
	},
}

// isImplementsHeritage reports whether this type reference is the target of an `implements` clause.
func isImplementsHeritage(node *ast.Node) bool {
	clause, isHeritage := enclosingHeritageClause(node)
	return isHeritage && clause.Token == ast.KindImplementsKeyword
}

// enclosingHeritageClause finds the heritage clause a type reference sits in, if any.
//
// The walk is bounded: a heritage clause is at most two levels above its type reference, so a deeper
// ancestor is a different construct that happens to contain one.
func enclosingHeritageClause(node *ast.Node) (*ast.HeritageClause, bool) {
	for ancestor, depth := node.Parent, 0; ancestor != nil && depth < 2; ancestor, depth = ancestor.Parent, depth+1 {
		if ancestor.Kind == ast.KindHeritageClause {
			return ancestor.AsHeritageClause(), true
		}
	}
	return nil, false
}
