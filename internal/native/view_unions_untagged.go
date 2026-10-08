package native

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// This named adapter is a shared-dispatch handoff. It uses already evaluated
// objects and shared slot metadata; field reads retain their original guards.
func (e *emitter) viewUntaggedObjectUnion(property ir.Property, object string) {
	if property.ViewType == "" && property.ViewContract > 0 {
		property.ViewType = e.program.ViewContracts[property.ViewContract-1].Name
	}

	name := e.temporary()
	var declarations strings.Builder
	declarations.WriteString("#include \"view_unions_untagged.h\"\n")
	rows := []string{}
	for i, contract := range e.program.ViewContracts {
		prefix := fmt.Sprintf("%s_%d", name, i)
		allowed, fields, members := "NULL", "NULL", "NULL"
		if len(contract.Allowed) != 0 {
			literals := []string{}
			for _, literal := range contract.Allowed {
				kind, payload := "adamic_view_union_number", ".number = "+strconv.FormatFloat(literal.Number, 'g', -1, 64)
				if literal.Of == ir.Boolean {
					kind = "adamic_view_union_boolean"
					payload = fmt.Sprintf(".boolean = %t", literal.Boolean)
				}
				if literal.Of == ir.String {
					kind = "adamic_view_union_string"
					text := prefix + "_text_" + strconv.Itoa(len(literals))
					fmt.Fprintf(&declarations, "static adamic_string %s = ADAMIC_STRING(%s);\n", text, cString(literal.String))
					payload = ".reference = &" + text
				}
				literals = append(literals, fmt.Sprintf("{.kind=%s,.literal=true,.value={%s}}", kind, payload))
			}
			allowed = prefix + "_allowed"
			fmt.Fprintf(&declarations, "static const adamic_view_union_member %s[] = {%s};\n", allowed, strings.Join(literals, ","))
		}
		if len(contract.Fields) != 0 {
			values := []string{}
			for _, field := range contract.Fields {
				values = append(values, fmt.Sprintf("{%s,%d,%t}", cString(field.Name), field.Contract, field.Optional))
			}
			fields = prefix + "_fields"
			fmt.Fprintf(&declarations, "static const adamic_view_untagged_field %s[] = {%s};\n", fields, strings.Join(values, ","))
		}
		if len(contract.Members) != 0 {
			values := []string{}
			for _, id := range contract.Members {
				values = append(values, strconv.Itoa(int(id)))
			}
			members = prefix + "_members"
			fmt.Fprintf(&declarations, "static const size_t %s[] = {%s};\n", members, strings.Join(values, ","))
		}
		kind := contract.Kind
		if contract.Nominal != "" || contract.Unsupported != "" && contract.Unsupported != "untagged object union" {
			kind = ir.ViewUnknown
		}
		rows = append(rows, fmt.Sprintf("{%d,%d,%t,%s,%d,%s,%d,%s,%d,%d}", kind, contract.Of, contract.Undefined, allowed, len(contract.Allowed), fields, len(contract.Fields), members, len(contract.Members), contract.Element))
	}
	fmt.Fprintf(&declarations, "static const adamic_view_untagged_contract %s[] = {%s};\n", name, strings.Join(rows, ","))
	e.declarations = append(e.declarations, declarations.String())
	e.line("(void)adamic_view_untagged_plain_select((const adamic_object *)%s, %s, %d, %d, %s, %s);", object, name, len(rows), property.ViewContract, cString(property.View), cString(property.ViewType))
}

// Preserve the existing shared-discriminant dispatch.
func untaggedObjectUnion(contracts []ir.ViewContract, contract ir.ViewContract) bool {
	return len(contract.Members) != 0 && !ir.ViewUnionHasDiscriminant(contracts, contract)
}
