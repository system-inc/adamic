// Functions and declarations moved unchanged from emit.go.
package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

// value emits what an expression needs evaluated first, in JavaScript's order, and returns a C
// expression that stays valid to the end of the statement.
func (e *emitter) evaluate(expression ir.Expression) string {
	switch expression := expression.(type) {
	case ir.TypedArrayNew, ir.TypedArrayFill, ir.TypedArraySet, ir.TypedArraySubarray:
		return e.typedArrayValue(expression)
	case ir.PromiseValue:
		return e.promiseValue(expression)
	case ir.Await:
		panic("native: await was not exposed before state cutting")
	case ir.RegExpNew:
		for _, argument := range expression.Arguments {
			e.value(argument)
		}
		return e.own(ir.Object, fmt.Sprintf("adamic_regex_new(&adamic_regex_%d, &adamic_string_%d, &adamic_string_%d)", expression.Index, expression.Source, expression.Flags))
	case ir.RegExpCall:
		return e.regexCall(expression)
	case ir.RegExpGroup:
		object := e.value(expression.Object)
		value := fmt.Sprintf("(%s)adamic_regex_group_lookup(%s, %s, %t)", cType(expression.Of), object, cString(expression.Name), expression.Optional)
		return e.own(expression.Of, "adamic_retain("+value+")")
	case ir.RegExpProperty:
		return e.regexProperty(expression)
	case ir.Null:
		return "NULL"
	case ir.IsNull:
		if expression.AlwaysFalse {
			e.value(expression.Value)
			return "false"
		}
		return fmt.Sprintf("(%s == NULL)", e.value(expression.Value))
	case ir.NumberConstant:
		return cNumber(expression.Value)
	case ir.BooleanConstant:
		return strconv.FormatBool(expression.Value)
	case ir.StringConstant:
		return fmt.Sprintf("&adamic_string_%d", expression.Index)
	case ir.Read:
		return e.read(expression)
	case ir.Unary:
		operand := e.value(expression.Operand)
		switch expression.Operator {
		case ir.Negate:
			return "(-" + operand + ")"
		case ir.Plus:
			return "(+" + operand + ")"
		case ir.Not:
			return "(!" + operand + ")"
		case ir.BitNot:
			return "adamic_bitwise_not(" + operand + ")"
		}
	case ir.Binary:
		if expression.Operator == ir.And || expression.Operator == ir.Or {
			return e.logical(expression)
		}
		return e.binary(expression.Operator, expression.Left.Type(), e.value(expression.Left), e.value(expression.Right))
	case ir.Call:
		if expression.Accessor != "" {
			result := e.accessorCall(expression)
			if e.program.CallMayThrow(expression) {
				e.checkThrown()
			}
			return result
		}
		region := e.regionFor(expression)
		arguments := e.arguments(expression)
		call := e.callCode(expression, arguments)
		var result string
		switch {
		case region != "":
			// What it returns is in that region: no count to own (region.go).
			result = e.regionValue(fmt.Sprintf("%s(%s)", e.regionFunctionName(expression.Function), strings.Join(append([]string{region}, arguments...), ", ")))
		case expression.Returns.IsReference():
			result = e.own(expression.Returns, call)
		default:
			result = e.temporary()
			e.line("%s %s = %s;", cType(expression.Returns), result, call)
		}
		if e.program.CallMayThrow(expression) {
			// A throw left the call: the result is the zero value it returned, owned like any.
			e.checkThrown()
		}
		return result
	case ir.HasAccessor:
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_accessor_find(%s, %s) != NULL", e.value(expression.Object), cString(expression.Name)))
	case ir.InstanceOf:
		value := e.value(expression.Value)
		if !expression.Value.Type().IsReference() {
			e.line("(void)%s;", value)
			return e.snapshot(ir.Boolean, "false")
		}
		if expression.Exact {
			return e.snapshot(ir.Boolean, fmt.Sprintf("(%s != NULL && ((adamic_object *)%s)->class == &adamic_class_%d)", value, value, expression.Class))
		}
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_instanceof(%s, &adamic_class_%d)", value, expression.Class))
	case ir.NumberToString:
		return e.own(ir.String, fmt.Sprintf("adamic_string_from_number(%s)", e.value(expression.Value)))
	case ir.BooleanToString:
		return fmt.Sprintf("((%s) ? &adamic_string_true : &adamic_string_false)", e.value(expression.Value))
	case ir.Concat:
		parts := make([]string, 0, len(expression.Parts))
		for _, part := range expression.Parts {
			parts = append(parts, e.value(part))
		}
		return e.own(ir.String, fmt.Sprintf("adamic_string_concat(%d, (adamic_string *const[]){%s})", len(parts), strings.Join(parts, ", ")))
	case ir.Conditional:
		return e.conditional(expression)
	case ir.ObjectLiteral:
		return e.objectLiteral(expression)
	case ir.Property:
		if taken, ok := e.take(expression); ok {
			return taken
		}
		lent := e.lendable && lendable(expression.Of) && !expression.Optional
		object := e.value(expression.Object)
		field := unslotted(expression.Of, fmt.Sprintf("%s->%s", e.fieldSlot(object, expression.Name, expression.Class), member(expression.Of)))
		if expression.Of == ir.MaybeNumber {
			field = fmt.Sprintf("adamic_object_maybe_number(%s, %s, &%s)", object, cString(expression.Name), e.cache())
		}
		if expression.Absent {
			slot := e.temporary()
			lookup := fmt.Sprintf("adamic_object_optional_field(%s, %s, &%s)", object, cString(expression.Name), e.cache())
			if expression.Optional {
				lookup = fmt.Sprintf("(%s == NULL ? NULL : %s)", object, lookup)
			}
			e.line("adamic_value *%s = %s;", slot, lookup)
			undefined := "NULL"
			if expression.Of.IsMaybe() {
				undefined = zero(expression.Of)
			}
			present := unslotted(expression.Of, fmt.Sprintf("%s->%s", slot, member(expression.Of)))
			if expression.Of == ir.MaybeNumber {
				present = field
			}
			field = fmt.Sprintf("(%s == NULL ? %s : %s)", slot, undefined, present)
		}
		if expression.Of == ir.MaybeNumber && expression.Optional {
			return e.snapshot(ir.MaybeNumber, fmt.Sprintf("(%s == NULL ? %s : %s)", object, zero(ir.MaybeNumber), field))
		}
		if expression.Of.IsReference() {
			// A field holds a reference as void *; read through the type the checker proved.
			field = fmt.Sprintf("((%s)%s)", cType(expression.Of), field)
		}
		if expression.Optional && expression.Type().IsMaybe() {
			return e.snapshot(expression.Type(), fmt.Sprintf("(%s == NULL ? %s : %s)", object, zero(expression.Type()), maybe(expression.Type(), field)))
		}
		if expression.Optional {
			field = fmt.Sprintf("(%s == NULL ? NULL : %s)", object, field)
		}
		// Read now, when JavaScript reads it: a call later in the statement may write the field. A
		// reference is retained, so that write can't free it from under its reader, unless nothing
		// can run before it's used (borrow.go).
		if lent {
			e.self = true
			return e.snapshot(expression.Of, field)
		}
		if expression.Of.IsReference() {
			return e.own(expression.Of, fmt.Sprintf("adamic_retain(%s)", field))
		}
		return e.snapshot(expression.Of, field)
	case ir.Undefined:
		return "NULL"
	case ir.IsUndefined:
		if expression.Value.Type().IsMaybe() {
			return fmt.Sprintf("(!(%s).present)", e.value(expression.Value))
		}
		return fmt.Sprintf("(%s == NULL)", e.value(expression.Value))
	case ir.Unwrap:
		// The checker narrowed undefined away, but a call since may have put it back (ir.Unwrap).
		value := e.snapshot(expression.Value.Type(), e.value(expression.Value))
		e.line("if (!%s.present) {", value)
		e.line("\tstatic const char message[] = %s;", cArray(narrowedAwayMessage))
		e.line("\tadamic_panic(message, sizeof message - 1);")
		e.line("}")
		return fmt.Sprintf("(%s).%s", value, member(expression.Type()))
	case ir.Defined:
		value := e.value(expression.Value)
		e.checkDefined(value, expression.Message)
		return value
	case ir.MaybeOf:
		if expression.Value == nil {
			return zero(expression.Of)
		}
		return maybe(expression.Of, e.value(expression.Value))
	case ir.MaybeToString:
		return e.maybeToString(expression.Value)
	case ir.Box:
		return e.box(expression.Value)
	case ir.MakeError:
		return e.makeError(expression)
	case ir.WeakOf:
		return e.own(ir.Weak, fmt.Sprintf("adamic_weak_of(%s)", e.value(expression.Value)))
	case ir.WeakTarget:
		target := "adamic_weak_target"
		if expression.Present {
			target = "adamic_weak_target_present"
		}
		// Retained, as any reference read out of a slot is: a call later in the statement may let go
		// of the last strong holder.
		return e.own(expression.To, fmt.Sprintf("(%s)adamic_retain(%s(%s))", cType(expression.To), target, e.value(expression.Value)))
	case ir.Narrow:
		return e.narrow(expression)
	case ir.TypeOf:
		return e.typeOf(expression)
	case ir.UnionToString:
		return e.own(ir.String, fmt.Sprintf("adamic_union_to_string(%s)", e.value(expression.Value)))
	case ir.Coalesce:
		return e.coalesce(expression)
	case ir.StringLength:
		if expression.Optional {
			text := e.value(expression.Value)
			return e.snapshot(ir.MaybeNumber, fmt.Sprintf("(%s == NULL ? %s : (adamic_maybe_number){true, adamic_string_length(%s)})", text, zero(ir.MaybeNumber), text))
		}
		return fmt.Sprintf("adamic_string_length(%s)", e.value(expression.Value))
	case ir.CharCodeAt:
		value := e.value(expression.Value)
		index := e.value(expression.Index)
		return fmt.Sprintf("adamic_string_char_code_at(%s, %s)", value, index)
	case ir.ObjectKeys, ir.ClosureSelf, ir.LibraryGlobal:
		return e.libraryLanguageValue(expression)
	case ir.MakeClosure:
		environment := e.program.Functions[expression.Function].Environment
		closure := e.own(ir.Closure, fmt.Sprintf("adamic_closure_new(%s, %d)", e.functionName(expression.Function), len(environment)))
		for index, local := range environment {
			e.line("%s->cells[%d] = adamic_retain(%s);", closure, index, e.cellReference(local))
		}
		return closure
	case ir.CallClosure:
		property, isProperty := expression.Closure.(ir.Property)
		if !isProperty || !property.Method {
			return e.callThrough(expression, e.value(expression.Closure), "")
		}
		// object.name(...) through an interface. The object stays alive for the call though an
		// argument may write where it was read from: a call lends nothing (borrow.go), so a read of a
		// global, a captured variable, a field or an element is retained when it's read, and a local
		// can't be written in the middle of an expression.
		receiver := e.value(property.Object)
		if !property.Optional {
			return e.callThrough(expression, "", receiver)
		}
		// object?.name(...): undefined, the method not looked up and the arguments not evaluated,
		// where the object is undefined, as JavaScript's chain stops there.
		text, value, owned := e.asideWith(func() string { return e.callThrough(expression, "", receiver) })
		result := "0"
		if expression.Returns != 0 {
			// Undefined: NULL for a reference, which zero isn't for a string.
			undefined := "NULL"
			if !expression.Returns.IsReference() {
				undefined = zero(expression.Returns)
			}
			result = e.temporary()
			e.line("%s %s = %s;", cType(expression.Returns), result, undefined)
		}
		e.line("if (%s != NULL) {", receiver)
		e.out.WriteString(text)
		e.indent++
		switch {
		case expression.Returns == 0:
		case expression.Returns.IsReference():
			e.line("%s = %s;", result, retained(value))
		default:
			e.line("%s = %s;", result, value)
		}
		for index := len(owned) - 1; index >= 0; index-- {
			e.line("adamic_release(%s);", owned[index])
		}
		e.indent--
		e.line("}")
		if expression.Returns.IsReference() {
			e.owned = append(e.owned, result)
		}
		return result
	case ir.ArrayMap:
		if mapped, ok := e.mapped(expression); ok {
			return mapped
		}
		source := e.temporary()
		e.line("adamic_array *%s = %s;", source, e.value(expression.Array))
		callback := e.value(expression.Callback)
		mapped := e.own(ir.Array, fmt.Sprintf("adamic_array_new(%s->length, %t)", source, expression.Result.IsReference()))
		count, index := e.temporary(), e.temporary()
		// The length is read once, as JavaScript's map does; an array the callback shrinks is a
		// panic here rather than JavaScript's holes, which 0.1 has no way to hold.
		e.line("size_t %s = %s->length;", count, source)
		e.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
		e.line("\tif (%s >= %s->length) {", index, source)
		e.line("\t\tstatic const char message[] = \"map: the array shrank while it was being mapped\";")
		e.line("\t\tadamic_panic(message, sizeof message - 1);")
		e.line("\t}")
		e.indent++
		element := e.temporary()
		e.line("adamic_value %s = %s->code(%s, (adamic_value[]){%s->elements[%s], {.number = (double)%s}, {.reference = %s}});", element, callback, callback, source, index, index, source)
		// What's mapped so far is the statement's, let go with its temporaries.
		e.closureThrown()
		e.line("adamic_array_push(%s, %s);", mapped, element)
		e.indent--
		e.line("}")
		return mapped
	case ir.ArrayVisit:
		return e.arrayVisit(expression)
	case ir.ArrayReduce:
		return e.arrayReduce(expression)
	case ir.ArraySearch:
		return e.libraryArraySearch(expression)
	case ir.ArrayFill:
		if expression.Array == nil {
			length := e.value(expression.Length)
			value := e.value(expression.Value)
			return e.own(ir.Array, fmt.Sprintf("adamic_array_filled(%s, %s, %t)", length, borrowed(expression.Element, value), expression.Element.IsReference()))
		}
		array := e.value(expression.Array)
		value := e.value(expression.Value)
		start, end := "0.0", "0.0"
		if expression.Start != nil {
			start = e.value(expression.Start)
		}
		if expression.End != nil {
			end = e.value(expression.End)
		}
		e.line("adamic_array_fill(%s, %s, %s, %s, %t, %t);", array, borrowed(expression.Element, value), start, end, expression.Start != nil, expression.End != nil)
		return array
	case ir.ArraySplice:
		return e.own(ir.Array, fmt.Sprintf("adamic_array_splice(%s)", e.spliceArguments(expression)))
	case ir.ArrayFrom:
		return e.arrayFrom(expression)
	case ir.ArrayReverse:
		array := e.value(expression.Array)
		e.line("adamic_array_reverse(%s);", array)
		return array
	case ir.ArrayConcat:
		arrays := []string{e.value(expression.Array)}
		for _, other := range expression.Others {
			arrays = append(arrays, e.value(other))
		}
		return e.own(ir.Array, fmt.Sprintf("adamic_array_concat(%d, (adamic_array *const[]){%s})", len(arrays), strings.Join(arrays, ", ")))
	case ir.CheckedCast:
		object := e.temporary()
		e.line("adamic_object *%s = %s;", object, e.value(expression.Value))
		field := fmt.Sprintf("adamic_object_field(%s, %s, &%s)->%s", object, cString(expression.Field), e.cache(), member(expression.FieldType))
		if expression.FieldType.IsReference() {
			field = fmt.Sprintf("((%s)%s)", cType(expression.FieldType), field)
		}
		tests := []string{}
		for _, allowed := range expression.Allowed {
			tests = append(tests, e.binary(ir.Equal, expression.FieldType, field, e.value(allowed)))
		}
		e.line("if (!(%s)) {", strings.Join(tests, " || "))
		e.line("\tstatic const char message[] = %s;", cArray(expression.Message))
		e.line("\tadamic_panic(message, sizeof message - 1);")
		e.line("}")
		return object
	case ir.ArrayIndex:
		if expression.Array.Type().IsTypedArray() {
			array := e.value(expression.Array)
			index := e.value(expression.Index)
			return e.snapshot(ir.MaybeNumber, fmt.Sprintf("adamic_typed_array_get(%s, %s)", array, index))
		}
		slot := e.arrayIndexSlot(expression)
		if expression.Type().IsMaybe() {
			return e.snapshot(expression.Type(), maybeSlot(expression.Element, slot))
		}
		// Retained, so a write later in the statement can't free it from under its reader.
		return e.own(expression.Element, fmt.Sprintf("%s == NULL ? NULL : (%s)adamic_retain(%s->reference)", slot, cType(expression.Element), slot))
	case ir.ArrayPop:
		array := e.temporary()
		e.line("adamic_array *%s = %s;", array, e.value(expression.Array))
		if expression.Type().IsMaybe() {
			popped := fmt.Sprintf("%s->elements[--%s->length].%s", array, array, member(expression.Element))
			if expression.Element == ir.MaybeNumber {
				popped = unslotted(ir.MaybeNumber, popped)
			} else {
				popped = maybe(expression.Type(), popped)
			}
			return e.snapshot(expression.Type(), fmt.Sprintf("%s->length == 0 ? %s : %s", array, zero(expression.Type()), popped))
		}
		// The array's reference to the element becomes the statement's.
		return e.own(expression.Element, fmt.Sprintf("%s->length == 0 ? NULL : %s->elements[--%s->length].%s", array, array, array, member(expression.Element)))
	case ir.MapEntries:
		pair := e.shapeOf([]string{"0", "1"}, []ir.Type{expression.KeyType, expression.ValueType})
		return e.own(ir.Array, fmt.Sprintf("adamic_map_entries(%s, &%s)", e.value(expression.Map), pair))
	case ir.ArraySlice:
		array := e.value(expression.Array)
		arguments := []string{"0.0", "0.0"}
		for index, argument := range expression.Arguments {
			arguments[index] = e.value(argument)
		}
		return e.own(ir.Array, fmt.Sprintf("adamic_array_slice(%s, %s, %s, %t)", array, arguments[0], arguments[1], len(expression.Arguments) == 2))
	case ir.ArraySort:
		array := e.value(expression.Array)
		sort := "adamic_array_sort"
		if expression.Element == ir.MaybeNumber {
			sort = "adamic_array_sort_undefined_last"
		}
		// A comparator that throws stops the sort, which leaves the array as it was, as V8's does
		// (sort.c), and the throw goes on from here.
		if expression.Callback != nil {
			e.line("%s(%s, adamic_compare_closure, %s);", sort, array, e.value(expression.Callback))
			e.closureThrown()
			return array
		}
		e.line("%s(%s, %s, NULL);", sort, array, e.comparator(expression))
		if e.program.ClosureMayThrow(expression) {
			e.checkThrown()
		}
		return array
	case ir.CodePoints:
		return e.own(ir.Array, fmt.Sprintf("adamic_string_code_points(%s)", e.value(expression.Value)))
	case ir.StringIndex:
		value := e.value(expression.Value)
		index := e.value(expression.Index)
		return e.own(ir.String, fmt.Sprintf("adamic_string_at(%s, %s)", value, index))
	case ir.StringCall:
		return e.stringCall(expression)
	case ir.Trim:
		return e.own(ir.String, fmt.Sprintf("adamic_string_trim(%s)", e.value(expression.Value)))
	case ir.CollectionIterator:
		return e.collectionIterator(expression)
	case ir.MapNew:
		entries := make([][2]string, 0, len(expression.Entries))
		for _, entry := range expression.Entries {
			entries = append(entries, [2]string{e.value(entry[0]), e.value(entry[1])})
		}
		var pairs string
		if expression.Pairs != nil {
			pairs = e.value(expression.Pairs)
		}
		created := e.own(ir.Map, newMap(expression.Key, expression.Value.IsReference()))
		for _, entry := range entries {
			e.line("adamic_map_set(%s, %s, %s);", created, held(expression.Key, entry[0]), held(expression.Value, entry[1]))
		}
		if expression.Pairs != nil {
			e.line("adamic_map_add_pairs(%s, %s);", created, pairs)
		}
		return created
	case ir.MapKeys:
		return e.own(ir.Array, fmt.Sprintf("adamic_map_keys(%s)", e.value(expression.Map)))
	case ir.MapValues:
		return e.own(ir.Array, fmt.Sprintf("adamic_map_values(%s)", e.value(expression.Map)))
	case ir.MapClear:
		e.line("adamic_map_clear(%s);", e.value(expression.Map))
		return "0"
	case ir.MapForEach:
		return e.mapForEach(expression)
	case ir.SetNew:
		var values string
		if expression.Values != nil {
			values = e.value(expression.Values)
		}
		created := e.own(ir.Map, newMap(expression.Element, false))
		if expression.Values != nil {
			e.line("adamic_set_add_all(%s, %s);", created, values)
		}
		return created
	case ir.SetAdd:
		set := e.value(expression.Set)
		value := e.value(expression.Value)
		e.line("adamic_map_set(%s, %s, (adamic_value){.number = 0});", set, held(expression.Element, value))
		return set
	case ir.SetValues:
		return e.own(ir.Array, fmt.Sprintf("adamic_set_values(%s)", e.value(expression.Set)))
	case ir.MapGet:
		slot := e.mapGetSlot(expression)
		if expression.Type().IsMaybe() {
			return e.snapshot(expression.Type(), maybeSlot(expression.ValueType, slot))
		}
		// Retained, so a set later in the same statement can't free it from under its reader.
		return e.own(expression.ValueType, fmt.Sprintf("%s == NULL ? NULL : adamic_retain(%s->%s)", slot, slot, member(expression.ValueType)))
	case ir.MapSet:
		object := e.value(expression.Map)
		key := e.value(expression.Key)
		value := e.value(expression.Value)
		e.line("adamic_map_set(%s, %s, %s);", object, held(expression.KeyType, key), held(expression.ValueType, value))
		return object
	case ir.MapHas:
		object := e.value(expression.Map)
		key := e.value(expression.Key)
		return e.snapshot(ir.Boolean, fmt.Sprintf("(adamic_map_get(%s, %s) != NULL)", object, borrowed(expression.KeyType, key)))
	case ir.MapDelete:
		object := e.value(expression.Map)
		key := e.value(expression.Key)
		result := e.temporary()
		e.line("bool %s = adamic_map_delete(%s, %s);", result, object, borrowed(expression.KeyType, key))
		return result
	case ir.MapSize:
		return e.snapshot(ir.Number, fmt.Sprintf("(double)%s->count", e.value(expression.Map)))
	case ir.HasOwn:
		return e.snapshot(ir.Boolean, fmt.Sprintf("adamic_object_has(%s, %s)", e.value(expression.Object), e.value(expression.Key)))
	case ir.ParallelMap:
		return e.parallelMap(expression)
	case ir.ReadTextFile:
		return e.own(ir.Object, fmt.Sprintf("adamic_read_text_file(%s)", e.value(expression.Path)))
	case ir.ProgramArguments:
		return e.own(ir.Array, "adamic_program_arguments()")
	case ir.Utf8Length:
		return fmt.Sprintf("adamic_utf8_length(%s)", e.value(expression.Text))
	case ir.Utf8At:
		text := e.value(expression.Text)
		return e.snapshot(ir.Number, fmt.Sprintf("adamic_utf8_at(%s, %s)", text, e.value(expression.Index)))
	case ir.ReadDirectory:
		return e.own(ir.Object, fmt.Sprintf("adamic_read_directory(%s)", e.value(expression.Path)))
	case ir.FileStatus:
		return e.own(ir.Object, fmt.Sprintf("adamic_file_status(%s)", e.value(expression.Path)))
	case ir.WriteTextFile:
		path := e.value(expression.Path)
		text := e.value(expression.Text)
		return e.own(ir.Object, fmt.Sprintf("adamic_write_text_file(%s, %s)", path, text))
	case ir.ArrayPush:
		array := e.value(expression.Array)
		value := e.value(expression.Value)
		if expression.Element.IsReference() {
			value = retained(value)
		}
		// The append happens here, in JavaScript's order, and the new length is the value.
		e.line("adamic_array_push(%s, (adamic_value){.%s = %s});", array, member(expression.Element), slotted(expression.Element, value))
		length := e.temporary()
		e.line("double %s = (double)%s->length;", length, array)
		return length
	case ir.ArrayJoin:
		if expression.Depth > 0 {
			return e.libraryArrayJoin(expression)
		}
		array := e.value(expression.Array)
		separator := e.value(expression.Separator)
		return e.own(ir.String, fmt.Sprintf("adamic_array_join(%s, %s, %s)", array, separator, joinKind(expression.Element)))
	case ir.ArrayLiteral:
		if spread, ok := e.spreadArray(expression); ok {
			return spread
		}
		if expression.Spread != nil {
			// A spread is iterated where it stands, before the elements after it are evaluated, so the
			// array is made first (which nothing can see) and each element appended as it comes.
			array := e.own(ir.Array, fmt.Sprintf("adamic_array_new(0, %t)", expression.Element.IsReference()))
			for index, element := range expression.Elements {
				value := e.value(element)
				switch {
				case expression.Spread[index]:
					e.line("adamic_array_append(%s, %s);", array, value)
				case expression.Element.IsReference():
					e.line("adamic_array_push(%s, (adamic_value){.reference = %s});", array, retained(value))
				default:
					e.line("adamic_array_push(%s, (adamic_value){.%s = %s});", array, member(expression.Element), slotted(expression.Element, value))
				}
			}
			return array
		}
		elements := make([]string, 0, len(expression.Elements))
		for _, element := range expression.Elements {
			elements = append(elements, e.value(element))
		}
		array := e.own(ir.Array, fmt.Sprintf("adamic_array_new(%d, %t)", len(elements), expression.Element.IsReference()))
		for _, element := range elements {
			if expression.Element.IsReference() {
				element = retained(element)
			}
			e.line("adamic_array_push(%s, (adamic_value){.%s = %s});", array, member(expression.Element), slotted(expression.Element, element))
		}
		return array
	case ir.Length:
		if expression.Array.Type().IsTypedArray() {
			array := e.value(expression.Array)
			length := fmt.Sprintf("adamic_typed_array_length(%s)", array)
			if expression.Optional {
				return e.snapshot(ir.MaybeNumber, fmt.Sprintf("(%s == NULL ? %s : (adamic_maybe_number){true, %s})", array, zero(ir.MaybeNumber), length))
			}
			return e.snapshot(ir.Number, length)
		}
		if expression.Optional {
			array := e.value(expression.Array)
			return e.snapshot(ir.MaybeNumber, fmt.Sprintf("(%s == NULL ? %s : (adamic_maybe_number){true, (double)%s->length})", array, zero(ir.MaybeNumber), array))
		}
		return e.snapshot(ir.Number, fmt.Sprintf("(double)%s->length", e.value(expression.Array)))
	case ir.JSONEncode:
		return e.jsonEncode(expression)
	case ir.JSONDecode:
		return e.jsonDecode(expression)
	case ir.JSONStringify:
		return e.jsonStringify(expression)
	case ir.JSONNull:
		return "NULL"
	case ir.MathCall:
		return e.mathCall(expression)
	case ir.StringFromCodes:
		function := "adamic_string_from_char_codes"
		if expression.CodePoints {
			function = "adamic_string_from_code_points"
		}
		if expression.Spread != nil {
			return e.own(ir.String, fmt.Sprintf("%s_of(%s)", function, e.value(expression.Spread)))
		}
		codes := make([]string, 0, len(expression.Codes))
		for _, code := range expression.Codes {
			codes = append(codes, e.value(code))
		}
		if len(codes) == 0 {
			return e.own(ir.String, function+"(0, NULL)")
		}
		return e.own(ir.String, fmt.Sprintf("%s(%d, (const double[]){%s})", function, len(codes), strings.Join(codes, ", ")))
	case ir.ObjectCall:
		return e.objectCall(expression)
	case ir.NumberCall:
		if value, known := e.libraryNumberCall(expression); known {
			return value
		}
		arguments := []string{}
		for _, argument := range expression.Arguments {
			arguments = append(arguments, e.value(argument))
		}
		switch expression.Function {
		case "parseInt":
			// A radix left out is undefined, which ToInt32 makes 0: detect it from the text.
			radix := "0.0"
			if len(arguments) == 2 {
				radix = arguments[1]
			}
			return e.snapshot(ir.Number, fmt.Sprintf("adamic_number_parse_int(%s, %s)", arguments[0], radix))
		case "parseFloat":
			return e.snapshot(ir.Number, fmt.Sprintf("adamic_number_parse_float(%s)", arguments[0]))
		case "isNaN":
			return fmt.Sprintf("isnan(%s)", arguments[0])
		case "isFinite":
			return fmt.Sprintf("isfinite(%s)", arguments[0])
		case "isInteger":
			return fmt.Sprintf("(isfinite(%s) && trunc(%s) == %s)", arguments[0], arguments[0], arguments[0])
		}
		return fmt.Sprintf("(isfinite(%s) && trunc(%s) == %s && fabs(%s) <= 9007199254740991.0)", arguments[0], arguments[0], arguments[0], arguments[0])
	case ir.ToFixed:
		value := e.value(expression.Value)
		digits := e.value(expression.Digits)
		return e.own(ir.String, fmt.Sprintf("adamic_number_to_fixed(%s, %s)", value, digits))
	case ir.NumberFormat:
		return e.numberFormat(expression)
	}
	panic(fmt.Sprintf("native: no C for %T", expression))
}

