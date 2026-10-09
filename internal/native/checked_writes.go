package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) shapeContracts(shape string, fields []ir.Field, allocationType int) string {
	// The header precedes the field-indexed entries. Runtime shapes with no contracts use NULL.
	rows := []string{fmt.Sprintf("{.type_id=%d}", allocationType)}
	for index, field := range fields {
		contract := field.Contract
		if contract == nil {
			rows = append(rows, "{0}")
			continue
		}
		rows = append(rows, e.contractDeclaration(fmt.Sprintf("%s_contract_%d", shape, index), contract))
	}
	name := shape + "_contracts"
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_field_contract %s[] = {%s};", name, strings.Join(rows, ", ")))
	return "&" + name + "[1]"
}

func (e *emitter) contractDeclaration(name string, contract *ir.FieldContract) string {
	allowed := "NULL"
	if len(contract.Allowed) > 0 {
		values := []string{}
		for _, value := range contract.Allowed {
			values = append(values, fmt.Sprintf("{.%s = %s}", member(value.Type()), e.value(value)))
		}
		allowed = name + "_allowed"
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_value %s[] = {%s};", allowed, strings.Join(values, ", ")))
	}
	writeProofs := e.contractProofs(name+"_write_proofs", contract.ProvenWrites)
	fieldProofs := e.contractProofs(name+"_field_proofs", contract.ProvenFields)
	fieldNames, optional, children := "NULL", "NULL", "NULL"
	if len(contract.Fields) > 0 {
		names, flags, rows := []string{}, []string{}, []string{}
		for index, field := range contract.Fields {
			names = append(names, cString(field.Name))
			flags = append(flags, fmt.Sprint(field.Optional))
			rows = append(rows, e.contractDeclaration(fmt.Sprintf("%s_field_%d", name, index), field.Contract))
		}
		fieldNames = name + "_names"
		optional = name + "_optional"
		children = name + "_fields"
		e.declarations = append(e.declarations, fmt.Sprintf("static const char *const %s[] = {%s};", fieldNames, strings.Join(names, ", ")), fmt.Sprintf("static const bool %s[] = {%s};", optional, strings.Join(flags, ", ")), fmt.Sprintf("static const adamic_field_contract %s[] = {%s};", children, strings.Join(rows, ", ")))
	}
	return fmt.Sprintf("{.nullish_only=%t, .kind=%d, .nullable=%t, .declared=%s, .count=%d, .allowed=%s, .type_id=%d, .reference=%t, .structural=%t, .write_proof_count=%d, .write_proofs=%s, .field_proof_count=%d, .field_proofs=%s, .field_count=%d, .field_names=%s, .field_optional=%s, .field_contracts=%s}", contract.NullishOnly, contract.Kind, contract.Nullable, cString(contract.Declared), len(contract.Allowed), allowed, contract.TypeID, contract.Reference, contract.Structural, len(contract.ProvenWrites), writeProofs, len(contract.ProvenFields), fieldProofs, len(contract.Fields), fieldNames, optional, children)
}

func (e *emitter) contractProofs(name string, proofs []int) string {
	if len(proofs) == 0 {
		return "NULL"
	}
	values := []string{}
	for _, proof := range proofs {
		values = append(values, fmt.Sprint(proof))
	}
	e.declarations = append(e.declarations, fmt.Sprintf("static const int %s[] = {%s};", name, strings.Join(values, ", ")))
	return name
}

func (e *emitter) contractContainer(expression ir.ContractContainer) string {
	value := e.value(expression.Value)
	contract := "NULL"
	if expression.Contract != nil {
		name := e.temporary() + "_element_contract"
		declaration := e.contractDeclaration(name, expression.Contract)
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_field_contract %s = %s;", name, declaration))
		contract = "&" + name
	}
	e.line("%s->element_contract = %s;", value, contract)
	e.line("%s->allocation_type = %d;", value, expression.AllocationType)
	return value
}

func (e *emitter) contractResult(expression ir.ContractResult) string {
	value := e.value(expression.Value)
	object := e.temporary()
	e.line("adamic_object *%s = %s;", object, value)
	for _, field := range expression.Fields {
		name := e.temporary() + "_result_contract"
		declaration := e.contractDeclaration(name, field.Contract)
		e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_field_contract %s = %s;", name, declaration))
		slot, cache := e.temporary(), e.cache()
		e.line("adamic_value *%s = adamic_object_read(%s, %s, &%s, %s);", slot, object, cString(field.Name), cache, cString(expression.Expression+"."+field.Name))
		e.line("adamic_check_contract(&%s, adamic_object_field_types(%s)[%s.index], *%s, 0, %s);", name, object, cache, slot, cString(expression.Expression+"."+field.Name))
	}
	return object
}
