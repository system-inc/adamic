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
// value whose producers are bounded is known. The same whole-program flow facts
// are used for cycle membership. Unsupported producers remain Unknown. ArrayVisit
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
	return p.ValueClosureTargets(value)
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

// ValueClosureTargets uses the call target analysis for a stored function value.
// An empty least fixed point is not evidence of a callable value's origin.
func (p *Program) ValueClosureTargets(value Expression) FunctionTargets {
	facts := p.closureFlow()
	targets := facts.value(value)
	if len(targets.Functions) == 0 {
		targets.Unknown = true
	}
	return targets
}

// StoredMapClosureTargets conservatively merges function values in all maps.
// Pooling maps can refuse a safe program, but cannot hide an indirect capture.
func (p *Program) StoredMapClosureTargets() FunctionTargets {
	return p.closureFlow().maps
}

// StoredClosureTargets answers the same producer graph for a collection or field.
// Slot names identify pooled fields, array elements, or Map keys and values.
func (p *Program) StoredClosureTargets(slot string) FunctionTargets {
	facts := p.closureFlow()
	if slot == "map" {
		return facts.maps
	}
	if len(slot) >= 6 && slot[:6] == "field:" && facts.foreignFields {
		return FunctionTargets{Unknown: true}
	}
	if len(slot) >= 6 && slot[:6] == "field:" {
		if _, known := facts.contents[slot]; !known {
			return FunctionTargets{Unknown: true}
		}
	}
	return facts.contents[slot]
}

type closureFacts struct {
	contents        map[string]FunctionTargets
	foreignFields   bool
	program         *Program
	locals, returns []FunctionTargets
	maps            FunctionTargets
	produced        []bool
}

func mergeClosureTargets(into *FunctionTargets, from FunctionTargets) bool {
	changed := false
	if from.Unknown && !into.Unknown {
		into.Unknown = true
		changed = true
	}
	for _, function := range from.Functions {
		if !slices.Contains(into.Functions, function) {
			into.Functions = append(into.Functions, function)
			changed = true
		}
	}
	return changed
}

func (f *closureFacts) value(expression Expression) FunctionTargets {
	switch value := expression.(type) {
	case nil, Undefined, NumberConstant, BooleanConstant, StringConstant:
		return FunctionTargets{}
	case MakeClosure:
		return FunctionTargets{Functions: []int{value.Function}}
	case Read:
		if value.Local < 0 || value.Local >= len(f.locals) {
			return FunctionTargets{Unknown: true}
		}
		if target := f.program.Locals[value.Local].ConstantClosure; target != 0 {
			return FunctionTargets{Functions: []int{target - 1}}
		}
		return f.locals[value.Local]
	case Defined:
		return f.value(value.Value)
	case Unwrap:
		return f.value(value.Value)
	case MaybeOf:
		return f.value(value.Value)
	case Box:
		return f.value(value.Value)
	case Conditional:
		result := FunctionTargets{}
		mergeClosureTargets(&result, f.value(value.WhenTrue))
		mergeClosureTargets(&result, f.value(value.WhenNot))
		return result
	case Coalesce:
		result := FunctionTargets{}
		mergeClosureTargets(&result, f.value(value.Value))
		mergeClosureTargets(&result, f.value(value.Fallback))
		return result
	case Call:
		return f.results(f.program.CallTargets(value))
	case CallClosure:
		targets := f.value(value.Closure)
		if value.Direct > 0 {
			targets = FunctionTargets{Functions: []int{value.Direct - 1}}
		}
		result := f.results(targets.Functions)
		result.Unknown = result.Unknown || targets.Unknown
		return result
	case Property:
		if value.Method {
			return FunctionTargets{Unknown: true}
		}
		if f.foreignFields {
			return FunctionTargets{Unknown: true}
		}
		return f.contents["field:"+value.Name]
	case ArrayIndex, ArrayPop:
		return f.contents["array"]
	case Narrow:
		return f.value(value.Value)
	case MapGet:
		return f.maps
	}
	return FunctionTargets{Unknown: true}
}

func (f *closureFacts) results(targets []int) FunctionTargets {
	result := FunctionTargets{}
	for _, target := range targets {
		if target < 0 || target >= len(f.returns) {
			result.Unknown = true
			continue
		}
		mergeClosureTargets(&result, f.returns[target])
	}
	return result
}

func walkClosureFlow(value reflect.Value, visit func(any)) {
	switch value.Kind() {
	case reflect.Interface:
		if !value.IsNil() {
			walkClosureFlow(value.Elem(), visit)
		}
	case reflect.Slice, reflect.Array:
		for index := 0; index < value.Len(); index++ {
			walkClosureFlow(value.Index(index), visit)
		}
	case reflect.Struct:
		if value.CanInterface() {
			visit(value.Interface())
		}
		for index := 0; index < value.NumField(); index++ {
			walkClosureFlow(value.Field(index), visit)
		}
	}
}

