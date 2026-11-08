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
func (lowering *lowering) iterated(node *ast.Node) (ir.Expression, ir.Type, error) {
	node = ast.SkipParentheses(node)
	// A call to a program's generator is collected as the generator it is, before any reading of its
	// declared iterator type (#qk8rztp, Compiler's risk 2). Collecting steps it to its end, which is
	// right only for a consumer that reads every value: an array literal's spread (object.go) is one,
	// and so are new Map (newMapFromGenerator) and new Set, which take every value a generator gives.
	if function, isGenerator, generatorCalleeError := lowering.generatorCallee(node); generatorCalleeError != nil {
		return nil, 0, generatorCalleeError
	} else if isGenerator {
		return lowering.collectGenerator(node, function)
	}
	if element, iterator := lowering.libraryIteratorElement(node); iterator {
		value, expressionError := lowering.expression(node)
		if expressionError != nil {
			return nil, 0, expressionError
		}
		return lowering.libraryIteratorArray(node, value, element), element, nil
	}
	if plan, planIterationError := lowering.planIteration(node); planIterationError != nil {
		return nil, 0, planIterationError
	} else if plan != nil {
		value, collectIterationError := lowering.collectIteration(node, node, nil, plan, plan.element)
		return value, plan.element, collectIterationError
	}
	if node.Kind == ast.KindCallExpression && len(node.AsCallExpression().Arguments.Nodes) == 0 {
		if callee := ast.SkipParentheses(
			node.AsCallExpression().Expression,
		); callee.Kind == ast.KindPropertyAccessExpression {
			receiver := callee.AsPropertyAccessExpression().Expression
			part := callee.Name().Text()
			receiverType, _ := lowering.representation(lowering.checker.GetTypeAtLocation(receiver))
			if receiverType == ir.Map && (part == "keys" || part == "values" || part == "entries") {
				return lowering.collectionPart(node, receiver, part)
			}
		}
	}
	if lowering.isArrayType(lowering.checker.GetTypeAtLocation(node)) {
		element, elementTypeError := lowering.elementType(node)
		if elementTypeError != nil {
			return nil, 0, elementTypeError
		}
		array, expressionError := lowering.expression(node)
		return array, element, expressionError
	}
	value, expressionError := lowering.expression(node)
	if expressionError != nil {
		return nil, 0, expressionError
	}
	switch value.Type() {
	case ir.String:
		return ir.CodePoints{Value: value}, ir.String, nil
	case ir.Map:
		if lowering.isSet(node) {
			element, setElementError := lowering.setElement(node)
			if setElementError != nil {
				return nil, 0, setElementError
			}
			return ir.SetValues{Set: value, Element: element}, element, nil
		}
		key, mapped, mapTypesError := lowering.mapTypes(node)
		if mapTypesError != nil {
			return nil, 0, mapTypesError
		}
		return ir.MapEntries{Map: value, KeyType: key, ValueType: mapped}, ir.Object, nil
	}
	return nil, 0, lowering.notYet(node, "iterating a "+typeName(value.Type()))
}

