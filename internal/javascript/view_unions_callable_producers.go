package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) unionCallableExpected(property ir.Property) string {
	id := property.ViewContract
	if id == 0 || int(id) > len(e.program.ViewContracts) {
		return "undefined"
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Kind != ir.ViewCallable || (contract.Result == 0 && !contract.DiscardResult) {
		return "undefined"
	}
	if contract.DiscardResult {
		return unionCallableSignature(nil, ir.Type(255), contract.Name)
	}
	parameters := make([]ir.Type, len(contract.Parameters))
	masks := make([]uint16, len(contract.Parameters)+1)
	for i, child := range contract.Parameters {
		parameters[i] = e.program.ViewContracts[child-1].Of
		masks[i] = e.program.ViewContracts[child-1].RepresentationMask
	}
	masks[len(masks)-1] = e.program.ViewContracts[contract.Result-1].RepresentationMask
	return unionCallableSignature(parameters, e.program.ViewContracts[contract.Result-1].Of, contract.Name, masks)
}

func unionCallableSignature(parameters []ir.Type, result ir.Type, name string, members ...[]uint16) string {
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

func unionCallableProducerResult(result ir.Type) ir.Type {
	if result == 0 {
		return ir.Type(254)
	}
	return result
}

func (e *emitter) unionCallableRecorded(value string) string {
	choices := []string{}
	for index, function := range e.program.Functions {
		if !function.Closure || function.Receiver {
			continue
		}
		parameters := make([]ir.Type, len(function.Parameters))
		for i, local := range function.Parameters {
			parameters[i] = e.program.Locals[local].Type
			if parameters[i] != ir.Number && parameters[i] != ir.Boolean && parameters[i] != ir.String && parameters[i] != ir.MaybeNumber && parameters[i] != ir.MaybeBoolean {
				parameters[i] = 0
			}
		}
		result := unionCallableProducerResult(function.Returns)
		if result != ir.Number && result != ir.Boolean && result != ir.String && result != 254 {
			result = 0
		}
		choices = append(choices, fmt.Sprintf("code === %s ? ({...%s,function:%d}) : ", functionName(e.program, index), unionCallableSignature(parameters, result, function.Name), index))
	}
	return "((code)=>" + strings.Join(choices, "") + "undefined)(" + value + " instanceof AdamicClosure ? " + value + ".code : " + value + ")"
}
