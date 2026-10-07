package lower

import (
	"math"

	"github.com/system-inc/adamic/internal/ir"
)

// refineExceptionBounds proves generated range guards from constants, immutable
// bindings, closed-program numeric field stores and bounded induction variables.
// An unknown store poisons a field's range; no observed runtime value is used.
func (l *lowering) refineExceptionBounds() {
	program := l.result
	known := facts{assigned: map[int]bool{}, declared: map[int]ir.Expression{}, counters: map[int]span{}}
	fields := map[string]span{}
	unknown := map[string]bool{}
	stores := map[string][]ir.Expression{}
	functionAssigned := map[int]bool{}
	for _, function := range program.Functions {
		findAssigned(function.Body, functionAssigned)
	}
	findAssigned(program.Main, known.assigned)
	findDeclared(program.Main, known.declared)
	for _, function := range program.Functions {
		findAssigned(function.Body, known.assigned)
		findDeclared(function.Body, known.declared)
	}
	note := func(node any) bool {
		switch node := node.(type) {
		case ir.ObjectLiteral:
			for _, field := range append(append([]ir.Field{}, node.Fields...), node.Empty...) {
				if field.Value != nil && field.Value.Type() == ir.Number {
					stores[field.Name] = append(stores[field.Name], field.Value)
				} else {
					unknown[field.Name] = true
				}
			}
		case ir.SetProperty:
			if node.Value.Type() == ir.Number {
				stores[node.Name] = append(stores[node.Name], node.Value)
			} else {
				unknown[node.Name] = true
			}
		case ir.ObjectCall:
			if node.Method == "assign" {
				unknown["*"] = true
			}
		}
		return true
	}
	walk(program.Main, note)
	for _, function := range program.Functions {
		walk(function.Body, note)
	}
	var number func(ir.Expression, int) (span, bool)
	number = func(value ir.Expression, depth int) (span, bool) {
		if depth > 8 || value == nil {
			return span{}, false
		}
		switch value := value.(type) {
		case ir.NumberConstant:
			return span{value.Value, value.Value}, !math.IsNaN(value.Value) && !math.IsInf(value.Value, 0)
		case ir.Read:
			if bound, ok := known.counters[value.Local]; ok {
				return bound, true
			}
			if !known.assigned[value.Local] {
				return number(known.declared[value.Local], depth+1)
			}
		case ir.Property:
			bound, ok := fields[value.Name]
			return bound, ok && !unknown[value.Name] && !unknown["*"]
		case ir.Unary:
			bound, ok := number(value.Operand, depth+1)
			if value.Operator == ir.Negate {
				return span{-bound.high, -bound.low}, ok
			}
			if value.Operator == ir.Plus {
				return bound, ok
			}
		case ir.Binary:
			left, okLeft := number(value.Left, depth+1)
			right, okRight := number(value.Right, depth+1)
			if !okLeft || !okRight {
				return span{}, false
			}
			switch value.Operator {
			case ir.Add:
				return span{left.low + right.low, left.high + right.high}, true
			case ir.Subtract:
				return span{left.low - right.high, left.high - right.low}, true
			case ir.Multiply:
				products := []float64{left.low * right.low, left.low * right.high, left.high * right.low, left.high * right.high}
				bound := span{products[0], products[0]}
				for _, product := range products[1:] {
					bound.low = math.Min(bound.low, product)
					bound.high = math.Max(bound.high, product)
				}
				return bound, !math.IsNaN(bound.low) && !math.IsNaN(bound.high)
			}
		case ir.Call:
			bound, found, safe := span{}, false, true
			for _, target := range program.CallTargets(value) {
				walk(program.Functions[target].Body, func(node any) bool {
					if returned, ok := node.(ir.Return); ok {
						next, ok := number(returned.Value, depth+1)
						safe = safe && ok
						if !found {
							bound = next
						} else {
							bound.low = math.Min(bound.low, next.low)
							bound.high = math.Max(bound.high, next.high)
						}
						found = true
					}
					return true
				})
			}
			return bound, found && safe
		}
		return span{}, false
	}
	// Field facts start empty. Install a range only after every store is
	// proven, so circular field equations cannot start at an empty range.
	for name, values := range stores {
		bound, first := span{}, true
		for _, value := range values {
			next, ok := number(value, 0)
			if !ok {
				unknown[name] = true
				break
			}
			if first {
				bound = next
				first = false
			} else {
				bound.low = math.Min(bound.low, next.low)
				bound.high = math.Max(bound.high, next.high)
			}
		}
		if !unknown[name] {
			fields[name] = bound
		}
	}
	within := func(value ir.Expression, low, high float64) bool {
		bound, ok := number(value, 0)
		return ok && math.Trunc(bound.low) >= low && math.Trunc(bound.high) <= high
	}
	simplify := func(node any) any {
		call, ok := node.(ir.Call)
		if !ok || call.Virtual != 0 || program.Functions[call.Function].Name != "error_checked_library" {
			return node
		}
		function := program.Functions[call.Function]
		operation := function.Body[1].(ir.Return).Value
		operation = rewriteChecks(operation, func(node any) any {
			if read, ok := node.(ir.Read); ok {
				for index, local := range function.Parameters {
					if read.Local == local {
						return call.Arguments[index]
					}
				}
			}
			return node
		}).(ir.Expression)
		safe := false
		switch operation := operation.(type) {
		case ir.ToFixed:
			safe = within(operation.Digits, 0, 100)
		case ir.NumberFormat:
			bounds := formatArguments[operation.Method]
			safe = within(operation.Argument, bounds[0], bounds[1])
		case ir.StringCall:
			if operation.Method == "normalize" {
				safe = isNormalizationForm(operation.Arguments[0], program.Strings)
			}
			if operation.Method == "repeat" {
				safe = within(operation.Arguments[0], 0, 1)
			}
		}
		if safe {
			return operation
		}
		return node
	}
	var body func([]ir.Statement) []ir.Statement
	body = func(statements []ir.Statement) []ir.Statement {
		result := append([]ir.Statement(nil), statements...)
		for index, statement := range result {
			if block, ok := statement.(ir.Block); ok && len(block.Body) > 0 {
				if loop, ok := block.Body[len(block.Body)-1].(ir.Loop); ok && !loop.CheckAfter && len(loop.Update) == 1 {
					if condition, ok := loop.Condition.(ir.Binary); ok {
						if read, ok := condition.Left.(ir.Read); ok && (!program.Locals[read.Local].Global || !functionAssigned[read.Local]) && !program.Locals[read.Local].Captured {
							start, declared := declaredIn(block, read.Local)
							step, stepping := stepOf(loop.Update[0], read.Local)
							written := map[int]bool{}
							findAssigned(loop.Body, written)
							low, high, bounded := boundRange(condition.Right, known, 0)
							startLow, startHigh, started := boundRange(start, known, 0)
							if declared && stepping && !written[read.Local] && bounded && started && ((step > 0 && (condition.Operator == ir.Less || condition.Operator == ir.LessOrEqual) && high <= exact-step) || (step < 0 && (condition.Operator == ir.Greater || condition.Operator == ir.GreaterOrEqual) && low >= -exact-step)) {
								bound := span{startLow, high}
								if step < 0 {
									bound = span{low, startHigh}
								}
								previous, had := known.counters[read.Local]
								known.counters[read.Local] = bound
								loop.Body = body(loop.Body)
								if had {
									known.counters[read.Local] = previous
								} else {
									delete(known.counters, read.Local)
								}
								block.Body = append(append([]ir.Statement{}, block.Body[:len(block.Body)-1]...), loop)
								result[index] = block
								continue
							}
						}
					}
				}
			}
			// Recurse through nested bodies while their enclosing counter facts hold.
			switch node := statement.(type) {
			case ir.Block:
				node.Body = body(node.Body)
				statement = node
			case ir.If:
				node.Then = body(node.Then)
				node.Else = body(node.Else)
				statement = node
			case ir.Loop:
				node.Body = body(node.Body)
				node.Update = body(node.Update)
				statement = node
			case ir.ForOf:
				node.Body = body(node.Body)
				statement = node
			case ir.Try:
				node.Body = body(node.Body)
				node.Catch = body(node.Catch)
				node.Finally = body(node.Finally)
				statement = node
			}
			result[index] = rewriteChecks(statement, simplify).(ir.Statement)
		}
		return result
	}
	program.Main = body(program.Main)
	for index := range program.Functions {
		program.Functions[index].Body = body(program.Functions[index].Body)
	}
}
