package lower

import (
	"math"
	"strconv"

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
	items := node.AsArrayLiteralExpression().Elements.Nodes
	if len(items) == 1 && items[0].Kind == ast.KindSpreadElement && !l.checker.IsArrayType(l.checker.GetTypeAtLocation(items[0].AsSpreadElement().Expression)) {
		// [...text] is the text's code points.
		spread, err := l.expression(items[0].AsSpreadElement().Expression)
		if err != nil {
			return nil, err
		}
		if spread.Type() == ir.String {
			return ir.CodePoints{Value: spread}, nil
		}
		if spread.Type() == ir.Map {
			key, value, err := l.mapTypes(items[0].AsSpreadElement().Expression)
			if err != nil {
				return nil, err
			}
			return ir.MapEntries{Map: spread, KeyType: key, ValueType: value}, nil
		}
		return nil, l.notYet(items[0], "spreading a "+typeName(spread.Type())+" into an array")
	}
	spreads := false
	for _, item := range items {
		if item.Kind == ast.KindOmittedExpression {
			return nil, l.notYet(item, describe(item)+" in an array literal")
		}
		spread := item.Kind == ast.KindSpreadElement
		if spread {
			item = item.AsSpreadElement().Expression
		}
		value, err := l.expression(item)
		if err != nil {
			return nil, err
		}
		if spread {
			if value.Type() != ir.Array {
				return nil, l.notYet(item, "spreading a "+typeName(value.Type())+" among other elements")
			}
			if other, err := l.elementType(item); err != nil || other != element {
				return nil, l.notYet(item, "spreading an array of other elements")
			}
			spreads = true
		}
		literal.Elements = append(literal.Elements, value)
		literal.Spread = append(literal.Spread, spread)
	}
	if !spreads {
		literal.Spread = nil
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
	if !isKnown || valueType == ir.MaybeNumber {
		// An element is one adamic_value, and number | undefined needs two words.
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
		case "random":
			return nil, refusedRandom(l, node)
		}
		return nil, l.notYet(node, "Math."+name)
	}
	if l.isLibraryGlobal(access.Expression, "Number") {
		if value, isConstant := numberConstants[name]; isConstant && name != "Infinity" {
			return ir.NumberConstant{Value: value}, nil
		}
		return nil, l.notYet(node, "Number."+name)
	}
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, err
	}
	if access.QuestionDotToken != nil && object.Type() != ir.Object {
		// words[0]?.length is number | undefined, which a length read as a number can't hold.
		return nil, l.notYet(node, "optional chaining on a "+typeName(object.Type()))
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
		if of == ir.MaybeNumber && optional {
			// box?.size is number | undefined because box may be; the field itself is what's stored.
			if field := l.checker.GetSymbolAtLocation(node.Name()); field != nil {
				if stored, isKnown := l.representation(l.checker.GetTypeOfSymbol(field)); isKnown && stored == ir.Number {
					return ir.Property{Object: object, Name: name, Of: ir.Number, Optional: true}, nil
				}
			}
		}
		if of == ir.MaybeNumber {
			return nil, l.notYet(node, "a field of type number | undefined")
		}
		if optional && !of.IsReference() {
			return nil, l.notYet(node, "?. to a "+typeName(of)+", which would be "+typeName(of)+" | undefined")
		}
		return ir.Property{Object: object, Name: name, Of: of, Optional: optional}, nil
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
	// V8's fdlibm, ported bit for bit (runtime/ieee754.c), and its hypot (runtime/hypot.c).
	"sin": 1, "cos": 1, "tan": 1, "asin": 1, "acos": 1, "atan": 1, "atan2": 2,
	"sinh": 1, "cosh": 1, "tanh": 1, "asinh": 1, "acosh": 1, "atanh": 1,
	"exp": 1, "expm1": 1, "log": 1, "log1p": 1, "log2": 1, "log10": 1, "cbrt": 1, "hypot": -1,
}

// refusedRandom refuses Math.random for good in 0.1: a program's output would no longer be a function
// of its source, and the oracle compares it with Node's byte for byte (docs/0.1.md).
func refusedRandom(l *lowering, node *ast.Node) error {
	return &Refused{Where: l.program.Where(node), What: "Math.random", Fix: "0.1 programs are deterministic, so the oracle can hold them to Node; compute the values you need, with a generator of your own seeded by a constant"}
}

