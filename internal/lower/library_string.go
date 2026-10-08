package lower

import (
	"math"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/microsoft/TypeScript/tsc/shim/scanner"
	"github.com/system-inc/adamic/internal/ir"
)

// stringPrototypeMethod recognizes the intrinsic, never a same-named local object.
func (l *lowering) stringPrototypeMethod(node *ast.Node) (string, bool) {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindPropertyAccessExpression {
		return "", false
	}
	prototype := ast.SkipParentheses(node.AsPropertyAccessExpression().Expression)
	if prototype.Kind != ast.KindPropertyAccessExpression || prototype.Name().Text() != "prototype" || !l.isLibraryGlobal(prototype.AsPropertyAccessExpression().Expression, "String") {
		return "", false
	}
	return node.Name().Text(), true
}

// A typeof observation doesn't detach or call a method. An immediate intrinsic .call supplies this
// explicitly. Neither is the unbound-method hole, and every other method read stays refused.
func (l *lowering) stringMethodObservation(node *ast.Node) bool {
	_, intrinsic := l.stringPrototypeMethod(node)
	if !intrinsic {
		access := node.AsPropertyAccessExpression()
		intrinsic = l.isLibraryGlobal(access.Expression, "String")
	}
	if !intrinsic {
		return false
	}
	if len(l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall)) == 0 {
		return false
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent != nil && outer.Parent.Kind == ast.KindTaggedTemplateExpression && outer.Parent.AsTaggedTemplateExpression().Tag == outer {
		return node.Name().Text() == "raw" && l.isLibraryGlobal(node.AsPropertyAccessExpression().Expression, "String")
	}
	if outer.Parent != nil && outer.Parent.Kind == ast.KindTypeOfExpression {
		return true
	}
	if _, prototype := l.stringPrototypeMethod(node); !prototype {
		return false
	}
	return outer.Parent != nil && outer.Parent.Kind == ast.KindPropertyAccessExpression && outer.Parent.Name().Text() == "call" && called(outer.Parent)
}

func (l *lowering) stringTypeOf(node *ast.Node) (ir.Expression, bool) {
	operand := ast.SkipParentheses(node.AsTypeOfExpression().Expression)
	if l.isLibraryGlobal(operand, "String") {
		return ir.StringConstant{Index: l.constant("function")}, true
	}
	if operand.Kind == ast.KindPropertyAccessExpression && l.stringMethodObservation(operand) {
		return ir.StringConstant{Index: l.constant("function")}, true
	}
	return nil, false
}

func (l *lowering) stringConversion(node *ast.Node) (ir.Expression, error) {
	if ast.SkipParentheses(node).Kind == ast.KindNullKeyword {
		return l.stringIdentity(ir.StringConstant{Index: l.constant("null")}), nil
	}
	value, err := l.expression(node)
	if err != nil {
		return nil, err
	}
	return l.stringConversionValue(node, value)
}

