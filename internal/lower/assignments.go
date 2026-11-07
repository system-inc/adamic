// Functions and declarations moved unchanged from lower.go.
package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
)

// assignment lowers =, and the compound assignments, to a local.
func (lowering *lowering) assignment(node *ast.Node) ([]ir.Statement, error) {
	if value, known, processValueError := lowering.processValue(node); known {
		if processValueError != nil {
			return nil, processValueError
		}
		return []ir.Statement{ir.Evaluate{Value: value}}, nil
	}
	binary := node.AsBinaryExpression()
	operator, isCompound := compoundAssignments[binary.OperatorToken.Kind]
	if binary.OperatorToken.Kind != ast.KindEqualsToken && !isCompound {
		return nil, lowering.notYet(node, describe(node)+" as a statement")
	}
	target := ast.SkipParentheses(binary.Left)
	if lowering.uninitializedInitializer(binary.Right) &&
		!ast.IsIdentifier(target) &&
		target.Kind != ast.KindPropertyAccessExpression {
		return nil, lowering.notYet(target, "a placeholder reset without a represented variable or field slot")
	}
	if isCompound && lowering.enumNeverIdentity(target, map[*ast.Node]bool{}) != nil {
		value, expressionError := lowering.expression(target)
		return []ir.Statement{ir.Evaluate{Value: value}}, expressionError
	}
	if target.Kind == ast.KindPropertyAccessExpression && !lowering.namespaceMember(target) {
		if !isCompound {
			if body, handled, entriesRecordWriteError := lowering.entriesRecordWrite(target, binary.Right); handled {
				return body, entriesRecordWriteError
			}
		}
		if isCompound {
			return lowering.updateProperty(node, target, operator, binary.Right)
		}
		return lowering.setProperty(target, binary.Right)
	}
	// array[index] += value never reaches here: the element may be missing, so the checker refuses it
	// (noUncheckedIndexedAccess).
	if target.Kind == ast.KindElementAccessExpression && binary.OperatorToken.Kind == ast.KindEqualsToken {
		if body, handled, constructionIndexWriteError := lowering.constructionIndexWrite(
			target,
			binary.Right,
		); handled {
			return body, constructionIndexWriteError
		}
		if body, handled, typedArrayWriteError := lowering.typedArrayWrite(target, binary.Right); handled {
			return body, typedArrayWriteError
		}
		return lowering.setIndex(target, binary.Right)
	}
	if target.Kind == ast.KindArrayLiteralExpression && binary.OperatorToken.Kind == ast.KindEqualsToken {
		return lowering.destructuringAssignment(target, binary.Right)
	}
	if target.Kind == ast.KindObjectLiteralExpression && binary.OperatorToken.Kind == ast.KindEqualsToken {
		return lowering.objectDestructuringAssignment(target, binary.Right)
	}
	if target.Kind == ast.KindElementAccessExpression && isCompound {
		return lowering.updateIndex(node, target, operator, binary.Right)
	}
	local, isLocal := lowering.local(target)
	if (!ast.IsIdentifier(target) && !lowering.namespaceMember(target)) || !isLocal {
		return nil, lowering.notYet(target, "assigning to "+describe(target))
	}
	if lowering.result.Locals[local].NestedFunction != 0 {
		return nil, lowering.notYet(target, "rebinding a nested function declaration")
	}
	if lowering.caught[lowering.symbol(target)] {
		return nil, lowering.notYet(target, "assigning to what a catch caught")
	}
	if lowering.alwaysUndefined[lowering.symbol(target)] {
		// Its type is unknown, so anything could be written to it, and it holds only undefined.
		return nil, lowering.notYet(target, "assigning to a parameter that only ever receives undefined")
	}
	if !isCompound && lowering.uninitializedInitializer(binary.Right) {
		if placeholderResetRefusalError := lowering.placeholderResetRefusal(
			target,
			local,
		); placeholderResetRefusalError != nil {
			return nil, placeholderResetRefusalError
		}
		lowering.result.Locals[local].Uninitialized = true
		return []ir.Statement{ir.Assign{
			Local:         local,
			Value:         lowering.placeholderInitialValue(binary.Right, lowering.result.Locals[local].Type),
			Checked:       lowering.checked(local),
			Uninitialized: true,
		}}, nil
	}
	value, expressionError := lowering.expression(binary.Right)
	if expressionError != nil {
		return nil, expressionError
	}
	if isCompound {
		current := ir.Expression(
			ir.Read{Local: local, Of: lowering.result.Locals[local].Type, Checked: lowering.checked(local)},
		)
		if origin := lowering.result.Locals[local].Placeholder; origin != "" {
			current = lowering.placeholderRead(
				target,
				ir.Read{
					Local:   local,
					Of:      lowering.result.Locals[local].Type,
					Checked: lowering.checked(local),
					Unset:   true,
				},
				origin,
			)
		}
		if operator == ast.KindPlusToken {
			current, value = lowering.spelled(target, current), lowering.spelled(binary.Right, value)
		}
		var combineError error
		if value, combineError = lowering.combine(node, operator, current, value); combineError != nil {
			return nil, combineError
		}
	}
	statements := lowering.namespaceReadyStatements(target, false)
	if !isCompound && target.Kind == ast.KindPropertyAccessExpression && lowering.namespaceMember(target) {
		// A simple property assignment evaluates its RHS before PutValue discovers
		// an undefined receiver. Earlier containers in a nested chain are reads.
		statements = lowering.namespaceReadyStatements(target.Expression(), false)
		temporary := len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: "namespace_assignment", Type: value.Type(), Function: lowering.functionIndex},
		)
		statements = append(statements, ir.Declare{Local: temporary, Value: value})
		statements = append(statements, lowering.namespaceReadyStatements(target, true)...)
		value = ir.Read{Local: temporary, Of: value.Type()}
	}
	return append(statements, ir.Assign{
		Local:   local,
		Value:   fit(value, lowering.result.Locals[local].Type),
		Checked: lowering.checked(local) && !lowering.result.Locals[local].NamespaceState,
	}), nil
}

