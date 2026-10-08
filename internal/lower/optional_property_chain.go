package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// optionalPropertyChain holds the receiver and each step, returning early only at ?.. Its
// element keys are capture helpers so their source effects remain inside the guarded path.
func (l *lowering) optionalPropertyChain(node *ast.Node) (ir.Expression, error) {
	steps := []*ast.Node{}
	base := node
	for base.Flags&ast.NodeFlagsOptionalChain != 0 {
		steps = append(steps, base)
		switch base.Kind {
		case ast.KindPropertyAccessExpression:
			base = base.AsPropertyAccessExpression().Expression
		case ast.KindElementAccessExpression:
			base = base.AsElementAccessExpression().Expression
		default:
			return nil, l.notYet(node, "an optional property chain containing a call")
		}
	}
	value, err := l.expression(base)
	if err != nil {
		return nil, err
	}
	result, err := l.typeOf(node)
	if err != nil {
		return nil, err
	}
	result = ir.Maybe(result)
	if !result.IsReference() && !result.IsMaybe() {
		return nil, l.notYet(node, "an optional property chain with an unrepresented result")
	}
	function := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "optional_property_chain", Returns: result})
	arguments := []ir.Expression{value}
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "receiver", Type: value.Type(), Function: function})
	parameters := []int{held}
	read := ir.Expression(ir.Read{Local: held, Of: value.Type()})
	body := []ir.Statement{}
	missing := fit(ir.Undefined{}, result)
	for index := len(steps) - 1; index >= 0; index-- {
		step := steps[index]
		part := ""
		optional := false
		if step.Kind == ast.KindPropertyAccessExpression {
			part = l.fieldName(step.Name())
			optional = step.AsPropertyAccessExpression().QuestionDotToken != nil
		} else {
			optional = step.AsElementAccessExpression().QuestionDotToken != nil
		}
		if !read.Type().IsReference() || read.Type() == ir.Union || read.Type() == ir.Weak {
			return nil, l.notYet(step, "an optional property chain through a "+typeName(read.Type()))
		}
		if optional {
			body = append(body, ir.If{
				Condition: ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: read}, Right: ir.IsNull{Value: read}},
				Then:      []ir.Statement{ir.Return{Value: missing}},
			})
		} else if step.Kind == ast.KindPropertyAccessExpression {
			read = ir.Defined{Value: read, Message: "TypeError: Cannot read properties of undefined (reading '" + part + "')"}
		}
		var next ir.Expression
		switch {
		case step.Kind == ast.KindElementAccessExpression:
			if read.Type() != ir.Array {
				return nil, l.notYet(step, "an optional element chain through a "+typeName(read.Type()))
			}
			// An index is delayed in a capture helper. Creating it has no source effects;
			// invoking it happens only after the earlier optional steps are present.
			thunk := len(l.result.Functions)
			l.result.Functions = append(l.result.Functions, ir.Function{Name: "chain_index", Closure: true, Returns: ir.Number})
			outer := l.functionIndex
			l.functionIndex = thunk
			l.closures = append(l.closures, thunk)
			key, keyErr := l.expression(step.AsElementAccessExpression().ArgumentExpression)
			l.closures = l.closures[:len(l.closures)-1]
			l.functionIndex = outer
			if keyErr != nil {
				return nil, keyErr
			}
			if key.Type() != ir.Number {
				return nil, l.notYet(step, "a non-number index in an optional chain")
			}
			l.result.Functions[thunk].Body = []ir.Statement{ir.Return{Value: key}}
			callback := len(l.result.Locals)
			l.result.Locals = append(l.result.Locals, ir.Local{Name: "index_callback", Type: ir.Closure, Function: function})
			parameters = append(parameters, callback)
			arguments = append(arguments, ir.MakeClosure{Function: thunk})
			position := len(l.result.Locals)
			l.result.Locals = append(l.result.Locals, ir.Local{Name: "index", Type: ir.Number, Function: function})
			body = append(body, ir.Declare{Local: position, Value: ir.CallClosure{Closure: ir.Read{Local: callback, Of: ir.Closure}, Returns: ir.Number}})
			key = ir.Read{Local: position, Of: ir.Number}
			// Even on an absent ordinary receiver, JavaScript evaluates the key before
			// throwing. The key is now held, so the check preserves both effects and text.
			if !optional {
				body = append(body, ir.If{Condition: ir.IsUndefined{Value: read}, Then: []ir.Statement{ir.Panic{Message: ir.Concat{Parts: []ir.Expression{
					ir.StringConstant{Index: l.constant("TypeError: Cannot read properties of undefined (reading '")}, ir.NumberToString{Value: key}, ir.StringConstant{Index: l.constant("')")},
				}}}}})
			}
			element, elementErr := l.elementType(step.AsElementAccessExpression().Expression)
			if elementErr != nil {
				return nil, elementErr
			}
			next = ir.ArrayIndex{Array: read, Index: key, Element: element}

		case read.Type() == ir.Array && part == "length":
			next = ir.Length{Array: read}
		case read.Type() == ir.String && part == "length":
			next = ir.StringLength{Value: read}
		case read.Type() == ir.Map && part == "size":
			next = ir.MapSize{Map: read}
		case read.Type() == ir.Object:
			field := l.checker.GetSymbolAtLocation(step.Name())
			if field == nil || (len(field.Declarations) > 0 && field.Declarations[0].Kind == ast.KindMethodDeclaration) || accessorSymbol(field) || l.accessorNames[step.Name().Text()] || l.inheritedLibraryMember(step) || checker.IsTupleType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(step.AsPropertyAccessExpression().Expression))) {
				return nil, l.notYet(step, "an optional property chain through a non-own data field")
			}
			stored, known := l.representation(l.checker.GetTypeOfSymbol(field))
			if !known || censusFieldSlotless(stored) || stored == ir.Weak || l.includesNull(l.checker.GetTypeOfSymbol(field)) {
				return nil, l.notYet(step, "an optional property chain through an unrepresented field")
			}
			next = l.readObjectField(step, ir.Property{Object: read, Name: part, Of: stored, Class: l.classOf(step)})
		default:
			return nil, l.notYet(step, "."+part+" in an optional property chain")
		}
		if index == 0 {
			body = append(body, ir.Return{Value: fit(next, result)})
		} else {
			local := len(l.result.Locals)
			l.result.Locals = append(l.result.Locals, ir.Local{Name: "chain_step", Type: next.Type(), Function: function})
			body = append(body, ir.Declare{Local: local, Value: next})
			read = ir.Read{Local: local, Of: next.Type()}
		}
	}
	l.result.Functions[function].Parameters = parameters
	l.result.Functions[function].Body = body
	return ir.Call{Function: function, Arguments: arguments, Returns: result}, nil
}
