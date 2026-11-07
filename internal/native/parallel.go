package native

import (
	"fmt"
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// parallelMap keeps both operands alive through the runtime's join. The result is owned.
func (nativeEmitter *emitter) parallelMap(expression ir.ParallelMap) string {
	name := "adamic_parallel_map"
	if expression.Moved {
		name = "adamic_parallel_map_move"
	}
	// Keep the ordinary program's explicit ABI declaration. Async programs use
	// adamic.h's declaration so split units need no external prototype ownership.
	if expression.Moved || !nativeEmitter.program.HasPromises() {
		declaration := fmt.Sprintf(
			"adamic_array *%s(adamic_array *items, adamic_closure *work, enum adamic_ir_type element);\n",
			name,
		)
		if !slices.Contains(nativeEmitter.declarations, declaration) {
			nativeEmitter.declarations = append(nativeEmitter.declarations, declaration)
		}
	}
	items := nativeEmitter.value(expression.Items)
	original := nativeEmitter.value(expression.Work)
	if expression.Moved {
		read, ok := expression.Items.(ir.Read)
		if !ok ||
			nativeEmitter.program.Locals[read.Local].Global ||
			nativeEmitter.program.Locals[read.Local].Captured ||
			nativeEmitter.program.Locals[read.Local].Borrowed {
			panic("parallel move without an owned local binding")
		}
		// Operand evaluation has completed. The statement now owns the shell;
		// source scope cleanup sees NULL, including when the join propagates error.
		items = nativeEmitter.own(ir.Array, items)
		nativeEmitter.line("%s = NULL;", nativeEmitter.localName(read.Local))
	}
	// A fresh adapter carries the original closure plus reference-valued globals.
	// Calling the original code with its original self preserves function identity.
	adapter, exists := nativeEmitter.sharedDeclaration("adamic_parallel_adapter", "closure_call/2")
	if !exists {
		nativeEmitter.declarations = append(nativeEmitter.declarations, fmt.Sprintf(
			"static adamic_value %s(adamic_closure *self, adamic_value *arguments) {\n\tadamic_closure *work = self->cells[0]->value.reference;\n\treturn adamic_closure_call(work, arguments, 2);\n}\n",
			adapter,
		))
	}
	work := nativeEmitter.own(ir.Closure, fmt.Sprintf("adamic_closure_new(%s, %d)", adapter, len(expression.Shared)+1))
	nativeEmitter.line(
		"%s->cells[0] = adamic_cell_new((adamic_value){.reference = %s}, true);",
		work,
		retained(original),
	)
	for index, global := range expression.Shared {
		value := nativeEmitter.value(global)
		nativeEmitter.line(
			"%s->cells[%d] = adamic_cell_new((adamic_value){.reference = %s}, true);",
			work,
			index+1,
			retained(value),
		)
	}
	result := nativeEmitter.own(
		ir.Array,
		fmt.Sprintf("%s(%s, %s, %s)", name, items, work, arrayElement(expression.Result)),
	)
	nativeEmitter.closureThrown()
	return result
}
