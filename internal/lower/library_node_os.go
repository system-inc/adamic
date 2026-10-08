package lower

import (
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// Node functions are ordinary stable callable values, including tsc's
// homedir feature probe. Each wrapper forwards to the real native operation.
func (l *lowering) nodeOSFunction(path string) ir.Expression {
	if local, found := l.nodeOSFunctions[path]; found {
		return ir.Read{Local: local, Of: ir.Closure}
	}
	operation := strings.TrimPrefix(path, "os.")
	index := len(l.result.Functions)
	nativeOperation := operation
	if operation == "platform" {
		nativeOperation = "osPlatform"
	}
	call := ir.ProcessCall{Operation: nativeOperation, Of: ir.String}
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "node_os_" + operation, Returns: ir.String, Closure: true, MayThrow: operation == "homedir", Body: []ir.Statement{ir.Return{Value: call}}})
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "node_os_" + operation, Type: ir.Closure, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: local, Value: ir.MakeClosure{Function: index}})
	if l.nodeOSFunctions == nil {
		l.nodeOSFunctions = map[string]int{}
	}
	l.nodeOSFunctions[path] = local
	return ir.Read{Local: local, Of: ir.Closure}
}
