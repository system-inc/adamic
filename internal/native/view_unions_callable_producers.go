package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"slices"
	"strings"
)

func (e *emitter) unionCallableExpected(property ir.Property) string {
	id := property.ViewContract
	if id == 0 || int(id) > len(e.program.ViewContracts) {
		return "NULL"
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Kind != ir.ViewCallable || (contract.Result == 0 && !contract.DiscardResult) {
		return "NULL"
	}
	if contract.DiscardResult {
		return e.unionCallableSignature(nil, ir.Type(255), contract.Name)
	}
	parameters := make([]ir.Type, len(contract.Parameters))
	masks := make([]uint16, len(contract.Parameters)+1)
	for i, child := range contract.Parameters {
		parameters[i] = e.program.ViewContracts[child-1].Of
		masks[i] = e.program.ViewContracts[child-1].RepresentationMask
	}
	masks[len(masks)-1] = e.program.ViewContracts[contract.Result-1].RepresentationMask
	return e.unionCallableSignature(parameters, e.program.ViewContracts[contract.Result-1].Of, contract.Name, masks)
}

func (e *emitter) unionCallableSignature(parameters []ir.Type, result ir.Type, name string, members ...[]uint16) string {
	signature := e.temporary()
	parameterName := "NULL"
	if len(parameters) != 0 {
		parameterName = e.temporary()
		values := make([]string, len(parameters))
		for i, of := range parameters {
			values[i] = fmt.Sprint(of)
		}
		e.declarations = append(e.declarations, fmt.Sprintf("static const unsigned char %s[] = {%s};", parameterName, strings.Join(values, ", ")))
	}
	maskName := "NULL"
	resultMask := uint16(0)
	if len(members) != 0 && len(members[0]) == len(parameters)+1 {
		resultMask = members[0][len(parameters)]
		if len(parameters) != 0 {
			maskName = e.temporary()
			values := make([]string, len(parameters))
			for i, mask := range members[0][:len(parameters)] {
				values[i] = fmt.Sprint(mask)
			}
			e.declarations = append(e.declarations, fmt.Sprintf("static const uint16_t %s[] = {%s};", maskName, strings.Join(values, ", ")))
		}
	}
	// Zero remains unknown. Void producer signatures cannot satisfy a valued result.
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_callable_signature %s = {%d, %s, %d, %s, %s, %d};", signature, len(parameters), parameterName, result, cString(name), maskName, resultMask))
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

func (e *emitter) unionCallableProducers() []unionCallableProducer {
	producers := []unionCallableProducer{}
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
		masks := make([]uint16, len(parameters)+1)
		// Exact checker identity certified these producer contracts in lowering.
		for _, contract := range e.program.ViewContracts {
			if contract.Kind != ir.ViewCallable || !contract.ProducerCertified || !slices.Contains(contract.Functions, index) {
				continue
			}
			for i, child := range contract.Parameters {
				parameters[i] = e.program.ViewContracts[child-1].Of
				masks[i] = e.program.ViewContracts[child-1].RepresentationMask
			}
		}
		result := unionCallableProducerResult(function.Returns)
		if result != ir.Number && result != ir.Boolean && result != ir.String && result != 254 {
			result = 0
		}
		producers = append(producers, unionCallableProducer{index, e.unionCallableSignature(parameters, result, function.Name, masks)})
	}
	return producers
}
func (e *emitter) unionClosureCodeIdentity(value string, index int) string {
	if e.program.ClosureConventionNeeded() {
		if e.program.PackedCountNeeded(index) {
			return fmt.Sprintf("(%s->counted && %s->counted_code == %s)", value, value, e.functionName(index))
		}
		return fmt.Sprintf("(!%s->counted && %s->code == %s)", value, value, e.functionName(index))
	}
	return fmt.Sprintf("%s->code == %s", value, e.functionName(index))
}
