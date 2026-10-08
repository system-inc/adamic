package native

import "github.com/system-inc/adamic/internal/ir"

// Assignment results retain their own count. A later operand may overwrite the
// target, so moving the result's count into the target would leave a dangling value.
func (e *emitter) assignmentValue(expression ir.AssignmentValue) string {
	switch store := expression.Store.(type) {
	case ir.Assign:
		value, stored := e.assignmentOperands(expression, store.Value.Type())
		if store.Checked {
			e.checkReady(store.Local)
		}
		e.store(store.Local, stored, false)
		return value
	case ir.SetIndex:
		array := e.value(store.Array)
		index := e.snapshot(ir.Number, e.value(store.Index))
		value, stored := e.assignmentOperands(expression, store.Value.Type())
		if store.Array.Type().IsTypedArray() {
			e.line("adamic_typed_array_set(%s, %s, %s);", array, index, stored)
		} else {
			e.line("adamic_array_set(%s, %s, %s);", array, index, held(store.Element, stored))
		}
		return value
	case ir.SetProperty:
		object := e.value(store.Object)
		value, stored := e.assignmentOperands(expression, store.Value.Type())
		e.line("if (%s == NULL) {", object)
		e.line("\tstatic const char message[] = %s;", cString("TypeError: Cannot set properties of undefined (setting '"+store.Name+"')"))
		e.line("\tadamic_panic(message, sizeof message - 1);")
		e.line("}")
		e.line("adamic_object_check_data_write(%s, %s);", object, cString(store.Name))
		slot := e.temporary()
		e.line("adamic_value *%s = %s;", slot, e.writeFieldSlot(object, store.Name, store.Class))
		if store.Value.Type().IsReference() {
			old := e.temporary()
			e.line("void *%s = %s->reference;", old, slot)
			e.line("%s->reference = %s;", slot, retained(stored))
			e.line("adamic_release(%s);", old)
		} else {
			e.line("%s->%s = %s;", slot, member(store.Value.Type()), slotted(store.Value.Type(), stored))
		}
		return value
	}
	panic("native: assignment value without a store")
}

// Existing ownership optimizations reason about statement-level writes. Until
// they model stores within operands, use counted heap ownership for these programs.
func hasAssignmentValues(program *ir.Program) bool {
	return program.AssignmentValues
}

func (e *emitter) assignmentOperands(expression ir.AssignmentValue, storage ir.Type) (string, string) {
	raw := expression.Value
	if raw == nil {
		switch store := expression.Store.(type) {
		case ir.Assign:
			raw = store.Value
		case ir.SetProperty:
			raw = store.Value
		case ir.SetIndex:
			raw = store.Value
		}
	}
	value := e.snapshot(raw.Type(), e.value(raw))
	if _, undefined := raw.(ir.Undefined); undefined {
		if storage.IsMaybe() {
			return value, zero(storage)
		}
		if storage.IsReference() && storage != ir.Weak {
			return value, "(" + cType(storage) + ")NULL"
		}
	}
	stored, fresh := converted(raw.Type(), storage, value)
	if fresh {
		stored = e.own(storage, stored)
	}
	return value, stored
}
