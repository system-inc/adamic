package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) extractedMethodValue(property ir.Property, object string) string {
	value := e.own(ir.Closure, fmt.Sprintf("adamic_object_method_value(%s, %s, &%s, %t)", object, cString(property.Name), e.cache(), property.Absent))
	if property.Bound != nil {
		receiver := e.value(property.Bound)
		return e.own(ir.Closure, fmt.Sprintf("adamic_method_bind(%s, %s)", value, receiver))
	}
	return value
}
