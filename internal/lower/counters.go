package lower

import (
	"math"

	"github.com/system-inc/adamic/internal/ir"
)

// counters marks the loop counters the native backend may keep in an integer (ir.Local.Counter).
//
// A number in JavaScript is a double, and a loop counter is one too. But a counter that starts at a
// whole number, steps by a whole number, and stops at a bound that's a whole number in a range known
// at compile time can only ever hold the whole numbers between its start and one step past its
// bound. When every one of those is no larger than 2^53 in size, each is a double exactly, and every
// step on the way is exact too, so an integer holds the same values, and read as a double it's the
// same number: the same answers, for an integer's step and an integer index into an array. The loop
// has to be one this can see whole:
//
//		for (let counter = start; counter < bound; counter += step)    (or <=, counter++, = counter + step)
//		for (let counter = start; counter > bound; counter -= step)    (or >=, counter--, = counter - step)
//
//	  - start is a whole number within 2^53, of the kind a bound is (below), that can't be -0, since
//	    an integer has no -0 and 1 / counter would tell; step is a whole-number constant, 1 or more;
//	  - nothing but the update writes the counter, and no closure captures it (a cell holds a double);
//	  - bound is, every time the condition reads it, a whole number in a known range: an array's, a
//	    string's or a map's size (from 0 to 2^32), a constant, a variable nothing writes but its
//	    declaration, the counter of a loop it's in the body of (in the range it holds there), or any
//	    of those negated, added, subtracted or multiplied, which keep whole numbers whole while they
//	    stay within 2^53. Division and remainder don't, and NaN, infinities and
//	    fractions aren't whole; their loops keep a double, as does anything else;
//	  - and the last value the counter can take, one step past the largest bound below which it still
//	    runs (or the smallest, counting down), is within 2^53: past it, JavaScript's counter + step is
//	    rounded, and the loop can stop moving, where an integer would go on.
//
// Counting up runs with < or <=, and down with > or >=; a counter that moves away from its bound
// keeps a double, since nothing bounds where it goes. A global counter keeps a double too, since a
// function may run before its loop has.
func counters(program *ir.Program) {
	known := facts{assigned: map[int]bool{}, declared: map[int]ir.Expression{}, counters: map[int]span{}}
	for _, function := range program.Functions {
		findAssigned(function.Body, known.assigned)
		findDeclared(function.Body, known.declared)
	}
	findAssigned(program.Main, known.assigned)
	findDeclared(program.Main, known.declared)
	// Blocks are visited outside in, so a loop's counter is marked, and its range known, before the
	// loops in its body are looked at.
	mark := func(statements []ir.Statement) {
		eachBlock(statements, func(block ir.Block) {
			markCounter(program, block, known)
		})
	}
	for _, function := range program.Functions {
		mark(function.Body)
	}
	mark(program.Main)
}

// facts is what the pass knows of a program's locals: which are written after their declaration,
// what each was declared with, and the range each counter it has marked holds in its loop's body.
type facts struct {
	assigned map[int]bool
	declared map[int]ir.Expression
	counters map[int]span
}

// span is the least and the most a counter can be.
type span struct {
	low, high float64
}

// exact is 2^53: every whole number no larger than it in size is a double exactly.
var exact = math.Ldexp(1, 53)

// markCounter marks the counter of a for loop's block, if it has one this can keep in an integer.
func markCounter(program *ir.Program, block ir.Block, known facts) {
	if len(block.Body) == 0 {
		return
	}
	loop, isLoop := block.Body[len(block.Body)-1].(ir.Loop)
	if !isLoop || loop.CheckAfter || len(loop.Update) != 1 {
		return
	}
	condition, isBinary := loop.Condition.(ir.Binary)
	if !isBinary {
		return
	}
	counterRead, isRead := condition.Left.(ir.Read)
	if !isRead {
		return
	}
	counter := counterRead.Local
	local := program.Locals[counter]
	if local.Type != ir.Number || local.Global || local.Captured || local.ExpressionAssigned || !perIteration(loop, counter) {
		return
	}
	// Its start: a declaration in this block, of a whole number within 2^53, as a bound is, that
	// can't be -0.
	start, isDeclared := declaredIn(block, counter)
	if !isDeclared || start == nil {
		return
	}
	startLow, startHigh, isWhole := boundRange(start, known, 0)
	if !isWhole || !neverNegativeZero(start, known, 0) {
		return
	}
	// Its update, a whole step toward the bound, and nothing in the body that writes it.
	step, isStep := stepOf(loop.Update[0], counter)
	if !isStep {
		return
	}
	written := map[int]bool{}
	findAssigned(loop.Body, written)
	if written[counter] {
		return
	}
	low, high, isWhole := boundRange(condition.Right, known, 0)
	if !isWhole {
		return
	}
	// The furthest the counter can go, one step past the last value the condition lets through, must
	// be within 2^53. It's checked as a limit on the bound, since high + step, done in doubles, would
	// round: 2^53 + 1 is the double 2^53. Each limit is exact, the step being whole and within 2^53.
	// Where the limit is met, the range the counter holds in the body follows: from its start to the
	// last value the condition lets through.
	size := math.Abs(step)
	var inBody span
	switch {
	case step > 0 && condition.Operator == ir.Less:
		// high - 1 + step <= 2^53
		isWhole = high <= exact-(size-1)
		inBody = span{startLow, high - 1}
	case step > 0 && condition.Operator == ir.LessOrEqual:
		// high + step <= 2^53
		isWhole = high <= exact-size
		inBody = span{startLow, high}
	case step < 0 && condition.Operator == ir.Greater:
		// low + 1 - size >= -2^53
		isWhole = low >= -exact+(size-1)
		inBody = span{low + 1, startHigh}
	case step < 0 && condition.Operator == ir.GreaterOrEqual:
		// low - size >= -2^53
		isWhole = low >= -exact+size
		inBody = span{low, startHigh}
	default:
		// A counter moving away from its bound, or a comparison that isn't one of these.
		isWhole = false
	}
	if !isWhole {
		return
	}
	program.Locals[counter].Counter = true
	// A loop in the body that reads the counter, as for (let j = i + 1; ...) does, reads a value in
	// this range: the body never writes the counter, and runs only while the condition holds. A body
	// that never runs has a range whose ends cross, which boundRange refuses.
	known.counters[counter] = inBody
}

