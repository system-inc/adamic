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
func (nativeEmitter *emitter) evaluate(expression ir.Expression) string {
	switch expression := expression.(type) {
	case ir.TypedArrayNew, ir.TypedArrayFill, ir.TypedArraySet, ir.TypedArraySubarray, ir.TypedArraySort:
		return nativeEmitter.typedArrayValue(expression)
	case ir.HasProperty:
		value := nativeEmitter.value(expression.Object)
		return nativeEmitter.snapshot(
			ir.Boolean,
			fmt.Sprintf("adamic_has_property(%s, %s)", value, cString(expression.Name)),
		)
	case ir.DynamicProperty:
		value := nativeEmitter.value(expression.Object)
		return nativeEmitter.own(
			ir.Union,
			fmt.Sprintf("adamic_dynamic_property(%s, %s)", value, cString(expression.Name)),
		)
	case ir.DynamicKeyedProperty:
		// The object, then the key, as Reflect.get's arguments are evaluated; both are borrowed for
		// the call, and the union it gives is owned (a getter's result, a retained slot or record value,
		// or a new box).
		value := nativeEmitter.value(expression.Object)
		key := nativeEmitter.value(expression.Key)
		classes := "NULL, 0"
		if rows := nativeEmitter.keyedClassRows(); len(rows) > 0 {
			classes = fmt.Sprintf("adamic_keyed_classes, %d", len(rows))
		}
		return nativeEmitter.own(
			ir.Union,
			fmt.Sprintf("adamic_dynamic_property_keyed(%s, %s, %s)", value, key, classes),
		)
	case ir.ProcessCall:
		return nativeEmitter.processCall(expression)
	case ir.NodeHostCall:
		return nativeEmitter.nodeHostCall(expression)
	case ir.NodeFSFile:
		return nativeEmitter.nodeFileSystemFile(expression)
	case ir.NodeBufferCall:
		return nativeEmitter.nodeBufferCall(expression)
	case ir.PromiseValue:
		return nativeEmitter.promiseValue(expression)
	case ir.Await:
		panic("native: await was not exposed before state cutting")
	case ir.Yield:
		panic("native: yield was not exposed before state cutting")
	case ir.GeneratorNext:
		return nativeEmitter.generatorNext(expression)
	case ir.GeneratorDone:
		return nativeEmitter.snapshot(ir.Boolean, fmt.Sprintf(
			"adamic_generator_done(%s)",
			generatorPointer(expression.Generator, nativeEmitter.value(expression.Generator)),
		))
	case ir.IterableOf:
		// The generator itself, seen as the object an Iterable<T> slot holds: no retain, no release.
		return fmt.Sprintf("((adamic_object *)%s)", nativeEmitter.value(expression.Value))
	case ir.IsGenerator:
		return nativeEmitter.snapshot(
			ir.Boolean,
			fmt.Sprintf("adamic_generator_is(%s)", nativeEmitter.value(expression.Value)),
		)
	case ir.IsCollectionIterator:
		return nativeEmitter.snapshot(
			ir.Boolean,
			fmt.Sprintf("adamic_collection_iterator_is(%s)", nativeEmitter.value(expression.Value)),
		)
	case ir.RegExpNew:
		for _, argument := range expression.Arguments {
			nativeEmitter.value(argument)
		}
		return nativeEmitter.own(ir.Object, fmt.Sprintf(
			"adamic_regex_new(&%s, &%s, &%s)",
			nativeEmitter.regexName(expression.Index),
			stringName(nativeEmitter.program.Strings[expression.Source]),
			stringName(nativeEmitter.program.Strings[expression.Flags]),
		))
	case ir.RegExpCall:
		return nativeEmitter.regexCall(expression)
	case ir.RegExpGroup:
		object := nativeEmitter.value(expression.Object)
		value := fmt.Sprintf(
			"(%s)adamic_regex_group_lookup(%s, %s, %t)",
			cType(expression.Of),
			object,
			cString(expression.Name),
			expression.Optional,
		)
		return nativeEmitter.own(expression.Of, "adamic_retain("+value+")")
	case ir.RegExpProperty:
		return nativeEmitter.regexProperty(expression)
	case ir.Null:
		return "NULL"
	case ir.IsNull:
		if expression.AlwaysFalse {
			nativeEmitter.value(expression.Value)
			return "false"
		}
		if expression.Value.Type() == ir.Union {
			return fmt.Sprintf("(%s == &adamic_null)", nativeEmitter.value(expression.Value))
		}
		return fmt.Sprintf("(%s == NULL)", nativeEmitter.value(expression.Value))
	case ir.NumberConstant:
		return cNumber(expression.Value)
	case ir.BooleanConstant:
		return strconv.FormatBool(expression.Value)
	case ir.StringConstant:
		return "&" + stringName(nativeEmitter.program.Strings[expression.Index])
	case ir.Read:
		return nativeEmitter.read(expression)
	case ir.Truthy:
		return nativeEmitter.toBoolean(expression.Value.Type(), nativeEmitter.value(expression.Value))
	case ir.Void:
		nativeEmitter.line("(void)%s;", nativeEmitter.value(expression.Value))
		if expression.Type().IsMaybe() {
			return zero(expression.Type())
		}
		return "NULL"
	case ir.Comma:
		nativeEmitter.line("(void)%s;", nativeEmitter.value(expression.Left))
		return nativeEmitter.value(expression.Right)
	case ir.Effects:
		return nativeEmitter.effects(expression)
	case ir.LogicalAssignment:
		return nativeEmitter.logicalAssignment(expression)
	case ir.Logical:
		return nativeEmitter.logicalValue(expression)
	case ir.Unary:
		operand := nativeEmitter.value(expression.Operand)
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
			return nativeEmitter.logical(expression)
		}
		return nativeEmitter.binary(
			expression.Operator,
			expression.Left.Type(),
			nativeEmitter.value(expression.Left),
			nativeEmitter.value(expression.Right),
		)
	case ir.Call:
		if expression.Accessor != "" {
			result := nativeEmitter.accessorCall(expression)
			if nativeEmitter.program.CallMayThrow(expression) {
				nativeEmitter.checkThrown()
			}
			return result
		}
		region := nativeEmitter.regionFor(expression)
		arguments := nativeEmitter.arguments(expression)
		call := nativeEmitter.callCode(expression, arguments)
		var result string
		switch {
		case region != "":
			// What it returns is in that region: no count to own (region.go).
			result = nativeEmitter.regionValue(fmt.Sprintf(
				"%s(%s)",
				nativeEmitter.regionFunctionName(expression.Function),
				strings.Join(append([]string{region}, arguments...), ", "),
			))
		case expression.Returns.IsReference():
			result = nativeEmitter.own(expression.Returns, call)
		default:
			result = nativeEmitter.temporary()
			nativeEmitter.line("%s %s = %s;", cType(expression.Returns), result, call)
		}
		if nativeEmitter.program.CallMayThrow(expression) {
			// A throw left the call: the result is the zero value it returned, owned like any.
			nativeEmitter.checkThrown()
		}
		return result
	case ir.HasAccessor:
		return nativeEmitter.snapshot(ir.Boolean, fmt.Sprintf(
			"adamic_accessor_find(%s, %s) != NULL",
			nativeEmitter.value(expression.Object),
			cString(expression.Name),
		))
	case ir.InstanceOf:
		value := nativeEmitter.value(expression.Value)
		if !expression.Value.Type().IsReference() {
			nativeEmitter.line("(void)%s;", value)
			return nativeEmitter.snapshot(ir.Boolean, "false")
		}
		if expression.Exact {
			return nativeEmitter.snapshot(ir.Boolean, fmt.Sprintf(
				"(%s != NULL && ((adamic_object *)%s)->class == &%s)",
				value,
				value,
				nativeEmitter.className(expression.Class),
			))
		}
		return nativeEmitter.snapshot(
			ir.Boolean,
			fmt.Sprintf("adamic_instanceof(%s, &%s)", value, nativeEmitter.className(expression.Class)),
		)
	case ir.NumberToString:
		return nativeEmitter.own(
			ir.String,
			fmt.Sprintf("adamic_string_from_number(%s)", nativeEmitter.value(expression.Value)),
		)
	case ir.BooleanToString:
		return fmt.Sprintf("((%s) ? &adamic_string_true : &adamic_string_false)", nativeEmitter.value(expression.Value))
	case ir.Concat:
		parts := make([]string, 0, len(expression.Parts))
		for _, part := range expression.Parts {
			parts = append(parts, nativeEmitter.value(part))
		}
		return nativeEmitter.own(ir.String, fmt.Sprintf(
			"adamic_string_concat(%d, (adamic_string *const[]){%s})",
			len(parts),
			strings.Join(parts, ", "),
		))
	case ir.Conditional:
		return nativeEmitter.conditional(expression)
	case ir.ObjectLiteral:
		return nativeEmitter.objectLiteral(expression)
	case ir.PhantomMember:
		return nativeEmitter.phantomMember(expression)
	case ir.Property:
		if expression.Object.Type() == ir.Array {
			return nativeEmitter.nodeArrayProperty(expression)
		}
		if expression.View != "" {
			return nativeEmitter.viewField(expression)
		}
		if expression.Readiness == "" {
			if taken, ok := nativeEmitter.take(expression); ok {
				return taken
			}
		}
		lent := nativeEmitter.lendable && lendable(expression.Of) && !expression.Optional
		object := nativeEmitter.value(expression.Object)
		slot := nativeEmitter.fieldSlot(object, expression.Name, expression.Class)
		if expression.Readiness != "" {
			slot = fmt.Sprintf(
				"adamic_object_read(%s, %s, &%s, %s)",
				object,
				cString(expression.Name),
				nativeEmitter.cache(),
				cString(expression.Readiness),
			)
			if expression.Optional {
				slot = fmt.Sprintf("(%s == NULL ? NULL : %s)", object, slot)
			}
		}
		field := unslotted(expression.Of, fmt.Sprintf("%s->%s", slot, member(expression.Of)))
		if expression.Of.IsMaybe() {
			if expression.Readiness != "" && !expression.Absent {
				nativeEmitter.line("(void)%s;", slot)
			}
			field = fmt.Sprintf(
				"adamic_object_maybe_number(%s, %s, &%s)",
				object,
				cString(expression.Name),
				nativeEmitter.cache(),
			)
			if expression.Of == ir.MaybeBoolean {
				field = fmt.Sprintf(
					"adamic_object_maybe_boolean(%s, %s, &%s)",
					object,
					cString(expression.Name),
					nativeEmitter.cache(),
				)
			}
		}
		if expression.Absent {
			slot := nativeEmitter.temporary()
			lookup := fmt.Sprintf(
				"adamic_object_optional_field(%s, %s, &%s)",
				object,
				cString(expression.Name),
				nativeEmitter.cache(),
			)
			if expression.Optional {
				lookup = fmt.Sprintf("(%s == NULL ? NULL : %s)", object, lookup)
			}
			nativeEmitter.line("adamic_value *%s = %s;", slot, lookup)
			if expression.Readiness != "" {
				nativeEmitter.line("if (%s != NULL) {", slot)
				nativeEmitter.line(
					"\t(void)adamic_object_read(%s, %s, &%s, %s);",
					object,
					cString(expression.Name),
					nativeEmitter.cache(),
					cString(expression.Readiness),
				)
				nativeEmitter.line("}")
			}
			undefined := "NULL"
			if expression.Of.IsMaybe() {
				undefined = zero(expression.Of)
			}
			present := unslotted(expression.Of, fmt.Sprintf("%s->%s", slot, member(expression.Of)))
			if expression.Of.IsMaybe() {
				present = field
			}
			field = fmt.Sprintf("(%s == NULL ? %s : %s)", slot, undefined, present)
		}
		if expression.Of.IsMaybe() && expression.Optional {
			return nativeEmitter.snapshot(
				expression.Of,
				fmt.Sprintf("(%s == NULL ? %s : %s)", object, zero(expression.Of), field),
			)
		}
		if expression.Of.IsReference() {
			// A field holds a reference as void *; read through the type the checker proved.
			field = fmt.Sprintf("((%s)%s)", cType(expression.Of), field)
		}
		if expression.Optional && expression.Type().IsMaybe() {
			return nativeEmitter.snapshot(
				expression.Type(),
				fmt.Sprintf("(%s == NULL ? %s : %s)", object, zero(expression.Type()), maybe(expression.Type(), field)),
			)
		}
		if expression.Optional {
			field = fmt.Sprintf("(%s == NULL ? NULL : %s)", object, field)
		}
		// Read now, when JavaScript reads it: a call later in the statement may write the field. A
		// reference is retained, so that write can't free it from under its reader, unless nothing
		// can run before it's used (borrow.go).
		if lent {
			nativeEmitter.self = true
			return nativeEmitter.snapshot(expression.Of, field)
		}
		if expression.Of.IsReference() {
			return nativeEmitter.own(expression.Of, fmt.Sprintf("adamic_retain(%s)", field))
		}
		return nativeEmitter.snapshot(expression.Of, field)
	case ir.Undefined:
		return "NULL"
	case ir.IsUndefined:
		if expression.Value.Type().IsMaybe() {
			return fmt.Sprintf("(!(%s).present)", nativeEmitter.value(expression.Value))
		}
		return fmt.Sprintf("(%s == NULL)", nativeEmitter.value(expression.Value))
	case ir.Unwrap:
		if expression.Proven {
			return fmt.Sprintf("(%s).%s", nativeEmitter.value(expression.Value), member(expression.Type()))
		}
		// The checker narrowed undefined away, but a call since may have put it back (ir.Unwrap).
		value := nativeEmitter.snapshot(expression.Value.Type(), nativeEmitter.value(expression.Value))
		nativeEmitter.line("if (!%s.present) {", value)
		nativeEmitter.line("\tstatic const char message[] = %s;", cArray(narrowedAwayMessage))
		nativeEmitter.line("\tadamic_panic(message, sizeof message - 1);")
		nativeEmitter.line("}")
		return fmt.Sprintf("(%s).%s", value, member(expression.Type()))
	case ir.Defined:
		value := nativeEmitter.value(expression.Value)
		nativeEmitter.checkDefined(value, expression.Message)
		return value
	case ir.MaybeOf:
		if expression.Value == nil {
			return zero(expression.Of)
		}
		return maybe(expression.Of, nativeEmitter.value(expression.Value))
	case ir.MaybeToString:
		return nativeEmitter.maybeToString(expression.Value)
	case ir.Box:
		return nativeEmitter.box(expression.Value)
	case ir.MakeError:
		return nativeEmitter.makeError(expression)
	case ir.WeakOf:
		return nativeEmitter.own(ir.Weak, fmt.Sprintf("adamic_weak_of(%s)", nativeEmitter.value(expression.Value)))
	case ir.WeakTarget:
		target := "adamic_weak_target"
		if expression.Present {
			target = "adamic_weak_target_present"
		}
		// Retained, as any reference read out of a slot is: a call later in the statement may let go
		// of the last strong holder.
		return nativeEmitter.own(expression.To, fmt.Sprintf(
			"(%s)adamic_retain(%s(%s))",
			cType(expression.To),
			target,
			nativeEmitter.value(expression.Value),
		))
	case ir.Narrow:
		return nativeEmitter.narrow(expression)
	case ir.ArrayIsArray:
		value := nativeEmitter.snapshot(ir.Union, nativeEmitter.value(expression.Value))
		return fmt.Sprintf("(%s != NULL && %s->kind == adamic_kind_array)", value, value)
	case ir.RecordStorage:
		// A record of numbers or booleans keeps its values unboxed, which a read by key into a union
		// would take for pointers. Only an entries record, whose values are references, is seen
		// through a fixed type (lower's recordStorageView), so another here is a road lowering missed.
		object := nativeEmitter.snapshot(ir.Object, nativeEmitter.value(expression.Value))
		const scalars = "compiler bug: a record of numbers or booleans reached a weak field read"
		nativeEmitter.line(
			"if (adamic_record_is(%s) && adamic_record_values(%s) != adamic_field_reference) adamic_panic(%s, %d);",
			object,
			object,
			cString(scalars),
			len(scalars),
		)
		return fmt.Sprintf("adamic_record_is(%s)", object)
	case ir.TypeOf:
		return nativeEmitter.typeOf(expression)
	case ir.UnionToString:
		return nativeEmitter.own(
			ir.String,
			fmt.Sprintf("adamic_union_to_string(%s)", nativeEmitter.value(expression.Value)),
		)
	case ir.UnknownKind:
		return nativeEmitter.snapshot(
			ir.Number,
			fmt.Sprintf("adamic_union_kind((const adamic_heap *)(%s))", nativeEmitter.value(expression.Value)),
		)
	case ir.OwnFunction:
		return nativeEmitter.snapshot(ir.Boolean, fmt.Sprintf(
			"adamic_object_own_function((const adamic_heap *)(%s), %s)",
			nativeEmitter.value(expression.Object),
			cString(expression.Name),
		))
	case ir.ErrorToString:
		return nativeEmitter.own(
			ir.String,
			fmt.Sprintf("adamic_error_to_string((const adamic_object *)(%s))", nativeEmitter.value(expression.Value)),
		)
	case ir.ArrayOfScalars:
		return nativeEmitter.snapshot(
			ir.Boolean,
			fmt.Sprintf("adamic_array_of_scalars((const adamic_heap *)(%s))", nativeEmitter.value(expression.Array)),
		)
	case ir.JoinEnter:
		return nativeEmitter.snapshot(
			ir.Boolean,
			fmt.Sprintf("adamic_join_enter((const adamic_heap *)(%s))", nativeEmitter.value(expression.Array)),
		)
	case ir.JoinLeave:
		return "adamic_join_leave()"
	case ir.Coalesce:
		return nativeEmitter.coalesce(expression)
	case ir.StringLength:
		if expression.Optional {
			text := nativeEmitter.value(expression.Value)
			return nativeEmitter.snapshot(ir.MaybeNumber, fmt.Sprintf(
				"(%s == NULL ? %s : (adamic_maybe_number){true, adamic_string_length(%s)})",
				text,
				zero(ir.MaybeNumber),
				text,
			))
		}
		return fmt.Sprintf("adamic_string_length(%s)", nativeEmitter.value(expression.Value))
	case ir.CharCodeAt:
		value := nativeEmitter.value(expression.Value)
		index := nativeEmitter.value(expression.Index)
		return fmt.Sprintf("adamic_string_char_code_at(%s, %s)", value, index)
	case ir.ObjectKeys, ir.ClosureSelf, ir.LibraryGlobal:
		return nativeEmitter.libraryLanguageValue(expression)
	case ir.MakeClosure:
		target := nativeEmitter.program.Functions[expression.Function]
		environment := target.Environment
		if target.NestedParent > 0 {
			identity := nativeEmitter.program.Functions[target.NestedParent-1].FrameIdentity
			if identity > 0 {
				cells := make([]string, 0, len(environment))
				for _, local := range environment {
					cells = append(cells, nativeEmitter.cellReference(local))
				}
				constructor := "adamic_closure_canonical"
				if nativeEmitter.program.PackedCountNeeded(expression.Function) {
					constructor = "adamic_counted_closure_canonical"
				}
				if target.GraphClosure {
					constructor += "_graph"
				}
				discard := ""
				if nativeEmitter.closureDiscards {
					// Stored once, where the closure is made, not on each return of the cached one.
					discard = ", " + nativeEmitter.discardOf(expression.Function)
				}
				closure := nativeEmitter.own(ir.Closure, fmt.Sprintf(
					"%s(%s, %s, %d, (adamic_cell *const[]){%s}%s)",
					constructor,
					nativeEmitter.cellReference(identity-1),
					nativeEmitter.functionName(expression.Function),
					len(environment),
					strings.Join(cells, ", "),
					discard,
				))
				nativeEmitter.line("%s->source_length = %d;", closure, target.SourceLength)
				return closure
			}
		}
		constructor := "adamic_closure_new"
		if nativeEmitter.program.PackedCountNeeded(expression.Function) {
			constructor = "adamic_counted_closure_new"
		}
		closure := nativeEmitter.own(
			ir.Closure,
			fmt.Sprintf("%s(%s, %d)", constructor, nativeEmitter.functionName(expression.Function), len(environment)),
		)
		nativeEmitter.line("%s->source_length = %d;", closure, target.SourceLength)
		if nativeEmitter.program.Functions[expression.Function].Receiver {
			nativeEmitter.line("%s->receiver = true;", closure)
		}
		nativeEmitter.closureDiscard(closure, expression.Function)
		nativeEmitter.adoptGraph(
			closure,
			fmt.Sprintf("sizeof *%s + %d * sizeof(adamic_cell *)", closure, len(environment)),
			target.GraphClosure && !expression.ProgramRegion && !target.ProgramRegion,
		)
		nativeEmitter.adoptProgram(
			closure,
			fmt.Sprintf("sizeof *%s + %d * sizeof %s->cells[0]", closure, len(environment), closure),
			expression.ProgramRegion || target.ProgramRegion,
		)
		for index, local := range environment {
			nativeEmitter.line(
				"%s->cells[%d] = %s;",
				closure,
				index,
				nativeEmitter.heldReferenceIn(closure, nativeEmitter.cellReference(local)),
			)
		}
		return closure
	case ir.CallClosure:
		property, isProperty := expression.Closure.(ir.Property)
		if !isProperty || !property.Method {
			return nativeEmitter.callThrough(expression, nativeEmitter.value(expression.Closure), "")
		}
		// object.name(...) through an interface. The object stays alive for the call though an
		// argument may write where it was read from: a call lends nothing (borrow.go), so a read of a
		// global, a captured variable, a field or an element is retained when it's read, and a local
		// can't be written in the middle of an expression.
		receiver := nativeEmitter.value(property.Object)
		if !property.Optional {
			return nativeEmitter.callThrough(expression, "", receiver)
		}
		// object?.name(...): undefined, the method not looked up and the arguments not evaluated,
		// where the object is undefined, as JavaScript's chain stops there.
		text, value, owned := nativeEmitter.asideWith(
			func() string { return nativeEmitter.callThrough(expression, "", receiver) },
		)
		result := "0"
		if expression.Returns != 0 {
			// Undefined: NULL for a reference, which zero isn't for a string.
			undefined := "NULL"
			if !expression.Returns.IsReference() {
				undefined = zero(expression.Returns)
			}
			result = nativeEmitter.temporary()
			nativeEmitter.line("%s %s = %s;", cType(expression.Returns), result, undefined)
		}
		nativeEmitter.line("if (%s != NULL) {", receiver)
		nativeEmitter.out.WriteString(text)
		nativeEmitter.indent++
		switch {
		case expression.Returns == 0:
		case expression.Returns.IsReference():
			nativeEmitter.line("%s = %s;", result, retained(value))
		default:
			nativeEmitter.line("%s = %s;", result, value)
		}
		for index := len(owned) - 1; index >= 0; index-- {
			nativeEmitter.line("adamic_release(%s);", owned[index])
		}
		nativeEmitter.indent--
		nativeEmitter.line("}")
		if expression.Returns.IsReference() {
			nativeEmitter.owned = append(nativeEmitter.owned, result)
		}
		return result
	case ir.ArrayMap:
		if mapped, ok := nativeEmitter.mapped(expression); ok {
			return mapped
		}
		source := nativeEmitter.temporary()
		nativeEmitter.line("adamic_array *%s = %s;", source, nativeEmitter.value(expression.Array))
		callback := nativeEmitter.value(expression.Callback)
		mapped := nativeEmitter.own(
			ir.Array,
			fmt.Sprintf("adamic_array_make(%s->length, %s)", source, arrayElement(expression.Result)),
		)
		nativeEmitter.adoptGraph(
			mapped,
			"sizeof *"+mapped,
			nativeEmitter.graphTypes(expression.GraphTypes) && !expression.ProgramRegion,
		)
		nativeEmitter.adoptProgram(mapped, "sizeof *"+mapped, expression.ProgramRegion)
		count, index := nativeEmitter.temporary(), nativeEmitter.temporary()
		// The length is read once, as JavaScript's map does; an array the callback shrinks is a
		// panic here rather than JavaScript's holes, which 0.1 has no way to hold.
		nativeEmitter.line("size_t %s = %s->length;", count, source)
		nativeEmitter.line("for (size_t %s = 0; %s < %s; %s++) {", index, index, count, index)
		nativeEmitter.line("\tif (%s >= %s->length) {", index, source)
		nativeEmitter.line("\t\tstatic const char message[] = \"map: the array shrank while it was being mapped\";")
		nativeEmitter.line("\t\tadamic_panic(message, sizeof message - 1);")
		nativeEmitter.line("\t}")
		nativeEmitter.indent++
		element := nativeEmitter.temporary()
		call := nativeEmitter.callbackCall(
			callback,
			expression.Callback,
			expression.CallbackType,
			fmt.Sprintf("%s->elements[%s]", source, index),
			fmt.Sprintf("{.number = (double)%s}", index),
			fmt.Sprintf("{.reference = %s}", source),
		)
		nativeEmitter.line("adamic_value %s = %s;", element, call)
		// What's mapped so far is the statement's, let go with its temporaries.
		nativeEmitter.closureThrown()
		if nativeEmitter.graphTypes(expression.GraphTypes) && expression.Result.IsReference() {
			nativeEmitter.line("%s.reference = adamic_graph_take(%s, %s.reference);", element, mapped, element)
		}
		nativeEmitter.line("adamic_array_push(%s, %s);", mapped, element)
		nativeEmitter.indent--
		nativeEmitter.line("}")
		return mapped
	case ir.ArrayVisit:
		return nativeEmitter.arrayVisit(expression)
	case ir.ArrayReduce:
		return nativeEmitter.arrayReduce(expression)
	case ir.ArraySearch:
		return nativeEmitter.libraryArraySearch(expression)
	case ir.ArrayFill:
		if expression.Array == nil {
			length := nativeEmitter.value(expression.Length)
			value := nativeEmitter.value(expression.Value)
			filled := fmt.Sprintf(
				"adamic_array_filled(%s, %s, %s)",
				length,
				borrowed(expression.Element, value),
				arrayElement(expression.Element),
			)
			if expression.MayThrow() {
				// A length out of range leaves a RangeError pending and makes no array, tested before
				// there's one to own.
				made := nativeEmitter.temporary()
				nativeEmitter.line("adamic_array *%s = %s;", made, filled)
				nativeEmitter.checkThrown()
				filled = made
			}
			return nativeEmitter.programArray(
				nativeEmitter.graphArray(filled, expression.GraphTypes, expression.ProgramRegion),
				expression.ProgramRegion,
			)
		}
		array := nativeEmitter.value(expression.Array)
		value := nativeEmitter.value(expression.Value)
		start, end := "0.0", "0.0"
		if expression.Start != nil {
			start = nativeEmitter.value(expression.Start)
		}
		if expression.End != nil {
			end = nativeEmitter.value(expression.End)
		}
		nativeEmitter.line(
			"adamic_array_fill(%s, %s, %s, %s, %t, %t);",
			array,
			borrowed(expression.Element, value),
			start,
			end,
			expression.Start != nil,
			expression.End != nil,
		)
		return array
	case ir.ArraySplice:
		return nativeEmitter.programArray(nativeEmitter.graphArray(fmt.Sprintf(
			"adamic_array_splice(%s, %s)",
			nativeEmitter.spliceArguments(expression),
			arrayElement(expression.Element),
		), expression.GraphTypes, expression.ProgramRegion), expression.ProgramRegion)
	case ir.ArrayFrom:
		return nativeEmitter.arrayFrom(expression)
	case ir.ArrayReverse:
		array := nativeEmitter.value(expression.Array)
		nativeEmitter.line("adamic_array_reverse(%s);", array)
		return array
	case ir.ArrayConcat:
		arrays := []string{nativeEmitter.value(expression.Array)}
		for _, other := range expression.Others {
			arrays = append(arrays, nativeEmitter.value(other))
		}
		return nativeEmitter.programArray(nativeEmitter.graphArray(fmt.Sprintf(
			"adamic_array_concat(%d, (adamic_array *const[]){%s}, %s)",
			len(arrays),
			strings.Join(arrays, ", "),
			arrayElement(expression.Element),
		), expression.GraphTypes, expression.ProgramRegion), expression.ProgramRegion)
	case ir.CheckedCast:
		object := nativeEmitter.temporary()
		nativeEmitter.line("adamic_object *%s = %s;", object, nativeEmitter.value(expression.Value))
		field := fmt.Sprintf(
			"adamic_object_field(%s, %s, &%s)->%s",
			object,
			cString(expression.Field),
			nativeEmitter.cache(),
			member(expression.FieldType),
		)
		if expression.CheckedFields {
			slot := nativeEmitter.temporary()
			nativeEmitter.line(
				"adamic_value %s = adamic_object_view(%s, %s, &%s, %d, %s, %s);",
				slot,
				object,
				cString(expression.Field),
				nativeEmitter.cache(),
				expression.FieldType,
				cString(map[ir.Type]string{
					ir.Number:  "number",
					ir.Boolean: "boolean",
					ir.String:  "string",
				}[expression.FieldType]),
				cString(expression.Field),
			)
			field = fmt.Sprintf("%s.%s", slot, member(expression.FieldType))
		}
		if expression.FieldType.IsReference() {
			field = fmt.Sprintf("((%s)%s)", cType(expression.FieldType), field)
		}
		tests := []string{}
		for _, allowed := range expression.Allowed {
			tests = append(
				tests,
				nativeEmitter.binary(ir.Equal, expression.FieldType, field, nativeEmitter.value(allowed)),
			)
		}
		nativeEmitter.line("if (!(%s)) {", strings.Join(tests, " || "))
		nativeEmitter.line("\tstatic const char message[] = %s;", cArray(expression.Message))
		nativeEmitter.line("\tadamic_panic(message, sizeof message - 1);")
		nativeEmitter.line("}")
		return object
	case ir.ArrayIndex:
		if expression.Array.Type().IsTypedArray() {
			array := nativeEmitter.value(expression.Array)
			index := nativeEmitter.value(expression.Index)
			lookup := "adamic_typed_array_get"
			if read, isRead := expression.Index.(ir.Read); isRead && nativeEmitter.integerCounter(read.Local) {
				lookup, index = "adamic_typed_array_get_integer", nativeEmitter.localName(read.Local)
			}
			return nativeEmitter.snapshot(ir.MaybeNumber, fmt.Sprintf("%s(%s, %s)", lookup, array, index))
		}
		slot := nativeEmitter.arrayIndexSlot(expression)
		if expression.Type().IsMaybe() {
			return nativeEmitter.snapshot(expression.Type(), maybeSlot(expression.Element, slot))
		}
		// Retained, so a write later in the statement can't free it from under its reader.
		return nativeEmitter.own(
			expression.Element,
			fmt.Sprintf("%s == NULL ? NULL : (%s)adamic_retain(%s->reference)", slot, cType(expression.Element), slot),
		)
	case ir.ArrayPop:
		array := nativeEmitter.temporary()
		nativeEmitter.line("adamic_array *%s = %s;", array, nativeEmitter.value(expression.Array))
		if expression.Type().IsMaybe() {
			popped := fmt.Sprintf("%s->elements[--%s->length].%s", array, array, member(expression.Element))
			if expression.Element == ir.MaybeNumber {
				popped = unslotted(ir.MaybeNumber, popped)
			} else {
				popped = maybe(expression.Type(), popped)
			}
			return nativeEmitter.snapshot(
				expression.Type(),
				fmt.Sprintf("%s->length == 0 ? %s : %s", array, zero(expression.Type()), popped),
			)
		}
		// The array's reference to the element becomes the statement's.
		popped := fmt.Sprintf("%s->elements[--%s->length].%s", array, array, member(expression.Element))
		if len(nativeEmitter.program.GraphTypes) != 0 {
			popped = fmt.Sprintf("adamic_graph_escape(%s, %s)", array, popped)
		}
		return nativeEmitter.own(expression.Element, fmt.Sprintf("%s->length == 0 ? NULL : %s", array, popped))
	case ir.MapEntries:
		pair := nativeEmitter.shapeOf([]string{"0", "1"}, []ir.Type{expression.KeyType, expression.ValueType})
		// map.c writes no representation bytes into a pair, so a run-time key read boxes a scalar slot
		// by the shape's registered types (union.c's keyed_tuple_slot), registered here as a literal's
		// are, not only once some literal of the same shape has been made.
		if nativeEmitter.dynamicProperties() {
			nativeEmitter.line("adamic_register_shape_types(&%s_metadata);", pair)
		}
		return nativeEmitter.programArray(nativeEmitter.graphArray(
			fmt.Sprintf("adamic_map_entries(%s, &%s)", nativeEmitter.value(expression.Map), pair),
			expression.GraphTypes,
			expression.ProgramRegion,
		), expression.ProgramRegion)
	case ir.ArraySlice:
		array := nativeEmitter.value(expression.Array)
		arguments := []string{"0.0", "0.0"}
		for index, argument := range expression.Arguments {
			arguments[index] = nativeEmitter.value(argument)
		}
		return nativeEmitter.programArray(nativeEmitter.graphArray(fmt.Sprintf(
			"adamic_array_slice(%s, %s, %s, %t, %s)",
			array,
			arguments[0],
			arguments[1],
			len(expression.Arguments) == 2,
			arrayElement(expression.Element),
		), expression.GraphTypes, expression.ProgramRegion), expression.ProgramRegion)
	case ir.ArraySort:
		array := nativeEmitter.value(expression.Array)
		sort := "adamic_array_sort"
		if expression.Element == ir.MaybeNumber {
			sort = "adamic_array_sort_undefined_last"
		}
		// A comparator that throws stops the sort, which leaves the array as it was, as V8's does
		// (sort.c), and the throw goes on from here.
		if expression.Callback != nil {
			nativeEmitter.line(
				"%s(%s, %s, %s);",
				sort,
				array,
				nativeEmitter.closureComparator(expression),
				nativeEmitter.value(expression.Callback),
			)
			nativeEmitter.closureThrown()
			return array
		}
		nativeEmitter.line("%s(%s, %s, NULL);", sort, array, nativeEmitter.comparator(expression))
		if nativeEmitter.program.ClosureMayThrow(expression) {
			nativeEmitter.checkThrown()
		}
		return array
	case ir.CodePoints:
		return nativeEmitter.own(
			ir.Array,
			fmt.Sprintf("adamic_string_code_points(%s)", nativeEmitter.value(expression.Value)),
		)
	case ir.StringIndex:
		value := nativeEmitter.value(expression.Value)
		index := nativeEmitter.value(expression.Index)
		return nativeEmitter.own(ir.String, fmt.Sprintf("adamic_string_at(%s, %s)", value, index))
	case ir.StringCall:
		return nativeEmitter.stringCall(expression)
	case ir.Trim:
		return nativeEmitter.own(
			ir.String,
			fmt.Sprintf("adamic_string_trim(%s)", nativeEmitter.value(expression.Value)),
		)
	case ir.StringIterator:
		return nativeEmitter.own(
			ir.Object,
			fmt.Sprintf("adamic_string_iterator(%s)", nativeEmitter.value(expression.Value)),
		)
	case ir.CollectionIterator:
		return nativeEmitter.collectionIterator(expression)
	case ir.MapNew:
		entries := make([][2]string, 0, len(expression.Entries))
		for _, entry := range expression.Entries {
			entries = append(entries, [2]string{nativeEmitter.value(entry[0]), nativeEmitter.value(entry[1])})
		}
		var pairs string
		if expression.Pairs != nil {
			pairs = nativeEmitter.value(expression.Pairs)
		}
		created := nativeEmitter.own(ir.Map, newMap(expression.Key, expression.Value.IsReference()))
		nativeEmitter.adoptGraph(
			created,
			"sizeof *"+created,
			nativeEmitter.graphTypes(expression.GraphTypes) && !expression.ProgramRegion,
		)
		nativeEmitter.adoptProgram(created, "sizeof *"+created, expression.ProgramRegion)
		for _, entry := range entries {
			nativeEmitter.line(
				"adamic_map_set(%s, %s, %s);",
				created,
				nativeEmitter.heldIn(created, expression.Key, entry[0]),
				nativeEmitter.heldIn(created, expression.Value, entry[1]),
			)
		}
		if expression.Pairs != nil {
			nativeEmitter.line("adamic_map_add_pairs(%s, %s);", created, pairs)
		}
		return created
	case ir.MapKeys:
		return nativeEmitter.programArray(nativeEmitter.graphArray(
			fmt.Sprintf("adamic_map_keys(%s, %s)", nativeEmitter.value(expression.Map), arrayElement(expression.Key)),
			expression.GraphTypes,
			expression.ProgramRegion,
		), expression.ProgramRegion)
	case ir.MapValues:
		return nativeEmitter.programArray(nativeEmitter.graphArray(fmt.Sprintf(
			"adamic_map_values(%s, %s)",
			nativeEmitter.value(expression.Map),
			arrayElement(expression.Value),
		), expression.GraphTypes, expression.ProgramRegion), expression.ProgramRegion)
	case ir.MapClear:
		nativeEmitter.line("adamic_map_clear(%s);", nativeEmitter.value(expression.Map))
		return "0"
	case ir.MapForEach:
		return nativeEmitter.mapForEach(expression)
	case ir.SetNew:
		var values string
		if expression.Values != nil {
			values = nativeEmitter.value(expression.Values)
		}
		created := nativeEmitter.own(ir.Map, newMap(expression.Element, false))
		nativeEmitter.adoptGraph(
			created,
			"sizeof *"+created,
			nativeEmitter.graphTypes(expression.GraphTypes) && !expression.ProgramRegion,
		)
		nativeEmitter.adoptProgram(created, "sizeof *"+created, expression.ProgramRegion)
		if expression.Values != nil {
			nativeEmitter.line("adamic_set_add_all(%s, %s);", created, values)
		}
		return created
	case ir.SetAdd:
		set := nativeEmitter.value(expression.Set)
		value := nativeEmitter.value(expression.Value)
		nativeEmitter.line(
			"adamic_map_set(%s, %s, (adamic_value){.number = 0});",
			set,
			nativeEmitter.heldIn(set, expression.Element, value),
		)
		return set
	case ir.SetValues:
		return nativeEmitter.programArray(nativeEmitter.graphArray(fmt.Sprintf(
			"adamic_set_values(%s, %s)",
			nativeEmitter.value(expression.Set),
			arrayElement(expression.Element),
		), expression.GraphTypes, expression.ProgramRegion), expression.ProgramRegion)
	case ir.MapGet:
		slot := nativeEmitter.mapGetSlot(expression)
		if expression.Type().IsMaybe() {
			return nativeEmitter.snapshot(expression.Type(), maybeSlot(expression.ValueType, slot))
		}
		// Retained, so a set later in the same statement can't free it from under its reader.
		return nativeEmitter.own(
			expression.ValueType,
			fmt.Sprintf("%s == NULL ? NULL : adamic_retain(%s->%s)", slot, slot, member(expression.ValueType)),
		)
	case ir.MapSet:
		object := nativeEmitter.value(expression.Map)
		key := nativeEmitter.value(expression.Key)
		value := nativeEmitter.value(expression.Value)
		nativeEmitter.line(
			"adamic_map_set(%s, %s, %s);",
			object,
			nativeEmitter.heldIn(object, expression.KeyType, key),
			nativeEmitter.heldIn(object, expression.ValueType, value),
		)
		return object
	case ir.MapHas:
		object := nativeEmitter.value(expression.Map)
		key := nativeEmitter.value(expression.Key)
		return nativeEmitter.snapshot(
			ir.Boolean,
			fmt.Sprintf("(adamic_map_get(%s, %s) != NULL)", object, borrowed(expression.KeyType, key)),
		)
	case ir.MapDelete:
		object := nativeEmitter.value(expression.Map)
		key := nativeEmitter.value(expression.Key)
		result := nativeEmitter.temporary()
		nativeEmitter.line("bool %s = adamic_map_delete(%s, %s);", result, object, borrowed(expression.KeyType, key))
		return result
	case ir.MapSize:
		return nativeEmitter.snapshot(ir.Number, fmt.Sprintf("(double)%s->count", nativeEmitter.value(expression.Map)))
	case ir.HasOwn:
		return nativeEmitter.snapshot(ir.Boolean, fmt.Sprintf(
			"adamic_object_has(%s, %s)",
			nativeEmitter.value(expression.Object),
			nativeEmitter.value(expression.Key),
		))
	case ir.ParallelMap:
		return nativeEmitter.parallelMap(expression)
	case ir.ReadTextFile:
		return nativeEmitter.own(
			ir.Object,
			fmt.Sprintf("adamic_read_text_file(%s)", nativeEmitter.value(expression.Path)),
		)
	case ir.ProgramArguments:
		return nativeEmitter.own(ir.Array, "adamic_program_arguments()")
	case ir.Utf8Length:
		return fmt.Sprintf("adamic_utf8_length(%s)", nativeEmitter.value(expression.Text))
	case ir.Utf8At:
		text := nativeEmitter.value(expression.Text)
		return nativeEmitter.snapshot(
			ir.Number,
			fmt.Sprintf("adamic_utf8_at(%s, %s)", text, nativeEmitter.value(expression.Index)),
		)
	case ir.ReadDirectory:
		return nativeEmitter.own(
			ir.Object,
			fmt.Sprintf("adamic_read_directory(%s)", nativeEmitter.value(expression.Path)),
		)
	case ir.FileStatus:
		return nativeEmitter.own(ir.Object, fmt.Sprintf("adamic_file_status(%s)", nativeEmitter.value(expression.Path)))
	case ir.WriteTextFile:
		path := nativeEmitter.value(expression.Path)
		text := nativeEmitter.value(expression.Text)
		return nativeEmitter.own(ir.Object, fmt.Sprintf("adamic_write_text_file(%s, %s)", path, text))
	case ir.ArrayPush:
		array := nativeEmitter.value(expression.Array)
		value := nativeEmitter.value(expression.Value)
		if expression.Element.IsReference() {
			value = nativeEmitter.heldReferenceIn(array, value)
		}
		// The append happens here, in JavaScript's order, and the new length is the value.
		nativeEmitter.line(
			"adamic_array_push(%s, (adamic_value){.%s = %s});",
			array,
			member(expression.Element),
			slotted(expression.Element, value),
		)
		length := nativeEmitter.temporary()
		nativeEmitter.line("double %s = (double)%s->length;", length, array)
		return length
	case ir.ArrayJoin:
		if expression.Depth > 0 {
			return nativeEmitter.libraryArrayJoin(expression)
		}
		array := nativeEmitter.value(expression.Array)
		separator := nativeEmitter.value(expression.Separator)
		return nativeEmitter.own(
			ir.String,
			fmt.Sprintf("adamic_array_join(%s, %s, %s)", array, separator, joinKind(expression.Element)),
		)
	case ir.ArrayLiteral:
		if len(expression.Metadata) != 0 {
			return nativeEmitter.nodeArrayLiteral(expression)
		}
		if array, ok := nativeEmitter.constantRecords(expression); ok {
			return array
		}
		if spread, ok := nativeEmitter.spreadArray(expression); ok {
			return spread
		}
		if expression.Spread != nil {
			// A spread is iterated where it stands, before the elements after it are evaluated, so the
			// array is made first (which nothing can see) and each element appended as it comes.
			array := nativeEmitter.own(
				ir.Array,
				fmt.Sprintf("adamic_array_make(0, %s)", arrayElement(expression.Element)),
			)
			nativeEmitter.adoptGraph(
				array,
				"sizeof *"+array,
				nativeEmitter.graphTypes(expression.GraphTypes) && !expression.ProgramRegion,
			)
			nativeEmitter.adoptProgram(array, "sizeof *"+array, expression.ProgramRegion)
			for index, element := range expression.Elements {
				value := nativeEmitter.value(element)
				switch {
				case expression.Spread[index]:
					nativeEmitter.line("adamic_array_append(%s, %s);", array, value)
				case expression.Element.IsReference():
					nativeEmitter.line(
						"adamic_array_push(%s, (adamic_value){.reference = %s});",
						array,
						nativeEmitter.heldReferenceIn(array, value),
					)
				default:
					nativeEmitter.line(
						"adamic_array_push(%s, (adamic_value){.%s = %s});",
						array,
						member(expression.Element),
						slotted(expression.Element, value),
					)
				}
			}
			return array
		}
		if array, constant := nativeEmitter.constantArray(expression); constant {
			return array
		}
		elements := make([]string, 0, len(expression.Elements))
		for _, element := range expression.Elements {
			elements = append(elements, nativeEmitter.value(element))
		}
		array := nativeEmitter.own(
			ir.Array,
			fmt.Sprintf("adamic_array_make(%d, %s)", len(elements), arrayElement(expression.Element)),
		)
		nativeEmitter.adoptGraph(
			array,
			"sizeof *"+array,
			nativeEmitter.graphTypes(expression.GraphTypes) && !expression.ProgramRegion,
		)
		nativeEmitter.adoptProgram(array, "sizeof *"+array, expression.ProgramRegion)
		for _, element := range elements {
			if expression.Element.IsReference() {
				element = nativeEmitter.heldReferenceIn(array, element)
			}
			nativeEmitter.line(
				"adamic_array_push(%s, (adamic_value){.%s = %s});",
				array,
				member(expression.Element),
				slotted(expression.Element, element),
			)
		}
		return array
	case ir.Length:
		if expression.Array.Type().IsTypedArray() {
			array := nativeEmitter.value(expression.Array)
			length := fmt.Sprintf("adamic_typed_array_length(%s)", array)
			if expression.Optional {
				return nativeEmitter.snapshot(ir.MaybeNumber, fmt.Sprintf(
					"(%s == NULL ? %s : (adamic_maybe_number){true, %s})",
					array,
					zero(ir.MaybeNumber),
					length,
				))
			}
			return nativeEmitter.snapshot(ir.Number, length)
		}
		if expression.Optional {
			array := nativeEmitter.value(expression.Array)
			return nativeEmitter.snapshot(ir.MaybeNumber, fmt.Sprintf(
				"(%s == NULL ? %s : (adamic_maybe_number){true, (double)%s->length})",
				array,
				zero(ir.MaybeNumber),
				array,
			))
		}
		return nativeEmitter.snapshot(
			ir.Number,
			fmt.Sprintf("(double)%s->length", nativeEmitter.value(expression.Array)),
		)
	case ir.JSONEncode:
		return nativeEmitter.jsonEncode(expression)
	case ir.JSONDecode:
		return nativeEmitter.jsonDecode(expression)
	case ir.JSONStringify:
		return nativeEmitter.jsonStringify(expression)
	case ir.JSONNull:
		return "NULL"
	case ir.MathCall:
		return nativeEmitter.mathCall(expression)
	case ir.StringFromCodes:
		function := "adamic_string_from_char_codes"
		if expression.CodePoints {
			function = "adamic_string_from_code_points"
		}
		var result string
		if expression.Spread != nil {
			result = nativeEmitter.own(
				ir.String,
				fmt.Sprintf("%s_of(%s)", function, nativeEmitter.value(expression.Spread)),
			)
		} else {
			codes := make([]string, 0, len(expression.Codes))
			for _, code := range expression.Codes {
				codes = append(codes, nativeEmitter.value(code))
			}
			if len(codes) == 0 {
				result = nativeEmitter.own(ir.String, function+"(0, NULL)")
			} else {
				result = nativeEmitter.own(
					ir.String,
					fmt.Sprintf("%s(%d, (const double[]){%s})", function, len(codes), strings.Join(codes, ", ")),
				)
			}
		}
		// A code point out of range leaves a RangeError pending, and an empty result, let go with the
		// rest (from_codes.c).
		if expression.MayThrow() {
			nativeEmitter.checkThrown()
		}
		return result
	case ir.RecordCoalesce:
		return nativeEmitter.recordCoalesce(expression)
	case ir.RecordCall:
		return nativeEmitter.recordCall(expression)
	case ir.RecordLiteral:
		return nativeEmitter.dictionaryLiteral(expression)
	case ir.ObjectCall:
		return nativeEmitter.objectCall(expression)
	case ir.NumberCall:
		if value, known := nativeEmitter.libraryNumberCall(expression); known {
			return value
		}
		arguments := []string{}
		for _, argument := range expression.Arguments {
			arguments = append(arguments, nativeEmitter.value(argument))
		}
		switch expression.Function {
		case "parseInt":
			// A radix left out is undefined, which ToInt32 makes 0: detect it from the text.
			radix := "0.0"
			if len(arguments) == 2 {
				radix = arguments[1]
			}
			return nativeEmitter.snapshot(
				ir.Number,
				fmt.Sprintf("adamic_number_parse_int(%s, %s)", arguments[0], radix),
			)
		case "parseFloat":
			return nativeEmitter.snapshot(ir.Number, fmt.Sprintf("adamic_number_parse_float(%s)", arguments[0]))
		case "isNaN":
			return fmt.Sprintf("isnan(%s)", arguments[0])
		case "isFinite":
			return fmt.Sprintf("isfinite(%s)", arguments[0])
		case "isInteger":
			return fmt.Sprintf("(isfinite(%s) && trunc(%s) == %s)", arguments[0], arguments[0], arguments[0])
		}
		return fmt.Sprintf(
			"(isfinite(%s) && trunc(%s) == %s && fabs(%s) <= 9007199254740991.0)",
			arguments[0],
			arguments[0],
			arguments[0],
			arguments[0],
		)
	case ir.ToFixed:
		value := nativeEmitter.value(expression.Value)
		digits := nativeEmitter.value(expression.Digits)
		result := nativeEmitter.own(ir.String, fmt.Sprintf("adamic_number_to_fixed(%s, %s)", value, digits))
		// Digits out of range leave a RangeError pending, and an empty result, let go with the rest.
		if expression.MayThrow() {
			nativeEmitter.checkThrown()
		}
		return result
	case ir.NumberFormat:
		return nativeEmitter.numberFormat(expression)
	}
	panic(fmt.Sprintf("native: no C for %T", expression))
}

