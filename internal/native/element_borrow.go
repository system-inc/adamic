package native

import (
	"slices"

	"github.com/system-inc/adamic/internal/ir"
)

// A variable borrowed from an array (docs/memory.md, "A variable borrowed from an array"): a local
// given an element, a[i] or a[i] ?? fallback, holds no count when nothing in the rest of its block can
// take an element out of any array. The array stays alive because a names it for the whole function,
// and the element stays in it, so the array's count is the variable's.

// planElementBorrows finds element and strong-field borrowing declarations and loop bindings that
// borrow, keyed by statement, and marks
// each variable Borrowed, as a borrowed parameter is: reuse in place and moves into consumed
// parameters refuse those, since the count they'd take is the array's.
//
// It also gives the arrays they borrow from (lending). The borrow leans on the array staying alive in
// its variable for the rest of the block, and changes can't see a move: a call to a function that
// only reads its array still frees it when the parameter is consumed (borrow_element_super_move.a).
// Every possible call target must preserve the elements for the borrow to stand.
// Reuse never moves an array that lends (reusePlan.movable).
func planElementBorrows(program *ir.Program) (map[*ir.Statement]bool, map[int]bool) {
	changing := changingFunctions(program)
	borrows := map[*ir.Statement]bool{}
	lending := map[int]bool{}
	for index := range program.Functions {
		function := &program.Functions[index]
		if function.Closure {
			continue
		}
		assigned := assignedLocals(function.Body)
		var statements func([]ir.Statement)
		statements = func(list []ir.Statement) {
			for position := range list {
				declare, declaration := list[position].(ir.Declare)
				element := declaration && borrowable(program, index, declare, assigned) && !changes(program, changing, list[position+1:])
				chain := declaration && borrowableChain(program, index, declare, assigned)
				if element || chain {
					borrows[&list[position]] = true
					program.Locals[declare.Local].Borrowed = true
					lending[borrowedArray(declare)] = true
				}
				if loop, ok := list[position].(ir.ForOf); ok && loopBorrowable(program, index, loop, assigned) && !changes(program, changing, loop.Body) {
					borrows[&list[position]] = true
					program.Locals[loop.Local].Borrowed = true
					lending[loop.Iterable.(ir.Read).Local] = true
				}
				walkStatement(list[position], func(ir.Expression) {}, statements)
			}
		}
		statements(function.Body)
	}
	return borrows, lending
}

// borrowedArray is the root holder a borrowing declaration relies on, an array
// for an element borrow or the first object in a strong field chain.
func borrowedArray(declare ir.Declare) int {
	value := declare.Value
	if coalesce, ok := value.(ir.Coalesce); ok {
		value = coalesce.Value
	}
	if _, ok := value.(ir.Property); ok {
		root, _ := chainRoot(value, map[string]bool{})
		return root
	}
	return value.(ir.ArrayIndex).Array.(ir.Read).Local
}

// borrowable reports whether a declaration's variable could borrow: a local of the function, never
// assigned after and not captured, given an element of an array a local or parameter of the same
// function names, never assigned and not captured, directly or as the left of ??.
func borrowable(program *ir.Program, function int, declare ir.Declare, assigned map[int]bool) bool {
	local := program.Locals[declare.Local]
	if local.Global || local.Captured || local.Function != function || !lendable(local.Type) || assigned[declare.Local] {
		return false
	}
	value := declare.Value
	if coalesce, ok := value.(ir.Coalesce); ok {
		if coalesce.Panic != nil || coalesce.Fallback == nil || coalesce.Of != local.Type {
			return false
		}
		value = coalesce.Value
	}
	element, ok := value.(ir.ArrayIndex)
	if !ok || element.Element != local.Type {
		return false
	}
	array, ok := element.Array.(ir.Read)
	if !ok {
		return false
	}
	held := program.Locals[array.Local]
	return !held.Global && !held.Captured && held.Function == function && !assigned[array.Local]
}

// assignedLocals is every variable an assignment in a function's body writes.
func assignedLocals(body []ir.Statement) map[int]bool {
	assigned := map[int]bool{}
	var statements func([]ir.Statement)
	statements = func(list []ir.Statement) {
		for _, statement := range list {
			if assign, ok := statement.(ir.Assign); ok {
				assigned[assign.Local] = true
			}
			walkStatement(statement, func(ir.Expression) {}, statements)
		}
	}
	statements(body)
	return assigned
}

// changingFunctions is every function whose body, or anything it calls, can take an element out of
// an array: the least fixed point, so a recursion changes only where one of its bodies does.
func changingFunctions(program *ir.Program) map[int]bool {
	changing := map[int]bool{}
	for changed := true; changed; {
		changed = false
		for index := range program.Functions {
			if !changing[index] && changes(program, changing, program.Functions[index].Body) {
				changing[index] = true
				changed = true
			}
		}
	}
	return changing
}

