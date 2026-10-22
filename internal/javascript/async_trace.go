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
