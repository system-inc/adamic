package outputflow

import (
	"slices"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
)

// IsRoot reports whether node owns a control-flow graph of its own.
func IsRoot(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindSourceFile,
		ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction,
		ast.KindMethodDeclaration, ast.KindConstructor,
		ast.KindGetAccessor, ast.KindSetAccessor,
		ast.KindClassStaticBlockDeclaration:
		return true
	case ast.KindPropertyDeclaration:
		return node.AsPropertyDeclaration().Initializer != nil
	}
	return false
}

// RootOf returns the code path root node executes in.
func RootOf(node *ast.Node) *ast.Node {
	previous := node
	var decorator *ast.Node
	for current := node.Parent; current != nil; current = current.Parent {
		if current.Kind == ast.KindDecorator {
			decorator = current
		}
		if current.Kind == ast.KindSourceFile {
			return current
		}
		if IsRoot(current) && runsInsideRoot(current, previous, decorator) {
			return current
		}
		previous = current
	}
	return nil
}

// runsInsideRoot reports whether a direct child of a code path root runs in
// that root rather than in the surrounding one. A member's name runs where the
// member is declared, and so do the decorators on the member and on each of its
// parameters.
func runsInsideRoot(root *ast.Node, child *ast.Node, decorator *ast.Node) bool {
	switch root.Kind {
	case ast.KindPropertyDeclaration:
		return root.AsPropertyDeclaration().Initializer == child
	case ast.KindClassStaticBlockDeclaration:
		return root.AsClassStaticBlockDeclaration().Body == child
	}
	if root.Name() == child {
		return false
	}
	if decorator != nil && (decorator.Parent == root || decorator.Parent == child) {
		return false
	}
	return true
}

// RootSummary is one code path root and what its own code holds, so a rule can tell whether the
// root's graph could give it anything to report before paying to build it.
//
// "Own code" stops at a nested root, which the builder lays out as a graph of its own and never
// descends into. A statement always runs in its nearest enclosing root, since a statement can only
// sit in a function body, a class static block or the file, so the nearest one is what is recorded.
type RootSummary struct {
	Node *ast.Node

	// BareReturn is whether the root's own code holds a `return;` with no value.
	BareReturn bool

	// Loops are the loop statement kinds the root's own code holds, each once, in the order first met.
	Loops []ast.Kind
}

// IndexRoots lists every code path root in a file, the file first and the rest in preorder, which is
// the order a rule walking the tree for IsRoot meets them in, with what each one's own code holds.
//
// One walk serves every rule that builds a graph per root. Three such rules each walked the file to
// list its roots and then built a graph for every one, though each can only report from the few
// roots holding what it looks for: on a cold ahra run, 4.4% of roots hold a bare return and 7.6% a
// loop (#tmn9n27).
func IndexRoots(sourceFile *ast.Node) []RootSummary {
	roots := []RootSummary{{Node: sourceFile}}
	var visit func(node *ast.Node, owner int)
	visit = func(node *ast.Node, owner int) {
		node.ForEachChild(func(child *ast.Node) bool {
			current := owner
			if IsRoot(child) {
				roots = append(roots, RootSummary{Node: child})
				current = len(roots) - 1
			}
			switch child.Kind {
			case ast.KindReturnStatement:
				if child.AsReturnStatement().Expression == nil {
					roots[current].BareReturn = true
				}
			case ast.KindWhileStatement, ast.KindDoStatement, ast.KindForStatement, ast.KindForInStatement,
				ast.KindForOfStatement:
				if !slices.Contains(roots[current].Loops, child.Kind) {
					roots[current].Loops = append(roots[current].Loops, child.Kind)
				}
			}
			visit(child, current)
			return false
		})
	}
	visit(sourceFile, 0)
	return roots
}
