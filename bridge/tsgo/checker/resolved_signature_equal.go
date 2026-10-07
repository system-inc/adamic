package checker

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/microsoft/TypeScript/tsc/shim/ast"
	"github.com/microsoft/TypeScript/tsc/shim/checker"
)

func signatureNode(node *ast.Node) bool {
	switch node.Kind {
	case ast.KindCallExpression, ast.KindNewExpression, ast.KindTaggedTemplateExpression, ast.KindJsxOpeningElement, ast.KindJsxSelfClosingElement:
		return true
	}
	return false
}

// Pointer equality is the production checker's signature identity relation.
func (p *Program) resolvedSignatureEqual(out *fields, c *checker.Checker, node *ast.Node, file, question string) (string, error) {
	parts := strings.Split(question, "\n")
	if len(parts) != 4 || !signatureNode(node) {
		return "", fmt.Errorf("resolved-signature-equal requires two call-like nodes")
	}
	bounds := make([]uint64, 2)
	for i := range bounds {
		n, err := strconv.ParseUint(parts[i+1], 10, 64)
		if err != nil || strconv.FormatUint(n, 10) != parts[i+1] {
			return "", fmt.Errorf("invalid node bound")
		}
		bounds[i] = n
	}
	_, other, err := p.exact(file, bounds[0], bounds[1], parts[3])
	if err != nil {
		return "", err
	}
	if !signatureNode(other) {
		return "", fmt.Errorf("target is not call-like")
	}
	out.yes(c.GetResolvedSignature(node) == c.GetResolvedSignature(other))
	return out.String(), nil
}
