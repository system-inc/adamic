package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// A method view shares the implementation's incoming and outgoing slots. A Weak
// handle cannot stand in for its target pointer without an explicit adapter.
func (l *lowering) sameWeakMethodSlots(from, to *checker.Type) bool {
	source := l.checker.GetSignaturesOfType(from, checker.SignatureKindCall)
	viewed := l.checker.GetSignaturesOfType(to, checker.SignatureKindCall)
	same := func(from, to *checker.Type) bool {
		source, _ := l.representation(l.concrete(from))
		viewed, _ := l.representation(l.concrete(to))
		return (source == ir.Weak) == (viewed == ir.Weak)
	}
	for _, sourceSignature := range source {
		for _, viewedSignature := range viewed {
			sourceParameters := sourceSignature.Parameters()
			viewedParameters := viewedSignature.Parameters()
			for index := 0; index < len(sourceParameters) && index < len(viewedParameters); index++ {
				if !same(l.checker.GetTypeOfSymbol(sourceParameters[index]), l.checker.GetTypeOfSymbol(viewedParameters[index])) {
					return false
				}
			}
			if !same(l.checker.GetReturnTypeOfSignature(sourceSignature), l.checker.GetReturnTypeOfSignature(viewedSignature)) {
				return false
			}
		}
	}
	return true
}

// Narrowing can leave Weak handles in an array whose callback parameter is a
// target pointer. Indexing and for...of already load those handles correctly;
// refuse only a callback call that needs the missing handle-to-target adapter.
func (l *lowering) weakArrayCallback(node *ast.Node) error {
	node = ast.SkipParentheses(node)
	if node.Kind != ast.KindCallExpression {
		return nil
	}
	call := node.AsCallExpression()
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression || !l.libraryMember(callee) || len(call.Arguments.Nodes) == 0 {
		return nil
	}
	parameter, count := 0, 1
	switch callee.Name().Text() {
	case "forEach", "find", "findIndex", "map", "some", "every", "filter":
	case "sort":
		count = 2
	case "reduce", "reduceRight":
		parameter = 1
	default:
		return nil
	}
	receiver := l.concrete(l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression))
	if !l.checker.IsArrayType(receiver) {
		return nil
	}
	arguments := l.typeArguments(receiver)
	if len(arguments) != 1 {
		return nil
	}
	element, known := l.kept(arguments[0])
	if !known || element != ir.Weak {
		return nil
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(call.Arguments.Nodes[0]), checker.SignatureKindCall)
	for _, signature := range signatures {
		parameters := signature.Parameters()
		for index := parameter; index < parameter+count && index < len(parameters); index++ {
			takes, known := l.representation(l.concrete(l.checker.GetTypeOfSymbol(parameters[index])))
			if known && takes != ir.Weak {
				return l.notYet(node, callee.Name().Text()+" callback over narrowed Weak elements whose parameter needs handle-to-target conversion; use a loop with an explicit narrowed copy")
			}
		}
	}
	return nil
}
