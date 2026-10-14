package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A Map lookup guards the receiver, not the library method. The key is
// evaluated only after that guard and the ordinary lookup keeps undefined.
func (l *lowering) optionalMapGet(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if call.QuestionDotToken != nil || callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	access := callee.AsPropertyAccessExpression()
	if access.QuestionDotToken == nil || access.Name().Text() != "get" || !l.libraryMember(callee) {
		return nil, false, nil
	}
	proven := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(access.Expression))
	if !l.isLibraryType(proven, "Map", "ReadonlyMap") {
		return nil, false, nil
	}
	if len(call.Arguments.Nodes) != 1 || hasSpread(node) {
		return nil, true, l.notYet(node, "an optional Map get with other than one key")
	}
	if l.optionalMapStructuralReceiver(access.Expression) {
		return nil, true, l.notYet(node, "an optional Map get through a structural receiver")
	}

	types := l.typeArguments(proven)
	if len(types) != 2 {
		return nil, true, l.notYet(node, "an optional Map get without key and value types")
	}
	keyType, keyKnown := l.representation(types[0])
	valueType, valueKnown := l.kept(types[1])
	if !keyKnown || !keyable(keyType) || !valueKnown || (slotless(valueType) && !(valueType == ir.Union && l.writable(types[1]))) {
		return nil, true, l.notYet(node, "an optional Map get with unsupported key or value storage")
	}
	base, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	key, err := l.expression(call.Arguments.Nodes[0])
	if err != nil {
		return nil, true, err
	}
	key = fit(key, keyType)
	if base.Type() != ir.Map || key.Type() != keyType {
		return nil, true, l.notYet(node, "an optional Map get with incompatible receiver or key storage")
	}
	function := -1
	if l.function != nil {
		function = l.functionIndex
	}
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "optional_map_base", Type: ir.Map, Function: function, ExpressionAssigned: true})
	read := ir.Read{Local: local, Of: ir.Map}
	value := ir.MapGet{Map: read, Key: key, KeyType: keyType, ValueType: valueType}
	of := value.Type()
	return ir.Effects{Body: []ir.Statement{ir.Declare{Local: local, Value: base}}, Result: ir.Conditional{
		Condition: ir.Unary{Operator: ir.Not, Operand: ir.Binary{Operator: ir.Or, Left: ir.IsUndefined{Value: read}, Right: ir.IsNull{Value: read}}},
		WhenTrue:  value, WhenNot: fit(ir.Undefined{}, of), Of: of,
	}}, true, nil
}

// A library interface alone does not prove the receiver uses native Map storage.
// Keep compatible object literals and constructed custom receivers out of the
// intrinsic path, including ones hidden behind a ReadonlyMap annotation.
func (l *lowering) optionalMapStructuralReceiver(receiver *ast.Node) bool {
	view := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(receiver))
	modules, err := l.moduleOrder(l.program.Files()[0])
	if err != nil {
		return true
	}
	hazard := false
	var visit ast.Visitor
	visit = func(node *ast.Node) bool {
		if hazard {
			return true
		}
		if node.Kind == ast.KindObjectLiteralExpression || node.Kind == ast.KindNewExpression {
			shape := l.checker.GetTypeAtLocation(node)
			if l.checker.IsTypeAssignableTo(shape, view) {
				representation, known := l.representation(shape)
				member := l.checker.GetPropertyOfType(shape, "get")
				if !known || representation != ir.Map || member == nil || !l.librarySymbol(member) {
					hazard = true
					return true
				}
			}
		}
		node.ForEachChild(visit)
		return false
	}
	for _, module := range modules {
		module.AsNode().ForEachChild(visit)
	}
	return hazard
}