func (l *lowering) stringConversionValue(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	if l.checker.GetTypeAtLocation(node).Flags() == checker.TypeFlagsUndefined {
		return ir.StringConstant{Index: l.constant("undefined")}, nil
	}
	if ast.SkipParentheses(node).Kind == ast.KindNullKeyword {
		return ir.StringConstant{Index: l.constant("null")}, nil
	}
	switch value.Type() {
	case ir.Number:
		return ir.NumberToString{Value: value}, nil
	case ir.Boolean:
		return ir.BooleanToString{Value: value}, nil
	case ir.MaybeNumber, ir.MaybeBoolean:
		return ir.MaybeToString{Value: value}, nil
	case ir.String:
		value = l.spelled(node, value)
		// Keep a literal conversion as a call result. typeof must observe its value rather
		// than emit an address-of-literal comparison against NULL, rejected by clang.
		if _, constant := value.(ir.StringConstant); constant {
			return l.stringIdentity(value), nil
		}
		return value, nil
	case ir.Union:
		if l.writable(l.checker.GetTypeAtLocation(node)) || l.dynamicScalarProperty(node) {
			return ir.UnionToString{Value: value}, nil
		}
	}
	if _, missing := value.(ir.Undefined); missing {
		return ir.StringConstant{Index: l.constant("undefined")}, nil
	}
	if value.Type() == ir.Array && !l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
		layer := l.checker.GetTypeAtLocation(node)
		for depth := 0; depth <= 32; depth++ {
			elementType := l.checker.GetElementTypeOfArrayType(l.checker.GetNonNullableType(layer))
			if elementType == nil {
				break
			}
			element, known := l.kept(elementType)
			if !known {
				break
			}
			if element == ir.Array {
				layer = elementType
				continue
			}
			if element == ir.Number || element == ir.MaybeNumber || element == ir.String || element == ir.Boolean {
				return ir.ArrayJoin{Array: value, Separator: ir.StringConstant{Index: l.constant(",")}, Element: element, Depth: depth}, nil
			}
			break
		}
	}
	if value.Type() == ir.Map && !l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
		// Maps and Sets have fixed intrinsic prototypes and refuse expando properties.
		text := "[object Map]"
		if !l.isLibraryType(l.checker.GetTypeAtLocation(node), "Map", "ReadonlyMap", "Set", "ReadonlySet") {
			return nil, l.notYet(node, "String conversion of a mixed Map and Set view")
		}
		if l.isLibraryType(l.checker.GetTypeAtLocation(node), "Set", "ReadonlySet") {
			text = "[object Set]"
		}
		function, _ := l.stringHelper("map_primitive", []ir.Expression{value})
		l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: ir.StringConstant{Index: l.constant(text)}}}
		return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.String}, nil
	}
	if value.Type() == ir.Object && !l.includesUndefined(l.checker.GetTypeAtLocation(node)) {
		return l.stringObjectConversion(node, value)
	}
	return nil, l.notYet(node, "String conversion of an object, array, map or function (ToPrimitive is not lowered)")
}

func (l *lowering) libraryString(node *ast.Node) (ir.Expression, bool, error) {
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	written := node.AsCallExpression().Arguments.Nodes
	if l.isLibraryGlobal(callee, "String") {
		if len(written) == 0 {
			return l.stringIdentity(ir.StringConstant{Index: l.constant("")}), true, nil
		}
		if len(written) != 1 {
			return nil, true, l.notYet(node, "String with spread or extra arguments")
		}
		value, err := l.stringConversion(written[0])
		if err == nil {
			if _, constant := value.(ir.StringConstant); constant {
				value = l.stringIdentity(value)
			}
		}
		return value, true, err
	}
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	receiver, name := callee.AsPropertyAccessExpression().Expression, callee.Name().Text()
	if l.isLibraryGlobal(receiver, "String") && name == "raw" {
		value, err := l.stringRaw(node, written)
		return value, true, err
	}
	if name == "call" {
		if method, intrinsic := l.stringPrototypeMethod(receiver); intrinsic {
			if len(written) == 0 {
				return nil, true, l.notYet(node, "String.prototype."+method+".call without a present receiver")
			}
			proven := l.checker.GetTypeAtLocation(written[0])
			if proven.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) != 0 {
				return l.stringNullPrototypeCall(node, method, written)
			}
			if l.mayBeUndefined(written[0]) {
				return nil, true, l.notYet(node, "String prototype call on a possibly undefined receiver (dynamic TypeError is not lowered)")
			}
			if method == "toString" || method == "valueOf" {
				if proven.Flags()&checker.TypeFlagsStringLike == 0 {
					return nil, true, l.notYet(node, "String.prototype."+method+" on a non-string receiver (requires a String internal slot)")
				}
			}
			if of, _ := l.representation(proven); of == ir.Object || of == ir.Array || of == ir.Map {
				return l.stringObjectPrototypeCall(node, method, written)
			}
			value, err := l.stringConversion(written[0])
			if err != nil {
				return nil, true, err
			}
			return l.libraryStringMethod(node, value, method, written[1:])
		}
	}
	if of, _ := l.representation(l.checker.GetTypeAtLocation(receiver)); of == ir.String {
		switch name {
		case "charAt", "substring", "concat", "toString", "valueOf", "startsWith", "endsWith", "isWellFormed", "toWellFormed":
			value, err := l.expression(receiver)
			if err != nil {
				return nil, true, err
			}
			return l.libraryStringMethod(node, value, name, written)
		case "localeCompare":
			return nil, true, l.notYet(node, "localeCompare (Node uses locale collation, not ordinal UTF-16 order)")
		}
	}
	return nil, false, nil
}

