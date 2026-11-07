package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// A named error keeps the existing Error ownership and uncaught formatting convention.
func (nativeEmitter *emitter) makeError(expression ir.MakeError) string {
	object := nativeEmitter.own(ir.Object, fmt.Sprintf("adamic_error_new(%s)", nativeEmitter.value(expression.Message)))
	if expression.Name != nil {
		name := nativeEmitter.value(expression.Name)
		slot := nativeEmitter.temporary()
		nativeEmitter.line(
			"adamic_value *%s = adamic_object_field(%s, \"name\", &%s);",
			slot,
			object,
			nativeEmitter.cache(),
		)
		nativeEmitter.line("adamic_release(%s->reference);", slot)
		nativeEmitter.line("%s->reference = %s;", slot, nativeEmitter.kept(name))
	}
	return object
}
