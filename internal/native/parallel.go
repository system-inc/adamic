package native

import (
	"fmt"
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// parallelMap keeps both operands alive through the runtime's join. The result is owned.
func (e *emitter) parallelMap(expression ir.ParallelMap) string {
	name := "adamic_parallel_map_typed"
	if expression.Moved {
		name = "adamic_parallel_map_move_typed"
	}
	declaration := fmt.Sprintf("adamic_array *%s(adamic_array *items, adamic_closure *work, bool references, const adamic_json_schema *schema);\n", name)
	if !slices.Contains(e.declarations, declaration) {
		e.declarations = append(e.declarations, declaration)
	}
	items := e.value(expression.Items)
	original := e.value(expression.Work)
	if expression.Moved {
		read, ok := expression.Items.(ir.Read)
		if !ok || e.program.Locals[read.Local].Global || e.program.Locals[read.Local].Captured || e.program.Locals[read.Local].Borrowed {
			panic("parallel move without an owned local binding")
		}
		// Operand evaluation has completed. The statement now owns the shell;
		// source scope cleanup sees NULL, including when the join propagates error.
		items = e.own(ir.Array, items)
		e.line("%s = NULL;", e.localName(read.Local))
	}
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
	result := e.own(ir.Array, fmt.Sprintf("%s(%s, %s, %t, %s)", name, items, work, expression.Result.IsReference(), jsonStorage(expression.Result)))
	e.closureThrown()
	return result
}
