package core

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/core"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/cohere/internal/lint/ecmascript/comments"
	"github.com/system-inc/cohere/internal/lint/ecmascript/property"
	"github.com/system-inc/cohere/internal/lint/rule"
)

var messagePreferObjectHasOwn = rule.Message{
	Id: "useHasOwn",
	Description: "Use 'Object.hasOwn()' instead of 'Object.prototype.hasOwnProperty.call()'. " +
		"The long form exists because an object can define its own `hasOwnProperty`, or have a null " +
		"prototype and none at all, so calling it through `Object.prototype` was the only safe " +
		"spelling. `Object.hasOwn` is that spelling with a name, and it cannot be shadowed by the " +
		"object being asked about.",
}

// PreferObjectHasOwn flags a call to `Object.prototype.hasOwnProperty.call()` and its variants.
//
//	valid:   Object.hasOwn(obj, prop)
//	valid:   foo.hasOwnProperty.call(obj, prop)
//	valid:   Object[hasOwnProperty].call(obj, prop)
//	valid:   (Object) => Object.hasOwnProperty.call(obj, prop)
//	invalid: Object.prototype.hasOwnProperty.call(obj, 'foo')
//	invalid: Object.hasOwnProperty.call(obj, property)
//	invalid: ({}).hasOwnProperty.call(obj, 'foo')
//
// # The shape being matched
//
// A call whose callee is a member access reading `call`, whose own object is a member access
// reading `hasOwnProperty`, whose own object is one of three things: the identifier `Object`, a
// member access reading `prototype` on the identifier `Object`, or an EMPTY object literal. That
// last one is why `({ foo }.hasOwnProperty.call(obj, prop))` is clean and `({}.hasOwnProperty.call
// (obj, prop))` is not: a literal with properties in it might be shadowing the very method being
// reached for, so upstream declines to reason about it.
//
// Both property names are read as STATIC names, so the bracket spellings count: `Object
// ['hasOwnProperty']['call']` and the template-literal form both report and both fix. A computed
// name that is an identifier does not, because its value is not settled by the syntax.
//
// # The scope guard, and what it is actually asking
//
// Upstream resolves `Object` in scope and requires the binding it finds to be the GLOBAL one, so a
// parameter or local named `Object` makes the whole thing clean. Three of upstream's valid cases
// are exactly that shadow, one for each of the three left-hand shapes.
//
// This asks the same question through the checker rather than through a scope table: resolve the
// identifier, and require its declaration to live in a declaration file. Probed on all four shapes
// before being built on: the real global comes back with two declarations in `lib.es5.d.ts`, while
// an arrow parameter, a function parameter and a block-scoped const each come back with one
// declaration in the source file being linted.
//
// That is a narrower question than upstream's `variable.scope.type === "global"`, and the
// difference has a live consequence this port does NOT reproduce. Upstream's guard also fails when
// the global is turned off, which its corpus writes as `/* global Object: off */`. Our
// configuration has no such directive and the checker has nothing to answer it with, so that case
// is recorded as reporting rather than silently made clean. It is in the fixtures, marked, with the
// reasoning at the line.
//
// # The fixer, and the two things it refuses
//
// The repair replaces the whole callee with `Object.hasOwn`, which turns any of the six accepted
// spellings into the one short form. Upstream declines in two situations and both are reproduced:
//
//	a comment INSIDE the callee     `Object/* comment */.prototype.hasOwnProperty.call(a, b)`
//	                                is upstream's only `output: null` case. The comment sits
//	                                between the tokens being replaced, so any repair drops it.
//	a token that would merge        `return{}.hasOwnProperty.call(a, b)` is written with no space,
//	                                and replacing the callee with `Object.hasOwn` would produce
//	                                `returnObject.hasOwn`, which is a different program. Upstream
//	                                inserts a leading space; so does this.
//
// The second decline has a companion that is easy to get backwards. A COMMENT before the callee
// already separates the tokens, so `return/*comment*/{}.hasOwnProperty.call(a, b)` needs no space
// and upstream does not add one. Four of its cases pair the two spellings for that comparison.
var PreferObjectHasOwn = rule.Rule{
	Name:             "prefer-object-has-own",
	NeedsTypeChecker: true,
	Run: func(ctx rule.Context, options any) rule.Listeners {
		return rule.Listeners{
			ast.KindCallExpression: func(node *ast.Node) {
				if ctx.TypeChecker == nil {
					return
				}
				callee := preferObjectHasOwnUnwrap(node.AsCallExpression().Expression)
				if !preferObjectHasOwnIsMemberAccess(callee) {
					return
				}
				calleeObject := preferObjectHasOwnUnwrap(preferObjectHasOwnAccessedObject(callee))
				if !preferObjectHasOwnIsMemberAccess(calleeObject) {
					return
				}

				if name, isStatic := property.AccessedName(callee, property.Static); !isStatic ||
					name != "call" {
					return
				}
				if name, isStatic := property.AccessedName(calleeObject, property.Static); !isStatic ||
					name != "hasOwnProperty" {
					return
				}
				if !preferObjectHasOwnHasLeftHandObject(ctx, calleeObject) {
					return
				}

				if fix, canFix := preferObjectHasOwnFix(ctx, callee); canFix {
					ctx.ReportNodeWithFixes(node, messagePreferObjectHasOwn, fix)
					return
				}
				ctx.ReportNode(node, messagePreferObjectHasOwn)
			},
		}
	},
}

