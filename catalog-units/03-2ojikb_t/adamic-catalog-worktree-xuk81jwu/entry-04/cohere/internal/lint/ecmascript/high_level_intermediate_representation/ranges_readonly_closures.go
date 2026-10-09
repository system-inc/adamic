package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

func hasReadOnlyClosureEffectsForCaptures(function *Function, captures []Place, kinds map[static_single_assignment.IdentifierId]mutation_aliasing.EffectValueKind) bool {
	if hasReadOnlyClosureEffects(function, map[*Function]bool{}) {
		return true
	}
	if function == nil || len(function.Context) != len(captures) {
		return false
	}
	contextKinds := map[static_single_assignment.IdentifierId]mutation_aliasing.EffectValueKind{}
	hasMixedCapture := false
	for index, capture := range captures {
		kind, known := kinds[capture.Identifier]
		if !known || kind == mutation_aliasing.EffectValueMutable || kind == mutation_aliasing.EffectValueGlobal {
			return false
		}
		hasMixedCapture = hasMixedCapture || kind == mutation_aliasing.EffectValueMaybeFrozen
		contextKinds[function.Context[index].Identifier] = kind
	}
	if !hasMixedCapture {
		return false
	}
	effects := InferAliasingEffects(function)
	calleeProducers := newCalleeProducers(function)
	producers := map[static_single_assignment.IdentifierId]InstructionValue{}
	for _, instruction := range function.Instructions {
		if instruction != nil {
			producers[instruction.LValue.Identifier] = instruction.Value
		}
	}
	for _, instruction := range function.Instructions {
		if instruction == nil {
			continue
		}
		switch value := instruction.Value.(type) {
		case *Await, *StoreGlobal, *StoreContext, *ObjectMethod, *NewExpression, *FunctionExpression:
			return false
		case *LoadGlobal:
			if value.Name == "Date" {
				return false
			}
		case *MethodCall:
			if _, known := lookupSignature(function, calleeProducers, instruction, calleeName(function, instruction, value.Property)); !known {
				return false
			}
		case *CallExpression:
			if _, known := lookupSignature(function, calleeProducers, instruction, calleeName(function, instruction, value.Callee)); !known {
				if _, global := producers[value.Callee.Identifier].(*LoadGlobal); !global {
					return false
				}
			}
		}
		for _, effect := range effects.Get(instruction.Id) {
			if effect.Kind == mutation_aliasing.AliasingEffectFreeze {
				return false
			}
		}
	}
	graph := mutation_aliasing.BuildAliasingGraph(newRangeGraph(function, effects), function,
		mutation_aliasing.Options{ContextKinds: contextKinds})
	return graph.MutationCount() == 0
}

func hasReadOnlyClosureEffects(function *Function, seen map[*Function]bool) bool {
	if function == nil || seen[function] {
		return false
	}
	seen[function] = true
	defer delete(seen, function)
	effects := InferAliasingEffects(function)
	for _, instruction := range function.Instructions {
		if instruction == nil {
			continue
		}
		switch value := instruction.Value.(type) {
		case *Await, *StoreGlobal, *StoreContext, *ObjectMethod:
			return false
		case *FunctionExpression:
			if int(value.Function) >= len(function.Functions) || !hasReadOnlyClosureEffects(function.Functions[value.Function], seen) {
				return false
			}
		case *LoadGlobal:
			if nested := nestedFunctionHeldBy(function, instruction.LValue.Identifier); nested != nil && !hasReadOnlyClosureEffects(nested, seen) {
				return false
			}
		}
		for _, effect := range effects.Get(instruction.Id) {
			if effect.Kind.IsMutation() || effect.Kind == mutation_aliasing.AliasingEffectFreeze {
				return false
			}
		}
	}
	return true
}

func refDerivedValues(function *Function) map[static_single_assignment.IdentifierId]bool {
	refs := useRefResultValues(function)
	for {
		changed := false
		mark := func(from, into static_single_assignment.IdentifierId) {
			if refs[from] && !refs[into] {
				refs[into] = true
				changed = true
			}
		}
		for _, block := range function.Blocks {
			if block == nil {
				continue
			}
			for _, phi := range block.Phis {
				for _, entry := range phi.Operands {
					mark(entry.Place.Identifier, phi.Place.Identifier)
				}
			}
			for _, instructionId := range block.Instructions {
				instruction := function.Instructions[instructionId]
				if instruction == nil {
					continue
				}
				switch value := instruction.Value.(type) {
				case *LoadLocal:
					mark(value.Place.Identifier, instruction.LValue.Identifier)
				case *StoreLocal:
					mark(value.Value.Identifier, value.LValue.Identifier)
					mark(value.Value.Identifier, instruction.LValue.Identifier)
				case *PropertyLoad:
					mark(value.Object.Identifier, instruction.LValue.Identifier)
				case *ComputedLoad:
					mark(value.Object.Identifier, instruction.LValue.Identifier)
				}
			}
		}
		if !changed {
			return refs
		}
	}
}
