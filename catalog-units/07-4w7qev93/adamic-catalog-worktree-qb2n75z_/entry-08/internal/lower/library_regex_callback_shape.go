package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	regex "github.com/system-inc/adamic/internal/regexp"
	"strings"
)

// Offset and input follow every capture. Prove zero captures from immutable producers,
// rather than assuming the callback's declared parameter types describe the regex.
func (l *lowering) regexCallbackNoCaptures(node *ast.Node, depth int) bool {
	if depth > 32 {
		return false
	}
	node = ast.SkipParentheses(node)
	pattern, flags := "", ""
	switch node.Kind {
	case ast.KindRegularExpressionLiteral:
		text := node.Text()
		end := strings.LastIndex(text, "/")
		if end <= 0 {
			return false
		}
		pattern, flags = text[1:end], text[end+1:]
	case ast.KindIdentifier:
		symbol := l.symbol(node)
		if symbol == nil || len(symbol.Declarations) != 1 {
			return false
		}
		declaration := symbol.Declarations[0]
		if declaration.Kind != ast.KindVariableDeclaration || declaration.Parent.Flags&ast.NodeFlagsConst == 0 || declaration.AsVariableDeclaration().Initializer == nil {
			return false
		}
		return l.regexCallbackNoCaptures(declaration.AsVariableDeclaration().Initializer, depth+1)
	case ast.KindConditionalExpression:
		conditional := node.AsConditionalExpression()
		return l.regexCallbackNoCaptures(conditional.WhenTrue, depth+1) && l.regexCallbackNoCaptures(conditional.WhenFalse, depth+1)
	default:
		return false
	}
	parsed, err := regex.Parse(pattern, flags)
	if err != nil {
		return false
	}
	var present func(regex.Node) bool
	present = func(node regex.Node) bool {
		switch n := node.(type) {
		case *regex.Disjunction:
			for _, alternative := range n.Alternatives {
				for _, term := range alternative.Terms {
					if !present(term) {
						return false
					}
				}
			}
		case *regex.Group:
			return n.Kind != regex.Capturing && present(n.Body)
		case *regex.Quantifier:
			return present(n.Atom)
		}
		return true
	}
	return present(parsed.Body)
}

// The intrinsic supplies a proven argument sequence, rather than lib.d.ts's any[] rest.
// This exemption applies only at the replacement argument, never to a stored function view.
func (l *lowering) regexCallbackIntrinsicView(node *ast.Node) bool {
	child := node
	for child.Parent != nil && child.Parent.Kind == ast.KindParenthesizedExpression {
		child = child.Parent
	}
	if child.Parent == nil || child.Parent.Kind != ast.KindCallExpression {
		return false
	}
	call := child.Parent.AsCallExpression()
	if call.Arguments == nil || len(call.Arguments.Nodes) != 2 || call.Arguments.Nodes[1] != child {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	access := callee.AsPropertyAccessExpression()
	if access.Name().Text() != "replace" && access.Name().Text() != "replaceAll" {
		return false
	}
	if l.checker.GetTypeAtLocation(access.Expression).Flags()&checker.TypeFlagsStringLike == 0 || !l.libraryMember(callee) {
		return false
	}
	if !l.isLibraryType(l.checker.GetTypeAtLocation(call.Arguments.Nodes[0]), "RegExp") || !l.regexCallbackNoCaptures(call.Arguments.Nodes[0], 0) {
		return false
	}
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall)
	if len(signatures) != 1 || len(signatures[0].Parameters()) < 2 || len(signatures[0].Parameters()) > 3 {
		return false
	}
	expected := []checker.TypeFlags{checker.TypeFlagsString, checker.TypeFlagsNumber, checker.TypeFlagsString}
	for i, parameter := range signatures[0].Parameters() {
		if l.checker.GetTypeOfSymbol(parameter).Flags() != expected[i] {
			return false
		}
	}
	return l.checker.GetReturnTypeOfSignature(signatures[0]).Flags() == checker.TypeFlagsString
}
