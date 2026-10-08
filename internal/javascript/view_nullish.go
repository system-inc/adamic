package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) nullishViewField(property ir.Property) string {
	value := fmt.Sprintf("adamicViewNullish(%s, %s, %s, %s, %d, %t, %t, [%s], %t, %t)", e.value(property.Object), quote(property.Name), quote(property.View), quote(property.ViewType), property.NullishKinds, property.NullAllowed, property.UndefinedAllowed, e.values(property.ViewAllowed), property.Absent, property.Optional)
	return e.viewIntersectionNullishRead(property, value)
}
