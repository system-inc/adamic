package lower

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) objectCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	refused := func(reason string) (ir.Expression, bool, error) {
		return nil, true, &Refused{Where: l.program.Where(node), What: "Object." + name, Fix: reason}
	}
	switch name {
	case "defineProperty", "defineProperties", "getOwnPropertyDescriptor", "getOwnPropertyDescriptors":
		return refused("property descriptors can change the presence, type or access behavior of fields; Adamic fields have a fixed shape and are plain loads and stores")
	case "getPrototypeOf", "setPrototypeOf", "create":
		return refused("prototypes expose or replace fields outside the declared shape; use a declared object or class with composition")
	case "fromEntries":
		return refused("tsc returns an index-signature object with unproven keys; Adamic fixes object shapes and refuses index signatures; use Map")
	case "groupBy":
		return nil, true, l.notYet(node, "Object.groupBy's partial record with dynamically present keys (use Map and an explicitly typed grouping loop)")
	}
	count := 1
	if name == "is" || name == "hasOwn" {
		count = 2
	}
	written := node.AsCallExpression().Arguments.Nodes
	if name == "assign" {
		if len(written) < 1 {
			return nil, true, l.notYet(node, "Object.assign without a target")
		}
	} else if len(written) != count {
		return nil, true, l.notYet(node, "Object."+name+" with these arguments")
	}
	if name == "freeze" {
		proven := l.checker.GetTypeAtLocation(written[0])
		if proven.Flags()&(checker.TypeFlagsNumberLike|checker.TypeFlagsBooleanLike|checker.TypeFlagsStringLike|checker.TypeFlagsUndefined) != 0 {
			value, err := l.expression(written[0])
			return value, true, err
		}
	}
	call := ir.ObjectCall{Method: name, Returns: ir.Boolean, Readiness: sourceExpression(node)}
	switch name {
	case "is":
		for _, argument := range written {
			value, err := l.expression(argument)
			if err != nil {
				return nil, true, err
			}
			call.Arguments = append(call.Arguments, fit(value, ir.Union))
		}
	case "isFrozen":
		value, err := l.expression(written[0])
		if err != nil {
			return nil, true, err
		}
		call.Arguments = []ir.Expression{fit(value, ir.Union)}
	case "keys", "values", "entries", "freeze", "hasOwn", "assign":
		// Reflection cannot use a widened view: a hidden field can have another representation.
		// A plain const's literal initializer proves the complete shape, including field presence.
		if !l.exactObject(written[0], 0) && !(name == "hasOwn" && isClassInstance(l.checker.GetTypeAtLocation(written[0]))) {
			return nil, true, l.notYet(written[0], "Object."+name+" on a shape not proven by a plain literal or its const binding")
		}
		value, err := l.expression(written[0])
		if err != nil {
			return nil, true, err
		}
		if value.Type() != ir.Object || l.includesUndefined(l.checker.GetTypeAtLocation(written[0])) || checker.IsTupleType(l.checker.GetTypeAtLocation(written[0])) {
			return nil, true, l.notYet(written[0], "Object."+name+" on other than a present plain object")
		}
		call.Arguments = []ir.Expression{value}
		switch name {
		case "hasOwn":
			key := ast.SkipParentheses(written[1])
			if key.Kind != ast.KindStringLiteral || !l.hasProperty(written[0], key.Text()) || len(key.Text()) > 0 && key.Text()[0] == '#' {
				return refused("hasOwn requires a string literal naming a declared public field or method; use Map for arbitrary keys")
			}
			keyValue, err := l.expression(key)
			if err != nil {
				return nil, true, err
			}
			call.Arguments = append(call.Arguments, keyValue)
		case "freeze":
			call.Returns = ir.Object
		case "keys":
			call.Returns = ir.Array
		case "values", "entries":
			result := l.checker.GetTypeAtLocation(node)
			arguments := l.checker.GetTypeArguments(result)
			if len(arguments) != 1 {
				return refused("tsc's result must be an array with a proven element type")
			}
			element := arguments[0]
			if name == "entries" {
				pair := l.checker.GetTypeArguments(element)
				if len(pair) != 2 {
					return refused("tsc's entries must have a proven value type")
				}
				element = pair[1]
			}
			call.Element, _ = l.representation(element)
			if call.Element != ir.Number && call.Element != ir.String && call.Element != ir.Boolean {
				return refused("tsc's result must have one homogeneous number, string or boolean value type; any and widened field views are unsound")
			}
			if declaration := l.enumObject(written[0]); declaration != nil {
				fields, err := l.enumFields(declaration)
				if err != nil {
					return nil, true, err
				}
				for _, field := range fields {
					if field.Value.Type() != call.Element {
						return refused("numeric enum reverse mappings add string values; use Object.keys or read members individually")
					}
				}
			}
			for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(written[0])) {
				of, known := l.representation(l.checker.GetTypeOfSymbol(field))
				if !known || of != call.Element || field.Flags&ast.SymbolFlagsOptional != 0 {
					return refused("every present field must have tsc's result element representation; optional or heterogeneous fields cannot be read soundly")
				}
			}
			call.Returns = ir.Array
		case "assign":
			if l.checker.GetTypeAtLocation(node).Flags()&(checker.TypeFlagsAny|checker.TypeFlagsUnknown) != 0 {
				return refused("tsc's result is any or unknown, not a proven object type; use one to three typed sources")
			}
			target := l.checker.GetTypeAtLocation(written[0])
			for _, argument := range written[1:] {
				if !l.exactObject(argument, 0) {
					return refused("each source must have its complete shape proven by a literal or const binding; a widened source can hide overwriting fields")
				}
				source := l.checker.GetTypeAtLocation(argument)
				for _, field := range l.checker.GetPropertiesOfType(source) {
					into := l.checker.GetPropertyOfType(target, field.Name)
					if into == nil {
						return nil, true, l.notYet(argument, "Object.assign adding a field to its target's fixed shape")
					}
					fromType, toType := l.checker.GetTypeOfSymbol(field), l.checker.GetTypeOfSymbol(into)
					of, known := l.representation(fromType)
					if !known || (of != ir.Number && of != ir.Boolean && of != ir.String) || !l.enumAssignable(fromType, toType) || !l.enumAssignable(toType, fromType) || !l.checker.IsTypeAssignableTo(fromType, toType) || !l.checker.IsTypeAssignableTo(toType, fromType) {
						return refused("source and target field types must agree in both directions with tsc's intersection result; widening, conflicting fields and reference cycles are refused")
					}
				}
				lowered, err := l.expression(argument)
				if err != nil {
					return nil, true, err
				}
				call.Arguments = append(call.Arguments, lowered)
			}
			call.Returns = ir.Object
		}
	default:
		return nil, true, l.notYet(node, "Object."+name)
	}
	return call, true, nil
}

