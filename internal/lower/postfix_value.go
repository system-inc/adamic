package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// postfix lowers the value of x++ or x-- on a numeric local: the old value, held in a temporary, with the
// write of old plus or minus one done before anything after it, even when a loop test then fails (`while
// (pos-- > 0)` leaves pos at -1). Numeric locals only; any other target keeps its stop (Adamic on #bbssdcy,
// family 102, #f3n0gyp).
func (l *lowering) postfix(node *ast.Node) (ir.Expression, error) {
	return l.updateValue(node, node.AsPostfixUnaryExpression().Operand, "postfix")
}

// updateValue lowers the value of an update on a numeric local, prefix or postfix, as an immediately called
// helper, which makes the write and the read one expression: capturing the local carries its cell, so the
// caller observes the write. A prefix update's value is the local after its increment; a postfix one's is the
// old value, read and held before the increment. The two differ only in that order, so one helper keeps them
// from drifting apart (Compiler's should on #f3n0gyp).
func (l *lowering) updateValue(node, operand *ast.Node, form string) (ir.Expression, error) {
	operand = ast.SkipParentheses(operand)
	if !ast.IsIdentifier(operand) {
		return nil, l.notYet(node, "a "+form+" update value whose target is not a plain name")
	}
	local, known := l.local(operand)
	if !known || l.result.Locals[local].Type != ir.Number {
		return nil, l.notYet(node, "a "+form+" update value whose name is not stored as a number")
	}
	index := len(l.result.Functions)
	l.result.Functions = append(l.result.Functions, ir.Function{Name: form + "_update", Closure: true, Returns: ir.Number})
	outerIndex := l.functionIndex
	l.functionIndex = index
	l.closures = append(l.closures, index)
	var body []ir.Statement
	var value ir.Expression
	var err error
	held := -1
	if form == "postfix" {
		if value, err = l.expression(operand); err == nil {
			held = len(l.result.Locals)
			l.result.Locals = append(l.result.Locals, ir.Local{Name: "postfix_old", Type: ir.Number, Function: index})
			body, err = l.increment(node)
		}
	} else if body, err = l.increment(node); err == nil {
		value, err = l.expression(operand)
	}
	l.closures = l.closures[:len(l.closures)-1]
	l.functionIndex = outerIndex
	if err != nil {
		return nil, err
	}
	if value.Type() != ir.Number {
		return nil, l.notYet(node, "a "+form+" update of a non-number")
	}
	if held >= 0 {
		body = append([]ir.Statement{ir.Declare{Local: held, Value: value}}, body...)
		value = ir.Read{Local: held, Of: ir.Number}
	}
	l.result.Functions[index].Body = append(body, ir.Return{Value: value})
	return ir.CallClosure{Closure: ir.MakeClosure{Function: index}, Returns: ir.Number}, nil
}
