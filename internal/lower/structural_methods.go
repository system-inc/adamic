package lower

import (
	"sort"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Instance methods have named shape thunks. Constructor objects instead carry
// virtual slots. Select by the actual receiver's identity, never by the mere
// presence of a static declaration or by a structural signature's name.
func (l *lowering) structuralMethodCall(node *ast.Node) (ir.Expression, error) {
	value, err := l.callClosure(node)
	if err != nil {
		return nil, err
	}
	call, ok := value.(ir.CallClosure)
	if !ok {
		return value, nil
	}
	property, ok := call.Closure.(ir.Property)
	if !ok {
		return value, nil
	}
	callee := ast.SkipParentheses(node.AsCallExpression().Expression)
	receiver := callee.AsPropertyAccessExpression().Expression
	receiverType := l.checker.GetNonNullableType(l.concrete(l.checker.GetTypeAtLocation(receiver)))
	candidates := []*instance{}
	for symbol, static := range l.statics {
		if !l.checker.IsTypeAssignableTo(l.checker.GetTypeOfSymbol(symbol), receiverType) {
			continue
		}
		if _, exists := static.methods[property.Name]; exists {
			candidates = append(candidates, static)
		}
	}
	if len(candidates) == 0 {
		return value, nil
	}
	if property.Optional {
		return nil, l.notYet(node, "an optional structural method call on a possible constructor object")
	}
	sort.Slice(candidates, func(i, j int) bool { return candidates[i].class < candidates[j].class })
	snapshot := l.structuralReceiver(property, receiver)
	index := len(l.result.Functions)
	function := ir.Function{Name: "structural_method", Returns: call.Returns}
	arguments := append([]ir.Expression{snapshot}, call.Arguments...)
	forwarded := []ir.Expression{}
	for position, argument := range arguments {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: argument.Type(), Function: index})
		function.Parameters = append(function.Parameters, local)
		forwarded = append(forwarded, ir.Read{Local: local, Of: argument.Type()})
		origin := receiver
		if position > 0 {
			origin = node.AsCallExpression().Arguments.Nodes[position-1]
		}
		l.noteLocal(local, l.concrete(l.checker.GetTypeAtLocation(origin)), origin)
	}
	finish := func(value ir.Expression) []ir.Statement {
		if function.Returns == 0 {
			return []ir.Statement{ir.Evaluate{Value: value}, ir.Return{}}
		}
		return []ir.Statement{ir.Return{Value: fit(value, function.Returns)}}
	}
	for _, static := range candidates {
		target := static.methods[property.Name]
		signature := l.result.Functions[target]
		if len(signature.Parameters) != len(forwarded) || signature.Returns != call.Returns {
			return nil, l.notYet(node, "a structural static method with a different native signature")
		}
		for position, parameter := range signature.Parameters {
			if l.result.Locals[parameter].Type != forwarded[position].Type() {
				return nil, l.notYet(node, "a structural static method with a different native signature")
			}
		}
		methodArguments := append([]ir.Expression{ir.Property{Object: forwarded[0], Name: "receiver", Of: ir.Object}}, forwarded[1:]...)
		dispatch := ir.Call{Function: target, Arguments: methodArguments, Returns: signature.Returns, Virtual: static.slots[property.Name] + 1}
		function.Body = append(function.Body, ir.If{Condition: ir.InstanceOf{Value: methodArguments[0], Class: static.class}, Then: finish(dispatch)})
	}
	property.Object = ir.Property{Object: forwarded[0], Name: "receiver", Of: ir.Object}
	call.Closure, call.Arguments = property, forwarded[1:]
	cached := ir.Property{Object: forwarded[0], Name: "callee", Of: ir.Closure}
	cachedCall := ir.CallClosure{Closure: cached, Arguments: forwarded[1:], Returns: call.Returns}
	function.Body = append(function.Body, ir.If{Condition: ir.IsUndefined{Value: cached}, Then: finish(call), Else: finish(cachedCall)})
	l.result.Functions = append(l.result.Functions, function)
	return ir.Call{Function: index, Arguments: arguments, Returns: function.Returns}, nil
}

// A method stored as a function-valued field can be replaced through another
// structural view while the arguments run. Snapshot that own field before the
// arguments, together with its receiver. Prototype and static methods cannot be
// replaced by the represented method-store rules, so their table lookup is stable.
func (l *lowering) structuralReceiver(property ir.Property, receiver *ast.Node) ir.Expression {
	index := len(l.result.Functions)
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "receiver", Type: ir.Object, Function: index})
	l.noteLocal(local, l.concrete(l.checker.GetTypeAtLocation(receiver)), receiver)
	object := ir.Read{Local: local, Of: ir.Object}
	key := ir.StringConstant{Index: l.constant(property.Name)}
	cached := ir.Conditional{
		Condition: ir.ObjectCall{Method: "hasOwn", Arguments: []ir.Expression{object, key}, Returns: ir.Boolean},
		WhenTrue:  ir.Property{Object: object, Name: property.Name, Of: ir.Closure, Absent: true},
		WhenNot:   ir.Undefined{Of: ir.Closure}, Of: ir.Closure,
	}
	body := []ir.Statement{ir.Return{Value: ir.ObjectLiteral{Fields: []ir.Field{
		{Name: "receiver", Value: object}, {Name: "callee", Value: cached},
	}}}}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "structural_receiver", Parameters: []int{local}, Returns: ir.Object, Body: body})
	return ir.Call{Function: index, Arguments: []ir.Expression{property.Object}, Returns: ir.Object}
}
