package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// closureArguments pads fixed slots and collects rest tails at the caller.
// The actual count travels separately through the selected typed convention.
func (nativeEmitter *emitter) closureArguments(call ir.CallClosure) (string, string) {
	if len(call.Spread) != 0 {
		packed := nativeEmitter.packCallArguments(call.Arguments, call.Spread)
		layout := nativeEmitter.program.ClosureArgumentLayout(call)
		slots := []string{}
		for index, of := range layout.Fixed {
			value := packedParameter(of, packed+"->elements", packed+"->length", index)
			slots = append(slots, fmt.Sprintf("{.%s = %s}", member(of), slotted(of, value)))
		}
		return nativeEmitter.closureSlots(call, slots, packed+"->elements", packed+"->length"), packed + "->length"
	}
	slots := []string{}
	for _, argument := range call.Arguments {
		slots = append(
			slots,
			fmt.Sprintf("{.%s = %s}", member(argument.Type()), slotted(argument.Type(), nativeEmitter.value(argument))),
		)
	}
	return nativeEmitter.closureSlots(call, slots, "", fmt.Sprint(len(call.Arguments))), fmt.Sprint(len(call.Arguments))
}

func (nativeEmitter *emitter) closureSlots(call ir.CallClosure, slots []string, source, count string) string {
	layout := nativeEmitter.program.ClosureArgumentLayout(call)
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
	base := nativeEmitter.program.FixedArgumentSlots
	entries := []string{}
	for index, slot := range slots[:min(base, len(slots))] {
		entries = append(entries, fmt.Sprintf("[%d] = %s", index, slot))
	}
	for _, rest := range layout.Rest {
		array := nativeEmitter.restArray(source, count, rest.Start, rest.Element)
		entries = append(
			entries,
			fmt.Sprintf("[%d] = {.reference = %s}", nativeEmitter.program.RestArgumentSlots[rest], array),
		)
	}

	return "(adamic_value[]){" + strings.Join(entries, ", ") + "}"
}

func (nativeEmitter *emitter) callbackCall(
	callback string,
	expression ir.Expression,
	kind int,
	slots ...string,
) string {
	call := ir.CallClosure{Closure: expression, FunctionType: kind}
	packed := nativeEmitter.closureSlots(call, slots, "", fmt.Sprint(len(slots)))
	return nativeEmitter.packedClosureCall(call, callback, packed, fmt.Sprint(len(slots)))
}

// The runtime's ordinary sort comparator still has main's ABI. A reader, rest
// tail or omitted optional slot gets an adapter only at that particular sort.
func (nativeEmitter *emitter) closureComparator(sort ir.ArraySort) string {
	call := ir.CallClosure{Closure: sort.Callback, FunctionType: sort.CallbackType}
	layout := nativeEmitter.program.ClosureArgumentLayout(call)
	if !layout.Count && len(layout.Rest) == 0 && len(layout.Fixed) <= 2 {
		return "adamic_compare_closure"
	}
	adapter := emitter{program: nativeEmitter.program, reuse: nativeEmitter.reuse, indent: 1, scopes: [][]string{{}}}
	packed := adapter.closureSlots(call, []string{"left", "right"}, "", "2")
	adapter.line("adamic_closure *compare = context;")
	adapter.line("double result = %s.number;", adapter.packedClosureCall(call, "compare", packed, "2"))
	adapter.releaseScopes(0)
	adapter.line("return result < 0 ? -1 : result > 0 ? 1 : 0;")
	name, exists := nativeEmitter.sharedDeclaration("adamic_compare", adapter.out.String())
	if exists {
		return name
	}
	nativeEmitter.declarations = append(
		nativeEmitter.declarations,
		"static int "+name+"(adamic_value left, adamic_value right, void *context) {\n"+adapter.out.String()+"}",
	)
	return name
}

