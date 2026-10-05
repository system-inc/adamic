package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// objectLiteral lowers { name: value, ... }, and the one spread 0.1 allows: { ...source, fields },
// where every field replaces one the source's type already has.
func (l *lowering) objectLiteral(node *ast.Node) (ir.Expression, error) {
	literal := ir.ObjectLiteral{}
	for index, property := range node.AsObjectLiteralExpression().Properties.Nodes {
		switch property.Kind {
		case ast.KindSpreadAssignment:
			if index != 0 {
				return nil, &Refused{Where: l.program.Where(property), What: "a spread after the first field", Fix: "spread once, first: { ...source, field: value } (adamic/single-spread)"}
			}
			spread, err := l.expression(property.AsSpreadAssignment().Expression)
			if err != nil {
				return nil, err
			}
			if spread.Type() != ir.Object {
				return nil, l.notYet(property, "spreading a "+typeName(spread.Type()))
			}
			literal.Spread = spread
		case ast.KindPropertyAssignment, ast.KindShorthandPropertyAssignment:
			name := property.Name()
			if !ast.IsIdentifier(name) && name.Kind != ast.KindStringLiteral {
				return nil, l.notYet(name, "a computed field name")
			}
			var value ir.Expression
			var err error
			if property.Kind == ast.KindPropertyAssignment {
				value, err = l.expression(property.AsPropertyAssignment().Initializer)
			} else {
				value, err = l.shorthand(property)
			}
			if err != nil {
				return nil, err
			}
			if literal.Spread != nil && !l.hasProperty(node.AsObjectLiteralExpression().Properties.Nodes[0].AsSpreadAssignment().Expression, name.Text()) {
				return nil, l.notYet(property, "a spread that adds a field the source doesn't have")
			}
			literal.Fields = append(literal.Fields, ir.Field{Name: name.Text(), Value: value})
		default:
			return nil, l.notYet(property, describe(property)+" in an object literal")
		}
	}
	return literal, nil
}

// hasProperty reports whether a value's type has a field of that name.
func (l *lowering) hasProperty(node *ast.Node, name string) bool {
	for _, property := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(node)) {
		if property.Name == name {
			return true
		}
	}
	return false
}

func (l *lowering) arrayLiteral(node *ast.Node) (ir.Expression, error) {
	element, err := l.elementType(node)
	if err != nil {
		return nil, err
	}
	literal := ir.ArrayLiteral{Element: element}
	for _, item := range node.AsArrayLiteralExpression().Elements.Nodes {
		if item.Kind == ast.KindSpreadElement || item.Kind == ast.KindOmittedExpression {
			return nil, l.notYet(item, describe(item)+" in an array literal")
		}
		value, err := l.expression(item)
		if err != nil {
			return nil, err
		}
		literal.Elements = append(literal.Elements, value)
	}
	return literal, nil
}

// elementType is the representation of an array's elements, from the checker's type for the node.
func (l *lowering) elementType(node *ast.Node) (ir.Type, error) {
	arrayType := l.checker.GetTypeAtLocation(node)
	if literal := ast.SkipParentheses(node); literal.Kind == ast.KindArrayLiteralExpression && len(literal.AsArrayLiteralExpression().Elements.Nodes) == 0 {
		// [] is never[] to the checker; what it will hold is the type it's written into, as in
		// const values: number[] = [].
		if contextual := l.checker.GetContextualType(literal, checker.ContextFlagsNone); contextual != nil && l.checker.IsArrayType(contextual) {
			arrayType = contextual
		}
	}
	if !l.checker.IsArrayType(arrayType) {
		return 0, l.notYet(node, "a value of type "+l.checker.TypeToString(arrayType)+" where an array goes")
	}
	element := l.checker.GetElementTypeOfArrayType(arrayType)
	valueType, isKnown := l.representation(element)
	if !isKnown || valueType == ir.Array {
		return 0, l.notYet(node, "an array of "+l.checker.TypeToString(element))
	}
	return valueType, nil
}

