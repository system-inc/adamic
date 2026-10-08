package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) emitViewCallableCertificate(property ir.Property, value string) string {
	expected := e.viewCallableExpected(property)
	choices := []string{}
	methods := map[int]bool{}
	for _, class := range e.program.Classes {
		for _, method := range class.Methods {
			methods[method] = true
		}
	}
	for index, function := range e.program.Functions {
		if !function.Closure && !methods[index] || function.Receiver {
			continue
		}
		locals := function.Parameters
		if !function.Closure {
			locals = locals[1:]
		}
		parameters := make([]ir.Type, len(locals))
		for i, local := range locals {
			parameters[i] = e.program.Locals[local].Type
		}
		choices = append(choices, fmt.Sprintf("code === %s ? %s : ", functionName(e.program, index), viewCallableSignature(parameters, viewCallableProducerResult(function.Returns), function.Name)))
	}
	recorded := "((code) => " + strings.Join(choices, "") + "undefined)(value instanceof AdamicClosure ? value.code : value)"
	return "((value) => " + emitViewCallableShape("value", recorded, expected, property.View, property.Absent || property.Optional) + ")(" + value + ")"
}

func (e *emitter) viewCallableExpected(property ir.Property) string {
	id := property.ViewContract
	if id == 0 || int(id) > len(e.program.ViewContracts) {
		return "undefined"
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Kind != ir.ViewCallable || (contract.Result == 0 && !contract.DiscardResult) {
		return "undefined"
	}
	if contract.DiscardResult {
		return viewCallableSignature(nil, ir.Type(255), contract.Name)
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

func viewCallableProducerResult(result ir.Type) ir.Type {
	if result == 0 {
		return ir.Type(254)
	}
	return result
}
