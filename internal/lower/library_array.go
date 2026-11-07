package lower

import (
	"strconv"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

var libraryArrayMethods = map[string]bool{
	"join": true, "indexOf": true, "includes": true, "lastIndexOf": true,
	"push": true, "unshift": true,
	"with": true, "flatMap": true, "copyWithin": true, "toSpliced": true, "flat": true,
	"findLast": true, "findLastIndex": true, "toReversed": true, "toSorted": true,
}

// libraryArrayMethod keeps this slice's additions out of the shared method dispatch.
func (lowering *lowering) libraryArrayMethod(node, receiver *ast.Node, name string) (ir.Expression, bool, error) {
	return lowering.libraryArrayMethodArguments(node, receiver, name, node.AsCallExpression().Arguments.Nodes)
}

func (lowering *lowering) libraryArrayMethodArguments(
	node, receiver *ast.Node,
	name string,
	written []*ast.Node,
) (ir.Expression, bool, error) {
	if name == "push" && len(written) == 1 && written[0].Kind != ast.KindSpreadElement {
		return lowering.arrayMethodArguments(node, receiver, name, written)
	}
	if name == "push" || name == "unshift" {
		return lowering.arrayInsertArguments(node, receiver, name, written)
	}
	element, elementTypeError := lowering.elementType(receiver)
	if elementTypeError != nil {
		return nil, true, elementTypeError
	}
	if name == "join" && element != ir.Array {
		return lowering.arrayMethod(node, receiver, name)
	}
	array, expressionError := lowering.expression(receiver)
	if expressionError != nil {
		return nil, true, expressionError
	}
	switch name {
	case "join":
		return lowering.libraryArrayJoin(node, receiver, array)
	case "with":
		return lowering.libraryArrayWith(node, array, element)
	case "flatMap":
		return lowering.libraryArrayFlatMap(node, array, element)
	case "toSpliced":
		return lowering.libraryArraySpliced(node, array, element)
	case "flat":
		return lowering.libraryArrayFlat(node, receiver, array)
	case "copyWithin":
		return lowering.libraryArrayCopyWithin(node, array, element)
	case "findLast", "findLastIndex":
		if len(written) != 1 {
			return nil, true, lowering.notYet(node, name+" without one callback")
		}
		signatures := lowering.checker.GetSignaturesOfType(
			lowering.checker.GetTypeAtLocation(written[0]),
			checker.SignatureKindCall,
		)
		if len(signatures) == 1 && len(signatures[0].Parameters()) > 3 {
			return nil, true, lowering.notYet(node, name+" with more than three callback parameters")
		}
		return lowering.arrayVisit(node, array, element, name)
	case "toReversed":
		if len(written) != 0 {
			return nil, true, lowering.notYet(node, "toReversed with arguments")
		}
		// Stage 0 arrays are dense. Copying first preserves identity and reference ownership.
		return ir.ArrayReverse{Array: ir.ArraySlice{Array: array, Element: element}}, true, nil
	case "toSorted":
		if lowering.includesUndefined(lowering.elementTypeOfArray(lowering.checker.GetTypeAtLocation(receiver))) {
			return nil, true, lowering.notYet(node, "toSorted on optional elements")
		}
		copy := ir.ArraySlice{Array: array, Element: element}
		if len(written) == 0 {
			if element != ir.Number && element != ir.Boolean && element != ir.String {
				return nil, true, lowering.notYet(node, "toSorted without a comparator on these elements")
			}
			if lowering.includesUndefined(lowering.elementTypeOfArray(lowering.checker.GetTypeAtLocation(receiver))) {
				return nil, true, lowering.notYet(node, "toSorted without a comparator on optional elements")
			}
			return ir.ArraySort{
				Array:      copy,
				Element:    element,
				Comparator: lowering.libraryArrayComparator(element),
			}, true, nil
		}
		// Reading a function name or making an arrow cannot change the source before the copy.
		// A call producing a comparator could; leave that case refused until operands can be bound.
		callback := ast.SkipParentheses(written[0])
		if len(written) != 1 || (!ast.IsIdentifier(callback) && callback.Kind != ast.KindArrowFunction) {
			return nil, true, lowering.notYet(node, "toSorted with an effectful comparator expression")
		}
		return lowering.arraySort(node, copy, element)
	}
	if len(written) < 1 || len(written) > 2 {
		return nil, true, lowering.notYet(node, name+" with these arguments")
	}
	value, expressionError := lowering.expression(written[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	value = fit(value, element)
	if value.Type() != element {
		return nil, true, lowering.notYet(node, name+" with a value of another type than the elements")
	}
	search := ir.ArraySearch{
		Array:    array,
		Value:    value,
		Element:  element,
		Includes: name == "includes",
		Last:     name == "lastIndexOf",
	}
	if len(written) == 2 {
		from, expressionError := lowering.expression(written[1])
		if expressionError != nil {
			return nil, true, expressionError
		}
		// undefined means 0 when explicitly supplied, including for lastIndexOf.
		if _, absent := from.(ir.Undefined); absent {
			from = ir.NumberConstant{Value: 0}
		}
		if from.Type() != ir.Number {
			return nil, true, lowering.notYet(node, name+" with a starting index that isn't a number")
		}
		search.From = from
	}
	return search, true, nil
}

// A default comparator is ordinary IR, so both backends and every ownership pass see its work.
// ECMAScript orders by UTF-16 strings even for numbers.
func (lowering *lowering) libraryArrayComparator(element ir.Type) int {
	function := len(lowering.result.Functions)
	parameters := []int{len(lowering.result.Locals), len(lowering.result.Locals) + 1}
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "left", Type: element, Function: function},
		ir.Local{Name: "right", Type: element, Function: function},
	)
	text := func(local int) ir.Expression {
		read := ir.Read{Local: local, Of: element}
		if element == ir.Number {
			return ir.NumberToString{Value: read}
		}
		if element == ir.Boolean {
			return ir.BooleanToString{Value: read}
		}
		return read
	}
	left, right := text(parameters[0]), text(parameters[1])
	result := ir.Conditional{
		Condition: ir.Binary{Operator: ir.Less, Left: left, Right: right},
		WhenTrue:  ir.NumberConstant{Value: -1},
		WhenNot: ir.Conditional{
			Condition: ir.Binary{Operator: ir.Greater, Left: left, Right: right},
			WhenTrue:  ir.NumberConstant{Value: 1},
			WhenNot:   ir.NumberConstant{Value: 0},
		},
	}
	lowering.result.Functions = append(lowering.result.Functions, ir.Function{
		Name:       "array_default_compare",
		Parameters: parameters,
		Returns:    ir.Number,
		Body:       []ir.Statement{ir.Return{Value: result}},
	})
	return function
}

// Helpers are ordinary functions in the IR. Their operands are evaluated before any copying,
// including an operand that changes the receiver, as JavaScript requires.
type libraryArrayBuilder struct {
	l          *lowering
	function   int
	parameters []int
	arguments  []ir.Expression
	body       []ir.Statement
}

func (lowering *lowering) libraryArrayBuilder(arguments []ir.Expression) *libraryArrayBuilder {
	b := &libraryArrayBuilder{l: lowering, function: len(lowering.result.Functions), arguments: arguments}
	for _, argument := range arguments {
		b.parameters = append(b.parameters, b.local("argument", argument.Type()))
	}
	return b
}
func (builder *libraryArrayBuilder) local(name string, of ir.Type) int {
	local := len(builder.l.result.Locals)
	builder.l.result.Locals = append(
		builder.l.result.Locals,
		ir.Local{Name: name, Type: of, Function: builder.function},
	)
	return local
}
func (builder *libraryArrayBuilder) read(local int) ir.Expression {
	return ir.Read{Local: local, Of: builder.l.result.Locals[local].Type}
}
func (builder *libraryArrayBuilder) declare(name string, value ir.Expression) int {
	local := builder.local(name, value.Type())
	builder.body = append(builder.body, ir.Declare{Local: local, Value: value})
	return local
}
func (builder *libraryArrayBuilder) finish(name string, result ir.Expression) ir.Expression {
	builder.body = append(builder.body, ir.Return{Value: result})
	builder.l.result.Functions = append(
		builder.l.result.Functions,
		ir.Function{Name: name, Parameters: builder.parameters, Returns: result.Type(), Body: builder.body},
	)
	return ir.Call{Function: builder.function, Arguments: builder.arguments, Returns: result.Type()}
}

func (lowering *lowering) libraryArraySpliced(
	node *ast.Node,
	array ir.Expression,
	element ir.Type,
) (ir.Expression, bool, error) {
	arguments := []ir.Expression{array}
	for index, written := range node.AsCallExpression().Arguments.Nodes {
		value, expressionError := lowering.expression(written)
		if expressionError != nil {
			return nil, true, expressionError
		}
		if index < 2 {
			if _, undefined := value.(ir.Undefined); undefined {
				value = ir.NumberConstant{Value: 0}
			}
			if value.Type() != ir.Number {
				return nil, true, lowering.notYet(node, "toSpliced with nonnumeric bounds")
			}
		} else {
			value = fit(value, element)
			if value.Type() != element {
				return nil, true, lowering.notYet(node, "toSpliced with different element types")
			}
		}
		arguments = append(arguments, value)
	}
	b := lowering.libraryArrayBuilder(arguments)
	copy := b.declare("copy", ir.ArraySlice{Array: b.read(b.parameters[0]), Element: element})
	if len(arguments) > 1 {
		splice := ir.ArraySplice{
			Array:   b.read(copy),
			Start:   b.read(b.parameters[1]),
			Element: element,
			Site:    lowering.writeSite(node),
		}
		if len(arguments) > 2 {
			splice.Count = b.read(b.parameters[2])
			for _, parameter := range b.parameters[3:] {
				splice.Items = append(splice.Items, b.read(parameter))
			}
		}
		b.body = append(b.body, ir.Evaluate{Value: splice})
	}
	return b.finish("array_to_spliced", b.read(copy)), true, nil
}

func (lowering *lowering) libraryArrayFlat(node, receiver *ast.Node, array ir.Expression) (ir.Expression, bool, error) {
	depth := 1
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) > 1 {
		return nil, true, lowering.notYet(node, "flat with multiple depths")
	}
	if len(written) == 1 {
		literal := ast.SkipParentheses(written[0])
		if literal.Kind != ast.KindNumericLiteral {
			return nil, true, lowering.notYet(node, "flat with a nonconstant depth")
		}
		number, parseError := strconv.ParseFloat(literal.Text(), 64)
		if parseError != nil || number > 32 {
			return nil, true, lowering.notYet(node, "flat with this depth")
		}
		depth = int(number)
	}
	// Only homogeneous layers of arrays: a mixed scalar/array union needs a runtime array tag.
	layer := lowering.checker.GetTypeAtLocation(receiver)
	types := []ir.Type{}
	for step := 0; step <= depth; step++ {
		item := lowering.elementTypeOfArray(layer)
		of, known := lowering.kept(item)
		if !known || slotless(of) {
			return nil, true, lowering.notYet(node, "flat on heterogeneous layers")
		}
		types = append(types, of)
		if of == ir.Array && step < depth && lowering.includesUndefined(item) {
			return nil, true, lowering.notYet(node, "flat on optional nested arrays")
		}
		if of != ir.Array || step == depth {
			break
		}
		layer = item
	}
	resultElement, elementTypeError := lowering.elementType(node)
	if elementTypeError != nil {
		return nil, true, elementTypeError
	}
	if resultElement != types[len(types)-1] {
		return nil, true, lowering.notYet(node, "flat whose result element type differs from its depth")
	}
	b := lowering.libraryArrayBuilder([]ir.Expression{array})
	result := b.declare("flattened", ir.ArrayLiteral{Element: resultElement})
	var loop func(ir.Expression, int) ir.Statement
	loop = func(source ir.Expression, level int) ir.Statement {
		local := b.local("element", types[level])
		body := []ir.Statement{ir.Evaluate{Value: ir.ArrayPush{
			Array:   b.read(result),
			Value:   b.read(local),
			Element: resultElement,
			Site:    lowering.writeSite(node),
		}}}
		if level+1 < len(types) {
			body = []ir.Statement{loop(b.read(local), level+1)}
		}
		return ir.ForOf{Iterable: source, Local: local, Element: types[level], Body: body}
	}
	b.body = append(b.body, loop(b.read(b.parameters[0]), 0))
	return b.finish("array_flat", b.read(result)), true, nil
}