// property lowers object.name, array.length, and Math's constants.
func (l *lowering) property(node *ast.Node) (ir.Expression, error) {
	access := node.AsPropertyAccessExpression()
	name := node.Name().Text()
	if access.QuestionDotToken == nil && node.Flags&ast.NodeFlagsOptionalChain != 0 {
		// The rest of a chain after a ?., which short-circuits with it.
		return nil, l.notYet(node, "an optional chain longer than one step")
	}
	if l.isLibraryGlobal(access.Expression, "Math") {
		switch name {
		case "PI":
			return ir.NumberConstant{Value: 3.141592653589793}, nil
		case "E":
			return ir.NumberConstant{Value: 2.718281828459045}, nil
		}
		return nil, l.notYet(node, "Math."+name)
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, err
	}
	switch {
	case object.Type() == ir.Array && name == "length":
		return ir.Length{Array: object}, nil
	case object.Type() == ir.String && name == "length":
		return ir.StringLength{Value: object}, nil
	case object.Type() == ir.Map && name == "size":
		return ir.MapSize{Map: object}, nil
	case object.Type() == ir.Object:
		of, err := l.typeOf(node)
		if err != nil {
			return nil, err
		}
		optional := access.QuestionDotToken != nil
		if optional && !of.IsReference() {
			return nil, l.notYet(node, "?. to a "+typeName(of)+", which would be "+typeName(of)+" | undefined")
		}
		return ir.Property{Object: object, Name: name, Of: of, Optional: optional}, nil
	}
	if access.QuestionDotToken != nil {
		return nil, l.notYet(node, "optional chaining on a "+typeName(object.Type()))
	}
	return nil, l.notYet(node, "."+name+" on a "+typeName(object.Type()))
}

// isLibraryGlobal reports whether a node names one of the library's globals, rather than something
// of the program's that happens to share its name.
func (l *lowering) isLibraryGlobal(node *ast.Node, name string) bool {
	node = ast.SkipParentheses(node)
	if !ast.IsIdentifier(node) || node.Text() != name {
		return false
	}
	symbol := l.checker.GetSymbolAtLocation(node)
	return symbol != nil && len(symbol.Declarations) > 0 && load.IsLibrary(ast.GetSourceFileOfNode(symbol.Declarations[0]))
}

// mathFunctions are the Math functions 0.1 has (docs/0.1.md), with how many arguments each takes;
// -1 is any number.
var mathFunctions = map[string]int{
	"abs": 1, "ceil": 1, "floor": 1, "round": 1, "trunc": 1, "sign": 1, "sqrt": 1,
	"pow": 2, "max": -1, "min": -1,
}

// builtin lowers a call to Math or a number's toFixed. isBuiltin is false for any other call.
func (l *lowering) builtin(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	receiver, name := callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
	isMath := l.isLibraryGlobal(receiver, "Math")
	receiverType, _ := l.representation(l.checker.GetTypeAtLocation(receiver))
	isToFixed := name == "toFixed" && receiverType == ir.Number
	if receiverType == ir.Array && (name == "push" || name == "join") {
		return l.arrayMethod(node, receiver, name)
	}
	if receiverType == ir.Map && (name == "get" || name == "set" || name == "has" || name == "delete") {
		return l.mapMethod(node, receiver, name)
	}
	if receiverType == ir.String && (name == "trim" || name == "charCodeAt") {
		return l.stringMethod(node, receiver, name)
	}
	if !isMath && !isToFixed {
		return nil, false, nil
	}
	// JavaScript's order: the receiver, then the arguments.
	var value ir.Expression
	if isToFixed {
		var err error
		if value, err = l.expression(receiver); err != nil {
			return nil, true, err
		}
	}
	arguments := []ir.Expression{}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		if lowered.Type() != ir.Number {
			return nil, true, l.notYet(argument, "a "+typeName(lowered.Type())+" argument to "+name)
		}
		arguments = append(arguments, lowered)
	}
	if isMath {
		count, isKnown := mathFunctions[name]
		if !isKnown {
			return nil, true, l.notYet(node, "Math."+name)
		}
		if count >= 0 && len(arguments) != count {
			return nil, true, l.notYet(node, "Math."+name+" with other than its arguments")
		}
		return ir.MathCall{Function: name, Arguments: arguments}, true, nil
	}
	if len(arguments) > 1 {
		return nil, true, l.notYet(node, "toFixed with more than one argument")
	}
	digits := ir.Expression(ir.NumberConstant{Value: 0})
	if len(arguments) == 1 {
		digits = arguments[0]
	}
	return ir.ToFixed{Value: value, Digits: digits}, true, nil
}