func (l *lowering) libraryStringMethod(node *ast.Node, value ir.Expression, name string, written []*ast.Node) (ir.Expression, bool, error) {
	return l.libraryStringMethodValues(node, value, name, written, nil)
}

func (l *lowering) libraryStringMethodValues(node *ast.Node, value ir.Expression, name string, written []*ast.Node, provided []ir.Expression) (ir.Expression, bool, error) {
	if name == "toString" || name == "valueOf" {
		if len(written) != 0 {
			return nil, true, l.notYet(node, name+" with arguments")
		}
		return value, true, nil
	}
	if name == "concat" {
		values := []ir.Expression{value}
		for index, arg := range written {
			var part ir.Expression
			var err error
			if provided != nil {
				part = provided[index]
			} else {
				part, err = l.expression(arg)
			}
			if err != nil {
				return nil, true, err
			}
			values = append(values, part)
		}
		function, reads := l.stringHelper("concat", values)
		parts := []ir.Expression{reads[0]}
		for index, arg := range written {
			part, err := l.stringConversionValue(arg, reads[index+1])
			if err != nil {
				return nil, true, err
			}
			parts = append(parts, part)
		}
		l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: ir.Concat{Parts: parts}}}
		return ir.Call{Function: function, Arguments: values, Returns: ir.String}, true, nil
	}
	shape, known := stringMethods[name]
	if name == "isWellFormed" || name == "toWellFormed" {
		shape.arguments = nil
		shape.optional = 0
		known = true
	}
	if name == "startsWith" || name == "endsWith" {
		shape.arguments = []ir.Type{ir.String, ir.Number}
		shape.optional = 1
		known = true
	}
	if name == "trim" {
		shape.arguments = nil
		shape.optional = 0
		known = true
	}
	if name == "charCodeAt" || name == "charAt" {
		shape.arguments = []ir.Type{ir.Number}
		shape.optional = 1
		known = true
	}
	if name == "substring" {
		shape.arguments = []ir.Type{ir.Number, ir.Number}
		shape.optional = 2
		known = true
	}
	if !known {
		return nil, true, l.notYet(node, "String.prototype."+name+" (no sound lowering for this method yet)")
	}
	if len(written) > len(shape.arguments) || len(written) < len(shape.arguments)-shape.optional {
		return nil, true, l.notYet(node, name+" with these arguments")
	}
	arguments := []ir.Expression{}
	for index, arg := range written {
		var lowered ir.Expression
		var err error
		if provided != nil {
			lowered = provided[index]
		} else {
			lowered, err = l.expression(arg)
		}
		if err != nil {
			return nil, true, err
		}
		if (name == "startsWith" || name == "endsWith") && index == 1 {
			fallback := ir.Expression(ir.NumberConstant{Value: 0})
			if name == "endsWith" {
				fallback = ir.NumberConstant{Value: math.Inf(1)}
			}
			if _, missing := lowered.(ir.Undefined); missing {
				lowered = fallback
			} else if lowered.Type() == ir.MaybeNumber {
				lowered = ir.Coalesce{Value: lowered, Fallback: fallback, Of: ir.Number}
			}
		}
		if fallback, optional := optionalStrings[name][index]; optional {
			lowered = l.orDefault(arg, lowered, fallback)
		}
		if lowered.Type() != shape.arguments[index] {
			return nil, true, l.notYet(arg, "a "+typeName(lowered.Type())+" argument to "+name)
		}
		arguments = append(arguments, lowered)
	}
	switch name {
	case "startsWith", "endsWith":
		return l.stringAffixMethod(value, name, arguments), true, nil
	case "trim":
		return ir.Trim{Value: value}, true, nil
	case "charCodeAt":
		position := ir.Expression(ir.NumberConstant{Value: 0})
		if len(arguments) > 0 {
			position = arguments[0]
		}
		return ir.CharCodeAt{Value: value, Index: position}, true, nil
	case "charAt", "substring":
		return l.stringIndexMethod(value, name, arguments), true, nil
	case "codePointAt":
		if len(arguments) == 0 {
			arguments = append(arguments, ir.NumberConstant{Value: 0})
		}
	case "normalize":
		if len(arguments) == 0 {
			arguments = append(arguments, ir.StringConstant{Index: l.constant("NFC")})
		}
	case "padStart", "padEnd":
		if len(arguments) == 1 {
			arguments = append(arguments, ir.StringConstant{Index: l.constant(" ")})
		}
	}
	return ir.StringCall{Method: name, Value: value, Arguments: arguments}, true, nil
}