func (lowering *lowering) libraryArrayCopyWithin(
	node *ast.Node,
	array ir.Expression,
	element ir.Type,
) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) < 1 || len(written) > 3 {
		return nil, true, lowering.notYet(node, "copyWithin with these arguments")
	}
	arguments := []ir.Expression{array}
	for _, item := range written {
		value, expressionError := lowering.expression(item)
		if expressionError != nil {
			return nil, true, expressionError
		}
		if value.Type() != ir.Number {
			return nil, true, lowering.notYet(node, "copyWithin with a nonnumeric bound")
		}
		arguments = append(arguments, value)
	}
	b := lowering.libraryArrayBuilder(arguments)
	source := b.read(b.parameters[0])
	length := b.declare("length", ir.Length{Array: source})
	zero := ir.NumberConstant{Value: 0}
	add := func(left, right ir.Expression) ir.Expression {
		return ir.Binary{Operator: ir.Add, Left: left, Right: right}
	}
	relative := func(value ir.Expression) ir.Expression {
		integer := ir.Conditional{
			Condition: ir.NumberCall{Function: "isNaN", Arguments: []ir.Expression{value}},
			WhenTrue:  zero,
			WhenNot:   ir.MathCall{Function: "trunc", Arguments: []ir.Expression{value}},
		}
		local := b.declare("integer", integer)
		read := b.read(local)
		return ir.Conditional{
			Condition: ir.Binary{Operator: ir.Less, Left: read, Right: zero},
			WhenTrue:  ir.MathCall{Function: "max", Arguments: []ir.Expression{add(b.read(length), read), zero}},
			WhenNot:   ir.MathCall{Function: "min", Arguments: []ir.Expression{read, b.read(length)}},
		}
	}
	target := b.declare("target", relative(b.read(b.parameters[1])))
	startValue, endValue := ir.Expression(zero), b.read(length)
	if len(arguments) > 2 {
		startValue = b.read(b.parameters[2])
	}
	if len(arguments) > 3 {
		endValue = b.read(b.parameters[3])
	}
	start := b.declare("start", relative(startValue))
	end := b.declare("end", relative(endValue))
	// Snapshot the source range so overlapping writes have memmove semantics, with references held.
	copy := b.declare(
		"range",
		ir.ArraySlice{Array: source, Arguments: []ir.Expression{b.read(start), b.read(end)}, Element: element},
	)
	count := b.declare("count", ir.MathCall{Function: "min", Arguments: []ir.Expression{
		ir.Length{Array: b.read(copy)},
		ir.Binary{Operator: ir.Subtract, Left: b.read(length), Right: b.read(target)},
	}})
	index := b.declare("index", zero)
	value := ir.Expression(ir.ArrayIndex{Array: b.read(copy), Index: b.read(index), Element: element})
	if value.Type().IsMaybe() && value.Type() != element {
		value = ir.Unwrap{Value: value}
	}
	b.body = append(b.body, ir.Loop{
		Condition: ir.Binary{Operator: ir.Less, Left: b.read(index), Right: b.read(count)},
		Body: []ir.Statement{
			ir.SetIndex{
				Array:   source,
				Index:   add(b.read(target), b.read(index)),
				Value:   value,
				Element: element,
				Site:    lowering.writeSite(node),
			},
		},
		Update: []ir.Statement{ir.Assign{Local: index, Value: add(b.read(index), ir.NumberConstant{Value: 1})}},
	})
	return b.finish("array_copy_within", source), true, nil
}

