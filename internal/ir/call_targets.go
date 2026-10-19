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
		if call.Direct > 0 {
			return FunctionTargets{Functions: []int{call.Direct - 1}}
		}
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
	// Effects preserve the identity of their result. In particular an intrinsic
	// adapter binds its callback before returning a known closure literal.
	for {
		effects, ok := value.(Effects)
		if !ok {
			break
		}
		value = effects.Result
	}
	switch value := value.(type) {
	case MakeClosure:
		return FunctionTargets{Functions: []int{value.Function}}
	case Read:
		if target := p.Locals[value.Local].ConstantClosure; target != 0 {
			return FunctionTargets{Functions: []int{target - 1}}
		}
	}
	return FunctionTargets{Unknown: true}
}

// Discard is what a void call does with the result it gives back, which it releases and never
// reads: a () => void slot may hold () => make(), whose result comes back owned, or () => 42, whose
// word is no reference. Every type's adamic_value member is fixed by the type alone, so a function's
// Returns says which.
type Discard int

const (
	// DiscardNothing is a result with no reference to let go: every target returns void, a number,
	// a boolean or a maybe of one.
	DiscardNothing Discard = iota
	// DiscardReference is a counted reference every target returns, owned, released at the site.
	DiscardReference
	// DiscardFlagged is a call whose targets are Unknown, or bounded and disagree: the closure called
	// says, in the discard field it's made with (ADAMIC_CLOSURE_DISCARDS).
	DiscardFlagged
	// DiscardMixed is bounded targets that disagree. A closure's call reads its flag instead; a
	// class's method has none, so lowering says not yet to a method arm that disagrees.
	DiscardMixed
)

// ReturnsDiscard is what a void call that can run any of functions releases of its result.
func (p *Program) ReturnsDiscard(functions []int) Discard {
	discard := DiscardNothing
	for index, function := range functions {
		returned := DiscardNothing
		if p.Functions[function].Returns.IsReference() {
			returned = DiscardReference
		}
		if index > 0 && returned != discard {
			return DiscardMixed
		}
		discard = returned
	}
	return discard
}

// ClosureDiscard is what a void call through a function value (a CallClosure, or forEach's callback)
// releases of the closure's result: by its targets' Returns where ClosureTargets bounds them and they
// agree, and by the closure's own flag where they're Unknown or disagree, since every closure carries
// its own answer.
func (p *Program) ClosureDiscard(call Expression) Discard {
	targets := p.ClosureTargets(call)
	if targets.Unknown {
		return DiscardFlagged
	}
	if discard := p.ReturnsDiscard(targets.Functions); discard != DiscardMixed {
		return discard
	}
	return DiscardFlagged
}

// MethodDiscard is what a void call through an interface (Property's Method) releases of a result
// its object's class's method gave, the arm a call takes when the object has no field of the name.
// MethodTargets is keyed by a class's method slot, which a call through an interface doesn't have,
// so the methods it can find are bounded by name, by every method table that holds one
// (MethodsNamed). Any other call has no such arm.
func (p *Program) MethodDiscard(call CallClosure) Discard {
	property, isProperty := call.Closure.(Property)
	if !isProperty || !property.Method {
		return DiscardNothing
	}
	return p.ReturnsDiscard(p.MethodsNamed(property.Name))
}

// methodTables caches MethodsNamed's walk, keyed to the program's size as argumentFacts is.
type methodTables struct {
	functions int
	named     map[string][]int
}

// MethodsNamed is every function a method table holds under the name, in function order: the
// methods of the objects class constructors make (ObjectLiteral's Methods), a class's static side's
// among them, which is all a call through an interface finds a method in.
func (p *Program) MethodsNamed(name string) []int {
	if tables := p.methodTables; tables != nil && tables.functions == len(p.Functions) {
		return tables.named[name]
	}
	tables := &methodTables{functions: len(p.Functions), named: map[string][]int{}}
	held := map[Method]bool{}
	visit := func(node any) {
		literal, isLiteral := node.(ObjectLiteral)
		if !isLiteral {
			return
		}
		for _, method := range literal.Methods {
			if !held[method] {
				held[method] = true
				tables.named[method.Name] = append(tables.named[method.Name], method.Function)
			}
		}
	}
	spreadStorageWalk(reflect.ValueOf(p.Main), visit)
	for _, function := range p.Functions {
		spreadStorageWalk(reflect.ValueOf(function.Body), visit)
	}
	for _, methods := range tables.named {
		slices.Sort(methods)
	}
	p.methodTables = tables
	return tables.named[name]
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

// ClosureReadsArgumentsCount includes every function the program can make of the
// call's function type. Missing type evidence uses bounded closure targets,
// falling back to the entire program only when those targets are unknown.
func (p *Program) ClosureReadsArgumentsCount(call CallClosure) bool {
	return p.ClosureArgumentLayout(call).Count
}

// CallExpandsArguments means source positions do not match parameter positions.
// Analyses must not replay a positional summary across this adapter boundary.
func (p *Program) CallExpandsArguments(call Call) bool {
	if len(call.Spread) != 0 {
		return true
	}
	if !call.RestPacked {
		for _, target := range p.CallTargets(call) {
			if p.Functions[target].RestElement != 0 {
				return true
			}
		}
	}
	return false
}
