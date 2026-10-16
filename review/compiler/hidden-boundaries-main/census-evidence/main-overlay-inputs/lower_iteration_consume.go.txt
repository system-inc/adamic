package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// collectIteration keeps stepping and mapping interleaved. Materializing before mapping would
// change side effects and would fail to close the iterator when the mapper throws.
func (l *lowering) collectIteration(where, sourceNode, callbackNode *ast.Node, plan *iterationPlan, element ir.Type) (ir.Expression, error) {
	source, err := l.expression(sourceNode)
	if err != nil {
		return nil, err
	}
	arguments := []ir.Expression{source}
	var callback ir.Expression
	var callbackReturns ir.Type
	callbackCount := 0
	if callbackNode != nil {
		callback, err = l.expression(callbackNode)
		if err != nil {
			return nil, err
		}
		signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(callbackNode), checker.SignatureKindCall)
		if len(signatures) != 1 || len(signatures[0].Parameters()) > 2 {
			return nil, l.notYet(callbackNode, "an Array.from mapper with overloads or more than two parameters")
		}
		callbackCount = len(signatures[0].Parameters())
		callbackReturns, _ = l.representation(l.checker.GetReturnTypeOfSignature(signatures[0]))
		if callbackReturns != element {
			return nil, l.notYet(callbackNode, "an Array.from mapper with a different result representation")
		}
		if callbackCount > 0 {
			parameter, _ := l.representation(l.checker.GetTypeOfSymbol(signatures[0].Parameters()[0]))
			if parameter != plan.element {
				return nil, l.notYet(callbackNode, "an Array.from mapper with a different input representation")
			}
		}
		if callbackCount > 1 {
			parameter, _ := l.representation(l.checker.GetTypeOfSymbol(signatures[0].Parameters()[1]))
			if parameter != ir.Number {
				return nil, l.notYet(callbackNode, "an Array.from mapper with a different index representation")
			}
		}
		arguments = append(arguments, callback)
	} else if element != plan.element {
		return nil, l.notYet(where, "a collected iterator value with a different element representation")
	}
	function := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "collect_iterator", Returns: ir.Array})
	outerIndex := l.functionIndex
	l.functionIndex = function
	defer func() { l.functionIndex = outerIndex }()
	held := l.iterationLocal("iterable", ir.Object, function)
	parameters := []int{held}
	if callback != nil {
		mapped := l.iterationLocal("mapper", ir.Closure, function)
		parameters = append(parameters, mapped)
		callback = ir.Read{Local: mapped, Of: ir.Closure}
	}
	setup, state, err := l.startIteration(where, plan, ir.Read{Local: held, Of: ir.Object})
	if err != nil {
		return nil, err
	}
	output := l.iterationLocal("collected", ir.Array, function)
	index := l.iterationLocal("index", ir.Number, function)
	setup = append(setup, ir.Declare{Local: output, Value: ir.ArrayLiteral{Element: element}}, ir.Declare{Local: index, Value: ir.NumberConstant{Value: 0}})
	step, value := l.iterationStep(state)
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
	step = append(step, ir.Evaluate{Value: ir.ArrayPush{Array: ir.Read{Local: output, Of: ir.Array}, Value: value, Element: element, Site: l.iterationArrayWriteSite(where)}}, ir.Assign{Local: index, Value: ir.Binary{Operator: ir.Add, Left: ir.Read{Local: index, Of: ir.Number}, Right: ir.NumberConstant{Value: 1}}})
	loop := ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: step, Update: []ir.Statement{ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: false}}}}
	body := []ir.Statement{loop}
	if callback != nil {
		body, err = l.closeIteration(where, state, body)
		if err != nil {
			return nil, err
		}
	}
	setup = append(setup, body...)
	setup = append(setup, ir.Return{Value: ir.Read{Local: output, Of: ir.Array}})
	l.result.Functions[function] = ir.Function{Name: "collect_iterator", Returns: ir.Array, Parameters: parameters, Body: setup}
	return ir.Call{Function: function, Arguments: arguments, Returns: ir.Array}, nil
}

func (l *lowering) iteratedType(node *ast.Node) (*iterationPlan, ir.Type, error) {
	if plan, err := l.planIteration(node); err != nil {
		return nil, 0, err
	} else if plan != nil {
		return plan, plan.element, nil
	}
	element, err := l.elementType(node)
	return nil, element, err
}

