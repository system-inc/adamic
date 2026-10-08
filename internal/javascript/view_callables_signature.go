package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) emitViewCallableCertificate(property ir.Property, value string, receiver ...string) string {
	expected := e.viewCallableExpected(property)
	recorded := e.viewCallableRecorded("value", property)
	if len(receiver) != 0 {
		recorded = "(adamicViewSetSignature(" + receiver[0] + ", value) ?? " + recorded + ")"
	}
	recorded = e.viewCallableLogicalProducer(property, recorded)
	expected = e.untaggedCallableUnionExpected(property, recorded, expected)
	recorded = e.untaggedCallableRecorded(property, recorded, expected)
	return "((value) => " + emitViewCallableShape("value", recorded, expected, property.View, property.Absent || property.Optional || property.UndefinedAllowed) + ")(" + value + ")"
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
	masks := make([]uint16, len(contract.Parameters)+1)
	for i, child := range contract.Parameters {
		parameters[i] = e.program.ViewContracts[child-1].Of
		masks[i] = e.program.ViewContracts[child-1].RepresentationMask
	}
	masks[len(masks)-1] = e.program.ViewContracts[contract.Result-1].RepresentationMask
	return viewCallableSignature(parameters, e.program.ViewContracts[contract.Result-1].Of, contract.Name, masks)
}

func viewCallableSignature(parameters []ir.Type, result ir.Type, name string, members ...[]uint16) string {
	values := make([]string, len(parameters))
	for i, of := range parameters {
		values[i] = fmt.Sprint(of)
	}
	maskValues := []string{}
	resultMask := uint16(0)
	if len(members) != 0 && len(members[0]) == len(parameters)+1 {
		resultMask = members[0][len(parameters)]
		for _, mask := range members[0][:len(parameters)] {
			maskValues = append(maskValues, fmt.Sprint(mask))
		}
	}
	return fmt.Sprintf("{parameters: [%s], result: %d, name: %s, parameterMasks: [%s], resultMask: %d}", strings.Join(values, ", "), result, quote(name), strings.Join(maskValues, ", "), resultMask)
}

func viewCallableProducerResult(result ir.Type) ir.Type {
	if result == 0 {
		return ir.Type(254)
	}
	return result
}

func (e *emitter) viewCallableRecorded(value string, properties ...ir.Property) string {
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
		offset := 0
		masks := function.CallableMasks
		result := viewCallableProducerResult(function.Returns)
		if !function.Closure {
			locals = locals[1:]
			offset = 1
			masks = nil
			for _, property := range properties {
				if property.ViewContract != 0 {
					callable := e.program.ViewContracts[property.ViewContract-1]
					if callable.Result != 0 && e.program.ViewContracts[callable.Result-1].Of == ir.Union && !function.Returns.IsReference() {
						result = 0
					}
				}
			}
		}
		parameters := make([]ir.Type, len(locals))
		for i, local := range locals {
			parameters[i] = e.program.Locals[local].Type
			if parameters[i] == ir.Object && len(function.CallableMasks) == len(function.Parameters)+1 && function.CallableMasks[i+offset] == 0 {
				parameters[i] = 0
			}
		}
		choices = append(choices, fmt.Sprintf("code === %s ? %s : ", functionName(e.program, index), fmt.Sprintf("({...%s, function: %d})", viewCallableSignature(parameters, result, function.Name, masks), index)))
	}
	recorded := "((code) => " + strings.Join(choices, "") + "undefined)(" + value + " instanceof AdamicClosure ? " + value + ".code : " + value + ")"
	return recorded
}
