package javascript

import (
	"fmt"
	"github.com/system-inc/adamic/internal/ir"
)

func (e *emitter) callWithReceiver(call ir.CallClosure) string {
	value := e.program.ClosureTargets(call).Value
	return fmt.Sprintf("adamicCallReceiver(%s, %s, [%s])", e.value(value), e.value(call.Receiver), e.callValues(call.Arguments, call.Spread))
}
