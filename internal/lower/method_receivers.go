package lower

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/system-inc/adamic/internal/ir"
)

// A TypeScript method's apparent object type can hold undefined after extraction.
// Follow that receiver through the lowered program before deciding which reads check.
// Slot names and callable parameters join conservatively across possible holders.
func (l *lowering) finishReceiverValues() error {
	nullable := map[int]bool{}
	slots := map[string]bool{}
	returns := map[int]bool{}
	arrays, maps := false, false
	for index, local := range l.result.Locals {
		nullable[index] = local.NullableReceiver
	}
	count := len(l.result.Functions)
	guard := func(function int) bool {
		name := l.result.Functions[function].Name
		return name == "extracted_this" || name == "receiver_read" || name == "receiver_write"
	}
	var carries func(ir.Expression) bool
	carries = func(value ir.Expression) bool {
		switch value := value.(type) {
		case ir.Read:
			return nullable[value.Local]
		case ir.Property:
			return slots[value.Name] || (value.Optional && carries(value.Object))
		case ir.ArrayIndex:
			return arrays
		case ir.ArrayPop:
			return arrays
		case ir.MapGet:
			return maps
		case ir.Conditional:
			return carries(value.WhenTrue) || carries(value.WhenNot)
		case ir.Coalesce:
			return carries(value.Fallback)
		case ir.Defined:
			return carries(value.Value)
		case ir.Narrow:
			return carries(value.Value)
		case ir.Box:
			return carries(value.Value)
		case ir.WeakOf:
			return carries(value.Value)
		case ir.WeakTarget:
			return carries(value.Value)
		case ir.Call:
			if guard(value.Function) {
				return false
			}
			for _, target := range l.result.CallTargets(value) {
				if returns[target] {
					return true
				}
			}
		case ir.CallClosure:
			for function := range returns {
				if returns[function] && l.result.Functions[function].Returns == value.Returns {
					return true
				}
			}
		}
		return false
	}
	for changed := true; changed; {
		changed = false
		mark := func(local int) {
			if !nullable[local] {
				nullable[local] = true
				changed = true
			}
		}
		slot := func(name string) {
			if !slots[name] {
				slots[name] = true
				changed = true
			}
		}
		note := func(body []ir.Statement, function int) {
			walk(body, func(node any) bool {
				switch node := node.(type) {
				case ir.Declare:
					if carries(node.Value) {
						mark(node.Local)
					}
				case ir.Assign:
					if carries(node.Value) {
						mark(node.Local)
					}
				case ir.Return:
					if function >= 0 && !guard(function) && carries(node.Value) && !returns[function] {
						returns[function] = true
						changed = true
					}
				case ir.ObjectLiteral:
					for _, field := range node.Fields {
						if carries(field.Value) {
							slot(field.Name)
						}
					}
				case ir.SetProperty:
					if carries(node.Value) {
						slot(node.Name)
					}
				case ir.ArrayLiteral:
					for _, element := range node.Elements {
						if carries(element) && !arrays {
							arrays = true
							changed = true
						}
					}
				case ir.ArrayPush:
					if carries(node.Value) && !arrays {
						arrays = true
						changed = true
					}
				case ir.SetIndex:
					if carries(node.Value) && !arrays {
						arrays = true
						changed = true
					}
				case ir.MapSet:
					if (carries(node.Value) || carries(node.Key)) && !maps {
						maps = true
						changed = true
					}
				case ir.ArrayMap:
					if node.Result == ir.Object && !arrays {
						for target, absent := range returns {
							if absent && l.result.Functions[target].Returns == ir.Object {
								arrays = true
								changed = true
								break
							}
						}
					}
				case ir.MapKeys, ir.MapValues:
					if maps && !arrays {
						arrays = true
						changed = true
					}
				case ir.ForOf:
					if (arrays && node.MapPart == "") || (maps && node.MapPart != "") {
						if len(node.Pattern) == 0 && l.result.Locals[node.Local].Type.IsReference() {
							mark(node.Local)
						}
						for _, binding := range node.Pattern {
							if l.result.Locals[binding.Local].Type.IsReference() {
								mark(binding.Local)
							}
						}
					}
				case ir.Call:
					if guard(node.Function) {
						break
					}
					for _, target := range l.result.CallTargets(node) {
						for position, parameter := range l.result.Functions[target].Parameters {
							if node.Virtual != 0 && position == 0 {
								continue
							} // the method lookup checks before entry
							if position < len(node.Arguments) && carries(node.Arguments[position]) {
								mark(parameter)
							}
						}
					}
				case ir.CallClosure:
					for _, target := range l.result.Functions[:count] {
						if !target.Closure && target.MethodName == "" {
							continue
						}
						offset := 0
						if target.Receiver || target.MethodName != "" {
							offset = 1
						}
						for position, argument := range node.Arguments {
							if carries(argument) {
								at := position + offset
								if target.Rest != 0 && at >= len(target.Parameters)-1 {
									if !arrays {
										arrays = true
										changed = true
									}
								} else if at < len(target.Parameters) && l.result.Locals[target.Parameters[at]].Type == argument.Type() {
									mark(target.Parameters[at])
								}
							}
						}
					}
				case ir.Property:
					if node.Extracted && node.Bound != nil && carries(node.Bound) {
						for index, local := range l.result.Locals {
							if local.MethodValue && local.Function >= 0 && l.result.Functions[local.Function].MethodName == node.Name {
								mark(index)
							}
						}
					}
				}
				return true
			})
		}
		for index := 0; index < count; index++ {
			if !guard(index) {
				note(l.result.Functions[index].Body, index)
			}
		}
		note(l.result.Main, -1)
		// A tuple or object field holding an absent receiver can feed an array.
		if len(slots) != 0 && !arrays {
			arrays = true
			changed = true
		}
	}
	// A derived receiver remains absent even when its apparent type is object.
	// Keep unsupported consumers refused instead of letting native library panics
	// or erased union tags stand in for JavaScript's undefined behavior.
	var unsupported any
	validate := func(body []ir.Statement) {
		walk(body, func(node any) bool {
			_, expression := node.(ir.Expression)
			_, statement := node.(ir.Statement)
			if !expression && !statement {
				return true
			}
			allowed := false
			switch node.(type) {
			case ir.Read, ir.Declare, ir.Assign, ir.Return, ir.Evaluate, ir.Call,
				ir.CallClosure, ir.Property, ir.SetProperty, ir.ObjectLiteral,
				ir.ArrayLiteral, ir.ArrayPush, ir.SetIndex, ir.ArrayIndex, ir.ArrayPop, ir.Length, ir.ForOf, ir.MapSet, ir.MapGet, ir.MapKeys, ir.MapValues,
				ir.Conditional, ir.Coalesce, ir.Defined, ir.Narrow,
				ir.Binary, ir.Unary, ir.IsUndefined, ir.TypeOf, ir.InstanceOf:
				allowed = true
			}
			// Qualified private storage keys are not JavaScript's displayed
			// names. Aliased private receivers need their source-name/brand adapter.
			privateReceiver := false
			switch operation := node.(type) {
			case ir.Property:
				privateReceiver = strings.HasPrefix(operation.Name, "#") && carries(operation.Object)
			case ir.SetProperty:
				privateReceiver = strings.HasPrefix(operation.Name, "#") && carries(operation.Object)
			case ir.Call:
				privateReceiver = operation.Virtual != 0 && strings.HasPrefix(l.result.Functions[operation.Function].MethodName, "#") && len(operation.Arguments) != 0 && carries(operation.Arguments[0])
			}
			if privateReceiver {
				unsupported = node
			}
			// Numeric collections cannot transport an absent object receiver.
			switch operation := node.(type) {
			case ir.ArrayMap:
				if operation.Element != ir.Object && operation.Result != ir.Object {
					allowed = true
				}
			case ir.ArrayVisit:
				if operation.Element != ir.Object {
					allowed = true
				}
			case ir.ArrayReduce:
				if operation.Element != ir.Object && operation.Result != ir.Object {
					allowed = true
				}
				if operation.Result == ir.Object {
					for target, absent := range returns {
						if absent && l.result.Functions[target].Returns == ir.Object {
							unsupported = node
							break
						}
					}
				}
			case ir.ArraySort:
				if operation.Element != ir.Object {
					allowed = true
				}
			case ir.MapForEach:
				if operation.Key != ir.Object && operation.Value != ir.Object {
					allowed = true
				}
			}
			if object, ok := node.(ir.ObjectLiteral); ok && object.Spread != nil && carries(object.Spread) {
				allowed = false
			}
			if join, ok := node.(ir.ArrayJoin); ok && arrays && join.Element == ir.Object {
				unsupported = node
			}
			// Arrays, maps, strings, and callables can themselves escape as
			// undefined through optional this reads. Their storage and observations
			// are safe; consuming them needs a separate undefined adapter.
			safeAlias := false
			switch node.(type) {
			case ir.Read, ir.Declare, ir.Assign, ir.Return, ir.Evaluate, ir.Call,
				ir.ObjectLiteral, ir.ArrayLiteral, ir.ArrayPush, ir.SetIndex, ir.MapSet,
				ir.Conditional, ir.Coalesce, ir.Defined, ir.Narrow, ir.IsUndefined,
				ir.TypeOf, ir.InstanceOf, ir.CallClosure:
				safeAlias = true
			}
			if call, ok := node.(ir.CallClosure); ok && carries(call.Closure) {
				unsupported = node
			}
			if !safeAlias && receiverOperand(reflect.ValueOf(node), func(value ir.Expression) bool {
				return value.Type().IsReference() && value.Type() != ir.Object && carries(value)
			}) {
				unsupported = node
			}
			if !allowed && receiverOperand(reflect.ValueOf(node), func(value ir.Expression) bool {
				return (value.Type().IsReference() && carries(value)) || (value.Type() == ir.Array && arrays) || (value.Type() == ir.Map && maps)
			}) {
				unsupported = node
			}
			return true
		})
	}
	for index := 0; index < count; index++ {
		if !guard(index) {
			validate(l.result.Functions[index].Body)
		}
	}
	validate(l.result.Main)
	if unsupported != nil {
		return l.notYet(l.program.Files()[0].AsNode(), fmt.Sprintf("an extracted receiver consumed by %T without an undefined adapter", unsupported))
	}
	fix := func(node any) any {
		switch node := node.(type) {
		case ir.Property:
			if node.Object.Type() == ir.Object && carries(node.Object) && !node.Optional {
				node.Object = l.receiverRead(node.Object, node.Name)
			}
			return node
		case ir.Call:
			if node.Virtual != 0 && len(node.Arguments) != 0 && carries(node.Arguments[0]) {
				name := l.result.Functions[node.Function].MethodName
				if name == "" {
					name = node.Accessor
				}
				node.Arguments[0] = l.receiverRead(node.Arguments[0], name)
			}
			return node
		case ir.SetProperty:
			if carries(node.Object) {
				return ir.Evaluate{Value: l.receiverWrite(node)}
			}
		}
		return node
	}
	for index := 0; index < count; index++ {
		if guard(index) {
			continue
		}
		body := rewriteReceiverIR(reflect.ValueOf(l.result.Functions[index].Body), fix).Interface().([]ir.Statement)
		l.result.Functions[index].Body = body
	}
	l.result.Main = rewriteReceiverIR(reflect.ValueOf(l.result.Main), fix).Interface().([]ir.Statement)
	return nil
}