// forOf lowers for (const element of array).
func (l *lowering) forOf(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsForInOrOfStatement()
	if statement.AwaitModifier != nil {
		return nil, l.notYet(node, "for await")
	}
	initializer := statement.Initializer
	if initializer.Kind != ast.KindVariableDeclarationList || initializer.Flags&ast.NodeFlagsBlockScoped == 0 {
		return nil, l.notYet(initializer, "a for...of that doesn't declare its variable with const or let")
	}
	declarations := initializer.AsVariableDeclarationList().Declarations.Nodes
	if len(declarations) != 1 || !ast.IsIdentifier(declarations[0].Name()) {
		return nil, l.notYet(initializer, "a destructuring for...of")
	}
	iterable, err := l.expression(statement.Expression)
	if err != nil {
		return nil, err
	}
	var element ir.Type
	switch iterable.Type() {
	case ir.Array:
		if element, err = l.elementType(statement.Expression); err != nil {
			return nil, err
		}
	case ir.String:
		// A string's elements are its code points, each a string.
		element = ir.String
	default:
		return nil, l.notYet(statement.Expression, "for...of over a "+typeName(iterable.Type()))
	}
	local, err := l.declareLocal(declarations[0].Name())
	if err != nil {
		return nil, err
	}
	body, err := l.statement(statement.Statement)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.ForOf{Iterable: iterable, Element: element, Local: local, Body: body}}, nil
}

// switchStatement lowers switch, whose cases 0.1 requires to be constants.
func (l *lowering) switchStatement(node *ast.Node) ([]ir.Statement, error) {
	statement := node.AsSwitchStatement()
	value, err := l.expression(statement.Expression)
	if err != nil {
		return nil, err
	}
	lowered := ir.Switch{Value: value}
	tests := []ir.Expression{}
	for _, clause := range statement.CaseBlock.AsCaseBlock().Clauses.Nodes {
		for _, inner := range clause.AsCaseOrDefaultClause().Statements.Nodes {
			if inner.Kind == ast.KindVariableStatement {
				// A declaration directly in a case is scoped to the whole switch in JavaScript, where
				// another case can see it (and hit its dead zone).
				return nil, l.notYet(inner, "a declaration directly in a case (wrap the case in a block)")
			}
		}
		isDefault := clause.Kind == ast.KindDefaultClause
		if !isDefault {
			test, err := l.expression(clause.AsCaseOrDefaultClause().Expression)
			if err != nil {
				return nil, err
			}
			switch test.(type) {
			case ir.NumberConstant, ir.StringConstant, ir.BooleanConstant:
			default:
				return nil, l.notYet(clause, "a case that isn't a constant")
			}
			if test.Type() != value.Type() {
				return nil, l.notYet(clause, "a case whose type differs from the switch's")
			}
			tests = append(tests, test)
		}
		body, err := l.statements(clause.AsCaseOrDefaultClause().Statements.Nodes)
		if err != nil {
			return nil, err
		}
		if len(body) == 0 && !isDefault {
			// case 'a': case 'b': share the next body.
			continue
		}
		if isDefault {
			// Cases grouped with default run its body, which is what not matching does anyway.
			lowered.Default = body
		} else {
			lowered.Cases = append(lowered.Cases, ir.Case{Tests: tests, Body: body})
		}
		tests = []ir.Expression{}
	}
	return []ir.Statement{lowered}, nil
}

// arrayMethod lowers array.push(value) and array.join(separator).
func (l *lowering) arrayMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	element, err := l.elementType(receiver)
	if err != nil {
		return nil, true, err
	}
	array, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments := []ir.Expression{}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, lowered)
	}
	if name == "push" {
		if len(arguments) != 1 {
			return nil, true, l.notYet(node, "push with other than one value")
		}
		return ir.ArrayPush{Array: array, Value: arguments[0], Element: element}, true, nil
	}
	if element == ir.Object {
		// JavaScript writes each object as "[object Object]"; 0.1 has no use for that.
		return nil, true, l.notYet(node, "join on an array of objects")
	}
	separator := ir.Expression(ir.StringConstant{Index: l.constant(",")})
	if len(arguments) == 1 {
		if arguments[0].Type() != ir.String {
			return nil, true, l.notYet(node, "join with a separator that isn't a string")
		}
		separator = arguments[0]
	} else if len(arguments) > 1 {
		return nil, true, l.notYet(node, "join with more than one argument")
	}
	return ir.ArrayJoin{Array: array, Separator: separator, Element: element}, true, nil
}

// mapTypes is a Map's key and value representations. 0.1's maps have string or number keys.
func (l *lowering) mapTypes(node *ast.Node) (ir.Type, ir.Type, error) {
	arguments := l.checker.GetTypeArguments(l.checker.GetTypeAtLocation(node))
	if len(arguments) != 2 {
		return 0, 0, l.notYet(node, "a Map whose key and value types aren't known")
	}
	key, keyKnown := l.representation(arguments[0])
	value, valueKnown := l.representation(arguments[1])
	if !keyKnown || (key != ir.String && key != ir.Number) {
		return 0, 0, l.notYet(node, "a Map whose keys aren't strings or numbers")
	}
	if !valueKnown || value == ir.MaybeNumber {
		return 0, 0, l.notYet(node, "a Map of "+l.checker.TypeToString(arguments[1]))
	}
	return key, value, nil
}

