package javascript

import "github.com/system-inc/adamic/internal/ir"

// tracedAwait preserves the real await turn while handing the tracing activation back to
// its synchronous caller, then reattaching it on either resumption edge.
func (javascriptEmitter *emitter) tracedAwait(at *ir.Statement, wait ir.Await, local int) {
	operand := javascriptEmitter.temporary()
	javascriptEmitter.line("const %s = %s;", operand, javascriptEmitter.value(wait.Value))
	if local >= 0 {
		javascriptEmitter.declare(local, "undefined")
	}
	javascriptEmitter.line("%s;", javascriptEmitter.options.Suspend())
	javascriptEmitter.line("try {")
	javascriptEmitter.indent++
	if local >= 0 {
		javascriptEmitter.line("%s = await %s;", javascriptEmitter.variable(local), operand)
	} else {
		javascriptEmitter.line("await %s;", operand)
	}
	javascriptEmitter.indent--
	javascriptEmitter.line("} finally {")
	javascriptEmitter.line("\t%s;", javascriptEmitter.options.Resume())
	javascriptEmitter.line("}")
	if local >= 0 {
		javascriptEmitter.markLine(at, 1)
	}
}
