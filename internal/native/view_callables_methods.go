package native

import (
	"github.com/system-inc/adamic/internal/ir"
)

// A resolved method pointer is callable, but its independently recorded shape
// still has to match before call arguments are evaluated. The witness below is
// only a local input to the shape validator, never an allocated/called closure.
func (e *emitter) emitViewCallableMethodCertificate(property ir.Property, method string) {
	e.declarations = append(e.declarations, "#include \"view_callables_contract.h\"")
	expected := e.viewCallableExpected(property)
	recorded := e.temporary()
	e.line("const adamic_callable_signature *%s = NULL;", recorded)
	seen := map[int]bool{}
	for _, class := range e.program.Classes {
		for _, index := range class.Methods {
			if seen[index] {
				continue
			}
			seen[index] = true
			function := e.program.Functions[index]
			if function.Closure || len(function.Parameters) == 0 || !e.dispatchable(index) {
				continue
			}
			parameters := make([]ir.Type, len(function.Parameters)-1)
			for i, local := range function.Parameters[1:] {
				parameters[i] = e.program.Locals[local].Type
			}
			signature := e.viewCallableSignature(parameters, function.Returns, function.Name)
			thunk := e.methodThunk(index)
			e.line("if (%s == %s) %s = %s;", method, thunk, recorded, signature)
		}
	}
	witness := e.temporary()
	e.line("adamic_closure %s = {.heap = {.references = 0, .kind = adamic_kind_closure, .slab = 0}, .code = NULL, .receiver = false, .count = 0};", witness)
	e.line("(void)adamic_view_callable_shape(&%s.heap, %s, %s, %s, false);", witness, recorded, expected, cString(property.View))
}