// tupleField reads element index of the tuple value, held in the local held, as the element is stored
// (tupleElementStorage), and makes it what the name it goes to holds: a string element going to a
// string | number name is boxed on the way.
func (lowering *lowering) tupleField(
	where *ast.Node,
	valueNode *ast.Node,
	value ir.Expression,
	held int,
	index int,
	to ir.Type,
) (ir.Expression, error) {
	of, slot, tupleElementStorageError := lowering.tupleElementStorage(where, valueNode, value, index)
	if tupleElementStorageError != nil {
		return nil, tupleElementStorageError
	}
	// A union is one counted reference, read as destructureFrom reads one: when every member is one a
	// template writes.
	if slotless(of) && !(of == ir.Union && lowering.writable(slot)) {
		return nil, lowering.notYet(where, "a tuple element of type "+lowering.checker.TypeToString(slot))
	}
	field := fit(
		ir.Expression(ir.Property{Object: ir.Read{Local: held, Of: ir.Object}, Name: strconv.Itoa(index), Of: of}),
		to,
	)
	if field.Type() != to {
		return nil, lowering.notYet(
			where,
			"a tuple element of type "+lowering.checker.TypeToString(slot)+" given to a "+typeName(to),
		)
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
func (lowering *lowering) tupleElementStorage(
	where *ast.Node,
	valueNode *ast.Node,
	value ir.Expression,
	index int,
) (ir.Type, *checker.Type, error) {
	written := ast.SkipParentheses(valueNode)
	var elements []*checker.Type
	if literal, isLiteral := value.(ir.ObjectLiteral); isLiteral {
		tuple := lowering.tupleType(written)
		if !literal.Tuple || written.Kind != ast.KindArrayLiteralExpression || tuple == nil {
			return 0, nil, lowering.notYet(where, "destructuring a value whose element storage isn't known")
		}
		elements = lowering.checker.GetTypeArguments(tuple)
		if index >= len(elements) {
			return 0, nil, lowering.notYet(where, "a name past its tuple's elements")
		}
		for _, field := range literal.Fields {
			if field.Name != strconv.Itoa(index) {
				continue
			}
			if of, isKnown := lowering.representation(elements[index]); !isKnown || of != field.Value.Type() {
				return 0, nil, lowering.notYet(where, "destructuring a value whose element storage isn't known")
			}
			return field.Value.Type(), elements[index], nil
		}
		return 0, nil, lowering.notYet(where, "destructuring a value whose element storage isn't known")
	}
	switch written.Kind {
	case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression:
	default:
		return 0, nil, lowering.notYet(where, "destructuring a value whose element storage isn't known")
	}
	elements = lowering.checker.GetTypeArguments(lowering.checker.GetTypeAtLocation(valueNode))
	if index >= len(elements) {
		return 0, nil, lowering.notYet(where, "a name past its tuple's elements")
	}
	of, isKnown := lowering.representation(elements[index])
	if !isKnown {
		return 0, nil, lowering.notYet(where, "a tuple element of type "+lowering.checker.TypeToString(elements[index]))
	}
	return of, elements[index], nil
}

// destructuringAssignment lowers [a, b] = tuple: the tuple is evaluated whole and held, then each
// name is assigned its field, in order, as JavaScript assigns them. A name left out ([, b]) is
// skipped. Anything but plain names, or a value that isn't a tuple, is not lowered yet.
func (lowering *lowering) destructuringAssignment(pattern *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {
	if !checker.IsTupleType(lowering.checker.GetTypeAtLocation(valueNode)) {
		return nil, lowering.notYet(pattern, "destructuring other than a tuple")
	}
	value, expressionError := lowering.expression(valueNode)
	if expressionError != nil {
		return nil, expressionError
	}
	if value.Type() != ir.Object {
		return nil, lowering.notYet(pattern, "destructuring a "+typeName(value.Type()))
	}
	held := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "tuple", Type: ir.Object, Function: lowering.functionIndex},
	)
	statements := []ir.Statement{ir.Declare{Local: held, Value: value}}
	for index, element := range pattern.AsArrayLiteralExpression().Elements.Nodes {
		if element.Kind == ast.KindOmittedExpression {
			continue
		}
		local, isLocal := lowering.local(element)
		if !ast.IsIdentifier(element) || !isLocal || lowering.alwaysUndefined[lowering.symbol(element)] {
			return nil, lowering.notYet(element, "assigning to "+describe(element)+" in a destructuring assignment")
		}
		if lowering.result.Locals[local].NestedFunction != 0 {
			return nil, lowering.notYet(element, "rebinding a nested function declaration")
		}
		field, tupleFieldError := lowering.tupleField(
			element,
			valueNode,
			value,
			held,
			index,
			lowering.result.Locals[local].Type,
		)
		if tupleFieldError != nil {
			return nil, tupleFieldError
		}
		statements = append(statements, ir.Assign{Local: local, Value: field, Checked: lowering.checked(local)})
	}
	// The held tuple is the block's, released when it ends.
	return []ir.Statement{ir.Block{Body: statements}}, nil
}

// objectDestructuringAssignment lowers { a, b: c } = value: the value is evaluated whole and held,
// then each name is assigned its field, read by name, in order, as JavaScript assigns them, and made
// what the name holds: a number field going to a number | string name is boxed on the way. A target
// that isn't a plain name, a default, a rest, a computed key, or a value that isn't one object, is
// not lowered yet.
func (lowering *lowering) objectDestructuringAssignment(
	pattern *ast.Node,
	valueNode *ast.Node,
) ([]ir.Statement, error) {
	destructured := lowering.checker.GetTypeAtLocation(valueNode)
	if destructured.Flags()&checker.TypeFlagsUnionOrIntersection != 0 {
		return nil, lowering.notYet(pattern, "destructuring a union or an intersection into { names }")
	}
	switch ast.SkipParentheses(valueNode).Kind {
	case ast.KindIdentifier, ast.KindPropertyAccessExpression, ast.KindElementAccessExpression, ast.KindCallExpression:
		// A value with storage of its own, held as its type says.
	default:
		// A literal written here, ({ a, b } = { a: b, b: a }), is stored as the pattern it's written
		// into, a number | string slot, while its own type says number: a field read by that type reads
		// the box as a number.
		return nil, lowering.notYet(valueNode, "destructuring a value made in the assignment itself")
	}
	value, expressionError := lowering.expression(valueNode)
	if expressionError != nil {
		return nil, expressionError
	}
	if value.Type() != ir.Object {
		return nil, lowering.notYet(pattern, "destructuring a "+typeName(value.Type()))
	}
	held := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "destructured", Type: ir.Object, Function: lowering.functionIndex},
	)
	statements := []ir.Statement{ir.Declare{Local: held, Value: value}}
	for _, property := range pattern.AsObjectLiteralExpression().Properties.Nodes {
		var target *ast.Node
		switch property.Kind {
		case ast.KindShorthandPropertyAssignment:
			// { a } reads a into a.
			if property.AsShorthandPropertyAssignment().ObjectAssignmentInitializer != nil {
				return nil, lowering.notYet(property, "a default in a destructuring assignment")
			}
			target = property.Name()
		case ast.KindPropertyAssignment:
			// { a: b } reads a into b.
			target = ast.SkipParentheses(property.AsPropertyAssignment().Initializer)
		default:
			return nil, lowering.notYet(property, describe(property)+" in a destructuring assignment")
		}
		name := property.Name()
		if name == nil || !ast.IsIdentifier(name) && name.Kind != ast.KindStringLiteral {
			return nil, lowering.notYet(property, "a computed field name")
		}
		if !ast.IsIdentifier(target) {
			return nil, lowering.notYet(target, "assigning to "+describe(target)+" in a destructuring assignment")
		}
		// A shorthand's name is the pattern's property; flagValueSymbol is the variable it writes.
		variable := lowering.flagValueSymbol(target)
		local, isLocal := lowering.localOf(target, variable)
		if !isLocal || lowering.alwaysUndefined[variable] {
			return nil, lowering.notYet(target, "assigning to "+describe(target)+" in a destructuring assignment")
		}
		if lowering.result.Locals[local].NestedFunction != 0 {
			return nil, lowering.notYet(target, "rebinding a nested function declaration")
		}
		if lowering.caught[variable] {
			return nil, lowering.notYet(target, "assigning to what a catch caught")
		}
		field := name.Text()
		symbol, destructuredFieldError := lowering.destructuredField(property, destructured, field)
		if destructuredFieldError != nil {
			return nil, destructuredFieldError
		}
		fieldType := lowering.checker.GetTypeOfSymbol(symbol)
		of, isKnown := lowering.representation(fieldType)
		if !isKnown || slotless(of) {
			return nil, lowering.notYet(
				property,
				"a destructured field of type "+lowering.checker.TypeToString(fieldType),
			)
		}
		read := ir.Property{
			Object: ir.Read{Local: held, Of: ir.Object},
			Name:   field,
			Of:     of,
			Absent: symbol.Flags&ast.SymbolFlagsOptional != 0,
		}
		read.View = sourceExpression(property) + " (field " + field + ")"
		read.ViewReceiverTypeID = lowering.viewFieldReceiverTypeID(nil, destructured)
		read.ViewTypeID = lowering.viewTypeID(fieldType)
		read.ViewWhere = lowering.program.Where(property)
		read.ViewType = lowering.checker.TypeToString(fieldType)
		read.ViewAllowed = lowering.viewLiterals(fieldType)
		to := lowering.result.Locals[local].Type
		assigned := fit(ir.Expression(read), to)
		if assigned.Type() != to {
			return nil, lowering.notYet(
				property,
				"a destructured field of type "+lowering.checker.TypeToString(fieldType)+" given to a "+typeName(to),
			)
		}
		statements = append(statements, ir.Assign{Local: local, Value: assigned, Checked: lowering.checked(local)})
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
func (lowering *lowering) increment(node *ast.Node) ([]ir.Statement, error) {
	var operator ast.Kind
	var operand *ast.Node
	if node.Kind == ast.KindPrefixUnaryExpression {
		operator, operand = node.AsPrefixUnaryExpression().Operator, node.AsPrefixUnaryExpression().Operand
	} else {
		operator, operand = node.AsPostfixUnaryExpression().Operator, node.AsPostfixUnaryExpression().Operand
	}
	if operator != ast.KindPlusPlusToken && operator != ast.KindMinusMinusToken {
		return nil, lowering.notYet(node, describe(node)+" as a statement")
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
			return nil, lowering.notYet(target, "incrementing a non-null target other than a stored numeric field")
		}
		object, expressionError := lowering.expression(target.AsPropertyAccessExpression().Expression)
		if expressionError != nil {
			return nil, expressionError
		}
		if object.Type() != ir.Object {
			return nil, lowering.notYet(target, "incrementing a non-null field of a "+typeName(object.Type()))
		}
		field := lowering.checker.GetSymbolAtLocation(target.Name())
		if field == nil || accessorSymbol(field) {
			return nil, lowering.notYet(target, "incrementing a non-null accessor field")
		}
		of, known := lowering.representation(lowering.checker.GetTypeOfSymbol(field))
		if !known || (of != ir.Number && of != ir.MaybeNumber) {
			return nil, lowering.notYet(target, "incrementing a non-null field without numeric storage")
		}
		// Hold the receiver once. Check the stored value before updating the slot.
		held := len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: "increment_object", Type: ir.Object, Function: lowering.functionIndex},
		)
		lowering.noteLocal(
			held,
			lowering.concrete(lowering.checker.GetTypeAtLocation(target.AsPropertyAccessExpression().Expression)),
			target,
		)
		object = lowering.privateStaticReceiver(target.Name(), object, false)
		read := ir.Read{Local: held, Of: ir.Object}
		name := lowering.fieldName(target.Name())
		current := lowering.readObjectField(
			target,
			ir.Property{Object: read, Name: name, Of: of, Class: lowering.classOf(target)},
		)
		current, nonNullValueError := lowering.nonNullValue(assertion, current)
		if nonNullValueError != nil {
			return nil, nonNullValueError
		}
		if current.Type() != ir.Number {
			return nil, lowering.notYet(target, "incrementing a non-null field whose checked value is not a number")
		}
		updated := ir.Binary{Operator: step, Left: current, Right: ir.NumberConstant{Value: 1}}
		return []ir.Statement{ir.Block{Body: []ir.Statement{
			ir.Declare{Local: held, Value: object},
			ir.SetProperty{
				Object: read,
				Name:   name,
				Value:  fit(updated, of),
				Class:  lowering.classOf(target),
				Site:   lowering.writeSite(target.AsPropertyAccessExpression().Expression),
				ViewReceiverTypeID: lowering.viewWriteReceiverTypeID(
					target,
					target.AsPropertyAccessExpression().Expression,
					name,
				),
			},
		}}}, nil
	}
	if lowering.enumNeverIdentity(operand, map[*ast.Node]bool{}) != nil {
		value, expressionError := lowering.expression(operand)
		return []ir.Statement{ir.Evaluate{Value: value}}, expressionError
	}
	if operand.Kind == ast.KindPropertyAccessExpression && !lowering.namespaceMember(operand) {
		step := ast.KindPlusToken
		if operator == ast.KindMinusMinusToken {
			step = ast.KindMinusToken
		}
		return lowering.updateProperty(node, operand, step, nil)
	}
	local, isLocal := lowering.local(operand)
	if (!ast.IsIdentifier(operand) && !lowering.namespaceMember(operand)) || !isLocal {
		return nil, lowering.notYet(operand, "incrementing "+describe(operand))
	}
	step := ir.Add
	if operator == ast.KindMinusMinusToken {
		step = ir.Subtract
	}
	current := ir.Read{Local: local, Of: ir.Number, Checked: lowering.checked(local)}
	return append(lowering.namespaceReadyStatements(operand, false), ir.Assign{
		Local:   local,
		Value:   ir.Binary{Operator: step, Left: current, Right: ir.NumberConstant{Value: 1}},
		Checked: lowering.checked(local),
	}), nil
}
