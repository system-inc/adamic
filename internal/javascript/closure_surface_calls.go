package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) callWithReceiver(call ir.CallClosure) string {
	value := e.program.ClosureTargets(call).Value
	arguments := "[" + e.callValues(call.Arguments, call.Spread) + "]"
	if call.ArgumentCount != nil {
		arguments += ".slice(0, " + e.value(call.ArgumentCount) + ")"
	}
	if call.CheckBound {
		return e.viewCallableInvoke(call, ir.Property{View: call.BoundView, ViewType: e.program.ViewContracts[call.CallContract-1].Name}) + "({value:" + e.value(value) + ", object:" + e.value(call.Receiver) + "}, " + arguments + ")"
	}
	return fmt.Sprintf("adamicCallReceiver(%s, %s, %s)", e.value(value), e.value(call.Receiver), arguments)
}
