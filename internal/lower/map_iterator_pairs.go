package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Consume each pair before asking for the next one. A readonly yield can still alias a mutable
// tuple the iterator updates on its next call, so collecting all tuples first is not equivalent.
func (l *lowering) mapFromIteration(where, sourceNode *ast.Node, plan *iterationPlan, key, value ir.Type) (ir.Expression, error) {
	yielded := l.concrete(l.checker.GetTypeOfSymbol(l.checker.GetPropertyOfType(plan.step, "value")))
	if plan.element != ir.Object || !checker.IsTupleType(yielded) {
		return nil, l.notYet(sourceNode, "new Map from something that isn't [key, value] pairs")
	}
	components := l.checker.GetTypeArguments(yielded)
	length := l.checker.GetPropertyOfType(yielded, "length")
	if len(components) != 2 || length == nil || l.checker.GetTypeOfSymbol(length).Flags()&checker.TypeFlagsNumberLiteral == 0 {
		return nil, l.notYet(sourceNode, "new Map from iterator pairs without exactly two required elements")
	}
	pairKey, keyKnown := l.representation(components[0])
	pairValue, valueKnown := l.representation(components[1])
	if !keyKnown || !valueKnown || pairKey != key || pairValue != value {
		return nil, l.notYet(sourceNode, "new Map from pairs held otherwise than the Map's keys and values")
	}
	source, err := l.expression(sourceNode)
	if err != nil {
		return nil, err
	}
	function := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "map_from_iterator", Returns: ir.Map})
	outerIndex := l.functionIndex
	l.functionIndex = function
	defer func() { l.functionIndex = outerIndex }()
	held := l.iterationLocal("iterable", ir.Object, function)
	output := l.iterationLocal("map", ir.Map, function)
	pair := l.iterationLocal("pair", ir.Object, function)
	setup, state, err := l.startIteration(where, plan, ir.Read{Local: held, Of: ir.Object})
	if err != nil {
		return nil, err
	}
	// Keep the strong key and value edges visible to the existing cycle proof.
	l.writeSites = append(l.writeSites, writeSite{holder: l.concrete(l.checker.GetTypeAtLocation(where)), node: where})
	site := len(l.writeSites)
	step, yieldedValue := l.iterationStep(state)
	tuple := ir.Read{Local: pair, Of: ir.Object}
	result := ir.Read{Local: output, Of: ir.Map}
	step = append(step,
		ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: true}},
		ir.Declare{Local: pair, Value: yieldedValue},
		ir.Evaluate{Value: ir.MapSet{Map: result, Key: ir.Property{Object: tuple, Name: "0", Of: key}, Value: ir.Property{Object: tuple, Name: "1", Of: value}, KeyType: key, ValueType: value, Site: site}},
	)
	loop := ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: step, Update: []ir.Statement{ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: false}}}}
	body, err := l.closeIteration(where, state, []ir.Statement{loop})
	if err != nil {
		return nil, err
	}
	body = append(setup, body...)
	body = append([]ir.Statement{ir.Declare{Local: output, Value: ir.MapNew{Key: key, Value: value}}}, body...)
	body = append(body, ir.Return{Value: result})
	l.result.Functions[function] = ir.Function{Name: "map_from_iterator", Returns: ir.Map, Parameters: []int{held}, Body: body}
	return ir.Call{Function: function, Arguments: []ir.Expression{source}, Returns: ir.Map}, nil
}
