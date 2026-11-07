package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

// One checked load for explicit reads and every implicit handout. The runtime
// checks storage before interpreting or retaining the slot and returns it owned.
func (e *emitter) checkedElement(check *ir.ViewElementCheck, array, index string, relative bool, element ir.Type) string {
	e.declarations = append(e.declarations, "#include \"view_unions_untagged.h\"")
	contracts, count := e.viewUntaggedContracts()
	return fmt.Sprintf("adamic_view_element(%s, %d, %d, %s, %s, %t, %d, %s, %s)", contracts, count, check.Contract, array, index, relative, element, cString(check.View), cString(check.ViewType))
}

// A checked read is a pure consumer, as ArrayIndex is (borrow.go's pureKind): adamic_view_element calls no program
// code and frees nothing, and it returns the element retained, so the array is read lent, with no count of its own.
func (e *emitter) viewElement(read ir.ViewElement) string {
	array := e.value(read.Array)
	index := e.value(read.Index)
	element := read.Of
	if read.Contract != 0 {
		element = e.program.ViewContracts[read.Contract-1].Of
	}
	check := &ir.ViewElementCheck{Contract: read.Contract, Where: read.Where, View: read.View, ViewType: read.ViewType}
	snapshot := e.temporary()
	e.line("adamic_view_union_value %s = %s;", snapshot, e.checkedElement(check, array, index, read.Relative, element))
	value := unslotted(element, snapshot+".payload."+member(element))
	// A maybe is only a number's or a boolean's (ir.Maybe keeps a reference type as it is, NULL for
	// undefined), and its contract's Of is that scalar or its maybe, so the runtime retains and boxes
	// nothing and a snapshot owns nothing. A maybe of a reference never reaches here
	// (optional-element.ts), and if one did, its retained count would leak, so it stops the build.
	if read.Of.IsMaybe() {
		if element.IsReference() {
			panic("native: a checked element read at " + read.Where + " is a maybe, but its contract holds a reference")
		}
		if !element.IsMaybe() {
			value = maybe(read.Of, value)
		}
		return e.snapshot(read.Of, fmt.Sprintf("(%s.kind == adamic_view_union_undefined ? %s : %s)", snapshot, zero(read.Of), value))
	}
	if read.Of.IsReference() {
		return e.own(read.Of, fmt.Sprintf("(%s)%s", cType(read.Of), value))
	}
	return e.snapshot(read.Of, value)
}
