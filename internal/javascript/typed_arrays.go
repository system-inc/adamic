package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func typedArrayName(kind ir.Type) string {
	switch kind {
	case ir.Uint16Array:
		return "Uint16Array"
	case ir.Uint8Array:
		return "Uint8Array"
	case ir.Int32Array:
		return "Int32Array"
	case ir.Float64Array:
		return "Float64Array"
	}
	panic(fmt.Sprintf("javascript: no typed array constructor for %d", kind))
}
