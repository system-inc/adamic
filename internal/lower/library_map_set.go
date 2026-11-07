package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Set operations become ordinary typed functions in the IR. Both backends and the ownership
// analyses then see every read and write, with no opaque runtime callback or hidden allocation.
func (lowering *lowering) librarySetMethod(node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	switch name {
	case "union", "intersection", "difference", "symmetricDifference", "isSubsetOf", "isSupersetOf", "isDisjointFrom":
	default:
		return nil, false, nil
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 || !lowering.isSet(arguments[0]) {
		return nil, true, lowering.notYet(node, name+" with anything but a concrete library Set")
	}
	element, setElementError := lowering.setElement(receiver)
	if setElementError != nil {
		return nil, true, setElementError
	}
	otherElement, setElementError := lowering.setElement(arguments[0])
	if setElementError != nil {
		return nil, true, setElementError
	}
	if element != otherElement {
		return nil, true, lowering.notYet(node, name+" between Sets whose elements have different representations")
	}
	left, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	right, expressionError := lowering.expression(arguments[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	index := len(lowering.result.Functions)
	function := ir.Function{Name: "set_" + name, Returns: ir.Map}
	local := func(name string, of ir.Type) ir.Read {
		id := len(lowering.result.Locals)
		lowering.result.Locals = append(lowering.result.Locals, ir.Local{Name: name, Type: of, Function: index})
		return ir.Read{Local: id, Of: of}
	}
	a, b := local("left", ir.Map), local("right", ir.Map)
	function.Parameters = []int{a.Local, b.Local}
	item := local("element", element)
	has := func(set ir.Expression) ir.Expression { return ir.MapHas{Map: set, Key: item, KeyType: element} }
	not := func(value ir.Expression) ir.Expression { return ir.Unary{Operator: ir.Not, Operand: value} }
	walk := func(set ir.Expression, body ...ir.Statement) ir.Statement {
		return ir.ForOf{Iterable: set, MapPart: "keys", Key: element, Value: ir.Number, Local: item.Local, Body: body}
	}
	smaller := ir.Binary{Operator: ir.LessOrEqual, Left: ir.MapSize{Map: a}, Right: ir.MapSize{Map: b}}
	switch name {
	case "isSubsetOf", "isSupersetOf", "isDisjointFrom":
		function.Returns = ir.Boolean
		no := ir.Return{Value: ir.BooleanConstant{Value: false}}
		check := func(source, other ir.Expression, present bool) ir.Statement {
			condition := has(other)
			if !present {
				condition = not(condition)
			}
			return walk(source, ir.If{Condition: condition, Then: []ir.Statement{no}})
		}
		if name == "isSubsetOf" {
			function.Body = []ir.Statement{ir.If{Condition: not(smaller), Then: []ir.Statement{no}}, check(a, b, false)}
		} else if name == "isSupersetOf" {
			function.Body = []ir.Statement{ir.If{
				Condition: ir.Binary{Operator: ir.Less, Left: ir.MapSize{Map: a}, Right: ir.MapSize{Map: b}},
				Then:      []ir.Statement{no},
			}, check(b, a, false)}
		} else {
			function.Body = []ir.Statement{ir.If{
				Condition: smaller,
				Then:      []ir.Statement{check(a, b, true)},
				Else:      []ir.Statement{check(b, a, true)},
			}}
		}
		function.Body = append(function.Body, ir.Return{Value: ir.BooleanConstant{Value: true}})
	default:
		result := local("result", ir.Map)
		// The result is a fresh Set of the call's result type. Record its writes for the cycle proof.
		lowering.writeSites = append(
			lowering.writeSites,
			writeSite{holder: lowering.concrete(lowering.checker.GetTypeAtLocation(node)), node: node},
		)
		site := len(lowering.writeSites)
		add := ir.Evaluate{Value: ir.SetAdd{Set: result, Value: item, Element: element, Site: site}}
		copy := func(source ir.Expression) ir.Statement { return walk(source, add) }
		selected := func(source, other ir.Expression, present bool) ir.Statement {
			condition := has(other)
			if !present {
				condition = not(condition)
			}
			return walk(source, ir.If{Condition: condition, Then: []ir.Statement{add}})
		}
		function.Body = []ir.Statement{ir.Declare{Local: result.Local, Value: ir.SetNew{Element: element}}}
		switch name {
		case "union":
			function.Body = append(function.Body, copy(a), copy(b))
		case "intersection":
			// The smaller Set supplies the order; ties visit the receiver (ECMA-262).
			function.Body = append(function.Body, ir.If{
				Condition: smaller,
				Then:      []ir.Statement{selected(a, b, true)},
				Else:      []ir.Statement{selected(b, a, true)},
			})
		case "difference":
			function.Body = append(function.Body, selected(a, b, false))
		case "symmetricDifference":
			function.Body = append(function.Body, selected(a, b, false), selected(b, a, false))
		}
		function.Body = append(function.Body, ir.Return{Value: result})
	}
	lowering.result.Functions = append(lowering.result.Functions, function)
	return ir.Call{Function: index, Arguments: []ir.Expression{left, right}, Returns: function.Returns}, true, nil
}

func (lowering *lowering) libraryCollectionIterator(
	node, receiver *ast.Node,
	part string,
) (ir.Expression, bool, error) {
	if len(node.AsCallExpression().Arguments.Nodes) != 0 {
		return nil, true, lowering.notYet(node, part+" with arguments")
	}
	set := lowering.isSet(receiver)
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
	return ir.CollectionIterator{Collection: collection, Part: part, Key: key, Value: value, Set: set}, true, nil
}

func (lowering *lowering) librarySetLiteral(node *ast.Node, element ir.Type) (ir.Expression, bool, error) {
	array := ir.ArrayLiteral{Element: element}
	for _, item := range ast.SkipParentheses(node).AsArrayLiteralExpression().Elements.Nodes {
		if item.Kind == ast.KindSpreadElement || item.Kind == ast.KindOmittedExpression {
			return nil, false, nil
		}
		value, expressionError := lowering.expression(item)
		if expressionError != nil {
			return nil, true, expressionError
		}
		value = fit(value, element)
		if value.Type() != element {
			return nil, true, lowering.notYet(item, "a Set literal element held otherwise than the Set's")
		}
		array.Elements = append(array.Elements, value)
	}
	return array, true, nil
}

func (lowering *lowering) libraryEmptyCollectionArgument(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNullKeyword {
		return true
	}
	_, local := lowering.local(node)
	return ast.IsIdentifier(node) && node.Text() == "undefined" && !local
}

// An arrow's this is lexical: thisArg is still evaluated in call order, but cannot bind it.
// Other callbacks stay refused until stage 0 can represent and check their dynamic this parameter.
func (lowering *lowering) libraryForEachThisArg(node, receiver *ast.Node, set bool) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if ast.SkipParentheses(arguments[0]).Kind != ast.KindArrowFunction {
		return nil, true, lowering.notYet(node, "forEach thisArg with a callback other than an arrow")
	}
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
	callback, expressionError := lowering.expression(arguments[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	context, expressionError := lowering.expression(arguments[1])
	if expressionError != nil {
		return nil, true, expressionError
	}
	// Pass the context to an unused parameter, preserving evaluation order, side effects,
	// exceptions and its lifetime across the visit.
	if context.Type() == 0 {
		return nil, true, lowering.notYet(arguments[1], "forEach thisArg without a value representation")
	}
	values := []ir.Expression{collection, callback, context}
	index := len(lowering.result.Functions)
	function := ir.Function{Name: "collection_forEach_thisArg"}
	for _, argument := range values {
		id := len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: "argument", Type: argument.Type(), Function: index},
		)
		function.Parameters = append(function.Parameters, id)
	}
	// Read the actual arrow's return representation, including an owned result discarded by forEach.
	returns := lowering.result.Functions[callback.(ir.MakeClosure).Function].Returns
	if slotless(returns) {
		return nil, true, lowering.notYet(node, "forEach with a callback result held in more than one word")
	}
	function.Body = []ir.Statement{ir.Evaluate{Value: ir.MapForEach{
		Map:          ir.Read{Local: function.Parameters[0], Of: ir.Map},
		Callback:     ir.Read{Local: function.Parameters[1], Of: ir.Closure},
		Key:          key,
		Value:        value,
		Set:          set,
		Returns:      returns,
		CallbackType: int(lowering.concrete(lowering.checker.GetTypeAtLocation(arguments[0])).Id()),
	}}}
	lowering.result.Functions = append(lowering.result.Functions, function)
	return ir.Call{Function: index, Arguments: values}, true, nil
}

func (lowering *lowering) libraryLocal(function int, name string, of ir.Type) ir.Read {
	id := len(lowering.result.Locals)
	lowering.result.Locals = append(lowering.result.Locals, ir.Local{Name: name, Type: of, Function: function})
	return ir.Read{Local: id, Of: of}
}

// A stored iterator is consumed by calls to next, rather than by snapshotting its collection.
func (lowering *lowering) libraryIteratorElement(node *ast.Node) (ir.Type, bool) {
	proven := lowering.checker.GetTypeAtLocation(node)
	if !lowering.isLibraryType(proven, "MapIterator", "SetIterator") {
		return 0, false
	}
	arguments := lowering.typeArguments(proven)
	if len(arguments) != 1 {
		return 0, false
	}
	element, known := lowering.representation(arguments[0])
	// Boxed unions already occupy one tagged reference slot in the iterator.
	return element, known && boxedSlot(element)
}

func (lowering *lowering) libraryIteratorLoop(
	function int,
	iterator ir.Expression,
	item ir.Read,
	body []ir.Statement,
) []ir.Statement {
	held := lowering.libraryLocal(function, "iterator", ir.Object)
	step := lowering.libraryLocal(function, "step", ir.Object)
	loop, value := libraryIteratorStep(held, step, item.Of)
	loop = append(loop, ir.Declare{Local: item.Local, Value: value})
	loop = append(loop, body...)
	return []ir.Statement{
		ir.Declare{Local: held.Local, Value: iterator},
		ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: loop},
	}
}

// libraryIteratorStep is one step of a library iterator held in held: its next called into step, a
// break when it says done, and the step's value, read as of. libraryIteratorLoop steps by it, and so
// does an iterable view's dispatch, its collection arm (iterable_views.go), never by a copy of it
// (Lower's condition 4 on #qk8rztp).
func libraryIteratorStep(held, step ir.Read, of ir.Type) ([]ir.Statement, ir.Expression) {
	return []ir.Statement{
		ir.Declare{
			Local: step.Local,
			Value: ir.CallClosure{Closure: ir.Property{Object: held, Name: "next", Of: ir.Closure}, Returns: ir.Object},
		},
		ir.If{Condition: ir.Property{Object: step, Name: "done", Of: ir.Boolean}, Then: []ir.Statement{ir.Break{}}},
	}, ir.Property{Object: step, Name: "value", Of: of}
}

func (lowering *lowering) libraryIteratorArray(node *ast.Node, iterator ir.Expression, element ir.Type) ir.Expression {
	index := len(lowering.result.Functions)
	function := ir.Function{Name: "collection_iterator_array", Returns: ir.Array}
	parameter := lowering.libraryLocal(index, "input", ir.Object)
	array := lowering.libraryLocal(index, "values", ir.Array)
	item := lowering.libraryLocal(index, "element", element)
	function.Parameters = []int{parameter.Local}
	function.Body = []ir.Statement{ir.Declare{Local: array.Local, Value: ir.ArrayLiteral{Element: element}}}
	lowering.writeSites = append(lowering.writeSites, writeSite{node: node})
	site := len(lowering.writeSites)
	function.Body = append(function.Body, lowering.libraryIteratorLoop(
		index,
		parameter,
		item,
		[]ir.Statement{ir.Evaluate{Value: ir.ArrayPush{Array: array, Value: item, Element: element, Site: site}}},
	)...)
	function.Body = append(function.Body, ir.Return{Value: array})
	lowering.result.Functions = append(lowering.result.Functions, function)
	return ir.Call{Function: index, Arguments: []ir.Expression{iterator}, Returns: ir.Array}
}

func (lowering *lowering) libraryForOfIterator(
	node *ast.Node,
	iterator ir.Expression,
	element ir.Type,
	name *ast.Node,
) ([]ir.Statement, error) {
	item := lowering.libraryLocal(lowering.functionIndex, "element", element)
	var bindings []ir.Statement
	if ast.IsIdentifier(name) {
		local, declareLocalError := lowering.declareLocal(name)
		if declareLocalError != nil {
			return nil, declareLocalError
		}
		bindings = append(bindings, ir.Declare{Local: local, Value: item})
	} else {
		if element != ir.Object {
			return nil, lowering.notYet(name, "destructuring a scalar iterator element")
		}
		for index, binding := range name.AsBindingPattern().Elements.Nodes {
			if skipped(binding) {
				continue
			}
			declared := binding.AsBindingElement()
			if !ast.IsIdentifier(binding.Name()) || declared.Initializer != nil || declared.DotDotDotToken != nil {
				return nil, lowering.notYet(binding, "an iterator binding that isn't plain")
			}
			local, declareLocalError := lowering.declareLocal(binding.Name())
			if declareLocalError != nil {
				return nil, declareLocalError
			}
			of := lowering.result.Locals[local].Type
			if !boxedSlot(of) {
				return nil, lowering.notYet(binding, "an iterator binding held in more than one word")
			}
			bindings = append(
				bindings,
				ir.Declare{Local: local, Value: ir.Property{Object: item, Name: strconv.Itoa(index), Of: of}},
			)
		}
	}
	body, statementError := lowering.statement(node.AsForInOrOfStatement().Statement)
	if statementError != nil {
		return nil, statementError
	}
	return []ir.Statement{
		ir.Block{Body: lowering.libraryIteratorLoop(lowering.functionIndex, iterator, item, append(bindings, body...))},
	}, nil
}

// groupBy copies elements into fresh bucket arrays while walking the original input live. Callback
// mutation therefore sees the same array/collection iterator that Node walks, not a snapshot.
func (lowering *lowering) libraryMapGroupBy(node *ast.Node) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 2 {
		return nil, true, lowering.notYet(node, "Map.groupBy with other than an iterable and callback")
	}
	key, _, err := lowering.mapTypes(node)
	if err != nil {
		return nil, true, err
	}
	sourceNode := arguments[0]
	source, err := lowering.expression(sourceNode)
	if err != nil {
		return nil, true, err
	}
	var element ir.Type
	iterator := false
	switch source.Type() {
	case ir.Array:
		element, err = lowering.elementType(sourceNode)
	case ir.String:
		element = ir.String
	case ir.Map:
		var k, value ir.Type
		if lowering.isSet(sourceNode) {
			k, err = lowering.setElement(sourceNode)
			value = k
			element = k
		} else {
			k, value, err = lowering.mapTypes(sourceNode)
			element = ir.Object
		}
		part := "entries"
		if lowering.isSet(sourceNode) {
			part = "keys"
		}
		source = ir.CollectionIterator{
			Collection: source,
			Part:       part,
			Key:        k,
			Value:      value,
			Set:        lowering.isSet(sourceNode),
		}
		iterator = true
	case ir.Object:
		var known bool
		element, known = lowering.libraryIteratorElement(sourceNode)
		if !known {
			return nil, true, lowering.notYet(sourceNode, "Map.groupBy over an unsupported iterable")
		}
		iterator = true
	default:
		return nil, true, lowering.notYet(sourceNode, "Map.groupBy over an unsupported iterable")
	}
	if err != nil {
		return nil, true, err
	}
	if slotless(element) {
		return nil, true, lowering.notYet(sourceNode, "Map.groupBy elements held in more than one word")
	}
	// The checker must retain the source element's exact type in each bucket; widening a mutable
	// element would expose an alias through the returned arrays. Refuse such inferred widening.
	var sourceType *checker.Type
	proven := lowering.checker.GetTypeAtLocation(sourceNode)
	if lowering.isArrayType(proven) {
		sourceType = lowering.elementTypeOfArray(proven)
	}
	if lowering.isSet(sourceNode) || lowering.isLibraryType(proven, "MapIterator", "SetIterator") {
		sourceType = lowering.typeArguments(proven)[0]
	}
	grouped := lowering.typeArguments(lowering.checker.GetTypeAtLocation(node))
	if len(grouped) != 2 {
		return nil, true, lowering.notYet(node, "Map.groupBy without known grouped elements")
	}
	groupedElement := lowering.elementTypeOfArray(grouped[1])
	if held, known := lowering.representation(groupedElement); !known || held != element {
		return nil, true, lowering.notYet(node, "Map.groupBy grouped elements held otherwise than the input's")
	}
	if lowering.isLibraryType(proven, "Map", "ReadonlyMap") {
		if !checker.IsTupleType(groupedElement) {
			return nil, true, lowering.notYet(node, "Map.groupBy widening its Map entries")
		}
		components := lowering.typeArguments(groupedElement)
		original := lowering.typeArguments(proven)
		if len(components) != 2 ||
			lowering.checker.TypeToString(components[0]) != lowering.checker.TypeToString(original[0]) ||
			lowering.checker.TypeToString(components[1]) != lowering.checker.TypeToString(original[1]) {
			return nil, true, lowering.notYet(node, "Map.groupBy widening its Map entry components")
		}
	}
	if sourceType != nil && lowering.checker.TypeToString(sourceType) != lowering.checker.TypeToString(groupedElement) {
		return nil, true, lowering.notYet(node, "Map.groupBy widening its mutable elements")
	}
	callback, err := lowering.expression(arguments[1])
	if err != nil {
		return nil, true, err
	}
	if callback.Type() != ir.Closure {
		return nil, true, lowering.notYet(node, "Map.groupBy without a function callback")
	}
	signatures := lowering.checker.GetSignaturesOfType(
		lowering.checker.GetTypeAtLocation(arguments[1]),
		checker.SignatureKindCall,
	)
	if len(signatures) != 1 {
		return nil, true, lowering.notYet(node, "Map.groupBy with an overloaded callback")
	}
	if parameters := signatures[0].Parameters(); len(parameters) > 0 {
		received, known := lowering.representation(lowering.checker.GetTypeOfSymbol(parameters[0]))
		if !known || received != element {
			return nil, true, lowering.notYet(node, "Map.groupBy callback elements held otherwise than the input's")
		}
	}
	returns, known := lowering.representation(lowering.checker.GetReturnTypeOfSignature(signatures[0]))
	if !known || returns != key {
		return nil, true, lowering.notYet(node, "Map.groupBy callback keys held otherwise than the result's")
	}
	index := len(lowering.result.Functions)
	function := ir.Function{Name: "map_groupBy", Returns: ir.Map}
	input := lowering.libraryLocal(index, "input", source.Type())
	fn := lowering.libraryLocal(index, "callback", ir.Closure)
	result := lowering.libraryLocal(index, "groups", ir.Map)
	position := lowering.libraryLocal(index, "index", ir.Number)
	item := lowering.libraryLocal(index, "element", element)
	groupKey := lowering.libraryLocal(index, "key", key)
	bucket := lowering.libraryLocal(index, "bucket", ir.Array)
	function.Parameters = []int{input.Local, fn.Local}
	lowering.writeSites = append(
		lowering.writeSites,
		writeSite{holder: lowering.concrete(lowering.checker.GetTypeAtLocation(node)), node: node},
	)
	mapSite := len(lowering.writeSites)
	lowering.writeSites = append(lowering.writeSites, writeSite{holder: grouped[1], node: node})
	arraySite := len(lowering.writeSites)
	body := []ir.Statement{
		ir.Declare{
			Local: groupKey.Local,
			Value: ir.CallClosure{Closure: fn, Arguments: []ir.Expression{item, position}, Returns: key},
		},
		ir.Assign{
			Local: position.Local,
			Value: ir.Binary{Operator: ir.Add, Left: position, Right: ir.NumberConstant{Value: 1}},
		},
		ir.Declare{Local: bucket.Local, Value: ir.Coalesce{
			Value:    ir.MapGet{Map: result, Key: groupKey, KeyType: key, ValueType: ir.Array},
			Fallback: ir.ArrayLiteral{Element: element},
			Of:       ir.Array,
		}},
		ir.Evaluate{Value: ir.ArrayPush{Array: bucket, Value: item, Element: element, Site: arraySite}},
		ir.Evaluate{Value: ir.MapSet{
			Map:       result,
			Key:       groupKey,
			Value:     bucket,
			KeyType:   key,
			ValueType: ir.Array,
			Site:      mapSite,
		}},
	}
	function.Body = []ir.Statement{
		ir.Declare{Local: result.Local, Value: ir.MapNew{Key: key, Value: ir.Array}},
		ir.Declare{Local: position.Local, Value: ir.NumberConstant{}},
	}
	if iterator {
		function.Body = append(function.Body, lowering.libraryIteratorLoop(index, input, item, body)...)
	} else {
		function.Body = append(
			function.Body,
			ir.ForOf{Iterable: input, Element: element, Local: item.Local, Body: body},
		)
	}
	function.Body = append(function.Body, ir.Return{Value: result})
	lowering.result.Functions = append(lowering.result.Functions, function)
	return ir.Call{Function: index, Arguments: []ir.Expression{source, callback}, Returns: ir.Map}, true, nil
}