// builtin lowers a call to Math or a number's toFixed. isBuiltin is false for any other call.
func (l *lowering) builtin(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	// The global parseInt and parseFloat are Number's, the same functions.
	if l.isLibraryGlobal(callee, "parseInt") || l.isLibraryGlobal(callee, "parseFloat") {
		return l.numberCall(node, callee.Text())
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	receiver, name := callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
	if l.isLibraryGlobal(receiver, "Number") {
		return l.numberCall(node, name)
	}
	if receiverType, _ := l.representation(l.checker.GetTypeAtLocation(receiver)); receiverType == ir.Number && name == "toString" && len(node.AsCallExpression().Arguments.Nodes) == 0 {
		value, err := l.expression(receiver)
		if err != nil {
			return nil, true, err
		}
		return ir.NumberToString{Value: value}, true, nil
	}
	if receiverType, _ := l.representation(l.checker.GetTypeAtLocation(receiver)); receiverType == ir.Number && (name == "toExponential" || name == "toPrecision") {
		return l.numberFormat(node, receiver, name)
	}
	isMath := l.isLibraryGlobal(receiver, "Math")
	receiverType, _ := l.representation(l.checker.GetTypeAtLocation(receiver))
	isToFixed := name == "toFixed" && receiverType == ir.Number
	if receiverType == ir.Array && name == "pop" && len(node.AsCallExpression().Arguments.Nodes) == 0 {
		element, err := l.elementType(receiver)
		if err != nil {
			return nil, true, err
		}
		array, err := l.expression(receiver)
		if err != nil {
			return nil, true, err
		}
		return ir.ArrayPop{Array: array, Element: element}, true, nil
	}
	if _, isVisit := visits[name]; receiverType == ir.Array && (isVisit || arrayMethods[name]) {
		return l.arrayMethod(node, receiver, name)
	}
	if receiverType == ir.Map && (name == "get" || name == "set" || name == "has" || name == "delete") {
		return l.mapMethod(node, receiver, name)
	}
	if receiverType == ir.String && (name == "trim" || name == "charCodeAt") {
		return l.stringMethod(node, receiver, name)
	}
	if _, isKnown := stringMethods[name]; isKnown && receiverType == ir.String {
		return l.stringCall(node, receiver, name)
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
		if name == "random" {
			return nil, true, refusedRandom(l, node)
		}
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

// numberFormat lowers value.toExponential(digits) and value.toPrecision(digits), the digits perhaps
// left out: the receiver first, then the argument, as JavaScript evaluates them.
func (l *lowering) numberFormat(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) > 1 {
		return nil, true, l.notYet(node, name+" with more than one argument")
	}
	format := ir.NumberFormat{Method: name, Value: value}
	if len(arguments) == 1 {
		argument, err := l.expression(arguments[0])
		if err != nil {
			return nil, true, err
		}
		if argument.Type() != ir.Number {
			return nil, true, l.notYet(arguments[0], "a "+typeName(argument.Type())+" argument to "+name)
		}
		format.Argument = argument
	}
	return format, true, nil
}

// numberFunctions are the functions of Number that 0.1 has (docs/0.1.md), with what each takes; the
// last of parseInt's, the radix, may be left out.
var numberFunctions = map[string][]ir.Type{
	"parseInt": {ir.String, ir.Number}, "parseFloat": {ir.String},
	"isNaN": {ir.Number}, "isFinite": {ir.Number}, "isInteger": {ir.Number}, "isSafeInteger": {ir.Number},
}

// numberCall lowers Number.parseInt, parseFloat, isNaN, isFinite, isInteger and isSafeInteger.
func (l *lowering) numberCall(node *ast.Node, name string) (ir.Expression, bool, error) {
	takes, isKnown := numberFunctions[name]
	if !isKnown {
		return nil, true, l.notYet(node, "Number."+name)
	}
	written := node.AsCallExpression().Arguments.Nodes
	optional := 0
	if name == "parseInt" {
		optional = 1
	}
	if len(written) > len(takes) || len(written) < len(takes)-optional {
		return nil, true, l.notYet(node, name+" with these arguments")
	}
	arguments := []ir.Expression{}
	for index, argument := range written {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		if lowered.Type() != takes[index] {
			// Number.isNaN('x') is false in JavaScript, and a string there is a mistake 0.1 won't guess at.
			return nil, true, l.notYet(argument, "a "+typeName(lowered.Type())+" argument to "+name)
		}
		arguments = append(arguments, lowered)
	}
	return ir.NumberCall{Function: name, Arguments: arguments}, true, nil
}

// numberConstants are Number's constants, and the globals NaN and Infinity.
var numberConstants = map[string]float64{
	"MAX_SAFE_INTEGER": 9007199254740991, "MIN_SAFE_INTEGER": -9007199254740991,
	"EPSILON": 2.220446049250313e-16, "MAX_VALUE": math.MaxFloat64, "MIN_VALUE": math.SmallestNonzeroFloat64,
	"POSITIVE_INFINITY": math.Inf(1), "NEGATIVE_INFINITY": math.Inf(-1), "NaN": math.NaN(), "Infinity": math.Inf(1),
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
	if len(declarations) != 1 {
		return nil, l.notYet(initializer, "a for...of declaring more than one variable")
	}
	name := declarations[0].Name()
	if !ast.IsIdentifier(name) && name.Kind != ast.KindArrayBindingPattern {
		return nil, l.notYet(initializer, "a for...of destructuring an object")
	}
	// map.entries(), map.keys() and map.values() are the map itself, iterated for that part.
	iterated, mapPart := statement.Expression, ""
	if call := ast.SkipParentheses(iterated); call.Kind == ast.KindCallExpression && len(call.AsCallExpression().Arguments.Nodes) == 0 {
		if callee := ast.SkipParentheses(call.AsCallExpression().Expression); callee.Kind == ast.KindPropertyAccessExpression {
			receiver := callee.AsPropertyAccessExpression().Expression
			if receiverType, _ := l.representation(l.checker.GetTypeAtLocation(receiver)); receiverType == ir.Map {
				switch part := callee.Name().Text(); part {
				case "entries", "keys", "values":
					iterated, mapPart = receiver, part
				}
			}
		}
	}
	iterable, err := l.expression(iterated)
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
	case ir.Map:
		return l.forOfMap(node, iterable, iterated, mapPart, name)
	default:
		return nil, l.notYet(statement.Expression, "for...of over a "+typeName(iterable.Type()))
	}
	lowered := ir.ForOf{Iterable: iterable, Element: element}
	if ast.IsIdentifier(name) {
		if lowered.Local, err = l.declareLocal(name); err != nil {
			return nil, err
		}
	} else {
		// for (const [a, b] of pairs): each name reads a field of the tuple, "0", "1", ...
		if element != ir.Object {
			return nil, l.notYet(name, "destructuring a "+typeName(element))
		}
		for index, binding := range name.AsBindingPattern().Elements.Nodes {
			if binding.Kind == ast.KindOmittedExpression {
				continue
			}
			bound := binding.AsBindingElement()
			if !ast.IsIdentifier(binding.Name()) || bound.Initializer != nil || bound.DotDotDotToken != nil {
				return nil, l.notYet(binding, "a destructured name that isn't plain")
			}
			local, err := l.declareLocal(binding.Name())
			if err != nil {
				return nil, err
			}
			if l.result.Locals[local].Type == ir.MaybeNumber {
				return nil, l.notYet(binding, "a tuple element of type number | undefined")
			}
			lowered.Pattern = append(lowered.Pattern, ir.Binding{Local: local, Field: strconv.Itoa(index)})
		}
	}
	if lowered.Body, err = l.statement(statement.Statement); err != nil {
		return nil, err
	}
	return []ir.Statement{lowered}, nil
}

// forOfMap lowers for...of over a map: for (const [key, value] of map), and over map.keys() or
// map.values() into one name. An entry as one name would be a tuple made each step, which stage 0
// doesn't do yet.
func (l *lowering) forOfMap(node *ast.Node, iterable ir.Expression, iterated *ast.Node, part string, name *ast.Node) ([]ir.Statement, error) {
	key, value, err := l.mapTypes(iterated)
	if err != nil {
		return nil, err
	}
	if part == "" {
		part = "entries"
	}
	lowered := ir.ForOf{Iterable: iterable, MapPart: part, Key: key, Value: value}
	switch {
	case part == "entries" && name.Kind == ast.KindArrayBindingPattern:
		elements := name.AsBindingPattern().Elements.Nodes
		if len(elements) > 2 {
			return nil, l.notYet(name, "destructuring more than a key and a value")
		}
		for index, binding := range elements {
			if binding.Kind == ast.KindOmittedExpression {
				continue
			}
			bound := binding.AsBindingElement()
			if !ast.IsIdentifier(binding.Name()) || bound.Initializer != nil || bound.DotDotDotToken != nil {
				return nil, l.notYet(binding, "a destructured name that isn't plain")
			}
			local, err := l.declareLocal(binding.Name())
			if err != nil {
				return nil, err
			}
			lowered.Pattern = append(lowered.Pattern, ir.Binding{Local: local, Field: strconv.Itoa(index)})
		}
	case part != "entries" && ast.IsIdentifier(name):
		if lowered.Local, err = l.declareLocal(name); err != nil {
			return nil, err
		}
	default:
		return nil, l.notYet(name, "for...of over a map's "+part+" into this name (write [key, value], or iterate keys() or values())")
	}
	if lowered.Body, err = l.statement(node.AsForInOrOfStatement().Statement); err != nil {
		return nil, err
	}
	return []ir.Statement{lowered}, nil
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
	if created := ast.SkipParentheses(receiver); name == "fill" && created.Kind == ast.KindNewExpression && l.isLibraryGlobal(created.AsNewExpression().Expression, "Array") {
		return l.newArrayFilled(node, created)
	}
	element, err := l.elementType(receiver)
	if err != nil {
		return nil, true, err
	}
	array, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	if name == "sort" {
		return l.arraySort(node, array, element)
	}
	if name == "map" {
		arguments := node.AsCallExpression().Arguments.Nodes
		if len(arguments) != 1 {
			return nil, true, l.notYet(node, "map with other than one callback")
		}
		callback, err := l.expression(arguments[0])
		if err != nil {
			return nil, true, err
		}
		if callback.Type() != ir.Closure {
			return nil, true, l.notYet(arguments[0], "map with a callback that isn't a function")
		}
		result, err := l.elementType(node)
		if err != nil {
			return nil, true, err
		}
		return ir.ArrayMap{Array: array, Callback: callback, Element: element, Result: result}, true, nil
	}
	if _, isVisit := visits[name]; isVisit {
		return l.arrayVisit(node, array, element, name)
	}
	if name == "reduce" {
		return l.arrayReduce(node, array, element)
	}
	if name == "concat" && element == ir.Array {
		// TypeScript says grid.concat(row) appends row as one element; JavaScript spreads any array it's
		// given, so the types and the program would disagree.
		return nil, true, l.notYet(node, "concat on an array of arrays")
	}
	arguments := []ir.Expression{}
	for _, argument := range node.AsCallExpression().Arguments.Nodes {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, lowered)
	}
	if name == "slice" {
		for _, argument := range arguments {
			if argument.Type() != ir.Number {
				return nil, true, l.notYet(node, "slice with an index that isn't a number")
			}
		}
		if len(arguments) > 2 {
			return nil, true, l.notYet(node, "slice with more than two arguments")
		}
		return ir.ArraySlice{Array: array, Arguments: arguments}, true, nil
	}
	if name == "push" {
		if len(arguments) != 1 {
			return nil, true, l.notYet(node, "push with other than one value")
		}
		return ir.ArrayPush{Array: array, Value: arguments[0], Element: element}, true, nil
	}
	switch name {
	case "includes", "indexOf":
		if len(arguments) != 1 {
			return nil, true, l.notYet(node, name+" with a starting index")
		}
		if arguments[0].Type() != element {
			return nil, true, l.notYet(node, name+" with a value of another type than the elements")
		}
		return ir.ArraySearch{Array: array, Value: arguments[0], Element: element, Includes: name == "includes"}, true, nil
	case "at":
		if len(arguments) != 1 || arguments[0].Type() != ir.Number {
			return nil, true, l.notYet(node, "at with other than one number")
		}
		if element == ir.Boolean {
			return nil, true, l.notYet(node, "at on an array of booleans (boolean | undefined)")
		}
		return ir.ArrayIndex{Array: array, Index: arguments[0], Element: element, Relative: true}, true, nil
	case "reverse":
		return ir.ArrayReverse{Array: array}, true, nil
	case "fill":
		if len(arguments) == 0 || len(arguments) > 3 || arguments[0].Type() != element {
			return nil, true, l.notYet(node, "fill with other than a value of the elements' type")
		}
		fill := ir.ArrayFill{Array: array, Value: arguments[0], Element: element}
		for index, bound := range arguments[1:] {
			if bound.Type() != ir.Number {
				return nil, true, l.notYet(node, "fill with a bound that isn't a number")
			}
			if index == 0 {
				fill.Start = bound
			} else {
				fill.End = bound
			}
		}
		return fill, true, nil
	case "splice":
		if len(arguments) == 0 || arguments[0].Type() != ir.Number || (len(arguments) > 1 && arguments[1].Type() != ir.Number) {
			return nil, true, l.notYet(node, "splice without a start and a count that are numbers")
		}
		splice := ir.ArraySplice{Array: array, Start: arguments[0], Element: element}
		if len(arguments) > 1 {
			splice.Count = arguments[1]
			for _, item := range arguments[2:] {
				if item.Type() != element {
					return nil, true, l.notYet(node, "splice inserting a value of another type than the elements")
				}
				splice.Items = append(splice.Items, item)
			}
		}
		return splice, true, nil
	case "concat":
		for index, argument := range arguments {
			if argument.Type() != ir.Array {
				return nil, true, l.notYet(node, "concat with a value that isn't an array (JavaScript appends it)")
			}
			if other, err := l.elementType(node.AsCallExpression().Arguments.Nodes[index]); err != nil || other != element {
				return nil, true, l.notYet(node, "concat of arrays of different elements")
			}
		}
		return ir.ArrayConcat{Array: array, Others: arguments}, true, nil
	}
	if element != ir.Number && element != ir.Boolean && element != ir.String {
		// JavaScript writes an object as "[object Object]", a function as its source, and an array as
		// its own join, flattened; 0.1 has no use for any of that.
		return nil, true, l.notYet(node, "join on an array of objects, arrays, maps or functions")
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

// arrayMethods are the array methods arrayMethod lowers, beside the visits.
var arrayMethods = map[string]bool{
	"push": true, "join": true, "slice": true, "sort": true, "map": true, "reduce": true,
	"includes": true, "indexOf": true, "at": true, "reverse": true, "concat": true, "splice": true,
	"fill": true,
}

// newArrayFilled lowers new Array(length).fill(value). new Array(length) alone is an array of holes,
// which 0.1 can't hold, so it lowers only filled, and filled whole.
func (l *lowering) newArrayFilled(node *ast.Node, created *ast.Node) (ir.Expression, bool, error) {
	element, err := l.elementType(node)
	if err != nil {
		return nil, true, err
	}
	constructed := created.AsNewExpression().Arguments
	filled := node.AsCallExpression().Arguments.Nodes
	if constructed == nil || len(constructed.Nodes) != 1 || len(filled) != 1 {
		return nil, true, l.notYet(node, "new Array filled with other than one length and one value (a part left unfilled is a hole)")
	}
	length, err := l.expression(constructed.Nodes[0])
	if err != nil {
		return nil, true, err
	}
	value, err := l.expression(filled[0])
	if err != nil {
		return nil, true, err
	}
	if length.Type() != ir.Number || value.Type() != element {
		return nil, true, l.notYet(node, "new Array(length).fill(value) with a length that isn't a number or a value of another type")
	}
	return ir.ArrayFill{Length: length, Value: value, Element: element}, true, nil
}

// visits are the array methods that call a function per element and look at what it returns.
var visits = map[string]struct{}{"forEach": {}, "filter": {}, "some": {}, "every": {}, "find": {}, "findIndex": {}}

// arrayVisit lowers forEach, filter, some, every, find and findIndex. All but forEach decide by what
// the callback returns, which 0.1 requires to be a boolean: JavaScript would take any value's
// truthiness there, and 0.1 has none.
func (l *lowering) arrayVisit(node *ast.Node, array ir.Expression, element ir.Type, name string) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 {
		return nil, true, l.notYet(node, name+" with other than one callback")
	}
	callback, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	if callback.Type() != ir.Closure {
		return nil, true, l.notYet(arguments[0], name+" with a callback that isn't a function")
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(arguments[0]), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, true, l.notYet(arguments[0], name+" with an overloaded callback")
	}
	var returns ir.Type
	if result := l.checker.GetReturnTypeOfSignature(signatures[0]); result.Flags()&checker.TypeFlagsVoid == 0 {
		var isKnown bool
		if returns, isKnown = l.representation(result); !isKnown {
			return nil, true, l.notYet(arguments[0], name+" with a callback returning "+l.checker.TypeToString(result))
		}
	}
	if name != "forEach" && returns != ir.Boolean {
		return nil, true, &Refused{Where: l.program.Where(arguments[0]), What: "a " + name + " callback that doesn't return a boolean", Fix: "return a comparison, like word.length > 0: 0.1 has no truthiness"}
	}
	if name == "find" && element == ir.Boolean {
		return nil, true, l.notYet(node, "find in an array of booleans (boolean | undefined)")
	}
	return ir.ArrayVisit{Method: name, Array: array, Callback: callback, Element: element, Returns: returns}, true, nil
}

// arrayReduce lowers array.reduce(callback, initial). 0.1 requires the initial value (docs/0.1.md):
// without one, an empty array throws and a one-element array returns it without a call.
func (l *lowering) arrayReduce(node *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) == 1 {
		return nil, true, &Refused{Where: l.program.Where(node), What: "reduce without an initial value", Fix: "pass one, like reduce((sum, value) => sum + value, 0): without it, an empty array throws"}
	}
	if len(arguments) != 2 {
		return nil, true, l.notYet(node, "reduce with other than a callback and an initial value")
	}
	callback, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	if callback.Type() != ir.Closure {
		return nil, true, l.notYet(arguments[0], "reduce with a callback that isn't a function")
	}
	initial, err := l.expression(arguments[1])
	if err != nil {
		return nil, true, err
	}
	result, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	if result != initial.Type() || result == ir.MaybeNumber {
		return nil, true, l.notYet(node, "reduce to a "+l.checker.TypeToString(l.checker.GetTypeAtLocation(node)))
	}
	return ir.ArrayReduce{Array: array, Callback: callback, Initial: initial, Element: element, Result: result}, true, nil
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
	if declaration, isClass := l.classes[l.symbol(ast.SkipParentheses(created.Expression))]; isClass {
		return l.construct(node, declaration)
	}
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
	l.touch(local)
	of := l.result.Locals[local].Type
	if of == ir.MaybeNumber {
		return nil, l.notYet(property, "a field from a number | undefined variable")
	}
	return ir.Read{Local: local, Of: of, Checked: l.checked(local)}, nil
}

