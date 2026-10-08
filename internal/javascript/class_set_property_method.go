package javascript

import "github.com/system-inc/adamic/internal/ir"

const ownMethodMessage = "replacing an inherited method would add an own field to a fixed shape"

func (e *emitter) ownMethodStore(statement ir.SetProperty) {
	object, value := e.temporary(), e.temporary()
	e.line("const %s = %s;", object, e.value(statement.Object))
	e.line("const %s = %s;", value, e.value(statement.Value))
	e.line("if (%s === undefined) panic(%s);", object, quote("TypeError: Cannot set properties of undefined (setting '"+statement.Name+"')"))
	e.line("if (!Object.prototype.hasOwnProperty.call(%s, %s)) panic(%s);", object, quote(statement.Name), quote(ownMethodMessage))
	e.line("%s[%s] = %s;", object, quote(statement.Name), value)
}
