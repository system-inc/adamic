package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// collectIteration keeps stepping and mapping interleaved. Materializing before mapping would
// change side effects and would fail to close the iterator when the mapper throws.
func (lowering *lowering) collectIteration(
	where, sourceNode, callbackNode *ast.Node,
	plan *iterationPlan,
	element ir.Type,
) (ir.Expression, error) {
	source, expressionError := lowering.expression(sourceNode)
	if expressionError != nil {
		return nil, expressionError
	}
	arguments := []ir.Expression{source}
	var callback ir.Expression
	var callbackReturns ir.Type
	callbackCount := 0
	if callbackNode != nil {
		callback, expressionError = lowering.expression(callbackNode)
		if expressionError != nil {
			return nil, expressionError
		}
		signatures := lowering.checker.GetSignaturesOfType(
			lowering.checker.GetTypeAtLocation(callbackNode),
			checker.SignatureKindCall,
		)
		if len(signatures) != 1 || len(signatures[0].Parameters()) > 2 {
			return nil, lowering.notYet(callbackNode, "an Array.from mapper with overloads or more than two parameters")
		}
		callbackCount = len(signatures[0].Parameters())
		callbackReturns, _ = lowering.representation(lowering.checker.GetReturnTypeOfSignature(signatures[0]))
		if callbackReturns != element {
			return nil, lowering.notYet(callbackNode, "an Array.from mapper with a different result representation")
		}
		if callbackCount > 0 {
			parameter, _ := lowering.representation(lowering.checker.GetTypeOfSymbol(signatures[0].Parameters()[0]))
			if parameter != plan.element {
				return nil, lowering.notYet(callbackNode, "an Array.from mapper with a different input representation")
			}
		}
		if callbackCount > 1 {
			parameter, _ := lowering.representation(lowering.checker.GetTypeOfSymbol(signatures[0].Parameters()[1]))
			if parameter != ir.Number {
				return nil, lowering.notYet(callbackNode, "an Array.from mapper with a different index representation")
			}
		}
		arguments = append(arguments, callback)
	} else if element != plan.element {
		return nil, lowering.notYet(where, "a collected iterator value with a different element representation")
	}
	function := len(lowering.result.Functions)
	lowering.result.Functions = append(
		lowering.result.Functions,
		ir.Function{Name: "collect_iterator", Returns: ir.Array},
	)
	outerIndex := lowering.functionIndex
	lowering.functionIndex = function
	defer func() { lowering.functionIndex = outerIndex }()
	held := lowering.iterationLocal("iterable", ir.Object, function)
	parameters := []int{held}
	if callback != nil {
		mapped := lowering.iterationLocal("mapper", ir.Closure, function)
		parameters = append(parameters, mapped)
		callback = ir.Read{Local: mapped, Of: ir.Closure}
	}
	setup, state, startIterationError := lowering.startIteration(where, plan, ir.Read{Local: held, Of: ir.Object})
	if startIterationError != nil {
		return nil, startIterationError
	}
	output := lowering.iterationLocal("collected", ir.Array, function)
	index := lowering.iterationLocal("index", ir.Number, function)
	setup = append(
		setup,
		ir.Declare{Local: output, Value: ir.ArrayLiteral{Element: element}},
		ir.Declare{Local: index, Value: ir.NumberConstant{Value: 0}},
	)
	step, value := lowering.iterationStep(state)
	if callback != nil {
		callArguments := []ir.Expression{}
		if callbackCount > 0 {
			callArguments = append(callArguments, value)
		} else {
			step = append(step, ir.Evaluate{Value: value})
		}
		if callbackCount > 1 {
			callArguments = append(callArguments, ir.Read{Local: index, Of: ir.Number})
		}
		step = append(step, ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: true}})
		value = ir.CallClosure{Closure: callback, Arguments: callArguments, Returns: callbackReturns}
	}
	step = append(step, ir.Evaluate{
		Value: ir.ArrayPush{
			Array:   ir.Read{Local: output, Of: ir.Array},
			Value:   value,
			Element: element,
			Site:    lowering.iterationArrayWriteSite(where),
		},
	}, ir.Assign{Local: index, Value: ir.Binary{
		Operator: ir.Add,
		Left:     ir.Read{Local: index, Of: ir.Number},
		Right:    ir.NumberConstant{Value: 1},
	}})
	loop := ir.Loop{
		Condition: ir.BooleanConstant{Value: true},
		Body:      step,
		Update:    []ir.Statement{ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: false}}},
	}
	body := []ir.Statement{loop}
	if callback != nil {
		var closeIterationError error
		body, closeIterationError = lowering.closeIteration(where, state, body)
		if closeIterationError != nil {
			return nil, closeIterationError
		}
	}
	setup = append(setup, body...)
	setup = append(setup, ir.Return{Value: ir.Read{Local: output, Of: ir.Array}})
	lowering.result.Functions[function] = ir.Function{
		Name:       "collect_iterator",
		Returns:    ir.Array,
		Parameters: parameters,
		Body:       setup,
	}
	return ir.Call{Function: function, Arguments: arguments, Returns: ir.Array}, nil
}

