package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// packCallArguments expands each spread immediately. Its snapshot owns reference
// elements until the statement ends, including across mutations in later arguments.
// The heterogeneous packed array itself borrows those snapshots and scalar temporaries.
func (e *emitter) packCallArguments(arguments []ir.Expression, spread []bool) string {
	packed := e.own(ir.Array, "adamic_array_new(0, false)")
	for index, argument := range arguments {
		if len(spread) > index && spread[index] {
			source := e.value(argument)
			snapshot := e.own(ir.Array, fmt.Sprintf("adamic_array_slice(%s, 0, 0, false)", source))
			position := e.temporary()
			e.line("for (size_t %s = 0; %s < %s->length; %s++) {", position, position, snapshot, position)
			e.line("\tadamic_array_push(%s, %s->elements[%s]);", packed, snapshot, position)
			e.line("}")
		} else {
			value := e.value(argument)
			e.line("adamic_array_push(%s, (adamic_value){.%s = %s});", packed, member(argument.Type()), slotted(argument.Type(), value))
		}
	}
	return packed
}

func packedParameter(of ir.Type, packed, count string, index int) string {
	value := unslotted(of, fmt.Sprintf("%s[%d].%s", packed, index, member(of)))
	if of.IsReference() {
		value = fmt.Sprintf("(%s)%s", cType(of), value)
	}
	return fmt.Sprintf("%s > %d ? %s : %s", count, index, value, missingArgument(of))
}

func (e *emitter) restArray(packed, count string, start int, element ir.Type) string {
	result := e.own(ir.Array, fmt.Sprintf("adamic_array_new(0, %t)", element.IsReference()))
	e.copyRest(result, packed, count, start, element)
	return result
}

func (e *emitter) copyRest(result, packed, count string, start int, element ir.Type) {
	index := e.temporary()
	e.line("for (size_t %s = %d; %s < %s; %s++) {", index, start, index, count, index)
	if element.IsReference() {
		e.line("\tadamic_array_push(%s, (adamic_value){.reference = adamic_retain(%s[%s].reference)});", result, packed, index)
	} else {
		e.line("\tadamic_array_push(%s, %s[%s]);", result, packed, index)
	}
	e.line("}")
}

func (e *emitter) restParameter(parameter int, element ir.Type, start int) {
	name := e.localName(parameter)
	e.line("adamic_array *%s = adamic_array_new(0, %t);", name, element.IsReference())
	e.copyRest(name, "arguments", "argument_count", start, element)
}

func (e *emitter) spreadArguments(call ir.Call) []string {
	function := e.program.Functions[call.Function]
	packed := e.packCallArguments(call.Arguments, call.Spread)
	values := []string{}
	for index, parameter := range function.Parameters {
		if function.RestElement != 0 && index == len(function.Parameters)-1 {
			value := e.restArray(packed+"->elements", packed+"->length", index, function.RestElement)
			if e.reuse.callConsumes(e.program, call, index) {
				value = "adamic_retain(" + value + ")"
			}
			values = append(values, value)
			continue
		}
		of := e.program.Locals[parameter].Type
		value := packedParameter(of, packed+"->elements", packed+"->length", index)
		if e.reuse.callConsumes(e.program, call, index) && of.IsReference() {
			value = "adamic_retain(" + value + ")"
		}
		values = append(values, value)
	}
	if function.ArgumentsCount != 0 {
		count := "(double)" + packed + "->length"
		if function.Receiver {
			count = "(" + count + " - 1)"
		}
		if call.ArgumentCount != nil {
			count = e.value(call.ArgumentCount)
		}
		values = append(values, count)
	}
	return values
}

// A missing reference parameter is undefined, not the empty value used when a
// declared local is initialized before its first assignment.
func missingArgument(of ir.Type) string {
	if of.IsReference() {
		return "NULL"
	}
	return zero(of)
}