// collectionPart lowers map.keys(), map.values() and map.entries(), and set.keys() and set.values(),
// where they're walked whole: each is a new array of what it gives.
func (lowering *lowering) collectionPart(
	node *ast.Node,
	receiver *ast.Node,
	part string,
) (ir.Expression, ir.Type, error) {
	if lowering.isSet(receiver) {
		if part == "entries" {
			iterator, _, libraryCollectionIteratorError := lowering.libraryCollectionIterator(node, receiver, part)
			if libraryCollectionIteratorError != nil {
				return nil, 0, libraryCollectionIteratorError
			}
			return lowering.libraryIteratorArray(node, iterator, ir.Object), ir.Object, nil
		}
		element, setElementError := lowering.setElement(receiver)
		if setElementError != nil {
			return nil, 0, setElementError
		}
		set, expressionError := lowering.expression(receiver)
		if expressionError != nil {
			return nil, 0, expressionError
		}
		return ir.SetValues{Set: set, Element: element}, element, nil
	}
	key, value, mapTypesError := lowering.mapTypes(receiver)
	if mapTypesError != nil {
		return nil, 0, mapTypesError
	}
	collection, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, 0, expressionError
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
func (lowering *lowering) newMapFrom(
	node *ast.Node,
	source *ast.Node,
	key ir.Type,
	value ir.Type,
) (ir.Expression, error) {
	// A call to a program's generator is read by its origin, before its declared iterator type can
	// route it as a collection's (#qk8rztp).
	if function, isGenerator, generatorCalleeError := lowering.generatorCallee(source); generatorCalleeError != nil {
		return nil, generatorCalleeError
	} else if isGenerator {
		return lowering.newMapFromGenerator(source, function, key, value)
	}
	if sourceType := lowering.checker.GetTypeAtLocation(
		ast.SkipParentheses(source),
	); sourceType.Flags()&checker.TypeFlagsUnion != 0 &&
		lowering.checker.GetNonNullableType(sourceType) != sourceType &&
		lowering.arrayOrUndefined(sourceType) {
		return lowering.newMapFromOptional(source, sourceType, key, value)
	}
	pairKey, pairValue, isKnown := lowering.pairTypes(source)
	if !isKnown {
		// A Generator that isn't a direct call to a program's generator, one stored or passed, waits on
		// generators as values (#qk8rztp S2), not on this constructor.
		if lowering.isLibraryType(lowering.checker.GetTypeAtLocation(ast.SkipParentheses(source)), "Generator") {
			return nil, lowering.notYet(source, "new Map from a generator's pairs")
		}
		return nil, lowering.notYet(source, "new Map from something that isn't [key, value] pairs")
	}
	if pairKey != key || pairValue != value {
		return nil, lowering.notYet(source, "new Map from pairs held otherwise than the Map's keys and values")
	}
	pairs, element, iteratedError := lowering.iterated(source)
	if iteratedError != nil {
		return nil, iteratedError
	}
	if element != ir.Object {
		return nil, lowering.notYet(source, "new Map from something that isn't [key, value] pairs")
	}
	return ir.MapNew{Key: key, Value: value, Pairs: pairs}, nil
}

// newMapFromOptional lowers new Map(pairs) where pairs is an array of [key, value] tuples or undefined
// as new Map(pairs ?? []): Node makes an empty Map from undefined, as it does with no argument.
func (lowering *lowering) newMapFromOptional(
	source *ast.Node,
	sourceType *checker.Type,
	key ir.Type,
	value ir.Type,
) (ir.Expression, error) {
	pairKey, pairValue, isKnown := lowering.pairArrayTypes(lowering.checker.GetNonNullableType(sourceType))
	if !isKnown {
		return nil, lowering.notYet(source, "new Map from something that isn't [key, value] pairs")
	}
	if pairKey != key || pairValue != value {
		return nil, lowering.notYet(source, "new Map from pairs held otherwise than the Map's keys and values")
	}
	pairs, expressionError := lowering.expression(source)
	if expressionError != nil {
		return nil, expressionError
	}
	if pairs.Type() != ir.Array {
		return nil, lowering.notYet(source, "new Map from something that isn't [key, value] pairs")
	}
	return ir.MapNew{
		Key:   key,
		Value: value,
		Pairs: ir.Coalesce{Value: pairs, Fallback: ir.ArrayLiteral{Element: ir.Object}, Of: ir.Array},
	}, nil
}

// pairTypes is how a source of [key, value] pairs holds its keys and values: an array of tuples by its
// tuple's types, a Map (and its entries()) by its own.
func (lowering *lowering) pairTypes(source *ast.Node) (ir.Type, ir.Type, bool) {
	source = ast.SkipParentheses(source)
	sourceType := lowering.checker.GetTypeAtLocation(source)
	if lowering.isArrayType(sourceType) ||
		lowering.isLibraryType(sourceType, "MapIterator", "SetIterator", "Set", "ReadonlySet") {
		return lowering.pairArrayTypes(sourceType)
	}
	if source.Kind == ast.KindCallExpression {
		callee := ast.SkipParentheses(source.AsCallExpression().Expression)
		if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "entries" {
			return 0, 0, false
		}
		source = callee.AsPropertyAccessExpression().Expression
	}
	if representation, _ := lowering.representation(
		lowering.checker.GetTypeAtLocation(source),
	); representation != ir.Map {
		return 0, 0, false
	}
	if lowering.isSet(source) {
		element, setElementError := lowering.setElement(source)
		return element, element, setElementError == nil
	}
	key, value, mapTypesError := lowering.mapTypes(source)
	return key, value, mapTypesError == nil
}

// pairArrayTypes is how an array (or a library iterator or Set) of [key, value] tuples holds its keys
// and values, by its tuple's types.
func (lowering *lowering) pairArrayTypes(sourceType *checker.Type) (ir.Type, ir.Type, bool) {
	element := lowering.elementTypeOfArray(sourceType)
	if element == nil {
		arguments := lowering.typeArguments(sourceType)
		if len(arguments) != 1 {
			return 0, 0, false
		}
		element = arguments[0]
	}
	if !checker.IsTupleType(element) {
		return 0, 0, false
	}
	arguments := lowering.checker.GetTypeArguments(element)
	if len(arguments) != 2 {
		return 0, 0, false
	}
	key, keyKnown := lowering.representation(arguments[0])
	value, valueKnown := lowering.representation(arguments[1])
	return key, value, keyKnown && valueKnown
}

// clearOrVisit lowers clear() and forEach(callback) on a Map or a Set.
func (lowering *lowering) clearOrVisit(
	node *ast.Node,
	receiver *ast.Node,
	name string,
	set bool,
) (ir.Expression, bool, error) {
	var key, value ir.Type
	var err error
	if set {
		key, err = lowering.setElement(receiver)
		value = key
	} else {
		key, value, err = lowering.mapTypes(receiver)
	}
	if err != nil {
		return nil, true, err
	}
	collection, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if name == "clear" {
		if len(arguments) != 0 {
			return nil, true, lowering.notYet(node, "clear with arguments")
		}
		return ir.MapClear{Map: collection}, true, nil
	}
	if len(arguments) == 2 {
		return lowering.libraryForEachThisArg(node, receiver, set)
	}
	if len(arguments) != 1 {
		return nil, true, lowering.notYet(node, "forEach with other than one callback")
	}
	callback, expressionError := lowering.expression(arguments[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	if callback.Type() != ir.Closure {
		return nil, true, lowering.notYet(arguments[0], "forEach with a callback that isn't a function")
	}
	signatures := lowering.checker.GetSignaturesOfType(
		lowering.checker.GetTypeAtLocation(arguments[0]),
		checker.SignatureKindCall,
	)
	if len(signatures) != 1 {
		return nil, true, lowering.notYet(arguments[0], "forEach with an overloaded callback")
	}
	var returns ir.Type
	if result := lowering.checker.GetReturnTypeOfSignature(signatures[0]); result.Flags()&checker.TypeFlagsVoid == 0 {
		var isKnown bool
		if returns, isKnown = lowering.representation(result); !isKnown || slotless(returns) {
			return nil, true, lowering.notYet(
				arguments[0],
				"forEach with a callback returning "+lowering.checker.TypeToString(result),
			)
		}
	}
	return ir.MapForEach{
		Map:          collection,
		Callback:     callback,
		Key:          key,
		Value:        value,
		Set:          set,
		Returns:      returns,
		CallbackType: int(lowering.concrete(lowering.checker.GetTypeAtLocation(arguments[0])).Id()),
	}, true, nil
}

// tupleLength is tuple.length for a tuple of fixed length: its type's count, which the checker gives
// as the type of the read. The tuple is a plain variable, so not reading it changes nothing, unless
// it's a global read from a function, which may throw for the temporal dead zone, as JavaScript does.
func (lowering *lowering) tupleLength(node *ast.Node, tuple *ast.Node) (ir.Expression, error) {
	tuple = ast.SkipParentheses(tuple)
	local, isLocal := lowering.local(tuple)
	if !ast.IsIdentifier(tuple) || !isLocal || lowering.checked(local) {
		return nil, lowering.notYet(node, "the length of a tuple that isn't a plain local")
	}
	if lowering.checker.GetTypeAtLocation(node).Flags()&checker.TypeFlagsNumberLiteral == 0 {
		return nil, lowering.notYet(node, "the length of a tuple with optional or rest elements")
	}
	return lowering.numericLiteral(node)
}

// destructure lowers const [a, b] = tuple and const { x, y } = object, and let: what's destructured
// is held in a local of its own, made here, and each name is declared from its field, in order, as
// JavaScript reads them.
func (lowering *lowering) destructure(pattern *ast.Node, initializer *ast.Node) ([]ir.Statement, error) {
	if initializer == nil {
		return nil, lowering.notYet(pattern, "a destructuring declaration without a value")
	}
	if pattern.Kind == ast.KindArrayBindingPattern {
		if plan, planIterationError := lowering.planIteration(initializer); planIterationError != nil {
			return nil, planIterationError
		} else if plan != nil {
			return lowering.destructureIterator(pattern, initializer, plan)
		}
	}
	value, expressionError := lowering.expression(initializer)
	if expressionError != nil {
		return nil, expressionError
	}
	if refuseIncompatibleDestructuredViewError := lowering.refuseIncompatibleDestructuredView(
		pattern,
		initializer,
	); refuseIncompatibleDestructuredViewError != nil {
		return nil, refuseIncompatibleDestructuredViewError
	}
	held := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "destructured", Type: value.Type(), Function: lowering.functionIndex},
	)
	declared, destructureFromError := lowering.destructureFrom(
		pattern,
		lowering.checker.GetTypeAtLocation(initializer),
		value.Type(),
		held,
	)
	if destructureFromError != nil {
		return nil, destructureFromError
	}
	return append([]ir.Statement{ir.Declare{Local: held, Value: value}}, declared...), nil
}

// destructureFrom declares a pattern's names from the local held, which holds a value of type
// destructured (held as heldAs): a tuple's elements by position, an object's fields by name.
func (lowering *lowering) destructureFrom(
	pattern *ast.Node,
	destructured *checker.Type,
	heldAs ir.Type,
	held int,
) ([]ir.Statement, error) {
	tuple := pattern.Kind == ast.KindArrayBindingPattern
	if tuple && heldAs == ir.Array && lowering.matchArray(destructured) {
		return lowering.destructureMatch(pattern, held)
	}
	if tuple && !checker.IsTupleType(destructured) {
		return nil, lowering.notYet(pattern, "destructuring anything but a tuple into [names]")
	}
	if heldAs == ir.String && !tuple {
		return lowering.destructureString(pattern, held)
	}
	if heldAs != ir.Object {
		return nil, lowering.notYet(pattern, "destructuring a "+typeName(heldAs))
	}
	var elementTypes []*checker.Type
	if tuple {
		elementTypes = lowering.checker.GetTypeArguments(destructured)
	}
	statements := []ir.Statement{}
	for index, binding := range pattern.AsBindingPattern().Elements.Nodes {
		// A hole, [, second], is a binding element with no name.
		if binding.Kind == ast.KindOmittedExpression || binding.Name() == nil {
			continue
		}
		declared := binding.AsBindingElement()
		if !ast.IsIdentifier(binding.Name()) || declared.Initializer != nil || declared.DotDotDotToken != nil {
			return nil, lowering.notYet(binding, "a destructured name that isn't plain")
		}
		var field string
		var fieldType *checker.Type
		absent := false
		if tuple {
			if index >= len(elementTypes) {
				return nil, lowering.notYet(binding, "destructuring past a tuple's end")
			}
			field, fieldType = strconv.Itoa(index), elementTypes[index]
		} else {
			// { x } reads x, and { x: other } reads x into other.
			field = binding.Name().Text()
			if declared.PropertyName != nil {
				if !ast.IsIdentifier(declared.PropertyName) && declared.PropertyName.Kind != ast.KindStringLiteral {
					return nil, lowering.notYet(declared.PropertyName, "a computed field name")
				}
				field = declared.PropertyName.Text()
			}
			property, destructuredFieldError := lowering.destructuredField(binding, destructured, field)
			if destructuredFieldError != nil {
				return nil, destructuredFieldError
			}
			fieldType = lowering.checker.GetTypeOfSymbol(property)
			absent = property.Flags&ast.SymbolFlagsOptional != 0
		}
		local, declareLocalError := lowering.declareLocal(binding.Name())
		if declareLocalError != nil {
			return nil, declareLocalError
		}
		of := lowering.result.Locals[local].Type
		if element, isKnown := lowering.representation(
			fieldType,
		); !isKnown || element != of || (slotless(of) && !(of == ir.Union && lowering.writable(fieldType))) {
			return nil, lowering.notYet(binding, "a destructured name held otherwise than its field")
		}
		value := ir.Property{Object: ir.Read{Local: held, Of: ir.Object}, Name: field, Of: of, Absent: absent}
		value.View = sourceExpression(binding) + " (field " + field + ")"
		value.ViewReceiverTypeID = lowering.viewFieldReceiverTypeID(nil, destructured)
		value.ViewTypeID = lowering.viewTypeId(fieldType)
		value.ViewWhere = lowering.program.Where(binding)
		value.ViewType = lowering.checker.TypeToString(fieldType)
		value.ViewAllowed = lowering.viewLiterals(fieldType)
		statements = append(statements, lowering.initializeLocal(local, value)...)
	}
	return statements, nil
}

// destructuredField is the field { x } reads off a value of type destructured, in a declaration or
// an assignment: one the type names and holds as its own, never a method, an inherited library
// member or an erased method.
func (lowering *lowering) destructuredField(
	where *ast.Node,
	destructured *checker.Type,
	field string,
) (*ast.Symbol, error) {
	property := lowering.checker.GetPropertyOfType(destructured, field)
	if property == nil {
		return nil, lowering.notYet(where, "destructuring a field the type doesn't name")
	}
	ownResultField := lowering.iteratorResultField(nil, property)
	if lowering.inheritedLibrarySymbol(property) && !ownResultField {
		return nil, lowering.prototypeRead(where, field)
	}
	if property.Flags&ast.SymbolFlagsMethod != 0 {
		return nil, &Refused{
			Where: lowering.program.Where(where),
			What:  "a method in object destructuring",
			Fix:   "call it on its receiver or wrap that call in an arrow; destructuring would lose this",
		}
	}
	for _, root := range lowering.checker.GetRootSymbols(property) {
		for _, declaration := range root.Declarations {
			if load.IsLibrary(ast.GetSourceFileOfNode(declaration)) && !ownResultField {
				return nil, lowering.notYet(
					where,
					"an inherited library member in object destructuring, which is not an own field",
				)
			}
		}
	}
	if erasedMethodFieldError := lowering.erasedMethodField(where, destructured, field); erasedMethodFieldError != nil {
		return nil, erasedMethodFieldError
	}
	return property, nil
}

// destructureString uses the same UTF-16 length read as string.length. The caller
// saves the source once, even for an empty pattern or repeated property reads.
func (lowering *lowering) destructureString(pattern *ast.Node, held int) ([]ir.Statement, error) {
	statements := []ir.Statement{}
	for _, binding := range pattern.AsBindingPattern().Elements.Nodes {
		declared := binding.AsBindingElement()
		if !ast.IsIdentifier(binding.Name()) || declared.Initializer != nil || declared.DotDotDotToken != nil {
			return nil, lowering.notYet(
				binding,
				"a string destructured name that isn't plain; save the string and read its properties explicitly",
			)
		}
		field := binding.Name().Text()
		if declared.PropertyName != nil {
			if !ast.IsIdentifier(declared.PropertyName) && declared.PropertyName.Kind != ast.KindStringLiteral {
				return nil, lowering.notYet(
					declared.PropertyName,
					"a computed string destructuring member; save the string and read its properties explicitly",
				)
			}
			field = declared.PropertyName.Text()
		}
		if field != "length" {
			return nil, lowering.notYet(binding, "string member '"+
				field+
				"' in object destructuring; call methods on the saved string receiver or wrap the call in an arrow")
		}
		local, declareLocalError := lowering.declareLocal(binding.Name())
		if declareLocalError != nil {
			return nil, declareLocalError
		}
		if lowering.result.Locals[local].Type != ir.Number {
			return nil, lowering.notYet(
				binding,
				"a destructured string length held otherwise than a number; use an explicitly typed string source",
			)
		}
		value := ir.StringLength{Value: ir.Read{Local: held, Of: ir.String}}
		statements = append(statements, lowering.initializeLocal(local, value)...)
	}
	return statements, nil
}

// optionalTupleElement is tuple?.[index], a tuple that may be missing: its field when it's there, and
// undefined when it isn't, which a number field gives as number | undefined and a reference field as
// a null one. Fields held any other way stay not yet, as for ?. on a field (object.go's property).
func (lowering *lowering) optionalTupleElement(
	node *ast.Node,
	object ir.Expression,
	index *ast.Node,
) (ir.Expression, error) {
	access := node.AsElementAccessExpression()
	tuple := lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(access.Expression))
	if index.Kind != ast.KindNumericLiteral || !checker.IsTupleType(tuple) {
		return nil, lowering.notYet(node, "?.[] on anything but a tuple, at a position written out")
	}
	elements := lowering.checker.GetTypeArguments(tuple)
	position, parseError := strconv.Atoi(index.Text())
	if parseError != nil || position >= len(elements) {
		return nil, lowering.notYet(node, "?.[] past a tuple's end")
	}
	held, isKnown := lowering.representation(elements[position])
	if !isKnown || (held != ir.Number && !held.IsReference()) || held == ir.Weak {
		return nil, lowering.notYet(
			node,
			"?.[] to a tuple element of type "+lowering.checker.TypeToString(elements[position]),
		)
	}
	return ir.Property{Object: object, Name: strconv.Itoa(position), Of: held, Optional: true}, nil
}

// everyKnown reports whether every type has a representation.
func (lowering *lowering) everyKnown(types []*checker.Type) bool {
	for _, each := range types {
		if _, isKnown := lowering.representation(each); !isKnown {
			return false
		}
	}
	return true
}
