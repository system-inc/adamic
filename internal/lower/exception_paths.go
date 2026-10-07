package lower

import (
	"fmt"
	"reflect"

	"github.com/system-inc/adamic/internal/ir"
)

// refineExceptionPaths carries successful presence tests to later reads on the
// same path. A mutation drops the facts; a join keeps only facts from both arms.
// Loops and handlers start without incoming facts, so back edges and exceptional
// entries cannot inherit a test that ran on only one earlier path.
func (l *lowering) refineExceptionPaths() {
	program := l.result
	mutates := make([]bool, len(program.Functions))
	for changed := true; changed; {
		changed = false
		for index, function := range program.Functions {
			if mutates[index] {
				continue
			}
			walk(function.Body, func(node any) bool {
				writes := false
				switch node := node.(type) {
				case ir.Assign, ir.SetProperty, ir.SetIndex, ir.ArrayPush, ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.MapForEach, ir.ArraySort:
					writes = true
				case ir.Call:
					for _, target := range program.CallTargets(node) {
						writes = writes || mutates[target]
					}
				default:
					// Unknown runtime operations may mutate a receiver. Only
					// the existing lowering purity list can exclude that.
					if expression, ok := node.(ir.Expression); ok {
						writes = !pathPureKind(expression)
					}
				}
				if writes {
					mutates[index] = true
					changed = true
				}
				return true
			})
		}
	}
	declared, assigned := map[int]ir.Expression{}, map[int]bool{}
	findDeclared(program.Main, declared)
	findAssigned(program.Main, assigned)
	for _, function := range program.Functions {
		findDeclared(function.Body, declared)
		findAssigned(function.Body, assigned)
	}
	writes := make([]map[int]bool, len(program.Functions))
	for index, function := range program.Functions {
		writes[index] = map[int]bool{}
		findAssigned(function.Body, writes[index])
		walk(function.Body, func(node any) bool {
			switch node.(type) {
			case ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.MapForEach, ir.ArraySort:
				for local := range assigned {
					writes[index][local] = true
				}
			}
			return true
		})
	}
	for changed := true; changed; {
		changed = false
		for index, function := range program.Functions {
			walk(function.Body, func(node any) bool {
				if call, ok := node.(ir.Call); ok {
					for _, target := range program.CallTargets(call) {
						for local := range writes[target] {
							if !writes[index][local] {
								writes[index][local] = true
								changed = true
							}
						}
					}
				}
				return true
			})
		}
	}
	roots := map[string]int{}
	for local := range program.Locals {
		roots[presenceKey(ir.Read{Local: local})] = local
	}
	var present func(ir.Expression, int) bool
	present = func(value ir.Expression, depth int) bool {
		if depth > 8 || value == nil {
			return false
		}
		switch value := value.(type) {
		case ir.ObjectLiteral, ir.ArrayLiteral, ir.StringConstant, ir.MakeClosure, ir.Defined:
			return true
		case ir.Read:
			return !assigned[value.Local] && present(declared[value.Local], depth+1)
		case ir.Call:
			found, safe := false, true
			for _, target := range program.CallTargets(value) {
				function := program.Functions[target]
				if checkedReferenceReturn(function, program.Strings) {
					found = true
					continue
				}
				walk(function.Body, func(node any) bool {
					if returned, ok := node.(ir.Return); ok {
						found = true
						safe = safe && present(returned.Value, depth+1)
					}
					return true
				})
			}
			return found && safe
		}
		return false
	}
	refine := func(body []ir.Statement) []ir.Statement {
		var transform func(reflect.Value, map[string]int) reflect.Value
		transform = func(value reflect.Value, facts map[string]int) reflect.Value {
			if value.Kind() == reflect.Interface && !value.IsNil() {
				result := reflect.New(value.Type()).Elem()
				result.Set(transform(value.Elem(), facts))
				return result
			}
			if value.Kind() == reflect.Slice {
				if value.IsNil() {
					return value
				}
				result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
				for index := 0; index < value.Len(); index++ {
					result.Index(index).Set(transform(value.Index(index), facts))
				}
				return result
			}
			if value.Kind() != reflect.Struct {
				return value
			}
			if node, ok := value.Interface().(ir.Conditional); ok {
				node.Condition = transform(reflect.ValueOf(node.Condition), facts).Interface().(ir.Expression)
				left, right := copyPresence(facts), copyPresence(facts)
				assumePresence(node.Condition, true, left)
				assumePresence(node.Condition, false, right)
				node.WhenTrue = transform(reflect.ValueOf(node.WhenTrue), left).Interface().(ir.Expression)
				node.WhenNot = transform(reflect.ValueOf(node.WhenNot), right).Interface().(ir.Expression)
				joinPresence(facts, left, right)
				return reflect.ValueOf(node)
			}
			if node, ok := value.Interface().(ir.If); ok {
				node.Condition = transform(reflect.ValueOf(node.Condition), facts).Interface().(ir.Expression)
				left, right := copyPresence(facts), copyPresence(facts)
				assumePresence(node.Condition, true, left)
				assumePresence(node.Condition, false, right)
				node.Then = transform(reflect.ValueOf(node.Then), left).Interface().([]ir.Statement)
				node.Else = transform(reflect.ValueOf(node.Else), right).Interface().([]ir.Statement)
				joinPresence(facts, left, right)
				return reflect.ValueOf(node)
			}
			if node, ok := value.Interface().(ir.Evaluate); ok {
				if call, ok := node.Value.(ir.Call); ok && call.Virtual == 0 && program.Functions[call.Function].Name == "error_set_property" && facts[presenceKey(call.Arguments[0])]&1 != 0 {
					write := program.Functions[call.Function].Body[1].(ir.SetProperty)
					write.Object = transform(reflect.ValueOf(call.Arguments[0]), facts).Interface().(ir.Expression)
					write.Value = transform(reflect.ValueOf(call.Arguments[1]), facts).Interface().(ir.Expression)
					clear(facts)
					return reflect.ValueOf(write)
				}
			}
			if node, ok := value.Interface().(ir.Loop); ok {
				// Each body entry follows a successful condition, including back edges.
				// Earlier iterations supply no facts; writes in the body still invalidate them.
				node.Condition = transform(reflect.ValueOf(node.Condition), map[string]int{}).Interface().(ir.Expression)
				inside := map[string]int{}
				assumePresence(node.Condition, true, inside)
				node.Body = transform(reflect.ValueOf(node.Body), inside).Interface().([]ir.Statement)
				node.Update = transform(reflect.ValueOf(node.Update), inside).Interface().([]ir.Statement)
				clear(facts)
				return reflect.ValueOf(node)
			}
			if node, ok := value.Interface().(ir.Try); ok {
				node.Body = transform(reflect.ValueOf(node.Body), map[string]int{}).Interface().([]ir.Statement)
				node.Catch = transform(reflect.ValueOf(node.Catch), map[string]int{}).Interface().([]ir.Statement)
				node.Finally = transform(reflect.ValueOf(node.Finally), map[string]int{}).Interface().([]ir.Statement)
				clear(facts)
				return reflect.ValueOf(node)
			}
			if node, ok := value.Interface().(ir.Switch); ok {
				node.Value = transform(reflect.ValueOf(node.Value), facts).Interface().(ir.Expression)
				for index := range node.Cases {
					node.Cases[index].Tests = transform(reflect.ValueOf(node.Cases[index].Tests), map[string]int{}).Interface().([]ir.Expression)
					node.Cases[index].Body = transform(reflect.ValueOf(node.Cases[index].Body), map[string]int{}).Interface().([]ir.Statement)
				}
				node.Default = transform(reflect.ValueOf(node.Default), map[string]int{}).Interface().([]ir.Statement)
				clear(facts)
				return reflect.ValueOf(node)
			}
			switch value.Interface().(type) {
			case ir.Loop, ir.ForOf, ir.Try:
				clear(facts)
			}
			result := reflect.New(value.Type()).Elem()
			for index := 0; index < value.NumField(); index++ {
				result.Field(index).Set(transform(value.Field(index), facts))
			}
			node := result.Interface()
			switch typed := node.(type) {
			case ir.Call:
				function := program.Functions[typed.Function]
				if typed.Virtual == 0 && function.Name == "error_defined" {
					guard := function.Body[0].(ir.If)
					bit := 1
					_, null := guard.Condition.(ir.IsNull)
					if null {
						bit = 2
					}
					if facts[presenceKey(typed.Arguments[0])]&bit != 0 || present(typed.Arguments[0], 0) {
						failure := guard.Then[0].(ir.Throw).Value.(ir.Call)
						message := program.Strings[failure.Arguments[0].(ir.StringConstant).Index]
						return reflect.ValueOf(ir.Defined{Proven: true, Value: typed.Arguments[0], Null: null, Message: "TypeError: " + message})
					}
				}
				for _, target := range program.CallTargets(typed) {
					if mutates[target] {
						for key := range facts {
							local, root := roots[key]
							if !root || writes[target][local] {
								delete(facts, key)
							}
						}
					}
				}
			case ir.Declare:
				if present(typed.Value, 0) {
					facts[presenceKey(ir.Read{Local: typed.Local})] = 3
				}
			case ir.Assign:
				clear(facts)
				if present(typed.Value, 0) {
					facts[presenceKey(ir.Read{Local: typed.Local})] = 3
				}
			case ir.SetProperty:
				clear(facts)
				if present(typed.Value, 0) {
					facts[presenceKey(ir.Property{Object: typed.Object, Name: typed.Name})] = 3
				}
			case ir.ArrayPush, ir.RegExpCall:
				// These runtime operations mutate objects but cannot assign a source binding.
				for key := range facts {
					if _, local := roots[key]; !local {
						delete(facts, key)
					}
				}
			case ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.MapForEach, ir.ArraySort, ir.SetIndex, ir.Loop, ir.ForOf, ir.Try, ir.Switch:
				clear(facts)
			default:
				if expression, ok := node.(ir.Expression); ok && !pathPureKind(expression) {
					clear(facts)
				}
			}
			return result
		}
		return transform(reflect.ValueOf(body), map[string]int{}).Interface().([]ir.Statement)
	}
	program.Main = refine(program.Main)
	for index := range program.Functions {
		program.Functions[index].Body = refine(program.Functions[index].Body)
	}
}

