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
		e.checkThrown()
		e.indent--
		e.line("}")
		// The error temporary only exists on the throwing path. checkThrown has released it;
		// normal statement cleanup must not refer to a name scoped inside that branch.
		e.owned = e.owned[:len(e.owned)-1]
	} else {
		e.line("(void)%s;", value)
	}
	return "NULL"
}