// Copy present numeric elements into packed slots when the Set's inferred or explicit type also
// admits undefined. Reusing the input array would change how its existing aliases read its NaNs.
func (lowering *lowering) libraryOptionalNumbers(node *ast.Node, source ir.Expression) ir.Expression {
	index := len(lowering.result.Functions)
	function := ir.Function{Name: "collection_optional_numbers", Returns: ir.Array}
	input := lowering.libraryLocal(index, "input", ir.Array)
	array := lowering.libraryLocal(index, "values", ir.Array)
	item := lowering.libraryLocal(index, "number", ir.Number)
	function.Parameters = []int{input.Local}
	lowering.writeSites = append(lowering.writeSites, writeSite{node: node})
	site := len(lowering.writeSites)
	function.Body = []ir.Statement{
		ir.Declare{Local: array.Local, Value: ir.ArrayLiteral{Element: ir.MaybeNumber}},
		ir.ForOf{Iterable: input, Element: ir.Number, Local: item.Local, Body: []ir.Statement{ir.Evaluate{
			Value: ir.ArrayPush{Array: array, Value: fit(item, ir.MaybeNumber), Element: ir.MaybeNumber, Site: site},
		}}},
		ir.Return{Value: array},
	}
	lowering.result.Functions = append(lowering.result.Functions, function)
	return ir.Call{Function: index, Arguments: []ir.Expression{source}, Returns: ir.Array}
}

