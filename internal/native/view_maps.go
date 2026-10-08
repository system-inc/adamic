package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func contractIDs(ids []ir.ViewContractID) (string, int) {
	values := []string{}
	for _, id := range ids {
		values = append(values, fmt.Sprint(id))
	}
	if len(values) == 0 {
		return "NULL", 0
	}
	return "(const unsigned int[]){" + strings.Join(values, ",") + "}", len(values)
}
func (e *emitter) mapViewCertificate(property ir.Property, value string) {
	id := property.ViewContract
	if id == 0 {
		return
	}
	c := e.program.ViewContracts[id-1]
	if c.Kind == ir.ViewNullable {
		id = c.Element
		if id == 0 {
			return
		}
		c = e.program.ViewContracts[id-1]
	}
	if c.Kind != ir.ViewMap {
		return
	}
	flat := []ir.ViewContractID{}
	for _, pair := range ir.MapCertificatePairs(e.program, id) {
		flat = append(flat, pair[0], pair[1])
	}
	pointer, count := contractIDs(flat)
	e.declarations = append(e.declarations, `#include "view_maps.h"`)
	e.line("if (%s != NULL && (adamic_heap *)%s != &adamic_null) adamic_map_view_certificate((const adamic_map *)%s, %s, %d, %s, %s);", value, value, value, pointer, count/2, cString(property.View), cString(property.ViewType))
}