func presenceKey(value ir.Expression) string {
	switch value := value.(type) {
	case ir.Read:
		return fmt.Sprintf("local:%d", value.Local)
	case ir.Property:
		return fmt.Sprintf("field:%q:%q", presenceKey(value.Object), value.Name)
	case ir.Defined:
		return presenceKey(value.Value)
	case ir.WeakTarget:
		return "weak:" + presenceKey(value.Value)
	}
	return fmt.Sprintf("%#v", value)
}

func assumePresence(condition ir.Expression, truth bool, facts map[string]int) {
	switch condition := condition.(type) {
	case ir.Unary:
		if condition.Operator == ir.Not {
			assumePresence(condition.Operand, !truth, facts)
		}
	case ir.IsUndefined:
		if !truth {
			facts[presenceKey(condition.Value)] |= 1
		}
	case ir.IsNull:
		if !truth {
			facts[presenceKey(condition.Value)] |= 2
		}
	}
}
func copyPresence(facts map[string]int) map[string]int {
	copy := map[string]int{}
	for key, value := range facts {
		copy[key] = value
	}
	return copy
}
func joinPresence(facts, left, right map[string]int) {
	clear(facts)
	for key, value := range left {
		if common := value & right[key]; common != 0 {
			facts[key] = common
		}
	}
}

