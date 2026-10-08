package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A void operation still runs when its result is read. Pass its operands into an
// ordinary helper so evaluation stays at the source site, including local reads,
// argument order, short-circuit branches and exception cleanup.
func (l *lowering) voidValue(node *ast.Node, value ir.Expression) (ir.Expression, error) {
	var arguments []ir.Expression
	switch call := value.(type) {
	case ir.Call:
		arguments = call.Arguments
	case ir.CallClosure:
		targets := l.result.ClosureTargets(call)
		if targets.Unknown || len(targets.Functions) == 0 {
			return nil, l.notYet(node, "a void call used as a value")
		}
		for _, target := range targets.Functions {
			if l.result.Functions[target].Returns != 0 {
				return nil, l.notYet(node, "a void call used as a value")
			}
		}
		arguments = append([]ir.Expression{call.Closure}, call.Arguments...)
	case ir.ArrayVisit:
		arguments = []ir.Expression{call.Array, call.Callback}
	case ir.MapClear:
		arguments = []ir.Expression{call.Map}
	case ir.MapForEach:
		arguments = []ir.Expression{call.Map, call.Callback}
	default:
		// In particular, () => void can hold () => number. Its erased result
		// is not proof that Node returns undefined.
		return nil, l.notYet(node, "a void call used as a value")
	}
	b := l.libraryArrayBuilder(arguments)
	reads := make([]ir.Expression, len(arguments))
	for index, parameter := range b.parameters {
		reads[index] = b.read(parameter)
	}
	switch call := value.(type) {
	case ir.Call:
		call.Arguments = reads
		value = call
	case ir.CallClosure:
		call.Closure, call.Arguments = reads[0], reads[1:]
		value = call
	case ir.ArrayVisit:
		call.Array, call.Callback = reads[0], reads[1]
		value = call
	case ir.MapClear:
		call.Map = reads[0]
		value = call
	case ir.MapForEach:
		call.Map, call.Callback = reads[0], reads[1]
		value = call
	}
	b.body = append(b.body, ir.Evaluate{Value: value})
	return b.finish("void_value", ir.Undefined{}), nil
}
