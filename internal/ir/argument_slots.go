package ir

// RestArguments names a collected tail separately from fixed parameter slots.
// A fixed-arity function and a rest function can inhabit the same function type.
type RestArguments struct {
	Start   int
	Element Type
}

func FunctionRestArguments(function Function) RestArguments {
	start := len(function.Parameters) - 1
	if function.Receiver {
		start--
	}
	return RestArguments{Start: start, Element: function.RestElement}
}

// PrepareArgumentSlots assigns stable offsets. Merely preparing a layout emits
// no instructions; the caller selects only its function type's needed slots.
func (p *Program) PrepareArgumentSlots() {
	fixed := 0
	for _, function := range p.Functions {
		n := len(function.Parameters)
		if function.RestElement != 0 {
			n--
		}
		if function.Receiver {
			n--
		}
		fixed = max(fixed, n)
	}
	p.RestArgumentSlots = map[RestArguments]int{}
	for _, function := range p.Functions {
		if function.RestElement != 0 {
			key := FunctionRestArguments(function)
			if _, exists := p.RestArgumentSlots[key]; !exists {
				p.RestArgumentSlots[key] = fixed + len(p.RestArgumentSlots)
			}
		}
	}
	p.ArgumentCountSlot = fixed + len(p.RestArgumentSlots)
}

type ArgumentLayout struct {
	Fixed []Type
	Rest  []RestArguments
	Count bool
}

// ClosureArgumentLayout pads optional slots and collects rest tails at the
// caller. A hidden count is selected only by an actual reader candidate.
func (p *Program) ClosureArgumentLayout(call CallClosure) ArgumentLayout {
	targets, resolved := p.FunctionTypeTargets[call.FunctionType]
	if !resolved {
		bounded := p.ClosureTargets(call)
		targets = bounded.Functions
		if bounded.Unknown {
			for index := range p.Functions {
				targets = append(targets, index)
			}
		}
	}
	if call.Direct > 0 {
		targets = []int{call.Direct - 1}
	}
	layout := ArgumentLayout{}
	for _, target := range targets {
		function := p.Functions[target]
		layout.Count = layout.Count || p.PackedCountNeeded(target)
		// Ordinary named functions bind their direct arguments elsewhere. Their
		// value adapters, when the program makes one, are separate closure records.
		if !function.Closure && !function.Receiver {
			continue
		}
		if !function.Closure && function.Receiver {
			property, method := call.Closure.(Property)
			if !method || !property.Method || property.Name != function.MethodName {
				continue
			}
		}
		parameters := function.Parameters
		if function.Receiver {
			parameters = parameters[1:]
		}
		if function.RestElement != 0 {
			key := FunctionRestArguments(function)
			found := false
			for _, rest := range layout.Rest {
				found = found || rest == key
			}
			if !found {
				layout.Rest = append(layout.Rest, key)
			}
			parameters = parameters[:len(parameters)-1]
		}
		for index, parameter := range parameters {
			of := p.Locals[parameter].Type
			if index == len(layout.Fixed) {
				layout.Fixed = append(layout.Fixed, of)
			} else if of.IsMaybe() {
				layout.Fixed[index] = of
			}
		}
	}
	return layout
}

// PackedCountNeeded is a callee observation, not a property of all function values.
func (p *Program) PackedCountNeeded(function int) bool {
	f := p.Functions[function]
	if f.ReadsArguments || f.RestElement != 0 && (f.Closure || f.Receiver) {
		return true
	}

	// A zero-argument view can reach optional parameters held differently. One
	// padding word cannot be absent in both representations, so those callees
	// distinguish presence with the count instead of interpreting that word.
	for _, targets := range p.FunctionTypeTargets {
		contains := false
		for _, target := range targets {
			contains = contains || target == function
		}
		if !contains {
			continue
		}
		parameters := f.Parameters
		if f.Receiver {
			parameters = parameters[1:]
		}
		for _, target := range targets {
			other := p.Functions[target]
			if !other.Closure && !other.Receiver {
				continue
			}
			if !f.Closure || !other.Closure {
				if f.MethodName != other.MethodName || f.Closure != other.Closure {
					continue
				}
			}
			compared := other.Parameters
			if other.Receiver {
				compared = compared[1:]
			}
			for index, parameter := range parameters {
				if index >= len(compared) || f.RestElement != 0 && index == len(parameters)-1 || other.RestElement != 0 && index == len(compared)-1 {
					continue
				}
				of, against := p.Locals[parameter].Type, p.Locals[compared[index]].Type
				if (of.IsMaybe() || of.IsReference()) && absentRepresentation(of) != absentRepresentation(against) {
					return true
				}
			}
		}
	}
	return false
}

func absentRepresentation(of Type) int {
	if of == MaybeNumber {
		return 1
	}
	if of == MaybeBoolean {
		return 2
	}
	return 0
}

func (p *Program) ClosureConventionNeeded() bool {
	for index := range p.Functions {
		if p.PackedCountNeeded(index) {
			return true
		}
	}
	return false
}

func (p *Program) PackedCountNeededFromCall(call CallClosure) bool {
	return p.ClosureArgumentLayout(call).Count
}

func (p *Program) CanonicalClosuresNeeded() bool {
	for _, f := range p.Functions {
		if f.FrameIdentity != 0 || f.NestedFrame {
			return true
		}
	}
	return false
}

func (p *Program) ClosureReceiversNeeded() bool {
	for _, f := range p.Functions {
		if f.Closure && f.Receiver {
			return true
		}
	}
	return false
}