// Destructuring steps only as far as the pattern asks. Defaults are evaluated after their step,
// and each binding is declared before the next step. Exhaustion is sticky, including across holes.
func (l *lowering) destructureIterator(pattern, sourceNode *ast.Node, plan *iterationPlan) ([]ir.Statement, error) {
	source, err := l.expression(sourceNode)
	if err != nil {
		return nil, err
	}
	held := l.iterationLocal("iterable", ir.Object, l.functionIndex)
	setup, state, err := l.startIteration(pattern, plan, ir.Read{Local: held, Of: ir.Object})
	if err != nil {
		return nil, err
	}
	statements := append([]ir.Statement{ir.Declare{Local: held, Value: source}}, setup...)
	done := l.iterationLocal("iterator_done", ir.Boolean, l.functionIndex)
	statements = append(statements, ir.Declare{Local: done, Value: ir.BooleanConstant{Value: false}})
	for _, binding := range pattern.AsBindingPattern().Elements.Nodes {
		hole := binding.Kind == ast.KindOmittedExpression || binding.Name() == nil
		if !hole && !ast.IsIdentifier(binding.Name()) {
			return nil, l.notYet(binding, "a nested custom iterator destructuring pattern")
		}
		if !hole && binding.AsBindingElement().DotDotDotToken != nil {
			local, err := l.declareLocal(binding.Name())
			if err != nil {
				return nil, err
			}
			element, err := l.elementType(binding.Name())
			if err != nil || element != plan.element {
				return nil, l.notYet(binding, "an iterator rest binding with a different element representation")
			}
			collected := l.iterationLocal("rest_values", ir.Array, l.functionIndex)
			statements = append(statements, ir.Declare{Local: collected, Value: ir.ArrayLiteral{Element: element}})
			step, value := l.iterationStep(state)
			step = append(step, ir.Evaluate{Value: ir.ArrayPush{Array: ir.Read{Local: collected, Of: ir.Array}, Value: value, Element: element, Site: l.writeSite(binding.Name())}})
			statements = append(statements, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: done, Of: ir.Boolean}}, Then: []ir.Statement{ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: step}}}, ir.Assign{Local: done, Value: ir.BooleanConstant{Value: true}})
			// The binding stays uninitialized while next runs; only the temporary owns the array.
			statements = append(statements, ir.Declare{Local: local, Value: ir.Read{Local: collected, Of: ir.Array}})
			continue
		}
		stepLocal := l.iterationLocal("iterator_result", ir.Object, l.functionIndex)
		result := ir.Read{Local: stepLocal, Of: ir.Object}
		step := []ir.Statement{ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: false}}, ir.Declare{Local: stepLocal, Value: invokeMember(state.next, state.receiver, state.direct, state.iterator, nil, ir.Object)}, ir.Assign{Local: done, Value: ir.Property{Object: result, Name: "done", Of: ir.Boolean}}}
		valueLocal := -1
		if !hole {
			valueLocal = l.iterationLocal("destructured_value", plan.element, l.functionIndex)
			statements = append(statements, ir.Declare{Local: valueLocal, Value: zeroValue(plan.element)})
			step = append(step, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: done, Of: ir.Boolean}}, Then: []ir.Statement{ir.Assign{Local: valueLocal, Value: ir.Property{Object: result, Name: "value", Of: plan.element}}}})
		}
		statements = append(statements, ir.If{Condition: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: done, Of: ir.Boolean}}, Then: step})
		if hole {
			continue
		}
		var fallback ir.Expression
		if initializer := binding.AsBindingElement().Initializer; initializer != nil {
			fallback, err = l.expression(initializer)
			if err != nil {
				return nil, err
			}
		} else {
			valueType := l.checker.GetTypeOfSymbol(l.checker.GetPropertyOfType(plan.step, "value"))
			if !l.includesUndefined(valueType) {
				return nil, &Refused{Where: l.program.Where(binding), What: "an iterator binding whose type omits undefined on exhaustion", Fix: "give the binding a default, [value = fallback], or make the yielded type include undefined; TypeScript does not account for early iterator exhaustion"}
			}
			fallback = fit(ir.Undefined{}, plan.element)
		}
		local, err := l.declareLocal(binding.Name())
		if err != nil {
			return nil, err
		}
		of := l.result.Locals[local].Type
		value := ir.Expression(ir.Read{Local: valueLocal, Of: plan.element})
		if binding.AsBindingElement().Initializer != nil && l.includesUndefined(l.checker.GetTypeOfSymbol(l.checker.GetPropertyOfType(plan.step, "value"))) {
			value = ir.Coalesce{Value: value, Fallback: fit(fallback, of), Of: of}
		} else {
			value = fit(value, of)
		}
		fallback = fit(fallback, of)
		if value.Type() != of || fallback.Type() != of {
			return nil, l.notYet(binding, "an iterator default with a different binding representation")
		}
		temporary := l.iterationLocal("binding_value", of, l.functionIndex)
		statements = append(statements, ir.Declare{Local: temporary, Value: zeroValue(of)}, ir.Assign{Local: state.active, Value: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: done, Of: ir.Boolean}}})
		evaluation := []ir.Statement{ir.Assign{Local: temporary, Value: ir.Conditional{Condition: ir.Read{Local: done, Of: ir.Boolean}, WhenTrue: fallback, WhenNot: value}}, ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: false}}}
		guarded, err := l.closeIteration(binding, state, evaluation)
		if err != nil {
			return nil, err
		}
		statements = append(statements, guarded...)
		statements = append(statements, ir.Declare{Local: local, Value: ir.Read{Local: temporary, Of: of}})
	}
	statements = append(statements, ir.Assign{Local: state.active, Value: ir.Unary{Operator: ir.Not, Operand: ir.Read{Local: done, Of: ir.Boolean}}})
	close, err := l.closeIteration(pattern, state, nil)
	if err != nil {
		return nil, err
	}
	return append(statements, close...), nil
}

// Generated spread collectors have no source-level array receiver. Keep unknown holder types
// conservative for cycle analysis; Array.from has its result type available on the call itself.
func (l *lowering) iterationArrayWriteSite(where *ast.Node) int {
	var holder *checker.Type
	proven := l.concrete(l.checker.GetTypeAtLocation(where))
	if l.checker.IsArrayType(proven) {
		holder = proven
	}
	l.writeSites = append(l.writeSites, writeSite{holder: holder, node: where})
	return len(l.writeSites)
}
