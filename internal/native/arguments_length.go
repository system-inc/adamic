package native

import (
	"fmt"
	"slices"
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
		return e.closureSlots(call, slots, nil, packed+"->elements", packed+"->length"), packed + "->length"
	}
	slots, of := []string{}, []ir.Type{}
	for _, argument := range call.Arguments {
		slots = append(slots, fmt.Sprintf("{.%s = %s}", member(argument.Type()), slotted(argument.Type(), e.value(argument))))
		of = append(of, argument.Type())
	}
	return e.closureSlots(call, slots, of, "", fmt.Sprint(len(call.Arguments))), fmt.Sprint(len(call.Arguments))
}

// closureSlots is a call's packed slots: the fixed ones, padded, and each rest tail collected from
// source, or from slots when source is "". of is each of slots' type when the caller made them, which
// collecting a rest from them reads (copyRest); nil when source is a pack made at run time, or when
// what slots hold isn't known here.
func (e *emitter) closureSlots(call ir.CallClosure, slots []string, of []ir.Type, source, count string) string {
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
		array := e.restArray(source, count, rest.Start, rest, of)
		entries = append(entries, fmt.Sprintf("[%d] = {.reference = %s}", e.program.RestArgumentSlots[rest], array))
	}

	return "(adamic_value[]){" + strings.Join(entries, ", ") + "}"
}

// callbackCall is a call a loop the emitter writes makes through a callback, slots its arguments
// and of their types: an element's, the index's number, the array's. A pack it makes is the loop
// body's own, so it's made inside an iteration (emit_ownership.go).
func (e *emitter) callbackCall(callback string, expression ir.Expression, kind int, of []ir.Type, slots ...string) string {
	if len(of) != len(slots) {
		panic(fmt.Sprintf("native: compiler bug: a callback call's %d slots come with %d types", len(slots), len(of)))
	}
	call := ir.CallClosure{Closure: expression, FunctionType: kind}
	packed := e.closureSlots(call, slots, of, "", fmt.Sprint(len(slots)))
	return e.packedClosureCall(call, callback, packed, fmt.Sprint(len(slots)))
}

// The runtime's ordinary sort comparator still has main's ABI. A reader, rest
// tail or omitted optional slot gets an adapter only at that particular sort.
func (e *emitter) closureComparator(sort ir.ArraySort) string {
	call := ir.CallClosure{Closure: sort.Callback, FunctionType: sort.CallbackType}
	layout := e.program.ClosureArgumentLayout(call)
	if !layout.Count && len(layout.Rest) == 0 && len(layout.Fixed) <= 2 {
		return "adamic_compare_closure"
	}
	adapter := emitter{program: e.program, reuse: e.reuse, indent: 1, scopes: [][]string{{}}}
	packed := adapter.closureSlots(call, []string{"left", "right"}, []ir.Type{sort.Element, sort.Element}, "", "2")
	adapter.line("adamic_closure *compare = context;")
	adapter.line("double result = %s.number;", adapter.packedClosureCall(call, "compare", packed, "2"))
	// What packing the call made, a rest tail for a comparator that collects one above all, is the
	// adapter's, let go of once the call is done, as timerInvoke does (#kwns2n5).
	adapter.end()
	adapter.releaseScopes(0)
	adapter.line("return result < 0 ? -1 : result > 0 ? 1 : 0;")
	name, exists := e.sharedDeclaration("adamic_compare", adapter.out.String())
	if exists {
		return name
	}
	e.declarations = append(e.declarations, "static int "+name+"(adamic_value left, adamic_value right, void *context) {\n"+adapter.out.String()+"}")
	return name
}

