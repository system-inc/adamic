package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// keyable reports whether values of a type can be a Map's keys or a Set's elements: strings and
// numbers, compared as === compares them (SameValueZero), and objects, arrays, maps and functions,
// each its own key, compared by identity.
func keyable(valueType ir.Type) bool {
	switch valueType {
	case ir.String, ir.Number, ir.Object, ir.Array, ir.Map, ir.Closure:
		return true
	}
	return false
}

// iterated lowers something a for...of or a spread walks as the array of what it gives, with that
// array's element type: an array, a string's code points, a Set's elements, a Map's [key, value]
// entries, and map.keys(), map.values(), map.entries(), set.keys() and set.values().
func (l *lowering) iterated(node *ast.Node) (ir.Expression, ir.Type, error) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindCallExpression && len(node.AsCallExpression().Arguments.Nodes) == 0 {
		if callee := ast.SkipParentheses(node.AsCallExpression().Expression); callee.Kind == ast.KindPropertyAccessExpression {
			receiver := callee.AsPropertyAccessExpression().Expression
			part := callee.Name().Text()
			receiverType, _ := l.representation(l.checker.GetTypeAtLocation(receiver))
			if receiverType == ir.Map && (part == "keys" || part == "values" || part == "entries") {
				return l.collectionPart(node, receiver, part)
			}
		}
	}
	if l.checker.IsArrayType(l.checker.GetTypeAtLocation(node)) {
		element, err := l.elementType(node)
		if err != nil {
			return nil, 0, err
		}
		array, err := l.expression(node)
		return array, element, err
	}
	value, err := l.expression(node)
	if err != nil {
		return nil, 0, err
	}
	switch value.Type() {
	case ir.String:
		return ir.CodePoints{Value: value}, ir.String, nil
	case ir.Map:
		if l.isSet(node) {
			element, err := l.setElement(node)
			if err != nil {
				return nil, 0, err
			}
			return ir.SetValues{Set: value, Element: element}, element, nil
		}
		key, mapped, err := l.mapTypes(node)
		if err != nil {
			return nil, 0, err
		}
		return ir.MapEntries{Map: value, KeyType: key, ValueType: mapped}, ir.Object, nil
	}
	return nil, 0, l.notYet(node, "iterating a "+typeName(value.Type()))
}

// collectionPart lowers map.keys(), map.values() and map.entries(), and set.keys() and set.values(),
// where they're walked whole: each is a new array of what it gives.
func (l *lowering) collectionPart(node *ast.Node, receiver *ast.Node, part string) (ir.Expression, ir.Type, error) {
	if l.isSet(receiver) {
		if part == "entries" {
			return nil, 0, l.notYet(node, "a Set's entries ([element, element] pairs) outside a for...of")
		}
		element, err := l.setElement(receiver)
		if err != nil {
			return nil, 0, err
		}
		set, err := l.expression(receiver)
		if err != nil {
			return nil, 0, err
		}
		return ir.SetValues{Set: set, Element: element}, element, nil
	}
	key, value, err := l.mapTypes(receiver)
	if err != nil {
		return nil, 0, err
	}
	collection, err := l.expression(receiver)
	if err != nil {
		return nil, 0, err
	}
	switch part {
	case "keys":
		return ir.MapKeys{Map: collection, Key: key}, key, nil
	case "values":
		return ir.MapValues{Map: collection, Value: value}, value, nil
	}
	return ir.MapEntries{Map: collection, KeyType: key, ValueType: value}, ir.Object, nil
}

// newMapFrom lowers new Map(pairs) from anything that gives [key, value] pairs: an array of tuples,
// another Map, or map.entries(). Each pair's key and value must be held as the Map holds its own, or
// a number would land where number | undefined goes without becoming the pair, so they're checked.
func (l *lowering) newMapFrom(node *ast.Node, source *ast.Node, key ir.Type, value ir.Type) (ir.Expression, error) {
	pairKey, pairValue, isKnown := l.pairTypes(source)
	if !isKnown {
		return nil, l.notYet(source, "new Map from something that isn't [key, value] pairs")
	}
	if pairKey != key || pairValue != value {
		return nil, l.notYet(source, "new Map from pairs held otherwise than the Map's keys and values")
	}
	pairs, element, err := l.iterated(source)
	if err != nil {
		return nil, err
	}
	if element != ir.Object {
		return nil, l.notYet(source, "new Map from something that isn't [key, value] pairs")
	}
	return ir.MapNew{Key: key, Value: value, Pairs: pairs}, nil
}