// closureFlow solves a monotone, context-insensitive target graph. No type
// compatibility inference supplies targets. Calls, returns and writes contribute
// actual producers; unbound inputs and unsupported values poison their consumers.
func (p *Program) closureFlow() *closureFacts {
	if p.closureTargets != nil {
		return p.closureTargets
	}
	f := &closureFacts{contents: map[string]FunctionTargets{}, program: p, locals: make([]FunctionTargets, len(p.Locals)), returns: make([]FunctionTargets, len(p.Functions)), produced: make([]bool, len(p.Locals))}
	for local, declared := range p.Locals {
		if declared.ConstantClosure != 0 {
			f.locals[local] = FunctionTargets{Functions: []int{declared.ConstantClosure - 1}}
			f.produced[local] = true
		}
	}
	changed := true
	store := func(slot string, value Expression) {
		targets := f.contents[slot]
		if mergeClosureTargets(&targets, f.value(value)) {
			changed = true
		}
		f.contents[slot] = targets
	}
	bind := func(targets FunctionTargets, arguments []Expression, expanded bool) {
		for _, target := range targets.Functions {
			if target < 0 || target >= len(p.Functions) {
				continue
			}
			for position, local := range p.Functions[target].Parameters {
				f.produced[local] = true
				input := FunctionTargets{Unknown: true}
				if !expanded && position < len(arguments) {
					input = f.value(arguments[position])
				}
				if mergeClosureTargets(&f.locals[local], input) {
					changed = true
				}
			}
		}
	}
	bindInputs := func(targets FunctionTargets, inputs []FunctionTargets) {
		for _, target := range targets.Functions {
			if target < 0 || target >= len(p.Functions) {
				continue
			}
			for position, local := range p.Functions[target].Parameters {
				f.produced[local] = true
				input := FunctionTargets{Unknown: true}
				if position < len(inputs) {
					input = inputs[position]
				}
				if mergeClosureTargets(&f.locals[local], input) {
					changed = true
				}
			}
		}
	}
	solve := func() {
		for changed {
			changed = false
			inspect := func(owner int) func(any) {
				return func(node any) {
					switch value := node.(type) {
					case Declare:
						if !value.Uninitialized {
							f.produced[value.Local] = true
							if mergeClosureTargets(&f.locals[value.Local], f.value(value.Value)) {
								changed = true
							}
						}
					case Assign:
						f.produced[value.Local] = true
						if mergeClosureTargets(&f.locals[value.Local], f.value(value.Value)) {
							changed = true
						}
					case Return:
						if owner >= 0 && mergeClosureTargets(&f.returns[owner], f.value(value.Value)) {
							changed = true
						}
					case Call:
						bind(FunctionTargets{Functions: p.CallTargets(value)}, value.Arguments, p.CallExpandsArguments(value))
					case CallClosure:
						targets := f.value(value.Closure)
						if value.Direct > 0 {
							targets = FunctionTargets{Functions: []int{value.Direct - 1}}
						}
						bind(targets, value.Arguments, len(value.Spread) > 0)
					case ForOf:
						if value.Local >= 0 && value.Local < len(f.locals) && len(value.Pattern) == 0 {
							f.produced[value.Local] = true
							input := FunctionTargets{Unknown: true}
							if value.Iterable != nil && value.Iterable.Type() == Array {
								input = f.contents["array"]
							}
							if value.MapPart == "keys" {
								input = f.contents["map-key"]
							}
							if value.MapPart == "values" {
								input = f.maps
							}
							if value.Element != Closure && value.Element != Union {
								input = FunctionTargets{}
							}
							if mergeClosureTargets(&f.locals[value.Local], input) {
								changed = true
							}
						}
						for _, binding := range value.Pattern {
							f.produced[binding.Local] = true
							input := FunctionTargets{Unknown: true}
							if value.MapPart == "entries" {
								if binding.Field == "0" {
									input = f.contents["map-key"]
								}
								if binding.Field == "1" {
									input = f.maps
								}
							}
							if mergeClosureTargets(&f.locals[binding.Local], input) {
								changed = true
							}
						}
					case MapForEach:
						targets := f.value(value.Callback)
						items := f.maps
						keys := f.contents["map-key"]
						if value.Set {
							items = f.contents["set"]
							keys = items
						}
						for _, target := range targets.Functions {
							if target >= 0 && target < len(p.Functions) {
								for position, local := range p.Functions[target].Parameters {
									f.produced[local] = true
									input := FunctionTargets{Unknown: true}
									if position == 0 {
										input = items
									}
									if position == 1 {
										input = keys
									}
									if position == 2 {
										input = FunctionTargets{}
									}
									if mergeClosureTargets(&f.locals[local], input) {
										changed = true
									}
								}
							}
						}
					case ArrayVisit:
						targets := f.value(value.Callback)
						for _, target := range targets.Functions {
							if target >= 0 && target < len(p.Functions) {
								for position, local := range p.Functions[target].Parameters {
									f.produced[local] = true
									input := FunctionTargets{Unknown: true}
									if position == 0 {
										input = f.contents["array"]
									}
									if position == 1 || position == 2 {
										input = FunctionTargets{}
									}
									if mergeClosureTargets(&f.locals[local], input) {
										changed = true
									}
								}
							}
						}
					case ObjectLiteral:
						for _, field := range value.Fields {
							if field.Value != nil && (field.Value.Type() == Closure || field.Value.Type() == Union) {
								store("field:"+field.Name, field.Value)
							}
						}
					case SetProperty:
						if value.Value != nil && (value.Value.Type() == Closure || value.Value.Type() == Union) {
							store("field:"+value.Name, value.Value)
						}
					case ArrayLiteral:
						if value.Element == Closure || value.Element == Union {
							for index, element := range value.Elements {
								if index < len(value.Spread) && value.Spread[index] {
									continue
								}
								store("array", element)
							}
						}
					case ArrayPush:
						if value.Element == Closure || value.Element == Union {
							store("array", value.Value)
						}
					case SetIndex:
						if value.Element == Closure || value.Element == Union {
							store("array", value.Value)
						}
					case ArrayFill:
						if value.Element == Closure || value.Element == Union {
							store("array", value.Value)
						}
					case ArraySplice:
						if value.Element == Closure || value.Element == Union {
							for _, item := range value.Items {
								store("array", item)
							}
						}
					case ArrayMap:
						bindInputs(f.value(value.Callback), []FunctionTargets{f.contents["array"], {}, {}})
						if value.Result == Closure || value.Result == Union {
							store("array", CallClosure{Closure: value.Callback, Returns: value.Result})
						}
					case ArrayFrom:
						bindInputs(f.value(value.Callback), []FunctionTargets{{}, {}})
						if value.Element == Closure || value.Element == Union {
							store("array", CallClosure{Closure: value.Callback, Returns: value.Element})
						}
					case ArrayReduce:
						accumulator := FunctionTargets{}
						mergeClosureTargets(&accumulator, f.value(value.Initial))
						mergeClosureTargets(&accumulator, f.results(f.value(value.Callback).Functions))
						bindInputs(f.value(value.Callback), []FunctionTargets{accumulator, f.contents["array"], {}, {}})
					case ArraySort:
						targets := f.value(value.Callback)
						if value.Callback == nil {
							targets = FunctionTargets{Functions: []int{value.Comparator}}
						}
						bindInputs(targets, []FunctionTargets{f.contents["array"], f.contents["array"]})
					case SetAdd:
						if value.Element == Closure || value.Element == Union {
							store("set", value.Value)
						}
					case SetNew:
						targets := f.contents["set"]
						if mergeClosureTargets(&targets, f.contents["array"]) {
							changed = true
						}
						f.contents["set"] = targets
					case SetValues:
						if value.Element == Closure || value.Element == Union {
							targets := f.contents["array"]
							if mergeClosureTargets(&targets, f.contents["set"]) {
								changed = true
							}
							f.contents["array"] = targets
						}
					case MapKeys:
						if value.Key == Closure || value.Key == Union {
							targets := f.contents["array"]
							if mergeClosureTargets(&targets, f.contents["map-key"]) {
								changed = true
							}
							f.contents["array"] = targets
						}
					case MapValues:
						if value.Value == Closure || value.Value == Union {
							targets := f.contents["array"]
							if mergeClosureTargets(&targets, f.maps) {
								changed = true
							}
							f.contents["array"] = targets
						}
					case MapSet:
						if value.KeyType == Closure || value.KeyType == Union {
							store("map-key", value.Key)
						}
						if (value.ValueType == Closure || value.ValueType == Union) && mergeClosureTargets(&f.maps, f.value(value.Value)) {
							changed = true
						}
					case MapNew:
						if value.Key == Closure || value.Key == Union {
							for _, entry := range value.Entries {
								store("map-key", entry[0])
							}
						}
						if value.Value == Closure || value.Value == Union {
							for _, entry := range value.Entries {
								if mergeClosureTargets(&f.maps, f.value(entry[1])) {
									changed = true
								}
							}
							if value.Pairs != nil && mergeClosureTargets(&f.maps, FunctionTargets{Unknown: true}) {
								changed = true
							}
						}
					}
				}
			}
			walkClosureFlow(reflect.ValueOf(p.Main), inspect(-1))
			for owner, function := range p.Functions {
				walkClosureFlow(reflect.ValueOf(function.Body), inspect(owner))
			}
		}
	}
	solve()
	// A local without any producer includes foreign inputs and callback parameters
	// invoked by operations this solver does not model. They are never empty proofs.
	for local := range f.locals {
		if !f.produced[local] {
			f.locals[local].Unknown = true
			changed = true
			if p.Locals[local].Type == Map {
				f.maps.Unknown = true
			}
		}
	}
	poisonMaps := func(node any) {
		expression, ok := node.(Expression)
		if !ok || expression.Type() != Map {
			return
		}
		switch expression.(type) {
		case MapNew, MapSet, SetNew, SetAdd, Read, Call, CallClosure, Defined, Coalesce, Conditional, Narrow, Box, WeakTarget, Property, ArrayIndex, Unwrap, Undefined:
		default:
			f.maps.Unknown = true
		}
	}
	walkClosureFlow(reflect.ValueOf(p.Main), poisonMaps)
	for _, function := range p.Functions {
		walkClosureFlow(reflect.ValueOf(function.Body), poisonMaps)
	}

	poisonContainers := func(node any) {
		expression, ok := node.(Expression)
		if !ok {
			return
		}
		if expression.Type() == Object {
			switch expression.(type) {
			case ObjectLiteral, Call, CallClosure, Read, Defined, Coalesce, Conditional, Property, ArrayIndex, MapGet, MaybeOf, Narrow, Box, Undefined, WeakTarget, Unwrap, CheckedCast, MakeError, RegExpNew, CollectionIterator, JSONNull:
			default:
				f.foreignFields = true
				f.maps.Unknown = true
				targets := f.contents["array"]
				targets.Unknown = true
				f.contents["array"] = targets
			}
		}
		if expression.Type() == Array {
			switch expression.(type) {
			case ArrayLiteral, ArrayPush, ArrayFill, ArraySplice, ArraySlice, ArrayConcat, ArrayReverse, ArraySort, ArrayMap, ArrayFrom, ArrayVisit, MapKeys, MapValues, MapEntries, SetValues, Read, Call, CallClosure, Defined, Coalesce, Conditional, ProgramArguments, CodePoints, StringCall, ObjectKeys, Undefined, Unwrap, Narrow, Box, WeakTarget, ArrayIndex, Property:
			default:
				targets := f.contents["array"]
				targets.Unknown = true
				f.contents["array"] = targets
			}
		}
	}
	walkClosureFlow(reflect.ValueOf(p.Main), poisonContainers)
	for _, function := range p.Functions {
		walkClosureFlow(reflect.ValueOf(function.Body), poisonContainers)
	}
	for local := range f.locals {
		if !f.produced[local] && p.Locals[local].Type == Array {
			targets := f.contents["array"]
			targets.Unknown = true
			f.contents["array"] = targets
		}
	}
	changed = true
	solve()
	// An unknown invocation or a foreign operation can call any function value
	// that escaped, including one which also has a known caller. Poison those
	// inputs rather than letting the known call conceal the foreign boundary.
	unknownCaller := false
	inspectBoundary := func(node any) {
		if call, ok := node.(CallClosure); ok && call.Direct == 0 {
			targets := f.value(call.Closure)
			if targets.Unknown || len(targets.Functions) == 0 {
				unknownCaller = true
			}
		}
		if expression, ok := node.(Expression); ok {
			name := reflect.TypeOf(expression).Name()
			if len(name) >= 4 && name[:4] == "Node" {
				unknownCaller = true
			}
		}
	}
	walkClosureFlow(reflect.ValueOf(p.Main), inspectBoundary)
	for _, function := range p.Functions {
		walkClosureFlow(reflect.ValueOf(function.Body), inspectBoundary)
	}
	if unknownCaller {
		poison := func(node any) {
			if closure, ok := node.(MakeClosure); ok && closure.Function >= 0 && closure.Function < len(p.Functions) {
				for _, local := range p.Functions[closure.Function].Parameters {
					f.locals[local].Unknown = true
					if p.Locals[local].Type == Map {
						f.maps.Unknown = true
					}
					if p.Locals[local].Type == Array {
						targets := f.contents["array"]
						targets.Unknown = true
						f.contents["array"] = targets
					}
				}
			}
		}
		walkClosureFlow(reflect.ValueOf(p.Main), poison)
		for _, function := range p.Functions {
			walkClosureFlow(reflect.ValueOf(function.Body), poison)
		}
		changed = true
		solve()
	}
	p.closureTargets = f
	return f
}
