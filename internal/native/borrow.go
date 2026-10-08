package native

import (
	"reflect"

	"github.com/system-inc/adamic/internal/ir"
)

// Borrowed reads (docs/memory.md, "Borrowed parameters", extended to reads).
//
// A reference read from somewhere a call could write (a global, a captured variable, a field) is
// retained the moment JavaScript reads it, because something later in the statement might write
// that place and free the value while it's still to be used. That's a retain and a release around
// every text.charCodeAt(position) of a global text: ten million pairs in the tokenizer benchmark.
//
// The count buys nothing when nothing can run between the read and its use. So a read is lent,
// taken without a count, when it's a direct operand of a consumer: an operation that uses the
// value and is done with it (its result is a number, a boolean, a new value, or one with its own
// count), whose every operand is pure, meaning it writes no variable, field, element or map, calls
// no code, and frees nothing. Then from the read to the consumer's operation, nothing but pure
// evaluation happens, and the place read still holds the value.
//
// One more thing makes that true in C: a consumer's result is often a C expression the emitter
// hands its parent, evaluated wherever the parent puts it, possibly after a sibling's call. A
// consumer that lent a read has its result evaluated where it stands (snapshotted), right after its
// operands, as JavaScript evaluates it.
//
// value is the wrapper the emitter calls for every expression; evaluate does the work.
func (e *emitter) value(expression ir.Expression) string {
	e.depth++
	defer func() { e.depth-- }()
	outerLendAt, outerLent := e.lendAt, e.lent
	// A read evaluated here may be lent when its parent is a consumer that said so.
	e.lendable = e.depth == outerLendAt || e.borrowChain(expression)
	if consumes(expression) {
		e.lendAt = e.depth + 1
	} else {
		e.lendAt = -1
	}
	e.lent, e.self = false, false
	result := e.evaluate(expression)
	// self is whether this expression's own value is lent (a read that took no count), which its
	// parent must hear; lent is whether one of its operands was.
	self, lent := e.self, e.lent
	e.lendAt, e.lent, e.self = outerLendAt, outerLent || self, false
	if lent && consumes(expression) && !expression.Type().IsReference() && expression.Type() != 0 {
		result = e.snapshot(expression.Type(), result)
	}
	return result
}

// lendable reports whether a read of a value of this type can be lent: a reference, held as a
// pointer. A union holds its reference inside a box, and stays counted.
func lendable(valueType ir.Type) bool {
	return valueType.IsReference() && valueType != ir.Union
}

// consumes reports whether an expression is a consumer whose operands it can lend: pure, done with
// its operands when it's evaluated (not handing one on as its own value, as a conditional or ??
// does), and with every operand pure.
//
// Every pure operation whose value is one of its operands, as it is (a narrowing, a cast, a box),
// must be on the list below: on pure's list without being here, it lends the read it hands on, and
// its parent keeps that uncounted pointer while a later operand frees it (ir.Defined, once pure,
// did: borrow_defined_lent.a). TestPassThroughsAreNotConsumers holds every pure kind to a decision.
func consumes(expression ir.Expression) bool {
	switch expression.(type) {
	case ir.Read, ir.Conditional, ir.Coalesce, ir.Box, ir.Narrow, ir.Unwrap, ir.CheckedCast, ir.MaybeOf, ir.Defined, ir.Undefined,
		ir.NumberConstant, ir.BooleanConstant, ir.StringConstant:
		return false
	}
	return pure(expression)
}

// pure reports whether evaluating an expression can't write a variable, a field, an element or a
// map, call code, or free anything a statement holds: a fixed list of the IR's operations, whatever
// they're made of being pure too. Anything not on the list, an operation added later included, is
// not pure.
func pure(expression ir.Expression) bool {
	if !pureKind(expression) {
		return false
	}
	operands := true
	eachOperand(expression, func(operand ir.Expression) {
		if !pure(operand) {
			operands = false
		}
	})
	return operands
}

// pureKind reports whether an expression's own operation is on pure's list, whatever its operands.
func pureKind(expression ir.Expression) bool {
	if call, ok := expression.(ir.ObjectCall); ok {
		switch call.Method {
		case "errorMember", "errorCause", "errorErrors", "errorToString", "errorPrototype", "errorGetPrototype", "errorInstanceOf", "errorIsPrototypeOf", "errorEnumerable":
			return true
		}
	}
	switch expression.(type) {
	case ir.NumberConstant, ir.BooleanConstant, ir.StringConstant, ir.Read, ir.Undefined,
		ir.Unary, ir.Binary, ir.NumberToString, ir.BooleanToString, ir.Concat, ir.Length, ir.StringLength,
		ir.CharCodeAt, ir.StringIndex, ir.ArrayIndex, ir.Property, ir.MapGet, ir.MapHas, ir.MapSize, ir.HasOwn,
		ir.IsUndefined, ir.Unwrap, ir.MaybeOf, ir.Box, ir.Narrow, ir.TypeOf, ir.Conditional, ir.Coalesce,
		ir.MathCall, ir.NumberCall, ir.ToFixed, ir.NumberFormat, ir.Trim, ir.StringCall, ir.CodePoints,
		ir.ArraySearch, ir.CheckedCast, ir.UnionToString, ir.MaybeToString, ir.Defined, ir.InstanceOf:
	default:
		return false
	}
	return true
}