// Helpers use the ordinary IR so both backends and ownership analyses see every evaluation. Each
// operand is passed once, before the helper reads it more than once.
func (l *lowering) stringHelper(name string, values []ir.Expression) (int, []ir.Expression) {
	index := len(l.result.Functions)
	function := ir.Function{Name: "library_string_" + name, Returns: ir.String}
	reads := []ir.Expression{}
	for _, value := range values {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: value.Type(), Function: index})
		function.Parameters = append(function.Parameters, local)
		reads = append(reads, ir.Read{Local: local, Of: value.Type()})
	}
	l.result.Functions = append(l.result.Functions, function)
	return index, reads
}

func stringInteger(value ir.Expression) ir.Expression {
	return ir.Conditional{Condition: ir.NumberCall{Function: "isNaN", Arguments: []ir.Expression{value}}, WhenTrue: ir.NumberConstant{Value: 0}, WhenNot: ir.MathCall{Function: "trunc", Arguments: []ir.Expression{value}}}
}

func (l *lowering) stringIndexMethod(value ir.Expression, name string, arguments []ir.Expression) ir.Expression {
	values := append([]ir.Expression{value}, arguments...)
	function, reads := l.stringHelper(name, values)
	position := ir.Expression(ir.NumberConstant{Value: 0})
	if len(reads) > 1 {
		position = stringInteger(reads[1])
	}
	var result ir.Expression
	if name == "charAt" {
		result = ir.Coalesce{Value: ir.StringIndex{Value: reads[0], Index: position}, Fallback: ir.StringConstant{Index: l.constant("")}, Of: ir.String}
	} else {
		length := ir.StringLength{Value: reads[0]}
		clamp := func(index ir.Expression) ir.Expression {
			return ir.MathCall{Function: "min", Arguments: []ir.Expression{ir.MathCall{Function: "max", Arguments: []ir.Expression{index, ir.NumberConstant{Value: 0}}}, length}}
		}
		start := clamp(position)
		end := ir.Expression(length)
		if len(reads) > 2 {
			end = clamp(stringInteger(reads[2]))
		}
		result = ir.StringCall{Method: "slice", Value: reads[0], Arguments: []ir.Expression{ir.MathCall{Function: "min", Arguments: []ir.Expression{start, end}}, ir.MathCall{Function: "max", Arguments: []ir.Expression{start, end}}}}
	}
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: result}}
	return ir.Call{Function: function, Arguments: values, Returns: ir.String}
}

