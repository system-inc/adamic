package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// keyable reports whether values of a type can be a Map's keys or a Set's elements: strings, numbers
// and booleans, plus packed number | undefined, compared as === compares them (numbers by
// SameValueZero), and objects, arrays, maps
// and functions, each its own key, compared by identity.
func keyable(valueType ir.Type) bool {
	switch valueType {
	case ir.String, ir.Number, ir.Boolean, ir.MaybeNumber, ir.Object, ir.Array, ir.Map, ir.Closure:
		return true
	}
	return false
}

// iterated lowers something a for...of or a spread walks as the array of what it gives, with that
// array's element type: an array, a string's code points, a Set's elements, a Map's [key, value]
// entries, and map.keys(), map.values(), map.entries(), set.keys() and set.values().
func (l *lowering) iterated(node *ast.Node) (ir.Expression, ir.Type, error) {
	node = ast.SkipParentheses(node)
	if element, iterator := l.libraryIteratorElement(node); iterator {
		value, err := l.expression(node)
		if err != nil {
			return nil, 0, err
		}
		return l.libraryIteratorArray(node, value, element), element, nil
	}
	if plan, err := l.planIteration(node); err != nil {
		return nil, 0, err
	} else if plan != nil {
		value, err := l.collectIteration(node, node, nil, plan, plan.element)
		return value, plan.element, err
	}
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
	if l.viewArrayBase(l.phantomArrayView(l.checker.GetTypeAtLocation(node))) != nil {
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
			iterator, _, err := l.libraryCollectionIterator(node, receiver, part)
			if err != nil {
				return nil, 0, err
			}
			return l.libraryIteratorArray(node, iterator, ir.Object), ir.Object, nil
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
	if l.checker.IsArrayType(sourceType) || l.isLibraryType(sourceType, "MapIterator", "SetIterator", "Set", "ReadonlySet") {
		element := l.checker.GetElementTypeOfArrayType(sourceType)
		if element == nil {
			arguments := l.typeArguments(sourceType)
			if len(arguments) != 1 {
				return 0, 0, false
			}
			element = arguments[0]
		}
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
	if representation, _ := l.representation(l.checker.GetTypeAtLocation(source)); representation != ir.Map {
		return 0, 0, false
	}
	if l.isSet(source) {
		element, err := l.setElement(source)
		return element, element, err == nil
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
	if len(arguments) == 2 {
		return l.libraryForEachThisArg(node, receiver, set)
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

// destructure lowers const [a, b] = tuple and const { x, y } = object, and let: what's destructured
// is held in a local of its own, made here, and each name is declared from its field, in order, as
// JavaScript reads them.
func (l *lowering) destructure(pattern *ast.Node, initializer *ast.Node) ([]ir.Statement, error) {
	if initializer == nil {
		return nil, l.notYet(pattern, "a destructuring declaration without a value")
	}
	if pattern.Kind == ast.KindArrayBindingPattern {
		if plan, err := l.planIteration(initializer); err != nil {
			return nil, err
		} else if plan != nil {
			return l.destructureIterator(pattern, initializer, plan)
		}
	}
	value, err := l.expression(initializer)
	if err != nil {
		return nil, err
	}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "destructured", Type: ir.Object, Function: l.functionIndex})
	declared, err := l.destructureFrom(pattern, l.checker.GetTypeAtLocation(initializer), value.Type(), held)
	if err != nil {
		return nil, err
	}
	return append([]ir.Statement{ir.Declare{Local: held, Value: value}}, declared...), nil
}

// destructureFrom declares a pattern's names from the local held, which holds a value of type
// destructured (held as heldAs): a tuple's elements by position, an object's fields by name.
func (l *lowering) destructureFrom(pattern *ast.Node, destructured *checker.Type, heldAs ir.Type, held int) ([]ir.Statement, error) {
	tuple := pattern.Kind == ast.KindArrayBindingPattern
	if tuple && !checker.IsTupleType(destructured) {
		return nil, l.notYet(pattern, "destructuring anything but a tuple into [names]")
	}
	if heldAs != ir.Object {
		return nil, l.notYet(pattern, "destructuring a "+typeName(heldAs))
	}
	var elementTypes []*checker.Type
	if tuple {
		elementTypes = l.checker.GetTypeArguments(destructured)
	}
	statements := []ir.Statement{}
	for index, binding := range pattern.AsBindingPattern().Elements.Nodes {
		// A hole, [, second], is a binding element with no name.
		if binding.Kind == ast.KindOmittedExpression || binding.Name() == nil {
			continue
		}
		declared := binding.AsBindingElement()
		if !ast.IsIdentifier(binding.Name()) || declared.Initializer != nil || declared.DotDotDotToken != nil {
			return nil, l.notYet(binding, "a destructured name that isn't plain")
		}
		var field string
		var fieldType *checker.Type
		absent := false
		if tuple {
			if index >= len(elementTypes) {
				return nil, l.notYet(binding, "destructuring past a tuple's end")
			}
			field, fieldType = strconv.Itoa(index), elementTypes[index]
		} else {
			// { x } reads x, and { x: other } reads x into other.
			field = binding.Name().Text()
			if declared.PropertyName != nil {
				if !ast.IsIdentifier(declared.PropertyName) && declared.PropertyName.Kind != ast.KindStringLiteral {
					return nil, l.notYet(declared.PropertyName, "a computed field name")
				}
				field = declared.PropertyName.Text()
			}
			property := l.checker.GetPropertyOfType(destructured, field)
			if property == nil {
				return nil, l.notYet(binding, "destructuring a field the type doesn't name")
			}
			if l.inheritedLibrarySymbol(property) {
				return nil, l.prototypeRead(binding, field)
			}
			if property.Flags&ast.SymbolFlagsMethod != 0 {
				return nil, &Refused{Where: l.program.Where(binding), What: "a method in object destructuring", Fix: "call it on its receiver or wrap that call in an arrow; destructuring would lose this"}
			}
			for _, root := range l.checker.GetRootSymbols(property) {
				for _, declaration := range root.Declarations {
					if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
						return nil, l.notYet(binding, "an inherited library member in object destructuring, which is not an own field")
					}
				}
			}
			if err := l.erasedMethodField(binding, destructured, field); err != nil {
				return nil, err
			}
			fieldType = l.checker.GetTypeOfSymbol(property)
			absent = property.Flags&ast.SymbolFlagsOptional != 0
		}
		local, err := l.declareLocal(binding.Name())
		if err != nil {
			return nil, err
		}
		of := l.result.Locals[local].Type
		if element, isKnown := l.representation(fieldType); !isKnown || element != of || slotless(of) {
			return nil, l.notYet(binding, "a destructured name held otherwise than its field")
		}
		value := ir.Property{Object: ir.Read{Local: held, Of: ir.Object}, Name: field, Of: of, Absent: absent}
		value.View = sourceExpression(binding) + " (field " + field + ")"
		value.ViewType = l.checker.TypeToString(fieldType)
		value.ViewTypeID = int(fieldType.Id())
		value.ViewReceiverTypeID = int(destructured.Id())
		value.ViewWhere = l.program.Where(binding)
		l.prepareViewCallableProperty(binding, fieldType, &value)
		value.ViewAllowed = l.viewLiterals(fieldType)
		statements = append(statements, ir.Declare{Local: local, Value: value})
	}
	return statements, nil
}

// optionalTupleElement is tuple?.[index], a tuple that may be missing: its field when it's there, and
// undefined when it isn't, which a number field gives as number | undefined and a reference field as
// a null one. Fields held any other way stay not yet, as for ?. on a field (object.go's property).
func (l *lowering) optionalTupleElement(node *ast.Node, object ir.Expression, index *ast.Node) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	tuple := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))
	if index.Kind != ast.KindNumericLiteral || !checker.IsTupleType(tuple) {
		return nil, l.notYet(node, "?.[] on anything but a tuple, at a position written out")
	}
	elements := l.checker.GetTypeArguments(tuple)
	position, err := strconv.Atoi(index.Text())
	if err != nil || position >= len(elements) {
		return nil, l.notYet(node, "?.[] past a tuple's end")
	}
	held, isKnown := l.representation(elements[position])
	if !isKnown || (held != ir.Number && !held.IsReference()) || held == ir.Weak {
		return nil, l.notYet(node, "?.[] to a tuple element of type "+l.checker.TypeToString(elements[position]))
	}
	return ir.Property{Object: object, Name: strconv.Itoa(position), Of: held, Optional: true}, nil
}

// everyKnown reports whether every type has a representation.
func (l *lowering) everyKnown(types []*checker.Type) bool {
	for _, each := range types {
		if _, isKnown := l.representation(each); !isKnown {
			return false
		}
	}
	return true
}
