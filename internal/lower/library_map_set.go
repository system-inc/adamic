package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Set operations become ordinary typed functions in the IR. Both backends and the ownership
// analyses then see every read and write, with no opaque runtime callback or hidden allocation.
func (l *lowering) librarySetMethod(node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	switch name {
	case "union", "intersection", "difference", "symmetricDifference", "isSubsetOf", "isSupersetOf", "isDisjointFrom":
	default:
		return nil, false, nil
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 1 || !l.isSet(arguments[0]) {
		return nil, true, l.notYet(node, name+" with anything but a concrete library Set")
	}
	element, err := l.setElement(receiver)
	if err != nil {
		return nil, true, err
	}
	otherElement, err := l.setElement(arguments[0])
	if err != nil {
		return nil, true, err
	}
	if element != otherElement {
		return nil, true, l.notYet(node, name+" between Sets whose elements have different representations")
	}
	left, err := l.expression(receiver)
	if err != nil {
		return nil, true, err
	}
	right, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	index := len(l.result.Functions)
	function := ir.Function{Name: "set_" + name, Returns: ir.Map}
	local := func(name string, of ir.Type) ir.Read {
		id := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: index})
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
			function.Body = []ir.Statement{ir.If{Condition: ir.Binary{Operator: ir.Less, Left: ir.MapSize{Map: a}, Right: ir.MapSize{Map: b}}, Then: []ir.Statement{no}}, check(b, a, false)}
		} else {
			function.Body = []ir.Statement{ir.If{Condition: smaller, Then: []ir.Statement{check(a, b, true)}, Else: []ir.Statement{check(b, a, true)}}}
		}
		function.Body = append(function.Body, ir.Return{Value: ir.BooleanConstant{Value: true}})
	default:
		result := local("result", ir.Map)
		// The result is a fresh Set of the call's result type. Record its writes for the cycle proof.
		l.writeSites = append(l.writeSites, writeSite{holder: l.concrete(l.checker.GetTypeAtLocation(node)), node: node})
		site := len(l.writeSites)
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
			function.Body = append(function.Body, ir.If{Condition: smaller, Then: []ir.Statement{selected(a, b, true)}, Else: []ir.Statement{selected(b, a, true)}})
		case "difference":
			function.Body = append(function.Body, selected(a, b, false))
		case "symmetricDifference":
			function.Body = append(function.Body, selected(a, b, false), selected(b, a, false))
		}
		function.Body = append(function.Body, ir.Return{Value: result})
	}
	l.result.Functions = append(l.result.Functions, function)
	return ir.Call{Function: index, Arguments: []ir.Expression{left, right}, Returns: function.Returns}, true, nil
}

func (l *lowering) libraryCollectionIterator(node, receiver *ast.Node, part string) (ir.Expression, bool, error) {
	if len(node.AsCallExpression().Arguments.Nodes) != 0 {
		return nil, true, l.notYet(node, part+" with arguments")
	}
	set := l.isSet(receiver)
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
	return ir.CollectionIterator{Collection: collection, Part: part, Key: key, Value: value, Set: set}, true, nil
}

func (l *lowering) librarySetLiteral(node *ast.Node, element ir.Type) (ir.Expression, bool, error) {
	array := ir.ArrayLiteral{Element: element}
	for _, item := range ast.SkipParentheses(node).AsArrayLiteralExpression().Elements.Nodes {
		if item.Kind == ast.KindSpreadElement || item.Kind == ast.KindOmittedExpression {
			return nil, false, nil
		}
		value, err := l.expression(item)
		if err != nil {
			return nil, true, err
		}
		value = fit(value, element)
		if value.Type() != element {
			return nil, true, l.notYet(item, "a Set literal element held otherwise than the Set's")
		}
		array.Elements = append(array.Elements, value)
	}
	return array, true, nil
}

func (l *lowering) libraryEmptyCollectionArgument(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindNullKeyword {
		return true
	}
	_, local := l.local(node)
	return ast.IsIdentifier(node) && node.Text() == "undefined" && !local
}

