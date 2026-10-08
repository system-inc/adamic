package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
)

// A tuple member target evaluates its receiver before reading this RHS element.
// Holding the RHS retains its identity: earlier writes can change later elements
// when target and source alias, as JavaScript's tuple iterator observes them.
func (l *lowering) destructuredTupleTarget(target *ast.Node, held int, source []*checker.Type, position int) (ir.Statement, error) {
	access := target.AsElementAccessExpression()
	tuple := l.checker.GetTypeAtLocation(access.Expression)
	index := ast.SkipParentheses(access.ArgumentExpression)
	if !checker.IsTupleType(tuple) || index.Kind != ast.KindNumericLiteral {
		return nil, l.notYet(target, "a destructuring member target other than a fixed tuple index")
	}
	elements := l.checker.GetTypeArguments(tuple)
	offset, err := strconv.Atoi(index.Text())
	if err != nil || offset < 0 || offset >= len(elements) {
		return nil, l.notYet(target, "a destructuring member target outside its tuple")
	}
	of, known := l.representation(elements[offset])
	if !known || destructuringFieldSlotless(of) && of != ir.Union {
		return nil, l.notYet(target, "a destructuring member target without a field representation")
	}
	receiver, err := l.expression(access.Expression)
	if err != nil {
		return nil, err
	}
	if receiver.Type() != ir.Object {
		return nil, l.notYet(target, "a destructuring member target without an object representation")
	}
	value, err := l.tupleField(target, held, source, position, of)
	if err != nil {
		return nil, err
	}
	return ir.SetProperty{Object: receiver, Name: strconv.Itoa(offset), Value: value, Site: l.writeSite(access.Expression)}, nil
}
