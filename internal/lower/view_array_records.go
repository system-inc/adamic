package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
)

// Fix the new own-field shape at production, without permitting reflection or
// arbitrary updates of an existing array's shape.
func (l *lowering) viewArrayRecordProduction(node *ast.Node, name string) (ir.Expression, bool, error) {
	if name != "assign" {
		return nil, false, nil
	}
	arguments := node.AsCallExpression().Arguments.Nodes
	if len(arguments) == 0 || l.viewArrayBase(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(arguments[0]))) == nil {
		return nil, false, nil
	}
	if len(arguments) != 2 || ast.SkipParentheses(arguments[0]).Kind != ast.KindArrayLiteralExpression || !l.exactObject(arguments[1], 0) {
		return nil, true, l.notYet(node, "array own-field production requiring a fresh array literal and one exact scalar record")
	}
	for _, field := range l.checker.GetPropertiesOfType(l.checker.GetTypeAtLocation(arguments[1])) {
		if _, err := strconv.ParseFloat(field.Name, 64); err == nil || arrayBrandMember(field.Name) {
			return nil, true, l.notYet(node, "array own-field production replacing index or intrinsic "+field.Name)
		}
		of, known := l.representation(l.checker.GetTypeOfSymbol(field))
		if !known || (of != ir.Number && of != ir.Boolean && of != ir.String) || field.Flags&ast.SymbolFlagsOptional != 0 {
			return nil, true, l.notYet(node, "array own-field production with unsupported field "+field.Name)
		}
	}
	array, err := l.expression(arguments[0])
	if err != nil {
		return nil, true, err
	}
	properties, err := l.expression(arguments[1])
	if err != nil {
		return nil, true, err
	}
	return ir.ArrayRecord{Array: array, Properties: properties, Where: sourceExpression(node)}, true, nil
}

func (l *lowering) viewArrayOwnReceiver(node *ast.Node, value ir.Expression) ir.Expression {
	if value.Type() != ir.Array || node.Kind != ast.KindPropertyAccessExpression {
		return value
	}
	receiver := l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(node.AsPropertyAccessExpression().Expression))
	if base := l.viewArrayBase(receiver); base != nil && l.checker.GetPropertyOfType(base, node.Name().Text()) == nil && l.checker.GetSymbolAtLocation(node.Name()) != nil {
		return ir.ArrayProperties{Array: value}
	}
	return value
}

// Own-field writes use the original copied slot's logical certificate, including
// finite literals. The current payload or the asserted array cannot certify it.
func (l *lowering) viewArrayOwnWriteReceiver(node *ast.Node, value ir.Expression) ir.Expression {
	receiver := l.viewArrayOwnReceiver(node, value)
	if _, own := receiver.(ir.ArrayProperties); own {
		name := l.fieldName(node.Name())
		l.optionalViewWriteField(name)
		if l.result.CheckedFields == nil {
			l.result.CheckedFields = map[string]bool{}
		}
		l.result.CheckedFields[name] = true
	}
	return receiver
}
