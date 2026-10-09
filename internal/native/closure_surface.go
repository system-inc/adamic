package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) closureSurface(property ir.Property) string {
	name := e.temporary()
	var body strings.Builder
	fmt.Fprintf(&body, "static %s %s(adamic_closure *value) {\n", cType(property.Of), name)
	body.WriteString("/* Source surface follows the underlying function, not the checking wrapper. */\nvalue = adamic_view_adapter_underlying(value);\n")
	if property.Name == "name" {
		body.WriteString("static adamic_string adapter_name = ADAMIC_STRING(\"\");\n")
		body.WriteString("if (value != NULL && value->heap.references != 0 && value->view != NULL) return &adapter_name;\n")
	} else {
		body.WriteString("if (value != NULL && value->heap.references != 0 && value->view != NULL) return 2.0;\n")
	}
	for index, function := range e.program.Functions {
		surface := e.program.FunctionValueSurface(index)
		if !function.Closure || surface == nil {
			continue
		}
		identity := e.unionClosureCodeIdentity("value", index)
		fmt.Fprintf(&body, "if (value != NULL && value->heap.references != 0 && %s) {\n", identity)
		if property.Name == "name" {
			fmt.Fprintf(&body, "static adamic_string source_name = ADAMIC_STRING(%s); return &source_name;\n", cString(surface.Name))
		} else {
			fmt.Fprintf(&body, "return %d.0;\n", surface.Length)
		}
		body.WriteString("}\n")
	}
	body.WriteString("static const char message[] = \"function surface has no source metadata\"; adamic_panic(message, sizeof message - 1);\n}\n")
	e.declarations = append(e.declarations, body.String())
	value := fmt.Sprintf("%s(%s)", name, e.value(property.Object))
	if property.Of == ir.String {
		return e.own(ir.String, "adamic_retain("+value+")")
	}
	return e.snapshot(ir.Number, value)
}
