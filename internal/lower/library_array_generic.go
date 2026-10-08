package lower

import (
	"math"
	"sort"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Only immediate calls to these intrinsic functions supply an explicit receiver.
// A detached method or a shadowed Array is still an ordinary, refused method read.
func (l *lowering) libraryArrayExplicitSearch(node *ast.Node) string {
	name := l.libraryArrayMethodName(node)
	if name != "indexOf" && name != "lastIndexOf" {
		return ""
	}
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	if outer.Parent != nil && outer.Parent.Kind == ast.KindPropertyAccessExpression && outer.Parent.Name().Text() == "call" && outer.Parent.AsPropertyAccessExpression().QuestionDotToken == nil && called(outer.Parent) {
		return name
	}
	return ""
}

func (l *lowering) libraryArrayGenericCall(node *ast.Node) (ir.Expression, bool, error) {
	if value, known, err := l.libraryArrayReceiverCall(node); known {
		return value, true, err
	}
	if value, known, err := l.libraryArrayFromIterable(node); known {
		return value, true, err
	}
	if value, known, err := l.libraryArrayConstruct(node); known {
		return value, true, err
	}
	if value, known, err := l.libraryArrayIsArray(node); known {
		return value, true, err
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "call" {
		return nil, false, nil
	}
	method := ast.SkipParentheses(callee.AsPropertyAccessExpression().Expression)
	name := l.libraryArrayExplicitSearch(method)
	if name == "" {
		return nil, false, nil
	}
	written := node.AsCallExpression().Arguments.Nodes
	if len(written) == 0 || len(written) > 3 {
		return nil, true, l.notYet(node, name+".call with missing receiver or extra arguments")
	}
	proven := l.checker.GetTypeAtLocation(written[0])
	if l.includesUndefined(proven) || l.includesNull(proven) {
		return nil, true, l.notYet(node, name+".call on null or undefined (ToObject requires a catchable TypeError)")
	}
	receiver, err := l.expression(written[0])
	if err != nil {
		return nil, true, err
	}
	needle := ir.Expression(ir.Undefined{})
	if len(written) > 1 {
		needle, err = l.expression(written[1])
		if err != nil {
			return nil, true, err
		}
	}
	arguments := []ir.Expression{receiver, needle}
	if len(written) > 2 {
		from, err := l.expression(written[2])
		if err != nil {
			return nil, true, err
		}
		arguments = append(arguments, from)
	}
	if receiver.Type() == ir.Array {
		if len(written) > 1 {
			searchType := l.checker.GetTypeAtLocation(written[1])
			elementType := l.checker.GetElementTypeOfArrayType(proven)
			if l.includesNull(searchType) || l.includesUndefined(searchType) && (elementType == nil || l.includesNull(elementType)) {
				return nil, true, l.notYet(node, "a generic array search requiring distinct null and undefined reference slots")
			}
		}
		element, err := l.elementType(written[0])
		if err != nil {
			return nil, true, err
		}
		if err := l.libraryArraySearchSlots(node, written[0], element); err != nil {
			return nil, true, err
		}
		value := fit(needle, element)
		if value.Type() != element {
			return nil, true, l.notYet(node, name+".call with a search value of another element representation")
		}
		search := ir.ArraySearch{Array: receiver, Value: value, Element: element, Last: name == "lastIndexOf"}
		if len(arguments) > 2 {
			search.From, err = l.libraryArrayPrimitiveNumber(node, arguments[2], l.checker.GetTypeAtLocation(written[2]))
			if err != nil {
				return nil, true, err
			}
		}
		return search, true, nil
	}
	if needle.Type() == ir.Union {
		return nil, true, l.notYet(node, "a generic array search with mixed representations")
	}
	if len(written) > 1 {
		searchType := l.checker.GetTypeAtLocation(written[1])
		if searchType.Flags()&checker.TypeFlagsUnion != 0 && (l.includesUndefined(searchType) || l.includesNull(searchType)) {
			return nil, true, l.notYet(node, "a generic array search with a nullable or undefined union search value")
		}
		if needle.Type() == ir.Object && searchType.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined) == 0 && !l.libraryArrayExactShape(written[1]) {
			return nil, true, l.notYet(node, "a generic array search with an object search view that can hide another representation")
		}
	}
	if receiver.Type() != ir.String && receiver.Type() != ir.Number && receiver.Type() != ir.Boolean && receiver.Type() != ir.Object {
		return nil, true, l.notYet(node, name+".call on an unrepresented array-like receiver")
	}
	if receiver.Type() == ir.Object && !l.libraryArrayExactShape(written[0]) {
		return nil, true, l.notYet(node, name+".call on a shape not proven by a plain literal or its unreassigned binding")
	}
	b := l.libraryArrayBuilder(arguments)
	source, sought := b.read(b.parameters[0]), b.read(b.parameters[1])
	if len(written) > 1 {
		// Pure null/undefined types can be normalized after binding the original
		// argument. This preserves effects without conflating their null pointers.
		searchType := l.checker.GetTypeAtLocation(written[1])
		if searchType.Flags()&checker.TypeFlagsUndefined != 0 {
			sought = ir.Undefined{}
		} else if searchType.Flags()&checker.TypeFlagsNull != 0 {
			sought = ir.Null{}
		}
	}
	if _, missing := needle.(ir.Undefined); missing {
		sought = needle
	}
	if _, null := needle.(ir.Null); null {
		sought = needle
	}
	zero, one := ir.NumberConstant{Value: 0}, ir.NumberConstant{Value: 1}
	lengthValue := ir.Expression(zero)
	if receiver.Type() == ir.String {
		lengthValue = ir.StringLength{Value: source}
	} else if receiver.Type() == ir.Object {
		if field := l.checker.GetPropertyOfType(proven, "length"); field != nil {
			typeOf := l.checker.GetTypeOfSymbol(field)
			raw, err := l.libraryArrayFieldValue(node, source, field)
			if err != nil {
				return nil, true, err
			}
			lengthValue, err = l.libraryArrayPrimitiveNumber(node, raw, typeOf)
			if err != nil {
				return nil, true, err
			}
		}
	}
	// GetLengthProperty / ToLength, as V8's Runtime_ArrayIndexOf and
	// ArrayPrototypeLastIndexOf do, before converting fromIndex.
	length := b.declare("length", ir.MathCall{Function: "min", Arguments: []ir.Expression{ir.NumberConstant{Value: 9007199254740991}, ir.MathCall{Function: "max", Arguments: []ir.Expression{zero, libraryArrayInteger(lengthValue)}}}})
	b.body = append(b.body, ir.If{Condition: ir.Binary{Operator: ir.Equal, Left: b.read(length), Right: zero}, Then: []ir.Statement{ir.Return{Value: ir.NumberConstant{Value: -1}}}})
	last := name == "lastIndexOf"
	fromValue := ir.Expression(zero)
	if last {
		fromValue = ir.Binary{Operator: ir.Subtract, Left: b.read(length), Right: one}
	}
	if len(arguments) > 2 {
		fromValue, err = l.libraryArrayPrimitiveNumber(node, b.read(b.parameters[2]), l.checker.GetTypeAtLocation(written[2]))
		if err != nil {
			return nil, true, err
		}
	}
	integer := b.declare("integer", libraryArrayInteger(fromValue))
	nonnegative := ir.Expression(ir.MathCall{Function: "max", Arguments: []ir.Expression{ir.Binary{Operator: ir.Add, Left: b.read(length), Right: b.read(integer)}, zero}})
	positive := b.read(integer)
	if last {
		nonnegative = ir.Binary{Operator: ir.Add, Left: b.read(length), Right: b.read(integer)}
		positive = ir.MathCall{Function: "min", Arguments: []ir.Expression{b.read(integer), ir.Binary{Operator: ir.Subtract, Left: b.read(length), Right: one}}}
	}
	from := b.declare("from", ir.Conditional{Condition: ir.Binary{Operator: ir.Less, Left: b.read(integer), Right: zero}, WhenTrue: nonnegative, WhenNot: positive})
	if receiver.Type() == ir.String {
		index := b.declare("index", b.read(from))
		condition := ir.Expression(ir.Binary{Operator: ir.Less, Left: b.read(index), Right: b.read(length)})
		operator := ir.Add
		if last {
			condition = ir.Binary{Operator: ir.GreaterOrEqual, Left: b.read(index), Right: zero}
			operator = ir.Subtract
		}
		character := ir.StringIndex{Value: source, Index: b.read(index)}
		b.body = append(b.body, ir.Loop{Condition: condition, Body: []ir.Statement{ir.If{Condition: libraryArrayStrictEqual(character, sought), Then: []ir.Statement{ir.Return{Value: b.read(index)}}}}, Update: []ir.Statement{ir.Assign{Local: index, Value: ir.Binary{Operator: operator, Left: b.read(index), Right: one}}}})
	} else if receiver.Type() == ir.Object {
		// Fixed own fields have no accessors or prototype elements. Visiting only
		// present numeric keys is V8's HasProperty loop with absent iterations elided.
		// This also handles huge lengths without iterating billions of absent keys.
		fields := append([]*ast.Symbol{}, l.checker.GetPropertiesOfType(proven)...)
		sort.Slice(fields, func(i, j int) bool {
			left, _ := strconv.ParseUint(fields[i].Name, 10, 53)
			right, _ := strconv.ParseUint(fields[j].Name, 10, 53)
			if last {
				return left > right
			}
			return left < right
		})
		for _, field := range fields {
			index, err := strconv.ParseUint(field.Name, 10, 53)
			if err != nil || strconv.FormatUint(index, 10) != field.Name || index >= 9007199254740991 {
				continue
			}
			if field.Flags&ast.SymbolFlagsOptional != 0 {
				return nil, true, l.notYet(node, name+".call on optional numeric fields (presence is not represented)")
			}
			value, err := l.libraryArrayFieldValue(node, source, field)
			if err != nil {
				return nil, true, err
			}
			key := ir.NumberConstant{Value: float64(index)}
			direction := ir.GreaterOrEqual
			if last {
				direction = ir.LessOrEqual
			}
			eligible := ir.Binary{Operator: ir.And, Left: ir.Binary{Operator: direction, Left: key, Right: b.read(from)}, Right: ir.Binary{Operator: ir.Less, Left: key, Right: b.read(length)}}
			b.body = append(b.body, ir.If{Condition: eligible, Then: []ir.Statement{ir.If{Condition: libraryArrayStrictEqual(value, sought), Then: []ir.Statement{ir.Return{Value: key}}}}})
		}
	}
	return b.finish("array_generic_"+name, ir.NumberConstant{Value: -1}), true, nil
}

func libraryArrayInteger(value ir.Expression) ir.Expression {
	return ir.Conditional{Condition: ir.NumberCall{Function: "isNaN", Arguments: []ir.Expression{value}}, WhenTrue: ir.NumberConstant{}, WhenNot: ir.MathCall{Function: "trunc", Arguments: []ir.Expression{value}}}
}

func (l *lowering) libraryArrayPrimitiveNumber(node *ast.Node, value ir.Expression, proven *checker.Type) (ir.Expression, error) {
	if proven.Flags()&checker.TypeFlagsNull != 0 {
		return ir.NumberConstant{}, nil
	}
	if proven.Flags()&checker.TypeFlagsUndefined != 0 {
		return ir.NumberConstant{Value: math.NaN()}, nil
	}
	switch value.Type() {
	case ir.Number:
		return value, nil
	case ir.Boolean, ir.String, ir.MaybeNumber, ir.MaybeBoolean:
		return ir.NumberCall{Function: "convert", Arguments: []ir.Expression{value}}, nil
	}
	return nil, l.notYet(node, "an array-like bound requiring object ToPrimitive or mixed conversion")
}

func libraryArrayStrictEqual(left, right ir.Expression) ir.Expression {
	// An absent numeric key is never supplied here. A present undefined field
	// is different from that absence, and is compared by its own representation.
	_, leftNull := left.(ir.Null)
	_, rightNull := right.(ir.Null)
	_, leftMissing := left.(ir.Undefined)
	_, rightMissing := right.(ir.Undefined)
	if leftNull && rightMissing || rightNull && leftMissing {
		return ir.BooleanConstant{Value: false}
	}
	if left.Type() != right.Type() {
		if left.Type().IsMaybe() && left.Type().Present() == right.Type() {
			right = fit(right, left.Type())
		} else if right.Type().IsMaybe() && right.Type().Present() == left.Type() {
			left = fit(left, right.Type())
		} else if left.Type() == ir.Union || right.Type() == ir.Union {
			left, right = fit(left, ir.Union), fit(right, ir.Union)
		} else {
			return ir.BooleanConstant{Value: false}
		}
	}
	return ir.Binary{Operator: ir.Equal, Left: left, Right: right}
}

// A let introduced by the test262 adapter can also prove a complete shape,
// provided no assignment anywhere changes its binding. Fields may still mutate:
// their representations and presence stay fixed, and searches read them at call time.
func (l *lowering) libraryArrayExactShape(node *ast.Node) bool {
	node = ast.SkipParentheses(node)
	if ast.IsIdentifier(node) {
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.AsVariableDeclaration().Type != nil || declaration.AsVariableDeclaration().Initializer == nil {
			return false
		}
		changed := false
		var visit ast.Visitor
		visit = func(part *ast.Node) bool {
			var target *ast.Node
			if part.Kind == ast.KindBinaryExpression && part.AsBinaryExpression().OperatorToken.Kind == ast.KindEqualsToken {
				target = part.AsBinaryExpression().Left
			}
			if part.Kind == ast.KindForOfStatement {
				target = part.AsForInOrOfStatement().Initializer
			}
			if target != nil {
				var check ast.Visitor
				check = func(inner *ast.Node) bool {
					if inner.Kind == ast.KindPropertyAccessExpression || inner.Kind == ast.KindElementAccessExpression {
						return false // A field write does not reassign the binding.
					}
					if ast.IsIdentifier(inner) && l.symbol(inner) == symbol {
						changed = true
					}
					return inner.ForEachChild(check)
				}
				check(target)
			}
			return part.ForEachChild(visit)
		}
		modules, err := l.moduleOrder(l.program.Files()[0])
		if err != nil {
			return false
		}
		for _, module := range modules {
			module.AsNode().ForEachChild(visit)
		}
		if changed {
			return false
		}
		node = ast.SkipParentheses(declaration.AsVariableDeclaration().Initializer)
	}
	if node.Kind != ast.KindObjectLiteralExpression {
		return false
	}
	for _, field := range node.AsObjectLiteralExpression().Properties.Nodes {
		if field.Kind != ast.KindPropertyAssignment && field.Kind != ast.KindShorthandPropertyAssignment {
			return false
		}
		name := field.Name()
		if name.Kind != ast.KindIdentifier && name.Kind != ast.KindStringLiteral && name.Kind != ast.KindNumericLiteral {
			return false
		}
		if strings.ContainsRune(name.Text(), 0) || name.Text() == "__proto__" {
			return false
		}
	}
	return true
}

// Numeric literal array-like keys use the checker's canonical property name,
// including hexadecimal and exponent spellings. Computed keys remain refused.
func (l *lowering) libraryArrayLikeFieldName(name *ast.Node) (string, bool) {
	if ast.IsIdentifier(name) || name.Kind == ast.KindStringLiteral {
		return name.Text(), true
	}
	if name.Kind == ast.KindNumericLiteral {
		if symbol := l.checker.GetSymbolAtLocation(name); symbol != nil {
			return symbol.Name, true
		}
	}
	return "", false
}

func (l *lowering) libraryArrayFieldValue(node *ast.Node, object ir.Expression, field *ast.Symbol) (ir.Expression, error) {
	proven := l.checker.GetTypeOfSymbol(field)
	if proven.Flags()&checker.TypeFlagsUnion != 0 && (l.includesUndefined(proven) || l.includesNull(proven)) {
		return nil, l.notYet(node, "an array-like field with nullable or undefined union values")
	}
	if proven.Flags()&checker.TypeFlagsUndefined != 0 {
		return ir.Undefined{}, nil
	}
	if proven.Flags()&checker.TypeFlagsNull != 0 {
		return ir.Null{}, nil
	}
	// An empty object type admits primitives. Such a field can later receive a
	// number while its native slot remains a pointer, so its value is not proven.
	if of, known := l.representation(proven); known && of == ir.Object && len(l.checker.GetPropertiesOfType(proven)) == 0 {
		return nil, l.notYet(node, "an array-like object field that can hide primitive slots")
	}
	if of, known := l.representation(proven); known && of == ir.Object {
		var initializer *ast.Node
		if len(field.Declarations) == 1 {
			declaration := field.Declarations[0]
			if declaration.Kind == ast.KindPropertyAssignment {
				initializer = declaration.AsPropertyAssignment().Initializer
			} else if declaration.Kind == ast.KindShorthandPropertyAssignment {
				initializer = declaration.Name()
			}
		}
		if initializer == nil || !l.libraryArrayExactShape(initializer) {
			return nil, l.notYet(node, "an array-like object field with an unproven runtime representation")
		}
	}
	of, known := l.kept(proven)
	if !known || slotless(of) || of == ir.Union {
		return nil, l.notYet(node, "an array-like field with an unrepresented or mixed value")
	}
	return ir.Property{Object: object, Name: field.Name, Of: of}, nil
}
