package native

import (
	"fmt"

	"github.com/system-inc/adamic/internal/ir"
)

// Exceptions by cleanup paths (docs/memory.md, "Exceptions, designed into counting"). The runtime
// has one pending-exception word, adamic_thrown. A throw sets it and jumps to its handler: the
// innermost try around it in this function, letting go of what the statement and every scope between
// hold, or the function's way out, letting go of everything the frame holds and returning a zero
// value. After a call to a function that can throw, the word is tested and the same jump made. Out of
// main, an error is a panic. A finally is emitted at every way out of its try.

// handler is one try being emitted: where a throw inside it lands, the scope depth its body starts
// at, and its finally, which every way out of it runs.
type handler struct {
	label      string
	depth      int
	finally    []ir.Statement
	hasFinally bool

	// jumped says something jumps to label: a try whose body can't throw has no landing to emit.
	jumped bool
}

// jumpThrown emits the jump for a pending exception, once the statement's temporaries are let go: to
// the innermost handler, or out of the function, or, out of main, the panic.
func (e *emitter) jumpThrown() {
	if len(e.handlers) > 0 {
		inner := e.handlers[len(e.handlers)-1]
		inner.jumped = true
		e.releaseScopes(inner.depth)
		e.line("goto %s;", inner.label)
		return
	}
	if e.function == nil {
		// Out of main: String(error) as a panic, stdout flushed first. What the program holds then
		// isn't let go, as with any panic.
		e.line("adamic_uncaught();")
		return
	}
	e.releaseScopes(e.functionDepth)
	switch {
	case e.function.Closure:
		// A function value returns a slot whatever its type; the caller tests the word before reading it.
		e.line("return (adamic_value){.number = 0};")
	case e.function.Returns == 0:
		e.line("return;")
	default:
		e.line("return %s;", zero(e.function.Returns))
	}
}

// checkThrown emits, after a call that can throw, the test of the pending word and, when it's set,
// the release of what holds names (what a loop around the call holds), then of every temporary the
// statement owns so far, and the jump.
func (e *emitter) checkThrown(holds ...string) {
	e.line("if (adamic_thrown != NULL) {")
	e.indent++
	for _, hold := range holds {
		e.line("adamic_release(%s);", hold)
	}
	for index := len(e.owned) - 1; index >= 0; index-- {
		e.line("adamic_release(%s);", e.owned[index])
	}
	for outer := len(e.outerOwned) - 1; outer >= 0; outer-- {
		for index := len(e.outerOwned[outer]) - 1; index >= 0; index-- {
			e.line("adamic_release(%s);", e.outerOwned[outer][index])
		}
	}
	if e.statementRegion != "" {
		// The statement's region ends on this way out of it too, or its blocks, and what its values
		// hold on the heap, leak (region.go).
		e.line("adamic_region_end(&%s);", e.statementRegion)
	}
	e.jumpThrown()
	e.indent--
	e.line("}")
}

// closureThrown is checkThrown after a call through a function value, written out or made by a loop
// the emitter writes (map, the visits, reduce, Array.from), when a function value in the program can
// throw: which one the call reaches isn't known. holds is what the loop holds across the call.
func (e *emitter) closureThrown(holds ...string) {
	if e.program.ClosuresMayThrow {
		e.checkThrown(holds...)
	}
}

// throwStatement emits throw: the error is the pending word's, and the jump is made.
func (e *emitter) throwStatement(statement ir.Throw) {
	value := e.value(statement.Value)
	e.line("adamic_thrown = adamic_retain(%s);", value)
	e.end()
	e.jumpThrown()
}

// finallies emits, for a way out that leaves the tries open from index on, each one's finally,
// innermost first, each with only the tries outside it open, so a throw from it goes on outward.
func (e *emitter) finallies(from int) {
	saved := e.handlers
	for index := len(saved) - 1; index >= from; index-- {
		if !saved[index].hasFinally {
			continue
		}
		e.handlers = saved[:index]
		e.line("{")
		e.nested(saved[index].finally, nil)
		e.line("}")
	}
	e.handlers = saved
}

