package native

import (
	"github.com/system-inc/adamic/internal/ir"
)

// A resolved method pointer is callable, but its independently recorded shape
// still has to match before call arguments are evaluated. The witness below is
// only a local input to the shape validator, never an allocated/called closure.
func (e *emitter) emitViewCallableMethodCertificate(property ir.Property, method string, exact ...int) {
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
				if parameters[i] == ir.Object && len(function.CallableMasks) == len(function.Parameters)+1 && function.CallableMasks[i+1] == 0 {
					parameters[i] = 0
				}
			}
			signature := e.viewCallableSignature(parameters, e.viewCallableMethodProducerResult(property, function.Returns), function.Name)
			thunk := e.methodThunk(index)
			identity := method
			if len(exact) != 0 {
				if e.program.PackedCountNeeded(exact[0]) != e.program.PackedCountNeeded(index) {
					continue
				}
			} else if e.program.ClosureConventionNeeded() {
				if e.program.PackedCountNeeded(index) {
					identity = method + ".counted_code"
				} else {
					identity = method + ".code"
				}
			}
			e.line("if (%s == %s) %s = %s;", identity, thunk, recorded, signature)
		}
	}
	witness := e.temporary()
	e.line("adamic_closure %s = {.heap = {.references = 0, .kind = adamic_kind_closure, .slab = 0}, .code = NULL, .count = 0};", witness)
	e.line("(void)adamic_view_callable_shape(&%s.heap, %s, %s, %s, false);", witness, recorded, expected, cString(property.View))
}

// Method thunks do not yet use the scalar-to-union result adapter. Keep that
// unsupported variance checked at the callable read instead of trusting the word.
func (e *emitter) viewCallableMethodProducerResult(property ir.Property, actual ir.Type) ir.Type {
	if property.ViewContract != 0 {
		callable := e.program.ViewContracts[property.ViewContract-1]
		if callable.Result != 0 && e.program.ViewContracts[callable.Result-1].Of == ir.Union && !actual.IsReference() {
			return 0
		}
	}
	return viewCallableProducerResult(actual)
}
