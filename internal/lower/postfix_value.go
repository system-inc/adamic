package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// postfix lowers the value of x++ or x-- on a numeric local: the old value, held in a temporary, with the
// write of old plus or minus one done before anything after it, even when a loop test then fails (`while
// (pos-- > 0)` leaves pos at -1). Numeric locals only; any other target keeps its stop (Adamic on #bbssdcy,
// family 102, #f3n0gyp). It's prefix's immediately called helper with the old value held first: capturing the
// local carries its cell, so the caller observes the write.
func (l *lowering) postfix(node *ast.Node) (ir.Expression, error) {
	postfix := node.AsPostfixUnaryExpression()
	operand := ast.SkipParentheses(postfix.Operand)
	if !ast.IsIdentifier(operand) {
		return nil, l.notYet(node, "a postfix update value whose target is not a plain name")
	}
	local, known := l.local(operand)
	if !known || l.result.Locals[local].Type != ir.Number {
		return nil, l.notYet(node, "a postfix update value whose name is not stored as a number")
	}
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "postfix_update", Closure: true, Returns: ir.Number})
	outerIndex := l.functionIndex
	l.functionIndex = index
	l.closures = append(l.closures, index)
	old, err := l.expression(operand)
	var body []ir.Statement
	held := len(l.result.Locals)
	if err == nil {
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "postfix_old", Type: ir.Number, Function: index})
		body, err = l.increment(node)
	}
	l.closures = l.closures[:len(l.closures)-1]
	l.functionIndex = outerIndex
	if err != nil {
		return nil, err
	}
	if old.Type() != ir.Number {
		return nil, l.notYet(node, "a postfix update of a non-number")
	}
	statements := append([]ir.Statement{ir.Declare{Local: held, Value: old}}, body...)
	l.result.Functions[index].Body = append(statements, ir.Return{Value: ir.Read{Local: held, Of: ir.Number}})
	return ir.CallClosure{Closure: ir.MakeClosure{Function: index}, Returns: ir.Number}, nil
}
