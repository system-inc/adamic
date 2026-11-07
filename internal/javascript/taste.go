package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (javascriptEmitter *emitter) logicalAssignment(expression ir.LogicalAssignment) string {
	var target, right string
	switch write := expression.Write.(type) {
	case ir.Assign:
		target, right = javascriptEmitter.variable(write.Local), javascriptEmitter.value(write.Value)
	case ir.SetProperty:
		if expression.Key == nil &&
			javascriptEmitter.program.CheckedField(write.ViewReceiverTypeID, write.Name) &&
			!write.Define &&
			!write.Uninitialized {
			return javascriptEmitter.viewedLogicalAssignment(expression, write)
		}
		target, right = javascriptEmitter.value(write.Object), javascriptEmitter.value(write.Value)
		if expression.Key != nil {
			target += "[" + javascriptEmitter.value(expression.Key) + "]"
		} else {
			target += "[" + quote(write.Name) + "]"
		}
	case ir.SetIndex:
		target, right = javascriptEmitter.value(write.Array)+
			"["+
			javascriptEmitter.value(write.Index)+
			"]", javascriptEmitter.value(write.Value)
	default:
		panic(fmt.Sprintf("javascript: no logical assignment for %T", expression.Write))
	}
	result := "(" + target + " " + expression.Operator + " " + right + ")"
	if write, ok := expression.Write.(ir.Assign); ok && write.Checked {
		result = fmt.Sprintf(
			"(%s ? %s : adamicUnready(%s))",
			readyName(write.Local),
			result,
			quote(javascriptEmitter.program.Locals[write.Local].Name),
		)
	}
	return result
}

// viewedLogicalAssignment reads and writes the field through a checked view, as a plain read and = through
// it do: the receiver once, then the checked read, and the right side and the checked write only when
// the operator takes the write.
func (javascriptEmitter *emitter) viewedLogicalAssignment(
	expression ir.LogicalAssignment,
	write ir.SetProperty,
) string {
	read, isProperty := expression.Read.(ir.Property)
	if !isProperty || read.View == "" {
		panic("compiler bug: a logical assignment through a checked view without its checked read")
	}
	receiver, current, value := javascriptEmitter.temporary(), javascriptEmitter.temporary(), javascriptEmitter.temporary()
	kept := map[string]string{
		"&&=": "!" + current,
		"||=": current,
		"??=": current + " !== null && " + current + " !== undefined",
	}[expression.Operator]
	return fmt.Sprintf(
		"((%s) => { const %s = %s; if (%s) return %s; const %s = %s; %s; return %s; })(%s)",
		receiver,
		current,
		javascriptEmitter.checkedViewFieldAt(read, receiver),
		kept,
		current,
		value,
		javascriptEmitter.value(write.Value),
		javascriptEmitter.checkedViewWrite(write, receiver, value),
		value,
		javascriptEmitter.value(write.Object),
	)
}

func (javascriptEmitter *emitter) effects(expression ir.Effects) string {
	savedOut, savedOptions := javascriptEmitter.out, javascriptEmitter.options
	javascriptEmitter.out = strings.Builder{}
	// Internal statements have no independent graph points: the enclosing expression is the point.
	javascriptEmitter.options.Mark = nil
	javascriptEmitter.statements(expression.Body)
	javascriptEmitter.line("return %s;", javascriptEmitter.value(expression.Result))
	text := javascriptEmitter.out.String()
	javascriptEmitter.out, javascriptEmitter.options = savedOut, savedOptions
	return "(() => {\n" + text + "})()"
}
