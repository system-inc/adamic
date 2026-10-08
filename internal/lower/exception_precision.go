package lower

import (
	"reflect"

	"github.com/system-inc/adamic/internal/ir"
)

// preciseChecks removes only checks whose failure is excluded by the program.
// Readiness is monotone at runtime, but a function may run during module
// initialization. Its entry facts are the intersection of all possible callers,
// including virtual calls and callbacks, solved to a fixed point.
func (l *lowering) preciseChecks() {
	program := l.result
	entries := make([]map[int]bool, len(program.Functions))
	dynamic := map[int]bool{}
	for _, instance := range l.instances {
		for _, method := range instance.methodList() {
			dynamic[method.Function] = true
		}
	}
	for index, function := range program.Functions {
		entries[index] = map[int]bool{}
		for local, declaration := range program.Locals {
			if declaration.Global {
				entries[index][local] = true
			}
		}
		if function.Closure {
			dynamic[index] = true
		}
	}
	for changed := true; changed; {
		changed = false
		call := func(target int, ready map[int]bool) {
			for local := range entries[target] {
				if !ready[local] {
					delete(entries[target], local)
					changed = true
				}
			}
		}
		var scan func(any, map[int]bool)
		scan = func(value any, ready map[int]bool) {
			walk(value, func(node any) bool {
				switch node := node.(type) {
				case ir.Call:
					for _, target := range program.CallTargets(node) {
						call(target, ready)
					}
				case ir.CallClosure, ir.ArrayMap, ir.ArrayVisit, ir.ArrayReduce, ir.ArrayFrom, ir.MapForEach:
					for target := range dynamic {
						call(target, ready)
					}
				case ir.ArraySort:
					if node.Callback == nil {
						call(node.Comparator, ready)
					} else {
						for target := range dynamic {
							call(target, ready)
						}
					}
				}
				return true
			})
		}
		var statements func([]ir.Statement, map[int]bool)
		statements = func(body []ir.Statement, ready map[int]bool) {
			for _, statement := range body {
				switch node := statement.(type) {
				case ir.Declare:
					scan(node.Value, ready)
					ready[node.Local] = true
				case ir.Block:
					statements(node.Body, ready)
				case ir.If:
					scan(node.Condition, ready)
					left, right := copyReady(ready), copyReady(ready)
					statements(node.Then, left)
					statements(node.Else, right)
					for local := range left {
						if right[local] {
							ready[local] = true
						}
					}
				default:
					// Loops may run zero times; try can leave at any point. Do
					// not export declarations from either, and scan all their
					// calls with the incoming facts.
					scan(statement, ready)
				}
			}
		}
		statements(program.Main, map[int]bool{})
		for index, function := range program.Functions {
			statements(function.Body, copyReady(entries[index]))
		}
	}
	for index := range program.Functions {
		program.Functions[index].Body = rewriteChecks(program.Functions[index].Body, func(node any) any {
			switch node := node.(type) {
			case ir.Read:
				if entries[index][node.Local] {
					node.NoMove = node.NoMove || node.Checked
					node.Checked = false
				}
				return node
			case ir.Assign:
				if entries[index][node.Local] {
					node.Checked = false
				}
				return node
			}
			return node
		}).([]ir.Statement)
	}
	// An uncaptured local cannot be changed by a call. The checker's own
	// control-flow narrowing is valid there, so a property receiver needs no
	// throwing wrapper. Keep Defined as the existing pure representation.
	assigned := map[int]bool{}
	fieldsWritten := map[string]bool{}
	noteWrites := func(node any) bool {
		switch node := node.(type) {
		case ir.Assign:
			assigned[node.Local] = true
		case ir.SetProperty:
			fieldsWritten[node.Name] = true
		case ir.ObjectCall:
			if node.Method == "assign" {
				fieldsWritten["*"] = true
			}
		}
		return true
	}
	walk(program.Main, noteWrites)
	for _, function := range program.Functions {
		walk(function.Body, noteWrites)
	}
	var stable func(ir.Expression) bool
	stable = func(value ir.Expression) bool {
		switch value := value.(type) {
		case ir.Read:
			local := program.Locals[value.Local]
			return (!local.Global && !local.Captured) || !assigned[value.Local]
		case ir.Narrow:
			return stable(value.Value)
		case ir.Property:
			return !fieldsWritten["*"] && !fieldsWritten[value.Name] && stable(value.Object)
		}
		return false
	}
	simplify := func(node any) any {
		if evaluation, ok := node.(ir.Evaluate); ok {
			if call, ok := evaluation.Value.(ir.Call); ok && call.Virtual == 0 && program.Functions[call.Function].Name == "error_set_property" && stable(call.Arguments[0]) {
				write := program.Functions[call.Function].Body[1].(ir.SetProperty)
				write.Object, write.Value = call.Arguments[0], call.Arguments[1]
				return write
			}
		}

		call, ok := node.(ir.Call)
		if !ok || call.Virtual != 0 || program.Functions[call.Function].Name != "error_defined" {
			return node
		}
		value := call.Arguments[0]
		if !stable(value) {
			return node
		}
		guard := program.Functions[call.Function].Body[0].(ir.If)
		failure := guard.Then[0].(ir.Throw).Value.(ir.Call)
		message := program.Strings[failure.Arguments[0].(ir.StringConstant).Index]
		_, null := guard.Condition.(ir.IsNull)
		return ir.Defined{Value: value, Message: "TypeError: " + message, Null: null}
	}
	program.Main = rewriteChecks(program.Main, simplify).([]ir.Statement)
	for index := range program.Functions {
		program.Functions[index].Body = rewriteChecks(program.Functions[index].Body, simplify).([]ir.Statement)
	}
}

func copyReady(ready map[int]bool) map[int]bool {
	copy := map[int]bool{}
	for local := range ready {
		copy[local] = true
	}
	return copy
}

// Rebuild value nodes, preserving sites and all fields the analysis does not
// change. IR nodes are values; modifying a reflected interface in place would
// otherwise silently leave the old node in its parent.
func rewriteChecks(value any, rewrite func(any) any) any {
	var transform func(reflect.Value) reflect.Value
	transform = func(value reflect.Value) reflect.Value {
		switch value.Kind() {
		case reflect.Interface:
			if value.IsNil() {
				return value
			}
			result := reflect.New(value.Type()).Elem()
			result.Set(transform(value.Elem()))
			return result
		case reflect.Slice:
			if value.IsNil() {
				return value
			}
			result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
			for index := 0; index < value.Len(); index++ {
				result.Index(index).Set(transform(value.Index(index)))
			}
			return result
		case reflect.Struct:
			result := reflect.New(value.Type()).Elem()
			for index := 0; index < value.NumField(); index++ {
				result.Field(index).Set(transform(value.Field(index)))
			}
			if value.Type().Implements(expressionType) || value.Type().Implements(statementType) {
				return reflect.ValueOf(rewrite(result.Interface()))
			}
			return result
		}
		return value
	}
	return transform(reflect.ValueOf(value)).Interface()
}