// innerHandlers is the index of the first open try inside scope depth: the ones a break or continue
// to a loop at that depth leaves.
func (e *emitter) innerHandlers(depth int) int {
	index := len(e.handlers)
	for index > 0 && e.handlers[index-1].depth > depth {
		index--
	}
	return index
}

// tryStatement emits try, catch and finally:
//
//	{ body }            a throw lands at landing
//	{ finally }         the body finished
//	goto done;
//	landing:;
//	{ error = adamic_thrown; catch }   a throw lands at rethrow (with a finally) or goes outward
//	{ finally }         the catch finished
//	goto done;
//	rethrow:;           with a finally: the pending error held while finally runs, then on outward
//	done:;
func (e *emitter) tryStatement(statement ir.Try) {
	e.temporaries++
	number := e.temporaries
	landing, rethrow, done := fmt.Sprintf("adamic_landing_%d", number), fmt.Sprintf("adamic_rethrow_%d", number), fmt.Sprintf("adamic_done_%d", number)
	if !statement.HasCatch {
		// Without a catch, a throw in the body goes straight to the finally that runs on the way out.
		landing = rethrow
	}

	body := &handler{label: landing, depth: len(e.scopes), finally: statement.Finally, hasFinally: statement.HasFinally}
	e.handlers = append(e.handlers, body)
	e.line("{")
	e.nested(statement.Body, nil)
	e.line("}")
	e.handlers = e.handlers[:len(e.handlers)-1]
	if statement.HasFinally {
		e.line("{")
		e.nested(statement.Finally, nil)
		e.line("}")
	}
	e.line("goto %s;", done)

	// A catch nothing can land in, or a rethrow nothing jumps to, isn't emitted: clang refuses a label
	// nothing uses, and the code would be dead.
	rethrown := !statement.HasCatch && body.jumped
	if statement.HasCatch && body.jumped {
		e.line("%s:;", landing)
		caughtHandler := &handler{label: rethrow, depth: len(e.scopes), finally: statement.Finally, hasFinally: true}
		if statement.HasFinally {
			e.handlers = append(e.handlers, caughtHandler)
		}
		e.line("{")
		e.indent++
		e.scopes = append(e.scopes, nil)
		caught := e.temporary()
		// What was thrown is the catch's now; with nothing bound to it, it's let go at once.
		e.line("adamic_object *%s = adamic_thrown;", caught)
		e.line("adamic_thrown = NULL;")
		if statement.CatchLocal >= 0 {
			e.declareLocal(statement.CatchLocal, caught, true)
		} else {
			e.line("adamic_release(%s);", caught)
		}
		for _, inner := range statement.Catch {
			e.statement(inner)
		}
		e.releaseScopes(len(e.scopes) - 1)
		e.scopes = e.scopes[:len(e.scopes)-1]
		e.indent--
		e.line("}")
		if statement.HasFinally {
			e.handlers = e.handlers[:len(e.handlers)-1]
			rethrown = caughtHandler.jumped
			e.line("{")
			e.nested(statement.Finally, nil)
			e.line("}")
		}
		e.line("goto %s;", done)
	}

	if statement.HasFinally && rethrown {
		e.line("%s:;", rethrow)
		e.line("{")
		e.indent++
		e.scopes = append(e.scopes, nil)
		// The pending error waits in a scope of the finally's while it runs, so a throw from the
		// finally, which replaces it, lets go of it on the way out.
		pending := e.temporary()
		e.line("adamic_object *%s = adamic_thrown;", pending)
		e.line("adamic_thrown = NULL;")
		e.hold(pending)
		e.line("{")
		e.nested(statement.Finally, nil)
		e.line("}")
		e.line("adamic_thrown = %s;", pending)
		e.line("%s = NULL;", pending)
		e.jumpThrown()
		e.scopes = e.scopes[:len(e.scopes)-1]
		e.indent--
		e.line("}")
	}
	e.line("%s:;", done)
}
