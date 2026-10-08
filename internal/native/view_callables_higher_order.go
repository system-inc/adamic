package native

import (
	"fmt"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) viewCallableLogicalProducer(property ir.Property, value, recorded string) {
	if property.ViewContract == 0 {
		return
	}
	contract := e.program.ViewContracts[property.ViewContract-1]
	if contract.Kind != ir.ViewCallable || !contract.ProducerCertified {
		return
	}
	choices := []string{}
	for _, function := range contract.Functions {
		choices = append(choices, fmt.Sprintf("%s->code == %s", value, e.functionName(function)))
	}
	allowed := "false"
	if len(choices) != 0 {
		allowed = "(" + strings.Join(choices, " || ") + ")"
	}
	e.line("if (%s != NULL && !%s) %s = NULL;", recorded, allowed, recorded)
}
