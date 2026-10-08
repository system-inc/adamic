package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) viewCallableNestedSignature(parameters []ir.Type, result ir.Type, name string, masks []uint16, contracts []ir.ViewContractID) string {
	children := []string{}
	active := false
	for _, id := range contracts {
		child := "undefined"
		if id != 0 && e.program.ViewContracts[id-1].Kind == ir.ViewCallable {
			child = e.viewCallableExpected(ir.Property{ViewContract: id})
			active = true
		}
		children = append(children, child)
	}
	signature := viewCallableSignature(parameters, result, name, masks)
	if !active {
		return signature
	}
	return "({..." + signature + ", parameterSignatures:[" + strings.Join(children, ", ") + "]})"
}

func (e *emitter) viewCallableCheckedArguments(call ir.CallClosure) string {
	values := []string{}
	for index, argument := range call.Arguments {
		value := e.value(argument)
		if index < len(call.CallableArguments) && call.CallableArguments[index] != 0 {
			value = e.emitViewCallableCertificate(ir.Property{ViewContract: call.CallableArguments[index], View: fmt.Sprintf("call argument %d", index+1)}, value)
		}
		values = append(values, value)
	}
	return strings.Join(values, ", ")
}