// An arrow's this is lexical: thisArg is still evaluated in call order, but cannot bind it.
// Other callbacks stay refused until stage 0 can represent and check their dynamic this parameter.
func (l *lowering) libraryForEachThisArg(node, receiver *ast.Node, set bool) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if ast.SkipParentheses(arguments[0]).Kind != ast.KindArrowFunction {
		return nil, true, l.notYet(node, "forEach thisArg with a callback other than an arrow")
	}
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
	callback, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	context, err := l.expression(arguments[1])
	if err != nil {
		return nil, true, err
	}
	// Pass the context to an unused parameter, preserving evaluation order, side effects,
	// exceptions and its lifetime across the visit.
	if context.Type() == 0 {
		return nil, true, l.notYet(arguments[1], "forEach thisArg without a value representation")
	}
	values := []ir.Expression{collection, callback, context}
	index := len(l.result.Functions)
	function := ir.Function{Name: "collection_forEach_thisArg"}
	for _, argument := range values {
		id := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: argument.Type(), Function: index})
		function.Parameters = append(function.Parameters, id)
	}
	// Read the actual arrow's return representation, including an owned result discarded by forEach.
	returns := l.result.Functions[callback.(ir.MakeClosure).Function].Returns
	if slotless(returns) {
		return nil, true, l.notYet(node, "forEach with a callback result held in more than one word")
	}
	function.Body = []ir.Statement{ir.Evaluate{Value: ir.MapForEach{Map: ir.Read{Local: function.Parameters[0], Of: ir.Map}, Callback: ir.Read{Local: function.Parameters[1], Of: ir.Closure}, Key: key, Value: value, Set: set, Returns: returns}}}
	l.result.Functions = append(l.result.Functions, function)
	return ir.Call{Function: index, Arguments: values}, true, nil
}

func (l *lowering) libraryLocal(function int, name string, of ir.Type) ir.Read {
	id := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: function})
	return ir.Read{Local: id, Of: of}
}

// A stored iterator is consumed by calls to next, rather than by snapshotting its collection.
func (l *lowering) libraryIteratorElement(node *ast.Node) (ir.Type, bool) {
	proven := l.checker.GetTypeAtLocation(node)
	if !l.isLibraryType(proven, "MapIterator", "SetIterator", "ArrayIterator", "StringIterator") {
		return 0, false
	}
	arguments := l.typeArguments(proven)
	if len(arguments) != 1 {
		return 0, false
	}
	element, known := l.representation(arguments[0])
	return element, known && !slotless(element)
}

func (l *lowering) libraryIteratorLoop(function int, iterator ir.Expression, item ir.Read, body []ir.Statement) []ir.Statement {
	held := l.libraryLocal(function, "iterator", ir.Object)
	step := l.libraryLocal(function, "step", ir.Object)
	loop := []ir.Statement{
		ir.Declare{Local: step.Local, Value: ir.CallClosure{Closure: ir.Property{Object: held, Name: "next", Of: ir.Closure, Method: true}, Returns: ir.Object}},
		ir.If{Condition: ir.Property{Object: step, Name: "done", Of: ir.Boolean}, Then: []ir.Statement{ir.Break{}}},
		ir.Declare{Local: item.Local, Value: ir.Property{Object: step, Name: "value", Of: item.Of}},
	}
	loop = append(loop, body...)
	return []ir.Statement{ir.Declare{Local: held.Local, Value: iterator}, ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: loop}}
}

func (l *lowering) libraryIteratorArray(node *ast.Node, iterator ir.Expression, element ir.Type) ir.Expression {
	index := len(l.result.Functions)
	function := ir.Function{Name: "collection_iterator_array", Returns: ir.Array}
	parameter := l.libraryLocal(index, "input", ir.Object)
	array := l.libraryLocal(index, "values", ir.Array)
	item := l.libraryLocal(index, "element", element)
	function.Parameters = []int{parameter.Local}
	function.Body = []ir.Statement{ir.Declare{Local: array.Local, Value: ir.ArrayLiteral{Element: element}}}
	l.writeSites = append(l.writeSites, writeSite{node: node})
	site := len(l.writeSites)
	function.Body = append(function.Body, l.libraryIteratorLoop(index, parameter, item, []ir.Statement{ir.Evaluate{Value: ir.ArrayPush{Array: array, Value: item, Element: element, Site: site}}})...)
	function.Body = append(function.Body, ir.Return{Value: array})
	l.result.Functions = append(l.result.Functions, function)
	return ir.Call{Function: index, Arguments: []ir.Expression{iterator}, Returns: ir.Array}
}