func (lowering *lowering) iteratedType(node *ast.Node) (*iterationPlan, ir.Type, error) {
	if function, isGenerator, generatorCalleeError := lowering.generatorCallee(node); generatorCalleeError != nil {
		return nil, 0, generatorCalleeError
	} else if isGenerator {
		return nil, lowering.result.Functions[function].Yields, nil
	}
	if plan, planIterationError := lowering.planIteration(node); planIterationError != nil {
		return nil, 0, planIterationError
	} else if plan != nil {
		return plan, plan.element, nil
	}
	element, elementTypeError := lowering.elementType(node)
	return nil, element, elementTypeError
}

// Destructuring steps only as far as the pattern asks. Defaults are evaluated after their step,
// and each binding is declared before the next step. Exhaustion is sticky, including across holes.
func (lowering *lowering) destructureIterator(
	pattern, sourceNode *ast.Node,
	plan *iterationPlan,
) ([]ir.Statement, error) {
	source, expressionError := lowering.expression(sourceNode)
	if expressionError != nil {
		return nil, expressionError
	}
	held := lowering.iterationLocal("iterable", ir.Object, lowering.functionIndex)
	setup, state, startIterationError := lowering.startIteration(pattern, plan, ir.Read{Local: held, Of: ir.Object})
	if startIterationError != nil {
		return nil, startIterationError
	}
	statements := append([]ir.Statement{ir.Declare{Local: held, Value: source}}, setup...)
	done := lowering.iterationLocal("iterator_done", ir.Boolean, lowering.functionIndex)
	statements = append(statements, ir.Declare{Local: done, Value: ir.BooleanConstant{Value: false}})
	for _, binding := range pattern.AsBindingPattern().Elements.Nodes {
		hole := binding.Kind == ast.KindOmittedExpression || binding.Name() == nil
		if !hole && !ast.IsIdentifier(binding.Name()) {
			return nil, lowering.notYet(binding, "a nested custom iterator destructuring pattern")
		}
		if !hole && binding.AsBindingElement().DotDotDotToken != nil {
			local, declareLocalError := lowering.declareLocal(binding.Name())
			if declareLocalError != nil {
				return nil, declareLocalError
			}
			element, elementTypeError := lowering.elementType(binding.Name())
			if elementTypeError != nil || element != plan.element {
				return nil, lowering.notYet(binding, "an iterator rest binding with a different element representation")
			}
			collected := lowering.iterationLocal("rest_values", ir.Array, lowering.functionIndex)
			statements = append(statements, ir.Declare{Local: collected, Value: ir.ArrayLiteral{Element: element}})
			step, value := lowering.iterationStep(state)
			step = append(step, ir.Evaluate{Value: ir.ArrayPush{
				Array:   ir.Read{Local: collected, Of: ir.Array},
				Value:   value,
				Element: element,
				Site:    lowering.writeSite(binding.Name()),
			}})
			statements = append(statements, ir.If{
				Condition: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: done, Of: ir.Boolean}},
				Then:      []ir.Statement{ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: step}},
			}, ir.Assign{Local: done, Value: ir.BooleanConstant{Value: true}})
			// The binding stays uninitialized while next runs; only the temporary owns the array.
			statements = append(statements, ir.Declare{Local: local, Value: ir.Read{Local: collected, Of: ir.Array}})
			continue
		}
		stepLocal := lowering.iterationLocal("iterator_result", ir.Object, lowering.functionIndex)
		result := ir.Read{Local: stepLocal, Of: ir.Object}
		step := []ir.Statement{ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: false}}, ir.Declare{
			Local: stepLocal,
			Value: invokeMember(state.next, state.receiver, state.direct, state.iterator, nil, ir.Object),
		}, ir.Assign{Local: done, Value: ir.Property{Object: result, Name: "done", Of: ir.Boolean}}}
		valueLocal := -1
		if !hole {
			valueLocal = lowering.iterationLocal("destructured_value", plan.element, lowering.functionIndex)
			statements = append(statements, ir.Declare{Local: valueLocal, Value: zeroValue(plan.element)})
			step = append(step, ir.If{
				Condition: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: done, Of: ir.Boolean}},
				Then: []ir.Statement{
					ir.Assign{Local: valueLocal, Value: ir.Property{Object: result, Name: "value", Of: plan.element}},
				},
			})
		}
		statements = append(
			statements,
			ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: done, Of: ir.Boolean}}, Then: step},
		)
		if hole {
			continue
		}
		var fallback ir.Expression
		if initializer := binding.AsBindingElement().Initializer; initializer != nil {
			fallback, expressionError = lowering.expression(initializer)
			if expressionError != nil {
				return nil, expressionError
			}
		} else {
			valueType := lowering.checker.GetTypeOfSymbol(lowering.checker.GetPropertyOfType(plan.step, "value"))
			if !lowering.includesUndefined(valueType) {
				return nil, &Refused{
					Where: lowering.program.Where(binding),
					What:  "an iterator binding whose type omits undefined on exhaustion",
					Fix:   "give the binding a default, [value = fallback], or make the yielded type include undefined; TypeScript does not account for early iterator exhaustion",
				}
			}
			fallback = fit(ir.Undefined{}, plan.element)
		}
		local, declareLocalError := lowering.declareLocal(binding.Name())
		if declareLocalError != nil {
			return nil, declareLocalError
		}
		of := lowering.result.Locals[local].Type
		value := ir.Expression(ir.Read{Local: valueLocal, Of: plan.element})
		if binding.AsBindingElement().Initializer != nil &&
			lowering.includesUndefined(
				lowering.checker.GetTypeOfSymbol(lowering.checker.GetPropertyOfType(plan.step, "value")),
			) {
			value = ir.Coalesce{Value: value, Fallback: fit(fallback, of), Of: of}
		} else {
			value = fit(value, of)
		}
		fallback = fit(fallback, of)
		if value.Type() != of || fallback.Type() != of {
			return nil, lowering.notYet(binding, "an iterator default with a different binding representation")
		}
		temporary := lowering.iterationLocal("binding_value", of, lowering.functionIndex)
		statements = append(statements, ir.Declare{Local: temporary, Value: zeroValue(of)}, ir.Assign{
			Local: state.active,
			Value: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: done, Of: ir.Boolean}},
		})
		evaluation := []ir.Statement{ir.Assign{
			Local: temporary,
			Value: ir.Conditional{Condition: ir.Read{Local: done, Of: ir.Boolean}, WhenTrue: fallback, WhenNot: value},
		}, ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: false}}}
		guarded, closeIterationError := lowering.closeIteration(binding, state, evaluation)
		if closeIterationError != nil {
			return nil, closeIterationError
		}
		statements = append(statements, guarded...)
		statements = append(statements, ir.Declare{Local: local, Value: ir.Read{Local: temporary, Of: of}})
	}
	statements = append(statements, ir.Assign{
		Local: state.active,
		Value: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: done, Of: ir.Boolean}},
	})
	close, closeIterationError := lowering.closeIteration(pattern, state, nil)
	if closeIterationError != nil {
		return nil, closeIterationError
	}
	return append(statements, close...), nil
}

// Generated spread collectors have no source-level array receiver. Keep unknown holder types
// conservative for cycle analysis; Array.from has its result type available on the call itself.
func (lowering *lowering) iterationArrayWriteSite(where *ast.Node) int {
	var holder *checker.Type
	proven := lowering.concrete(lowering.checker.GetTypeAtLocation(where))
	if lowering.isArrayType(proven) {
		holder = proven
	}
	lowering.writeSites = append(lowering.writeSites, writeSite{holder: holder, node: where})
	return len(lowering.writeSites)
}
