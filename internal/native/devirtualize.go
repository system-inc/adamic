package native

import (
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// exactReceiverClass proves an allocation's identity, including a local whose
// only binding is that allocation. A static class type alone is not exact: this
// in a base constructor can already have a derived class's identity. Parameters,
// captured bindings, assignments, joins and property reads remain unbounded.
// This emission proof does not narrow any ownership or exception summary.
func (e *emitter) exactReceiverClass(value ir.Expression) int {
	switch value := value.(type) {
	case ir.Call:
		if value.Virtual != 0 {
			return 0
		}
		target := e.program.CallTargets(value)[0]
		for index, class := range e.program.Classes {
			if !class.Static && class.Constructor == target {
				return index + 1
			}
		}
	case ir.Read:
		local := e.program.Locals[value.Local]
		if local.Captured || (local.Function >= 0 && slices.Contains(e.program.Functions[local.Function].Parameters, value.Local)) {
			return 0
		}
		if class, known := e.exactReceiverClasses[value.Local]; known {
			return class
		}
		if e.exactReceiverClasses == nil {
			e.exactReceiverClasses = map[int]int{}
		}
		e.receiverClassWalks++
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
							class = e.exactReceiverClass(call)
						}
					}
				case ir.Assign:
					written = written || statement.Local == value.Local
				}
				walkStatement(statement, func(ir.Expression) {}, statements)
			}
		}
		statements(e.program.Main)
		for _, function := range e.program.Functions {
			statements(function.Body)
		}
		if declarations != 1 || written {
			class = 0
		}
		e.exactReceiverClasses[value.Local] = class
		return class
	}
	return 0
}

// An interface can also hold an object literal with an own closure. Counting
// implementing classes is not sufficient; the receiver must be proven exact.
func (e *emitter) exactReceiverMethod(receiver ir.Expression, name string) (int, bool) {
	class := e.exactReceiverClass(receiver)
	if class == 0 {
		return 0, false
	}
	constructor := e.program.Functions[e.program.Classes[class-1].Constructor]
	for _, statement := range constructor.Body {
		if declare, ok := statement.(ir.Declare); ok {
			if literal, ok := declare.Value.(ir.ObjectLiteral); ok && literal.Class == class {
				for _, method := range literal.Methods {
					if method.Name == name && e.dispatchable(method.Function) {
						return method.Function, true
					}
				}
			}
		}
	}
	return 0, false
}
