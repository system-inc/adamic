package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (nativeEmitter *emitter) graphTypes(types []int) bool {
	return nativeEmitter.program.IsGraph(types)
}

func (nativeEmitter *emitter) adoptGraph(value, bytes string, graph bool) {
	if graph {
		nativeEmitter.line("%s = adamic_graph_adopt_owned(%s, %s);", value, value, bytes)
	}
}

func (nativeEmitter *emitter) adoptGraphObject(value string, literal ir.ObjectLiteral) {
	if literal.ProgramRegion {
		return
	}
	graph := nativeEmitter.graphTypes(literal.GraphTypes)
	if literal.Class != 0 {
		graph = graph || nativeEmitter.program.Classes[literal.Class-1].Graph
	}
	nativeEmitter.adoptGraph(value, fmt.Sprintf("adamic_object_allocation_size(%s)", value), graph)
}

// The runtime decides boundary ownership from both actual allocations. This
// keeps structural views and unions consistent and skips internal graph counts.
func (nativeEmitter *emitter) heldReferenceIn(holder, value string) string {
	if len(nativeEmitter.program.GraphTypes) == 0 || constantUndefined.MatchString(value) {
		return retained(value)
	}
	return fmt.Sprintf("adamic_graph_hold(%s, %s)", holder, value)
}
func (nativeEmitter *emitter) keptIn(holder, value string) string {
	if len(nativeEmitter.program.GraphTypes) == 0 {
		return nativeEmitter.kept(value)
	}
	return nativeEmitter.heldReferenceIn(holder, value)
}
func (nativeEmitter *emitter) heldIn(holder string, of ir.Type, value string) string {
	if of.IsReference() {
		return fmt.Sprintf("(adamic_value){.reference = %s}", nativeEmitter.heldReferenceIn(holder, value))
	}
	return borrowed(of, value)
}
func (nativeEmitter *emitter) dropIn(holder, value string) {
	if len(nativeEmitter.program.GraphTypes) == 0 {
		nativeEmitter.line("adamic_release(%s);", value)
	} else {
		nativeEmitter.line("adamic_graph_drop(%s, %s);", holder, value)
	}
}
func (nativeEmitter *emitter) graphUnique(value string) string {
	result := uniquelyHeld(value)
	if len(nativeEmitter.program.GraphTypes) != 0 {
		result = fmt.Sprintf("(adamic_graph_counted(%s) && %s)", value, result)
	}
	return result
}

func (nativeEmitter *emitter) graphArray(code string, types []int, program ...bool) string {
	value := nativeEmitter.own(ir.Array, code)
	nativeEmitter.adoptGraph(
		value,
		"sizeof *"+value,
		nativeEmitter.graphTypes(types) && !(len(program) != 0 && program[0]),
	)
	return value
}