// String.raw's supported shape has a present raw string array. Index signatures, holes, getters
// remain refused. Substitutions convert lazily; reading raw occurs after all call arguments.
func (l *lowering) stringRaw(node *ast.Node, written []*ast.Node) (ir.Expression, error) {
	if len(written) == 0 {
		return nil, l.notYet(node, "String.raw without a template")
	}
	rawType := l.checker.GetTypeOfPropertyOfType(l.checker.GetTypeAtLocation(written[0]), "raw")
	if rawType == nil || !l.checker.IsArrayType(rawType) || l.includesUndefined(rawType) {
		if value, accepted, err := l.stringRawEmptyLiteral(node, written); accepted {
			return value, err
		}
		return nil, l.notYet(node, "String.raw without a present array of strings in raw")
	}
	element := l.checker.GetElementTypeOfArrayType(rawType)
	if element.Flags()&checker.TypeFlagsNever != 0 {
		if value, accepted, err := l.stringRawEmptyLiteral(node, written); accepted {
			return value, err
		}
		return nil, l.notYet(node, "String.raw with an unrepresented empty raw array")
	}
	if element.Flags()&checker.TypeFlagsStringLike == 0 {
		return nil, l.notYet(node, "String.raw with raw elements that are not strings")
	}
	template, err := l.expression(written[0])
	if err != nil {
		return nil, err
	}
	if template.Type() != ir.Object {
		return nil, l.notYet(node, "String.raw with a template that is not an object")
	}
	values := []ir.Expression{template}
	for _, arg := range written[1:] {
		value, err := l.expression(arg)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	function, reads := l.stringHelper("raw", values)
	local := func(name string, of ir.Type) int {
		index := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: of, Function: function})
		return index
	}
	raw, index, text := local("raw", ir.Array), local("index", ir.Number), local("text", ir.String)
	readRaw := ir.Read{Local: raw, Of: ir.Array}
	readIndex := ir.Read{Local: index, Of: ir.Number}
	readText := ir.Read{Local: text, Of: ir.String}
	next := ir.Binary{Operator: ir.Add, Left: readIndex, Right: ir.NumberConstant{Value: 1}}
	lengthLocal := local("length", ir.Number)
	length := ir.Read{Local: lengthLocal, Of: ir.Number}
	body := []ir.Statement{ir.Assign{Local: text, Value: ir.Concat{Parts: []ir.Expression{readText, ir.Coalesce{Value: ir.ArrayIndex{Array: readRaw, Index: readIndex, Element: ir.String}, Fallback: ir.StringConstant{Index: l.constant("undefined")}, Of: ir.String}}}}}
	for sub, value := range reads[1:] {
		converted, err := l.stringConversionValue(written[sub+1], value)
		if err != nil {
			return nil, err
		}
		body = append(body, ir.If{Condition: ir.Binary{Operator: ir.And, Left: ir.Binary{Operator: ir.Equal, Left: readIndex, Right: ir.NumberConstant{Value: float64(sub)}}, Right: ir.Binary{Operator: ir.Less, Left: next, Right: length}}, Then: []ir.Statement{ir.Assign{Local: text, Value: ir.Concat{Parts: []ir.Expression{readText, converted}}}}})
	}
	l.result.Functions[function].Body = []ir.Statement{
		ir.Declare{Local: raw, Value: ir.Property{Object: reads[0], Name: "raw", Of: ir.Array}},
		ir.Declare{Local: lengthLocal, Value: ir.Length{Array: readRaw}},
		ir.Declare{Local: index, Value: ir.NumberConstant{Value: 0}}, ir.Declare{Local: text, Value: ir.StringConstant{Index: l.constant("")}},
		ir.Loop{Condition: ir.Binary{Operator: ir.Less, Left: readIndex, Right: length}, Body: body, Update: []ir.Statement{ir.Assign{Local: index, Value: next}}}, ir.Return{Value: readText},
	}
	return ir.Call{Function: function, Arguments: values, Returns: ir.String}, nil
}