// stringMethods are the string methods stringCall lowers: the types of their arguments, and how many
// may be left out, each filled in with JavaScript's default.
var stringMethods = map[string]struct {
	arguments []ir.Type
	optional  int
}{
	"slice":       {[]ir.Type{ir.Number, ir.Number}, 2},
	"codePointAt": {[]ir.Type{ir.Number}, 1},
	"padStart":    {[]ir.Type{ir.Number, ir.String}, 1},
	"padEnd":      {[]ir.Type{ir.Number, ir.String}, 1},
	"repeat":      {[]ir.Type{ir.Number}, 0},
	"indexOf":     {[]ir.Type{ir.String}, 0},
	"includes":    {[]ir.Type{ir.String}, 0},
	"startsWith":  {[]ir.Type{ir.String}, 0},
	"endsWith":    {[]ir.Type{ir.String}, 0},
	"split":       {[]ir.Type{ir.String}, 0},
	"lastIndexOf": {[]ir.Type{ir.String}, 0},
	"trimStart":   {nil, 0},
	"trimEnd":     {nil, 0},
	"toUpperCase": {nil, 0},
	"toLowerCase": {nil, 0},
	"at":          {[]ir.Type{ir.Number}, 0},
	// A pattern that's a string, and a replacement that's a string: a regular expression or a
	// function there isn't a string, and stays not yet.
	"replace":    {[]ir.Type{ir.String, ir.String}, 0},
	"replaceAll": {[]ir.Type{ir.String, ir.String}, 0},
}

