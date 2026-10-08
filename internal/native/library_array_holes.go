package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) arrayHoles(expression ir.ArrayHoles) string {
	result := e.own(ir.Array, fmt.Sprintf("adamic_array_holes(%s, %t)", e.value(expression.Length), expression.Element.IsReference()))
	e.checkThrown()
	return result
}

func (e *emitter) hasArrayHoles() bool {
	found := false
	var statements func([]ir.Statement)
	statements = func(list []ir.Statement) {
		for _, statement := range list {
			walkStatement(statement, func(expression ir.Expression) {
				switch expression.(type) {
				case ir.ArrayHoles, ir.ArraySetLength:
					found = true
				}
			}, statements)
		}
	}
	statements(e.program.Main)
	for _, f := range e.program.Functions {
		statements(f.Body)
	}
	return found
}

func (e *emitter) arrayHolesMap(expression ir.ArrayMap) string {
	source := e.snapshot(ir.Array, e.value(expression.Array))
	callback := e.value(expression.Callback)
	count := e.temporary()
	e.line("size_t %s = %s->length;", count, source)
	result := e.own(ir.Array, fmt.Sprintf("adamic_array_holes((double)%s, %t)", count, expression.Result.IsReference()))
	owner := e.viewArrayReadOwner(expression.Element)
	index, slot, element, answer := e.temporary(), e.temporary(), e.temporary(), e.temporary()
	e.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
	e.indent++
	e.line("adamic_value *%s = adamic_array_holes_at(%s, (double)%s);", slot, source, index)
	e.line("if (%s == NULL) continue;", slot)
	if expression.ViewRead.View != "" {
		checked := e.viewArrayElementSlot(expression.ViewRead, source, index, owner)
		e.line("adamic_value %s = %s;", element, checked)
	} else {
		e.line("adamic_value %s = *%s;", element, slot)
	}
	if expression.Element.IsReference() {
		e.line("adamic_retain(%s.reference);", element)
	}
	invoke := e.viewCallableBoxedInvokeTypes([]ir.Type{expression.Element, ir.Number, ir.Array}, expression.Result, false)
	e.line("adamic_value %s = %s(%s, (adamic_value[]){%s, {.number = (double)%s}, {.reference = %s}}, 3, false);", answer, invoke, callback, element, index, source)
	if expression.Element.IsReference() {
		e.closureThrown(element + ".reference")
		e.line("adamic_release(%s.reference);", element)
	} else {
		e.closureThrown()
	}
	e.line("adamic_array_holes_set(%s, (double)%s, %s);", result, index, answer)
	e.indent--
	e.line("}")
	return result
}

func (e *emitter) arrayHolesLength(statement ir.ArraySetLength) string {
	array := e.value(statement.Array)
	length := e.value(statement.Length)
	e.line("adamic_array_holes_set_length(%s, %s);", array, length)
	e.checkThrown()
	return length
}