func (l *lowering) libraryForOfIterator(node *ast.Node, iterator ir.Expression, element ir.Type, name *ast.Node) ([]ir.Statement, error) {
	item := l.libraryLocal(l.functionIndex, "element", element)
	var bindings []ir.Statement
	if ast.IsIdentifier(name) {
		local, err := l.declareLocal(name)
		if err != nil {
			return nil, err
		}
		bindings = append(bindings, ir.Declare{Local: local, Value: item})
	} else {
		if element != ir.Object {
			return nil, l.notYet(name, "destructuring a scalar iterator element")
		}
		for index, binding := range name.AsBindingPattern().Elements.Nodes {
			if skipped(binding) {
				continue
			}
			declared := binding.AsBindingElement()
			if !ast.IsIdentifier(binding.Name()) || declared.Initializer != nil || declared.DotDotDotToken != nil {
				return nil, l.notYet(binding, "an iterator binding that isn't plain")
			}
			local, err := l.declareLocal(binding.Name())
			if err != nil {
				return nil, err
			}
			of := l.result.Locals[local].Type
			if slotless(of) {
				return nil, l.notYet(binding, "an iterator binding held in more than one word")
			}
			bindings = append(bindings, ir.Declare{Local: local, Value: ir.Property{Object: item, Name: strconv.Itoa(index), Of: of}})
		}
	}
	body, err := l.statement(node.AsForInOrOfStatement().Statement)
	if err != nil {
		return nil, err
	}
	return []ir.Statement{ir.Block{Body: l.libraryIteratorLoop(l.functionIndex, iterator, item, append(bindings, body...))}}, nil
}