// cBitwise are the runtime's bitwise operators (bitwise.c).
var cBitwise = map[ir.Operator]string{
	ir.BitAnd:             "adamic_bitwise_and",
	ir.BitOr:              "adamic_bitwise_or",
	ir.BitXor:             "adamic_bitwise_xor",
	ir.ShiftLeft:          "adamic_shift_left",
	ir.ShiftRight:         "adamic_shift_right",
	ir.ShiftRightUnsigned: "adamic_shift_right_unsigned",
}

var cOperators = map[ir.Operator]string{
	ir.Add: "+", ir.Subtract: "-", ir.Multiply: "*", ir.Divide: "/",
	ir.Less: "<", ir.LessOrEqual: "<=", ir.Greater: ">", ir.GreaterOrEqual: ">=",
	ir.Equal: "==", ir.NotEqual: "!=",
}

func (nativeEmitter *emitter) binary(operator ir.Operator, operandType ir.Type, left string, right string) string {
	switch {
	case operator == ir.Remainder:
		// JavaScript's % is C's fmod: truncated, with the sign of the dividend.
		return fmt.Sprintf("fmod(%s, %s)", left, right)
	case operator == ir.Power:
		return fmt.Sprintf("adamic_power(%s, %s)", left, right)
	case cBitwise[operator] != "":
		return fmt.Sprintf("%s(%s, %s)", cBitwise[operator], left, right)
	case operandType == ir.String &&
		(operator == ir.Less || operator == ir.LessOrEqual || operator == ir.Greater || operator == ir.GreaterOrEqual):
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
	if operator == ir.Equal || operator == ir.NotEqual {
		return nativeEmitter.scalarEquality(operator, operandType, left, right)
	}
	return fmt.Sprintf("(%s %s %s)", left, cOperators[operator], right)
}

// narrowedAwayMessage is the panic of a number or a boolean the checker narrowed undefined out of,
// read where a call since put it back (ir.Unwrap). JavaScript would go on computing with undefined.
const narrowedAwayMessage = "undefined where the checker narrowed it away: a call since the narrowing put it back"
