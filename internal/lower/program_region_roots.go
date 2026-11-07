package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// The Program root, chosen by structure and never by name (#64pg02g). A run that builds a second
// Program begins a new generation at the entry of each function that allocates the root, so only
// the first Program's sites adopt and later ones fall back to their graph region or count
// (program_region.c). Membership names no root: it's every owning component the selection keeps.
// The root is a top of their condensation, a component no owning edge enters from a selected type
// outside it. Containers and type views (unions, intersections) are never sources: an edge is
// followed through them to the component it reaches. So a Program[] that merely reaches Program
// doesn't hide Program's component, while a holder that keeps Programs in an array still holds
// them. Where a begin lands decides retention only, never safety: an extra one shuts the latch
// early and a missed one keeps a later Program until exit.

// programRegionTops is every member type in a top component, by type id.
func (f *cycleFinder) programRegionTops(graph programRegionOwning, members ir.SortedMap[int, bool]) map[int]bool {
	transparent := func(node cycleNode) bool {
		return node.proven != nil && (f.programContainer(node.proven) || node.proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0)
	}
	component := map[cycleNode]int{}
	for index, nodes := range graph.components {
		for node := range nodes {
			if !transparent(node) {
				component[node] = index
			}
		}
	}
	entered := make([]bool, len(graph.components))
	for node, index := range component {
		seen := map[cycleNode]bool{}
		queue := append([]cycleNode{}, graph.owning[node]...)
		for len(queue) != 0 {
			target := queue[len(queue)-1]
			queue = queue[:len(queue)-1]
			if seen[target] {
				continue
			}
			seen[target] = true
			if reached, ok := component[target]; ok {
				if reached != index {
					entered[reached] = true
				}
				continue
			}
			if transparent(target) {
				queue = append(queue, graph.owning[target]...)
			}
		}
	}
	tops := map[int]bool{}
	for node, index := range component {
		if !entered[index] && node.proven != nil && members[int(node.proven.Id())] {
			tops[int(node.proven.Id())] = true
		}
	}
	return tops
}

// programRegionBegin is a function that allocates a top type in its own body (rule 1). Returned is
// set when its declared return type is a top type too (rule 2).
type programRegionBegin struct {
	function *ast.Node
	returned bool
}

// programRegionBegins finds every function a begin goes first in. An allocation counts in the
// nearest function around it, so a nested function counts at its own entry and never its
// parent's. A function value is allocated where it's made, in its parent; a class instance in its
// constructor, or at the class when it declares none. Its types are the ones programAllocation
// reads: the value's own, its slot's and its implied target's.
func (l *lowering) programRegionBegins(modules []*ast.SourceFile, tops map[int]bool) []programRegionBegin {
	if len(tops) == 0 {
		return nil
	}
	top := func(proven *checker.Type) bool {
		if proven == nil {
			return false
		}
		proven = l.concrete(proven)
		if proven.Flags()&checker.TypeFlagsUnion != 0 {
			for _, member := range proven.Types() {
				if tops[int(member.Id())] {
					return true
				}
			}
			return false
		}
		return tops[int(proven.Id())]
	}
	allocating := map[*ast.Node]bool{}
	order := []*ast.Node{}
	allocates := func(function *ast.Node) {
		if function != nil && !allocating[function] {
			allocating[function] = true
			order = append(order, function)
		}
	}
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if ast.IsPartOfTypeNode(node) {
			return false
		}
		switch node.Kind {
		case ast.KindObjectLiteralExpression, ast.KindArrayLiteralExpression, ast.KindFunctionExpression, ast.KindArrowFunction:
			types := []*checker.Type{l.checker.GetTypeAtLocation(node), l.slotType(node), l.impliedTarget(node)}
			for _, proven := range types {
				if top(proven) {
					allocates(programRegionFunction(node.Parent))
					break
				}
			}
		case ast.KindFunctionDeclaration, ast.KindMethodDeclaration:
			// A nested declaration's closure, or an object literal's method, is made in its parent.
			if parent := programRegionFunction(node.Parent); parent != nil && (node.Kind == ast.KindFunctionDeclaration || node.Parent.Kind == ast.KindObjectLiteralExpression) {
				if symbol := l.checker.GetSymbolOfDeclaration(node); symbol != nil && top(l.checker.GetTypeOfSymbol(symbol)) {
					allocates(parent)
				}
			}
		case ast.KindClassDeclaration, ast.KindClassExpression:
			if symbol := l.checker.GetSymbolOfDeclaration(node); symbol != nil && top(l.checker.GetDeclaredTypeOfSymbol(symbol)) {
				constructor := node
				for _, member := range node.Members() {
					if member.Kind == ast.KindConstructor && member.Body() != nil {
						constructor = member
					}
				}
				allocates(constructor)
			}
		}
		return node.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	begins := make([]programRegionBegin, 0, len(order))
	for _, function := range order {
		// A constructor, or a class standing in for its implicit one, returns the instance it allocates.
		returned := true
		if function.Kind != ast.KindClassDeclaration && function.Kind != ast.KindClassExpression && function.Kind != ast.KindConstructor {
			signature := l.checker.GetSignatureFromDeclaration(function)
			returned = signature != nil && top(l.checker.GetReturnTypeOfSignature(signature))
		}
		begins = append(begins, programRegionBegin{function: function, returned: returned})
	}
	return begins
}

// programRegionFunction is the function whose body holds node, or nil at module level.
func programRegionFunction(node *ast.Node) *ast.Node {
	for node != nil {
		switch node.Kind {
		case ast.KindFunctionDeclaration, ast.KindFunctionExpression, ast.KindArrowFunction, ast.KindMethodDeclaration, ast.KindConstructor, ast.KindGetAccessor, ast.KindSetAccessor:
			return node
		case ast.KindSourceFile:
			return nil
		}
		node = node.Parent
	}
	return nil
}
