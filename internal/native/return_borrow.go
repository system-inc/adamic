package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// returnedField recognizes an effect-free accessor, including every override.
// The caller can specialize its entire body to the field read. The ordinary
// function remains an owning return for callers that keep the result.
func returnedField(program *ir.Program, call ir.Call) (ir.Property, bool) {
	if call.Setter || len(call.Arguments) != 1 || !lendable(call.Returns) {
		return ir.Property{}, false
	}
	if inner, _, ok := getterDispatch(program, call); ok {
		return returnedField(program, inner)
	}
	var result ir.Property
	targets := program.CallTargets(call)
	if len(targets) == 0 {
		return ir.Property{}, false
	}
	for position, target := range targets {
		function := program.Functions[target]
		if function.Closure || function.MayThrow || len(function.Parameters) != 1 || len(function.Body) != 1 {
			return ir.Property{}, false
		}
		returned, ok := function.Body[0].(ir.Return)
		if !ok {
			return ir.Property{}, false
		}
		field, ok := returned.Value.(ir.Property)
		if !ok || field.Method || field.Optional || field.Absent || field.Of != call.Returns {
			return ir.Property{}, false
		}
		receiver, ok := field.Object.(ir.Read)
		if !ok || receiver.Local != function.Parameters[0] || program.Locals[receiver.Local].Name != "this" || !program.Locals[receiver.Local].Borrowed {
			return ir.Property{}, false
		}
		if position != 0 && (field.Name != result.Name || field.Of != result.Of) {
			return ir.Property{}, false
		}
		result = field
	}
	result.Object = call.Arguments[0]
	return result, true
}

// getterDispatch accepts the exact lowering wrapper: accessor lookup followed
// by an ordinary data-field fallback. Its accessor call still joins all targets.
func getterDispatch(program *ir.Program, call ir.Call) (ir.Call, string, bool) {
	if call.Setter || call.Virtual != 0 || len(call.Arguments) != 1 {
		return ir.Call{}, "", false
	}
	targets := program.CallTargets(call)
	if len(targets) != 1 {
		return ir.Call{}, "", false
	}
	f := program.Functions[targets[0]]
	if f.Closure || f.MayThrow || len(f.Parameters) != 1 || len(f.Body) != 2 {
		return ir.Call{}, "", false
	}
	branch, ok := f.Body[0].(ir.If)
	if !ok || len(branch.Then) != 1 || len(branch.Else) != 0 {
		return ir.Call{}, "", false
	}
	condition, ok := branch.Condition.(ir.HasAccessor)
	if !ok {
		return ir.Call{}, "", false
	}
	root, ok := condition.Object.(ir.Read)
	if !ok || root.Local != f.Parameters[0] {
		return ir.Call{}, "", false
	}
	returned, ok := branch.Then[0].(ir.Return)
	if !ok {
		return ir.Call{}, "", false
	}
	inner, ok := returned.Value.(ir.Call)
	if !ok || inner.Virtual != -1 || inner.Setter || inner.Accessor != condition.Name || inner.Returns != call.Returns || len(inner.Arguments) != 1 {
		return ir.Call{}, "", false
	}
	read, ok := inner.Arguments[0].(ir.Read)
	if !ok || read.Local != root.Local {
		return ir.Call{}, "", false
	}
	fallback, ok := f.Body[1].(ir.Return)
	if !ok {
		return ir.Call{}, "", false
	}
	field, ok := fallback.Value.(ir.Property)
	if !ok || field.Name != condition.Name || field.Of != call.Returns || field.Method || field.Optional || field.Absent {
		return ir.Call{}, "", false
	}
	read, ok = field.Object.(ir.Read)
	if !ok || read.Local != root.Local {
		return ir.Call{}, "", false
	}
	// A setter-only descriptor must retain the original missing-getter failure.
	for _, class := range program.Classes {
		for _, accessor := range class.Accessors {
			if accessor.Name == condition.Name && accessor.Getter < 0 {
				return ir.Call{}, "", false
			}
		}
	}
	inner.Arguments = call.Arguments
	return inner, condition.Name, true
}