// The raw intrinsic only reads its template. Its bundled declaration exposes a mutable raw field,
// but passing the template here cannot write through that wider view. This exemption is confined
// to the intrinsic's first argument; storing the same value through that view still refuses.
func (l *lowering) stringReadOnlyArgument(node *ast.Node) bool {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	if len(call.Arguments.Nodes) == 0 || call.Arguments.Nodes[0] != outer {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	return callee.Kind == ast.KindPropertyAccessExpression && callee.Name().Text() == "raw" && l.isLibraryGlobal(callee.AsPropertyAccessExpression().Expression, "String")
}

func (l *lowering) refuseStringWidening(node *ast.Node) error {
	if l.stringReadOnlyArgument(node) {
		return nil
	}
	return l.refuseWidening(node)
}

// String.raw as a tag uses the source's raw template text, including invalid cooked escapes.
// ECMAScript normalizes CR and CRLF to LF even in the raw parts.
func (l *lowering) stringRawTemplate(node *ast.Node) (ir.Expression, error) {
	tagged := node.AsTaggedTemplateExpression()
	tag := ast.SkipParentheses(tagged.Tag)
	if tagged.QuestionDotToken != nil || tag.Kind != ast.KindPropertyAccessExpression || tag.Name().Text() != "raw" || !l.isLibraryGlobal(tag.AsPropertyAccessExpression().Expression, "String") {
		return nil, l.notYet(node, "a tagged template other than the intrinsic String.raw")
	}
	raw := func(text string) ir.Expression {
		text = strings.ReplaceAll(strings.ReplaceAll(text, "\r\n", "\n"), "\r", "\n")
		return ir.StringConstant{Index: l.constant(text)}
	}
	template := tagged.Template
	if template.Kind == ast.KindNoSubstitutionTemplateLiteral {
		source := ast.GetSourceFileOfNode(template)
		start := scanner.GetTokenPosOfNode(template, source, false)
		return raw(source.Text()[start+1 : template.End()-1]), nil
	}
	parts := []ir.Expression{raw(template.AsTemplateExpression().Head.RawText())}
	for _, span := range template.AsTemplateExpression().TemplateSpans.Nodes {
		substitution, err := l.stringConversion(span.AsTemplateSpan().Expression)
		if err != nil {
			return nil, err
		}
		parts = append(parts, substitution, raw(span.AsTemplateSpan().Literal.RawText()))
	}
	return ir.Concat{Parts: parts}, nil
}

// .call evaluates every explicit argument before entering the intrinsic's ToPrimitive step.
func (l *lowering) stringObjectPrototypeCall(node *ast.Node, method string, written []*ast.Node) (ir.Expression, bool, error) {
	values := []ir.Expression{}
	for _, argument := range written {
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		values = append(values, value)
	}
	function, reads := l.stringHelper("prototype_primitive", values)
	converted, err := l.stringConversionValue(written[0], reads[0])
	if err != nil {
		return nil, true, err
	}
	result, _, err := l.libraryStringMethodValues(node, converted, method, written[1:], reads[1:])
	if err != nil {
		return nil, true, err
	}
	l.result.Functions[function].Returns = result.Type()
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: result}}
	return ir.Call{Function: function, Arguments: values, Returns: result.Type()}, true, nil
}

// Positioned affixes follow V8's src/builtins/string-startswith.tq and
// src/builtins/string-endswith.tq: ToIntegerOrInfinity, clamp, then compare units.
func (l *lowering) stringAffixMethod(value ir.Expression, name string, arguments []ir.Expression) ir.Expression {
	values := append([]ir.Expression{value}, arguments...)
	function, reads := l.stringHelper(name, values)
	position := ir.Expression(ir.NumberConstant{Value: 0})
	if name == "endsWith" {
		position = ir.StringLength{Value: reads[0]}
	}
	if len(reads) > 2 {
		position = stringInteger(reads[2])
	}
	position = ir.MathCall{Function: "min", Arguments: []ir.Expression{ir.MathCall{Function: "max", Arguments: []ir.Expression{position, ir.NumberConstant{Value: 0}}}, ir.StringLength{Value: reads[0]}}}
	bounds := []ir.Expression{position}
	if name == "endsWith" {
		bounds = []ir.Expression{ir.NumberConstant{Value: 0}, position}
	}
	result := ir.StringCall{Method: name, Value: ir.StringCall{Method: "slice", Value: reads[0], Arguments: bounds}, Arguments: []ir.Expression{reads[1]}}
	l.result.Functions[function].Returns = ir.Boolean
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: result}}
	return ir.Call{Function: function, Arguments: values, Returns: ir.Boolean}
}