func (lowering *lowering) libraryArrayWith(
	node *ast.Node,
	array ir.Expression,
	element ir.Type,
) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) != 2 {
		return nil, true, lowering.notYet(node, "with without an index and value")
	}
	index, expressionError := lowering.expression(written[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	value, expressionError := lowering.expression(written[1])
	if expressionError != nil {
		return nil, true, expressionError
	}
	value = fit(value, element)
	if index.Type() != ir.Number || value.Type() != element {
		return nil, true, lowering.notYet(node, "with with incompatible arguments")
	}
	b := lowering.libraryArrayBuilder([]ir.Expression{array, index, value})
	source, raw := b.read(b.parameters[0]), b.read(b.parameters[1])
	zero := ir.NumberConstant{Value: 0}
	integer := b.declare("integer", ir.Conditional{
		Condition: ir.NumberCall{Function: "isNaN", Arguments: []ir.Expression{raw}},
		WhenTrue:  zero,
		WhenNot:   ir.MathCall{Function: "trunc", Arguments: []ir.Expression{raw}},
	})
	actual := b.declare("actual", ir.Conditional{
		Condition: ir.Binary{Operator: ir.Less, Left: b.read(integer), Right: zero},
		WhenTrue:  ir.Binary{Operator: ir.Add, Left: ir.Length{Array: source}, Right: b.read(integer)},
		WhenNot:   b.read(integer),
	})
	outside := ir.Binary{
		Operator: ir.Or,
		Left:     ir.Binary{Operator: ir.Less, Left: b.read(actual), Right: zero},
		Right:    ir.Binary{Operator: ir.GreaterOrEqual, Left: b.read(actual), Right: ir.Length{Array: source}},
	}
	errorLocal := b.local("range_error", ir.Object)
	message := ir.Concat{
		Parts: []ir.Expression{
			ir.StringConstant{Index: lowering.constant("Invalid index : ")},
			ir.NumberToString{Value: raw},
		},
	}
	b.body = append(b.body, ir.If{Condition: outside, Then: []ir.Statement{
		ir.Declare{Local: errorLocal, Value: ir.MakeError{Message: message}},
		ir.SetProperty{
			Object: b.read(errorLocal),
			Name:   "name",
			Value:  ir.StringConstant{Index: lowering.constant("RangeError")},
			Site:   lowering.libraryArraySyntheticWriteSite(node),
		},
		ir.Throw{Value: b.read(errorLocal)},
	}})
	copy := b.declare("copy", ir.ArraySlice{Array: source, Element: element})
	b.body = append(b.body, ir.SetIndex{
		Array:   b.read(copy),
		Index:   b.read(actual),
		Value:   b.read(b.parameters[2]),
		Element: element,
		Site:    lowering.writeSite(node),
	})
	return b.finish("array_with", b.read(copy)), true, nil
}

func (lowering *lowering) libraryArrayFlatMap(
	node *ast.Node,
	array ir.Expression,
	element ir.Type,
) (ir.Expression, bool, error) {
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) != 1 {
		return nil, true, lowering.notYet(node, "flatMap with other than one callback")
	}
	callback, expressionError := lowering.expression(written[0])
	if expressionError != nil {
		return nil, true, expressionError
	}
	signatures := lowering.checker.GetSignaturesOfType(
		lowering.checker.GetTypeAtLocation(written[0]),
		checker.SignatureKindCall,
	)
	if callback.Type() != ir.Closure || len(signatures) != 1 || len(signatures[0].Parameters()) > 3 {
		return nil, true, lowering.notYet(node, "flatMap with this callback")
	}
	returned := lowering.checker.GetReturnTypeOfSignature(signatures[0])
	resultType, known := lowering.representation(returned)
	output, elementTypeError := lowering.elementType(node)
	if elementTypeError != nil {
		return nil, true, elementTypeError
	}
	if !known || (resultType != ir.Array && resultType != output) {
		return nil, true, lowering.notYet(node, "flatMap with a mixed scalar and array callback result")
	}
	if resultType == ir.Array {
		if lowering.includesUndefined(returned) {
			return nil, true, lowering.notYet(node, "flatMap with optional array results")
		}
		inner, known := lowering.kept(lowering.elementTypeOfArray(returned))
		if !known || inner != output {
			return nil, true, lowering.notYet(node, "flatMap with this callback element type")
		}
	}
	b := lowering.libraryArrayBuilder([]ir.Expression{array, callback})
	source := b.read(b.parameters[0])
	result := b.declare("mapped", ir.ArrayLiteral{Element: output})
	count := b.declare("count", ir.Length{Array: source})
	index := b.declare("index", ir.NumberConstant{Value: 0})
	input := ir.Expression(ir.ArrayIndex{Array: source, Index: b.read(index), Element: element})
	if input.Type().IsMaybe() && input.Type() != element {
		input = ir.Unwrap{Value: input}
	}
	mapped := b.local("callback_result", resultType)
	callArguments := []ir.Expression{input, b.read(index), source}
	for index, parameter := range signatures[0].Parameters() {
		takes, known := lowering.representation(lowering.checker.GetTypeOfSymbol(parameter))
		if !known || slotless(takes) {
			return nil, true, lowering.notYet(node, "flatMap with this callback parameter type")
		}
		callArguments[index] = fit(callArguments[index], takes)
	}
	call := ir.CallClosure{Closure: b.read(b.parameters[1]), Arguments: callArguments, Returns: resultType}
	body := []ir.Statement{ir.Declare{Local: mapped, Value: call}}
	if resultType == ir.Array {
		item := b.local("element", output)
		body = append(
			body,
			ir.ForOf{Iterable: b.read(mapped), Element: output, Local: item, Body: []ir.Statement{ir.Evaluate{
				Value: ir.ArrayPush{
					Array:   b.read(result),
					Value:   b.read(item),
					Element: output,
					Site:    lowering.writeSite(node),
				},
			}}},
		)
	} else {
		body = append(body, ir.Evaluate{
			Value: ir.ArrayPush{
				Array:   b.read(result),
				Value:   b.read(mapped),
				Element: output,
				Site:    lowering.writeSite(node),
			},
		})
	}
	// An index the callback removed is skipped, as flatMap's HasProperty requires.
	b.body = append(b.body, ir.Loop{
		Condition: ir.Binary{Operator: ir.Less, Left: b.read(index), Right: b.read(count)},
		Body: []ir.Statement{ir.If{
			Condition: ir.Binary{Operator: ir.Less, Left: b.read(index), Right: ir.Length{Array: source}},
			Then:      body,
		}},
		Update: []ir.Statement{ir.Assign{
			Local: index,
			Value: ir.Binary{Operator: ir.Add, Left: b.read(index), Right: ir.NumberConstant{Value: 1}},
		}},
	})
	return b.finish("array_flat_map", b.read(result)), true, nil
}

