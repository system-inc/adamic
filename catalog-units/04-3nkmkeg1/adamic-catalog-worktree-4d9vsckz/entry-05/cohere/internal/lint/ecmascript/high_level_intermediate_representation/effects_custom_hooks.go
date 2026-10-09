package high_level_intermediate_representation

import (
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

func customHookSignature(function *Function) effectSignature {
	if function != nil && function.Node != nil {
		if sourceFile := ast.GetSourceFileOfNode(function.Node); sourceFile != nil {
			firstLine, _, _ := strings.Cut(sourceFile.Text(), "\n")
			if strings.Contains(firstLine, "@enableAssumeHooksFollowRulesOfReact:false") {
				return effectSignature{Receiver: EffectRead, Rest: EffectConditionallyMutate, HasRest: true, Result: mutation_aliasing.EffectValueMutable}
			}
		}
	}
	return effectSignature{Receiver: EffectRead, Rest: EffectFreeze, HasRest: true, Result: mutation_aliasing.EffectValueFrozen}
}

// calleeProducers is the table isModuleHookCall resolves a callee through: each value's producing
// instruction value, plus each StoreLocal under the binding it writes.
//
// It is built on first use and then shared by every call a pass asks about in one function. Building
// it inside isModuleHookCall made each call cost a walk of every instruction in the function, which
// is quadratic in a function's size and was 85 MB of map on a cold ahra run. A pass makes one with
// newCalleeProducers for the function it walks, and must not mutate that function's instructions
// while holding it, since the table is not rebuilt.
type calleeProducers struct {
	function *Function
	table    map[static_single_assignment.IdentifierId]InstructionValue
	// builds counts how many times the table was built, for the test that it is built once.
	builds int
}

func newCalleeProducers(function *Function) *calleeProducers {
	return &calleeProducers{function: function}
}

// producer returns the value that produced identifier, building the table on the first ask.
func (producers *calleeProducers) producer(identifier static_single_assignment.IdentifierId) InstructionValue {
	if producers.table == nil {
		producers.builds++
		producers.table = make(map[static_single_assignment.IdentifierId]InstructionValue, len(producers.function.Instructions))
		for _, candidate := range producers.function.Instructions {
			if candidate == nil {
				continue
			}
			producers.table[candidate.LValue.Identifier] = candidate.Value
			if store, ok := candidate.Value.(*StoreLocal); ok {
				producers.table[store.LValue.Identifier] = store
			}
		}
	}
	return producers.table[identifier]
}

func isModuleHookCall(function *Function, producers *calleeProducers, instruction *Instruction) bool {
	if function == nil || instruction == nil {
		return false
	}
	var callee Place
	switch call := instruction.Value.(type) {
	case *CallExpression:
		callee = call.Callee
	case *MethodCall:
		callee = call.Property
	default:
		return false
	}
	seen := map[static_single_assignment.IdentifierId]bool{}
	var resolve func(static_single_assignment.IdentifierId, bool) bool
	resolve = func(identifier static_single_assignment.IdentifierId, requireHook bool) bool {
		if seen[identifier] {
			return false
		}
		seen[identifier] = true
		defer delete(seen, identifier)
		switch value := producers.producer(identifier).(type) {
		case *LoadGlobal:
			if value.BindingKind == GlobalBindingKindGlobal || value.Source == "react" {
				return false
			}
			return !requireHook || isHookName(value.Name) || isHookName(value.Imported)
		case *LoadLocal:
			return resolve(value.Place.Identifier, requireHook)
		case *StoreLocal:
			return resolve(value.Value.Identifier, requireHook)
		case *PropertyLoad:
			return (!requireHook || isHookName(string(value.Property))) && resolve(value.Object.Identifier, false)
		case *ComputedLoad:
			property, ok := producers.producer(value.Property.Identifier).(*Primitive)
			if !ok {
				return false
			}
			name, ok := property.Value.(string)
			return ok && (!requireHook || isHookName(name)) && resolve(value.Object.Identifier, false)
		default:
			return false
		}
	}
	if call, ok := instruction.Value.(*MethodCall); ok {
		if property, ok := producers.producer(call.Property.Identifier).(*Primitive); ok {
			name, ok := property.Value.(string)
			return ok && isHookName(name) && resolve(call.Receiver.Identifier, false)
		}
	}
	return resolve(callee.Identifier, true)
}
