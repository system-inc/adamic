package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strconv"
	"strings"
)

// The aggregate's checker-combined fields retain every constituent obligation,
// including intersections on duplicate names. Every physical slot is read once.
func (e *emitter) viewObjectIntersection(property ir.Property, object string) {
	e.viewIntersectionFields(property.ViewContract, object, property.View, map[ir.ViewContractID]bool{})
}

func (e *emitter) viewIntersectionFields(id ir.ViewContractID, object, expression string, active map[ir.ViewContractID]bool) {
	if active[id] {
		return
	} // Recursive descendants keep their checks on subsequent reads.
	active[id] = true
	defer delete(active, id)
	contract := e.program.ViewContracts[id-1]
	for _, field := range contract.Fields {
		child := e.program.ViewContracts[field.Contract-1]
		// Unsupported descendants are refused by lazy demand at their own reads.
		if child.Unsupported != "" || child.Kind == ir.ViewUnknown || child.Kind == ir.ViewCallable || child.Kind == ir.ViewUndefined || child.Of == ir.Union {
			continue
		}
		slot := e.temporary()
		path := expression + "." + field.Name
		optional := field.Optional || child.Undefined
		e.line("adamic_value %s = adamic_object_optional_view(%s, %s, &%s, %d, %s, %s, %t, %t);", slot, object, cString(field.Name), e.cache(), child.Of, cString(child.Name), cString(path), optional, optional)
		// A snapshot used only for validation must still compile with -Werror.
		e.line("(void)%s;", slot)
		if child.Of == ir.Object {
			if optional {
				e.line("if (%s.reference != NULL) {", slot)
			}
			nested := fmt.Sprintf("(const adamic_object *)%s.reference", slot)
			if child.Kind == ir.ViewUnion {
				e.viewObjectUnion(ir.Property{View: path, ViewContract: field.Contract}, nested)
			} else {
				e.viewIntersectionFields(field.Contract, nested, path, active)
			}
			if optional {
				e.line("}")
			}
		}
		if len(child.Allowed) != 0 {
			actualSlot := slot
			literalType := child.Of
			present := "true"
			if child.Of == ir.MaybeNumber {
				unpacked := fmt.Sprintf("adamic_maybe_number_unpack(%s.number)", slot)
				actualSlot = e.temporary()
				e.line("adamic_value %s = {.number = %s.number};", actualSlot, unpacked)
				present = unpacked + ".present"
				literalType = ir.Number
			}
			if child.Of == ir.MaybeBoolean {
				actualSlot = e.temporary()
				e.line("adamic_value %s = {.boolean = %s.reference != NULL && ((const adamic_boolean_box *)%s.reference)->boolean};", actualSlot, slot, slot)
				present = slot + ".reference != NULL"
				literalType = ir.Boolean
			}

			tests := []string{}
			for _, literal := range child.Allowed {
				switch literal.Of {
				case ir.String:
					name := e.temporary()
					e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(literal.String)))
					tests = append(tests, fmt.Sprintf("adamic_string_equal((const adamic_string *)%s.reference, &%s)", actualSlot, name))
				case ir.Number:
					tests = append(tests, fmt.Sprintf("%s.number == %s", actualSlot, strconv.FormatFloat(literal.Number, 'g', -1, 64)))
				case ir.Boolean:
					tests = append(tests, fmt.Sprintf("%s.boolean == %t", actualSlot, literal.Boolean))
				}
			}
			test := strings.Join(tests, " || ")
			if literalType != child.Of {
				test = "!(" + present + ") || (" + test + ")"
			}
			if optional && child.Of.IsReference() {
				test = slot + ".reference == NULL || (" + test + ")"
			}
			e.line("if (!(%s)) adamic_view_literal_failure(%s, %s, %d, %s);", test, cString(path), cString(child.Name), literalType, actualSlot)
		}
	}
}
