// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
)

// assignment lowers =, and the compound assignments, to a local.
func (l *lowering) assignment(node *ast.Node) ([]ir.Statement, error) {
	if value, known, err := l.processValue(node); known {
		if err != nil {
			return nil, err
		}
		return []ir.Statement{ir.Evaluate{Value: value}}, nil
	}
	binary := node.AsBinaryExpression()
	operator, isCompound := compoundAssignments[binary.OperatorToken.Kind]
	if binary.OperatorToken.Kind != ast.KindEqualsToken && !isCompound {
		return nil, l.notYet(node, describe(node)+" as a statement")
	}
	target := ast.SkipParentheses(binary.Left)
	if l.uninitializedInitializer(binary.Right) && !ast.IsIdentifier(target) && target.Kind != ast.KindPropertyAccessExpression {
		return nil, l.notYet(target, "a placeholder reset without a represented variable or field slot")
	}
	if isCompound && l.enumNeverIdentity(target, map[*ast.Node]bool{}) != nil {
		value, err := l.expression(target)
		return []ir.Statement{ir.Evaluate{Value: value}}, err
	}
	if target.Kind == ast.KindPropertyAccessExpression && !l.namespaceMember(target) {
		if !isCompound {
			if body, handled, err := l.entriesRecordWrite(target, binary.Right); handled {
				return body, err
			}
		}
		if isCompound {
			return l.updateProperty(node, target, operator, binary.Right)
		}
		return l.setProperty(target, binary.Right)
	}
	// array[index] += value never reaches here: the element may be missing, so the checker refuses it
	// (noUncheckedIndexedAccess).
	if target.Kind == ast.KindElementAccessExpression && binary.OperatorToken.Kind == ast.KindEqualsToken {
		if body, handled, err := l.constructionIndexWrite(target, binary.Right); handled {
			return body, err
		}
		if body, handled, err := l.typedArrayWrite(target, binary.Right); handled {
			return body, err
		}
		return l.setIndex(target, binary.Right)
	}
	if target.Kind == ast.KindArrayLiteralExpression && binary.OperatorToken.Kind == ast.KindEqualsToken {
		return l.destructuringAssignment(target, binary.Right)
	}
	if target.Kind == ast.KindObjectLiteralExpression && binary.OperatorToken.Kind == ast.KindEqualsToken {
		return l.objectDestructuringAssignment(target, binary.Right)
	}
	if target.Kind == ast.KindElementAccessExpression && isCompound {
		return l.updateIndex(node, target, operator, binary.Right)
	}
	local, isLocal := l.local(target)
	if (!ast.IsIdentifier(target) && !l.namespaceMember(target)) || !isLocal {
		return nil, l.notYet(target, "assigning to "+describe(target))
	}
	if l.result.Locals[local].NestedFunction != 0 {
		return nil, l.notYet(target, "rebinding a nested function declaration")
	}
	if l.caught[l.symbol(target)] {
		return nil, l.notYet(target, "assigning to what a catch caught")
	}
	if l.alwaysUndefined[l.symbol(target)] {
		// Its type is unknown, so anything could be written to it, and it holds only undefined.
		return nil, l.notYet(target, "assigning to a parameter that only ever receives undefined")
	}
	if !isCompound && l.uninitializedInitializer(binary.Right) {
		if err := l.placeholderResetRefusal(target, local); err != nil {
			return nil, err
		}
		l.result.Locals[local].Uninitialized = true
		return []ir.Statement{ir.Assign{Local: local, Value: l.placeholderInitialValue(binary.Right, l.result.Locals[local].Type), Checked: l.checked(local), Uninitialized: true}}, nil
	}
	value, err := l.expression(binary.Right)
	if err != nil {
		return nil, err
	}
	if isCompound {
		current := ir.Expression(ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.checked(local)})
		if origin := l.result.Locals[local].Placeholder; origin != "" {
			current = l.placeholderRead(target, ir.Read{Local: local, Of: l.result.Locals[local].Type, Checked: l.checked(local), Unset: true}, origin)
		}
		if operator == ast.KindPlusToken {
			current, value = l.spelled(target, current), l.spelled(binary.Right, value)
		}
		if value, err = l.combine(node, operator, current, value); err != nil {
			return nil, err
		}
	}
	statements := l.namespaceReadyStatements(target, false)
	if !isCompound && target.Kind == ast.KindPropertyAccessExpression && l.namespaceMember(target) {
		// A simple property assignment evaluates its RHS before PutValue discovers
		// an undefined receiver. Earlier containers in a nested chain are reads.
		statements = l.namespaceReadyStatements(target.Expression(), false)
		temporary := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "namespace_assignment", Type: value.Type(), Function: l.functionIndex})
		statements = append(statements, ir.Declare{Local: temporary, Value: value})
		statements = append(statements, l.namespaceReadyStatements(target, true)...)
		value = ir.Read{Local: temporary, Of: value.Type()}
	}
	return append(statements, ir.Assign{Local: local, Value: fit(value, l.result.Locals[local].Type), Checked: l.checked(local) && !l.result.Locals[local].NamespaceState}), nil
}