func (l *lowering) stringCall(node *ast.Node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	shape := stringMethods[name]
	value, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) > len(shape.arguments) || len(written) < len(shape.arguments)-shape.optional {
		return nil, true, l.notYet(node, name+" with these arguments")
	}
	arguments := []ir.Expression{}
	for index, argument := range written {
		lowered, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		if lowered.Type() != shape.arguments[index] {
			return nil, true, l.notYet(argument, "a "+typeName(lowered.Type())+" argument to "+name)
		}
		arguments = append(arguments, lowered)
	}
	// JavaScript's defaults: codePointAt() is position 0, and padStart's fill is a space. slice's end
	// stays missing, which is different from any number, so the emitter is told how many were given.
	switch {
	case name == "codePointAt" && len(arguments) == 0:
		arguments = append(arguments, ir.NumberConstant{Value: 0})
	case (name == "padStart" || name == "padEnd") && len(arguments) == 1:
		arguments = append(arguments, ir.StringConstant{Index: l.constant(" ")})
	}
	return ir.StringCall{Method: name, Value: value, Arguments: arguments}, true, nil
}

// arraySort lowers array.sort(comparator). 0.1 requires the comparator (the default sorts numbers as
// strings), and stage 0 takes one of the module's functions by name.
func (l *lowering) arraySort(node *ast.Node, array ir.Expression, element ir.Type) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) == 0 {
		return nil, true, &Refused{Where: l.program.Where(node), What: "sort without a comparator", Fix: "pass one: the default compares numbers as strings, so [10, 9, 1].sort() is [1, 10, 9]"}
	}
	comparator := ast.SkipParentheses(arguments[0])
	function, isFunction := l.functions[l.symbol(comparator)]
	if !ast.IsIdentifier(comparator) || !isFunction {
		// A function value: an arrow, or a variable holding one.
		callback, err := l.expression(comparator)
		if err != nil {
			return nil, true, err
		}
		signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(comparator), checker.SignatureKindCall)
		if callback.Type() != ir.Closure || len(signatures) != 1 {
			return nil, true, l.notYet(comparator, "a comparator that isn't a function")
		}
		if returns, _ := l.representation(l.checker.GetReturnTypeOfSignature(signatures[0])); returns != ir.Number {
			return nil, true, l.notYet(comparator, "a comparator that doesn't return a number")
		}
		return ir.ArraySort{Array: array, Callback: callback, Element: element}, true, nil
	}
	declared := l.result.Functions[function]
	if declared.Returns != ir.Number || len(declared.Parameters) != 2 || l.result.Locals[declared.Parameters[0]].Type != element || l.result.Locals[declared.Parameters[1]].Type != element {
		return nil, true, l.notYet(comparator, "a comparator that doesn't take two elements and return a number")
	}
	return ir.ArraySort{Array: array, Comparator: function, Element: element}, true, nil
}

