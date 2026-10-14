package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// TypeScript can infer {}[] for [0, {}]: {} admits primitive values too. A native
// object slot cannot hold those values. Check the source shapes before any search
// can reach emission, including a literal hidden behind an assignable array view.
func (l *lowering) libraryArraySearchSlots(node, receiver *ast.Node, element ir.Type) error {
	if element != ir.Object {
		return nil
	}
	view := l.checker.GetElementTypeOfArrayType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver)))
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return err
	}
	unsafe := false
	var visit ast.Visitor
	visit = func(source *ast.Node) bool {
		if unsafe {
			return true
		}
		if source.Kind == ast.KindArrayLiteralExpression {
			for _, item := range source.AsArrayLiteralExpression().Elements.Nodes {
				proven := l.checker.GetTypeAtLocation(item)
				if item.Kind == ast.KindSpreadElement {
					proven = l.checker.GetElementTypeOfArrayType(l.checker.GetTypeAtLocation(item.AsSpreadElement().Expression))
				}
				if proven == nil {
					continue
				}
				if of, known := l.representation(proven); known && of != ir.Object && l.checker.IsTypeAssignableTo(proven, view) {
					unsafe = true
					return true
				}
			}
		}
		return source.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if unsafe {
		return l.notYet(node, "an array search whose object element type can hide primitive slots (heterogeneous arrays need tagged elements)")
	}
	return nil
}
