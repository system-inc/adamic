package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// The first copy has already observed the spread's getters and own fields.
// Merge its actual layout, including fields hidden by a structural view, before
// evaluating replacements. Layout metadata has process lifetime, like static shapes.
func (e *emitter) extendedSpread(literal ir.ObjectLiteral, snapshot string) string {
	e.declarations = append(e.declarations, `#include "object_spread_extend.h"`)
	return e.own(ir.Object, fmt.Sprintf("adamic_object_spread_extend(%s, &%s)", snapshot, e.shape(literal.Fields)))
}