// Collection iterators hold their source strongly, though the public iterator type only names its
// yielded elements. Keep this hidden edge visible to the cycle finder just like a closure capture.
type libraryIteratorCapture struct {
	iterator, collection *checker.Type
}

func (finder *cycleFinder) libraryIteratorMade(node *ast.Node) {
	if node.Kind != ast.KindCallExpression {
		return
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return
	}
	access := callee.AsPropertyAccessExpression()
	iterator := finder.lowering.checker.GetTypeAtLocation(node)
	collection := finder.lowering.checker.GetTypeAtLocation(access.Expression)
	if finder.lowering.isLibraryType(iterator, "MapIterator", "SetIterator") &&
		finder.lowering.isLibraryType(collection, "Map", "ReadonlyMap", "Set", "ReadonlySet") {
		finder.libraryIterators = append(finder.libraryIterators, libraryIteratorCapture{iterator, collection})
	}
}

func (finder *cycleFinder) libraryIteratorCaptures(proven *checker.Type) []cycleNode {
	var held []cycleNode
	for _, capture := range finder.libraryIterators {
		if finder.lowering.checker.IsTypeAssignableTo(capture.iterator, proven) {
			held = append(held, cycleNode{proven: capture.collection})
		}
	}
	return held
}

// Iterator next is inherited in JavaScript. Copying our internal closure slot as an own
// enumerable field would let a copy drive the original iterator, unlike Node.
func (lowering *lowering) libraryIteratorUnsupportedUse(node *ast.Node) error {
	isIterator := func(source *ast.Node) bool {
		return lowering.libraryIteratorType(lowering.checker.GetTypeAtLocation(source))
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind == ast.KindSpreadAssignment && isIterator(property.AsSpreadAssignment().Expression) {
				return lowering.notYet(
					property,
					"spreading a Map or Set iterator: next is inherited, not an own enumerable field; use the original iterator or collect its elements into an array",
				)
			}
		}
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind == ast.KindPropertyAccessExpression &&
			callee.Name().Text() == "assign" &&
			lowering.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
			for index, source := range call.Arguments.Nodes {
				if index == 0 {
					continue
				}
				if isIterator(source) {
					return lowering.notYet(
						source,
						"Object.assign copying a Map or Set iterator: next is inherited, not an own enumerable field; use the original iterator",
					)
				}
			}
		}
	}
	return nil
}

func (lowering *lowering) libraryIteratorType(proven *checker.Type) bool {
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if lowering.libraryIteratorType(member) {
				return true
			}
		}
	}
	return lowering.isLibraryType(proven, "MapIterator", "SetIterator")
}
