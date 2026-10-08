package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) mapViewCertificate(property ir.Property, value string) string {
	id := property.ViewContract
	if id == 0 {
		return value
	}
	c := e.program.ViewContracts[id-1]
	if c.Kind == ir.ViewNullable {
		id = c.Element
		if id == 0 {
			return value
		}
		c = e.program.ViewContracts[id-1]
	}
	if c.Kind != ir.ViewMap {
		return value
	}
	pairs := []string{}
	for _, p := range ir.MapCertificatePairs(e.program, id) {
		pairs = append(pairs, fmt.Sprintf("[%d,%d]", p[0], p[1]))
	}
	return fmt.Sprintf("adamicMapView(%s,[%s],%s,%s)", value, strings.Join(pairs, ","), quote(property.View), quote(property.ViewType))
}
