// Package scope answers which construct encloses a node.
//
// Four rules walked parents looking for a function-like kind, with four different kind sets, and a
// census measured them disagreeing on five shapes: a getter, a setter, a constructor, a static
// block, and module scope. None of the disagreements could reach a verdict, which is stated here
// rather than left for the next reader to rediscover, because the sets were guesses that happened
// not to matter rather than decisions anyone made.
package scope

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// EnclosingFunctionLike returns the nearest ancestor that introduces a function scope, or nil at
// module scope.
//
// Accessors and constructors count, because they introduce a scope exactly as a method does. A
// caller asking the narrower question "which named function is this in" wants
// `EnclosingNamedFunction`, which is a different question rather than a subset: a getter has a name
// and is still not a function anyone calls by it.
//
// # What this deliberately does NOT answer
//
// A static block and the source file are absent. `no-useless-assignment` includes both, because it
// is asking which flow graph a node belongs to rather than which function encloses it: its set is
// the bodies the binder gives a fresh start node, so it is a boundary its reachability walk cannot
// cross. That is a different question wearing a similar name, and its own doc says so. Merging the
// two would give one of them the wrong answer at module scope, where this correctly reports nil and
// a flow walk must report the file.
//
// # The five shapes the four lifted sets disagreed on, and why none mattered
//
// Measured with a marker identifier in each position:
//
//	getter, setter, constructor   the four-kind set answered nil, the others answered the accessor
//	static block                  only the nine-kind set answered
//	module scope                  only the nine-kind set answered, with the source file
//
// The four-kind set's caller passes the result straight to a name predicate that answers "" for an
// accessor and then requires a component name, so a widened walk returns the accessor and is
// rejected one line later. Driven at the rule with a control that fires: a destructure inside a
// getter, a setter and a constructor reports nothing either way, while the same destructure inside a
// component reports once. So the narrow set was not protecting anything, and the wide set does not
// break anything.
func EnclosingFunctionLike(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	for current := node.Parent; current != nil; current = current.Parent {
		switch current.Kind {
		case ast.KindFunctionDeclaration,
			ast.KindFunctionExpression,
			ast.KindArrowFunction,
			ast.KindMethodDeclaration,
			ast.KindGetAccessor,
			ast.KindSetAccessor,
			ast.KindConstructor:
			return current
		}
	}
	return nil
}

// NameOf returns the name a function-like node is known by, or "" when it has none.
//
// An anonymous function expression or arrow assigned to a variable or an object property takes that
// name, which is how `const useThing = () => {}` is recognized as a hook. Without it every
// arrow-bodied function in a modern codebase would be anonymous and the rules built on this would
// see none of them.
//
// Only the immediate parent supplies a borrowed name. A function nested inside an object passed to
// something else does not inherit that thing's name, and walking further would attribute a callback
// to whatever declaration happened to enclose it.
//
// An accessor and a constructor answer "" rather than their property name, matching the four rules
// this was lifted from. A getter named `p` is not a function anyone calls `p`, so a rule asking
// "is this function named like a component" must not be handed one.
func NameOf(node *ast.Node) string {
	if node == nil {
		return ""
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		if name := node.AsFunctionDeclaration().Name(); name != nil {
			return name.Text()
		}

	case ast.KindMethodDeclaration:
		if name := node.AsMethodDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
			return name.Text()
		}

	case ast.KindFunctionExpression:
		if name := node.AsFunctionExpression().Name(); name != nil {
			return name.Text()
		}
		return borrowedName(node)

	case ast.KindArrowFunction:
		return borrowedName(node)
	}
	return ""
}

// EnclosingNamedFunction walks to the nearest function-like ancestor that has a name, and returns
// it with that name.
//
// Different from calling `EnclosingFunctionLike` and then `NameOf`: that pair stops at the first
// function whether or not it has a name, and answers "" for an anonymous one. This continues
// outward. The distinction matters wherever the nearest function is the anonymous callback passed to
// a hook, which is most of the call sites: stopping there would silence a rule everywhere it is
// meant to fire.
//
// A caller that must stop at the first function regardless, matching an original that does, should
// use the pair instead. Both shapes exist in the rules this was lifted from.
func EnclosingNamedFunction(node *ast.Node) (*ast.Node, string) {
	if node == nil {
		return nil, ""
	}
	for current := node.Parent; current != nil; current = current.Parent {
		if name := NameOf(current); name != "" {
			return current, name
		}
	}
	return nil, ""
}

// BodyOf returns a function-like node's body, or nil when it has none.
//
// An accessor and a constructor answer, unlike `NameOf`, because a body is a body whatever the
// construct is called. The asymmetry is deliberate: a caller wanting to search a body has no reason
// to care whether the enclosing thing has a callable name.
func BodyOf(node *ast.Node) *ast.Node {
	if node == nil {
		return nil
	}
	switch node.Kind {
	case ast.KindFunctionDeclaration:
		return node.AsFunctionDeclaration().Body
	case ast.KindFunctionExpression:
		return node.AsFunctionExpression().Body
	case ast.KindArrowFunction:
		return node.AsArrowFunction().Body
	case ast.KindMethodDeclaration:
		return node.AsMethodDeclaration().Body
	case ast.KindGetAccessor:
		return node.AsGetAccessorDeclaration().Body
	case ast.KindSetAccessor:
		return node.AsSetAccessorDeclaration().Body
	case ast.KindConstructor:
		return node.AsConstructorDeclaration().Body
	}
	return nil
}

// borrowedName reads the variable or property name an anonymous function is assigned to.
func borrowedName(node *ast.Node) string {
	parent := node.Parent
	if parent == nil {
		return ""
	}

	switch parent.Kind {
	case ast.KindVariableDeclaration:
		if name := parent.AsVariableDeclaration().Name(); name != nil && name.Kind == ast.KindIdentifier {
			return name.Text()
		}

	case ast.KindPropertyAssignment:
		if name := parent.AsPropertyAssignment().Name(); name != nil && name.Kind == ast.KindIdentifier {
			return name.Text()
		}
	}
	return ""
}