func borrowableReturn(program *ir.Program, function int, declare ir.Declare, assigned map[int]bool, remaining []ir.Statement, preserving map[int]bool) (int, bool) {
	call, ok := declare.Value.(ir.Call)
	if !ok {
		return 0, false
	}
	if _, ok := returnedField(program, call); !ok {
		return 0, false
	}
	receiver, ok := variableRead(call.Arguments[0])
	if !ok {
		return 0, false
	}
	root, local := program.Locals[receiver.Local], program.Locals[declare.Local]
	if root.Global || root.Captured || root.Function != function || assigned[receiver.Local] ||
		local.Global || local.Captured || local.Function != function || assigned[declare.Local] {
		return 0, false
	}
	return receiver.Local, preservesFields(program, preserving, remaining) && parameterOnlyRead(program, remaining, declare.Local)
}

// preservingFunctions is a conservative whole-program fixed point. Even an
// unrelated field write stops a borrow: it could reach the receiver through an
// alias. Unknown calls and operations cannot prove preservation.
func preservingFunctions(program *ir.Program) map[int]bool {
	preserving := map[int]bool{}
	for function := range program.Functions {
		preserving[function] = true
	}
	for changed := true; changed; {
		changed = false
		for function, body := range program.Functions {
			if preserving[function] && !preservesFields(program, preserving, body.Body) {
				preserving[function] = false
				changed = true
			}
		}
	}
	return preserving
}

func preservesFields(program *ir.Program, preserving map[int]bool, body []ir.Statement) bool {
	safe := true
	var statements func([]ir.Statement)
	statements = func(list []ir.Statement) {
		for _, statement := range list {
			switch statement := statement.(type) {
			case ir.SetProperty, ir.SetIndex:
				safe = false
			case ir.Assign:
				if program.Locals[statement.Local].Type.IsReference() {
					safe = false
				}
			}
			walkStatement(statement, func(expression ir.Expression) {
				switch expression := expression.(type) {
				case ir.Call:
					for _, target := range program.CallTargets(expression) {
						if !preserving[target] {
							safe = false
						}
					}
				case ir.CallClosure:
					targets := program.ClosureTargets(expression)
					if targets.Unknown {
						safe = false
					}
					for _, target := range targets.Functions {
						if !preserving[target] {
							safe = false
						}
					}
				case ir.ObjectLiteral:
					if expression.Spread != nil {
						safe = false
					}
				case ir.ArrayLiteral, ir.MakeError, ir.HasAccessor:
					// Fresh containers and accessor lookup cannot invalidate an existing field.
				default:
					if !pureKind(expression) {
						safe = false
					}
				}
			}, statements)
		}
	}
	statements(body)
	return safe
}

func (e *emitter) borrowReturnedField(declare ir.Declare) {
	call := declare.Value.(ir.Call)
	field, ok := returnedField(e.program, call)
	if !ok {
		panic("native: planned borrowed return lost its field proof")
	}
	receiver := e.value(field.Object)
	// Snapshot the field at JavaScript's read point. Nothing in this scope may
	// replace it, and the receiver's owner is stable and protected against moves.
	value := fmt.Sprintf("(%s)adamic_object_field(%s, %s, &%s)->reference", cType(field.Of), receiver, cString(field.Name), e.cache())
	if _, name, ok := getterDispatch(e.program, call); ok {
		fallback := fmt.Sprintf("(%s)adamic_object_field(%s, %s, &%s)->reference", cType(field.Of), receiver, cString(name), e.cache())
		value = fmt.Sprintf("(adamic_accessor_find(%s, %s) != NULL ? %s : %s)", receiver, cString(name), value, fallback)
	}
	e.line("%s %s = %s;", cType(field.Of), e.localName(declare.Local), value)
	e.end()
}
