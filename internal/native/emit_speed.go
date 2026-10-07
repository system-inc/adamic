package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// writeFieldSlot uses the same layout proof as reads. Static constructor
// objects still need write_field to mark inherited fields as own properties.
// A constructor object can arrive through a structural view, so programs with
// static layouts retain that runtime distinction even when the field is uniform.
func (e *emitter) writeFieldSlot(object, name string, class int) string {
	if !cName.MatchString(object) {
		object = e.snapshotObjectForStore(object)
	}
	if !strings.HasPrefix(name, "#") {
		e.line("if (%s->frozen) {", object)
		e.line("\tadamic_object_check_write(%s, %s);", object, cString(name))
		e.line("}")
	}
	field := e.fieldSlot(object, name, class)
	for _, metadata := range e.program.Classes {
		if metadata.Static {
			field = fmt.Sprintf("(%s->class != NULL && %s->class->is_static ? adamic_object_write_field(%s, %s, &%s) : %s)", object, object, object, cString(name), e.cache(), field)
			break
		}
	}
	slot := e.temporary()
	e.line("adamic_value *%s = %s;", slot, field)
	return slot
}

func (e *emitter) snapshotObjectForStore(object string) string {
	name := e.temporary()
	e.line("adamic_object *%s = %s;", name, object)
	return name
}

// primitiveArrayStore elides the reference-ownership branch only where the
// element type proves there is no reference. Index validation is unchanged;
// invalid writes use the runtime's exact diagnostic after all operands ran.
func (e *emitter) primitiveArrayStore(element ir.Type, array, index, value string) bool {
	if element != ir.Boolean && element != ir.Number && element != ir.MaybeNumber {
		return false
	}
	slot := e.temporary()
	e.line("adamic_value *%s = adamic_array_at(%s, %s);", slot, array, index)
	e.line("if (%s != NULL) {", slot)
	e.line("\t*%s = (adamic_value){.%s = %s};", slot, member(element), slotted(element, value))
	e.line("} else {")
	e.line("\tadamic_array_set(%s, %s, (adamic_value){.%s = %s});", array, index, member(element), slotted(element, value))
	e.line("}")
	return true
}

// booleanLiteralEquality needs only the optional operand's tag: the literal's
// tag is proven present. The optional operand is evaluated once and its missing
// case remains unequal even to false.
func (e *emitter) booleanLiteralEquality(binary ir.Binary) (string, bool) {
	if binary.Left.Type() != ir.MaybeBoolean || (binary.Operator != ir.Equal && binary.Operator != ir.NotEqual) {
		return "", false
	}
	operand := binary.Left
	literal, known := binary.Right.(ir.MaybeOf)
	if !known {
		return "", false
	}
	boolean, known := literal.Value.(ir.BooleanConstant)
	if !known {
		return "", false
	}
	value := e.snapshot(ir.MaybeBoolean, e.value(operand))
	equal := fmt.Sprintf("(%s.present && %s.boolean == %t)", value, value, boolean.Value)
	if binary.Operator == ir.NotEqual {
		equal = "(!" + equal + ")"
	}
	return equal, true
}