var libraryArrayLengths = map[string]float64{
	"at": 1, "concat": 1, "copyWithin": 2, "entries": 0, "every": 1, "fill": 1,
	"filter": 1, "find": 1, "findIndex": 1, "findLast": 1, "findLastIndex": 1,
	"flat": 0, "flatMap": 1, "forEach": 1, "includes": 1, "indexOf": 1, "join": 1,
	"keys": 0, "lastIndexOf": 1, "map": 1, "pop": 0, "push": 1, "reduce": 1,
	"reduceRight": 1, "reverse": 0, "shift": 0, "slice": 2, "some": 1, "sort": 1,
	"splice": 2, "toLocaleString": 0, "toReversed": 0, "toSorted": 1, "toSpliced": 2,
	"toString": 0, "unshift": 1, "values": 0, "with": 2,
}

// Observing the standard function's metadata never calls it with a lost receiver. Recognition
// includes the library global's identity; an object's method or a shadowed Array is never folded.
func (lowering *lowering) libraryArrayMethodName(node *ast.Node) string {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression {
		return ""
	}
	access := node.AsPropertyAccessExpression()
	if access.QuestionDotToken != nil {
		return ""
	}
	prototype := ast.SkipParentheses(access.Expression)
	if prototype.Kind != ast.KindPropertyAccessExpression ||
		prototype.Name().Text() != "prototype" ||
		prototype.AsPropertyAccessExpression().QuestionDotToken != nil ||
		!lowering.isLibraryGlobal(prototype.AsPropertyAccessExpression().Expression, "Array") {
		return ""
	}
	name := access.Name().Text()
	if _, known := libraryArrayLengths[name]; !known {
		return ""
	}
	return name
}

