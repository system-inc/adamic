// Mutable ranges: over which span of a function's evaluation a value is still being written.
//
// The pass is mutation_aliasing's, which this IR shares with Adamic's flow graph: React's
// `InferMutationAliasingRanges`, its alias graph, and the reasoning and measurements behind every rule
// in it. What lives here is this IR's side of it: rangeGraph, which is how the pass reads a Function
// and the effect table this IR infers, and the entry points that infer the effects and run it.
// RangesForNested and MutationSites read this IR's own instruction shapes, so they stay with it.
package high_level_intermediate_representation

import (
	"github.com/system-inc/cohere/mutation_aliasing"
	"github.com/system-inc/cohere/static_single_assignment"
)

// InferMutableRanges computes the mutable range of every value in function, returning them as a
// table keyed by value. mutation_aliasing.InferMutableRanges is the pass and says what it requires.
//
// Infers effects for itself, with InferAliasingEffects. A caller holding an effect table already can
// pass it to `InferMutableRangesWithEffects` and avoid the second inference.
func InferMutableRanges(function *Function) *mutation_aliasing.MutableRanges {
	if function == nil {
		return &mutation_aliasing.MutableRanges{}
	}
	return InferMutableRangesWithEffects(function, InferAliasingEffects(function))
}

// InferMutableRangesWithEffects is the full pass, taking an effect table the caller already has.
//
// React's rules hold by default: a component's or hook's parameters are Frozen, so none of
// Options is set.
func InferMutableRangesWithEffects(function *Function, effects *AliasingEffects) *mutation_aliasing.MutableRanges {
	if function == nil {
		return &mutation_aliasing.MutableRanges{}
	}
	return mutation_aliasing.InferMutableRanges(newRangeGraph(function, effects), function, mutation_aliasing.Options{})
}

// RangesForNested returns one range table per nested function, keyed by FunctionId.
//
// Separate from `InferMutableRanges` because identifier tables are per-function: an id means one
// value in this function and a different value in a nested one, so merging the tables would answer
// confidently and wrongly. Upstream keeps them separate for the same reason and additionally resets
// a nested function's context ranges after analysing it.
func RangesForNested(function *Function) map[FunctionId]*mutation_aliasing.MutableRanges {
	if function == nil {
		return nil
	}
	out := map[FunctionId]*mutation_aliasing.MutableRanges{}
	for index, nested := range function.Functions {
		out[FunctionId(index)] = InferMutableRanges(nested)
	}
	return out
}

// MutationSite is one instruction whose range widening needs effect facts.
//
// It is the seam to Stage 1: when `Place.Effect` is filled, the extension half of this pass is a
// walk over these calling `AliasingState.mutate`, and nothing in the definition half changes.
type MutationSite struct {
	// Instruction is the instruction that may mutate a value.
	Instruction InstructionId
	// Order is where it sits in evaluation.
	Order static_single_assignment.EvaluationOrder
	// Target is the value the instruction shape says is written, where the shape names one.
	Target static_single_assignment.IdentifierId
	// Kind names the shape, for reporting a count by category rather than a bare total.
	Kind MutationSiteKind
}

// MutationSiteKind is the instruction shape that makes a site a mutation candidate.
type MutationSiteKind uint8

const (
	// MutationSitePropertyStore is `a.b = c`, which mutates the receiver.
	MutationSitePropertyStore MutationSiteKind = iota
	// MutationSiteComputedStore is `a[b] = c`, which mutates the receiver.
	MutationSiteComputedStore
	// MutationSiteStoreContext is a write into a captured binding.
	MutationSiteStoreContext
	// MutationSiteCall is a call, whose arguments may be mutated by the callee.
	//
	// This is the shape effects exist to resolve and the reason the count below is a CEILING rather
	// than a measurement of real mutations: most calls mutate nothing, and only an effect table can
	// say which.
	MutationSiteCall
)

func (k MutationSiteKind) String() string {
	switch k {
	case MutationSitePropertyStore:
		return "property-store"
	case MutationSiteComputedStore:
		return "computed-store"
	case MutationSiteStoreContext:
		return "store-context"
	case MutationSiteCall:
		return "call"
	default:
		return "<unknown>"
	}
}

