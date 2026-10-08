package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) emitViewCallableRead(property ir.Property, receiver, method string) string {
	contract := property.ViewContract
	if contract == 0 || (e.program.ViewContracts[contract-1].Result == 0 && !e.program.ViewContracts[contract-1].DiscardResult) {
		return e.own(ir.Closure, fmt.Sprintf("adamic_retain(adamic_view_callable(%s, %s, &%s, &%s, %s))", receiver, cString(property.Name), e.cache(), method, cString(property.View)))
	}
	expected := e.program.ViewContracts[contract-1].Name
	value := e.own(ir.Closure, fmt.Sprintf("adamic_retain(adamic_view_callable_typed(%s, %s, &%s, &%s, %s, %s, %t, %t))", receiver, cString(property.Name), e.cache(), method, cString(property.View), cString(expected), property.Absent, property.Optional))
	present := method + " != NULL"
	if e.program.ClosureConventionNeeded() {
		present = "(" + method + ".counted ? " + method + ".counted_code != NULL : " + method + ".code != NULL)"
	}
	e.line("if (%s == NULL && %s) {", value, present)
	e.indent++
	e.emitViewCallableMethodCertificate(property, method)
	e.indent--
	e.line("} else {")
	e.indent++
	checked := e.emitViewCallableCertificate(property, value)
	e.line("%s = %s;", value, checked)
	e.indent--
	e.line("}")
	return value
}