// tupleField reads element index of the tuple value, held in the local held, as the element is stored
// (tupleElementStorage), and makes it what the name it goes to holds: a string element going to a
// string | number name is boxed on the way.
func (l *lowering) tupleField(where *ast.Node, valueNode *ast.Node, value ir.Expression, held int, index int, to ir.Type) (ir.Expression, error) {
	of, slot, err := l.tupleElementStorage(where, valueNode, value, index)
	if err != nil {
		return nil, err
	}
	// A union is one counted reference, read as destructureFrom reads one: when every member is one a
	// template writes.
	if slotless(of) && !(of == ir.Union && l.writable(slot)) {
		return nil, l.notYet(where, "a tuple element of type "+l.checker.TypeToString(slot))
	}
	field := fit(ir.Expression(ir.Property{Object: ir.Read{Local: held, Of: ir.Object}, Name: strconv.Itoa(index), Of: of}), to)
	if field.Type() != to {
		return nil, l.notYet(where, "a tuple element of type "+l.checker.TypeToString(slot)+" given to a "+typeName(to))
	}
	return field, nil
}

// tupleElementStorage is how element index of the tuple value is stored, and the type it's stored as.
// A literal written here is stored as the slots it's written into, the pattern's, while the checker
// narrows each element by what's written in it: with both lets number | string, [second, first] is
// [string, number], and reading the first's number | string slot as a number printed a box's pointer
// bits (#fsb6n82). So a literal's elements are read as its lowered fields hold them, and the slot it
// was made for must say the same. A value read from where it's held (a local, a field, an element, a
// call's result) is stored as its own type. Anything else may hold a literal the checker narrows the
// same way ((0, [second, first]), or a conditional between two), so it is not yet.
func (l *lowering) tupleElementStorage(where *ast.Node, valueNode *ast.Node, value ir.Expression, index int) (ir.Type, *checker.Type, error) {
	written := ast.SkipParentheses(valueNode)
	var elements []*checker.Type
	if literal, isLiteral := value.(ir.ObjectLiteral); isLiteral {
		tuple := l.tupleType(written)
		if !literal.Tuple || written.Kind != ast.KindArrayLiteralExpression || tuple == nil {
			return 0, nil, l.notYet(where, "destructuring a value whose element storage isn't known")
		}
		elements = l.checker.GetTypeArguments(tuple)
		if index >= len(elements) {
			return 0, nil, l.notYet(where, "a name past its tuple's elements")
		}
		for _, field := range literal.Fields {
			if field.Name != strconv.Itoa(index) {
				continue
			}
			if of, isKnown := l.representation(elements[index]); !isKnown || of != field.Value.Type() {
				return 0, nil, l.notYet(where, "destructuring a value whose element storage isn't known")
			}
			return field.Value.Type(), elements[index], nil
		}
		return 0, nil, l.notYet(where, "destructuring a value whose element storage isn't known")
	}
	switch written.Kind {
	case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression:
	default:
		return 0, nil, l.notYet(where, "destructuring a value whose element storage isn't known")
	}
	elements = l.checker.GetTypeArguments(l.checker.GetTypeAtLocation(valueNode))
	if index >= len(elements) {
		return 0, nil, l.notYet(where, "a name past its tuple's elements")
	}
	of, isKnown := l.representation(elements[index])
	if !isKnown {
		return 0, nil, l.notYet(where, "a tuple element of type "+l.checker.TypeToString(elements[index]))
	}
	return of, elements[index], nil
}

