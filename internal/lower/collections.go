package lower

import (
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