// groupBy copies elements into fresh bucket arrays while walking the original input live. Callback
// mutation therefore sees the same array/collection iterator that Node walks, not a snapshot.
func (l *lowering) libraryMapGroupBy(node *ast.Node) (ir.Expression, bool, error) {
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) != 2 {
		return nil, true, l.notYet(node, "Map.groupBy with other than an iterable and callback")
	}
	key, _, err := l.mapTypes(node)
	if err != nil {
		return nil, true, err
	}
	sourceNode := arguments[0]
	source, err := l.expression(sourceNode)
	if err != nil {
		return nil, true, err
	}
	var element ir.Type
	iterator := false
	switch source.Type() {
	case ir.Array:
		element, err = l.elementType(sourceNode)
	case ir.String:
		element = ir.String
	case ir.Map:
		var k, v ir.Type
		if l.isSet(sourceNode) {
			k, err = l.setElement(sourceNode)
			v = k
			element = k
		} else {
			k, v, err = l.mapTypes(sourceNode)
			element = ir.Object
		}
		part := "entries"
		if l.isSet(sourceNode) {
			part = "keys"
		}
		source = ir.CollectionIterator{Collection: source, Part: part, Key: k, Value: v, Set: l.isSet(sourceNode)}
		iterator = true
	case ir.Object:
		var known bool
		element, known = l.libraryIteratorElement(sourceNode)
		if !known {
			return nil, true, l.notYet(sourceNode, "Map.groupBy over an unsupported iterable")
		}
		iterator = true
	default:
		return nil, true, l.notYet(sourceNode, "Map.groupBy over an unsupported iterable")
	}
	if err != nil {
		return nil, true, err
	}
	if slotless(element) {
		return nil, true, l.notYet(sourceNode, "Map.groupBy elements held in more than one word")
	}
	// The checker must retain the source element's exact type in each bucket; widening a mutable
	// element would expose an alias through the returned arrays. Refuse such inferred widening.
	var sourceType *checker.Type
	proven := l.checker.GetTypeAtLocation(sourceNode)
	if l.checker.IsArrayType(proven) {
		sourceType = l.checker.GetElementTypeOfArrayType(proven)
	}
	if l.isSet(sourceNode) || l.isLibraryType(proven, "MapIterator", "SetIterator", "ArrayIterator", "StringIterator") {
		sourceType = l.typeArguments(proven)[0]
	}
	grouped := l.typeArguments(l.checker.GetTypeAtLocation(node))
	if len(grouped) != 2 {
		return nil, true, l.notYet(node, "Map.groupBy without known grouped elements")
	}
	groupedElement := l.checker.GetElementTypeOfArrayType(grouped[1])
	if held, known := l.representation(groupedElement); !known || held != element {
		return nil, true, l.notYet(node, "Map.groupBy grouped elements held otherwise than the input's")
	}
	if l.isLibraryType(proven, "Map", "ReadonlyMap") {
		if !checker.IsTupleType(groupedElement) {
			return nil, true, l.notYet(node, "Map.groupBy widening its Map entries")
		}
		components := l.typeArguments(groupedElement)
		original := l.typeArguments(proven)
		if len(components) != 2 || l.checker.TypeToString(components[0]) != l.checker.TypeToString(original[0]) || l.checker.TypeToString(components[1]) != l.checker.TypeToString(original[1]) {
			return nil, true, l.notYet(node, "Map.groupBy widening its Map entry components")
		}
	}
	if sourceType != nil && l.checker.TypeToString(sourceType) != l.checker.TypeToString(groupedElement) {
		return nil, true, l.notYet(node, "Map.groupBy widening its mutable elements")
	}
	callback, err := l.expression(arguments[1])
	if err != nil {
		return nil, true, err
	}
	if callback.Type() != ir.Closure {
		return nil, true, l.notYet(node, "Map.groupBy without a function callback")
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(arguments[1]), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, true, l.notYet(node, "Map.groupBy with an overloaded callback")
	}
	if parameters := signatures[0].Parameters(); len(parameters) > 0 {
		received, known := l.representation(l.checker.GetTypeOfSymbol(parameters[0]))
		if !known || received != element {
			return nil, true, l.notYet(node, "Map.groupBy callback elements held otherwise than the input's")
		}
	}
	returns, known := l.representation(l.checker.GetReturnTypeOfSignature(signatures[0]))
	if !known || returns != key {
		return nil, true, l.notYet(node, "Map.groupBy callback keys held otherwise than the result's")
	}
	index := len(l.result.Functions)
	function := ir.Function{Name: "map_groupBy", Returns: ir.Map}
	input := l.libraryLocal(index, "input", source.Type())
	fn := l.libraryLocal(index, "callback", ir.Closure)
	result := l.libraryLocal(index, "groups", ir.Map)
	position := l.libraryLocal(index, "index", ir.Number)
	item := l.libraryLocal(index, "element", element)
	groupKey := l.libraryLocal(index, "key", key)
	bucket := l.libraryLocal(index, "bucket", ir.Array)
	function.Parameters = []int{input.Local, fn.Local}
	l.writeSites = append(l.writeSites, writeSite{holder: l.concrete(l.checker.GetTypeAtLocation(node)), node: node})
	mapSite := len(l.writeSites)
	l.writeSites = append(l.writeSites, writeSite{holder: grouped[1], node: node})
	arraySite := len(l.writeSites)
	body := []ir.Statement{
		ir.Declare{Local: groupKey.Local, Value: ir.CallClosure{Closure: fn, Arguments: []ir.Expression{item, position}, Returns: key}},
		ir.Assign{Local: position.Local, Value: ir.Binary{Operator: ir.Add, Left: position, Right: ir.NumberConstant{Value: 1}}},
		ir.Declare{Local: bucket.Local, Value: ir.Coalesce{Value: ir.MapGet{Map: result, Key: groupKey, KeyType: key, ValueType: ir.Array}, Fallback: ir.ArrayLiteral{Element: element}, Of: ir.Array}},
		ir.Evaluate{Value: ir.ArrayPush{Array: bucket, Value: item, Element: element, Site: arraySite}},
		ir.Evaluate{Value: ir.MapSet{Map: result, Key: groupKey, Value: bucket, KeyType: key, ValueType: ir.Array, Site: mapSite}},
	}
	function.Body = []ir.Statement{ir.Declare{Local: result.Local, Value: ir.MapNew{Key: key, Value: ir.Array}}, ir.Declare{Local: position.Local, Value: ir.NumberConstant{}}}
	if iterator {
		function.Body = append(function.Body, l.libraryIteratorLoop(index, input, item, body)...)
	} else {
		function.Body = append(function.Body, ir.ForOf{Iterable: input, Element: element, Local: item.Local, Body: body})
	}
	function.Body = append(function.Body, ir.Return{Value: result})
	l.result.Functions = append(l.result.Functions, function)
	return ir.Call{Function: index, Arguments: []ir.Expression{source, callback}, Returns: ir.Map}, true, nil
}

// Copy present numeric elements into packed slots when the Set's inferred or explicit type also
// admits undefined. Reusing the input array would change how its existing aliases read its NaNs.
func (l *lowering) libraryOptionalNumbers(node *ast.Node, source ir.Expression) ir.Expression {
	index := len(l.result.Functions)
	function := ir.Function{Name: "collection_optional_numbers", Returns: ir.Array}
	input := l.libraryLocal(index, "input", ir.Array)
	array := l.libraryLocal(index, "values", ir.Array)
	item := l.libraryLocal(index, "number", ir.Number)
	function.Parameters = []int{input.Local}
	l.writeSites = append(l.writeSites, writeSite{node: node})
	site := len(l.writeSites)
	function.Body = []ir.Statement{
		ir.Declare{Local: array.Local, Value: ir.ArrayLiteral{Element: ir.MaybeNumber}},
		ir.ForOf{Iterable: input, Element: ir.Number, Local: item.Local, Body: []ir.Statement{ir.Evaluate{Value: ir.ArrayPush{Array: array, Value: fit(item, ir.MaybeNumber), Element: ir.MaybeNumber, Site: site}}}},
		ir.Return{Value: array},
	}
	l.result.Functions = append(l.result.Functions, function)
	return ir.Call{Function: index, Arguments: []ir.Expression{source}, Returns: ir.Array}
}