// MutationSites returns the instructions whose range widening this pass cannot perform.
//
// Computed from the instruction SHAPE alone, which is exactly what is available without effects. It
// is the count the package comment reports and the list a Stage 1 consumer walks.
//
// # Calls are included and they are the reason this is a ceiling
//
// A `PropertyStore` definitely mutates its receiver, so its presence here is a real missing
// widening. A `CallExpression` only MIGHT mutate an argument, and upstream decides which by reading
// the callee's signature into an effect. Counting calls here would overstate the gap by an order of
// magnitude, so they are returned as their own kind and the package comment reports the two totals
// separately rather than summing them into one misleading number.
func MutationSites(function *Function) []MutationSite {
	if function == nil {
		return nil
	}
	var sites []MutationSite
	for _, block := range function.Blocks {
		for _, instructionId := range block.Instructions {
			instruction := function.Instructions[instructionId]
			if instruction == nil {
				continue
			}
			switch value := instruction.Value.(type) {
			case *PropertyStore:
				sites = append(sites, MutationSite{
					Instruction: instructionId,
					Order:       instruction.Order,
					Target:      value.Object.Identifier,
					Kind:        MutationSitePropertyStore,
				})
			case *ComputedStore:
				sites = append(sites, MutationSite{
					Instruction: instructionId,
					Order:       instruction.Order,
					Target:      value.Object.Identifier,
					Kind:        MutationSiteComputedStore,
				})
			case *StoreContext:
				sites = append(sites, MutationSite{
					Instruction: instructionId,
					Order:       instruction.Order,
					Target:      value.Value.Identifier,
					Kind:        MutationSiteStoreContext,
				})
			case *CallExpression:
				sites = append(sites, MutationSite{
					Instruction: instructionId,
					Order:       instruction.Order,
					Target:      value.Callee.Identifier,
					Kind:        MutationSiteCall,
				})
			case *MethodCall:
				sites = append(sites, MutationSite{
					Instruction: instructionId,
					Order:       instruction.Order,
					Target:      value.Receiver.Identifier,
					Kind:        MutationSiteCall,
				})
			}
		}
	}
	return sites
}

// rangeGraph is this IR's side of mutation_aliasing.Graph: ssaGraph's answers for single assignment,
// and this IR's answer at each seam where it differs from Adamic's.
//
// It carries the effect table the pass reads, and the values derived from a useRef result, which
// React's closure rule reads. A zero rangeGraph answers everything but Effects and Closure, which is
// all mutation_aliasing.BlockFirstOrder asks.
type rangeGraph struct {
	ssaGraph
	effects *AliasingEffects
	refs    map[static_single_assignment.IdentifierId]bool
}

// newRangeGraph is the adapter for one function and its effects.
func newRangeGraph(function *Function, effects *AliasingEffects) rangeGraph {
	return rangeGraph{effects: effects, refs: refDerivedValues(function)}
}

func (rangeGraph) InstructionOrder(function *Function, block *BasicBlock, index int) (static_single_assignment.EvaluationOrder, bool) {
	instruction := function.Instructions[block.Instructions[index]]
	if instruction == nil {
		return 0, false
	}
	return instruction.Order, true
}

func (rangeGraph) TerminalOrder(block *BasicBlock) static_single_assignment.EvaluationOrder {
	return TerminalOrder(block.Terminal)
}

func (g rangeGraph) Effects(function *Function, block *BasicBlock, index int) []AliasingEffect {
	return g.effects.Get(block.Instructions[index])
}

// ParametersFrozen is React's rule: a component's or a hook's parameters are Frozen, and only a nested
// function expression's are Mutable.
func (rangeGraph) ParametersFrozen(function *Function) bool {
	return function.Kind == FunctionKindComponent || function.Kind == FunctionKindHook
}

func (rangeGraph) Context(function *Function) []Place { return function.Context }

func (rangeGraph) ReturnValue(block *BasicBlock) (Place, bool) {
	returnTerminal, ok := block.Terminal.(*Return)
	if !ok {
		return Place{}, false
	}
	return returnTerminal.Value, true
}

func (rangeGraph) StoredContextValue(function *Function, block *BasicBlock, index int) (static_single_assignment.IdentifierId, bool) {
	store, ok := function.Instructions[block.Instructions[index]].Value.(*StoreContext)
	if !ok {
		return 0, false
	}
	return store.Value.Identifier, true
}

// Closure is React's rule for a FunctionExpression: created Frozen when every capture is immutable,
// none is derived from a useRef result, and the closure's body only reads them
// (hasReadOnlyClosureEffectsForCaptures).
func (g rangeGraph) Closure(function *Function, block *BasicBlock, index int, into static_single_assignment.IdentifierId,
	kinds map[static_single_assignment.IdentifierId]mutation_aliasing.EffectValueKind) ([]Place, bool) {
	instruction := function.Instructions[block.Instructions[index]]
	expression, ok := instruction.Value.(*FunctionExpression)
	if !ok || into != instruction.LValue.Identifier {
		return nil, false
	}
	allImmutable := len(expression.Captures) > 0
	for _, capture := range expression.Captures {
		_, immutable := kinds[capture.Identifier]
		allImmutable = allImmutable && immutable && !g.refs[capture.Identifier]
	}
	frozen := allImmutable && int(expression.Function) < len(function.Functions) &&
		hasReadOnlyClosureEffectsForCaptures(function.Functions[expression.Function], expression.Captures, kinds)
	return expression.Captures, frozen
}
