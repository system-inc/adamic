package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Surface invocation snapshots the callable, then thisArg, then every argument.
// It uses the callable's signature, never Function.call's erased generic ABI.
func (l *lowering) closureSurfaceCall(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return nil, false, nil
	}
	access := callee.AsPropertyAccessExpression()
	name := access.Name().Text()
	if name != "call" && name != "apply" {
		return nil, false, nil
	}
	proven := l.concrete(l.checker.GetTypeAtLocation(access.Expression))
	of, known := l.representation(proven)
	if !known || of != ir.Closure || l.librarySymbol(l.symbol(access.Expression)) {
		return nil, false, nil
	}
	refuse := func(what string) (ir.Expression, bool, error) { return nil, true, l.notYet(node, what) }
	if access.QuestionDotToken != nil || call.QuestionDotToken != nil {
		return refuse("optional call or apply on a function value")
	}
	signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].TypeParameters()) != 0 || signatures[0].HasRestParameter() {
		return refuse("call or apply without one concrete fixed callable signature")
	}
	nodes := call.Arguments.Nodes
	if len(nodes) == 0 {
		return refuse("call or apply without an explicit thisArg")
	}
	value, err := l.expression(access.Expression)
	if err != nil {
		return nil, true, err
	}
	receiver, err := l.expression(nodes[0])
	if err != nil {
		return nil, true, err
	}
	if receiver.Type() != ir.Object {
		if _, undefined := receiver.(ir.Undefined); !undefined && !l.includesNull(l.checker.GetTypeAtLocation(nodes[0])) {
			return refuse("call or apply with a thisArg without an object representation")
		}
		receiver = fit(receiver, ir.Object)
	}
	arguments := nodes[1:]
	if name == "apply" {
		if len(nodes) != 2 {
			return refuse("apply without exactly thisArg and a dense argument literal")
		}
		array := ast.SkipParentheses(nodes[1])
		if array.Kind == ast.KindAsExpression {
			array = ast.SkipParentheses(array.AsAsExpression().Expression)
		}
		if array.Kind != ast.KindArrayLiteralExpression {
			return refuse("apply without a dense argument literal; spell the arguments with call")
		}
		arguments = array.AsArrayLiteralExpression().Elements.Nodes
	}
	lowered := []ir.Expression{}
	parameters := signatures[0].Parameters()
	for i, argument := range arguments {
		if argument.Kind == ast.KindSpreadElement || argument.Kind == ast.KindOmittedExpression {
			return refuse("call or apply with a spread or missing argument slot")
		}
		argumentValue, err := l.expression(argument)
		if err != nil {
			return nil, true, err
		}
		if i < len(parameters) {
			takes, known := l.representation(l.checker.GetTypeOfSymbol(parameters[i]))
			if !known {
				return refuse("call or apply with an unrepresented parameter")
			}
			argumentValue = fit(argumentValue, takes)
		}
		if censusCallableSlotless(argumentValue.Type()) {
			return refuse("call or apply with a slotless argument")
		}
		lowered = append(lowered, argumentValue)
	}
	returns := ir.Type(0)
	result := l.checker.GetReturnTypeOfSignature(signatures[0])
	if result.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) == 0 {
		returns, known = l.representation(result)
		if !known || censusCallableSlotless(returns) {
			return refuse("call or apply with a slotless result")
		}
	}
	l.result.ViewAdapters = true
	return ir.CallClosure{CallWhere: l.program.Where(node), Closure: value, Receiver: receiver, Arguments: lowered, FunctionType: int(proven.Id()), Returns: returns}, true, nil
}
