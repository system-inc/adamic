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
	for _, producer := range e.viewCallableProducers() {
		e.line("if (%s != NULL && %s->heap.kind == adamic_kind_closure && %s->code == %s) %s = %s;", value, value, value, e.functionName(producer.Function), recorded, producer.Signature)
	}
	expected = e.untaggedCallableUnionExpected(property, recorded, expected)
	e.certifyUntaggedCallableRecorded(property, value, recorded, expected)
	return emitViewCallableShape(value, recorded, expected, property.View, property.Absent || property.Optional || property.UndefinedAllowed)
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
	for i, child := range contract.Parameters {
		parameters[i] = e.program.ViewContracts[child-1].Of
	}
	return e.viewCallableSignature(parameters, e.program.ViewContracts[contract.Result-1].Of, contract.Name)
}

func (e *emitter) viewCallableSignature(parameters []ir.Type, result ir.Type, name string) string {
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
	// Zero remains unknown. Void producer signatures cannot satisfy a valued result.
	e.declarations = append(e.declarations, fmt.Sprintf("static const adamic_callable_signature %s = {%d, %s, %d, %s};", signature, len(parameters), parameterName, result, cString(name)))
	return "&" + signature
}

func viewCallableProducerResult(result ir.Type) ir.Type {
	if result == 0 {
		return ir.Type(254)
	}
	return result
}

type viewCallableProducer struct {
	Function  int
	Signature string
}

func (e *emitter) viewCallableProducers() []viewCallableProducer {
	producers := []viewCallableProducer{}
	for index, function := range e.program.Functions {
		if !function.Closure || function.Receiver {
			continue
		}
		parameters := make([]ir.Type, len(function.Parameters))
		for i, local := range function.Parameters {
			parameters[i] = e.program.Locals[local].Type
		}
		producers = append(producers, viewCallableProducer{index, e.viewCallableSignature(parameters, viewCallableProducerResult(function.Returns), function.Name)})
	}
	return producers
}
