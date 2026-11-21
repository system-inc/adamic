package ir

import (
	"math/rand"
	"testing"
)

// packedCountNeededByPairs is the derivation as it was before #67w3g0x: each function asks every member of
// every set holding it, pair by pair. It is the reference the per-set tables (packedPeers) are held to.
func (p *Program) packedCountNeededByPairs(function int) bool {
	f := p.Functions[function]
	if f.ReadsArguments || f.RestElement != 0 && (f.Closure || f.Receiver) {
		return true
	}
	for _, targets := range p.FunctionTypeTargets {
		holds := false
		for _, target := range targets {
			holds = holds || target == function
		}
		if !holds {
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

// randomArgumentProgram is a program of functions drawn from every shape the derivation tells apart:
// closures and methods and ordinary functions, receivers, rest elements, readers of arguments, two
// method names, and parameters of every absent representation, in sets that overlap, repeat, hold a
// function twice and are empty.
func randomArgumentProgram(random *rand.Rand) Program {
	types := []Type{Number, MaybeNumber, Boolean, MaybeBoolean, String, Object, Closure, Union}
	program := Program{FunctionTypeTargets: map[int][]int{}}
	for range 12 {
		program.Locals = append(program.Locals, Local{Type: types[random.Intn(len(types))]})
	}
	names := []string{"", "visit", "close"}
	for range 3 + random.Intn(10) {
		function := Function{Closure: random.Intn(2) == 0, Receiver: random.Intn(3) == 0, MethodName: names[random.Intn(len(names))], ReadsArguments: random.Intn(12) == 0}
		count := random.Intn(4)
		if function.Receiver {
			count++
		}
		for range count {
			function.Parameters = append(function.Parameters, random.Intn(len(program.Locals)))
		}
		if len(function.Parameters) > 0 && random.Intn(5) == 0 {
			function.RestElement = Number
		}
		program.Functions = append(program.Functions, function)
	}
	for set := range 1 + random.Intn(6) {
		var targets []int
		for range random.Intn(len(program.Functions) + 2) {
			targets = append(targets, random.Intn(len(program.Functions)))
		}
		program.FunctionTypeTargets[set] = targets
		if random.Intn(3) == 0 {
			program.FunctionTypeTargets[set+100] = append([]int(nil), targets...)
		}
	}
	return program
}

// The per-set tables give every function the answer the pair-by-pair walk gives, over 20,000 programs drawn
// from every shape the walk tells apart. Mutants: closures and methods read from one class's table (a
// closure needs the count for a method's parameter); a member's own representation counted against it (the
// mask's own bit kept); a peer's rest slot gathered into the table.
func TestPackedPeersAnswerAsThePairWalkDoes(t *testing.T) {
	t.Parallel()
	random := rand.New(rand.NewSource(67))
	for round := range 20000 {
		program := randomArgumentProgram(random)
		facts, _ := program.deriveArgumentFacts()
		for function := range program.Functions {
			if want := program.packedCountNeededByPairs(function); facts.packed[function] != want {
				t.Fatalf("program %d, function %d: per-set tables say %t, the pair walk %t\n%+v", round, function, facts.packed[function], want, program)
			}
		}
	}
}

// Deriving the facts visits each distinct target set's members a fixed number of times, never once per
// pair, and a set held under many function types is read once: 200 copies of one set of 400 closures cost
// 800 visits, where the pair walk made 200·400² = 32,000,000 (#67w3g0x). Mutant: the copies read each time (160,000
// visits).
func TestPackedPeersVisitEachSetOnceAndLinearly(t *testing.T) {
	t.Parallel()
	const width, copies = 400, 200
	program := Program{Locals: []Local{{Type: Number}, {Type: MaybeNumber}}, FunctionTypeTargets: map[int][]int{}}
	var targets []int
	for index := range width {
		program.Functions = append(program.Functions, Function{Closure: true, Parameters: []int{index % 2}})
		targets = append(targets, index)
	}
	for held := range copies {
		program.FunctionTypeTargets[held] = targets
	}
	facts, visits := program.deriveArgumentFacts()
	if visits != 2*width {
		t.Fatalf("%d copies of a set of %d made %d visits, want %d: each distinct set once, each member twice", copies, width, visits, 2*width)
	}
	// A MaybeNumber beside a Number at the same index: every MaybeNumber needs the count, and no Number does.
	for index, packed := range facts.packed {
		if want := index%2 == 1; packed != want {
			t.Fatalf("function %d: packed %t, want %t", index, packed, want)
		}
	}
	t.Logf("%d copies of a set of %d: %d visits", copies, width, visits)
}
