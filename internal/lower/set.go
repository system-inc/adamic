package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A Set is held as a Map whose values aren't used (ir.Map), which is what JavaScript's Set is: the
// same insertion order, the same SameValueZero keys, the same rules for what an iteration sees. What
// tells them apart is the checker's type, so every Set operation is found here, by it, before the
// Map ones. Anything Map-shaped asked of a Set (its key and value types) finds one type argument, not
// two, and stops as not yet.

// isSet reports whether a node is one of the library's Sets.
func (l *lowering) isSet(node *ast.Node) bool {
	return l.isLibraryType(l.checker.GetTypeAtLocation(node), "Set", "ReadonlySet")
}

// setElement is the representation of a Set's elements, strings or numbers in 0.2.
func (l *lowering) setElement(node *ast.Node) (ir.Type, error) {
	arguments := l.checker.GetTypeArguments(l.checker.GetTypeAtLocation(node))
	if len(arguments) != 1 {
		return 0, l.notYet(node, "a Set whose element type isn't known")
	}
	element, isKnown := l.representation(arguments[0])
	if !isKnown || (element != ir.String && element != ir.Number) {
		return 0, l.notYet(node, "a Set of "+l.checker.TypeToString(arguments[0])+" (a Set holds strings or numbers so far)")
	}
	return element, nil
}

// newSet lowers new Set() and new Set(array).
func (l *lowering) newSet(node *ast.Node) (ir.Expression, error) {
	element, err := l.setElement(node)
	if err != nil {
		return nil, err
	}
	lowered := ir.SetNew{Element: element}
	arguments := node.AsNewExpression().Arguments
	if arguments == nil || len(arguments.Nodes) == 0 {
		return lowered, nil
	}
	if len(arguments.Nodes) != 1 {
		return nil, l.notYet(node, "new Set with more than one argument")
	}
	values, err := l.expression(arguments.Nodes[0])
	if err != nil {
		return nil, err
	}
	if values.Type() != ir.Array {
		return nil, l.notYet(arguments.Nodes[0], "new Set from a "+typeName(values.Type())+" (an array is what it takes so far)")
	}
	if from, err := l.elementType(arguments.Nodes[0]); err != nil || from != element {
		return nil, l.notYet(arguments.Nodes[0], "new Set from an array of other elements")
	}
	lowered.Values = values
	return lowered, nil
}

// setMethod lowers set.add, set.has and set.delete. isBuiltin is false for any other method, which
// is left to be refused as a call stage 0 doesn't know.
func (l *lowering) setMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	if name != "add" && name != "has" && name != "delete" {
		return nil, false, nil
	}
	element, err := l.setElement(receiver)
	if err != nil {
		return nil, true, err
	}
	set, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		return nil, true, l.notYet(node, "set."+name+" with other than one value")
	}
	value, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	if value.Type() != element {
		return nil, true, l.notYet(arguments[0], "set."+name+" with a value of another type than the elements")
	}
	switch name {
	case "add":
		return ir.SetAdd{Set: set, Value: value, Element: element}, true, nil
	case "has":
		return ir.MapHas{Map: set, Key: value, KeyType: element}, true, nil
	}
	return ir.MapDelete{Map: set, Key: value, KeyType: element}, true, nil
}

// forOfSet lowers for (const element of set), and of set.keys() and set.values(), which give the
// same elements. Its steps are a Map's keys.
func (l *lowering) forOfSet(node *ast.Node, iterable ir.Expression, iterated *ast.Node, part string, name *ast.Node) ([]ir.Statement, error) {
	if part == "entries" {
		return nil, l.notYet(iterated, "for...of over a Set's entries ([element, element] pairs)")
	}
	if !ast.IsIdentifier(name) {
		return nil, l.notYet(name, "destructuring a Set's element")
	}
	element, err := l.setElement(iterated)
	if err != nil {
		return nil, err
	}
	lowered := ir.ForOf{Iterable: iterable, MapPart: "keys", Key: element, Value: ir.Number}
	if lowered.Local, err = l.declareLocal(name); err != nil {
		return nil, err
	}
	if lowered.Body, err = l.statement(node.AsForInOrOfStatement().Statement); err != nil {
		return nil, err
	}
	return []ir.Statement{lowered}, nil
}

// setValues lowers [...set]: a new array of its elements, in order.
func (l *lowering) setValues(spread *ast.Node, set ir.Expression) (ir.Expression, error) {
	element, err := l.setElement(spread)
	if err != nil {
		return nil, err
	}
	return ir.SetValues{Set: set, Element: element}, nil
}