// cBitwise are the runtime's bitwise operators (bitwise.c).
var cBitwise = map[ir.Operator]string{
	ir.BitAnd: "adamic_bitwise_and", ir.BitOr: "adamic_bitwise_or", ir.BitXor: "adamic_bitwise_xor",
	ir.ShiftLeft: "adamic_shift_left", ir.ShiftRight: "adamic_shift_right", ir.ShiftRightUnsigned: "adamic_shift_right_unsigned",
}

var cOperators = map[ir.Operator]string{
	ir.Add: "+", ir.Subtract: "-", ir.Multiply: "*", ir.Divide: "/",
	ir.Less: "<", ir.LessOrEqual: "<=", ir.Greater: ">", ir.GreaterOrEqual: ">=",
	ir.Equal: "==", ir.NotEqual: "!=",
}

func (e *emitter) binary(operator ir.Operator, operandType ir.Type, left string, right string) string {
	switch {
	case operator == ir.Remainder:
		// JavaScript's % is C's fmod: truncated, with the sign of the dividend.
		return fmt.Sprintf("fmod(%s, %s)", left, right)
	case operator == ir.Power:
		return fmt.Sprintf("adamic_power(%s, %s)", left, right)
	case cBitwise[operator] != "":
		return fmt.Sprintf("%s(%s, %s)", cBitwise[operator], left, right)
	case operandType == ir.String && (operator == ir.Less || operator == ir.LessOrEqual || operator == ir.Greater || operator == ir.GreaterOrEqual):
		return fmt.Sprintf("(adamic_string_compare(%s, %s) %s 0)", left, right, cOperators[operator])
	case operandType == ir.String && operator == ir.Equal:
		return fmt.Sprintf("adamic_string_equal(%s, %s)", left, right)
	case operandType == ir.String && operator == ir.NotEqual:
		return fmt.Sprintf("(!adamic_string_equal(%s, %s))", left, right)
	case operandType.IsMaybe() && (operator == ir.Equal || operator == ir.NotEqual):
		return pairEquality(operator, operandType, left, right)
	case operandType == ir.Union && operator == ir.Equal:
		return fmt.Sprintf("adamic_union_equal(%s, %s)", left, right)
	case operandType == ir.Union && operator == ir.NotEqual:
		return fmt.Sprintf("(!adamic_union_equal(%s, %s))", left, right)
	}
	return fmt.Sprintf("(%s %s %s)", left, cOperators[operator], right)
}

// narrowedAwayMessage is the panic of a number or a boolean the checker narrowed undefined out of,
// read where a call since put it back (ir.Unwrap). JavaScript would go on computing with undefined.
const narrowedAwayMessage = "undefined where the checker narrowed it away: a call since the narrowing put it back"
