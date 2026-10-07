package native

import (
	"crypto/sha256"
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
	condition := object + " == NULL"
	if !strings.HasPrefix(name, "#") {
		condition += " || " + object + "->frozen"
	}
	e.line("if (%s) {", condition)
	e.line("\t%s(%s);", e.coldFieldFailure(name), object)
	e.line("}")
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
	e.line("\tadamic_array_write_failure(%s, %s);", array, index)
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

// coldHelper uses content names only for these new internal helpers. It leaves
// user symbols and temporary/cache counter allocation unchanged.
func (e *emitter) coldHelper(kind, key, parameters, body string) string {
	name := fmt.Sprintf("adamic_cold_%s_%x", kind, sha256.Sum256([]byte(key)))
	declaration := fmt.Sprintf("static ADAMIC_COLD _Noreturn void %s(%s) {\n%s\n}\n", name, parameters, body)
	prefix := "static ADAMIC_COLD _Noreturn void " + name + "("
	for _, existing := range e.declarations {
		if strings.HasPrefix(existing, prefix) {
			if existing != declaration {
				panic("native: compiler bug: cold helper name collision")
			}
			return name
		}
	}
	e.declarations = append(e.declarations, declaration)
	return name
}

// coldStop shares a constant-message failure body, including its argument setup,
// instead of repeating that setup in each generated caller.
func (e *emitter) coldStop(message string) string {
	body := fmt.Sprintf("\tstatic const char message[] = %s;\n\tadamic_panic(message, sizeof message - 1);", cString(message))
	return e.coldHelper("stop", message, "void", body)
}

func (e *emitter) coldFieldFailure(name string) string {
	message := "TypeError: Cannot set properties of undefined (setting '" + name + "')"
	body := fmt.Sprintf("\tif (object == NULL) {\n\t\tstatic const char message[] = %s;\n\t\tadamic_panic(message, sizeof message - 1);\n\t}\n", cString(message))
	if !strings.HasPrefix(name, "#") {
		body += fmt.Sprintf("\tadamic_object_check_write(object, %s);\n", cString(name))
	}
	body += "\tadamic_unreachable();"
	return e.coldHelper("field", name, "const adamic_object *object", body)
}

// programHasExpression is conservative across calls and aliases: every lowered
// function body participates, even when no call from main has been proved.
// Runtime-created layouts must be included whenever their producing API occurs.
func programHasExpression(program *ir.Program, matches func(ir.Expression) bool) bool {
	found := false
	walkExpressions(program, func(expression ir.Expression) {
		if !found && matches(expression) {
			found = true
		}
	})
	return found
}
