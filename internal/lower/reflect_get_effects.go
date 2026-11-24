package lower

import "github.com/system-inc/adamic/internal/ir"

// keyedGetterReadsAndThrows checks the reachable code of a getter that exceeded
// the syntax judgment. A back edge adds no effects to the current walk.
func (l *lowering) keyedGetterReadsAndThrows(getter int) bool {
	visited := make(map[int]bool)
	var inspect func(int) bool
	inspect = func(function int) bool {
		if function < 0 || function >= len(l.result.Functions) {
			return false
		}
		if visited[function] {
			return true
		}
		visited[function] = true
		body := l.result.Functions[function]
		return !body.Async && !body.Generator && l.keyedGetterReadBody(body.Body, function, inspect)
	}
	return inspect(getter)
}

func (l *lowering) keyedGetterReadBody(body any, function int, inspect func(int) bool) bool {
	allowed := true
	walk(body, func(node any) bool {
		if !allowed {
			return false
		}
		switch node := node.(type) {
		case ir.Return, ir.Evaluate, ir.If, ir.Block, ir.Try,
			ir.Read, ir.HasAccessor, ir.ArrayIndex, ir.MapGet, ir.MapHas,
			ir.MapSize, ir.Length, ir.StringLength,
			ir.NumberConstant, ir.BooleanConstant, ir.StringConstant, ir.Null, ir.Undefined,
			ir.Conditional, ir.Binary, ir.Unary, ir.Logical, ir.Coalesce, ir.TypeOf,
			ir.IsNull, ir.IsUndefined, ir.Truthy, ir.InstanceOf, ir.Unwrap, ir.Defined,
			ir.Narrow, ir.CheckedCast, ir.MaybeOf, ir.Effects, ir.RecordStorage:
			// These operations read or organize their operands. The walk checks the operands.
		case ir.Property:
			allowed = node.View == "" // A checked view read can box its field for the caller. #7j0xwmz narrows it.
		case ir.Declare:
			local := l.result.Locals[node.Local]
			allowed = local.Function == function && !local.Global && !local.Captured &&
				!node.Uninitialized && !node.CopyCell
		case ir.RecordCall:
			allowed = node.Method == "get" || node.Method == "has" || node.Method == "hasOwn"
		case ir.Throw:
			if made, ok := node.Value.(ir.MakeError); ok && node.DirectError {
				// Exempt just the Error itself. Its message and name have to pass the ordinary proof.
				allowed = l.keyedGetterReadBody([]ir.Expression{made.Message, made.Name}, function, inspect)
				return false
			}
		case ir.Box:
			allowed = node.Value.Type() != ir.Number && node.Value.Type() != ir.MaybeNumber
		case ir.Call:
			allowed = !l.result.CallExpandsArguments(node)
			if allowed {
				for _, target := range l.result.CallTargets(node) {
					if !inspect(target) {
						allowed = false
						break
					}
				}
			}
		case ir.DynamicKeyedProperty:
			allowed = node.Reflect
			for _, target := range l.result.DynamicKeyedGetterTargets(node) {
				if !inspect(target) {
					allowed = false
					break
				}
			}
		case ir.Assign, ir.SetProperty, ir.SetIndex, ir.LogicalAssignment,
			ir.RecordCoalesce, ir.MapSet, ir.MapDelete, ir.MapClear, ir.SetAdd,
			ir.ArrayPush, ir.ArrayPop, ir.ArrayFill, ir.ArrayReverse, ir.ArraySplice:
			allowed = false // Write.
		case ir.MakeError, ir.ArrayLiteral, ir.ObjectLiteral, ir.RecordLiteral, ir.MapNew, ir.SetNew,
			ir.MakeClosure, ir.BoundMethod, ir.BoundClosure, ir.Concat, ir.NumberToString, ir.StringIndex:
			allowed = false // Allocation.
		case ir.FunctionLength:
			allowed = false // Not admitted here yet. #7j0xwmz admits it with its pin.
		default:
			allowed = false // Unknown operation or reach.
		}
		return allowed
	})
	return allowed
}
