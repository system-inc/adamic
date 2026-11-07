package lower

import (
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
	"github.com/system-inc/adamic/internal/ir"
)

func (lowering *lowering) regexReplacement(
	node *ast.Node,
	input ir.Expression,
	arguments []ir.Expression,
	method string,
	written *ast.Node,
) (ir.Expression, bool, error) {
	signatures := lowering.checker.GetSignaturesOfType(
		lowering.checker.GetTypeAtLocation(written),
		checker.SignatureKindCall,
	)
	if len(signatures) != 1 {
		return nil, true, lowering.notYet(written, "RegExp replacement with multiple callback signatures")
	}
	signature := signatures[0]
	returned := lowering.checker.GetReturnTypeOfSignature(signature)
	result, known := lowering.representation(returned)
	if returned.Flags()&checker.TypeFlagsUndefined != 0 {
		result, known = ir.String, true
	}
	if returned.Flags()&checker.TypeFlagsNull != 0 {
		result, known = ir.Object, true
	}
	if returned.Flags()&(checker.TypeFlagsVoid|checker.TypeFlagsNever) != 0 {
		result, known = 0, true
	}
	if !known || result != 0 &&
		result != ir.String &&
		result != ir.Number &&
		result != ir.Boolean &&
		result != ir.Union &&
		!(result == ir.Object && returned.Flags()&checker.TypeFlagsNull != 0) {
		return nil, true, lowering.notYet(written, "RegExp replacement callback return conversion")
	}
	if result == ir.Union && !lowering.writable(lowering.checker.GetReturnTypeOfSignature(signature)) {
		return nil, true, lowering.notYet(written, "RegExp replacement callback ToPrimitive")
	}
	replacement := &ir.RegExpReplacement{Returns: result}
	for argument, parameter := range signature.Parameters() {
		proven := lowering.checker.GetTypeOfSymbol(parameter)
		rest := len(parameter.Declarations) == 1 &&
			parameter.Declarations[0].Kind == ast.KindParameter &&
			parameter.Declarations[0].AsParameterDeclaration().DotDotDotToken != nil
		if rest {
			proven = lowering.elementTypeOfArray(proven)
		}
		of, known := lowering.representation(proven)
		if !known || of != ir.String && of != ir.Number && of != ir.Object && of != ir.Union {
			return nil, true, lowering.notYet(written, "RegExp replacement callback argument representation")
		}
		for _, member := range lowering.definedMembers(proven) {
			represented, _ := lowering.representation(member)
			if represented != ir.Object {
				continue
			}
			for _, field := range lowering.checker.GetPropertiesOfType(member) {
				fieldType := lowering.checker.GetTypeOfSymbol(field)
				if lowering.withoutUndefined(fieldType).Flags()&checker.TypeFlagsString == 0 {
					return nil, true, lowering.notYet(
						written,
						"RegExp replacement named-group field other than string or undefined",
					)
				}
				replacement.Groups = append(replacement.Groups, ir.RegExpReplacementGroup{
					Argument: argument,
					Rest:     rest,
					Name:     field.Name,
					Optional: lowering.includesUndefined(fieldType),
				})
			}
		}
		kind := lowering.regexReplacementKind(proven, of)
		if rest {
			replacement.RestKind = kind
			replacement.Rest = of
		} else {
			replacement.Parameters = append(replacement.Parameters, of)
			replacement.Kinds = append(replacement.Kinds, kind)
		}
	}
	return ir.RegExpCall{
		Value:       input,
		Arguments:   arguments,
		Method:      method,
		Returns:     ir.String,
		Replacement: replacement,
	}, true, nil
}

// The intrinsic does not call through lib.d.ts's any[] signature: its adapter checks actual
// argument kinds before entering the typed closure, including dynamic capture counts.
func (lowering *lowering) regexReplacementArgument(node *ast.Node) bool {
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
	receiver, known := lowering.representation(
		lowering.checker.GetTypeAtLocation(callee.AsPropertyAccessExpression().Expression),
	)
	return known && receiver == ir.String && (name == "replace" || name == "replaceAll") && lowering.isLibraryType(
		lowering.checker.GetNonNullableType(lowering.checker.GetTypeAtLocation(call.Arguments.Nodes[0])),
		"RegExp",
	) && len(
		lowering.checker.GetSignaturesOfType(lowering.checker.GetTypeAtLocation(node), checker.SignatureKindCall),
	) == 1
}

// A union argument has a box and a mask of the actual kinds its declared type admits.
func (lowering *lowering) regexReplacementKind(proven *checker.Type, of ir.Type) int {
	kind := map[ir.Type]int{ir.String: 1, ir.Number: 2, ir.Object: 3}[of]
	if of == ir.Union {
		kind = 64
		for _, member := range lowering.definedMembers(proven) {
			represented, known := lowering.representation(member)
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
		if lowering.includesUndefined(proven) {
			kind |= 1
		}
	} else if lowering.includesUndefined(proven) {
		kind |= 16
	}
	return kind
}
