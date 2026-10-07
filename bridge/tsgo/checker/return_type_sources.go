package checker

import (
	"fmt"
	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

// returnTypeSources exposes declared, contextual and inferred return types.
// It supplies both the written and awaited/iterator types, without a lint verdict.
func (p *Program) returnTypeSources(out *fields, c *checker.Checker, node *ast.Node, question string) (string, error) {
	if question != "return-type-sources" || !ast.IsFunctionLike(node) {
		return "", fmt.Errorf("return-type-sources requires a function")
	}
	flags := ast.GetFunctionFlags(node)
	async := flags&ast.FunctionFlagsAsync != 0
	generator := flags&ast.FunctionFlagsGenerator != 0
	out.yes(async)
	out.yes(generator)
	var declared, contextual, inferred []*checker.Type
	if annotation := node.Type(); annotation != nil {
		declared = append(declared, checker.Checker_getTypeFromTypeNode(c, annotation))
	}
	var contextType *checker.Type
	if node.Kind == ast.KindArrowFunction || node.Kind == ast.KindFunctionExpression {
		contextType = checker.Checker_getContextualType(c, node, checker.ContextFlagsSignature)
	} else if ast.IsObjectLiteralMethod(node) {
		contextType = c.GetContextualTypeForObjectLiteralElement(node, checker.ContextFlagsSignature)
	}
	if contextType != nil {
		parts := []*checker.Type{contextType}
		if contextType.Flags()&checker.TypeFlagsUnion != 0 {
			parts = contextType.Types()
		}
		ambiguous := false
		for _, part := range parts {
			apparent := checker.Checker_getApparentType(c, part)
			signatures := c.GetSignaturesOfType(apparent, checker.SignatureKindCall)
			if len(signatures) > 1 {
				ambiguous = true
				break
			}
			if len(signatures) == 1 {
				contextual = append(contextual, checker.Checker_getReturnTypeOfSignature(c, signatures[0]))
			}
		}
		if ambiguous {
			contextual = nil
		} else if len(contextual) > 1 {
			contextual = []*checker.Type{c.GetUnionType(contextual)}
		}
	}
	actual := c.GetTypeAtLocation(node)
	if node.Kind == ast.KindGetAccessor {
		inferred = append(inferred, actual)
	} else if actual != nil {
		for _, signature := range c.GetSignaturesOfType(actual, checker.SignatureKindCall) {
			inferred = append(inferred, checker.Checker_getReturnTypeOfSignature(c, signature))
		}
	}
	for _, group := range [][]*checker.Type{declared, contextual, inferred} {
		out.number(uint64(len(group)))
		for _, written := range group {
			resolved := written
			if generator {
				resolved = nil
				if written != nil && written.Flags()&checker.TypeFlagsObject != 0 && written.ObjectFlags()&checker.ObjectFlagsReference != 0 {
					args := checker.Checker_getTypeArguments(c, written)
					if len(args) > 1 {
						resolved = args[1]
					}
				}
			} else if async {
				resolved = checker.Checker_getAwaitedType(c, written)
			}
			out.yes(resolved != nil)
			if resolved == nil {
				continue
			}
			out.text(c.TypeToString(written))
			out.number(uint64(resolved.Flags()))
			parts := []*checker.Type{resolved}
			if resolved.Flags()&checker.TypeFlagsUnion != 0 {
				parts = resolved.Types()
			}
			out.number(uint64(len(parts)))
			for _, part := range parts {
				out.number(uint64(part.Flags()))
			}
		}
	}
	return out.String(), nil
}
