package native

import "github.com/system-inc/adamic/internal/ir"

func (e *emitter) phantomMember(expression ir.PhantomMember) string {
	value := e.value(expression.Value)
	if !expression.Optional && (expression.Value.Type().IsReference() || expression.Value.Type().IsMaybe()) {
		test := value + " == NULL"
		if expression.Value.Type().IsMaybe() {
			test = "!(" + value + ").present"
		}
		e.line("if (%s) {", test)
		e.indent++
		errorValue := e.makeError(expression.Failure)
		e.line("adamic_thrown = adamic_retain(%s);", errorValue)
		// The error temporary only exists on the throwing path, in this branch. The check releases
		// it as a hold, first, as it would the statement's last temporary; neither the statement's
		// cleanup nor its ladder may name it past the branch.
		e.owned = e.owned[:len(e.owned)-1]
		e.checkThrown(errorValue)
		e.indent--
		e.line("}")
	} else {
		e.line("(void)%s;", value)
	}
	return "NULL"
}
