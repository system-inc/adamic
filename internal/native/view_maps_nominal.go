package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func mapNominalContract(program *ir.Program, id ir.ViewContractID) (ir.ViewContract, bool, bool) {
	nullable, undefined := false, false
	if id == 0 {
		return ir.ViewContract{}, false, false
	}
	contract := program.ViewContracts[id-1]
	if contract.Kind == ir.ViewNullable {
		nullable, undefined = contract.Null, contract.Undefined
		if contract.Element == 0 {
			return ir.ViewContract{}, false, false
		}
		contract = program.ViewContracts[contract.Element-1]
	}
	return contract, nullable, undefined || contract.Undefined
}

func (e *emitter) mapEntryNominalCertificate(id ir.ViewContractID, value, where string) {
	contract, nullable, undefined := mapNominalContract(e.program, id)
	if contract.NominalClass == 0 {
		return
	}
	stored := e.temporary()
	e.line("const void *%s = (const void *)%s;", stored, value)
	accepted := []string{fmt.Sprintf("adamic_instanceof(%s, &adamic_class_%d)", stored, contract.NominalClass)}
	if nullable {
		accepted = append(accepted, fmt.Sprintf("%s == &adamic_null", stored))
	}
	if undefined {
		accepted = append(accepted, fmt.Sprintf("%s == NULL", stored))
	}
	message := cString("Map nominal producer failed: " + where + " expected " + contract.Name + ", found value without its class identity")
	e.line("if (!(%s)) adamic_panic(%s, sizeof %s - 1);", strings.Join(accepted, " || "), message, message)
}

func (e *emitter) mapPairNominalCertificates(key, value ir.ViewContractID, pairs string) {
	for position, id := range []ir.ViewContractID{key, value} {
		contract, _, _ := mapNominalContract(e.program, id)
		if contract.NominalClass == 0 {
			continue
		}
		index, cache, entry := e.temporary(), e.temporary(), e.temporary()
		e.line("adamic_slot_cache %s = {0};", cache)
		e.line("for (size_t %s = 0; %s < %s->length; %s++) {", index, index, pairs, index)
		e.indent++
		e.line("void *%s = adamic_object_field((adamic_object *)%s->elements[%s].reference, %s, &%s)->reference;", entry, pairs, index, cString(fmt.Sprint(position)), cache)
		e.mapEntryNominalCertificate(id, entry, "Map copied entry")
		e.indent--
		e.line("}")
	}
}
