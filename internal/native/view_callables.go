package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) emitViewCallableRead(property ir.Property, receiver, method string) string {
	return e.own(ir.Closure, fmt.Sprintf("adamic_retain(adamic_view_callable(%s, %s, &%s, &%s, %s))", receiver, cString(property.Name), e.cache(), method, cString(property.View)))
}