// preferObjectHasOwnIsMemberAccess answers whether this node reads a property off something, in
// either spelling. Both arms matter: the bracket forms are half of upstream's invalid cases.
func preferObjectHasOwnIsMemberAccess(node *ast.Node) bool {
	return node != nil && (node.Kind == ast.KindPropertyAccessExpression ||
		node.Kind == ast.KindElementAccessExpression)
}

// preferObjectHasOwnAccessedObject returns what a member access reads FROM, or nil.
func preferObjectHasOwnAccessedObject(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindPropertyAccessExpression:
		return node.AsPropertyAccessExpression().Expression
	case ast.KindElementAccessExpression:
		return node.AsElementAccessExpression().Expression
	}
	return nil
}

// preferObjectHasOwnHasLeftHandObject is upstream's `hasLeftHandObject`, which accepts three
// shapes for the thing `hasOwnProperty` is being read off.
//
// Parentheses are unwrapped in a loop, and that is not an improvement on upstream: its parser folds
// them away, so `(( Object )).prototype.hasOwnProperty.call(a, b)` reaches its check as a bare
// identifier already. Six of its invalid cases write exactly that shape, so without the unwrap this
// port would go silent on inputs the corpus asserts. `((x))` nests, which is why this is a loop.
func preferObjectHasOwnHasLeftHandObject(ctx rule.Context, memberAccess *ast.Node) bool {
	objectNode := preferObjectHasOwnAccessedObject(memberAccess)
	for objectNode != nil && objectNode.Kind == ast.KindParenthesizedExpression {
		objectNode = objectNode.AsParenthesizedExpression().Expression
	}
	if objectNode == nil {
		return false
	}

	// An EMPTY object literal. A literal with properties is declined, because one of them could be
	// a `hasOwnProperty` of its own and upstream will not reason about that.
	//
	// This arm still has to answer the scope question, and it is the arm where forgetting to is
	// most tempting: there is no `Object` identifier anywhere in `({}).hasOwnProperty.call(a, b)`,
	// so it reads as though scope cannot be relevant. Upstream checks it for all three shapes and
	// writes a passing case for exactly this one, because the FIX writes `Object.hasOwn` and under
	// a shadow that names something else entirely. Answered by resolving the name the repair would
	// introduce, at the position it would be introduced.
	if objectNode.Kind == ast.KindObjectLiteralExpression {
		if len(objectNode.AsObjectLiteralExpression().Properties.Nodes) != 0 {
			return false
		}
		return preferObjectHasOwnGlobalObjectIsInScopeAt(ctx, objectNode)
	}

	// `Object.prototype`, in either spelling, unwraps one more level to the `Object` beneath it.
	nodeToCheck := objectNode
	if preferObjectHasOwnIsMemberAccess(objectNode) {
		if name, isStatic := property.AccessedName(objectNode, property.Static); isStatic &&
			name == "prototype" {
			nodeToCheck = preferObjectHasOwnAccessedObject(objectNode)
			for nodeToCheck != nil && nodeToCheck.Kind == ast.KindParenthesizedExpression {
				nodeToCheck = nodeToCheck.AsParenthesizedExpression().Expression
			}
		}
	}

	if nodeToCheck == nil || nodeToCheck.Kind != ast.KindIdentifier ||
		nodeToCheck.Text() != "Object" {
		return false
	}
	return preferObjectHasOwnResolvesToTheGlobalObject(ctx, nodeToCheck)
}