// eachOperand visits the expressions an IR node holds directly.
func eachOperand(expression ir.Expression, visit func(ir.Expression)) {
	var each func(value reflect.Value)
	each = func(value reflect.Value) {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return
			}
			if operand, ok := value.Interface().(ir.Expression); ok {
				visit(operand)
				return
			}
			each(value.Elem())
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				each(value.Field(index))
			}
		case reflect.Slice, reflect.Array:
			for index := 0; index < value.Len(); index++ {
				each(value.Index(index))
			}
		}
	}
	each(reflect.ValueOf(expression))
}

// lentArgument lends only a direct global read to a borrowed parameter. Every later
// argument is pure, and all possible callees are proved not to touch the global.
// touches also rejects unknown closure and callback effects. Reads of the global
// are conservatively refused along with writes.
func (e *emitter) lentArgument(call ir.Call, index int) (string, bool) {
	read, ok := call.Arguments[index].(ir.Read)
	if !ok || !e.program.Locals[read.Local].Global || !lendable(read.Of) {
		return "", false
	}
	for _, argument := range call.Arguments[index+1:] {
		if !pure(argument) {
			return "", false
		}
	}
	// Every target that can run must borrow this parameter, so each one's own parameters are
	// read rather than the static callee's.
	for _, target := range e.program.CallTargets(call) {
		parameters := e.program.Functions[target].Parameters
		if index >= len(parameters) || touches(e.program, target, read.Local, map[int]bool{}) || !e.program.Locals[parameters[index]].Borrowed || e.reuse.consumed[parameters[index]] {
			return "", false
		}
	}
	if read.Checked {
		e.checkReady(read.Local)
	}
	return e.snapshot(read.Of, e.localName(read.Local)), true
}

// chainRoot follows strong field loads only. A weak handle is not an owner of its
// target, so a weak read deliberately ends the chain. Names conservatively match
// every holder: the native IR does not retain the checker's related-holder types.
func chainRoot(expression ir.Expression, names map[string]bool) (int, bool) {
	switch expression := expression.(type) {
	case ir.Property:
		if expression.Method || expression.Optional || expression.Object.Type() == ir.Weak {
			return 0, false
		}
		names[expression.Name] = true
		return chainRoot(expression.Object, names)
	case ir.Defined:
		return chainRoot(expression.Value, names)
	case ir.Read:
		return expression.Local, true
	}
	return 0, false
}

// chainUnchanged proves fields remain held throughout a scope and every call it
// can make. Visiting a recursive target once suffices: all its statements are
// checked, not just the path returning to the call. Unknown callbacks refuse.
func chainUnchanged(program *ir.Program, body []ir.Statement, names map[string]bool) bool {
	seen := map[int]bool{}
	safe := true
	var statements func([]ir.Statement)
	var target func(int)
	target = func(index int) {
		if safe && !seen[index] {
			seen[index] = true
			statements(program.Functions[index].Body)
		}
	}
	statements = func(body []ir.Statement) {
		for _, statement := range body {
			if !safe {
				return
			}
			if write, ok := statement.(ir.SetProperty); ok && names[write.Name] {
				safe = false
			}
			walkStatement(statement, func(expression ir.Expression) {
				if !safe {
					return
				}
				switch expression := expression.(type) {
				case ir.Call:
					for _, index := range program.CallTargets(expression) {
						target(index)
					}
				case ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.MapForEach, ir.ArraySort:
					targets := program.ClosureTargets(expression)
					if targets.Unknown {
						safe = false
					}
					for _, index := range targets.Functions {
						target(index)
					}
				case ir.ObjectLiteral:
					if expression.Spread != nil {
						safe = false
					}
				case ir.ArrayLiteral, ir.ArrayPush, ir.MakeClosure, ir.MakeError:
					// These operations do not remove fields. Their operands are walked too.
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

// borrowChain extends the existing lent-read emission to a strong chain whose
// root and fields survive the whole function. Stores and owned returns still go
// through kept/retained; a declaration borrows only when its planner says so.
func (e *emitter) borrowChain(expression ir.Expression) bool {
	if _, ok := expression.(ir.Property); !ok || e.function == nil {
		return false
	}
	names := map[string]bool{}
	root, ok := chainRoot(expression, names)
	if !ok {
		return false
	}
	local := e.program.Locals[root]
	if local.Global || local.Captured || local.Function != e.functionIndex || assignedLocals(e.function.Body)[root] {
		return false
	}
	if e.reuse != nil {
		if e.reuse.consumed[root] {
			return false
		}
		for _, moves := range e.reuse.moves {
			if moves[root] {
				return false
			}
		}
		for _, spreads := range e.reuse.spreads {
			if spreads[root] {
				return false
			}
		}
		for _, arrays := range e.reuse.arrays {
			if arrays[root] {
				return false
			}
		}
	}
	return chainUnchanged(e.program, e.function.Body, names)
}
