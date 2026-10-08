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
	if e.program.ViewContracts[property.ViewContract-1].IntersectionBounded {
		e.viewBoundedIntersection(property, object)
		return
	}
	if e.program.ViewContracts[property.ViewContract-1].IntersectionRecursive {
		e.viewRecursiveIntersection(property, object)
		return
	}
	e.viewIntersectionFields(property.ViewContract, object, property.View, map[ir.ViewContractID]bool{})
}

func (e *emitter) viewIntersectionFields(id ir.ViewContractID, object, expression string, active map[ir.ViewContractID]bool, skip ...string) {
	if active[id] {
		return
	} // Recursive descendants keep their checks on subsequent reads.
	active[id] = true
	defer delete(active, id)
	contract := e.program.ViewContracts[id-1]
	for _, field := range contract.Fields {
		if len(skip) > 0 && field.Name == skip[0] {
			continue
		}
		child := e.program.ViewContracts[field.Contract-1]
		// Unsupported descendants are refused by lazy demand at their own reads.
		if child.Unsupported != "" || child.Kind == ir.ViewUnknown || child.Kind == ir.ViewUndefined || child.Of == ir.Union {
			continue
		}
		slot := e.temporary()
		path := expression + "." + field.Name
		optional := field.Optional || child.Undefined
		e.line("adamic_value %s = adamic_object_optional_view_undefined(%s, %s, &%s, %d, %s, %s, %t, false, %t);", slot, object, cString(field.Name), e.cache(), child.Of, cString(child.Name), cString(path), field.Optional, child.Undefined)
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

// Validate the selector snapshot once; it then discharges the selected arm's
// tag obligation without reevaluating the physical slot during the field walk.
func (e *emitter) viewIntersectionUnion(property ir.Property, object string) {
	root := e.program.ViewContracts[property.ViewContract-1]
	if root.IntersectionBounded {
		e.viewBoundedIntersection(property, object)
		return
	}
	var tag ir.ViewContract
	for _, field := range root.Fields {
		if field.Name == root.IntersectionTag {
			tag = e.program.ViewContracts[field.Contract-1]
		}
	}
	slot := e.temporary()
	path := property.View + "." + root.IntersectionTag
	e.line("adamic_value %s = adamic_object_view(%s, %s, &%s, %d, %s, %s);", slot, object, cString(root.IntersectionTag), e.cache(), tag.Of, cString(tag.Name), cString(path))
	for index, member := range root.Members {
		arm := e.program.ViewContracts[member-1]
		var allowed []ir.ViewLiteral
		for _, field := range arm.Fields {
			if field.Name == root.IntersectionTag {
				allowed = e.program.ViewContracts[field.Contract-1].Allowed
			}
		}
		tests := []string{}
		for _, literal := range allowed {
			switch literal.Of {
			case ir.Number:
				tests = append(tests, fmt.Sprintf("%s.number == %s", slot, strconv.FormatFloat(literal.Number, 'g', -1, 64)))
			case ir.Boolean:
				tests = append(tests, fmt.Sprintf("%s.boolean == %t", slot, literal.Boolean))
			case ir.String:
				name := e.temporary()
				e.declarations = append(e.declarations, fmt.Sprintf("static adamic_string %s = ADAMIC_STRING(%s);", name, cString(literal.String)))
				tests = append(tests, fmt.Sprintf("adamic_string_equal((const adamic_string *)%s.reference, &%s)", slot, name))
			}
		}
		prefix := "if"
		if index != 0 {
			prefix = "else if"
		}
		e.line("%s (%s) {", prefix, strings.Join(tests, " || "))
		e.viewIntersectionFields(member, object, property.View, map[ir.ViewContractID]bool{}, root.IntersectionTag)
		e.line("}")
	}
	e.line("else { adamic_view_literal_failure(%s, %s, %d, %s); }", cString(path), cString(tag.Name), tag.Of, slot)
}
