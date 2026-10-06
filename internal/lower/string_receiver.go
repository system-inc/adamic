package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/system-inc/adamic/internal/ir"
)

// A prototype call evaluates every argument before checking its receiver. The helper holds each
// operand once, then checks the raw receiver before ToString can spell undefined as a string.
func (l *lowering) stringPrototypeCall(call, node *ast.Node, method string, written []*ast.Node) (ir.Expression, bool, error) {
	value, err := l.expression(node)
	if err != nil {
		return nil, true, err
	}
	if !l.mayBeUndefined(node) && !l.includesNull(l.checker.GetTypeAtLocation(node)) {
		converted, err := l.stringConverted(node, value)
		if err != nil {
			return nil, true, err
		}
		return l.libraryStringMethod(call, converted, method, written)
	}
	function, reads := l.stringHelper("receiver", []ir.Expression{value})
	converted, err := l.stringConverted(node, reads[0])
	if err != nil {
		return nil, true, err
	}
	errorLocal := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "type_error", Type: ir.Object, Function: function})
	readError := ir.Read{Local: errorLocal, Of: ir.Object}
	l.writeSites = append(l.writeSites, writeSite{node: node})
	condition := ir.Expression(ir.IsUndefined{Value: reads[0]})
	if l.includesNull(l.checker.GetTypeAtLocation(node)) {
		condition = ir.IsNull{Value: reads[0]}
	}
	l.result.Functions[function].Body = []ir.Statement{
		ir.If{Condition: condition, Then: []ir.Statement{
			ir.Declare{Local: errorLocal, Value: ir.MakeError{Message: ir.StringConstant{Index: l.constant("String.prototype." + method + " called on null or undefined")}}},
			ir.SetProperty{Object: readError, Name: "name", Value: ir.StringConstant{Index: l.constant("TypeError")}, Site: len(l.writeSites)},
			ir.Throw{Value: readError},
		}},
	}
	receiver := &stringReceiver{lowering: l, function: function, values: []ir.Expression{value}}
	return l.libraryStringMethodWith(call, converted, method, written, receiver)
}

// stringReceiver supplies the method with reads of arguments evaluated at the helper's call site.
type stringReceiver struct {
	lowering *lowering
	function int
	values   []ir.Expression
}

func (receiver *stringReceiver) argument(value ir.Expression) ir.Expression {
	l := receiver.lowering
	local := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "argument", Type: value.Type(), Function: receiver.function})
	l.result.Functions[receiver.function].Parameters = append(l.result.Functions[receiver.function].Parameters, local)
	receiver.values = append(receiver.values, value)
	return ir.Read{Local: local, Of: value.Type()}
}

func (receiver *stringReceiver) finish(result ir.Expression) ir.Expression {
	function := &receiver.lowering.result.Functions[receiver.function]
	function.Body = append(function.Body, ir.Return{Value: result})
	return ir.Call{Function: receiver.function, Arguments: receiver.values, Returns: ir.String}
}