// packCallArguments expands each spread immediately. Its snapshot owns reference
// elements until the statement ends, including across mutations in later arguments.
// The heterogeneous packed array itself borrows those snapshots and scalar temporaries.
func (e *emitter) packCallArguments(arguments []ir.Expression, spread []bool) string {
	packed := e.own(ir.Array, "adamic_array_make(0, 0)")
	for index, argument := range arguments {
		if len(spread) > index && spread[index] {
			source := e.value(argument)
			snapshot := e.own(ir.Array, fmt.Sprintf("adamic_array_slice(%s, 0, 0, false, %s->element)", source, source))
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

// restArray packs a tail for a spread or closure call, copying from start, which counts a
// receiver where the packed source holds one, as rest.Start doesn't. It's made where an array
// literal of the rest's type would be, so in a graph program a pack on a cycle lives in the
// region (#2zhhwaz). of is each packed slot's type, or nil (closureSlots).
func (e *emitter) restArray(packed, count string, start int, rest ir.RestArguments, of []ir.Type) string {
	result := e.graphArray(fmt.Sprintf("adamic_array_make(0, %s)", arrayElement(rest.Element)), e.program.RestGraphTypes[rest])
	e.copyRest(result, packed, count, start, rest.Element, of)
	return result
}

// copyRest pushes the packed slots from start to count onto a rest array of element. Slots the rest
// holds as they are (every one a reference, into a rest of references, or every one the element's
// own type) are copied by a loop, as is a pack made at run time (of nil), whose slots lowering made
// the rest's. A pack whose slots aren't all that, as a callback's (an element, the index's number,
// the array), is pushed slot by slot, each made what the rest holds (restSlot): only a reference is
// retained, since a number's bits retained as a pointer write a count into memory that has none
// (#9grybqc).
func (e *emitter) copyRest(result, packed, count string, start int, element ir.Type, of []ir.Type) {
	if slices.ContainsFunc(of[min(start, len(of)):], func(slot ir.Type) bool { return !restHoldsAlike(element, slot) }) {
		for index := start; index < len(of); index++ {
			e.line("adamic_array_push(%s, %s);", result, e.restSlot(result, fmt.Sprintf("%s[%d]", packed, index), of[index], element))
		}
		return
	}
	index := e.temporary()
	e.line("for (size_t %s = %d; %s < %s; %s++) {", index, start, index, count, index)
	if element.IsReference() {
		e.line("\tadamic_array_push(%s, (adamic_value){.reference = %s});", result, e.heldReferenceIn(result, fmt.Sprintf("%s[%s].reference", packed, index)))
	} else {
		e.line("\tadamic_array_push(%s, %s[%s]);", result, packed, index)
	}
	e.line("}")
}

// restHoldsAlike says a rest of element holds a slot of type slot as the slot is: a reference among
// references, or a scalar of the element's own type.
func restHoldsAlike(element, slot ir.Type) bool {
	return element.IsReference() && slot.IsReference() || element == slot
}

// restSlot is the packed slot of type of as an element of a rest array of element: a reference
// retained, a number or a boolean boxed where the rest is a union's (its count the rest's, as
// lowering's fit boxes an argument to a rest of a union), a scalar made the rest's maybe, and a
// scalar as it is where it's the element's type. Any other slot is one the rest's element doesn't
// admit, which reaches the pack only for a target the call never runs: the layout collects the rest
// of every function its type admits (ir.ClosureArgumentLayout), and lowering refuses to let a rest
// function be seen through a type that hides its rest (restHiddenByView), so the function that runs
// is never one whose rest this is. It's undefined there, a word no count is taken through.
func (e *emitter) restSlot(result, slot string, of, element ir.Type) string {
	value := unslotted(of, slot+"."+member(of))
	switch {
	case of.IsReference() && element.IsReference():
		return fmt.Sprintf("(adamic_value){.reference = %s}", e.heldReferenceIn(result, value))
	case of == element:
		return slot
	case element == ir.Union && (of == ir.Number || of == ir.Boolean || of.IsMaybe()):
		boxed, fresh := converted(of, ir.Union, value)
		if !fresh {
			// A boolean's box is a constant one, retained as any reference the rest holds.
			boxed = e.heldReferenceIn(result, boxed)
		}
		return fmt.Sprintf("(adamic_value){.reference = %s}", boxed)
	case element.IsMaybe() && of == element.Present():
		return fmt.Sprintf("(adamic_value){.%s = %s}", member(element), slotted(element, maybe(element, value)))
	}
	// Undefined, never the slot's own value: no function that reads this rest runs with this pack.
	// Lowering holds a type that admits a number or a boolean as a union (lower's representation:
	// unknown, object and {} among them), so a scalar reaches here only into a rest whose element's
	// type doesn't admit it, or holding undefined itself, where undefined is the value. The function
	// that runs is the closure in the callback's or callee's slot, and a rest function reaches a slot
	// whose type hides its rest only past restHiddenByView, which stops it as NotYet; where the type
	// shows the rest, TypeScript's strictFunctionTypes checked each argument against its element.
	// So this rest is read only by a target the call never runs, as q1's collect, which its layout
	// collects only because its type is assignable to map's callback (#9grybqc).
	return fmt.Sprintf("(adamic_value){.%s = %s}", member(element), slotted(element, missingArgument(element)))
}

func (e *emitter) spreadArguments(call ir.Call) []string {
	function := e.program.Functions[call.Function]
	packed := e.packCallArguments(call.Arguments, call.Spread)
	values := []string{}
	for index, parameter := range function.Parameters {
		if function.RestElement != 0 && index == len(function.Parameters)-1 {
			value := e.restArray(packed+"->elements", packed+"->length", index, ir.FunctionRestArguments(function), nil)
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
	if e.program.PackedCountNeededFromCall(call) {
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
