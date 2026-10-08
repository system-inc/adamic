package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) shapeContracts(shape string, fields []ir.Field) string {
	rows := []string{}
	for index, field := range fields {
		contract := field.Contract
		if contract == nil {
			rows = append(rows, "{0, false, NULL, 0, NULL}")
			continue
		}
		allowed := "NULL"
		if len(contract.Allowed) > 0 {
			values := []string{}
			for _, value := range contract.Allowed {
				values = append(values, fmt.Sprintf("{.%s = %s}", member(value.Type()), e.value(value)))
			}
			allowed = fmt.Sprintf("%s_allowed_%d", shape, index)
			e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_value %s[] = {%s};", allowed, strings.Join(values, ", ")))
		}
		rows = append(rows, fmt.Sprintf("{%d, %t, %s, %d, %s}", contract.Kind, contract.Nullable, cString(contract.Declared), len(contract.Allowed), allowed))
	}
	name := shape + "_contracts"
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_field_contract %s[] = {%s};", name, strings.Join(rows, ", ")))
	return name
}
