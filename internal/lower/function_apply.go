package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
	"github.com/system-inc/adamic/internal/load"
)

// A dense apply literal preserves actual argument presence, including an empty
// list versus a supplied undefined. Other receivers and argument lists need a
// separate proof before they can enter this convention.
func (l *lowering) functionApply(node *ast.Node) (ir.Expression, bool, error) {
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || callee.Name().Text() != "apply" {
		return nil, false, nil
	}
	target := callee.AsPropertyAccessExpression().Expression
	proven := l.checker.GetTypeAtLocation(target)
	if held, known := l.representation(proven); !known || held != ir.Closure {
		return nil, false, nil
	}
	member := l.checker.GetSymbolAtLocation(callee)
	if member == nil || len(member.Declarations) == 0 || !load.IsLibrary(ast.GetSourceFileOfNode(member.Declarations[0])) {
		return nil, false, nil
	}
	args := call.Arguments.Nodes
	if len(args) != 2 || hasSpread(node) {
		return nil, true, l.notYet(node, "function apply without a proven dense argument literal")
	}
	receiver := ast.SkipParentheses(args[0])
	_, localReceiver := l.local(receiver)
	if !ast.IsIdentifier(receiver) || receiver.Text() != "undefined" || localReceiver {
		return nil, true, l.notYet(node, "function apply with a this receiver (receiver binding is not proven)")
	}
	list := ast.SkipParentheses(args[1])
	if list.Kind != ast.KindArrayLiteralExpression {
		return nil, true, l.notYet(node, "function apply without a proven dense argument literal")
	}
	for _, item := range list.AsArrayLiteralExpression().Elements.Nodes {
		if item.Kind == ast.KindSpreadElement || item.Kind == ast.KindOmittedExpression {
			return nil, true, l.notYet(node, "function apply with holes or spread")
		}
	}
	signatures := l.checker.GetSignaturesOfType(proven, checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, true, l.notYet(node, "function apply without one proven call signature")
	}
	closure, err := l.expression(target)
	if err != nil {
		return nil, true, err
	}
	arguments, _, err := l.callArguments(list.AsArrayLiteralExpression().Elements.Nodes)
	if err != nil {
		return nil, true, err
	}
	parameters := signatures[0].Parameters()
	for index := range arguments {
		if index >= len(parameters) {
			return nil, true, l.notYet(node, "function apply beyond its fixed parameter signature")
		}
		parameter := parameters[index]
		if len(parameter.Declarations) > 0 && parameter.Declarations[0].Kind == ast.KindParameter && parameter.Declarations[0].AsParameterDeclaration().DotDotDotToken != nil {
			return nil, true, l.notYet(node, "function apply with rest parameters")
		}
		takes, known := l.representation(l.checker.GetTypeOfSymbol(parameter))
		if !known {
			return nil, true, l.notYet(node, "function apply argument without a proven representation")
		}
		if len(parameter.Declarations) > 0 && parameter.Declarations[0].Kind == ast.KindParameter && parameter.Declarations[0].AsParameterDeclaration().Initializer != nil {
			takes = ir.Maybe(takes)
		}
		arguments[index] = fit(arguments[index], takes)
		if censusCallableSlotless(arguments[index].Type()) {
			return nil, true, l.notYet(node, "function apply with an unsupported argument representation")
		}
	}
	result := l.checker.GetTypeAtLocation(node)
	var returns ir.Type
	if result.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) == 0 {
		var known bool
		returns, known = l.representation(result)
		if !known || censusCallableSlotless(returns) {
			return nil, true, l.notYet(node, "function apply without a proven result representation")
		}
	}
	return ir.CallClosure{Closure: closure, Arguments: arguments, FunctionType: int(l.concrete(proven).Id()), Returns: returns}, true, nil
}
