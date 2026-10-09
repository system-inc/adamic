package control_flow_graph

// This file is a LOCAL ADDITION, not vendored. See the change list in cfg.go's header.
//
// It is the monotone dataflow framework the graph never had. `no-useless-assignment` wrote its own
// lattice, meet, and transfer inline and unexported inside `package core`, and two class rules
// (`constructor-super`, `no-this-before-super`) describe a five-state lattice in their doc comments
// and then do not use the graph at all. This gives all three one shape to reach for.

// Direction is which way a dataflow propagates.
type Direction uint8

const (
	// Forward propagates from predecessors to successors: reaching definitions, constant
	// propagation, the "has super() been called" question the two class rules describe.
	Forward Direction = iota
	// Backward propagates from successors to predecessors: liveness, and any "is this value ever
	// used again" question.
	Backward
)

// Solution is the settled boundary values of one dataflow, indexed by Block.Index.
type Solution[V any, E any] struct {

	// entry[i] is the value flowing INTO block i in the analysis's own direction: on entry to the
	// block for a Forward analysis, on exit from it for a Backward one.
	entry []V

	// exit[i] is the value flowing OUT of block i in the analysis's own direction, which is
	// Transfer applied to entry[i].
	exit []V

	reachable []bool
	rounds    int
}

// maximumRounds bounds the fixed-point iteration.
//
// A monotone dataflow over a lattice of height h converges in at most h * blocks rounds, and every
// lattice this framework is built for has small height (a boolean per symbol, a five-state enum).
// The bound is deliberately far above any real convergence and exists to turn a non-monotone
// Transfer or a broken Equal into a panic naming the analysis rather than a hang: a linter that
// stops and says which file is debuggable, one that spins is not. Measured on this tree, the
// deepest real convergence is well under ten rounds.
const maximumRounds = 10000

// LatticeOf is the value a dataflow carries at each block boundary, and how two of them combine.
//
// V is the lattice value, E is the graph's event type. Both are type parameters on the interface
// rather than on its methods, because Go has no method-level type parameters. The consequence is
// that a consumer writes one small struct implementing this rather than passing five closures, and
// a value that is cheap (a bitset, a small enum) costs no interface boxing per block while an
// expensive one (a symbol set) is free to be a pointer.
//
// # What the framework requires of an implementation, and what happens when it is not met
//
// Meet must be commutative, associative, and idempotent, and Equal must agree with it. Transfer
// must be monotone: a higher input never produces a lower output. Together those are what make the
// fixed point exist and be unique, and `Solve` cannot check any of them.
//
// What it CAN check is termination, and it does. A lattice whose values keep changing forever — the
// usual causes are a Meet that is not idempotent, or an Equal comparing pointers where it should
// compare contents — would make `Solve` iterate without end, so `Solve` bounds its rounds and
// panics instead. A linter that stops and says why is debuggable; one that spins on one file in a
// build is not. The bound is stated at `maximumRounds`.
type LatticeOf[V any, E any] interface {
	// Bottom is the value a block starts at before anything has flowed into it. For a "must"
	// analysis whose meet is intersection this is the TOP of the lattice rather than its bottom,
	// which is the standard confusion; what the framework needs is only that it is the identity for
	// Meet, so meeting it with anything yields that thing.
	Bottom() V

	// Entry is the value at the boundary the analysis starts from: the graph entry for a Forward
	// analysis, each exit block for a Backward one. It is where the analysis's assumptions about the
	// world outside this code path root live — a liveness pass says nothing is live at an exit, a
	// reaching-definitions pass says the parameters are already defined.
	Entry() V

	// Meet combines the values arriving from two neighbours. It is folded left-to-right over a
	// block's neighbours and must not mutate either argument, because a value may be a neighbour's
	// settled state and reachable from more than one merge.
	Meet(left, right V) V

	// Transfer computes the value leaving a block from the value entering it. The block is handed
	// over whole rather than event by event, so the implementation reads its Events in whatever
	// order its own direction wants — forward for a Forward analysis, backward for a Backward one.
	//
	// It must not mutate the incoming value. `Solve` holds that as the block's settled boundary and
	// compares against it to decide whether anything moved.
	Transfer(block *Block[E], incoming V) V

	// Equal reports whether two values are the same point in the lattice. It is what decides the
	// fixed point has been reached, so an Equal stricter than the lattice's real equality makes
	// Solve iterate longer than it needs to, and one that is looser makes it stop early with a wrong
	// answer.
	Equal(left, right V) bool
}

