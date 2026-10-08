package lower

import (
	"math"

	"github.com/system-inc/adamic/internal/ir"
)

// extendedCounter supplements the static proof without reinterpreting any of
// its unknown expressions. Unknown invariant numeric locals get a guarded native
// specialization; the ordinary IR and the double fallback remain unchanged.
func extendedCounter(program *ir.Program, block ir.Block, known facts) {
	if len(block.Body) != 2 {
		return
	}
	declaration, ok := block.Body[0].(ir.Declare)
	if !ok || declaration.Value == nil {
		return
	}
	loop, ok := block.Body[1].(ir.Loop)
	if !ok || loop.CheckAfter || len(loop.Update) != 1 {
		return
	}
	condition, ok := loop.Condition.(ir.Binary)
	if !ok {
		return
	}
	read, direct := condition.Left.(ir.Read)
	square := false
	if !direct {
		product, ok := condition.Left.(ir.Binary)
		if !ok || product.Operator != ir.Multiply {
			return
		}
		left, a := product.Left.(ir.Read)
		right, b := product.Right.(ir.Read)
		if !a || !b || left.Local != right.Local {
			return
		}
		read, square = left, true
	}
	counter := read.Local
	local := program.Locals[counter]
	if declaration.Local != counter || local.Counter || local.Type != ir.Number || local.Global || local.Captured || !perIteration(loop, counter) {
		return
	}
	written := map[int]bool{}
	findAssigned(loop.Body, written)
	if written[counter] {
		return
	}
	findAssigned(loop.Update, written)
	startLow, startHigh, startWhole := boundRange(declaration.Value, known, 0)
	if !startWhole || !counterStartHasPositiveZero(declaration.Value, known) {
		return
	}
	update, ok := loop.Update[0].(ir.Assign)
	if !ok || update.Local != counter || update.Checked {
		return
	}
	sum, ok := update.Value.(ir.Binary)
	if !ok || (sum.Operator != ir.Add && sum.Operator != ir.Subtract) {
		return
	}
	previous, ok := sum.Left.(ir.Read)
	if !ok || previous.Local != counter || !invariantNumber(program, sum.Right, written) {
		return
	}
	stepLow, stepHigh, stepWhole := boundRange(sum.Right, known, 0)
	if stepWhole {
		if stepLow < 1 {
			return
		}
	} else {
		if !unknownInvariantNumber(program, sum.Right, written, known) {
			return
		}
		stepLow, stepHigh = 1, exact
	}
	ascending := sum.Operator == ir.Add
	if ascending && condition.Operator != ir.Less && condition.Operator != ir.LessOrEqual {
		return
	}
	if !ascending && condition.Operator != ir.Greater && condition.Operator != ir.GreaterOrEqual {
		return
	}
	if square && (!ascending || startLow < 0) {
		return
	}
	low, high, boundWhole := boundRange(condition.Right, known, 0)
	if !boundWhole {
		if !unknownInvariantNumber(program, condition.Right, written, known) {
			return
		}
		low, high = -exact, exact
	}
	// A changing numeric bound cannot be snapshotted for the new proof. The
	// old proof's lengths may change, but always stay within their static range.
	if read, ok := condition.Right.(ir.Read); ok && written[read.Local] {
		return
	}
	if !boundWhole && !invariantNumber(program, condition.Right, written) {
		return
	}
	inclusive := condition.Operator == ir.LessOrEqual || condition.Operator == ir.GreaterOrEqual
	var inBody span
	safe := false
	maxStep := exact
	if square {
		last := high
		if !inclusive {
			last--
		}
		// The counter is nonnegative and its double square must be <= last.
		// Whole squares through 2^53 are exact. sqrt at an integer square is exact;
		// rounding up at a nonsquare can only overestimate this inclusive range.
		end := math.Floor(math.Sqrt(math.Max(0, last)))
		inBody = span{startLow, end}
		maxStep = exact - math.Max(startHigh, end)
		safe = stepHigh <= maxStep
	} else if ascending {
		end := high
		if !inclusive {
			end--
		}
		safe = high <= exact-stepHigh
		if !inclusive {
			safe = high <= exact-(stepHigh-1)
		}
		// Under the entry guard even an unknown bound cannot let the next
		// update cross 2^53. stepLow is a conservative minimum step.
		if !boundWhole || !safe {
			end = math.Min(end, exact-stepLow)
		}
		if !inclusive && (!boundWhole || !safe) {
			end = math.Min(high-1, exact-stepLow)
		}
		inBody = span{startLow, end}
	} else {
		first := low
		if !inclusive {
			first++
		}
		safe = low >= -exact+stepHigh
		if !inclusive {
			safe = low >= -exact+(stepHigh-1)
		}
		if !boundWhole || !safe {
			first = math.Max(first, -exact+stepLow)
		}
		inBody = span{first, startHigh}
	}
	// Constant steps against statically known bounds keep all the old overflow
	// refusals. A variable step may instead pass the same test at loop entry.
	_, constantStep := constantOf(sum.Right)
	if boundWhole && stepWhole && !safe && constantStep && !square {
		return
	}
	dependencies := counterDependencies(known, declaration.Value, sum.Right, condition.Right)
	guard := !boundWhole || !stepWhole || !safe
	// A static range proof can tolerate changing lengths, just as the original
	// proof does. An entry guard must instead read a truly invariant bound.
	if guard && !invariantNumber(program, condition.Right, written) {
		return
	}
	if !guard && len(dependencies) == 0 {
		program.Locals[counter].Counter = true
	} else {
		plan := &ir.CounterGuard{Dependencies: dependencies}
		if guard {
			plan.Bound, plan.Step = condition.Right, sum.Right
			plan.Ascending, plan.Inclusive, plan.Square = ascending, inclusive, square
			plan.MaxStep = maxStep
		}
		program.Locals[counter].CounterGuard = plan
		known.guarded[counter] = true
	}
	known.counters[counter] = inBody
}

