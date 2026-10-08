package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Stored Array and String iterator objects reuse the collection next closure ABI.
// The existing compiler loop lowering consumes these through the same next protocol.
func (l *lowering) librarySequenceIterator(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	var receiver *ast.Node
	part := ""
	if callee.Kind == ast.KindElementAccessExpression {
		access := callee.AsElementAccessExpression()
		if l.symbolIterator(access.ArgumentExpression) {
			receiver, part = access.Expression, "iterator"
		}
	} else if callee.Kind == ast.KindPropertyAccessExpression {
		access := callee.AsPropertyAccessExpression()
		if name := callee.Name().Text(); name == "keys" || name == "values" || name == "entries" {
			receiver, part = access.Expression, name
		}
	}
	if receiver == nil {
		return nil, false, nil
	}
	proven := l.checker.GetTypeAtLocation(receiver)
	of, _ := l.representation(proven)
	if part == "iterator" && l.libraryIteratorType(proven) {
		if len(call.Arguments.Nodes) != 0 {
			return nil, true, l.notYet(node, "a built-in iterator factory with arguments")
		}
		value, err := l.expression(receiver)
		return value, true, err
	}
	if part == "iterator" && of == ir.Map {
		part = "entries"
		if l.isSet(receiver) {
			part = "values"
		}
		return l.libraryCollectionIterator(node, receiver, part)
	}
	if of != ir.Array && (of != ir.String || part != "iterator") {
		return nil, false, nil
	}
	if len(call.Arguments.Nodes) != 0 {
		return nil, true, l.notYet(node, "a built-in iterator factory with arguments")
	}
	element := ir.String
	var err error
	if of == ir.Array {
		element, err = l.elementType(receiver)
		if err != nil {
			return nil, true, err
		}
		if part == "iterator" {
			part = "values"
		}
		if element.IsReference() && element != ir.String {
			return nil, true, l.notYet(node, "an Array iterator over reference elements: compiler cycle analysis does not yet follow its hidden array capture")
		}
		if (element == ir.Boolean && part == "values") || element == ir.Weak || slotless(element) {
			return nil, true, l.notYet(node, "an Array iterator whose completion or element needs compiler slot representation support")
		}
	}
	source, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	return ir.CollectionIterator{Collection: source, Part: part, Key: ir.Number, Value: element}, true, nil
}

// A library interface can also describe a user object with a different iterator
// factory. The compiler must prove protocol origin before treating that as built-in.
func (l *lowering) libraryIteratorOrigin(node *ast.Node) error {
	view := l.checker.GetTypeAtLocation(node)
	if !l.libraryIteratorType(view) {
		return nil
	}
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
		if source.Kind == ast.KindObjectLiteralExpression {
			// A copy is refused at its construction; it is not a custom protocol origin.
			for _, property := range source.AsObjectLiteralExpression().Properties.Nodes {
				if property.Kind == ast.KindSpreadAssignment && l.libraryIteratorType(l.checker.GetTypeAtLocation(property.AsSpreadAssignment().Expression)) {
					return false
				}
			}
		}
		if source.Kind == ast.KindObjectLiteralExpression || source.Kind == ast.KindNewExpression {
			actual := l.checker.GetTypeAtLocation(source)
			if actual.Flags()&checker.TypeFlagsObject != 0 && l.checker.IsTypeAssignableTo(actual, view) && !l.libraryIteratorType(actual) {
				unsafe = true
				return true
			}
		}
		return source.ForEachChild(visit)
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	if unsafe {
		return l.notYet(node, "a built-in iterator view that can hide a custom Symbol.iterator method; compiler protocol origin is not proved")
	}
	return nil
}