// pairTypes is how a source of [key, value] pairs holds its keys and values: an array of tuples by its
// tuple's types, a Map (and its entries()) by its own.
func (l *lowering) pairTypes(source *ast.Node) (ir.Type, ir.Type, bool) {
	source = ast.SkipParentheses(source)
	sourceType := l.checker.GetTypeAtLocation(source)
	if l.checker.IsArrayType(sourceType) {
		element := l.checker.GetElementTypeOfArrayType(sourceType)
		if !checker.IsTupleType(element) {
			return 0, 0, false
		}
		arguments := l.checker.GetTypeArguments(element)
		if len(arguments) != 2 {
			return 0, 0, false
		}
		key, keyKnown := l.representation(arguments[0])
		value, valueKnown := l.representation(arguments[1])
		return key, value, keyKnown && valueKnown
	}
	if source.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(source.AsCallExpression().Expression)
		if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "entries" {
			return 0, 0, false
		}
		source = callee.AsPropertyAccessExpression().Expression
	}
	if representation, _ := l.representation(l.checker.GetTypeAtLocation(source)); representation != ir.Map || l.isSet(source) {
		return 0, 0, false
	}
	key, value, err := l.mapTypes(source)
	return key, value, err == nil
}

// clearOrVisit lowers clear() and forEach(callback) on a Map or a Set.
func (l *lowering) clearOrVisit(node *ast.Node, receiver *ast.Node, name string, set bool) (ir.Expression, bool, error) {
	var key, value ir.Type
	var err error
	if set {
		key, err = l.setElement(receiver)
		value = key
	} else {
		key, value, err = l.mapTypes(receiver)
	}
	if err != nil {
		return nil, true, err
	}
	collection, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if name == "clear" {
		if len(arguments) != 0 {
			return nil, true, l.notYet(node, "clear with arguments")
		}
		return ir.MapClear{Map: collection}, true, nil
	}
	if len(arguments) != 1 {
		return nil, true, l.notYet(node, "forEach with other than one callback")
	}
	callback, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	if callback.Type() != ir.Closure {
		return nil, true, l.notYet(arguments[0], "forEach with a callback that isn't a function")
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(arguments[0]), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, true, l.notYet(arguments[0], "forEach with an overloaded callback")
	}
	var returns ir.Type
	if result := l.checker.GetReturnTypeOfSignature(signatures[0]); result.Flags()&checker.TypeFlagsVoid == 0 {
		var isKnown bool
		if returns, isKnown = l.representation(result); !isKnown || slotless(returns) {
			return nil, true, l.notYet(arguments[0], "forEach with a callback returning "+l.checker.TypeToString(result))
		}
	}
	return ir.MapForEach{Map: collection, Callback: callback, Key: key, Value: value, Set: set, Returns: returns}, true, nil
}

