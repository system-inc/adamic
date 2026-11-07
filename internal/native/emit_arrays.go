// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"slices"
	"strconv"
	"strings"
)

// arrayVisit reads the length once. Searches Get removed indices as undefined,
// checked at the callback call; the other visits skip removed indices.
// Each element is held across its call, since the callback
// may take it out of the array; filter and find hand that hold to what they return.
func (nativeEmitter *emitter) arrayVisit(visit ir.ArrayVisit) string {
	source := nativeEmitter.temporary()
	nativeEmitter.line("adamic_array *%s = %s;", source, nativeEmitter.value(visit.Array))
	callback := nativeEmitter.value(visit.Callback)
	references := visit.Element.IsReference()
	result := "0"
	switch visit.Method {
	case "filter":
		result = nativeEmitter.graphArray(
			fmt.Sprintf("adamic_array_make(0, %s)", arrayElement(visit.Element)),
			visit.GraphTypes,
			visit.ProgramRegion,
		)
		nativeEmitter.programArray(result, visit.ProgramRegion)
	case "some", "every":
		result = nativeEmitter.snapshot(ir.Boolean, strconv.FormatBool(visit.Method == "every"))
	case "findIndex", "findLastIndex":
		result = nativeEmitter.snapshot(ir.Number, "-1.0")
	case "find", "findLast":
		if visit.Type().IsMaybe() {
			result = nativeEmitter.snapshot(visit.Type(), zero(visit.Type()))
		} else {
			result = nativeEmitter.own(visit.Element, "NULL")
		}
	}
	count, index, element, answer := nativeEmitter.temporary(), nativeEmitter.temporary(), nativeEmitter.temporary(), nativeEmitter.temporary()
	nativeEmitter.line("size_t %s = %s->length;", count, source)
	if visit.Method == "findLast" || visit.Method == "findLastIndex" {
		nativeEmitter.line("for (size_t %s = %s; %s-- > 0;) {", index, count, index)
	} else {
		nativeEmitter.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
	}
	nativeEmitter.indent++
	search := visit.Method == "find" ||
		visit.Method == "findIndex" ||
		visit.Method == "findLast" ||
		visit.Method == "findLastIndex"
	present := "true"
	if search {
		present = nativeEmitter.temporary()
		nativeEmitter.line("bool %s = %s < %s->length;", present, index, source)
		if !visit.SearchUndefined {
			nativeEmitter.declarations = append(nativeEmitter.declarations, "#include <stdio.h>")
			nativeEmitter.line("if (!(%s)) {", present)
			nativeEmitter.line("\tchar message[sizeof %s + 128];", cString(visit.SearchElementName))
			nativeEmitter.line(
				"\tint length = snprintf(message, sizeof message, %s, %s);",
				cString(visit.Method+": index %zu is undefined; element type %s does not admit undefined"),
				index+", "+cString(visit.SearchElementName),
			)
			nativeEmitter.line(
				"\tadamic_panic(message, (size_t)length < sizeof message ? (size_t)length : sizeof message - 1);",
			)
			nativeEmitter.line("}")
		}
	} else {
		nativeEmitter.line("if (%s >= %s->length) continue;", index, source)
	}
	missing := "{.reference = NULL}"
	if visit.Element.IsMaybe() {
		missing = fmt.Sprintf("{.%s = %s}", member(visit.Element), slotted(visit.Element, zero(visit.Element)))
	}
	nativeEmitter.line(
		"adamic_value %s = (%s) ? %s->elements[%s] : (adamic_value)%s;",
		element,
		present,
		source,
		index,
		missing,
	)
	if references {
		nativeEmitter.line("adamic_retain(%s.reference);", element)
	}
	argument, fresh := element, false
	if search && visit.SearchFirst != 0 && visit.SearchFirst != visit.Element {
		value, allocated := converted(
			visit.Element,
			visit.SearchFirst,
			unslotted(visit.Element, element+"."+member(visit.Element)),
		)
		if visit.SearchNull {
			value = fmt.Sprintf("(%s.reference != NULL ? %s : &adamic_null)", element, value)
		}
		fresh = allocated
		argument = nativeEmitter.temporary()
		absent := missingArgument(visit.SearchFirst)
		nativeEmitter.line(
			"adamic_value %s = {.%s = %s};",
			argument,
			member(visit.SearchFirst),
			slotted(visit.SearchFirst, fmt.Sprintf("(%s) ? %s : %s", present, value, absent)),
		)
	}
	call := nativeEmitter.callbackCall(
		callback,
		visit.Callback,
		visit.CallbackType,
		argument,
		fmt.Sprintf("{.number = (double)%s}", index),
		fmt.Sprintf("{.reference = %s}", source),
	)
	// A callback typed void may still return a value, which forEach lets go of as any void call does.
	discard := ir.DiscardNothing
	if visit.Method == "forEach" && visit.Returns == 0 {
		discard = nativeEmitter.program.ClosureDiscard(visit)
	}
	if visit.Method == "forEach" && !visit.Returns.IsReference() && !nativeEmitter.discarding(discard) {
		nativeEmitter.line("%s;", call)
	} else {
		nativeEmitter.line("adamic_value %s = %s;", answer, call)
	}
	if fresh {
		nativeEmitter.line("adamic_release(%s.reference);", argument)
	}
	// The element held across the call is let go; what the visit made so far is the statement's.
	if references {
		nativeEmitter.closureThrown(element + ".reference")
	} else {
		nativeEmitter.closureThrown()
	}
	release := func() {
		if references {
			nativeEmitter.line("adamic_release(%s.reference);", element)
		}
	}
	switch visit.Method {
	case "forEach":
		if visit.Returns.IsReference() {
			// A callback's result comes back owned, and forEach has no use for it.
			nativeEmitter.line("adamic_release(%s.reference);", answer)
		}
		if nativeEmitter.discarding(discard) {
			nativeEmitter.line("%s", nativeEmitter.discard(discard, callback, answer))
		}
		release()
	case "filter":
		nativeEmitter.line("if (%s.boolean) {", answer)
		if references && nativeEmitter.graphTypes(visit.GraphTypes) {
			nativeEmitter.line("\t%s.reference = adamic_graph_take(%s, %s.reference);", element, result, element)
		}
		nativeEmitter.line("\tadamic_array_push(%s, %s);", result, element)
		if references {
			nativeEmitter.line("} else {")
			nativeEmitter.line("\tadamic_release(%s.reference);", element)
		}
		nativeEmitter.line("}")
	case "find", "findLast":
		nativeEmitter.line("if (%s.boolean) {", answer)
		if visit.Type().IsMaybe() {
			found := element + "." + member(visit.Element)
			if visit.Element.IsMaybe() {
				found = unslotted(visit.Element, found)
			} else {
				found = maybe(visit.Type(), found)
			}
			nativeEmitter.line("\t%s = (%s) ? %s : %s;", result, present, found, zero(visit.Type()))
		} else {
			nativeEmitter.line("\t%s = %s.reference;", result, element)
		}
		nativeEmitter.line("\tbreak;")
		nativeEmitter.line("}")
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
		nativeEmitter.line("if (%s) {", found)
		nativeEmitter.line("\t%s = %s;", result, value)
		nativeEmitter.line("\tbreak;")
		nativeEmitter.line("}")
	}
	nativeEmitter.indent--
	nativeEmitter.line("}")
	return result
}

