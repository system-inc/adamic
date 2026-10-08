package ir

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
