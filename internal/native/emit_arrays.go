// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

// arrayVisit emits forEach, filter, some, every, find and findIndex as one loop. The length is read
// once and an index the array has lost is skipped, as JavaScript does (0.1's arrays have no holes, so
// a lost index is one past a shrunken end). Each element is held across its call, since the callback
// may take it out of the array; filter and find hand that hold to what they return.
func (e *emitter) arrayVisit(visit ir.ArrayVisit) string {
	source := e.temporary()
	e.line("adamic_array *%s = %s;", source, e.value(visit.Array))
	callback := e.value(visit.Callback)
	references := visit.Element.IsReference()
	result := "0"
	switch visit.Method {
	case "filter":
		result = e.own(ir.Array, fmt.Sprintf("adamic_array_new(0, %t)", references))
	case "some", "every":
		result = e.snapshot(ir.Boolean, strconv.FormatBool(visit.Method == "every"))
	case "findIndex", "findLastIndex":
		result = e.snapshot(ir.Number, "-1.0")
	case "find", "findLast":
		if visit.Type().IsMaybe() {
			result = e.snapshot(visit.Type(), zero(visit.Type()))
		} else {
			result = e.own(visit.Element, "NULL")
		}
	}
	count, index, element, answer := e.temporary(), e.temporary(), e.temporary(), e.temporary()
	e.line("size_t %s = %s->length;", count, source)
	if visit.Method == "findLast" || visit.Method == "findLastIndex" {
		e.line("for (size_t %s = %s; %s-- > 0;) {", index, count, index)
	} else {
		e.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
	}
	e.indent++
	e.line("if (%s >= %s->length) {", index, source)
	if visit.Method == "find" || visit.Method == "findIndex" || visit.Method == "findLast" || visit.Method == "findLastIndex" {
		// The other visits skip an index the callback took away, as JavaScript's do; find and
		// findIndex call it with undefined there, which the element's type can't hold. A panic, the
		// same in both backends.
		e.line("\tstatic const char message[] = \"%s: the array shrank while it was being searched\";", visit.Method)
		e.line("\tadamic_panic(message, sizeof message - 1);")
	} else {
		e.line("\tcontinue;")
	}
	e.line("}")
	e.line("adamic_value %s = %s->elements[%s];", element, source, index)
	if references {
		e.line("adamic_retain(%s.reference);", element)
	}
	argument := e.arrayCallbackSlot(source, element, visit.Callback, visit.CallbackType, 0, visit.Element)
	call := e.callbackCall(callback, visit.Callback, visit.CallbackType, argument, fmt.Sprintf("{.number = (double)%s}", index), fmt.Sprintf("{.reference = %s}", source))
	if visit.Method == "forEach" && !visit.Returns.IsReference() {
		e.line("%s;", call)
	} else {
		e.line("adamic_value %s = %s;", answer, call)
	}
	// The element held across the call is let go; what the visit made so far is the statement's.
	if references {
		e.closureThrown(element + ".reference")
	} else {
		e.closureThrown()
	}
	release := func() {
		if references {
			e.line("adamic_release(%s.reference);", element)
		}
	}
	switch visit.Method {
	case "forEach":
		if visit.Returns.IsReference() {
			// A callback's result comes back owned, and forEach has no use for it.
			e.line("adamic_release(%s.reference);", answer)
		}
		release()
	case "filter":
		e.line("if (%s.boolean) {", answer)
		e.line("\tadamic_array_push(%s, %s);", result, element)
		if references {
			e.line("} else {")
			e.line("\tadamic_release(%s.reference);", element)
		}
		e.line("}")
	case "find", "findLast":
		e.line("if (%s.boolean) {", answer)
		if visit.Type().IsMaybe() {
			found := element + "." + member(visit.Element)
			if visit.Element == ir.MaybeNumber {
				found = unslotted(ir.MaybeNumber, found)
			} else {
				found = maybe(visit.Type(), found)
			}
			e.line("\t%s = %s;", result, found)
		} else {
			e.line("\t%s = %s.reference;", result, element)
		}
		e.line("\tbreak;")
		e.line("}")
		release()
	default:
		release()
		found, value := answer+".boolean", "true"
		switch visit.Method {
		case "every":
			found, value = "!"+found, "false"
		case "findIndex", "findLastIndex":
			value = "(double)" + index
		}
		e.line("if (%s) {", found)
		e.line("\t%s = %s;", result, value)
		e.line("\tbreak;")
		e.line("}")
	}
	e.indent--
	e.line("}")
	return result
}

