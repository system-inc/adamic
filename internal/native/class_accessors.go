package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) accessorDeclarations(builder *strings.Builder, classID int, class ir.Class) string {
	if len(class.Accessors) == 0 {
		return "NULL"
	}
	entries := []string{}
	for i, accessor := range class.Accessors {
		getter, setter := "NULL", "NULL"
		of := ir.Type(0)
		if accessor.Getter >= 0 {
			getter = fmt.Sprintf("adamic_getter_%d_%d", classID, i)
			function := e.program.Functions[accessor.Getter]
			of = function.Returns
			fmt.Fprintf(builder, "static adamic_value %s(adamic_object *object) {\n", getter)
			self := "object"
			if e.reuse.consumed[function.Parameters[0]] {
				self = "adamic_retain(object)"
			}
			if function.Closure {
				packed := e.closureSlots(ir.CallClosure{Closure: ir.MakeClosure{Function: accessor.Getter}}, []string{fmt.Sprintf("{.reference = %s}", self)}, "", "1")
				count := ""
				if e.program.PackedCountNeeded(accessor.Getter) {
					count = ", 1"
				}
				fmt.Fprintf(builder, "\tadamic_slot_cache cache = {NULL, 0};\n\tadamic_closure *closure = adamic_object_field(object, %s, &cache)->reference;\n\treturn %s(closure, %s%s);\n", cString(fmt.Sprintf("#accessor:%d", accessor.Getter)), e.functionName(accessor.Getter), packed, count)
			} else {
				count := ""
				if function.ArgumentsCount != 0 {
					count = ", 0"
				}
				code := fmt.Sprintf("%s(%s%s)", e.functionName(accessor.Getter), self, count)
				fmt.Fprintf(builder, "\treturn (adamic_value){.%s = %s};\n", member(of), slotted(of, code))
			}
			builder.WriteString("}\n")
		}
		if accessor.Setter >= 0 {
			setter = fmt.Sprintf("adamic_setter_%d_%d", classID, i)
			function := e.program.Functions[accessor.Setter]
			input := e.program.Locals[function.Parameters[1]].Type
			fmt.Fprintf(builder, "static void %s(adamic_object *object, adamic_value value, int type) {\n\t(void)type;\n", setter)
			raw := unslotted(input, "value."+member(input))
			fresh := false
			if input == ir.MaybeNumber {
				raw = fmt.Sprintf("(type == %d ? (adamic_maybe_number){true, value.number} : %s)", ir.Number, raw)
			}
			if input == ir.Union {
				raw = fmt.Sprintf("(type == %d ? adamic_box_number(value.number) : type == %d ? (value.boolean ? &adamic_box_true.heap : &adamic_box_false.heap) : value.reference)", ir.Number, ir.Boolean)
				fresh = true
			}
			fmt.Fprintf(builder, "\t%s incoming = (%s)(%s);\n", cType(input), cType(input), raw)
			self, argument := "object", "incoming"
			if e.reuse.consumed[function.Parameters[0]] {
				self = "adamic_retain(object)"
			}
			consumed := e.reuse.consumed[function.Parameters[1]]
			if consumed && input.IsReference() {
				if fresh {
					argument = fmt.Sprintf("(type == %d ? incoming : adamic_retain(incoming))", ir.Number)
				} else {
					argument = "adamic_retain(incoming)"
				}
			}
			if function.Closure {
				packed := e.closureSlots(ir.CallClosure{Closure: ir.MakeClosure{Function: accessor.Setter}}, []string{fmt.Sprintf("{.reference = %s}", self), fmt.Sprintf("{.%s = %s}", member(input), slotted(input, argument))}, "", "2")
				count := ""
				if e.program.PackedCountNeeded(accessor.Setter) {
					count = ", 2"
				}
				fmt.Fprintf(builder, "\tadamic_slot_cache cache = {NULL, 0};\n\tadamic_closure *closure = adamic_object_field(object, %s, &cache)->reference;\n\t(void)%s(closure, %s%s);\n", cString(fmt.Sprintf("#accessor:%d", accessor.Setter)), e.functionName(accessor.Setter), packed, count)
			} else {
				count := ""
				if function.ArgumentsCount != 0 {
					count = ", 1"
				}
				fmt.Fprintf(builder, "\t%s(%s, %s%s);\n", e.functionName(accessor.Setter), self, argument, count)
			}
			if fresh && !consumed {
				fmt.Fprintf(builder, "\tif (type == %d) { adamic_release(incoming); }\n", ir.Number)
			}
			builder.WriteString("}\n")
		}
		entries = append(entries, fmt.Sprintf("{%s, %s, %s, %d}", cString(accessor.Name), getter, setter, of))
	}
	name := fmt.Sprintf("adamic_accessors_%d", classID)
	fmt.Fprintf(builder, "static const adamic_accessor %s[] = {%s};\n", name, strings.Join(entries, ", "))
	return name
}

// The adapter boundary borrows its inputs. Individual adapters implement the target's ownership
// convention, so literal closures, consuming methods and borrowed methods can share a descriptor.
func (e *emitter) accessorCall(call ir.Call) string {
	object := e.value(call.Arguments[0])
	if call.Setter {
		argument := call.Arguments[1]
		value := e.value(argument)
		e.line("adamic_accessor_set(%s, %s, (adamic_value){.%s = %s}, %d);", object, cString(call.Accessor), member(argument.Type()), slotted(argument.Type(), value), argument.Type())
		return "0"
	}
	descriptor := e.temporary()
	e.line("const adamic_accessor *%s = adamic_accessor_find(%s, %s);", descriptor, object, cString(call.Accessor))
	result := e.temporary()
	e.line("adamic_value %s = adamic_accessor_get(%s, %s);", result, object, cString(call.Accessor))
	value := unslotted(call.Returns, result+"."+member(call.Returns))
	switch call.Returns {
	case ir.MaybeNumber:
		value = fmt.Sprintf("(%s->type == %d ? (adamic_maybe_number){true, %s.number} : %s)", descriptor, ir.Number, result, value)
	case ir.Number:
		value = fmt.Sprintf("(%s->type == %d ? (%s).number : %s->type == %d ? ((const adamic_number_box *)%s.reference)->number : %s.number)", descriptor, ir.MaybeNumber, unslotted(ir.MaybeNumber, result+".number"), descriptor, ir.Union, result, result)
	case ir.Boolean:
		value = fmt.Sprintf("(%s->type == %d ? ((const adamic_boolean_box *)%s.reference)->boolean : %s.boolean)", descriptor, ir.Union, result, result)
	case ir.Union:
		value = fmt.Sprintf("(%s->type == %d ? adamic_box_number(%s.number) : %s->type == %d ? (%s.boolean ? &adamic_box_true.heap : &adamic_box_false.heap) : %s.reference)", descriptor, ir.Number, result, descriptor, ir.Boolean, result, result)
	default:
		e.line("(void)%s;", descriptor)
	}
	if call.Returns.IsReference() {
		return e.own(call.Returns, fmt.Sprintf("(%s)(%s)", cType(call.Returns), value))
	}
	return e.snapshot(call.Returns, value)
}
