package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) extractedMethodValue(property ir.Property, object string) string {
	value := e.own(ir.Closure, fmt.Sprintf("adamic_object_method_value(%s, %s, &%s, %t)", object, cString(property.Name), e.cache(), property.Absent))
	if property.Bound != nil {
		receiver := e.value(property.Bound)
		return e.own(ir.Closure, fmt.Sprintf("adamic_method_bind(%s, %s)", value, receiver))
	}
	return value
}

// Spreads are copied at their source position, before later arguments run.
func (e *emitter) callSpreadMethod(expression ir.CallClosure, closure, receiver, method string) string {
	packed := e.own(ir.Array, "adamic_array_new(0, false)")
	holds := e.own(ir.Array, "adamic_array_new(0, true)")
	for index, argument := range expression.Arguments {
		value := e.value(argument)
		if index < len(expression.Spread) && expression.Spread[index] {
			at := e.temporary()
			e.line("for (size_t %s = 0; %s < %s->length; %s++) {", at, at, value, at)
			e.line("\tadamic_array_push(%s, %s->elements[%s]);", packed, value, at)
			e.line("\tif (%s->references) { adamic_array_push(%s, (adamic_value){.reference = adamic_retain(%s->elements[%s].reference)}); }", value, holds, value, at)
			e.line("}")
		} else {
			e.line("adamic_array_push(%s, (adamic_value){.%s = %s});", packed, member(argument.Type()), slotted(argument.Type(), value))
		}
	}
	call := fmt.Sprintf("%s->code(%s, %s->elements, %s->length)", closure, closure, packed, packed)
	if receiver != "" {
		if closure == "" {
			call = fmt.Sprintf("%s(%s, %s->elements, %s->length)", method, receiver, packed, packed)
		} else {
			received, at := e.temporary(), e.temporary()
			e.line("adamic_value %s[%s->length + 1];", received, packed)
			e.line("%s[0].reference = %s;", received, receiver)
			e.line("for (size_t %s = 0; %s < %s->length; %s++) { %s[%s + 1] = %s->elements[%s]; }", at, at, packed, at, received, at, packed, at)
			call = fmt.Sprintf("(%s != NULL ? %s->code(%s, %s->receiver ? %s : %s->elements, %s->length + (%s->receiver ? 1 : 0)) : %s(%s, %s->elements, %s->length))", closure, closure, closure, closure, received, packed, packed, closure, method, receiver, packed, packed)
		}
	}
	result := e.temporary()
	e.line("adamic_value %s = %s;", result, call)
	e.closureThrown()
	if expression.Returns == 0 {
		e.line("(void)%s;", result)
		return "0"
	}
	if expression.Returns.IsReference() {
		return e.own(expression.Returns, fmt.Sprintf("(%s)%s.reference", cType(expression.Returns), result))
	}
	return e.snapshot(expression.Returns, unslotted(expression.Returns, result+"."+member(expression.Returns)))
}
