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
	if !ir.HasMapNominalWitness(e.program, id) {
		return
	}
	stored := e.temporary()
	e.line("const void *%s = (const void *)%s;", stored, value)
	accepted := []string{fmt.Sprintf("%s != NULL && ((const adamic_heap *)%s)->kind == adamic_kind_object", stored, stored)}
	if contract.NominalClass != 0 {
		accepted = []string{fmt.Sprintf("adamic_instanceof(%s, &adamic_class_%d)", stored, contract.NominalClass)}
	}
	if nullable {
		accepted = append(accepted, fmt.Sprintf("%s == &adamic_null", stored))
	}
	if undefined {
		accepted = append(accepted, fmt.Sprintf("%s == NULL", stored))
	}
	message := cString("Map nominal producer failed: " + where + " expected " + contract.Name + ", found value without its class identity")
	e.line("if (!(%s)) adamic_panic(%s, sizeof %s - 1);", strings.Join(accepted, " || "), message, message)
	for _, field := range contract.Fields {
		if !ir.HasMapNominalWitness(e.program, field.Contract) {
			continue
		}
		e.line("if (%s != NULL && %s != &adamic_null) {", stored, stored)
		e.indent++
		child := e.temporary()
		schema := e.program.ViewContracts[field.Contract-1]
		reading := child + ".reference"
		if schema.Kind == ir.ViewNullable || schema.Undefined {
			e.declarations = append(e.declarations, "#include \"view_nullish.h\"")
			present, _, _ := mapNominalContract(e.program, field.Contract)
			reading = child
			e.line("adamic_heap *%s = %s;", child, fmt.Sprintf("adamic_object_nullish_view((adamic_object *)%s,%s,&%s,%d,%t,%t,false,false,%s,%s)", stored, cString(field.Name), e.cache(), 1<<present.Of, schema.Null, schema.Undefined, cString(where+"."+field.Name), cString(schema.Name)))
		} else {
			e.line("adamic_value %s = adamic_object_view((adamic_object *)%s,%s,&%s,%d,%s,%s);", child, stored, cString(field.Name), e.cache(), schema.Of, cString(schema.Name), cString(where+"."+field.Name))
		}
		e.mapEntryNominalCertificate(field.Contract, reading, where+"."+field.Name)
		if schema.Kind == ir.ViewNullable || schema.Undefined {
			e.line("adamic_release(%s);", reading)
		}
		e.indent--
		e.line("}")
	}
}

func (e *emitter) mapPairNominalCertificates(key, value ir.ViewContractID, pairs string) {
	for position, id := range []ir.ViewContractID{key, value} {
		if !ir.HasMapNominalWitness(e.program, id) {
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