// tupleLiteral lowers a tuple written out, [key, value], where the checker typed it a tuple: an
// object whose fields are "0", "1" and on, each held as the tuple's type says, as [...map]'s entries
// are made.
func (l *lowering) tupleLiteral(node *ast.Node) (ir.Expression, error) {
	elements := node.AsArrayLiteralExpression().Elements.Nodes
	// What it's written into decides how each element is held: [key, undefined] pushed onto an array
	// of [string, number | undefined] holds number | undefined, which its own type wouldn't say.
	tuple := l.checker.GetTypeAtLocation(node)
	if contextual := l.checker.GetContextualType(node, checker.ContextFlagsNone); contextual != nil && checker.IsTupleType(contextual) {
		tuple = contextual
	}
	types := l.checker.GetTypeArguments(tuple)
	if len(types) != len(elements) {
		return nil, l.notYet(node, "a tuple with optional or rest elements")
	}
	literal := ir.ObjectLiteral{Tuple: true}
	for index, element := range elements {
		if element.Kind == ast.KindSpreadElement || element.Kind == ast.KindOmittedExpression {
			return nil, l.notYet(element, describe(element)+" in a tuple")
		}
		held, isKnown := l.representation(types[index])
		if !isKnown || slotless(held) {
			return nil, l.notYet(element, "a tuple element of type "+l.checker.TypeToString(types[index]))
		}
		value, err := l.expression(element)
		if err != nil {
			return nil, err
		}
		// A number, or undefined, where number | undefined goes is made that pair.
		value = fit(value, held)
		// undefined where a reference that may be missing goes is a null one, as anywhere.
		_, isUndefined := value.(ir.Undefined)
		if value.Type() != held && !(isUndefined && held.IsReference()) {
			return nil, l.notYet(element, "a tuple element held otherwise than its type")
		}
		literal.Fields = append(literal.Fields, ir.Field{Name: strconv.Itoa(index), Value: value})
	}
	return literal, nil
}

// tupleLength is tuple.length for a tuple of fixed length: its type's count, which the checker gives
// as the type of the read. The tuple is a plain variable, so not reading it changes nothing, unless
// it's a global read from a function, which may throw for the temporal dead zone, as JavaScript does.
func (l *lowering) tupleLength(node *ast.Node, tuple *ast.Node) (ir.Expression, error) {
	tuple = ast.SkipParentheses(tuple)
	local, isLocal := l.local(tuple)
	if !ast.IsIdentifier(tuple) || !isLocal || l.checked(local) {
		return nil, l.notYet(node, "the length of a tuple that isn't a plain local")
	}
	if l.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsNumberLiteral == 0 {
		return nil, l.notYet(node, "the length of a tuple with optional or rest elements")
	}
	return l.numericLiteral(node)
}

// destructure lowers const [a, b] = tuple, and let: the tuple is held in a local of its own, made
// here, and each name is declared from its field, in order, as JavaScript reads them.
func (l *lowering) destructure(pattern *ast.Node, initializer *ast.Node) ([]ir.Statement, error) {
	if initializer == nil {
		return nil, l.notYet(pattern, "a destructuring declaration without a value")
	}
	tupleType := l.checker.GetTypeAtLocation(initializer)
	if !checker.IsTupleType(tupleType) {
		return nil, l.notYet(pattern, "destructuring anything but a tuple")
	}
	elementTypes := l.checker.GetTypeArguments(tupleType)
	value, err := l.expression(initializer)
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.Object {
		return nil, l.notYet(initializer, "destructuring a "+typeName(value.Type()))
	}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "tuple", Type: ir.Object, Function: l.functionIndex})
	statements := []ir.Statement{ir.Declare{Local: held, Value: value}}
	for index, binding := range pattern.AsBindingPattern().Elements.Nodes {
		// A hole, [, second], is a binding element with no name.
		if binding.Kind == ast.KindOmittedExpression || binding.Name() == nil {
			continue
		}
		declared := binding.AsBindingElement()
		if !ast.IsIdentifier(binding.Name()) || declared.Initializer != nil || declared.DotDotDotToken != nil {
			return nil, l.notYet(binding, "a destructured name that isn't plain")
		}
		if index >= len(elementTypes) {
			return nil, l.notYet(binding, "destructuring past a tuple's end")
		}
		local, err := l.declareLocal(binding.Name())
		if err != nil {
			return nil, err
		}
		of := l.result.Locals[local].Type
		if element, isKnown := l.representation(elementTypes[index]); !isKnown || element != of || slotless(of) {
			return nil, l.notYet(binding, "a destructured name held otherwise than its tuple element")
		}
		field := ir.Property{Object: ir.Read{Local: held, Of: ir.Object}, Name: strconv.Itoa(index), Of: of}
		statements = append(statements, ir.Declare{Local: local, Value: field})
	}
	return statements, nil
}