// preferObjectHasOwnUnwrap strips parentheses in a loop.
//
// Not an improvement on upstream: its parser folds parentheses away, so a callee written
// `(( Object.prototype.hasOwnProperty.call ))` reaches its listener as a bare member access
// already. Five of its invalid cases write that shape at the callee or one level in, so without
// this the port goes silent on inputs the corpus asserts. `((x))` nests, which is why it loops.
//
// Written out rather than using ast.SkipParentheses, which dereferences its argument and would
// panic on the optional nodes this is handed.
func preferObjectHasOwnUnwrap(node *ast.Node) *ast.Node {
	for node != nil && node.Kind == ast.KindParenthesizedExpression {
		node = node.AsParenthesizedExpression().Expression
	}
	return node
}

// preferObjectHasOwnGlobalObjectIsInScopeAt answers whether the name `Object` at this position
// resolves to the global, for the empty-literal shape where no such identifier is written.
//
// The question is about the code the FIX would produce rather than the code that is there, so it is
// asked by resolving the name at the site the replacement would occupy.
func preferObjectHasOwnGlobalObjectIsInScopeAt(ctx rule.Context, at *ast.Node) bool {
	symbol := ctx.TypeChecker.ResolveName("Object", at, ast.SymbolFlagsValue, false)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	for _, declaration := range symbol.Declarations {
		declaringFile := ast.GetSourceFileOfNode(declaration)
		if declaringFile == nil || !declaringFile.IsDeclarationFile {
			return false
		}
	}
	return true
}

// preferObjectHasOwnResolvesToTheGlobalObject answers upstream's scope question through the
// checker: is this `Object` the global one, or a binding shadowing it?
//
// A declaration in a declaration file is the global. Deliberately NOT written as
// `resolvesToAGlobal`'s complement, because that helper is the shipped exemplar for this question
// and its shape is what is wanted here: a name with zero declarations is not something this rule
// should report on, since `Object` always has declarations when the library is loaded.
func preferObjectHasOwnResolvesToTheGlobalObject(ctx rule.Context, identifier *ast.Node) bool {
	symbol := ctx.TypeChecker.GetSymbolAtLocation(identifier)
	if symbol == nil || len(symbol.Declarations) == 0 {
		return false
	}
	// Every declaration must be ambient rather than just the first.
	//
	// Measured equivalent on this shape and kept anyway, with the measurement recorded so the next
	// reader does not have to retake it. Probed over five ways to write `Object` in source beside
	// the global: a var, a function, a class and a namespace each RESOLVE THE VALUE POSITION to a
	// symbol carrying only their own source declaration, so the global is replaced rather than
	// merged with. The one shape that does merge, `interface Object`, contributes a TYPE, so the
	// value position still comes back with two ambient declarations and no source one. No probed
	// input produces a mixed list here, so the loop and an index-zero read cannot disagree.
	//
	// The loop stands because the verdict is about what the checker does today rather than about
	// the grammar, and because index-zero is the pattern this tree has been bitten by before. If a
	// merging shape is ever found, this is already right and the note is what expires.
	for _, declaration := range symbol.Declarations {
		declaringFile := ast.GetSourceFileOfNode(declaration)
		if declaringFile == nil || !declaringFile.IsDeclarationFile {
			return false
		}
	}
	return true
}

// preferObjectHasOwnFix builds the repair, or declines.
//
// The replacement text is the whole callee rather than a surgical edit, because the six accepted
// spellings share no common structure and all collapse to the same eleven characters.
func preferObjectHasOwnFix(ctx rule.Context, callee *ast.Node) (rule.Fix, bool) {
	calleeRange := rule.TokenRange(ctx.SourceFile, callee)

	// A comment inside the callee would be deleted by the replacement, so upstream declines. This
	// is its only `output: null` case and the decline is the port rather than a detail.
	for _, comment := range comments.ForFile(ctx) {
		if comment.Range.Pos() >= calleeRange.Pos() && comment.Range.End() <= calleeRange.End() {
			return rule.Fix{}, false
		}
	}

	// If the character immediately before the callee could merge with the `O` of `Object`, the
	// replacement needs a space in front of it. `return{}.hasOwnProperty.call(a, b)` becomes
	// `returnObject.hasOwn(a, b)` without one, which is a different program that still parses.
	//
	// Upstream asks this by re-tokenizing, which it has to do because its right-hand side is a
	// caller-supplied string. Here the right-hand side is always `Object.hasOwn`, whose first token
	// is an identifier, so the question is only whether the character before is one an identifier
	// would absorb.
	replacement := "Object.hasOwn"
	if calleeRange.Pos() > 0 {
		previous := ctx.SourceFile.Text()[calleeRange.Pos()-1]
		if scanner.IsIdentifierPart(rune(previous)) {
			replacement = " " + replacement
		}
	}
	return rule.ReplaceRange(core.NewTextRange(calleeRange.Pos(), calleeRange.End()), replacement), true
}