func (l *lowering) receiverRead(value ir.Expression, name string) ir.Expression {
	b := l.libraryArrayBuilder([]ir.Expression{value})
	self := b.read(b.parameters[0])
	b.body = []ir.Statement{l.receiverFailure(self, "Cannot read properties of undefined (reading '"+name+"')")}
	return b.finish("receiver_read", self)
}

func (l *lowering) receiverWrite(write ir.SetProperty) ir.Expression {
	// Both arguments arrive before the check, preserving RHS effects before SetValue.
	b := l.libraryArrayBuilder([]ir.Expression{write.Object, write.Value})
	self := b.read(b.parameters[0])
	write.Object, write.Value = self, b.read(b.parameters[1])
	b.body = []ir.Statement{l.receiverFailure(self, "Cannot set properties of undefined (setting '"+write.Name+"')"), write}
	return b.finish("receiver_write", ir.NumberConstant{})
}

func (l *lowering) receiverFailure(value ir.Expression, message string) ir.Statement {
	return ir.If{Condition: ir.IsUndefined{Value: value}, Then: []ir.Statement{ir.Throw{Value: ir.MakeError{Message: ir.StringConstant{Index: l.constant(message)}, Name: ir.StringConstant{Index: l.constant("TypeError")}}}}}
}

// Copy the tree before rewriting value interfaces; appending helper functions must not
// leave a pointer into the old function slice. Metadata and original write sites survive.
func rewriteReceiverIR(value reflect.Value, fix func(any) any) reflect.Value {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return value
		}
		result := reflect.New(value.Type()).Elem()
		result.Set(rewriteReceiverIR(value.Elem(), fix))
		return result
	case reflect.Slice:
		if value.IsNil() {
			return value
		}
		result := reflect.MakeSlice(value.Type(), value.Len(), value.Len())
		for index := 0; index < value.Len(); index++ {
			result.Index(index).Set(rewriteReceiverIR(value.Index(index), fix))
		}
		return result
	case reflect.Struct:
		result := reflect.New(value.Type()).Elem()
		for index := 0; index < value.NumField(); index++ {
			result.Field(index).Set(rewriteReceiverIR(value.Field(index), fix))
		}
		if result.Type().Implements(expressionType) || result.Type().Implements(statementType) {
			return reflect.ValueOf(fix(result.Interface()))
		}
		return result
	}
	return value
}

// Inspect operands without crossing another expression node: walk validates that
// node separately, so an array holding undefined is not itself an absent receiver.
func receiverOperand(value reflect.Value, carries func(ir.Expression) bool) bool {
	switch value.Kind() {
	case reflect.Interface:
		if value.IsNil() {
			return false
		}
		if expression, ok := value.Interface().(ir.Expression); ok {
			return carries(expression)
		}
		if _, statement := value.Interface().(ir.Statement); statement {
			return false
		}
		return receiverOperand(value.Elem(), carries)
	case reflect.Struct:
		for index := 0; index < value.NumField(); index++ {
			if receiverOperand(value.Field(index), carries) {
				return true
			}
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			if receiverOperand(value.Index(index), carries) {
				return true
			}
		}
	}
	return false
}