// arrayReduce emits reduce as arrayVisit's loop, carrying what the callback last returned. The callee
// holds its own copy of each argument, so after each call the old accumulator is let go and the new
// one, which comes back owned, is taken.
func (nativeEmitter *emitter) arrayReduce(reduce ir.ArrayReduce) string {
	source := nativeEmitter.temporary()
	nativeEmitter.line("adamic_array *%s = %s;", source, nativeEmitter.value(reduce.Array))
	callback := nativeEmitter.value(reduce.Callback)
	initial := nativeEmitter.value(reduce.Initial)
	var accumulator string
	if reduce.Result.IsReference() {
		accumulator = nativeEmitter.own(reduce.Result, fmt.Sprintf("adamic_retain(%s)", initial))
	} else {
		accumulator = nativeEmitter.snapshot(reduce.Result, initial)
	}
	count, index, element, answer := nativeEmitter.temporary(), nativeEmitter.temporary(), nativeEmitter.temporary(), nativeEmitter.temporary()
	nativeEmitter.line("size_t %s = %s->length;", count, source)
	nativeEmitter.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
	nativeEmitter.indent++
	nativeEmitter.line("if (%s >= %s->length) {", index, source)
	nativeEmitter.line("\tcontinue;")
	nativeEmitter.line("}")
	nativeEmitter.line("adamic_value %s = %s->elements[%s];", element, source, index)
	if reduce.Element.IsReference() {
		nativeEmitter.line("adamic_retain(%s.reference);", element)
	}
	call := nativeEmitter.callbackCall(
		callback,
		reduce.Callback,
		reduce.CallbackType,
		fmt.Sprintf("{.%s = %s}", member(reduce.Result), slotted(reduce.Result, accumulator)),
		element,
		fmt.Sprintf("{.number = (double)%s}", index),
		fmt.Sprintf("{.reference = %s}", source),
	)
	nativeEmitter.line("adamic_value %s = %s;", answer, call)
	// The element held across the call is let go; the accumulator is the statement's.
	if reduce.Element.IsReference() {
		nativeEmitter.closureThrown(element + ".reference")
	} else {
		nativeEmitter.closureThrown()
	}
	if reduce.Element.IsReference() {
		nativeEmitter.line("adamic_release(%s.reference);", element)
	}
	if reduce.Result.IsReference() {
		nativeEmitter.line("adamic_release(%s);", accumulator)
		nativeEmitter.line("%s = (%s)%s.reference;", accumulator, cType(reduce.Result), answer)
	} else {
		nativeEmitter.line("%s = %s;", accumulator, unslotted(reduce.Result, answer+"."+member(reduce.Result)))
	}
	nativeEmitter.indent--
	nativeEmitter.line("}")
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

// arrayElement is the element kind an array of these elements is made with (adamic.h's adamic_array),
// as the C name of its enum adamic_ir_type: a scalar names itself, and every reference is Union, since
// each element names its own kind by its heap header. An array whose kind isn't known is made with 0,
// written where it's made; here a missing element type is a compiler bug.
func arrayElement(element ir.Type) string {
	if element.IsReference() {
		element = ir.Union
	}
	for name, tag := range ir.TypeTags() {
		if tag == element {
			return "adamic_ir_" + name
		}
	}
	panic(fmt.Sprintf("native: an array made with no element type (%d)", element))
}

// comparator declares, once per sort, an adapter from the runtime's comparison to the program's
// comparator: JavaScript reads its result's sign, and NaN as 0.
func (nativeEmitter *emitter) comparator(sort ir.ArraySort) string {
	name, exists := nativeEmitter.sharedDeclaration(
		"adamic_compare",
		nativeEmitter.functionName(sort.Comparator)+fmt.Sprint(sort.Element),
	)
	if exists {
		return name
	}
	argument := unslotted(sort.Element, "left."+member(sort.Element)) +
		", " +
		unslotted(sort.Element, "right."+member(sort.Element))
	if nativeEmitter.program.Functions[sort.Comparator].ArgumentsCount != 0 {
		argument += ", 2"
	}
	nativeEmitter.declarations = append(nativeEmitter.declarations, fmt.Sprintf(
		"static int %s(adamic_value left, adamic_value right, void *context) {\n\t(void)context;\n\tdouble result = %s(%s);\n\treturn result < 0 ? -1 : result > 0 ? 1 : 0;\n}",
		name, nativeEmitter.functionName(sort.Comparator), argument))
	return name
}

// spliceArguments evaluates a splice's operands, in order, as the runtime's splice takes them.
func (nativeEmitter *emitter) spliceArguments(splice ir.ArraySplice) string {
	array := nativeEmitter.value(splice.Array)
	start := nativeEmitter.value(splice.Start)
	count := "0.0"
	if splice.Count != nil {
		count = nativeEmitter.value(splice.Count)
	}
	items := []string{}
	for _, item := range splice.Items {
		items = append(items, nativeEmitter.heldIn(array, splice.Element, nativeEmitter.value(item)))
	}
	packed := "NULL"
	if len(items) > 0 {
		packed = "(adamic_value[]){" + strings.Join(items, ", ") + "}"
	}
	return fmt.Sprintf("%s, %s, %s, %t, %d, %s", array, start, count, splice.Count != nil, len(items), packed)
}

// constantNumbers is how many constant numbers an array literal holds before they're emitted as
// static data a loop pushes, not a push a line. clang spends little on data and much on a long run of
// calls: tsprinter's emoji width table, 1,472 arrays of numbers, was one 11,409-line statement
// (docs/tsc-generated-c.md).
const constantNumbers = 64

// constantArray uses a bounded fresh-copy helper for flat numeric literals. The
// existing table-of-rows path still requires constantNumbers numbers, and leaves
// placement-bearing rows to ordinary emission. Immutable data shares content names.
func (nativeEmitter *emitter) constantArray(literal ir.ArrayLiteral) (string, bool) {
	if array, ok := nativeEmitter.constantNumericArray(literal); ok {
		return array, true
	}
	// Flat literals not selected above use ordinary emission. This remaining
	// path only packs rows.
	if literal.ProgramRegion ||
		nativeEmitter.graphTypes(literal.GraphTypes) ||
		!literal.Element.IsReference() ||
		len(literal.Elements) == 0 ||
		slices.Contains(literal.Spread, true) {
		return "", false
	}
	var numbers []float64
	var lengths []int
	for _, element := range literal.Elements {
		row, ok := element.(ir.ArrayLiteral)
		if !ok || row.ProgramRegion || nativeEmitter.graphTypes(row.GraphTypes) {
			return "", false
		}
		values, ok := constantNumbersOf(row)
		if !ok {
			return "", false
		}
		numbers = append(numbers, values...)
		lengths = append(lengths, len(values))
	}
	if len(numbers) < constantNumbers {
		return "", false
	}
	values := make([]string, len(numbers))
	for index, number := range numbers {
		values[index] = cNumber(number)
	}
	table := nativeEmitter.numberTable(values)
	counts := make([]string, len(lengths))
	for index, length := range lengths {
		counts[index] = strconv.Itoa(length)
	}
	rows := nativeEmitter.lengthTable(counts)
	array := nativeEmitter.own(
		ir.Array,
		fmt.Sprintf("adamic_array_make(%d, %s)", len(lengths), arrayElement(literal.Element)),
	)
	row, at, column, inner := nativeEmitter.temporary(), nativeEmitter.temporary(), nativeEmitter.temporary(), nativeEmitter.temporary()
	nativeEmitter.line("for (size_t %s = 0, %s = 0; %s < %d; %s++) {", row, at, row, len(lengths), row)
	nativeEmitter.line("\tadamic_array *%s = adamic_array_make(%s[%s], %s);", inner, rows, row, arrayElement(ir.Number))
	nativeEmitter.line("\tfor (size_t %s = 0; %s < %s[%s]; %s++, %s++) {", column, column, rows, row, column, at)
	nativeEmitter.line("\t\tadamic_array_push(%s, (adamic_value){.number = %s[%s]});", inner, table, at)
	nativeEmitter.line("\t}")
	// The row is fresh: the outer array takes its only reference.
	nativeEmitter.line(
		"\tadamic_array_push(%s, (adamic_value){.%s = %s});",
		array,
		member(literal.Element),
		slotted(literal.Element, inner),
	)
	nativeEmitter.line("}")
	return array, true
}

// constantNumbersOf is an array literal's numbers, when it holds only constant numbers, a negative
// one among them (a constant negated, which is exact).
func constantNumbersOf(literal ir.ArrayLiteral) ([]float64, bool) {
	if literal.Element != ir.Number || slices.Contains(literal.Spread, true) {
		return nil, false
	}
	numbers := make([]float64, 0, len(literal.Elements))
	for _, element := range literal.Elements {
		sign := 1.0
		if unary, isUnary := element.(ir.Unary); isUnary && unary.Operator == ir.Negate {
			sign, element = -1, unary.Operand
		}
		constant, isConstant := element.(ir.NumberConstant)
		if !isConstant {
			return nil, false
		}
		numbers = append(numbers, sign*constant.Value)
	}
	return numbers, true
}

// constantList is a static table's initializer, sixteen values a line.
func constantList(values []string) string {
	var builder strings.Builder
	for index, value := range values {
		if index > 0 {
			builder.WriteString(",")
			if index%16 == 0 {
				builder.WriteString("\n\t")
			} else {
				builder.WriteString(" ")
			}
		}
		builder.WriteString(value)
	}
	return builder.String()
}

// All literal paths use the same content keys and formatting for immutable data.
func (nativeEmitter *emitter) numberTable(values []string) string {
	data := constantList(values)
	name, exists := nativeEmitter.sharedDeclaration("adamic_numbers", data)
	if !exists {
		nativeEmitter.declarations = append(
			nativeEmitter.declarations,
			fmt.Sprintf("static const double %s[%d] = {%s};", name, len(values), data),
		)
	}
	return name
}
func (nativeEmitter *emitter) lengthTable(values []string) string {
	data := constantList(values)
	name, exists := nativeEmitter.sharedDeclaration("adamic_lengths", data)
	if !exists {
		nativeEmitter.declarations = append(
			nativeEmitter.declarations,
			fmt.Sprintf("static const size_t %s[%d] = {%s};", name, len(values), data),
		)
	}
	return name
}
