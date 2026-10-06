package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// parallelMap keeps both operands alive through the runtime's join. The result is owned.
func (e *emitter) parallelMap(expression ir.ParallelMap) string {
	items := e.value(expression.Items)
	original := e.value(expression.Work)
	// A fresh adapter carries the original closure plus reference-valued globals.
	// Calling the original code with its original self preserves function identity.
	adapter := e.temporary() + "_parallel"
	e.declarations = append(e.declarations, fmt.Sprintf("static adamic_value %s(adamic_closure *self, adamic_value *arguments) {\n\tadamic_closure *work = self->cells[0]->value.reference;\n\treturn work->code(work, arguments);\n}\n", adapter))
	work := e.own(ir.Closure, fmt.Sprintf("adamic_closure_new(%s, %d)", adapter, len(expression.Shared)+1))
	e.line("%s->cells[0] = adamic_cell_new((adamic_value){.reference = %s}, true);", work, retained(original))
	for index, global := range expression.Shared {
		value := e.value(global)
		e.line("%s->cells[%d] = adamic_cell_new((adamic_value){.reference = %s}, true);", work, index+1, retained(value))
	}
	result := e.own(ir.Array, fmt.Sprintf("adamic_parallel_map(%s, %s)", items, work))
	e.closureThrown()
	return result
}
