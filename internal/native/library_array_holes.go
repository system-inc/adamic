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
				if _, ok := expression.(ir.ArrayHoles); ok {
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
