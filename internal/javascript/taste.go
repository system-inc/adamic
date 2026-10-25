package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) logicalAssignment(expression ir.LogicalAssignment) string {
	var target, right string
	switch write := expression.Write.(type) {
	case ir.Assign:
		target, right = e.variable(write.Local), e.value(write.Value)
	case ir.SetProperty:
		if expression.Key == nil && e.program.CheckedField(write.ViewReceiverTypeID, write.Name) && !write.Define && !write.Uninitialized {
			return e.viewedLogicalAssignment(expression, write)
		}
		target, right = e.value(write.Object), e.value(write.Value)
		if expression.Key != nil {
			target += "[" + e.value(expression.Key) + "]"
		} else {
			target += "[" + quote(write.Name) + "]"
		}
	case ir.SetIndex:
		target, right = e.value(write.Array)+"["+e.value(write.Index)+"]", e.value(write.Value)
	default:
		panic(fmt.Sprintf("javascript: no logical assignment for %T", expression.Write))
	}
	result := "(" + target + " " + expression.Operator + " " + right + ")"
	if write, ok := expression.Write.(ir.Assign); ok && write.Checked {
		result = fmt.Sprintf("(%s ? %s : adamicUnready(%s))", readyName(write.Local), result, quote(e.program.Locals[write.Local].Name))
	}
	return result
}

// viewedLogicalAssignment reads and writes the field through a checked view, as a plain read and = through
// it do: the receiver once, then the checked read, and the right side and the checked write only when
// the operator takes the write.
func (e *emitter) viewedLogicalAssignment(expression ir.LogicalAssignment, write ir.SetProperty) string {
	read, isProperty := expression.Read.(ir.Property)
	if !isProperty || read.View == "" {
		panic("compiler bug: a logical assignment through a checked view without its checked read")
	}
	receiver, current, value := e.temporary(), e.temporary(), e.temporary()
	kept := map[string]string{"&&=": "!" + current, "||=": current, "??=": current + " !== null && " + current + " !== undefined"}[expression.Operator]
	return fmt.Sprintf("((%s) => { const %s = %s; if (%s) return %s; const %s = %s; %s; return %s; })(%s)", receiver, current, e.checkedViewFieldAt(read, receiver), kept, current, value, e.value(write.Value), e.checkedViewWrite(write, receiver, value), value, e.value(write.Object))
}

func (e *emitter) effects(expression ir.Effects) string {
	savedOut, savedOptions := e.out, e.options
	e.out = strings.Builder{}
	// Internal statements have no independent graph points: the enclosing expression is the point.
	e.options.Mark = nil
	e.statements(expression.Body)
	e.line("return %s;", e.value(expression.Result))
	text := e.out.String()
	e.out, e.options = savedOut, savedOptions
	return "(() => {\n" + text + "})()"
}
