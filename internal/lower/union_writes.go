package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// TypeScript types a write through a union against the union of the members' slot types, so a
// { value: number } seen as { value: string } | { value: number } can be written a string, and a
// discriminant relabeled, and native code then reads the string's bits as a number (Language H1,
// ruled Oct 18, #yawc0z5). Adamic accepts the write only when the value fits the slot in every
// member, the intersection tsc already uses for calls on a union. The same holds for a key that is
// a union (person[key] with key: K extends keyof Person). A discriminant fits only after narrowing.
func (l *lowering) unionSlotWriteRefusal(node *ast.Node) error {
	switch node.Kind {
	case ast.KindBinaryExpression:
		binary := node.AsBinaryExpression()
		operator := binary.OperatorToken.Kind
		if !ast.IsAssignmentOperator(operator) {
			return nil
		}
		target := ast.SkipParentheses(binary.Left)
		if operator == ast.KindEqualsToken && (target.Kind == ast.KindObjectLiteralExpression || target.Kind == ast.KindArrayLiteralExpression) {
			return l.unionSlotDestructuringRefusal(target)
		}
		// A compound write stores its result; a logical one stores its right side.
		value := l.checker.GetTypeAtLocation(binary.Right)
		if operator != ast.KindEqualsToken && !ast.IsLogicalOrCoalescingAssignmentOperator(operator) {
			value = l.checker.GetTypeAtLocation(node)
		}
		return l.unionSlotWrite(node, target, value)
	case ast.KindPrefixUnaryExpression:
		unary := node.AsPrefixUnaryExpression()
		if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
			return l.unionSlotWrite(node, ast.SkipParentheses(unary.Operand), l.checker.GetTypeAtLocation(node))
		}
	case ast.KindPostfixUnaryExpression:
		unary := node.AsPostfixUnaryExpression()
		if unary.Operator == ast.KindPlusPlusToken || unary.Operator == ast.KindMinusMinusToken {
			return l.unionSlotWrite(node, ast.SkipParentheses(unary.Operand), l.checker.GetTypeAtLocation(node))
		}
	}
	return nil
}

// A write through a union is refused before any gap says not yet: it's never a lowering.
func init() { refusalRules.register(100, "unionSlotWriteRefusal", (*lowering).unionSlotWriteRefusal) }

// A destructuring assignment's targets each take a part of the right side. Without that part's
// type in hand, a target through a union is held to its own read type: it fits only when every
// member's slot holds the same values.
func (l *lowering) unionSlotDestructuringRefusal(pattern *ast.Node) error {
	var elements []*ast.Node
	if pattern.Kind == ast.KindObjectLiteralExpression {
		for _, property := range pattern.AsObjectLiteralExpression().Properties.Nodes {
			switch property.Kind {
			case ast.KindPropertyAssignment:
				elements = append(elements, property.AsPropertyAssignment().Initializer)
			case ast.KindSpreadAssignment:
				elements = append(elements, property.AsSpreadAssignment().Expression)
			}
		}
	} else {
		for _, element := range pattern.AsArrayLiteralExpression().Elements.Nodes {
			if element.Kind == ast.KindSpreadElement {
				element = element.AsSpreadElement().Expression
			}
			elements = append(elements, element)
		}
	}
	for _, element := range elements {
		element = ast.SkipParentheses(element)
		// { value: cell.value = fallback } defaults the target, which is still the left side.
		if element.Kind == ast.KindBinaryExpression && element.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
			element = ast.SkipParentheses(element.AsBinaryExpression().Left)
		}
		if element.Kind == ast.KindObjectLiteralExpression || element.Kind == ast.KindArrayLiteralExpression {
			if err := l.unionSlotDestructuringRefusal(element); err != nil {
				return err
			}
			continue
		}
		if err := l.unionSlotWrite(element, element, l.checker.GetTypeAtLocation(element)); err != nil {
			return err
		}
	}
	return nil
}

