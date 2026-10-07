package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// Presence is outside the phantom contract even when the checker says a member is
// required. Search every union arm and constrained generic before flow can use it.
func (l *lowering) phantomArrayFields(proven *checker.Type, seen map[*checker.Type]bool) []*ast.Symbol {
	proven = l.concrete(proven)
	if proven == nil || seen[proven] {
		return nil
	}
	seen[proven] = true
	if proven.Flags()&checker.TypeFlagsTypeParameter != 0 {
		return l.phantomArrayFields(l.checker.GetBaseConstraintOfType(proven), seen)
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		var fields []*ast.Symbol
		for _, part := range proven.Types() {
			fields = append(fields, l.phantomArrayFields(part, seen)...)
		}
		return fields
	}
	if l.phantomArrayBase(proven) == nil {
		return nil
	}
	_, fields := l.phantomArrayParts(proven)
	return fields
}

func (l *lowering) phantomArrayObservation(node, receiver, key *ast.Node, operation string, write bool) error {
	fields := l.phantomArrayFields(l.checker.GetTypeAtLocation(receiver), map[*checker.Type]bool{})
	// A literal key known to name a real array member does not observe the brand.
	// A dynamic key can name any brand field, so it cannot prove absence of observation.
	name, known := "", false
	if key != nil {
		proven := l.checker.GetTypeAtLocation(key)
		if key.Parent != nil && key.Parent.Kind == ast.KindPropertyAccessExpression && key.Parent.Name() == key {
			name, known = key.Text(), true
		} else if proven.Flags()&checker.TypeFlagsStringLiteral != 0 {
			name, known = proven.AsLiteralType().Value().(string), true
		}
	}
	for _, field := range fields {
		if known && name != field.Name {
			continue
		}
		if write {
			return &Refused{Where: l.program.Where(node), What: "a write to phantom array brand member " + field.Name, Fix: "keep the brand member phantom and unwritten; store observable state in a separate real field or Map"}
		}
		return &Refused{Where: l.program.Where(node), What: "presence of phantom array brand member " + field.Name + " via " + operation, Fix: "read the brand member's undefined value without testing its presence; use a separate real field or Map for observable presence"}
	}
	return nil
}

// Written targets include computed members and nested destructuring. Inspect them
// before slot lowering, which must never create storage for a checker-only member.
func (l *lowering) phantomArrayWrite(node, target *ast.Node) error {
	target = ast.SkipParentheses(target)
	switch target.Kind {
	case ast.KindPropertyAccessExpression:
		access := target.AsPropertyAccessExpression()
		return l.phantomArrayObservation(node, access.Expression, access.Name(), "", true)
	case ast.KindElementAccessExpression:
		access := target.AsElementAccessExpression()
		// Numeric indexes change elements, not brand presence.
		keyType := l.checker.GetTypeAtLocation(access.ArgumentExpression)
		if keyType.Flags()&checker.TypeFlagsNumberLike != 0 {
			return nil
		}
		return l.phantomArrayObservation(node, access.Expression, access.ArgumentExpression, "", true)
	case ast.KindArrayLiteralExpression:
		for _, element := range target.AsArrayLiteralExpression().Elements.Nodes {
			if err := l.phantomArrayWrite(node, element); err != nil {
				return err
			}
		}
	case ast.KindObjectLiteralExpression:
		for _, property := range target.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind == ast.KindPropertyAssignment {
				if err := l.phantomArrayWrite(node, property.AsPropertyAssignment().Initializer); err != nil {
					return err
				}
			}
		}
	case ast.KindSpreadElement:
		return l.phantomArrayWrite(node, target.AsSpreadElement().Expression)
	case ast.KindSpreadAssignment:
		return l.phantomArrayWrite(node, target.AsSpreadAssignment().Expression)
	case ast.KindBinaryExpression:
		// A default in a destructuring target still writes its left side.
		return l.phantomArrayWrite(node, target.AsBinaryExpression().Left)
	}
	return nil
}

func (l *lowering) phantomArrayPresenceRefusal(node *ast.Node) error {
	switch node.Kind {
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		if binary.OperatorToken.Kind == ast.KindInKeyword {
			return l.phantomArrayObservation(node, binary.Right, binary.Left, "in", false)
		}
		if ast.IsAssignmentOperator(binary.OperatorToken.Kind) {
			return l.phantomArrayWrite(node, binary.Left)
		}
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
			return l.phantomArrayWrite(node, unary.Operand)
		}
	case ast.KindPostfixUnaryExpression:
		return l.phantomArrayWrite(node, node.AsPostfixUnaryExpression().Operand)
	case ast.KindSpreadAssignment:
		return l.phantomArrayObservation(node, node.AsSpreadAssignment().Expression, nil, "object spread", false)
	case ast.KindCallExpression:
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		var receiver *ast.Node
		var name string
		switch callee.Kind {
		case ast.KindPropertyAccessExpression:
			receiver, name = callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
		case ast.KindElementAccessExpression:
			access := callee.AsElementAccessExpression()
			key := ast.SkipParentheses(access.ArgumentExpression)
			if key.Kind != ast.KindStringLiteral {
				return nil
			}
			receiver, name = access.Expression, key.Text()
		default:
			return nil
		}
		arguments := call.Arguments.Nodes
		if l.isLibraryGlobal(ast.SkipParentheses(receiver), "Object") {
			switch name {
			case "keys", "values", "entries", "getOwnPropertyNames", "hasOwn":
				if len(arguments) == 0 {
					return nil
				}
				var key *ast.Node
				if name == "hasOwn" && len(arguments) > 1 {
					key = arguments[1]
				}
				return l.phantomArrayObservation(node, arguments[0], key, "Object."+name, false)
			case "assign":
				if len(arguments) == 0 {
					return nil
				}
				for _, source := range arguments[1:] {
					if err := l.phantomArrayObservation(node, source, nil, "Object.assign", false); err != nil {
						return err
					}
				}
				if len(arguments) > 1 {
					return l.phantomArrayObservation(node, arguments[0], nil, "", true)
				}
			}
		}
		if l.isLibraryGlobal(ast.SkipParentheses(receiver), "JSON") && name == "stringify" && len(arguments) > 0 {
			return l.phantomArrayObservation(node, arguments[0], nil, "JSON.stringify", false)
		}
		if (name == "hasOwnProperty" || name == "propertyIsEnumerable") && l.libraryMember(callee) {
			var key *ast.Node
			if len(arguments) > 0 {
				key = arguments[0]
			}
			return l.phantomArrayObservation(node, receiver, key, name, false)
		}
	}
	return nil
}
