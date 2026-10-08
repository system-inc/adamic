package lower

import (
	"fmt"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// ordinaryString runs OrdinaryToPrimitive before spelling its primitive result.
// The operand is passed once. Default/number hints try valueOf first; the string
// hint tries toString first. Returning an object does not terminate the protocol.
func (l *lowering) ordinaryString(node *ast.Node, value ir.Expression, stringHint bool) (ir.Expression, error) {
	proven := l.concrete(l.checker.GetTypeAtLocation(node))
	if proven.Flags()&checker.TypeFlagsUnion != 0 && (l.includesUndefined(proven) || l.includesNull(proven)) && (value.Type() == ir.Object || value.Type() == ir.Array) {
		index := len(l.result.Functions)
		l.result.Functions = append(l.result.Functions, ir.Function{Name: "ordinary_optional", Returns: ir.String})
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "operand", Type: value.Type(), Function: index})
		read := ir.Read{Local: local, Of: value.Type()}
		text, err := l.ordinaryStringPresent(node, read, stringHint)
		if err != nil {
			return nil, err
		}
		missing := "undefined"
		if l.includesNull(proven) {
			missing = "null"
		}
		result := ir.Conditional{Condition: ir.IsUndefined{Value: read}, WhenTrue: ir.StringConstant{Index: l.constant(missing)}, WhenNot: text}
		l.result.Functions[index].Parameters = []int{local}
		l.result.Functions[index].Body = []ir.Statement{ir.Return{Value: result}}
		return ir.Call{Function: index, Arguments: []ir.Expression{value}, Returns: ir.String}, nil
	}
	return l.ordinaryStringPresent(node, value, stringHint)
}

func (l *lowering) ordinaryStringPresent(node *ast.Node, value ir.Expression, stringHint bool) (ir.Expression, error) {
	proven := l.concrete(l.checker.GetTypeAtLocation(node))
	if proven.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return l.primitiveString(node, value)
	}
	if checker.IsTupleType(proven) {
		return nil, l.notYet(node, "OrdinaryToPrimitive of a tuple held as an object")
	}
	if l.isLibraryType(proven, "Error", "RegExp", "RegExpStringIterator") {
		return nil, l.notYet(node, "OrdinaryToPrimitive of a host object with intrinsic conversion behavior")
	}
	if value.Type() != ir.Object && value.Type() != ir.Array {
		return l.primitiveString(node, value)
	}
	for _, property := range l.checker.GetPropertiesOfType(proven) {
		if strings.Contains(property.Name, "toPrimitive") {
			return nil, l.notYet(node, "Symbol.toPrimitive conversion")
		}
	}
	proven = l.checker.GetNonNullableType(proven)
	if value.Type() == ir.Array {
		// Built-in valueOf returns the array, so either hint reaches the built-in join.
		for _, name := range []string{"valueOf", "toString", "join"} {
			if property := l.checker.GetPropertyOfType(proven, name); property != nil && !l.librarySymbol(property) {
				return nil, l.notYet(node, "OrdinaryToPrimitive of an array with a custom conversion")
			}
		}
		element, known := l.kept(l.checker.GetElementTypeOfArrayType(proven))
		if !known {
			return nil, l.notYet(node, "OrdinaryToPrimitive of an array with unrepresented elements")
		}
		if element != ir.Number && element != ir.String && element != ir.Boolean && element != ir.MaybeNumber {
			return nil, l.notYet(node, "OrdinaryToPrimitive of an array with non-primitive elements")
		}
		return ir.ArrayJoin{Array: value, Separator: ir.StringConstant{Index: l.constant(",")}, Element: element}, nil
	}
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "ordinary_primitive", Returns: ir.String})
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "operand", Type: ir.Object, Function: index})
	read := ir.Read{Local: local, Of: ir.Object}
	body := []ir.Statement{}
	names := []string{"valueOf", "toString"}
	if stringHint {
		names = []string{"toString", "valueOf"}
	}
	for _, name := range names {
		property := l.checker.GetPropertyOfType(proven, name)
		if property == nil || l.librarySymbol(property) {
			// A structural view may hide either conversion. Absence in its static type
			// is not proof of the intrinsic: use the existing closed-program shape scan.
			if !l.exactPlainObject(node) {
				if reason := l.prototypeHazard(node, name); reason != "" {
					return nil, l.notYet(node, "OrdinaryToPrimitive through a view ("+reason+")")
				}
			}
			if name == "toString" {
				body = append(body, ir.Return{Value: ir.StringConstant{Index: l.constant("[object Object]")}})
				break
			}
			continue // Object.prototype.valueOf returns the receiver.
		}
		for _, declaration := range property.Declarations {
			if declaration.Kind == ast.KindGetAccessor || declaration.Kind == ast.KindSetAccessor {
				return nil, l.notYet(node, "OrdinaryToPrimitive through a conversion accessor")
			}
		}
		memberType := l.checker.GetTypeOfSymbol(property)
		if l.includesUndefined(memberType) && len(l.checker.GetSignaturesOfType(l.checker.GetNonNullableType(memberType), checker.SignatureKindCall)) != 0 {
			return nil, l.notYet(node, "OrdinaryToPrimitive with a possibly absent conversion method")
		}
		signatures := l.checker.GetSignaturesOfType(memberType, checker.SignatureKindCall)
		if len(signatures) == 0 {
			continue
		} // IsCallable is false.
		if len(signatures) != 1 || len(signatures[0].Parameters()) != 0 {
			return nil, l.notYet(node, "OrdinaryToPrimitive with an unproven zero-argument conversion signature")
		}
		result := l.checker.GetReturnTypeOfSignature(signatures[0])
		of, known := l.representation(result)
		if result.Flags()&checker.TypeFlagsVoid != 0 {
			of, known = 0, true
		}
		if !known {
			return nil, l.notYet(node, "OrdinaryToPrimitive with an unrepresented conversion result")
		}
		call := ir.CallClosure{Closure: ir.Property{Object: read, Name: name, Of: ir.Closure, Method: true}, Returns: of}
		if of == 0 {
			body = append(body, ir.Evaluate{Value: call}, ir.Return{Value: ir.StringConstant{Index: l.constant("undefined")}})
			break
		}
		resultLocal := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "primitive_result", Type: of, Function: index})
		body = append(body, ir.Declare{Local: resultLocal, Value: call})
		held := ir.Read{Local: resultLocal, Of: of}
		if l.writable(result) || result.Flags()&(checker.TypeFlagsNull|checker.TypeFlagsVoid) != 0 {
			text, err := l.primitiveStringType(result, held)
			if err != nil {
				return nil, l.notYet(node, err.Error())
			}
			body = append(body, ir.Return{Value: text})
			break
		}
		if of == ir.Union {
			tag := ir.TypeOf{Value: held}
			primitive := ir.Expression(ir.Binary{Operator: ir.And,
				Left:  ir.Binary{Operator: ir.NotEqual, Left: tag, Right: ir.StringConstant{Index: l.constant("object")}},
				Right: ir.Binary{Operator: ir.NotEqual, Left: tag, Right: ir.StringConstant{Index: l.constant("function")}}})
			primitive = ir.Binary{Operator: ir.Or, Left: primitive, Right: ir.IsNull{Value: held}}
			body = append(body, ir.If{Condition: primitive, Then: []ir.Statement{ir.Return{Value: ir.Conditional{Condition: ir.IsNull{Value: held}, WhenTrue: ir.StringConstant{Index: l.constant("null")}, WhenNot: ir.UnionToString{Value: held}}}}})
		}
		// Non-primitives are released on the helper's exit, including exceptional exits.
	}
	body = append(body, ir.Throw{Value: ir.MakeError{
		Name:    ir.StringConstant{Index: l.constant("TypeError")},
		Message: ir.StringConstant{Index: l.constant("Cannot convert object to primitive value")},
	}})
	l.result.Functions[index].Parameters = []int{local}
	l.result.Functions[index].Body = body
	return ir.Call{Function: index, Arguments: []ir.Expression{value}, Returns: ir.String}, nil
}

