package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) viewTuple(property ir.Property, value string) (string, bool) {
	if property.ViewContract == 0 {
		return value, false
	}
	contract := e.program.ViewContracts[property.ViewContract-1]
	if !contract.FixedTuple {
		return value, false
	}
	if contract.TupleVariable {
		upper := fmt.Sprintf("v.length <= %d", len(contract.Tuple))
		if contract.TupleRest != 0 {
			upper = "true"
		}
		return fmt.Sprintf("((v) => {if (!(Array.isArray(v) && v.length >= %d && %s)) panic('cast failed: field read failed: '+%s+' is not a '+%s+'; expected '+%s+', found '+(v === undefined ? 'undefined' : Array.isArray(v) ? 'array' : typeof v)); return v;})(%s)", contract.TupleMinimum, upper, quote(property.View), quote(contract.Name), quote(contract.Name), value), true
	}
	return fmt.Sprintf("((v) => { if (!(Array.isArray(v) && v.length === %d)) panic('cast failed: field read failed: ' + %s + ' is not a ' + %s + '; expected ' + %s + ', found ' + (v === undefined ? 'undefined' : Array.isArray(v) ? 'array' : typeof v)); return v; })(%s)", len(contract.Tuple), quote(property.View), quote(contract.Name), quote(contract.Name), value), true
}

func (e *emitter) tupleViewRepresentation(contract ir.ViewContractID, of ir.Type) ir.Type {
	if contract != 0 && (e.program.ViewContracts[contract-1].FixedTuple || e.program.ViewContracts[contract-1].TupleUnion) {
		return ir.Array
	}
	return of
}

func (e *emitter) tupleViewContract(id ir.ViewContractID) bool {
	return id != 0 && e.program.ViewContracts[id-1].FixedTuple
}
