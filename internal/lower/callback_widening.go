package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Intrinsic callers provide element, index and array slots in their own storage.
// Capture the callback once and adapt those slots before invoking its signature.
// The adapter is ordinary IR, including every box and its ownership.
func (lowering *lowering) adaptArrayCallback(
	node *ast.Node,
	callback ir.Expression,
	supplied []ir.Type,
) (ir.Expression, int, error) {
	// Generic function values use the signature inferred for this operand's
	// concrete callback slot, the same signature used to lower their closure.
	proven := lowering.contextualGenericType(node, lowering.checker.GetTypeAtLocation(node))
	kind := int(lowering.concrete(proven).Id())
	signatures := lowering.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	if len(signatures) != 1 {
		return callback, kind, nil
	}
	parameters := signatures[0].Parameters()
	// A nullable reference uses NULL in its source storage. A boxed union
	// instead distinguishes the immortal null token from undefined (NULL).
	var incoming []*ast.Symbol
	if contextual := lowering.slotType(node); contextual != nil {
		if source := lowering.checker.GetSignaturesOfType(
			lowering.concrete(contextual),
			checker.SignatureKindCall,
		); len(source) == 1 {
			incoming = source[0].Parameters()
		}
	}
	nullable := make([]bool, len(supplied))
	takes := append([]ir.Type(nil), supplied...)
	changed := false
	for index, given := range supplied {
		if len(parameters) == 0 {
			break
		}
		parameter := parameters[min(index, len(parameters)-1)]
		rest := len(parameter.Declarations) > 0 &&
			parameter.Declarations[0].Kind == ast.KindParameter &&
			parameter.Declarations[0].AsParameterDeclaration().DotDotDotToken != nil
		if !rest && index >= len(parameters) {
			break
		}
		target := lowering.concrete(lowering.checker.GetTypeOfSymbol(parameter))
		if rest {
			target = lowering.elementTypeOfArray(target)
		}
		if target == nil {
			return nil, 0, lowering.notYet(node, "an intrinsic callback with an unknown rest element representation")
		}
		of, known := lowering.representation(target)
		if !known {
			return nil, 0, lowering.notYet(
				node,
				"an intrinsic callback parameter "+parameter.Name+" with an unknown representation",
			)
		}
		if !rest &&
			len(parameter.Declarations) > 0 &&
			parameter.Declarations[0].Kind == ast.KindParameter &&
			parameter.Declarations[0].AsParameterDeclaration().Initializer != nil {
			of = ir.Maybe(of)
		}
		if fit(ir.Read{Of: given}, of).Type() != of {
			source := typeName(given)
			if index < len(incoming) {
				source = lowering.checker.TypeToString(
					lowering.concrete(lowering.checker.GetTypeOfSymbol(incoming[index])),
				)
			}
			return nil, 0, &Refused{
				Where: lowering.program.Where(node),
				What: "an intrinsic callback parameter " +
					parameter.Name +
					" of type " +
					lowering.checker.TypeToString(target) +
					" receiving array element type " +
					source +
					" without a representation adapter",
				Fix: "annotate parameter " + parameter.Name + " as the element type " + source,
			}
		}
		if given.IsReference() && given != ir.Union && of == ir.Union && index < len(incoming) {
			source := lowering.concrete(lowering.checker.GetTypeOfSymbol(incoming[index]))
			nullable[index] = lowering.includesNull(source) && !lowering.includesUndefined(source)
		}
		takes[index] = of
		changed = changed || given != of
	}
	if !changed {
		return callback, kind, nil
	}
	// A void callback's adapter is void too, as callClosure's void call is: its slot may hold a
	// function returning a reference or a number, which the adapter's own Returns couldn't say, so
	// the call inside lets go of what comes back by the closure's own discard (ir.Discard).
	var returns ir.Type
	if result := lowering.concrete(
		lowering.checker.GetReturnTypeOfSignature(signatures[0]),
	); result.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) == 0 {
		var known bool
		if returns, known = lowering.representation(result); !known {
			return nil, 0, lowering.notYet(node, "an intrinsic callback with an unknown result representation")
		}
	}
	captured := len(lowering.result.Locals)
	lowering.result.Locals = append(
		lowering.result.Locals,
		ir.Local{Name: "intrinsic_callback", Type: ir.Closure, Function: lowering.functionIndex, Captured: true},
	)
	index := len(lowering.result.Functions)
	function := ir.Function{
		Name:        "array_callback_adapter",
		Closure:     true,
		Environment: []int{captured},
		Returns:     returns,
	}
	arguments := []ir.Expression{}
	for position, of := range supplied {
		local := len(lowering.result.Locals)
		lowering.result.Locals = append(lowering.result.Locals, ir.Local{Name: "argument", Type: of, Function: index})
		function.Parameters = append(function.Parameters, local)
		arguments = append(
			arguments,
			fitCallbackArgument(ir.Read{Local: local, Of: of}, takes[position], nullable[position]),
		)
	}
	call := ir.CallClosure{
		Closure:      ir.Read{Local: captured, Of: ir.Closure},
		Arguments:    arguments,
		FunctionType: kind,
		Returns:      returns,
	}
	if returns == 0 {
		function.Body = []ir.Statement{ir.Evaluate{Value: call}, ir.Return{}}
	} else {
		function.Body = []ir.Statement{ir.Return{Value: call}}
	}
	lowering.result.Functions = append(lowering.result.Functions, function)
	return ir.Effects{
		Body:   []ir.Statement{ir.Declare{Local: captured, Value: callback}},
		Result: ir.MakeClosure{Function: index},
	}, 0, nil
}

// Preserve a source nullable-reference protocol when fitting a boxed parameter.
// Ordinary optional references still use NULL as undefined.
func fitCallbackArgument(value ir.Expression, to ir.Type, nullPointer bool) ir.Expression {
	fitted := fit(value, to)
	if nullPointer {
		return ir.Conditional{
			Condition: ir.Truthy{Value: value},
			WhenTrue:  fitted,
			WhenNot:   ir.Box{Value: ir.Null{Of: value.Type()}},
			Of:        ir.Union,
		}
	}
	return fitted
}