// destructuringAssignment lowers [a, b] = tuple: the tuple is evaluated whole and held, then each
// name is assigned its field, in order, as JavaScript assigns them. A name left out ([, b]) is
// skipped. Anything but plain names, or a value that isn't a tuple, is not lowered yet.
func (l *lowering) destructuringAssignment(pattern *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {
	if !checker.IsTupleType(l.checker.GetTypeAtLocation(valueNode)) {
		return nil, l.notYet(pattern, "destructuring other than a tuple")
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.Object {
		return nil, l.notYet(pattern, "destructuring a "+typeName(value.Type()))
	}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "tuple", Type: ir.Object, Function: l.functionIndex})
	statements := []ir.Statement{ir.Declare{Local: held, Value: value}}
	for index, element := range pattern.AsArrayLiteralExpression().Elements.Nodes {
		if element.Kind == ast.KindOmittedExpression {
			continue
		}
		local, isLocal := l.local(element)
		if !ast.IsIdentifier(element) || !isLocal || l.alwaysUndefined[l.symbol(element)] {
			return nil, l.notYet(element, "assigning to "+describe(element)+" in a destructuring assignment")
		}
		if l.result.Locals[local].NestedFunction != 0 {
			return nil, l.notYet(element, "rebinding a nested function declaration")
		}
		field, err := l.tupleField(element, valueNode, value, held, index, l.result.Locals[local].Type)
		if err != nil {
			return nil, err
		}
		statements = append(statements, ir.Assign{Local: local, Value: field, Checked: l.checked(local)})
	}
	// The held tuple is the block's, released when it ends.
	return []ir.Statement{ir.Block{Body: statements}}, nil
}