// boundRange is the least and the most a bound can be, every time it's read, when it's always a whole
// number within 2^53, and false otherwise.
func boundRange(bound ir.Expression, known facts, depth int) (float64, float64, bool) {
	if depth > 8 {
		return 0, 0, false
	}
	var low, high float64
	switch bound := bound.(type) {
	case ir.Length, ir.StringLength, ir.MapSize:
		// An array's length is below 2^32, and a string's and a map's size far below.
		low, high = 0, math.Ldexp(1, 32)
	case ir.NumberConstant:
		low, high = bound.Value, bound.Value
	case ir.Read:
		// The counter of a loop this is in the body of, in the range it holds there.
		if counter, isCounter := known.counters[bound.Local]; isCounter {
			low, high = counter.low, counter.high
			break
		}
		// A variable nothing writes but its declaration: in effect a const, of its declared bound.
		value, isDeclared := known.declared[bound.Local]
		if bound.Of != ir.Number || known.assigned[bound.Local] || !isDeclared || value == nil {
			return 0, 0, false
		}
		return boundRange(value, known, depth+1)
	case ir.Unary:
		operandLow, operandHigh, isWhole := boundRange(bound.Operand, known, depth+1)
		switch {
		case !isWhole:
			return 0, 0, false
		case bound.Operator == ir.Negate:
			low, high = -operandHigh, -operandLow
		case bound.Operator == ir.Plus:
			low, high = operandLow, operandHigh
		default:
			return 0, 0, false
		}
	case ir.Binary:
		leftLow, leftHigh, isLeft := boundRange(bound.Left, known, depth+1)
		rightLow, rightHigh, isRight := boundRange(bound.Right, known, depth+1)
		if !isLeft || !isRight {
			return 0, 0, false
		}
		switch bound.Operator {
		case ir.Add:
			low, high = leftLow+rightLow, leftHigh+rightHigh
		case ir.Subtract:
			low, high = leftLow-rightHigh, leftHigh-rightLow
		case ir.Multiply:
			products := []float64{leftLow * rightLow, leftLow * rightHigh, leftHigh * rightLow, leftHigh * rightHigh}
			low, high = products[0], products[0]
			for _, product := range products[1:] {
				low, high = math.Min(low, product), math.Max(high, product)
			}
		default:
			// Division and remainder needn't give a whole number, and ** is V8's pow, not Go's.
			return 0, 0, false
		}
	default:
		return 0, 0, false
	}
	// Whole numbers within 2^53, at both ends, so every value between, and every sum and product on
	// the way to them, is a whole number a double holds exactly.
	if !whole(low, exact) || !whole(high, exact) || low > high {
		return 0, 0, false
	}
	return low, high, true
}

// neverNegativeZero reports whether an expression boundRange accepts can never be -0: a sum or a
// difference is -0 only when one side is, while a negation or a product of a zero is -0 when its
// signs say so, so those are refused rather than followed.
func neverNegativeZero(expression ir.Expression, known facts, depth int) bool {
	if depth > 8 {
		return false
	}
	switch expression := expression.(type) {
	case ir.Length, ir.StringLength, ir.MapSize:
		return true
	case ir.NumberConstant:
		return !(expression.Value == 0 && math.Signbit(expression.Value))
	case ir.Read:
		// A counter is never -0: it starts as something that isn't, and x + step and x - step are -0
		// only when x is.
		if _, isCounter := known.counters[expression.Local]; isCounter {
			return true
		}
		value, isDeclared := known.declared[expression.Local]
		if expression.Of != ir.Number || known.assigned[expression.Local] || !isDeclared || value == nil {
			return false
		}
		return neverNegativeZero(value, known, depth+1)
	case ir.Unary:
		if expression.Operator == ir.Negate {
			// A constant negated is a constant, and says what it is.
			value, isConstant := constantOf(expression)
			return isConstant && !(value == 0 && math.Signbit(value))
		}
		return expression.Operator == ir.Plus && neverNegativeZero(expression.Operand, known, depth+1)
	case ir.Binary:
		switch expression.Operator {
		case ir.Add, ir.Subtract:
			return neverNegativeZero(expression.Left, known, depth+1) && neverNegativeZero(expression.Right, known, depth+1)
		case ir.Multiply:
			value, isConstant := constantOf(expression)
			return isConstant && !(value == 0 && math.Signbit(value))
		}
	}
	return false
}