// packCallArguments expands each spread immediately. Its snapshot owns reference
// elements until the statement ends, including across mutations in later arguments.
// The heterogeneous packed array itself borrows those snapshots and scalar temporaries.
func (nativeEmitter *emitter) packCallArguments(arguments []ir.Expression, spread []bool) string {
	packed := nativeEmitter.own(ir.Array, "adamic_array_make(0, 0)")
	for index, argument := range arguments {
		if len(spread) > index && spread[index] {
			source := nativeEmitter.value(argument)
			snapshot := nativeEmitter.own(
				ir.Array,
				fmt.Sprintf("adamic_array_slice(%s, 0, 0, false, %s->element)", source, source),
			)
			position := nativeEmitter.temporary()
			nativeEmitter.line("for (size_t %s = 0; %s < %s->length; %s++) {", position, position, snapshot, position)
			nativeEmitter.line("\tadamic_array_push(%s, %s->elements[%s]);", packed, snapshot, position)
			nativeEmitter.line("}")
		} else {
			value := nativeEmitter.value(argument)
			nativeEmitter.line(
				"adamic_array_push(%s, (adamic_value){.%s = %s});",
				packed,
				member(argument.Type()),
				slotted(argument.Type(), value),
			)
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

func (nativeEmitter *emitter) restArray(packed, count string, start int, element ir.Type) string {
	result := nativeEmitter.own(ir.Array, fmt.Sprintf("adamic_array_make(0, %s)", arrayElement(element)))
	nativeEmitter.copyRest(result, packed, count, start, element)
	return result
}

func (nativeEmitter *emitter) copyRest(result, packed, count string, start int, element ir.Type) {
	index := nativeEmitter.temporary()
	nativeEmitter.line("for (size_t %s = %d; %s < %s; %s++) {", index, start, index, count, index)
	if element.IsReference() {
		nativeEmitter.line(
			"\tadamic_array_push(%s, (adamic_value){.reference = adamic_retain(%s[%s].reference)});",
			result,
			packed,
			index,
		)
	} else {
		nativeEmitter.line("\tadamic_array_push(%s, %s[%s]);", result, packed, index)
	}
	nativeEmitter.line("}")
}

func (nativeEmitter *emitter) spreadArguments(call ir.Call) []string {
	function := nativeEmitter.program.Functions[call.Function]
	packed := nativeEmitter.packCallArguments(call.Arguments, call.Spread)
	values := []string{}
	for index, parameter := range function.Parameters {
		if function.RestElement != 0 && index == len(function.Parameters)-1 {
			value := nativeEmitter.restArray(packed+"->elements", packed+"->length", index, function.RestElement)
			if nativeEmitter.reuse.callConsumes(nativeEmitter.program, call, index) {
				value = "adamic_retain(" + value + ")"
			}
			values = append(values, value)
			continue
		}
		of := nativeEmitter.program.Locals[parameter].Type
		value := packedParameter(of, packed+"->elements", packed+"->length", index)
		if nativeEmitter.reuse.callConsumes(nativeEmitter.program, call, index) && of.IsReference() {
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
			count = nativeEmitter.value(call.ArgumentCount)
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

func (nativeEmitter *emitter) packedClosureCall(call ir.CallClosure, closure, packed, count string) string {
	if nativeEmitter.program.PackedCountNeededFromCall(call) {
		return fmt.Sprintf("adamic_closure_call(%s, %s, %s)", closure, packed, count)
	}
	return fmt.Sprintf("%s->code(%s, %s)", closure, closure, packed)
}

// A receiver is prefixed to every stored slot, including reserved rest tails.
// This length is distinct from the actual count observed by arguments.length.
func (nativeEmitter *emitter) packedArgumentSize(call ir.CallClosure) int {
	layout := nativeEmitter.program.ClosureArgumentLayout(call)
	n := len(layout.Fixed)
	if len(call.Spread) == 0 && len(layout.Rest) == 0 {
		n = max(n, len(call.Arguments))
	}
	for _, rest := range layout.Rest {
		n = max(n, nativeEmitter.program.RestArgumentSlots[rest]+1)
	}
	return n
}