// elementAccess lowers array[index], string[index], and tuple[index] with a constant index: the
// tuple's field of that name.
func (l *lowering) elementAccess(node *ast.Node) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	index := ast.SkipParentheses(access.ArgumentExpression)
	object, err := l.expression(access.Expression)
	if err != nil {
		return nil, err
	}
	if object.Type() == ir.String {
		position, err := l.expression(access.ArgumentExpression)
		if err != nil {
			return nil, err
		}
		if position.Type() != ir.Number {
			return nil, l.notYet(node, "a string index that isn't a number")
		}
		return ir.StringIndex{Value: object, Index: position}, nil
	}
	if object.Type() == ir.Array {
		element, err := l.elementType(access.Expression)
		if err != nil {
			return nil, err
		}
		if element == ir.Boolean {
			return nil, l.notYet(node, "an index into an array of booleans (boolean | undefined)")
		}
		position, err := l.expression(access.ArgumentExpression)
		if err != nil {
			return nil, err
		}
		if position.Type() != ir.Number {
			return nil, l.notYet(node, "an array index that isn't a number")
		}
		return ir.ArrayIndex{Array: object, Index: position, Element: element}, nil
	}
	if object.Type() != ir.Object || index.Kind != ast.KindNumericLiteral || !checker.IsTupleType(l.checker.GetTypeAtLocation(access.Expression)) {
		return nil, l.notYet(node, describe(node))
	}
	of, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	if of == ir.MaybeNumber {
		return nil, l.notYet(node, "a tuple element of type number | undefined")
	}
	return ir.Property{Object: object, Name: index.Text(), Of: of}, nil
}

// setIndex lowers array[index] = value, as a statement.
func (l *lowering) setIndex(target *ast.Node, valueNode *ast.Node) ([]ir.Statement, error) {
	access := target.AsElementAccessExpression()
	array, err := l.expression(access.Expression)
	if err != nil {
		return nil, err
	}
	if array.Type() != ir.Array {
		return nil, l.notYet(target, "assigning an element of a "+typeName(array.Type()))
	}
	element, err := l.elementType(access.Expression)
	if err != nil {
		return nil, err
	}
	index, err := l.expression(access.ArgumentExpression)
	if err != nil {
		return nil, err
	}
	if index.Type() != ir.Number {
		return nil, l.notYet(target, "an array index that isn't a number")
	}
	value, err := l.expression(valueNode)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.SetIndex{Array: array, Index: index, Value: value, Element: element}}, nil
}
