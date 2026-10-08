package native

import "github.com/system-inc/adamic/internal/ir"

// The callee and receiver have already been read and retained. Only the
// present branch emits argument work and invokes the saved callable.
func (e *emitter) optionalClosureCall(call ir.CallClosure, closure, receiver, method string) string {
	of := call.Type()
	call.Optional = false
	text, value, owned := e.asideWith(func() string { return e.invokeClosure(call, closure, receiver, method) })
	present := "true"
	if closure != "" {
		present = closure + " != NULL"
		if method != "" {
			present += " || " + method + " != NULL"
		}
	}
	result := "0"
	if of != 0 {
		result = e.temporary()
		undefined := "NULL"
		if of.IsMaybe() {
			undefined = zero(of)
		}
		e.line("%s %s = %s;", cType(of), result, undefined)
	}
	e.line("if (%s) {", present)
	e.out.WriteString(text)
	e.indent++
	if of != 0 {
		value, _ = converted(call.Returns, of, value)
		if of.IsReference() {
			value = retained(value)
		}
		e.line("%s = %s;", result, value)
	}
	for index := len(owned) - 1; index >= 0; index-- {
		e.line("adamic_release(%s);", owned[index])
	}
	e.indent--
	e.line("}")
	if of.IsReference() {
		e.owned = append(e.owned, result)
	}
	return result
}
