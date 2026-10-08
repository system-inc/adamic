package native

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) mapEntryCallableCertificate(id ir.ViewContractID, value, where string) {
	if id == 0 {
		return
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Kind == ir.ViewNullable {
		if contract.Element == 0 {
			return
		}
		child := e.program.ViewContracts[contract.Element-1]
		if child.Kind != ir.ViewCallable {
			return
		}
		stored := e.temporary()
		e.line("adamic_heap *%s = (adamic_heap *)%s;", stored, value)
		e.line("if (%s != NULL && %s != &adamic_null) {", stored, stored)
		e.indent++
		e.mapEntryCallableCertificate(contract.Element, value, where)
		e.indent--
		e.line("}")
		return
	}
	if contract.Kind != ir.ViewCallable {
		return
	}
	property := ir.Property{Of: ir.Closure, ViewContract: id, View: where}
	e.line("(void)%s;", e.emitViewCallableCertificate(property, "((adamic_closure *)"+value+")"))
}

func (e *emitter) mapPairCallableCertificates(id ir.ViewContractID, pairs string) {
	if id == 0 {
		return
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Kind == ir.ViewNullable && contract.Element != 0 {
		contract = e.program.ViewContracts[contract.Element-1]
	}
	if contract.Kind != ir.ViewCallable {
		return
	}
	index, cache, value := e.temporary(), e.temporary(), e.temporary()
	e.line("adamic_slot_cache %s = {0};", cache)
	e.line("for (size_t %s = 0; %s < %s->length; %s++) {", index, index, pairs, index)
	e.indent++
	e.line("void *%s = adamic_object_field((adamic_object *)%s->elements[%s].reference, \"1\", &%s)->reference;", value, pairs, index, cache)
	e.mapEntryCallableCertificate(id, value, "Map constructor entry")
	e.indent--
	e.line("}")
}