func (lowering *lowering) libraryArrayObservation(node *ast.Node) (ir.Expression, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind == ast.KindTypeOfExpression {
		if name := lowering.libraryArrayMethodName(node.AsTypeOfExpression().Expression); name != "" {
			return ir.StringConstant{Index: lowering.constant("function")}, true
		}
	}
	if node.Kind == ast.KindPropertyAccessExpression && node.AsPropertyAccessExpression().QuestionDotToken == nil {
		if name := lowering.libraryArrayMethodName(node.AsPropertyAccessExpression().Expression); name != "" {
			switch node.Name().Text() {
			case "length":
				return ir.NumberConstant{Value: libraryArrayLengths[name]}, true
			case "name":
				return ir.StringConstant{Index: lowering.constant(name)}, true
			}
		}
	}
	return nil, false
}

func (lowering *lowering) libraryArrayObservedMethod(node *ast.Node) bool {
	if lowering.libraryArrayMethodName(node) == "" {
		return false
	}
	for node.Parent != nil && node.Parent.Kind == ast.KindParenthesizedExpression {
		node = node.Parent
	}
	parent := node.Parent
	if parent == nil {
		return false
	}
	if parent.Kind == ast.KindTypeOfExpression {
		return true
	}
	return parent.Kind == ast.KindPropertyAccessExpression &&
		parent.AsPropertyAccessExpression().Expression == node &&
		parent.AsPropertyAccessExpression().QuestionDotToken == nil &&
		(parent.Name().Text() == "length" || parent.Name().Text() == "name")
}