// Products of two nonnegative, already whole counters cannot manufacture -0.
// Other formerly refused products (including a possibly zero value times a
// negative constant) still go through the original negative-zero proof.
func counterStartHasPositiveZero(value ir.Expression, known facts) bool {
	if neverNegativeZero(value, known, 0) {
		return true
	}
	product, ok := value.(ir.Binary)
	if !ok || product.Operator != ir.Multiply {
		return false
	}
	left, a := product.Left.(ir.Read)
	right, b := product.Right.(ir.Read)
	if !a || !b {
		return false
	}
	x, a := known.counters[left.Local]
	y, b := known.counters[right.Local]
	return a && b && x.low >= 0 && y.low >= 0 && x.low <= x.high && y.low <= y.high
}

// Only numeric values whose dependencies cannot change during this loop may
// be evaluated by its entry guard. No globals, capture cells, property reads,
// mutable lengths, or calls: a callee could change those without an Assign here.
func invariantNumber(program *ir.Program, value ir.Expression, written map[int]bool) bool {
	switch value := value.(type) {
	case ir.NumberConstant:
		return true
	case ir.Read:
		local := program.Locals[value.Local]
		return value.Of == ir.Number && !local.Global && !local.Captured && !value.Checked && !written[value.Local]
	case ir.Unary:
		return (value.Operator == ir.Plus || value.Operator == ir.Negate) && invariantNumber(program, value.Operand, written)
	case ir.Binary:
		return (value.Operator == ir.Add || value.Operator == ir.Subtract || value.Operator == ir.Multiply) && invariantNumber(program, value.Left, written) && invariantNumber(program, value.Right, written)
	}
	return false
}

// An invariant numeric read can be checked at entry regardless of its
// initializer or writes outside this loop. Keep a known bad constant's old
// refusal rather than emitting an integer branch that can never be taken.
func unknownInvariantNumber(program *ir.Program, value ir.Expression, written map[int]bool, known facts) bool {
	read, ok := value.(ir.Read)
	if !ok || !invariantNumber(program, value, written) {
		return false
	}
	if initializer, declared := known.declared[read.Local]; declared && !known.assigned[read.Local] {
		if constant, knownConstant := constantOf(initializer); knownConstant && !whole(constant, exact) {
			return false
		}
	}
	return true
}

func counterDependencies(known facts, values ...ir.Expression) []int {
	var dependencies []int
	seen := map[int]bool{}
	var visit func(ir.Expression)
	visit = func(value ir.Expression) {
		switch value := value.(type) {
		case ir.Read:
			if seen[value.Local] {
				return
			}
			seen[value.Local] = true
			if known.guarded[value.Local] {
				dependencies = append(dependencies, value.Local)
			}
			// A const can carry the range and guarded proof of an outer counter.
			if initializer, ok := known.declared[value.Local]; ok && !known.assigned[value.Local] {
				visit(initializer)
			}
		case ir.Unary:
			visit(value.Operand)
		case ir.Binary:
			visit(value.Left)
			visit(value.Right)
		}
	}
	for _, value := range values {
		visit(value)
	}
	return dependencies
}
