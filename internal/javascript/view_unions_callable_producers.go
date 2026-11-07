package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"slices"
	"strings"
)

func (javascriptEmitter *emitter) unionCallableExpected(property ir.Property) string {
	id := property.ViewContract
	if id == 0 || int(id) > len(javascriptEmitter.program.ViewContracts) {
		return "undefined"
	}
	contract := javascriptEmitter.program.ViewContracts[id-1]
	if contract.Kind != ir.ViewCallable || (contract.Result == 0 && !contract.DiscardResult) {
		return "undefined"
	}
	if contract.DiscardResult {
		return unionCallableSignature(nil, ir.Type(255), contract.Name)
	}
	parameters := make([]ir.Type, len(contract.Parameters))
	masks := make([]uint16, len(contract.Parameters)+1)
	for index, child := range contract.Parameters {
		parameters[index] = javascriptEmitter.program.ViewContracts[child-1].Of
		masks[index] = javascriptEmitter.program.ViewContracts[child-1].RepresentationMask
	}
	masks[len(masks)-1] = javascriptEmitter.program.ViewContracts[contract.Result-1].RepresentationMask
	return unionCallableSignature(
		parameters,
		javascriptEmitter.program.ViewContracts[contract.Result-1].Of,
		contract.Name,
		masks,
	)
}

func unionCallableSignature(parameters []ir.Type, result ir.Type, name string, members ...[]uint16) string {
	values := make([]string, len(parameters))
	for index, of := range parameters {
		values[index] = fmt.Sprint(of)
	}
	maskValues := []string{}
	resultMask := uint16(0)
	if len(members) != 0 && len(members[0]) == len(parameters)+1 {
		resultMask = members[0][len(parameters)]
		for _, mask := range members[0][:len(parameters)] {
			maskValues = append(maskValues, fmt.Sprint(mask))
		}
	}
	return fmt.Sprintf(
		"{parameters: [%s], result: %d, name: %s, parameterMasks: [%s], resultMask: %d}",
		strings.Join(values, ", "),
		result,
		quote(name),
		strings.Join(maskValues, ", "),
		resultMask,
	)
}

func unionCallableProducerResult(result ir.Type) ir.Type {
	if result == 0 {
		return ir.Type(254)
	}
	return result
}

func (javascriptEmitter *emitter) unionCallableRecorded(value string) string {
	choices := []string{}
	for index, function := range javascriptEmitter.program.Functions {
		if !function.Closure || function.Receiver {
			continue
		}
		parameters := make([]ir.Type, len(function.Parameters))
		for index, local := range function.Parameters {
			parameters[index] = javascriptEmitter.program.Locals[local].Type
			if parameters[index] != ir.Number &&
				parameters[index] != ir.Boolean &&
				parameters[index] != ir.String &&
				parameters[index] != ir.MaybeNumber &&
				parameters[index] != ir.MaybeBoolean {
				parameters[index] = 0
			}
		}
		masks := make([]uint16, len(parameters)+1)
		// Exact checker identity certified these producer contracts in lowering.
		for _, contract := range javascriptEmitter.program.ViewContracts {
			if contract.Kind != ir.ViewCallable ||
				!contract.ProducerCertified ||
				!slices.Contains(contract.Functions, index) {
				continue
			}
			for index, child := range contract.Parameters {
				parameters[index] = javascriptEmitter.program.ViewContracts[child-1].Of
				masks[index] = javascriptEmitter.program.ViewContracts[child-1].RepresentationMask
			}
		}
		result := unionCallableProducerResult(function.Returns)
		if result != ir.Number && result != ir.Boolean && result != ir.String && result != 254 {
			result = 0
		}
		choices = append(choices, fmt.Sprintf(
			"code === %s ? ({...%s,function:%d}) : ",
			functionName(javascriptEmitter.program, index),
			unionCallableSignature(parameters, result, function.Name, masks),
			index,
		))
	}
	return "((code)=>" +
		strings.Join(choices, "") +
		"undefined)(" +
		value +
		" instanceof AdamicClosure ? " +
		value +
		".code : " +
		value +
		")"
}
