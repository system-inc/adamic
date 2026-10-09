package lower

import (
	"github.com/system-inc/adamic/internal/ir"
	"slices"
)

// closedFrameInput proves a narrow exception to type-based capture cycles.
// Every call supplies a literal of scalars and capture-free closures, the
// parameter is never replaced, and no operation can rewrite an object field.
// This is a closed input graph proof, not freshness of arbitrary captured cells.
func (l *lowering) closedFrameInput(local int) bool {
	owner := l.result.Locals[local].Function
	if owner < 0 || l.result.Functions[owner].Closure {
		return false
	}
	position := slices.Index(l.result.Functions[owner].Parameters, local)
	if position < 0 {
		return false
	}
	safe, called := true, false
	inspect := func(node any) bool {
		switch value := node.(type) {
		case ir.Assign:
			if value.Local == local {
				safe = false
			}
		case ir.MakeClosure:
			if value.Function == owner {
				safe = false
			}
		case ir.Call:
			if !slices.Contains(l.result.CallTargets(value), owner) {
				break
			}
			called = true
			if value.Virtual != 0 || position >= len(value.Arguments) || !l.closedInputLiteral(value.Arguments[position]) {
				safe = false
			}
		}
		if expression, ok := node.(ir.Expression); ok {
			switch expression.(type) {
			case ir.NumberConstant, ir.BooleanConstant, ir.StringConstant, ir.Undefined, ir.Read,
				ir.Binary, ir.Unary, ir.Conditional, ir.Coalesce, ir.Box, ir.Narrow, ir.Unwrap,
				ir.MaybeOf, ir.Defined, ir.IsUndefined, ir.NumberToString, ir.BooleanToString,
				ir.Concat, ir.Length, ir.Property, ir.ArrayIndex, ir.ArrayLiteral, ir.ObjectLiteral,
				ir.Call, ir.CallClosure, ir.MakeClosure, ir.ClosureSelf:
			default:
				safe = false
			}
		}
		if statement, ok := node.(ir.Statement); ok {
			switch statement.(type) {
			case ir.AllocateEnvironment, ir.Declare, ir.Assign, ir.Evaluate, ir.Return, ir.WriteLine,
				ir.If, ir.Block, ir.Loop, ir.ForOf, ir.Switch, ir.Break, ir.Continue:
			default:
				safe = false
			}
		}
		return true
	}
	for _, function := range l.result.Functions {
		walk(function.Body, inspect)
	}
	walk(l.result.Main, inspect)
	return safe && called
}

func (l *lowering) closedInputLiteral(value ir.Expression) bool {
	literal, ok := value.(ir.ObjectLiteral)
	if !ok || literal.Spread != nil || literal.Class != 0 {
		return false
	}
	for _, field := range literal.Fields {
		switch value := field.Value.(type) {
		case ir.NumberConstant, ir.BooleanConstant, ir.StringConstant, ir.Undefined:
		case ir.MakeClosure:
			if len(l.result.Functions[value.Function].Environment) != 0 {
				return false
			}
		default:
			return false
		}
	}
	return true
}
