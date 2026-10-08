package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) viewArraySourceCertificate(expression ir.Expression, value string) {
	switch source := expression.(type) {
	case ir.ArrayHoles:
		e.line("adamic_view_array_element_certificate(%s, %d);", value, source.ElementContract)
	case ir.ArrayLiteral:
		e.line("adamic_view_array_element_certificate(%s, %d);", value, source.ElementContract)
	case ir.ObjectLiteral:
		e.line("adamic_view_array_object_certificate(%s, %d);", value, source.ArrayWriteContract)
	}
}

func (e *emitter) viewArrayReferenceWrite(array, value string) {
	incoming := e.temporary()
	e.line("const void *%s = (const void *)%s;", incoming, value)
	value = incoming
	e.line("if (%s->element_kind != 10 || !%s->references) adamic_view_array_storage_check(%s,4,%s);", array, array, array, cString("<array write>"))
	checked := e.temporary()
	e.line("bool %s = false;", checked)
	e.line("switch (%s->element_contract) {", array)
	for index, source := range e.program.ViewContracts {
		if ir.PrimitiveArrayContract(e.program, ir.ViewContractID(index+1)) {
			e.line("case %d:", index+1)
			e.arrayPrimitiveUnionRead(ir.ViewContractID(index+1), value, "<array write>", source.Name)
			e.line("%s = true; break;", checked)
			continue
		}
		if source.Of != ir.Object && source.Of != ir.Union {
			continue
		}
		class, _, _ := mapNominalContract(e.program, ir.ViewContractID(index+1))
		if class.NominalClass == 0 {
			continue
		}
		e.line("case %d:", index+1)
		e.nominalViewRead(ir.ViewContractID(index+1), value, "<array write>", false)
		e.line("%s = true; break;", checked)
	}
	e.line("default: break; }")
	e.line("if (!%s) {", checked)
	e.viewArrayReferencePolicy()
	e.line("adamic_view_array_reference_write(%s, %s, adamic_array_reference_pairs, %d, adamic_array_reference_names, %d);", array, "(const adamic_object *)"+value, len(ir.ArrayRecordWritePairs(e.program)), len(e.program.ViewContracts)+1)
	e.line("adamic_verify_array_record((const adamic_object *)%s);", value)
	e.line("}")
}

// Generate one complete scalar-record policy per compiled program. Type names
// are diagnostics only; compatibility uses immutable contract ids and structure.
func (e *emitter) viewArrayReferencePolicy() {
	for _, declaration := range e.declarations {
		if strings.HasPrefix(declaration, "static const adamic_array_write_pair adamic_array_reference_pairs") {
			return
		}
	}
	pairs := []string{}
	for _, pair := range ir.ArrayRecordWritePairs(e.program) {
		pairs = append(pairs, fmt.Sprintf("{%d,%d}", pair[0], pair[1]))
	}
	if len(pairs) == 0 {
		pairs = append(pairs, "{0,0}")
	}
	names := []string{cString("uncertified contract")}
	for _, contract := range e.program.ViewContracts {
		names = append(names, cString(contract.Name))
	}
	e.declarations = append(e.declarations, "static const adamic_array_write_pair adamic_array_reference_pairs[] = {"+strings.Join(pairs, ",")+"};", "static const char *const adamic_array_reference_names[] = {"+strings.Join(names, ",")+"};")
	lines := []string{"static void adamic_verify_array_record(const adamic_object *value) {", "switch (value->array_write_contract) {"}
	for index, contract := range e.program.ViewContracts {
		if !ir.FlatArrayRecordContract(e.program, ir.ViewContractID(index+1)) {
			continue
		}
		lines = append(lines, fmt.Sprintf("case %d: {", index+1))
		for position, field := range contract.Fields {
			child := e.program.ViewContracts[field.Contract-1]
			call := fmt.Sprintf("adamic_object_view(value, %s, &cache_%d, %d, %s, %s)", cString(field.Name), position, child.Of, cString(child.Name), cString("<array write>."+field.Name))
			lines = append(lines, fmt.Sprintf("adamic_slot_cache cache_%d = {NULL,0};", position))
			if len(child.Allowed) == 0 {
				lines = append(lines, "(void)"+call+";")
				continue
			}
			snapshot := fmt.Sprintf("slot_%d", position)
			lines = append(lines, "adamic_value "+snapshot+" = "+call+";")
			tests := []string{}
			for literalIndex, literal := range child.Allowed {
				switch literal.Of {
				case ir.Number:
					tests = append(tests, fmt.Sprintf("%s.number == %s", snapshot, cNumber(literal.Number)))
				case ir.Boolean:
					tests = append(tests, fmt.Sprintf("%s.boolean == %t", snapshot, literal.Boolean))
				case ir.String:
					name := fmt.Sprintf("literal_%d_%d", position, literalIndex)
					lines = append(lines, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(literal.String)))
					tests = append(tests, fmt.Sprintf("adamic_string_equal(%s.reference, &%s)", snapshot, name))
				}
			}
			lines = append(lines, fmt.Sprintf("if (!(%s)) adamic_view_literal_failure(%s, %s, %d, %s);", strings.Join(tests, " || "), cString("<array write>."+field.Name), cString(child.Name), child.Of, snapshot))
		}
		lines = append(lines, "return; }")
	}
	lines = append(lines, "default: adamic_unreachable();", "}", "}")
	e.declarations = append(e.declarations, strings.Join(lines, "\n"))
}
