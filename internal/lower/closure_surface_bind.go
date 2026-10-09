package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// All factory operands evaluate before its prefix check. Captured cells keep
// the callable, receiver and argument values alive after the factory returns.
func (l *lowering) bindCallable(node *ast.Node, proven *checker.Type, signature *checker.Signature, source, receiver ir.Expression, arguments []ir.Expression, returns ir.Type) (ir.Expression, bool, error) {
	contract := l.viewCallableInvocationContract(node, proven, []*checker.Signature{signature})
	if contract == 0 || len(arguments) > len(signature.Parameters()) {
		return nil, true, &Refused{Where: l.program.Where(node), What: "bind without a runtime-checkable fixed callable contract", Fix: "prove the callable parameter and result relation and bind a fixed argument prefix"}
	}
	maker := len(l.result.Functions)
	bound := maker + 1
	l.result.Functions = append(l.result.Functions, ir.Function{Name: "checked_bind_factory", Returns: ir.Closure}, ir.Function{Name: "checked_bound", Closure: true, Returns: returns, ReadsArguments: true, Surface: &ir.FunctionSurface{}})
	operands := append([]ir.Expression{source, receiver}, arguments...)
	captures := []int{}
	reads := []ir.Expression{}
	for i, value := range operands {
		local := len(l.result.Locals)
		name := "bound_argument"
		if i == 0 {
			name = "bound_callable"
		}
		if i == 1 {
			name = "bound_receiver"
		}
		l.result.Locals = append(l.result.Locals, ir.Local{Name: name, Type: value.Type(), Function: maker, Captured: true})
		captures = append(captures, local)
		reads = append(reads, ir.Read{Local: local, Of: value.Type()})
	}
	where := l.program.Where(node)
	callee := ast.SkipParentheses(node.AsCallExpression().Expression).AsPropertyAccessExpression().Expression
	view := sourceExpression(callee)
	l.result.Functions[maker].Parameters = captures
	check := ir.CallClosure{CheckBound: true, SurfaceOperation: "bind", BoundView: view, CallWhere: where, CallContract: contract, Closure: reads[0], Receiver: reads[1], Arguments: reads[2:], FunctionType: int(proven.Id())}
	l.result.Functions[maker].Body = []ir.Statement{ir.Evaluate{Value: check}, ir.Return{Value: ir.MakeClosure{Function: bound}}}
	function := l.result.Functions[bound]
	function.Environment = captures
	function.Bound = &ir.BoundCallable{Source: captures[0], Arguments: captures[2:]}
	function.OptionalParameters = map[int]bool{}
	forwarded := append([]ir.Expression{}, reads[2:]...)
	for _, parameter := range signature.Parameters()[len(arguments):] {
		takes, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		if !known || censusCallableSlotless(takes) {
			return nil, true, l.notYet(node, "a bound parameter without a callable ABI")
		}
		local := len(l.result.Locals)
		l.result.Locals = append(l.result.Locals, ir.Local{Name: "remaining_argument", Type: takes, Function: bound})
		function.Parameters = append(function.Parameters, local)
		forwarded = append(forwarded, ir.Read{Local: local, Of: takes})
		if l.includesUndefined(l.checker.GetTypeOfSymbol(parameter)) {
			function.OptionalParameters[local] = true
		}
	}
	// The captured producer remains the source of truth. Do not give this
	// wrapper producer domains derived from the wider view signature.
	count := len(l.result.Locals)
	l.result.Locals = append(l.result.Locals, ir.Local{Name: "remaining_argument_count", Type: ir.Number, Function: bound})
	function.ArgumentsCount = count + 1
	actual := ir.Binary{Operator: ir.Add, Left: ir.NumberConstant{Value: float64(len(arguments))}, Right: ir.Read{Local: count, Of: ir.Number}}
	call := ir.CallClosure{CallWhere: where, Closure: reads[0], Receiver: reads[1], Arguments: forwarded, ArgumentCount: actual, FunctionType: int(proven.Id()), Returns: returns}
	if returns == 0 {
		function.Body = []ir.Statement{ir.Evaluate{Value: call}}
	} else {
		function.Body = []ir.Statement{ir.Return{Value: call}}
	}
	l.result.Functions[bound] = function
	l.closureRecords = append(l.closureRecords, closureRecord{proven: l.concrete(l.checker.GetTypeAtLocation(node)), function: bound, node: node})
	return ir.Call{Function: maker, Arguments: operands, Returns: ir.Closure}, true, nil
}