// Collection iterators hold their source strongly, though the public iterator type only names its
// yielded elements. Keep this hidden edge visible to the cycle finder just like a closure capture.
type libraryIteratorCapture struct {
	iterator, collection *checker.Type
}

func (f *cycleFinder) libraryIteratorMade(node *ast.Node) {
	if node.Kind != ast.KindCallExpression {
		return
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	var receiver *ast.Node
	if callee.Kind == ast.KindPropertyAccessExpression {
		receiver = callee.AsPropertyAccessExpression().Expression
	} else if callee.Kind == ast.KindElementAccessExpression && f.l.symbolIterator(callee.AsElementAccessExpression().ArgumentExpression) {
		receiver = callee.AsElementAccessExpression().Expression
	} else {
		return
	}
	iterator := f.l.checker.GetTypeAtLocation(node)
	collection := f.l.checker.GetTypeAtLocation(receiver)
	if f.l.isLibraryType(iterator, "MapIterator", "SetIterator", "ArrayIterator", "StringIterator") &&
		(f.l.isLibraryType(collection, "Map", "ReadonlyMap", "Set", "ReadonlySet") || f.l.checker.IsArrayType(collection)) {
		f.libraryIterators = append(f.libraryIterators, libraryIteratorCapture{iterator, collection})
	}
}

func (f *cycleFinder) libraryIteratorCaptures(proven *checker.Type) []cycleNode {
	var held []cycleNode
	for _, capture := range f.libraryIterators {
		if f.l.checker.IsTypeAssignableTo(capture.iterator, proven) {
			held = append(held, cycleNode{proven: capture.collection})
		}
	}
	return held
}

// Iterator next is inherited in JavaScript. Copying our internal closure slot as an own
// enumerable field would let a copy drive the original iterator, unlike Node.
func (l *lowering) libraryIteratorUnsupportedUse(node *ast.Node) error {
	isIterator := func(source *ast.Node) bool {
		return l.libraryIteratorType(l.checker.GetTypeAtLocation(source))
	}
	if node.Kind == ast.KindObjectLiteralExpression {
		for _, property := range node.AsObjectLiteralExpression().Properties.Nodes {
			if property.Kind == ast.KindSpreadAssignment && isIterator(property.AsSpreadAssignment().Expression) {
				return l.notYet(property, "spreading a built-in iterator: next is inherited, not an own enumerable field; use the original iterator or collect its elements into an array")
			}
		}
	}
	if node.Kind == ast.KindCallExpression {
		call := node.AsCallExpression()
		callee := ast.SkipParentheses(call.Expression)
		if callee.Kind == ast.KindPropertyAccessExpression {
			receiver, name := callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
			if l.isLibraryGlobal(receiver, "Object") && len(call.Arguments.Nodes) > 0 && isIterator(call.Arguments.Nodes[0]) {
				switch name {
				case "keys", "values", "entries", "getOwnPropertyNames", "getOwnPropertySymbols", "getOwnPropertyDescriptor", "getOwnPropertyDescriptors", "hasOwn":
					return l.notYet(node, "own-property reflection on a built-in iterator: next is inherited, and its native closure slot is private")
				}
			}
			if isIterator(receiver) && (name == "hasOwnProperty" || name == "propertyIsEnumerable") {
				return l.notYet(node, "own-property reflection on a built-in iterator: next is inherited, and its native closure slot is private")
			}
		}
		if callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "assign" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "Object") {
			for index, source := range call.Arguments.Nodes {
				if index == 0 {
					continue
				}
				if isIterator(source) {
					return l.notYet(source, "Object.assign copying a built-in iterator: next is inherited, not an own enumerable field; use the original iterator")
				}
			}
		}
	}
	return l.libraryIteratorCustomProtocolBoundary(node)
}

func (l *lowering) libraryIteratorType(proven *checker.Type) bool {
	if proven.Flags()&(checker.TypeFlagsUnion|checker.TypeFlagsIntersection) != 0 {
		for _, member := range proven.Types() {
			if l.libraryIteratorType(member) {
				return true
			}
		}
	}
	return l.isLibraryType(proven, "MapIterator", "SetIterator", "ArrayIterator", "StringIterator")
}
