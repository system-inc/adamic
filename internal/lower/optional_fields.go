package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A contextual type determines storage even when the literal omits its optional fields.
func (l *lowering) optionalLiteralSlots(node *ast.Node, own []ir.Field) ([]ir.Field, error) {
	contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
	// Object.assign returns its target. A contextual result therefore also constrains
	// the target literal's storage, even when generic inference sees only {}.
	parent := node.Parent
	if parent != nil && parent.Kind == ast.KindCallExpression {
		call := parent.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if len(call.Arguments.Nodes) > 0 && call.Arguments.Nodes[0] == node && callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "assign" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
			if result := l.checker.GetContextualType(parent, checker.ContextFlagsNone); result != nil {
				contextual = result
			}
		}
	}
	if contextual == nil {
		return nil, nil
	}
	members := []*checker.Type{contextual}
	if contextual.Flags()&checker.TypeFlagsUnion != 0 {
		members = contextual.Types()
	}
	given := map[string]bool{}
	for _, field := range own {
		given[field.Name] = true
	}
	var missing []ir.Field
	made := l.checker.GetTypeAtLocation(node)
	for _, member := range members {
		if !l.checker.IsTypeAssignableTo(made, member) {
			continue
		}
		for _, field := range l.checker.GetPropertiesOfType(member) {
			if given[field.Name] || field.Flags&ast.SymbolFlagsOptional == 0 {
				continue
			}
			of, _ := l.representation(l.checker.GetTypeOfSymbol(field))
			if len(members) > 1 {
				of = l.declaredField(node, field.Name)
			}
			if of == 0 || slotless(of) || !(of.IsReference() || of == ir.MaybeNumber) {
				continue // Unsupported representations remain refused at their reads and writes.
			}
			missing = append(missing, ir.Field{Name: field.Name, Value: fit(ir.Undefined{}, of)})
			given[field.Name] = true
		}
	}
	return missing, nil
}

// Presence operators are limited to declared optional data fields of plain objects.
func (l *lowering) optionalPresence(node, receiver, key *ast.Node, remove bool) (ir.Expression, error) {
	key = ast.SkipParentheses(key)
	proven := l.checker.GetTypeAtLocation(receiver)
	if (key.Kind != ast.KindStringLiteral && !(remove && ast.IsIdentifier(key))) || isClassInstance(proven) || l.includesUndefined(proven) {
		return nil, l.notYet(node, "presence on other than a declared optional plain-object field")
	}
	if !remove {
		switch key.Text() {
		case "__proto__", "constructor", "toString", "toLocaleString", "valueOf", "hasOwnProperty", "isPrototypeOf", "propertyIsEnumerable", "__defineGetter__", "__defineSetter__", "__lookupGetter__", "__lookupSetter__":
			return nil, l.notYet(node, "in on an inherited Object prototype name")
		}
	}
	field := l.checker.GetPropertyOfType(proven, key.Text())
	if field == nil || field.Flags&ast.SymbolFlagsOptional == 0 {
		return nil, &Refused{Where: l.program.Where(node), What: "presence on an undeclared or required field", Fix: "use a declared optional data field"}
	}
	object, err := l.expression(receiver)
	if err != nil {
		return nil, err
	}
	if object.Type() != ir.Object {
		return nil, l.notYet(node, "presence on other than a plain object")
	}
	method := "optionalIn"
	if remove {
		method = "optionalDelete"
	}
	return ir.ObjectCall{Method: method, Arguments: []ir.Expression{object, ir.StringConstant{Index: l.constant(key.Text())}}, Returns: ir.Boolean}, nil
}

// A const's literal or spread initializer proves its keys regardless of its annotated view.
func (l *lowering) literalObjectKeys(node *ast.Node, depth int) bool {
	if depth > 16 {
		return false
	}
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindBinaryExpression {
		binary := node.AsBinaryExpression()
		// Every non-nullish object is truthy, so a present result has the right origin.
		if binary.OperatorToken.Kind == ast.KindAmpersandAmpersandToken && l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(binary.Left)).Flags()&checker.TypeFlagsObject != 0 {
			return l.literalObjectKeys(binary.Right, depth+1)
		}
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
			if field.Kind == ast.KindSpreadAssignment {
				source := field.AsSpreadAssignment().Expression
				if l.includesUndefined(l.checker.GetTypeAtLocation(source)) {
					contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone)
					if contextual == nil {
						return false
					}
					for _, property := range l.checker.GetPropertiesOfType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(source))) {
						if property.Flags&ast.SymbolFlagsOptional == 0 {
							return false
						}
					}
				} else if !l.literalObjectKeys(source, depth+1) {
					return false
				}
			}
		}
		return true
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "assign" || !l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
			return false
		}
		for _, argument := range call.Arguments.Nodes {
			if !l.literalObjectKeys(argument, depth+1) {
				return false
			}
		}
		return len(call.Arguments.Nodes) > 0
	}
	if !ast.IsIdentifier(node) {
		return false
	}
	symbol := l.symbol(node)
	if symbol == nil || len(symbol.Declarations) != 1 {
		return false
	}
	declaration := symbol.Declarations[0]
	if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 {
		return false
	}
	initializer := declaration.AsVariableDeclaration().Initializer
	return initializer != nil && l.literalObjectKeys(initializer, depth+1)
}
