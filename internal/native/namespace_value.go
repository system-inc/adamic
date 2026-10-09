package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) namespaceField(property ir.Property) string {
	e.declarations = append(e.declarations, `#include "namespace.h"`)
	object := e.value(property.Object)
	slot := e.temporary()
	e.line("adamic_value %s = adamic_namespace_read(%s, %s, %d);", slot, object, e.recordKey(property.Name), property.Of)
	value := unslotted(property.Of, fmt.Sprintf("%s.%s", slot, member(property.Of)))
	if property.Of.IsReference() {
		return e.own(property.Of, fmt.Sprintf("(%s)%s", cType(property.Of), value))
	}
	return e.snapshot(property.Of, value)
}