// A nested raw literal with ToLength(length) == 0 has no observable indexed reads.
// It cannot escape to a substitution, so later arguments cannot make it nonempty.
func (l *lowering) stringRawEmptyLiteral(node *ast.Node, written []*ast.Node) (ir.Expression, bool, error) {
	template := ast.SkipParentheses(written[0])
	if template.Kind != ast.KindObjectLiteralExpression {
		return nil, false, nil
	}
	fields := template.AsObjectLiteralExpression().Properties.Nodes
	if len(fields) != 1 || fields[0].Kind != ast.KindPropertyAssignment || fields[0].Name().Text() != "raw" {
		return nil, false, nil
	}
	raw := ast.SkipParentheses(fields[0].AsPropertyAssignment().Initializer)
	if raw.Kind == ast.KindArrayLiteralExpression && len(raw.AsArrayLiteralExpression().Elements.Nodes) == 0 {
		// No template expression has effects here; evaluate all unused arguments exactly once.
	} else if raw.Kind == ast.KindObjectLiteralExpression {
		members := raw.AsObjectLiteralExpression().Properties.Nodes
		if len(members) != 1 || members[0].Kind != ast.KindPropertyAssignment || members[0].Name().Text() != "length" {
			return nil, false, nil
		}
		length, err := l.expression(members[0].AsPropertyAssignment().Initializer)
		if err != nil {
			return nil, true, err
		}
		var number float64
		switch value := length.(type) {
		case ir.NumberConstant:
			number = value.Value
		case ir.Unary:
			constant, known := value.Operand.(ir.NumberConstant)
			if !known || value.Operator != ir.Negate {
				return nil, false, nil
			}
			number = -constant.Value
		default:
			return nil, false, nil
		}
		if !math.IsNaN(number) && number >= 1 {
			return nil, false, nil
		}
	} else {
		return nil, false, nil
	}
	values := []ir.Expression{}
	for _, argument := range written[1:] {
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		values = append(values, value)
	}
	function, _ := l.stringHelper("raw_empty", values)
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: ir.StringConstant{Index: l.constant("")}}}
	return ir.Call{Function: function, Arguments: values, Returns: ir.String}, true, nil
}

// A statically nullish receiver fails before any ToPrimitive or method-specific work.
// The ordinary IR throw carries the error through existing cleanup paths.
func (l *lowering) stringNullPrototypeCall(node *ast.Node, method string, written []*ast.Node) (ir.Expression, bool, error) {
	returns, err := l.typeOf(node)
	if err != nil {
		return nil, true, err
	}
	values := []ir.Expression{}
	for _, argument := range written {
		value, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		// An undefined literal has no slot and no effects.
		if _, missing := value.(ir.Undefined); !missing {
			values = append(values, value)
		}
	}
	function, _ := l.stringHelper("null_receiver", values)
	message := "String.prototype." + method + " called on null or undefined"
	if method == "toString" || method == "valueOf" {
		message = "String.prototype." + method + " requires that 'this' be a String"
	}
	l.result.Functions[function].Returns = returns
	l.result.Functions[function].Body = []ir.Statement{ir.Throw{Value: ir.MakeError{Name: ir.StringConstant{Index: l.constant("TypeError")}, Message: ir.StringConstant{Index: l.constant(message)}}}}
	return ir.Call{Function: function, Arguments: values, Returns: returns}, true, nil
}

func (l *lowering) stringIdentity(value ir.Expression) ir.Expression {
	function, reads := l.stringHelper("identity", []ir.Expression{value})
	l.result.Functions[function].Body = []ir.Statement{ir.Return{Value: reads[0]}}
	return ir.Call{Function: function, Arguments: []ir.Expression{value}, Returns: ir.String}
}