// arrayReduce emits reduce as arrayVisit's loop, carrying what the callback last returned. The callee
// holds its own copy of each argument, so after each call the old accumulator is let go and the new
// one, which comes back owned, is taken.
func (e *emitter) arrayReduce(reduce ir.ArrayReduce) string {
	source := e.temporary()
	e.line("adamic_array *%s = %s;", source, e.value(reduce.Array))
	callback := e.value(reduce.Callback)
	initial := e.value(reduce.Initial)
	var accumulator string
	if reduce.Result.IsReference() {
		accumulator = e.own(reduce.Result, fmt.Sprintf("adamic_retain(%s)", initial))
	} else {
		accumulator = e.snapshot(reduce.Result, initial)
	}
	count, index, element, answer := e.temporary(), e.temporary(), e.temporary(), e.temporary()
	e.line("size_t %s = %s->length;", count, source)
	e.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
	e.indent++
	e.line("if (%s >= %s->length) {", index, source)
	e.line("\tcontinue;")
	e.line("}")
	e.line("adamic_value %s = %s->elements[%s];", element, source, index)
	if reduce.Element.IsReference() {
		e.line("adamic_retain(%s.reference);", element)
	}
	argument := e.arrayCallbackSlot(source, element, reduce.Callback, reduce.CallbackType, 1, reduce.Element)
	call := e.callbackCall(callback, reduce.Callback, reduce.CallbackType, fmt.Sprintf("{.%s = %s}", member(reduce.Result), slotted(reduce.Result, accumulator)), argument, fmt.Sprintf("{.number = (double)%s}", index), fmt.Sprintf("{.reference = %s}", source))
	e.line("adamic_value %s = %s;", answer, call)
	// The element held across the call is let go; the accumulator is the statement's.
	if reduce.Element.IsReference() {
		e.closureThrown(element + ".reference")
	} else {
		e.closureThrown()
	}
	if reduce.Element.IsReference() {
		e.line("adamic_release(%s.reference);", element)
	}
	if reduce.Result.IsReference() {
		e.line("adamic_release(%s);", accumulator)
		e.line("%s = (%s)%s.reference;", accumulator, cType(reduce.Result), answer)
	} else {
		e.line("%s = %s;", accumulator, unslotted(reduce.Result, answer+"."+member(reduce.Result)))
	}
	e.indent--
	e.line("}")
	return accumulator
}

// equality tells the runtime how indexOf and includes compare elements.
func equality(element ir.Type) string {
	switch element {
	case ir.Number:
		return "adamic_equal_numbers"
	case ir.Boolean:
		return "adamic_equal_booleans"
	case ir.String:
		return "adamic_equal_strings"
	case ir.Union:
		return "adamic_equal_unions"
	case ir.MaybeNumber:
		return "adamic_equal_maybe_numbers"
	}
	return "adamic_equal_identity"
}

// joinKind tells the runtime how join writes an element.
func joinKind(element ir.Type) string {
	switch element {
	case ir.Number:
		return "adamic_join_numbers"
	case ir.Boolean:
		return "adamic_join_booleans"
	case ir.MaybeNumber:
		return "adamic_join_maybe_numbers"
	}
	return "adamic_join_strings"
}

// comparator declares, once per sort, an adapter from the runtime's comparison to the program's
// comparator: JavaScript reads its result's sign, and NaN as 0.
func (e *emitter) comparator(sort ir.ArraySort) string {
	e.temporaries++
	name := fmt.Sprintf("adamic_compare_%d", e.temporaries)
	argument := unslotted(sort.Element, "left."+member(sort.Element)) + ", " + unslotted(sort.Element, "right."+member(sort.Element))
	if e.program.Functions[sort.Comparator].ArgumentsCount != 0 {
		argument += ", 2"
	}
	e.declarations = append(e.declarations, fmt.Sprintf(
		"static int %s(adamic_value left, adamic_value right, void *context) {\n\t(void)context;\n\tdouble result = %s(%s);\n\treturn result < 0 ? -1 : result > 0 ? 1 : 0;\n}",
		name, e.functionName(sort.Comparator), argument))
	return name
}

// spliceArguments evaluates a splice's operands, in order, as the runtime's splice takes them.
func (e *emitter) spliceArguments(splice ir.ArraySplice) string {
	array := e.value(splice.Array)
	start := e.value(splice.Start)
	count := "0.0"
	if splice.Count != nil {
		count = e.value(splice.Count)
	}
	items := []string{}
	for _, item := range splice.Items {
		items = append(items, held(splice.Element, e.value(item)))
	}
	packed := "NULL"
	if len(items) > 0 {
		packed = "(adamic_value[]){" + strings.Join(items, ", ") + "}"
	}
	return fmt.Sprintf("%s, %s, %s, %t, %d, %s", array, start, count, splice.Count != nil, len(items), packed)
}
