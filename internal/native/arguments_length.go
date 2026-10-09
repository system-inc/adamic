package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// closureArguments pads fixed slots and collects rest tails at the caller.
// The actual count travels separately through the selected typed convention.
func (e *emitter) closureArguments(call ir.CallClosure) (string, string) {
	if len(call.Spread) != 0 {
		packed := e.packCallArguments(call.Arguments, call.Spread)
		layout := e.program.ClosureArgumentLayout(call)
		slots := []string{}
		for index, of := range layout.Fixed {
			value := packedParameter(of, packed+"->elements", packed+"->length", index)
			slots = append(slots, fmt.Sprintf("{.%s = %s}", member(of), slotted(of, value)))
		}
		return e.closureSlots(call, slots, packed+"->elements", packed+"->length"), packed + "->length"
	}
	slots := []string{}
	for _, argument := range call.Arguments {
		slots = append(slots, fmt.Sprintf("{.%s = %s}", member(argument.Type()), slotted(argument.Type(), e.value(argument))))
	}
	return e.closureSlots(call, slots, "", fmt.Sprint(len(call.Arguments))), fmt.Sprint(len(call.Arguments))
}

func (e *emitter) closureSlots(call ir.CallClosure, slots []string, source, count string) string {
	layout := e.program.ClosureArgumentLayout(call)
	if source == "" && len(slots) > 0 {
		source = "(adamic_value[]){" + strings.Join(slots, ", ") + "}"
	}
	if source == "" {
		source = "((adamic_value *)NULL)"
	}
	for index := len(slots); index < len(layout.Fixed); index++ {
		of := layout.Fixed[index]
		slots = append(slots, fmt.Sprintf("{.%s = %s}", member(of), slotted(of, missingArgument(of))))
	}
	if len(layout.Rest) == 0 {
		if len(slots) == 0 {
			return "NULL"
		}
		return "(adamic_value[]){" + strings.Join(slots, ", ") + "}"
	}
	// Extra actual arguments are evaluated, but only parameter slots are kept.
	// The original source remains available while collecting each rest tail.
	base := e.program.FixedArgumentSlots
	entries := []string{}
	for index, slot := range slots[:min(base, len(slots))] {
		entries = append(entries, fmt.Sprintf("[%d] = %s", index, slot))
	}
	for _, rest := range layout.Rest {
		array := e.restArray(source, count, rest.Start, rest.Element)
		entries = append(entries, fmt.Sprintf("[%d] = {.reference = %s}", e.program.RestArgumentSlots[rest], array))
	}

	return "(adamic_value[]){" + strings.Join(entries, ", ") + "}"
}

func (e *emitter) callbackCall(callback string, expression ir.Expression, kind int, slots ...string) string {
	call := ir.CallClosure{Closure: expression, FunctionType: kind}
	packed := e.closureSlots(call, slots, "", fmt.Sprint(len(slots)))
	return e.packedClosureCall(call, callback, packed, fmt.Sprint(len(slots)))
}

// The runtime's ordinary sort comparator still has main's ABI. A reader, rest
// tail or omitted optional slot gets an adapter only at that particular sort.
func (e *emitter) closureComparator(sort ir.ArraySort) string {
	call := ir.CallClosure{Closure: sort.Callback, FunctionType: sort.CallbackType}
	layout := e.program.ClosureArgumentLayout(call)
	if !e.program.ViewAdapters && !layout.Count && len(layout.Rest) == 0 && len(layout.Fixed) <= 2 {
		return "adamic_compare_closure"
	}
	e.temporaries++
	name := fmt.Sprintf("adamic_compare_%d", e.temporaries)
	adapter := emitter{program: e.program, reuse: e.reuse, indent: 1, scopes: [][]string{{}}}
	packed := adapter.closureSlots(call, []string{"left", "right"}, "", "2")
	adapter.line("adamic_closure *compare = context;")
	adapter.line("double result = %s.number;", adapter.packedClosureCall(call, "compare", packed, "2"))
	adapter.releaseScopes(0)
	adapter.line("return result < 0 ? -1 : result > 0 ? 1 : 0;")
	e.declarations = append(e.declarations, "static int "+name+"(adamic_value left, adamic_value right, void *context) {\n"+adapter.out.String()+"}")
	return name
}

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

func (e *emitter) packedClosureCall(call ir.CallClosure, closure, packed, count string) string {
	if e.program.ClosureReceiversNeeded() {
		return fmt.Sprintf("adamic_closure_receiver_call(%s, NULL, %s, %s, %d)", closure, packed, count, e.packedArgumentSize(call))
	}
	if e.program.PackedCountNeededFromCall(call) || e.program.ClosureConventionNeeded() {
		return fmt.Sprintf("adamic_closure_call(%s, %s, %s)", closure, packed, count)
	}
	return fmt.Sprintf("%s->code(%s, %s)", closure, closure, packed)
}

// A receiver is prefixed to every stored slot, including reserved rest tails.
// This length is distinct from the actual count observed by arguments.length.
func (e *emitter) packedArgumentSize(call ir.CallClosure) int {
	layout := e.program.ClosureArgumentLayout(call)
	n := len(layout.Fixed)
	if len(call.Spread) == 0 && len(layout.Rest) == 0 {
		n = max(n, len(call.Arguments))
	}
	for _, rest := range layout.Rest {
		n = max(n, e.program.RestArgumentSlots[rest]+1)
	}
	return n
}
