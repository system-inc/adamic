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
func (e *emitter) arrayVisit(visit ir.ArrayVisit) string {
	source := e.temporary()
	e.line("adamic_array *%s = %s;", source, e.value(visit.Array))
	callback := e.value(visit.Callback)
	references := visit.Element.IsReference()
	result := "0"
	switch visit.Method {
	case "filter":
		result = e.graphArray(fmt.Sprintf("adamic_array_new(0, %t)", references), visit.GraphTypes, visit.ProgramRegion)
		e.programArray(result, visit.ProgramRegion)
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
	search := visit.Method == "find" || visit.Method == "findIndex" || visit.Method == "findLast" || visit.Method == "findLastIndex"
	present := "true"
	if search {
		present = e.temporary()
		e.line("bool %s = %s < %s->length;", present, index, source)
		if !visit.SearchUndefined {
			e.declarations = append(e.declarations, "#include <stdio.h>")
			e.line("if (!(%s)) {", present)
			e.line("\tchar message[sizeof %s + 128];", cString(visit.SearchElementName))
			e.line("\tint length = snprintf(message, sizeof message, %s, %s);", cString(visit.Method+": index %zu is undefined; element type %s does not admit undefined"), index+", "+cString(visit.SearchElementName))
			e.line("\tadamic_panic(message, (size_t)length < sizeof message ? (size_t)length : sizeof message - 1);")
			e.line("}")
		}
	} else {
		e.line("if (%s >= %s->length) continue;", index, source)
	}
	missing := "{.reference = NULL}"
	if visit.Element.IsMaybe() {
		missing = fmt.Sprintf("{.%s = %s}", member(visit.Element), slotted(visit.Element, zero(visit.Element)))
	}
	e.line("adamic_value %s = (%s) ? %s->elements[%s] : (adamic_value)%s;", element, present, source, index, missing)
	if references {
		e.line("adamic_retain(%s.reference);", element)
	}
	argument, fresh := element, false
	if search && visit.SearchFirst != 0 && visit.SearchFirst != visit.Element {
		value, allocated := converted(visit.Element, visit.SearchFirst, unslotted(visit.Element, element+"."+member(visit.Element)))
		if visit.SearchNull {
			value = fmt.Sprintf("(%s.reference != NULL ? %s : &adamic_null)", element, value)
		}
		fresh = allocated
		argument = e.temporary()
		absent := missingArgument(visit.SearchFirst)
		e.line("adamic_value %s = {.%s = %s};", argument, member(visit.SearchFirst), slotted(visit.SearchFirst, fmt.Sprintf("(%s) ? %s : %s", present, value, absent)))
	}
	call := e.callbackCall(callback, visit.Callback, visit.CallbackType, argument, fmt.Sprintf("{.number = (double)%s}", index), fmt.Sprintf("{.reference = %s}", source))
	// A callback typed void may still return a value, which forEach lets go of as any void call does.
	discard := ir.DiscardNothing
	if visit.Method == "forEach" && visit.Returns == 0 {
		discard = e.program.ClosureDiscard(visit)
	}
	if visit.Method == "forEach" && !visit.Returns.IsReference() && !e.discarding(discard) {
		e.line("%s;", call)
	} else {
		e.line("adamic_value %s = %s;", answer, call)
	}
	if fresh {
		e.line("adamic_release(%s.reference);", argument)
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
		if e.discarding(discard) {
			e.line("%s", e.discard(discard, callback, answer))
		}
		release()
	case "filter":
		e.line("if (%s.boolean) {", answer)
		if references && e.graphTypes(visit.GraphTypes) {
			e.line("\t%s.reference = adamic_graph_take(%s, %s.reference);", element, result, element)
		}
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
			if visit.Element.IsMaybe() {
				found = unslotted(visit.Element, found)
			} else {
				found = maybe(visit.Type(), found)
			}
			e.line("\t%s = (%s) ? %s : %s;", result, present, found, zero(visit.Type()))
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
	call := e.callbackCall(callback, reduce.Callback, reduce.CallbackType, fmt.Sprintf("{.%s = %s}", member(reduce.Result), slotted(reduce.Result, accumulator)), element, fmt.Sprintf("{.number = (double)%s}", index), fmt.Sprintf("{.reference = %s}", source))
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
	name, exists := e.sharedDeclaration("adamic_compare", e.functionName(sort.Comparator)+fmt.Sprint(sort.Element))
	if exists {
		return name
	}
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
		items = append(items, e.heldIn(array, splice.Element, e.value(item)))
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
func (e *emitter) constantArray(literal ir.ArrayLiteral) (string, bool) {
	if array, ok := e.constantNumericArray(literal); ok {
		return array, true
	}
	// Flat literals not selected above use ordinary emission. This remaining
	// path only packs rows.
	if literal.ProgramRegion || e.graphTypes(literal.GraphTypes) || !literal.Element.IsReference() || len(literal.Elements) == 0 || slices.Contains(literal.Spread, true) {
		return "", false
	}
	var numbers []float64
	var lengths []int
	for _, element := range literal.Elements {
		row, ok := element.(ir.ArrayLiteral)
		if !ok || row.ProgramRegion || e.graphTypes(row.GraphTypes) {
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
	table := e.numberTable(values)
	counts := make([]string, len(lengths))
	for index, length := range lengths {
		counts[index] = strconv.Itoa(length)
	}
	rows := e.lengthTable(counts)
	array := e.own(ir.Array, fmt.Sprintf("adamic_array_new(%d, true)", len(lengths)))
	row, at, column, inner := e.temporary(), e.temporary(), e.temporary(), e.temporary()
	e.line("for (size_t %s = 0, %s = 0; %s < %d; %s++) {", row, at, row, len(lengths), row)
	e.line("\tadamic_array *%s = adamic_array_new(%s[%s], false);", inner, rows, row)
	e.line("\tfor (size_t %s = 0; %s < %s[%s]; %s++, %s++) {", column, column, rows, row, column, at)
	e.line("\t\tadamic_array_push(%s, (adamic_value){.number = %s[%s]});", inner, table, at)
	e.line("\t}")
	// The row is fresh: the outer array takes its only reference.
	e.line("\tadamic_array_push(%s, (adamic_value){.%s = %s});", array, member(literal.Element), slotted(literal.Element, inner))
	e.line("}")
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
func (e *emitter) numberTable(values []string) string {
	data := constantList(values)
	name, exists := e.sharedDeclaration("adamic_numbers", data)
	if !exists {
		e.declarations = append(e.declarations, fmt.Sprintf("static const double %s[%d] = {%s};", name, len(values), data))
	}
	return name
}
func (e *emitter) lengthTable(values []string) string {
	data := constantList(values)
	name, exists := e.sharedDeclaration("adamic_lengths", data)
	if !exists {
		e.declarations = append(e.declarations, fmt.Sprintf("static const size_t %s[%d] = {%s};", name, len(values), data))
	}
	return name
}
