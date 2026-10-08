package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) graphTypes(types []int) bool { return e.program.IsGraph(types) }

func (e *emitter) adoptGraph(value, bytes string, graph bool) {
	if graph {
		e.line("%s = adamic_graph_adopt_owned(%s, %s);", value, value, bytes)
	}
}

func (e *emitter) adoptGraphObject(value string, literal ir.ObjectLiteral) {
	graph := e.graphTypes(literal.GraphTypes)
	if literal.Class != 0 {
		graph = graph || e.program.Classes[literal.Class-1].Graph
	}
	// The whole object: its slots, readiness and representation bytes and its aligned slot contracts.
	e.adoptGraph(value, fmt.Sprintf("adamic_object_size(%s->shape->count)", value), graph)
}

// The runtime decides boundary ownership from both actual allocations. This
// keeps structural views and unions consistent and skips internal graph counts.
func (e *emitter) heldReferenceIn(holder, value string) string {
	if len(e.program.GraphTypes) == 0 || constantUndefined.MatchString(value) {
		return retained(value)
	}
	return fmt.Sprintf("adamic_graph_hold(%s, %s)", holder, value)
}
func (e *emitter) keptIn(holder, value string) string {
	if len(e.program.GraphTypes) == 0 {
		return e.kept(value)
	}
	return e.heldReferenceIn(holder, value)
}
func (e *emitter) heldIn(holder string, of ir.Type, value string) string {
	if of.IsReference() {
		return fmt.Sprintf("(adamic_value){.reference = %s}", e.heldReferenceIn(holder, value))
	}
	return borrowed(of, value)
}
func (e *emitter) dropIn(holder, value string) {
	if len(e.program.GraphTypes) == 0 {
		e.line("adamic_release(%s);", value)
	} else {
		e.line("adamic_graph_drop(%s, %s);", holder, value)
	}
}
func (e *emitter) graphUnique(value string) string {
	result := uniquelyHeld(value)
	if len(e.program.GraphTypes) != 0 {
		result = fmt.Sprintf("(adamic_graph_counted(%s) && %s)", value, result)
	}
	return result
}

func (e *emitter) graphArray(code string, types []int) string {
	value := e.own(ir.Array, code)
	e.adoptGraph(value, "sizeof *"+value, e.graphTypes(types))
	return value
}