// exactObject proves there are no hidden fields and no synthetic absent slots. An explicit type
// annotation, alias, spread or call would need a separate proof and is deliberately not guessed.
func (l *lowering) exactObject(node *ast.Node, depth int) bool {
	if depth > 16 {
		return false
	}
	node = ast.SkipParentheses(node)
	if l.enumObject(node) != nil {
		return true
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
			if field.Kind != ast.KindPropertyAssignment && field.Kind != ast.KindShorthandPropertyAssignment {
				return false
			}
			if field.Name().Kind != ast.KindIdentifier && field.Name().Kind != ast.KindStringLiteral {
				return false
			}
			if strings.ContainsRune(field.Name().Text(), 0) || field.Name().Text() == "__proto__" || len(field.Name().Text()) > 0 && field.Name().Text()[0] == '#' {
				return false
			}
		}
		return true
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration {
		return false
	}
	variable := declaration.AsVariableDeclaration()
	if variable.Type != nil || variable.Initializer == nil || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	return l.exactObject(variable.Initializer, depth+1)
}

// objectCanFreeze is conservative across aliases and calls. A try around a write in a program
// that freezes objects is NotYet until library TypeErrors participate in exception cleanup.
func (l *lowering) objectCanFreeze() bool {
	found := false
	walk(l.result.Main, func(node any) bool {
		if call, ok := node.(ir.ObjectCall); ok && call.Method == "freeze" {
			found = true
		}
		return !found
	})
	for _, function := range l.result.Functions {
		walk(function.Body, func(node any) bool {
			if call, ok := node.(ir.ObjectCall); ok && call.Method == "freeze" {
				found = true
			}
			return !found
		})
	}
	return found
}

// objectIntersection assigns plain structural intersections an object reference
// slot. Checked views retain separate lazy obligations for the resulting fields.
func (l *lowering) objectIntersection(proven *checker.Type) (ir.Type, bool) {
	if l.objectPrimitiveIntersectionStorage(proven) {
		return ir.Object, true
	}
	for _, part := range proven.Types() {
		if part.Flags()&checker.TypeFlagsObject == 0 || l.checker.IsArrayType(part) || checker.IsTupleType(part) || isClassInstance(part) || len(l.checker.GetSignaturesOfType(part, checker.SignatureKindCall)) != 0 {
			return 0, false
		}
	}
	for _, field := range l.checker.GetPropertiesOfType(proven) {
		of, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if !known || (of != ir.Number && of != ir.String && of != ir.Boolean) {
			return 0, false
		}
	}
	return ir.Object, true
}
