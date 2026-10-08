package lower

import (
	"math"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// Math's numeric operations ignore this. Give them ordinary callable identities,
// shared across reads, rather than the opaque token used by specialized adapters.
func (l *lowering) mathMethodValue(node *ast.Node, method libraryMethod) ir.Expression {
	identity := "Math." + method.name
	for index, function := range l.result.Functions {
		if function.Intrinsic == identity {
			return ir.Read{Local: l.forwarders[index], Of: ir.Closure}
		}
	}
	index := len(l.result.Functions)
	function := ir.Function{Name: "library_math_" + method.name, Intrinsic: identity, Closure: true, Returns: ir.Number}
	result := ir.MathCall{Function: method.name}
	count, _ := methodMathArity(method.name)
	if count < 0 {
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "arguments", Type: ir.Array, Function: index})
		function.Parameters = []int{local}
		function.Rest = ir.Number
		result.Spread = ir.Read{Local: local, Of: ir.Array}
	} else {
		for position := 0; position < count; position++ {
			local := len(l.result.Locals)
			l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: ir.MaybeNumber, Function: index})
			function.Parameters = append(function.Parameters, local)
			result.Arguments = append(result.Arguments, ir.Coalesce{Value: ir.Read{Local: local, Of: ir.MaybeNumber}, Fallback: ir.NumberConstant{Value: math.NaN()}, Of: ir.Number})
		}
	}
	function.Body = []ir.Statement{ir.Return{Value: result}}
	l.result.Functions = append(l.result.Functions, function)
	l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(l.checker.GetTypeAtLocation(node)), function: index, node: node})
	held := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: identity, Type: ir.Closure, Global: true, Function: -1})
	l.forwarderValues = append(l.forwarderValues, ir.Declare{Local: held, Value: ir.MakeClosure{Function: index}})
	if l.forwarders == nil {
		l.forwarders = map[int]int{}
	}
	l.forwarders[index] = held
	return ir.Read{Local: held, Of: ir.Closure}
}

func methodMathArity(name string) (int, bool) {
	switch name {
	case "clz32", "fround":
		return 1, true
	case "imul":
		return 2, true
	}
	count, supported := mathFunctions[name]
	return count, supported
}

// A checked alias must throw a catchable ReferenceError rather than bypass its
// declaration via the intrinsic identity, or use the native raw TDZ panic.
func (l *lowering) checkedMathAlias(local int) ir.Expression {
	b := l.libraryArrayBuilder([]ir.Expression{ir.Read{Local: local, Of: ir.Closure}})
	value := b.read(b.parameters[0])
	message := "Cannot access '" + l.result.Locals[local].Name + "' before initialization"
	b.body = []ir.Statement{ir.If{Condition: ir.IsUndefined{Value: value}, Then: []ir.Statement{ir.Throw{Value: ir.MakeError{Message: ir.StringConstant{Index: l.constant(message)}, Name: ir.StringConstant{Index: l.constant("ReferenceError")}}}}}}
	return b.finish("library_math_alias", value)
}
