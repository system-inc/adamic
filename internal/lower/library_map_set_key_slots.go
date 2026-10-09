package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Readonly collection covariance does not convert existing scalar key slots
// into boxes. Keep such views refused until collection views can adapt reads
// and lookups to the source representation. Constructors and typed factories
// in the closed program supply the possible runtime collection representations.
func (l *lowering) libraryCollectionKeySlots(node *ast.Node, element ir.Type) error {
	if element != ir.Union {
		return nil
	}
	view := l.checker.GetTypeAtLocation(node)
	set := l.isSet(node)
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
		if source.Kind == ast.KindNewExpression || source.Kind == ast.KindCallExpression {
			actual := l.checker.GetTypeAtLocation(source)
			collection := l.isLibraryType(actual, "Map", "ReadonlyMap")
			if set {
				collection = l.isLibraryType(actual, "Set", "ReadonlySet")
			}
			if collection && l.checker.IsTypeAssignableTo(actual, view) {
				arguments := l.typeArguments(actual)
				if len(arguments) > 0 {
					stored, known := l.representation(arguments[0])
					if known && stored != ir.Union {
						unsafe = true
						return true
					}
				}
			}
		}
		return source.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if unsafe {
		return l.notYet(node, "a boxed collection key view that can hide unboxed source keys; copy into an explicitly typed Map or Set")
	}
	return nil
}
