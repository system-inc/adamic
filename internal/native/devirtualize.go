package native

import "github.com/system-inc/adamic/internal/ir"

// exactReceiverClass proves an allocation's identity, including a local whose
// only binding is that allocation. A static class type alone is not exact: this
// in a base constructor can already have a derived class's identity. Parameters,
// captured bindings, assignments, joins and property reads remain unbounded.
// This emission proof does not narrow any ownership or exception summary.
func (nativeEmitter *emitter) exactReceiverClass(value ir.Expression) int {
	switch value := value.(type) {
	case ir.Call:
		if value.Virtual != 0 {
			return 0
		}
		target := nativeEmitter.program.CallTargets(value)[0]
		for index, class := range nativeEmitter.program.Classes {
			if !class.Static && class.Constructor == target {
				return index + 1
			}
		}
	case ir.Read:
		if nativeEmitter.program.Locals[value.Local].Captured {
			return 0
		}
		class, declarations, written := 0, 0, false
		var statements func([]ir.Statement)
		statements = func(list []ir.Statement) {
			for _, statement := range list {
				switch statement := statement.(type) {
				case ir.Declare:
					if statement.Local == value.Local {
						declarations++
						// Only an allocation, never another read: no alias cycles.
						if call, ok := statement.Value.(ir.Call); ok {
							class = nativeEmitter.exactReceiverClass(call)
						}
					}
				case ir.Assign:
					written = written || statement.Local == value.Local
				}
				walkStatement(statement, func(ir.Expression) {}, statements)
			}
		}
		statements(nativeEmitter.program.Main)
		for _, function := range nativeEmitter.program.Functions {
			statements(function.Body)
		}
		if declarations == 1 && !written {
			return class
		}
	}
	return 0
}

// An interface can also hold an object literal with an own closure. Counting
// implementing classes is not sufficient; the receiver must be proven exact.
func (nativeEmitter *emitter) exactReceiverMethod(receiver ir.Expression, name string) (int, bool) {
	class := nativeEmitter.exactReceiverClass(receiver)
	if class == 0 {
		return 0, false
	}
	constructor := nativeEmitter.program.Functions[nativeEmitter.program.Classes[class-1].Constructor]
	for _, statement := range constructor.Body {
		if declare, ok := statement.(ir.Declare); ok {
			if literal, ok := declare.Value.(ir.ObjectLiteral); ok && literal.Class == class {
				for _, method := range literal.Methods {
					if method.Name == name && nativeEmitter.dispatchable(method.Function) {
						return method.Function, true
					}
				}
			}
		}
	}
	return 0, false
}
