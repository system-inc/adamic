package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (l *lowering) regexReplacement(node *ast.Node, input ir.Expression, arguments []ir.Expression, method string, written *ast.Node) (ir.Expression, bool, error) {
	signatures := l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(written), checker.SignatureKindCall)
	if len(signatures) != 1 {
		return nil, true, l.notYet(written, "RegExp replacement with multiple callback signatures")
	}
	signature := signatures[0]
	returned := l.checker.GetReturnTypeOfSignature(signature)
	result, known := l.representation(returned)
	if returned.Flags()&checker.TypeFlagsUndefined != 0 {
		result, known = ir.String, true
	}
	if returned.Flags()&checker.TypeFlagsNull != 0 {
		result, known = ir.Object, true
	}
	if returned.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) != 0 {
		result, known = 0, true
	}
	if !known || result != 0 && result != ir.String && result != ir.Number && result != ir.Boolean && result != ir.Union && !(result == ir.Object && returned.Flags()&checker.TypeFlagsNull != 0) {
		return nil, true, l.notYet(written, "RegExp replacement callback return conversion")
	}
	if result == ir.Union && !l.writable(l.checker.GetReturnTypeOfSignature(signature)) {
		return nil, true, l.notYet(written, "RegExp replacement callback ToPrimitive")
	}
	replacement := &ir.RegExpReplacement{Returns: result}
	for argument, parameter := range signature.Parameters() {
		proven := l.checker.GetTypeOfSymbol(parameter)
		rest := len(parameter.Declarations) == 1 && parameter.Declarations[0].Kind == ast.KindParameter && parameter.Declarations[0].AsParameterDeclaration().DotDotDotToken != nil
		if rest {
			proven = l.checker.GetElementTypeOfArrayType(proven)
		}
		of, known := l.representation(proven)
		if !known || of != ir.String && of != ir.Number && of != ir.Object && of != ir.Union {
			return nil, true, l.notYet(written, "RegExp replacement callback argument representation")
		}
		for _, member := range l.definedMembers(proven) {
			represented, _ := l.representation(member)
			if represented != ir.Object {
				continue
			}
			for _, field := range l.checker.GetPropertiesOfType(member) {
				fieldType := l.checker.GetTypeOfSymbol(field)
				if l.withoutUndefined(fieldType).Flags()&checker.TypeFlagsString == 0 {
					return nil, true, l.notYet(written, "RegExp replacement named-group field other than string or undefined")
				}
				replacement.Groups = append(replacement.Groups, ir.RegExpReplacementGroup{Argument: argument, Rest: rest, Name: field.Name, Optional: l.includesUndefined(fieldType)})
			}
		}
		kind := l.regexReplacementKind(proven, of)
		if rest {
			replacement.RestKind = kind
			replacement.Rest = of
		} else {
			replacement.Parameters = append(replacement.Parameters, of)
			replacement.Kinds = append(replacement.Kinds, kind)
		}
	}
	return ir.RegExpCall{Value: input, Arguments: arguments, Method: method, Returns: ir.String, Replacement: replacement}, true, nil
}

// The intrinsic does not call through lib.d.ts's any[] signature: its adapter checks actual
// argument kinds before entering the typed closure, including dynamic capture counts.
func (l *lowering) regexReplacementArgument(node *ast.Node) bool {
	outer := node
	for outer.Parent != nil && outer.Parent.Kind == ast.KindParenthesizedExpression {
		outer = outer.Parent
	}
	parent := outer.Parent
	if parent == nil || parent.Kind != ast.KindCallExpression {
		return false
	}
	call := parent.AsCallExpression()
	if len(call.Arguments.Nodes) != 2 || call.Arguments.Nodes[1] != outer {
		return false
	}
	callee := ast.SkipParentheses(call.Expression)
	if callee.Kind != ast.KindPropertyAccessExpression {
		return false
	}
	name := callee.Name().Text()
	receiver, known := l.representation(l.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression))
	return known && receiver == ir.String && (name == "replace" || name == "replaceAll") && l.isLibraryType(l.checker.GetNonNullableType(l.checker.GetTypeAtLocation(call.Arguments.Nodes[0])), "RegExp") && len(l.checker.GetSignaturesOfType(l.checker.GetTypeAtLocation(node), checker.SignatureKindCall)) == 1
}

// A union argument has a box and a mask of the actual kinds its declared type admits.
func (l *lowering) regexReplacementKind(proven *checker.Type, of ir.Type) int {
	kind := map[ir.Type]int{ir.String: 1, ir.Number: 2, ir.Object: 3}[of]
	if of == ir.Union {
		kind = 64
		for _, member := range l.definedMembers(proven) {
			represented, known := l.representation(member)
			if known {
				switch represented {
				case ir.String:
					kind |= 2
				case ir.Number:
					kind |= 4
				case ir.Object:
					kind |= 8
				}
			}
		}
		if l.includesUndefined(proven) {
			kind |= 1
		}
	} else if l.includesUndefined(proven) {
		kind |= 16
	}
	return kind
}
