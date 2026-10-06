package lower

import (
	"math"

	"github.com/system-inc/adamic/internal/ir"
)

// counters marks the loop counters the native backend may keep in an integer (ir.Local.Counter).
//
// A number in JavaScript is a double, and a loop counter is one too. But a counter that starts at a
// whole number, only ever has 1 added to it, and runs while it's below (or at) a bound that's always a
// whole number no larger than 2^53 never holds anything but a whole number no larger than 2^53, and
// every one of those is a double exactly. So an integer holds the same values, and read as a double
// it's the same number: the same answers, for an integer's increment and an integer index into an
// array. The loop has to be one this can see whole:
//
//		for (let counter = start; counter < bound; counter++)    (or <=, or counter += 1)
//
//	  - start is a whole-number constant within 2^53 that isn't -0, since an integer has no -0 and
//	    1 / counter would tell;
//	  - nothing but the update writes the counter, and no closure captures it (a cell holds a double);
//	  - bound is, every time the condition reads it, a whole number: an array's, a string's or a map's
//	    size, a whole-number constant, or a variable nothing writes but its declaration, of one of those.
//	    A constant must be no larger than 2^53 for <, and smaller for <=: at 2^53 JavaScript's counter + 1
//	    is the counter again, and the loop never ends, where an integer would go on. NaN, Infinity and
//	    fractions aren't whole numbers, so their loops keep a double, as does anything else.
//
// A global counter keeps a double too, since a function may run before its loop has.
func counters(program *ir.Program) {
	assigned := map[int]bool{}
	declared := map[int]ir.Expression{}
	for _, function := range program.Functions {
		findAssigned(function.Body, assigned)
		findDeclared(function.Body, declared)
	}
	findAssigned(program.Main, assigned)
	findDeclared(program.Main, declared)
	mark := func(statements []ir.Statement) {
		eachBlock(statements, func(block ir.Block) {
			markCounter(program, block, assigned, declared)
		})
	}
	for _, function := range program.Functions {
		mark(function.Body)
	}
	mark(program.Main)
}

// markCounter marks the counter of a for loop's block, if it has one this can keep in an integer.
func markCounter(program *ir.Program, block ir.Block, assigned map[int]bool, declared map[int]ir.Expression) {
	if len(block.Body) == 0 {
		return
	}
	loop, isLoop := block.Body[len(block.Body)-1].(ir.Loop)
	if !isLoop || loop.CheckAfter || len(loop.Update) != 1 {
		return
	}
	condition, isBinary := loop.Condition.(ir.Binary)
	if !isBinary || (condition.Operator != ir.Less && condition.Operator != ir.LessOrEqual) {
		return
	}
	counterRead, isRead := condition.Left.(ir.Read)
	if !isRead {
		return
	}
	counter := counterRead.Local
	local := program.Locals[counter]
	if local.Type != ir.Number || local.Global || local.Captured || !perIteration(loop, counter) {
		return
	}
	// Its start: a declaration in this block, of a whole-number constant within 2^53, not -0.
	start, isDeclared := declaredIn(block, counter)
	value, isConstant := constantOf(start)
	if !isDeclared || !isConstant || !whole(value, math.Ldexp(1, 53)) || (value == 0 && math.Signbit(value)) {
		return
	}
	// Its update, counter + 1, and nothing in the body that writes it.
	if !increments(loop.Update[0], counter) {
		return
	}
	written := map[int]bool{}
	findAssigned(loop.Body, written)
	if written[counter] {
		return
	}
	limit := math.Ldexp(1, 53)
	if condition.Operator == ir.LessOrEqual {
		// counter <= 2^53 would end at 2^53 + 1, which JavaScript's counter + 1 never reaches.
		limit = math.Nextafter(limit, 0)
	}
	if !wholeBound(condition.Right, limit, assigned, declared, 0) {
		return
	}
	program.Locals[counter].Counter = true
}

// wholeBound reports whether a bound is a whole number no larger than limit every time it's read.
func wholeBound(bound ir.Expression, limit float64, assigned map[int]bool, declared map[int]ir.Expression, depth int) bool {
	switch bound := bound.(type) {
	case ir.Length, ir.StringLength, ir.MapSize:
		// Sizes are whole, and far below 2^53.
		return true
	case ir.NumberConstant, ir.Unary, ir.Binary:
		value, isConstant := constantOf(bound)
		return isConstant && whole(value, limit)
	case ir.Read:
		// A variable nothing writes but its declaration, of a whole bound: in effect a const.
		value, isDeclared := declared[bound.Local]
		if bound.Of != ir.Number || assigned[bound.Local] || !isDeclared || value == nil || depth > 8 {
			return false
		}
		return wholeBound(value, limit, assigned, declared, depth+1)
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

// increments reports whether a statement is counter = counter + 1, which counter++, ++counter and
// counter += 1 all lower to.
func increments(statement ir.Statement, counter int) bool {
	assign, isAssign := statement.(ir.Assign)
	if !isAssign || assign.Local != counter || assign.Checked {
		return false
	}
	sum, isBinary := assign.Value.(ir.Binary)
	if !isBinary || sum.Operator != ir.Add {
		return false
	}
	read, isRead := sum.Left.(ir.Read)
	one, isConstant := sum.Right.(ir.NumberConstant)
	return isRead && read.Local == counter && isConstant && one.Value == 1
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
	case ir.Block:
		visit(statement.Body)
	case ir.ForOf:
		visit(statement.Body)
	case ir.Switch:
		for _, switchCase := range statement.Cases {
			visit(switchCase.Body)
		}
		visit(statement.Default)
	}
}
