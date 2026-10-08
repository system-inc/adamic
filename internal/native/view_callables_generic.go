package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) emitGenericCallableCertificate(property ir.Property, value, expected string) string {
	tests := []string{}
	for _, function := range e.program.ViewContracts[property.ViewContract-1].Functions {
		tests = append(tests, fmt.Sprintf("%s->code == %s", value, e.functionName(function)))
	}
	known := "false"
	if len(tests) != 0 {
		known = strings.Join(tests, " || ")
	}
	return fmt.Sprintf("(%s != NULL && %s->heap.kind == adamic_kind_closure && (%s) ? %s : %s)", value, value, known, value, emitViewCallableShape(value, "NULL", expected, property.View, false))
}

func (e *emitter) genericCallableInvoke(call ir.CallClosure) string {
	name := e.temporary()
	var body strings.Builder
	fmt.Fprintf(&body, "static adamic_value %s(adamic_closure *self, adamic_value *arguments, size_t count, bool discard) { (void)discard;\n", name)
	for _, pair := range call.GenericInstances {
		if pair[0] >= 0 {
			fmt.Fprintf(&body, "if (self != NULL && self->code == %s) return %s(self, arguments, count);\n", e.functionName(pair[0]), e.functionName(pair[1]))
		}
	}
	body.WriteString("(void)self; (void)arguments; (void)count; static const char message[] = \"generic callable: no compatible producer instantiation\"; adamic_panic(message, sizeof message - 1);\n}\n")
	e.declarations = append(e.declarations, body.String())
	return name
}
