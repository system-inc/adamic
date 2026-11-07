package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"slices"
	"strings"
)

func (nativeEmitter *emitter) unionCallableExpected(property ir.Property) string {
	id := property.ViewContract
	if id == 0 || int(id) > len(nativeEmitter.program.ViewContracts) {
		return "NULL"
	}
	contract := nativeEmitter.program.ViewContracts[id-1]
	if contract.Kind != ir.ViewCallable || (contract.Result == 0 && !contract.DiscardResult) {
		return "NULL"
	}
	if contract.DiscardResult {
		return nativeEmitter.unionCallableSignature(nil, ir.Type(255), contract.Name)
	}
	parameters := make([]ir.Type, len(contract.Parameters))
	masks := make([]uint16, len(contract.Parameters)+1)
	for index, child := range contract.Parameters {
		parameters[index] = nativeEmitter.program.ViewContracts[child-1].Of
		masks[index] = nativeEmitter.program.ViewContracts[child-1].RepresentationMask
	}
	masks[len(masks)-1] = nativeEmitter.program.ViewContracts[contract.Result-1].RepresentationMask
	return nativeEmitter.unionCallableSignature(
		parameters,
		nativeEmitter.program.ViewContracts[contract.Result-1].Of,
		contract.Name,
		masks,
	)
}

func (nativeEmitter *emitter) unionCallableSignature(
	parameters []ir.Type,
	result ir.Type,
	name string,
	members ...[]uint16,
) string {
	signature, exists := nativeEmitter.sharedDeclaration(
		"adamic_callable",
		fmt.Sprintf("%v/%d/%q/%v", parameters, result, name, members),
	)
	if exists {
		return "&" + signature
	}
	parameterName := "NULL"
	if len(parameters) != 0 {
		parameterName = signature + "_parameters"
		values := make([]string, len(parameters))
		for index, of := range parameters {
			values[index] = fmt.Sprint(of)
		}
		nativeEmitter.declarations = append(
			nativeEmitter.declarations,
			fmt.Sprintf("static const unsigned char %s[] = {%s};", parameterName, strings.Join(values, ", ")),
		)
	}
	maskName := "NULL"
	resultMask := uint16(0)
	if len(members) != 0 && len(members[0]) == len(parameters)+1 {
		resultMask = members[0][len(parameters)]
		if len(parameters) != 0 {
			maskName = signature + "_masks"
			values := make([]string, len(parameters))
			for index, mask := range members[0][:len(parameters)] {
				values[index] = fmt.Sprint(mask)
			}
			nativeEmitter.declarations = append(
				nativeEmitter.declarations,
				fmt.Sprintf("static const uint16_t %s[] = {%s};", maskName, strings.Join(values, ", ")),
			)
		}
	}
	// Zero remains unknown. Void producer signatures cannot satisfy a valued result.
	nativeEmitter.declarations = append(nativeEmitter.declarations, fmt.Sprintf(
		"static const adamic_callable_signature %s = {%d, %s, %d, %s, %s, %d};",
		signature,
		len(parameters),
		parameterName,
		result,
		cString(name),
		maskName,
		resultMask,
	))
	return "&" + signature
}

func unionCallableProducerResult(result ir.Type) ir.Type {
	if result == 0 {
		return ir.Type(254)
	}
	return result
}

type unionCallableProducer struct {
	Function  int
	Signature string
}

func (nativeEmitter *emitter) unionCallableProducers() []unionCallableProducer {
	producers := []unionCallableProducer{}
	for index, function := range nativeEmitter.program.Functions {
		if !function.Closure || function.Receiver {
			continue
		}
		parameters := make([]ir.Type, len(function.Parameters))
		for index, local := range function.Parameters {
			parameters[index] = nativeEmitter.program.Locals[local].Type
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
		for _, contract := range nativeEmitter.program.ViewContracts {
			if contract.Kind != ir.ViewCallable ||
				!contract.ProducerCertified ||
				!slices.Contains(contract.Functions, index) {
				continue
			}
			for index, child := range contract.Parameters {
				parameters[index] = nativeEmitter.program.ViewContracts[child-1].Of
				masks[index] = nativeEmitter.program.ViewContracts[child-1].RepresentationMask
			}
		}
		result := unionCallableProducerResult(function.Returns)
		if result != ir.Number && result != ir.Boolean && result != ir.String && result != 254 {
			result = 0
		}
		producers = append(producers, unionCallableProducer{
			index,
			nativeEmitter.unionCallableSignature(parameters, result, function.Name, masks),
		})
	}
	return producers
}
func (nativeEmitter *emitter) unionClosureCodeIdentity(value string, index int) string {
	if nativeEmitter.program.ClosureConventionNeeded() {
		if nativeEmitter.program.PackedCountNeeded(index) {
			return fmt.Sprintf(
				"(%s->counted && %s->counted_code == %s)",
				value,
				value,
				nativeEmitter.functionName(index),
			)
		}
		return fmt.Sprintf("(!%s->counted && %s->code == %s)", value, value, nativeEmitter.functionName(index))
	}
	return fmt.Sprintf("%s->code == %s", value, nativeEmitter.functionName(index))
}
