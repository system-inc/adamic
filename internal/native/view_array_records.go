package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) emitViewArrayRecord(record ir.ArrayRecord) string {
	array := e.value(record.Array)
	properties := e.value(record.Properties)
	// Only a fresh array reaches this hook; the properties are copied so source
	// aliases can mutate independently. The existing array destructor owns them.
	e.line("%s->properties = adamic_object_copy_checked(%s, %s);", array, properties, cString(record.Where))
	e.line("%s->properties->real_type = \"array\";", array)
	return e.own(ir.Array, fmt.Sprintf("adamic_retain(%s)", array))
}

func (e *emitter) emitViewArrayProperties(properties ir.ArrayProperties) string {
	array := e.value(properties.Array)
	return e.own(ir.Object, fmt.Sprintf("adamic_retain(%s->properties)", array))
}
