package ir

import (
	"reflect"
	"slices"
)

// CallTargets answers which functions a call can run. A direct call has exactly
// its Function as target. A virtual call (Virtual != 0) has every implementation
// that can run, including overrides in every instantiated subclass, recorded by
// lowering in MethodTargets after all classes are known. This also includes
// structural accessor dispatch (Virtual == -1). A missing virtual target set is
// malformed IR and must fail loudly rather than imply the call has no effects.
func (p *Program) CallTargets(call Call) []int {
	if call.Virtual != 0 {
		targets := p.MethodTargets[call.Function]
		if len(targets) == 0 {
			panic("ir: virtual call has no target set")
		}
		return targets
	}
	return []int{call.Function}
}

func (p *Program) CallMayThrow(call Call) bool {
	for _, target := range p.CallTargets(call) {
		if p.Functions[target].MayThrow {
			return true
		}
	}
	return false
}

// FunctionTargets is a proven set of possible functions. Unknown means the value
// cannot be bounded, and analyses must treat it as anything, never as an empty set.
type FunctionTargets struct {
	Functions []int
	Unknown   bool
}

// ClosureTargets answers which functions a function-value call can run. Only a
// closure literal or a const initialized directly with one is bounded. Parameters,
// mutable bindings, properties, returned values and joins are Unknown. ArrayVisit
// covers forEach, filter, find, some and every; MapForEach covers Map and Set.
// A sort's named comparator is a direct function, rather than a function value.
func (p *Program) ClosureTargets(call Expression) FunctionTargets {
	var value Expression
	switch call := call.(type) {
	case CallClosure:
		value = call.Closure
	case ArrayMap:
		value = call.Callback
	case ArrayVisit:
		value = call.Callback
	case ArrayReduce:
		value = call.Callback
	case ArrayFrom:
		value = call.Callback
	case MapForEach:
		value = call.Callback
	case ArraySort:
		if call.Callback == nil {
			return FunctionTargets{Functions: []int{call.Comparator}}
		}
		value = call.Callback
	default:
		panic("ir: ClosureTargets requires a function-value call")
	}
	switch value := value.(type) {
	case MakeClosure:
		return FunctionTargets{Functions: []int{value.Function}}
	case Property:
		if value.Method {
			return p.structuralTargets(value.Name)
		}
	case Read:
		if target := p.Locals[value.Local].ConstantClosure; target != 0 {
			return FunctionTargets{Functions: []int{target - 1}}
		}
	}
	return FunctionTargets{Unknown: true}
}

// ClosureMayThrow treats Unknown as anything in the program, including methods
// callable through structural interfaces, rather than trusting a static signature.
func (p *Program) ClosureMayThrow(call Expression) bool {
	targets := p.ClosureTargets(call)
	if targets.Unknown {
		for _, function := range p.Functions {
			if function.MayThrow {
				return true
			}
		}
		return false
	}
	for _, target := range targets.Functions {
		if p.Functions[target].MayThrow {
			return true
		}
	}
	return false
}

// CallBorrows is the convention at one argument position, receiver included.
// Every possible implementation must borrow. Missing parameters are owned.
func (p *Program) CallBorrows(call Call, position int) bool {
	return p.targetsBorrow(FunctionTargets{Functions: p.CallTargets(call)}, position)
}

// ClosureBorrows uses the same join for a bounded function value. Unknown is owned.
func (p *Program) ClosureBorrows(call Expression, position int) bool {
	targets := p.ClosureTargets(call)
	if closure, ok := call.(CallClosure); ok {
		if property, ok := closure.Closure.(Property); ok && property.Method {
			if targets.Unknown || len(targets.Functions) == 0 {
				return false
			}
			for _, target := range targets.Functions {
				offset := 0
				if !p.Functions[target].Closure {
					offset = 1
				}
				if !p.targetsBorrow(FunctionTargets{Functions: []int{target}}, position+offset) {
					return false
				}
			}
			return true
		}
	}
	return p.targetsBorrow(targets, position)
}

func (p *Program) targetsBorrow(targets FunctionTargets, position int) bool {
	if targets.Unknown || len(targets.Functions) == 0 || position < 0 {
		return false
	}
	for _, target := range targets.Functions {
		parameters := p.Functions[target].Parameters
		if position >= len(parameters) || !p.Locals[parameters[position]].Borrowed {
			return false
		}
	}
	return true
}

// ClosureReceiverBorrows joins the implicit receivers of structural methods.
// Function-valued fields do not receive this, but an unknown field stays owned.
func (p *Program) ClosureReceiverBorrows(call CallClosure) bool {
	targets := p.ClosureTargets(call)
	if targets.Unknown || len(targets.Functions) == 0 {
		return false
	}
	for _, target := range targets.Functions {
		if !p.Functions[target].Closure && !p.targetsBorrow(FunctionTargets{Functions: []int{target}}, 0) {
			return false
		}
	}
	return true
}

// structuralTargets is deliberately a whole-program superset. Every construction
// and field write with this name contributes, regardless of the receiver's type.
// An unbounded function-valued field or accessor makes the answer Unknown.
func (p *Program) structuralTargets(name string) FunctionTargets {
	result := FunctionTargets{}
	add := func(function int) {
		if !slices.Contains(result.Functions, function) {
			result.Functions = append(result.Functions, function)
		}
	}
	field := func(value Expression) {
		switch value := value.(type) {
		case MakeClosure:
			add(value.Function)
		case Read:
			if target := p.Locals[value.Local].ConstantClosure; target != 0 {
				add(target - 1)
			} else {
				result.Unknown = true
			}
		default:
			result.Unknown = true
		}
	}
	var visit func(reflect.Value)
	visit = func(value reflect.Value) {
		if !value.IsValid() {
			return
		}
		if value.CanInterface() {
			switch node := value.Interface().(type) {
			case ObjectCall:
				// Dynamic key construction and copying can introduce a function-valued
				// field without a named Field or SetProperty node.
				if node.Method == "assign" || node.Method == "fromEntries" {
					result.Unknown = true
				}
			case ObjectLiteral:
				for _, method := range node.Methods {
					if method.Name == name {
						add(method.Function)
					}
				}
				for _, entry := range node.Fields {
					if entry.Name == name {
						field(entry.Value)
					}
				}
			case SetProperty:
				if node.Name == name {
					field(node.Value)
				}
			}
		}
		switch value.Kind() {
		case reflect.Interface, reflect.Pointer:
			if !value.IsNil() {
				visit(value.Elem())
			}
		case reflect.Struct:
			for index := 0; index < value.NumField(); index++ {
				visit(value.Field(index))
			}
		case reflect.Slice:
			for index := 0; index < value.Len(); index++ {
				visit(value.Index(index))
			}
		}
	}
	for _, function := range p.Functions {
		visit(reflect.ValueOf(function.Body))
	}
	visit(reflect.ValueOf(p.Main))
	for _, class := range p.Classes {
		for _, accessor := range class.Accessors {
			if accessor.Name == name {
				result.Unknown = true
			}
		}
	}
	if len(result.Functions) == 0 {
		result.Unknown = true
	}
	return result
}

// ClosureReceiver exposes the implicit receiver without consumers reinterpreting
// the function-value target representation. Other calls have no such receiver.
func (p *Program) ClosureReceiver(call CallClosure) Expression {
	if property, ok := call.Closure.(Property); ok && property.Method {
		return property.Object
	}
	return nil
}
