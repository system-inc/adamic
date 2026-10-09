package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Relational comparison uses the number hint. Date's default/string hint selects
// local toString and is forbidden by ruling 10. Equality retains object identity.
func (l *lowering) dateBinary(node *ast.Node, operator ast.Kind, left, right ir.Expression) (ir.Expression, bool, error) {
	if node.Kind != ast.KindBinaryExpression {
		return nil, false, nil
	}
	binary := node.AsBinaryExpression()
	leftDate := l.isLibraryType(l.checker.GetTypeAtLocation(binary.Left), "Date")
	rightDate := l.isLibraryType(l.checker.GetTypeAtLocation(binary.Right), "Date")
	if !leftDate && !rightDate {
		return nil, false, nil
	}
	if operator == ast.KindPlusToken {
		return nil, true, l.dateForbidden(node, "Date addition (default ToPrimitive uses local toString)")
	}
	if operator != ast.KindLessThanToken && operator != ast.KindLessThanEqualsToken && operator != ast.KindGreaterThanToken && operator != ast.KindGreaterThanEqualsToken {
		return nil, false, nil
	}
	if leftDate {
		left = ir.NodeFSFile{Operation: "date_time", Arguments: []ir.Expression{left}, Of: ir.Number}
	}
	if rightDate {
		right = ir.NodeFSFile{Operation: "date_time", Arguments: []ir.Expression{right}, Of: ir.Number}
	}
	if left.Type() != ir.Number || right.Type() != ir.Number {
		return nil, true, l.notYet(node, "Date relational comparison against a non-number representation")
	}
	return ir.Binary{Operator: comparisons[operator], Left: left, Right: right}, true, nil
}