// unionSlotWrite refuses a write of value to target when target's receiver or key is a union and
// some member's slot can't hold the value.
func (l *lowering) unionSlotWrite(where, target *ast.Node, value *checker.Type) error {
	var receiver, key *ast.Node
	var names []string
	switch target.Kind {
	case ast.KindPropertyAccessExpression:
		receiver = target.AsPropertyAccessExpression().Expression
		names = []string{target.AsPropertyAccessExpression().Name().Text()}
	case ast.KindElementAccessExpression:
		receiver = target.AsElementAccessExpression().Expression
		key = ast.SkipParentheses(target.AsElementAccessExpression().ArgumentExpression)
		names = l.unionWriteKeys(key)
	default:
		return nil
	}
	members := l.unionWriteMembers(receiver)
	if len(members)*max(len(names), 1) < 2 {
		// One member and one slot: the checker has already held the value to it.
		return nil
	}
	for _, member := range members {
		for _, name := range names {
			slot := l.unionWriteSlot(member, name, key)
			if slot == nil {
				return &Refused{Where: l.program.Where(where), What: "a write through a union to a slot " + l.checker.TypeToString(member) + " doesn't name", Fix: "narrow to one member before writing (union-write)"}
			}
			if !l.checker.IsTypeAssignableTo(value, slot) {
				return &Refused{Where: l.program.Where(where),
					What: "a write through a union of " + l.checker.TypeToString(value) + " into a slot that holds " + l.checker.TypeToString(slot) + " in " + l.checker.TypeToString(member),
					Fix:  "narrow to one member before writing, so the value fits the slot it lands in; a discriminant is written only after narrowing (union-write)"}
			}
		}
	}
	return nil
}

// unionWriteMembers is the receiver's members, a type parameter read through its constraint: one
// member when the receiver isn't a union.
func (l *lowering) unionWriteMembers(receiver *ast.Node) []*checker.Type {
	proven := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
	if proven.Flags()&checker.TypeFlagsInstantiable != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(proven); constraint != nil {
			proven = l.checker.GetNonNullableType(constraint)
		}
	}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		return proven.Types()
	}
	return []*checker.Type{proven}
}

// unionWriteKeys is the property names an element write's key can be: a literal's own name, each
// literal of a union key (a type parameter through its constraint), or "" for an index.
func (l *lowering) unionWriteKeys(key *ast.Node) []string {
	if key.Kind == ast.KindStringLiteral || key.Kind == ast.KindNoSubstitutionTemplateLiteral || key.Kind == ast.KindNumericLiteral {
		return []string{key.Text()}
	}
	proven := l.checker.GetTypeAtLocation(key)
	if proven.Flags()&checker.TypeFlagsInstantiable != 0 {
		if constraint := l.checker.GetBaseConstraintOfType(proven); constraint != nil {
			proven = constraint
		}
	}
	members := []*checker.Type{proven}
	if proven.Flags()&checker.TypeFlagsUnion != 0 {
		members = proven.Types()
	}
	var names []string
	for _, member := range members {
		switch {
		case member.Flags()&checker.TypeFlagsStringLiteral != 0:
			names = append(names, member.AsLiteralType().Value().(string))
		case member.Flags()&checker.TypeFlagsNumberLiteral != 0:
			names = append(names, l.checker.TypeToString(member))
		default:
			return []string{""}
		}
	}
	return names
}

// unionWriteSlot is the type a member's slot accepts: a named property's declared type without
// exactOptionalPropertyTypes' missing marker, or the index the key's kind reaches.
func (l *lowering) unionWriteSlot(member *checker.Type, name string, key *ast.Node) *checker.Type {
	if name != "" {
		if property := l.checker.GetPropertyOfType(member, name); property != nil {
			return checker.Checker_getNonMissingTypeOfSymbol(l.checker, property)
		}
	}
	indexKey := l.checker.GetStringType()
	if key != nil && l.checker.GetTypeAtLocation(key).Flags()&checker.TypeFlagsNumberLike != 0 || name != "" && isArrayIndexName(name) {
		indexKey = l.checker.GetNumberType()
	}
	return l.checker.GetIndexTypeOfType(member, indexKey)
}

// isArrayIndexName reports whether a property name is a canonical array index, as tsc reads "0".
func isArrayIndexName(name string) bool {
	if name == "" || len(name) > 1 && name[0] == '0' {
		return false
	}
	for _, character := range name {
		if character < '0' || character > '9' {
			return false
		}
	}
	return true
}