// changes reports whether statements can take an element out of an array, as far as changing says
// the program's functions can: anything not on the list of what can't is a change.
func changes(program *ir.Program, changing map[int]bool, list []ir.Statement) bool {
	found := false
	var statements func([]ir.Statement)
	statements = func(list []ir.Statement) {
		for _, statement := range list {
			if _, isStore := statement.(ir.SetIndex); isStore {
				found = true
			}
			walkStatement(statement, func(expression ir.Expression) {
				if !unchanging(program, changing, expression) {
					found = true
				}
			}, statements)
		}
	}
	statements(list)
	return found
}

// unchanging reports whether an expression, on its own, can't take an element out of an array: a
// pure one (borrow.go), an object literal, an array literal with no spread (reuse could take one
// over), a push, an Error made without calling user code, a closure made but not called, or a call to a
// function that doesn't change any. The walk still checks every operand, including an Error's message.
func unchanging(program *ir.Program, changing map[int]bool, expression ir.Expression) bool {
	switch expression := expression.(type) {
	case ir.ObjectLiteral, ir.ArrayPush, ir.MakeClosure, ir.MakeError, ir.Defined:
		return true
	case ir.ArrayLiteral:
		for _, spread := range expression.Spread {
			if spread {
				return false
			}
		}
		return true
	case ir.Call:
		for _, target := range program.CallTargets(expression) {
			if changing[target] {
				return false
			}
		}
		return true
	case ir.CallClosure:
		targets := program.ClosureTargets(expression)
		if targets.Unknown {
			return false
		}
		for _, target := range targets.Functions {
			if changing[target] {
				return false
			}
		}
		return true
	}
	return pureKind(expression)
}

// borrowElement declares a variable that borrows an element: the element itself, with no count. With
// a ?? fallback, a missing element gives the fallback, which is fresh: a hidden owner holds it (NULL
// when the element was there), and the scope lets go of the owner.
func (e *emitter) borrowElement(declare ir.Declare) {
	local := e.program.Locals[declare.Local]
	if _, chain := declare.Value.(ir.Property); chain {
		outer := e.lendAt
		e.lendAt = e.depth + 1
		value := e.value(declare.Value)
		e.lendAt = outer
		e.line("%s %s = %s;", cType(local.Type), e.localName(declare.Local), value)
		e.end()
		return
	}
	value := declare.Value
	coalesce, hasFallback := value.(ir.Coalesce)
	if hasFallback {
		value = coalesce.Value
	}
	// Share the ordinary lookup's integer-counter and relative-index rules.
	slot := e.arrayIndexSlot(value.(ir.ArrayIndex))
	name := e.localName(declare.Local)
	e.line("%s %s = %s == NULL ? NULL : (%s)%s->reference;", cType(local.Type), name, slot, cType(local.Type), slot)
	if hasFallback {
		owner := name + "_owner"
		e.line("%s %s = NULL;", cType(local.Type), owner)
		e.line("if (%s == NULL) {", name)
		text, fallback, owned := e.aside(coalesce.Fallback)
		e.out.WriteString(text)
		e.indent++
		if index := slices.Index(owned, fallback); index >= 0 {
			owned = slices.Delete(slices.Clone(owned), index, index+1)
			e.line("%s = %s;", owner, fallback)
		} else {
			e.line("%s = %s;", owner, retained(fallback))
		}
		e.line("%s = %s;", name, owner)
		for index := len(owned) - 1; index >= 0; index-- {
			e.line("adamic_release(%s);", owned[index])
		}
		e.indent--
		e.line("}")
		if e.mostlyNull == nil {
			e.mostlyNull = map[string]bool{}
		}
		e.mostlyNull[owner] = true
		e.hold(owner)
	}
	e.end()
}

// loopBorrowable uses the same local and lending facts as indexed element declarations.
// Together with the body's changes check, this also proves the iterator needs no array count:
// an unassigned, uncaptured local or parameter keeps the array alive until the loop exits.
// Lending prevents moves into consumed parameters from taking that owner away. A throw or early
// exit ends the iterator before the enclosing variable's scope or the caller's hold ends.
func loopBorrowable(program *ir.Program, function int, loop ir.ForOf, assigned map[int]bool) bool {
	if loop.Iterable.Type() != ir.Array || loop.MapPart != "" || loop.RegexIterator || loop.Pattern != nil || loop.Element != program.Locals[loop.Local].Type {
		return false
	}
	return borrowable(program, function, ir.Declare{Local: loop.Local, Value: ir.ArrayIndex{Array: loop.Iterable, Element: loop.Element}}, assigned)
}

// borrowableChain lends a strong field to an uncaptured, unassigned local. The
// whole function is checked, including enclosing blocks after this block ends,
// so a later iteration or finally cannot invalidate the holder underneath it.
func borrowableChain(program *ir.Program, function int, declare ir.Declare, assigned map[int]bool) bool {
	if _, ok := declare.Value.(ir.Property); !ok {
		return false
	}
	local := program.Locals[declare.Local]
	if local.Global || local.Captured || local.Function != function || !lendable(local.Type) || assigned[declare.Local] {
		return false
	}
	names := map[string]bool{}
	root, ok := chainRoot(declare.Value, names)
	if !ok {
		return false
	}
	held := program.Locals[root]
	if held.Global || held.Captured || held.Function != function || assigned[root] {
		return false
	}
	return chainUnchanged(program, program.Functions[function].Body, names)
}
