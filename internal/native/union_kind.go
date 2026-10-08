package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// A typeof object result does not prove the layout a narrowing reads.
func (e *emitter) kindIs(expression ir.KindIs) string {
	kind := ""
	switch expression.Of {
	case ir.Object:
		kind = "object"
	case ir.Array:
		kind = "array"
	case ir.Map:
		kind = "map"
	default:
		panic("native: unsupported union object kind")
	}
	value := e.snapshot(ir.Union, e.value(expression.Value))
	if expression.Of == ir.Object {
		// Library iterator references share the Object representation.
		return fmt.Sprintf("(%s != NULL && (%s->kind == adamic_kind_%s || %s->kind == adamic_kind_map_iterator || %s->kind == adamic_kind_typed_array_iterator))", value, value, kind, value, value)
	}
	return fmt.Sprintf("(%s != NULL && %s->kind == adamic_kind_%s)", value, value, kind)
}
