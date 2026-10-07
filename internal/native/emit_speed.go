package native

import (
	"fmt"
	"strings"
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
