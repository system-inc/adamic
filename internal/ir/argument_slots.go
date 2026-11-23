package ir

import (
	"encoding/binary"
	"math/bits"
	"slices"
)

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
	p.FixedArgumentSlots = fixed
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
			property, _, method := MethodCallee(call.Closure)
			if !method || property.Name != function.MethodName {
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

// argumentFacts is what PackedCountNeeded and ClosureConventionNeeded derive from the whole program,
// computed once: the emitter asks for every function, and for every closure call, so deriving it per
// question walked every function-type target set each time (quadratic in a program the size of a port).
// It's keyed to the sizes of what it reads, so a program that changes after the first question gets
// fresh facts rather than stale ones.
type argumentFacts struct {
	functions, sets int
	packed          []bool
	convention      bool
}

func (p *Program) argumentFactsOf() *argumentFacts {
	if facts := p.argumentFacts; facts != nil && facts.functions == len(p.Functions) && facts.sets == len(p.FunctionTypeTargets) {
		return facts
	}
	facts, _ := p.deriveArgumentFacts()
	p.argumentFacts = facts
	return facts
}

// deriveArgumentFacts is argumentFactsOf's derivation, and how many peer visits it made, which the visit
// count's test reads. Each distinct target set is read once: a program holds one set under many function
// types (every type whose call sites see the same declarations), and asking each member of each copy about
// every other member made m copies of a set of k cost m·k² (#67w3g0x: 83 to 97% of C emission on a program
// shaped like tsc's imports, growing about cubically). Once per set, every member reads its peers' absent
// representations from one table of them (packedPeers), so a set costs its members' parameters.
func (p *Program) deriveArgumentFacts() (*argumentFacts, int) {
	facts := &argumentFacts{functions: len(p.Functions), sets: len(p.FunctionTypeTargets), packed: make([]bool, len(p.Functions))}
	for index, function := range p.Functions {
		facts.packed[index] = function.ReadsArguments || function.RestElement != 0 && (function.Closure || function.Receiver)
	}
	visits := 0
	read := map[string]bool{}
	var words []uint64
	var key []byte
	for _, targets := range p.FunctionTypeTargets {
		words, key = targetSetKey(targets, words, key)
		if read[string(key)] {
			continue
		}
		read[string(key)] = true
		visits += p.packedPeers(setMembers(words), facts.packed)
	}
	for _, packed := range facts.packed {
		facts.convention = facts.convention || packed
	}
	return facts, visits
}

// PackedCountNeeded is a callee observation, not a property of all function values.
func (p *Program) PackedCountNeeded(function int) bool {
	return p.argumentFactsOf().packed[function]
}

// targetSetKey writes a target set's functions into words as a bitset, one bit per function index, and into key as
// the bitset's bytes. Two sets of the same functions give one key in any order and with any repeats, so a set held
// under many function types costs one pass over each copy, never a copy sorted or formatted for its key (#bqwjdep:
// at a tsc-shaped program of 22,861 lines, 3,178 lists of about 2,350 members held 5 distinct sets). words and key
// are reused from the last call.
func targetSetKey(targets []int, words []uint64, key []byte) ([]uint64, []byte) {
	highest := -1
	for _, target := range targets {
		highest = max(highest, target)
	}
	count := highest/64 + 1
	words = slices.Grow(words[:0], count)[:count]
	clear(words)
	for _, target := range targets {
		words[target/64] |= uint64(1) << (target % 64)
	}
	key = key[:0]
	for _, word := range words {
		key = binary.LittleEndian.AppendUint64(key, word)
	}
	return words, key
}

// setMembers is a bitset's functions, each once and in order, so two sets of the same functions are read as one.
func setMembers(words []uint64) []int {
	var members []int
	for index, word := range words {
		for word != 0 {
			members = append(members, index*64+bits.TrailingZeros64(word))
			word &= word - 1
		}
	}
	return members
}

// absentMasks are, for each parameter index, the absent representations (absentRepresentation, one bit
// each) the peers of one class hold at that index.
type absentMasks []uint8

// packedPeers marks in packed each member of one target set that a peer makes need the count, and returns
// how many member visits it made. A zero-argument view can reach optional parameters held differently: one
// padding word cannot be absent in both representations, so those callees distinguish presence with the
// count instead of interpreting that word.
//
// A member is compared with the set's peers of its class: a closure with every closure of the set, any
// other function with the set's methods (receivers that aren't closures) of its own name. At each of its
// parameter indexes that is neither its own rest slot nor a peer's, a member whose parameter may be missing
// (a Maybe pair or a reference) needs the count when a peer holds absence another way there. So the peers'
// representations at each index are gathered once into a class's absentMasks, and each member reads them:
// the answer a walk of every pair of members gives, in a pass over the members twice.
func (p *Program) packedPeers(members []int, packed []bool) int {
	visits := 0
	var closures absentMasks
	methods := map[string]absentMasks{}
	for _, target := range members {
		visits++
		peer := p.Functions[target]
		if !peer.Closure && !peer.Receiver {
			continue
		}
		masks := methods[peer.MethodName]
		if peer.Closure {
			masks = closures
		}
		compared := peer.Parameters
		if peer.Receiver {
			compared = compared[1:]
		}
		for index, parameter := range compared {
			if peer.RestElement != 0 && index == len(compared)-1 {
				continue
			}
			for len(masks) <= index {
				masks = append(masks, 0)
			}
			masks[index] |= 1 << absentRepresentation(p.Locals[parameter].Type)
		}
		if peer.Closure {
			closures = masks
		} else {
			methods[peer.MethodName] = masks
		}
	}
	for _, target := range members {
		visits++
		if packed[target] {
			continue
		}
		f := p.Functions[target]
		masks := methods[f.MethodName]
		if f.Closure {
			masks = closures
		}
		parameters := f.Parameters
		if f.Receiver {
			parameters = parameters[1:]
		}
		for index, parameter := range parameters {
			if index >= len(masks) {
				break
			}
			if f.RestElement != 0 && index == len(parameters)-1 {
				continue
			}
			of := p.Locals[parameter].Type
			if (of.IsMaybe() || of.IsReference()) && masks[index]&^(1<<absentRepresentation(of)) != 0 {
				packed[target] = true
				break
			}
		}
	}
	return visits
}

// absentRepresentation is how a parameter of the type holds absence: 0 for a reference's null, 1 for a
// MaybeNumber's pair, 2 for a MaybeBoolean's. packedPeers keeps one bit of a uint8 for each, so an answer past 7
// would fall off its mask: widen absentMasks before adding one.
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
	return p.argumentFactsOf().convention
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
