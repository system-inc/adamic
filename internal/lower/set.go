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
	arguments := l.typeArguments(l.checker.GetTypeAtLocation(node))
	if len(arguments) != 1 {
		return 0, l.notYet(node, "a Set whose element type isn't known")
	}
	element, isKnown := l.representation(arguments[0])
	if !isKnown || !keyable(element) {
		return 0, l.notYet(node, "a Set of "+l.checker.TypeToString(arguments[0])+" (a Set holds strings, numbers, booleans, objects, arrays, maps or functions so far)")
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
	if ast.SkipParentheses(arguments.Nodes[0]).Kind == ast.KindArrayLiteralExpression {
		if values, known, err := l.librarySetLiteral(arguments.Nodes[0], element); known {
			lowered.Values = values
			return lowered, err
		}
	}
	if l.libraryEmptyCollectionArgument(arguments.Nodes[0]) {
		return lowered, nil
	}
	// From anything iterable: an array, a string's code points, a Set, a Map's entries, or what keys()
	// and values() give (collections.go).
	values, from, err := l.iterated(arguments.Nodes[0])
	if err != nil {
		return nil, err
	}
	if from == ir.Number && element == ir.MaybeNumber {
		values = l.libraryOptionalNumbers(arguments.Nodes[0], values)
		from = element
	}
	if from != element {
		return nil, l.notYet(arguments.Nodes[0], "new Set from elements held otherwise than the Set's")
	}
	lowered.Values = values
	return lowered, nil
}

// setMethod lowers set.add, set.has and set.delete. isBuiltin is false for any other method, which
// is left to be refused as a call stage 0 doesn't know.
func (l *lowering) setMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	if lowered, known, err := l.librarySetMethod(node, receiver, name); known {
		return lowered, true, err
	}
	if name == "clear" || name == "forEach" {
		return l.clearOrVisit(node, receiver, name, true)
	}
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
	// set.has(undefined), written out, which the checker allows only where the elements may be
	// undefined: a reference's undefined is its null one, which the set holds as any element.
	_, isUndefined := value.(ir.Undefined)
	if !isUndefined || !element.IsReference() {
		if value = fit(value, element); value.Type() != element {
			return nil, true, l.notYet(arguments[0], "set."+name+" with a value of another type than the elements")
		}
	}
	switch name {
	case "add":
		return ir.SetAdd{Set: set, Value: value, Element: element, Site: l.writeSite(receiver)}, true, nil
	case "has":
		return ir.MapHas{Map: set, Key: value, KeyType: element}, true, nil
	}
	return ir.MapDelete{Map: set, Key: value, KeyType: element}, true, nil
}

// forOfSet lowers for (const element of set), and of set.keys() and set.values(), which give the
// same elements. Its steps are a Map's keys.
func (l *lowering) forOfSet(node *ast.Node, iterable ir.Expression, iterated *ast.Node, part string, name *ast.Node) ([]ir.Statement, error) {
	element, err := l.setElement(iterated)
	if err != nil {
		return nil, err
	}
	lowered := ir.ForOf{Iterable: iterable, MapPart: "keys", Key: element, Value: ir.Number}
	// Each entry is [element, element]: the first name bound takes the element, and any other is
	// declared from it, each its own variable.
	var others []ir.Statement
	switch {
	case part == "entries" && name.Kind == ast.KindArrayBindingPattern:
		bound := false
		elements := name.AsBindingPattern().Elements.Nodes
		if len(elements) > 2 {
			return nil, l.notYet(name, "destructuring more than a Set entry's two elements")
		}
		for _, binding := range elements {
			// A hole, [, second], is a binding element with no name.
			if binding.Kind == ast.KindOmittedExpression || binding.Name() == nil {
				continue
			}
			declared := binding.AsBindingElement()
			if !ast.IsIdentifier(binding.Name()) || declared.Initializer != nil || declared.DotDotDotToken != nil {
				return nil, l.notYet(binding, "a destructured name that isn't plain")
			}
			local, err := l.declareLocal(binding.Name())
			if err != nil {
				return nil, err
			}
			if !bound {
				lowered.Local, bound = local, true
				continue
			}
			others = append(others, ir.Declare{Local: local, Value: ir.Read{Local: lowered.Local, Of: element}})
		}
		if !bound {
			return nil, l.notYet(name, "a Set entry destructured into no names")
		}
	case part != "entries" && ast.IsIdentifier(name):
		if lowered.Local, err = l.declareLocal(name); err != nil {
			return nil, err
		}
	default:
		return nil, l.notYet(name, "for...of over a Set's "+part+" into this name")
	}
	body, err := l.statement(node.AsForInOrOfStatement().Statement)
	if err != nil {
		return nil, err
	}
	lowered.Body = append(others, body...)
	return []ir.Statement{lowered}, nil
}