// newExpression lowers new Map(), and new Map([[key, value], ...]) with its pairs written out, which
// is what the array of pairs means.
func (l *lowering) newExpression(node *ast.Node) (ir.Expression, error) {
	created := node.AsNewExpression()
	if !l.isLibraryGlobal(created.Expression, "Map") {
		return nil, l.notYet(node, "new "+describe(created.Expression))
	}
	key, value, err := l.mapTypes(node)
	if err != nil {
		return nil, err
	}
	lowered := ir.MapNew{Key: key, Value: value}
	if created.Arguments == nil || len(created.Arguments.Nodes) == 0 {
		return lowered, nil
	}
	pairs := ast.SkipParentheses(created.Arguments.Nodes[0])
	if len(created.Arguments.Nodes) != 1 || pairs.Kind != ast.KindArrayLiteralExpression {
		return nil, l.notYet(node, "new Map from anything but pairs written out")
	}
	for _, pair := range pairs.AsArrayLiteralExpression().Elements.Nodes {
		pair = ast.SkipParentheses(pair)
		if pair.Kind != ast.KindArrayLiteralExpression || len(pair.AsArrayLiteralExpression().Elements.Nodes) != 2 {
			return nil, l.notYet(pair, "a Map entry that isn't [key, value] written out")
		}
		elements := pair.AsArrayLiteralExpression().Elements.Nodes
		entryKey, err := l.expression(elements[0])
		if err != nil {
			return nil, err
		}
		entryValue, err := l.expression(elements[1])
		if err != nil {
			return nil, err
		}
		lowered.Entries = append(lowered.Entries, [2]ir.Expression{entryKey, entryValue})
	}
	return lowered, nil
}

// mapMethod lowers map.get, set, has and delete.
func (l *lowering) mapMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	key, value, err := l.mapTypes(receiver)
	if err != nil {
		return nil, true, err
	}
	object, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments := []ir.Expression{}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, lowered)
	}
	want := 1
	if name == "set" {
		want = 2
	}
	if len(arguments) != want || arguments[0].Type() != key || (name == "set" && arguments[1].Type() != value) {
		return nil, true, l.notYet(node, "map."+name+" with arguments of other types")
	}
	switch name {
	case "get":
		return ir.MapGet{Map: object, Key: arguments[0], KeyType: key, ValueType: value}, true, nil
	case "set":
		return ir.MapSet{Map: object, Key: arguments[0], Value: arguments[1], KeyType: key, ValueType: value}, true, nil
	case "has":
		return ir.MapHas{Map: object, Key: arguments[0], KeyType: key}, true, nil
	}
	return ir.MapDelete{Map: object, Key: arguments[0], KeyType: key}, true, nil
}

// stringMethod lowers string.trim() and string.charCodeAt(index).
func (l *lowering) stringMethod(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if name == "trim" {
		if len(arguments) != 0 {
			return nil, true, l.notYet(node, "trim with arguments")
		}
		return ir.Trim{Value: value}, true, nil
	}
	index := ir.Expression(ir.NumberConstant{Value: 0})
	if len(arguments) > 1 {
		return nil, true, l.notYet(node, "charCodeAt with more than one argument")
	}
	if len(arguments) == 1 {
		if index, err = l.expression(arguments[0]); err != nil {
			return nil, true, err
		}
		if index.Type() != ir.Number {
			return nil, true, l.notYet(node, "charCodeAt with an index that isn't a number")
		}
	}
	return ir.CharCodeAt{Value: value, Index: index}, true, nil
}

// shorthand lowers the value of { value }. The name there is the field's, and asked for its symbol
// the checker gives the field; the variable it reads is a separate question.
func (l *lowering) shorthand(property *ast.Node) (ir.Expression, error) {
	symbol := l.checker.GetShorthandAssignmentValueSymbol(property)
	if symbol != nil && symbol.Flags&ast.SymbolFlagsAlias != 0 {
		symbol = l.checker.GetAliasedSymbol(symbol)
	}
	if symbol != nil {
		symbol = l.checker.GetExportSymbolOfSymbol(symbol)
	}
	local, isLocal := l.locals[symbol]
	if !isLocal {
		return nil, l.notYet(property, "reading "+property.Name().Text())
	}
	of := l.result.Locals[local].Type
	if of == ir.MaybeNumber {
		return nil, l.notYet(property, "a field from a number | undefined variable")
	}
	return ir.Read{Local: local, Of: of, Checked: l.checked(local)}, nil
}
