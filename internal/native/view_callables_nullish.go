package native

import "github.com/system-inc/adamic/internal/ir"

// The owner validates presence and nullish kind first. Check each present
// callable before returning its existing owned snapshot, without reboxing it.
func (e *emitter) viewCallableNullishCertificate(property ir.Property, value string) {
	if property.ViewContract == 0 {
		return
	}
	contract := e.program.ViewContracts[property.ViewContract-1]
	if contract.Kind != ir.ViewCallable || contract.Result == 0 && !contract.DiscardResult {
		return
	}
	e.line("if (%s != NULL && %s != &adamic_null) {", value, value)
	callable := e.temporary()
	e.line("adamic_closure *%s = (adamic_closure *)%s;", callable, value)
	e.line("(void)%s;", e.emitViewCallableCertificate(property, callable))
	e.line("}")
}
