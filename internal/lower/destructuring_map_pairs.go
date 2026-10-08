package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// iterableMapPairs proves a custom iterable yields two-slot tuples. Library
// structural iterable types can reveal the pair type but still need a concrete
// runtime protocol origin; planIteration keeps that boundary explicit.
func (l *lowering) iterableMapPairs(source *ast.Node) (ir.Type, ir.Type, *iterationPlan, bool, error) {
	sourceType := l.checker.GetTypeAtLocation(source)
	var pair *checker.Type
	if l.isLibraryType(sourceType, "Generator", "Iterable", "IterableIterator", "IteratorObject") {
		arguments := l.typeArguments(sourceType)
		if len(arguments) > 0 {
			pair = arguments[0]
		}
	}
	plan, err := l.planIteration(source)
	if err != nil {
		return 0, 0, nil, false, err
	}
	if plan != nil {
		pair = l.checker.GetTypeOfSymbol(l.checker.GetPropertyOfType(plan.step, "value"))
	}
	if pair == nil || !checker.IsTupleType(pair) {
		return 0, 0, nil, false, nil
	}
	elements := l.checker.GetTypeArguments(pair)
	if len(elements) != 2 {
		return 0, 0, nil, false, nil
	}
	key, keyKnown := l.representation(elements[0])
	value, valueKnown := l.representation(elements[1])
	return key, value, plan, keyKnown && valueKnown, nil
}

// mapFromIteration inserts before asking for the next entry. Collecting tuples
// first would miscompile an iterator which mutates and yields the same tuple.
func (l *lowering) mapFromIteration(where, sourceNode *ast.Node, plan *iterationPlan, key, value ir.Type) (ir.Expression, error) {
	source, err := l.expression(sourceNode)
	if err != nil {
		return nil, err
	}
	function := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "map_from_iterator", Returns: ir.Map})
	outer := l.functionIndex
	l.functionIndex = function
	defer func() { l.functionIndex = outer }()
	held := l.iterationLocal("iterable", ir.Object, function)
	setup, state, err := l.startIteration(where, plan, ir.Read{Local: held, Of: ir.Object})
	if err != nil {
		return nil, err
	}
	output := l.iterationLocal("constructed_map", ir.Map, function)
	setup = append(setup, ir.Declare{Local: output, Value: ir.MapNew{Key: key, Value: value}})
	step, entry := l.iterationStep(state)
	pair := l.iterationLocal("map_entry", ir.Object, function)
	step = append(step, ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: true}}, ir.Declare{Local: pair, Value: entry})
	read := ir.Read{Local: pair, Of: ir.Object}
	step = append(step, ir.Evaluate{Value: ir.MapSet{Map: ir.Read{Local: output, Of: ir.Map}, Key: ir.Property{Object: read, Name: "0", Of: key}, Value: ir.Property{Object: read, Name: "1", Of: value}, KeyType: key, ValueType: value}})
	loop := ir.Loop{Condition: ir.BooleanConstant{Value: true}, Body: step, Update: []ir.Statement{ir.Assign{Local: state.active, Value: ir.BooleanConstant{Value: false}}}}
	body, err := l.closeIteration(where, state, []ir.Statement{loop})
	if err != nil {
		return nil, err
	}
	setup = append(setup, body...)
	setup = append(setup, ir.Return{Value: ir.Read{Local: output, Of: ir.Map}})
	l.result.Functions[function] = ir.Function{Name: "map_from_iterator", Returns: ir.Map, Parameters: []int{held}, Body: setup}
	return ir.Call{Function: function, Arguments: []ir.Expression{source}, Returns: ir.Map}, nil
}
