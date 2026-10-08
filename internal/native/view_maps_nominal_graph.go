package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Cyclic producer descriptors are emitted as data, never unrolled recursively.
func (e *emitter) mapNominalGraphCertificate(id ir.ViewContractID, value, where string) {
	table := e.temporary() + "_nominal_schemas"
	e.declarations = append(e.declarations, "#include \"view_nominal_graph.h\"")
	var rows []string
	for index := range e.program.ViewContracts {
		slot := ir.ViewContractID(index + 1)
		if !ir.HasMapNominalWitness(e.program, slot) {
			rows = append(rows, "{0}")
			continue
		}
		contract, nullable, undefined := mapNominalContract(e.program, slot)
		class := "NULL"
		if contract.NominalClass != 0 {
			class = fmt.Sprintf("&adamic_class_%d", contract.NominalClass)
			e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_class adamic_class_%d;", contract.NominalClass))
		}
		fields := []string{}
		for _, field := range contract.Fields {
			if ir.HasMapNominalWitness(e.program, field.Contract) {
				fields = append(fields, fmt.Sprintf("{%s,%d,%t}", cString(field.Name), field.Contract-1, field.Optional))
			}
		}
		pointer := "NULL"
		if len(fields) != 0 {
			pointer = fmt.Sprintf("%s_fields_%d", table, index)
			e.declarations = append(e.declarations, "static const adamic_nominal_graph_field "+pointer+"[]={"+strings.Join(fields, ",")+"};")
		}
		rows = append(rows, fmt.Sprintf("{%s,%s,%t,%t,%s,%d}", cString(contract.Name), class, nullable, undefined, pointer, len(fields)))
	}
	e.declarations = append(e.declarations, "static const adamic_nominal_graph_schema "+table+"[]={"+strings.Join(rows, ",")+"};")
	e.line("adamic_nominal_graph_check((const void *)%s,%d,%s,%s);", value, id-1, table, cString(where))
}
