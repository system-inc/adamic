package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) emitViewCallableCertificate(property ir.Property, value string) string {
	expected := e.viewCallableExpected(property)
	choices := []string{}
	for index, function := range e.program.Functions {
		if !function.Closure || function.Receiver {
			continue
		}
		parameters := make([]ir.Type, len(function.Parameters))
		for i, local := range function.Parameters {
			parameters[i] = e.program.Locals[local].Type
		}
		choices = append(choices, fmt.Sprintf("code === %s ? %s : ", functionName(e.program, index), viewCallableSignature(parameters, function.Returns, function.Name)))
	}
	recorded := "((code) => " + strings.Join(choices, "") + "undefined)(value instanceof AdamicClosure ? value.code : undefined)"
	return "((value) => " + emitViewCallableShape("value", recorded, expected, property.View, property.Absent) + ")(" + value + ")"
}

func (e *emitter) viewCallableExpected(property ir.Property) string {
	id := property.ViewContract
	if id == 0 || int(id) > len(e.program.ViewContracts) {
		return "undefined"
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Kind != ir.ViewCallable || contract.Result == 0 {
		return "undefined"
	}
	parameters := make([]ir.Type, len(contract.Parameters))
	for i, child := range contract.Parameters {
		parameters[i] = e.program.ViewContracts[child-1].Of
	}
	return viewCallableSignature(parameters, e.program.ViewContracts[contract.Result-1].Of, contract.Name)
}

func viewCallableSignature(parameters []ir.Type, result ir.Type, name string) string {
	values := make([]string, len(parameters))
	for i, of := range parameters {
		values[i] = fmt.Sprint(of)
	}
	return fmt.Sprintf("{parameters: [%s], result: %d, name: %s}", strings.Join(values, ", "), result, quote(name))
}
