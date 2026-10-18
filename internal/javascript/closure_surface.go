package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) closureSurface(property ir.Property) string {
	var body strings.Builder
	body.WriteString("(function adamicSurface(value) { /* Source surface follows the underlying function. */ value = adamicViewAdapterUnderlying(value); ")
	fmt.Fprintf(&body, "if (adamicViewAdapterRoots.has(value)) return value.code[%s]; ", quote(property.Name))
	body.WriteString("switch (value.code) { ")
	for index, function := range e.program.Functions {
		surface := e.program.FunctionValueSurface(index)
		if !function.Closure || surface == nil {
			continue
		}
		result := fmt.Sprint(surface.Length)
		if property.Name == "name" {
			result = quote(surface.Name)
		}
		if binding := function.Bound; binding != nil {
			source := fmt.Sprintf("adamicSurface(value.cells[%d].value)", e.program.BoundCaptureSlot(function, binding.Source))
			result = fmt.Sprintf("Math.max(0, %s-%d)", source, len(binding.Arguments))
			if property.Name == "name" {
				result = "'bound ' + " + source
			}
		}
		fmt.Fprintf(&body, "case %s: return %s; ", functionName(e.program, index), result)
	}
	body.WriteString("} panic('function surface has no source metadata'); })(")
	body.WriteString(e.value(property.Object))
	body.WriteString(")")
	return body.String()
}
