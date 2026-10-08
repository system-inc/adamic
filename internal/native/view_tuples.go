package native

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) viewTuple(property ir.Property, value string) bool {
	if property.ViewContract == 0 {
		return false
	}
	contract := e.program.ViewContracts[property.ViewContract-1]
	if !contract.FixedTuple {
		return false
	}
	e.declarations = append(e.declarations, "#include \"view_tuples.h\"")
	e.line("adamic_view_tuple((const adamic_object *)%s, %d, %s, %s);", value, len(contract.Tuple), cString(contract.Name), cString(property.View))
	return true
}
