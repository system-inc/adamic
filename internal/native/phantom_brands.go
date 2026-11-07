package native

import "github.com/system-inc/adamic/internal/ir"

func (nativeEmitter *emitter) phantomMember(expression ir.PhantomMember) string {
	value := nativeEmitter.value(expression.Value)
	if !expression.Optional && (expression.Value.Type().IsReference() || expression.Value.Type().IsMaybe()) {
		test := value + " == NULL"
		if expression.Value.Type().IsMaybe() {
			test = "!(" + value + ").present"
		}
		nativeEmitter.line("if (%s) {", test)
		nativeEmitter.indent++
		errorValue := nativeEmitter.makeError(expression.Failure)
		nativeEmitter.line("adamic_thrown = adamic_retain(%s);", errorValue)
		// The error temporary only exists on the throwing path, in this branch. The check releases
		// it as a hold, first, as it would the statement's last temporary; neither the statement's
		// cleanup nor its ladder may name it past the branch.
		nativeEmitter.owned = nativeEmitter.owned[:len(nativeEmitter.owned)-1]
		nativeEmitter.checkThrown(errorValue)
		nativeEmitter.indent--
		nativeEmitter.line("}")
	} else {
		nativeEmitter.line("(void)%s;", value)
	}
	return "NULL"
}
