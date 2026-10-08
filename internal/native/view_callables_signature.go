package native

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

// Producer certificates are selected by immutable generated code identity. This
// closed-program adapter needs no inferred Function.length, view-supplied
// implementation ID, heap mutation or second closure calling convention.
func (e *emitter) emitViewCallableCertificate(property ir.Property, value string) string {
	e.declarations = append(e.declarations, "#include \"view_callables_contract.h\"")
	expected := e.viewCallableExpected(property)
	recorded := e.temporary()
	e.line("const adamic_callable_signature *%s = NULL;", recorded)
	for index, function := range e.program.Functions {
		if !function.Closure || function.Receiver {
			continue
		}
		parameters := make([]ir.Type, len(function.Parameters))
		for i, local := range function.Parameters {
			parameters[i] = e.program.Locals[local].Type
			if parameters[i] == ir.Object && len(function.CallableMasks) == len(function.Parameters)+1 && function.CallableMasks[i] == 0 {
				parameters[i] = 0
			}
		}
		signature := e.viewCallableSignature(parameters, viewCallableProducerResult(function.Returns), function.Name, function.CallableMasks)
		e.line("if (%s != NULL && %s->heap.kind == adamic_kind_closure && %s->code == %s) %s = %s;", value, value, value, e.functionName(index), recorded, signature)
	}
	return emitViewCallableShape(value, recorded, expected, property.View, property.Absent || property.Optional)
}

func (e *emitter) viewCallableExpected(property ir.Property) string {
	id := property.ViewContract
	if id == 0 || int(id) > len(e.program.ViewContracts) {
		return "NULL"
	}
	contract := e.program.ViewContracts[id-1]
	if contract.Kind != ir.ViewCallable || (contract.Result == 0 && !contract.DiscardResult) {
		return "NULL"
	}
	if contract.DiscardResult {
		return e.viewCallableSignature(nil, ir.Type(255), contract.Name)
	}
	parameters := make([]ir.Type, len(contract.Parameters))
	masks := make([]uint16, len(contract.Parameters)+1)
	for i, child := range contract.Parameters {
		parameters[i] = e.program.ViewContracts[child-1].Of
		masks[i] = e.program.ViewContracts[child-1].RepresentationMask
	}
	masks[len(masks)-1] = e.program.ViewContracts[contract.Result-1].RepresentationMask
	return e.viewCallableSignature(parameters, e.program.ViewContracts[contract.Result-1].Of, contract.Name, masks)
}

func (e *emitter) viewCallableSignature(parameters []ir.Type, result ir.Type, name string, members ...[]uint16) string {
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

func viewCallableProducerResult(result ir.Type) ir.Type {
	if result == 0 {
		return ir.Type(254)
	}
	return result
}
