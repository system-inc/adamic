package ir

// RestArguments names a collected tail separately from fixed parameter slots.
// A fixed-arity function and a rest function can inhabit the same function type.
type RestArguments struct {
	Start   int
	Element Type
}

func FunctionRestArguments(function Function) RestArguments {
	start := len(function.Parameters) - 1
	if function.Receiver && !function.Closure {
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
		if function.Receiver && !function.Closure {
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
	layout := ArgumentLayout{}
	for _, target := range targets {
		function := p.Functions[target]
		layout.Count = layout.Count || function.ReadsArguments
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
		packable := function.Returns != MaybeBoolean && function.Returns != Union
		for _, parameter := range function.Parameters {
			of := p.Locals[parameter].Type
			packable = packable && of != MaybeBoolean && of != Union
		}
		if !packable {
			continue
		}
		parameters := function.Parameters
		if function.Receiver && !function.Closure {
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
