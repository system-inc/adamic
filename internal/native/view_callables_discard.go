package native

import (
	"github.com/system-inc/adamic/internal/ir"
)

// Checked marker calls use the ordinary closure ABI. Producer metadata determines
// whether the returned word owns a reference; the erased view never decides it.
func (e *emitter) emitViewCallableDiscard(property ir.Property, closure string) string {
	e.declarations = append(e.declarations, "#include \"view_callables_contract.h\"")
	recorded := e.temporary()
	e.line("const adamic_callable_signature *%s = NULL;", recorded)
	for index, function := range e.program.Functions {
		if !function.Closure || function.Receiver {
			continue
		}
		parameters := make([]ir.Type, len(function.Parameters))
		for i, local := range function.Parameters {
			parameters[i] = e.program.Locals[local].Type
		}
		signature := e.viewCallableSignature(parameters, viewCallableProducerResult(function.Returns), function.Name)
		e.line("if (%s != NULL && %s->heap.kind == adamic_kind_closure && %s->code == %s) %s = %s;", closure, closure, closure, e.functionName(index), recorded, signature)
	}
	expected := e.viewCallableExpected(property)
	e.line("(void)adamic_view_callable_shape((const adamic_heap *)%s, %s, %s, %s, false);", closure, recorded, expected, cString(property.View))
	result := e.temporary()
	e.line("adamic_value %s = adamic_node_performance_invoke(%s, NULL, 0, false);", result, closure)
	e.closureThrown()
	e.line("if (%s->result == %d || %s->result == %d || %s->result == %d || %s->result == %d || %s->result == %d || %s->result == %d || %s->result == %d) adamic_release(%s.reference);", recorded, ir.String, recorded, ir.Object, recorded, ir.Array, recorded, ir.Map, recorded, ir.Closure, recorded, ir.Weak, recorded, ir.Union, result)
	return "0"
}
