package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) arrayCallableCertificate(property ir.Property, value string) string {
	if !ir.ArrayCallableContract(e.program, property.ViewContract) {
		return e.emitViewCallableCertificate(property, value)
	}
	e.declarations = append(e.declarations, "#include \"view_array_callables.h\"")
	recorded := e.temporary()
	e.line("const adamic_callable_signature *%s = NULL;", recorded)
	for _, producer := range e.viewCallableProducers() {
		e.line("if (%s != NULL && %s->heap.kind == adamic_kind_closure && %s->code == %s) %s = %s;", value, value, value, e.functionName(producer.Function), recorded, producer.Signature)
	}
	return fmt.Sprintf("((adamic_closure *)adamic_array_callable_shape((const adamic_heap *)(%s),%s,%s,%s,%t))", value, recorded, e.viewCallableExpected(property), cString(property.View), property.Optional || property.Absent || property.UndefinedAllowed)
}