// objectDestructuringAssignment lowers { a, b: c } = value: the value is evaluated whole and held,
// then each name is assigned its field, read by name, in order, as JavaScript assigns them, and made
// what the name holds: a number field going to a number | string name is boxed on the way. A target
// that isn't a plain name, a default, a rest, a computed key, or a value that isn't one object, is
// not lowered yet.
func (l *lowering) objectDestructuringAssignment(pattern *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {
	destructured := l.checker.GetTypeAtLocation(valueNode)
	if destructured.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		return nil, l.notYet(pattern, "destructuring a union or an intersection into { names }")
	}
	switch ast.SkipParentheses(valueNode).Kind {
	case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression:
		// A value with storage of its own, held as its type says.
	default:
		// A literal written here, ({ a, b } = { a: b, b: a }), is stored as the pattern it's written
		// into, a number | string slot, while its own type says number: a field read by that type reads
		// the box as a number.
		return nil, l.notYet(valueNode, "destructuring a value made in the assignment itself")
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.Object {
		return nil, l.notYet(pattern, "destructuring a "+typeName(value.Type()))
	}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "destructured", Type: ir.Object, Function: l.functionIndex})
	statements := []ir.Statement{ir.Declare{Local: held, Value: value}}
	for _, property := range pattern.AsObjectLiteralExpression().Properties.Nodes {
		var target *ast.Node
		switch property.Kind {
		case ast.KindShorthandPropertyAssignment:
			// { a } reads a into a.
			if property.AsShorthandPropertyAssignment().ObjectAssignmentInitializer != nil {
				return nil, l.notYet(property, "a default in a destructuring assignment")
			}
			target = property.Name()
		case ast.KindPropertyAssignment:
			// { a: b } reads a into b.
			target = ast.SkipParentheses(property.AsPropertyAssignment().Initializer)
		default:
			return nil, l.notYet(property, describe(property)+" in a destructuring assignment")
		}
		name := property.Name()
		if name == nil || !ast.IsIdentifier(name) && name.Kind != ast.KindStringLiteral {
			return nil, l.notYet(property, "a computed field name")
		}
		if !ast.IsIdentifier(target) {
			return nil, l.notYet(target, "assigning to "+describe(target)+" in a destructuring assignment")
		}
		// A shorthand's name is the pattern's property; flagValueSymbol is the variable it writes.
		variable := l.flagValueSymbol(target)
		local, isLocal := l.localOf(target, variable)
		if !isLocal || l.alwaysUndefined[variable] {
			return nil, l.notYet(target, "assigning to "+describe(target)+" in a destructuring assignment")
		}
		if l.result.Locals[local].NestedFunction != 0 {
			return nil, l.notYet(target, "rebinding a nested function declaration")
		}
		if l.caught[variable] {
			return nil, l.notYet(target, "assigning to what a catch caught")
		}
		field := name.Text()
		symbol, err := l.destructuredField(property, destructured, field)
		if err != nil {
			return nil, err
		}
		fieldType := l.checker.GetTypeOfSymbol(symbol)
		of, isKnown := l.representation(fieldType)
		if !isKnown || slotless(of) {
			return nil, l.notYet(property, "a destructured field of type "+l.checker.TypeToString(fieldType))
		}
		read := ir.Property{Object: ir.Read{Local: held, Of: ir.Object}, Name: field, Of: of, Absent: symbol.Flags&ast.SymbolFlagsOptional != 0}
		read.View = sourceExpression(property) + " (field " + field + ")"
		read.ViewReceiverTypeID = l.viewFieldReceiverTypeID(nil, destructured)
		read.ViewTypeID = l.viewTypeID(fieldType)
		read.ViewWhere = l.program.Where(property)
		read.ViewType = l.checker.TypeToString(fieldType)
		read.ViewAllowed = l.viewLiterals(fieldType)
		to := l.result.Locals[local].Type
		assigned := fit(ir.Expression(read), to)
		if assigned.Type() != to {
			return nil, l.notYet(property, "a destructured field of type "+l.checker.TypeToString(fieldType)+" given to a "+typeName(to))
		}
		statements = append(statements, ir.Assign{Local: local, Value: assigned, Checked: l.checked(local)})
	}
	// The held value is the block's, released when it ends.
	return []ir.Statement{ir.Block{Body: statements}}, nil
}

var compoundAssignments = map[ast.Kind]ast.Kind{
	ast.KindPlusEqualsToken:             ast.KindPlusToken,
	ast.KindMinusEqualsToken:            ast.KindMinusToken,
	ast.KindAsteriskEqualsToken:         ast.KindAsteriskToken,
	ast.KindSlashEqualsToken:            ast.KindSlashToken,
	ast.KindPercentEqualsToken:          ast.KindPercentToken,
	ast.KindAsteriskAsteriskEqualsToken: ast.KindAsteriskAsteriskToken,

	ast.KindAmpersandEqualsToken:                         ast.KindAmpersandToken,
	ast.KindBarEqualsToken:                               ast.KindBarToken,
	ast.KindCaretEqualsToken:                             ast.KindCaretToken,
	ast.KindLessThanLessThanEqualsToken:                  ast.KindLessThanLessThanToken,
	ast.KindGreaterThanGreaterThanEqualsToken:            ast.KindGreaterThanGreaterThanToken,
	ast.KindGreaterThanGreaterThanGreaterThanEqualsToken: ast.KindGreaterThanGreaterThanGreaterThanToken,
}