// A closed list: anything newly added to IR starts conservatively mutating.
func pathPureKind(expression ir.Expression) bool {
	switch expression.(type) {
	case ir.NumberConstant, ir.BooleanConstant, ir.StringConstant, ir.Read, ir.Undefined,
		ir.Unary, ir.Binary, ir.NumberToString, ir.BooleanToString, ir.Concat, ir.Length, ir.StringLength,
		ir.CharCodeAt, ir.StringIndex, ir.ArrayIndex, ir.Property, ir.MapGet, ir.MapHas, ir.MapSize, ir.HasOwn,
		ir.IsUndefined, ir.IsNull, ir.Unwrap, ir.MaybeOf, ir.Box, ir.Narrow, ir.TypeOf, ir.Conditional, ir.Coalesce,
		ir.MathCall, ir.NumberCall, ir.ToFixed, ir.NumberFormat, ir.Trim, ir.StringCall, ir.CodePoints,
		ir.ArraySearch, ir.CheckedCast, ir.UnionToString, ir.MaybeToString, ir.Defined, ir.InstanceOf,
		ir.ObjectLiteral, ir.ArrayLiteral, ir.MakeClosure, ir.WeakTarget, ir.WeakOf, ir.ArrayJoin:
		return true
	}
	return false
}

// A checked union conversion can return a reference only after its matching
// runtime tag test succeeds. A second undefined check cannot fail afterward.
func checkedReferenceReturn(function ir.Function, strings []string) bool {
	if len(function.Body) != 2 {
		return false
	}
	guard, ok := function.Body[0].(ir.If)
	if !ok || len(guard.Then) != 1 || len(guard.Else) != 0 {
		return false
	}
	if _, stop := guard.Then[0].(ir.Panic); !stop {
		return false
	}
	returned, ok := function.Body[1].(ir.Return)
	if !ok {
		return false
	}
	narrowed, ok := returned.Value.(ir.Narrow)
	if !ok || (narrowed.To != ir.String && narrowed.To != ir.Closure) {
		return false
	}
	read, ok := narrowed.Value.(ir.Read)
	if !ok || read.Of != ir.Union {
		return false
	}
	inverse, ok := guard.Condition.(ir.Unary)
	if !ok || inverse.Operator != ir.Not {
		return false
	}
	equal, ok := inverse.Operand.(ir.Binary)
	if !ok || equal.Operator != ir.Equal {
		return false
	}
	tag, ok := equal.Left.(ir.TypeOf)
	if !ok {
		return false
	}
	checked, ok := tag.Value.(ir.Read)
	if !ok || checked.Local != read.Local || checked.Of != ir.Union {
		return false
	}
	name, ok := equal.Right.(ir.StringConstant)
	if !ok {
		return false
	}
	expected := "string"
	if narrowed.To == ir.Closure {
		expected = "function"
	}
	return strings[name.Index] == expected
}
