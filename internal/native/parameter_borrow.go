package native

import "github.com/system-inc/adamic/internal/ir"

// inferParameterBorrows refines lowering's lifetime flags into read-only conventions.
// The fixed point starts at borrowing and only removes proofs. A recursive read-only
// walk therefore borrows, while a store in any reachable implementation propagates.
// Local.Borrowed remains the single ownership flag consumed by emission and reuse.
func inferParameterBorrows(program *ir.Program) {
	changing := changingFunctions(program)
	for index, function := range program.Functions {
		assigned := assignedLocals(function.Body)
		for _, parameter := range function.Parameters {
			local := &program.Locals[parameter]
			local.Borrowed = lendable(local.Type) && !local.Captured && !assigned[parameter] && (!function.Closure || !changing[index])
		}
	}
	for changed := true; changed; {
		changed = false
		for _, function := range program.Functions {
			for _, parameter := range function.Parameters {
				if program.Locals[parameter].Borrowed && !parameterOnlyRead(program, function.Body, parameter) {
					program.Locals[parameter].Borrowed = false
					changed = true
				}
			}
		}
	}
}

// parameterOnlyRead follows immutable local aliases and reads through containers.
// Unknown operations are owning. A local alias may stay within the call, but a
// store, return, capture, mutation or owning call stops the proof.
func parameterOnlyRead(program *ir.Program, body []ir.Statement, parameter int) bool {
	aliases := map[int]bool{parameter: true}
	assigned := assignedLocals(body)
	owned := false
	var derived func(ir.Expression) bool
	derived = func(expression ir.Expression) bool {
		switch expression := expression.(type) {
		case nil:
			return false
		case ir.Read:
			return aliases[expression.Local] && expression.Of.IsReference()
		case ir.Property:
			return derived(expression.Object) && expression.Of.IsReference()
		case ir.ArrayIndex:
			source := derived(expression.Array)
			derived(expression.Index)
			return source && expression.Element.IsReference()
		case ir.Defined:
			return derived(expression.Value)
		case ir.Narrow:
			return derived(expression.Value)
		case ir.Unwrap:
			return derived(expression.Value)
		case ir.CheckedCast:
			return derived(expression.Value)
		case ir.HasAccessor:
			derived(expression.Object)
			return false
		case ir.Call:
			for position, argument := range expression.Arguments {
				if derived(argument) && !program.CallBorrows(expression, position) {
					owned = true
				}
			}
			return false
		case ir.CallClosure:
			if receiver := program.ClosureReceiver(expression); receiver != nil {
				if derived(receiver) && !program.ClosureReceiverBorrows(expression) {
					owned = true
				}
			}
			// Evaluate nested effects in the callee expression as well as its arguments.
			eachOperand(expression, func(operand ir.Expression) { derived(operand) })
			for position, argument := range expression.Arguments {
				if derived(argument) && !program.ClosureBorrows(expression, position) {
					owned = true
				}
			}
			return false
		}
		operands := false
		eachOperand(expression, func(operand ir.Expression) { operands = derived(operand) || operands })
		if operands && !consumes(expression) {
			owned = true
		}
		return false
	}
	var statements func([]ir.Statement)
	statements = func(list []ir.Statement) {
		for _, statement := range list {
			switch statement := statement.(type) {
			case ir.Return:
				// Returning a direct field takes an independent count on the field, not
				// the receiver. The return can later be lent within the receiver's lifetime.
				if field, ok := statement.Value.(ir.Property); ok {
					if receiver, ok := field.Object.(ir.Read); ok && receiver.Local == parameter && !field.Method && !field.Optional {
						continue
					}
				}
				if derived(statement.Value) {
					owned = true
				}
				continue
			case ir.Declare:
				if derived(statement.Value) {
					local := program.Locals[statement.Local]
					if local.Global || local.Captured || assigned[statement.Local] {
						owned = true
					} else {
						aliases[statement.Local] = true
					}
				}
				continue
			case ir.ForOf:
				if derived(statement.Iterable) {
					if statement.Pattern != nil || program.Locals[statement.Local].Captured || assigned[statement.Local] {
						owned = true
					} else {
						aliases[statement.Local] = true
					}
				}
				statements(statement.Body)
				continue
			case ir.SetProperty:
				if derived(statement.Object) || derived(statement.Value) {
					owned = true
				}
				continue
			case ir.SetIndex:
				if derived(statement.Array) || derived(statement.Value) {
					owned = true
				}
				derived(statement.Index)
				continue
			}
			eachOperandOfStatement(statement, func(expression ir.Expression) {
				if derived(expression) {
					owned = true
				}
			})
			walkStatement(statement, func(ir.Expression) {}, statements)
		}
	}
	// Aliases can be encountered on a later branch or loop pass. Repeat until the
	// set no longer grows, so those uses cannot hide a store in an earlier branch.
	for previous := -1; previous != len(aliases); {
		previous = len(aliases)
		statements(body)
	}
	return !owned
}
