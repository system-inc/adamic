package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

// Zero or one formal parameter can observe only the whole match. Additional
// capture/index/input/group parameters require proving their varying layout.
func (l *lowering) regexReplacementCallback(node *ast.Node) bool {
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return false
	}
	signature := signatures[0]
	if result, known := l.representation(l.checker.GetReturnTypeOfSignature(signature)); !known || result != ir.String {
		return false
	}
	parameters := signature.Parameters()
	if len(parameters) > 1 {
		return false
	}
	if len(parameters) == 1 {
		parameter := parameters[0]
		if of, known := l.representation(l.checker.GetTypeOfSymbol(parameter)); !known || of != ir.String {
			return false
		}
		for _, declaration := range parameter.Declarations {
			if declaration.Kind == ast.KindParameter && declaration.AsParameterDeclaration().DotDotDotToken != nil {
				return false
			}
		}
	}
	return true
}