// constantOf is a constant expression's value: a number, negated or not, or two of those added,
// subtracted or multiplied, which Go's doubles and C's (built with -ffp-contract=off) round the same.
// ** isn't, since V8's pow, which the runtime carries, can differ from Go's in the last bit.
func constantOf(expression ir.Expression) (float64, bool) {
	switch expression := expression.(type) {
	case ir.NumberConstant:
		return expression.Value, true
	case ir.Unary:
		operand, isConstant := constantOf(expression.Operand)
		switch {
		case !isConstant:
			return 0, false
		case expression.Operator == ir.Negate:
			return -operand, true
		case expression.Operator == ir.Plus:
			return operand, true
		}
	case ir.Binary:
		left, isLeft := constantOf(expression.Left)
		right, isRight := constantOf(expression.Right)
		if !isLeft || !isRight {
			return 0, false
		}
		switch expression.Operator {
		case ir.Add:
			return left + right, true
		case ir.Subtract:
			return left - right, true
		case ir.Multiply:
			return left * right, true
		}
	}
	return 0, false
}

// whole reports whether a value is a whole number of magnitude no larger than limit: not NaN, not
// infinite, not a fraction.
func whole(value float64, limit float64) bool {
	return value == math.Trunc(value) && math.Abs(value) <= limit
}

// stepOf is how far a counter's update moves it: counter = counter + step or counter - step, which
// counter++, counter--, counter += step and counter -= step lower to, for a whole constant step of 1
// or more (negative counting down), and false for any other update.
func stepOf(statement ir.Statement, counter int) (float64, bool) {
	assign, isAssign := statement.(ir.Assign)
	if !isAssign || assign.Local != counter || assign.Checked {
		return 0, false
	}
	sum, isBinary := assign.Value.(ir.Binary)
	if !isBinary || (sum.Operator != ir.Add && sum.Operator != ir.Subtract) {
		return 0, false
	}
	read, isRead := sum.Left.(ir.Read)
	step, isConstant := constantOf(sum.Right)
	if !isRead || read.Local != counter || !isConstant || !whole(step, exact) || step < 1 {
		return 0, false
	}
	if sum.Operator == ir.Subtract {
		step = -step
	}
	return step, true
}

func perIteration(loop ir.Loop, local int) bool {
	for _, each := range loop.PerIteration {
		if each == local {
			return true
		}
	}
	return false
}

func declaredIn(block ir.Block, local int) (ir.Expression, bool) {
	for _, statement := range block.Body {
		if declare, isDeclare := statement.(ir.Declare); isDeclare && declare.Local == local {
			return declare.Value, true
		}
	}
	return nil, false
}

// findDeclared records every local's declared value, however deeply nested.
func findDeclared(statements []ir.Statement, declared map[int]ir.Expression) {
	for _, statement := range statements {
		if declare, isDeclare := statement.(ir.Declare); isDeclare {
			declared[declare.Local] = declare.Value
		}
		eachNested(statement, func(nested []ir.Statement) { findDeclared(nested, declared) })
	}
}

// eachBlock calls visit with every block in statements, however deeply nested.
func eachBlock(statements []ir.Statement, visit func(ir.Block)) {
	for _, statement := range statements {
		if block, isBlock := statement.(ir.Block); isBlock {
			visit(block)
		}
		eachNested(statement, func(nested []ir.Statement) { eachBlock(nested, visit) })
	}
}

// eachNested calls visit with each list of statements a statement holds.
func eachNested(statement ir.Statement, visit func([]ir.Statement)) {
	switch statement := statement.(type) {
	case ir.If:
		visit(statement.Then)
		visit(statement.Else)
	case ir.Loop:
		visit(statement.Body)
		visit(statement.Update)
	case ir.Labeled:
		visit(statement.Body)
	case ir.Block:
		visit(statement.Body)
	case ir.ForOf:
		visit(statement.Body)
	case ir.Switch:
		for _, switchCase := range statement.Cases {
			visit(switchCase.Body)
		}
		visit(statement.Default)
	case ir.Try:
		visit(statement.Body)
		visit(statement.Catch)
		visit(statement.Finally)
	}
}
