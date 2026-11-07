package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
	"math"
)

// A coercion call must preserve its receiver and evaluate the original operand once.
// Only a statically promised primitive result can terminate OrdinaryToPrimitive.
func (lowering *lowering) censusNumberObject(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	proven := lowering.checker.GetTypeAtLocation(node)
	if value.Type() == ir.Array {
		if !lowering.isArrayType(proven) {
			return nil, lowering.notYet(node, "numeric coercion of a non-array view")
		}
		for _, name := range []string{"valueOf", "toString", "join"} {
			property := lowering.checker.GetPropertyOfType(proven, name)
			if property != nil {
				for _, declaration := range property.Declarations {
					if !load.IsLibrary(ast.GetSourceFileOfNode(declaration)) {
						return nil, lowering.notYet(node, "numeric coercion of an array with a custom conversion")
					}
				}
			}
		}
		element, known := lowering.kept(lowering.elementTypeOfArray(proven))
		if !known ||
			(element != ir.Number && element != ir.String && element != ir.Boolean && element != ir.MaybeNumber) {
			return nil, lowering.notYet(node, "numeric coercion of an array with non-primitive elements")
		}
		joined := ir.ArrayJoin{
			Array:     value,
			Separator: ir.StringConstant{Index: lowering.constant(",")},
			Element:   element,
		}
		return ir.NumberCall{Function: "convert", Arguments: []ir.Expression{joined}}, nil
	}
	if value.Type() != ir.Object {
		return nil, lowering.notYet(node, "numeric coercion with an unrepresented object conversion")
	}
	index := len(lowering.result.Functions)
	local := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "coercion_operand", Type: ir.Object, Function: index},
	)
	read := ir.Read{Local: local, Of: ir.Object}
	body := []ir.Statement{}
	for _, name := range []string{"valueOf", "toString"} {
		property := lowering.checker.GetPropertyOfType(proven, name)
		if property == nil ||
			len(property.Declarations) == 0 ||
			load.IsLibrary(ast.GetSourceFileOfNode(property.Declarations[0])) {
			continue
		}
		for _, declaration := range property.Declarations {
			if declaration.Kind == ast.KindGetAccessor {
				return nil, lowering.notYet(node, "numeric coercion through a conversion getter")
			}
		}
		signatures := lowering.checker.GetSignaturesOfType(
			lowering.checker.GetTypeOfSymbol(property),
			checker.SignatureKindCall,
		)
		if len(signatures) != 1 {
			return nil, lowering.notYet(node, "numeric coercion with a non-callable or overloaded conversion member")
		}
		for _, parameter := range signatures[0].Parameters() {
			takes := lowering.censusCallableParameterType(parameter)
			if !lowering.censusRelated(lowering.checker.GetUndefinedType(), takes) {
				return nil, lowering.notYet(node, "numeric coercion invoking a required conversion parameter")
			}
		}
		result := lowering.checker.GetReturnTypeOfSignature(signatures[0])
		returns, known := lowering.representation(result)
		if !known || censusCallableSlotless(returns) {
			return nil, lowering.notYet(node, "numeric coercion with an unrepresented conversion result")
		}
		call := ir.CallClosure{
			Closure: ir.Property{Object: read, Name: name, Of: ir.Closure, Method: true},
			Returns: returns,
		}
		if lowering.writable(result) {
			body = append(body, ir.Return{Value: ir.NumberCall{Function: "convert", Arguments: []ir.Expression{call}}})
			lowering.result.Functions = append(
				lowering.result.Functions,
				ir.Function{Name: "numeric_coercion", Parameters: []int{local}, Returns: ir.Number, Body: body},
			)
			return ir.Call{Function: index, Arguments: []ir.Expression{value}, Returns: ir.Number}, nil
		}
		// A returned object is evaluated and released before trying toString.
		body = append(body, ir.Evaluate{Value: call})
	}
	// A fresh literal has no hidden conversion members. An erased structural view can,
	// so absence from its type is never interpreted as the absence of user code.
	if ast.SkipParentheses(node).Kind == ast.KindObjectLiteralExpression && len(body) == 0 {
		body = append(body, ir.Return{Value: ir.NumberConstant{Value: math.NaN()}})
		lowering.result.Functions = append(
			lowering.result.Functions,
			ir.Function{Name: "numeric_coercion", Parameters: []int{local}, Returns: ir.Number, Body: body},
		)
		return ir.Call{Function: index, Arguments: []ir.Expression{value}, Returns: ir.Number}, nil
	}
	return nil, lowering.notYet(node, "numeric coercion requiring dynamic ToPrimitive")
}