// Solve runs a monotone dataflow over graph's reachable blocks to a fixed point.
//
// # Why this exists rather than each analysis writing its own loop
//
// `no-useless-assignment` wrote one, and its own doc comment names two things it settled for: it
// iterates blocks by DESCENDING INDEX rather than reverse postorder, and it rebuilds each block's
// successor union from scratch every round. Both are consequences of the graph not exposing an
// order or predecessor edges, and both are fixed here for every consumer at once rather than in
// one rule.
//
// Descending index is not reverse postorder. It happens to be close for a graph laid out in
// construction order and it is not the same thing: a `finally` block is laid out twice, and a loop's
// incrementor is laid out after its body, so the index order crosses back edges in places the
// postorder does not. It still converges — a fixed point does not care what order it is reached in
// — but it takes more rounds, and the number of extra rounds is the loop nesting depth.
//
// # Unreachable blocks
//
// They are excluded, and every consumer already wanted that. `no-useless-assignment`'s comment
// spells out why in its own terms: an unreachable block's events are the same source positions laid
// out a second time, and letting them contribute lets a `finally` copy's own kill mask a live read.
// The framework makes that the default rather than something each analysis remembers.
//
// A Backward analysis seeds from FinalBlocks and ThrownBlocks both. Measured on this substrate,
// those two sets cover every reachable terminal block, including the `finally` duplicate laid out on
// the abrupt path — the shape that looks like it would escape them. The measurement is recorded at
// the seeding code below, because it is a property of the builder rather than of this file.
func Solve[V any, E any](graph *Graph[E], direction Direction, lattice LatticeOf[V, E]) *Solution[V, E] {
	solution := &Solution[V, E]{rounds: 0}
	if graph == nil || len(graph.Blocks) == 0 {
		return solution
	}

	blockCount := len(graph.Blocks)
	solution.entry = make([]V, blockCount)
	solution.exit = make([]V, blockCount)
	solution.reachable = make([]bool, blockCount)

	dominators := AnalyzeDominators(graph)
	order := dominators.ReversePostOrder()
	if direction == Backward {
		reversed := make([]*Block[E], len(order))
		for index, block := range order {
			reversed[len(order)-1-index] = block
		}
		order = reversed
	}

	for _, block := range order {
		solution.reachable[block.Index()] = true
		solution.entry[block.Index()] = lattice.Bottom()
		solution.exit[block.Index()] = lattice.Bottom()
	}

	// seeds are the blocks whose incoming value comes from outside the graph rather than from a
	// neighbour, so Meet must not lower them to Bottom.
	seeds := make([]bool, blockCount)
	if direction == Forward {
		seeds[0] = true
	} else {
		for _, block := range graph.FinalBlocks {
			if block != nil && block.Reachable {
				seeds[block.Index()] = true
			}
		}
		for _, block := range graph.ThrownBlocks {
			if block != nil && block.Reachable {
				seeds[block.Index()] = true
			}
		}
		// A reachable terminal block outside both exit sets would also need seeding, and it does not
		// exist on this substrate. Measured rather than assumed, because the shape it would come
		// from is exactly the one that has already cost this graph's consumers a false positive: the
		// `finally` duplicate laid out on the abrupt path is a terminal block with no successors,
		// and the question is whether the builder puts it in an exit set.
		//
		// It does. Probed over `try { return 1; } finally {}`, the same with a `throw`, the same
		// with a conditional return, a generator yielding inside a `try`, `try/catch/finally` with a
		// return, and `for (;;)`, and every one reports ZERO reachable terminal blocks outside
		// FinalBlocks ∪ ThrownBlocks. `markFinal`/`markThrown` are called on the abrupt-path copy,
		// which is what puts it in.
		//
		// So the defensive third clause is not written here. It would be unreachable, no test could
		// see it, and an unreachable branch guarding a correctness property reads as coverage
		// without being it. `TestSolveSeedsEveryTerminalBlockBackward` asserts the property this
		// relies on, so if the builder ever stops marking that copy, the seeding gap is named rather
		// than silently opening.
	}

	for changed := true; changed; {
		changed = false
		solution.rounds++
		if solution.rounds > maximumRounds {
			panic("controlflow: dataflow did not converge — the lattice's Meet is probably not idempotent, or Equal disagrees with it")
		}
		for _, block := range order {
			index := block.Index()

			incoming := lattice.Bottom()
			if seeds[index] {
				incoming = lattice.Entry()
			}
			if direction == Forward {
				for _, predecessor := range dominators.Predecessors(block) {
					incoming = lattice.Meet(incoming, solution.exit[predecessor.Index()])
				}
			} else {
				for _, successor := range block.Successors {
					if successor == nil || !successor.Reachable {
						continue
					}
					incoming = lattice.Meet(incoming, solution.exit[successor.Index()])
				}
			}

			outgoing := lattice.Transfer(block, incoming)
			if !lattice.Equal(incoming, solution.entry[index]) {
				solution.entry[index] = incoming
				changed = true
			}
			if !lattice.Equal(outgoing, solution.exit[index]) {
				solution.exit[index] = outgoing
				changed = true
			}
		}
	}

	return solution
}

// In returns the value flowing into block in the analysis's direction: on entry for a Forward
// analysis, on exit for a Backward one. The second result is false for an unreachable block.
func (s *Solution[V, E]) In(block *Block[E]) (V, bool) {
	var zero V
	if s == nil || block == nil || block.Index() >= len(s.entry) || !s.reachable[block.Index()] {
		return zero, false
	}
	return s.entry[block.Index()], true
}

// Out returns the value flowing out of block in the analysis's direction. The second result is
// false for an unreachable block.
func (s *Solution[V, E]) Out(block *Block[E]) (V, bool) {
	var zero V
	if s == nil || block == nil || block.Index() >= len(s.exit) || !s.reachable[block.Index()] {
		return zero, false
	}
	return s.exit[block.Index()], true
}

// Rounds reports how many times the solver swept every block before nothing moved.
//
// It is exposed because it is the measurement that says whether the iteration order is doing its
// job. A Forward analysis over an acyclic graph in reverse postorder settles in two rounds, one to
// propagate and one to observe nothing moved; more than that means back edges, and a lot more than
// that means the order is wrong or the lattice is taller than it looks.
func (s *Solution[V, E]) Rounds() int {
	if s == nil {
		return 0
	}
	return s.rounds
}
