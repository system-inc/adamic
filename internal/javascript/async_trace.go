package javascript

import "github.com/system-inc/adamic/internal/ir"

// tracedAwait preserves the real await turn while handing the tracing activation back to
// its synchronous caller, then reattaching it on either resumption edge.
func (e *emitter) tracedAwait(at *ir.Statement, wait ir.Await, local int) {
	operand := e.temporary()
	e.line("const %s = %s;", operand, e.value(wait.Value))
	if local >= 0 {
		e.declare(local, "undefined")
	}
	e.line("%s;", e.options.Suspend())
	e.line("try {")
	e.indent++
	if local >= 0 {
		e.line("%s = await %s;", e.variable(local), operand)
	} else {
		e.line("await %s;", operand)
	}
	e.indent--
	e.line("} finally {")
	e.line("\t%s;", e.options.Resume())
	e.line("}")
	if local >= 0 {
		e.markLine(at, 1)
	}
}

// tracedYield is tracedAwait's twin for a generator: the step that ran the body to this yield gets
// the tracing activation back before the consumer goes on, and whatever comes back into the body
// here reattaches it, a next() going on after the yield or a return() closing the generator, whose
// finally then leaves through the function's own. The yield's point was marked before this, so it
// is the body's last point before it parks.
func (e *emitter) tracedYield(yield ir.Yield) {
	value := e.temporary()
	e.line("const %s = %s;", value, e.value(yield.Value))
	e.line("%s;", e.options.Suspend())
	e.line("try {")
	e.line("\tyield %s;", value)
	e.line("} finally {")
	e.line("\t%s;", e.options.Resume())
	e.line("}")
}