// increment lowers ++ and -- on a number local or field, as a statement, where prefix and postfix
// agree.
func (l *lowering) increment(node *ast.Node) ([]ir.Statement, error) {
	var operator ast.Kind
	var operand *ast.Node
	if node.Kind == ast.KindPrefixUnaryExpression {
		operator, operand = node.AsPrefixUnaryExpression().Operator, node.AsPrefixUnaryExpression().Operand
	} else {
		operator, operand = node.AsPostfixUnaryExpression().Operator, node.AsPostfixUnaryExpression().Operand
	}
	if operator != ast.KindPlusPlusToken && operator != ast.KindMinusMinusToken {
		return nil, l.notYet(node, describe(node)+" as a statement")
	}
	operand = ast.SkipParentheses(operand)
	if operand.Kind == ast.KindNonNullExpression {
		assertion := operand
		target := ast.SkipParentheses(assertion.AsNonNullExpression().Expression)
		step := ir.Add
		if operator == ast.KindMinusMinusToken {
			step = ir.Subtract
		}
		if target.Kind != ast.KindPropertyAccessExpression {
			return nil, l.notYet(target, "incrementing a non-null target other than a stored numeric field")
		}
		object, err := l.expression(target.AsPropertyAccessExpression().Expression)
		if err != nil {
			return nil, err
		}
		if object.Type() != ir.Object {
			return nil, l.notYet(target, "incrementing a non-null field of a "+typeName(object.Type()))
		}
		field := l.checker.GetSymbolAtLocation(target.Name())
		if field == nil || accessorSymbol(field) {
			return nil, l.notYet(target, "incrementing a non-null accessor field")
		}
		of, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if !known || (of != ir.Number && of != ir.MaybeNumber) {
			return nil, l.notYet(target, "incrementing a non-null field without numeric storage")
		}
		// Hold the receiver once. Check the stored value before updating the slot.
		held := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "increment_object", Type: ir.Object, Function: l.functionIndex})
		l.noteLocal(held, l.concrete(l.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression)), target)
		object = l.privateStaticReceiver(target.Name(), object, false)
		read := ir.Read{Local: held, Of: ir.Object}
		name := l.fieldName(target.Name())
		current := l.readObjectField(target, ir.Property{Object: read, Name: name, Of: of, Class: l.classOf(target)})
		current, err = l.nonNullValue(assertion, current)
		if err != nil {
			return nil, err
		}
		if current.Type() != ir.Number {
			return nil, l.notYet(target, "incrementing a non-null field whose checked value is not a number")
		}
		updated := ir.Binary{Operator: step, Left: current, Right: ir.NumberConstant{Value: 1}}
		return []ir.Statement{ir.Block{Body: []ir.Statement{
			ir.Declare{Local: held, Value: object},
			ir.SetProperty{Object: read, Name: name, Value: fit(updated, of), Class: l.classOf(target), Site: l.writeSite(target.AsPropertyAccessExpression().Expression), ViewReceiverTypeID: l.viewWriteReceiverTypeID(target, target.AsPropertyAccessExpression().Expression, name)},
		}}}, nil
	}
	if l.enumNeverIdentity(operand, map[*ast.Node]bool{}) != nil {
		value, err := l.expression(operand)
		return []ir.Statement{ir.Evaluate{Value: value}}, err
	}
	if operand.Kind == ast.KindPropertyAccessExpression && !l.namespaceMember(operand) {
		step := ast.KindPlusToken
		if operator == ast.KindMinusMinusToken {
			step = ast.KindMinusToken
		}
		return l.updateProperty(node, operand, step, nil)
	}
	local, isLocal := l.local(operand)
	if (!ast.IsIdentifier(operand) && !l.namespaceMember(operand)) || !isLocal {
		return nil, l.notYet(operand, "incrementing "+describe(operand))
	}
	step := ir.Add
	if operator == ast.KindMinusMinusToken {
		step = ir.Subtract
	}
	current := ir.Read{Local: local, Of: ir.Number, Checked: l.checked(local)}
	return append(l.namespaceReadyStatements(operand, false), ir.Assign{Local: local, Value: ir.Binary{Operator: step, Left: current, Right: ir.NumberConstant{Value: 1}}, Checked: l.checked(local)}), nil
}
