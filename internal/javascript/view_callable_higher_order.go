package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
	"strings"
)

func (e *emitter) viewCallableCallbackAdapter(id ir.ViewContractID, value string, origin ir.Property, where string) string {
	name := fmt.Sprintf("adamicViewCallback%d", id)
	prefix := "function " + name + "(value)"
	for _, declaration := range e.prototypes {
		if strings.HasPrefix(declaration, prefix) {
			return name + "(" + value + ")"
		}
	}
	index := len(e.prototypes)
	e.prototypes = append(e.prototypes, prefix)
	target := e.program.ViewContracts[id-1]
	property := ir.Property{ViewEscape: true, View: origin.View + " callback", ViewWhere: where, ViewType: target.Name}
	call := ir.CallClosure{CallContract: id, CallWhere: "callback adapted at " + where, Returns: e.program.ViewContracts[target.Result-1].Of}
	if e.program.ViewContracts[target.Result-1].Kind == ir.ViewUndefined {
		call.Returns = 0
	}
	invoke := e.viewCallableInvoke(call, property)
	e.prototypes[index] = fmt.Sprintf("%s { return adamicViewAdapterIntern(value, adamicViewAdapterKey(%d), underlying => new AdamicClosure((self, arguments_, receiver=undefined) => %s({value:underlying,object:receiver},arguments_), [])); }\n", prefix, target.CallableTypeID, invoke)
	return name + "(" + value + ")"
}