func (l *lowering) primitiveString(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	text, err := l.primitiveStringType(l.concrete(l.checker.GetTypeAtLocation(node)), l.spelled(node, value))
	if err != nil {
		return nil, l.notYet(node, err.Error())
	}
	return text, nil
}

func (l *lowering) primitiveStringType(proven *checker.Type, value ir.Expression) (ir.Expression, error) {
	if proven.Flags()&checker.TypeFlagsNull != 0 {
		return ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: value}}, Result: ir.StringConstant{Index: l.constant("null")}}, nil
	}
	if proven.Flags()&(checker.TypeFlagsUndefined|checker.TypeFlagsVoid) != 0 {
		return ir.Effects{Body: []ir.Statement{ir.Evaluate{Value: value}}, Result: ir.StringConstant{Index: l.constant("undefined")}}, nil
	}
	switch value.Type() {
	case ir.String:
		if l.includesUndefined(proven) {
			return ir.Coalesce{Value: value, Fallback: ir.StringConstant{Index: l.constant("undefined")}, Of: ir.String}, nil
		}
		return value, nil
	case ir.Number:
		return ir.NumberToString{Value: value}, nil
	case ir.Boolean:
		return ir.BooleanToString{Value: value}, nil
	case ir.MaybeNumber, ir.MaybeBoolean:
		return ir.MaybeToString{Value: value}, nil
	case ir.Union:
		if l.writable(proven) {
			return ir.UnionToString{Value: value}, nil
		}
	}
	return nil, fmt.Errorf("primitive string conversion requiring dynamic ToPrimitive")
}

// Addition evaluates BOTH operands before either conversion. This also preserves
// the receiver if evaluating the right operand changes the left operand's alias.
func (l *lowering) ordinaryStringAddition(node *ast.Node, left, right ir.Expression) (ir.Expression, error) {
	binary := node.AsBinaryExpression()
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "ordinary_addition", Returns: ir.String})
	parameters := []int{}
	reads := []ir.Expression{}
	for _, value := range []ir.Expression{left, right} {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "operand", Type: value.Type(), Function: index})
		parameters = append(parameters, local)
		reads = append(reads, ir.Read{Local: local, Of: value.Type()})
	}
	a, err := l.ordinaryString(binary.Left, reads[0], false)
	if err != nil {
		return nil, err
	}
	b, err := l.ordinaryString(binary.Right, reads[1], false)
	if err != nil {
		return nil, err
	}
	l.result.Functions[index].Parameters = parameters
	l.result.Functions[index].Body = []ir.Statement{ir.Return{Value: ir.Concat{Parts: []ir.Expression{a, b}}}}
	return ir.Call{Function: index, Arguments: []ir.Expression{left, right}, Returns: ir.String}, nil
}