// Direct for...of over a standard array iterator keeps the source alive and reads its length
// each time. Materializing its elements up front would lose writes and pushes in the body.
func (lowering *lowering) libraryArrayForOf(node *ast.Node) ([]ir.Statement, bool, error) {
	statement := node.AsForInOrOfStatement()
	call := ast.SkipParentheses(statement.Expression)
	if call.Kind != ast.KindCallExpression || len(call.AsCallExpression().Arguments.Nodes) != 0 {
		return nil, false, nil
	}
	callee := ast.SkipParentheses(call.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	access := callee.AsPropertyAccessExpression()
	name := access.Name().Text()
	if name != "entries" && name != "keys" && name != "values" {
		return nil, false, nil
	}
	if receiver, known := lowering.representation(
		lowering.checker.GetTypeAtLocation(access.Expression),
	); !known || receiver != ir.Array {
		return nil, false, nil
	}
	if optionalCallError := lowering.optionalCall(call); optionalCallError != nil {
		return nil, true, optionalCallError
	}
	if statement.AwaitModifier != nil {
		return nil, true, lowering.notYet(node, "for await")
	}
	initializer := statement.Initializer
	if initializer.Kind != ast.KindVariableDeclarationList || initializer.Flags&ast.NodeFlagsBlockScoped == 0 {
		return nil, true, lowering.notYet(node, "an array iterator loop without const or let")
	}
	declarations := initializer.AsVariableDeclarationList().Declarations.Nodes
	if len(declarations) != 1 {
		return nil, true, lowering.notYet(node, "an array iterator loop with multiple declarations")
	}
	binding := declarations[0].Name()
	array, expressionError := lowering.expression(access.Expression)
	if expressionError != nil {
		return nil, true, expressionError
	}
	element, elementTypeError := lowering.elementType(access.Expression)
	if elementTypeError != nil {
		return nil, true, elementTypeError
	}
	local := func(name string, of ir.Type) int {
		index := len(lowering.result.Locals)
		lowering.result.Locals = append(
			lowering.result.Locals,
			ir.Local{Name: name, Type: of, Function: lowering.functionIndex},
		)
		return index
	}
	source, index := local("array_iterator_source", ir.Array), local("array_iterator_index", ir.Number)
	read := func(local int) ir.Expression { return ir.Read{Local: local, Of: lowering.result.Locals[local].Type} }
	value := ir.Expression(ir.ArrayIndex{Array: read(source), Index: read(index), Element: element})
	if value.Type().IsMaybe() && value.Type() != element {
		value = ir.Unwrap{Value: value}
	}
	body := []ir.Statement{}
	if ast.IsIdentifier(binding) {
		bound, declareLocalError := lowering.declareLocal(binding)
		if declareLocalError != nil {
			return nil, true, declareLocalError
		}
		given := value
		if name == "keys" {
			given = read(index)
		}
		if name == "entries" {
			given = ir.ObjectLiteral{
				Tuple:  true,
				Fields: []ir.Field{{Name: "0", Value: read(index)}, {Name: "1", Value: value}},
			}
		}
		body = append(body, ir.Declare{Local: bound, Value: given})
	} else if name == "entries" && binding.Kind == ast.KindArrayBindingPattern {
		entries := binding.AsBindingPattern().Elements.Nodes
		if len(entries) > 2 {
			return nil, true, lowering.notYet(node, "destructuring more than two array entry fields")
		}
		for position, bound := range entries {
			if skipped(bound) {
				continue
			}
			item := bound.AsBindingElement()
			if !ast.IsIdentifier(bound.Name()) || item.Initializer != nil || item.DotDotDotToken != nil {
				return nil, true, lowering.notYet(node, "an array entry binding with a default or rest")
			}
			variable, declareLocalError := lowering.declareLocal(bound.Name())
			if declareLocalError != nil {
				return nil, true, declareLocalError
			}
			given := value
			if position == 0 {
				given = read(index)
			}
			body = append(body, ir.Declare{Local: variable, Value: given})
		}
	} else {
		return nil, true, lowering.notYet(node, "destructuring this array iterator")
	}
	inner, statementError := lowering.statement(statement.Statement)
	if statementError != nil {
		return nil, true, statementError
	}
	body = append(body, inner...)
	loop := ir.Loop{
		Condition: ir.Binary{Operator: ir.Less, Left: read(index), Right: ir.Length{Array: read(source)}},
		Body:      body,
		Update: []ir.Statement{ir.Assign{
			Local: index,
			Value: ir.Binary{Operator: ir.Add, Left: read(index), Right: ir.NumberConstant{Value: 1}},
		}},
	}
	return []ir.Statement{ir.Block{Body: []ir.Statement{
		ir.Declare{Local: source, Value: array},
		ir.Declare{Local: index, Value: ir.NumberConstant{Value: 0}},
		loop,
	}}}, true, nil
}

// A generated Error has no source expression whose checker type names its holder. A nil holder
// keeps the cycle finder conservative if the write is not proven, rather than borrowing an array type.
func (lowering *lowering) libraryArraySyntheticWriteSite(node *ast.Node) int {
	lowering.writeSites = append(lowering.writeSites, writeSite{node: node})
	return len(lowering.writeSites)
}

// Nested homogeneous arrays stringify with a comma at every inner layer, regardless of the
// separator supplied for the outer join. Optional inner arrays stringify as an empty element.
func (lowering *lowering) libraryArrayJoin(node, receiver *ast.Node, array ir.Expression) (ir.Expression, bool, error) {
	layer := lowering.checker.GetTypeAtLocation(receiver)
	depth := 0
	var element ir.Type
	for {
		item := lowering.elementTypeOfArray(lowering.checker.GetNonNullableType(layer))
		of, known := lowering.kept(item)
		if !known {
			return nil, true, lowering.notYet(node, "join on heterogeneous nested arrays")
		}
		if of != ir.Array {
			element = of
			break
		}
		depth++
		if depth > 32 {
			return nil, true, lowering.notYet(node, "join on recursively nested arrays")
		}
		layer = item
	}
	if element != ir.Number && element != ir.MaybeNumber && element != ir.Boolean && element != ir.String {
		return nil, true, lowering.notYet(node, "join on nested arrays of objects or functions")
	}
	written := node.AsCallExpression().Arguments.Nodes
	separator := ir.Expression(ir.StringConstant{Index: lowering.constant(",")})
	if len(written) > 1 {
		return nil, true, lowering.notYet(node, "join with more than one separator")
	}
	if len(written) == 1 {
		value, expressionError := lowering.expression(written[0])
		if expressionError != nil {
			return nil, true, expressionError
		}
		separator = lowering.orDefault(written[0], value, ",")
		if separator.Type() != ir.String {
			return nil, true, lowering.notYet(node, "join with a nonstring separator")
		}
	}
	return ir.ArrayJoin{Array: array, Separator: separator, Element: element, Depth: depth}, true, nil
}
